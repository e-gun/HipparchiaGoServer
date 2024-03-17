//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/generic"
	"github.com/e-gun/HipparchiaGoServer/internal/launch"
	"github.com/e-gun/HipparchiaGoServer/internal/search"
	"github.com/e-gun/HipparchiaGoServer/internal/structs"
	"github.com/e-gun/HipparchiaGoServer/internal/vaults"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/labstack/echo/v4"
	"regexp"
	"strconv"
	"strings"
)

// BrowsedPassage - a JSON output struct
type BrowsedPassage struct {
	Browseforwards    string `json:"browseforwards"`
	Browseback        string `json:"browseback"`
	Authornumber      string `json:"authornumber"`
	Workid            string `json:"workid"`
	Worknumber        string `json:"worknumber"`
	Authorboxcontents string `json:"authorboxcontents"`
	Workboxcontents   string `json:"workboxcontents"`
	Browserhtml       string `json:"browserhtml"`
}

//
// ROUTING
//

// RtBrowseLocus - open a browser if sent '/browse/locus/gr0086/025/999a|_0'
func RtBrowseLocus(c echo.Context) error {
	sep := "|"
	bp := Browse(c, sep)
	return generic.JSONresponse(c, bp)
}

// RtBrowsePerseus - open a browser if sent '/browse/perseus/lt0550/001/2:717'
func RtBrowsePerseus(c echo.Context) error {
	sep := ":"
	bp := Browse(c, sep)
	return generic.JSONresponse(c, bp)
}

// RtBrowseRaw - open a browser if sent '/browse/rawlocus/lt0474/055/1.1.1'
func RtBrowseRaw(c echo.Context) error {
	sep := "."
	bp := Browse(c, sep)
	return generic.JSONresponse(c, bp)
}

// RtBrowseLine - open a browser if sent '/browse/index/lt0550/001/1855'
func RtBrowseLine(c echo.Context) error {
	// sample input: '/browse/index/lt0550/001/1855'
	// the one route that calls GenerateBrowsedPassage() directly
	c.Response().After(func() { msg.LogPaths("RtBrowseLine()") })

	const (
		FAIL = "RtBrowseLine() could not parse %s"
	)

	user := ReadUUIDCookie(c)
	if !vaults.AllAuthorized.Check(user) {
		bp := BrowsedPassage{Browserhtml: vv.AUTHWARN}
		return generic.JSONresponse(c, bp)
	}

	s := vaults.AllSessions.GetSess(user)
	locus := c.Param("locus")
	elem := strings.Split(locus, "/")
	if len(elem) == 3 {
		au := elem[0]
		wk := elem[1]
		ln, e := strconv.Atoi(elem[2])
		msg.EC(e)
		ctx := s.BrowseCtx
		bp := GenerateBrowsedPassage(au, wk, ln, ctx)
		return generic.JSONresponse(c, bp)
	} else {
		msg.FYI(fmt.Sprintf(FAIL, locus))
		return emptyjsreturn(c)
	}
}

//
// BROWSING
//

// Browse - parse request and send a request to GenerateBrowsedPassage
func Browse(c echo.Context, sep string) BrowsedPassage {
	// sample input: http://localhost:8000//browse/perseus/lt0550/001/2:717
	const (
		FAIL  = "Browse() could not parse %s"
		FIRST = "_firstwork"
	)

	user := ReadUUIDCookie(c)
	s := vaults.AllSessions.GetSess(user)

	if !vaults.AllAuthorized.Check(user) {
		return BrowsedPassage{Browserhtml: vv.AUTHWARN}
	}

	locus := c.Param("locus")
	elem := strings.Split(locus, "/")
	if len(elem) == 3 {
		au := elem[0]
		wk := elem[1]

		if wk == FIRST {
			wk = vv.AllWorks[vv.AllAuthors[au].WorkList[0]].WkID()
		}
		uid := au + "w" + wk

		// findendpointsfromlocus() lives in rt-selection.go
		ln := findendpointsfromlocus(uid, elem[2], sep)
		ctx := s.BrowseCtx

		return GenerateBrowsedPassage(au, wk, ln[0], ctx)
	} else {
		msg.FYI(fmt.Sprintf(FAIL, locus))
		return BrowsedPassage{}
	}
}

