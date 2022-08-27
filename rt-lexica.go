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

func RtLexFindByForm(c echo.Context) error {
	// be able to respond to "GET /lexica/findbyform/ἀμιϲθὶ/gr0062 HTTP/1.1"
	req := c.Param("id")
	elem := strings.Split(req, "/")
	fmt.Println(elem)
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

	b := findbyform(word, au)

	return c.String(http.StatusOK, string(b))
}

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

/*
	def _buildfullentry(self) -> str:
		fullentrystring = '<br /><br />\n<span class="lexiconhighlight">Full entry:</span><br />'
		suppressedmorph = '<br /><br />\n<span class="lexiconhighlight">(Morphology notes hidden)</span><br />'
		w = self.thiswordobject
		w.constructsensehierarchy()
		w.runbodyxrefsuite()
		w.insertclickablelookups()
		# next is optional, really: a good CSS file will parse what you have thus far
		# (HipparchiaServer v.1.1.2 has the old XML CSS)
		w.xmltohtmlconversions()
		segments = list()
		segments.append(w.grabheadmaterial())
		# segments.append(suppressedmorph)
		segments.append(fullentrystring)
		segments.append(w.grabnonheadmaterial())
		fullentry = '\n'.join(segments)
		return fullentry


<div style="position: absolute; height: auto; width: 624.03px; top: 47px; left: 385.167px;" tabindex="-1" role="dialog" class="ui-dialog ui-corner-all ui-widget ui-widget-content ui-front ui-draggable ui-resizable" aria-describedby="lexicadialogtext" aria-labelledby="ui-id-43"><div class="ui-dialog-titlebar ui-corner-all ui-widget-header ui-helper-clearfix ui-draggable-handle"><span id="ui-id-43" class="ui-dialog-title">finitio</span><button type="button" class="ui-button ui-corner-all ui-widget ui-button-icon-only ui-dialog-titlebar-close" title="Close"><span class="ui-button-icon ui-icon ui-icon-closethick"></span><span class="ui-button-icon-space"> </span>Close</button></div><div id="lexicadialogtext" class="ui-dialog-content ui-widget-content" style="width: auto; min-height: 86.3334px; max-height: 1253.93px; height: auto;"><p class="wordcounts">Prevalence (this form):
    <!-- lexicaformatting.py formatprevalencedata() output begins -->
    <span class="prevalence">Ⓖ</span> 5 / <span class="prevalence">Ⓛ</span> 44 / <span class="prevalence">Ⓣ</span> 49
    <!-- lexicaformatting.py formatprevalencedata() output ends -->
    </p>

    <!-- lexicaformatting.py formatparsinginformation() output begins -->


    <span class="obsv">
        <span class="dictionaryform">finitio</span>&nbsp;:&nbsp;
        <a class="parsing" href="#finitio_30382620"><span class="obsv">
            from <span class="baseform">finitio</span>
            <span class="baseformtranslation">&nbsp;(“<span class="transtree">I.</span> A <cb n="FIRM"> limiting; <span class="transtree">II.</span> A determining; <span class="transtree">III.</span> An end”)</cb></span>
        </span></a>
    </span>


    <table class="morphtable">
        <tbody>
            <tr><td class="morphcell invisible">[a]</td>
<td class="morphcell">fem</td>
<td class="morphcell">nom/voc</td>
<td class="morphcell">sg</td>
</tr>
        </tbody>
    </table>

    <!-- lexicaformatting.py formatparsinginformation() output ends -->

<div id="finitio_30382620">
<hr><p class="dictionaryheading" id="finitio_18205.0">finitio
&nbsp;<span class="metrics">[fīnītĭo]</span>
</p>

    <!-- lexicaloutputobjects.py _buildprincipleparts() output begins -->

        <formsummary parserxref="30382620" lexicalid="18205.0" headword="finitio" lang="latin">known forms in use: 15</formsummary>
        <table class="morphtable">
            <tbody>
                <tr><th class="morphcell labelcell" rowspan="1" colspan="2"></th></tr>

            </tbody>
        </table>

    <!-- lexicaloutputobjects.py _buildprincipleparts() output ends -->

<p class="wordcounts">Prevalence (all forms):

    <!-- lexicaformatting.py formatprevalencedata() output begins -->
    <span class="prevalence">Ⓖ</span> 5 / <span class="prevalence">Ⓛ</span> 148 / <span class="prevalence">Ⓣ</span> 153

</p><p class="wordcounts">Weighted distribution by corpus:
<span class="prevalence">Ⓛ</span> 100 / <span class="prevalence">Ⓖ</span> 0 / <span class="prevalence">Ⓘ</span> 0 / <span class="prevalence">Ⓓ</span> 0 / <span class="prevalence">Ⓒ</span> 0
</p>
<p class="wordcounts">Predominant genres:
<span class="emph">allrhet</span>&nbsp;(100), <span class="emph">agric</span>&nbsp;(77), <span class="emph">astron</span>&nbsp;(33), <span class="emph">nathist</span>&nbsp;(32), <span class="emph">lexicogr</span>&nbsp;(27), <span class="emph">phil</span>&nbsp;(20), <span class="emph">polyhist</span>&nbsp;(14), <span class="emph">gramm</span>&nbsp;(11)
    <!-- lexicaformatting.py formatprevalencedata() output ends -->

</p>

    <!-- lexicaloutputobjects.py _buildentrysummary() output begins -->

    <!-- lexicaformatting.py formatdictionarysummary() output begins -->
    <div class="sensesummary"><span class="lexiconhighlight">Senses</span><br>
<span class="sensesum">13 senses</span><br>
</div>
<div class="authorsummary"><span class="lexiconhighlight">Citations from</span><br>
<span class="authorsum">7 authors</span><br>
</div>
<div class="quotessummary"><span class="lexiconhighlight">Quotes</span><br>
<span class="quotesum">4 quotes</span><br>
</div><br>
    <!-- lexicaformatting.py formatdictionarysummary() output ends -->

    <!-- lexicaloutputobjects.py _buildentrysummary() output ends -->

<span class="dictorth dictlang_la">fīnītĭo</span>, <span class="dictitype">ōnis</span>, <span class="dictgen">f.</span> <span class="dictetym"><dictionaryentry id="finio">finio</dictionaryentry></span> (post-Aug.).

<br><br>
<span class="lexiconhighlight">Full entry:</span><br>
        <p class="level1">
            <span class="levellabel1">I</span>
            <sense id="n18205.0" level="1"><span class="dicthi dictrend_ital">A <span class="dictcb"> limiting</span>, <span class="dicthi dictrend_ital">limit</span>, <span class="dicthi dictrend_ital">boundary</span>, <bibl id="perseus/lt1056/001/2:1"><span class="dictauthor">Vitruvius</span> 2, 1 <span class="dicthi dictrend_ital">fin.</span></bibl>; <bibl id="perseus/lt1056/001/5:4">5, 4</bibl> <span class="dicthi dictrend_ital">fin.</span>; 8, 1.—</span></sense>
        </p>
        <p class="level1">
            <span class="levellabel1">II</span>
            <sense id="n18205.1" level="1"> <span class="dicthi dictrend_ital">A determining</span>, <span class="dicthi dictrend_ital">assigning</span>, viz., </sense>
        </p>
        <p class="level2">
            <span class="levellabel2">A</span>
            <sense id="n18205.2" level="2"> <span class="dictusg dicttype_style">Lit.</span>, <span class="dicthi dictrend_ital">a division</span>, <span class="dicthi dictrend_ital">part</span>, <bibl><span class="dictauthor">Hyginus</span> Astr. 1, 6 <span class="dicthi dictrend_ital">fin.</span></bibl>—</sense>
        </p>
        <p class="level2">
            <span class="levellabel2">B</span>
            <sense id="n18205.3" level="2"> <span class="dictusg dicttype_style">Trop.</span> </sense>
        </p>
        <p class="level3">
            <span class="levellabel3">1</span>
            <sense id="n18205.4" level="3"> <span class="dicthi dictrend_ital">A definition</span>, <span class="dicthi dictrend_ital">explanation</span> (esp. freq. in Quint.): <span class="dictcit"><span class="dictquote dictlang_la">finitio est rei propositae propria et dilucida et breviter comprehensa verbis enunciatio,</span> <bibl><span class="dictauthor">Quintilian</span> 7, 3, 2 sq.</bibl></span>; <bibl id="perseus/lt1002/001/2:15:34">2, 15, 34</bibl>; <bibl id="perseus/lt1002/001/3:6:49">3, 6, 49</bibl>; 5, 10, 63 et saep.; <bibl id="perseus/lt1254/001/15:9:11"><span class="dictauthor">Gellius</span> 15, 9, 11</bibl>.—</sense>
        </p>
        <p class="level3">
            <span class="levellabel3">2</span>
            <sense id="n18205.5" level="3"> <span class="dicthi dictrend_ital">A rule</span>: <span class="dictcit"><span class="dictquote dictlang_la">illam quasi finitionem veluti quandam legem sanxerunt, eos tantum surculos posse coalescere, qui, etc.,</span> <bibl id="perseus/lt0845/001/5:11:12"><span class="dictauthor">Columella</span> 5, 11, 12</bibl></span>.—</sense>
        </p>
        <p class="level1">
            <span class="levellabel1">III</span>
            <sense id="n18205.6" level="1"> <span class="dicthi dictrend_ital">An end;</span> esp., </sense>
        </p>
        <p class="level2">
            <span class="levellabel2">A</span>
            <sense id="n18205.7" level="2"> <span class="dicthi dictrend_ital">The end of life</span>, <span class="dicthi dictrend_ital">death</span>, <unclickablebibl><span class="dictauthor">Inscr. Grut.</span> 810, 10</unclickablebibl>: <span class="dictcit"><span class="dictquote dictlang_la">FATI,</span> <unclickablebibl><span class="dictauthor">Inscr. Orell.</span> 4776</unclickablebibl></span>.—</sense>
        </p>
        <p class="level2">
            <span class="levellabel2">B</span>
            <sense id="n18205.8" level="2"> <span class="dicthi dictrend_ital">Completeness</span>: <span class="dictcit"><span class="dictquote dictlang_la">progressum esse ad hanc finitionem,</span> <bibl id="perseus/lt1056/001/2:1:8"><span class="dictauthor">Vitruvius</span> 2, 1, 8</bibl></span>.</sense>
        </p>

        <table class="navtable">
        <tbody><tr>
            <td class="alignleft">
                <span class="label">Previous: </span>
                <dictionaryidsearch entryid="18204.0" language="latin">finitimus</dictionaryidsearch>
            </td>
            <td>&nbsp;</td>
            <td class="alignright">
                <span class="label">Next: </span>
                <dictionaryidsearch entryid="18206.0" language="latin">finitivus</dictionaryidsearch>
            </td>
        </tr><tr>
        </tr></tbody></table>

</div></div><div class="ui-resizable-handle ui-resizable-n" style="z-index: 90;"></div><div class="ui-resizable-handle ui-resizable-e" style="z-index: 90;"></div><div class="ui-resizable-handle ui-resizable-s" style="z-index: 90;"></div><div class="ui-resizable-handle ui-resizable-w" style="z-index: 90;"></div><div class="ui-resizable-handle ui-resizable-se ui-icon ui-icon-gripsmall-diagonal-se" style="z-index: 90;"></div><div class="ui-resizable-handle ui-resizable-sw" style="z-index: 90;"></div><div class="ui-resizable-handle ui-resizable-ne" style="z-index: 90;"></div><div class="ui-resizable-handle ui-resizable-nw" style="z-index: 90;"></div></div>
*/
