package main

import (
	"fmt"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
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
	SIUpdateSummMsg <- SIKVs{s.ID, sum}
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

//
// THREAD SAFE INFRASTRUCTURE: MUTEX FOR STATE
// (and not channel: https://github.com/golang/go/wiki/MutexOrChannel)
//

// MakeSearchVault - called only once; yields the AllSearches vault
func MakeSearchVault() SearchVault {
	return SearchVault{
		SearchMap: make(map[string]SearchStruct),
		mutex:     sync.RWMutex{},
	}
}

// SearchVault - there should be only one of these; and it contains all the searches
type SearchVault struct {
	SearchMap map[string]SearchStruct
	mutex     sync.RWMutex
}

// InsertSS - add a search to the vault
func (sv *SearchVault) InsertSS(s SearchStruct) {
	sv.mutex.Lock()
	defer sv.mutex.Unlock()
	sv.SearchMap[s.ID] = s
	SIUpdateHits <- SIKVi{s.ID, 0}
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
		SIUpdateHits <- SIKVi{s.ID, 0}
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
		SIUpdateHits <- SIKVi{s.ID, 0}
	}
	return s
}

// Delete - get rid of a search (probably for good, but see "Purge")
func (sv *SearchVault) Delete(id string) {
	// msg("SearchVault deleting "+id, 1)
	SIDel <- id
	sv.mutex.Lock()
	defer sv.mutex.Unlock()
	delete(sv.SearchMap, id)
}

// Purge is just delete; makes the code logic more legible; "Purge" implies that this search is likely to reappear with an "Update"
func (sv *SearchVault) Purge(id string) {
	// msg("SearchVault purging "+id, 3)
	SIDel <- id
	sv.Delete(id)
}

// CountTotal - how many searches is the server already running?
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
// CHANNEL-BASED SEARCHINFO REPORTING TO COMMUNICATE RESULTS BETWEEN ROUTINES
//

// SrchInfo - struct used to deliver info about searches in progress
type SrchInfo struct {
	ID        string
	Exists    bool
	Hits      int
	Remain    int
	TableCt   int
	SrchCount int
	VProgStrg string
	Summary   string
}

// SIKVi - SearchInfoHub helper struct for setting an int val on the item at map[key]
type SIKVi struct {
	key string
	val int
}

// SIKVs - SearchInfoHub helper struct for setting a string val on the item at map[key]
type SIKVs struct {
	key string
	val string
}

// SIReply - SearchInfoHub helper struct for returning the SrchInfo stored at map[key]
type SIReply struct {
	key      string
	response chan SrchInfo
}

var (
	SIUpdateHits     = make(chan SIKVi, 2*runtime.NumCPU())
	SIUpdateRemain   = make(chan SIKVi, 2*runtime.NumCPU())
	SIUpdateVProgMsg = make(chan SIKVs, 2*runtime.NumCPU())
	SIUpdateSummMsg  = make(chan SIKVs, 2*runtime.NumCPU())
	SIUpdateTW       = make(chan SIKVi)
	SIRequest        = make(chan SIReply)
	SIDel            = make(chan string)
)

// SearchInfoHub - the loop that lets you read/write from/to the searchinfo channels
func SearchInfoHub() {
	var (
		Allinfo  = make(map[string]SrchInfo)
		Finished = make(map[string]bool)
	)

	reporter := func(r SIReply) {
		if _, ok := Allinfo[r.key]; ok {
			r.response <- Allinfo[r.key]
		} else {
			// "false" triggers a break in rt-websocket.go
			r.response <- SrchInfo{Exists: false}
		}
	}

	fetchifexists := func(id string) SrchInfo {
		if _, ok := Allinfo[id]; ok {
			return Allinfo[id]
		} else {
			// any non-zero value for SrchCount is fine; the test in re-websocket.go is just for 0
			return SrchInfo{ID: id, Exists: true, SrchCount: 1}
		}
	}

	// this silly mechanism because selftest had 2nd round of nn vector tests respawning after deletion; rare, but...
	storeunlessfinished := func(si SrchInfo) {
		if _, ok := Finished[si.ID]; !ok {
			Allinfo[si.ID] = si
		}
	}

	// the main loop; it will never exit
	for {
		select {
		case rq := <-SIRequest:
			reporter(rq)
		case tw := <-SIUpdateTW:
			x := fetchifexists(tw.key)
			x.TableCt = tw.val
			storeunlessfinished(x)
		case wr := <-SIUpdateHits:
			x := fetchifexists(wr.key)
			x.Hits = wr.val
			storeunlessfinished(x)
		case wr := <-SIUpdateRemain:
			x := fetchifexists(wr.key)
			x.Remain = wr.val
			storeunlessfinished(x)
		case wr := <-SIUpdateVProgMsg:
			x := fetchifexists(wr.key)
			x.VProgStrg = wr.val
			storeunlessfinished(x)
		case wr := <-SIUpdateSummMsg:
			x := fetchifexists(wr.key)
			x.Summary = wr.val
			storeunlessfinished(x)
		case del := <-SIDel:
			Finished[del] = true
			delete(Allinfo, del)

		}
	}
}

//
// FOR DEBUGGING ONLY
//

// searchvaultreport - report the # and names of the registered searches every N seconds
func searchvaultreport() {
	// add the following to main.go: "go searchvaultreport()"
	// it would be possible to "garbage collect" all searches where IsActive is "false" for too long
	// but these really are not supposed to be a problem
	for {
		as := AllSearches.SearchMap
		var ss []string
		for k := range as {
			ss = append(ss, k)
		}
		msg(fmt.Sprintf("%d in AllSearches: %s", len(as), strings.Join(ss, ", ")), MSGNOTE)
		time.Sleep(4 * time.Second)
	}
}
