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
	"strconv"
	"strings"
)

const (
	JOINER = "_"
	BLANK  = " --- "
)

var (
	GKCASES       = []string{"nom", "gen", "dat", "acc", "voc"}
	GKNUMB        = []string{"sg", "dual", "pl"}
	GKMOODS       = []string{"ind", "subj", "opt", "imperat", "inf", "part"}
	GKVOICE       = []string{"act", "mid", "pass"}
	GKTENSES      = []string{"pres", "imperf", "fut", "aor", "perf", "plup", "futperf"} // order matters
	GKINTTENSEMAP = map[int]string{1: "Present", 2: "Imperfect", 3: "Future", 4: "Aorist", 5: "Perfect", 6: "Pluperfect", 7: "Future Perfect"}
	GKTENSEMAP    = map[string]int{"pres": 1, "imperf": 2, "fut": 3, "aor": 4, "perf": 5, "plup": 6, "futperf": 7}
	GKVERBS       = getgkvbmap()
	GKDIALECT     = []string{"attic"} // TODO: INCOMPLETE
	LTCASES       = []string{"nom", "gen", "dat", "acc", "abl", "voc"}
	LTNUMB        = []string{"sg", "pl"}
	LTMOODS       = []string{"ind", "subj", "imperat", "inf", "part", "gerundive", "supine"}
	LTVOICE       = []string{"act", "pass"}
	LTTENSES      = []string{"pres", "imperf", "fut", "perf", "plup", "futperf"} // order needs to match LTINTTENSEMAP
	LTINTTENSEMAP = map[int]string{1: "Present", 2: "Imperfect", 3: "Future", 5: "Perfect", 6: "Pluperfect", 7: "Future Perfect"}
	LTTENSEMAP    = map[string]int{"pres": 1, "imperf": 2, "fut": 3, "perf": 5, "plup": 6, "futperf": 7}
	LTVERBS       = getltvbmap()
	GENDERS       = []string{"masc", "fem", "neut"}
	PERSONS       = []string{"1st", "2nd", "3rd"}
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
		thehit.Observed = strings.ToLower(thehit.Observed)
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
	// but they will in fact be stored less the apostrophe...
	// this leads to a clash between the wordcounts_κ data and the x_morphology data: todo - address this

	var esc []string
	for _, w := range ww {
		// esc = append(esc, strings.Replace(w, "'", "''", -1))
		esc = append(esc, strings.Replace(w, "'", "", -1))
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
		if []rune(l)[0] == 0 {
			continue
		}

		rnd := strings.Replace(uuid.New().String(), "-", "", -1)
		_, e := dbpool.Exec(context.Background(), fmt.Sprintf(tt, rnd, arr))
		chke(e)

		q := fmt.Sprintf(qt, l, rnd, l)
		rr, e := dbpool.Query(context.Background(), q)
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
				key := strings.Replace(v, " ", JOINER, -1)
				if _, ok := pdm[key]; !ok {
					pdm[key] = k
				} else {
					pdm[key] = pdm[key] + " / " + k
				}
			} else {
				// need to decompress "nom/voc/acc" into three entries, etc
				// todo: breaks on "masc/fem nom/voc sg"
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
					key = strings.Replace(key, " ", JOINER, -1)
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
	newpdm := make(map[string]string)
	if lg == "greek" {
		for k, v := range pdm {
			if strings.Contains(k, "(") {
				k = strings.Replace(k, ")", "", 1)
				parts := strings.Split(k, "(")
				diall := strings.Split(parts[1], JOINER)
				for _, d := range diall {
					newkey := parts[0] + JOINER + d
					newkey = strings.Replace(newkey, JOINER+JOINER, JOINER, 1)
					newpdm[newkey] = v
				}
			} else {
				if !strings.Contains(k, "attic") {
					newkey := k + JOINER + "attic"
					newpdm[newkey] = v
				} else {
					newpdm[k] = v
				}
			}
		}
	} else {
		// add the "blank" dialect to latin
		for k, v := range pdm {
			newpdm[k+JOINER] = v
		}
	}
	pdm = newpdm

	// [e3] TODO: mp needs a splitter: middle and passive

	// [e4] get counts for each word
	pdcm := make(map[string]map[string]int64)
	for k, v := range pdm {
		wds := strings.Split(v, " / ")
		mm := make(map[string]int64)
		for _, w := range wds {
			mm[w] = wcc[w].Total
		}
		pdcm[k] = mm
	}

	ctm := `<verbform searchterm="%s">%s</verbform> (<span class="counter">%d</span>)`
	pdxm := make(map[string]string)
	for kk, pd := range pdcm {
		var vv []string
		for k, v := range pd {
			vv = append(vv, fmt.Sprintf(ctm, k, k, v))
		}
		pdxm[kk] = strings.Join(vv, " / ")

		//if strings.Contains(kk, "part") {
		//	fmt.Printf("%s\t%s\n", kk, pdxm[kk])
		//}
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

		//[HGS] aor_part_mid_fem_nom_sg_attic
		//[HGS] perf_part_mp_fem_voc_pl_attic
		//[HGS] pres_part_act_fem_voc_pl_ionic
		//[HGS] pres_part_act_fem_nom_sg_attic
	}

	//zz := stringmapkeysintoslice(pdxm)
	//for _, z := range zz {
	//	msg(z, 1)
	//}

	isverb := func() bool {
		kk := stringmapkeysintoslice(pdxm)
		return arraystringseeker(GKTENSES, kk)
	}()

	var jb JSB

	// [g] build the table

	if isverb {
		jb.HTML = generateverbtable(lg, pdxm)
	} else {
		// jb.HTML = "[RtMorphchart() is a work in progress...]<br>" + strings.Join(oo, "<br>")
		jb.HTML = generatedeclinedtable(lg, pdxm)
	}

	jb.JS = MORPHJS

	return c.JSONPretty(http.StatusOK, jb, JSONINDENT)
}

func generateverbtable(lang string, words map[string]string) string {
	// first voice
	// then mood
	// then tense as columns and number_and_person as rows

	const (
		DIALTR = `
		<tr align="center">
			<td rowspan="1" colspan="%d" class="dialectlabel">%s<br>
			</td>
		</tr>`

		VOICETR = `
		<tr align="center">
			<td rowspan="1" colspan="%d" class="voicelabel">%s<br>
			</td>
		</tr>`

		MOODTR = `
		<tr align="center">
			<td rowspan="1" colspan="%d" class="moodlabel">%s<br>
			</td>
		</tr>`
	)

	vm := make(map[string]map[string]map[int]bool)
	tm := make(map[string]int)
	var dialect []string
	var voices []string
	var moods []string
	var numbers []string
	var tenses []string
	var cases []string
	gend := GENDERS

	switch lang {
	case "greek":
		vm = getgkvbmap()
		tm = GKTENSEMAP
		dialect = GKDIALECT
		voices = GKVOICE
		moods = GKMOODS
		numbers = GKNUMB
		tenses = GKTENSES
		cases = GKCASES
	case "latin":
		vm = getltvbmap()
		tm = LTTENSEMAP
		dialect = []string{""}
		voices = LTVOICE
		moods = LTMOODS
		numbers = LTNUMB
		tenses = LTTENSES
		cases = LTCASES
	}

	kk := stringmapkeysintoslice(words)
	needgend := func() []string {
		var need []string
		for _, g := range gend {
			if sliceseeker(g, kk) {
				need = append(need, g)
			}
		}
		return need
	}()

	//
	// HEAD ROW PRODUCERS
	//

	maketnshdr := func(v string, m string) string {
		hdr := `
		<tr>
			<td class="tenselabel">&nbsp;</td>
			`
		for i := 1; i < 8; i++ {
			// have to do it in numerical order...
			if vm[v][m][i] {
				hdr += fmt.Sprintf("<td class=\"tensecell\">%s<br></td>\n\t", GKINTTENSEMAP[i])
			}
		}
		hdr += `</tr>`
		return hdr
	}

	makepcphdr := func() string {
		hdr := `
		<tr>
			<td class="tenselabel">&nbsp;</td>
			`
		for _, g := range needgend {
			hdr += fmt.Sprintf("<td class=\"tensecell\">%s<br></td>\n\t", g)
		}
		hdr += `</tr>`
		return hdr
	}()

	//
	// TRR PRODUCERS
	//

	makevftrr := func(d string, v string, m string) ([]string, bool) {
		// for vanilla verbs only; this will NOT do participles, supines, gerundives, infinitives

		// <tr class="morphrow">
		//	<td class="morphlabelcell">sg 1st</td>
		//	<td class="morphcell"><verbform searchterm="πίτνω">πίτνω</verbform> (<span class="counter">15</span>) / <verbform searchterm="πίπτω">πίπτω</verbform> (<span class="counter">117</span>)</td>
		//	<td class="morphcell"><verbform searchterm="ἔπιπτον">ἔπιπτον</verbform> (<span class="counter">259</span>) / <verbform searchterm="ἔπιτνον">ἔπιτνον</verbform> (<span class="counter">3</span>)</td>
		//	<td class="morphcell">---</td>
		//	<td class="morphcell"><verbform searchterm="ἔπεϲον">ἔπεϲον</verbform> (<span class="counter">686</span>)</td>
		//	<td class="morphcell"><verbform searchterm="πέπτηκα">πέπτηκα</verbform> (<span class="counter">14</span>) / <verbform searchterm="πέπτωκα">πέπτωκα</verbform> (<span class="counter">67</span>)</td>
		//	<td class="morphcell"><verbform searchterm="ἐπεπτώκειν">ἐπεπτώκειν</verbform> (<span class="counter">1</span>)</td>
		//</tr>
		blankcount := 0
		cellcount := 0

		var trr []string
		for _, n := range numbers {
			for _, p := range PERSONS {
				// np := fmt.Sprintf("%s %s", n, p)
				trr = append(trr, `<tr class="morphrow">`)
				trr = append(trr, fmt.Sprintf(`<td class="morphlabelcell">%s %s</td>`, n, p))
				var tdd []string
				for _, t := range tenses {
					// not ever combination should be generated
					thevm := vm[v][m]
					if !thevm[tm[t]] {
						continue
					}
					k := fmt.Sprintf("%s_%s_%s_%s_%s_%s", t, m, v, p, n, d)
					if _, ok := words[k]; ok {
						tdd = append(tdd, words[k])
					} else {
						tdd = append(tdd, BLANK)
						blankcount += 1
					}
					cellcount += 1
				}
				for _, td := range tdd {
					trr = append(trr, fmt.Sprintf(`<td class="morphcell">%s</td>`, td))
				}
				trr = append(trr, `</tr>`)
			}
		}
		isblank := false
		if cellcount == blankcount {
			isblank = true
		}
		return trr, isblank
	}

	makepcpltrr := func(d string, m string, v string) ([]string, bool) {
		// problem: the header row has been pre-set to "tenses" not genders

		// LATIN PROBLEM
		// sent: pres_part_neut_acc_sg_
		// want: pres_part_act_neut_acc_sg_

		//[HGS] aor_part_mid_fem_nom_sg_attic
		//[HGS] perf_part_mp_fem_voc_pl_attic
		var trr []string
		// need to loop the tenses...
		blankcount := 0
		cellcount := 0
		for _, t := range tenses {
			// not ever combination should be generated
			thevm := vm[v][m]
			if !thevm[tm[t]] {
				continue
			}
			tl := `<tr align="center"><td rowspan="1" colspan="%d" class="morphrow emph">%s<br></td></tr>`
			trr = append(trr, fmt.Sprintf(tl, len(numbers)+2, t))
			for _, n := range numbers {
				for _, c := range cases {
					trr = append(trr, `<tr class="morphrow">`)
					trr = append(trr, fmt.Sprintf(`<td class="morphlabelcell">%s %s</td>`, n, c))
					var tdd []string
					for _, g := range needgend {
						// not every combination should be generated
						k := fmt.Sprintf("%s_%s_%s_%s_%s_%s_%s", t, m, v, g, c, n, d)
						// fix the irregular original data
						if lang == "latin" && t == "pres" {
							k = fmt.Sprintf("%s_%s_%s_%s_%s_%s", t, m, g, c, n, d)
						}
						if _, ok := words[k]; ok {
							tdd = append(tdd, words[k])
						} else {
							tdd = append(tdd, BLANK)
							blankcount += 1
						}
						cellcount += 1
					}
					for _, td := range tdd {
						trr = append(trr, fmt.Sprintf(`<td class="morphcell">%s</td>`, td))
					}
					trr = append(trr, `</tr>`)
				}
			}
		}
		isblank := false
		if cellcount == blankcount {
			isblank = true
		}
		return trr, isblank
	}

	makegertrr := func(d string, m string, v string) ([]string, bool) {
		// problem: the header row has been pre-set to "tenses" not genders

		// [HGS] gerundive_neut_abl_pl_

		// can also do supines...
		// [HGS] supine_neut_dat_sg_

		// todo: some sort of key collision will leave "laudando" in the wrong boxes

		var trr []string

		if v == "act" {
			return trr, true
		}

		nn := numbers
		cc := cases
		if m == "supine" {
			nn = []string{"sg"}
			cc = []string{"dat", "acc", "abl"}
		}

		tl := `<tr align="center"><td rowspan="1" colspan="%d" class="morphrow emph center">%s<br></td></tr>`
		trr = append(trr, fmt.Sprintf(tl, len(numbers)+1, ""))
		blankcount := 0
		cellcount := 0
		for _, n := range nn {
			for _, c := range cc {
				trr = append(trr, `<tr class="morphrow">`)
				trr = append(trr, fmt.Sprintf(`<td class="morphlabelcell">%s %s</td>`, n, c))
				var tdd []string
				for _, g := range needgend {
					// not every combination should be generated
					// fem_acc_dual_doric
					k := fmt.Sprintf("%s_%s_%s_%s_%s", m, g, c, n, d)
					if _, ok := words[k]; ok {
						tdd = append(tdd, words[k])
					} else {
						tdd = append(tdd, BLANK)
						blankcount += 1
					}
					cellcount += 1
				}
				for _, td := range tdd {
					trr = append(trr, fmt.Sprintf(`<td class="morphcell">%s</td>`, td))
				}
				trr = append(trr, `</tr>`)
			}
		}
		isblank := false
		if cellcount == blankcount {
			isblank = true
		}
		return trr, isblank
	}

	makeinftrr := func(d string, m string, v string) ([]string, bool) {
		// 	<tr align="center">
		//		<td rowspan="1" colspan="7" class="moodlabel">inf<br>
		//		</td>
		//	</tr><tr>
		//		<td class="tenselabel">&nbsp;</td>
		//		<td class="tensecell">Present<br></td>
		//		<td class="tensecell">Future<br></td>
		//		<td class="tensecell">Aorist<br></td>
		//		<td class="tensecell">Perfect<br></td>
		//	</tr>
		// 	<tr class="morphrow">
		//		<td class="morphlabelcell">infinitive</td>
		//		<td class="morphcell">---</td>
		//		<td class="morphcell">---</td>
		//		<td class="morphcell"><verbform searchterm="θρέψαι">θρέψαι</verbform> (<span class="counter">284</span>)</td>
		//		<td class="morphcell"><verbform searchterm="τετροφέναι">τετροφέναι</verbform> (<span class="counter">2</span>)</td>
		//	</tr>
		//
		var trr []string
		trr = append(trr, `<td class="tenselabel">&nbsp;</td>`)
		// need to loop the tenses...
		blankcount := 0
		cellcount := 0
		var tdd []string
		for _, t := range tenses {
			// not ever combination should be generated
			thevm := vm[v][m]
			if !thevm[tm[t]] {
				continue
			}
			//[HGS] fut_inf_mid_attic
			//[HGS] perf_inf_act_attic
			k := fmt.Sprintf("%s_%s_%s_%s", t, m, v, d)
			if _, ok := words[k]; ok {
				tdd = append(tdd, words[k])
			} else {
				tdd = append(tdd, BLANK)
				blankcount += 1
			}
			cellcount += 1
		}
		for _, td := range tdd {
			trr = append(trr, fmt.Sprintf(`<td class="morphcell">%s</td>`, td))
		}
		isblank := false
		if cellcount == blankcount {
			isblank = true
		}
		return trr, isblank
	}

	counttns := func(v string, m string) int {
		c := 0
		for _, t := range vm[v][m] {
			if t {
				c += 1
			}
		}
		return c
	}

	//
	// THE MAIN TABLE GENERATOR
	//

	var html []string

	for _, d := range dialect {
		// each dialect is a major section
		// but latin has only one dialect
		for _, v := range voices {
			// each voice is a section
			for _, m := range moods {
				// each mood is a table
				// not every item needs generating
				isblank := false
				// the top
				ct := counttns(v, m)
				html = append(html, `<table class="verbanalysis">`)
				html = append(html, fmt.Sprintf(DIALTR, ct, d))
				html = append(html, fmt.Sprintf(VOICETR, ct, v))
				html = append(html, fmt.Sprintf(MOODTR, ct, m))

				var trrhtml []string
				switch m {
				case "part":
					trrhtml, isblank = makepcpltrr(d, m, v)
				case "inf":
					trrhtml, isblank = makeinftrr(d, m, v)
				case "gerundive":
					trrhtml, isblank = makegertrr(d, m, v)
				case "supine":
					// exact same issues as gerundives
					trrhtml, isblank = makegertrr(d, m, v)
				default:
					trrhtml, isblank = makevftrr(d, v, m)
				}

				if isblank {
					trrhtml = []string{"<tr><td>[nothing found]</td></tr>"}
				} else {
					switch m {
					case "part":
						html = append(html, makepcphdr)
					case "inf":
						html = append(html, maketnshdr(v, m))
					case "gerundive":
						html = append(html, makepcphdr)
					case "supine":
						html = append(html, makepcphdr)
					default:
						html = append(html, maketnshdr(v, m))
					}
				}

				html = append(html, trrhtml...)
				html = append(html, "</table>")
			}
		}
	}

	h := strings.Join(html, "")
	return h
}

func generatedeclinedtable(lang string, words map[string]string) string {
	// need something to determine which gender columns are needed
	const (
		DIALTR = `
		<tr align="center">
			<td rowspan="1" colspan="%d" class="dialectlabel">%s<br>
			</td>
		</tr>`
	)

	var dialect []string
	var cases []string
	var numbers []string
	var gend []string

	switch lang {
	case "greek":
		dialect = GKDIALECT
		cases = GKCASES
		numbers = GKNUMB
		gend = GENDERS
	case "latin":
		dialect = []string{""}
		cases = LTCASES
		numbers = LTNUMB
		gend = GENDERS
	}

	kk := stringmapkeysintoslice(words)
	needgend := func() []string {
		var need []string
		for _, g := range gend {
			if sliceseeker(g, kk) {
				need = append(need, g)
			}
		}
		return need
	}()

	makehdr := func() string {
		hd := `
		<tr>
			%s
		</tr>`
		tdd := []string{`<td class="genderlabel">&nbsp;</td>`}
		for _, g := range needgend {
			tdd = append(tdd, fmt.Sprintf(`<td class="gendercell">%s<br></td>`, g))
		}
		td := strings.Join(tdd, "")
		return fmt.Sprintf(hd, td)
	}()

	maketrr := func(d string) []string {
		// this code fragment is highly convergent with what is needed for participles; duplicating for now
		var trr []string
		for _, n := range numbers {
			for _, c := range cases {
				trr = append(trr, `<tr class="morphrow">`)
				trr = append(trr, fmt.Sprintf(`<td class="morphlabelcell">%s %s</td>`, n, c))
				var tdd []string
				blankcount := 0
				for _, g := range needgend {
					// not every combination should be generated
					// fem_acc_dual_doric
					k := fmt.Sprintf("%s_%s_%s_%s", g, c, n, d)
					if _, ok := words[k]; ok {
						tdd = append(tdd, words[k])
					} else {
						tdd = append(tdd, BLANK)
						blankcount += 1
					}
				}
				for _, td := range tdd {
					trr = append(trr, fmt.Sprintf(`<td class="morphcell">%s</td>`, td))
				}
				trr = append(trr, `</tr>`)
			}
		}
		return trr
	}

	var html []string

	for _, d := range dialect {
		// each dialect is a major section
		// but latin has only one dialect
		html = append(html, `<table class="verbanalysis">`)
		html = append(html, makehdr)
		html = append(html, fmt.Sprintf(DIALTR, 3, d))
		trr := maketrr(d)
		html = append(html, trr...)
		html = append(html, "</table>")
	}

	h := strings.Join(html, "")
	return h
}

//
// HELPERS
//

// stringseeker - if s is in the []string produced via splitting, then true
func stringseeker(skg string, split string) bool {
	slc := strings.Split(split, JOINER)
	for _, s := range slc {
		if s == skg {
			return true
		}
	}
	return false
}

// sliceseeker - if s is in the []strings produced via splitting spp, then true
func sliceseeker(s string, spp []string) bool {
	for _, sp := range spp {
		if stringseeker(s, sp) {
			return true
		}
	}
	return false
}

// multistringseeker - if any s in []string is in the []string produced via splitting, then true
func multistringseeker(ss []string, split string) bool {
	for _, s := range ss {
		if stringseeker(s, split) {
			return true
		}
	}
	return false
}

// arraystringseeker - if any s in []string is in the []strings produced via splitting each of spp, then true
func arraystringseeker(ss []string, spp []string) bool {
	for _, sp := range spp {
		if multistringseeker(ss, sp) {
			return true
		}
	}
	return false
}
