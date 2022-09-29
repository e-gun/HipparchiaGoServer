//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

//
// ROUTING
//

// RtBrowseLocus - open a browser if sent '/browse/locus/gr0086/025/999a|_0'
func RtBrowseLocus(c echo.Context) error {
	// sample input: http://localhost:8000/browse/locus/gr0086/025/999a|_0
	sep := "|"
	bp := Browse(c, sep)
	return c.JSONPretty(http.StatusOK, bp, JSONINDENT)
}

func RtBrowsePerseus(c echo.Context) error {
	// sample input: http://localhost:8000//browse/perseus/lt0550/001/2:717
	sep := ":"
	bp := Browse(c, sep)
	return c.JSONPretty(http.StatusOK, bp, JSONINDENT)
}

func RtBrowseRaw(c echo.Context) error {
	// uri: /browse/rawlocus/lt0474/055/1.1.1
	sep := "."
	bp := Browse(c, sep)
	return c.JSONPretty(http.StatusOK, bp, JSONINDENT)
}

// RtBrowseline - open a browser if sent '/browse/linenumber/lt0550/001/1855'
func RtBrowseline(c echo.Context) error {
	// sample input: '/browse/linenumber/lt0550/001/1855'
	// the one route that calls generatebrowsedpassage() directly

	user := readUUIDCookie(c)

	locus := c.Param("locus")
	elem := strings.Split(locus, "/")
	if len(elem) == 3 {
		au := elem[0]
		wk := elem[1]
		ln, e := strconv.Atoi(elem[2])
		chke(e)
		ctx := sessions[user].BrowseCtx
		bp := generatebrowsedpassage(au, wk, int64(ln), ctx)
		return c.JSONPretty(http.StatusOK, bp, JSONINDENT)
	} else {
		msg(fmt.Sprintf("RtBrowseline() could not parse %s", locus), 3)
		return c.JSONPretty(http.StatusOK, "", JSONINDENT)
	}
}

//
// BROWSING
//

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

// Browse - parse request and send a request to generatebrowsedpassage
func Browse(c echo.Context, sep string) BrowsedPassage {
	// sample input: http://localhost:8000//browse/perseus/lt0550/001/2:717
	user := readUUIDCookie(c)

	locus := c.Param("locus")
	elem := strings.Split(locus, "/")
	if len(elem) == 3 {
		au := elem[0]
		wk := elem[1]
		uid := au + "w" + wk

		// findendpointsfromlocus() lives in rt-selection.go
		ln := findendpointsfromlocus(uid, elem[2], sep)
		ctx := sessions[user].BrowseCtx
		return generatebrowsedpassage(au, wk, ln[0], ctx)
	} else {
		msg(fmt.Sprintf("Browse() could not parse %s", locus), 3)
		return BrowsedPassage{}
	}
}

// generatebrowsedpassage - browse Author A at line X with a context of Y lines
func generatebrowsedpassage(au string, wk string, fc int64, ctx int64) BrowsedPassage {
	// build a response to "GET /browse/linenumber/gr0062/028/14672 HTTP/1.1"

	k := fmt.Sprintf("%sw%s", au, wk)

	// [a] validate
	w := DbWork{}
	w.UID = "null"
	if _, ok := AllWorks[k]; ok {
		w = AllWorks[k]
	} else {
		if _, y := AllAuthors[au]; y {
			// firstwork; otherwise we are still set to "null"
			w = AllWorks[AllAuthors[au].WorkList[0]]
		}
	}

	if w.UID == "null" {
		msg(fmt.Sprintf("could not find a work for %s", k), 1)
		return BrowsedPassage{}
	}

	// [b] acquire the lines we need to display in the body

	lines := simplecontextgrabber(au, fc, ctx/2)

	// [b1] drop lines that are part of another work (matters in DP, IN, and CH)
	var trimmed []DbWorkline
	for _, l := range lines {
		if l.WkUID == w.UID {
			trimmed = append(trimmed, l)
		}
	}

	lines = trimmed

	if len(lines) == 0 {
		msg(fmt.Sprintf("generatebrowsedpassage() called simplecontextgrabber() and failed: %s, %d, %d", au, fc, ctx/2), 1)
		return BrowsedPassage{}
	}

	// want to do what follows in some sort of regular order
	nk := []string{"#", "", "loc", "pub", "c:", "r:", "d:"}

	for i, _ := range lines {
		lines[i].GatherMetadata()
		if len(lines[i].EmbNotes) != 0 {
			nt := `%s %s<br>`
			lines[i].Annotations = ""
			for _, key := range nk {
				if v, y := lines[i].EmbNotes[key]; y {
					lines[i].Annotations += fmt.Sprintf(nt, key, v)
				}
			}
		}
	}

	// [c] acquire and format the HTML

	ci := formatbrowsercitationinfo(w, lines[0], lines[len(lines)-1])
	tr := buildbrowsertable(fc, lines)

	// [d] fill out the JSON-ready struct
	p := fc - ctx
	if p < AllWorks[k].FirstLine {
		p = AllWorks[k].FirstLine
	}

	n := fc + ctx
	if n > AllWorks[k].LastLine {
		n = AllWorks[k].LastLine
	}

	bw := fmt.Sprintf(`linenumber/%s/%s/%d`, au, wk, p)
	fw := fmt.Sprintf(`linenumber/%s/%s/%d`, au, wk, n)
	ab := fmt.Sprintf(`%s [%s]`, AllAuthors[au].Cleaname, au)
	wb := fmt.Sprintf(`%s (w%s)`, w.Title, w.FindWorknumber())

	var bp BrowsedPassage
	bp.Browseforwards = fw
	bp.Browseback = bw
	bp.Authornumber = au
	bp.Workid = lines[0].WkUID
	bp.Authorboxcontents = ab
	bp.Workboxcontents = wb
	bp.Browserhtml = ci + tr
	bp.Worknumber = wk

	return bp
}

