//    HipparchiaGoDBHelper: search and vector helper app and functions for HipparchiaServer
//    Copyright: E Gunderson 2021
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v4/pgxpool"
	"regexp"
	"sort"
	"strings"
)

type WeightedHeadword struct {
	Word  string
	Count int
}

type WHWList []WeightedHeadword

func (w WHWList) Len() int {
	return len(w)
}

func (w WHWList) Less(i, j int) bool {
	return w[i].Count > w[j].Count
}

func (w WHWList) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
}

type BagWithLocus struct {
	Loc string
	Bag string
}

// findtherows - use a redis.Conn to acquire []DbWorkline
func findtherows(thequery string, thecaller string, searchkey string, clientnumber int, rc redis.Conn, dbpool *pgxpool.Pool) []DbWorkline {
	// called by both linegrabber() and HipparchiaBagger()
	// this VERSION contains polling data
	// it also assumes that thequery arrived via popping redis

	// [ii] update the polling data
	if thecaller != "bagger" {
		remain, err := redis.Int64(rc.Do("SCARD", searchkey))
		chke(err)

		k := fmt.Sprintf("%s_remaining", searchkey)
		_, e := rc.Do("SET", k, remain)
		chke(e)
		msg(fmt.Sprintf("%s #%d says that %d items remain", thecaller, clientnumber, remain), 3)
	}

	// [iii] decode the query
	var prq PrerolledQuery
	err := json.Unmarshal([]byte(thequery), &prq)
	chke(err)

	// fmt.Println(prq)
	foundlines := worklinequery(prq, dbpool)

	return foundlines
}

//
// CLEANING
//

func stripper(item string, purge []string) string {
	// delete each in a list of items from a string
	for i := 0; i < len(purge); i++ {
		re := regexp.MustCompile(purge[i])
		item = re.ReplaceAllString(item, "")
	}
	return item
}

func makesubstitutions(thetext string) string {
	// https://golang.org/pkg/strings/#NewReplacer
	// cf cleanvectortext() in vectorhelpers.py
	swap := strings.NewReplacer("v", "u", "j", "i", "σ", "ϲ", "ς", "ϲ", "A.", "Aulus", "App.", "Appius",
		"C.", "Caius", "G.", "Gaius", "Cn.", "Cnaius", "Gn.", "Gnaius", "D.", "Decimus", "L.", "Lucius", "M.", "Marcus",
		"M.’", "Manius", "N.", "Numerius", "P.", "Publius", "Q.", "Quintus", "S.", "Spurius", "Sp.", "Spurius",
		"Ser.", "Servius", "Sex.", "Sextus", "T.", "Titus", "Ti", "Tiberius", "V.", "Vibius", "a.", "ante",
		"d.", "dies", "Id.", "Idibus", "Kal.", "Kalendas", "Non.", "Nonas", "prid.", "pridie", "Ian.", "Ianuarias",
		"Feb.", "Februarias", "Mart.", "Martias", "Apr.", "Aprilis", "Mai.", "Maias", "Iun.", "Iunias",
		"Quint.", "Quintilis", "Sext.", "Sextilis", "Sept.", "Septembris", "Oct.", "Octobris", "Nov.", "Novembris",
		"Dec.", "Decembris")

	return swap.Replace(thetext)
}

func splitonpunctuaton(thetext string) []string {
	// replacement for recursivesplitter(): perhaps very slightly faster, but definitely much more direct and legible
	// swap all punctuation for one item; then split on it...
	swap := strings.NewReplacer("?", ".", "!", ".", "·", ".", ";", ".")
	thetext = swap.Replace(thetext)
	split := strings.Split(thetext, ".")

	// slower way of doing the same...

	//re := regexp.MustCompile("[?!;·]")
	//thetext = re.ReplaceAllString(thetext, ".")
	//split := strings.Split(thetext, ".")

	return split
}

func acuteforgrave(thetext string) string {
	swap := strings.NewReplacer("ὰ", "ά", "ὲ", "έ", "ὶ", "ί", "ὸ", "ό", "ὺ", "ύ", "ὴ", "ή", "ὼ", "ώ",
		"ἂ", "ἄ", "ἒ", "ἔ", "ἲ", "ἴ", "ὂ", "ὄ", "ὒ", "ὔ", "ἢ", "ἤ", "ὢ", "ὤ", "ᾃ", "ᾅ", "ᾓ", "ᾕ", "ᾣ", "ᾥ",
		"ᾂ", "ᾄ", "ᾒ", "ᾔ", "ᾢ", "ᾤ")
	return swap.Replace(thetext)
}

