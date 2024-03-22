//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/search"
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/labstack/echo/v4"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"net/http"
	"strings"
)

//
// ROUTES
//

// RtTextMaker - make a text of whatever collection of lines you would be searching
func RtTextMaker(c echo.Context) error {
	c.Response().After(func() { Msg.LogPaths("RtTextMaker()") })
	// text generation works like a simple search for "anything" in each line of the selected texts
	// the results then gett output as a big "browser table"...

	const (
		TBLRW = `
            <tr class="browser">
                <td class="browserembeddedannotations">%s</td>
                <td class="browsedline">%s</td>
                <td class="browsercite">%s</td>
            </tr>
		`
		SUMM = `
		<div id="searchsummary">%s,&nbsp;<span class="foundwork">%s</span><br>
		citation format:&nbsp;%s<br></div>`

		SNIP   = `✃✃✃`
		HITCAP = `<span class="small"><span class="red emph">text generation incomplete:</span> hit the cap of %d on allowed lines</span>`
	)

	type JSFeeder struct {
		SU string `json:"searchsummary"`
		HT string `json:"thehtml"`
		JS string `json:"newjs"`
	}

	user := vlt.ReadUUIDCookie(c)
	if !vlt.AllAuthorized.Check(user) {
		return c.JSONPretty(http.StatusOK, JSFeeder{JS: vv.JSVALIDATION}, vv.JSONINDENT)
	}

	sess := vlt.AllSessions.GetSess(user)
	srch := search.SessionIntoBulkSearch(c, vv.MAXTEXTLINEGENERATION)

	if srch.Results.Len() == 0 {
		return emptyjsreturn(c)
	}

	// now we have the lines we need....
	firstline := srch.Results.FirstLine()
	firstwork := search.DbWlnMyWk(&firstline)
	firstauth := search.DbWlnMyAu(&firstline)

	lines := srch.Results.YieldAll()
	block := make([]string, srch.Results.Len())

	i := 0
	for l := range lines {
		l.PurgeMetadata()
		block[i] = l.MarkedUp
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
	lines = srch.Results.YieldAll()
	for l := range lines {
		cit := selectivelydisplaycitations(l, previous, -1)
		trr[i] = fmt.Sprintf(TBLRW, l.Annotations, l.MarkedUp, cit)
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
	top += `<table><tbody>`
	top += `<tr class="spacing">` + strings.Repeat("&nbsp;", vv.MINBROWSERWIDTH) + `</tr>`

	tab = top + tab + `</tbody></table>`

	// but we don't want/need "observed" tags

	// <div id="searchsummary">Cicero,&nbsp;<span class="foundwork">Philippicae</span><br><br>citation format:&nbsp;oration 3, section 13, line 1<br></div>

	sui := sess.Inclusions

	au := firstauth.Shortname
	if len(sui.Authors) > 1 || len(sui.AuGenres) > 0 || len(sui.AuLocations) > 0 {
		au += " (and others)"
	}

	ti := firstwork.Title
	if len(sui.Works) > 1 || len(sui.WkGenres) > 0 || len(sui.WkLocations) > 0 {
		ti += " (and others)"
	}

	ct := basiccitation(firstline)

	sum := fmt.Sprintf(SUMM, au, ti, ct)

	cp := ""
	if srch.Results.Len() == vv.MAXTEXTLINEGENERATION {
		m := message.NewPrinter(language.English)
		cp = m.Sprintf(HITCAP, vv.MAXTEXTLINEGENERATION)
	}
	sum = sum + cp

	if lnch.Config.ZapLunates {
		tab = gen.DeLunate(tab)
	}

	var jso JSFeeder
	jso.SU = sum
	jso.HT = tab
	jso.JS = ""

	vlt.WSInfo.Del <- srch.WSID

	return gen.JSONresponse(c, jso)
}