//
// HELPERS
//

// formatpublicationinfo - does just what you think it does
func formatpublicationinfo(w DbWork) string {
	// 	in:
	//		<volumename>FHG </volumename>4 <press>Didot </press><city>Paris </city><year>1841–1870</year><pages>371 </pages><pagesintocitations>Frr. 1–2</pagesintocitations><editor>Müller, K. </editor>
	//	out:
	//		<span class="pubvolumename">FHG <br /></span><span class="pubpress">Didot , </span><span class="pubcity">Paris , </span><span class="pubyear">1841–1870. </span><span class="pubeditor"> (Müller, K. )</span>

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
		tag := fmt.Sprintf("<%s>(?P<data>.*?)</%s>", t.Name, t.Name)
		pattern := regexp.MustCompile(tag)
		found := pattern.MatchString(w.Pub)
		if found {
			subs := pattern.FindStringSubmatch(w.Pub)
			data := subs[pattern.SubexpIndex("data")]
			pub := fmt.Sprintf(`<%d>%s%s%s</%d>`, t.Sub, t.Left, data, t.Right, t.Sub)
			pubinfo += pub
		}
	}

	pubinfo = avoidlonglines(pubinfo, MINBROWSERWIDTH+(MINBROWSERWIDTH/2))

	// restore the strings
	var reconstituted string
	for _, t := range tags {
		tag := fmt.Sprintf("<%d>(?P<data>.*?)</%d>", t.Sub, t.Sub)
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
func formatbrowsercitationinfo(w DbWork, f DbWorkline, l DbWorkline) string {
	cv := `
		<p class="currentlyviewing">
		%s<br>
		<span class="currentlyviewingcitation">%s — %s</span>
		%s
		%s</p>`

	ct := `<cvauthor">%s</span>, <cvwork">%s</span>`

	au := AllAuthors[w.FindAuthor()].Name
	ti := w.Title

	ci := fmt.Sprintf(ct, au, ti)
	ci = avoidlonglines(ci, MINBROWSERWIDTH)
	ci = strings.Replace(ci, "<cv", `<span class="currentlyviewing`, -1)

	dt := `<br>(Assigned date of %s)`
	beg := basiccitation(w, f)
	end := basiccitation(w, l)

	pi := formatpublicationinfo(AllWorks[f.WkUID])
	id := formatinscriptiondates(dt, f)

	cv = fmt.Sprintf(cv, ci, beg, end, pi, id)

	return cv
}

// basiccitation - produce a comma-separated citation from a DbWorkline
func basiccitation(w DbWork, l DbWorkline) string {
	cf := w.CitationFormat()
	loc := l.FindLocus()
	cf = cf[6-(len(loc)) : 6]

	var cit []string
	for i, _ := range loc {
		cit = append(cit, fmt.Sprintf("%s %s", cf[i], loc[i]))
	}
	fullcit := strings.Join(cit, ", ")
	return fullcit
}

// buildbrowsertable - where the actual HTML gets generated
func buildbrowsertable(focus int64, lines []DbWorkline) string {
	tr := `
            <tr class="browser">
                <td class="browserembeddedannotations">%s</td>
                <td class="browsedline">%s</td>
                <td class="browsercite">%s</td>
            </tr>
		`
	fla := `<span class="focusline">`
	flb := `</span>`

	block := make([]string, len(lines))
	for i, l := range lines {
		block[i] = l.MarkedUp
	}

	whole := strings.Join(block, "✃✃✃")

	whole = textblockcleaner(whole)

	// reassemble
	block = strings.Split(whole, "✃✃✃")
	for i, b := range block {
		lines[i].MarkedUp = b
	}

	trr := make([]string, len(lines))
	previous := lines[0]

	// complication: hyphenated words at the end of a line
	// this will already have markup from bracketformatting and so have to be handled carefully

	terminalhyph := regexp.MustCompile("(\\S+-)$")

	for i, _ := range lines {
		// turn "abc def" into "<observed id="abc">abc</observed> <observed id="def">def</observed>"
		// the complication is that x.MarkedUp contains html; use x.Accented to find the words

		// further complications: hyphenated words & capitalized words

		wds := strings.Split(lines[i].Accented, " ")
		lastwordindex := len(wds) - 1
		lwd := wds[lastwordindex] // preserve this before potentially shrinking wds
		wds = unique(wds)

		newline := lines[i].MarkedUp
		mw := strings.Split(lines[i].MarkedUp, " ")
		lmw := mw[len(mw)-1]
		for w, _ := range wds {
			cv := capsvariants(wds[w])
			if w == len(wds)-1 && terminalhyph.MatchString(lmw) {
				cv = capsvariants(lmw)
			}
			pattern, e := regexp.Compile(fmt.Sprintf("(^|\\s|\\[|\\>|⟨|‘|;)(%s)(\\s|\\.|\\]|\\<|⟩|’|\\!|,|:|;|\\?|·|$)", cv))
			if e == nil && w == len(wds)-1 && terminalhyph.MatchString(lmw) {
				// wds[lastwordindex] is the unhyphenated word
				r := fmt.Sprintf(`$1<observed id="%s">$2</observed>$3`, lwd)
				newline = pattern.ReplaceAllString(newline, r)
			} else if e == nil {
				newline = pattern.ReplaceAllString(newline, `$1<observed id="$2">$2</observed>$3`)
			} else {
				// you will barf if wds[w] = *
				msg(fmt.Sprintf("buildbrowsertable() could not regex compile %s", wds[w]), 4)
			}

			// complication: elision: <observed id="ἀλλ">ἀλλ</observed>’
			// but you can't deal with that here: the ’ will not turn up a find in the dictionary; the ' will yield bad SQL
			// so the dictionary lookup has to be reworked
		}

		var bl string
		if lines[i].TbIndex != focus {
			bl = newline
		} else {
			bl = fmt.Sprintf("%s%s%s", fla, newline, flb)
		}

		cit := selectivelydisplaycitations(lines[i], previous, focus)

		an := lines[i].Annotations
		if cfg.DbDebug {
			an = fmt.Sprintf("%s: %d", lines[i].FindAuthor(), lines[i].TbIndex)
			// bl = fmt.Sprintf(`<span class="small">%s</span>`, lines[i].ShowMarkup())
		}

		trr[i] = fmt.Sprintf(tr, an, bl, cit)
		previous = lines[i]
	}
	tab := strings.Join(trr, "")

	// that was the body, now do the head and tail
	top := fmt.Sprintf(`<div id="browsertableuid" uid="%s"></div>`, lines[0].FindAuthor())
	top += `<table><tbody>`
	top += `<tr class="spacing">` + strings.Repeat("&nbsp;", MINBROWSERWIDTH) + `</tr>`

	tab = top + tab + `</tbody></table>`

	return tab
}

// selectivelydisplaycitations - only show line numbers every N lines, etc.
func selectivelydisplaycitations(theline DbWorkline, previous DbWorkline, focus int64) string {
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

	if !theline.SameLevelAs(previous) || z%SHOWCITATIONEVERYNLINES == 0 || theline.TbIndex == focus {
		// display citation
	} else {
		citation = ""
	}
	return citation
}

func avoidlonglines(untrimmed string, maxlen int) string {
	if len(untrimmed) > maxlen {
		untrimmed = strings.Replace(untrimmed, ";", "; ", -1)
		pi := strings.Split(untrimmed, " ")
		var trimmed string
		breaks := 0
		reset := 0
		crop := maxlen
		for i := 0; i < len(pi); i++ {
			trimmed += pi[i] + " "
			if len(trimmed) > reset+crop {
				trimmed += "<br>"
				breaks += 1
				reset = len(trimmed)
			}
		}
		untrimmed = trimmed
	}
	return untrimmed
}
