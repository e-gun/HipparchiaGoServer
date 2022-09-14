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
)

// full set of verbs: lookup, findbyform, idlookup, morphologychart

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

func RtLexLookup(c echo.Context) error {
	req := c.Param("wd")
	seeking := purgechars(UNACCEPTABLEINPUT, req)
	seeking = swapacuteforgrave(seeking)

	dict := "latin"
	if isGreek.MatchString(seeking) {
		dict = "greek"
	}

	html := dictsearch(seeking, dict)

	var jb JSB
	jb.HTML = html
	jb.JS = insertlexicaljs()

	jsonbundle, ee := json.Marshal(jb)
	chke(ee)
	return c.String(http.StatusOK, string(jsonbundle))
}

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

	word := purgechars(UNACCEPTABLEINPUT, elem[0])

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

func RtLexReverse(c echo.Context) error {
	// be able to respond to "/lexica/reverselookup/0ae94619/sorrow"
	req := c.Param("wd")
	elem := strings.Split(req, "/")

	if len(elem) == 0 || elem[0] == "" {
		return c.String(http.StatusOK, "")
	}

	word := purgechars(UNACCEPTABLEINPUT, elem[1])

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

	mpp := dbmorthintomorphpossib(thesefinds)

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
		err := foundrows.Scan(&wc.Word, &wc.Total, &wc.Gr, &wc.Lt, &wc.Dp, &wc.In, &wc.Ch)
		chke(err)
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
	// this is not the fast way to do it; just the first draft of a way to do it...
	type EntryStruct struct {
		Lex   DbLexicon
		Count DbHeadwordCount
		HTML  string
	}

	var entries []EntryStruct
	dbpool := GetPSQLconnection()
	defer dbpool.Close()

	// [a] look for the words
	qt := `SELECT %s FROM %s_dictionary WHERE translations ~ '%s' LIMIT %d`
	fld := `entry_name, metrical_entry, id_number, pos, translations, html_body`
	for _, d := range dicts {
		psq := fmt.Sprintf(qt, fld, d, word, MAXDICTLOOKUP)

		var foundrows pgx.Rows
		var err error
		foundrows, err = dbpool.Query(context.Background(), psq)
		chke(err)

		for foundrows.Next() {
			var newword DbLexicon
			err = foundrows.Scan(&newword.Word, &newword.Metrical, &newword.ID, &newword.POS, &newword.Transl, &newword.Entry)
			chke(err)
			newword.Lang = d

			var newentry EntryStruct
			newentry.Lex = newword
			entries = append(entries, newentry)
		}
		foundrows.Close()
	}

	// [b] attach the counts to the finds
	for i, f := range entries {
		hwc := headwordlookup(f.Lex.Word)
		entries[i].Count = hwc
	}

	// [c] attach the html to the entries
	// this is the slow bit and could be parallelized
	for i, e := range entries {
		b := formatlexicaloutput(e.Lex)
		entries[i].HTML = b
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Count.Total > entries[j].Count.Total })

	// [d] prepare the output

	// et := `<span class="sensesum">(INDEX)&nbsp;<a class="nounderline" href="ENTRY_ENTRYID">ENTRY</a><span class="small">(COUNT)</span></span><br />`
	et := `<span class="sensum">(%d)&nbsp;<a class="nounderline" href="%s_%f">%s</a><span class="small">(%d)</span></span><br />`

	// [d1] insert the overview
	var htmlchunks []string
	for i, e := range entries {
		h := fmt.Sprintf(et, i+1, e.Lex.Word, e.Lex.ID, e.Lex.Word, e.Count.Total)
		htmlchunks = append(htmlchunks, h)
	}

	// [d2] insert the actual entries
	for _, e := range entries {
		htmlchunks = append(htmlchunks, e.HTML)
	}

	thehtml := strings.Join(htmlchunks, "")

	return thehtml
}

// dictsearch - word into HTML dictionary entry
func dictsearch(seeking string, dict string) string {
	dbpool := GetPSQLconnection()
	defer dbpool.Close()

	// note that "html_body" is only available via HipparchiaBuilder 1.6.0+
	fld := `entry_name, metrical_entry, id_number, pos, translations, html_body`
	psq := `SELECT %s FROM %s_dictionary WHERE %s ~* '%s' ORDER BY id_number ASC LIMIT %d`
	col := "entry_name"
	q := fmt.Sprintf(psq, fld, dict, col, seeking, MAXDICTLOOKUP)

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

	sort.Slice(lexicalfinds, func(i, j int) bool { return lexicalfinds[i].Word < lexicalfinds[j].Word })

	// [d1] insert the overview
	et := `<span class="sensum">(%d)&nbsp;<a class="nounderline" href="#%s_%f">%s</a><br />`

	var htmlchunks []string
	for i, l := range lexicalfinds {
		h := fmt.Sprintf(et, i+1, l.Word, l.ID, l.Word)
		htmlchunks = append(htmlchunks, h)
	}

	// the entries
	for _, l := range lexicalfinds {
		htmlchunks = append(htmlchunks, formatlexicaloutput(l))
	}

	html := strings.Join(htmlchunks, "")

	return html
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

// dbmorthintomorphpossib - from []DbMorphology yield up []MorphPossib
func dbmorthintomorphpossib(dbmm []DbMorphology) []MorphPossib {
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
				msg(fmt.Sprintf("dbmorthintomorphpossib() could not unmarshal %s", p), 5)
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
			err := foundrows.Scan(&thehit.Word, &thehit.Metrical, &thehit.ID, &thehit.POS, &thehit.Transl, &thehit.Entry)
			chke(err)
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

// formatprevalencedata - turn a wordcount into an HTML summary
func formatprevalencedata(w DbWordCount, s string) string {
	// <p class="wordcounts">Prevalence (all forms): <span class="prevalence">Ⓣ</span> 1482 / <span class="prevalence">Ⓖ</span> 1415 / <span class="prevalence">Ⓓ</span> 54 / <span class="prevalence">Ⓘ</span> 11 / <span class="prevalence">Ⓒ</span> 2</p>
	m := message.NewPrinter(language.English)

	pdp := `<p class="wordcounts">Prevalence of %s: %s</p>`
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
	<span class="obsv"><a class="parsing" href="#%s_%s"><span class="obsv"> from <span class="baseform">%s</span>
	`
	bft := `<span class="baseformtranslation">&nbsp;(“%s”)</span></span></a></span>`
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

	// TODO

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
	</script>`

	jscore := fmt.Sprintf(BROWSERJS, "bibl")

	thejs := fmt.Sprintf(js, jscore)
	return thejs
}
