//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	// hipparchiaDB=# select * from authors limit 0;
	// universalid | language | idxname | akaname | shortname | cleanname | genres | recorded_date | converted_date | location
	//-------------+----------+---------+---------+-----------+-----------+--------+---------------+----------------+----------

	AUTHORTEMPLATE = ` universalid, language, idxname, akaname, shortname, cleanname, genres, recorded_date, converted_date, location `

	// hipparchiaDB=# select * from works limit 0;
	// universalid | title | language | publication_info | levellabels_00 | levellabels_01 | levellabels_02 | levellabels_03 | levellabels_04 | levellabels_05 | workgenre | transmission | worktype | provenance | recorded_date | converted_date | wordcount | firstline | lastline | authentic
	//-------------+-------+----------+------------------+----------------+----------------+----------------+----------------+----------------+----------------+-----------+--------------+----------+------------+---------------+----------------+-----------+-----------+----------+-----------
	//(0 rows)

	WORKTEMPLATE = ` universalid, title, language, publication_info,
		levellabels_00, levellabels_01, levellabels_02, levellabels_03, levellabels_04, levellabels_05,
		workgenre, transmission, worktype, provenance, recorded_date, converted_date, wordcount,
		firstline, lastline, authentic`
)

var (
	// order matters
	cfg         CurrentConfiguration
	sessions    = make(map[string]ServerSession)
	searches    = make(map[string]SearchStruct)
	proghits    = sync.Map{}
	progremain  = sync.Map{}
	AllWorks    = make(map[string]DbWork)
	AllAuthors  = make(map[string]DbAuthor)
	AllLemm     = make(map[string]DbLemma)
	NestedLemm  = make(map[string]map[string]DbLemma)
	WkCorpusMap = make(map[string][]string)
	AuCorpusMap = make(map[string][]string)
	AuGenres    = make(map[string]bool)
	WkGenres    = make(map[string]bool)
	AuLocs      = make(map[string]bool)
	WkLocs      = make(map[string]bool)
)

type CurrentConfiguration struct {
	WorkerCount int
	LogLevel    int
	PSQL        string
	EchoLog     int // "none", "terse", "verbose"
	PGLogin     PostgresLogin
}

type DbAuthor struct {
	UID       string
	Language  string
	IDXname   string
	Name      string
	Shortname string
	Cleaname  string
	Genres    string
	RecDate   string
	ConvDate  int64
	Location  string
	// beyond the DB starts here
	WorkList []string
}

func (dba DbAuthor) AddWork(w string) {
	dba.WorkList = append(dba.WorkList, w)
}

type DbWork struct {
	UID       string
	Title     string
	Language  string
	Pub       string
	LL0       string
	LL1       string
	LL2       string
	LL3       string
	LL4       string
	LL5       string
	Genre     string
	Xmit      string
	Type      string
	Prov      string
	RecDate   string
	ConvDate  int64
	WdCount   int64
	FirstLine int64
	LastLine  int64
	Authentic bool
}

func (dbw DbWork) FindWorknumber() string {
	// ex: gr2017w068
	return dbw.UID[7:]
}

func (dbw DbWork) FindAuthor() string {
	// ex: gr2017w068
	if len(dbw.UID) < 6 {
		return ""
	} else {
		return dbw.UID[:6]
	}
}

func (dbw DbWork) CitationFormat() []string {
	cf := []string{
		dbw.LL5, dbw.LL4, dbw.LL3, dbw.LL2, dbw.LL1, dbw.LL0,
	}
	return cf
}

func (dbw DbWork) CountLevels() int {
	ll := 0
	for _, l := range []string{dbw.LL5, dbw.LL4, dbw.LL3, dbw.LL2, dbw.LL1, dbw.LL0} {
		if len(l) > 0 {
			ll += 1
		}
	}
	return ll
}

func (dbw DbWork) DateInRange(earliest int64, latest int64) bool {
	if earliest >= dbw.ConvDate && dbw.ConvDate <= latest {
		return true
	} else {
		return false
	}
}

type DbLemma struct {
	// dictionary_entry | xref_number |    derivative_forms
	Entry string
	Xref  int64
	Deriv []string
}

func (dbl DbLemma) EntryRune() []rune {
	return []rune(dbl.Entry)
}

type SearchSummary struct {
	Time time.Time
	Sum  string
}

// all functions in here should be run in order to prepare the core data

