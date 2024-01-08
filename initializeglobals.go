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
	"sort"
	"strings"
	"time"
)

const (
	AUTHORTEMPLATE = ` universalid, language, idxname, akaname, shortname, cleanname, genres, recorded_date, converted_date, location `

	WORKTEMPLATE = ` universalid, title, language, publication_info,
		levellabels_00, levellabels_01, levellabels_02, levellabels_03, levellabels_04, levellabels_05,
		workgenre, transmission, worktype, provenance, recorded_date, converted_date, wordcount,
		firstline, lastline, authentic`
)

// The maps below are all race condition candidates, but most of them are write once, read many.

// But note that the races are in practice a "small" and "technical" complaint. HipparchiaGoServer is not supposed to
// be exposed to everyone, everywhere all the time. Someone requesting 100 searches is a bigger worry than any race on
// 10m requests condition. There needs to be a fairly high degree of user trust to begin with.

var (
	Config        CurrentConfiguration
	SQLPool       *pgxpool.Pool
	AllSearches   = MakeSearchVault()
	AllSessions   = MakeSessionVault()
	AllAuthorized = MakeAuthorizedVault()
	UserPassPairs = make(map[string]string)
	AllWorks      = make(map[string]*DbWork)
	AllAuthors    = make(map[string]*DbAuthor)
	AllLemm       = make(map[string]*DbLemma)
	NestedLemm    = make(map[string]map[string]*DbLemma)
	WkCorpusMap   = make(map[string][]string)
	AuCorpusMap   = make(map[string][]string)
	LoadedCorp    = make(map[string]bool)
	AuGenres      = make(map[string]bool)
	WkGenres      = make(map[string]bool)
	AuLocs        = make(map[string]bool)
	WkLocs        = make(map[string]bool)
	TheCorpora    = []string{GREEKCORP, LATINCORP, INSCRIPTCORP, CHRISTINSC, PAPYRUSCORP}
	TheLanguages  = []string{"greek", "latin"}
	ServableFonts = map[string]FontTempl{"Noto": NotoFont, "Roboto": RobotoFont, "Fira": FiraFont}
	LaunchTime    = time.Now()
	WebsocketPool = WSFillNewPool()
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
func (dbw *DbWork) WkID() string {
	return dbw.UID[7:]
}

// AuID - ex: gr2017w068 --> gr2017
func (dbw *DbWork) AuID() string {
	if len(dbw.UID) < 6 {
		return ""
	} else {
		return dbw.UID[:6]
	}
}

// MyAu - return the work's DbAuthor
func (dbw *DbWork) MyAu() *DbAuthor {
	a, ok := AllAuthors[dbw.AuID()]
	if !ok {
		msg(fmt.Sprintf("DbWork.MyAu() failed to find '%s'", dbw.AuID()), MSGWARN)
		a = &DbAuthor{}
	}
	return a
}

func (dbw *DbWork) CitationFormat() []string {
	cf := []string{
		dbw.LL5, dbw.LL4, dbw.LL3, dbw.LL2, dbw.LL1, dbw.LL0,
	}
	return cf
}

// CountLevels - the work structure employs how many levels?
func (dbw *DbWork) CountLevels() int {
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

// FindValidLevelValues - tell me some of a citation and I can tell you what is a valid choice at the next step
func (dbw *DbWork) FindValidLevelValues(locc []string) LevelValues {
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

	lvls := dbw.CountLevels() - 1 // count vs indexing adjustment
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
	prq.PsqlQuery = fmt.Sprintf(SEL, dbw.AuID(), dbw.UID, and, andnot)

	dbconn := GetDBConnection()
	defer dbconn.Release()
	wlb := WorklineQuery(prq, dbconn)

	// [c] extract info from the hitlines returned
	var vals LevelValues
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
	r = Unique(r)
	sort.Strings(r)
	vals.Range = r

	return vals
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

// these functions should be run in order to initialize/reinitialize the core data

// [1] WORKMAPPING

// activeworkmapper - build a map of all works in the *active* corpora; keyed to the authorUID: map[string]*DbWork
func activeworkmapper() map[string]*DbWork {
	// note that you are still on the hook for adding to the global workmap when someone cals "/setoption/papyruscorpus/yes"
	// AND you should never drop from the map because that is only session-specific: "no" is only "no for me"
	// so the memory footprint can only grow: but G&L is an 82M launch vs an 189M launch for everything

	// the bookkeeping is handled by modifyglobalmapsifneeded() inside of RtSetOption()

	workmap := make(map[string]*DbWork)

	for k, b := range Config.DefCorp {
		if b {
			workmap = mapnewworkcorpus(k, workmap)
		}
	}
	return workmap
}

// mapnewworkcorpus - add a corpus to a workmap
func mapnewworkcorpus(corpus string, workmap map[string]*DbWork) map[string]*DbWork {
	const (
		MSG = "mapnewworkcorpus() added %d works from '%s'"
	)
	toadd := sliceworkcorpus(corpus)
	for i := 0; i < len(toadd); i++ {
		w := toadd[i]
		workmap[w.UID] = &w
	}

	LoadedCorp[corpus] = true

	msg(fmt.Sprintf(MSG, len(toadd), corpus), MSGPEEK)
	return workmap
}

// sliceworkcorpus - fetch all relevant works from the db as a DbWork slice
func sliceworkcorpus(corpus string) []DbWork {
	// this is far and away the "heaviest" bit of the whole program if you grab every known work
	// Total: 204MB
	// 65.35MB (flat, cum) 32.03% of Total

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

	const (
		CT = `SELECT count(*) FROM works WHERE universalid ~* '^%s'`
		QT = `SELECT %s FROM works WHERE universalid ~* '^%s'`
	)

	var cc int
	cq := fmt.Sprintf(CT, corpus)
	qq := fmt.Sprintf(QT, WORKTEMPLATE, corpus)

	countrow := SQLPool.QueryRow(context.Background(), cq)
	err := countrow.Scan(&cc)

	foundrows, err := SQLPool.Query(context.Background(), qq)
	chke(err)

	workslice := make([]DbWork, cc)
	var w DbWork

	foreach := []any{&w.UID, &w.Title, &w.Language, &w.Pub, &w.LL0, &w.LL1, &w.LL2, &w.LL3, &w.LL4, &w.LL5, &w.Genre,
		&w.Xmit, &w.Type, &w.Prov, &w.RecDate, &w.ConvDate, &w.WdCount, &w.FirstLine, &w.LastLine, &w.Authentic}

	index := 0
	rwfnc := func() error {
		workslice[index] = w
		index++
		return nil
	}

	_, e := pgx.ForEachRow(foundrows, foreach, rwfnc)
	chke(e)

	return workslice
}

// [2] AUTHORS AND CORPUS DATA (DO NOT RUN THESE AS UPDATES BEFORE UPDATING AllWorks)

// [2a] AUTHORS

// activeauthormapper - build a map of all authors in the *active* corpora; keyed to the authorUID: map[string]*DbAuthor
func activeauthormapper() map[string]*DbAuthor {
	// see comments at top of activeworkmapper(): they apply here too
	authmap := make(map[string]*DbAuthor)
	for k, b := range Config.DefCorp {
		if b {
			authmap = mapnewauthorcorpus(k, authmap)
		}
	}
	return authmap
}

// mapnewauthorcorpus - add a corpus to an authormap
func mapnewauthorcorpus(corpus string, authmap map[string]*DbAuthor) map[string]*DbAuthor {
	const (
		MSG = "mapnewauthorcorpus() added %d authors from '%s'"
	)

	toadd := sliceauthorcorpus(corpus)
	for i := 0; i < len(toadd); i++ {
		a := toadd[i]
		authmap[a.UID] = &a
	}

	LoadedCorp[corpus] = true

	msg(fmt.Sprintf(MSG, len(toadd), corpus), MSGPEEK)

	return authmap
}

// sliceauthorcorpus - fetch all relevant works from the db as a DbAuthor slice
func sliceauthorcorpus(corpus string) []DbAuthor {
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

	const (
		CT = `SELECT count(*) FROM authors WHERE universalid ~* '^%s'`
		QT = `SELECT %s FROM authors WHERE universalid ~* '^%s'`
	)

	// need to be ready to load the worklists into the authors
	// so: build a map of {UID: WORKLIST...}; map called by rfnc()

	worklists := make(map[string][]string)
	for _, w := range AllWorks {
		wk := w.UID
		au := wk[0:6]
		if _, y := worklists[au]; !y {
			worklists[au] = []string{wk}
		} else {
			worklists[au] = append(worklists[au], wk)
		}
	}

	var cc int
	cq := fmt.Sprintf(CT, corpus)
	qq := fmt.Sprintf(QT, AUTHORTEMPLATE, corpus)

	countrow := SQLPool.QueryRow(context.Background(), cq)
	err := countrow.Scan(&cc)

	foundrows, err := SQLPool.Query(context.Background(), qq)
	chke(err)

	authslice := make([]DbAuthor, cc)
	var a DbAuthor
	foreach := []any{&a.UID, &a.Language, &a.IDXname, &a.Name, &a.Shortname, &a.Cleaname, &a.Genres, &a.RecDate, &a.ConvDate, &a.Location}

	index := 0
	rfnc := func() error {
		a.WorkList = worklists[a.UID]
		authslice[index] = a
		index++
		return nil
	}

	_, e := pgx.ForEachRow(foundrows, foreach, rfnc)
	chke(e)

	return authslice
}

// [2b] CORPUSMAPS

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

// populateglobalmaps - full up WkCorpusMap, AuCorpusMap, ...
func populateglobalmaps() {
	WkCorpusMap = buildwkcorpusmap()
	AuCorpusMap = buildaucorpusmap()
	AuGenres = buildaugenresmap()
	WkGenres = buildwkgenresmap()
	AuLocs = buildaulocationmap()
	WkLocs = buildwklocationmap()
}

// [3] LEMMATA

// lemmamapper - map[string]DbLemma for all lemmata
func lemmamapper() map[string]*DbLemma {
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

	// a list of 152k words is too long to send to 'getlemmahint' without offering quicker access
	// [HGS] [B1: 0.167s][Δ: 0.167s] unnested lemma map built (152382 items)

	// move to pgx v5 slows this function down (and will add .1s to startup time...):
	// [HGS] [B1: 0.436s][Δ: 0.436s] unnested lemma map built (152382 items)
	// see devel-mutex 1.0.7 at e841c135f22ffaae26cb5cc29e20be58bf4801d7 vs 9457ace03e048c0e367d132cef595ed1661a8c12
	// but pgx v5 does seem faster and more memory efficient in general: must not like returning huge lists

	const (
		THEQUERY = `SELECT dictionary_entry, xref_number, derivative_forms FROM %s_lemmata`
	)

	// note that the v --> u here will push us to stripped_line SearchMap instead of accented_line
	// clean := strings.NewReplacer("-", "", "¹", "", "²", "", "³", "", "j", "i", "v", "u")
	clean := strings.NewReplacer("-", "", "j", "i", "v", "u")

	unnested := make(map[string]*DbLemma, DBLMMAPSIZE)

	// use the older iterative idiom to facilitate working with pointers: "foreach" idiom will fight you...
	for _, lg := range TheLanguages {
		q := fmt.Sprintf(THEQUERY, lg)
		foundrows, err := SQLPool.Query(context.Background(), q)
		chke(err)
		for foundrows.Next() {
			thehit := &DbLemma{}
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
func nestedlemmamapper(unnested map[string]*DbLemma) map[string]map[string]*DbLemma {
	// 20.96MB    20.96MB (flat, cum)  7.91% of Total
	// you need both a nested and the unnested version; nested is for the hinter

	nested := make(map[string]map[string]*DbLemma, NESTEDLEMMASIZE)
	swap := strings.NewReplacer("j", "i", "v", "u")
	for k, v := range unnested {
		rbag := []rune(v.Entry)[0:2]
		rbag = StripaccentsRUNE(rbag)
		bag := strings.ToLower(string(rbag))
		bag = swap.Replace(bag)
		if _, y := nested[bag]; !y {
			nested[bag] = make(map[string]*DbLemma)
		}
		nested[bag][k] = v
	}
	return nested
}
