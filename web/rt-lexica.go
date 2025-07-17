//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"encoding/json"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/format"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/labstack/echo/v4"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type jsb struct {
	HTML string `json:"newhtml"`
	JS   string `json:"newjs"`
	ST   string `json:"entryname"`
}

// RtLexLookup - search the dictionary for a headword substring
func RtLexLookup(c echo.Context) error {
	c.Response().After(func() { vlt.LogPaths("RtLexLookup()") })

	user := vlt.ReadUUIDCookie(c)
	if !vlt.AllAuthorized.Check(user) {
		return jsonresponse(c, jsb{JS: vv.JSVALIDATION})
	}
	s := vlt.AllSessions.GetSess(user)

	req := c.Param("wd")
	seeking := gen.Purgechars(lnch.Config.BadChars, req)
	seeking = gen.SwapAcuteForGrave(seeking)

	originalseeking := seeking

	dict := "latin"
	if vv.IsGreek.MatchString(seeking) {
		dict = "greek"
	}

	seeking = gen.UVσςϲ(seeking)
	seeking = gen.UniversalPatternMaker(seeking) // UniversalPatternMaker() returns the term with brackets around it

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

	html := dictsearch(seeking, dict, s.ZapLunates)

	if s.ZapLunates {
		seeking = gen.DeLunate(seeking)
	}

	var jb jsb
	jb.HTML = html
	jb.JS = insertlexicaljs()
	jb.ST = originalseeking

	return jsonresponse(c, jb)
}

// RtLexFindByForm - search the dictionary for a specific headword
func RtLexFindByForm(c echo.Context) error {
	// be able to respond to "GET /lexica/findbyform/ἀμιϲθὶ/gr0062 HTTP/1.1"
	user := vlt.ReadUUIDCookie(c)
	if !vlt.AllAuthorized.Check(user) {
		return jsonresponse(c, jsb{JS: vv.JSVALIDATION})
	}
	s := vlt.AllSessions.GetSess(user)
	c.Response().After(func() { vlt.LogPaths("RtLexFindByForm()") })

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

	word := gen.Purgechars(lnch.Config.BadChars, elem[0])
	word = gen.NumberStripper.Replace(word)

	word = gen.SwapAcuteForGrave(word)
	word = gen.UVσςϲ(word)

	html := findbyform(word, au, s.ZapLunates)
	js := insertlexicaljs()

	if s.ZapLunates {
		word = gen.DeLunate(word)
	}

	var jb jsb
	jb.HTML = html
	jb.JS = js
	jb.ST = word

	return jsonresponse(c, jb)
}

// RtLexId - grab a word by its entry value
func RtLexId(c echo.Context) error {
	// http://127.0.0.1:8000/lexica/idlookup/latin/24236.0
	const (
		FAIL1 = "RtLexId() received bad request: '%s'"
		FAIL2 = "RtLexId() found nothing at idval '%s'"
	)

	user := vlt.ReadUUIDCookie(c)
	if !vlt.AllAuthorized.Check(user) {
		return jsonresponse(c, jsb{JS: vv.JSVALIDATION})
	}

	s := vlt.AllSessions.GetSess(user)

	req := c.Param("wd")
	elem := strings.Split(req, "/")
	if len(elem) != 2 {
		Msg.WARN(fmt.Sprintf(FAIL1, req))
		return emptyjsreturn(c)
	}
	d := gen.Purgechars(lnch.Config.BadChars, elem[0])
	w := gen.Purgechars(lnch.Config.BadChars, elem[1])

	f := db.DictEntryGrabber(w, d, "idval", "=")
	if len(f) == 0 {
		Msg.WARN(fmt.Sprintf(FAIL2, w))
		return emptyjsreturn(c)
	}

	html := format.FormatLexicalOutput(f[0])
	js := insertlexicaljs()

	if s.ZapLunates {
		html = gen.DeLunate(html)
	}

	var jb jsb
	jb.HTML = html
	jb.JS = js

	return jsonresponse(c, jb)
}