// workmapper - build a map of all works keyed to the authorUID: map[string]DbWork
func workmapper() map[string]DbWork {
	dbpool := GetPSQLconnection()
	defer dbpool.Close()
	qt := "SELECT %s FROM works"
	q := fmt.Sprintf(qt, WORKTEMPLATE)

	foundrows, err := dbpool.Query(context.Background(), q)
	chke(err)

	var thefinds []DbWork

	defer foundrows.Close()
	for foundrows.Next() {
		// fmt.Println(foundrows.Values())
		// this will die if <nil> comes back inside any of the columns
		var thehit DbWork
		err := foundrows.Scan(&thehit.UID, &thehit.Title, &thehit.Language, &thehit.Pub, &thehit.LL0,
			&thehit.LL1, &thehit.LL2, &thehit.LL3, &thehit.LL4, &thehit.LL5, &thehit.Genre,
			&thehit.Xmit, &thehit.Type, &thehit.Prov, &thehit.RecDate, &thehit.ConvDate, &thehit.WdCount,
			&thehit.FirstLine, &thehit.LastLine, &thehit.Authentic)
		chke(err)
		thefinds = append(thefinds, thehit)
	}

	workmap := make(map[string]DbWork, DBWKMAPSIZE)
	for _, val := range thefinds {
		workmap[val.UID] = val
	}
	return workmap
}

// authormapper - build a map of all authors keyed to the authorUID: map[string]DbAuthor
func authormapper() map[string]DbAuthor {
	dbpool := GetPSQLconnection()
	defer dbpool.Close()
	qt := "SELECT %s FROM authors ORDER by universalid ASC"
	q := fmt.Sprintf(qt, AUTHORTEMPLATE)

	foundrows, err := dbpool.Query(context.Background(), q)
	chke(err)

	var thefinds []DbAuthor

	defer foundrows.Close()
	for foundrows.Next() {
		// fmt.Println(foundrows.Values())
		// this will die if <nil> comes back inside any of the columns: "cannot scan null into *string"
		// the builder should address this: fixing it here is less ideal
		var thehit DbAuthor
		err := foundrows.Scan(&thehit.UID, &thehit.Language, &thehit.IDXname, &thehit.Name, &thehit.Shortname,
			&thehit.Cleaname, &thehit.Genres, &thehit.RecDate, &thehit.ConvDate, &thehit.Location)
		chke(err)
		thefinds = append(thefinds, thehit)
	}

	authormap := make(map[string]DbAuthor, DBAUMAPSIZE)
	for _, val := range thefinds {
		authormap[val.UID] = val
	}
	return authormap
}

