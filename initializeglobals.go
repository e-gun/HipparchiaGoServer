//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"sync"
)

const (
	AUTHORTEMPLATE = ` universalid, language, idxname, akaname, shortname, cleanname, genres, recorded_date, converted_date, location `

	WORKTEMPLATE = ` universalid, title, language, publication_info,
		levellabels_00, levellabels_01, levellabels_02, levellabels_03, levellabels_04, levellabels_05,
		workgenre, transmission, worktype, provenance, recorded_date, converted_date, wordcount,
		firstline, lastline, authentic`
)

// The maps below are all race condition candidates, but most of them are write once, read many. It does not seem possible
// to race on read. 20 goroutines reading the same static map simultaneously can each get to 3,540,000,000 reads without
// triggering a race.

// SessionMap and SearchMap receive infrequent additions and deletions. SearchLocker should be locking these but that does
// not mean there can never be a race: 20 goroutines reading the same L/U updated map simultaneously as fast as they can
// are each able get to >100,000,000 on average if you update the map every .1 seconds. [without L/U writes, this will
// die right away]

// You will seemingly never race if you L/U all writes and RL/RU all reads. This is astonishingly slow: good luck
// getting to 100,000,000 reads this way to confirm that you can get to 100x that number...

// So, the current policy here is: try to catch all potential race conditions. Writes are higher priority. Reads lower.
// "Correctiness" has something to be said for it. But note that the races are in practice a "small" and "technical"
// complaint. HipparchiaGoServer is not supposed to be exposed to everyone, everywhere all the time. Someone requesting
// 100 searches is a bigger worry than any race on 10m requests condition. There needs to be a fairly high degree of user
// trust to begin with.

var (
	Config         CurrentConfiguration
	SQLPool        *pgxpool.Pool
	SessionMap     = make(map[string]ServerSession)
	SearchMap      = make(map[string]SearchStruct)
	AuthorizedMap  = make(map[string]bool)
	UserPassPairs  = make(map[string]string)
	AllWorks       = make(map[string]DbWork)
	AllAuthors     = make(map[string]DbAuthor)
	AllLemm        = make(map[string]DbLemma)
	NestedLemm     = make(map[string]map[string]DbLemma)
	WkCorpusMap    = make(map[string][]string)
	AuCorpusMap    = make(map[string][]string)
	AuGenres       = make(map[string]bool)
	WkGenres       = make(map[string]bool)
	AuLocs         = make(map[string]bool)
	WkLocs         = make(map[string]bool)
	TheCorpora     = [5]string{"gr", "lt", "in", "ch", "dp"}
	TheLanguages   = [2]string{"greek", "latin"}
	SearchLocker   sync.RWMutex
	SessionLocker  sync.RWMutex
	AuthorizLocker sync.RWMutex
	WebsocketPool  = WSFillNewPool()
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
	ConvDate  int
	Location  string
	// beyond the DB starts here
	WorkList []string
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
	ConvDate  int
	WdCount   int
	FirstLine int
	LastLine  int
	Authentic bool
}

// WkID - ex: gr2017w068 --> 068
func (dbw DbWork) WkID() string {
	return dbw.UID[7:]
}

// AuID - ex: gr2017w068 --> gr2017
func (dbw DbWork) AuID() string {
	if len(dbw.UID) < 6 {
		return ""
	} else {
		return dbw.UID[:6]
	}
}

// MyAu - return the work's DbAuthor
func (dbw DbWork) MyAu() DbAuthor {
	a, ok := AllAuthors[dbw.AuID()]
	if !ok {
		msg(fmt.Sprintf("DbWork.MyAu() failed to find '%s'", dbw.AuID()), 1)
		a = DbAuthor{}
	}
	return a
}

func (dbw DbWork) CitationFormat() []string {
	cf := []string{
		dbw.LL5, dbw.LL4, dbw.LL3, dbw.LL2, dbw.LL1, dbw.LL0,
	}
	return cf
}

// CountLevels - the work structure employs how many levels?
func (dbw DbWork) CountLevels() int {
	ll := 0
	for _, l := range []string{dbw.LL5, dbw.LL4, dbw.LL3, dbw.LL2, dbw.LL1, dbw.LL0} {
		if len(l) > 0 {
			ll += 1
		}
	}
	return ll
}

// DateInRange - is the work dated between X and Y?
func (dbw DbWork) DateInRange(earliest int, latest int) bool {
	if earliest >= dbw.ConvDate && dbw.ConvDate <= latest {
		return true
	} else {
		return false
	}
}

