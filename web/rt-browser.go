//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/mps"
	"github.com/e-gun/HipparchiaGoServer/internal/search"
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/labstack/echo/v4"
	"regexp"
	"strconv"
	"strings"
)

var (
	// lt0474w057 138013 has no annotations registered in the database, but there are notes still in the line proper
	// <hmu_metadata_documentnumber value="445" /><hmu_metadata_date value="prid.(?) Id. Nov. 44" /> ... ⟨CICERO ATTICO SAL.⟩
	// so this and similar lines need patching; these precompiled values are used by adhocfixforbadannotations() below
	noteregex        = regexp.MustCompile("<hmu_metadata_([^\\s]*) value=\"([^\"]*)\" />")
	noteandlineregex = regexp.MustCompile("<hmu_metadata_.* value=.* />(.*)")
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
}

//
// ROUTING
//

// RtBrowseLocus - open a browser if sent '/browse/locus/gr0086/025/999a|_0'
func RtBrowseLocus(c echo.Context) error {
	sep := "|"
	bp := browse(c, sep)
	return jsonresponse(c, bp)
}

// RtBrowsePerseus - open a browser if sent '/browse/perseus/lt0550/001/2:717'
func RtBrowsePerseus(c echo.Context) error {
	sep := ":"
	bp := browse(c, sep)
	return jsonresponse(c, bp)
}

// RtBrowseRaw - open a browser if sent '/browse/rawlocus/lt0474/055/1.1.1'
func RtBrowseRaw(c echo.Context) error {
	sep := "."
	bp := browse(c, sep)
	return jsonresponse(c, bp)
}

// RtBrowseLine - open a browser if sent '/browse/index/lt0550/001/1855'
func RtBrowseLine(c echo.Context) error {
	// sample input: '/browse/index/lt0550/001/1855'
	// the one route that calls generatebrowsedpassage() directly
	c.Response().After(func() { vlt.LogPaths("RtBrowseLine()") })

	const (
		FAIL = "RtBrowseLine() could not parse %s"
	)

	user := vlt.ReadUUIDCookie(c)
	if !vlt.AllAuthorized.Check(user) {
		bp := browsedpassage{Browserhtml: vv.AUTHWARN}
		return jsonresponse(c, bp)
	}

	s := vlt.AllSessions.GetSess(user)
	locus := c.Param("locus")
	elem := strings.Split(locus, "/")
	if len(elem) == 3 {
		au := elem[0]
		wk := elem[1]
		ln, e := strconv.Atoi(elem[2])
		Msg.EC(e)
		ctx := s.BrowseCtx
		bp := generatebrowsedpassage(au, wk, ln, ctx, s.ZapLunates)
		return jsonresponse(c, bp)
	} else {
		Msg.FYI(fmt.Sprintf(FAIL, locus))
		return emptyjsreturn(c)
	}
}

// RtEmptyBrowse - to stave off 404s
func RtEmptyBrowse(c echo.Context) error {
	bp := browsedpassage{}
	return jsonresponse(c, bp)
}

//
// BROWSING
//

// browse - parse request and send a request to generatebrowsedpassage
func browse(c echo.Context, sep string) browsedpassage {
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

		return generatebrowsedpassage(au, wk, ln[0], ctx, s.ZapLunates)
	} else {
		Msg.FYI(fmt.Sprintf(FAIL, locus))
		return browsedpassage{}
	}
}

