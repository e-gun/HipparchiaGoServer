//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/base/mm"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/mps"
	"github.com/e-gun/HipparchiaGoServer/internal/search"
	"github.com/e-gun/HipparchiaGoServer/internal/vec"
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"slices"
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
	// curl localhost:8000/vbot/nn/gr0011

	if lnch.Config.VectorsDisabled {
		return nil
	}

	if c.RealIP() != lnch.Config.HostIP {
		Msg.NOTE(fmt.Sprintf(MSG3, c.RealIP()))
		return nil
	}

	req := c.Param("typeandselection")
	ts := strings.Split(req, "/")

	if len(ts) != 2 {
		return nil
	}

	vtype := ts[0]
	a := ts[1]

	s := search.BuildDefaultSearch(c)
	s.ID = "RtVectorBot-" + vtype + "-" + strings.Replace(uuid.New().String(), "-", "", -1)

	allof := func(db string) []string {
		allauth := gen.StringMapKeysIntoSlice(mps.AllAuthors)
		var dbauth []string
		for _, au := range allauth {
			if strings.HasPrefix(au, db) {
				dbauth = append(dbauth, au)
			}
		}
		return dbauth
	}

	var dbs []string
	for db := range mps.LoadedCorp {
		dbs = append(dbs, db)
	}

	if slices.Contains(dbs, a) {
		s.SearchIn.Authors = allof(a)
		m := fmt.Sprintf(MSG4, a, len(s.SearchIn.Authors), lnch.Config.VectorMaxlines)
		Msg.FYI(m)
	} else {
		if _, ok := mps.AllAuthors[a]; !ok {
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

	vlt.WSInfo.Del <- s.WSID
	return nil
}

// ldamodelbot - automate the building of lda models
func ldamodelbot(c echo.Context, s str.SearchStruct, a string) {
	// note that only the selftestsuite suite is calling this right now

	// there is no storage mechanism for lda

	// in fact pre-building works makes more sense than authors
	// and the caps need to be borne in mind

	// SessionIntoBulkSearch() can't be used because there is no real session...

	s.CurrentLimit = lnch.Config.VectorMaxlines
	s.Seeking = ""

	// do not edit the next variable without appreciating that there are "if srch.ID == "ldamodelbot()" checks elsewhere
	// note also that you REALLY do not want multiple vectorbots running at once...
	s.ID = "ldamodelbot()"

	s.Seeking = ""
	s.Proximate = ""
	s.LemmaOne = ""
	s.LemmaTwo = ""
	s.SkgSlice = []string{}

	search.SSBuildQueries(&s)
	s.IsActive = true
	s.TableSize = 1
	search.SearchAndInsertResults(&s)
	e := vec.LDASearch(c, s)
	if e != nil {
		Msg.WARN("ldamodelbot() could not execute LDASearch()")
	}
}

// nnmodelbot - automate the building of nn models
func nnmodelbot(c echo.Context, s str.SearchStruct, a string) {
	const (
		MSG1    = "RtVectorBot() found model for %s"
		MSG2    = "RtVectorBot() skipping %s - only %d lines found"
		MINSIZE = 10000
	)
	m := lnch.Config.VectorModel
	fp := vec.FingerprintNNVectorSearch(s)

	// bot hangs here on gr0063 (Dionysius Thrax)
	isstored := vec.VectorDBCheckNN(fp)

	if isstored {
		// 'gr' is a possible collection, but not a possible au name
		nm := a
		if _, ok := mps.AllAuthors[a]; ok {
			nm = mps.AllAuthors[a].Name
		}
		Msg.PEEK(fmt.Sprintf(MSG1, nm))
	} else {
		// SessionIntoBulkSearch() can't be used because there is no real session...
		s.CurrentLimit = lnch.Config.VectorMaxlines
		s.Seeking = ""
		s.Proximate = ""
		s.LemmaOne = ""
		s.LemmaTwo = ""
		s.SkgSlice = []string{}

		search.SSBuildQueries(&s)
		s.IsActive = true
		s.TableSize = 1
		search.SearchAndInsertResults(&s)
		if s.Results.Len() > MINSIZE {
			embs := vec.GenerateVectEmbeddings(c, m, s)
			vec.VectorDBAddNN(fp, embs)
		} else {
			Msg.TMI(fmt.Sprintf(MSG2, a, s.Results.Len()))
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
		THROTTLE   = 25
		SIZEVERY   = 500
		STARTDELAY = 2
	)

	if !lnch.Config.VectorBot {
		return
	}

	// currently only autovectorizes nn
	// lda unsupported, but a possibility later

	Msg.NOTE(MSG1)

	time.Sleep(STARTDELAY * time.Second)

	start := time.Now()
	previous := time.Now()

	auu := gen.StringMapKeysIntoSlice(mps.AllAuthors)
	sort.Strings(auu)

	var dbs []string
	for db := range mps.LoadedCorp {
		dbs = append(dbs, db)
	}

	var trimmedauu []string

	for _, k := range dbs {
		for _, a := range auu {
			if strings.HasPrefix(a, k) {
				trimmedauu = append(trimmedauu, a)
			}
		}
	}

	auu = append(trimmedauu, dbs...)

	tot := float32(len(auu))
	count := 0

	for _, a := range auu {
		mustnotify := false

		var an string
		if _, ok := mps.AllAuthors[a]; ok {
			// full corpora "gr", "lt", etc. can be in here too; AllAuthors[a].Name will not find  AllAuthors["gr"]
			an = mps.AllAuthors[a].Name
		} else {
			an = a
		}

		if slices.Contains(dbs, a) {
			mustnotify = true
			an = a
		}

		count += 1
		if count%COUNTEVERY == 0 || mustnotify {
			Msg.Timer("AV", fmt.Sprintf(MSG2, float32(count)/tot*100, an, a), start, previous)
			previous = time.Now()
		}
		u := fmt.Sprintf(URL, lnch.Config.HostIP, lnch.Config.HostPort, "nn", a)
		_, err := http.Get(u)
		Msg.EC(err)

		// if you do not throttle the bot it will violate MAXECHOREQPERSECONDPERIP
		time.Sleep(THROTTLE * time.Millisecond)

		if count%SIZEVERY == 0 {
			vec.VectorDBSizeNN(mm.MSGNOTE)
		}
	}

	Msg.Timer("VB", MSG3, start, previous)
	vec.VectorDBSizeNN(mm.MSGNOTE)
	vec.VectorDBCountNN(mm.MSGNOTE)
	lnch.Config.VectorBot = false
}
