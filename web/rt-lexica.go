//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/generic"
	"github.com/e-gun/HipparchiaGoServer/internal/launch"
	"github.com/e-gun/HipparchiaGoServer/internal/mps"
	"github.com/e-gun/HipparchiaGoServer/internal/search"
	"github.com/e-gun/HipparchiaGoServer/internal/structs"
	"github.com/e-gun/HipparchiaGoServer/internal/vaults"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"math"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
)

var (
	// IsGreek regex compiled here instead of inside various loops

	// quantityfixer = strings.NewReplacer("ă_", "ā̆", "ā^", "ā̆", "ē^", "ē̆", "ĭ_", "ī̆", "ō^", "ō̆", "A_^", "Ā̆", "A^", "Ă", "A_", "Ā", "E_", "Ē", "E^", "Ĕ", "I_^", "Ī̆", "I_", "Ī", "I^", "Ĭ", "O_", "Ō", "O^", "Ŏ", "U^", "Ŭ", "U_", "Ū", "_^", "̆̄", "_", "̄", "^", "̆")
	quantityfixer = strings.NewReplacer("_^", "̄̆", "_", "̄", "^", "̆")
)

// hipparchiaDB-# \d latin_morphology
//                           Table "public.latin_morphology"
//          Column           |          Type          | Collation | Nullable | Default
//---------------------------+------------------------+-----------+----------+---------
// observed_form             | character varying(64)  |           |          |
// xrefs                     | character varying(128) |           |          |
// prefixrefs                | character varying(128) |           |          |
// possible_dictionary_forms | jsonb                  |           |          |
// related_headwords         | character varying(256) |           |          |
//Indexes:
//    "latin_analysis_trgm_idx" gin (related_headwords gin_trgm_ops)
//    "latin_morphology_idx" btree (observed_form)

// hipparchiaDB-# \d latin_dictionary
//                     Table "public.latin_dictionary"
//     Column     |          Type          | Collation | Nullable | Default
//----------------+------------------------+-----------+----------+---------
// entry_name     | character varying(256) |           |          |
// metrical_entry | character varying(256) |           |          |
// id_number      | real                   |           |          |
// entry_key      | character varying(64)  |           |          |
// pos            | character varying(64)  |           |          |
// translations   | text                   |           |          |
// entry_body     | text                   |           |          |
// html_body      | text                   |           |          |
//Indexes:
//    "latin_dictionary_idx" btree (entry_name)

type JSB struct {
	HTML string `json:"newhtml"`
	JS   string `json:"newjs"`
}

//
// ROUTING
//

// RtLexLookup - search the dictionary for a headword substring
func RtLexLookup(c echo.Context) error {
	c.Response().After(func() { msg.LogPaths("RtLexLookup()") })

	user := ReadUUIDCookie(c)
	if !vaults.AllAuthorized.Check(user) {
		return generic.JSONresponse(c, JSB{JS: vv.JSVALIDATION})
	}

	req := c.Param("wd")
	seeking := generic.Purgechars(launch.Config.BadChars, req)
	seeking = generic.SwapAcuteForGrave(seeking)

	dict := "latin"
	if vv.IsGreek.MatchString(seeking) {
		dict = "greek"
	}

	seeking = generic.UVσςϲ(seeking)
	seeking = generic.UniversalPatternMaker(seeking) // UniversalPatternMaker() returns the term with brackets around it

	seeking = strings.Replace(seeking, "(", "", -1)
	seeking = strings.Replace(seeking, ")", "", -1)

	initialspace := regexp.MustCompile("^\\s")
	if initialspace.MatchString(seeking) {
		seeking = "^" + initialspace.ReplaceAllString(seeking, "")
	}

	terminalspace := regexp.MustCompile("\\s$")
	if terminalspace.MatchString(seeking) {
		seeking = terminalspace.ReplaceAllString(seeking, "") + "$"
	}

	html := dictsearch(seeking, dict)

	var jb JSB
	jb.HTML = html
	jb.JS = insertlexicaljs()

	return generic.JSONresponse(c, jb)
}