func dropstopwords(skipper string, bagsofwords []BagWithLocus) []BagWithLocus {
	// set up the skiplist; then iterate through the bags returning new, clean bags
	s := strings.Split(skipper, " ")
	sm := make(map[string]bool)
	for i := 0; i < len(s); i++ {
		sm[s[i]] = true
	}

	for i := 0; i < len(bagsofwords); i++ {
		wl := strings.Split(bagsofwords[i].Bag, " ")
		wl = stopworddropper(sm, wl)
		bagsofwords[i].Bag = strings.Join(wl, " ")
	}

	return bagsofwords
}

func stopworddropper(stops map[string]bool, wordlist []string) []string {
	// if word is in stops, drop the word
	var returnlist []string
	for i := 0; i < len(wordlist); i++ {
		if _, t := stops[wordlist[i]]; t {
			continue
		} else {
			returnlist = append(returnlist, wordlist[i])
		}
	}
	return returnlist
}

//
// BAGGING
//

func buildflatbagsofwords(bags []BagWithLocus, parsemap map[string][]string) []BagWithLocus {
	// turn a list of sentences into a list of list of headwords; here we put alternate possibilities next to one another:
	// flatbags: ϲυγγενεύϲ ϲυγγενήϲ
	// composite: ϲυγγενεύϲ·ϲυγγενήϲ

	for i := 0; i < len(bags); i++ {
		var newwords []string
		words := strings.Split(bags[i].Bag, " ")
		for j := 0; j < len(words); j++ {
			newwords = append(newwords, parsemap[words[j]]...)
		}
		bags[i].Bag = strings.Join(newwords, " ")
	}

	return bags
}

func buildcompositebagsofwords(bags []BagWithLocus, parsemap map[string][]string) []BagWithLocus {
	// turn a list of sentences into a list of list of headwords; here we put yoked alternate possibilities next to one another:
	// flatbags: ϲυγγενεύϲ ϲυγγενήϲ
	// composite: ϲυγγενεύϲ·ϲυγγενήϲ

	for i := 0; i < len(bags); i++ {
		var newwords []string
		words := strings.Split(bags[i].Bag, " ")
		for j := 0; j < len(words); j++ {
			comp := strings.Join(parsemap[words[j]], "·")
			newwords = append(newwords, comp)
		}
		bags[i].Bag = strings.Join(newwords, " ")
	}
	return bags
}

func buildwinnertakesallbagsofwords(bags []BagWithLocus, parsemap map[string][]string, dbpool *pgxpool.Pool) []BagWithLocus {
	// turn a list of sentences into a list of list of headwords; here we figure out which headword is the dominant homonym
	// then we just use that term; "esse" always comes from "sum" and never "edo", etc.

	// [a] figure out all headwords in use

	allheadwords := make(map[string]bool)
	for i := range parsemap {
		for j := range parsemap[i] {
			allheadwords[parsemap[i][j]] = true
		}
	}

	// [b] generate scoremap and assign scores to each of the headwords

	scoremap := fetchheadwordcounts(allheadwords, dbpool)

	// [c] note that there are capital words in the parsemap that need lowering

	// [c1] lower the internal values first
	for i := range parsemap {
		for j := 0; j < len(parsemap[i]); j++ {
			parsemap[i][j] = strings.ToLower(parsemap[i][j])
		}
	}

	// [c2] lower the parsemap keys; how worried should we be about the collisions...
	lcparsemap := make(map[string][]string)
	for i := range parsemap {
		lcparsemap[strings.ToLower(i)] = parsemap[i]
	}

	// [d] run through the parsemap and kill off the losers

	newparsemap := make(map[string][]string)
	for i := range lcparsemap {
		var hwl WHWList
		for j := 0; j < len(lcparsemap[i]); j++ {
			var thishw WeightedHeadword
			thishw.Word = lcparsemap[i][j]
			thishw.Count = scoremap[lcparsemap[i][j]]
			hwl = append(hwl, thishw)
		}
		sort.Sort(hwl)

		newparsemap[i] = make([]string, 0, 1)
		newparsemap[i] = append(newparsemap[i], hwl[0].Word)
	}

	// [e] now you can just buildflatbagsofwords() with the new pruned parsemap

	bags = buildflatbagsofwords(bags, newparsemap)

	return bags
}
