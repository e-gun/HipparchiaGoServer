//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	// cpu profile via import of "github.com/pkg/profile":
	// defer profile.Start().Stop()

	// mem profile:
	// defer profile.Start(profile.MemProfile).Stop()

	// go tool pprof --pdf ./HipparchiaGoServer /var/folders/d8/_gb2lcbn0klg22g_cbwcxgmh0000gn/T/profile1880749830/cpu.pprof > profile.pdf

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
	ResetScreen()

	printversion()

	if !Config.QuietStart {
		msg(fmt.Sprintf(TERMINALTEXT, PROJYEAR, PROJAUTH, PROJMAIL), MSGMAND)
	}

	SQLPool = FillPSQLPoool()
	go WebsocketPool.WSPoolStartListening()
	go UptimeTicker(TICKERDELAY)

	// concurrent launching
	var awaiting sync.WaitGroup
	awaiting.Add(1)
	go func(awaiting *sync.WaitGroup) {
		defer awaiting.Done()

		start := time.Now()
		previous := time.Now()

		AllWorks = workmapper()
		TimeTracker("A1", fmt.Sprintf(MSG1, len(AllWorks)), start, previous)

		previous = time.Now()
		AllAuthors = authormapper(AllWorks)
		TimeTracker("A2", fmt.Sprintf(MSG2, len(AllAuthors)), start, previous)

		previous = time.Now()
		WkCorpusMap = buildwkcorpusmap()
		AuCorpusMap = buildaucorpusmap()
		AuGenres = buildaugenresmap()
		WkGenres = buildwkgenresmap()
		AuLocs = buildaulocationmap()
		WkLocs = buildwklocationmap()
		TimeTracker("A3", MSG3, start, previous)
	}(&awaiting)

	awaiting.Add(1)
	go func(awaiting *sync.WaitGroup) {
		defer awaiting.Done()

		start := time.Now()
		previous := time.Now()

		AllLemm = lemmamapper()
		TimeTracker("B1", fmt.Sprintf(MSG4, len(AllLemm)), start, previous)

		previous = time.Now()
		NestedLemm = nestedlemmamapper(AllLemm)
		TimeTracker("B2", MSG5, start, previous)
	}(&awaiting)

	awaiting.Wait()

	vectordbreset()

	SelfStats("main() post-initialization")
	msg(fmt.Sprintf(SUMM, time.Now().Sub(LaunchTime).Seconds()), MSGWARN)
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
	versioninfo = styleoutput(coloroutput(versioninfo))
	fmt.Println(versioninfo)
}

func printbuildinfo() {
	bi := ""
	if BuildDate != "" {
		bi = styleoutput(coloroutput(fmt.Sprintf("\tS1Built:S0\tC3%sC0\n", BuildDate)))

	}
	bi += styleoutput(coloroutput(fmt.Sprintf("\tS1Go:S0\tC3%sC0", runtime.Version())))
	fmt.Println(bi)
}
