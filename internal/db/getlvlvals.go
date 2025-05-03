//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package db

import (
	"context"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/jackc/pgx/v5"
	"sort"
	"strings"
)

// FindValidLevelValues - tell me some of a citation and I can tell you what is a valid choice at the next step
func FindValidLevelValues(dbw str.DbWork, locc []string, thisisthesecondtry bool) str.LevelValues {
	// route: /get/json/workstructure/gr0033/003/11 -> RtGetJSWorksStruct -> here

	// curl localhost:5000/get/json/workstructure/lt0959/001
	// {"totallevels": 3, "level": 2, "label": "book", "low": "1", "high": "3", "range": ["1", "2", "3"]}
	// curl localhost:5000/get/json/workstructure/lt0959/001/2
	// {"totallevels": 3, "level": 1, "label": "poem", "low": "1", "high": "19", "range": ["1", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "2", "3", "4", "5", "6", "7", "8", "9a", "9b"]}

	// select levellabels_00, levellabels_01, levellabels_02, levellabels_03, levellabels_04, levellabels_05 from works where universalid = 'lt0959w001';
	// levellabels_00 | levellabels_01 | levellabels_02 | levellabels_03 | levellabels_04 | levellabels_05
	//----------------+----------------+----------------+----------------+----------------+----------------
	// verse          | poem           | book           |                |                |

	const (
		SELECTFROM = `
		SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, 
			marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM %s`
		SEL    = SELECTFROM + ` WHERE wkuniversalid='%s' %s %s ORDER BY index ASC`
		ANDNOT = `AND %s NOT IN ('t')`
	)

	var vals str.LevelValues

	// [a] what do we need?

	lmap := map[int]string{0: dbw.LL0, 1: dbw.LL1, 2: dbw.LL2, 3: dbw.LL3, 4: dbw.LL4, 5: dbw.LL5}

	lvls := dbw.CountLevels() - 1 // Count vs indexing adjustment
	atlvl := 0
	if locc[0] == "" {
		// at top
		atlvl = lvls
	} else {
		atlvl = lvls - len(locc)
	}

	need := lvls - atlvl

	if atlvl < 0 || need < 0 {
		// logic bug in here somewhere...
		// FAIL = "FindValidLevelValues() sent negative levels"
		// mm(FAIL, MSGWARN)
		return str.LevelValues{}
	}

	// [b] make a query

	qmap := map[int]string{0: "level_00_value", 1: "level_01_value", 2: "level_02_value", 3: "level_03_value",
		4: "level_04_value", 5: "level_05_value"}

	syntax := "="
	if thisisthesecondtry {
		syntax = "~*"
	}

	var ands []string
	for i := 0; i < need; i++ {
		// example: xen's anabasis (gr0032w006) has 4 levels
		// top is 3; need just all vals @ 3; so no ands
		// next is 2; need "level_03_value='X'" (ie, qmap[3] and locc[0])
		// next is 1; need "level_03_value='X' AND level_02_value='Y'" (ie, qmap[3] and locc[0] + qmap[2] and locc[1])
		// next is 0; need "level_03_value='X' AND level_02_value='Y' AND level_01_value='Z'"
		q := lvls - i

		// 'fr%20a' might come in, but you are looking for 'fr a'...
		locc[i] = strings.ReplaceAll(locc[i], "%20", " ")

		a := fmt.Sprintf(`%s %s '%s'`, qmap[q], syntax, locc[i])
		ands = append(ands, a)
	}

	var and string
	if len(ands) > 0 {
		and = " AND " + strings.Join(ands, " AND ")
	}
	andnot := fmt.Sprintf(ANDNOT, qmap[atlvl])

	var prq str.PrerolledQuery
	prq.PsqlQuery = fmt.Sprintf(SEL, dbw.AuID(), dbw.UID, and, andnot)

	wlb := GetWorklineBundle(prq)

	var validlines []str.DbWorkline

	if wlb.Len() == 0 && dbw.IsPHI() && !thisisthesecondtry {
		// A NOTE CONCERING PHI IRREGULARITIES...

		// see the notes in HGB gathernewinsworkinfo() concerning `fr a` and `fr b` level values both occupying `fr`
		// you will also see there why the builder cannot fix the issue and so an ugly hack has to appear here

		// FindValidLevelValues() will do `= 'fr'` at l2, but only `= 'fr a'` finds anything
		// so, on the assumption that you are doing such a search, you reach this moment of code with `wlb` empty
		// wlb should never normally be empty; let's make a second try... this time we do `~*` so that `fr` finds `fr a` and `fr b`

		vals = FindValidLevelValues(dbw, locc, true)

		// [c0] - maybe I have too many lines...

		// notice that you now have a NEW PROBLEM if you did a `~*` search for `fr`: you capture too many lines at the next level down
		// `fr a` can be associated with `col I` in the results even though they in fact are not so associated in the data:
		// `col I` is only a `fr b` feature...

		// no returns for
		// SELECT index FROM inz043 WHERE wkuniversalid='inz043w00d' AND level_02_value='fr a' AND level_01_value='col II' AND level_00_value='4' ORDER BY index ASC

		//    58 | inz043w00d    | -1             | -1             | -1             | fr a           | col II/III/IV? | 10             | &nbsp;&nbsp;&nbsp;[ηʹ Φλαύ(ιοϲ) {27Κλαύ(διοϲ)}27 Ὑ]ψ̣ικλῆϲ̣ [Λυϲιϲτράτου {27βʹ καθ’ ὑ(οθεϲίαν δὲ) Ποϲιδωνίου}27]
		//    59 | inz043w00d    | -1             | -1             | -1             | fr a           | col II/III/IV? | 11             | [—  —  —  —  —  —  —  —  —  —  —  —  —  —  —  —  —  — ]
		//    60 | inz043w00d    | -1             | -1             | -1             | fr b           | col I          | 26             | [—  —  —  —  —  — ]ευϲ
		//    61 | inz043w00d    | -1             | -1             | -1             | fr b           | col I          | 27             | [—  —  —  —  —  — ]∙
		//    62 | inz043w00d    | -1             | -1             | -1             | fr b           | col I          | 28             | [—  —  —  —  —  — ]
		//
		//    98 | inz043w00d    | -1             | -1             | -1             | fr b           | col I          | 64             | [—  —  —  —  —  —  — ]
		//    99 | inz043w00d    | -1             | -1             | -1             | fr b           | col I          | 65             | [—  —  —  —  —  —  — ]
		//   100 | inz043w00d    | -1             | -1             | -1             | fr b           | col II         | 1              | [ηʹ —  —  —  —  —  —  —  —  —  —  —  —  —  — ]
		//   101 | inz043w00d    | -1             | -1             | -1             | fr b           | col II         | 2              | [θʹ —  —  —  —  —  —  —  —  —  —  —  —  —  — ]
		//   102 | inz043w00d    | -1             | -1             | -1             | fr b           | col II         | 3              | [ιʹ —  —  —  —  —  —  —  —  —  —  —  —  —  — ]

		// more problems here: `col II/III/IV?` will freak out browse() because `elem := strings.Split(locus, "/")` and "?" is disallowed...

		// if you hit lvl0 and locc = `[fr%20a col%20II 5]`, you have failed
		// NB: at lvl1 locc = `[fr a col]`

		// what you need to do is manually take care of the `select ... where...` that could not be sent to psql
		// i.e., trim out all results where the l2 criterion is not met... [ie, atlvl + 1 constrains atlvl]

		for _, l := range wlb.Lines {
			if handcheckiflocusisok(l, locc, atlvl) {
				validlines = append(validlines, l)
			}
		}

		if len(validlines) == 0 {
			return vals
		}
	} else {
		validlines = wlb.Lines
	}

	// [c1] extract info from the hitlines returned
	vals.AtLvl = atlvl
	vals.Label = lmap[atlvl]

	first := validlines[0]
	vals.Total = first.Lvls()
	vals.Low = first.LvlVal(atlvl)
	vals.High = wlb.Lines[wlb.Len()-1].LvlVal(atlvl)
	var r []string

	for i := range validlines {
		r = append(r, validlines[i].LvlVal(atlvl))
	}

	r = gen.Unique(r)
	sort.Strings(r) // todo: ? sort by index which is the real order anyway...
	vals.Range = r
	vals.CleanVals()

	return vals
}

