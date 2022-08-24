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
	if l > 1 {
		return false
	} else {
		return true
	}
}

// sessionintosearchlist - converts the stored set of selections into a calculated pair of SearchIncExl w/ Authors, Works, Passages
func sessionintosearchlist(s Session) [2]SearchIncExl {
	var inc SearchIncExl
	var exc SearchIncExl

	// note that we do all the initial stuff by adding WORKS to the list individually

	// [a] trim mappers by active corpora
	auu := make(map[string]DbAuthor)
	wkk := make(map[string]DbWork)

	for k, v := range s.ActiveCorp {
		for _, a := range AllAuthors {
			if a.UID[0:2] == k && v == true {
				auu[a.UID] = a
			}
		}
		for _, w := range AllWorks {
			if w.UID[0:2] == k && v == true {
				wkk[w.UID] = w
			}
		}
	}

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
				if strings.Contains(a.Genres, g) {
					inc.Works = append(inc.Works, a.WorkList...)
				}
			}
		}
		// [b2] work genres to include
		for _, g := range sessincl.WkGenres {
			for _, w := range wkk {
				if w.Genre == g {
					inc.Works = append(inc.Works, w.UID)
				}
			}
		}

		// [b3] author locations to include
		for _, l := range sessincl.AuLocations {
			for _, a := range auu {
				if a.Location == l {
					inc.Works = append(inc.Works, a.WorkList...)
				}
			}
		}

		// [b4] work locations to include
		for _, l := range sessincl.WkLocations {
			for _, w := range wkk {
				if w.Prov == l {
					inc.Works = append(inc.Works, w.UID)
				}
			}
		}

		// 		a tricky spot: when/how to apply prunebydate()
		//		if you want to be able to seek 5th BCE oratory and Plutarch, then you need to let auselections take precedence
		//		accordingly we will do classes and genres first, then trim by date, then add inc individual choices

		// [b5] prune by date

		inc.Works = prunebydate(inc.Works, sessincl, wkk, s)

		// [b6] add all works of the authors selected

		for _, au := range sessincl.Authors {
			// this should be superfluous, but...
			_, remains := auu[au]
			if remains {
				inc.Works = append(inc.Works, auu[au].WorkList...)
			}
		}

		// [b7] add the individual works selected

		for _, wk := range sessincl.Works {
			// this should be superfluous, but...
			_, remains := wkk[wk]
			if remains {
				inc.Works = append(inc.Works, wk)
			}
		}

		// [b8] add the individual passages selected

		inc.Passages = append(inc.Passages, sessincl.Passages...)

	} else {
		// you want everything. well, maybe everything...
		for _, w := range wkk {
			inc.Works = append(inc.Works, w.UID)
		}

		// but maybe the only restriction is time...
		inc.Works = prunebydate(inc.Works, sessincl, wkk, s)
	}

	// [c] subtract the exclusions from the searchlist

	// [c1] do we allow spuria, incerta, varia?
	// note that the following will kill explicitly selected spuria: basically a logic bug, but not a priority...

	if !s.SpuriaOK {
		var trimmed []string
		for _, w := range inc.Works {
			if wkk[w[0:10]].Authentic {
				trimmed = append(trimmed, w)
			}
		}
		inc.Works = trimmed
	}

	if !s.VariaOK {
		var trimmed []string
		for _, w := range inc.Works {
			if wkk[w[0:10]].ConvDate != VARIADATE {
				trimmed = append(trimmed, w)
			}
		}
		inc.Works = trimmed
	}

	if !s.IncertaOK {
		var trimmed []string
		for _, w := range inc.Works {
			if wkk[w].ConvDate != INCERTADATE {
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
				if strings.Contains(a.Genres, g) {
					blacklist = append(blacklist, a.UID)
				}
			}
		}

		// [c2c] the author locations
		for _, l := range sessexl.AuLocations {
			for _, a := range auu {
				if a.Location == l {
					blacklist = append(blacklist, a.UID)
				}
			}
		}

		blacklist = unique(blacklist)

		// [c2d] all works of all excluded authors are themselves excluded
		// we are now moving over from AuUIDs to WkUIDS...

		for _, b := range blacklist {
			exc.Works = append(exc.Works, auu[b].WorkList...)
		}

		// [c2e] + the plain old work exclusions
		exc.Works = append(exc.Works, sessexl.Works...)

		// [c2f] works excluded by genre
		for _, l := range sessexl.WkGenres {
			for _, w := range wkk {
				if w.Genre == l {
					exc.Works = append(exc.Works, w.UID)
				}
			}
		}

		// [c2g] works excluded by provenance
		for _, l := range sessexl.WkLocations {
			for _, w := range wkk {
				if w.Prov == l {
					exc.Works = append(exc.Works, w.UID)
				}
			}
		}

		inc.Works = setsubtraction(inc.Works, exc.Works)
	}

	// this is the moment when you know the total # of locations searched: worth recording somewhere

	// now we lose that info in the name of making the search quicker...
	inc.Authors = calculatewholeauthorsearches(inc.Works, auu)

	// still need to clean the whole authors out of inc.Works
	var trim []string
	for _, i := range inc.Works {
		if !contains(inc.Authors, i[0:6]) {
			trim = append(trim, i)
		}
	}

	inc.Works = trim
	return [2]SearchIncExl{inc, exc}
}

// prunebydate - drop items from searchlist if they are not inside the valid date range
func prunebydate(searchlist []string, incl SearchIncExl, wkk map[string]DbWork, s Session) []string {
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
			if wkk[uid].DateInRange(b, a) {
				trimmed = append(trimmed, uid)
			}
		}

		// [b5b] add back in any varia and/or incerta as needed
		if s.VariaOK {
			for _, uid := range searchlist {
				if wkk[uid].ConvDate == VARIADATE {
					trimmed = append(trimmed, uid)
				}
			}
		}

		if s.IncertaOK {
			for _, uid := range searchlist {
				if wkk[uid].ConvDate == INCERTADATE {
					trimmed = append(trimmed, uid)
				}
			}
		}

		searchlist = trimmed
	}
	return searchlist
}

// calculatewholeauthorsearches - find all authors where 100% of works are requested in the searchlist
func calculatewholeauthorsearches(sl []string, aa map[string]DbAuthor) []string {
	// 	we have applied all of our inclusions and exclusions by this point and we might well be sitting on a pile of authorsandworks
	//	that is really a pile of full author dbs. for example, imagine we have not excluded anything from 'Cicero'
	//
	//	there is no reason to search that DB work by work since that just means doing a series of "WHERE" searches
	//	instead of a single, faster search of the whole thing: hits are turned into full citations via the info contained in the
	//	hit itself and there is no need to derive the work from the item name sent to the dispatcher
	//
	//	this function will figure out if the list of work uids contains all of the works for an author and can accordingly be collapsed

	var wholes []string

	members := make(map[string]int)
	for _, s := range sl {
		// count the works
		members[s] += 1
	}

	for k, v := range members {
		if len(aa[k].WorkList) == v {
			wholes = append(wholes, k)
		}
	}

	return wholes
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
	in := sl[0]

	sort.Slice(in.Authors, func(i, j int) bool { return in.Authors[i] < in.Authors[j] })
	sort.Slice(in.Works, func(i, j int) bool { return in.Works[i] < in.Works[j] })
	fmt.Println(in.Authors)
	fmt.Println(in.Works)
	timetracker("-", "searchlist compiled", start, previous)
}
