//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package mm

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

//
// TERMINAL OUTPUT/MESSAGES
//

const (
	MSGMAND              = -1
	MSGCRIT              = 0
	MSGWARN              = 1
	MSGNOTE              = 2
	MSGFYI               = 3
	MSGPEEK              = 4
	MSGTMI               = 5
	TIMETRACKERMSGTHRESH = MSGFYI
	RESET                = "\033[0m"
	BLUE1                = "\033[38;5;38m"  // DeepSkyBlue2
	BLUE2                = "\033[38;5;68m"  // SteelBlue3
	CYAN1                = "\033[38;5;109m" // LightSkyBlue3
	CYAN2                = "\033[38;5;117m" // SkyBlue1
	GREEN                = "\033[38;5;70m"  // Chartreuse3
	RED1                 = "\033[38;5;160m" // Red3
	RED2                 = "\033[38;5;168m" // HotPink3
	YELLOW1              = "\033[38;5;178m" // Gold3
	YELLOW2              = "\033[38;5;143m" // DarkKhaki
	GREY1                = "\033[38;5;254m" // Grey89
	GREY2                = "\033[38;5;247m" // Grey62
	GREY3                = "\033[38;5;242m" // Grey42
	WHITE                = "\033[38;5;255m" // Grey93
	BLINK                = "\033[30;0;5m"
	PANIC                = "[%s%s v.%s%s] %sUNRECOVERABLE ERROR%s\n"
	PANIC2               = "[%s%s v.%s%s] (%s%s%s) %sUNRECOVERABLE ERROR%s\n"
)

// tedious because we need to avoid circular imports
// see also lnch.msgwithcfg.go which fleshes this out

func NewMessageMaker() *MessageMaker {
	w := false
	if runtime.GOOS == "windows" {
		w = true
	}
	return &MessageMaker{
		Lnc:  time.Now(),
		BW:   false,
		Clr:  "",
		GC:   false,
		LLvl: 0,
		LNm:  "",
		SNm:  "",
		Tick: false,
		Ver:  "",
		Win:  w,
		mtx:  sync.RWMutex{},
	}
}

type MessageMaker struct {
	Lnc  time.Time
	BW   bool
	Clr  string
	GC   bool
	LLvl int
	LNm  string
	SNm  string
	Tick bool
	Ver  string
	Win  bool
	mtx  sync.RWMutex
}

func (m *MessageMaker) MAND(s string) {
	m.Emit(s, MSGMAND)
}

func (m *MessageMaker) CRIT(s string) {
	m.Emit(s, MSGCRIT)
}

func (m *MessageMaker) WARN(s string) {
	m.Emit(s, MSGWARN)
}

func (m *MessageMaker) NOTE(s string) {
	m.Emit(s, MSGNOTE)
}

func (m *MessageMaker) FYI(s string) {
	m.Emit(s, MSGFYI)
}

func (m *MessageMaker) PEEK(s string) {
	m.Emit(s, MSGPEEK)
}

func (m *MessageMaker) TMI(s string) {
	m.Emit(s, MSGTMI)
}

// Emit - send a message to the terminal, perhaps adding color and style to it
func (m *MessageMaker) Emit(message string, threshold int) {
	// sample output: "[HGS] findbyform() found no results for 'Romani'"

	if m.LLvl < threshold {
		return
	}

	if !m.Win && !m.BW {
		var color string

		switch threshold {
		case MSGMAND:
			color = GREEN
		case MSGCRIT:
			color = RED1
		case MSGWARN:
			color = YELLOW2
		case MSGNOTE:
			color = YELLOW1
		case MSGFYI:
			color = CYAN2
		case MSGPEEK:
			color = BLUE2
		case MSGTMI:
			color = GREY3
		default:
			color = WHITE
		}
		fmt.Printf("[%s%s%s] %s%s%s\n", YELLOW1, m.SNm, RESET, color, message, RESET)
	} else {
		// terminal color codes not w's friend
		if threshold < 0 {
			fmt.Printf("[%s] %s\n", m.SNm, message)
		} else {
			fmt.Printf("[%s] [LL%d] %s\n", m.SNm, threshold, message)
		}
	}
}

// Color - color text with ANSI codes by swapping out pseudo-tags
func (m *MessageMaker) Color(tagged string) string {
	// "[git: C4%sC0]" ==> green text for the %s
	swap := strings.NewReplacer("C1", "", "C2", "", "C3", "", "C4", "", "C5", "", "C6", "", "C7", "", "C0", "")

	if !m.Win && !m.BW {
		swap = strings.NewReplacer("C1", YELLOW1, "C2", CYAN2, "C3", BLUE1, "C4", GREEN, "C5", RED1,
			"C6", GREY3, "C7", BLINK, "C0", RESET)
	}
	tagged = swap.Replace(tagged)
	return tagged
}

