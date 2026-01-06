//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package format

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"strings"
)

var (
	GKCASES       = []string{"nom", "gen", "dat", "acc", "voc"}
	GKNUMB        = []string{"sg", "dual", "pl"}
	GKMOODS       = []string{"ind", "subj", "opt", "imperat", "inf", "part"}
	GKVOICE       = []string{"act", "mid", "pass"}
	GKTENSES      = []string{"pres", "imperf", "fut", "aor", "perf", "plup", "futperf"} // order matters
	GKINTTENSEMAP = map[int]string{1: "Present", 2: "Imperfect", 3: "Future", 4: "Aorist", 5: "Perfect", 6: "Pluperfect", 7: "Future Perfect"}
	GKTENSEMAP    = map[string]int{"pres": 1, "imperf": 2, "fut": 3, "aor": 4, "perf": 5, "plup": 6, "futperf": 7}
	GKDIALECT     = []string{"attic", "aeolic", "doric", "epic", "homeric", "ionic"}

	LTCASES  = []string{"nom", "gen", "dat", "acc", "abl", "voc"}
	LTNUMB   = []string{"sg", "pl"}
	LTMOODS  = []string{"ind", "subj", "imperat", "inf", "part", "gerundive", "supine"}
	LTVOICE  = []string{"act", "pass"}
	LTTENSES = []string{"pres", "imperf", "fut", "perf", "plup", "futperf"} // order needs to match LTINTTENSEMAP

	LTTENSEMAP = map[string]int{"pres": 1, "imperf": 2, "fut": 3, "perf": 5, "plup": 6, "futperf": 7}
	GENDERS    = []string{"masc", "fem", "neut"}
	PERSONS    = []string{"1st", "2nd", "3rd"}
	// unused
	// GKDIALINVALID = []string{"parad", "form"}
	// LTINTTENSEMAP = map[int]string{1: "Present", 2: "Imperfect", 3: "Future", 5: "Perfect", 6: "Pluperfect", 7: "Future Perfect"}
)

const (
	BLANK  = " --- "
	DIALTR = `
		<tr align="center">
			<td rowspan="1" colspan="%d" class="dialectlabel">%s<br>
			</td>
		</tr>`
	JOINER = "_"
)