func handcheckiflocusisok(ln str.DbWorkline, locc []string, atlvl int) bool {
	// curl localhost:8001/get/json/workstructure/inz043/00d
	//{
	//  "totallevels": 3,
	//  "level": 2,
	//  "label": "fr",
	//  "low": "fr a",
	//  "high": "fr b",
	//  "range": [
	//    "fr a",
	//    "fr b"
	//  ]
	//}

	// curl localhost:8001/get/json/workstructure/inz043/00d/fr%20b
	//{
	//  "totallevels": 3,
	//  "level": 1,
	//  "label": "col",
	//  "low": "col I",
	//  "high": "col III",
	//  "range": [
	//    "col I",
	//    "col II",
	//    "col III"
	//  ]
	//}

	//  trim out all results where the higher criterion is not met... [ie, atlvl + 1 constrains atlvl]

	// the index count of locc moves in the opposite direction of atlvl: 0, 2; 1, 1; 2, 0
	isok := true

	// should check all higher levels; for now checking just one up
	// this will produce index errors until you catch them
	tocheck := len(locc) - atlvl - 1
	if tocheck < 0 {
		return true
	}
	if len(locc) == 0 {
		return false
	}

	//fmt.Println("tocheck", tocheck)
	//fmt.Println("ln.LvlVal(atlvl+1)", ln.LvlVal(atlvl+1))
	//fmt.Println("locc[tocheck]", locc[tocheck])

	if ln.LvlVal(atlvl+1) != locc[tocheck] {
		isok = false
	}
	return isok
}