// GenerateBrowsedPassage - browse Author A at line X with a context of Y lines
func GenerateBrowsedPassage(au string, wk string, fc int, ctx int) BrowsedPassage {
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
		msg.FYI(fmt.Sprintf(FAIL1, k))
		return BrowsedPassage{}
	}

	// [b] acquire the wlb we need to display in the body

	wlb := search.SimpleContextGrabber(au, fc, ctx/2)

	// [b1] drop wlb that are part of another work (matters in DP, IN, and CH)
	var trimmed []structs.DbWorkline

	ll := wlb.YieldAll()
	for l := range ll {
		if l.WkUID == w.UID {
			trimmed = append(trimmed, l)
		}
	}

	wlb.Lines = trimmed

	if wlb.Len() == 0 {
		var bp BrowsedPassage
		bp.Browserhtml = fmt.Sprintf(FAIL2, au, wk, fc)
		return bp
	}

	// want to do what follows in some sort of regular order
	nk := []string{"#", "", "loc", "pub", "c:", "r:", "d:"}

	ll = wlb.YieldAll()
	for l := range ll {
		l.GatherMetadata()
		if len(l.GetNotes()) != 0 {
			nt := `%s %s<br>`
			l.Annotations = ""
			for _, key := range nk {
				if v, y := l.GetNotes()[key]; y {
					l.Annotations += fmt.Sprintf(nt, key, v)
				}
			}
		}
	}

	// [c] acquire and format the HTML

	ci := formatbrowsercitationinfo(wlb.FirstLine(), wlb.Lines[wlb.Len()-1])
	tr := buildbrowsertable(fc, wlb.Lines)

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
	ab := fmt.Sprintf(`%s [%s]`, vv.AllAuthors[au].Cleaname, au)
	wb := fmt.Sprintf(`%s (w%s)`, w.Title, w.WkID())

	bp := BrowsedPassage{
		Browseforwards:    fw,
		Browseback:        bw,
		Authornumber:      au,
		Workid:            wlb.FirstLine().WkUID,
		Worknumber:        wk,
		Authorboxcontents: ab,
		Workboxcontents:   wb,
		Browserhtml:       ci + tr,
	}

	return bp
}

//
// HELPERS
//

// formatpublicationinfo - does just what you think it does
func formatpublicationinfo(w structs.DbWork) string {
	// 	in:
	//		<volumename>FHG </volumename>4 <press>Didot </press><city>Paris </city><year>1841–1870</year><pages>371 </pages><pagesintocitations>Frr. 1–2</pagesintocitations><editor>Müller, K. </editor>
	//	out:
	//		<span class="pubvolumename">FHG <br /></span><span class="pubpress">Didot , </span><span class="pubcity">Paris , </span><span class="pubyear">1841–1870. </span><span class="pubeditor"> (Müller, K. )</span>

	const (
		REGS = "<%s>(?P<data>.*?)</%s>"
		REGD = "<%d>(?P<data>.*?)</%d>"
	)

	type Swapper struct {
		Name  string
		Sub   int
		Left  string
		Right string
	}

	tags := []Swapper{
		{"volumename", 1, "", " "},
		{"press", 2, " ", ", "},
		{"city", 3, " ", ", "},
		{"year", 4, " ", ". "},
		{"yearreprinted", 5, "[", "] "},
		{"series", 6, " ", ""},
		{"editor", 7, "(", ")"},
		{"work", 8, " ", " "},
		{"pages", 9, " pp. ", ". "},
	}

	pubinfo := ""

	// shorten the strings so you can split
	for _, t := range tags {
		tag := fmt.Sprintf(REGS, t.Name, t.Name)
		pattern := regexp.MustCompile(tag)
		found := pattern.MatchString(w.Pub)
		if found {
			subs := pattern.FindStringSubmatch(w.Pub)
			data := subs[pattern.SubexpIndex("data")]
			pub := fmt.Sprintf(`<%d>%s%s%s</%d>`, t.Sub, t.Left, data, t.Right, t.Sub)
			pubinfo += pub
		}
	}

	pubinfo = generic.AvoidLongLines(pubinfo, vv.MINBROWSERWIDTH+(vv.MINBROWSERWIDTH/2))

	// restore the strings
	var reconstituted string
	for _, t := range tags {
		tag := fmt.Sprintf(REGD, t.Sub, t.Sub)
		pattern := regexp.MustCompile(tag)
		found := pattern.MatchString(pubinfo)
		if found {
			subs := pattern.FindStringSubmatch(pubinfo)
			data := subs[pattern.SubexpIndex("data")]
			pub := fmt.Sprintf(`<span class="pub%s">%s</span>`, t.Name, data)
			reconstituted += pub
		}
	}

	readability := `<br>
	%s
	`
	return fmt.Sprintf(readability, reconstituted)
}

