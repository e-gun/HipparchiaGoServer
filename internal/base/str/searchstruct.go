//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package str

import (
	"context"
	"fmt"
	"regexp"
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
		TWO = `[%s%s]\s[%s%s]`
		ALN = "accented_line"
	)
	ps := s.Proximate != ""
	psl := s.LemmaTwo != ""

	if ps || psl {
		s.Twobox = true
	}

	comp := fmt.Sprintf(TWO, REG, ACC, REG, ACC)
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
			s.SrchColumn = ALN
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

func (s *SearchStruct) DedupeResults() {
	// LemmaIntoRegexSlice() might mean you are doing many searches
	// `seruus¹` in Plautus will come back as 3 collections of strings for which to search
	// but then you can end up with non-unique hits since (alas) some of those search strings are non-unique
	// here we toss the duplicates

	// you only need to do this check `if len(s.SkgSlice) > 1`

	dedupmap := map[string]bool{}
	dupemap := map[string]bool{}
	var dedupedlines WorkLineBundle
	ll := s.Results.Yield()
	dupsfound := 0

	for l := range ll {
		_, present := dedupmap[l.BuildHyperlink()]
		if !present {
			dedupmap[l.BuildHyperlink()] = true
			dedupedlines.AppendOne(l)
		} else {
			dupsfound++
			dupemap[l.BuildHyperlink()] = true
		}
	}

	Msg.TMI(fmt.Sprintf("DedupeResults(): %d duplicate results found out of %d total results", dupsfound, s.Results.Len()))

	s.Results = dedupedlines
}
