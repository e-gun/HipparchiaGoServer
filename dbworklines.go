//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"regexp"
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

type WorkLineBundle struct {
	Lines []DbWorkline
}

// Generate - don't copy everything at once; send it over a channel
func (wlb *WorkLineBundle) Generate() chan DbWorkline {
	c := make(chan DbWorkline)
	go func() {
		for i := 0; i < len(wlb.Lines); i++ {
			c <- wlb.Lines[i]
		}
		close(c)
	}()
	return c
}

func (wlb *WorkLineBundle) ResizeTo(i int) {
	if i < len(wlb.Lines) {
		wlb.Lines = wlb.Lines[0:i]
	}
}

func (wlb *WorkLineBundle) Len() int {
	return len(wlb.Lines)
}

func (wlb *WorkLineBundle) IsEmpty() bool {
	if len(wlb.Lines) == 0 {
		return true
	} else {
		return false
	}
}

func (wlb *WorkLineBundle) FirstLine() DbWorkline {
	if len(wlb.Lines) != 0 {
		return wlb.Lines[0]
	} else {
		return DbWorkline{}
	}
}

func (wlb *WorkLineBundle) AppendLines(toadd []DbWorkline) {
	wlb.Lines = append(wlb.Lines, toadd...)
}

func (wlb *WorkLineBundle) AppendOne(toadd DbWorkline) {
	wlb.Lines = append(wlb.Lines, toadd)
}

//
// QUERY FUNCTIONS
//

// WorklineQuery - use a PrerolledQuery to acquire a WorkLineBundle
func WorklineQuery(prq PrerolledQuery, dbconn *pgxpool.Conn) WorkLineBundle {
	// NB: you have to use a dbconn.Exec() and can't use SQLPool.Exex() because with the latter the temp table will
	// get separated from the main query:
	// ERROR: relation "{ttname}" does not exist (SQLSTATE 42P01)

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

	return WorkLineBundle{Lines: thesefinds}
}

// GrabOneLine - return a single DbWorkline from a table
func GrabOneLine(table string, line int) DbWorkline {
	const (
		QTMPL = "SELECT %s FROM %s WHERE index = %d"
	)

	dbconn := GetDBConnection()
	defer dbconn.Release()

	var prq PrerolledQuery
	prq.TempTable = ""
	prq.PsqlQuery = fmt.Sprintf(QTMPL, WORLINETEMPLATE, table, line)
	foundlines := WorklineQuery(prq, dbconn)
	if foundlines.Len() != 0 {
		// "index = %d" in QTMPL ought to mean you can never have len(foundlines) > 1 because index values are unique
		return foundlines.FirstLine()
	} else {
		return DbWorkline{}
	}
}

// SimpleContextGrabber - grab a pile of Lines centered around the focusline
func SimpleContextGrabber(table string, focus int, context int) WorkLineBundle {
	const (
		QTMPL = "SELECT %s FROM %s WHERE (index BETWEEN %d AND %d) ORDER by index"
	)

	dbconn := GetDBConnection()
	defer dbconn.Release()

	low := focus - context
	high := focus + context

	var prq PrerolledQuery
	prq.TempTable = ""
	prq.PsqlQuery = fmt.Sprintf(QTMPL, WORLINETEMPLATE, table, low, high)

	foundlines := WorklineQuery(prq, dbconn)

	return foundlines
}
