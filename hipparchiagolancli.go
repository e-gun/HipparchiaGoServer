//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"C"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	MYNAME        = "Hipparchia Golang Server"
	SHORTNAME     = "HGS"
	VERSION       = "0.1.2"
	PSQ           = `{"Host": "localhost", "Port": 5432, "User": "hippa_wr", "Pass": "", "DBName": "hipparchiaDB"}`
	PSDefaultHost = "localhost"
	PSDefaultUser = "hippa_wr"
	PSDefaultPort = 5432
	PSDefaultDB   = "hipparchiaDB"
	TwoPassThresh = 100 // cicero has >70 works
)

func main() {

	makeconfig()

	// the command line arguments are getting lost

	// main() instead has a cfg with the defaults burned into it
	// so we do this the stupid/bound to fail way...

	// fmt.Println(os.Args[1:len(os.Args)])

	args := os.Args[1:len(os.Args)]

	for i, a := range args {
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
		if a == "-skip" {
			cfg.SkipLemm = true
		} else {
			cfg.SkipLemm = false
		}
	}

	versioninfo := fmt.Sprintf("%s CLI Debugging Interface (v.%s)", MYNAME, VERSION)
	versioninfo = versioninfo + fmt.Sprintf(" [loglevel=%d]", cfg.LogLevel)

	if cfg.LogLevel > 5 {
		cfg.LogLevel = 5
	}

	if cfg.LogLevel < 0 {
		cfg.LogLevel = 0
	}

	cfg.PGLogin = decodepsqllogin([]byte(cfg.PosgresInfo))

	fmt.Println(versioninfo)

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
		timetracker("A3", "corpus maps built", start, previous)
	}(&awaiting)

	if !cfg.SkipLemm {
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
	}

	awaiting.Wait()
	StartEchoServer()
}
