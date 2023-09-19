//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"os"
	"sort"
	"strings"
)

var ()

// NeighborsSearch - a special case for RtSearch() where you requested vectorization of the results
func NeighborsSearch(c echo.Context, srch SearchStruct) error {
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

		SETTINGS = `model type: %s; text prep: %s`
	)

	c.Response().After(func() { messenger.GCStats("NeighborsSearch()") })
	se := srch.StoredSession

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
	set := fmt.Sprintf(SETTINGS, se.VecModeler, se.VecTextPrep)

	// [a] prepare the image output
	blank := buildblanknngraph(set, term, srch.InclusionOverview(se.Inclusions))
	graph := formatnngraph(c, blank, term, nn)
	img := customnngraphhtmlandjs(graph)

	neighbors := nn[term]

	// [b] prepare text output
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

	out := fmt.Sprintf(THETABLE, term, strings.Join(tablerows, "\n"), se.VecModeler, se.VecTextPrep)

	soj := SearchOutputJSON{
		Title:         fmt.Sprintf("Neighbors of '%s'", term),
		Searchsummary: "",
		Found:         out,
		Image:         img,
		JS:            VECTORJS,
	}

	AllSearches.Delete(srch.ID)

	return JSONresponse(c, soj)
}

// fingerprintnnvectorsearch - derive a unique md5 for any given mix of search items & vector settings
func fingerprintnnvectorsearch(srch SearchStruct) string {
	const (
		MSG1 = "NeighborsSearch() fingerprint: "
		FAIL = "fingerprintnnvectorsearch() failed to Marshal"
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

	switch srch.VecModeler {
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
	f1 = append(f1, []byte(srch.VecTextPrep)...)

	m := fmt.Sprintf("%x", md5.Sum(f1))
	msg(MSG1+m, MSGTMI)

	return m
}