// Styled - style text with ANSI codes by swapping out pseudo-tags
func (m *MessageMaker) Styled(tagged string) string {
	const (
		BOLD    = "\033[1m"
		ITAL    = "\033[3m"
		UNDER   = "\033[4m"
		REVERSE = "\033[7m"
		STRIKE  = "\033[9m"
	)
	swap := strings.NewReplacer("S1", "", "S2", "", "S3", "", "S4", "", "S5", "", "S0", "")

	if !m.Win && !m.BW {
		swap = strings.NewReplacer("S1", BOLD, "S2", ITAL, "S3", UNDER, "S4", STRIKE, "S5", REVERSE,
			"S0", RESET)
	}
	tagged = swap.Replace(tagged)
	return tagged
}

func (m *MessageMaker) ColStyle(tagged string) string {
	return m.Styled(m.Color(tagged))
}

// Error - just panic...
func (m *MessageMaker) Error(err error) {
	if err != nil {
		fmt.Printf(PANIC, YELLOW2, m.LNm, m.Ver, RESET, RED1, RESET)
		fmt.Println(err)
		m.ExitOrHang(1)
	}
}

// EF - report error and function
func (m *MessageMaker) EF(err error, fn string) {
	if err != nil {
		fmt.Printf(PANIC2, YELLOW2, m.LNm, m.Ver, RESET, CYAN2, fn, RESET, RED1, RESET)
		fmt.Println(err)
		m.ExitOrHang(1)
	}
}

// EC - report error and page that called function
func (m *MessageMaker) EC(err error) {
	var c string
	if m.Clr != "" {
		c = m.Clr
	}
	if err != nil {
		fmt.Printf(PANIC2, YELLOW2, m.LNm, m.Ver, RESET, CYAN2, c, RESET, RED1, RESET)
		fmt.Println(err)
		m.ExitOrHang(1)
	}
}

// ExitOrHang - Windows should hang to keep the error visible before the window closes and hides it
func (m *MessageMaker) ExitOrHang(e int) {
	const (
		HANG = `Execution suspended. %s is now frozen. Note any errors above. Execution will halt after %d seconds.`
		SUSP = 60
	)
	if !m.Win {
		os.Exit(e)
	} else {
		m.Emit(fmt.Sprintf(HANG, m.LNm, SUSP), -1)
		time.Sleep(SUSP * time.Second)
		os.Exit(e)
	}
}

// ResetScreen - ANSI reset of console
func (m *MessageMaker) ResetScreen() {
	const (
		ERASESCRN = "\033[2J"
		CURSHOME  = "\033[1;1H"
		DOWNONE   = "\033[1B"
	)
	if !m.Tick || m.Win {
		return
	}
	fmt.Println(ERASESCRN + CURSHOME + DOWNONE + DOWNONE)
}

// LogPaths - increment path counter for this path; optionally do runtime.GC as well
func (m *MessageMaker) LogPaths(fn string) {
	// sample output:
	// [a] "[HGS] RtLexReverse() runtime.GC() 426M --> 408M"
	// [b] "[HGS] RtLexLookup() current heap: 340M"

	// users of this service:

	//rt-browser.go:	c.Response().After(func() { messenger.LogPaths("RtBrowseLine()") })
	//rt-lexica.go:	c.Response().After(func() { messenger.LogPaths("RtLexLookup()") })
	//rt-lexica.go:	c.Response().After(func() { messenger.LogPaths("RtLexFindByForm()") })
	//rt-lexica.go:	c.Response().After(func() { messenger.LogPaths("RtLexReverse()") })
	//rt-morphtables.go:	c.Response().After(func() { messenger.LogPaths("RtMorphchart()") })
	//rt-search.go:	c.Response().After(func() { messenger.LogPaths("RtSearch()") })
	//rt-textmaker.go:	c.Response().After(func() { messenger.LogPaths("RtTextMaker()") })
	//rt-textmaker.go:	c.Response().After(func() { messenger.LogPaths("RtVocabMaker()") })
	//rt-textmaker.go:	c.Response().After(func() { messenger.LogPaths("RtIndexMaker()") })
	//vectorquerylda.go:	c.Response().After(func() { messenger.LogPaths("LDASearch()") })
	//vectorqueryneighbors.go:	c.Response().After(func() { messenger.LogPaths("NeighborsSearch()") })

	const (
		MSG  = "%s runtime.GC() %s --> %s"
		HEAP = "%s current heap: %s"
		MPR  = MSGPEEK
	)

	// GENERAL STATS
	PIUpdate <- fn

	// GC INFO

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	b := fmt.Sprintf("%dM", mem.HeapAlloc/1024/1024)

	if !m.GC {
		m.Emit(fmt.Sprintf(HEAP, fn, b), MPR)
	} else {
		runtime.GC()
		runtime.ReadMemStats(&mem)
		a := fmt.Sprintf("%dM", mem.HeapAlloc/1024/1024)
		m.Emit(fmt.Sprintf(MSG, fn, b, a), MPR)
	}

	return
}

