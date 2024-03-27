package search

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/mps"
)

// OptimizeSrearch - think about rewriting the search to make it faster
func OptimizeSrearch(s *str.SearchStruct) {
	// only zero or one of the following should be true

	// if BoxA has a lemma and BoxB has a phrase, it is almost certainly faster to search B, then A...
	if s.HasLemmaBoxA && s.HasPhraseBoxB {
		s.SwapPhraseAndLemma()
		return
	}

	// all forms of an uncommon word should (usually) be sought before all forms of a common word...
	if s.HasLemmaBoxA && s.HasLemmaBoxB {
		PickFastestLemma(s)
		return
	}

	// a single word should be faster than a lemma; but do not swap an empty string
	if s.HasLemmaBoxA && !s.HasPhraseBoxB && s.Proximate != "" {
		s.SwapWordAndLemma()
		return
	}

	// consider looking for the string with more characters in it first
	if len(s.Seeking) > 0 && len(s.Proximate) > 0 {
		s.SearchQuickestFirst()
		return
	}
}

// PickFastestLemma - all forms of an uncommon word should (usually) be sought before all forms of a common word
func PickFastestLemma(s *str.SearchStruct) {
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
		NOTE1 = "PickFastestLemma() is swapping %s for %s: possible hits %d < %d; known forms %d < %d"
		NOTE2 = "PickFastestLemma() is NOT swapping %s for %s: possible hits %d vs %d; known forms %d vs %d"
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
		s.LemmaTwo = hw1.Entry
		s.LemmaOne = hw2.Entry
		Msg.PEEK(fmt.Sprintf(NOTE1, hw2.Entry, hw1.Entry, hw2.Total, hw1.Total, fc2, fc1))
	} else {
		Msg.PEEK(fmt.Sprintf(NOTE2, hw1.Entry, hw2.Entry, hw1.Total, hw2.Total, fc1, fc2))
	}
}
