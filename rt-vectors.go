//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
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
        <td class="vectorrank small" colspan = "7">(model type: <span class="dbb">%s</span>)</td>
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
	out := fmt.Sprintf(THETABLE, term, strings.Join(tablerows, "\n"), se.VecModeler)

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
func fingerprintvectorsearch(srch SearchStruct, modeltype string) string {
	const (
		MSG1 = "VectorSearch() fingerprint: "
		FAIL = "fingerprintvectorsearch() failed to Marshal"
	)

	// vectorbot vs normal surfer requires passing the model type (bot: configtype; surfer: sessiontype)

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
	default:
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
	m := fmt.Sprintf("%x", md5.Sum(f1))
	msg(MSG1+m, MSGTMI)

	return m
}

// buildtextblock - turn []DbWorkline into a single long string
func buildtextblock(lines []DbWorkline) string {
	const (
		FAIL1 = "failed to unmarshal %s into objmap\n"
		FAIL2 = "failed second pass unmarshal of %s into newmap\n"
	)
	// [a] get all the words we need
	var slicedwords []string
	for i := 0; i < len(lines); i++ {
		wds := lines[i].AccentedSlice()
		for _, w := range wds {
			slicedwords = append(slicedwords, UVσςϲ(SwapAcuteForGrave(w)))
		}
	}

	// [b] get basic morphology info for those words
	morphmapdbm := arraytogetrequiredmorphobjects(slicedwords) // map[string]DbMorphology

	// [c] figure out which headwords to associate with the collection of words

	// this information is inside DbMorphology.RawPossib
	// but it needs to be parsed
	// example: `{"1": {"transl": "A. nom. plur; II. a guardian god", "analysis": "masc gen pl", "headword": "deus", "scansion": "deūm", "xref_kind": "9", "xref_value": "22568216"}, "2": {"transl": "A. nom. plur; II. a guardian god", "analysis": "masc acc sg", "headword": "deus", "scansion": "", "xref_kind": "9", "xref_value": "22568216"}}`

	type possib struct {
		Trans string `json:"found"`
		Anal  string `json:"analysis"`
		Head  string `json:"headword"`
		Scans string `json:"scansion"`
		Xrefk string `json:"xref_kind"`
		Xrefv string `json:"xref_value"`
	}

	morphmapstrslc := make(map[string]map[string]bool, len(morphmapdbm))
	for m := range morphmapdbm {
		morphmapstrslc[m] = make(map[string]bool)
		// first pass: {"1": bytes1, "2": bytes2, ...}
		var objmap map[string]json.RawMessage
		err := json.Unmarshal([]byte(morphmapdbm[m].RawPossib), &objmap)
		if err != nil {
			fmt.Printf(FAIL1, morphmapdbm[m].Observed)
		}
		// second pass: : {"1": possib1, "2": possib2, ...}
		newmap := make(map[string]possib)
		for key, v := range objmap {
			var pp possib
			e := json.Unmarshal(v, &pp)
			if e != nil {
				fmt.Printf(FAIL2, morphmapdbm[m].Observed)
			}
			newmap[key] = pp
		}

		for _, v := range newmap {
			morphmapstrslc[m][v.Head] = true
		}
	}

	// if you just iterate over morphmapstrslc, you drop unparsed terms: the next will retain them
	for _, w := range slicedwords {
		if _, t := morphmapstrslc[w]; t {
			continue
		} else {
			morphmapstrslc[w] = make(map[string]bool)
			morphmapstrslc[w][w] = true
		}
	}

	// "morphmapstrslc" for Albinus , poet. [lt2002]
	// map[abscondere:map[abscondo:true] apte:map[apte:true aptus:true] capitolia:map[capitolium:true] celsa:map[celsus¹:true] concludere:map[concludo:true] cui:map[quis²:true quis¹:true qui²:true qui¹:true] dactylum:map[dactylus:true] de:map[de:true] deum:map[deus:true] fieri:map[fio:true] freta:map[fretum:true fretus¹:true] i:map[eo¹:true] ille:map[ille:true] iungens:map[jungo:true] liber:map[liber¹:true liber⁴:true libo¹:true] metris:map[metrum:true] moenibus:map[moenia¹:true] non:map[non:true] nulla:map[nullus:true] patuere:map[pateo:true patesco:true] posse:map[possum:true] repostos:map[re-pono:true] rerum:map[res:true] romanarum:map[romanus:true] sed:map[sed:true] sinus:map[sinus¹:true] spondeum:map[spondeum:true spondeus:true] sponte:map[sponte:true] ternis:map[terni:true] totum:map[totus²:true totus¹:true] triumphis:map[triumphus:true] tutae:map[tueor:true] uersum:map[verro:true versum:true versus³:true verto:true] urbes:map[urbs:true] †uilem:map[†uilem:true]]
	//

	// [d] swap out words for headwords
	winnermap := buildwinnertakesallparsemap(morphmapstrslc)

	// "winnermap" for Albinus , poet. [lt2002]
	// map[abscondere:[abscondo] apte:[aptus] capitolia:[capitolium] celsa:[celsus¹] concludere:[concludo] cui:[qui¹] dactylum:[dactylus] de:[de] deum:[deus] fieri:[fio] freta:[fretus¹] i:[eo¹] ille:[ille] iungens:[jungo] liber:[liber⁴] metris:[metrum] moenibus:[moenia¹] non:[non] nulla:[nullus] patuere:[pateo] posse:[possum] repostos:[re-pono] rerum:[res] romanarum:[romanus] sed:[sed] sinus:[sinus¹] spondeum:[spondeus] sponte:[sponte] ternis:[terni] totum:[totus¹] triumphis:[triumphus] tutae:[tueor] uersum:[verro] urbes:[urbs] †uilem:[†uilem]]

	// [e] turn results into unified text block

	// string addition will use a huge amount of time: 120s to concatinate Cicero: txt = txt + newtxt...
	// with strings.Builder we only need .1s to build the text...

	var sb strings.Builder
	preallocate := CHARSPERLINE * len(lines) // NB: a long line has 60 chars
	sb.Grow(preallocate)

	winner := true

	if winner {
		winnerstring(&sb, slicedwords, winnermap)
	} else {
		flatstring(&sb, slicedwords)
	}

	return strings.TrimSpace(sb.String())
}

