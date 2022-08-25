package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"sort"
	"strconv"
	"strings"
)

type LevelValues struct {
	// for JSON output...
	// {"totallevels": 3, "level": 2, "label": "book", "low": "1", "high": "3", "range": ["1", "2", "3"]}
	Total int      `json:"totallevels"`
	AtLvl int      `json:"level"`
	Label string   `json:"label"`
	Low   string   `json:"low"`
	High  string   `json:"high"`
	Range []string `json:"range"`
}

// findtherows - use a redis.Conn to acquire []DbWorkline
func findtherows(thequery string, thecaller string, searchkey string, clientnumber int, rc redis.Conn, dbpool *pgxpool.Pool) []DbWorkline {
	// called by both linegrabber() and HipparchiaBagger()
	// this version contains polling data
	// it also assumes that thequery arrived via popping redis

	// [ii] update the polling data
	if thecaller != "bagger" {
		remain, err := redis.Int64(rc.Do("SCARD", searchkey))
		checkerror(err)

		k := fmt.Sprintf("%s_remaining", searchkey)
		_, e := rc.Do("SET", k, remain)
		checkerror(e)
		msg(fmt.Sprintf("%s #%d says that %d items remain", thecaller, clientnumber, remain), 3)
	}

	// [iii] decode the query
	var prq PrerolledQuery
	err := json.Unmarshal([]byte(thequery), &prq)
	checkerror(err)

	// fmt.Println(prq)
	foundlines := worklinequery(prq, dbpool)

	return foundlines
}

// worklinequery - use a PrerolledQuery to acquire []DbWorkline
func worklinequery(prq PrerolledQuery, dbpool *pgxpool.Pool) []DbWorkline {
	// [a] build a temp table if needed
	if prq.TempTable != "" {
		_, err := dbpool.Exec(context.Background(), prq.TempTable)
		checkerror(err)
	}

	// [b] execute the main query
	var foundrows pgx.Rows
	var err error

	if prq.PsqlData != "" {
		foundrows, err = dbpool.Query(context.Background(), prq.PsqlQuery, prq.PsqlData)
		checkerror(err)
	} else {
		foundrows, err = dbpool.Query(context.Background(), prq.PsqlQuery)
		checkerror(err)
	}

	// [c] convert the finds into []DbWorkline
	var thesefinds []DbWorkline

	defer foundrows.Close()
	for foundrows.Next() {
		// [vi.1] convert the finds into DbWorklines
		var thehit DbWorkline
		err := foundrows.Scan(&thehit.WkUID, &thehit.TbIndex, &thehit.Lvl5Value, &thehit.Lvl4Value, &thehit.Lvl3Value,
			&thehit.Lvl2Value, &thehit.Lvl1Value, &thehit.Lvl0Value, &thehit.MarkedUp, &thehit.Accented,
			&thehit.Stripped, &thehit.Hypenated, &thehit.Annotations)
		checkerror(err)
		thesefinds = append(thesefinds, thehit)
	}

	return thesefinds
}

// simplecontextgrabber - grab a pile of lines centered around the focusline
func simplecontextgrabber(table string, focus int64, context int64) []DbWorkline {
	dbpool := grabpgsqlconnection()

	qt := "SELECT %s FROM %s WHERE (index BETWEEN %s AND %s) ORDER by index"

	low := focus - (context / 2)
	high := focus + (context / 2)

	var prq PrerolledQuery

	prq.TempTable = ""
	prq.PsqlData = ""
	prq.PsqlQuery = fmt.Sprintf(qt, WORLINETEMPLATE, table, strconv.FormatInt(low, 10), strconv.FormatInt(high, 10))

	foundlines := worklinequery(prq, dbpool)

	return foundlines
}

