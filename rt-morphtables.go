//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/labstack/echo/v4"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

const (
	DECLTABTMPL = `
	<table class="verbanalysis">
	<tbody>
	{{.Header}}
	{{.Rows}}
	</tbody>
	</table>
	<hr class="styled">
`
	DECLHEADERTMPL = `
	<tr align="center">
		<td rowspan="1" colspan="{{.Span}}" class="dialectlabel">{{.Dialect}}<br>
		</td>
	</tr>
	{{.Gendlabel}}
	</tr>
`
	DECLGENDLABEL = `<tr>
		<td class="genderlabel">&nbsp;</td>
		{{.AllGenders}}
	</tr>`

	DECLBLANK = `
	<tr><td>&nbsp;</td>{{.Columns}}</tr>
`
	DECLGENDERCELL = `<td class="gendercell">{{.G}}<br></td>`

	DECLMORPHROW = `	
	<tr class="morphrow">
		{{.AllCells}}
	</tr>`

	MORPHLABELCELL = `<td class="morphlabelcell">{{.Ml}}</td>`

	// MORPHCELL - 2nd sg attic mid indic of τίκτω should yield: τέξηι / τέξει / τέξῃ / τεκῇ
	MORPHCELL = `<td class="morphcell">{{.Mo}}</td>`

	// 	Smythe §383ff
	//
	//	verb cells look like:
	//
	//		<td class="morphcell">_attic_subj_pass_pl_2nd_pres_</td>

	VBTABTMPL = `
	<table class="verbanalysis">
	<tbody>
	{{.Header}}
	{{.Rows}}
	</tbody>
	</table>
	<hr class="styled">`

	VBHEADERTMPL = `	
	<tr align="center">
		<td rowspan="1" colspan="{s}" class="dialectlabel">{dialect}<br>
		</td>
	</tr>
	<tr align="center">
		<td rowspan="1" colspan="{s}" class="voicelabel">{voice}<br>
		</td>
	</tr>
	<tr align="center">
		<td rowspan="1" colspan="{s}" class="moodlabel">{mood}<br>
		</td>
	{{.Tenseheader}}
	</tr>`

	VBTENSETEMPL = `
	<tr>
		<td class="tenselabel">&nbsp;</td>
		{{.Alltenses}}
	</tr>`

	VBBLANK      = DECLBLANK
	VMMORPHROW   = DECLMORPHROW
	VBREGEXTEMPL = `_{{.D}}_{{.M}}_{{.V}}_{{.N}}_{{.P}}_{{.T}}_`
	PCPLTEMPL    = `_{{.D}}_{{.M}}_{{.V}}_{{.N}}_{{.T}}_{{.G}}_{{.C}}_`
)

var (
	GKCASES  = []string{"nom", "gen", "dat", "acc", "voc"}
	GKNUMB   = []string{"sg", "dual", "pl"}
	GKMOODS  = []string{"ind", "subj", "opt", "imperat", "inf", "part"}
	GKVOICE  = []string{"act", "mid", "pass"}
	GKTENSES = []string{"aor", "fut", "futperf", "imperf", "perf", "pres", "plup"}
	// GKTENSEMAP = map[int]string{1: "Present", 2: "Imperfect", 3: "Future", 4: "Aorist", 5: "Perfect", 6: "Pluperfect", 7: "Future Perfect"}
	GKTENSEMAP = map[string]int{"pres": 1, "imperf": 2, "fut": 3, "aor": 4, "perf": 5, "plup": 6, "futperf": 7}
	GKVERBS    = getgkvbmap()
	GKDIALECT  = []string{"attic"} // INCOMPLETE
	LTCASES    = []string{"nom", "gen", "dat", "acc", "abl", "voc"}
	LTNUMB     = []string{"sg", "pl"}
	LTMOODS    = []string{"ind", "subj", "imperat", "inf", "part", "gerundive", "supine"}
	LTVOICE    = []string{"act", "pass"}
	LTTENSES   = []string{"fut", "futperf", "imperf", "perf", "pres", "plup"}
	LTTENSEMAP = map[int]string{1: "Present", 2: "Imperfect", 3: "Future", 5: "Perfect", 6: "Pluperfect", 7: "Future Perfect"}
	LTVERBS    = getltvbmap()
	GENDERS    = []string{"masc", "fem", "neut"}
	PERSONS    = []string{"1st", "2nd", "3rd"}
)

func getgkvbmap() map[string]map[string]map[int]bool {
	gvm := make(map[string]map[string]map[int]bool)
	for _, v := range GKVOICE {
		gvm[v] = make(map[string]map[int]bool)
		for _, m := range GKMOODS {
			gvm[v][m] = make(map[int]bool)
		}
	}

	gvm["act"]["ind"] = map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: false}
	gvm["act"]["subj"] = map[int]bool{1: true, 2: false, 3: false, 4: true, 5: true, 6: false, 7: false}
	gvm["act"]["opt"] = map[int]bool{1: true, 2: false, 3: true, 4: true, 5: true, 6: false, 7: false}
	gvm["act"]["imperat"] = map[int]bool{1: true, 2: false, 3: false, 4: true, 5: true, 6: false, 7: false}
	gvm["act"]["inf"] = map[int]bool{1: true, 2: false, 3: true, 4: true, 5: true, 6: false, 7: false}
	gvm["act"]["part"] = map[int]bool{1: true, 2: false, 3: true, 4: true, 5: true, 6: false, 7: false}
	gvm["mid"]["ind"] = map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: false}
	gvm["mid"]["subj"] = map[int]bool{1: true, 2: false, 3: false, 4: true, 5: true, 6: false, 7: false}
	gvm["mid"]["opt"] = map[int]bool{1: true, 2: false, 3: true, 4: true, 5: true, 6: false, 7: false}
	gvm["mid"]["imperat"] = map[int]bool{1: true, 2: false, 3: false, 4: true, 5: true, 6: false, 7: false}
	gvm["mid"]["inf"] = map[int]bool{1: true, 2: false, 3: true, 4: true, 5: true, 6: false, 7: false}
	gvm["mid"]["part"] = map[int]bool{1: true, 2: false, 3: true, 4: true, 5: true, 6: false, 7: false}
	gvm["pass"]["ind"] = map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true}
	gvm["pass"]["subj"] = map[int]bool{1: true, 2: false, 3: false, 4: true, 5: true, 6: false, 7: false}
	gvm["pass"]["opt"] = map[int]bool{1: true, 2: false, 3: true, 4: true, 5: true, 6: false, 7: true}
	gvm["pass"]["imperat"] = map[int]bool{1: true, 2: false, 3: false, 4: true, 5: true, 6: false, 7: false}
	gvm["pass"]["inf"] = map[int]bool{1: true, 2: false, 3: true, 4: true, 5: true, 6: false, 7: true}
	gvm["pass"]["part"] = map[int]bool{1: true, 2: false, 3: true, 4: true, 5: true, 6: false, 7: true}
	return gvm
}

func getltvbmap() map[string]map[string]map[int]bool {
	// note that ppf subj pass, etc are "false" because "laudātus essem" is not going to be found

	lvm := make(map[string]map[string]map[int]bool)
	for _, v := range LTVOICE {
		lvm[v] = make(map[string]map[int]bool)
		for _, m := range LTMOODS {
			lvm[v][m] = make(map[int]bool)
		}
	}
	lvm["act"]["ind"] = map[int]bool{1: true, 2: true, 3: true, 5: true, 6: true, 7: true}
	lvm["act"]["subj"] = map[int]bool{1: true, 2: false, 3: false, 5: true, 6: true, 7: false}
	lvm["act"]["imperat"] = map[int]bool{1: true, 2: false, 3: true, 5: false, 6: false, 7: false}
	lvm["act"]["inf"] = map[int]bool{1: true, 2: false, 3: false, 5: true, 6: false, 7: false}
	lvm["act"]["part"] = map[int]bool{1: true, 2: false, 3: true, 5: false, 6: false, 7: false}
	lvm["pass"]["ind"] = map[int]bool{1: true, 2: true, 3: true, 5: false, 6: false, 7: false}
	lvm["pass"]["subj"] = map[int]bool{1: true, 2: true, 3: false, 5: false, 6: false, 7: false}
	lvm["pass"]["imperat"] = map[int]bool{1: true, 2: false, 3: true, 5: false, 6: false, 7: false}
	lvm["pass"]["inf"] = map[int]bool{1: true, 2: false, 3: false, 5: false, 6: false, 7: false}
	lvm["pass"]["part"] = map[int]bool{1: false, 2: false, 3: false, 5: true, 6: false, 7: false}
	return lvm
}

