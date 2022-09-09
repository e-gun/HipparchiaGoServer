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
	"net/http"
	"reflect"
	"regexp"
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

//
// ROUTING
//

func RtLexFindByForm(c echo.Context) error {
	// be able to respond to "GET /lexica/findbyform/ἀμιϲθὶ/gr0062 HTTP/1.1"
	req := c.Param("id")
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

	word := elem[0]

	word = acuteforgrave(word)

	b := findbyform(word, au)

	return c.String(http.StatusOK, string(b))
}

//
// LOOKUPS
//

func findbyform(word string, author string) []byte {

	// [a] clean the search term

	// TODO...

	// [b] pick a dictionary

	// naive and could/should be improved

	var d string
	if author[0:2] != "lt" {
		d = "greek"
	} else {
		d = "latin"
	}

	// [c] search for morphology matches

	// the python is funky because we need to poke at words several times and to try combinations of fixes
	// skipping that stuff here for now because 'findbyform' should usually see a known form
	// TODO: accute/grave issues should be handled ASAP

	dbpool := GetPSQLconnection()
	defer dbpool.Close()

	fld := `observed_form, xrefs, prefixrefs, possible_dictionary_forms, related_headwords`
	psq := fmt.Sprintf("SELECT %s FROM %s_morphology WHERE observed_form = '%s'", fld, d, word)

	var foundrows pgx.Rows
	var err error

	foundrows, err = dbpool.Query(context.Background(), psq)
	chke(err)

	var thesefinds []DbMorphology
	defer foundrows.Close()
	for foundrows.Next() {
		var thehit DbMorphology
		err := foundrows.Scan(&thehit.Observed, &thehit.Xrefs, &thehit.PrefixXrefs, &thehit.RawPossib, &thehit.RelatedHW)
		chke(err)
		thesefinds = append(thesefinds, thehit)
	}

	// [c1] turn morph matches into []MorphPossib

	var mpp []MorphPossib

	for _, h := range thesefinds {
		// RawPossib is JSON + JSON; nested JSON is a PITA, but the structure is: {"1": {...}, "2": {...}, ...}
		// that is splittable
		// just need to clean the '}}' at the end

		boundary := regexp.MustCompile(`(\{|, )"\d": `)
		possible := boundary.Split(h.RawPossib, -1)

		for _, p := range possible {
			// fmt.Println(p)
			p = strings.Replace(p, "}}", "}", -1)
			p = strings.TrimSpace(p)
			var mp MorphPossib
			if len(p) > 0 {
				err := json.Unmarshal([]byte(p), &mp)
				chke(err)
			}
			mpp = append(mpp, mp)
		}
	}

	// [d] take the []MorphPossib and find the set of headwords we are interested in

	var hwm []string
	for _, p := range mpp {
		if strings.TrimSpace(p.Headwd) != "" {
			hwm = append(hwm, p.Headwd)
		}
	}

	// the next is primed to produce problems: see καρποῦ which will turn καρπόϲ1 and καρπόϲ2 into just καρπόϲ; need xref_value?
	// but we have probably taken care of this below: see the comments
	hwm = unique(hwm)
	// fmt.Println(hwm)

	// [e] get the wordobjects for each unique headword: probedictionary()

	// note that "html_body" is only available via HipparchiaBuilder 1.6.0+
	fld = `entry_name, metrical_entry, id_number, pos, translations, html_body`
	psq = `SELECT %s FROM %s_dictionary WHERE %s ~* '^%s(|¹|²|³|⁴)$' ORDER BY id_number ASC`
	col := "entry_name"

	var lexicalfinds []DbLexicon
	dedup := make(map[float32]bool)
	for _, w := range hwm {
		// var foundrows pgx.Rows
		var err error
		q := fmt.Sprintf(psq, fld, d, col, w)
		foundrows, err = dbpool.Query(context.Background(), q)
		chke(err)

		defer foundrows.Close()
		for foundrows.Next() {
			var thehit DbLexicon
			err := foundrows.Scan(&thehit.Word, &thehit.Metrical, &thehit.ID, &thehit.POS, &thehit.Transl, &thehit.Entry)
			chke(err)
			thehit.Lang = d
			if _, dup := dedup[thehit.ID]; !dup {
				// use ID and not Word because καρπόϲ.53442 is not καρπόϲ.53443
				dedup[thehit.ID] = true
				lexicalfinds = append(lexicalfinds, thehit)
			}
		}
	}

	// [f] generate and format the prevalence data for this form: cf formatprevalencedata() in lexicalformatting.py

	fld = `entry_name, total_count, gr_count, lt_count, dp_count, in_count, ch_count`
	psq = `SELECT %s FROM wordcounts_%s where entry_name = '%s'`
	// golang hates indexing unicode strings: strings are bytes, and unicode chars take more than one byte
	c := []rune(word)
	q := fmt.Sprintf(psq, fld, stripaccentsSTR(string(c[0])), word)

	foundrows, err = dbpool.Query(context.Background(), q)
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

	// [g] format the parsing summary

	parsing := formatparsingdata(mpp)

	// [h] generate the lexical output: multiple entries possible - <div id="δημόϲιοϲ_23337644"> ... <div id="δημοϲίᾳ_23333080"> ...

	var entries string
	for _, w := range lexicalfinds {
		entries += formatlexicaloutput(w)
	}

	// [i] add the HTML + JS to inject `{"newhtml": "...", "newjs":"..."}`

	html := allformpd + parsing + entries
	// html = strings.Replace(html, `"`, `\"`, -1)
	js := insertlexicaljs()

	type JSB struct {
		HTML string `json:"newhtml"`
		JS   string `json:"newjs"`
	}

	var jb JSB
	jb.HTML = html
	jb.JS = js

	jsonbundle, ee := json.Marshal(jb)
	chke(ee)

	// jsonbundle := []byte(fmt.Sprintf(`{"newhtml":"%s","newjs":"%s"}`, html, js))

	return jsonbundle
}

