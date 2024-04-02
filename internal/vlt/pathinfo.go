package vlt

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"runtime"
	"sort"
	"strings"
	"time"
)

// TerminalTicker - requires running with the "-tk" flag; feed basic use states to the console and update them indefinitely
func TerminalTicker(wait time.Duration) {
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
	if !Msg.Tick || Msg.Win {
		return
	}

	var mem runtime.MemStats

	// the uptime line
	t := func(up time.Duration) {
		runtime.ReadMemStats(&mem)
		heap := fmt.Sprintf("%dM", mem.HeapAlloc/1024/1024)
		// stack := fmt.Sprintf("%dM", mem.StackInuse/1024/1024)
		tick := fmt.Sprintf(UPTIME, time.Now().Format(time.TimeOnly), up.Truncate(time.Second), heap)
		tick = Msg.ColStyle(PADDING + tick + PADDING)
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
		out := Msg.Color(strings.Join(pairs, " C6*C0 "))
		fmt.Printf(out + CLEARRT)
		fmt.Println()
		fmt.Printf(CLEAR + CURSREST)
	}

	// this loop will never exit
	for {
		up := time.Since(Msg.Lnc)
		t(up)
		s()
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