// formatbrowsercitationinfo - the prolix bibliographic info for a line/work
func formatbrowsercitationinfo(f structs.DbWorkline, l structs.DbWorkline) string {
	const (
		CV = `
		<p class="currentlyviewing">
		%s<br>
		<span class="currentlyviewingcitation">%s — %s</span>
		%s
		%s</p>`

		CT = `<cvauthor">%s</span>, <cvwork">%s</span>`
	)

	w := f.MyWk()

	au := w.MyAu().Name
	ti := w.Title

	ci := fmt.Sprintf(CT, au, ti)
	ci = generic.AvoidLongLines(ci, vv.MINBROWSERWIDTH)
	ci = strings.Replace(ci, "<cv", `<span class="currentlyviewing`, -1)

	dt := `<br>(Assigned date of %s)`
	beg := basiccitation(f)
	end := basiccitation(l)
	pi := formatpublicationinfo(*w)
	id := search.FormatInscriptionDates(dt, &f)

	cv := fmt.Sprintf(CV, ci, beg, end, pi, id)

	return cv
}

// basiccitation - produce a comma-separated citation from a DbWorkline: e.g., "book 5, chapter 37, section 5, line 3"
func basiccitation(l structs.DbWorkline) string {
	w := l.MyWk()
	cf := w.CitationFormat()
	loc := l.FindLocus()
	cf = cf[vv.NUMBEROFCITATIONLEVELS-(len(loc)) : vv.NUMBEROFCITATIONLEVELS]

	var cit []string
	for i := range loc {
		cit = append(cit, fmt.Sprintf("%s %s", cf[i], loc[i]))
	}
	fullcit := strings.Join(cit, ", ")
	return fullcit
}