// loadworksintoauthors - load all works in the workmap into the authormap WorkList
func loadworksintoauthors(aa map[string]DbAuthor, ww map[string]DbWork) map[string]DbAuthor {
	// https://stackoverflow.com/questions/32751537/why-do-i-get-a-cannot-assign-error-when-setting-value-to-a-struct-as-a-value-i
	// the following does not work: aa[a].WorkList = append(aa[w.FindAuthor()].WorkList, w.UID)
	// that means you have to rebuild the damn authormap unless you want to use pointers in DbAuthor: itself a hassle

	// [1] build a map of {UID: WORKLIST...}
	worklists := make(map[string][]string)
	for _, w := range ww {
		wk := w.UID
		au := wk[0:6]
		if _, y := worklists[au]; !y {
			worklists[au] = []string{wk}
		} else {
			worklists[au] = append(worklists[au], wk)
		}
	}

	// [2] decompose aa and rebuild but this time be in possession of all relevant data...
	// [2a] find all keys and sort them
	keys := make([]string, len(aa))
	ct := 0
	for _, a := range aa {
		keys[ct] = a.UID
		ct += 1
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	// fmt.Printf("wl %d vs al %d", len(worklists), len(keys))  : wl 3455 vs al 3455

	// [2b] build a *slice* of []DbAuthor since we can's modify a.WorkList in a map VERSION
	asl := make([]DbAuthor, len(keys))
	ct = 0
	for _, k := range keys {
		asl[ct] = aa[k]
		ct += 1
	}

	// [2c] add the worklists to the slice
	for i, _ := range keys {
		asl[i].WorkList = worklists[asl[i].UID]
	}

	// [3] convert slice to map
	na := make(map[string]DbAuthor, DBAUMAPSIZE)
	for i, a := range asl {
		na[a.UID] = asl[i]
	}
	return na
}

// lemmamapper - map[string]DbLemma for all lemmata
func lemmamapper() map[string]DbLemma {
	// hipparchiaDB=# \d greek_lemmata
	//                       Table "public.greek_lemmata"
	//      Column      |         Type          | Collation | Nullable | Default
	//------------------+-----------------------+-----------+----------+---------
	// dictionary_entry | character varying(64) |           |          |
	// xref_number      | integer               |           |          |
	// derivative_forms | text[]                |           |          |
	//Indexes:
	//    "greek_lemmata_idx" btree (dictionary_entry)

	// a list of 140k words is too long to send to 'getlemmahint' without offering quicker access
	// [HGS] [D: 0.199s][Δ: 0.199s] unnested lemma map built

	unnested := make(map[string]DbLemma, DBLMMAPSIZE)

	langs := [2]string{"greek", "latin"}
	t := `SELECT dictionary_entry, xref_number, derivative_forms FROM %s_lemmata`

	dbpool := GetPSQLconnection()
	defer dbpool.Close()

	var thefinds []DbLemma
	for _, lg := range langs {
		q := fmt.Sprintf(t, lg)
		foundrows, err := dbpool.Query(context.Background(), q)
		chke(err)
		defer foundrows.Close()
		for foundrows.Next() {
			var thehit DbLemma
			err := foundrows.Scan(&thehit.Entry, &thehit.Xref, &thehit.Deriv)
			chke(err)
			thefinds = append(thefinds, thehit)
		}
	}
	clean := strings.NewReplacer("-", "", "¹", "", "²", "", "³", "", "j", "i", "v", "u")

	for _, lm := range thefinds {
		cl := clean.Replace(lm.Entry)
		lm.Entry = cl
		unnested[cl] = lm
	}

	// fmt.Println(unnested["dorsum"])
	// {dorsum 24563373 [dorsum dorsone dorsa dorsoque dorso dorsoue dorsis dorsi dorsisque dorsumque]}
	return unnested
}

// nestedlemmamapper - map[string]map[string]DbLemma for the hinter
func nestedlemmamapper(unnested map[string]DbLemma) map[string]map[string]DbLemma {
	// you need both a nested and the unnested VERSION; nested for the hinter
	// [HGS] [E: 2.284s][Δ: 2.284s] nested lemma map built
	nested := make(map[string]map[string]DbLemma)
	for k, v := range unnested {
		bag := string([]rune(v.Entry)[0:2])
		bag = stripaccentsSTR(bag)
		bag = strings.ToLower(bag)
		bag = strings.Replace(bag, "j", "i", -1)
		bag = strings.Replace(bag, "v", "u", -1)
		if _, y := nested[bag]; !y {
			nested[bag] = make(map[string]DbLemma)
		} else {
			nested[bag][k] = v
		}
	}
	return nested
}

func buildwkcorpusmap() map[string][]string {
	// sessionintosearchlist() could just grab a pre-rolled list instead of calculating every time...
	wkcorpusmap := make(map[string][]string)
	corp := [5]string{"gr", "lt", "in", "ch", "dp"}
	for _, w := range AllWorks {
		for _, c := range corp {
			if w.UID[0:2] == c {
				wkcorpusmap[c] = append(wkcorpusmap[c], w.UID)
			}
		}
	}
	return wkcorpusmap
}

func buildaucorpusmap() map[string][]string {
	// sessionintosearchlist() could just grab a pre-rolled list instead of calculating every time...
	aucorpusmap := make(map[string][]string)
	corp := [5]string{"gr", "lt", "in", "ch", "dp"}
	for _, a := range AllAuthors {
		for _, c := range corp {
			if a.UID[0:2] == c {
				aucorpusmap[c] = append(aucorpusmap[c], a.UID)
			}
		}
	}
	return aucorpusmap
}

func buildaugenresmap() map[string]bool {
	genres := make(map[string]bool)
	for _, a := range AllAuthors {
		gg := strings.Split(a.Genres, ",")
		for _, g := range gg {
			genres[g] = true
		}
	}
	return genres
}

func buildwkgenresmap() map[string]bool {
	genres := make(map[string]bool)
	for _, w := range AllWorks {
		genres[w.Genre] = true
	}
	return genres
}

func buildaulocationmap() map[string]bool {
	locations := make(map[string]bool)
	for _, a := range AllAuthors {
		ll := strings.Split(a.Location, ",")
		for _, l := range ll {
			locations[l] = true
		}
	}
	return locations
}

func buildwklocationmap() map[string]bool {
	locations := make(map[string]bool)
	for _, w := range AllWorks {
		locations[w.Prov] = true
	}
	return locations
}
