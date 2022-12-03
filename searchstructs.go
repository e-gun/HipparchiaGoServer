package main

import (
	"fmt"
	"regexp"
	"strings"
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
	// Hits          *SrchCounter
	// Remain        *SrchCounter
	// lock   *sync.RWMutex
	Remain *SCClient
	Hits   *SCClient
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

	s.Seeking = uvσςϲ(s.Seeking)
	s.Proximate = uvσςϲ(s.Proximate)

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
		OrderedBy(nameIncreasing, titleIncreasing, increasingLines).Sort(s.Results)
	case crit == "converted_date":
		OrderedBy(dateIncreasing, nameIncreasing, titleIncreasing, increasingLines).Sort(s.Results)
	case crit == "universalid":
		OrderedBy(increasingID).Sort(s.Results)
	case crit == "provenance":
		// as this is likely an inscription search, why not sort next by date?
		OrderedBy(increasingWLOC, dateIncreasing).Sort(s.Results)
	default:
		// author nameIncreasing
		OrderedBy(nameIncreasing, increasingLines).Sort(s.Results)
	}
}

// AcqHitCounter - get a SrchCounter for storing Hits values
func (s *SearchStruct) AcqHitCounter() {
	//h := func() *SrchCounter { return &SrchCounter{} }()
	//s.Hits = h
	m := &SCMessage{
		ID:  s.ID + "-hits",
		Val: 0,
	}
	s.Hits = &SCClient{ID: s.ID + "-hits", Count: 0, SCM: m}
	SearchCountPool.Add <- s.Hits
}

// GetHitCount - concurrency aware way to read a SrchCounter
func (s *SearchStruct) GetHitCount() int {
	return s.Hits.Get()
}

// SetHitCount - concurrency aware way to write to a SrchCounter
func (s *SearchStruct) SetHitCount(c int) {
	s.Hits.Set(c)
}

// AcqRemainCounter - get a SrchCounter for storing Remain values
func (s *SearchStruct) AcqRemainCounter() {
	//r := func() *SrchCounter { return &SrchCounter{} }()
	// s.Remain = r
	m := &SCMessage{
		ID:  s.ID + "-remain",
		Val: 0,
	}
	s.Remain = &SCClient{ID: s.ID + "-remain", Count: 0, SCM: m}
	SearchCountPool.Add <- s.Remain
}

// GetRemainCount - concurrency aware way to read a SrchCounter
func (s *SearchStruct) GetRemainCount() int {
	return s.Remain.Get()
}

// SetRemainCount - concurrency aware way to write to a SrchCounter
func (s *SearchStruct) SetRemainCount(c int) {
	s.Remain.Set(c)
}

// Finished - clean up a finished search
func (s *SearchStruct) Finished() {
	s.IsActive = false
	SearchCountPool.Remove <- s.Remain
	SearchCountPool.Remove <- s.Hits
}

//
// SEARCHCOUNTERS
//

//type SrchCounter struct {
//	// atomic package could do this more simply, but this architecture is more flexible in the long term
//	count int
//	lock  sync.RWMutex
//}
//
//// Get - concurrency aware way to read a SrchCounter
//func (h *SrchCounter) Get() int {
//	h.lock.RLock()
//	defer h.lock.RUnlock()
//	return h.count
//}
//
//// Set - concurrency aware way to write to a SrchCounter
//func (h *SrchCounter) Set(c int) {
//	h.lock.Lock()
//	defer h.lock.Unlock()
//	h.count = c
//}

//
// POOLED SEARCHCOUNTERS
//

type SCPool struct {
	Add         chan *SCClient
	Remove      chan *SCClient
	SCClientMap map[*SCClient]bool
	Get         chan *SCMessage
	Set         chan *SCMessage
}

type SCClient struct {
	ID    string
	Count int
	SCM   *SCMessage
}

type SCMessage struct {
	ID  string
	Val int
}

func (s *SCClient) Get() int {
	SearchCountPool.Get <- s.SCM
	return s.SCM.Val
}

func (s *SCClient) Set(n int) int {
	s.SCM.Val = n
	SearchCountPool.Set <- s.SCM
	return s.Count
}

func SCFillNewPool() *SCPool {
	return &SCPool{
		Add:         make(chan *SCClient),
		Remove:      make(chan *SCClient),
		SCClientMap: make(map[*SCClient]bool),
		Get:         make(chan *SCMessage),
		Set:         make(chan *SCMessage),
	}
}

func (pool *SCPool) SCPoolStartListening() {
	for {
		select {
		case sc := <-pool.Add:
			pool.SCClientMap[sc] = true
			break
		case sc := <-pool.Remove:
			delete(pool.SCClientMap, sc)
			break
		case set := <-pool.Set:
			for cl := range pool.SCClientMap {
				if cl.ID == set.ID {
					cl.Count = set.Val
					break
				}
			}
			break
		case get := <-pool.Get:
			for cl := range pool.SCClientMap {
				if cl.ID == get.ID {
					get.Val = cl.Count
					break
				}
			}
		}
	}
}

//
// POOLED SEARCHSTRUCTURES
//

type SStructPool struct {
	Add           chan *SearchStruct
	Remove        chan *SearchStruct
	SSClientMap   map[*SearchStruct]bool
	Update        chan *SearchStruct
	RequestRemain chan *SearchStruct
	SendRemain    chan SSOut
	SendHits      chan SSOut
	RequestSS     chan string
	RequestHits   chan string
	RequestStats  chan string
	SendSS        chan *SearchStruct
	RequestExist  chan string
	Exists        chan SSOut
}

type SSOut struct {
	ID  string
	Val int
}

func SStructFillNewPool() *SStructPool {
	return &SStructPool{
		Add:           make(chan *SearchStruct),
		Remove:        make(chan *SearchStruct),
		SSClientMap:   make(map[*SearchStruct]bool),
		Update:        make(chan *SearchStruct),
		RequestRemain: make(chan *SearchStruct),
		SendRemain:    make(chan SSOut),
		SendHits:      make(chan SSOut),
		RequestSS:     make(chan string),
		RequestHits:   make(chan string),
		RequestStats:  make(chan string),
		SendSS:        make(chan *SearchStruct),
		RequestExist:  make(chan string),
		Exists:        make(chan SSOut),
	}
}

func (pool *SStructPool) SStructPoolStartListening() {
	for {
		select {
		case sc := <-pool.Add:
			pool.SSClientMap[sc] = true
			break
		case sc := <-pool.Remove:
			delete(pool.SSClientMap, sc)
			break
		case set := <-pool.Update:
			for cl := range pool.SSClientMap {
				if cl.ID == set.ID {
					cl = set
					break
				}
			}
			break
		case get := <-pool.RequestSS:
			for cl := range pool.SSClientMap {
				if cl.ID == get {
					pool.SendSS <- cl
					break
				}
			}
			break
		case stats := <-pool.RequestStats:
			for cl := range pool.SSClientMap {
				if cl.ID == stats {
					pool.SendRemain <- SSOut{ID: stats, Val: cl.Remain.Get()}
					pool.SendHits <- SSOut{ID: stats, Val: cl.Hits.Get()}
					break
				}
			}
			break
		case exists := <-pool.RequestExist:
			found := 0
			for cl := range pool.SSClientMap {
				if cl.ID == exists {
					found = 1
					break
				}
			}
			pool.Exists <- SSOut{ID: exists, Val: found}
			break
		}

	}
}
