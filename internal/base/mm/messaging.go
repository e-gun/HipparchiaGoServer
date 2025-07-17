//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package mm

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

//
// TERMINAL OUTPUT/MESSAGES
//

const (
	LOGFILE              = "hgs-msg.log" // circular import problem; need to edit "vv.miscinternalconstants.go" too if changing this
	UNCOLORED            = -9
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
		Lnc:    time.Now(),
		BW:     false,
		Clr:    "",
		GC:     false,
		LLvl:   0,
		LNm:    "",
		SNm:    "",
		Tick:   false,
		Ver:    "",
		Win:    w,
		ToFile: false,
		mtx:    sync.RWMutex{},
	}
}

type MessageMaker struct {
	Lnc    time.Time
	BW     bool
	Clr    string
	GC     bool
	LLvl   int
	LNm    string
	SNm    string
	Tick   bool
	Ver    string
	Win    bool
	ToFile bool
	mtx    sync.RWMutex
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

// AlwaysEmit - - send a message to the terminal or to the logfile; ignore threshold check
func (m *MessageMaker) AlwaysEmit(s string) {
	m.Emit(s, UNCOLORED)
}

// Emit - send a message to the terminal or to the logfile, perhaps adding color and style to it
func (m *MessageMaker) Emit(message string, threshold int) {
	// sample output: "[HGS] findbyform() found no results for 'Romani'"

	if m.LLvl < threshold {
		return
	}

	if m.ToFile && !m.BW {
		msg := m.ColorMessage(message, threshold)
		m.EmitToFile(msg)
		return
	} else if m.ToFile {
		msg := m.BWMessage(message, threshold)
		m.EmitToFile(msg)
	}

	if !m.Win && !m.BW {
		m.EmitColor(message, threshold)
	} else {
		m.EmitBW(message, threshold)
	}
}

// EmitBW - send a black and white message to the console
func (m *MessageMaker) EmitBW(message string, threshold int) {
	// terminal color codes not win's friend
	if threshold < 0 {
		fmt.Printf("[%s] %s\n", m.SNm, message)
	} else {
		fmt.Printf("[%s] [%d] %s\n", m.SNm, threshold, message)
	}
}

// EmitColor - send a color message to the console
func (m *MessageMaker) EmitColor(message string, threshold int) {
	fmt.Printf(m.ColorMessage(message, threshold))
}

func (m *MessageMaker) ColorMessage(message string, threshold int) string {
	var color string

	switch threshold {
	case UNCOLORED:
		// leave color as ""
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
	return fmt.Sprintf("[%s%s%s] %s%s%s\n", YELLOW1, m.SNm, RESET, color, message, RESET)
}

func (m *MessageMaker) BWMessage(message string, threshold int) string {
	tn := time.Now().Format(time.RFC3339)
	return fmt.Sprintf("[%s] [%s] [%d] %s\n", tn, m.SNm, threshold, message)
}

// EmitToFile - send the message to a file
func (m *MessageMaker) EmitToFile(message string) {
	uh, _ := os.UserHomeDir()
	f, err := os.OpenFile(uh+"/"+LOGFILE, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err = f.WriteString(message); err != nil {
		panic(err)
	}
	return
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

// EC - report error and branch that called function
func (m *MessageMaker) EC(err error) {
	if err != nil {
		fmt.Printf(PANIC2, YELLOW2, m.LNm, m.Ver, RESET, CYAN2, m.SNm, RESET, RED1, RESET)
		fmt.Println("\t" + err.Error())
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

// Timer - report how much time elapsed between A and B
func (m *MessageMaker) Timer(letter string, o string, start time.Time, previous time.Time) {
	// sample output: "[D2: 33.764s][Δ: 8.024s] look up 48 specific words"
	d := fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	o = fmt.Sprintf("[%s: %.3fs]", letter, time.Now().Sub(start).Seconds()) + d + o
	m.Emit(o, TIMETRACKERMSGTHRESH)
}
