//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/mm"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/clr"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/debug"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/mps"
	"github.com/e-gun/HipparchiaGoServer/internal/search"
	"github.com/e-gun/HipparchiaGoServer/internal/vec"
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/e-gun/HipparchiaGoServer/web"
	"github.com/pkg/profile"
	_ "net/http/pprof"
	"os"
	"runtime"
	"sync"
	"time"
)

// these next variables should be injected at build time: 'go build -ldflags "-X main.GitCommit=$GIT_COMMIT"', etc

var GitCommit string
var VersSuppl string
var BuildDate string
var PGOInfo string

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

	// lnch.PrintVersion() needs to know this
	lnch.GitCommit = GitCommit
	lnch.VersSuppl = VersSuppl
	lnch.BuildDate = BuildDate
	lnch.PGOInfo = PGOInfo

	//
	// [0] debugging code block #1 of 2
	//

	// memory use debugging runs have to be custom-built

	// UNCOMMENT next and then: "curl http://localhost:8080/debug/pprof/heap > heap.0.pprof"
	// "go tool pprof heap.0.pprof" -> "top 20", etc.

	//go func() {
	//	lnch.Msg.CRIT("**THIS BUILD IS NOT FOR RELEASE** PPROF server is active")
	//	http.ListenAndServe("localhost:8080", nil)
	//}()

	vv.LaunchTime = time.Now()

	//
	// [1] set up the runtime configuration
	//

	vv.ServableFonts = web.CheckFontsAvailable(vv.ServableFonts)

	lnch.LookForConfigFile()
	lnch.ConfigAtLaunch()
	lnch.Msg.LLvl = lnch.Config.LogLevel
	msg := lnch.NewMessageMakerConfigured()
	msg.ResetScreen()

	clr.CssColorModes = clr.GenerateColorModes()
	clr.CssSamples = clr.GenerateCSSSamples()

	val := lnch.Config.CssColors
	vec.InjectedDotHue, vec.InjectedDotLum, vec.InjectedDotPeriphLum = clr.FindVectorDotHueAndLums(val, clr.CssColorHSLs[val])
	vec.InjectedLineHue, vec.InjectedLineLum = clr.FindVectorLineHueAndLums(val, clr.CssColorHSLs[val])
	vec.InjectedFontHue, vec.InjectedFontLum = clr.FindVectorFontHueAndLums(val, clr.CssColorHSLs[val])

	// [1a] maybe we are profiling
	// profiling runs are requested from the command line

	// e.g. running: ./HipparchiaGoServer -pc -st
	// vectorless: ./HipparchiaGoServer -pc -st -dv

	// profile into pdf:
	// 	"go tool pprof --pdf ./HipparchiaGoServer /var/folders/d8/_gb2lcbn0klg22g_cbwcxgmh0000gn/T/profile1075644045/cpu.pprof > ./fyi/CPUProfile.pdf"
	//	"cp /var/folders/d8/_gb2lcbn0klg22g_cbwcxgmh0000gn/T/profile1075644045/cpu.pprof ./pgo/default.pgo"

	if lnch.Config.ProfileCPU {
		defer profile.Start().Stop()
	}

	if lnch.Config.ProfileMEM {
		// mem profile:
		defer profile.Start(profile.MemProfile).Stop()
	}

	// [1b] need to update all the message makers out there now that Config is set
	mkr := []*mm.MessageMaker{db.Msg, debug.Msg, lnch.Msg, mps.Msg, search.Msg, str.Msg, vec.Msg, vlt.Msg, web.Msg}
	// need to keep these in the right order otherwise the coming loop will misassociate items...
	suff := []string{"-DBI", "-DBG", "-LNC", "-MPS", "-SEA", "-STR", "-VEC", "-VLT", "-WEB"}
	for i := range mkr {
		lnch.UpdateMessageMakerWithConfig(mkr[i])
		mkr[i].SNm = vv.SHORTNAME + suff[i]
	}

	// blank out the old message file
	if lnch.Config.LogToFile {
		uh, _ := os.UserHomeDir()
		f, err := os.Create(uh + "/" + vv.LOGFILEML)
		msg.EC(err)
		f.Close()
	}

	lnch.PrintVersion(*lnch.Config)
	lnch.PrintBuildInfo(*lnch.Config)

	if !lnch.Config.QuietStart {
		msg.MAND(fmt.Sprintf(vv.TERMINALTEXT, vv.PROJYEAR, vv.PROJAUTH, vv.PROJMAIL))
	}

	//
	// [2] set up things that will run forever in the background
	//

	db.SQLPool = db.FillDBConnectionPool(*lnch.Config)

	//
	// [3] now that we have access to the db, set some variables
	//

	vv.GrCt = vv.SetCorpusCount(vv.GREEKCORP, db.SQLPool)
	vv.LtCt = vv.SetCorpusCount(vv.LATINCORP, db.SQLPool)
	vv.InCt = vv.SetCorpusCount(vv.INSCRIPTCORP, db.SQLPool)
	vv.ChCt = vv.SetCorpusCount(vv.CHRISTINSC, db.SQLPool)
	vv.DpCt = vv.SetCorpusCount(vv.PAPYRUSCORP, db.SQLPool)
	vv.DbLmMapSize = vv.SetLemmCount(db.SQLPool)

	go vlt.WebsocketPool.WSPoolStartListening()

	go vlt.WSSearchInfoHub()
	go vlt.PathInfoHub()
	go vlt.TerminalTicker(vv.TICKERDELAY)

	//
	// [4] concurrent loading of the core data
	//

	var awaiting sync.WaitGroup
	awaiting.Add(1)
	go func(awaiting *sync.WaitGroup) {
		defer awaiting.Done()

		start := time.Now()
		previous := time.Now()

		mps.AllWorks = mps.ActiveWorkMapper()
		msg.Timer("A1", fmt.Sprintf(MSG1, len(mps.AllWorks)), start, previous)
		previous = time.Now()

		mps.AllAuthors = mps.ActiveAuthorMapper()
		msg.Timer("A2", fmt.Sprintf(MSG2, len(mps.AllAuthors)), start, previous)
		previous = time.Now()

		// full up WkCorpusMap, AuCorpusMap, ...
		mps.RePopulateGlobalMaps()
		msg.Timer("A3", MSG3, start, previous)
	}(&awaiting)

	awaiting.Add(1)
	go func(awaiting *sync.WaitGroup) {
		defer awaiting.Done()

		start := time.Now()
		previous := time.Now()

		mps.AllLemm = mps.LemmaMapper()
		msg.Timer("B1", fmt.Sprintf(MSG4, len(mps.AllLemm)), start, previous)

		previous = time.Now()
		mps.NestedLemm = mps.NestedLemmaMapper(mps.AllLemm)

		mps.LoadUnparsedWordCountWeights()
		mps.LoadParsedWordCountWeights()

		msg.Timer("B2", MSG5, start, previous)
	}(&awaiting)

	awaiting.Add(1)
	go func(awaiting *sync.WaitGroup) {
		defer awaiting.Done()
		if lnch.Config.ResetVectors {
			db.VectorDBReset()
		} else if lnch.Config.LogLevel >= mm.MSGNOTE {
			db.VectorDBCountNN(mm.MSGNOTE)
		}
	}(&awaiting)

	awaiting.Wait()

	runtime.GC()
	msg.Emit(msg.ColStyle(fmt.Sprintf(SUMM, time.Now().Sub(vv.LaunchTime).Seconds())), -1)

	//
	// [5] debugging code block #2 of 2
	// uncomment the following but very spammy in the console...
	//

	// go debug.WSClientReport(2 * time.Second)

	// [6] uncomment the next to generate "fontsubsetting/inuse.txt" (once in a very blue moon)
	// (after that you should run "fontsubsetting/pyftsubset.sh" and regenerate the embedded filesystem fonts

	// çmps.SubsetChars()

	msg.MAND(QUIT)

	if lnch.Config.Authenticate {
		vlt.BuildUserPassPairs(*lnch.Config)
	}

	//
	// [7] done: start the server (with a fnc call that will never return)
	//

	// fmt.Println(db.LoadBuildMetadata())
	web.StartEchoServer()
}
