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

	msg(fmt.Sprintf(TERMINALTEXT, PROJ, PROJYEAR, PROJAUTH, PROJMAIL), -1)

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
		AllAuthors = loadworksintoauthors(authormapper(), AllWorks)
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

	StartEchoServer()
}

type CurrentConfiguration struct {
	WorkerCount int
	LogLevel    int
	EchoLog     int // "none", "terse", "verbose"
	PGLogin     PostgresLogin
	HostIP      string
	HostPort    int
	Font        string
	Gzip        bool
	MaxText     int
	BadChars    string
}

// configatlaunch - read the configuration values from JSON and/or command line
func configatlaunch() {
	cfg.HostIP = SERVEDFROMHOST
	cfg.HostPort = SERVEDFROMPORT
	cfg.Font = FONTSETTING
	cfg.Gzip = USEGZIP
	cfg.MaxText = MAXTEXTLINEGENERATION
	cfg.BadChars = UNACCEPTABLEINPUT

	cf := fmt.Sprintf("%s/%s", CONFIGLOCATION, CONFIGNAME)

	var pl PostgresLogin

	args := os.Args[1:len(os.Args)]

	for i, a := range args {
		switch a {
		case "-v":
			printversion()
			os.Exit(1)
		case "-gl":
			ll, e := strconv.Atoi(args[i+1])
			chke(e)
			cfg.LogLevel = ll
		case "-gz":
			cfg.Gzip = true
		case "-el":
			ll, e := strconv.Atoi(args[i+1])
			chke(e)
			cfg.EchoLog = ll
		case "-cf":
			cf = args[i+1]
		case "-ft":
			cfg.Font = args[i+1]
		case "-h":
			printversion()
			fmt.Println(fmt.Sprintf(HELPTEXT, CONFIGLOCATION, CONFIGNAME, DEFAULTECHOLOGLEVEL, DEFAULTGOLOGLEVEL,
				SERVEDFROMHOST, SERVEDFROMPORT, MAXTEXTLINEGENERATION, UNACCEPTABLEINPUT))
			os.Exit(1)
		case "-p":
			js := args[i+1]
			err := json.Unmarshal([]byte(js), &pl)
			if err != nil {
				msg("Could not parse your information as a valid collection of credentials. Use the following template:", -1)
				msg(`"{\"Pass\": \"YOURPASSWORDHERE\" ,\"Host\": \"127.0.0.1\", \"Port\": 5432, \"DBName\": \"hipparchiaDB\" ,\"User\": \"hippa_wr\"}"`, 0)
			}
		case "-sa":
			cfg.HostIP = args[i+1]
		case "-sp":
			p, e := strconv.Atoi(args[i+1])
			chke(e)
			cfg.HostPort = p
		case "-ti":
			tt, e := strconv.Atoi(args[i+1])
			chke(e)
			cfg.MaxText = tt
		case "-ui":
			cfg.BadChars = args[i+1]
		default:
			// do nothing
		}
	}

	type ConfigFile struct {
		PosgreSQLPassword string
	}

	cfg.PGLogin = PostgresLogin{}
	if pl.Pass != "" {
		cfg.PGLogin = pl
	} else {
		file, _ := os.Open(cf)
		decoder := json.NewDecoder(file)
		conf := ConfigFile{}
		err := decoder.Decode(&conf)
		if err != nil {
			msg(fmt.Sprintf("FAILED to load the configuration file: '%s'", cf), 0)
			msg(fmt.Sprintf("Make sure that the file exists and that it has the following format:"), 0)
			fmt.Println(MINCONFIG)
			os.Exit(0)
		}
		cfg.PGLogin = PostgresLogin{
			Host:   PSQLHOST,
			Port:   PSQLPORT,
			User:   PSQLUSER,
			Pass:   conf.PosgreSQLPassword,
			DBName: PSQLDB,
		}
	}
}

func printversion() {
	versioninfo := fmt.Sprintf("%s (v%s)", MYNAME, VERSION)
	versioninfo = versioninfo + fmt.Sprintf(" [loglevel=%d]", cfg.LogLevel)
	msg(versioninfo, 0)
}
