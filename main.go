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

	versioninfo := fmt.Sprintf("%s CLI Debugging Interface (v.%s)", MYNAME, VERSION)
	versioninfo = versioninfo + fmt.Sprintf(" [loglevel=%d]", cfg.LogLevel)
	msg(versioninfo, 0)

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

// configatlaunch - read the configuration values from JSON and/or command line
func configatlaunch() {
	config := fmt.Sprintf("%s/%s", CONFIGLOCATION, CONFIGNAME)

	args := os.Args[1:len(os.Args)]

	for i, a := range args {
		if a == "-v" {
			// version always printed anyway
			os.Exit(1)
		}
		if a == "-gl" {
			ll, e := strconv.Atoi(args[i+1])
			chke(e)
			cfg.LogLevel = ll
		}
		if a == "-el" {
			ll, e := strconv.Atoi(args[i+1])
			chke(e)
			cfg.EchoLog = ll
		}
		if a == "-c" {
			config = args[i+1]
		}
	}

	type ConfigFile struct {
		PosgreSQL PostgresLogin
	}

	file, _ := os.Open(config)
	decoder := json.NewDecoder(file)
	conf := ConfigFile{}
	err := decoder.Decode(&conf)
	if err != nil {
		msg(fmt.Sprintf("failed to load configuration file: '%s'", config), 0)
	}

	cfg.PGLogin = conf.PosgreSQL
}