func findvalidlevelvalues(wkid string, locc []string) LevelValues {
	// tell me some of a citation and i can tell you what is a valid choice at the next step
	// curl localhost:5000/get/json/workstructure/lt0959/001
	// {"totallevels": 3, "level": 2, "label": "book", "low": "1", "high": "3", "range": ["1", "2", "3"]}
	// curl localhost:5000/get/json/workstructure/lt0959/001/2
	// {"totallevels": 3, "level": 1, "label": "poem", "low": "1", "high": "19", "range": ["1", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "2", "3", "4", "5", "6", "7", "8", "9a", "9b"]}

	// select levellabels_00, levellabels_01, levellabels_02, levellabels_03, levellabels_04, levellabels_05 from works where universalid = 'lt0959w001';
	// levellabels_00 | levellabels_01 | levellabels_02 | levellabels_03 | levellabels_04 | levellabels_05
	//----------------+----------------+----------------+----------------+----------------+----------------
	// verse          | poem           | book           |                |                |

	// [a] what do we need?

	w := AllWorks[wkid]
	lmap := map[int]string{0: w.LL0, 1: w.LL1, 2: w.LL2, 3: w.LL3, 4: w.LL4, 5: w.LL5}

	lvls := w.CountLevels() - 1 // count vs indexing adjustment
	atlvl := 0
	if locc[0] == "" {
		// at top
		atlvl = lvls
	} else {
		atlvl = lvls - len(locc)
	}

	if atlvl < 0 {
		return LevelValues{}
	}
	need := lvls - atlvl - 1

	// [b] make a query

	// top: SELECT ... FROM lt0959 WHERE ( wkuniversalid=%s ) AND level_02_value NOT IN (%s) ORDER BY index ('lt0959w001', 't')
	// first: SELECT ... FROM lt0959 WHERE ( wkuniversalid=%s ) AND  level_02_value=%s AND level_01_value NOT IN (%s) ORDER BY index ('lt0959w001', '1', 't')
	// second: SELECT ... FROM lt0959 WHERE ( wkuniversalid=%s ) AND  level_02_value=%s AND  level_01_value=%s AND level_00_value NOT IN (%s) ORDER BY index ('lt0959w001', '1', '3', 't')

	qmap := map[int]string{0: "level_00_value", 1: "level_01_value", 2: "level_02_value", 3: "level_03_value",
		4: "level_04_value", 5: "level_05_value"}

	t := SELECTFROM + ` WHERE wkuniversalid='%s' %s %s ORDER BY index ASC`

	var ands []string
	for i := atlvl; i < need; i-- {
		a := fmt.Sprintf(`%s='%s'`, qmap[i], locc[len(locc)-i])
		ands = append(ands, a)
	}
	and := strings.Join(ands, " AND ")
	andnot := fmt.Sprintf(`AND %s NOT IN ('t')`, qmap[atlvl])

	var prq PrerolledQuery
	prq.PsqlQuery = fmt.Sprintf(t, w.FindAuthor(), wkid, and, andnot)

	dbpool := grabpgsqlconnection()
	lines := worklinequery(prq, dbpool)

	// [c] extract info from the hitlines returned
	var vals LevelValues
	vals.AtLvl = atlvl
	vals.Label = lmap[atlvl]

	if len(lines) == 0 {
		return vals
	}
	vals.Low = picklvlval(atlvl, lines[0])
	vals.High = picklvlval(atlvl, lines[len(lines)-1])
	var r []string
	for i, _ := range lines {
		r = append(r, picklvlval(atlvl, lines[i]))
	}
	r = unique(r)
	sort.Strings(r)
	vals.Range = r
	return vals
}

func picklvlval(lvl int, ln DbWorkline) string {
	// reflection and type checking is every bit as cumbersome as this stupid solution
	switch lvl {
	case 0:
		return ln.Lvl0Value
	case 1:
		return ln.Lvl1Value
	case 2:
		return ln.Lvl2Value
	case 3:
		return ln.Lvl3Value
	case 4:
		return ln.Lvl4Value
	case 5:
		return ln.Lvl5Value
	default:
		return ""
	}
}