//
// FORMATTING
//

// formatprevalencedata - turn a wordcount into an HTML summary
func formatprevalencedata(w DbWordCount, s string) string {
	// <p class="wordcounts">Prevalence (all forms): <span class="prevalence">Ⓣ</span> 1482 / <span class="prevalence">Ⓖ</span> 1415 / <span class="prevalence">Ⓓ</span> 54 / <span class="prevalence">Ⓘ</span> 11 / <span class="prevalence">Ⓒ</span> 2</p>

	pdp := `<p class="wordcounts">Prevalence of %s: %s</p>`
	pds := `<span class="prevalence">%s</span> %d`
	labels := map[string]string{"Total": "Ⓣ", "Gr": "Ⓖ", "Lt": "Ⓛ", "Dp": "Ⓓ", "In": "Ⓘ", "Ch": "Ⓒ"}

	var pdd []string
	for _, l := range []string{"Total", "Gr", "Lt", "Dp", "In", "Ch"} {
		v := reflect.ValueOf(w).FieldByName(l).Int()
		if v > 0 {
			pd := fmt.Sprintf(pds, labels[l], v)
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
	if len(mpp) > 1 {
		usecounter = true
	}
	ct := 0
	for _, m := range mpp {
		if strings.TrimSpace(m.Headwd) == "" {
			continue
		}
		if usecounter {
			ct += 1
			html += fmt.Sprintf("(%d)&nbsp;", ct)
		}
		html += fmt.Sprintf(obs, m.Headwd, m.Xrefval, m.Headwd)
		if strings.TrimSpace(m.Transl) != "" {
			html += fmt.Sprintf(bft, m.Transl)
		}
		pos := strings.Split(m.Anal, " ")
		var tab string
		for _, p := range pos {
			tab += fmt.Sprintf(mtd, "morphcell", p)
		}
		tab = fmt.Sprintf(mtr, tab)
		tab = fmt.Sprintf(mtb, tab)
		html += tab
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

	elem = append(elem, headwordprevalence(hwc))
	elem = append(elem, headworddistrib(hwc))

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
