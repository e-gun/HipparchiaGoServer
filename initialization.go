package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

var (
	AllWorks   = workmapper()
	AllAuthors = loadworksintoauthors(authormapper(), AllWorks)
	// AllLemm    = lemmamapper()
)

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
	// not in the DB, but derived: gr2017w068 --> 068
	WorkNum string
}

func (dbw DbWork) FindWorknumber() string {
	// ex: gr2017w068
	return dbw.UID[7:]
}

func (dbw DbWork) FindAuthor() string {
	// ex: gr2017w068
	return dbw.UID[:6]
}

func (dbw DbWork) CitationFormat() []string {
	cf := []string{
		dbw.LL5,
		dbw.LL4,
		dbw.LL3,
		dbw.LL2,
		dbw.LL1,
		dbw.LL0,
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

func (dbw DbWork) DateInRange(b int64, a int64) bool {
	if b <= dbw.ConvDate && dbw.ConvDate <= a {
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

// all functions in here should be run in order to prepare the core data

// workmapper - build a map of all works keyed to the authorUID: map[string]DbWork
func workmapper() map[string]DbWork {
	start := time.Now()
	previous := time.Now()
	dbpool := grabpgsqlconnection()
	qt := "SELECT %s FROM works"
	q := fmt.Sprintf(qt, WORKTEMPLATE)

	foundrows, err := dbpool.Query(context.Background(), q)
	checkerror(err)

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
		checkerror(err)
		thefinds = append(thefinds, thehit)
	}

	for _, val := range thefinds {
		val.WorkNum = val.FindWorknumber()
	}

	workmap := make(map[string]DbWork)
	for _, val := range thefinds {
		workmap[val.UID] = val
	}
	timetracker("A", "works built: map[string]DbWork", start, previous)
	return workmap

}

// authormapper - build a map of all authors keyed to the authorUID: map[string]DbAuthor
func authormapper() map[string]DbAuthor {
	start := time.Now()
	previous := time.Now()

	dbpool := grabpgsqlconnection()
	qt := "SELECT %s FROM authors ORDER by universalid ASC"
	q := fmt.Sprintf(qt, AUTHORTEMPLATE)

	foundrows, err := dbpool.Query(context.Background(), q)
	checkerror(err)

	var thefinds []DbAuthor

	defer foundrows.Close()
	for foundrows.Next() {
		// fmt.Println(foundrows.Values())
		// this will die if <nil> comes back inside any of the columns: "cannot scan null into *string"
		// the builder should address this: fixing it here is less ideal
		var thehit DbAuthor
		err := foundrows.Scan(&thehit.UID, &thehit.Language, &thehit.IDXname, &thehit.Name, &thehit.Shortname,
			&thehit.Cleaname, &thehit.Genres, &thehit.RecDate, &thehit.ConvDate, &thehit.Location)
		checkerror(err)
		thefinds = append(thefinds, thehit)
	}

	authormap := make(map[string]DbAuthor)
	for _, val := range thefinds {
		authormap[val.UID] = val
	}
	timetracker("B", "authors built: map[string]DbAuthor", start, previous)
	return authormap

}

// loadworksintoauthors - load all works in the workmap into the authormap WorkList
func loadworksintoauthors(aa map[string]DbAuthor, ww map[string]DbWork) map[string]DbAuthor {
	start := time.Now()
	previous := time.Now()
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
	keys := make([]string, 0, len(aa))
	for _, a := range aa {
		keys = append(keys, a.UID)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	// fmt.Printf("wl %d vs al %d", len(worklists), len(keys))  : wl 3455 vs al 3455

	// [2b] build a *slice* of []DbAuthor since we can's modify a.WorkList in a map version
	asl := make([]DbAuthor, 0, len(keys))
	for _, k := range keys {
		asl = append(asl, aa[k])
	}

	// [2c] add the worklists to the slice
	for i, _ := range keys {
		asl[i].WorkList = worklists[asl[i].UID]
	}

	// [3] convert slice to map
	na := make(map[string]DbAuthor)
	for i, a := range asl {
		na[a.UID] = asl[i]
	}
	timetracker("C", "works loaded into authors: map[string]DbAuthor", start, previous)
	return na
}

func lemmamapper() map[string]map[string]DbLemma {
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
	// nest a map: [HGS] [-: 2.497s][Δ: 2.497s] lemma

	start := time.Now()
	previous := time.Now()

	nested := make(map[string]map[string]DbLemma)
	unnested := make(map[string]DbLemma)

	langs := [2]string{"greek", "latin"}
	t := `SELECT dictionary_entry, xref_number, derivative_forms FROM %s_lemmata`

	dbpool := grabpgsqlconnection()
	var thefinds []DbLemma
	for _, lg := range langs {
		q := fmt.Sprintf(t, lg)
		foundrows, err := dbpool.Query(context.Background(), q)
		checkerror(err)
		defer foundrows.Close()
		for foundrows.Next() {
			var thehit DbLemma
			err := foundrows.Scan(&thehit.Entry, &thehit.Xref, &thehit.Deriv)
			checkerror(err)
			thefinds = append(thefinds, thehit)
		}
	}

	for _, lm := range thefinds {
		unnested[lm.Entry] = lm
	}

	for k, v := range unnested {
		bag := string([]rune(v.Entry)[0:2])
		bag = stripaccents(bag)
		bag = strings.ToLower(bag)
		bag = strings.Replace(bag, "j", "i", -1)
		bag = strings.Replace(bag, "v", "u", -1)
		if _, y := nested[bag]; !y {
			nested[bag] = make(map[string]DbLemma)
		} else {
			nested[bag][k] = v
		}
	}

	//fmt.Println("lemmata count")
	//fmt.Println(len(nested))
	//fmt.Println(nested["ζω"])

	timetracker("D", "lemma map built", start, previous)
	return nested
}
