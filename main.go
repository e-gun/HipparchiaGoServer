//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"github.com/pkg/profile"
	_ "net/http/pprof"
	"runtime"
	"sync"
	"time"
)

func main() {
	const (
		MSG1 = "%d works built: map[string]DbWork"
		MSG2 = "%d authors built: map[string]DbAuthor"
		MSG3 = "corpus maps built"
		MSG4 = "unnested lemma map built (%d items)"
		MSG5 = "nested lemma map built"
		SUMM = "C3initialization took %.3fsC0"
		QUIT = "to stop the server press Control-C or close this window"
	)

	//
	// [0] debugging code block #1 of 2
	//

	// memory use debugging runs have to be custom-built

	// UNCOMMENT next and then: "curl http://localhost:8080/debug/pprof/heap > heap.0.pprof"
	// "go tool pprof heap.0.pprof" -> "top 20", etc.

	//go func() {
	//	msg("**THIS BUILD IS NOT FOR RELEASE** PPROF server is active", MSGCRIT)
	//	http.ListenAndServe("localhost:8080", nil)
	//}()

	// testing sqlite...
	msg("**THIS BUILD IS NOT FOR RELEASE AND IS CERTAIN TO BREAK**", MSGCRIT)

	LaunchTime = time.Now()

	//
	// [1] set up the runtime configuration
	//

	LookForConfigFile()
	ConfigAtLaunch()

	// profiling runs are requested from the command line

	// e.g. running: ./HipparchiaGoServer -pc -st
	// vectorless: ./HipparchiaGoServer -pc -st -dv

	// profile into pdf:
	// 	"go tool pprof --pdf ./HipparchiaGoServer /var/folders/d8/_gb2lcbn0klg22g_cbwcxgmh0000gn/T/profile1075644045/cpu.pprof > ./fyi/CPUProfile.pdf"
	//	"cp /var/folders/d8/_gb2lcbn0klg22g_cbwcxgmh0000gn/T/profile1075644045/cpu.pprof ./pgo/default.pgo"

	if Config.ProfileCPU {
		defer profile.Start().Stop()
	}

	if Config.ProfileMEM {
		// mem profile:
		defer profile.Start(profile.MemProfile).Stop()
	}

	messenger.Cfg = Config
	messenger.Lnc.LaunchTime = LaunchTime
	messenger.ResetScreen()

	printversion()
	printbuildinfo()

	if !Config.QuietStart {
		msg(fmt.Sprintf(TERMINALTEXT, PROJYEAR, PROJAUTH, PROJMAIL), MSGMAND)
	}

	//
	// [2] set up things that will run forever in the background
	//

	SQLPool = FillPSQLPoool()
	go WebsocketPool.WSPoolStartListening()

	go SearchInfoHub()
	go PathInfoHub()

	go messenger.Ticker(TICKERDELAY)

	//
	// [3] concurrent loading of the core data
	//

	var awaiting sync.WaitGroup
	awaiting.Add(1)
	go func(awaiting *sync.WaitGroup) {
		defer awaiting.Done()

		start := time.Now()
		previous := time.Now()

		AllWorks = activeworkmapper()
		messenger.Timer("A1", fmt.Sprintf(MSG1, len(AllWorks)), start, previous)
		previous = time.Now()

		AllAuthors = activeauthormapper()
		messenger.Timer("A2", fmt.Sprintf(MSG2, len(AllAuthors)), start, previous)
		previous = time.Now()

		// full up WkCorpusMap, AuCorpusMap, ...
		populateglobalmaps()
		messenger.Timer("A3", MSG3, start, previous)
	}(&awaiting)

	awaiting.Add(1)
	go func(awaiting *sync.WaitGroup) {
		defer awaiting.Done()

		start := time.Now()
		previous := time.Now()

		AllLemm = lemmamapper()
		messenger.Timer("B1", fmt.Sprintf(MSG4, len(AllLemm)), start, previous)

		previous = time.Now()
		NestedLemm = nestedlemmamapper(AllLemm)
		messenger.Timer("B2", MSG5, start, previous)
	}(&awaiting)

	awaiting.Add(1)
	go func(awaiting *sync.WaitGroup) {
		defer awaiting.Done()
		if Config.ResetVectors {
			vectordbreset()
		} else if Config.LogLevel >= MSGNOTE {
			vectordbcountnn(MSGNOTE)
		}
	}(&awaiting)

	awaiting.Wait()

	// developer only fnc: extract pgsql DB to filesystem
	// DBtoCSV()

	// sqlite...
	start := time.Now()

	if SQLProvider == "sqlite" {
		sqliteloadactiveauthors()
		previous := time.Now()
		messenger.Timer("C", "sqliteloadactiveauthors()", start, previous)
	}

	messenger.LogPaths("main() post-initialization")
	msg(messenger.ColStyle(fmt.Sprintf(SUMM, time.Now().Sub(LaunchTime).Seconds())), -999)

	//
	// [4] debugging code block #2 of 2
	// uncomment one or more; they are very spammy in the console...
	//

	// go searchvaultreport(2 * time.Second)
	// go wsclientreport(2 * time.Second)

	msg(QUIT, MSGMAND)

	//
	// [5] done: start the server (which will never return)
	//

	StartEchoServer()
}

//
// VERSION INFO BUILD TIME INJECTION
//

// these next variables should be injected at build time: 'go build -ldflags "-X main.GitCommit=$GIT_COMMIT"', etc

var GitCommit string
var VersSuppl string
var BuildDate string
var PGOInfo string

func printversion() {
	// example:
	// [HGS] Hipparchia Golang Server (v1.2.16-pre) [git: 64974732] [default.pgo] [gl=3; el=0]
	const (
		SN = "[C1%sC0] "
		GC = " [C4git: C4%sC0]"
		LL = " [C6gl=%d; el=%dC0]"
		ME = "C5%sC0 (C2v%sC0)"
		PG = " [C3%sC0]"
	)
	sn := fmt.Sprintf(SN, SHORTNAME)
	gc := ""
	if GitCommit != "" {
		gc = fmt.Sprintf(GC, GitCommit)
	}

	pg := ""
	if PGOInfo != "" {
		pg = fmt.Sprintf(PG, PGOInfo)
	} else {
		pg = fmt.Sprintf(PG, "no pgo")
	}

	ll := fmt.Sprintf(LL, Config.LogLevel, Config.EchoLog)
	versioninfo := fmt.Sprintf(ME, MYNAME, VERSION+VersSuppl)
	versioninfo = sn + versioninfo + gc + pg + ll
	versioninfo = messenger.ColStyle(versioninfo)
	fmt.Println(versioninfo)
}

func printbuildinfo() {
	// example:
	// 	Built:	2023-11-14@19:02:51		Golang:	go1.21.4
	//	System:	darwin-arm64			WKvCPU:	20/20
	const (
		BD = "\tS1Built:S0\tC3%sC0\t"
		GV = "\tS1Golang:S0\tC3%sC0\n"
		SY = "\tS1System:S0\tC3%s-%sC0\t"
		WC = "\t\tS1WKvCPU:S0\tC3%dC0/C3%dC0"
	)

	bi := ""
	if BuildDate != "" {
		bi = messenger.ColStyle(fmt.Sprintf(BD, BuildDate))
	}
	bi += messenger.ColStyle(fmt.Sprintf(GV, runtime.Version()))
	bi += messenger.ColStyle(fmt.Sprintf(SY, runtime.GOOS, runtime.GOARCH))
	bi += messenger.ColStyle(fmt.Sprintf(WC, Config.WorkerCount, runtime.NumCPU()))
	fmt.Println(bi)
}
