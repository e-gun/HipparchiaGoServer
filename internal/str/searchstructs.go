//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package str

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
)

//
// SEARCHSTRUCTS
//

type SearchStruct struct {
	User          string
	IPAddr        string
	ID            string
	WSID          string
	Seeking       string
	Proximate     string
	LemmaOne      string
	LemmaTwo      string
	InitSum       string
	Summary       string
	ProxScope     string // "lines" or "words"
	ProxType      string // "near" or "not near"
	ProxDist      int
	HasLemmaBoxA  bool
	HasLemmaBoxB  bool
	HasPhraseBoxA bool
	HasPhraseBoxB bool
	IsLemmAndPhr  bool
	OneHit        bool
	Twobox        bool
	NotNear       bool
	SkgRewritten  bool
	Type          string
	PhaseNum      int
	SrchColumn    string // usually "stripped_line", sometimes "accented_line"
	SrchSyntax    string // almost always "~"
	OrderBy       string // almost always "index" + ASC
	VecTextPrep   string
	VecModeler    string
	CurrentLimit  int
	OriginalLimit int
	SkgSlice      []string // either just Seeking or a decomposed version of a Lemma's possibilities
	PrxSlice      []string
	SearchIn      SearchIncExl
	SearchEx      SearchIncExl
	Queries       []PrerolledQuery
	Results       WorkLineBundle // pointer here yields problem: WithinXLinesSearch() has S1 & S2 where S2 is a modified copy of S1
	Launched      time.Time
	TTName        string
	SearchSize    int // # of works searched
	TableSize     int // # of tables searched
	ExtraMsg      string
	StoredSession ServerSession
	RealIP        string
	Context       context.Context
	CancelFnc     context.CancelFunc
	IsActive      bool
}

// SetType - set internal values via self-probe
func (s *SearchStruct) SetType() {
	const (
		ACC = `ϲῥἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάἐἑἒἓἔἕὲέἰἱἲἳἴἵἶἷὶίῐῑῒΐῖῗὀὁὂὃὄὅόὸὐὑὒὓὔὕὖὗϋῠῡῢΰῦῧύὺᾐᾑᾒᾓᾔᾕᾖᾗῂῃῄῆῇἤἢἥἣὴήἠἡἦἧὠὡὢὣὤὥὦὧᾠᾡᾢᾣᾤᾥᾦᾧῲῳῴῶῷώὼ`
		REG = `a-zα-ω`
	)
	ps := s.Proximate != ""
	psl := s.LemmaTwo != ""

	if ps || psl {
		s.Twobox = true
	}

	comp := fmt.Sprintf(`[%s%s]\s[%s%s]`, REG, ACC, REG, ACC)
	twowords := regexp.MustCompile(comp)

	if twowords.MatchString(s.Seeking) {
		s.HasPhraseBoxA = true
	}

	if twowords.MatchString(s.Proximate) {
		s.HasPhraseBoxB = true
	}

	if s.LemmaOne != "" {
		s.HasLemmaBoxA = true
		// accented line has "volat" in latin; and "uolo" will not find it
		if IsGreek.MatchString(s.LemmaOne) {
			s.SrchColumn = "accented_line"
		}
	}

	if s.LemmaTwo != "" {
		s.HasLemmaBoxB = true
	}

	if (s.HasLemmaBoxA && s.HasPhraseBoxB) || (s.HasLemmaBoxB && s.HasPhraseBoxA) {
		s.IsLemmAndPhr = true
	}

	return
}

// todo: refactor to avoid circularity
// Optimize - think about rewriting the search to make it faster
//func (s *SearchStruct) Optimize() {
//	// only zero or one of the following should be true
//
//	// if BoxA has a lemma and BoxB has a phrase, it is almost certainly faster to search B, then A...
//	if s.HasLemmaBoxA && s.HasPhraseBoxB {
//		s.SwapPhraseAndLemma()
//		return
//	}
//
//	// all forms of an uncommon word should (usually) be sought before all forms of a common word...
//	if s.HasLemmaBoxA && s.HasLemmaBoxB {
//		s.PickFastestLemma()
//		return
//	}
//
//	// a single word should be faster than a lemma; but do not swap an empty string
//	if s.HasLemmaBoxA && !s.HasPhraseBoxB && s.Proximate != "" {
//		s.SwapWordAndLemma()
//		return
//	}
//
//	// consider looking for the string with more characters in it first
//	if len(s.Seeking) > 0 && len(s.Proximate) > 0 {
//		s.SearchQuickestFirst()
//		return
//	}
//}