// RtLexReverse - look for the headwords that have the sought word in their body
func RtLexReverse(c echo.Context) error {
	// be able to respond to "/lexica/reverselookup/0ae94619/sorrow"
	c.Response().After(func() { vlt.LogPaths("RtLexReverse()") })

	user := vlt.ReadUUIDCookie(c)
	if !vlt.AllAuthorized.Check(user) {
		return jsonresponse(c, jsb{JS: vv.JSVALIDATION})
	}

	req := c.Param("wd")
	elem := strings.Split(req, "/")

	if len(elem) == 0 || elem[0] == "" {
		return emptyjsreturn(c)
	}

	word := gen.Purgechars(lnch.Config.BadChars, elem[1])

	s := vlt.AllSessions.GetSess(user)

	var dd []string
	// map[string]bool{"gr": true, "lt": true, "in": false, "ch": false, "dp": false}
	if s.ActiveCorp[vv.LATABBREV] || s.ActiveCorp[vv.CHRABBREV] {
		dd = append(dd, "latin")
	}

	if s.ActiveCorp[vv.TLGABBREV] || s.ActiveCorp[vv.INSABBREV] || s.ActiveCorp[vv.DDPABREV] {
		dd = append(dd, "greek")
	}

	if len(dd) == 0 {
		return emptyjsreturn(c)
	}

	html := reversefind(word, dd)

	if s.ZapLunates {
		html = gen.DeLunate(html)
		word = gen.DeLunate(word)
	}

	var jb jsb
	jb.HTML = html
	jb.JS = insertlexicaljs()
	jb.ST = word

	return jsonresponse(c, jb)
}

//
// LOOKUPS
//

