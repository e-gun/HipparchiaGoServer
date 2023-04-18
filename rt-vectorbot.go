//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"sort"
	"strings"
	"time"
)

//
// VECTORBOT
//

// RtVectorBot - build a model for the author table requested; should only be called by activatevectorbot()
func RtVectorBot(c echo.Context) error {
	const (
		MSG3 = "attempted access to RtVectorBot() by foreign IP: '%s'"
		MSG4 = "RtVectorBot() building a model for '%s' (%d tables) [maxlines=%d]"
	)

	// the question is how much time are you saving vs how much space are you wasting
	// Catullus is 2555 lines and can be vectorized in just a couple of seconds: easy to leave this as ad hoc
	// Caesar is 11038 lines and requires 5s or so on a fast machine; 10s on a 'medium' machine

	// first 500 will take 13MG if MINSIZE = 10000; 118MB if MINSIZE = 1000...
	// if MINSIZE = 10000: 170MB @ 1000; 319MB @ 2000; 729MB @ 2500; 930MB @ 3000; 1021MB @ 3456
	// if VectorMaxlines = 1M, then 'gr' adds only 58MB to the db and takes 365s
	// all authors and all categories: 1311MB after 4804.478s (i.e. 80min) on a mac studio

	// the default is just 'gr' and 'lt'; and this yields different results
	//[HGS] [AV: 3278.550s][Δ: 0.007s] (#2186) checking need to model gr (gr)
	//[HGS] RtVectorBot() building a model for 'gr' (1823 tables) [maxlines=1000000]
	//[HGS] [AV: 3643.051s][Δ: 364.501s] (#2187) checking need to model lt (lt)
	//[HGS] RtVectorBot() building a model for 'lt' (362 tables) [maxlines=1000000]
	//[HGS] [VB: 4060.712s][Δ: 417.661s] The vectorbot has checked all authors and is now shutting down
	//[HGS] Disk space used by stored vectors is currently 997MB

	// if you adjust maxlines...
	//[HGS] [AV: 504.021s][Δ: 0.046s] (100.0%) checking need to model gr (gr)
	//[HGS] RtVectorBot() building a model for 'gr' (1823 tables) [maxlines=1250000]
	//[HGS] [AV: 946.632s][Δ: 442.611s] (100.0%) checking need to model lt (lt)
	//[HGS] RtVectorBot() building a model for 'lt' (362 tables) [maxlines=1250000]

	// 6 cores of intel 9900k
	// [HGS] [VB: 7602.934s][Δ: 877.251s] The vectorbot has checked all authors and is now shutting down

	// testable via:
	// curl localhost:8000/vbot/ch0d09

	if Config.VectorsDisabled {
		return nil
	}

	if c.RealIP() != Config.HostIP {
		msg(fmt.Sprintf(MSG3, c.RealIP()), MSGNOTE)
		return nil
	}

	req := c.Param("typeandselection")
	ts := strings.Split(req, "/")

	if len(ts) != 2 {
		return nil
	}

	vtype := ts[0]
	a := ts[1]

	s := BuildDefaultSearch(c)
	AllSearches.InsertSS(s)

	allof := func(db string) []string {
		allauth := StringMapKeysIntoSlice(AllAuthors)
		var dbauth []string
		for _, au := range allauth {
			if strings.HasPrefix(au, db) {
				dbauth = append(dbauth, au)
			}
		}
		return dbauth
	}

	dbs := []string{"lt", "gr", "in", "dp", "ch"}

	if IsInSlice(a, dbs) {
		s.SearchIn.Authors = allof(a)
		m := fmt.Sprintf(MSG4, a, len(s.SearchIn.Authors), Config.VectorMaxlines)
		msg(m, MSGFYI)
	} else {
		if _, ok := AllAuthors[a]; !ok {
			return nil
		}
		s.SearchIn.Authors = []string{a}
	}

	switch vtype {
	case "lda":
		ldamodelbot(c, s, a)
	default:
		// "nn"
		nnmodelbot(c, s, a)
	}
	AllSearches.Delete(s.ID)
	return nil
}