// RtLexFindByForm - search the dictionary for a specific headword
func RtLexFindByForm(c echo.Context) error {
	// be able to respond to "GET /lexica/findbyform/ἀμιϲθὶ/gr0062 HTTP/1.1"
	user := ReadUUIDCookie(c)
	if !vaults.AllAuthorized.Check(user) {
		return generic.JSONresponse(c, JSB{JS: vv.JSVALIDATION})
	}
	c.Response().After(func() { msg.LogPaths("RtLexFindByForm()") })

	req := c.Param("wd")
	elem := strings.Split(req, "/")

	if len(elem) == 0 || elem[0] == "" {
		return emptyjsreturn(c)
	}

	var au string
	if len(elem) == 1 {
		au = ""
	} else {
		au = elem[1]
	}

	word := generic.Purgechars(launch.Config.BadChars, elem[0])

	clean := strings.NewReplacer("-", "", "¹", "", "²", "", "³", "") // you can get sent here by the indexer ...
	word = clean.Replace(word)

	word = generic.SwapAcuteForGrave(word)
	word = generic.UVσςϲ(word)

	html := findbyform(word, au)
	js := insertlexicaljs()

	var jb JSB
	jb.HTML = html
	jb.JS = js

	return generic.JSONresponse(c, jb)
}

// RtLexId - grab a word by its entry value
func RtLexId(c echo.Context) error {
	// http://127.0.0.1:8000/lexica/idlookup/latin/24236.0
	const (
		FAIL1 = "RtLexId() received bad request: '%s'"
		FAIL2 = "RtLexId() found nothing at id_number '%s'"
	)

	user := ReadUUIDCookie(c)
	if !vaults.AllAuthorized.Check(user) {
		return generic.JSONresponse(c, JSB{JS: vv.JSVALIDATION})
	}

	req := c.Param("wd")
	elem := strings.Split(req, "/")
	if len(elem) != 2 {
		msg.WARN(fmt.Sprintf(FAIL1, req))
		return emptyjsreturn(c)
	}
	d := generic.Purgechars(launch.Config.BadChars, elem[0])
	w := generic.Purgechars(launch.Config.BadChars, elem[1])

	f := dictgrabber(w, d, "id_number", "=")
	if len(f) == 0 {
		msg.WARN(fmt.Sprintf(FAIL2, w))
		return emptyjsreturn(c)
	}

	html := formatlexicaloutput(f[0])
	js := insertlexicaljs()

	var jb JSB
	jb.HTML = html
	jb.JS = js

	return generic.JSONresponse(c, jb)
}

// RtLexReverse - look for the headwords that have the sought word in their body
func RtLexReverse(c echo.Context) error {
	// be able to respond to "/lexica/reverselookup/0ae94619/sorrow"
	c.Response().After(func() { msg.LogPaths("RtLexReverse()") })

	user := ReadUUIDCookie(c)
	if !vaults.AllAuthorized.Check(user) {
		return generic.JSONresponse(c, JSB{JS: vv.JSVALIDATION})
	}

	req := c.Param("wd")
	elem := strings.Split(req, "/")

	if len(elem) == 0 || elem[0] == "" {
		return emptyjsreturn(c)
	}

	word := generic.Purgechars(launch.Config.BadChars, elem[1])

	s := vaults.AllSessions.GetSess(user)

	var dd []string
	// map[string]bool{"gr": true, "lt": true, "in": false, "ch": false, "dp": false}
	if s.ActiveCorp[vv.LATINCORP] || s.ActiveCorp[vv.CHRISTINSC] {
		dd = append(dd, "latin")
	}

	if s.ActiveCorp[vv.GREEKCORP] || s.ActiveCorp[vv.INSCRIPTCORP] || s.ActiveCorp[vv.PAPYRUSCORP] {
		dd = append(dd, "greek")
	}

	if len(dd) == 0 {
		return emptyjsreturn(c)
	}

	html := reversefind(word, dd)

	var jb JSB
	jb.HTML = html
	jb.JS = insertlexicaljs()

	return generic.JSONresponse(c, jb)
}

//
// LOOKUPS
//