// GetLocusEndpoints - query db for index values correspond to the start and end of a text segment like "book 2"
func GetLocusEndpoints(wk *str.DbWork, locus string, sep string) ([2]int, bool) {
	// [HGS] wuid: 'lt0474w049'; locus: '3|14|_0'; sep: '|'
	// [HGS] wuid: 'lt0474w049'; locus: '4:8:18'; sep: ':'

	const (
		QTMP = `SELECT index FROM %s WHERE wkuniversalid='%s' AND %s ORDER BY index ASC`
		FAIL = "GetLocusEndpoints() failed to find the following inside of %s: '%s'"
	)

	dbconn := getdbconnection()
	defer dbconn.Release()

	// 'fr%20a' will come in, but you are looking for 'fr a'...
	locus = strings.ReplaceAll(locus, "%20", " ")
	locus = strings.ReplaceAll(locus, "%EF%BC%8F", "/") // "／"
	locus = strings.ReplaceAll(locus, "%EF%BC%9F", "?") // "？"

	fl := [2]int{0, 0}
	success := false

	wl := wk.CountLevels()

	ll := strings.Split(locus, sep)
	if len(ll) > wl {
		ll = ll[0:wl]
	}

	if len(ll) == 0 || ll[0] == "_0" {
		fl = [2]int{wk.FirstLine, wk.LastLine}
		return fl, true
	}

	if ll[len(ll)-1] == "_0" {
		ll = ll[0 : len(ll)-1]
	}

	col := []string{"level_00_value", "level_01_value", "level_02_value", "level_03_value", "level_04_value", "level_05_value"}
	tem := `%s='%s'`
	var use []string
	for i, l := range ll {
		s := fmt.Sprintf(tem, col[wl-i-1], l)
		use = append(use, s)
	}

	tb := wk.AuID()

	a := strings.Join(use, " AND ")
	q := fmt.Sprintf(QTMP, tb, wk.UID, a)

	foundrows, err := dbconn.Query(context.Background(), q)
	Msg.EC(err)

	idx, err := pgx.CollectRows(foundrows, pgx.RowTo[int])
	Msg.EC(err)

	if len(idx) == 0 {
		// bogus input
		Msg.PEEK(fmt.Sprintf(FAIL, wk.UID, locus))
		fl = [2]int{1, 1}
	} else {
		fl = [2]int{idx[0], idx[len(idx)-1]}
		success = true
	}
	return fl, success
}
