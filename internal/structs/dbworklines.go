//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package structs

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/generic"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
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
	MarkedUp    string // converting this and others to pointers will not save you memory
	Accented    string // converting to pointers might give you a very slight speed boost
	Stripped    string // converting to pointers can produce nil pointer problems that need constant checks
	Hyphenated  string
	Annotations string
	// beyond the db stuff; do not make this "public": pgx.RowToStructByPos will balk
	embnotes map[string]string
}

func (dbw *DbWorkline) GetNotes() map[string]string {
	return dbw.embnotes
}

func (dbw *DbWorkline) FindLocus() []string {
	loc := [vv.NUMBEROFCITATIONLEVELS]string{
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
	return dbw.WkUID[:vv.LENGTHOFAUTHORID]
}

// MyAu - get the DbAuthor for this line
func (dbw *DbWorkline) MyAu() *DbAuthor {
	a, ok := AllAuthors[dbw.AuID()]
	if !ok {
		msg.WARN(fmt.Sprintf("DbWorkline.MyAu() failed to find '%s'", dbw.AuID()))
		a = &DbAuthor{}
	}
	return a
}

// WkID - gr0001w001 --> 001
func (dbw *DbWorkline) WkID() string {
	return dbw.WkUID[vv.LENGTHOFAUTHORID+1:]
}

// MyWk - get the DbWork for this line
func (dbw *DbWorkline) MyWk() *DbWork {
	w, ok := AllWorks[dbw.WkUID]
	if !ok {
		msg.WARN(fmt.Sprintf("MyAu() failed to find '%s'", dbw.AuID()))
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
		msg.TMI("BuildHyperlink() on empty dbworkline")
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
	vl := []string{dbw.Lvl0Value, dbw.Lvl1Value, dbw.Lvl2Value, dbw.Lvl3Value, dbw.Lvl4Value, dbw.Lvl5Value}
	empty := generic.ContainsN(vl, "-1")
	return vv.NUMBEROFCITATIONLEVELS - empty
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

// YieldAll - don't copy everything at once; send everything over a channel
func (wlb *WorkLineBundle) YieldAll() chan DbWorkline {
	// assuming the receiver will grab everything
	// the code is always of the following format: `ll = wlb.YieldAll()` + `for l := range ll { ... }`

	// a YieldSome() is not yet needed: yield some but listen on a stop channel, etc.

	msg.TMI(fmt.Sprintf("WorkLineBundle.YieldAll() sending %d lines", wlb.Len()))

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
// SORTING: https://pkg.go.dev/sort#example__sortMultiKeys
//

type WLLessFunc func(p1, p2 *DbWorkline) bool

// WLMultiSorter implements the Sort interface, sorting the changes within.
type WLMultiSorter struct {
	changes []DbWorkline
	less    []WLLessFunc
}

// Sort sorts the argument slice according to the less functions passed to WLOrderedBy.
func (ms *WLMultiSorter) Sort(changes []DbWorkline) {
	ms.changes = changes
	sort.Sort(ms)
}

// WLOrderedBy returns a Sorter that sorts using the less functions, in order.
// Call its Sort method to sort the data.
func WLOrderedBy(less ...WLLessFunc) *WLMultiSorter {
	return &WLMultiSorter{
		less: less,
	}
}

// Len is part of sort.Interface.
func (ms *WLMultiSorter) Len() int {
	return len(ms.changes)
}

// Swap is part of sort.Interface.
func (ms *WLMultiSorter) Swap(i, j int) {
	ms.changes[i], ms.changes[j] = ms.changes[j], ms.changes[i]
}

func (ms *WLMultiSorter) Less(i, j int) bool {
	p, q := &ms.changes[i], &ms.changes[j]
	// Try all but the last comparison.
	var k int
	for k = 0; k < len(ms.less)-1; k++ {
		less := ms.less[k]
		switch {
		case less(p, q):
			// p < q, so we have a decision.
			return true
		case less(q, p):
			// p > q, so we have a decision.
			return false
		}
		// p == q; try the next comparison.
	}
	// All comparisons to here said "equal", so just return whatever
	// the final comparison reports.
	return ms.less[k](p, q)
}