// GenerateVerbTable - given a map of grammar IDs to words, build a verb table
func GenerateVerbTable(lang string, words map[string]string) string {
	// first voice
	// then mood
	// then tense as columns and number_and_person as rows

	const (
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
		TRT   = "<tr class=\"morphrow\">\n"
		TDT   = "\t<td class=\"morphcell\">%s</td>\n"
		TDLT  = "\t<td class=\"tenselabel\">%s %s</td>\n"
		TLT   = "<tr align=\"center\"><td rowspan=\"1\" colspan=\"%d\" class=\"morphrow emph\">%s<br></td></tr>\n"
		EMPTY = "<tr><td align=\"center\">[n/a]</td></tr>\n"
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

	kk := gen.StringMapKeysIntoSlice(words)

	// do we need all theoretically possible categories?
	needy := func(someslice []string) []string {
		var need []string
		for _, g := range someslice {
			if SliceSeeker(g, kk) {
				need = append(need, g)
			}
		}
		return need
	}

	needgend := needy(gend)
	needdial := needy(dialect)
	neednumb := needy(numbers)

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

	makepcphdr := func(gg []string) string {
		hdr := `
		<tr>
			<td class="tenselabel">&nbsp;</td>
			`
		for _, g := range gg {
			hdr += fmt.Sprintf("<td class=\"tensecell\">%s<br></td>\n\t", g)
		}
		hdr += `</tr>`
		return hdr
	}

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
		for _, n := range neednumb {
			for _, p := range PERSONS {
				// tempting to build a skipper for duals...
				if n == "dual" && p == "1st" {
					continue
				}
				// np := fmt.Sprintf("%s %s", n, p)
				trr = append(trr, TRT)
				trr = append(trr, fmt.Sprintf(TDLT, n, p))
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
					trr = append(trr, fmt.Sprintf(TDT, td))
				}
				trr = append(trr, "</tr>\n")
			}
		}
		isblank := false
		if cellcount == blankcount {
			isblank = true
		}
		return trr, isblank
	}

	makepcpltrr := func(d string, m string, v string) ([]string, bool) {
		// LATIN PROBLEM REQUIRING TWEAK BELOW
		// sent: pres_part_neut_acc_sg_
		// want: pres_part_act_neut_acc_sg_

		//[HGS] aor_part_mid_fem_nom_sg_attic
		//[HGS] perf_part_mp_fem_voc_pl_attic
		const (
			TNAT = `<tr align="center"><td rowspan="1" colspan="%d" class="morphrow">[n/a]<br></td></tr>`
		)

		var trr []string
		blankcount := 0
		cellcount := 0
		for _, t := range tenses {
			// not ever combination should be generated
			thevm := vm[v][m]
			if !thevm[tm[t]] {
				continue
			}
			trr = append(trr, fmt.Sprintf(TLT, len(neednumb)+2, t))

			// we are going to skip building individual tenses that yield nothing but blanks
			var provisional []string
			emptytense := 0
			totaltense := 0
			for _, n := range neednumb {
				for _, c := range cases {
					provisional = append(provisional, TRT)
					provisional = append(provisional, fmt.Sprintf(TDLT, n, c))
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
							emptytense += 1
						}
						cellcount += 1
						totaltense += 1
					}
					for _, td := range tdd {
						provisional = append(provisional, fmt.Sprintf(TDT, td))
					}
					provisional = append(provisional, "</tr>\n")
				}
			}
			// skip empty tenses
			if emptytense == totaltense {
				trr = append(trr, fmt.Sprintf(TNAT, len(neednumb)+2))
			} else {
				trr = append(trr, provisional...)
			}
		}
		isblank := false
		if cellcount == blankcount {
			isblank = true
		}
		return trr, isblank
	}

	makegdvtrr := func(d string, m string, v string) ([]string, bool) {
		// [HGS] gerundive_neut_abl_pl_
		// [HGS] supine_neut_dat_sg_

		var trr []string

		if v == "act" {
			return trr, true
		}

		nn := neednumb
		cc := cases
		if m == "supine" {
			nn = []string{"sg"}
			cc = []string{"dat", "acc", "abl"}
			needgend = []string{"neut"}
		}

		trr = append(trr, fmt.Sprintf(TLT, len(nn)+1, ""))
		blankcount := 0
		cellcount := 0
		for _, n := range nn {
			for _, c := range cc {
				trr = append(trr, TRT)
				trr = append(trr, fmt.Sprintf(TDLT, n, c))
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
					trr = append(trr, fmt.Sprintf(TDT, td))
				}
				trr = append(trr, "</tr>\n")
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

		const (
			TDLTNBSP = "<td class=\"tenselabel\">&nbsp;</td>\n"
		)

		var trr []string
		trr = append(trr, TDLTNBSP)
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
			trr = append(trr, fmt.Sprintf(TDT, td))
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

	for _, d := range needdial {
		// each dialect is a major section
		// but latin has only one dialect
		for _, v := range voices {
			// each voice is a section
			for _, m := range moods {
				if (m == "gerundive" || m == "supine") && v == "act" {
					continue
				}

				// each mood is a table
				// not every item needs generating
				isblank := false
				// the top

				html = append(html, `<table class="verbanalysis">`)

				ct := 1

				var trrhtml []string
				switch m {
				case "part":
					ct = len(gend) + 1
					trrhtml, isblank = makepcpltrr(d, m, v)
				case "inf":
					ct = counttns(v, m) + 1
					trrhtml, isblank = makeinftrr(d, m, v)
				case "gerundive":
					ct = len(gend) + 1
					trrhtml, isblank = makegdvtrr(d, m, v)
				case "supine":
					ct = 2 // only masculine exists
					// exact same issues as gerundives
					trrhtml, isblank = makegdvtrr(d, m, v)
				default:
					ct = counttns(v, m) + 1
					trrhtml, isblank = makevftrr(d, v, m)
				}

				html = append(html, fmt.Sprintf(DIALTR, ct, d))
				html = append(html, fmt.Sprintf(VOICETR, ct, v))
				html = append(html, fmt.Sprintf(MOODTR, ct, m))

				if isblank {
					trrhtml = []string{EMPTY}
				} else {
					switch m {
					case "part":
						html = append(html, makepcphdr(needgend))
					case "inf":
						html = append(html, maketnshdr(v, m))
					case "gerundive":
						html = append(html, makepcphdr(needgend))
					case "supine":
						html = append(html, makepcphdr([]string{"neut"}))
					default:
						html = append(html, maketnshdr(v, m))
					}
				}

				html = append(html, trrhtml...)
				html = append(html, "</table>\n")
			}
		}
	}

	h := strings.Join(html, "")
	return h
}

// GenerateDeclinedTable - given a map of grammar IDs to words, build a declined from table
func GenerateDeclinedTable(lang string, words map[string]string) string {
	const (
		TDGL  = "<td class=\"genderlabel\">&nbsp;</td>\n"
		TDGC  = "\t<td class=\"gendercell\">%s<br></td>\n"
		TRMR  = "<tr class=\"morphrow\">\n"
		TDMLC = "\t<td class=\"morphlabelcell\">%s %s</td>\n"
		TDMC  = "<td class=\"morphcell\">%s</td>\n"
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

	kk := gen.StringMapKeysIntoSlice(words)
	needy := func(someslice []string) []string {
		var need []string
		for _, g := range someslice {
			if SliceSeeker(g, kk) {
				need = append(need, g)
			}
		}
		return need
	}

	needgend := needy(gend)
	needdial := needy(dialect)
	neednumb := needy(numbers)

	makehdr := func() string {
		hd := `
		<tr>
			%s
		</tr>`
		tdd := []string{TDGL}
		for _, g := range needgend {
			tdd = append(tdd, fmt.Sprintf(TDGC, g))
		}
		td := strings.Join(tdd, "")
		return fmt.Sprintf(hd, td)
	}()

	maketrr := func(d string) []string {
		// this code fragment is highly convergent with what is needed for participles; duplicating for now
		var trr []string
		for _, n := range neednumb {
			for _, c := range cases {
				trr = append(trr, TRMR)
				trr = append(trr, fmt.Sprintf(TDMLC, n, c))
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
					trr = append(trr, fmt.Sprintf(TDMC, td))
				}
				trr = append(trr, "</tr>\n")
			}
		}
		return trr
	}

	var html []string

	for _, d := range needdial {
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

// getgkvbmap - return a map that tells you what Greek verbal forms in fact exist
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

// getltvbmap - return a map that tells you what Latin verbal forms in fact exist
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

// StringSeeker - if s is in the []string produced via splitting, then true
func StringSeeker(skg string, split string) bool {
	slc := strings.Split(split, JOINER)
	for _, s := range slc {
		if s == skg {
			return true
		}
	}
	return false
}

// SliceSeeker - if s is in the []strings produced via splitting spp, then true
func SliceSeeker(s string, spp []string) bool {
	for _, sp := range spp {
		if StringSeeker(s, sp) {
			return true
		}
	}
	return false
}
