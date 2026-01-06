//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/format"
	"github.com/e-gun/HipparchiaGoServer/internal/search"
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/labstack/echo/v4"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"net/http"
	"strings"
)

// RtTextMaker - make a text of whatever collection of lines you would be searching
func RtTextMaker(c echo.Context) error {
	c.Response().After(func() { vlt.LogPaths("RtTextMaker()") })

	// text generation works like a simple search for "anything" in each line of the selected texts
	// the results then get output as a big "browser table"...

	// it would be nice to make this into three columns so cut-and-paste was easy like the revised browser
	// but keeping the notes and citations aligned over hundreds/thousands of lines is not at all trivial

	// multiple stabs at this added a lot of complexity while still exposing tons of corner cases, etc.

	// things also got very slow... not much gain; much pain

	// for a working-ish draft see dbacac7cc06a1c2ce67efb3b1cb69df363c7cfa1

	const (
		TBLRW = `
            <tr class="browser">
                <td class="textembeddedannotations">%s</td>
                <td class="textline">%s</td>
                <td class="textcite">%s</td>
            </tr>
		`
		SUMM = `
		%s
		citation format:&nbsp;%s<br></div>`

		SNIP   = `✃✃✃`
		HITCAP = `<span class="small"><span class="red emph">text generation incomplete:</span> hit the cap of %d on allowed lines</span>`
	)

	user := vlt.ReadUUIDCookie(c)
	if !vlt.AllAuthorized.Check(user) {
		return c.JSONPretty(http.StatusOK, str.CommonJSONOutput{JS: vv.JSVALIDATION}, vv.JSONINDENT)
	}

	s := vlt.AllSessions.GetSess(user)
	srch := search.SessionIntoBulkSearch(c, vv.MAXTEXTLINEGENERATION)

	if srch.Results.Len() == 0 {
		return emptyjsreturn(c)
	}

	// now we have the lines we need....
	firstline := srch.Results.FirstLine()
	firstwork := search.DbWlnMyWk(&firstline)
	firstauth := search.DbWlnMyAu(&firstline)

	lines := srch.Results.Yield()
	block := make([]string, srch.Results.Len())

	i := 0
	for l := range lines {
		block[i] = l.GetMarked()
		i++
	}

	whole := strings.Join(block, SNIP)
	whole = search.TextBlockCleaner(whole)
	block = strings.Split(whole, SNIP)

	for i = 0; i < len(block); i++ {
		srch.Results.Lines[i].MarkedUp = block[i]
	}

	// delete after use...
	whole = ""
	block = []string{""}

	trr := make([]string, srch.Results.Len())
	previous := srch.Results.FirstLine()
	workcount := 1

	i = 0
	lines = srch.Results.Yield()
	for l := range lines {
		cit := format.SelectivelyDisplayCitations(l, previous, -1)
		trr[i] = fmt.Sprintf(TBLRW, format.FormatAnnotations(l), l.GetMarked(), cit)
		if l.WkUID != previous.WkUID {
			// you were doing multi-text generation
			workcount += 1
			aw := search.DbWlnMyAu(&l).Name + fmt.Sprintf(`, <span class="italic">%s</span>`, search.DbWlnMyWk(&l).Title)
			aw = fmt.Sprintf(`<hr><span class="emph">[%d] %s</span>`, workcount, aw)
			extra := fmt.Sprintf(TBLRW, "", aw, "")
			trr[i] = extra + trr[i]
		}
		previous = l
		i++
	}

	tab := strings.Join(trr, "")
	// that was the body, now do the head and tail
	top := fmt.Sprintf(`<div id="browsertableuid" uid="%s"></div>`, srch.Results.Lines[0].AuID())
	top += `<table id="textmaker"><tbody>`
	top += `<tr class="spacing">` + strings.Repeat("&nbsp;", vv.MINBROWSERWIDTH) + `</tr>`

	tab = top + tab + `</tbody></table>`

	// but we don't want/need "observed" tags

	// <div id="searchsummary">Cicero,&nbsp;<span class="foundwork">Philippicae</span><br><br>citation format:&nbsp;oration 3, section 13, line 1<br></div>

	sui := s.Inclusions

	au := firstauth.Shortname
	if len(sui.Authors) > 1 || len(sui.AuGenres) > 0 || len(sui.AuLocations) > 0 {
		au += " (and others)"
	}

	ti := firstwork.Title
	if len(sui.Works) > 1 || len(sui.WkGenres) > 0 || len(sui.WkLocations) > 0 {
		ti += " (and others)"
	}

	ct := format.BasicCitation(firstline)

	// use the browser info...
	bc := format.ProlixBrowswerCitations(firstline, firstline)
	sum := fmt.Sprintf(SUMM, bc, ct)

	cp := ""
	if srch.Results.Len() == vv.MAXTEXTLINEGENERATION {
		m := message.NewPrinter(language.English)
		cp = m.Sprintf(HITCAP, vv.MAXTEXTLINEGENERATION)
	}
	sum = sum + cp

	if s.ZapLunates {
		tab = gen.DeLunate(tab)
		// overshot in DeLunate...: ς’ for σ’
		tab = strings.Replace(tab, " ς’ ", " σ’ ", -1)
	}

	var jso str.CommonJSONOutput
	jso.Sum = sum
	jso.Htm = tab
	jso.JS = ""
	jso.Tit = fmt.Sprintf("Text of %s, %s", au, ti)

	vlt.WSInfo.Del <- srch.WSID

	return jsonresponse(c, jso)
}
