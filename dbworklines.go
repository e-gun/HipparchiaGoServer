//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// hipparchiaDB-# \d gr0001
//                                     Table "public.gr0001"
//      Column      |          Type          | Collation | Nullable |           Default
//------------------+------------------------+-----------+----------+-----------------------------
// index            | integer                |           | not null | nextval('gr0001'::regclass)
// wkuniversalid    | character varying(10)  |           |          |
// level_05_value   | character varying(64)  |           |          |
// level_04_value   | character varying(64)  |           |          |
// level_03_value   | character varying(64)  |           |          |
// level_02_value   | character varying(64)  |           |          |
// level_01_value   | character varying(64)  |           |          |
// level_00_value   | character varying(64)  |           |          |
// marked_up_line   | text                   |           |          |
// accented_line    | text                   |           |          |
// stripped_line    | text                   |           |          |
// hyphenated_words | character varying(128) |           |          |
// annotations      | character varying(256) |           |          |
//Indexes:
//    "gr0001_index_key" UNIQUE CONSTRAINT, btree (index)
//    "gr0001_mu_trgm_idx" gin (accented_line gin_trgm_ops)
//    "gr0001_st_trgm_idx" gin (stripped_line gin_trgm_ops)

const (
	WORLINETEMPLATE = `wkuniversalid, index,
			level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value,
			marked_up_line, accented_line, stripped_line, hyphenated_words, annotations`
)

