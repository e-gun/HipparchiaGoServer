//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// VectorSearch - a special case for RtSearch() where you requested vectorization of the results
func VectorSearch(c echo.Context, srch SearchStruct) error {
	const (
		NTH      = 3
		THETABLE = `
	<table class="vectortable"><tbody>
    <tr class="vectorrow">
        <td class="vectorrank" colspan = "7">Nearest neighbors of »<span class="colorhighlight">%s</span>«</td>
    </tr>
	<tr class="vectorrow">
		<td class="vectorrank">Rank</td>
		<td class="vectorrank">Distance</td>
		<td class="vectorrank">Word</td>
		<td class="vectorrank">&nbsp;&nbsp;&nbsp;</td>
		<td class="vectorrank">Rank</td>
		<td class="vectorrank">Distance</td>
		<td class="vectorrank">Word</td>
	</tr>
    %s
    <tr class="vectorrow">
        <td class="vectorrank small" colspan = "7">(model type: <code>%s</code>; text prep: <code>%s</code>)</td>
    </tr>
	</tbody></table>
	<hr>`

		TABLEROW = `
	<tr class="%s">%s
		<td class="vectorword">&nbsp;&nbsp;&nbsp;</td>%s
	</tr>`

		TABLEELEM = `
		<td class="vectorrank">%d</td>
		<td class="vectorscore">%.4f</td>
		<td class="vectorword"><vectorheadword id="%s">%s</vectorheadword></td>`
	)

	c.Response().After(func() { SelfStats("VectorSearch()") })

	term := srch.LemmaOne
	if term == "" {
		// JS not supposed to let this happen, but...
		term = srch.Seeking
	}

	// the words in the model have different formation rules from the hints supplied...
	term = strings.ToLower(term)
	term = RestoreInitialVJ(term) // a quick, partial fix...
	srch.LemmaOne = term

	nn := generateneighborsdata(c, srch)
	img := buildgraph(c, term, nn)

	neighbors := nn[term]
	// [c] prepare text output

	var columnone []string
	var columntwo []string

	half := len(neighbors) / 2
	for i, n := range neighbors {
		r := fmt.Sprintf(TABLEELEM, n.Rank, n.Similarity, n.Word, n.Word)
		if i < half {
			columnone = append(columnone, r)
		} else {
			columntwo = append(columntwo, r)
		}
	}

	var tablerows []string
	for i := range columnone {
		rn := "vectorrow"
		if i%NTH == 0 {
			rn = "nthrow"
		}
		tablerows = append(tablerows, fmt.Sprintf(TABLEROW, rn, columnone[i], columntwo[i]))
	}

	se := AllSessions.GetSess(readUUIDCookie(c))
	out := fmt.Sprintf(THETABLE, term, strings.Join(tablerows, "\n"), se.VecModeler, se.VecTextPrep)

	soj := SearchOutputJSON{
		Title:         fmt.Sprintf("Neighbors of '%s'", term),
		Searchsummary: "",
		Found:         out,
		Image:         img,
		JS:            VECTORJS,
	}

	AllSearches.Delete(srch.ID)

	return c.JSONPretty(http.StatusOK, soj, JSONINDENT)
}