// todo: refactor to avoid circularity
// PickFastestLemma - all forms of an uncommon word should (usually) be sought before all forms of a common word
//func (s *SearchStruct) PickFastestLemma() {
//	// Sought all 65 forms of »δημηγορέω« within 1 lines of all 386 forms of »γιγνώϲκω«
//	// swapped: 20s vs 80s
//
//	// Sought all 68 forms of »διαμάχομαι« within 1 lines of all 644 forms of »ποιέω«
//	// similar to previous: 20s vs forever...
//
//	// Sought all 12 forms of »αὐτοκράτωρ« within 1 lines of all 50 forms of »πόλιϲ«
//	// swapped: 4.17s vs 10.09s
//
//	// it does not *always* save time to just pick the uncommon word:
//
//	// Sought all 50 forms of »πόλιϲ« within 1 lines of all 191 forms of »ὁπλίζω«
//	// this fnc will COST you 10s when you swap 33s instead of 23s.
//
//	// the "191 forms" take longer to find than the "50 forms"; that is, the fast first pass of πόλιϲ is fast enough
//	// to offset the cost of looking for ὁπλίζω among the 125274 initial hits (vs 2547 initial hits w/ ὁπλίζω run first)
//
//	// note that it is *usually* the case that words with more forms also have more hits
//	// the penalty for being wrong is relatively low; the savings when you get this right can be significant
//
//	const (
//		NOTE1 = "PickFastestLemma() is swapping %s for %s: possible hits %d < %d; known forms %d < %d"
//		NOTE2 = "PickFastestLemma() is NOT swapping %s for %s: possible hits %d vs %d; known forms %d vs %d"
//	)
//
//	hw1 := headwordlookup(s.LemmaOne)
//	hw2 := headwordlookup(s.LemmaTwo)
//
//	// how many forms to look up?
//	fc1 := len(AllLemm[s.LemmaOne].Deriv)
//	fc2 := len(AllLemm[s.LemmaTwo].Deriv)
//
//	// the "&&" tries to address the »πόλιϲ« vs »ὁπλίζω« problem: see the notes above
//	if (hw1.Total > hw2.Total) && (fc1 > fc2) {
//		s.LemmaTwo = hw1.Entry
//		s.LemmaOne = hw2.Entry
//		msg.PEEK(fmt.Sprintf(NOTE1, hw2.Entry, hw1.Entry, hw2.Total, hw1.Total, fc2, fc1))
//	} else {
//		msg.PEEK(fmt.Sprintf(NOTE2, hw1.Entry, hw2.Entry, hw1.Total, hw2.Total, fc1, fc2))
//	}
//}

// LemmaBoxSwap - swap 'seeking' and 'proximate' to do a lemma as the second search (in the Name of speed)
func (s *SearchStruct) LemmaBoxSwap() {
	boxa := s.LemmaOne
	boxb := s.Proximate
	s.Seeking = boxb
	s.LemmaOne = ""
	s.LemmaTwo = boxa
	s.Proximate = ""

	if HasAccent.MatchString(boxb) {
		s.SrchColumn = "accented_line"
	} else {
		s.SrchColumn = DEFAULTCOLUMN
	}

	// zap some bools
	s.HasPhraseBoxA = false
	s.HasLemmaBoxA = false
	s.HasPhraseBoxB = false
	s.HasLemmaBoxB = false

	// reset the type and the bools...
	s.SetType()
}

// SwapPhraseAndLemma -  if BoxA has a lemma and BoxB has a phrase, it very likely faster to search B, then A...
func (s *SearchStruct) SwapPhraseAndLemma() {
	// we will swap elements and reset the relevant elements of the SearchStruct

	// no  SwapPhraseAndLemma(): [Δ: 4.564s] lemma near phrase: 'γαῖα' near 'ἐϲχάτη χθονόϲ'
	// yes SwapPhraseAndLemma(): [Δ: 1.276s] lemma near phrase: 'γαῖα' near 'ἐϲχάτη χθονόϲ'

	const (
		CALLED = `SwapPhraseAndLemma() was called: lemmatized '%s' swapped with '%s'`
	)

	msg.PEEK(fmt.Sprintf(CALLED, s.LemmaOne, s.Proximate))
	s.LemmaBoxSwap()
}

// SwapWordAndLemma - if BoxA has a lemma and BoxB has a single word, it very likely faster to search B, then A...
func (s *SearchStruct) SwapWordAndLemma() {
	// [swapped]
	// Sought »χρηματα« within 1 lines of all 45 forms of »ἄνθρωποϲ«
	// Searched 7,461 works and found 298 passages (2.86s)

	// [unswapped]
	// Sought all 45 forms of »ἄνθρωποϲ« within 1 lines of »χρηματα«
	// Searched 7,461 works and found 1 passages (8.89s)

	const (
		CALLED = `SwapWordAndLemma() was called: lemmatized '%s' swapped with '%s'`
	)

	msg.PEEK(fmt.Sprintf(CALLED, s.LemmaOne, s.Proximate))
	s.LemmaBoxSwap()
}

// SearchQuickestFirst - look for the string with more characters in it first; it will typically generate fewer initial hits
func (s *SearchStruct) SearchQuickestFirst() {
	const (
		NOTE = "SearchQuickestFirst() swapping '%s' and '%s'"
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
		msg.PEEK(fmt.Sprintf(NOTE, skg, prx))
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

// showinterimresults - print out the current results
func showinterimresults(s *SearchStruct) {
	const (
		NB  = "showinterimresults()"
		FMT = "[%d] %s\t%s\t%s"
	)

	msg.WARN(NB)

	mp := make(map[string]DbWorkline, s.Results.Len())
	kk := make([]string, s.Results.Len())

	for i := 0; i < s.Results.Len(); i++ {
		r := s.Results.Lines[i]
		mp[r.BuildHyperlink()] = r
		kk[i] = r.BuildHyperlink()
	}

	slices.Sort(kk)

	for i, k := range kk {
		r := mp[k]
		v := fmt.Sprintf(FMT, i, r.BuildHyperlink(), s.Seeking, r.MarkedUp)
		msg.NOTE(v)
	}
}
