//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vec

import (
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"math/rand"
	"sort"
	"strings"
)

//
// PARSEMAPPERS
//

//
// this code uses many large maps; it is possible that https://github.com/dolthub/swiss might one day speed this up
//

// note that building these can be quite slow when you send 1m lines into the modeler
// it might be possible to speed them up, but they are still <25% of the model building time so...

type hwguesser struct {
	total int
	words map[int]string
}

// buildmontecarloparsemap
func buildmontecarloparsemap(parsemap map[string]map[string]bool) map[string]hwguesser {
	// turn a list of sentences into a list of headwords; here we figure out the chances of any given homonym
	// then we set ourselves up to do a weighted guess of which one is in use

	// if a word might be A, B, or C and A appears 50 times, B appears 25 times, and C appears 5 times, then you
	// want to randomly assign the word to A 5/8 of the time, etc.
	// this can be achieved by a cumulative weight: [A -> 1-50, B -> 51-75, C -> 76-80]; a guess of 66 is a "B"

	// [a] figure out all headwords in use

	allheadwords := make(map[string]bool)
	for i := range parsemap {
		for k, _ := range parsemap[i] {
			allheadwords[k] = true
		}
	}

	// [b] generate scoremap and assign scores to each of the headwords

	scoremap := db.MapHeadwordCounts(allheadwords)

	// [c] note that there are capital words in the parsemap that need lowering

	// [c1] lower the internal values first
	for i := range parsemap {
		newmap := make(map[string]bool)
		for k, _ := range parsemap[i] {
			newmap[strings.ToLower(k)] = true
		}
		parsemap[i] = newmap
	}

	// [c2] lower the parsemap keys; how worried should we be about the collisions...
	lcparsemap := make(map[string]map[string]bool)
	for i := range parsemap {
		lcparsemap[strings.ToLower(i)] = parsemap[i]
	}

	// [d] run through the parsemap convert to a hwguesser map

	guessermap := make(map[string]hwguesser)
	for i := range lcparsemap {
		var g hwguesser
		g.words = make(map[int]string)
		t := 0
		for j, _ := range parsemap[i] {
			t += scoremap[j]
			g.words[t] = j
		}
		g.total = t
		guessermap[i] = g
	}

	return guessermap
}

// buildwinnertakesallparsemap - figure out which is the most common of the possible headwords for any given word
func buildwinnertakesallparsemap(parsemap map[string]map[string]bool) map[string]string {
	// turn a list of sentences into a list of headwords; here we figure out which headword is the dominant homonym
	// then we just use that term; "esse" always comes from "sum" and never "edo", etc.

	// [a] figure out all headwords in use

	allheadwords := make(map[string]bool)
	for i := range parsemap {
		for k, _ := range parsemap[i] {
			allheadwords[k] = true
		}
	}

	// [b] generate scoremap and assign scores to each of the headwords

	scoremap := db.MapHeadwordCounts(allheadwords)

	// [c] note that there are capital words in the parsemap that need lowering

	// [c1] lower the internal values first
	for i := range parsemap {
		newmap := make(map[string]bool)
		for k, _ := range parsemap[i] {
			newmap[strings.ToLower(k)] = true
		}
		parsemap[i] = newmap
	}

	// [c2] lower the parsemap keys; how worried should we be about the collisions...
	lcparsemap := make(map[string]map[string]bool)
	for i := range parsemap {
		lcparsemap[strings.ToLower(i)] = parsemap[i]
	}

	// [d] run through the parsemap and kill off the losers

	winnermap := make(map[string]string)
	for i := range lcparsemap {
		var hwl str.WHWList
		for j, _ := range parsemap[i] {
			var thishw str.WeightedHeadword
			thishw.Word = j
			thishw.Count = scoremap[j]
			hwl = append(hwl, thishw)
		}
		sort.Sort(hwl)
		winnermap[i] = hwl[0].Word
	}

	return winnermap
}

// buildyokedparsemap - set of sentences --> set of headwords; accept all headwords and yoke them: "esse" is "sum" + "edo", etc.
func buildyokedparsemap(parsemap map[string]map[string]bool) map[string]string {
	const (
		SEPARATOR = `ˣ` // wrong and nlp.NewPipeline(vectoriser, lda) will split yokedbags: spondeum•spondeus --> spondeum:25 spondeus:26
	)

	// [a] figure out all headwords in use

	allheadwords := make(map[string]bool)
	for i := range parsemap {
		for k, _ := range parsemap[i] {
			allheadwords[k] = true
		}
	}

	// [b] note that there are capital words in the parsemap that need lowering

	// [b1] lower the internal values first
	for i := range parsemap {
		newmap := make(map[string]bool)
		for k, _ := range parsemap[i] {
			newmap[strings.ToLower(k)] = true
		}
		parsemap[i] = newmap
	}

	// [b2] lower the parsemap keys; how worried should we be about the collisions...
	lcparsemap := make(map[string]map[string]bool)
	for i := range parsemap {
		lcparsemap[strings.ToLower(i)] = parsemap[i]
	}

	// [c] build the yoked map

	yoked := make(map[string]string)
	for i := range lcparsemap {
		var ww []string
		for j, _ := range parsemap[i] {
			ww = append(ww, j)
		}
		sort.Strings(ww)

		yoked[i] = strings.Join(ww, SEPARATOR)
	}

	return yoked
}

//
// buildtextblock() HELPERS
//

// flatstring - helper for buildtextblock() to generate unmodified text
func flatstring(sb *strings.Builder, slicedwords []string) {
	stops := getstopset()
	for i := 0; i < len(slicedwords); i++ {
		// drop skipwords
		_, s := stops[slicedwords[i]]
		if s {
			continue
		} else {
			sb.WriteString(slicedwords[i] + " ")
		}
	}

	for i := 0; i < len(slicedwords); i++ {
		sb.WriteString(slicedwords[i] + " ")
	}
}

// yokedstring - helper for buildtextblock() to generate conjoined string substitutions
func yokedstring(sb *strings.Builder, slicedwords []string, yokedmap map[string]string, stops map[string]struct{}) {
	// exact same logic as winnerstring()
	winnerstring(sb, slicedwords, yokedmap, stops)
}

// winnerstring - helper for buildtextblock() to generate winner takes all substitutions
func winnerstring(sb *strings.Builder, slicedwords []string, winnermap map[string]string, stops map[string]struct{}) {
	for i := 0; i < len(slicedwords); i++ {
		// drop skipwords
		w := winnermap[slicedwords[i]]
		_, s := stops[w]
		if s {
			continue
		} else {
			sb.WriteString(w + " ")
		}
	}
}

// montecarlostring - helper for buildtextblock() to generate lucky-ducky substitutions
func montecarlostring(sb *strings.Builder, slicedwords []string, guessermap map[string]hwguesser, stops map[string]struct{}) {
	var w string
	for i := 0; i < len(slicedwords); i++ {
		w = ""
		// pick a word...
		mc := guessermap[slicedwords[i]]
		if mc.total > 0 {
			g := rand.Intn(mc.total)
			for k, v := range mc.words {
				if g < k {
					w = v
					break
				}
			}
		} else {
			// just grab the first one
			for _, v := range mc.words {
				w = v
				break
			}
		}

		if w == "" {
			w = slicedwords[i]
		}

		_, s := stops[w]
		if s {
			continue
		} else {
			sb.WriteString(w + " ")
		}
	}
}
