//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type SearchIncExl struct {
	// the first are for internal use
	AuGenres    []string
	WkGenres    []string
	AuLocations []string
	WkLocations []string
	Authors     []string
	Works       []string
	Passages    []string // "lt0474_FROM_36136_TO_36151"
	// the next are for output to the browser
	MappedPsgByName  map[string]string // "lt0474_FROM_36136_TO_36151": "Cicero, Pro Caelio, section 1
	MappedAuthByName map[string]string
	MappedWkByName   map[string]string
	// "val.Interface().([]string)" assertion in makeselections.go means we have to insist on the slices
	ListedPBN []string
	ListedABN []string
	ListedWBN []string
}

func (i *SearchIncExl) isEmpty() bool {
	l := len(i.AuGenres) + len(i.WkGenres) + len(i.AuLocations) + len(i.WkLocations) + len(i.Authors)
	l += len(i.Works) + len(i.Passages)
	if l > 0 {
		return false
	} else {
		return true
	}
}

func (i *SearchIncExl) CountItems() int {
	l := len(i.AuGenres) + len(i.WkGenres) + len(i.AuLocations) + len(i.WkLocations) + len(i.Authors)
	l += len(i.Works) + len(i.Passages)
	return l
}

func (i *SearchIncExl) BuildAuByName() {
	bn := make(map[string]string, len(i.MappedAuthByName))
	for _, a := range i.Authors {
		bn[a] = AllAuthors[a].Cleaname
	}
	i.MappedAuthByName = bn

	var nn []string
	for _, v := range bn {
		nn = append(nn, v)
	}

	sort.Slice(nn, func(i, j int) bool { return nn[i] < nn[j] })
	i.ListedABN = nn
}

func (i *SearchIncExl) BuildWkByName() {
	const (
		TMPL = `%s, <i>%s</i>`
	)
	bn := make(map[string]string, len(i.MappedWkByName))
	for _, w := range i.Works {
		ws := AllWorks[w]
		au := ws.MyAu().Name
		ti := ws.Title
		bn[w] = fmt.Sprintf(TMPL, au, ti)
	}
	i.MappedWkByName = bn

	var nn []string
	for _, v := range bn {
		nn = append(nn, v)
	}

	sort.Slice(nn, func(i, j int) bool { return nn[i] < nn[j] })
	i.ListedWBN = nn
}

func (i *SearchIncExl) BuildPsgByName() {
	var nn []string
	for _, v := range i.MappedPsgByName {
		nn = append(nn, v)
	}
	sort.Slice(nn, func(i, j int) bool { return nn[i] < nn[j] })
	i.ListedPBN = nn
}

type ProcessedList struct {
	Inc  SearchIncExl
	Excl SearchIncExl
	Size int
}

