package main

import (
	"fmt"
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

var (
	messenger = NewGenericMessageMaker(BuildDefaultConfig(), LaunchStruct{
		Name:       MYNAME,
		Version:    VERSION,
		Shortname:  SHORTNAME,
		LaunchTime: time.Now(),
	})
)

// msg - send a message to the terminal; alias for "messenger.Emit(s, i)"
func msg(s string, i int) {
	messenger.Emit(s, i)
}

// chke - check an error; alias for messenger.Error(e)
func chke(e error) {
	messenger.Error(e)
}

func chkf(e error, s string) {
	messenger.EF(e, s)
}

func coloroutput(s string) string {
	return messenger.Color(s)
}

func styleoutput(s string) string {
	return messenger.Styled(s)
}

func NewGenericMessageMaker(cc CurrentConfiguration, ls LaunchStruct) *MessageMaker {
	w := false
	if runtime.GOOS == "windows" {
		w = true
	}
	return &MessageMaker{
		Cfg: cc,
		Lnc: ls,
		Win: w,
	}
}

func NewFncMessageMaker(c string) *MessageMaker {
	cc := messenger.Cfg
	ls := messenger.Lnc
	ls.Caller = c
	return &MessageMaker{
		Cfg: cc,
		Lnc: ls,
	}
}

type MessageMaker struct {
	Cfg CurrentConfiguration
	Lnc LaunchStruct
	Pgn string
	Win bool
	mtx sync.RWMutex
}

type LaunchStruct struct {
	Name       string
	Version    string
	Shortname  string
	LaunchTime time.Time
	Caller     string
}

// Emit - send a message to the terminal, perhaps adding color and style to it
func (m *MessageMaker) Emit(message string, threshold int) {
	// sample output: "[HGS] findbyform() found no results for 'Romani'"

	if m.Cfg.LogLevel < threshold {
		return
	}

	if !m.Win && !m.Cfg.BlackAndWhite {
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
		fmt.Printf("[%s%s%s] %s%s%s\n", YELLOW1, m.Lnc.Shortname, RESET, color, message, RESET)
	} else {
		// terminal color codes not w's friend
		fmt.Printf("[%s] %s\n", m.Lnc.Shortname, message)
	}
}

// Color - color text with ANSI codes by swapping out pseudo-tags
func (m *MessageMaker) Color(tagged string) string {
	// "[git: C4%sC0]" ==> green text for the %s
	swap := strings.NewReplacer("C1", "", "C2", "", "C3", "", "C4", "", "C5", "", "C6", "", "C7", "", "C0", "")

	if !m.Win && !m.Cfg.BlackAndWhite {
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

	if !m.Win && !m.Cfg.BlackAndWhite {
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
		fmt.Printf(PANIC, YELLOW2, m.Lnc.Name, m.Lnc.Version, RESET, RED1, RESET)
		fmt.Println(err)
		m.ExitOrHang(1)
	}
}

// EF - report error and function
func (m *MessageMaker) EF(err error, fn string) {
	if err != nil {
		fmt.Printf(PANIC2, YELLOW2, m.Lnc.Name, m.Lnc.Version, RESET, CYAN2, fn, RESET, RED1, RESET)
		fmt.Println(err)
		m.ExitOrHang(1)
	}
}

// EC - report error and page that called function
func (m *MessageMaker) EC(err error) {
	var c string
	if m.Lnc.Caller != "" {
		c = m.Lnc.Caller
	}
	if err != nil {
		fmt.Printf(PANIC2, YELLOW2, m.Lnc.Name, m.Lnc.Version, RESET, CYAN2, c, RESET, RED1, RESET)
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
		m.Emit(fmt.Sprintf(HANG, m.Lnc.Name, SUSP), -1)
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
	if !m.Cfg.TickerActive || m.Win {
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
	//rt-textsindicesandvocab.go:	c.Response().After(func() { messenger.LogPaths("RtTextMaker()") })
	//rt-textsindicesandvocab.go:	c.Response().After(func() { messenger.LogPaths("RtVocabMaker()") })
	//rt-textsindicesandvocab.go:	c.Response().After(func() { messenger.LogPaths("RtIndexMaker()") })
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

	if !m.Cfg.ManualGC {
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
	if !m.Cfg.TickerActive || m.Win {
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
		responder := PIReply{req: true, response: make(chan map[string]int)}
		PIRequest <- responder
		ctr := <-responder.response

		exclude := []string{"main() post-initialization"}
		keys := StringMapKeysIntoSlice(ctr)
		keys = SetSubtraction(keys, exclude)

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
		up := time.Since(m.Lnc.LaunchTime)
		t(up)
		s()
		time.Sleep(wait)
	}
}

//
// CHANNEL-BASED PATHINFO REPORTING TO COMMUNICATE STATS BETWEEN ROUTINES
//

// PIReply - PathInfoHub helper struct for returning the PathInfo
type PIReply struct {
	req      bool
	response chan map[string]int
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
			req.response <- PathsCalled
		}
	}
}
