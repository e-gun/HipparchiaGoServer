//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"C"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

func main() {

	// cpu profile:
	// defer profile.Start().Stop()

	// mem profile:
	// defer profile.Start(profile.MemProfile).Stop()

	// go tool pprof --pdf ./HipparchiaGoServer /var/folders/d8/_gb2lcbn0klg22g_cbwcxgmh0000gn/T/profile1880749830/cpu.pprof > profile.pdf

	configatlaunch()

	printversion()

	if !cfg.QuietStart {
		msg(fmt.Sprintf(TERMINALTEXT, PROJYEAR, PROJAUTH, PROJMAIL), -1)
	}

	dbp = FillPSQLPoool()

	// concurrent launching
	var awaiting sync.WaitGroup

	awaiting.Add(1)
	go func(awaiting *sync.WaitGroup) {
		defer awaiting.Done()

		start := time.Now()
		previous := time.Now()

		AllWorks = workmapper()
		timetracker("A1", fmt.Sprintf("%d works built: map[string]DbWork", len(AllWorks)), start, previous)

		previous = time.Now()
		AllAuthors = authormapper(AllWorks)
		timetracker("A2", fmt.Sprintf("%d authors built: map[string]DbAuthor", len(AllAuthors)), start, previous)

		previous = time.Now()
		WkCorpusMap = buildwkcorpusmap()
		AuCorpusMap = buildaucorpusmap()
		AuGenres = buildaugenresmap()
		WkGenres = buildwkgenresmap()
		AuLocs = buildaulocationmap()
		WkLocs = buildwklocationmap()
		timetracker("A3", "corpus maps built", start, previous)
	}(&awaiting)

	awaiting.Add(1)
	go func(awaiting *sync.WaitGroup) {
		defer awaiting.Done()

		start := time.Now()
		previous := time.Now()

		AllLemm = lemmamapper()
		timetracker("B1", fmt.Sprintf("unnested lemma map built (%d items)", len(AllLemm)), start, previous)

		previous = time.Now()
		NestedLemm = nestedlemmamapper(AllLemm)
		timetracker("B2", "nested lemma map built", start, previous)
	}(&awaiting)

	awaiting.Wait()

	gcstats("main() post-initialization")

	StartEchoServer()
}

type CurrentConfiguration struct {
	BadChars    string
	BrowserCtx  int64
	DbDebug     bool
	DefCorp     map[string]bool
	EchoLog     int // "none", "terse", "verbose"
	Font        string
	Gzip        bool
	HostIP      string
	HostPort    int
	LogLevel    int
	ManualGC    bool // see gcstats()
	MaxText     int
	PGLogin     PostgresLogin
	QuietStart  bool
	WorkerCount int
}

