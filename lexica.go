package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4"
	"reflect"
	"regexp"
	"strings"
)

// initial target: be able to respond to "GET /lexica/findbyform/ἀμιϲθὶ/gr0062 HTTP/1.1"
// full set of verbs: lookup, findbyform, idlookup, morphologychart

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

	dbpool := grabpgsqlconnection()
	fld := `observed_form, xrefs, prefixrefs, possible_dictionary_forms, related_headwords`
	psq := fmt.Sprintf("SELECT %s FROM %s_morphology WHERE observed_form = '%s'", fld, d, word)

	var foundrows pgx.Rows
	var err error

	foundrows, err = dbpool.Query(context.Background(), psq)
	checkerror(err)

	var thesefinds []DbMorphology
	defer foundrows.Close()
	for foundrows.Next() {
		var thehit DbMorphology
		err := foundrows.Scan(&thehit.Observed, &thehit.Xrefs, &thehit.PrefixXrefs, &thehit.RawPossib, &thehit.RelatedHW)
		checkerror(err)
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
				checkerror(err)
			}
			mpp = append(mpp, mp)
		}
	}

	// fmt.Println(mpp)

	// [d] take the []MorphPossib and find the set of headwords we are interested in

	var hwm []string
	for _, p := range mpp {
		if len(p.Headwd) > 0 {
			hwm = append(hwm, p.Headwd)
		}
	}

	// the next is primed to produce problems: see καρποῦ which will turn καρπόϲ1 and καρπόϲ2 into just καρπόϲ; need xref_value?
	// but we have probably taken care of this below: see the comments
	hwm = unique(hwm)
	// fmt.Println(hwm)

	// [e] get the wordobjects for each unique headword: probedictionary()

	// note that the greek and latin dictionaries have extra fields that we are not using (right?)
	//var ec string
	//if d == "latin" {
	//	ec = "entry_key"
	//} else {
	//	ec = "unaccented_entry"
	//}

	// fld = fmt.Sprintf(`entry_name, metrical_entry, id_number, pos, translations, entry_body, %s`, ec)

	fld = `entry_name, metrical_entry, id_number, pos, translations, entry_body`
	psq = `SELECT %s FROM %s_dictionary WHERE %s ~* '^%s(|¹|²|³|⁴)$' ORDER BY id_number ASC`
	col := "entry_name"

	var lexicalfinds []DbLexicon
	dedup := make(map[int64]bool)
	for _, w := range hwm {
		// var foundrows pgx.Rows
		var err error
		q := fmt.Sprintf(psq, fld, d, col, w)
		foundrows, err = dbpool.Query(context.Background(), q)
		checkerror(err)

		defer foundrows.Close()
		for foundrows.Next() {
			var thehit DbLexicon
			err := foundrows.Scan(&thehit.Word, &thehit.Metrical, &thehit.ID, &thehit.POS, &thehit.Transl, &thehit.Entry)
			checkerror(err)
			if _, dup := dedup[thehit.ID]; !dup {
				// use ID and not Word because καρπόϲ.53442 is not καρπόϲ.53443
				dedup[thehit.ID] = true
				lexicalfinds = append(lexicalfinds, thehit)
			}
		}
	}

	//for _, x := range lexicalfinds {
	//	fmt.Println(fmt.Sprintf("%s: %s", x.Word, x.Transl))
	//}

	// [f] generate and format the prevalence data for this form: cf formatprevalencedata() in lexicalformatting.py

	fld = `entry_name, total_count, gr_count, lt_count, dp_count, in_count, ch_count`
	psq = `SELECT %s FROM wordcounts_%s where entry_name = '%s'`
	// golang hates indexing unicode strings: strings are bytes, and unicode chars take more than one byte
	c := []rune(word)
	q := fmt.Sprintf(psq, fld, stripaccents(string(c[0])), word)

	foundrows, err = dbpool.Query(context.Background(), q)
	checkerror(err)
	var wc DbWordCount
	defer foundrows.Close()
	for foundrows.Next() {
		// only one should ever return...
		err := foundrows.Scan(&wc.Word, &wc.Total, &wc.Gr, &wc.Lt, &wc.Dp, &wc.In, &wc.Ch)
		checkerror(err)
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

	//fmt.Println(allformpd)
	//fmt.Println(parsing)

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
	checkerror(ee)

	// jsonbundle := []byte(fmt.Sprintf(`{"newhtml":"%s","newjs":"%s"}`, html, js))
	return jsonbundle
}

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
	<span class="baseformtranslation">&nbsp;(“%s”)</span></span></a></span>`
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
	for i, m := range mpp {
		if usecounter {
			html += fmt.Sprintf("(%d)&nbsp;", i+1)
		}
		html += fmt.Sprintf(obs, m.Headwd, m.Xrefval, m.Headwd, m.Transl)
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
	// [h1] first part of a lexical entry:

	// [h1a] known forms in use

	// requires probing dictionary_headword_wordcounts

	// 		SELECT
	//			entry_name , total_count, gr_count, lt_count, dp_count, in_count, ch_count,
	//			frequency_classification, early_occurrences, middle_occurrences ,late_occurrences,
	//			acta, agric, alchem, anthol, apocalyp, apocryph, apol, astrol, astron, biogr, bucol, caten, chronogr, comic, comm,
	//			concil, coq, dialog, docu, doxogr, eccl, eleg, encom, epic, epigr, epist, evangel, exeget, fab, geogr, gnom, gramm,
	//			hagiogr, hexametr, hist, homilet, hymn, hypoth, iamb, ignotum, invectiv, inscr, jurisprud, lexicogr, liturg, lyr,
	//			magica, math, mech, med, metrolog, mim, mus, myth, narrfict, nathist, onir, orac, orat, paradox, parod, paroem,
	//			perieg, phil, physiognom, poem, polyhist, prophet, pseudepigr, rhet, satura, satyr, schol, tact, test, theol, trag
	//		FROM dictionary_headword_wordcounts WHERE entry_name='%s'

	// [h1b] principle parts

	// [h2] wordcounts data including weighted distributions

	// [h3]  _buildentrysummary() which gives senses, flagged senses, citations, quotes (summarized or not)

	// [h4] the actual body of the entry

	// more formatting to come

	entrybody := w.Entry

	// [h5] previous & next entry
	enfolded := `<div id="%s_%d">%s</div>
	`
	html := fmt.Sprintf(enfolded, w.Word, w.ID, entrybody)
	return html
}

func insertlexicaljs() string {
	js := `
	<script>
	// Chromium can send poll data after the search is done... 
	$('#pollingdata').hide();
	
	$('%s').click( function() {
		$.getJSON('/browse/'+this.id, function (passagereturned) {
			$('#browseforward').unbind('click');
			$('#browseback').unbind('click');
			var fb = parsepassagereturned(passagereturned)
			// left and right arrow keys
			$('#browserdialogtext').keydown(function(e) {
				switch(e.which) {
					case 37: browseuponclick(fb[1]); break;
					case 39: browseuponclick(fb[0]); break;
				}
			});
			$('#browseforward').bind('click', function(){ browseuponclick(fb[0]); });
			$('#browseback').bind('click', function(){ browseuponclick(fb[1]); });
		});
	});
	</script>`

	tag := "bibl"

	thejs := fmt.Sprintf(js, tag)
	return thejs
}
