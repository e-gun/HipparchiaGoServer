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

	Msg.PEEK(fmt.Sprintf(CALLED, s.LemmaOne, s.Proximate))
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

	Msg.PEEK(fmt.Sprintf(CALLED, s.LemmaOne, s.Proximate))
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

// showinterimresults - print out the current results
func showinterimresults(s *SearchStruct) {
	const (
		NB  = "showinterimresults()"
		FMT = "[%d] %s\t%s\t%s"
	)

	Msg.WARN(NB)

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
		Msg.NOTE(v)
	}
}
