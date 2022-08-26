package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type SearchIncExl struct {
	AuGenres       []string
	WkGenres       []string
	AuLocations    []string
	WkLocations    []string
	Authors        []string
	Works          []string
	Passages       []string          // "lt0474_FROM_36136_TO_36151"
	PassagesByName map[string]string // "lt0474_FROM_36136_TO_36151": "Cicero, Pro Caelio, section 1
	DateRange      [2]string
}

func (i SearchIncExl) isEmpty() bool {
	l := len(i.AuGenres) + len(i.WkGenres) + len(i.AuLocations) + len(i.WkLocations) + len(i.Authors)
	l += len(i.Works) + len(i.Passages)
	if l > 0 {
		return false
	} else {
		return true
	}
}

type ProcessedList struct {
	Inc  SearchIncExl
	Excl SearchIncExl
	Size int
}

// sessionintosearchlist - converts the stored set of selections into a calculated pair of SearchIncExl w/ Authors, Works, Passages
func sessionintosearchlist(s Session) ProcessedList {
	// pretty slow, really if you have all lists active: [HGS] [B: 1.112s][Δ: 1.112s] sessionintosearchlist()
	// https://medium.com/scum-gazeta/golang-simple-optimization-notes-70bc64673980
	start := time.Now()
	previous := time.Now()
	var inc SearchIncExl
	var exc SearchIncExl

	// note that we do all the initial stuff by adding WORKS to the list individually

	// [a] trim mappers by active corpora
	// SLOW: [Δ: 0.244s] sessionintosearchlist(): trim mappers by active corpora
	//auu := make(map[string]DbAuthor)
	//wkk := make(map[string]DbWork)
	//
	//for k, v := range s.ActiveCorp {
	//	for _, a := range AllAuthors {
	//		if a.UID[0:2] == k && v == true {
	//			auu[a.UID] = a
	//		}
	//	}
	//	for _, w := range AllWorks {
	//		if w.UID[0:2] == k && v == true {
	//			wkk[w.UID] = w
	//		}
	//	}
	//}

	// faster, but requires a lot of edits below: [Δ: 0.116s] sessionintosearchlist(): trim mappers by active corpora
	auu := make([]string, 0, len(AllAuthors))
	wkk := make([]string, 0, len(AllWorks))
	for k, v := range s.ActiveCorp {
		for _, a := range AllAuthors {
			if a.UID[0:2] == k && v == true {
				auu = append(auu, a.UID)
			}
		}
	}
	for k, v := range s.ActiveCorp {
		for _, w := range AllWorks {
			if w.UID[0:2] == k && v == true {
				wkk = append(wkk, w.UID)
			}
		}
	}

	timetracker("1", "sessionintosearchlist(): trim mappers by active corpora", start, previous)
	previous = time.Now()

	sessincl := s.Inclusions
	sessexl := s.Exclusions

	// retain in unmodified form
	inc.Passages = sessincl.Passages
	exc.Passages = sessexl.Passages

	// [b] build the inclusion list: inc.Works is the core searchlist
	if !sessincl.isEmpty() {
		// you only want *some* things
		// [b1] author genres to include
		for _, g := range sessincl.AuGenres {
			for _, a := range auu {
				if strings.Contains(AllAuthors[a].Genres, g) {
					inc.Works = append(inc.Works, AllAuthors[a].WorkList...)
				}
			}
		}
		// [b2] work genres to include
		for _, g := range sessincl.WkGenres {
			for _, w := range wkk {
				if AllWorks[w].Genre == g {
					inc.Works = append(inc.Works, AllWorks[w].UID)
				}
			}
		}

		// [b3] author locations to include
		for _, l := range sessincl.AuLocations {
			for _, a := range auu {
				if AllAuthors[a].Location == l {
					inc.Works = append(inc.Works, AllAuthors[a].WorkList...)
				}
			}
		}

		// [b4] work locations to include
		for _, l := range sessincl.WkLocations {
			for _, w := range wkk {
				if AllWorks[w].Prov == l {
					inc.Works = append(inc.Works, AllWorks[w].UID)
				}
			}
		}
		timetracker("2", "sessionintosearchlist(): build inclusion list", start, previous)
		previous = time.Now()
		// 		a tricky spot: when/how to apply prunebydate()
		//		if you want to be able to seek 5th BCE oratory and Plutarch, then you need to let auselections take precedence
		//		accordingly we will do classes and genres first, then trim by date, then add inc individual choices

		// [b5] prune by date

		inc.Works = prunebydate(inc.Works, sessincl, s)
		timetracker("3", "sessionintosearchlist(): prune by date", start, previous)
		previous = time.Now()
		// [b6] add all works of the authors selected

		for _, au := range sessincl.Authors {
			// this should be superfluous, but...
			_, remains := AllAuthors[au]
			if remains {
				inc.Works = append(inc.Works, AllAuthors[au].WorkList...)
			}
		}

		// [b7] add the individual works selected

		for _, wk := range sessincl.Works {
			// this should be superfluous, but...
			_, remains := AllWorks[wk]
			if remains {
				inc.Works = append(inc.Works, wk)
			}
		}

		// [b8] add the individual passages selected

		inc.Passages = append(inc.Passages, sessincl.Passages...)

	} else {
		// you want everything. well, maybe everything...
		for _, w := range wkk {
			inc.Works = append(inc.Works, AllWorks[w].UID)
		}
		timetracker("4", "sessionintosearchlist(): all of everything", start, previous)
		previous = time.Now()
		// but maybe the only restriction is time...
		inc.Works = prunebydate(inc.Works, sessincl, s)
		timetracker("5", "sessionintosearchlist(): prune by date", start, previous)
		previous = time.Now()
	}

	// [c] subtract the exclusions from the searchlist

	// [c1] do we allow spuria, incerta, varia?
	// note that the following will kill explicitly selected spuria: basically a logic bug, but not a priority...

	if !s.SpuriaOK {
		var trimmed []string
		for _, w := range inc.Works {
			if AllWorks[w[0:10]].Authentic {
				trimmed = append(trimmed, w)
			}
		}
		inc.Works = trimmed
	}

	if !s.VariaOK {
		var trimmed []string
		for _, w := range inc.Works {
			if AllWorks[w[0:10]].ConvDate != VARIADATE {
				trimmed = append(trimmed, w)
			}
		}
		inc.Works = trimmed
	}

	if !s.IncertaOK {
		var trimmed []string
		for _, w := range inc.Works {
			if AllWorks[w].ConvDate != INCERTADATE {
				trimmed = append(trimmed, w)
			}
		}
		inc.Works = trimmed
	}

	// [c2] walk through the exclusions categories; note that excluded passages are handled via the querybuilder

	if !sessexl.isEmpty() {
		// [c2a] the authors
		blacklist := sessexl.Authors

		// [c2c] the author genres
		for _, g := range sessexl.AuGenres {
			for _, a := range auu {
				if strings.Contains(AllAuthors[a].Genres, g) {
					blacklist = append(blacklist, AllAuthors[a].UID)
				}
			}
		}

		// [c2c] the author locations
		for _, l := range sessexl.AuLocations {
			for _, a := range auu {
				if AllAuthors[a].Location == l {
					blacklist = append(blacklist, AllAuthors[a].UID)
				}
			}
		}

		blacklist = unique(blacklist)

		// [c2d] all works of all excluded authors are themselves excluded
		// we are now moving over from AuUIDs to WkUIDS...

		for _, b := range blacklist {
			exc.Works = append(exc.Works, AllAuthors[b].WorkList...)
		}

		// [c2e] + the plain old work exclusions
		exc.Works = append(exc.Works, sessexl.Works...)

		// [c2f] works excluded by genre
		for _, l := range sessexl.WkGenres {
			for _, w := range wkk {
				if AllWorks[w].Genre == l {
					exc.Works = append(exc.Works, AllWorks[w].UID)
				}
			}
		}

		// [c2g] works excluded by provenance
		for _, l := range sessexl.WkLocations {
			for _, w := range wkk {
				if AllWorks[w].Prov == l {
					exc.Works = append(exc.Works, AllWorks[w].UID)
				}
			}
		}

		inc.Works = setsubtraction(inc.Works, exc.Works)
		timetracker("5", "sessionintosearchlist(): setsubtraction", start, previous)
		previous = time.Now()
	}

	// this is the moment when you know the total # of locations searched: worth recording
	sl := len(inc.Works)

	// now we lose that info in the name of making the search quicker...
	wp := calculatewholeauthorsearches(inc.Works)
	inc.Authors = wp[0]
	pruner := wp[1]
	prunemap := make(map[string]bool)
	for _, p := range pruner {
		prunemap[p] = true
	}

	// still need to clean the whole authors out of inc.Works

	// SLOW: [HGS] [6: 1.221s][Δ: 0.989s] sessionintosearchlist(): clean the whole authors out of inc.Works
	//var trim []string
	//for _, i := range inc.Works {
	//	if !contains(inc.Authors, i[0:6]) {
	//		trim = append(trim, i)
	//	}
	//}

	// FAST: [Δ: 0.045s] sessionintosearchlist(): clean the whole authors out of inc.Works
	var trim []string
	for _, i := range inc.Works {
		if _, ok := prunemap[i]; !ok {
			trim = append(trim, i)
		}
	}

	timetracker("6", "sessionintosearchlist(): clean the whole authors out of inc.Works", start, previous)
	previous = time.Now()

	inc.Works = trim

	var proc ProcessedList
	proc.Inc = inc
	proc.Excl = exc
	proc.Size = sl
	timetracker("7", "sessionintosearchlist(): done", start, previous)
	return proc
}