// findbyform - observed word into HTML dictionary entry
func findbyform(word string, author string, zaplunates bool) string {
	const (
		SRCH = `<bibl id="perseus/%s/`
		REPL = `<bibl class="flagged" id="perseus/%s/`
		NOTH = "(no match for '%s' in the morphology lookup tables)"
	)

	d := "latin"
	if vv.IsGreek.MatchString(word) {
		d = "greek"
	}

	// [a] search for morphology matches
	thesefinds := db.GetMorphMatch(word, d)
	if len(thesefinds) == 0 {
		// was it a capitalization issue... a double accent...
		// there is an argument for doing the ToLower version first, but we will stick with this for now
		cleaned := strings.ToLower(word)
		cleaned = gen.StripExtraAccent(cleaned)
		thesefinds = db.GetMorphMatch(cleaned, d)
	}

	if len(thesefinds) == 0 {
		return fmt.Sprintf(NOTH, word)
	}

	// [b] turn morph matches into []MorphPossib

	mpp := dbmorphintomorphpossib(thesefinds)

	// [c] take the []MorphPossib and find the set of headwords we are interested in; store this in a []dblexicon

	lexicalfinds := db.MorphPossibIntoLexPossib(d, mpp)

	// [d] generate and format the prevalence data for this form: cf formatprevalencedata() in lexicalformatting.py

	wc := db.GetIndividualUnparsedWordCount(word)
	label := wc.Word
	allformpd := format.FormatLexPrevalenceData(wc, label)

	// [e] format the parsing summary

	parsing := format.FormatLexParsingData(mpp)

	// [f] generate the lexical output: multiple entries possible - <div id="δημόϲιοϲ_23337644"> ... <div id="δημοϲίᾳ_23333080"> ...

	var entries string
	for _, lf := range lexicalfinds {
		entries += format.FormatLexicalOutput(lf)
	}

	// [g] add the HTML + JS to inject `{"newhtml": "...", "newjs":"..."}`

	html := allformpd + parsing + entries

	// [h] conditionally rewrite the html
	if zaplunates {
		html = gen.DeLunate(html)
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

	var lexicalfinds []str.DbLexicon
	// [a] look for the words
	for _, d := range dicts {
		ff := db.DictEntryGrabber(word, d, "translations", "~")
		lexicalfinds = append(lexicalfinds, ff...)
	}

	// [b] the counts for the finds
	countmap := make(map[string]str.DbHeadwordCounts)
	for _, f := range lexicalfinds {
		ct := db.GetIndividualHeadwordCount(f.EntryName)
		if ct.Word == "" {
			ct.Word = f.EntryName
		}
		countmap[f.IdString] = ct
	}

	// [c] get the html for the entries

	htmlmap := paralleldictformatter(lexicalfinds)

	var keys []string
	for k := range htmlmap {
		keys = append(keys, k)
	}

	// sort by number of hits
	sort.Slice(keys, func(i, j int) bool { return countmap[keys[i]].Total > countmap[keys[j]].Total })

	// [d] prepare the output

	// [d1] insert the overview
	ov := make([]string, len(lexicalfinds))
	for i, k := range keys {
		ov[i] = fmt.Sprintf(ENTRYSPAN, i+1, countmap[k].Word, k, countmap[k].Word, countmap[k].Total)
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
func dictsearch(seeking string, dict string, zaplunates bool) string {
	// this is pretty slow if you do 100 entries... so run it in parallel

	const (
		ENTRYLINE = `<span class="sensum">(%d)&nbsp;<a class="nounderline" href="#%s_%f">%s</a><span class="small">&nbsp;(%d)</span><br>`
		HITCAP    = `<span class="small">[stopped searching after %d entries found]</span><br>`
		SEPARATOR = `<hr>`
		CHUNKHEAD = `<hr><span class="small">(%d)</span>`
		COLUMN    = "entry_name"
		SYNTAX    = "~*"
	)

	lexicalfinds := db.DictEntryGrabber(seeking, dict, COLUMN, SYNTAX)

	htmlmap := paralleldictformatter(lexicalfinds)

	var keys []string
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

	countmap := make(map[string]str.DbHeadwordCounts)
	for _, f := range lexicalfinds {
		ct := db.GetIndividualHeadwordCount(f.EntryName)
		if ct.Word == "" {
			ct.Word = f.EntryName
		}
		countmap[f.IdString] = ct
	}

	// [d1] insert the overview

	ov := make([]string, len(lexicalfinds))
	for i, e := range lexicalfinds {
		ov[i] = fmt.Sprintf(ENTRYLINE, i+1, e.EntryName, e.IdString, e.EntryName, countmap[e.IdString].Total)
	}

	if len(lexicalfinds) == vv.MAXDICTLOOKUP {
		ov = append(ov, fmt.Sprintf(HITCAP, vv.MAXDICTLOOKUP))
	}

	htmlchunks = append(ov, htmlchunks...)

	html := strings.Join(htmlchunks, "")

	if len(html) == 0 {
		html = "(nothing found)"
	}

	if zaplunates {
		html = gen.DeLunate(html)
	}

	return html
}

// dbmorphintomorphpossib - from []DbMorphology yield up []MorphPossib
func dbmorphintomorphpossib(dbmm []str.DbMorphology) []str.MorphPossib {

	var mpp []str.MorphPossib

	for _, d := range dbmm {
		mpp = append(mpp, extractmorphpossibilities(d.RawPossib)...)
	}

	return mpp
}

// extractmorphpossibilities - turn nested morphological JSON into []MorphPossib
func extractmorphpossibilities(raw string) []str.MorphPossib {
	// Input:     {"1": {"transl": "A.I. stem, tree; II. shaft of a spear", "analysis": "neut nom/voc/acc sg", "headword": "δόρυ", "scansion": "", "xref_kind": "9", "xref_value": "26874791"}}
	// Unmarshal: map[1:{A.I. stem, tree; II. shaft of a spear neut nom/voc/acc sg δόρυ  9 26874791}]

	// note that HGB will only set a value for Transl for greek; "latin-analyses.txt" provides no info
	// so any (latin) translations associated with a possibility ultimately have to be derived from the dictionary data
	// db.BulkEntryTranslations() does this job; in the case of greek, the morph transl values are very thin: one
	// serviceable meaning, but no range of senses

	const (
		FAIL = "extractmorphpossibilities() could not unmarshal %s"
	)

	nested := make(map[string]str.MorphPossib)
	e := json.Unmarshal([]byte(raw), &nested)
	if e != nil {
		Msg.TMI(fmt.Sprintf(FAIL, raw))
	}

	// ob-caec --> obcaec, dēmorsico --> demorsico...
	// note that there is a macron in there in the second pair: ̄
	clean := strings.NewReplacer("-", "", "̄", "")

	mpp := gen.StringMapIntoSlice(nested)
	for i := 0; i < len(mpp); i++ {
		// "ob-caec" --> "obcaec", etc.
		mpp[i].Headwd = clean.Replace(mpp[i].Headwd)
	}
	return mpp
}

//
// FORMATTING
//

// paralleldictformatter - send N workers off to turn []DbLexicon into a map: [entryid]entryhtml
func paralleldictformatter(lexicalfinds []str.DbLexicon) map[string]string {
	workers := lnch.Config.WorkerCount
	totalwork := len(lexicalfinds)
	chunksize := totalwork / workers
	leftover := totalwork % workers
	entrymap := make(map[int][]str.DbLexicon, workers)

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
	var collector []map[string]string

	outputchannels := make(chan map[string]string, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		j := i
		go func(lexlist []str.DbLexicon, workerid int) {
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
	htmlmap := make(map[string]string)

	for _, hmap := range collector {
		for w := range hmap {
			htmlmap[w] = hmap[w]
		}
	}

	return htmlmap
}

// multipleentriesashtml - turn []DbLexicon into a map: [entryid]entryhtml
func multipleentriesashtml(ee []str.DbLexicon) map[string]string {
	oneentry := func(e str.DbLexicon) (string, string) {
		body := format.FormatLexicalOutput(e)
		return e.IdString, body
	}

	entries := make(map[string]string, len(ee))
	for _, e := range ee {
		id, ent := oneentry(e)
		entries[id] = ent
	}
	return entries
}

func insertlexicaljs() string {
	const (
		LJS = `
	<script>
	%s
	%s
	</script>`
	)

	jscore := fmt.Sprintf(vv.CLICKTOBROWSE, "bibl")

	thejs := fmt.Sprintf(LJS, jscore, vv.DICTIDJS)
	return thejs
}