// buildbrowsertable - where the actual HTML gets generated
func buildbrowsertable(focus int, lines []structs.DbWorkline) string {
	const (
		OBSREGTEMPL = "(^|\\s|\\[|\\>|⟨|‘|“|;)(%s)" + vv.TERMINATIONS
		UIDDIV      = `<div id="browsertableuid" uid="%s"></div>`
		TRTMPL      = `
            <tr class="browser">
                <td class="browserembeddedannotations">%s</td>
                <td class="browsedline">%s</td>
                <td class="browsercite">%s</td>
            </tr>
		`
		FOCA = `<span class="focusline">`
		FOCB = `</span>`
		SNIP = "✃✃✃"
		FAIL = "buildbrowsertable() could not regex compile %s"
	)

	block := make([]string, len(lines))
	for i, l := range lines {
		block[i] = l.MarkedUp
	}

	whole := strings.Join(block, SNIP)

	whole = search.TextBlockCleaner(whole)

	// reassemble
	block = strings.Split(whole, SNIP)
	for i, b := range block {
		lines[i].MarkedUp = b
	}

	trr := make([]string, len(lines))
	previous := lines[0]

	// complication: hyphenated words at the end of a line
	// this will already have markup from bracketformatting and so have to be handled carefully

	terminalhyph := regexp.MustCompile("(\\S+-)$")

	allwords := func() []string {
		wm := make(map[string]bool)
		for i := range lines {
			wds := strings.Split(lines[i].Accented, " ")
			for _, w := range wds {
				wm[w] = true
			}
		}
		return generic.StringMapKeysIntoSlice(wm)
	}()

	almostallregex := func() map[string]*regexp.Regexp {
		// you will have "ἱματίῳ", but the marked up line has "ἱμα- | τίῳ"
		ar := make(map[string]*regexp.Regexp)
		for _, w := range allwords {
			r := fmt.Sprintf(OBSREGTEMPL, generic.CapsVariants(w))
			pattern, e := regexp.Compile(r)
			if e != nil {
				// you will barf if w = *
				msg.PEEK(fmt.Sprintf(FAIL, w))
				pattern = regexp.MustCompile("FIND_NOTHING")
			}
			ar[w] = pattern
		}
		return ar
	}()

	for i := range lines {
		// turn "abc def" into "<observed id="abc">abc</observed> <observed id="def">def</observed>"
		// the complication is that x.MarkedUp contains html; use x.Accented to find the words

		// further complications: hyphenated words & capitalized words

		wds := strings.Split(lines[i].Accented, " ")
		lastwordindex := len(wds) - 1
		lwd := wds[lastwordindex] // preserve this before potentially shrinking wds
		wds = generic.Unique(wds)

		newline := lines[i].MarkedUp
		mw := strings.Split(lines[i].MarkedUp, " ")
		lmw := mw[len(mw)-1]

		for j := range wds {
			p := almostallregex[wds[j]]
			if j == len(wds)-1 && terminalhyph.MatchString(lmw) {
				// wds[lastwordindex] is the unhyphenated word
				// almostallregex does not contain this pattern: "ἱμα-", e.g.
				np, e := regexp.Compile(fmt.Sprintf(OBSREGTEMPL, generic.CapsVariants(lmw)))
				if e != nil {
					msg.PEEK(fmt.Sprintf(FAIL, lmw))
					np = regexp.MustCompile("FIND_NOTHING")
				}

				// without strings.Replace() gr2042@81454 browser formatting error: τὴν ἐκκληϲίαν, τὸν οἶκον τῆϲ class="expanded_text">προϲ-
				// the html ends up as: <span <observed="" id="προϲευχῆϲ">class="expanded_text"&gt;προϲ-</span>
				newline = strings.Replace(newline, "<span ", "<span_", -1)
				r := fmt.Sprintf(`$1<observed id="%s">$2</observed>$3`, lwd)
				newline = np.ReplaceAllString(newline, r)
				newline = strings.Replace(newline, "<span_", "<span ", -1)
			} else {
				newline = p.ReplaceAllString(newline, `$1<observed id="$2">$2</observed>$3`)
			}

			// complication: elision: <observed id="ἀλλ">ἀλλ</observed>’
			// but you can't deal with that here: the ’ will not turn up a find in the dictionary; the ' will yield bad SQL
			// so the dictionary lookup has to be reworked
		}

		var bl string
		if lines[i].TbIndex != focus {
			bl = newline
		} else {
			bl = fmt.Sprintf("%s%s%s", FOCA, newline, FOCB)
		}

		cit := selectivelydisplaycitations(lines[i], previous, focus)

		an := lines[i].Annotations
		if launch.Config.DbDebug {
			an = fmt.Sprintf("%s: %d", lines[i].AuID(), lines[i].TbIndex)
			// bl = fmt.Sprintf(`<span class="small">%s</span>`, lines[i].ShowMarkup())
		}

		trr[i] = fmt.Sprintf(TRTMPL, an, bl, cit)
		previous = lines[i]
	}
	tab := strings.Join(trr, "")

	// that was the body, now do the head and tail
	top := fmt.Sprintf(UIDDIV, lines[0].AuID())
	top += `<table><tbody>`
	// top += `<tr class="spacing">` + strings.Repeat("&nbsp;", MINBROWSERWIDTH) + `</tr>`

	tab = top + tab + `</tbody></table>`

	if launch.Config.ZapLunates {
		tab = generic.DeLunate(tab)
	}

	return tab
}

// selectivelydisplaycitations - only show line numbers every N lines, etc.
func selectivelydisplaycitations(theline structs.DbWorkline, previous structs.DbWorkline, focus int) string {
	// figure out whether to display a citation
	// pulled this out because it is common with the textbuilder (who will always send "0" as the focus)

	// [a] if thisline.samelevelas(previousline) is not True:...
	// [b] if linenumber % linesevery == 0
	// [c] always give a citation for the focus line
	citation := strings.Join(theline.FindLocus(), ".")

	z, e := strconv.Atoi(theline.Lvl0Value)
	if e != nil {
		z = 0
	}

	if !theline.SameLevelAs(previous) || z%vv.SHOWCITATIONEVERYNLINES == 0 || theline.TbIndex == focus {
		// display citation
	} else {
		citation = ""
	}
	return citation
}