// ldamodelbot - automate the building of lda models
func ldamodelbot(c echo.Context, s SearchStruct, a string) {
	// note that only the selftest suite is calling this right now

	// there is no storage mechanism for lda

	// in fact pre-building works makes more sense than authors
	// and the caps need to be borne in mind

	// sessionintobulksearch() can't be used because there is no real session...

	s.CurrentLimit = Config.VectorMaxlines
	s.Seeking = ""
	// s.ID = strings.Replace(uuid.New().String(), "-", "", -1)
	s.ID = "ldamodelbot()"
	// note that you do not want multiple vectorbots running at once...
	SSBuildQueries(&s)
	s.IsActive = true
	s.TableSize = 1
	s = HGoSrch(s)
	e := LDASearch(c, s)
	if e != nil {
		msg("ldamodelbot() could not execute LDASearch()", MSGWARN)
	}
}

// nnmodelbot - automate the building of nn models
func nnmodelbot(c echo.Context, s SearchStruct, a string) {
	const (
		MSG1    = "RtVectorBot() found model for %s"
		MSG2    = "RtVectorBot() skipping %s - only %d lines found"
		MINSIZE = 10000
	)
	m := Config.VectorModel
	fp := fingerprintnnvectorsearch(s)

	isstored := vectordbcheck(fp)

	if isstored {
		msg(fmt.Sprintf(MSG1, AllAuthors[a].Name), MSGPEEK)
	} else {
		// sessionintobulksearch() can't be used because there is no real session...
		s.CurrentLimit = Config.VectorMaxlines
		s.Seeking = ""
		s.ID = strings.Replace(uuid.New().String(), "-", "", -1)
		SSBuildQueries(&s)
		s.IsActive = true
		s.TableSize = 1
		s = HGoSrch(s)
		if len(s.Results) > MINSIZE {
			embs := generateembeddings(c, m, s)
			vectordbadd(fp, embs)
		} else {
			msg(fmt.Sprintf(MSG2, a, len(s.Results)), MSGTMI)
		}
	}
}

// activatevectorbot - build a vector model for every author
func activatevectorbot() {
	const (
		MSG1       = "activatevectorbot(): launching"
		MSG2       = "(%.1f%%) checking need to model %s (%s)"
		MSG3       = "The vectorbot has checked all authors and is now shutting down"
		URL        = "http://%s:%d/vbot/%s/%s"
		COUNTEVERY = 10
		THROTTLE   = 4
		SIZEVERY   = 500
		STARTDELAY = 2
	)

	// currently only autovectorizes nn
	// lda unsupported, but a possibility later

	msg(MSG1, MSGNOTE)

	time.Sleep(STARTDELAY * time.Second)

	start := time.Now()
	previous := time.Now()

	auu := StringMapKeysIntoSlice(AllAuthors)
	sort.Strings(auu)

	// dbs := []string{"lt", "gr", "in", "dp", "ch"}

	// only model the ones you actually use regularly
	var trimmedauu []string
	var dbs []string

	for k, v := range Config.DefCorp {
		if v {
			for _, a := range auu {
				if strings.HasPrefix(a, k) {
					trimmedauu = append(trimmedauu, a)
				}
			}
			dbs = append(dbs, k)
		}
	}

	auu = append(trimmedauu, dbs...)

	tot := float32(len(auu))
	count := 0

	for _, a := range auu {
		mustnotify := false
		an := AllAuthors[a].Name
		if IsInSlice(a, dbs) {
			mustnotify = true
			an = a
		}

		count += 1
		if count%COUNTEVERY == 0 || mustnotify {
			TimeTracker("AV", fmt.Sprintf(MSG2, float32(count)/tot*100, an, a), start, previous)
			previous = time.Now()
		}
		u := fmt.Sprintf(URL, Config.HostIP, Config.HostPort, "nn", a)
		_, err := http.Get(u)
		chke(err)

		// if you do not throttle the bot it will violate MAXECHOREQPERSECONDPERIP
		time.Sleep(THROTTLE * time.Millisecond)

		if count%SIZEVERY == 0 {
			vectordbsize(MSGNOTE)
		}
	}

	TimeTracker("VB", MSG3, start, previous)
	vectordbsize(MSGNOTE)
	vectordbcount(MSGNOTE)
}