// generatebrowsedpassage - browse Author A at line X with a context of Y lines
func generatebrowsedpassage(au string, wk string, fc int, ctx int, zaplunates bool) browsedpassage {
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

	// want to do what follows in some sort of regular order
	nk := []string{"#", "", "loc", "pub", "c:", "r:", "d:"}

	ll = wlb.Yield()
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
	tr := buildbrowsertable(fc, wlb.Lines, zaplunates)

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

	bp := browsedpassage{
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
func formatpublicationinfo(w str.DbWork) string {
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

	pubinfo = gen.AvoidLongLines(pubinfo, vv.MINBROWSERWIDTH+(vv.MINBROWSERWIDTH/2))

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
func formatbrowsercitationinfo(f str.DbWorkline, l str.DbWorkline) string {
	const (
		CV = `
		<p class="currentlyviewing">
		%s<br>
		<span class="currentlyviewingcitation">%s — %s</span>
		%s
		%s</p>`

		CT = `<cvauthor">%s</span>, <cvwork">%s</span>`
	)

	w := search.DbWlnMyWk(&f)

	au := mps.DbWkMyAu(w).Name
	ti := w.Title

	ci := fmt.Sprintf(CT, au, ti)
	ci = gen.AvoidLongLines(ci, vv.MINBROWSERWIDTH)
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
func basiccitation(l str.DbWorkline) string {
	w := search.DbWlnMyWk(&l)
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

// buildbrowsertable - where the actual HTML gets generated; this table is better for cut-and-paste...
func buildbrowsertable(focus int, lines []str.DbWorkline, zaplunates bool) string {
	const (
		OBSREGTEMPL = "(^|\\s|\\[|\\>|⟨|‘|“|;)(%s)" + vv.TERMINATIONS
		UIDDIV      = `<div id="browsertableuid" uid="%s"></div>`
		TRTMPLNOTES = `
            <tr class="browser">
                <td class="browserembeddedannotations">%s</td>
                <td class="browsedline">%s</td>
                <td class="browsercite">%s</td>
            </tr>
		`
		TRTMPL = `
            <tr class="browser">
                <td class="browsedlinewithoutnotes">%s</td>
                <td class="browsercite">%s</td>
            </tr>
		`

		FOCA = `<span class="focusline">`
		FOCB = `</span>`
		SNIP = "✃✃✃"
		FAIL = "buildbrowsertable() could not regex compile %s"
	)

	// the builder has failed to parse some notes; do something about that
	longnotes := false // cap chars per line in notes
	lines = reannotatelines(lines, longnotes)

	block := make([]string, len(lines))
	longestline := 0
	longestnote := 0

	for i, l := range lines {
		block[i] = l.GetMarked()

		chars := len([]rune(l.Stripped)) + len(l.Annotations) + len(l.Citation())
		if chars > longestline {
			longestline = chars
		}

		if len(l.Annotations) > longestnote {
			longestnote = len(l.Annotations)
		}
	}

	whole := strings.Join(block, SNIP)

	whole = search.TextBlockCleaner(whole)

	// reassemble
	block = strings.Split(whole, SNIP)
	for i, b := range block {
		lines[i].MarkedUp = b
	}

	blines := make([]string, len(lines))
	bnotes := make([]string, len(lines))
	bcites := make([]string, len(lines))
	previous := lines[0]

	// complication: hyphenated words at the end of a line
	// this will already have markup from bracketformatting and so have to be handled carefully

	terminalhyph := regexp.MustCompile("(\\S+-)$")

	allwords := func() []string {
		wm := make(map[string]bool)
		for i := range lines {
			wds := strings.Split(lines[i].GetAccented(), " ")
			for _, w := range wds {
				wm[w] = true
			}
		}
		return gen.StringMapKeysIntoSlice(wm)
	}()

	almostallregex := func() map[string]*regexp.Regexp {
		// you will have "ἱματίῳ", but the marked up line has "ἱμα- | τίῳ"
		ar := make(map[string]*regexp.Regexp)
		for _, w := range allwords {
			r := fmt.Sprintf(OBSREGTEMPL, gen.CapsVariants(w))
			pattern, e := regexp.Compile(r)
			if e != nil {
				// you will barf if w = *
				Msg.PEEK(fmt.Sprintf(FAIL, w))
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
		wds := strings.Split(lines[i].GetAccented(), " ")
		lastwordindex := len(wds) - 1
		lwd := wds[lastwordindex] // preserve this before potentially shrinking wds
		wds = gen.Unique(wds)

		newline := lines[i].GetMarked()
		mw := strings.Split(lines[i].GetMarked(), " ")
		lmw := mw[len(mw)-1]

		for j := range wds {
			p := almostallregex[wds[j]]
			if j == len(wds)-1 && terminalhyph.MatchString(lmw) {
				// wds[lastwordindex] is the unhyphenated word
				// almostallregex does not contain this pattern: "ἱμα-", e.g.
				np, e := regexp.Compile(fmt.Sprintf(OBSREGTEMPL, gen.CapsVariants(lmw)))
				if e != nil {
					Msg.PEEK(fmt.Sprintf(FAIL, lmw))
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

		an := strings.Replace(lines[i].Annotations, "documentnumber: ", "#", 1)
		if lnch.Config.DbDebug {
			an = fmt.Sprintf("%s: %d", lines[i].AuID(), lines[i].TbIndex)
			// bl = fmt.Sprintf(`<span class="small">%s</span>`, lines[i].ShowMarkup())
		}

		blines[i] = bl
		bcites[i] = fmt.Sprintf("<span class=\"eighty\">%s</span>&nbsp;", cit) // the "normal" sized space is to maintain vertical alignment
		bnotes[i] = fmt.Sprintf("<span class=\"eighty\">%s</span>&nbsp;", an)
		previous = lines[i]
	}

	// we are building a table with one row and three columns; the pre v1.3.8 way was len(lines) rows
	// but if you want to cut and paste from the browser, that is not so good

	// note that font/style differences can/will throw these out of visual alignment unless something is done...
	ll := strings.Join(blines, "<br>\n")
	cc := strings.Join(bcites, "<br>\n")
	nn := strings.Join(bnotes, "<br>\n")

	var tab string
	if longestnote == 0 {
		tab = fmt.Sprintf(TRTMPL, ll, cc)
	} else {
		tab = fmt.Sprintf(TRTMPLNOTES, nn, ll, cc)
	}

	// that was the body, now do the head and tail
	top := fmt.Sprintf(UIDDIV, lines[0].AuID())
	top += `<table><tbody>`
	top += stabilizebrowserwidth(longestnote)
	tab = top + tab + `</tbody></table>`

	if zaplunates {
		tab = gen.DeLunate(tab)
		// overshot in DeLunate...: ς’ for σ’
		tab = strings.Replace(tab, "id=\"σ\">ς</observed>’ ", "id=\"σ\">σ</observed>’ ", -1)
	}

	return tab
}

// stabilizebrowserwidth - cut down on the bounce between widths in the browser
func stabilizebrowserwidth(longestline int) string {
	const (
		MINBROWSERWIDTH    = 130
		NOTESANDLOCUSFUDGE = 60 // because the table cells have pixel padding
		NOTESANDLOCUSPAD   = 15
	)

	//if 1 > 0 {
	//	return ""
	//}

	maxlen := longestline + NOTESANDLOCUSFUDGE

	var row string
	if maxlen < MINBROWSERWIDTH {
		row = `<tr class="spacing">` + strings.Repeat("·", MINBROWSERWIDTH) + `</tr>`
	} else {
		row = `<tr class="spacing">` + strings.Repeat("·", maxlen+NOTESANDLOCUSPAD) + `</tr>`
	}
	return row
}

// selectivelydisplaycitations - only show line numbers every N lines, etc.
func selectivelydisplaycitations(theline str.DbWorkline, previous str.DbWorkline, focus int) string {
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

// adhocfixforbadannotations - the current build of the data has errors in its hmu_metadata handling; patch them
func adhocfixforbadannotations(line string, longlines bool) (string, string) {
	// lt0474w057 138013 has no annotations in the database, but there are notes up front

	// have:
	// <hmu_metadata_documentnumber value="445" /><hmu_metadata_date value="prid.(?) Id. Nov. 44" / ... ⟨CICERO ATTICO SAL.⟩

	// want:
	// "445; prid.(?) Id. Nov. 44"

	// 	noteregex := regexp.MustCompile("<hmu_metadata_([^\\s]*) value=\"([^\"]*)\" />")
	//	noteandlineregex := regexp.MustCompile("<hmu_metadata_.* value=.* />(.*)")

	const (
		LONGLINE = 25
	)

	finds := noteregex.FindAllStringSubmatch(line, -1)
	if len(finds) == 0 {
		return line, ""
	}

	nl := noteandlineregex.FindAllStringSubmatch(line, -1)
	strippedline := nl[0][1]

	var hmu []string
	for _, find := range finds {
		// if you include "documentnumber" and "date" the lines can get very long...
		// hmu = append(hmu, fmt.Sprintf("%s: %s", find[1], find[2]))
		hmu = append(hmu, fmt.Sprintf("%s", find[2]))
	}

	metadata := strings.Join(hmu, "; ")

	// insert "<br>" to avoid hogging the screen; lots on the same sheet will ntl see the notes cluster around the right spot
	// not so hot for the textmaker, though
	if !longlines {
		metadata = gen.AvoidLongLines(metadata, LONGLINE)
	}

	return strippedline, metadata
}

// reannotatelines - bulk apply adhocfixforbadannotations()
func reannotatelines(lines []str.DbWorkline, longlines bool) []str.DbWorkline {
	// the builder has failed to parse some notes; do something about that
	for i, l := range lines {
		cln, en := adhocfixforbadannotations(l.MarkedUp, longlines)
		lines[i].MarkedUp = cln
		if isoktoshownotes(l) {
			// the original data also reset the notes at block ends so you can re-see them in the middle of a text
			// should really only display notes that go with "t" or "sa" lines, vel sim
			if lines[i].Annotations == "" {
				lines[i].Annotations = en
			} else {
				an := strings.ReplaceAll(lines[i].Annotations, "documentnumber: ", "")
				lines[i].Annotations += an + "; " + en
			}
		}
	}
	return lines
}

// isoktoshownotes - try to avoid re-showing notes that pop-up mid-text owing to original data block reset issue
func isoktoshownotes(l str.DbWorkline) bool {
	// the original data also reset the notes at block ends so you can re-see them in the middle of a text
	// should really only display notes that go with "t" or "sa" lines, vel sim
	w := mps.AllWorks[l.WkUID]

	// inscriptions, etc
	if l.TbIndex == w.FirstLine {
		return true
	}

	// letters of cicero (and who else?)
	if l.Lvl1Value == "sa" || l.Lvl0Value == "t" {
		return true
	}
	return false
}