// fingerprintvectorsearch - derive a unique md5 for any given mix of search items & vector settings
func fingerprintvectorsearch(srch SearchStruct, modeltype string, textprep string) string {
	const (
		MSG1 = "VectorSearch() fingerprint: "
		FAIL = "fingerprintvectorsearch() failed to Marshal"
	)

	// vectorbot vs normal surfer requires passing the model type and textprep style (bot: configtype; surfer: sessiontype)

	// unless you sort, you do not get repeatable results with a md5sum of srch.SearchIn if you look at "all latin"
	var inc []string
	sort.Strings(srch.SearchIn.AuGenres)
	sort.Strings(srch.SearchIn.WkGenres)
	sort.Strings(srch.SearchIn.AuLocations)
	sort.Strings(srch.SearchIn.WkLocations)
	sort.Strings(srch.SearchIn.Authors)
	sort.Strings(srch.SearchIn.Works)
	sort.Strings(srch.SearchIn.Passages)
	inc = append(inc, srch.SearchIn.AuGenres...)
	inc = append(inc, srch.SearchIn.WkGenres...)
	inc = append(inc, srch.SearchIn.AuLocations...)
	inc = append(inc, srch.SearchIn.WkLocations...)
	inc = append(inc, srch.SearchIn.Authors...)
	inc = append(inc, srch.SearchIn.Works...)
	inc = append(inc, srch.SearchIn.Passages...)

	var exc []string
	sort.Strings(srch.SearchEx.AuGenres)
	sort.Strings(srch.SearchEx.WkGenres)
	sort.Strings(srch.SearchEx.AuLocations)
	sort.Strings(srch.SearchEx.WkLocations)
	sort.Strings(srch.SearchEx.Authors)
	sort.Strings(srch.SearchEx.Works)
	sort.Strings(srch.SearchEx.Passages)
	exc = append(exc, srch.SearchEx.AuGenres...)
	exc = append(exc, srch.SearchEx.WkGenres...)
	exc = append(exc, srch.SearchEx.AuLocations...)
	exc = append(exc, srch.SearchEx.WkLocations...)
	exc = append(exc, srch.SearchEx.Authors...)
	exc = append(exc, srch.SearchEx.Works...)
	exc = append(exc, srch.SearchEx.Passages...)

	var stops []string
	stops = readstopconfig("latin")
	stops = append(stops, readstopconfig("greek")...)
	sort.Strings(stops)

	f1, e1 := json.Marshal(inc)
	f2, e2 := json.Marshal(exc)
	f3, e3 := json.Marshal(stops)

	var f4 []byte
	var e4 error
	switch modeltype {
	case "glove":
		ff, ee := json.Marshal(glovevectorconfig())
		f4 = ff
		e4 = ee
	case "lexvec":
		ff, ee := json.Marshal(lexvecvectorconfig())
		f4 = ff
		e4 = ee
	default: // w2v
		ff, ee := json.Marshal(w2vvectorconfig())
		f4 = ff
		e4 = ee
	}

	if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
		msg(FAIL, 0)
		os.Exit(1)
	}

	f1 = append(f1, f2...)
	f1 = append(f1, f3...)
	f1 = append(f1, f4...)
	f1 = append(f1, []byte(textprep)...)

	m := fmt.Sprintf("%x", md5.Sum(f1))
	msg(MSG1+m, MSGTMI)

	return m
}

//
// VECTORBOT
//

// RtVectorBot - build a model for the author table requested; should only be called by activatevectorbot()
func RtVectorBot(c echo.Context) error {
	const (
		MSG1    = "vectorbot found model for %s"
		MSG2    = "vectorbot skipping %s - only %d lines found"
		MSG3    = "attempted access to vectorbot route by foreign IP: '%s'"
		MINSIZE = 1000
	)

	a := c.Param("au")

	if _, ok := AllAuthors[a]; !ok {
		return nil
	}

	if Config.VectorsDisabled {
		return nil
	}

	if c.RealIP() != Config.HostIP {
		msg(fmt.Sprintf(MSG3, c.RealIP()), MSGNOTE)
	}

	s := BuildDefaultSearch(c)
	s.SearchIn.Authors = []string{a}

	m := Config.VectorModel
	fp := fingerprintvectorsearch(s, m, Config.VectorTextPrep)

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

	return nil
}

// activatevectorbot - build a vector model for every author
func activatevectorbot() {
	const (
		MSG2       = "(#%d) ensure model for %s (%s)"
		MSG3       = "The vectorbot has checked all authors and is now shutting down"
		URL        = "http://%s:%d/vbot/%s"
		COUNTEVERY = 5
		THROTTLE   = 5
		SIZEVERY   = 500
	)

	time.Sleep(2 * time.Second)

	count := 0

	start := time.Now()
	previous := time.Now()

	auu := StringMapKeysIntoSlice(AllAuthors)
	sort.Strings(auu)

	for _, a := range auu {
		count += 1
		if count%COUNTEVERY == 0 {
			TimeTracker("VB", fmt.Sprintf(MSG2, count, AllAuthors[a].Name, a), start, previous)
			previous = time.Now()
		}
		u := fmt.Sprintf(URL, Config.HostIP, Config.HostPort, a)
		_, err := http.Get(u)
		chke(err)
		// if you do not throttle the bot it will violate MAXECHOREQPERSECONDPERIP
		time.Sleep(THROTTLE * time.Millisecond)

		if count%SIZEVERY == 0 {
			vectordbsize(MSGFYI)
		}
	}

	TimeTracker("VB", MSG3, start, previous)
	vectordbsize(MSGNOTE)
}
