//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
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