//
// VECTORBOT
//

// RtVectorBot - build a model for the author table requested; should only be called by activatevectorbot()
func RtVectorBot(c echo.Context) error {
	const (
		MSG1    = "vectorbot found model for %s"
		MSG2    = "vectorbot skipping %s - only %d lines found"
		MINSIZE = 1000
	)

	a := c.Param("au")

	if _, ok := AllAuthors[a]; !ok {
		return nil
	}

	if Config.VectorsDisabled {
		return nil
	}

	s := BuildDefaultSearch(c)
	s.SearchIn.Authors = []string{a}

	m := Config.VectorModel
	fp := fingerprintvectorsearch(s, m)
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
		MSG4       = "Total size of stored vectors is %dMB"
		URL        = "http://%s:%d/vbot/%s"
		COUNTEVERY = 5
		THROTTLE   = 5
		SZQ        = "SELECT SUM(vectorsize) AS total FROM semantic_vectors"
		SIZEVERY   = 500
	)

	time.Sleep(2 * time.Second)

	count := 0
	var size int64

	vs := func() int64 {
		dbconn := GetPSQLconnection()
		defer dbconn.Release()
		err := dbconn.QueryRow(context.Background(), SZQ).Scan(&size)
		chke(err)
		return size
	}()

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
			msg(fmt.Sprintf(MSG4, vs/1024/1024), MSGNOTE)
		}
	}

	TimeTracker("VB", MSG3, start, previous)
	msg(fmt.Sprintf(MSG4, vs/1024/1024), MSGNOTE)
}
