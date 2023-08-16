package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

//
// SEARCHSTRUCTS
//

type SearchStruct struct {
	User          string
	IPAddr        string
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
	VecTextPrep   string
	VecModeler    string
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
	StoredSession ServerSession
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
	// atomic package is faster and mutext architecture is more robust/flexible in the long term
	// OR can do a counter via channels (Count int; Reply chan int; Fetch chan bool; Mod chan int)...
	// but that version is slightly slower than mutex: c. 5%
	count atomic.Int64
}

// Get - concurrency aware way to read a SrchCounter
func (c *SrchCounter) Get() int {
	return int(c.count.Load())
}

// Set - concurrency aware way to write to a SrchCounter
func (c *SrchCounter) Set(v int) {
	c.count.Store(int64(v))
}

//
// THREAD SAFE INFRASTRUCTURE: MUTEX
// (and not channel: https://github.com/golang/go/wiki/MutexOrChannel)
//

// MakeSearchVault - called only once; yields the AllSearches vault
func MakeSearchVault() SearchVault {
	return SearchVault{
		SearchMap: make(map[string]SearchStruct),
		mutex:     sync.RWMutex{},
	}
}

// SrchInfo - struct used to deliver info about searches in progress
type SrchInfo struct {
	ID        string
	Exists    bool
	Hits      int
	Remain    int
	SrchCount int
	VProgStrg string
	Summary   string
}

// SearchVault - there should be only one of these; and it contains all the searches
type SearchVault struct {
	SearchMap map[string]SearchStruct
	mutex     sync.RWMutex
}

func (sv *SearchVault) InsertSS(s SearchStruct) {
	sv.mutex.Lock()
	defer sv.mutex.Unlock()
	sv.SearchMap[s.ID] = s
}

// GetSS will fetch a SearchStruct; if it does not find one, it makes and registers a hollow search
func (sv *SearchVault) GetSS(id string) SearchStruct {
	sv.mutex.Lock()
	defer sv.mutex.Unlock()
	s, e := sv.SearchMap[id]
	if e != true {
		s = BuildHollowSearch()
		s.ID = id
		s.IsActive = false
		sv.SearchMap[id] = s
	}
	return s
}

// SimpleGetSS will fetch a SearchStruct; if it does not find one, it makes but does not register a hollow search
func (sv *SearchVault) SimpleGetSS(id string) SearchStruct {
	sv.mutex.Lock()
	defer sv.mutex.Unlock()
	s, e := sv.SearchMap[id]
	if e != true {
		s = BuildHollowSearch()
		s.ID = id
		s.IsActive = false
		// do not let WSMessageLoop() register searches in the SearchMap. This wreaks havoc with the MAXSEARCHTOTAL code
		// sv.SearchMap[id] = s
	}
	return s
}

func (sv *SearchVault) Delete(id string) {
	sv.mutex.Lock()
	defer sv.mutex.Unlock()
	delete(sv.SearchMap, id)
}

func (sv *SearchVault) GetInfo(id string) SrchInfo {
	sv.mutex.RLock()
	defer sv.mutex.RUnlock()
	var m SrchInfo
	_, m.Exists = sv.SearchMap[id]
	if m.Exists {
		m.Remain = sv.SearchMap[id].Remain.Get()
		m.Hits = sv.SearchMap[id].Hits.Get()
		m.VProgStrg = sv.SearchMap[id].ExtraMsg
		m.Summary = sv.SearchMap[id].InitSum
	}
	m.SrchCount = len(sv.SearchMap)
	return m
}

func (sv *SearchVault) SetRemain(id string, r int) {
	sv.mutex.Lock()
	defer sv.mutex.Unlock()
	sv.SearchMap[id].Remain.Set(r)
}

// CountIP - how many searches is the server already running?
func (sv *SearchVault) CountTotal() int {
	sv.mutex.Lock()
	defer sv.mutex.Unlock()
	return len(sv.SearchMap)
}

// CountIP - how many searches is this IP address already running?
func (sv *SearchVault) CountIP(ip string) int {
	sv.mutex.Lock()
	defer sv.mutex.Unlock()
	count := 0
	for _, v := range sv.SearchMap {
		if v.IPAddr == ip {
			count += 1
		}
	}
	return count
}

//
// FOR DEBUGGING ONLY
//

// svreport - report the # and names of the registered searches every 5 seconds
func svreport() {
	// add the following to main.go: "go svreport()"
	// it would be possible to "garbage collect" all searches where IsActive is "false" for too long
	// but these really are not supposed to be a problem
	for {
		as := AllSearches.SearchMap
		var ss []string
		for k := range as {
			ss = append(ss, k)
		}
		msg(fmt.Sprintf("%d in AllSearches: %s", len(as), strings.Join(ss, ", ")), MSGNOTE)
		time.Sleep(5 * time.Second)
	}
}