// prunebydate - drop items from searchlist if they are not inside the valid date range
func prunebydate(searchlist []string, incl SearchIncExl, s Session) []string {
	// 'varia' and 'incerta' have special dates: incerta = 2500; varia = 2000
	before, _ := strconv.Atoi(incl.DateRange[0])
	after, _ := strconv.Atoi(incl.DateRange[1])
	b := int64(before)
	a := int64(after)

	if b != MINDATE || a != MAXDATE {
		// should have already been validated elsewhere...
		if b > a {
			b = a
		}

		// [b5a] first prune the bad dates
		var trimmed []string
		for _, uid := range searchlist {
			if AllWorks[uid].DateInRange(b, a) {
				trimmed = append(trimmed, uid)
			}
		}

		// [b5b] add back in any varia and/or incerta as needed
		if s.VariaOK {
			for _, uid := range searchlist {
				if AllWorks[uid].ConvDate == VARIADATE {
					trimmed = append(trimmed, uid)
				}
			}
		}

		if s.IncertaOK {
			for _, uid := range searchlist {
				if AllWorks[uid].ConvDate == INCERTADATE {
					trimmed = append(trimmed, uid)
				}
			}
		}

		searchlist = trimmed
	}
	return searchlist
}

// calculatewholeauthorsearches - find all authors where 100% of works are requested in the searchlist
func calculatewholeauthorsearches(sl []string) [2][]string {
	// 	we have applied all of our inclusions and exclusions by this point and we might well be sitting on a pile of authorsandworks
	//	that is really a pile of full author dbs. for example, imagine we have not excluded anything from 'Cicero'
	//
	//	there is no reason to search that DB work by work since that just means doing a series of "WHERE" searches
	//	instead of a single, faster search of the whole thing: hits are turned into full citations via the info contained in the
	//	hit itself and there is no need to derive the work from the item name sent to the dispatcher
	//
	//	this function will figure out if the list of work uids contains all of the works for an author and can accordingly be collapsed

	start := time.Now()
	previous := time.Now()

	var wholes []string
	var pruner []string

	members := make(map[string]int)
	for _, s := range sl {
		// count the works
		members[s[0:6]] += 1
	}

	for k, v := range members {
		if len(AllAuthors[k].WorkList) == v {
			wholes = append(wholes, k)
			pruner = append(pruner, AllAuthors[k].WorkList...)
		}
	}

	//fmt.Printf("len(aa[lt0474].WorkList): %d\n", len(aa["lt0474"].WorkList))
	//fmt.Printf("members[lt0474]: %d\n", members["lt0474"])
	timetracker("-", "calculatewholeauthorsearches()", start, previous)

	return [2][]string{wholes, pruner}
}

