//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"regexp"
	"sort"
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
	WORLINETEMPLATE = `wkuniversalid, "index",
			level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value,
			marked_up_line, accented_line, stripped_line, hyphenated_words, annotations`
	WKLNHYPERLNKTEMPL = `index/%s/%s/%d`
	WLNMETADATATEMPL  = `<span class="embeddedannotations foundwork">$1</span>`
)

var (
	NoHTML   = regexp.MustCompile("<[^>]*>") // crude, and will not do all of everything
	Metadata = regexp.MustCompile(`<hmu_metadata_(.*?) value="(.*?)" />`)
	MDFormat = regexp.MustCompile(`&3(.*?)&`) // see andsubstitutes in betacodefontshifts.py
	// MDRemap has to be kept in sync w/ l 150 of rt-browser.go ()
	MDRemap = map[string]string{"provenance": "loc:", "documentnumber": "#", "publicationinfo": "pub:", "notes": "",
		"city": "c:", "region": "r:", "date": "d:"}
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
	TbIndex     int
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
	// beyond the db stuff; do not make this "public": pgx.RowToStructByPos will balk
	embnotes map[string]string
}

func (dbw *DbWorkline) FindLocus() []string {
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

// AuID - gr0001w001 --> gr0001
func (dbw *DbWorkline) AuID() string {
	return dbw.WkUID[:6]
}

// MyAu - get the DbAuthor for this line
func (dbw *DbWorkline) MyAu() *DbAuthor {
	a, ok := AllAuthors[dbw.AuID()]
	if !ok {
		msg(fmt.Sprintf("DbWorkline.MyAu() failed to find '%s'", dbw.AuID()), MSGWARN)
		a = &DbAuthor{}
	}
	return a
}

// WkID - gr0001w001 --> 001
func (dbw *DbWorkline) WkID() string {
	return dbw.WkUID[7:]
}

// MyWk - get the DbWork for this line
func (dbw *DbWorkline) MyWk() *DbWork {
	w, ok := AllWorks[dbw.WkUID]
	if !ok {
		msg(fmt.Sprintf("MyAu() failed to find '%s'", dbw.AuID()), MSGWARN)
		w = &DbWork{}
	}
	return w
}

func (dbw *DbWorkline) FindCorpus() string {
	// gr0001w001 --> gr
	return dbw.WkUID[0:2]
}

func (dbw *DbWorkline) BuildHyperlink() string {
	if len(dbw.WkUID) == 0 {
		// FormatWithContextResults() will trigger this
		msg("BuildHyperlink() on empty dbworkline", MSGTMI)
		return ""
	}
	return fmt.Sprintf(WKLNHYPERLNKTEMPL, dbw.AuID(), dbw.WkID(), dbw.TbIndex)
}

func (dbw *DbWorkline) GatherMetadata() {
	md := make(map[string]string)
	if Metadata.MatchString(dbw.MarkedUp) {
		mm := Metadata.FindAllStringSubmatch(dbw.MarkedUp, -1)
		for _, m := range mm {
			// sample location:
			// hipparchiaDB=# select index, marked_up_line from lt0474 where index = 116946;
			md[m[1]] = m[2]
		}

		dbw.MarkedUp = Metadata.ReplaceAllString(dbw.MarkedUp, "")
		for k, v := range md {
			md[k] = MDFormat.ReplaceAllString(v, WLNMETADATATEMPL)
			if _, y := MDRemap[k]; y {
				md[MDRemap[k]] = md[k]
				delete(md, k)
			}
		}
	}
	dbw.embnotes = md
}

// PurgeMetadata - delete the line Metadata
func (dbw *DbWorkline) PurgeMetadata() {
	if Metadata.MatchString(dbw.MarkedUp) {
		dbw.MarkedUp = Metadata.ReplaceAllString(dbw.MarkedUp, "")
	}
}

// ShowMarkup - reveal markup in a line
func (dbw *DbWorkline) ShowMarkup() string {
	clean := strings.NewReplacer("<", "&lt;", ">", "&gt;")
	return clean.Replace(dbw.MarkedUp)
}

func (dbw *DbWorkline) SameLevelAs(other DbWorkline) bool {
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

func (dbw *DbWorkline) StrippedSlice() []string {
	return strings.Split(dbw.Stripped, " ")
}

func (dbw *DbWorkline) AccentedSlice() []string {
	return strings.Split(dbw.Accented, " ")
}

func (dbw *DbWorkline) MarkedUpSlice() []string {
	cln := NoHTML.ReplaceAllString(dbw.MarkedUp, "")
	return strings.Split(cln, " ")
}

func (dbw *DbWorkline) Citation() string {
	return strings.Join(dbw.FindLocus(), ".")
}

// Lvls - report the number of active levels for this line
func (dbw *DbWorkline) Lvls() int {
	//alternate is: "return dbw.MyWk().CountLevels()"
	vv := []string{dbw.Lvl0Value, dbw.Lvl1Value, dbw.Lvl2Value, dbw.Lvl3Value, dbw.Lvl4Value, dbw.Lvl5Value}
	empty := ContainsN(vv, "-1")
	return 6 - empty
}

func (dbw *DbWorkline) LvlVal(lvl int) string {
	// what is the value at level N?
	switch lvl {
	case 0:
		return dbw.Lvl0Value
	case 1:
		return dbw.Lvl1Value
	case 2:
		return dbw.Lvl2Value
	case 3:
		return dbw.Lvl3Value
	case 4:
		return dbw.Lvl4Value
	case 5:
		return dbw.Lvl5Value
	default:
		return ""
	}
}

func WorklineQuery(prq PrerolledQuery, dbconn *pgxpool.Conn) []DbWorkline {
	if SQLProvider == "pgsql" {
		return PGXWorklineQuery(prq, dbconn)
	} else {
		return SQLITEWorklineQuery(prq)
	}

}

// SQLITEWorklineQuery - use a PrerolledQuery to acquire []DbWorkline
func SQLITEWorklineQuery(prq PrerolledQuery) []DbWorkline {
	ltconn := GetSQLiteConn()
	defer ltconn.Close()

	// todo: prq.TempTable

	rows, err := ltconn.QueryContext(context.Background(), prq.PsqlQuery)
	//msg(prq.PsqlQuery, 3)
	chke(err)

	defer rows.Close()
	var rr []DbWorkline
	for rows.Next() {
		var rw DbWorkline
		// note the order: [1] WkUID, [2] TbIndex and not [1] TbIndex, [2] WkUID
		e := rows.Scan(&rw.WkUID, &rw.TbIndex, &rw.Lvl5Value, &rw.Lvl4Value, &rw.Lvl3Value, &rw.Lvl2Value, &rw.Lvl1Value, &rw.Lvl0Value, &rw.MarkedUp, &rw.Accented, &rw.Stripped, &rw.Hyphenated, &rw.Annotations)
		chke(e)
		rr = append(rr, rw)
	}
	fmt.Println(fmt.Sprintf("SQLITEWorklineQuery() found %d rows", len(rr)))
	return rr
}

// PGXWorklineQuery - use a PrerolledQuery to acquire []DbWorkline
func PGXWorklineQuery(prq PrerolledQuery, dbconn *pgxpool.Conn) []DbWorkline {
	// NB: you have to use a dbconn.Exec() and can't use SQLPool.Exex() because with the latter
	// the temp table will get separated from the main query: ERROR: relation "{ttname}" does not exist (SQLSTATE 42P01)

	// [a] build a temp table if needed

	if prq.TempTable != "" {
		_, err := dbconn.Exec(context.Background(), prq.TempTable)
		chke(err)
	}

	// [b] execute the main query (nb: query needs to satisfy needs of RowToStructByPos in [c])

	foundrows, err := dbconn.Query(context.Background(), prq.PsqlQuery)
	chke(err)

	// [c] convert the finds into []DbWorkline

	thesefinds, err := pgx.CollectRows(foundrows, pgx.RowToStructByPos[DbWorkline])
	chke(err)

	return thesefinds
}

// GrabOneLine - return a single DbWorkline from a table
func GrabOneLine(table string, line int) DbWorkline {
	const (
		QTMPL = `SELECT %s FROM %s WHERE "index" = %d`
	)

	dbconn := GetPSQLconnection()
	defer dbconn.Release()

	var prq PrerolledQuery
	prq.TempTable = ""
	prq.PsqlQuery = fmt.Sprintf(QTMPL, WORLINETEMPLATE, table, line)
	foundlines := WorklineQuery(prq, dbconn)
	if len(foundlines) != 0 {
		// "index = %d" in QTMPL ought to mean you can never have len(foundlines) > 1 because index values are unique
		return foundlines[0]
	} else {
		return DbWorkline{}
	}
}

// SimpleContextGrabber - grab a pile of lines centered around the focusline
func SimpleContextGrabber(table string, focus int, context int) []DbWorkline {
	const (
		QTMPL = `SELECT %s FROM %s WHERE ("index" BETWEEN %d AND %d) ORDER by "index"`
	)

	dbconn := GetPSQLconnection()
	defer dbconn.Release()

	low := focus - context
	high := focus + context

	var prq PrerolledQuery
	prq.TempTable = ""
	prq.PsqlQuery = fmt.Sprintf(QTMPL, WORLINETEMPLATE, table, low, high)
	foundlines := WorklineQuery(prq, dbconn)

	for _, ln := range foundlines {
		fmt.Println(ln.BuildHyperlink())
	}
	return foundlines
}

// findvalidlevelvalues - tell me some of a citation and I can tell you what is a valid choice at the next step
func findvalidlevelvalues(wkid string, locc []string) LevelValues {
	// curl localhost:5000/get/json/workstructure/lt0959/001
	// {"totallevels": 3, "level": 2, "label": "book", "low": "1", "high": "3", "range": ["1", "2", "3"]}
	// curl localhost:5000/get/json/workstructure/lt0959/001/2
	// {"totallevels": 3, "level": 1, "label": "poem", "low": "1", "high": "19", "range": ["1", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "2", "3", "4", "5", "6", "7", "8", "9a", "9b"]}

	// select levellabels_00, levellabels_01, levellabels_02, levellabels_03, levellabels_04, levellabels_05 from works where universalid = 'lt0959w001';
	// levellabels_00 | levellabels_01 | levellabels_02 | levellabels_03 | levellabels_04 | levellabels_05
	//----------------+----------------+----------------+----------------+----------------+----------------
	// verse          | poem           | book           |                |                |

	const (
		SEL    = SELECTFROM + ` WHERE wkuniversalid='%s' %s %s ORDER BY "index" ASC`
		ANDNOT = `AND %s NOT IN ('t')`
	)

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
		// logic bug in here somewhere...
		// FAIL = "findvalidlevelvalues() sent negative levels"
		// msg(FAIL, MSGWARN)
		return LevelValues{}
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

	var prq PrerolledQuery
	prq.PsqlQuery = fmt.Sprintf(SEL, w.AuID(), wkid, and, andnot)

	dbconn := GetPSQLconnection()
	defer dbconn.Release()
	lines := WorklineQuery(prq, dbconn)

	// [c] extract info from the hitlines returned
	var vals LevelValues
	vals.AtLvl = atlvl
	vals.Label = lmap[atlvl]

	if len(lines) == 0 {
		return vals
	}

	vals.Total = lines[0].Lvls()
	vals.Low = lines[0].LvlVal(atlvl)
	vals.High = lines[len(lines)-1].LvlVal(atlvl)
	var r []string
	for i := range lines {
		r = append(r, lines[i].LvlVal(atlvl))
	}
	r = Unique(r)
	sort.Strings(r)
	vals.Range = r

	return vals
}
