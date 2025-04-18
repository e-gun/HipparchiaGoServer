//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/format"
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/labstack/echo/v4"
	"slices"
	"strconv"
	"strings"
)

// RtMorphchart - return a chart mapping known forms of a word to their grammatical identification
func RtMorphchart(c echo.Context) error {
	// /lexica/morphologychart/greek/39046.0/37925260/ἐπιγιγνώϲκω

	// should reach this route exclusively via a click from rt-lexica.go
	c.Response().After(func() { vlt.LogPaths("RtMorphchart()") })
	user := vlt.ReadUUIDCookie(c)
	if !vlt.AllAuthorized.Check(user) {
		return jsonresponse(c, str.SearchOutputJSON{JS: vv.VALIDATIONBOX})
	}
	s := vlt.AllSessions.GetSess(user)

	const (
		CTM   = `<verbform searchterm="%s">%s</verbform> (<span class="counter">%d</span>)`
		TBTOP = `
		<div class="center">
			<span class="verylarge">All known forms of <dictionaryidsearch entryid="%f" language="%s">%s</dictionaryidsearch></span>
		</div>`
	)

	// [a] parse request

	req := c.Param("wd")
	elem := strings.Split(req, "/")

	if len(elem) != 4 || elem[0] == "" {
		return emptyjsreturn(c)
	}

	lg := elem[0]
	id, e1 := strconv.ParseFloat(elem[1], 32)
	_, e2 := strconv.Atoi(elem[2])
	wd := gen.Purgechars(vv.UNACCEPTABLEINPUT, elem[3])
	gl := lg == "greek" || lg == "latin"

	if !gl || e1 != nil || e2 != nil {
		return emptyjsreturn(c)
	}

	// if e2 == nil it is safe to use elem[2] as the (string) xref val
	xr := elem[2]

	// [b] get all forms of the word:  map[string]str.DbMorphology

	dbmmap := db.GetAllFormsOf(lg, xr)

	// [c] get all counts for all forms: [c] and [d-e] can run concurrently
	// [c1] slice of the words; map of the first letters of those words
	ww := make([]string, len(dbmmap))
	lett := make(map[string]bool)
	var words []string

	count := 0
	for _, f := range dbmmap {
		fo := f.Observed
		ww[count] = fo
		r := []rune(fo)
		init := gen.StripaccentsRUNE(r)
		lett[string(init[0])] = true
		count += 1
		words = append(words, f.Observed)
	}

	// [c2] query the database

	wcc := db.GetMultipleWordCounts(words)

	// [d] extract parsing info for all forms

	mpp := make(map[string][]string)
	// will look like:
	// credam:[ fut ind act 1st sg , pres subj act 1st sg]
	// credamus:[ pres subj act 1st pl]
	// credamusque:[ pres subj act 1st pl]
	// credant:[ pres subj act 3rd pl]
	// ...

	for k, v := range dbmmap {
		vall := []str.DbMorphology{v} // dbmorphintomorphpossib() wants a slice, we fake a slice
		mp := dbmorphintomorphpossib(vall)
		for _, m := range mp {
			// item 0 is always ""; item 1 is an actual analysis
			mpp[k] = append(mpp[k], m.Analysis)
		}
	}

	// WARNING: you just keyed »ἥρμοττ'« (mpp[ἥρμοττ']), but the value is associated with »ἥρμοττ« at wcc[ἥρμοττ]
	// NB: mpp keys will next be seen in pdm

	// [e] generate parsing map: [parsedata]form
	// this effectively flips the preceding map: k, v --> v, k
	// 	fut ind act 1st sg: credam
	// 	pres subj act 1st sg: credam
	// 	...

	// NB have to decompress "nom/voc/acc" into three entries: getparsercombinations()

	// [e1] first pass: make the map and deal with cases

	pdm := make(map[string]string)

	for k, vall := range mpp {
		for _, v := range vall {
			if len(v) == 0 {
				continue
			}
			// "imperfect" will be ruined by next if you are not careful
			v = strings.Replace(v, " mp ", " mid/pass ", -1)
			if !strings.Contains(v, "/") {
				key := strings.Replace(v, " ", format.JOINER, -1)
				if _, ok := pdm[key]; !ok {
					pdm[key] = k
				} else {
					pdm[key] = pdm[key] + " / " + k
				}
			} else {
				multiplier := getparsercombinations(v)
				for _, m := range multiplier {
					key := strings.Replace(m, " ", format.JOINER, -1)
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
				diall := strings.Split(parts[1], format.JOINER)
				for _, d := range diall {
					if slices.Contains(format.GKDIALECT, d) {
						newkey := parts[0] + format.JOINER + d
						newkey = strings.Replace(newkey, format.JOINER+format.JOINER, format.JOINER, 1)
						newpdm[newkey] = v
					}
				}
			} else {
				if !strings.Contains(k, "attic") {
					newkey := k + format.JOINER + "attic"
					newpdm[newkey] = v
				} else {
					newpdm[k] = v
				}
			}
		}
	} else {
		// add the "blank" dialect to latin
		for k, v := range pdm {
			newpdm[k+format.JOINER] = v
		}
	}
	pdm = newpdm

	// [e3] get counts for each word
	pdcm := make(map[string]map[string]int)
	for k, v := range pdm {
		wds := strings.Split(v, " / ")
		mm := make(map[string]int)
		for _, w := range wds {
			//  reassociate »ἥρμοττ'« and »ἥρμοττ«
			mm[w] = wcc[strings.Replace(w, "'", "", -1)].Total
		}
		pdcm[k] = mm
	}

	// [e4] add markup and format the counts
	pdxm := make(map[string]string)
	for kk, pd := range pdcm {
		var vcc []string
		for k, v := range pd {
			vcc = append(vcc, fmt.Sprintf(CTM, k, k, v))
		}
		pdxm[kk] = strings.Join(vcc, " / ")

		// tense_mood_voice_pers_numb_dial
		//aor_imperat_act_2nd_pl_attic: παραθλίψατε (1)
		//aor_imperat_act_2nd_sg_attic: θλῖψον (2)
		//aor_imperat_act_3rd_pl_attic: θλιψάντων (18)
		//aor_imperat_mid_2nd_sg_attic: θλῖψαι (25)
		// ...

	}

	isverb := func() bool {
		const (
			MSG = "RtMorphchart() isverb counts cases: %d; tenses: %d"
		)
		kk := gen.StringMapKeysIntoSlice(pdxm)
		// ῥώμη will trigger "verb"... : you can't choose via a single verb hit; you have to compare total form counts
		// NO GOOD: return arraystringseeker(GKTENSES, kk)

		tc := arraystringcounter(format.GKTENSES, kk)
		cc := arraystringcounter(format.LTCASES, kk)
		Msg.PEEK(fmt.Sprintf(MSG, cc, tc))

		// in greek tc is 2x cc or (far) more; in latin tc can just squeak by cc: "[HGS] cc: 94; tc: 104"
		// the "(3*tc)/2" below is required to keep ἀγαθόϲ from returning as if ἀγαθόω; otherwise "2*" would make sense
		if cc < (3*tc)/2 {
			return true
		} else {
			return false
		}
	}()

	var jb jsb

	// [f] build the table

	if isverb {
		jb.HTML = format.GenerateVerbTable(lg, pdxm)
	} else {
		jb.HTML = format.GenerateDeclinedTable(lg, pdxm)
	}

	jb.HTML = fmt.Sprintf(TBTOP, id, lg, wd) + jb.HTML

	jb.JS = vv.MORPHJS

	if s.ZapLunates {
		jb.HTML = gen.DeLunate(jb.HTML)
	}

	return jsonresponse(c, jb)
}

//
// COMBINATORIALS
//

// getparsercombinations - turn "pres part masc/fem/neut nom/voc sg" into a slice of all of its individual possibilities
func getparsercombinations(ps string) []string {
	// [a] ps := "pres part masc/fem/neut nom/voc sg"
	// [b] numpossible := [1 1 3 2 1]
	// [c] items := map[0:[pres] 1:[part] 2:[masc fem neut] 3:[nom voc] 4:[sg]]
	// [d] intcombinations := [[1 1 3 2 1] [1 1 3 1 1] [1 1 2 2 1] [1 1 2 1 1] [1 1 1 2 1] [1 1 1 1 1] [1 1 3 2 1] [1 1 3 1 1]]
	// [e] stringcombinations:
	//	pres part neut voc sg
	//	pres part neut nom sg
	//	pres part fem voc sg
	//	pres part fem nom sg
	//	pres part masc voc sg
	//	pres part masc nom sg
	//	pres part neut voc sg
	//	pres part neut nom sg

	ss := strings.Split(ps, " ")
	numpossible := make([]int, len(ss))
	items := make(map[int][]string)
	for i, s := range ss {
		items[i] = strings.Split(s, "/")
		numpossible[i] = len(items[i])
	}

	var intcombinations [][]int
	for i, n := range numpossible {
		if n > 1 {
			intcombinations = append(intcombinations, rcombinator(numpossible, n, i)...)
		}
	}

	var stringcombinations []string
	for _, cc := range intcombinations {
		var pp []string
		for i, c := range cc {
			p := items[i][c-1]
			pp = append(pp, p)
		}
		stringcombinations = append(stringcombinations, strings.Join(pp, " "))
	}

	return stringcombinations
}

// rcombinator - recursively produce combinations of integers
func rcombinator(slc []int, start int, posit int) [][]int {
	// [1 1 3 2 1] --> [[1 1 3 2 1] [1 1 3 1 1] [1 1 2 2 1] [1 1 2 1 1] [1 1 1 2 1] [1 1 1 1 1] [1 1 3 2 1] [1 1 3 1 1]]
	var combin [][]int
	if posit > len(slc) {
		return combin
	}

	if start == 1 {
		return [][]int{slc}
	}

	head := slc[0:posit]
	tail := slc[posit+1:]
	for j := start; j > 0; j-- {
		// the following overwrites the slices in the end...: [[1 1 1 1 1] [1 1 1 1 1] [1 1 1 1 1] [1 1 1 1 1]]
		// c := make([]int, len(head)+len(tail)+1)
		// c = append(append(head, j), tail...)

		// so we will do it the tedious way: copy()
		c := make([]int, len(head)+len(tail)+1)
		copy(c[:], head[:])
		copy(c[len(head):], []int{j})
		copy(c[len(head)+1:], tail[:])

		if posit+1 >= len(slc) {
			return combin
		} else {
			combin = append(combin, rcombinator(c, slc[posit+1], posit+1)...)
		}
	}
	return combin
}

//
// HELPERS
//

// multistringseeker - if any s in []string is in the []string produced via splitting, then true
func multistringseeker(ss []string, split string) bool {
	for _, s := range ss {
		if format.StringSeeker(s, split) {
			return true
		}
	}
	return false
}

// arraystringcounter - if any s in []string is in the []strings produced via splitting each of spp, then add to the count
func arraystringcounter(ss []string, spp []string) int {
	count := 0
	for _, sp := range spp {
		if multistringseeker(ss, sp) {
			count += 1
		}
	}
	return count
}
