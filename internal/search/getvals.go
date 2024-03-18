package search

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/str"
	"sort"
	"strings"
)

// FindValidLevelValues - tell me some of a citation and I can tell you what is a valid choice at the next step
func FindValidLevelValues(dbw str.DbWork, locc []string) str.LevelValues {
	// curl localhost:5000/get/json/workstructure/lt0959/001
	// {"totallevels": 3, "level": 2, "label": "book", "low": "1", "high": "3", "range": ["1", "2", "3"]}
	// curl localhost:5000/get/json/workstructure/lt0959/001/2
	// {"totallevels": 3, "level": 1, "label": "poem", "low": "1", "high": "19", "range": ["1", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "2", "3", "4", "5", "6", "7", "8", "9a", "9b"]}

	// select levellabels_00, levellabels_01, levellabels_02, levellabels_03, levellabels_04, levellabels_05 from works where universalid = 'lt0959w001';
	// levellabels_00 | levellabels_01 | levellabels_02 | levellabels_03 | levellabels_04 | levellabels_05
	//----------------+----------------+----------------+----------------+----------------+----------------
	// verse          | poem           | book           |                |                |

	const (
		SEL    = SELECTFROM + ` WHERE wkuniversalid='%s' %s %s ORDER BY index ASC`
		ANDNOT = `AND %s NOT IN ('t')`
	)

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
		// m(FAIL, MSGWARN)
		return str.LevelValues{}
	}

	// [b] make a query

	qmap := map[int]string{0: "level_00_value", 1: "level_01_value", 2: "level_02_value", 3: "level_03_value",
		4: "level_04_value", 5: "level_05_value"}

	var ands []string
	for i := 0; i < need; i++ {
		// example: xen's anabasis (gr0032w006) has 4 levels
		// top is 3; need just all vals @ 3; so no ands
		// next is 2; need "level_03_value='X'" (ie, qmap[3] and locc[0])
		// next is 1; need "level_03_value='X' AND level_02_value='Y'" (ie, qmap[3] and locc[0] + qmap[2] and locc[1])
		// next is 0; need "level_03_value='X' AND level_02_value='Y' AND level_01_value='Z'"
		q := lvls - i
		a := fmt.Sprintf(`%s='%s'`, qmap[q], locc[i])
		ands = append(ands, a)
	}

	var and string
	if len(ands) > 0 {
		and = " AND " + strings.Join(ands, " AND ")
	}
	andnot := fmt.Sprintf(ANDNOT, qmap[atlvl])

	var prq str.PrerolledQuery
	prq.PsqlQuery = fmt.Sprintf(SEL, dbw.AuID(), dbw.UID, and, andnot)

	dbconn := db.GetDBConnection()
	defer dbconn.Release()
	wlb := AcquireWorkLineBundle(prq, dbconn)

	// [c] extract info from the hitlines returned
	var vals str.LevelValues
	vals.AtLvl = atlvl
	vals.Label = lmap[atlvl]

	if wlb.Len() == 0 {
		return vals
	}

	first := wlb.FirstLine()
	vals.Total = first.Lvls()
	vals.Low = first.LvlVal(atlvl)
	vals.High = wlb.Lines[wlb.Len()-1].LvlVal(atlvl)
	var r []string

	for i := range wlb.Lines {
		r = append(r, wlb.Lines[i].LvlVal(atlvl))
	}
	r = gen.Unique(r)
	sort.Strings(r)
	vals.Range = r

	return vals
}