// findbyform - observed word into HTML dictionary entry
func findbyform(word string, author string) string {
	const (
		FLDS = `entry_name, total_count, gr_count, lt_count, dp_count, in_count, ch_count`
		PSQQ = `SELECT %s FROM wordcounts_%s where entry_name = '%s'`
		SRCH = `<bibl id="perseus/%s/`
		REPL = `<bibl class="flagged" id="perseus/%s/`
		NOTH = `findbyform() found no results for '%s'`
	)

	d := "latin"
	if vv.IsGreek.MatchString(word) {
		d = "greek"
	}

	// [a] search for morphology matches
	thesefinds := getmorphmatch(strings.ToLower(word), d)
	if len(thesefinds) == 0 {
		// Νέαιρα can be found, νέαιρα can't
		thesefinds = getmorphmatch(word, d)
	}

	if len(thesefinds) == 0 {
		return fmt.Sprintf("(no match for '%s' in the morphology lookup tables)", word)
	}

	// [b] turn morph matches into []MorphPossib

	mpp := dbmorphintomorphpossib(thesefinds)

	// [c] take the []MorphPossib and find the set of headwords we are interested in; store this in a []dblexicon

	lexicalfinds := morphpossibintolexpossib(d, mpp)

	// [d] generate and format the prevalence data for this form: cf formatprevalencedata() in lexicalformatting.py

	// golang hates indexing unicode strings: strings are bytes, and unicode chars take more than one byte
	c := []rune(word)
	q := fmt.Sprintf(PSQQ, FLDS, generic.StripaccentsSTR(string(c[0])), word)

	var wc structs.DbWordCount
	ct := db.SQLPool.QueryRow(context.Background(), q)
	e := ct.Scan(&wc.Word, &wc.Total, &wc.Gr, &wc.Lt, &wc.Dp, &wc.In, &wc.Ch)
	if e != nil {
		msg.FYI(fmt.Sprintf(NOTH, word))
	}

	label := wc.Word
	allformpd := formatprevalencedata(wc, label)

	// [e] format the parsing summary

	parsing := formatparsingdata(mpp)

	// [f] generate the lexical output: multiple entries possible - <div id="δημόϲιοϲ_23337644"> ... <div id="δημοϲίᾳ_23333080"> ...

	var entries string
	for _, w := range lexicalfinds {
		entries += formatlexicaloutput(w)
	}

	// [g] add the HTML + JS to inject `{"newhtml": "...", "newjs":"..."}`

	html := allformpd + parsing + entries

	// [h] conditionally rewrite the html
	if launch.Config.ZapLunates {
		html = generic.DeLunate(html)
	}

	// author flagging: "<bibl id="perseus/lt0474" --> "<bibl class="flagged" id="perseus/lt0474"
	html = strings.ReplaceAll(html, fmt.Sprintf(SRCH, author), fmt.Sprintf(REPL, author))

	return html
}

// reversefind - english word into collection of HTML dictionary entries
func reversefind(word string, dicts []string) string {
	const (
		ENTRYSPAN = `<span class="sensum">(%d)&nbsp;<a class="nounderline" href="#%s_%f">%s</a>
			<span class="small">&nbsp;(%d)</span></span><br />`
		SEPARATOR = `<hr>`
		ITEMIZER  = `<hr><span class="small">(%d)</span>`
	)

	var lexicalfinds []structs.DbLexicon
	// [a] look for the words
	for _, d := range dicts {
		ff := dictgrabber(word, d, "translations", "~")
		lexicalfinds = append(lexicalfinds, ff...)
	}

	// [b] the counts for the finds
	countmap := make(map[float32]structs.DbHeadwordCount)
	for _, f := range lexicalfinds {
		ct := search.HeadwordLookup(f.Word)
		if ct.Entry == "" {
			ct.Entry = f.Word
		}
		countmap[f.ID] = ct
	}

	// [c] get the html for the entries

	htmlmap := paralleldictformatter(lexicalfinds)

	var keys []float32
	for k := range htmlmap {
		keys = append(keys, k)
	}

	// sort by number of hits
	sort.Slice(keys, func(i, j int) bool { return countmap[keys[i]].Total > countmap[keys[j]].Total })

	// [d] prepare the output

	// [d1] insert the overview
	ov := make([]string, len(lexicalfinds))
	for i, k := range keys {
		ov[i] = fmt.Sprintf(ENTRYSPAN, i+1, countmap[k].Entry, k, countmap[k].Entry, countmap[k].Total)
	}

	htmlchunks := make([]string, len(keys))
	for i, k := range keys {
		n := fmt.Sprintf(ITEMIZER, i+1)
		h := strings.Replace(htmlmap[k], SEPARATOR, n, 1)
		htmlchunks[i] = h
	}

	htmlchunks = append(ov, htmlchunks...)

	thehtml := strings.Join(htmlchunks, "")

	if len(thehtml) == 0 {
		thehtml = "(nothing found)"
	}

	return thehtml
}

