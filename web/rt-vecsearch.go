//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/search"
	"github.com/e-gun/HipparchiaGoServer/internal/vec"
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/e-gun/nlp"
	"github.com/labstack/echo/v4"
	"gonum.org/v1/gonum/mat"
	"strings"
)

//
// these are two pseudo-routes: RtSearch() will send you here conditionally before you exit via jsonresponse()
//

// RtLDASearch - a special case for RtSearch() where you requested Latent Dirichlet Allocation vectorization of the results
func RtLDASearch(c echo.Context, srch str.SearchStruct) error {
	const (
		LDAMSG = `Building LDA model for the current selections`
		ESM1   = "Preparing the text for modeling"
		ESM2   = "Building topic models"
		ESM3   = "Building the graph (please be patient this can be very slow...)"
	)
	c.Response().After(func() { vlt.LogPaths("RtLDASearch()") })

	se := srch.StoredSession
	ntopics := se.LDAtopics
	if ntopics < 1 {
		ntopics = vv.LDATOPICS
	}

	var vs str.SearchStruct
	if srch.ID != "ldamodelbot()" {
		vs = search.SessionIntoBulkSearch(c, lnch.Config.VectorMaxlines)
		vlt.WSInfo.UpdateRemain <- vlt.WSSIKVi{vs.WSID, 1}
		vlt.WSInfo.UpdateSummMsg <- vlt.WSSIKVs{vs.WSID, LDAMSG}
		vlt.WSInfo.UpdateVProgMsg <- vlt.WSSIKVs{vs.WSID, fmt.Sprintf(ESM1)}
	} else {
		vs = srch
	}

	bags := vec.LDAPrepText(se.VecTextPrep, &vs)

	corpus := make([]string, len(bags))
	for i := 0; i < len(bags); i++ {
		corpus[i] = bags[i].ModifiedBag
	}

	stops := gen.StringMapKeysIntoSlice(vec.GetStopSet())
	vectoriser := nlp.NewCountVectoriser(stops...)

	vlt.WSInfo.UpdateSummMsg <- vlt.WSSIKVs{vs.WSID, fmt.Sprintf(ESM2)}

	// consider building TESTITERATIONS models and making a table for each
	var dot mat.Matrix
	var tables []string

	// a chance to bail if you hit RtResetSession() in time
	if lnch.Config.SelfTest == 0 && !lnch.Config.VectorBot && !vlt.AllSessions.IsInVault(vs.User) {
		return jsonresponse(c, str.SearchOutputJSON{})
	}

	docsOverTopics, topicsOverWords, ok := vec.LDAModel(ntopics, corpus, vectoriser, &vs)
	if !ok {
		return jsonresponse(c, str.SearchOutputJSON{})
	}

	tables = append(tables, vec.LDATopicSummary(ntopics, topicsOverWords, vectoriser, docsOverTopics))
	tables = append(tables, vec.LDATopicSentences(ntopics, bags, corpus, docsOverTopics))
	dot = docsOverTopics

	htmltables := strings.Join(tables, "")

	incl := search.InclusionOverview(&srch, se.Inclusions)

	var img string
	if se.LDAgraph || srch.ID == "ldamodelbot()" {
		vlt.WSInfo.UpdateSummMsg <- vlt.WSSIKVs{vs.WSID, fmt.Sprintf(ESM3)}
		img = vec.LDAPlot(vs.Context, se.LDA2D, ntopics, incl, se.VecTextPrep, dot, bags)
	}

	soj := str.SearchOutputJSON{
		Title:         "",
		Searchsummary: "",
		Found:         htmltables,
		Image:         img,
		JS:            vv.VECTORJS,
	}

	vlt.WSInfo.Del <- srch.ID
	vlt.WSInfo.Del <- vs.ID

	return jsonresponse(c, soj)
}

// RtNeighborsSearch - a special case for RtSearch() where you requested NN vectorization of the results
func RtNeighborsSearch(c echo.Context, srch str.SearchStruct) error {
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

	c.Response().After(func() { vlt.LogPaths("RtNeighborsSearch()") })
	sess := srch.StoredSession

	term := srch.LemmaOne
	if term == "" {
		// JS not supposed to let this happen, but...
		term = srch.Seeking
	}

	// the words in the model have different formation rules from the hints supplied...
	term = strings.ToLower(term)
	term = gen.RestoreInitialVJ(term) // a quick, partial fix...
	srch.LemmaOne = term

	nn := vec.GenerateNeighborsData(c, srch)
	set := fmt.Sprintf(SETTINGS, sess.VecModeler, sess.VecTextPrep)

	// [a] prepare the image output
	io := search.InclusionOverview(&srch, sess.Inclusions)
	blank := vec.BuildBlankNNGraph(set, term, io)
	graph := vec.FormatNNGraph(c, blank, term, nn)
	img := vec.CustomNNGraphHTMLandJS(graph)

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

	soj := str.SearchOutputJSON{
		Title:         fmt.Sprintf("Neighbors of '%s'", term),
		Searchsummary: "",
		Found:         out,
		Image:         img,
		JS:            vv.VECTORJS,
	}

	vlt.WSInfo.Del <- srch.WSID
	return jsonresponse(c, soj)
}
