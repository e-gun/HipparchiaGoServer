//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"github.com/pkg/profile"
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
		SUMM = "initialization took %.3fs before reaching StartEchoServer()"
	)
	LaunchTime = time.Now()

	LookForConfigFile()
	ConfigAtLaunch()

	// profiling runs...

	// e.g. running: ./HipparchiaGoServer -pc -st
	// vectorless: ./HipparchiaGoServer -pc -st -dv

	// profile into pdf:
	// 	"go tool pprof --pdf ./HipparchiaGoServer /var/folders/d8/_gb2lcbn0klg22g_cbwcxgmh0000gn/T/profile1075644045/cpu.pprof > ./fyi/CPUProfile.pdf"
	//	"cp /var/folders/d8/_gb2lcbn0klg22g_cbwcxgmh0000gn/T/profile1075644045/cpu.pprof ./default.pgo"

	if Config.ProfileCPU {
		defer profile.Start().Stop()
	}

	if Config.ProfileMEM {
		// mem profile:
		defer profile.Start(profile.MemProfile).Stop()
	}

	messenger.Cfg = Config
	messenger.Lnc.LaunchTime = LaunchTime
	messenger.Ctr = StatCounter
	messenger.ResetScreen()

	printversion()

	if Config.LogLevel > 0 {
		printbuildinfo()
	}

	if !Config.QuietStart {
		msg(fmt.Sprintf(TERMINALTEXT, PROJYEAR, PROJAUTH, PROJMAIL), MSGMAND)
	}

	SQLPool = FillPSQLPoool()
	go WebsocketPool.WSPoolStartListening()
	go messenger.Ticker(TICKERDELAY)

	// concurrent launching
	var awaiting sync.WaitGroup
	awaiting.Add(1)
	go func(awaiting *sync.WaitGroup) {
		defer awaiting.Done()

		start := time.Now()
		previous := time.Now()

		AllWorks = workmapper()
		messenger.Timer("A1", fmt.Sprintf(MSG1, len(AllWorks)), start, previous)

		previous = time.Now()
		AllAuthors = authormapper(AllWorks)
		messenger.Timer("A2", fmt.Sprintf(MSG2, len(AllAuthors)), start, previous)

		previous = time.Now()
		WkCorpusMap = buildwkcorpusmap()
		AuCorpusMap = buildaucorpusmap()
		AuGenres = buildaugenresmap()
		WkGenres = buildwkgenresmap()
		AuLocs = buildaulocationmap()
		WkLocs = buildwklocationmap()
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

	awaiting.Wait()

	if Config.ResetVectors {
		vectordbreset()
	} else {
		vectordbcountnn(MSGNOTE)
	}

	messenger.GCStats("main() post-initialization")
	msg(fmt.Sprintf(SUMM, time.Now().Sub(LaunchTime).Seconds()), MSGWARN)

	// uncomment one or more of the next if debugging; they are very spammy for the console...

	// go svreport()
	// go wsclientreport()

	StartEchoServer()
}

//
// VERSION INFO BUILD TIME INJECTION
//

// these next variables should be injected at build time: 'go build -ldflags "-X main.GitCommit=$GIT_COMMIT"', etc

var GitCommit string
var VersSuppl string
var BuildDate string

func printversion() {
	sn := fmt.Sprintf("[C1%sC0] ", SHORTNAME)
	gc := ""
	if GitCommit != "" {
		gc = fmt.Sprintf(" [C4git: C4%sC0]", GitCommit)
	}
	ll := fmt.Sprintf(" [C6gl=%d; el=%dC0]", Config.LogLevel, Config.EchoLog)
	versioninfo := fmt.Sprintf("C5%sC0 (C2v%sC0)", MYNAME, VERSION+VersSuppl)
	versioninfo = sn + versioninfo + gc + ll
	versioninfo = messenger.ColStyle(versioninfo)
	fmt.Println(versioninfo)
}

func printbuildinfo() {
	bi := ""
	if BuildDate != "" {
		bi = messenger.ColStyle(fmt.Sprintf("\tS1Built:S0\tC3%sC0\t", BuildDate))
	}
	bi += messenger.ColStyle(fmt.Sprintf("\tS1Golang:S0\tC3%sC0\n", runtime.Version()))
	bi += messenger.ColStyle(fmt.Sprintf("\tS1System:S0\tC3%s-%sC0\t", runtime.GOOS, runtime.GOARCH))
	bi += messenger.ColStyle(fmt.Sprintf("\t\tS1WKvCPU:S0\tC3%dC0/C3%dC0", Config.WorkerCount, runtime.NumCPU()))
	fmt.Println(bi)
}