// SessionIntoSearchlist - converts the stored set of selections into a calculated pair of SearchIncExl w/ Authors, Works, Passages
func SessionIntoSearchlist(s ServerSession) ProcessedList {
	// https://medium.com/scum-gazeta/golang-simple-optimization-notes-70bc64673980

	var inc SearchIncExl
	var exc SearchIncExl

	// note that we do all the initial stuff by adding WORKS to the list individually

	// [a] trim mappers by active corpora

	var activeauthors []string
	var activeworks []string

	for k, v := range s.ActiveCorp {
		if v {
			activeauthors = append(activeauthors, AuCorpusMap[k]...)
			activeworks = append(activeworks, WkCorpusMap[k]...)
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
			for _, a := range activeauthors {
				if strings.Contains(AllAuthors[a].Genres, g) {
					inc.Works = append(inc.Works, AllAuthors[a].WorkList...)
				}
			}
		}
		// [b2] work genres to include
		for _, g := range sessincl.WkGenres {
			for _, w := range activeworks {
				if AllWorks[w].Genre == g {
					inc.Works = append(inc.Works, AllWorks[w].UID)
				}
			}
		}

		// [b3] author locations to include
		for _, l := range sessincl.AuLocations {
			for _, a := range activeauthors {
				if AllAuthors[a].Location == l {
					inc.Works = append(inc.Works, AllAuthors[a].WorkList...)
				}
			}
		}

		// [b4] work locations to include
		for _, l := range sessincl.WkLocations {
			for _, w := range activeworks {
				if AllWorks[w].Prov == l {
					inc.Works = append(inc.Works, AllWorks[w].UID)
				}
			}
		}

		// 		a tricky spot: when/how to apply prunebydate()
		//		if you want to be able to seek 5th BCE oratory and Plutarch, then you need to let auselections take precedence
		//		accordingly we will do classes and genres first, then trim by date, then add inc individual choices

		// [b5] prune by date

		inc.Works = prunebydate(inc.Works, s)

		// [b6] add all works of the authors registerselection

		for _, au := range sessincl.Authors {
			// this should be superfluous, but...
			_, remains := AllAuthors[au]
			if remains {
				inc.Works = append(inc.Works, AllAuthors[au].WorkList...)
			}
		}

		// [b7] add the individual works registerselection

		for _, wk := range sessincl.Works {
			// this should be superfluous, but...
			_, remains := AllWorks[wk]
			if remains {
				inc.Works = append(inc.Works, wk)
			}
		}

		// [b8] add the individual passages registerselection

		inc.Passages = append(inc.Passages, sessincl.Passages...)

	} else {
		// you want everything. well, maybe everything...
		for _, w := range activeworks {
			inc.Works = append(inc.Works, AllWorks[w].UID)
		}

		// but maybe the only restriction is time...
		inc.Works = prunebydate(inc.Works, s)
	}

	// [c] subtract the exclusions from the searchlist

	// [c1] do we allow spuria (varia and incerta already lost via prunebydate)

	// note that the following will kill explicitly registerselection spuria: basically a logic bug, but not a priority...

	if !s.SpuriaOK {
		var trimmed []string
		for _, w := range inc.Works {
			if AllWorks[w[0:10]].Authentic {
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
			for _, a := range activeauthors {
				if strings.Contains(AllAuthors[a].Genres, g) {
					blacklist = append(blacklist, AllAuthors[a].UID)
				}
			}
		}

		// [c2c] the author locations
		for _, l := range sessexl.AuLocations {
			for _, a := range activeauthors {
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
			for _, w := range activeworks {
				if AllWorks[w].Genre == l {
					exc.Works = append(exc.Works, AllWorks[w].UID)
				}
			}
		}

		// [c2g] works excluded by provenance
		for _, l := range sessexl.WkLocations {
			for _, w := range activeworks {
				if AllWorks[w].Prov == l {
					exc.Works = append(exc.Works, AllWorks[w].UID)
				}
			}
		}

		inc.Works = setsubtraction(inc.Works, exc.Works)
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

	var trim []string
	for _, i := range inc.Works {
		if _, ok := prunemap[i]; !ok {
			trim = append(trim, i)
		}
	}

	inc.Works = trim

	inc.Passages = unique(inc.Passages)
	inc.Works = unique(inc.Works)
	inc.Authors = unique(inc.Authors)

	exc.Passages = unique(exc.Passages)
	exc.Works = unique(exc.Works)
	exc.Authors = unique(exc.Authors)

	var proc ProcessedList
	proc.Inc = inc
	proc.Excl = exc
	proc.Size = sl

	// fmt.Println(fmt.Sprintf("SessionIntoSearchlist(): proc is\n\t%s\n", proc.Inc))
	return proc
}

// prunebydate - drop items from searchlist if they are not inside the valid date range
func prunebydate(searchlist []string, s ServerSession) []string {
	// 'varia' and 'incerta' have special dates: incerta = 2500; varia = 2000

	if s.Earliest == MINDATESTR && s.Latest == MAXDATESTR && s.VariaOK && s.IncertaOK {
		// no work for us to do...
		return searchlist
	}

	earliest, _ := strconv.Atoi(s.Earliest)
	latest, _ := strconv.Atoi(s.Latest)

	e := int64(earliest)
	l := int64(latest)

	// should have already been validated elsewhere...
	if e > l {
		e = l
	}

	// [b5a] first prune the bad dates; nb: the inscriptions have lots of work dates; the gl and lt works don't
	var trimmed []string
	for _, uid := range searchlist {
		cda := AllAuthors[AllWorks[uid].AuID()].ConvDate
		cdb := AllWorks[uid].ConvDate
		if (cda >= e && cda <= l) || (cdb >= e && cdb <= l) {
			trimmed = append(trimmed, uid)
			// msg(fmt.Sprintf("added: %s w/ date of %d", uid, cd), 1)
		}
	}

	// [b5b] add back in any varia and/or incerta as needed
	if s.VariaOK {
		for _, uid := range searchlist {
			cda := AllAuthors[AllWorks[uid].AuID()].ConvDate
			cdb := AllWorks[uid].ConvDate
			if (cda == INCERTADATE || cda == VARIADATE) && cdb == VARIADATE {
				trimmed = append(trimmed, uid)
			}
		}
	}

	if s.IncertaOK {
		for _, uid := range searchlist {
			cda := AllAuthors[AllWorks[uid].AuID()].ConvDate
			cdb := AllWorks[uid].ConvDate
			if (cda == INCERTADATE || cda == VARIADATE) && cdb == INCERTADATE {
				trimmed = append(trimmed, uid)
			}
		}
	}

	searchlist = trimmed

	return searchlist
}

// calculatewholeauthorsearches - find all authors where 100% of works are requested in the searchlist
func calculatewholeauthorsearches(sl []string) [2][]string {
	// 	we have applied all of our inclusions and exclusions by this point and we might well be sitting on a pile of authorsandworks
	//	that is really a pile of full author dbs. for example, imagine we have not excluded anything from 'Cicero'
	//
	//	there is no reason to search that DB work by work since that just means doing a series of "WHERE" SearchMap
	//	instead of a single, faster search of the whole thing: hits are turned into full citations via the info contained in the
	//	hit itself and there is no need to derive the work from the item name sent to the dispatcher
	//
	//	this function will figure out if the list of work uids contains all of the works for an author and can accordingly be collapsed

	//start := time.Now()
	//previous := time.Now()

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
	// timetracker("-", "calculatewholeauthorsearches()", start, previous)

	return [2][]string{wholes, pruner}
}
