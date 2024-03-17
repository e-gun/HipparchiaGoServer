//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/launch"
	"github.com/e-gun/HipparchiaGoServer/internal/m"
	"github.com/e-gun/HipparchiaGoServer/internal/pools"
	"github.com/e-gun/HipparchiaGoServer/internal/vaults"
	"github.com/e-gun/HipparchiaGoServer/internal/vect"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/e-gun/HipparchiaGoServer/web"
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
	//	m("**THIS BUILD IS NOT FOR RELEASE** PPROF server is active", MSGCRIT)
	//	http.ListenAndServe("localhost:8080", nil)
	//}()

	vv.LaunchTime = time.Now()

	//
	// [1] set up the runtime configuration
	//

	launch.LookForConfigFile()
	launch.ConfigAtLaunch()

	// profiling runs are requested from the command line

	// e.g. running: ./HipparchiaGoServer -pc -st
	// vectorless: ./HipparchiaGoServer -pc -st -dv

	// profile into pdf:
	// 	"go tool pprof --pdf ./HipparchiaGoServer /var/folders/d8/_gb2lcbn0klg22g_cbwcxgmh0000gn/T/profile1075644045/cpu.pprof > ./fyi/CPUProfile.pdf"
	//	"cp /var/folders/d8/_gb2lcbn0klg22g_cbwcxgmh0000gn/T/profile1075644045/cpu.pprof ./pgo/default.pgo"

	if launch.Config.ProfileCPU {
		defer profile.Start().Stop()
	}

	if launch.Config.ProfileMEM {
		// mem profile:
		defer profile.Start(profile.MemProfile).Stop()
	}
	messenger := m.NewMessageMaker(launch.Config, m.LaunchStruct{})
	messenger.Cfg = launch.Config
	messenger.Lnc.LaunchTime = vv.LaunchTime
	messenger.ResetScreen()

	launch.PrintVersion(*launch.Config)
	launch.PrintBuildInfo(*launch.Config)

	if !launch.Config.QuietStart {
		messenger.MAND(fmt.Sprintf(vv.TERMINALTEXT, vv.PROJYEAR, vv.PROJAUTH, vv.PROJMAIL))
	}

	//
	// [2] set up things that will run forever in the background
	//

	db.SQLPool = pools.FillDBConnectionPool(*launch.Config)
	go vaults.WebsocketPool.WSPoolStartListening()

	go vaults.WSSearchInfoHub()
	go m.PathInfoHub()

	go messenger.Ticker(vv.TICKERDELAY)

	//
	// [3] concurrent loading of the core data
	//

	var awaiting sync.WaitGroup
	awaiting.Add(1)
	go func(awaiting *sync.WaitGroup) {
		defer awaiting.Done()

		start := time.Now()
		previous := time.Now()

		vv.AllWorks = vv.ActiveWorkMapper()
		messenger.Timer("A1", fmt.Sprintf(MSG1, len(vv.AllWorks)), start, previous)
		previous = time.Now()

		vv.AllAuthors = vv.ActiveAuthorMapper()
		messenger.Timer("A2", fmt.Sprintf(MSG2, len(vv.AllAuthors)), start, previous)
		previous = time.Now()

		// full up WkCorpusMap, AuCorpusMap, ...
		vv.RePopulateGlobalMaps()
		messenger.Timer("A3", MSG3, start, previous)
	}(&awaiting)

	awaiting.Add(1)
	go func(awaiting *sync.WaitGroup) {
		defer awaiting.Done()

		start := time.Now()
		previous := time.Now()

		vv.AllLemm = vv.LemmaMapper()
		messenger.Timer("B1", fmt.Sprintf(MSG4, len(vv.AllLemm)), start, previous)

		previous = time.Now()
		vv.NestedLemm = vv.NestedLemmaMapper(vv.AllLemm)
		messenger.Timer("B2", MSG5, start, previous)
	}(&awaiting)

	awaiting.Add(1)
	go func(awaiting *sync.WaitGroup) {
		defer awaiting.Done()
		if launch.Config.ResetVectors {
			vect.VectorDBReset()
		} else if launch.Config.LogLevel >= m.MSGNOTE {
			vect.VectorDBCountNN(m.MSGNOTE)
		}
	}(&awaiting)

	awaiting.Wait()

	runtime.GC()
	messenger.LogPaths("main() post-initialization")
	messenger.Emit(messenger.ColStyle(fmt.Sprintf(SUMM, time.Now().Sub(vv.LaunchTime).Seconds())), -999)

	//
	// [4] debugging code block #2 of 2
	// uncomment the following but very spammy in the console...
	//

	// go wsclientreport(2 * time.Second)

	messenger.MAND(QUIT)

	if launch.Config.Authenticate {
		vaults.BuildUserPassPairs(*launch.Config)
	}

	//
	// [5] done: start the server (which will never return)
	//

	web.StartEchoServer()
}