// configatlaunch - read the configuration values from JSON and/or command line
func configatlaunch() {
	cfg.BadChars = UNACCEPTABLEINPUT
	cfg.BrowserCtx = DEFAULTBROWSERCTX
	cfg.DbDebug = false
	cfg.Font = FONTSETTING
	cfg.Gzip = USEGZIP
	cfg.HostIP = SERVEDFROMHOST
	cfg.HostPort = SERVEDFROMPORT
	cfg.ManualGC = true
	cfg.MaxText = MAXTEXTLINEGENERATION
	cfg.QuietStart = false
	cfg.WorkerCount = runtime.NumCPU()

	e := json.Unmarshal([]byte(DEFAULTCORPORA), &cfg.DefCorp)
	chke(e)

	cf := fmt.Sprintf("%s/%s", CONFIGLOCATION, CONFIGNAME)

	uh, _ := os.UserHomeDir()
	h := fmt.Sprintf(CONFIGALTAPTH, uh)
	acf := fmt.Sprintf("%s/%s", h, CONFIGNAME)
	pcf := fmt.Sprintf("%s/%s", h, PROLIXCONFIGFILE)

	cfc, _ := os.Open(pcf)
	decoderc := json.NewDecoder(cfc)
	confc := CurrentConfiguration{}
	errc := decoderc.Decode(&confc)

	if errc == nil {
		cfg = confc
	}

	var pl PostgresLogin

	args := os.Args[1:len(os.Args)]

	for i, a := range args {
		switch a {
		case "-v":
			printversion()
			os.Exit(1)
		case "-ac":
			err := json.Unmarshal([]byte(args[i+1]), &cfg.DefCorp)
			if err != nil {
				msg(fmt.Sprintf("Improperly formatted corpus list. Using:\n\t%s", DEFAULTCORPORA), 0)
			}
		case "-bc":
			bc, err := strconv.Atoi(args[i+1])
			chke(err)
			cfg.BrowserCtx = int64(bc)
		case "-cf":
			cf = args[i+1]
		case "-db":
			cfg.DbDebug = true
		case "-ft":
			cfg.Font = args[i+1]
		case "-el":
			ll, err := strconv.Atoi(args[i+1])
			chke(err)
			cfg.EchoLog = ll
		case "-gl":
			ll, err := strconv.Atoi(args[i+1])
			chke(err)
			cfg.LogLevel = ll
		case "-gz":
			cfg.Gzip = true
		case "-h":
			printversion()
			fmt.Println(fmt.Sprintf(HELPTEXT, DEFAULTBROWSERCTX, CONFIGLOCATION, CONFIGNAME, h, CONFIGNAME,
				DEFAULTECHOLOGLEVEL, DEFAULTGOLOGLEVEL, SERVEDFROMHOST, SERVEDFROMPORT, MAXTEXTLINEGENERATION,
				UNACCEPTABLEINPUT, PROLIXCONFIGFILE, h, PROJURL))
			os.Exit(1)
		case "-pg":
			js := args[i+1]
			err := json.Unmarshal([]byte(js), &pl)
			if err != nil {
				msg("Could not parse your information as a valid collection of credentials. Use the following template:", -1)
				msg(`"{\"Pass\": \"YOURPASSWORDHERE\" ,\"Host\": \"127.0.0.1\", \"Port\": 5432, \"DBName\": \"hipparchiaDB\" ,\"User\": \"hippa_wr\"}"`, 0)
			}
		case "-q":
			cfg.QuietStart = true
		case "-sa":
			cfg.HostIP = args[i+1]
		case "-sp":
			p, err := strconv.Atoi(args[i+1])
			chke(err)
			cfg.HostPort = p
		case "-ti":
			tt, err := strconv.Atoi(args[i+1])
			chke(err)
			cfg.MaxText = tt
		case "-ui":
			cfg.BadChars = args[i+1]
		case "-wc":
			wc, err := strconv.Atoi(args[i+1])
			chke(err)
			cfg.WorkerCount = wc
		default:
			// do nothing
		}
	}

	y := ""
	if errc != nil {
		y = " *not*"
	}
	msg(fmt.Sprintf("'%s%s'%s loaded", h, PROLIXCONFIGFILE, y), 5)

	type ConfigFile struct {
		PosgreSQLPassword string
	}

	cfg.PGLogin = PostgresLogin{}

	if pl.Pass != "" {
		cfg.PGLogin = pl
	} else {
		cfa, _ := os.Open(cf)
		cfb, _ := os.Open(acf)

		decodera := json.NewDecoder(cfa)
		confa := ConfigFile{}
		erra := decodera.Decode(&confa)

		decoderb := json.NewDecoder(cfb)
		confb := ConfigFile{}
		errb := decoderb.Decode(&confb)

		if erra != nil && errb != nil {
			msg(fmt.Sprintf("FAILED to load the configuration from either '%s' or '%s'", cf, acf), 0)
			msg(fmt.Sprintf("Make sure that the file exists and that it has the following format:"), 0)
			fmt.Println(MINCONFIG)
			os.Exit(0)
		}
		conf := ConfigFile{}
		if erra == nil {
			conf = confa
		} else {
			conf = confb
		}

		cfg.PGLogin = PostgresLogin{
			Host:   DEFAULTPSQLHOST,
			Port:   DEFAULTPSQLPORT,
			User:   DEFAULTPSQLUSER,
			DBName: DEFAULTPSQLDB,
			Pass:   conf.PosgreSQLPassword,
		}
	}
}

func printversion() {
	ll := fmt.Sprintf(" [gl=%d; el=%d]", cfg.LogLevel, cfg.EchoLog)
	versioninfo := fmt.Sprintf("%s (v%s)", MYNAME, VERSION)
	versioninfo = versioninfo + ll
	msg(versioninfo, 0)
}