type DbLemma struct {
	// dictionary_entry | xref_number |    derivative_forms
	Entry string
	Xref  int
	Deriv []string
}

func (dbl DbLemma) EntryRune() []rune {
	return []rune(dbl.Entry)
}

// all functions in here should be run in order to prepare the core data

// workmapper - build a map of all works keyed to the authorUID: map[string]DbWork
func workmapper() map[string]DbWork {
	// this is far and away the "heaviest" bit of the whole program:
	// Total: 265.13MB
	// 78.86MB   103.62MB (flat, cum) 39.08% of Total

	// hipparchiaDB-# \d works
	//                            Table "public.works"
	//      Column      |          Type          | Collation | Nullable | Default
	//------------------+------------------------+-----------+----------+---------
	// universalid      | character(10)          |           |          |
	// title            | character varying(512) |           |          |
	// language         | character varying(10)  |           |          |
	// publication_info | text                   |           |          |
	// levellabels_00   | character varying(64)  |           |          |
	// levellabels_01   | character varying(64)  |           |          |
	// levellabels_02   | character varying(64)  |           |          |
	// levellabels_03   | character varying(64)  |           |          |
	// levellabels_04   | character varying(64)  |           |          |
	// levellabels_05   | character varying(64)  |           |          |
	// workgenre        | character varying(32)  |           |          |
	// transmission     | character varying(32)  |           |          |
	// worktype         | character varying(32)  |           |          |
	// provenance       | character varying(64)  |           |          |
	// recorded_date    | character varying(64)  |           |          |
	// converted_date   | integer                |           |          |
	// wordcount        | integer                |           |          |
	// firstline        | integer                |           |          |
	// lastline         | integer                |           |          |
	// authentic        | boolean                |           |          |

	dbconn := GetPSQLconnection()
	defer dbconn.Release()
	qt := "SELECT %s FROM works"
	q := fmt.Sprintf(qt, WORKTEMPLATE)

	foundrows, err := dbconn.Query(context.Background(), q)
	chke(err)

	workmap := make(map[string]DbWork, DBWKMAPSIZE)

	// THE EVEN MORE RAM-INTENSIVE SOLUTION...

	//works, err := pgx.CollectRows(foundrows, pgx.RowToStructByPos[DbWork])
	//chke(err)
	//
	//for i := 0; i < len(works); i++ {
	//	workmap[works[i].UID] = works[i]
	//}

	defer foundrows.Close()
	for foundrows.Next() {
		// this will die if <nil> comes back inside any of the columns; and so you have to use builds from HipparchiaBuilder 1.6.0+
		var w DbWork
		e := foundrows.Scan(&w.UID, &w.Title, &w.Language, &w.Pub, &w.LL0,
			&w.LL1, &w.LL2, &w.LL3, &w.LL4, &w.LL5, &w.Genre,
			&w.Xmit, &w.Type, &w.Prov, &w.RecDate, &w.ConvDate, &w.WdCount,
			&w.FirstLine, &w.LastLine, &w.Authentic)
		chke(e)
		workmap[w.UID] = w
	}
	return workmap
}

// authormapper - build a map of all authors keyed to the authorUID: map[string]DbAuthor
func authormapper(ww map[string]DbWork) map[string]DbAuthor {
	//  5.26MB     5.80MB (flat, cum)  2.19% of Total

	// hipparchiaDB-# \d authors
	//                          Table "public.authors"
	//     Column     |          Type          | Collation | Nullable | Default
	//----------------+------------------------+-----------+----------+---------
	// universalid    | character(6)           |           |          |
	// language       | character varying(10)  |           |          |
	// idxname        | character varying(128) |           |          |
	// akaname        | character varying(128) |           |          |
	// shortname      | character varying(128) |           |          |
	// cleanname      | character varying(128) |           |          |
	// genres         | character varying(512) |           |          |
	// recorded_date  | character varying(64)  |           |          |
	// converted_date | integer                |           |          |
	// location       | character varying(128) |           |          |

	// need to be ready to load the worklists into the authors
	// so: build a map of {UID: WORKLIST...}
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

	dbconn := GetPSQLconnection()
	defer dbconn.Release()
	qt := "SELECT %s FROM authors ORDER by universalid ASC"
	q := fmt.Sprintf(qt, AUTHORTEMPLATE)

	foundrows, err := dbconn.Query(context.Background(), q)
	chke(err)

	authormap := make(map[string]DbAuthor, DBAUMAPSIZE)
	defer foundrows.Close()
	for foundrows.Next() {
		// this will die if <nil> comes back inside any of the columns; and so you have to use builds from HipparchiaBuilder 1.6.0+
		var a DbAuthor
		e := foundrows.Scan(&a.UID, &a.Language, &a.IDXname, &a.Name, &a.Shortname,
			&a.Cleaname, &a.Genres, &a.RecDate, &a.ConvDate, &a.Location)
		chke(e)
		a.WorkList = worklists[a.UID]
		authormap[a.UID] = a
	}

	return authormap
}