func RtMorphchart(c echo.Context) error {
	// /lexica/morphologychart/greek/39046.0/37925260/ἐπιγιγνώϲκω

	// hipparchiaDB=# \d greek_morphology
	//                           Table "public.greek_morphology"
	//          Column           |          Type          | Collation | Nullable | Default
	//---------------------------+------------------------+-----------+----------+---------
	// observed_form             | character varying(64)  |           |          |
	// xrefs                     | character varying(128) |           |          |
	// prefixrefs                | character varying(128) |           |          |
	// possible_dictionary_forms | jsonb                  |           |          |
	// related_headwords         | character varying(256) |           |          |
	//Indexes:
	//    "greek_analysis_trgm_idx" gin (related_headwords gin_trgm_ops)
	//    "greek_morphology_idx" btree (observed_form)

	// should reach this route exclusively via a click from rt-lexica.go
	//kf := `<formsummary parserxref="%d" lexicalid="%.1f" headword="%s" lang="%s">%d known forms</formsummary>`
	//kf = fmt.Sprintf(kf, AllLemm[w.Word].Xref, w.ID, w.Word, w.Lang, len(AllLemm[w.Word].Deriv))

	if cfg.LogLevel < 3 {
		// no calling this route unless debugging
		return emptyjsreturn(c)
	}

	// [a] parse request

	req := c.Param("wd")
	elem := strings.Split(req, "/")

	if len(elem) != 4 || elem[0] == "" {
		return emptyjsreturn(c)
	}

	lg := elem[0]
	_, e1 := strconv.ParseFloat(elem[1], 32)
	_, e2 := strconv.Atoi(elem[2])
	// wd := purgechars(UNACCEPTABLEINPUT, elem[3])
	gl := lg == "greek" || lg == "latin"

	if !gl || e1 != nil || e2 != nil {
		return emptyjsreturn(c)
	}

	// if e2 == nil it is safe to use elem[2] as the (string) xref val
	xr := elem[2]

	// [b] get all forms of the word

	// for ἐπιγιγνώϲκω...
	// select * from greek_morphology where greek_morphology.xrefs='37925260';

	dbpool := GetPSQLconnection()
	defer dbpool.Close()

	fld := `observed_form, xrefs, prefixrefs, possible_dictionary_forms, related_headwords`
	psq := fmt.Sprintf(`SELECT %s FROM %s_morphology WHERE xrefs='%s'`, fld, lg, xr)

	var foundrows pgx.Rows
	var err error

	foundrows, err = dbpool.Query(context.Background(), psq)
	chke(err)

	dbmmap := make(map[string]DbMorphology)
	defer foundrows.Close()
	for foundrows.Next() {
		var thehit DbMorphology
		e := foundrows.Scan(&thehit.Observed, &thehit.Xrefs, &thehit.PrefixXrefs, &thehit.RawPossib, &thehit.RelatedHW)
		chke(e)
		dbmmap[thehit.Observed] = thehit
	}

	// [c] get all counts for all forms: [c] and [d-e] can run concurrently
	// [c1] slice of the words; map of the first letters of those words
	ww := make([]string, len(dbmmap))
	lett := make(map[string]bool)

	count := 0
	for _, f := range dbmmap {
		fo := f.Observed
		ww[count] = fo
		r := []rune(fo)
		init := stripaccentsRUNE(r)
		lett[string(init[0])] = true
		count += 1
	}

	// [c2] query the database

	// pgsql single quote escape: quote followed by a single quote to be escaped: κρυφθεῖϲ''

	var esc []string
	for _, w := range ww {
		//x := swapgraveforacute(w)
		//x = strings.Replace(x, "'", "''", -1)
		//esc = append(esc, x)
		esc = append(esc, strings.Replace(w, "'", "''", -1))
	}
	arr := fmt.Sprintf("'%s'", strings.Join(esc, "', '"))

	// hipparchiaDB=# CREATE TEMPORARY TABLE ttw AS
	//    SELECT values AS wordforms FROM
	//      unnest(ARRAY['κόραϲ', 'κόραι', 'κῶραι', 'κούρῃϲιν', 'κούραϲ', 'κούραιϲιν', 'κόραν', 'κώρα', 'κόραιϲιν', 'κόραιϲι', 'κόρα', 'κόρᾳϲ'])
	//    values;
	//
	//SELECT entry_name, total_count FROM wordcounts_κ WHERE EXISTS (
	//  (SELECT 1 FROM ttw temptable WHERE temptable.wordforms = wordcounts_κ.entry_name )
	//);
	//SELECT 12
	// entry_name | total_count
	//------------+-------------
	// κόραν      |          59
	// κούραιϲιν  |           1
	// κῶραι      |           4
	// κόρᾳϲ      |           1
	// κούρῃϲιν   |           9
	// κόραι      |         363
	// κόραϲ      |         668
	// κόραιϲιν   |           2
	// κόραιϲι    |           8
	// κούραϲ     |          89
	// κόρα       |          72
	// κώρα       |           9
	//(12 rows)

	tt := `CREATE TEMPORARY TABLE ttw_%s AS SELECT values AS wordforms FROM unnest(ARRAY[%s]) values`
	qt := `SELECT entry_name, total_count FROM wordcounts_%s WHERE EXISTS 
		(SELECT 1 FROM ttw_%s temptable WHERE temptable.wordforms = wordcounts_%s.entry_name)`

	wcc := make(map[string]DbWordCount)
	for l, _ := range lett {
		rnd := strings.Replace(uuid.New().String(), "-", "", -1)
		_, e := dbpool.Exec(context.Background(), fmt.Sprintf(tt, rnd, arr))
		chke(e)
		rr, e := dbpool.Query(context.Background(), fmt.Sprintf(qt, l, rnd, l))
		chke(e)
		var wc DbWordCount
		defer rr.Close()
		for rr.Next() {
			ee := rr.Scan(&wc.Word, &wc.Total)
			chke(ee)
			wcc[wc.Word] = wc
		}
	}

	// [d] extract parsing info for all forms

	mpp := make(map[string][]string)

	for k, v := range dbmmap {
		vv := []DbMorphology{v} // dbmorphintomorphpossib() wants a slice, we fake a slice
		mp := dbmorphintomorphpossib(vv)
		for _, m := range mp {
			// item 0 is always ""; item 1 is an actual analysis
			mpp[k] = append(mpp[k], m.Anal)
		}
	}

	// [e] generate parsing map: [parsedata]form
	// NB have to decompress "nom/voc/acc" into three entries

	// [e1] first pass: make the map and deal with cases
	pdm := make(map[string]string)
	for k, vv := range mpp {
		for _, v := range vv {
			if len(v) == 0 {
				continue
			}
			if !strings.Contains(v, "/") {
				key := strings.Replace(v, " ", "_", -1)
				if _, ok := pdm[key]; !ok {
					pdm[key] = k
				} else {
					pdm[key] = pdm[key] + " / " + k
				}
			} else {
				// need to decompress "nom/voc/acc" into three entries, etc
				var rebuild []string
				var multiplier []string
				ell := strings.Split(v, " ")
				for _, e := range ell {
					if !strings.Contains(e, "/") {
						rebuild = append(rebuild, e)
					} else {
						multiplier = strings.Split(e, "/")
						rebuild = append(rebuild, "CLONE_ME")
					}
				}
				templ := strings.Join(rebuild, " ")
				for _, m := range multiplier {
					key := strings.Replace(templ, "CLONE_ME", m, 1)
					key = strings.Replace(key, " ", "_", -1)
					if _, ok := pdm[key]; !ok {
						pdm[key] = k
					} else {
						pdm[key] = pdm[key] + " / " + k
					}
				}
			}
		}
	}

	// [e2] second pass at the map to deal with dialects
	if lg == "greek" {
		for k, v := range pdm {
			delete(pdm, k)
			if strings.Contains(k, "(") {
				k = strings.Replace(k, ")", "", 1)
				parts := strings.Split(k, "(")
				diall := strings.Split(parts[1], "_")
				for _, d := range diall {
					newkey := parts[0] + "_" + d
					newkey = strings.Replace(newkey, "__", "_", 1)
					pdm[newkey] = v
				}
			} else {
				if !strings.Contains(k, "attic") {
					newkey := k + "_attic"
					pdm[newkey] = v
				} else {
					pdm[k] = v
				}
			}
		}
	}

	// [e3] get counts for each word
	pdcm := make(map[string]map[string]int64)
	for k, v := range pdm {
		wds := strings.Split(v, " / ")
		mm := make(map[string]int64)
		for _, w := range wds {
			mm[w] = wcc[w].Total
		}
		pdcm[k] = mm
	}

	var oo []string
	pdxm := make(map[string]string)
	for kk, pd := range pdcm {
		var vv []string
		for k, v := range pd {
			vv = append(vv, fmt.Sprintf("%s (%d)", k, v))
		}
		pdxm[kk] = strings.Join(vv, " / ")

		oo = append(oo, fmt.Sprintf("%s: %s\n", kk, strings.Join(vv, " / ")))

		//gen_cas_num_dial
		//fem_acc_dual_attic: κόρα (72)
		//fem_acc_dual_doric: κώρα (9)
		//fem_acc_dual_epic_attic: κούρα (62)
		//fem_acc_dual_ionic_attic: κούρα (62)
		// ...

		// tense_mood_voice_pers_numb_dial
		//aor_imperat_act_2nd_pl_attic: παραθλίψατε (1)
		//aor_imperat_act_2nd_sg_attic: θλῖψον (2)
		//aor_imperat_act_3rd_pl_attic: θλιψάντων (18)
		//aor_imperat_mid_2nd_sg_attic: θλῖψαι (25)
		// ...
	}

	// [f] determine if it is a verb or declined
	dc := 0
	vc := 0
	for key := range pdxm {
		k := strings.Split(key, "_")
		if isinslice(GKTENSES, k[0]) {
			vc += 1
		}
		if isinslice(LTCASES, k[0]) {
			dc += 1
		}
	}

	isverb := true
	if dc > vc {
		isverb = false
	}

	var jb JSB
	sort.Strings(oo)

	if isverb {
		jb.HTML = generateverbtable(pdxm)
	} else {
		jb.HTML = "[RtMorphchart() is a work in progress...]<br>" + strings.Join(oo, "<br>")
	}

	// [g] build the table head

	// [h] build the table body

	// return emptyjsreturn(c)

	jb.JS = insertlexicaljs()

	return c.JSONPretty(http.StatusOK, jb, JSONINDENT)
}

