package main

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync/atomic"
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
)

var (
	messenger = NewMessageMaker(BuildDefaultConfig(), nil, LaunchStruct{
		Name:       MYNAME,
		Version:    VERSION,
		Shortname:  SHORTNAME,
		LaunchTime: time.Now(),
	})
)

func msg(s string, i int) {
	messenger.Emit(s, i)
}

func chke(e error) {
	messenger.Error(e)
}

func coloroutput(s string) string {
	return messenger.Color(s)
}

func styleoutput(s string) string {
	return messenger.Styled(s)
}

func NewMessageMaker(cc CurrentConfiguration, ct map[string]*atomic.Int32, ls LaunchStruct) *MessageMaker {
	return &MessageMaker{
		Cfg: cc,
		Ctr: ct,
		Lnc: ls,
	}
}

type MessageMaker struct {
	Cfg CurrentConfiguration
	Ctr map[string]*atomic.Int32
	Lnc LaunchStruct
}

type LaunchStruct struct {
	Name       string
	Version    string
	Shortname  string
	LaunchTime time.Time
}

func (m *MessageMaker) Emit(message string, threshold int) {
	if m.Cfg.LogLevel < threshold {
		return
	}

	if runtime.GOOS != "windows" && !m.Cfg.BlackAndWhite {
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

func (m *MessageMaker) Color(tagged string) string {
	// "[git: C4%sC0]" ==> green text for the %s
	swap := strings.NewReplacer("C1", "", "C2", "", "C3", "", "C4", "", "C5", "", "C6", "", "C7", "", "C0", "")

	if runtime.GOOS != "windows" && !m.Cfg.BlackAndWhite {
		swap = strings.NewReplacer("C1", YELLOW1, "C2", CYAN2, "C3", BLUE1, "C4", GREEN, "C5", RED1,
			"C6", GREY3, "C7", BLINK, "C0", RESET)
	}
	tagged = swap.Replace(tagged)
	return tagged
}

func (m *MessageMaker) Styled(tagged string) string {
	const (
		BOLD    = "\033[1m"
		ITAL    = "\033[3m"
		UNDER   = "\033[4m"
		REVERSE = "\033[7m"
		STRIKE  = "\033[9m"
	)
	swap := strings.NewReplacer("S1", "", "S2", "", "S3", "", "S4", "", "S5", "", "S0", "")

	if runtime.GOOS != "windows" && !m.Cfg.BlackAndWhite {
		swap = strings.NewReplacer("S1", BOLD, "S2", ITAL, "S3", UNDER, "S4", STRIKE, "S5", REVERSE,
			"S0", RESET)
	}
	tagged = swap.Replace(tagged)
	return tagged
}

func (m *MessageMaker) ColStyle(tagged string) string {
	return m.Styled(m.Color(tagged))
}

func (m *MessageMaker) Error(err error) {
	if err != nil {
		fmt.Printf(PANIC, YELLOW2, m.Lnc.Name, m.Lnc.Version, RESET, RED1, RESET)
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
	if runtime.GOOS != "windows" {
		os.Exit(e)
	} else {
		m.Emit(fmt.Sprintf(HANG, m.Lnc.Name, SUSP), -1)
		time.Sleep(SUSP * time.Second)
		os.Exit(e)
	}
}

func (m *MessageMaker) ResetScreen() {
	const (
		ERASESCRN = "\033[2J"
		CURSHOME  = "\033[1;1H"
		DOWNONE   = "\033[1B"
	)
	if !m.Cfg.TickerActive || runtime.GOOS == "windows" {
		return
	}
	fmt.Println(ERASESCRN + CURSHOME + DOWNONE + DOWNONE)
}

func (m *MessageMaker) Stats(fn string) {
	//rt-lexica.go:   c.Response().After(func() { messenger.Stats("RtLexLookup()") })
	//rt-lexica.go:   c.Response().After(func() { messenger.Stats("RtLexReverse()") })
	//rt-search.go:   c.Response().After(func() { messenger.Stats("RtSearch()") })
	//rt-textsindicesandvocab.go:     c.Response().After(func() { messenger.Stats("RtTextMaker()") })
	//rt-textsindicesandvocab.go:     c.Response().After(func() { messenger.Stats("RtVocabMaker()") })
	//rt-textsindicesandvocab.go:     c.Response().After(func() { messenger.Stats("RtIndexMaker()") })

	const (
		MSG = "%s runtime.GC() %s --> %s"
		MPR = MSGPEEK
	)

	if !m.Cfg.ManualGC {
		return
	}
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	b := fmt.Sprintf("%dM", mem.HeapAlloc/1024/1024)
	runtime.GC()
	runtime.ReadMemStats(&mem)
	a := fmt.Sprintf("%dM", mem.HeapAlloc/1024/1024)
	m.Emit(fmt.Sprintf(MSG, fn, b, a), MPR)

	_, ok := m.Ctr[fn]
	if !ok {
		m.Ctr[fn] = &atomic.Int32{}
	}
	_ = m.Ctr[fn].Add(1)
}

func (m *MessageMaker) Timer(letter string, o string, start time.Time, previous time.Time) {
	d := fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	o = fmt.Sprintf("[%s: %.3fs]", letter, time.Now().Sub(start).Seconds()) + d + o
	m.Emit(o, TIMETRACKERMSGTHRESH)
}

func (m *MessageMaker) Ticker(wait time.Duration) {
	const (
		CLEAR     = "\033[2K"
		CLEARRT   = "\033[0K"
		HEAD      = "\r"
		CURSHOME  = "\033[1;1H"
		FIRSTLINE = "\033[2;1H"
		CURSSAVE  = "\033[s"
		CURSREST  = "\033[u"
		PADDING   = " ----------------- "
		STATTMPL  = "%s: C2%dC0"
	)
	// ANSI escape codes do not work in windows
	if !m.Cfg.TickerActive || runtime.GOOS == "windows" {
		return
	}

	t := func(up time.Duration) {
		tick := fmt.Sprintf("[S1C6%vC0] C5S1HGS uptime: C1%vC0", time.Now().Format(time.TimeOnly), up.Truncate(time.Minute))
		tick = m.ColStyle(PADDING + tick + PADDING)
		fmt.Printf(CURSSAVE + CURSHOME + CLEAR + HEAD + tick + CURSREST)
	}

	s := func() {
		exclude := []string{"main() post-initialization"}
		keys := StringMapKeysIntoSlice(m.Ctr)
		keys = SetSubtraction(keys, exclude)
		sort.Strings(keys)

		var pairs []string
		for k := range keys {
			this := strings.TrimPrefix(keys[k], "Rt")
			this = strings.TrimSuffix(this, "()")
			pairs = append(pairs, fmt.Sprintf(STATTMPL, this, m.Ctr[keys[k]].Load()))
		}
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