// Timer - report how much time elapsed between A and B
func (m *MessageMaker) Timer(letter string, o string, start time.Time, previous time.Time) {
	// sample output: "[D2: 33.764s][Δ: 8.024s] look up 48 specific words"
	d := fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	o = fmt.Sprintf("[%s: %.3fs]", letter, time.Now().Sub(start).Seconds()) + d + o
	m.Emit(o, TIMETRACKERMSGTHRESH)
}

// Ticker - requires running with the "-tk" flag; feed basic use states to the console and update them indefinitely
func (m *MessageMaker) Ticker(wait time.Duration) {
	// sample output:

	//  ----------------- [10:24:28] HGS uptime: 1m0s -----------------
	//BrowseLine: 51 * LexFindByForm: 48 * LexLookup: 6 * LexReverse: 6 * NeighborsSearch: 1 * Search: 7

	const (
		CLEAR     = "\033[2K"
		CLEARRT   = "\033[0K"
		HEAD      = "\r"
		CURSHOME  = "\033[1;1H"
		FIRSTLINE = "\033[2;1H"
		CURSSAVE  = "\033[s"
		CURSREST  = "\033[u"
		PADDING   = "  -----------------  "
		STATTMPL  = "%s: C2%dC0"
		UPTIME    = "[S1C6%vC0]  C5S1HGS uptime: C1%vC0  [S1C6%sC0]"
	)

	// ANSI escape codes do not work in windows
	if !m.Tick || m.Win {
		return
	}
	var mem runtime.MemStats

	// the uptime line
	t := func(up time.Duration) {
		runtime.ReadMemStats(&mem)
		heap := fmt.Sprintf("%dM", mem.HeapAlloc/1024/1024)
		// stack := fmt.Sprintf("%dM", mem.StackInuse/1024/1024)
		tick := fmt.Sprintf(UPTIME, time.Now().Format(time.TimeOnly), up.Truncate(time.Second), heap)
		tick = m.ColStyle(PADDING + tick + PADDING)
		fmt.Printf(CURSSAVE + CURSHOME + CLEAR + HEAD + tick + CURSREST)
	}

	// the searches run line
	s := func() {
		responder := PIReply{Request: true, Response: make(chan map[string]int)}
		PIRequest <- responder
		ctr := <-responder.Response

		exclude := []string{"main() post-initialization"}
		keys := gen.StringMapKeysIntoSlice(ctr)
		keys = gen.SetSubtraction(keys, exclude)

		var pairs []string
		for k := range keys {
			this := strings.TrimPrefix(keys[k], "Rt")
			this = strings.TrimSuffix(this, "()")
			pairs = append(pairs, fmt.Sprintf(STATTMPL, this, ctr[keys[k]]))
		}

		sort.Strings(pairs)

		fmt.Printf(CURSSAVE + FIRSTLINE)
		out := m.Color(strings.Join(pairs, " C6*C0 "))
		fmt.Printf(out + CLEARRT)
		fmt.Println()
		fmt.Printf(CLEAR + CURSREST)
	}

	for {
		up := time.Since(m.Lnc)
		t(up)
		s()
		time.Sleep(wait)
	}
}

//
// CHANNEL-BASED PATHINFO REPORTING TO COMMUNICATE STATS BETWEEN ROUTINES; this can't go to 'vlt': circular imports
//

// PIReply - PathInfoHub helper struct for returning the PathInfo
type PIReply struct {
	Request  bool
	Response chan map[string]int
}

var (
	PIUpdate  = make(chan string, 2*runtime.NumCPU())
	PIRequest = make(chan PIReply)
)

// PathInfoHub - log paths that pass through MessageMaker.LogPaths; note that we are assuming only one mm is logging
func PathInfoHub() {
	var (
		PathsCalled = make(map[string]int)
	)

	increm := func(p string) {
		if _, ok := PathsCalled[p]; ok {
			PathsCalled[p]++
		} else {
			PathsCalled[p] = 1
		}
	}

	// the main loop; it will never exit
	for {
		select {
		case upd := <-PIUpdate:
			increm(upd)
		case req := <-PIRequest:
			req.Response <- PathsCalled
		}
	}
}