func test_compilesearchlist() {
	start := time.Now()
	previous := time.Now()
	fmt.Println("testing sessionintosearchlist()")
	var s Session
	s.IncertaOK = true
	s.VariaOK = true
	s.SpuriaOK = true
	c := make(map[string]bool)
	c["gr"] = true
	c["lt"] = true
	c["dp"] = false
	c["in"] = false
	c["ch"] = false
	s.ActiveCorp = c
	i := s.Inclusions
	i.Authors = []string{"lt0474", "lt0917"}
	i.AuGenres = []string{"Apologetici", "Doxographi"}
	i.WkGenres = []string{"Eleg."}
	i.Passages = []string{"gr0032w002_FROM_11313_TO_11843"}
	i.Works = []string{"gr0062w001"}
	i.AuLocations = []string{"Abdera"}
	e := s.Exclusions
	e.Works = []string{"lt0474w001"}
	e.Passages = []string{"lt0474_FROM_36136_TO_36151"}
	s.Inclusions = i
	s.Exclusions = e

	sl := sessionintosearchlist(s)
	in := sl.Inc

	sort.Slice(in.Authors, func(i, j int) bool { return in.Authors[i] < in.Authors[j] })
	sort.Slice(in.Works, func(i, j int) bool { return in.Works[i] < in.Works[j] })
	fmt.Println(in.Authors)
	fmt.Println(in.Works)
	timetracker("-", "searchlist compiled", start, previous)
}
