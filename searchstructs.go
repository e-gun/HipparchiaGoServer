//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
	"fmt"
	"regexp"
	"sort"
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

// CleanInput - remove bad chars, etc. from the submitted data
func (s *SearchStruct) CleanInput() {
	// address uv issues; lunate issues; ...
	// no need to parse a lemma: this bounces if there is not a key match to a map
	dropping := USELESSINPUT + Config.BadChars
	s.ID = Purgechars(dropping, s.ID)
	s.Seeking = strings.ToLower(s.Seeking)
	s.Proximate = strings.ToLower(s.Proximate)

	if hasAccent.MatchString(s.Seeking) || hasAccent.MatchString(s.Proximate) {
		// lemma search will select accented automatically
		s.SrchColumn = "accented_line"
	}

	rs := []rune(s.Seeking)
	if len(rs) > MAXINPUTLEN {
		s.Seeking = string(rs[0:MAXINPUTLEN])
	}

	rp := []rune(s.Proximate)
	if len(rp) > MAXINPUTLEN {
		s.Proximate = string(rs[0:MAXINPUTLEN])
	}

	s.Seeking = UVσςϲ(s.Seeking)
	s.Proximate = UVσςϲ(s.Proximate)

	s.Seeking = Purgechars(dropping, s.Seeking)
	s.Proximate = Purgechars(dropping, s.Proximate)

	// don't let BoxA be blank if BoxB is not
	BoxA := s.Seeking == "" && s.LemmaOne == ""
	NotBoxB := s.Proximate != "" || s.LemmaTwo != ""

	if BoxA && NotBoxB {
		if s.Proximate != "" {
			s.Seeking = s.Proximate
			s.Proximate = ""
		}
		if s.LemmaTwo != "" {
			s.LemmaOne = s.LemmaTwo
			s.LemmaTwo = ""
		}
	}
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
		if isGreek.MatchString(s.LemmaOne) {
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

// FormatInitialSummary - build HTML for the search summary
func (s *SearchStruct) FormatInitialSummary() {
	// ex:
	// Sought <span class="sought">»ἡμέρα«</span> within 2 lines of all 79 forms of <span class="sought">»ἀγαθόϲ«</span>
	const (
		TPM = `Sought %s<span class="sought">»%s«</span>%s`
		WIN = `%s within %d %s of %s<span class="sought">»%s«</span>`
		ADF = "all %d forms of "
	)

	yn := ""
	if s.NotNear {
		yn = " not "
	}

	af1 := ""
	sk := s.Seeking
	if len(s.LemmaOne) != 0 {
		sk = s.LemmaOne
		if _, ok := AllLemm[sk]; ok {
			af1 = fmt.Sprintf(ADF, len(AllLemm[sk].Deriv))
		}
	}

	two := ""
	if s.Twobox {
		sk2 := s.Proximate
		af2 := ""
		if len(s.LemmaTwo) != 0 {
			sk2 = s.LemmaTwo
			if _, ok := AllLemm[sk]; ok {
				af2 = fmt.Sprintf(ADF, len(AllLemm[sk].Deriv))
			}
		}
		two = fmt.Sprintf(WIN, yn, s.ProxDist, s.ProxScope, af2, sk2)
	}
	sum := fmt.Sprintf(TPM, af1, sk, two)
	s.InitSum = sum
}

// InclusionOverview - yield a summary of the inclusions; NeighborsSearch will use this when calling buildblanknngraph()
func (s *SearchStruct) InclusionOverview(sessincl SearchIncExl) string {
	// possible to get burned, but this cheat is "good enough"
	// hipparchiaDB=# SELECT COUNT(universalid) FROM authors WHERE universalid LIKE 'gr%';
	// gr: 1823
	// lt: 362
	// in: 463
	// ch: 291
	// dp: 516

	const (
		MAXITEMS = 4
		GRCT     = 1823
		LTCT     = 362
		INCT     = 463
		CHCT     = 291
		DPCT     = 516
		FULL     = "all %d of the %s tables"
	)

	in := s.SearchIn
	in.BuildAuByName()
	in.BuildWkByName()

	// the named passages are available to a SeverSession, not a SearchStruct
	namemap := sessincl.MappedPsgByName
	var nameslc []string
	for _, v := range namemap {
		nameslc = append(nameslc, v)
	}
	sort.Strings(nameslc)

	var ov []string
	ov = append(ov, in.AuGenres...)
	ov = append(ov, in.WkGenres...)
	ov = append(ov, in.ListedABN...)
	ov = append(ov, in.ListedWBN...)
	ov = append(ov, nameslc...)

	notall := func() string {
		sort.Strings(ov)

		var enum []string

		if len(ov) != 1 {
			for i, p := range ov {
				enum = append(enum, fmt.Sprintf("(%d) %s", i+1, p))
			}
		} else {
			enum = append(enum, fmt.Sprintf("%s", ov[0]))
		}

		if len(enum) > MAXITEMS {
			diff := len(enum) - MAXITEMS
			enum = enum[0:MAXITEMS]
			enum = append(enum, fmt.Sprintf("and %d others", diff))
		}

		o := strings.Join(enum, "; ")
		nomarkup := strings.NewReplacer("<i>", "", "</i>", "")
		return nomarkup.Replace(o)
	}

	tt := len(ov)
	if tt != len(in.Authors) {
		tt = -1
	}

	r := ""
	switch tt {
	case GRCT:
		r = fmt.Sprintf(FULL, GRCT, "Greek author")
	case LTCT:
		r = fmt.Sprintf(FULL, LTCT, "Latin author")
	case INCT:
		r = fmt.Sprintf(FULL, INCT, "classical inscriptions")
	case DPCT:
		r = fmt.Sprintf(FULL, DPCT, "documentary papyri")
	case CHCT:
		r = fmt.Sprintf(FULL, CHCT, "christian era inscriptions")
	default:
		r = notall()
	}

	return r
}

// Optimize - think about rewriting the search to make it faster
func (s *SearchStruct) Optimize() {
	// only zero or one of the following should be true

	// if BoxA has a lemma and BoxB has a phrase, it is almost certainly faster to search B, then A...
	if s.HasLemmaBoxA && s.HasPhraseBoxB {
		s.SwapPhraseAndLemma()
		return
	}

	// all forms of an uncommon word should (usually) be sought before all forms of a common word...
	if s.HasLemmaBoxA && s.HasLemmaBoxB {
		s.PickFastestLemma()
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
func (s *SearchStruct) PickFastestLemma() {
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

	hw1 := headwordlookup(s.LemmaOne)
	hw2 := headwordlookup(s.LemmaTwo)

	// how many forms to look up?
	fc1 := len(AllLemm[s.LemmaOne].Deriv)
	fc2 := len(AllLemm[s.LemmaTwo].Deriv)

	// the "&&" tries to address the »πόλιϲ« vs »ὁπλίζω« problem: see the notes above
	if (hw1.Total > hw2.Total) && (fc1 > fc2) {
		s.LemmaTwo = hw1.Entry
		s.LemmaOne = hw2.Entry
		msg(fmt.Sprintf(NOTE1, hw2.Entry, hw1.Entry, hw2.Total, hw1.Total, fc2, fc1), MSGPEEK)
	} else {
		msg(fmt.Sprintf(NOTE2, hw1.Entry, hw2.Entry, hw1.Total, hw2.Total, fc1, fc2), MSGPEEK)
	}
}

// LemmaBoxSwap - swap 'seeking' and 'proximate' to do a lemma as the second search (in the name of speed)
func (s *SearchStruct) LemmaBoxSwap() {
	boxa := s.LemmaOne
	boxb := s.Proximate
	s.Seeking = boxb
	s.LemmaOne = ""
	s.LemmaTwo = boxa
	s.Proximate = ""

	if hasAccent.MatchString(boxb) {
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

	msg(fmt.Sprintf(CALLED, s.LemmaOne, s.Proximate), MSGPEEK)
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

	msg(fmt.Sprintf(CALLED, s.LemmaOne, s.Proximate), MSGPEEK)
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
		msg(fmt.Sprintf(NOTE, skg, prx), MSGPEEK)
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

// SortResults - sort the search results by the session's registerselection criterion
func (s *SearchStruct) SortResults() {
	// Closures that order the DbWorkline structure:
	// see generichelpers.go and https://pkg.go.dev/sort#example__sortMultiKeys

	const (
		NULL = `Unavailable`
	)

	nameIncreasing := func(one, two *DbWorkline) bool {
		a1 := one.MyAu().Shortname
		a2 := two.MyAu().Shortname
		return a1 < a2
	}

	titleIncreasing := func(one, two *DbWorkline) bool {
		return one.MyWk().Title < two.MyWk().Title
	}

	dateIncreasing := func(one, two *DbWorkline) bool {
		d1 := one.MyWk().RecDate
		d2 := two.MyWk().RecDate
		if d1 != NULL && d2 != NULL {
			return one.MyWk().ConvDate < two.MyWk().ConvDate
		} else if d1 == NULL && d2 != NULL {
			return one.MyAu().ConvDate < two.MyAu().ConvDate
		} else if d1 != NULL && d2 == NULL {
			return one.MyAu().ConvDate < two.MyAu().ConvDate
		} else {
			return one.MyAu().ConvDate < two.MyAu().ConvDate
		}
	}

	increasingLines := func(one, two *DbWorkline) bool {
		return one.TbIndex < two.TbIndex
	}

	increasingID := func(one, two *DbWorkline) bool {
		return one.BuildHyperlink() < two.BuildHyperlink()
	}

	increasingWLOC := func(one, two *DbWorkline) bool {
		return one.MyWk().Prov < two.MyWk().Prov
	}

	sess := AllSessions.GetSess(s.User)
	crit := sess.SortHitsBy

	switch {
	case crit == "shortname":
		WLOrderedBy(nameIncreasing, titleIncreasing, increasingLines).Sort(s.Results.Lines)
	case crit == "converted_date":
		WLOrderedBy(dateIncreasing, nameIncreasing, titleIncreasing, increasingLines).Sort(s.Results.Lines)
	case crit == "universalid":
		WLOrderedBy(increasingID).Sort(s.Results.Lines)
	case crit == "provenance":
		// as this is likely an inscription search, why not sort next by date?
		WLOrderedBy(increasingWLOC, dateIncreasing).Sort(s.Results.Lines)
	default:
		// author nameIncreasing
		WLOrderedBy(nameIncreasing, increasingLines).Sort(s.Results.Lines)
	}
}
