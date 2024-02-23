//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"os"
	"slices"
	"strings"
)

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

	c.Response().After(func() { messenger.LogPaths("NeighborsSearch()") })
	sess := srch.StoredSession

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
	set := fmt.Sprintf(SETTINGS, sess.VecModeler, sess.VecTextPrep)

	// [a] prepare the image output
	blank := buildblanknngraph(set, term, srch.InclusionOverview(sess.Inclusions))
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

	out := fmt.Sprintf(THETABLE, term, strings.Join(tablerows, "\n"), sess.VecModeler, sess.VecTextPrep)

	soj := SearchOutputJSON{
		Title:         fmt.Sprintf("Neighbors of '%s'", term),
		Searchsummary: "",
		Found:         out,
		Image:         img,
		JS:            VECTORJS,
	}

	WSSIDel <- srch.WSID
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

	var fp []string

	// [1] start with the searchlist + the stoplists + VecTextPrep (which are all collections of strings + one string)

	// includes
	fp = append(fp, srch.SearchIn.AuGenres...)
	fp = append(fp, srch.SearchIn.WkGenres...)
	fp = append(fp, srch.SearchIn.AuLocations...)
	fp = append(fp, srch.SearchIn.WkLocations...)
	fp = append(fp, srch.SearchIn.Authors...)
	fp = append(fp, srch.SearchIn.Works...)
	fp = append(fp, srch.SearchIn.Passages...)

	// excludes
	fp = append(fp, srch.SearchEx.AuGenres...)
	fp = append(fp, srch.SearchEx.WkGenres...)
	fp = append(fp, srch.SearchEx.AuLocations...)
	fp = append(fp, srch.SearchEx.WkLocations...)
	fp = append(fp, srch.SearchEx.Authors...)
	fp = append(fp, srch.SearchEx.Works...)
	fp = append(fp, srch.SearchEx.Passages...)

	// stops
	fp = append(fp, readstopconfig("greek")...)
	fp = append(fp, readstopconfig("latin")...)

	// one last item...
	fp = append(fp, srch.VecTextPrep)
	slices.Sort(fp)

	f1, e1 := json.Marshal(fp)

	// [2] now add in the vector settings (which have an underlying Options struct)

	var f2 []byte
	var e2 error

	switch srch.VecModeler {
	case "glove":
		ff, ee := json.Marshal(glovevectorconfig())
		f2 = ff
		e2 = ee
	case "lexvec":
		ff, ee := json.Marshal(lexvecvectorconfig())
		f2 = ff
		e2 = ee
	default: // w2v
		ff, ee := json.Marshal(w2vvectorconfig())
		f2 = ff
		e2 = ee
	}

	if e1 != nil || e2 != nil {
		msg(FAIL, 0)
		os.Exit(1)
	}

	// [3] merge the previous two into a single byte array

	f1 = append(f1, f2...)

	// [4] generate the md5 fingerprint from this

	m := fmt.Sprintf("%x", md5.Sum(f1))
	msg(MSG1+m, MSGTMI)

	return m
}