var (
	nohtml   = regexp.MustCompile("<[^>]*>") // crude, and will not do all of everything
	metadata = regexp.MustCompile(`<hmu_metadata_(.*?) value="(.*?)" />`)
	mdformat = regexp.MustCompile(`&3(.*?)&`) // see andsubstitutes in betacodefontshifts.py
	mdremap  = map[string]string{"provenance": "loc:", "documentnumber": "#", "publicationinfo": "pub:", "notes": "",
		"city": "c:", "region": "r:", "date": "d:"}
	// the above has to be kept in sync w/ l 130 of rt-browser.go ()
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

type DbWorkline struct {
	WkUID       string
	TbIndex     int64
	Lvl5Value   string
	Lvl4Value   string
	Lvl3Value   string
	Lvl2Value   string
	Lvl1Value   string
	Lvl0Value   string
	MarkedUp    string
	Accented    string
	Stripped    string
	Hyphenated  string
	Annotations string
	// beyond the db stuff
	EmbNotes map[string]string
}

func (dbw DbWorkline) FindLocus() []string {
	loc := [6]string{
		dbw.Lvl5Value,
		dbw.Lvl4Value,
		dbw.Lvl3Value,
		dbw.Lvl2Value,
		dbw.Lvl1Value,
		dbw.Lvl0Value,
	}

	var trim []string
	for _, l := range loc {
		if l != "-1" {
			trim = append(trim, l)
		}
	}
	return trim
}

func (dbw DbWorkline) FindAuthor() string {
	return dbw.WkUID[:6]
}

func (dbw DbWorkline) FindWork() string {
	return dbw.WkUID[7:]
}

func (dbw DbWorkline) FindCorpus() string {
	return dbw.WkUID[0:2]
}

func (dbw DbWorkline) BuildHyperlink() string {
	if len(dbw.WkUID) == 0 {
		// formatwithcontextresults() will trigger this
		msg("BuildHyperlink() on empty dbworkline", 5)
		return ""
	}
	t := `linenumber/%s/%s/%d`
	return fmt.Sprintf(t, dbw.FindAuthor(), dbw.FindWork(), dbw.TbIndex)
}

func (dbw *DbWorkline) GatherMetadata() {
	md := make(map[string]string)
	if metadata.MatchString(dbw.MarkedUp) {
		mm := metadata.FindAllStringSubmatch(dbw.MarkedUp, -1)
		for _, m := range mm {
			// sample location:
			// hipparchiaDB=# select index, marked_up_line from lt0474 where index = 116946;
			md[m[1]] = m[2]
		}

		// fmt.Println(md)

		dbw.MarkedUp = metadata.ReplaceAllString(dbw.MarkedUp, "")
		for k, v := range md {
			md[k] = mdformat.ReplaceAllString(v, `<span class="embeddedannotations foundwork">$1</span>`)
			if _, y := mdremap[k]; y {
				md[mdremap[k]] = md[k]
				delete(md, k)
			}
		}
	}

	dbw.EmbNotes = md
}

func (dbw *DbWorkline) PurgeMetadata() {
	if metadata.MatchString(dbw.MarkedUp) {
		dbw.MarkedUp = metadata.ReplaceAllString(dbw.MarkedUp, "")
	}
}

func (dbw DbWorkline) SameLevelAs(other DbWorkline) bool {
	// to help toggle the counters when building texts
	one := dbw.Lvl1Value == other.Lvl1Value
	two := dbw.Lvl2Value == other.Lvl2Value
	three := dbw.Lvl3Value == other.Lvl3Value
	four := dbw.Lvl4Value == other.Lvl4Value
	five := dbw.Lvl5Value == other.Lvl5Value
	if one && two && three && four && five {
		return true
	} else {
		return false
	}
}

func (dbw DbWorkline) StrippedSlice() []string {
	return strings.Split(dbw.Stripped, " ")
}

func (dbw DbWorkline) AccentedSlice() []string {
	return strings.Split(dbw.Accented, " ")
}

func (dbw DbWorkline) MarkedUpSlice() []string {
	cln := nohtml.ReplaceAllString(dbw.MarkedUp, "")
	return strings.Split(cln, " ")
}

func (dbw DbWorkline) Citation() string {
	return strings.Join(dbw.FindLocus(), ".")
}

// worklinequery - use a PrerolledQuery to acquire []DbWorkline
func worklinequery(prq PrerolledQuery, dbpool *pgxpool.Pool) []DbWorkline {
	// [a] build a temp table if needed

	// fmt.Printf("TT:\n%s\n", prq.TempTable)
	if prq.TempTable != "" {
		_, err := dbpool.Exec(context.Background(), prq.TempTable)
		chke(err)
	}

	// [b] execute the main query
	var foundrows pgx.Rows
	var err error

	// fmt.Printf("Q:\n%s\n", prq.PsqlQuery)
	if prq.PsqlData != "" {
		foundrows, err = dbpool.Query(context.Background(), prq.PsqlQuery, prq.PsqlData)
		chke(err)
	} else {
		foundrows, err = dbpool.Query(context.Background(), prq.PsqlQuery)
		chke(err)
	}

	// [c] convert the finds into []DbWorkline
	var thesefinds []DbWorkline

	defer foundrows.Close()
	for foundrows.Next() {
		// [vi.1] convert the finds into DbWorklines
		var thehit DbWorkline
		err := foundrows.Scan(&thehit.WkUID, &thehit.TbIndex, &thehit.Lvl5Value, &thehit.Lvl4Value, &thehit.Lvl3Value,
			&thehit.Lvl2Value, &thehit.Lvl1Value, &thehit.Lvl0Value, &thehit.MarkedUp, &thehit.Accented,
			&thehit.Stripped, &thehit.Hyphenated, &thehit.Annotations)
		chke(err)
		thesefinds = append(thesefinds, thehit)
	}

	return thesefinds
}

// graboneline - return a single DbWorkline from a table
func graboneline(table string, line int64) DbWorkline {
	dbpool := GetPSQLconnection()
	defer dbpool.Close()
	qt := "SELECT %s FROM %s WHERE index = %s ORDER by index"
	var prq PrerolledQuery
	prq.TempTable = ""
	prq.PsqlData = ""
	prq.PsqlQuery = fmt.Sprintf(qt, WORLINETEMPLATE, table, strconv.FormatInt(line, 10))
	foundlines := worklinequery(prq, dbpool)
	if len(foundlines) != 0 {
		return foundlines[0]
	} else {
		return DbWorkline{}
	}
}

// simplecontextgrabber - grab a pile of lines centered around the focusline
func simplecontextgrabber(table string, focus int64, context int64) []DbWorkline {
	dbpool := GetPSQLconnection()
	defer dbpool.Close()

	qt := "SELECT %s FROM %s WHERE (index BETWEEN %s AND %s) ORDER by index"

	low := focus - context
	high := focus + context

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

	need := lvls - atlvl

	if atlvl < 0 || need < 0 {
		// logic bug in here somehwere...
		msg("findvalidlevelvalues() sent negative levels", -1)
		return LevelValues{}
	}

	// [b] make a query

	qmap := map[int]string{0: "level_00_value", 1: "level_01_value", 2: "level_02_value", 3: "level_03_value",
		4: "level_04_value", 5: "level_05_value"}

	t := SELECTFROM + ` WHERE wkuniversalid='%s' %s %s ORDER BY index ASC`

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
		// fmt.Println(ands)
	}

	var and string
	if len(ands) > 0 {
		and = " AND " + strings.Join(ands, " AND ")
	}
	andnot := fmt.Sprintf(`AND %s NOT IN ('t')`, qmap[atlvl])

	var prq PrerolledQuery
	prq.PsqlQuery = fmt.Sprintf(t, w.FindAuthor(), wkid, and, andnot)

	dbpool := GetPSQLconnection()
	defer dbpool.Close()
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
