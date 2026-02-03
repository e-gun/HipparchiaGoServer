//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package search

import (
	"fmt"
	"strings"

	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/mps"
)

// optimizesearch - think about rewriting the search to make it faster
func optimizesearch(s *str.SearchStruct) {
	// only zero or one of the following should be true

	// if BoxA has a lemma and BoxB has a phrase, it is almost certainly faster to search B, then A...
	if s.HasLemmaBoxA && s.HasPhraseBoxB {
		swapphraseandlemma(s)
		return
	}

	// all forms of an uncommon word should (usually) be sought before all forms of a common word...
	if s.HasLemmaBoxA && s.HasLemmaBoxB {
		pickfastestlemma(s)
		return
	}

	// a single word should be faster than a lemma; but do not swap an empty string
	if s.HasLemmaBoxA && !s.HasPhraseBoxB && s.Proximate != "" {
		swapwordandlemma(s)
		return
	}

	// consider looking for the string with more characters in it first
	if len(s.Seeking) > 0 && len(s.Proximate) > 0 {
		searchquickestfirst(s)
		return
	}
}

// swapphraseandlemma -  if BoxA has a lemma and BoxB has a phrase, it very likely faster to search B, then A...
func swapphraseandlemma(s *str.SearchStruct) {
	// we will swap elements and reset the relevant elements of the SearchStruct

	// no  swapphraseandlemma(): [Δ: 4.564s] lemma near phrase: 'γαῖα' near 'ἐϲχάτη χθονόϲ'
	// yes swapphraseandlemma(): [Δ: 1.276s] lemma near phrase: 'γαῖα' near 'ἐϲχάτη χθονόϲ'

	const (
		CALLED = `swapphraseandlemma() was called: lemmatized '%s' swapped with '%s'`
	)

	Msg.PEEK(fmt.Sprintf(CALLED, s.LemmaOne, s.Proximate))
	lemmaboxswap(s)
}

// swapwordandlemma - if BoxA has a lemma and BoxB has a single word, it very likely faster to search B, then A...
func swapwordandlemma(s *str.SearchStruct) {
	// [swapped]
	// Sought »χρηματα« within 1 lines of all 45 forms of »ἄνθρωποϲ«
	// Searched 7,461 works and found 298 passages (2.86s)

	// [unswapped]
	// Sought all 45 forms of »ἄνθρωποϲ« within 1 lines of »χρηματα«
	// Searched 7,461 works and found 1 passages (8.89s)

	const (
		CALLED = `swapwordandlemma() was called: lemmatized '%s' swapped with '%s'`
	)

	Msg.PEEK(fmt.Sprintf(CALLED, s.LemmaOne, s.Proximate))
	lemmaboxswap(s)
}