func generateverbtable(words map[string]string) string {
	// first voice
	// then mood
	// then tense as columns and number_and_person as rows
	const (
		BLANK = " --- "
	)

	vm := getgkvbmap()
	tm := GKTENSEMAP
	skipper := false
	var html []string

	for _, d := range GKDIALECT {
		// each dialect is a major section
		html = append(html, fmt.Sprintf("<h3>%s</h3>", d))
		for _, v := range GKVOICE {
			// each voice is a section
			html = append(html, fmt.Sprintf("<h4>%s</h4>", v))
			for _, m := range GKMOODS {
				// each mood is a table
				html = append(html, "<table>")
				var trrhtml []string
				// <tr class="morphrow">
				//	<td class="morphlabelcell">sg 1st</td>
				//	<td class="morphcell"><verbform searchterm="πίτνω">πίτνω</verbform> (<span class="counter">15</span>) / <verbform searchterm="πίπτω">πίπτω</verbform> (<span class="counter">117</span>)</td>
				//	<td class="morphcell"><verbform searchterm="ἔπιπτον">ἔπιπτον</verbform> (<span class="counter">259</span>) / <verbform searchterm="ἔπιτνον">ἔπιτνον</verbform> (<span class="counter">3</span>)</td>
				//	<td class="morphcell">---</td>
				//	<td class="morphcell"><verbform searchterm="ἔπεϲον">ἔπεϲον</verbform> (<span class="counter">686</span>)</td>
				//	<td class="morphcell"><verbform searchterm="πέπτηκα">πέπτηκα</verbform> (<span class="counter">14</span>) / <verbform searchterm="πέπτωκα">πέπτωκα</verbform> (<span class="counter">67</span>)</td>
				//	<td class="morphcell"><verbform searchterm="ἐπεπτώκειν">ἐπεπτώκειν</verbform> (<span class="counter">1</span>)</td>
				//</tr>
				for _, n := range GKNUMB {
					for _, p := range PERSONS {
						// np := fmt.Sprintf("%s %s", n, p)
						html = append(html, `<tr class="morphrow">`)
						var tdd []string
						for _, t := range GKTENSES {
							k := fmt.Sprintf("%s_%s_%s_%s_%s_%s", t, m, v, p, n, d)
							if _, ok := words[k]; ok {
								tdd = append(tdd, words[k])
							} else {
								tdd = append(tdd, BLANK)
							}
							tv := tm[t]
							if vm[m][t][tv] {
								// for skipping impossible moot/tense combinations
								skipper = true
							}
						}
						for _, td := range tdd {
							html = append(html, fmt.Sprintf(`<td class="morphcell">%s</td>`, td))
						}
						html = append(html, `</tr>`)
					}
				}
				html = append(html, trrhtml...)
				html = append(html, "</table>")
			}
		}
	}
	fmt.Print(skipper)
	h := strings.Join(html, "")
	return h
}

func generatedeclinedformmap() {
	// 		miles :  from miles :
	//		[a]	masc/fem	nom	sg
	//
	//		operibus :  from opus¹  (“work”):
	//		[a]	neut	abl	pl
	//		[b]	neut
	//
	//		λοιβή :  from λοιβή  (“pouring.”):
	//		[a]	fem	nom/voc	sg	(attic	epic	ionic)
	//
	//		fd {'_ _sg_masc_gen_': ['dolorisque', 'dolorist', 'doloris', 'dolorisue'], '_ _pl_masc_acc_': ['dolores', 'doloresque'],
	//		'_ _pl_masc_nom_': ['dolores', 'doloresque'], '_ _pl_masc_voc_': ['dolores', 'doloresque'], ... }

}

func generateverbformmap() {
	// 		e.g. {'_attic_imperf_ind_mp_1st_pl_': 'ἠλαττώμεθα', ...}
	//		[a]	imperf	ind	mp	1st	pl	(attic	doric	aeolic)
	//		[b]	plup	ind	mp	1st	pl	(attic)
	//		[c]	perf	ind	mp	1st	pl	(attic)
	//		[d]	plup	ind	mp	1st	pl	(homeric	ionic)
	//
	//		cell arrangement: left to right and top to bottom is vmnpt
	//		i.e., voice, mood, number, person, tense
	//
	//		to be used with greekverbtabletemplate()

}

