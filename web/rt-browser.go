//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/format"
	"github.com/e-gun/HipparchiaGoServer/internal/mps"
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/labstack/echo/v5"
	"slices"
	"strconv"
	"strings"
)

// browsedpassage - a JSON output struct
type browsedpassage struct {
	Browseforwards    string `json:"browseforwards"`
	Browseback        string `json:"browseback"`
	Authornumber      string `json:"authornumber"`
	Workid            string `json:"workid"`
	Worknumber        string `json:"worknumber"`
	Authorboxcontents string `json:"authorboxcontents"`
	Workboxcontents   string `json:"workboxcontents"`
	Browserhtml       string `json:"browserhtml"`
	Newtitle          string `json:"newtitle"`
}

//
// ROUTING
//

// RtBrowseLocus - open a browser if sent '/browse/locus/gr0086/025/999a|_0'
func RtBrowseLocus(c *echo.Context) error {
	sep := "|"
	bp := browse(c, sep)
	return jsonresponse(c, bp)
}

// RtBrowsePerseus - open a browser if sent '/browse/perseus/lt0550/001/2:717'
func RtBrowsePerseus(c *echo.Context) error {
	sep := ":"
	bp := browse(c, sep)
	return jsonresponse(c, bp)
}

// RtBrowseRaw - open a browser if sent '/browse/rawlocus/lt0474/055/1.1.1'
func RtBrowseRaw(c *echo.Context) error {
	sep := "."
	bp := browse(c, sep)
	return jsonresponse(c, bp)
}

// RtBrowseLine - open a browser if sent '/browse/index/lt0550/001/1855'
func RtBrowseLine(c *echo.Context) error {
	// sample input: '/browse/index/lt0550/001/1855'
	// the one route that calls generatebrowsedpassage() directly
	// c.Response().After(func() { vlt.LogPaths("RtBrowseLine()") })

	const (
		FAIL = "RtBrowseLine() could not parse %s"
	)

	user := vlt.ReadUUIDCookie(c)
	if !vlt.AllAuthorized.Check(user) {
		bp := browsedpassage{Browserhtml: vv.AUTHWARN}
		return jsonresponse(c, bp)
	}

	s := vlt.AllSessions.GetSess(user)
	regularizewidth := true
	if slices.Contains(vv.MonspaceFonts, s.FontSel) {
		regularizewidth = false
	}

	locus := c.Param("locus")
	elem := strings.Split(locus, "/")
	if len(elem) == 3 {
		au := elem[0]
		wk := elem[1]
		ln, e := strconv.Atoi(elem[2])
		if e != nil {
			Msg.NOTE("Error in RtBrowseLine() - " + e.Error())
			return emptyjsreturn(c)
		}
		ctx := s.BrowseCtx
		bp := generatebrowsedpassage(au, wk, ln, ctx, s.ZapLunates, regularizewidth)
		return jsonresponse(c, bp)
	} else {
		Msg.FYI(fmt.Sprintf(FAIL, locus))
		return emptyjsreturn(c)
	}
}

// RtEmptyBrowse - to stave off 404s
func RtEmptyBrowse(c *echo.Context) error {
	bp := browsedpassage{}
	return jsonresponse(c, bp)
}

//
// BROWSING
//

// browse - parse request and send a request to generatebrowsedpassage
func browse(c *echo.Context, sep string) browsedpassage {
	// sample input: http://localhost:8000//browse/perseus/lt0550/001/2:717
	const (
		FAIL  = "browse() could not parse %s"
		FIRST = "_firstwork"
	)

	user := vlt.ReadUUIDCookie(c)
	s := vlt.AllSessions.GetSess(user)

	if !vlt.AllAuthorized.Check(user) {
		return browsedpassage{Browserhtml: vv.AUTHWARN}
	}

	locus := c.Param("locus")
	elem := strings.Split(locus, "/")
	if len(elem) == 3 {
		au := elem[0]
		wk := elem[1]

		if wk == FIRST {
			wk = mps.AllWorks[mps.AllAuthors[au].WorkList[0]].WkID()
		}
		uid := au + "w" + wk

		// findendpointsfromlocus() lives in rt-selection.go
		ln := findendpointsfromlocus(uid, elem[2], sep)
		ctx := s.BrowseCtx

		regularizewidth := true
		if s.FontSel == "Iosevka" {
			regularizewidth = false
		}
		return generatebrowsedpassage(au, wk, ln[0], ctx, s.ZapLunates, regularizewidth)
	} else {
		Msg.FYI(fmt.Sprintf(FAIL, locus))
		return browsedpassage{}
	}
}

// generatebrowsedpassage - browse Author A at line X with a context of Y lines
func generatebrowsedpassage(au string, wk string, fc int, ctx int, zaplunates bool, regularizewidth bool) browsedpassage {
	// build a response to "GET /browse/index/gr0062/028/14672 HTTP/1.1"

	const (
		FAIL1 = "could not find a work for %s"
		FAIL2 = "<br>Called SimpleContextGrabber() and failed.<br><br><code>No data for %sw%s where idx=%d</code><br>"
	)

	k := fmt.Sprintf("%sw%s", au, wk)

	// [a] validate
	w := validateworkselection(k)

	if w.UID == "work_not_found" {
		// some problem cases (that arise via rt-lexica.go and the bad clicks embedded in the lexical data):
		// gr0161w001
		Msg.FYI(fmt.Sprintf(FAIL1, k))
		return browsedpassage{}
	}

	// [b] acquire the wlb we need to display in the body

	wlb := db.SimpleContextGrabber(au, fc, ctx/2)

	// [b1] drop wlb that are part of another work (matters in DP, IN, and CH)
	var trimmed []str.DbWorkline

	ll := wlb.Yield()
	for l := range ll {
		if l.WkUID == w.UID {
			trimmed = append(trimmed, l)
		}
	}

	wlb.Lines = trimmed

	if wlb.Len() == 0 {
		var bp browsedpassage
		bp.Browserhtml = fmt.Sprintf(FAIL2, au, wk, fc)
		return bp
	}

	// [c] acquire and format the HTML

	ci := format.ProlixBrowswerCitations(wlb.FirstLine(), wlb.Lines[wlb.Len()-1])
	tr := format.BuildBrowserTable(fc, wlb.Lines, zaplunates, regularizewidth)

	// [d] fill out the JSON-ready struct
	p := fc - ctx
	if p < w.FirstLine {
		p = w.FirstLine
	}

	n := fc + ctx
	if n > w.LastLine {
		n = w.LastLine
	}

	bw := fmt.Sprintf(`index/%s/%s/%d`, au, wk, p)
	fw := fmt.Sprintf(`index/%s/%s/%d`, au, wk, n)
	ab := fmt.Sprintf(`%s [%s]`, mps.AllAuthors[au].Cleaname, au)
	wb := fmt.Sprintf(`%s (w%s)`, w.Title, w.WkID())
	nt := fmt.Sprintf("%s, %s", mps.AllAuthors[au].Shortname, w.Title)

	bp := browsedpassage{
		Browseforwards:    fw,
		Browseback:        bw,
		Authornumber:      au,
		Workid:            wlb.FirstLine().WkUID,
		Worknumber:        wk,
		Authorboxcontents: ab,
		Workboxcontents:   wb,
		Browserhtml:       ci + tr,
		Newtitle:          nt,
	}

	return bp
}
