//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/labstack/echo/v4"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
)

var (
	// regex compiled here instead of inside of various loops
	isGreek = regexp.MustCompile("[α-ωϲῥἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάἐἑἒἓἔἕὲέἰἱἲἳἴἵἶἷὶίῐῑῒΐῖῗὀὁὂὃὄὅόὸὐὑὒὓὔὕὖὗϋῠῡῢΰῦῧύὺᾐᾑᾒᾓᾔᾕᾖᾗῂῃῄῆῇἤἢἥἣὴήἠἡἦἧὠὡὢὣὤὥὦὧᾠᾡᾢᾣᾤᾥᾦᾧῲῳῴῶῷώὼ]")
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

type DbLexicon struct {
	// skipping 'unaccented_entry' from greek_dictionary
	// skipping 'entry_key' from latin_dictionary
	Word     string
	Metrical string
	ID       float32
	POS      string
	Transl   string
	Entry    string
	Lang     string
}

type DbMorphology struct {
	Observed    string
	Xrefs       string
	PrefixXrefs string
	RawPossib   string
	RelatedHW   string
}

func (dbm DbMorphology) PossibSlice() []string {
	return strings.Split(dbm.RawPossib, " ")
}

type DbWordCount struct {
	Word  string
	Total int64
	Gr    int64
	Lt    int64
	Dp    int64
	In    int64
	Ch    int64
}

type MorphPossib struct {
	Transl   string `json:"transl"`
	Anal     string `json:"analysis"`
	Headwd   string `json:"headword"`
	Scansion string `json:"scansion"`
	Xrefkind string `json:"xref_kind"`
	Xrefval  string `json:"xref_value"`
}

type JSB struct {
	HTML string `json:"newhtml"`
	JS   string `json:"newjs"`
}

//
// ROUTING
//

// RtLexLookup - search the dictionary for a headword substring
func RtLexLookup(c echo.Context) error {
	req := c.Param("wd")
	seeking := purgechars(cfg.BadChars, req)
	seeking = swapacuteforgrave(seeking)

	dict := "latin"
	if isGreek.MatchString(seeking) {
		dict = "greek"
	}

	seeking = uvσςϲ(seeking)
	seeking = universalpatternmaker(seeking)
	// universalpatternmaker() returns the term with brackets around it
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

	jsonbundle, ee := json.Marshal(jb)
	chke(ee)
	return c.String(http.StatusOK, string(jsonbundle))
}

// RtLexFindByForm - search the dictionary for a specific headword
func RtLexFindByForm(c echo.Context) error {
	// be able to respond to "GET /lexica/findbyform/ἀμιϲθὶ/gr0062 HTTP/1.1"
	req := c.Param("wd")
	elem := strings.Split(req, "/")

	if len(elem) == 0 || elem[0] == "" {
		return c.String(http.StatusOK, "")
	}

	var au string
	if len(elem) == 1 {
		au = ""
	} else {
		au = elem[1]
	}

	word := purgechars(cfg.BadChars, elem[0])

	word = swapacuteforgrave(word)

	word = uvσςϲ(word)

	html := findbyform(word, au)

	// html = strings.Replace(html, `"`, `\"`, -1)
	js := insertlexicaljs()

	var jb JSB
	jb.HTML = html
	jb.JS = js

	jsonbundle, ee := json.Marshal(jb)
	chke(ee)

	// jsonbundle := []byte(fmt.Sprintf(`{"newhtml":"%s","newjs":"%s"}`, html, js))

	return c.String(http.StatusOK, string(jsonbundle))
}

// RtLexId - grab a word by its entry value
func RtLexId(c echo.Context) error {
	// http://127.0.0.1:8000/lexica/idlookup/latin/24236.0
	req := c.Param("wd")
	elem := strings.Split(req, "/")
	if len(elem) != 2 {
		msg(fmt.Sprintf("RtLexId() received bad request: '%s'", req), 1)
		return c.String(http.StatusOK, "")
	}
	d := purgechars(cfg.BadChars, elem[0])
	w := purgechars(cfg.BadChars, elem[1])

	f := dictgrabber(w, d, "id_number", "=")
	if len(f) == 0 {
		msg(fmt.Sprintf("RtLexId() found nothing at id_number '%s'", w), 1)
		return c.String(http.StatusOK, "")
	}

	html := formatlexicaloutput(f[0])
	js := insertlexicaljs()

	var jb JSB
	jb.HTML = html
	jb.JS = js

	jsonbundle, ee := json.Marshal(jb)
	chke(ee)

	return c.String(http.StatusOK, string(jsonbundle))
}

// RtLexReverse - look for the headwords that have the sought word in their body
func RtLexReverse(c echo.Context) error {
	// be able to respond to "/lexica/reverselookup/0ae94619/sorrow"
	req := c.Param("wd")
	elem := strings.Split(req, "/")

	if len(elem) == 0 || elem[0] == "" {
		return c.String(http.StatusOK, "")
	}

	word := purgechars(cfg.BadChars, elem[1])

	s := sessions[readUUIDCookie(c)]

	var dd []string
	// map[string]bool{"gr": true, "lt": true, "in": false, "ch": false, "dp": false}
	if s.ActiveCorp["lt"] || s.ActiveCorp["ch"] {
		dd = append(dd, "latin")
	}

	if s.ActiveCorp["gr"] || s.ActiveCorp["in"] || s.ActiveCorp["dp"] {
		dd = append(dd, "greek")
	}

	if len(dd) == 0 {
		return c.String(http.StatusOK, "")
	}

	html := reversefind(word, dd)

	var jb JSB
	jb.HTML = html
	jb.JS = insertlexicaljs()

	jsb, ee := json.Marshal(jb)
	chke(ee)

	return c.String(http.StatusOK, string(jsb))
}

//
// LOOKUPS
//

// findbyform - observed word into HTML dictionary entry
func findbyform(word string, author string) string {

	d := "latin"
	if isGreek.MatchString(word) {
		d = "greek"
	}

	// [a] search for morphology matches
	thesefinds := getmorphmatch(strings.ToLower(word), d)
	if len(thesefinds) == 0 {
		// Νέαιρα can be found, νέαιρα can't
		thesefinds = getmorphmatch(word, d)
	}

	if len(thesefinds) == 0 {
		return "(nothing found)"
	}

	// [b] turn morph matches into []MorphPossib

	mpp := dbmorphintomorphpossib(thesefinds)

	// [c] take the []MorphPossib and find the set of headwords we are interested in; store this in a []dblexicon

	lexicalfinds := morphpossibintolexpossib(d, mpp)

	// [d] generate and format the prevalence data for this form: cf formatprevalencedata() in lexicalformatting.py

	fld := `entry_name, total_count, gr_count, lt_count, dp_count, in_count, ch_count`
	psq := `SELECT %s FROM wordcounts_%s where entry_name = '%s'`
	// golang hates indexing unicode strings: strings are bytes, and unicode chars take more than one byte
	c := []rune(word)
	q := fmt.Sprintf(psq, fld, stripaccentsSTR(string(c[0])), word)

	dbpool := GetPSQLconnection()
	defer dbpool.Close()
	foundrows, err := dbpool.Query(context.Background(), q)
	chke(err)
	var wc DbWordCount
	defer foundrows.Close()
	for foundrows.Next() {
		// only one should ever return...
		e := foundrows.Scan(&wc.Word, &wc.Total, &wc.Gr, &wc.Lt, &wc.Dp, &wc.In, &wc.Ch)
		chke(e)
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

	return html
}

// reversefind - english word into collection of HTML dictionary entries
func reversefind(word string, dicts []string) string {
	var lexicalfinds []DbLexicon
	// [a] look for the words
	for _, d := range dicts {
		ff := dictgrabber(word, d, "translations", "~")
		lexicalfinds = append(lexicalfinds, ff...)
	}

	// [b] the counts for the finds
	countmap := make(map[float32]DbHeadwordCount)
	for _, f := range lexicalfinds {
		ct := headwordlookup(f.Word)
		if ct.Entry == "" {
			ct.Entry = f.Word
		}
		countmap[f.ID] = ct
	}

	// [c] get the html for the entries

	htmlmap := paralleldictformatter(lexicalfinds)

	var keys []float32
	for k, _ := range htmlmap {
		keys = append(keys, k)
	}

	// sort by number of hits
	sort.Slice(keys, func(i, j int) bool { return countmap[keys[i]].Total > countmap[keys[j]].Total })

	// [d] prepare the output

	// et := `<span class="sensesum">(INDEX)&nbsp;<a class="nounderline" href="ENTRY_ENTRYID">ENTRY</a><span class="small">(COUNT)</span></span><br />`
	et := `<span class="sensum">(%d)&nbsp;<a class="nounderline" href="#%s_%f">%s</a><span class="small">&nbsp;(%d)</span></span><br />`

	// [d1] insert the overview
	ov := make([]string, len(lexicalfinds))
	for i, k := range keys {
		ov[i] = fmt.Sprintf(et, i+1, countmap[k].Entry, k, countmap[k].Entry, countmap[k].Total)
	}

	htmlchunks := make([]string, len(keys))
	for i, k := range keys {
		n := fmt.Sprintf(`<hr><span class="small">(%d)</span>`, i+1)
		h := strings.Replace(htmlmap[k], "<hr>", n, 1)
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

	lexicalfinds := dictgrabber(seeking, dict, "entry_name", "~*")

	htmlmap := paralleldictformatter(lexicalfinds)

	var keys []float32
	for k, _ := range htmlmap {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	htmlchunks := make([]string, len(keys))
	for i, k := range keys {
		n := fmt.Sprintf(`<hr><span class="small">(%d)</span>`, i+1)
		h := strings.Replace(htmlmap[k], "<hr>", n, 1)
		htmlchunks[i] = h
	}

	countmap := make(map[float32]DbHeadwordCount)
	for _, f := range lexicalfinds {
		ct := headwordlookup(f.Word)
		if ct.Entry == "" {
			ct.Entry = f.Word
		}
		countmap[f.ID] = ct
	}

	// [d1] insert the overview
	et := `<span class="sensum">(%d)&nbsp;<a class="nounderline" href="#%s_%f">%s</a><span class="small">&nbsp;(%d)</span><br />`
	ov := make([]string, len(lexicalfinds))
	for i, e := range lexicalfinds {
		ov[i] = fmt.Sprintf(et, i+1, e.Word, e.ID, e.Word, countmap[e.ID].Total)
	}

	htmlchunks = append(ov, htmlchunks...)

	html := strings.Join(htmlchunks, "")

	if len(html) == 0 {
		html = "(nothing found)"
	}

	return html
}

// dictgrabber - search postgres tables and return []DbLexicon
func dictgrabber(seeking string, dict string, col string, syntax string) []DbLexicon {
	dbpool := GetPSQLconnection()
	defer dbpool.Close()

	// note that "html_body" is only available via HipparchiaBuilder 1.6.0+
	fld := `entry_name, metrical_entry, id_number, pos, translations, html_body`
	psq := `SELECT %s FROM %s_dictionary WHERE %s %s '%s' ORDER BY id_number ASC LIMIT %d`
	q := fmt.Sprintf(psq, fld, dict, col, syntax, seeking, MAXDICTLOOKUP)

	var lexicalfinds []DbLexicon
	var foundrows pgx.Rows
	var err error
	foundrows, err = dbpool.Query(context.Background(), q)
	chke(err)

	defer foundrows.Close()
	for foundrows.Next() {
		var thehit DbLexicon
		err := foundrows.Scan(&thehit.Word, &thehit.Metrical, &thehit.ID, &thehit.POS, &thehit.Transl, &thehit.Entry)
		chke(err)
		thehit.Lang = dict
		lexicalfinds = append(lexicalfinds, thehit)
	}
	return lexicalfinds
}

// getmorphmatch - word into []DbMorphology
func getmorphmatch(word string, lang string) []DbMorphology {
	dbpool := GetPSQLconnection()
	defer dbpool.Close()

	fld := `observed_form, xrefs, prefixrefs, possible_dictionary_forms, related_headwords`
	psq := fmt.Sprintf("SELECT %s FROM %s_morphology WHERE observed_form = '%s'", fld, lang, word)

	var foundrows pgx.Rows
	var err error

	foundrows, err = dbpool.Query(context.Background(), psq)
	chke(err)

	var thesefinds []DbMorphology
	defer foundrows.Close()
	for foundrows.Next() {
		var thehit DbMorphology
		e := foundrows.Scan(&thehit.Observed, &thehit.Xrefs, &thehit.PrefixXrefs, &thehit.RawPossib, &thehit.RelatedHW)
		chke(e)
		thesefinds = append(thesefinds, thehit)
	}
	return thesefinds
}

// dbmorphintomorphpossib - from []DbMorphology yield up []MorphPossib
func dbmorphintomorphpossib(dbmm []DbMorphology) []MorphPossib {
	var mpp []MorphPossib
	boundary := regexp.MustCompile(`(\{|, )"\d": `)

	for _, d := range dbmm {
		mpp = append(mpp, extractmorphpossibilities(d.RawPossib, boundary)...)
	}

	return mpp
}

func extractmorphpossibilities(raw string, boundary *regexp.Regexp) []MorphPossib {
	// RawPossib is JSON + JSON; nested JSON is a PITA, but the structure is: {"1": {...}, "2": {...}, ...}
	// that is splittable
	// just need to clean the '}}' at the end
	possible := boundary.Split(raw, -1)

	var mpp []MorphPossib
	for _, p := range possible {
		p = strings.Replace(p, "}}", "}", -1)
		p = strings.TrimSpace(p)
		var mp MorphPossib
		if len(p) > 0 {
			err := json.Unmarshal([]byte(p), &mp)
			if err != nil {
				msg(fmt.Sprintf("dbmorphintomorphpossib() could not unmarshal %s", p), 5)
			}
		}
		mpp = append(mpp, mp)
	}

	return mpp
}

// morphpossibintolexpossib - []MorphPossib into []DbLexicon
func morphpossibintolexpossib(d string, mpp []MorphPossib) []DbLexicon {
	var hwm []string
	for _, p := range mpp {
		if strings.TrimSpace(p.Headwd) != "" {
			hwm = append(hwm, p.Headwd)
		}
	}

	// the next is primed to produce problems: see καρποῦ which will turn καρπόϲ1 and καρπόϲ2 into just καρπόϲ; need xref_value?
	// but we have probably taken care of this below: see the comments
	hwm = unique(hwm)

	// [d] get the wordobjects for each unique headword: probedictionary()
	dbpool := GetPSQLconnection()
	defer dbpool.Close()
	// note that "html_body" is only available via HipparchiaBuilder 1.6.0+
	fld := `entry_name, metrical_entry, id_number, pos, translations, html_body`
	psq := `SELECT %s FROM %s_dictionary WHERE %s ~* '^%s(|¹|²|³|⁴)$' ORDER BY id_number ASC`
	col := "entry_name"

	var lexicalfinds []DbLexicon
	dedup := make(map[float32]bool)
	for _, w := range hwm {
		// var foundrows pgx.Rows
		var err error
		q := fmt.Sprintf(psq, fld, d, col, w)
		foundrows, err := dbpool.Query(context.Background(), q)
		chke(err)

		defer foundrows.Close()
		for foundrows.Next() {
			var thehit DbLexicon
			e := foundrows.Scan(&thehit.Word, &thehit.Metrical, &thehit.ID, &thehit.POS, &thehit.Transl, &thehit.Entry)
			chke(e)
			thehit.Lang = d
			if _, dup := dedup[thehit.ID]; !dup {
				// use ID and not Lex because καρπόϲ.53442 is not καρπόϲ.53443
				dedup[thehit.ID] = true
				lexicalfinds = append(lexicalfinds, thehit)
			}
		}
	}
	return lexicalfinds
}

//
// FORMATTING
//

// paralleldictformatter - send N workers off to turn []DbLexicon into a map: [entryid]entryhtml
func paralleldictformatter(lexicalfinds []DbLexicon) map[float32]string {
	workers := cfg.WorkerCount
	totalwork := len(lexicalfinds)
	chunksize := totalwork / workers
	leftover := totalwork % workers
	entrymap := make(map[int][]DbLexicon, workers)

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
		go func(lexlist []DbLexicon, workerid int) {
			defer wg.Done()
			dbp := GetPSQLconnection()
			defer dbp.Close()
			outputchannels <- multipleentriesashtml(j, entrymap[j])
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
func multipleentriesashtml(workerid int, ee []DbLexicon) map[float32]string {
	msg(fmt.Sprintf("multipleentriesashtml() - worker %d sent %d entries", workerid, len(ee)), 5)

	oneentry := func(e DbLexicon) (float32, string) {
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
func formatprevalencedata(w DbWordCount, s string) string {
	// <p class="wordcounts">Prevalence (all forms): <span class="prevalence">Ⓣ</span> 1482 / <span class="prevalence">Ⓖ</span> 1415 / <span class="prevalence">Ⓓ</span> 54 / <span class="prevalence">Ⓘ</span> 11 / <span class="prevalence">Ⓒ</span> 2</p>
	m := message.NewPrinter(language.English)

	pdp := `<p class="wordcounts">Prevalence of <span class="emph">%s</span>: %s</p>`
	pds := `<span class="prevalence">%s</span> %d`
	labels := map[string]string{"Total": "Ⓣ", "Gr": "Ⓖ", "Lt": "Ⓛ", "Dp": "Ⓓ", "In": "Ⓘ", "Ch": "Ⓒ"}

	var pdd []string
	for _, l := range []string{"Total", "Gr", "Lt", "Dp", "In", "Ch"} {
		v := reflect.ValueOf(w).FieldByName(l).Int()
		if v > 0 {
			pd := m.Sprintf(pds, labels[l], v)
			pdd = append(pdd, pd)
		}
	}

	spans := strings.Join(pdd, " / ")
	html := fmt.Sprintf(pdp, s, spans)
	return html
}

// formatparsingdata - turn []MorphPossib into HTML
func formatparsingdata(mpp []MorphPossib) string {
	obs := `
	<span class="obsv"><span class="obsv"> from <span class="baseform"><a class="parsing" href="#%s_%s">%s</a></span>
	`
	bft := `<span class="baseformtranslation">&nbsp;(“%s”)</span></span></span>`
	mtb := `
	<table class="morphtable">
		<tbody>
		%s
		</tbody>
	</table>
	`
	mtr := `<tr>%s</tr>`
	mtd := `<td class="%s">%s</td>`

	var html string
	usecounter := false
	// on mpp is always empty: why?
	if len(mpp) > 2 {
		usecounter = true
	}
	ct := 0
	memo := ""
	for _, m := range mpp {
		if strings.TrimSpace(m.Headwd) == "" {
			continue
		}
		if usecounter && m.Xrefval != memo {
			ct += 1
			html += fmt.Sprintf("(%d)&nbsp;", ct)
		}

		if m.Xrefval != memo {
			html += fmt.Sprintf(obs, m.Headwd, m.Xrefval, m.Headwd)
			if strings.TrimSpace(m.Transl) != "" {
				html += fmt.Sprintf(bft, m.Transl)
			}
		}

		pos := strings.Split(m.Anal, " ")
		var tab string
		for _, p := range pos {
			tab += fmt.Sprintf(mtd, "morphcell", p)
		}
		tab = fmt.Sprintf(mtr, tab)
		tab = fmt.Sprintf(mtb, tab)
		html += tab
		memo = m.Xrefval
	}

	return html
}

// formatlexicaloutput - turn a DbLexicon word into HTML
func formatlexicaloutput(w DbLexicon) string {

	var elem []string

	// [h1] first part of a lexical entry:

	ht := `<div id="%s_%f"><hr>
		<p class="dictionaryheading" id="%s_%.1f">%s&nbsp;<span class="metrics">%s</span></p>
	`
	var met string
	if w.Metrical != "" {
		met = fmt.Sprintf("[%s]", w.Metrical)
	}

	elem = append(elem, fmt.Sprintf(ht, w.Word, w.ID, w.Word, w.ID, w.Word, met))

	// [h1a] known forms in use

	if _, ok := AllLemm[w.Word]; ok {
		kf := `<p class="wordcounts"><zformsummary parserxref="%d" lexicalid="%.1f" headword="%s" lang="%s">%d known forms</zformsummary></p>`
		// kf := `<formsummary parserxref="%d" lexicalid="%.1f" headword="%s" lang="%s">%d known forms</formsummary>`
		kf = fmt.Sprintf(kf, AllLemm[w.Word].Xref, w.ID, w.Word, w.Lang, len(AllLemm[w.Word].Deriv))
		fmt.Println(AllLemm[w.Word].Deriv)
		elem = append(elem, kf)
	}

	// [h1b] principle parts

	// TODO

	// [h2] wordcounts data including weighted distributions

	hwc := headwordlookup(w.Word)

	elem = append(elem, `<div class="wordcounts">`)
	elem = append(elem, headwordprevalence(hwc))
	elem = append(elem, headworddistrib(hwc))
	elem = append(elem, headwordchronology(hwc))
	elem = append(elem, headwordgenres(hwc))
	elem = append(elem, `</div>`)

	// [h4] the actual body of the entry

	elem = append(elem, w.Entry)

	// [h5] previous & next entry
	nt := `
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

	qt := `SELECT entry_name, id_number from %s_dictionary WHERE id_number %s %.0f ORDER BY id_number %s LIMIT 1`
	dbpool := GetPSQLconnection()
	defer dbpool.Close()

	foundrows, err := dbpool.Query(context.Background(), fmt.Sprintf(qt, w.Lang, "<", w.ID, "DESC"))
	chke(err)
	var prev DbLexicon
	defer foundrows.Close()
	for foundrows.Next() {
		err = foundrows.Scan(&prev.Entry, &prev.ID)
		chke(err)
	}

	foundrows, err = dbpool.Query(context.Background(), fmt.Sprintf(qt, w.Lang, ">", w.ID, "ASC"))
	chke(err)
	var nxt DbLexicon
	defer foundrows.Close()
	for foundrows.Next() {
		err = foundrows.Scan(&nxt.Entry, &nxt.ID)
		chke(err)
	}

	pn := fmt.Sprintf(nt, prev.ID, w.Lang, prev.Entry, nxt.ID, w.Lang, nxt.Entry)
	elem = append(elem, pn)

	html := strings.Join(elem, "")

	return html
}

func insertlexicaljs() string {
	js := `
	<script>
	%s
	%s
	</script>`

	jscore := fmt.Sprintf(BROWSERJS, "bibl")

	thejs := fmt.Sprintf(js, jscore, DICTIDJS)
	return thejs
}
