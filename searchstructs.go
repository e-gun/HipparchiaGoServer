package main

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

//
// SEARCHSTRUCTS
//

type SearchStruct struct {
	User          string
	ID            string
	Seeking       string
	Proximate     string
	LemmaOne      string
	LemmaTwo      string
	InitSum       string
	Summary       string
	ProxScope     string // "lines" or "words"
	ProxType      string // "near" or "not near"
	ProxDist      int
	HasLemma      bool
	HasPhrase     bool
	IsVector      bool
	IsActive      bool
	OneHit        bool
	Twobox        bool
	NotNear       bool
	SkgRewritten  bool
	PhaseNum      int
	SrchColumn    string // usually "stripped_line", sometimes "accented_line"
	SrchSyntax    string // almost always "~"
	OrderBy       string // almost always "index" + ASC
	CurrentLimit  int
	OriginalLimit int
	SkgSlice      []string // either just Seeking or a decomposed version of a Lemma's possibilities
	PrxSlice      []string
	SearchIn      SearchIncExl
	SearchEx      SearchIncExl
	Queries       []PrerolledQuery
	Results       []DbWorkline
	Launched      time.Time
	TTName        string
	SearchSize    int // # of works searched
	TableSize     int // # of tables searched
	ExtraMsg      string
	Hits          *SrchCounter
	Remain        *SrchCounter
	lock          *sync.RWMutex
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
}

// SetType - set internal values via self-probe
func (s *SearchStruct) SetType() {
	// skip detailed proximate checks because second pass search just feeds all of that into the primary fields
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
		s.HasPhrase = true
	}

	if len(s.LemmaOne) != 0 {
		s.HasLemma = true
		// accented line has "volat" in latin; and "uolo" will not find it
		if isGreek.MatchString(s.LemmaOne) {
			s.SrchColumn = "accented_line"
		}
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
		af1 = fmt.Sprintf(ADF, len(AllLemm[sk].Deriv))
	}

	two := ""
	if s.Twobox {
		sk2 := s.Proximate
		af2 := ""
		if len(s.LemmaTwo) != 0 {
			sk2 = s.LemmaTwo
			af2 = fmt.Sprintf(ADF, len(AllLemm[sk2].Deriv))
		}
		two = fmt.Sprintf(WIN, yn, s.ProxDist, s.ProxScope, af2, sk2)
	}
	sum := fmt.Sprintf(TPM, af1, sk, two)
	s.InitSum = sum
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

	sess := SafeSessionRead(s.User)
	crit := sess.SortHitsBy

	switch {
	case crit == "shortname":
		WLOrderedBy(nameIncreasing, titleIncreasing, increasingLines).Sort(s.Results)
	case crit == "converted_date":
		WLOrderedBy(dateIncreasing, nameIncreasing, titleIncreasing, increasingLines).Sort(s.Results)
	case crit == "universalid":
		WLOrderedBy(increasingID).Sort(s.Results)
	case crit == "provenance":
		// as this is likely an inscription search, why not sort next by date?
		WLOrderedBy(increasingWLOC, dateIncreasing).Sort(s.Results)
	default:
		// author nameIncreasing
		WLOrderedBy(nameIncreasing, increasingLines).Sort(s.Results)
	}
}

// AcqHitCounter - get a SrchCounter for storing Hits values
func (s *SearchStruct) AcqHitCounter() {
	h := func() *SrchCounter { return &SrchCounter{} }()
	s.Hits = h
}

// AcqRemainCounter - get a SrchCounter for storing Remain values
func (s *SearchStruct) AcqRemainCounter() {
	r := func() *SrchCounter { return &SrchCounter{} }()
	s.Remain = r
}

//
// SEARCHCOUNTERS
//

type SrchCounter struct {
	// atomic package could do this more simply, but this architecture is more flexible in the long term
	count int
	lock  sync.RWMutex
}

// Get - concurrency aware way to read a SrchCounter
func (h *SrchCounter) Get() int {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.count
}

// Set - concurrency aware way to write to a SrchCounter
func (h *SrchCounter) Set(c int) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.count = c
}

//
// THREAD SAFE INFRASTRUCTURE: MUTEX
// (and not channel: https://github.com/golang/go/wiki/MutexOrChannel)
//

// SrchInfo - struct used to deliver info about searches in progress
type SrchInfo struct {
	ID        string
	Exists    bool
	Hits      int
	Remain    int
	Summary   string
	SrchCount int
}

// SearchVault - there should be only one of these; and it contains all the searches
type SearchVault struct {
	SearchMap    map[string]SearchStruct
	SearchLocker sync.RWMutex
}

func (sv *SearchVault) Insert(s SearchStruct) {
	sv.SearchLocker.Lock()
	defer sv.SearchLocker.Unlock()
	sv.SearchMap[s.ID] = s
}

func (sv *SearchVault) Get(id string) SearchStruct {
	sv.SearchLocker.Lock()
	defer sv.SearchLocker.Unlock()
	s, e := sv.SearchMap[id]
	if e != true {
		s = BuildHollowSearch()
		s.ID = id
		s.IsActive = false
	}
	return s
}

func (sv *SearchVault) Delete(id string) {
	sv.SearchLocker.Lock()
	defer sv.SearchLocker.Unlock()
	delete(sv.SearchMap, id)
}

func (sv *SearchVault) GetInfo(id string) SrchInfo {
	sv.SearchLocker.Lock()
	defer sv.SearchLocker.Unlock()
	var m SrchInfo
	_, m.Exists = sv.SearchMap[id]
	if m.Exists {
		m.Remain = sv.SearchMap[id].Remain.Get()
		m.Hits = sv.SearchMap[id].Hits.Get()
		m.Summary = sv.SearchMap[id].InitSum
	}
	m.SrchCount = len(sv.SearchMap)
	return m
}

func (sv *SearchVault) Exists(id string) bool {
	sv.SearchLocker.Lock()
	defer sv.SearchLocker.Unlock()
	_, exists := SearchMap[id]
	return exists
}

func (sv *SearchVault) Count() int {
	sv.SearchLocker.Lock()
	defer sv.SearchLocker.Unlock()
	return len(SearchMap)
}

func (sv *SearchVault) GetRemain(id string) int {
	sv.SearchLocker.Lock()
	defer sv.SearchLocker.Unlock()
	return sv.SearchMap[id].Remain.Get()
}

func (sv *SearchVault) SetRemain(id string, r int) {
	sv.SearchLocker.Lock()
	defer sv.SearchLocker.Unlock()
	sv.SearchMap[id].Remain.Set(r)
}

func (sv *SearchVault) GetHits(id string) int {
	sv.SearchLocker.Lock()
	defer sv.SearchLocker.Unlock()
	return sv.SearchMap[id].Hits.Get()
}

func (sv *SearchVault) InitSum(id string) string {
	sv.SearchLocker.Lock()
	defer sv.SearchLocker.Unlock()
	return sv.SearchMap[id].InitSum
}