// dictsearch - word into HTML dictionary entry
func dictsearch(seeking string, dict string) string {
	// this is pretty slow if you do 100 entries... so run it in parallel

	const (
		ENTRYLINE = `<span class="sensum">(%d)&nbsp;<a class="nounderline" href="#%s_%f">%s</a><span class="small">&nbsp;(%d)</span><br>`
		HITCAP    = `<span class="small">[stopped searching after %d entries found]</span><br>`
		SEPARATOR = `<hr>`
		CHUNKHEAD = `<hr><span class="small">(%d)</span>`
		COLUMN    = "entry_name"
		SYNTAX    = "~*"
	)

	lexicalfinds := dictgrabber(seeking, dict, COLUMN, SYNTAX)

	htmlmap := paralleldictformatter(lexicalfinds)

	var keys []float32
	for k := range htmlmap {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	htmlchunks := make([]string, len(keys))
	for i, k := range keys {
		n := fmt.Sprintf(CHUNKHEAD, i+1)
		h := strings.Replace(htmlmap[k], SEPARATOR, n, 1)
		htmlchunks[i] = h
	}

	countmap := make(map[float32]structs.DbHeadwordCount)
	for _, f := range lexicalfinds {
		ct := search.HeadwordLookup(f.Word)
		if ct.Entry == "" {
			ct.Entry = f.Word
		}
		countmap[f.ID] = ct
	}

	// [d1] insert the overview

	ov := make([]string, len(lexicalfinds))
	for i, e := range lexicalfinds {
		ov[i] = fmt.Sprintf(ENTRYLINE, i+1, e.Word, e.ID, e.Word, countmap[e.ID].Total)
	}

	if len(lexicalfinds) == vv.MAXDICTLOOKUP {
		ov = append(ov, fmt.Sprintf(HITCAP, vv.MAXDICTLOOKUP))
	}

	htmlchunks = append(ov, htmlchunks...)

	html := strings.Join(htmlchunks, "")

	if len(html) == 0 {
		html = "(nothing found)"
	}

	if launch.Config.ZapLunates {
		html = generic.DeLunate(html)
	}

	return html
}

// dictgrabber - search postgres tables and return []DbLexicon
func dictgrabber(seeking string, dict string, col string, syntax string) []structs.DbLexicon {
	const (
		FLDS = `entry_name, metrical_entry, id_number, pos, translations, html_body`
		PSQQ = `SELECT %s FROM %s_dictionary WHERE %s %s '%s' ORDER BY id_number ASC LIMIT %d`
	)

	// note that "html_body" is only available via HipparchiaBuilder 1.6.0+
	q := fmt.Sprintf(PSQQ, FLDS, dict, col, syntax, seeking, vv.MAXDICTLOOKUP)

	var lexicalfinds []structs.DbLexicon
	var thehit structs.DbLexicon
	dedup := make(map[float32]bool)

	foreach := []any{&thehit.Word, &thehit.Metrical, &thehit.ID, &thehit.POS, &thehit.Transl, &thehit.Entry}
	rwfnc := func() error {
		thehit.SetLang(dict)
		if _, dup := dedup[thehit.ID]; !dup {
			// use ID and not Lex because καρπόϲ.53442 is not καρπόϲ.53443
			dedup[thehit.ID] = true
			lexicalfinds = append(lexicalfinds, thehit)
		}
		return nil
	}

	foundrows, err := db.SQLPool.Query(context.Background(), q)
	msg.EC(err)

	_, e := pgx.ForEachRow(foundrows, foreach, rwfnc)
	msg.EC(e)

	return lexicalfinds
}

// getmorphmatch - word into []DbMorphology
func getmorphmatch(word string, lang string) []structs.DbMorphology {
	const (
		FLDS = `observed_form, xrefs, prefixrefs, possible_dictionary_forms, related_headwords`
		PSQQ = "SELECT %s FROM %s_morphology WHERE observed_form = '%s'"
	)

	psq := fmt.Sprintf(PSQQ, FLDS, lang, word)

	foundrows, err := db.SQLPool.Query(context.Background(), psq)
	msg.EC(err)

	thesefinds, err := pgx.CollectRows(foundrows, pgx.RowToStructByPos[structs.DbMorphology])
	msg.EC(err)

	return thesefinds
}

// dbmorphintomorphpossib - from []DbMorphology yield up []MorphPossib
func dbmorphintomorphpossib(dbmm []structs.DbMorphology) []structs.MorphPossib {

	var mpp []structs.MorphPossib

	for _, d := range dbmm {
		mpp = append(mpp, extractmorphpossibilities(d.RawPossib)...)
	}

	return mpp
}

// extractmorphpossibilities - turn nested morphological JSON into []MorphPossib
func extractmorphpossibilities(raw string) []structs.MorphPossib {
	// Input:     {"1": {"transl": "A.I. stem, tree; II. shaft of a spear", "analysis": "neut nom/voc/acc sg", "headword": "δόρυ", "scansion": "", "xref_kind": "9", "xref_value": "26874791"}}
	// Unmarshal: map[1:{A.I. stem, tree; II. shaft of a spear neut nom/voc/acc sg δόρυ  9 26874791}]

	const (
		FAIL = "dbmorphintomorphpossib() could not unmarshal %s"
	)

	nested := make(map[string]structs.MorphPossib)
	e := json.Unmarshal([]byte(raw), &nested)
	if e != nil {
		msg.TMI(fmt.Sprintf(FAIL, raw))
	}

	// ob-caec --> obcaec, dēmorsico --> demorsico...
	// note that there is a macron in there in the second pair: ̄
	clean := strings.NewReplacer("-", "", "̄", "")

	mpp := generic.StringMapIntoSlice(nested)
	for i := 0; i < len(mpp); i++ {
		// "ob-caec" --> "obcaec", etc.
		mpp[i].Headwd = clean.Replace(mpp[i].Headwd)
	}
	return mpp
}

// morphpossibintolexpossib - []MorphPossib into []DbLexicon
func morphpossibintolexpossib(d string, mpp []structs.MorphPossib) []structs.DbLexicon {
	const (
		FLDS = `entry_name, metrical_entry, id_number, pos, translations, html_body`
		PSQQ = `SELECT %s FROM %s_dictionary WHERE %s ~* '^%s(|¹|²|³|⁴|1|2)$' ORDER BY id_number ASC`
		COLM = "entry_name"
	)
	var hwm []string
	for _, p := range mpp {
		if strings.TrimSpace(p.Headwd) != "" {
			hwm = append(hwm, p.Headwd)
		}
	}

	// the next is primed to produce problems: see καρποῦ which will turn καρπόϲ1 and καρπόϲ2 into just καρπόϲ; need xref_value?
	// but we have probably taken care of this below: see the comments
	hwm = generic.Unique(hwm)

	// [d] get the wordobjects for each Unique headword: probedictionary()

	// note that "html_body" is only available via HipparchiaBuilder 1.6.0+

	var lexicalfinds []structs.DbLexicon
	var thehit structs.DbLexicon
	dedup := make(map[float32]bool)

	foreach := []any{&thehit.Word, &thehit.Metrical, &thehit.ID, &thehit.POS, &thehit.Transl, &thehit.Entry}

	rwfnc := func() error {
		thehit.SetLang(d)
		if _, dup := dedup[thehit.ID]; !dup {
			// use ID and not Lex because καρπόϲ.53442 is not καρπόϲ.53443
			dedup[thehit.ID] = true
			lexicalfinds = append(lexicalfinds, thehit)
		}
		return nil
	}

	for _, w := range hwm {
		q := fmt.Sprintf(PSQQ, FLDS, d, COLM, w)
		foundrows, err := db.SQLPool.Query(context.Background(), q)
		msg.EC(err)

		_, e := pgx.ForEachRow(foundrows, foreach, rwfnc)
		msg.EC(e)

	}
	return lexicalfinds
}

//
// FORMATTING
//

// paralleldictformatter - send N workers off to turn []DbLexicon into a map: [entryid]entryhtml
func paralleldictformatter(lexicalfinds []structs.DbLexicon) map[float32]string {
	workers := launch.Config.WorkerCount
	totalwork := len(lexicalfinds)
	chunksize := totalwork / workers
	leftover := totalwork % workers
	entrymap := make(map[int][]structs.DbLexicon, workers)

	if totalwork <= workers {
		chunksize = 1
		workers = totalwork
		leftover = 0
	}

	thestart := 0
	for i := 0; i < workers; i++ {
		entrymap[i] = lexicalfinds[thestart : thestart+chunksize]
		thestart = thestart + chunksize
	}

	if leftover > 0 {
		entrymap[workers-1] = append(entrymap[workers-1], lexicalfinds[totalwork-leftover-1:totalwork-1]...)
	}

	var wg sync.WaitGroup
	var collector []map[float32]string

	outputchannels := make(chan map[float32]string, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		j := i
		go func(lexlist []structs.DbLexicon, workerid int) {
			defer wg.Done()
			outputchannels <- multipleentriesashtml(entrymap[j])
		}(entrymap[i], i)
	}

	go func() {
		wg.Wait()
		close(outputchannels)
	}()

	// merge the results into []map[float32]string
	for c := range outputchannels {
		collector = append(collector, c)
	}

	// reduce the results map
	htmlmap := make(map[float32]string)

	for _, hmap := range collector {
		for w := range hmap {
			htmlmap[w] = hmap[w]
		}
	}

	return htmlmap
}

// multipleentriesashtml - turn []DbLexicon into a map: [entryid]entryhtml
func multipleentriesashtml(ee []structs.DbLexicon) map[float32]string {
	oneentry := func(e structs.DbLexicon) (float32, string) {
		body := formatlexicaloutput(e)
		return e.ID, body
	}

	entries := make(map[float32]string, len(ee))
	for _, e := range ee {
		id, ent := oneentry(e)
		entries[id] = ent
	}
	return entries
}

// formatprevalencedata - turn a wordcount into an HTML summary
func formatprevalencedata(w structs.DbWordCount, s string) string {
	// <p class="wordcounts">Prevalence (all forms): <span class="prevalence">Ⓣ</span> 1482 / <span class="prevalence">Ⓖ</span> 1415 / <span class="prevalence">Ⓓ</span> 54 / <span class="prevalence">Ⓘ</span> 11 / <span class="prevalence">Ⓒ</span> 2</p>
	const (
		PDPAR = `<p class="wordcounts">Prevalence of <span class="emph">%s</span>: %s</p>`
		PDSPA = `<span class="rarechars prevalence">%s</span> %d`
	)

	m := message.NewPrinter(language.English)

	labels := map[string]string{"Total": "Ⓣ", "Gr": "Ⓖ", "Lt": "Ⓛ", "Dp": "Ⓓ", "In": "Ⓘ", "Ch": "Ⓒ"}

	var pdd []string
	for _, l := range []string{"Total", "Gr", "Lt", "Dp", "In", "Ch"} {
		v := reflect.ValueOf(w).FieldByName(l).Int()
		if v > 0 {
			pdd = append(pdd, m.Sprintf(PDSPA, labels[l], v))
		}
	}

	spans := strings.Join(pdd, " / ")
	html := fmt.Sprintf(PDPAR, s, spans)
	return html
}

// formatparsingdata - turn []MorphPossib into HTML
func formatparsingdata(mpp []structs.MorphPossib) string {
	const (
		OBSERVED = `<span class="obsv"><span class="obsv"> from <span class="baseform"><a class="lex" href="#%s_%s">%s</a></span>
	`
		BFTRANS  = `<span class="baseformtranslation">&nbsp;(“%s”)</span></span></span>`
		MORPHTAB = `
		<table class="morphtable">
			<tbody>
			%s
			</tbody>
		</table>
	`
		MORPHTR = `<tr>%s</tr>`
		MORPHTD = `<td class="%s">%s</td>`
	)
	pat := regexp.MustCompile("^(.{1,3}\\.)\\s")

	mpmap := make(map[string]structs.MorphPossib, len(mpp))
	for _, p := range mpp {
		k := p.Headwd + " - " + p.Anal + " - " + p.Transl
		mpmap[k] = p
	}

	keys := generic.StringMapKeysIntoSlice(mpmap)
	sort.Strings(keys)

	var html string
	usecounter := false
	// on mpp is always empty: why?
	if len(mpp) > 2 {
		usecounter = true
	}
	ct := 0
	memo := ""
	// there are duplicates in the original parsing data
	dedup := make(map[string]bool)
	letter := 0

	for _, k := range keys {
		m := mpmap[k]

		if strings.TrimSpace(m.Anal) == "" {
			continue
		}

		getlett := func() string {
			if len(mpp) > 2 {
				return fmt.Sprintf("[%s]", string(rune(letter+97)))
			}
			return ""
		}()

		if usecounter && m.Xrefval != memo {
			ct += 1
			html += fmt.Sprintf("(%d)&nbsp;", ct)
		}

		if m.Xrefval != memo {
			html += fmt.Sprintf(OBSERVED, m.Headwd, m.Xrefval, m.Headwd)
			if strings.TrimSpace(m.Transl) != "" {
				m.Transl = polishtrans(m.Transl, pat)
				html += fmt.Sprintf(BFTRANS, m.Transl)
			}
		}

		dd := m.Headwd + " - " + m.Anal
		if _, ok := dedup[dd]; !ok {
			pos := strings.Split(m.Anal, " ")
			var tab string
			tab = fmt.Sprintf(MORPHTD, "morphcell", getlett)
			for _, p := range pos {
				tab += fmt.Sprintf(MORPHTD, "morphcell", p)
			}
			tab = fmt.Sprintf(MORPHTR, tab)
			tab = fmt.Sprintf(MORPHTAB, tab)
			html += tab
			memo = m.Xrefval
			dedup[dd] = true
		} else {
			letter -= 1
		}
		letter += 1
	}

	return html
}

// formatlexicaloutput - turn a DbLexicon word into HTML
func formatlexicaloutput(w structs.DbLexicon) string {
	const (
		HEADTEMPL = `<div id="%s_%f"><hr>
		<p class="dictionaryheading" id="%s_%.1f">%s&nbsp;<span class="metrics">%s</span></p>
	`
		FORMSUMM = `<formsummary parserxref="%d" lexicalid="%.1f" headword="%s" lang="%s">%d known forms</formsummary>`

		FRQSUM = `<p class="wordcounts">Relative frequency: <span class="blue">%s</span></p>`

		NAVTABLE = `
		<table class="navtable">
			<tbody>
			<tr>
				<td class="alignleft">
					<span class="label">Previous: </span>
					<dictionaryidsearch entryid="%.1f" language="%s">%s</dictionaryidsearch>
				</td>
				<td>&nbsp;</td>
				<td class="alignright">
					<span class="label">Next: </span>
					<dictionaryidsearch entryid="%.1f" language="%s">%s</dictionaryidsearch>
				</td>
			</tr>
			</tbody>
		</table>`

		PROXENTRYQUERY = `SELECT entry_name, id_number from %s_dictionary WHERE id_number %s %.0f ORDER BY id_number %s LIMIT 1`
		NOTH           = `formatlexicaloutput() found no entry %s '%s'`
	)

	var elem []string

	// [h1] first part of a lexical entry:

	var met string
	if w.Metrical != "" {
		met = fmt.Sprintf("[%s]", quantityfixer.Replace(w.Metrical))
	}

	elem = append(elem, fmt.Sprintf(HEADTEMPL, w.Word, w.ID, w.Word, w.ID, w.Word, met))

	// [h1a] known forms in use

	hwc := search.HeadwordLookup(w.Word)
	elem = append(elem, fmt.Sprintf(FRQSUM, hwc.FrqCla))

	lw := generic.UVσςϲ(w.Word) // otherwise "venio" will hit AllLemm instead of "uenio"
	if _, ok := mps.AllLemm[lw]; ok {
		elem = append(elem, fmt.Sprintf(FORMSUMM, mps.AllLemm[lw].Xref, w.ID, w.Word, w.GetLang(), len(mps.AllLemm[lw].Deriv)))
	}

	// [h1b] principle parts

	// TODO: but not at all a priority for v1.x

	// [h2] wordcounts data including weighted distributions

	elem = append(elem, `<div class="wordcounts">`)
	elem = append(elem, headwordprevalence(hwc))
	elem = append(elem, headworddistrib(hwc))
	elem = append(elem, headwordchronology(hwc))
	elem = append(elem, headwordgenres(hwc))
	elem = append(elem, `</div>`)

	// [h4] the actual body of the entry

	w.Entry = entryqickfixes(w.Entry)

	elem = append(elem, w.Entry)

	// [h5] previous & next entry

	// todo: push all db interaction into 'search'
	var prev structs.DbLexicon
	p := db.SQLPool.QueryRow(context.Background(), fmt.Sprintf(PROXENTRYQUERY, w.GetLang(), "<", w.ID, "DESC"))
	e := p.Scan(&prev.Entry, &prev.ID)
	if e != nil {
		msg.FYI(fmt.Sprintf(NOTH, "before", w.Entry))
	}

	var nxt structs.DbLexicon
	n := db.SQLPool.QueryRow(context.Background(), fmt.Sprintf(PROXENTRYQUERY, w.GetLang(), ">", w.ID, "ASC"))
	e = n.Scan(&nxt.Entry, &nxt.ID)
	if e != nil {
		msg.FYI(fmt.Sprintf(NOTH, "after", w.Entry))
	}

	pn := fmt.Sprintf(NAVTABLE, prev.ID, w.GetLang(), prev.Entry, nxt.ID, w.GetLang(), nxt.Entry)
	elem = append(elem, pn)

	html := strings.Join(elem, "")

	return html
}

// entryqickfixes - tidy up wonky things that are in "html_body" in the DB; the builder should be doing this instead...
func entryqickfixes(html string) string {
	// [a]
	// <span class="dictquote dictlang_la">sedile"><span class="dictcit"><span class="dictquote dictlang_la">sedile</dictionaryentry>
	//     "sedile" is here twice and will print as 'sedile">sedile,'

	badpatt1, err := regexp.Compile("<span class=\"dictquote dictlang_la\">(\\w+)\"><span class=\"dictcit\"><span class=\"dictquote dictlang_la\">(\\w+)")
	msg.EC(err)
	html = badpatt1.ReplaceAllString(html, "<span class=\"dictcit\"><span class=\"dictquote dictlang_la\">$1")

	// [b] ē^ -> ē̆

	html = quantityfixer.Replace(html)

	return html
}

func insertlexicaljs() string {
	const (
		LJS = `
	<script>
	%s
	%s
	</script>`
	)

	jscore := fmt.Sprintf(vv.BROWSERJS, "bibl")

	thejs := fmt.Sprintf(LJS, jscore, vv.DICTIDJS)
	return thejs
}

func headwordprevalence(wc structs.DbHeadwordCount) string {
	// Prevalence (all forms): Ⓖ 95,843 / Ⓛ 10 / Ⓘ 151 / Ⓓ 751 / Ⓒ 64 / Ⓣ 96,819
	const (
		PREVSPAN = `<span class="prevalence rarechars">%s</span>&nbsp;%d`
		PREVSUM  = `<br>Prevalence (all forms): `
	)

	m := message.NewPrinter(language.English)

	cv := wc.CorpVal

	var pd []string

	for _, c := range cv {
		if c.Count > 0 {
			pd = append(pd, m.Sprintf(PREVSPAN, c.Name, c.Count))
		}
	}
	pd = append(pd, m.Sprintf(PREVSPAN, "Ⓣ", wc.Total))

	p := PREVSUM + strings.Join(pd, " / ")

	return p
}

func headworddistrib(wc structs.DbHeadwordCount) string {
	// Weighted distribution by corpus: Ⓖ 100 / Ⓓ 14 / Ⓒ 6 / Ⓘ 2 / Ⓛ 0
	const (
		DIST = `<br>Distribution by corpus: `
	)
	cv := wc.CorpVal

	for i, c := range cv {
		cv[i].Count = int(float32(c.Count) * structs.CORPUSWEIGTING[c.Name])
	}

	// descending order
	slices.SortFunc(cv, func(a, b structs.HWData) int { return cmp.Compare(b.Count, a.Count) })

	mx := cv[0].Count

	p := ""
	if mx != 0 {
		pd := weightedpdslice(cv, true)
		p = DIST + strings.Join(pd, "; ")
	}

	return p
}

func headwordchronology(wc structs.DbHeadwordCount) string {
	// Weighted chronological distribution: ℯ 100 / ℓ 84 / 𝓂 62
	const (
		DIST = `<br>Distribution by time: `
	)
	cv := wc.TimeVal

	for i, c := range cv {
		cv[i].Count = int(float32(c.Count) * structs.ERAWEIGHTING[c.Name])
	}

	// descending order
	slices.SortFunc(cv, func(a, b structs.HWData) int { return cmp.Compare(b.Count, a.Count) })

	mx := cv[0].Count

	p := ""
	if mx != 0 {
		pd := weightedpdslice(cv, true)
		p = DIST + strings.Join(pd, "; ")

	}

	return p
}

func headwordgenres(wc structs.DbHeadwordCount) string {
	// Predominant genres: comm (100), mech (97), jurisprud (93), med (84), mus (75), nathist (61), paroem (60), allrelig (57)
	const (
		DIST = `<br>Distribution by genre: `
	)

	cv := wc.GenreVal

	wt := map[string]float32{}
	if vv.IsGreek.MatchString(wc.Entry) {
		wt = structs.GKGENREWEIGHT
	} else {
		wt = structs.LATGENREWEIGHT
	}

	for i, c := range cv {
		w := wt[c.Name]
		if w > vv.MINORGENREWTCAP {
			w = 0
		}
		cv[i].Count = int(float32(c.Count) * w)
	}

	// descending order
	slices.SortFunc(cv, func(a, b structs.HWData) int { return cmp.Compare(b.Count, a.Count) })

	mx := cv[0].Count

	p := ""
	if mx != 0 {
		pd := weightedpdslice(cv, false)
		lim := math.Min(vv.GENRESTOCOUNT, float64(len(pd)))
		pd = pd[0:int(lim)]
		p = DIST + strings.Join(pd, "; ")
	}

	return p
}

// weightedpdslice - convert count values into a formatted string slice
func weightedpdslice(cv []structs.HWData, rare bool) []string {
	const (
		PREVSPANA = `<span class="rarechars prevalence">%s</span>&nbsp;%d`
		PREVSPANB = `<span class="rarechars prevalence">%s</span>&nbsp;%d`
	)

	ps := PREVSPANA
	if !rare {
		// headwordgenres()
		ps = PREVSPANB
	}

	mx := cv[0].Count
	var pd []string
	for _, c := range cv {
		cpt := (float32(c.Count) / float32(mx)) * 100
		if int(cpt) > 0 {
			pd = append(pd, fmt.Sprintf(ps, c.Name, int(cpt)))
		}
	}
	return pd
}