// lemmamapper - map[string]DbLemma for all lemmata
func lemmamapper() map[string]DbLemma {
	// example: {dorsum 24563373 [dorsum dorsone dorsa dorsoque dorso dorsoue dorsis dorsi dorsisque dorsumque]}

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

	t := `SELECT dictionary_entry, xref_number, derivative_forms FROM %s_lemmata`

	dbconn := GetPSQLconnection()
	defer dbconn.Release()

	// note that the v --> u here will push us to stripped_line SearchMap instead of accented_line
	// clean := strings.NewReplacer("-", "", "¹", "", "²", "", "³", "", "j", "i", "v", "u")
	clean := strings.NewReplacer("-", "", "j", "i", "v", "u")

	unnested := make(map[string]DbLemma, DBLMMAPSIZE)
	for _, lg := range TheLanguages {
		q := fmt.Sprintf(t, lg)
		foundrows, err := dbconn.Query(context.Background(), q)
		chke(err)
		for foundrows.Next() {
			var thehit DbLemma
			e := foundrows.Scan(&thehit.Entry, &thehit.Xref, &thehit.Deriv)
			chke(e)
			thehit.Entry = clean.Replace(thehit.Entry)
			unnested[thehit.Entry] = thehit
		}
		foundrows.Close()
	}

	return unnested
}

// nestedlemmamapper - map[string]map[string]DbLemma for the hinter
func nestedlemmamapper(unnested map[string]DbLemma) map[string]map[string]DbLemma {
	// 20.96MB    20.96MB (flat, cum)  7.91% of Total
	// you need both a nested and the unnested version; nested is for the hinter
	nested := make(map[string]map[string]DbLemma, NESTEDLEMMASIZE)
	swap := strings.NewReplacer("j", "i", "v", "u")
	for k, v := range unnested {
		rbag := []rune(v.Entry)[0:2]
		rbag = stripaccentsRUNE(rbag)
		bag := strings.ToLower(string(rbag))
		bag = swap.Replace(bag)
		if _, y := nested[bag]; !y {
			nested[bag] = make(map[string]DbLemma)
		}
		nested[bag][k] = v
	}
	return nested
}

// buildaucorpusmap - populate global variable used by SessionIntoSearchlist()
func buildwkcorpusmap() map[string][]string {
	// SessionIntoSearchlist() could just grab a pre-rolled list instead of calculating every time...
	wkcorpusmap := make(map[string][]string)
	for _, w := range AllWorks {
		for _, c := range TheCorpora {
			if w.UID[0:2] == c {
				wkcorpusmap[c] = append(wkcorpusmap[c], w.UID)
			}
		}
	}
	return wkcorpusmap
}

// buildaucorpusmap - populate global variable used by SessionIntoSearchlist()
func buildaucorpusmap() map[string][]string {
	// SessionIntoSearchlist() could just grab a pre-rolled list instead of calculating every time...
	aucorpusmap := make(map[string][]string)
	for _, a := range AllAuthors {
		for _, c := range TheCorpora {
			if a.UID[0:2] == c {
				aucorpusmap[c] = append(aucorpusmap[c], a.UID)
			}
		}
	}
	return aucorpusmap
}

// buildaugenresmap - populate global variable used by hinter
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

// buildwkgenresmap - populate global variable used by hinter
func buildwkgenresmap() map[string]bool {
	genres := make(map[string]bool)
	for _, w := range AllWorks {
		genres[w.Genre] = true
	}
	return genres
}

// buildaulocationmap - populate global variable used by hinter
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

// buildwklocationmap - populate global variable used by hinter
func buildwklocationmap() map[string]bool {
	locations := make(map[string]bool)
	for _, w := range AllWorks {
		locations[w.Prov] = true
	}
	return locations
}