// pickfastestlemma - all forms of an uncommon word should (usually) be sought before all forms of a common word
func pickfastestlemma(s *str.SearchStruct) {
	// Sought all 65 forms of »δημηγορέω« within 1 lines of all 386 forms of »γιγνώϲκω«
	// swapped: 20s vs 80s

	// Sought all 68 forms of »διαμάχομαι« within 1 lines of all 644 forms of »ποιέω«
	// similar to previous: 20s vs forever...

	// Sought all 12 forms of »αὐτοκράτωρ« within 1 lines of all 50 forms of »πόλιϲ«
	// swapped: 4.17s vs 10.09s

	// it does not *always* save time to just pick the uncommon word:

	// Sought all 50 forms of »πόλιϲ« within 1 lines of all 191 forms of »ὁπλίζω«
	// this fnc will COST you 10s when you swap 33s instead of 23s.

	// the "191 forms" take longer to find than the "50 forms"; that is, the fast first pass of πόλιϲ is fast enough
	// to offset the cost of looking for ὁπλίζω among the 125274 initial hits (vs 2547 initial hits w/ ὁπλίζω run first)

	// note that it is *usually* the case that words with more forms also have more hits
	// the penalty for being wrong is relatively low; the savings when you get this right can be significant

	const (
		NOTE1 = "pickfastestlemma() is swapping %s for %s: possible hits %d < %d; known forms %d < %d"
		NOTE2 = "pickfastestlemma() is NOT swapping %s for %s: possible hits %d vs %d; known forms %d vs %d"
	)

	hw1 := db.GetIndividualHeadwordCount(s.LemmaOne)
	hw2 := db.GetIndividualHeadwordCount(s.LemmaTwo)

	// how many forms to look up?

	fc1 := 0
	fc2 := 0
	if _, ok := mps.AllLemm[s.LemmaOne]; ok {
		fc1 = len(mps.AllLemm[s.LemmaOne].Deriv)
	}
	if _, ok := mps.AllLemm[s.LemmaTwo]; ok {
		fc2 = len(mps.AllLemm[s.LemmaTwo].Deriv)
	}

	// the "&&" tries to address the »πόλιϲ« vs »ὁπλίζω« problem: see the notes above
	if (hw1.Total > hw2.Total) && (fc1 > fc2) {
		s.LemmaTwo = hw1.Word
		s.LemmaOne = hw2.Word
		Msg.PEEK(fmt.Sprintf(NOTE1, hw2.Word, hw1.Word, hw2.Total, hw1.Total, fc2, fc1))
	} else {
		Msg.PEEK(fmt.Sprintf(NOTE2, hw1.Word, hw2.Word, hw1.Total, hw2.Total, fc1, fc2))
	}
}

// searchquickestfirst - look for the string with more characters in it first; it will typically generate fewer initial hits
func searchquickestfirst(s *str.SearchStruct) {
	const (
		NOTE = "searchquickestfirst() swapping '%s' and '%s'"
	)

	// a long phrase is slower than a single word:
	// faster: Sought »ἡδονήν« within 1 lines of »τέλουϲ τῆϲ φιλοϲοφίαϲ«
	// slower: Sought »τέλουϲ τῆϲ φιλοϲοφίαϲ« within 1 lines of »ἡδονήν«

	isphraseskg := strings.Split(strings.TrimSpace(s.Seeking), " ")
	isphraseprx := strings.Split(strings.TrimSpace(s.Proximate), " ")

	test1 := len(s.Seeking) < len(s.Proximate)
	test2 := len(isphraseskg) == 1 && len(isphraseprx) == 1
	test3 := len(isphraseskg) != 1 && len(isphraseprx) != 1
	test4 := len(isphraseskg) != 1 || len(isphraseprx) != 1

	skg := s.Seeking
	prx := s.Proximate

	swap := func() {
		s.Proximate = skg
		s.Seeking = prx
		Msg.PEEK(fmt.Sprintf(NOTE, skg, prx))
	}

	// sequence of checks matters... test4 logic can't come until test3 has been cleared

	if test1 && test2 {
		// two single words
		swap()
		return
	}

	if test1 && test3 {
		// two phrases
		swap()
		return
	}

	if test4 {
		// there is a phrase in here somewhere; the other term is a single word because "two phrase" was already tested
		if len(isphraseprx) != 1 {
			// single word + a phrase
			// fastest to do nothing
		} else {
			// phrase + single word
			// quicker to swap because single words beat phrases
			swap()
		}
	}
}

// lemmaboxswap - swap 'seeking' and 'proximate' to do a lemma as the second search (in the Name of speed)
func lemmaboxswap(s *str.SearchStruct) {
	boxa := s.LemmaOne
	boxb := s.Proximate
	s.Seeking = boxb
	s.LemmaOne = ""
	s.LemmaTwo = boxa
	s.Proximate = ""

	if str.HasAccent.MatchString(boxb) {
		s.SrchColumn = "accented_line"
	} else {
		s.SrchColumn = str.DEFAULTCOLUMN
	}

	// zap some bools
	s.HasPhraseBoxA = false
	s.HasLemmaBoxA = false
	s.HasPhraseBoxB = false
	s.HasLemmaBoxB = false

	// reset the type and the bools...
	s.SetType()
}