/*  SAMPLE: πίπτω

<div class="center">
	<span class="verylarge">All known forms of <dictionaryidsearch entryid="83253.0" language="greek">πίπτω</dictionaryidsearch></span>
</div>

<!-- morphologytableformatting.py filloutmorphtabletemplate() output begins -->

<table class="verbanalysis">
<tbody>

<tr align="center">
	<td rowspan="1" colspan="7" class="dialectlabel">attic<br>
	</td>
</tr>
<tr align="center">
	<td rowspan="1" colspan="7" class="voicelabel">act<br>
	</td>
</tr>
<tr align="center">
	<td rowspan="1" colspan="7" class="moodlabel">ind<br>
	</td>
</tr><tr>
	<td class="tenselabel">&nbsp;</td>
	<td class="tensecell">Present<br></td>
	<td class="tensecell">Imperfect<br></td>
	<td class="tensecell">Future<br></td>
	<td class="tensecell">Aorist<br></td>
	<td class="tensecell">Perfect<br></td>
	<td class="tensecell">Pluperfect<br></td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">sg 1st</td>
	<td class="morphcell"><verbform searchterm="πίτνω">πίτνω</verbform> (<span class="counter">15</span>) / <verbform searchterm="πίπτω">πίπτω</verbform> (<span class="counter">117</span>)</td>
	<td class="morphcell"><verbform searchterm="ἔπιπτον">ἔπιπτον</verbform> (<span class="counter">259</span>) / <verbform searchterm="ἔπιτνον">ἔπιτνον</verbform> (<span class="counter">3</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="ἔπεϲον">ἔπεϲον</verbform> (<span class="counter">686</span>)</td>
	<td class="morphcell"><verbform searchterm="πέπτηκα">πέπτηκα</verbform> (<span class="counter">14</span>) / <verbform searchterm="πέπτωκα">πέπτωκα</verbform> (<span class="counter">67</span>)</td>
	<td class="morphcell"><verbform searchterm="ἐπεπτώκειν">ἐπεπτώκειν</verbform> (<span class="counter">1</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">sg 2nd</td>
	<td class="morphcell"><verbform searchterm="πίτνειϲ">πίτνειϲ</verbform> (<span class="counter">2</span>) / <verbform searchterm="πίπτειϲ">πίπτειϲ</verbform> (<span class="counter">11</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="ἔπεϲεϲ">ἔπεϲεϲ</verbform> (<span class="counter">26</span>)</td>
	<td class="morphcell"><verbform searchterm="πέπτωκαϲ">πέπτωκαϲ</verbform> (<span class="counter">27</span>)</td>
	<td class="morphcell"><verbform searchterm="ἐπεπτώκειϲ">ἐπεπτώκειϲ</verbform> (<span class="counter">1</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">sg 3rd</td>
	<td class="morphcell"><verbform searchterm="πίτνει">πίτνει</verbform> (<span class="counter">21</span>) / <verbform searchterm="πίπτει">πίπτει</verbform> (<span class="counter">1125</span>)</td>
	<td class="morphcell"><verbform searchterm="ἔπιπτε">ἔπιπτε</verbform> (<span class="counter">56</span>) / <verbform searchterm="ἔπιτνε">ἔπιτνε</verbform> (<span class="counter">3</span>) / <verbform searchterm="ἔπιπτεν">ἔπιπτεν</verbform> (<span class="counter">75</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="ἔπεϲε">ἔπεϲε</verbform> (<span class="counter">526</span>) / <verbform searchterm="κἄπεϲε">κἄπεϲε</verbform> (<span class="counter">1</span>) / <verbform searchterm="ἔπεϲεν">ἔπεϲεν</verbform> (<span class="counter">995</span>)</td>
	<td class="morphcell"><verbform searchterm="πέπτωκεν">πέπτωκεν</verbform> (<span class="counter">460</span>) / <verbform searchterm="πέπτωκε">πέπτωκε</verbform> (<span class="counter">237</span>)</td>
	<td class="morphcell"><verbform searchterm="ἐπεπτώκει">ἐπεπτώκει</verbform> (<span class="counter">42</span>) / <verbform searchterm="πεπτώκει">πεπτώκει</verbform> (<span class="counter">4</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">dual 2nd</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">dual 3rd</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">pl 1st</td>
	<td class="morphcell"><verbform searchterm="πίτνομεν">πίτνομεν</verbform> (<span class="counter">1</span>) / <verbform searchterm="πίπτομεν">πίπτομεν</verbform> (<span class="counter">14</span>)</td>
	<td class="morphcell"><verbform searchterm="ἐπίπτομεν">ἐπίπτομεν</verbform> (<span class="counter">1</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="ἐπέϲομεν">ἐπέϲομεν</verbform> (<span class="counter">2</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτώκαμεν">πεπτώκαμεν</verbform> (<span class="counter">21</span>)</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">pl 2nd</td>
	<td class="morphcell"><verbform searchterm="πίπτετε">πίπτετε</verbform> (<span class="counter">29</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεπτώκατε">πεπτώκατε</verbform> (<span class="counter">6</span>)</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">pl 3rd</td>
	<td class="morphcell"><verbform searchterm="πίπτουϲιν">πίπτουϲιν</verbform> (<span class="counter">346</span>) / <verbform searchterm="πίπτουϲι">πίπτουϲι</verbform> (<span class="counter">262</span>)</td>
	<td class="morphcell"><verbform searchterm="ἔπιπτον">ἔπιπτον</verbform> (<span class="counter">259</span>) / <verbform searchterm="ἔπιτνον">ἔπιτνον</verbform> (<span class="counter">3</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="ἔπεϲον">ἔπεϲον</verbform> (<span class="counter">686</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτώκαϲιν">πεπτώκαϲιν</verbform> (<span class="counter">112</span>) / <verbform searchterm="πεπτώκαϲι">πεπτώκαϲι</verbform> (<span class="counter">97</span>)</td>
	<td class="morphcell"><verbform searchterm="ἐπεπτώκειϲαν">ἐπεπτώκειϲαν</verbform> (<span class="counter">3</span>) / <verbform searchterm="ἐπεπτώκεϲαν">ἐπεπτώκεϲαν</verbform> (<span class="counter">6</span>)</td>
</tr>

</tbody>
</table>
<hr class="styled">

<!-- morphologytableformatting.py filloutmorphtabletemplate() output ends -->


<!-- morphologytableformatting.py filloutmorphtabletemplate() output begins -->

<table class="verbanalysis">
<tbody>

<tr align="center">
	<td rowspan="1" colspan="7" class="dialectlabel">attic<br>
	</td>
</tr>
<tr align="center">
	<td rowspan="1" colspan="7" class="voicelabel">act<br>
	</td>
</tr>
<tr align="center">
	<td rowspan="1" colspan="7" class="moodlabel">subj<br>
	</td>
</tr><tr>
	<td class="tenselabel">&nbsp;</td>
	<td class="tensecell">Present<br></td>
	<td class="tensecell">Aorist<br></td>
	<td class="tensecell">Perfect<br></td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">sg 1st</td>
	<td class="morphcell"><verbform searchterm="πίτνω">πίτνω</verbform> (<span class="counter">15</span>) / <verbform searchterm="πίπτω">πίπτω</verbform> (<span class="counter">117</span>)</td>
	<td class="morphcell"><verbform searchterm="πέϲω">πέϲω</verbform> (<span class="counter">34</span>)</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">sg 2nd</td>
	<td class="morphcell"><verbform searchterm="πίπτῃϲ">πίπτῃϲ</verbform> (<span class="counter">3</span>)</td>
	<td class="morphcell"><verbform searchterm="πέϲῃϲ">πέϲῃϲ</verbform> (<span class="counter">57</span>) / <verbform searchterm="πέϲηιϲ">πέϲηιϲ</verbform> (<span class="counter">3</span>)</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">sg 3rd</td>
	<td class="morphcell"><verbform searchterm="πίπτῃ">πίπτῃ</verbform> (<span class="counter">57</span>)</td>
	<td class="morphcell"><verbform searchterm="πέϲῃ">πέϲῃ</verbform> (<span class="counter">496</span>) / <verbform searchterm="πέϲηι">πέϲηι</verbform> (<span class="counter">15</span>)</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">dual 2nd</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πέϲητον">πέϲητον</verbform> (<span class="counter">4</span>)</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">dual 3rd</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πέϲητον">πέϲητον</verbform> (<span class="counter">4</span>)</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">pl 1st</td>
	<td class="morphcell"><verbform searchterm="πίπτωμεν">πίπτωμεν</verbform> (<span class="counter">11</span>)</td>
	<td class="morphcell"><verbform searchterm="πέϲωμεν">πέϲωμεν</verbform> (<span class="counter">20</span>)</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">pl 2nd</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πέϲητε">πέϲητε</verbform> (<span class="counter">11</span>)</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">pl 3rd</td>
	<td class="morphcell"><verbform searchterm="πίπτωϲι">πίπτωϲι</verbform> (<span class="counter">14</span>) / <verbform searchterm="πίπτωϲιν">πίπτωϲιν</verbform> (<span class="counter">11</span>)</td>
	<td class="morphcell"><verbform searchterm="πέϲωϲι">πέϲωϲι</verbform> (<span class="counter">54</span>) / <verbform searchterm="πέϲωϲιν">πέϲωϲιν</verbform> (<span class="counter">70</span>)</td>
	<td class="morphcell">---</td>
</tr>

</tbody>
</table>
<hr class="styled">

<!-- morphologytableformatting.py filloutmorphtabletemplate() output ends -->


<!-- morphologytableformatting.py filloutmorphtabletemplate() output begins -->

<table class="verbanalysis">
<tbody>

<tr align="center">
	<td rowspan="1" colspan="7" class="dialectlabel">attic<br>
	</td>
</tr>
<tr align="center">
	<td rowspan="1" colspan="7" class="voicelabel">act<br>
	</td>
</tr>
<tr align="center">
	<td rowspan="1" colspan="7" class="moodlabel">opt<br>
	</td>
</tr><tr>
	<td class="tenselabel">&nbsp;</td>
	<td class="tensecell">Present<br></td>
	<td class="tensecell">Future<br></td>
	<td class="tensecell">Aorist<br></td>
	<td class="tensecell">Perfect<br></td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">sg 1st</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πέϲοιμι">πέϲοιμι</verbform> (<span class="counter">7</span>)</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">sg 2nd</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πέϲοιϲ">πέϲοιϲ</verbform> (<span class="counter">2</span>) / <verbform searchterm="πεϲοίηϲ">πεϲοίηϲ</verbform> (<span class="counter">1</span>)</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">sg 3rd</td>
	<td class="morphcell"><verbform searchterm="πίπτοι">πίπτοι</verbform> (<span class="counter">24</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πέϲοι">πέϲοι</verbform> (<span class="counter">100</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτώκοι">πεπτώκοι</verbform> (<span class="counter">2</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">dual 2nd</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">dual 3rd</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">pl 1st</td>
	<td class="morphcell"><verbform searchterm="πίπτοιμεν">πίπτοιμεν</verbform> (<span class="counter">1</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πέϲοιμεν">πέϲοιμεν</verbform> (<span class="counter">7</span>)</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">pl 2nd</td>
	<td class="morphcell"><verbform searchterm="πίπτοιτε">πίπτοιτε</verbform></td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">pl 3rd</td>
	<td class="morphcell"><verbform searchterm="πίπτοιεν">πίπτοιεν</verbform> (<span class="counter">13</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πέϲοιεν">πέϲοιεν</verbform> (<span class="counter">21</span>)</td>
	<td class="morphcell">---</td>
</tr>

</tbody>
</table>
<hr class="styled">

<!-- morphologytableformatting.py filloutmorphtabletemplate() output ends -->


<!-- morphologytableformatting.py filloutmorphtabletemplate() output begins -->

<table class="verbanalysis">
<tbody>

<tr align="center">
	<td rowspan="1" colspan="7" class="dialectlabel">attic<br>
	</td>
</tr>
<tr align="center">
	<td rowspan="1" colspan="7" class="voicelabel">act<br>
	</td>
</tr>
<tr align="center">
	<td rowspan="1" colspan="7" class="moodlabel">imperat<br>
	</td>
</tr><tr>
	<td class="tenselabel">&nbsp;</td>
	<td class="tensecell">Present<br></td>
	<td class="tensecell">Aorist<br></td>
	<td class="tensecell">Perfect<br></td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">sg 1st</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">sg 2nd</td>
	<td class="morphcell"><verbform searchterm="πῖπτε">πῖπτε</verbform> (<span class="counter">34</span>)</td>
	<td class="morphcell"><verbform searchterm="πέϲε">πέϲε</verbform> (<span class="counter">78</span>)</td>
	<td class="morphcell"><verbform searchterm="πέπτωκε">πέπτωκε</verbform> (<span class="counter">237</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">sg 3rd</td>
	<td class="morphcell"><verbform searchterm="πιπτέτω">πιπτέτω</verbform> (<span class="counter">83</span>)</td>
	<td class="morphcell"><verbform searchterm="πεϲέτω">πεϲέτω</verbform> (<span class="counter">12</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτωκέτω">πεπτωκέτω</verbform> (<span class="counter">1</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">dual 2nd</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">dual 3rd</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">pl 1st</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">pl 2nd</td>
	<td class="morphcell"><verbform searchterm="πίπτετε">πίπτετε</verbform> (<span class="counter">29</span>)</td>
	<td class="morphcell"><verbform searchterm="πέϲετε">πέϲετε</verbform> (<span class="counter">11</span>)</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">pl 3rd</td>
	<td class="morphcell"><verbform searchterm="πιπτέτωϲαν">πιπτέτωϲαν</verbform> (<span class="counter">14</span>) / <verbform searchterm="πιτνόντων">πιτνόντων</verbform> (<span class="counter">3</span>) / <verbform searchterm="πιπτόντων">πιπτόντων</verbform> (<span class="counter">275</span>)</td>
	<td class="morphcell"><verbform searchterm="πεϲόντων">πεϲόντων</verbform> (<span class="counter">385</span>)</td>
	<td class="morphcell">---</td>
</tr>

</tbody>
</table>
<hr class="styled">

<!-- morphologytableformatting.py filloutmorphtabletemplate() output ends -->


<!-- morphologytableformatting.py filloutmorphtabletemplate() output begins -->

<table class="verbanalysis">
<tbody>

<tr align="center">
	<td rowspan="1" colspan="7" class="dialectlabel">attic<br>
	</td>
</tr>
<tr align="center">
	<td rowspan="1" colspan="7" class="voicelabel">act<br>
	</td>
</tr>
<tr align="center">
	<td rowspan="1" colspan="7" class="moodlabel">part<br>
	</td>
</tr><tr>
	<td class="tenselabel">&nbsp;</td>
	<td class="tensecell">Present<br></td>
	<td class="tensecell">Future<br></td>
	<td class="tensecell">Aorist<br></td>
	<td class="tensecell">Perfect<br></td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc sg nom</td>
	<td class="morphcell"><verbform searchterm="πίπτων">πίπτων</verbform> (<span class="counter">181</span>) / <verbform searchterm="πίτνων">πίτνων</verbform> (<span class="counter">9</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲών">πεϲών</verbform> (<span class="counter">1131</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτεώϲ">πεπτεώϲ</verbform> (<span class="counter">1</span>) / <verbform searchterm="πεπτωκώϲ">πεπτωκώϲ</verbform> (<span class="counter">84</span>) / <verbform searchterm="πεπτηώϲ">πεπτηώϲ</verbform> (<span class="counter">15</span>) / <verbform searchterm="πεπτηκώϲ">πεπτηκώϲ</verbform> (<span class="counter">3</span>) / <verbform searchterm="πεπτώϲ">πεπτώϲ</verbform> (<span class="counter">1</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc sg gen</td>
	<td class="morphcell"><verbform searchterm="πίπτοντοϲ">πίπτοντοϲ</verbform> (<span class="counter">73</span>) / <verbform searchterm="πίτνοντοϲ">πίτνοντοϲ</verbform> (<span class="counter">1</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόντοϲ">πεϲόντοϲ</verbform> (<span class="counter">548</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτωκότοϲ">πεπτωκότοϲ</verbform> (<span class="counter">93</span>) / <verbform searchterm="πεπτηῶτοϲ">πεπτηῶτοϲ</verbform> (<span class="counter">2</span>) / <verbform searchterm="πεπτῶτοϲ">πεπτῶτοϲ</verbform> (<span class="counter">3</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc sg dat</td>
	<td class="morphcell"><verbform searchterm="πίπτοντι">πίπτοντι</verbform> (<span class="counter">21</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόντι">πεϲόντι</verbform> (<span class="counter">77</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτηκότι">πεπτηκότι</verbform> (<span class="counter">1</span>) / <verbform searchterm="πεπτωκότι">πεπτωκότι</verbform> (<span class="counter">39</span>) / <verbform searchterm="πεπτηῶτι">πεπτηῶτι</verbform> (<span class="counter">1</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc sg acc</td>
	<td class="morphcell"><verbform searchterm="πίπτοντα">πίπτοντα</verbform> (<span class="counter">176</span>) / <verbform searchterm="πίτνοντα">πίτνοντα</verbform> (<span class="counter">7</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόντα">πεϲόντα</verbform> (<span class="counter">460</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτῶτα">πεπτῶτα</verbform> (<span class="counter">3</span>) / <verbform searchterm="πεπτεῶτα">πεπτεῶτα</verbform> (<span class="counter">5</span>) / <verbform searchterm="πεπτωκότα">πεπτωκότα</verbform> (<span class="counter">188</span>) / <verbform searchterm="πεπτηῶτα">πεπτηῶτα</verbform> (<span class="counter">16</span>) / <verbform searchterm="πεπτηότα">πεπτηότα</verbform> (<span class="counter">5</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc sg voc</td>
	<td class="morphcell"><verbform searchterm="πίτνον">πίτνον</verbform> (<span class="counter">12</span>) / <verbform searchterm="πῖπτον">πῖπτον</verbform> (<span class="counter">85</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόν">πεϲόν</verbform> (<span class="counter">123</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτεώϲ">πεπτεώϲ</verbform> (<span class="counter">1</span>) / <verbform searchterm="πεπτωκώϲ">πεπτωκώϲ</verbform> (<span class="counter">84</span>) / <verbform searchterm="πεπτηώϲ">πεπτηώϲ</verbform> (<span class="counter">15</span>) / <verbform searchterm="πεπτηκώϲ">πεπτηκώϲ</verbform> (<span class="counter">3</span>) / <verbform searchterm="πεπτώϲ">πεπτώϲ</verbform> (<span class="counter">1</span>)</td>
</tr>


<tr><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td></tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem sg nom</td>
	<td class="morphcell"><verbform searchterm="πίτνουϲα">πίτνουϲα</verbform> (<span class="counter">1</span>) / <verbform searchterm="πίπτουϲα">πίπτουϲα</verbform> (<span class="counter">57</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲοῦϲα">πεϲοῦϲα</verbform> (<span class="counter">197</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτωκυῖα">πεπτωκυῖα</verbform> (<span class="counter">10</span>) / <verbform searchterm="πεπτηυῖα">πεπτηυῖα</verbform> (<span class="counter">2</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem sg gen</td>
	<td class="morphcell"><verbform searchterm="πιτνούϲηϲ">πιτνούϲηϲ</verbform> (<span class="counter">1</span>) / <verbform searchterm="πιπτούϲηϲ">πιπτούϲηϲ</verbform> (<span class="counter">58</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲούϲηϲ">πεϲούϲηϲ</verbform> (<span class="counter">110</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτηυίαϲ">πεπτηυίαϲ</verbform> (<span class="counter">2</span>) / <verbform searchterm="πεπτωκυίαϲ">πεπτωκυίαϲ</verbform> (<span class="counter">18</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem sg dat</td>
	<td class="morphcell"><verbform searchterm="πιπτούϲῃ">πιπτούϲῃ</verbform> (<span class="counter">7</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲούϲῃ">πεϲούϲῃ</verbform> (<span class="counter">15</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτωκυίᾳ">πεπτωκυίᾳ</verbform> (<span class="counter">3</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem sg acc</td>
	<td class="morphcell"><verbform searchterm="πίπτουϲαν">πίπτουϲαν</verbform> (<span class="counter">36</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲοῦϲαν">πεϲοῦϲαν</verbform> (<span class="counter">109</span>) / <verbform searchterm="πετοῦϲαν">πετοῦϲαν</verbform> (<span class="counter">1</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτωκυῖαν">πεπτωκυῖαν</verbform> (<span class="counter">64</span>) / <verbform searchterm="πεπτηυῖαν">πεπτηυῖαν</verbform> (<span class="counter">5</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem sg voc</td>
	<td class="morphcell"><verbform searchterm="πίτνουϲα">πίτνουϲα</verbform> (<span class="counter">1</span>) / <verbform searchterm="πίπτουϲα">πίπτουϲα</verbform> (<span class="counter">57</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲοῦϲα">πεϲοῦϲα</verbform> (<span class="counter">197</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτωκυῖα">πεπτωκυῖα</verbform> (<span class="counter">10</span>) / <verbform searchterm="πεπτηυῖα">πεπτηυῖα</verbform> (<span class="counter">2</span>)</td>
</tr>


<tr><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td></tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut sg nom</td>
	<td class="morphcell"><verbform searchterm="πίτνον">πίτνον</verbform> (<span class="counter">12</span>) / <verbform searchterm="πῖπτον">πῖπτον</verbform> (<span class="counter">85</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόν">πεϲόν</verbform> (<span class="counter">123</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτωκόϲ">πεπτωκόϲ</verbform> (<span class="counter">111</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut sg gen</td>
	<td class="morphcell"><verbform searchterm="πίπτοντοϲ">πίπτοντοϲ</verbform> (<span class="counter">73</span>) / <verbform searchterm="πίτνοντοϲ">πίτνοντοϲ</verbform> (<span class="counter">1</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόντοϲ">πεϲόντοϲ</verbform> (<span class="counter">548</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτωκότοϲ">πεπτωκότοϲ</verbform> (<span class="counter">93</span>) / <verbform searchterm="πεπτηῶτοϲ">πεπτηῶτοϲ</verbform> (<span class="counter">2</span>) / <verbform searchterm="πεπτῶτοϲ">πεπτῶτοϲ</verbform> (<span class="counter">3</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut sg dat</td>
	<td class="morphcell"><verbform searchterm="πίπτοντι">πίπτοντι</verbform> (<span class="counter">21</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόντι">πεϲόντι</verbform> (<span class="counter">77</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτηκότι">πεπτηκότι</verbform> (<span class="counter">1</span>) / <verbform searchterm="πεπτωκότι">πεπτωκότι</verbform> (<span class="counter">39</span>) / <verbform searchterm="πεπτηῶτι">πεπτηῶτι</verbform> (<span class="counter">1</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut sg acc</td>
	<td class="morphcell"><verbform searchterm="πίτνον">πίτνον</verbform> (<span class="counter">12</span>) / <verbform searchterm="πῖπτον">πῖπτον</verbform> (<span class="counter">85</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόν">πεϲόν</verbform> (<span class="counter">123</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτωκόϲ">πεπτωκόϲ</verbform> (<span class="counter">111</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut sg voc</td>
	<td class="morphcell"><verbform searchterm="πίτνον">πίτνον</verbform> (<span class="counter">12</span>) / <verbform searchterm="πῖπτον">πῖπτον</verbform> (<span class="counter">85</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόν">πεϲόν</verbform> (<span class="counter">123</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτωκόϲ">πεπτωκόϲ</verbform> (<span class="counter">111</span>)</td>
</tr>


<tr><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td></tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc dual nom</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόντε">πεϲόντε</verbform> (<span class="counter">5</span>)</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc dual gen</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc dual dat</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc dual acc</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόντε">πεϲόντε</verbform> (<span class="counter">5</span>)</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc dual voc</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόντε">πεϲόντε</verbform> (<span class="counter">5</span>)</td>
	<td class="morphcell">---</td>
</tr>


<tr><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td></tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem dual nom</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem dual gen</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem dual dat</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem dual acc</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem dual voc</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td></tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut dual nom</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόντε">πεϲόντε</verbform> (<span class="counter">5</span>)</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut dual gen</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut dual dat</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut dual acc</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόντε">πεϲόντε</verbform> (<span class="counter">5</span>)</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut dual voc</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόντε">πεϲόντε</verbform> (<span class="counter">5</span>)</td>
	<td class="morphcell">---</td>
</tr>


<tr><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td></tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc pl nom</td>
	<td class="morphcell"><verbform searchterm="πίτνοντεϲ">πίτνοντεϲ</verbform> (<span class="counter">1</span>) / <verbform searchterm="πίπτοντεϲ">πίπτοντεϲ</verbform> (<span class="counter">126</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόντεϲ">πεϲόντεϲ</verbform> (<span class="counter">281</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτωκότεϲ">πεπτωκότεϲ</verbform> (<span class="counter">71</span>) / <verbform searchterm="πεπτηῶτεϲ">πεπτηῶτεϲ</verbform> (<span class="counter">13</span>) / <verbform searchterm="πεπτηότεϲ">πεπτηότεϲ</verbform> (<span class="counter">5</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc pl gen</td>
	<td class="morphcell"><verbform searchterm="πιτνόντων">πιτνόντων</verbform> (<span class="counter">3</span>) / <verbform searchterm="πιπτόντων">πιπτόντων</verbform> (<span class="counter">275</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόντων">πεϲόντων</verbform> (<span class="counter">385</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτωκότων">πεπτωκότων</verbform> (<span class="counter">160</span>) / <verbform searchterm="πεπτηκότων">πεπτηκότων</verbform> (<span class="counter">1</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc pl dat</td>
	<td class="morphcell"><verbform searchterm="πίπτουϲιν">πίπτουϲιν</verbform> (<span class="counter">346</span>) / <verbform searchterm="πίπτουϲι">πίπτουϲι</verbform> (<span class="counter">262</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲοῦϲι">πεϲοῦϲι</verbform> (<span class="counter">59</span>) / <verbform searchterm="πεϲοῦϲιν">πεϲοῦϲιν</verbform> (<span class="counter">50</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτωκόϲι">πεπτωκόϲι</verbform> (<span class="counter">14</span>) / <verbform searchterm="πεπτωκόϲιν">πεπτωκόϲιν</verbform> (<span class="counter">15</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc pl acc</td>
	<td class="morphcell"><verbform searchterm="πίτνονταϲ">πίτνονταϲ</verbform> (<span class="counter">1</span>) / <verbform searchterm="πίπτονταϲ">πίπτονταϲ</verbform> (<span class="counter">115</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόνταϲ">πεϲόνταϲ</verbform> (<span class="counter">210</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτεῶταϲ">πεπτεῶταϲ</verbform> (<span class="counter">1</span>) / <verbform searchterm="πεπτωκόταϲ">πεπτωκόταϲ</verbform> (<span class="counter">85</span>) / <verbform searchterm="πεπτηόταϲ">πεπτηόταϲ</verbform> (<span class="counter">2</span>) / <verbform searchterm="πεπτηῶταϲ">πεπτηῶταϲ</verbform> (<span class="counter">8</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc pl voc</td>
	<td class="morphcell"><verbform searchterm="πίτνοντεϲ">πίτνοντεϲ</verbform> (<span class="counter">1</span>) / <verbform searchterm="πίπτοντεϲ">πίπτοντεϲ</verbform> (<span class="counter">126</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόντεϲ">πεϲόντεϲ</verbform> (<span class="counter">281</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτωκότεϲ">πεπτωκότεϲ</verbform> (<span class="counter">71</span>) / <verbform searchterm="πεπτηῶτεϲ">πεπτηῶτεϲ</verbform> (<span class="counter">13</span>) / <verbform searchterm="πεπτηότεϲ">πεπτηότεϲ</verbform> (<span class="counter">5</span>)</td>
</tr>


<tr><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td></tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem pl nom</td>
	<td class="morphcell"><verbform searchterm="πίπτουϲαι">πίπτουϲαι</verbform> (<span class="counter">27</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲοῦϲαι">πεϲοῦϲαι</verbform> (<span class="counter">22</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτηυῖαι">πεπτηυῖαι</verbform> (<span class="counter">6</span>) / <verbform searchterm="πεπτωκυῖαι">πεπτωκυῖαι</verbform> (<span class="counter">6</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem pl gen</td>
	<td class="morphcell"><verbform searchterm="πιπτουϲῶν">πιπτουϲῶν</verbform> (<span class="counter">26</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲουϲῶν">πεϲουϲῶν</verbform> (<span class="counter">9</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτωκυιῶν">πεπτωκυιῶν</verbform> (<span class="counter">1</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem pl dat</td>
	<td class="morphcell"><verbform searchterm="πιπτούϲαιϲ">πιπτούϲαιϲ</verbform> (<span class="counter">4</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲούϲαιϲ">πεϲούϲαιϲ</verbform> (<span class="counter">13</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτηυίαιϲ">πεπτηυίαιϲ</verbform> (<span class="counter">2</span>) / <verbform searchterm="πεπτηυίαιϲ">πεπτηυίαιϲ</verbform> (<span class="counter">2</span>) / <verbform searchterm="πεπτωκυίαιϲ">πεπτωκυίαιϲ</verbform> (<span class="counter">7</span>) / <verbform searchterm="πεπτωκυίαιϲ">πεπτωκυίαιϲ</verbform> (<span class="counter">7</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem pl acc</td>
	<td class="morphcell"><verbform searchterm="πιπτούϲαϲ">πιπτούϲαϲ</verbform> (<span class="counter">22</span>) / <verbform searchterm="πιτνούϲαϲ">πιτνούϲαϲ</verbform> (<span class="counter">1</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲούϲαϲ">πεϲούϲαϲ</verbform> (<span class="counter">18</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτηυίαϲ">πεπτηυίαϲ</verbform> (<span class="counter">2</span>) / <verbform searchterm="πεπτωκυίαϲ">πεπτωκυίαϲ</verbform> (<span class="counter">18</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem pl voc</td>
	<td class="morphcell"><verbform searchterm="πίπτουϲαι">πίπτουϲαι</verbform> (<span class="counter">27</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲοῦϲαι">πεϲοῦϲαι</verbform> (<span class="counter">22</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτηυῖαι">πεπτηυῖαι</verbform> (<span class="counter">6</span>) / <verbform searchterm="πεπτωκυῖαι">πεπτωκυῖαι</verbform> (<span class="counter">6</span>)</td>
</tr>


<tr><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td></tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut pl nom</td>
	<td class="morphcell"><verbform searchterm="πίπτοντα">πίπτοντα</verbform> (<span class="counter">176</span>) / <verbform searchterm="πίτνοντα">πίτνοντα</verbform> (<span class="counter">7</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόντα">πεϲόντα</verbform> (<span class="counter">460</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτῶτα">πεπτῶτα</verbform> (<span class="counter">3</span>) / <verbform searchterm="πεπτεῶτα">πεπτεῶτα</verbform> (<span class="counter">5</span>) / <verbform searchterm="πεπτωκότα">πεπτωκότα</verbform> (<span class="counter">188</span>) / <verbform searchterm="πεπτηῶτα">πεπτηῶτα</verbform> (<span class="counter">16</span>) / <verbform searchterm="πεπτηότα">πεπτηότα</verbform> (<span class="counter">5</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut pl gen</td>
	<td class="morphcell"><verbform searchterm="πιτνόντων">πιτνόντων</verbform> (<span class="counter">3</span>) / <verbform searchterm="πιπτόντων">πιπτόντων</verbform> (<span class="counter">275</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόντων">πεϲόντων</verbform> (<span class="counter">385</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτωκότων">πεπτωκότων</verbform> (<span class="counter">160</span>) / <verbform searchterm="πεπτηκότων">πεπτηκότων</verbform> (<span class="counter">1</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut pl dat</td>
	<td class="morphcell"><verbform searchterm="πίπτουϲιν">πίπτουϲιν</verbform> (<span class="counter">346</span>) / <verbform searchterm="πίπτουϲι">πίπτουϲι</verbform> (<span class="counter">262</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲοῦϲι">πεϲοῦϲι</verbform> (<span class="counter">59</span>) / <verbform searchterm="πεϲοῦϲιν">πεϲοῦϲιν</verbform> (<span class="counter">50</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτωκόϲι">πεπτωκόϲι</verbform> (<span class="counter">14</span>) / <verbform searchterm="πεπτωκόϲιν">πεπτωκόϲιν</verbform> (<span class="counter">15</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut pl acc</td>
	<td class="morphcell"><verbform searchterm="πίπτοντα">πίπτοντα</verbform> (<span class="counter">176</span>) / <verbform searchterm="πίτνοντα">πίτνοντα</verbform> (<span class="counter">7</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόντα">πεϲόντα</verbform> (<span class="counter">460</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτῶτα">πεπτῶτα</verbform> (<span class="counter">3</span>) / <verbform searchterm="πεπτεῶτα">πεπτεῶτα</verbform> (<span class="counter">5</span>) / <verbform searchterm="πεπτωκότα">πεπτωκότα</verbform> (<span class="counter">188</span>) / <verbform searchterm="πεπτηῶτα">πεπτηῶτα</verbform> (<span class="counter">16</span>) / <verbform searchterm="πεπτηότα">πεπτηότα</verbform> (<span class="counter">5</span>)</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut pl voc</td>
	<td class="morphcell"><verbform searchterm="πίπτοντα">πίπτοντα</verbform> (<span class="counter">176</span>) / <verbform searchterm="πίτνοντα">πίτνοντα</verbform> (<span class="counter">7</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲόντα">πεϲόντα</verbform> (<span class="counter">460</span>)</td>
	<td class="morphcell"><verbform searchterm="πεπτῶτα">πεπτῶτα</verbform> (<span class="counter">3</span>) / <verbform searchterm="πεπτεῶτα">πεπτεῶτα</verbform> (<span class="counter">5</span>) / <verbform searchterm="πεπτωκότα">πεπτωκότα</verbform> (<span class="counter">188</span>) / <verbform searchterm="πεπτηῶτα">πεπτηῶτα</verbform> (<span class="counter">16</span>) / <verbform searchterm="πεπτηότα">πεπτηότα</verbform> (<span class="counter">5</span>)</td>
</tr>


<tr><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td></tr>

</tbody>
</table>
<hr class="styled">

<!-- morphologytableformatting.py filloutmorphtabletemplate() output ends -->


<!-- morphologytableformatting.py filloutmorphtabletemplate() output begins -->

<table class="verbanalysis">
<tbody>

<tr align="center">
	<td rowspan="1" colspan="7" class="dialectlabel">attic<br>
	</td>
</tr>
<tr align="center">
	<td rowspan="1" colspan="7" class="voicelabel">act<br>
	</td>
</tr>
<tr align="center">
	<td rowspan="1" colspan="7" class="moodlabel">inf<br>
	</td>
</tr><tr>
	<td class="tenselabel">&nbsp;</td>
	<td class="tensecell">Present<br></td>
	<td class="tensecell">Future<br></td>
	<td class="tensecell">Aorist<br></td>
	<td class="tensecell">Perfect<br></td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">infinitive</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεπτωκέναι">πεπτωκέναι</verbform> (<span class="counter">107</span>)</td>
</tr>

</tbody>
</table>
<hr class="styled">

<!-- morphologytableformatting.py filloutmorphtabletemplate() output ends -->


<!-- morphologytableformatting.py filloutmorphtabletemplate() output begins -->

<table class="verbanalysis">
<tbody>

<tr align="center">
	<td rowspan="1" colspan="7" class="dialectlabel">attic<br>
	</td>
</tr>
<tr align="center">
	<td rowspan="1" colspan="7" class="voicelabel">mid<br>
	</td>
</tr>
<tr align="center">
	<td rowspan="1" colspan="7" class="moodlabel">ind<br>
	</td>
</tr><tr>
	<td class="tenselabel">&nbsp;</td>
	<td class="tensecell">Present<br></td>
	<td class="tensecell">Imperfect<br></td>
	<td class="tensecell">Future<br></td>
	<td class="tensecell">Aorist<br></td>
	<td class="tensecell">Perfect<br></td>
	<td class="tensecell">Pluperfect<br></td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">sg 1st</td>
	<td class="morphcell"><verbform searchterm="πίπτομαι">πίπτομαι</verbform> (<span class="counter">1</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲοῦμαι">πεϲοῦμαι</verbform> (<span class="counter">7</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">sg 2nd</td>
	<td class="morphcell"><verbform searchterm="πίτνει">πίτνει</verbform> (<span class="counter">21</span>) / <verbform searchterm="πίπτει">πίπτει</verbform> (<span class="counter">1125</span>) / <verbform searchterm="πίπτῃ">πίπτῃ</verbform> (<span class="counter">57</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲῆι">πεϲῆι</verbform> (<span class="counter">4</span>) / <verbform searchterm="πεϲῇ">πεϲῇ</verbform> (<span class="counter">17</span>) / <verbform searchterm="πεϲεῖ">πεϲεῖ</verbform> (<span class="counter">1</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">sg 3rd</td>
	<td class="morphcell"><verbform searchterm="πίπτεται">πίπτεται</verbform> (<span class="counter">1</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲεῖται">πεϲεῖται</verbform> (<span class="counter">745</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">dual 2nd</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">dual 3rd</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">pl 1st</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲούμεθα">πεϲούμεθα</verbform> (<span class="counter">10</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">pl 2nd</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲεῖϲθε">πεϲεῖϲθε</verbform> (<span class="counter">22</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">pl 3rd</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲοῦνται">πεϲοῦνται</verbform> (<span class="counter">399</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>

</tbody>
</table>
<hr class="styled">

<!-- morphologytableformatting.py filloutmorphtabletemplate() output ends -->


<!-- morphologytableformatting.py filloutmorphtabletemplate() output begins -->

<table class="verbanalysis">
<tbody>

<tr align="center">
	<td rowspan="1" colspan="7" class="dialectlabel">attic<br>
	</td>
</tr>
<tr align="center">
	<td rowspan="1" colspan="7" class="voicelabel">mid<br>
	</td>
</tr>
<tr align="center">
	<td rowspan="1" colspan="7" class="moodlabel">opt<br>
	</td>
</tr><tr>
	<td class="tenselabel">&nbsp;</td>
	<td class="tensecell">Present<br></td>
	<td class="tensecell">Future<br></td>
	<td class="tensecell">Aorist<br></td>
	<td class="tensecell">Perfect<br></td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">sg 1st</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">sg 2nd</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πέϲοιο">πέϲοιο</verbform> (<span class="counter">1</span>)</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">sg 3rd</td>
	<td class="morphcell"><verbform searchterm="πίπτοιτο">πίπτοιτο</verbform> (<span class="counter">1</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">dual 2nd</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">dual 3rd</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">pl 1st</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">pl 2nd</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">pl 3rd</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>

</tbody>
</table>
<hr class="styled">

<!-- morphologytableformatting.py filloutmorphtabletemplate() output ends -->


<!-- morphologytableformatting.py filloutmorphtabletemplate() output begins -->

<table class="verbanalysis">
<tbody>

<tr align="center">
	<td rowspan="1" colspan="7" class="dialectlabel">attic<br>
	</td>
</tr>
<tr align="center">
	<td rowspan="1" colspan="7" class="voicelabel">mid<br>
	</td>
</tr>
<tr align="center">
	<td rowspan="1" colspan="7" class="moodlabel">part<br>
	</td>
</tr><tr>
	<td class="tenselabel">&nbsp;</td>
	<td class="tensecell">Present<br></td>
	<td class="tensecell">Future<br></td>
	<td class="tensecell">Aorist<br></td>
	<td class="tensecell">Perfect<br></td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc sg nom</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲούμενοϲ">πεϲούμενοϲ</verbform> (<span class="counter">7</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc sg gen</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲουμένου">πεϲουμένου</verbform> (<span class="counter">4</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc sg dat</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc sg acc</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲούμενον">πεϲούμενον</verbform> (<span class="counter">10</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc sg voc</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td></tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem sg nom</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem sg gen</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲουμένηϲ">πεϲουμένηϲ</verbform> (<span class="counter">3</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem sg dat</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem sg acc</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem sg voc</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td></tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut sg nom</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲούμενον">πεϲούμενον</verbform> (<span class="counter">10</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut sg gen</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲουμένου">πεϲουμένου</verbform> (<span class="counter">4</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut sg dat</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut sg acc</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲούμενον">πεϲούμενον</verbform> (<span class="counter">10</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut sg voc</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲούμενον">πεϲούμενον</verbform> (<span class="counter">10</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td></tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc dual nom</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc dual gen</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc dual dat</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc dual acc</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc dual voc</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td></tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem dual nom</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem dual gen</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem dual dat</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem dual acc</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem dual voc</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td></tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut dual nom</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut dual gen</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut dual dat</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut dual acc</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut dual voc</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td></tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc pl nom</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲούμενοι">πεϲούμενοι</verbform> (<span class="counter">1</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc pl gen</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲουμένων">πεϲουμένων</verbform> (<span class="counter">6</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc pl dat</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲουμένοιϲ">πεϲουμένοιϲ</verbform> (<span class="counter">1</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc pl acc</td>
	<td class="morphcell"><verbform searchterm="πιπτομένουϲ">πιπτομένουϲ</verbform> (<span class="counter">1</span>)</td>
	<td class="morphcell"><verbform searchterm="πεϲουμένουϲ">πεϲουμένουϲ</verbform> (<span class="counter">1</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">masc pl voc</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲούμενοι">πεϲούμενοι</verbform> (<span class="counter">1</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td></tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem pl nom</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem pl gen</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲουμένων">πεϲουμένων</verbform> (<span class="counter">6</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem pl dat</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲουμέναιϲ">πεϲουμέναιϲ</verbform> (<span class="counter">1</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem pl acc</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">fem pl voc</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td></tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut pl nom</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲούμενα">πεϲούμενα</verbform> (<span class="counter">5</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut pl gen</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲουμένων">πεϲουμένων</verbform> (<span class="counter">6</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut pl dat</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲουμένοιϲ">πεϲουμένοιϲ</verbform> (<span class="counter">1</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut pl acc</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲούμενα">πεϲούμενα</verbform> (<span class="counter">5</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">neut pl voc</td>
	<td class="morphcell">---</td>
	<td class="morphcell"><verbform searchterm="πεϲούμενα">πεϲούμενα</verbform> (<span class="counter">5</span>)</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>


<tr><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td></tr>

</tbody>
</table>
<hr class="styled">

<!-- morphologytableformatting.py filloutmorphtabletemplate() output ends -->


<!-- morphologytableformatting.py filloutmorphtabletemplate() output begins -->

<table class="verbanalysis">
<tbody>

<tr align="center">
	<td rowspan="1" colspan="7" class="dialectlabel">attic<br>
	</td>
</tr>
<tr align="center">
	<td rowspan="1" colspan="7" class="voicelabel">mid<br>
	</td>
</tr>
<tr align="center">
	<td rowspan="1" colspan="7" class="moodlabel">inf<br>
	</td>
</tr><tr>
	<td class="tenselabel">&nbsp;</td>
	<td class="tensecell">Present<br></td>
	<td class="tensecell">Future<br></td>
	<td class="tensecell">Aorist<br></td>
	<td class="tensecell">Perfect<br></td>
</tr>


<tr class="morphrow">
	<td class="morphlabelcell">infinitive</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
	<td class="morphcell">---</td>
</tr>

</tbody>
</table>
<hr class="styled">

<!-- morphologytableformatting.py filloutmorphtabletemplate() output ends -->
*/
