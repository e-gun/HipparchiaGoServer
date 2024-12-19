//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vlt

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/e-gun/lastnlines"
	"math"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"
)

// TerminalTicker - requires running with the "-tk" flag; feed basic use states to the console and update them indefinitely
func TerminalTicker(wait time.Duration) {
	// sample output:

	//   -----------------  [11:24:34]  HGS uptime: 00d 01h 24m 17s  [87M]  -----------------
	//BrowseLine: 51 * LexFindByForm: 48 * LexLookup: 6 * LexReverse: 6 * NeighborsSearch: 1 * Search: 7

	const (
		CLEAR     = "\033[2K"
		CLEARLN   = "\033[2K"
		CLEARRT   = "\033[0K"
		CLEARSCR  = "\033[2J"
		HEAD      = "\r"
		CURSHOME  = "\033[1;1H"
		FIRSTLINE = "\033[2;1H"
		FROMTOP   = 4
		NLINE     = "\033[%d;1H"
		CURSSAVE  = "\033[s"
		CURSREST  = "\033[u"
		PADDING   = "  -----------------  "
		STATTMPL  = "%s: C2%dC0"
		UPTIME    = "[S1C6%vC0]  C5S1%s uptime: C1%vC0  [S1C6%sC0]"
		TIMESTR   = "%02dd %02dh %02dm %02ds"
	)

	// ANSI escape codes do not work in windows
	if !Msg.Tick || Msg.Win {
		return
	}

	var mem runtime.MemStats

	// --> "00d 01h 24m 17s"
	formattime := func(t time.Duration) string {
		// test := time.Date(2024, 7, 11, 9, 20, 0, 0, time.UTC)
		// t = time.Since(test)
		raws := float64(t.Seconds())
		s := int(math.Mod(raws, 60))
		rawm := float64(t.Minutes())
		m := int(math.Mod(rawm, 60))
		rawh := float64(t.Hours())
		h := int(math.Mod(rawh, 24))
		d := int(rawh / 24)
		return fmt.Sprintf(TIMESTR, d, h, m, s)
	}

	// the uptime line
	//   -----------------  [11:24:34]  HGS uptime: 00d 01h 24m 17s  [87M]  -----------------
	uptimer := func(up time.Duration) {
		runtime.ReadMemStats(&mem)
		heap := fmt.Sprintf("%dM", mem.HeapAlloc/1024/1024)
		// stack := fmt.Sprintf("%dM", mem.StackInuse/1024/1024)
		tick := fmt.Sprintf(UPTIME, time.Now().Format(time.TimeOnly), vv.SHORTNAME, formattime(up), heap)
		tick = Msg.ColStyle(PADDING + tick + PADDING)
		fmt.Printf(CURSSAVE + CURSHOME + CLEAR + HEAD)
		fmt.Printf(tick + CURSREST)
	}

	// the searches run line
	// BrowseLine: 51 * LexFindByForm: 48 * LexLookup: 6 * LexReverse: 6 * NeighborsSearch: 1 * Search: 7
	srchinfo := func() {
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
		out := Msg.Color(strings.Join(pairs, " C6*C0 "))
		fmt.Printf(out + CLEARRT)
		fmt.Println()
		fmt.Println()
		fmt.Printf(CLEAR + CURSREST)
	}

	// start collecting tails of the two logfiles
	uh, _ := os.UserHomeDir()

	lnl1 := lastnlines.NewLNL(uh + "/" + vv.LOGFILEEL)
	lnl1.SetDepth(lnch.Config.TickerLines)
	lnl1.Start()
	defer lnl1.Stop()

	lnl2 := lastnlines.NewLNL(uh + "/" + vv.LOGFILEML)
	lnl2.SetDepth(lnch.Config.TickerLines)
	lnl2.Start()
	defer lnl2.Stop()

	// output the tails
	tailinfo := func(info []string) {
		for i := 0; i < len(info); i++ {
			fmt.Printf(CLEARLN)
			fmt.Println(info[i])
		}

	}

	hdr := func(f string) {
		fmt.Printf(Msg.ColStyle("C6((C0 C1%sC0 C6))C0\n"), f)
	}

	// this loop will never exit
	fmt.Printf(CLEARSCR)
	for {
		up := time.Since(Msg.Lnc)
		uptimer(up)
		srchinfo()
		fmt.Printf(CURSSAVE + fmt.Sprintf(NLINE, FROMTOP))
		hdr(vv.LOGFILEEL)
		tailinfo(lnl1.Get())
		fmt.Printf(CURSSAVE + fmt.Sprintf(NLINE, FROMTOP+lnch.Config.TickerLines))
		fmt.Println()
		hdr(vv.LOGFILEML)
		tailinfo(lnl2.Get())
		time.Sleep(wait)
	}
}

// LogPaths - increment path counter for this path; optionally do runtime.GC as well
func LogPaths(fn string) {
	// sample output:
	// [a] "[HGS] RtLexReverse() runtime.GC() 426M --> 408M"
	// [b] "[HGS] RtLexLookup() current heap: 340M"

	const (
		MSG  = "%s runtime.GC() %s --> %s"
		HEAP = "%s current heap: %s"
	)

	// GENERAL STATS
	piupdate <- fn

	// GC INFO

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	b := fmt.Sprintf("%dM", mem.HeapAlloc/1024/1024)

	if !Msg.GC {
		Msg.PEEK(fmt.Sprintf(HEAP, fn, b))
	} else {
		runtime.GC()
		runtime.ReadMemStats(&mem)
		a := fmt.Sprintf("%dM", mem.HeapAlloc/1024/1024)
		Msg.PEEK(fmt.Sprintf(MSG, fn, b, a))
	}

	return
}

//
// CHANNEL-BASED PATHINFO REPORTING TO COMMUNICATE STATS BETWEEN ROUTINES
//

// PIReply - PathInfoHub helper struct for returning the PathInfo
type PIReply struct {
	Request  bool
	Response chan map[string]int
}

var (
	piupdate  = make(chan string, 2*runtime.NumCPU())
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
		case upd := <-piupdate:
			increm(upd)
		case req := <-PIRequest:
			req.Response <- PathsCalled
		}
	}
}
