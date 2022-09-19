//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"encoding/json"
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
	return c.String(http.StatusOK, bp)
}

func RtBrowsePerseus(c echo.Context) error {
	// sample input: http://localhost:8000//browse/perseus/lt0550/001/2:717
	sep := ":"
	bp := Browse(c, sep)
	return c.String(http.StatusOK, bp)
}

func RtBrowseRaw(c echo.Context) error {
	// uri: /browse/rawlocus/lt0474/055/1.1.1
	sep := "."
	bp := Browse(c, sep)
	return c.String(http.StatusOK, bp)
}

// RtBrowseline - open a browser if sent '/browse/linenumber/lt0550/001/1855'
func RtBrowseline(c echo.Context) error {
	// sample input: '/browse/linenumber/lt0550/001/1855'
	// the one route that calls HipparchiaBrowser() directly

	user := readUUIDCookie(c)

	locus := c.Param("locus")
	elem := strings.Split(locus, "/")
	if len(elem) == 3 {
		au := elem[0]
		wk := elem[1]
		ln, e := strconv.Atoi(elem[2])
		chke(e)
		ctx := sessions[user].UI.BrowseCtx
		js := HipparchiaBrowser(au, wk, int64(ln), ctx)
		return c.String(http.StatusOK, string(js))
	} else {
		msg(fmt.Sprintf("RtBrowseline() could not parse %s", locus), 3)
		return c.String(http.StatusOK, "")
	}
}

//
// BROWSING
//

func Browse(c echo.Context, sep string) string {
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
		ctx := sessions[user].UI.BrowseCtx
		js := HipparchiaBrowser(au, wk, ln[0], ctx)
		return string(js)
	} else {
		msg(fmt.Sprintf("Browse() could not parse %s", locus), 3)
		return ""
	}
}

// HipparchiaBrowser - browse Author A at line X with a context of Y lines
func HipparchiaBrowser(au string, wk string, fc int64, ctx int64) []byte {
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
		return []byte{}
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

	for i, _ := range lines {
		lines[i].GatherMetadata()
		if len(lines[i].EmbNotes) != 0 {
			nt := `<span class="red">%s:</span> %s<br>`
			lines[i].Annotations = ""
			for key, v := range lines[i].EmbNotes {
				lines[i].Annotations += fmt.Sprintf(nt, key, v)
			}
		}
	}

	// [c] acquire and format the HTML

	// need to set lines[0] to the focus, ie the middle of the pile of lines
	ci := formatcitationinfo(w, lines[0])
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

	// a JSON output struct
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

	var bp BrowsedPassage
	bp.Browseforwards = fw
	bp.Browseback = bw
	bp.Authornumber = au
	bp.Workid = lines[0].WkUID
	bp.Authorboxcontents = ab
	bp.Workboxcontents = wb
	bp.Browserhtml = ci + tr
	bp.Worknumber = wk

	// debugging
	// fmt.Println(ci)

	js, e := json.Marshal(bp)
	chke(e)

	return js
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

	// do the splits
	if len(pubinfo) > MINBROWSERWIDTH+(MINBROWSERWIDTH/2) {
		pubinfo = strings.Replace(pubinfo, ";", "; ", -1)
		pi := strings.Split(pubinfo, " ")
		var trimmed string
		breaks := 0
		reset := 0
		crop := MINBROWSERWIDTH + (MINBROWSERWIDTH / 2)
		for i := 0; i < len(pi); i++ {
			trimmed += pi[i] + " "
			if len(trimmed) > reset+crop {
				trimmed += "<br>"
				breaks += 1
				reset = len(trimmed)
			}
		}
		pubinfo = trimmed
	}

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

// formatcitationinfo - the prolix bibliographic info for a line/work
func formatcitationinfo(w DbWork, l DbWorkline) string {
	cv := `
		<p class="currentlyviewing">
		<span class="currentlyviewingauthor">%s</span>, 
		<span class="currentlyviewingwork">%s</span><br />
		<span class="currentlyviewingcitation">%s</span>
		%s
		%s</p>`

	dt := `<br>(Assigned date of %s)`

	au := AllAuthors[w.FindAuthor()].Name
	ti := w.Title
	fc := basiccitation(w, l)
	pi := formatpublicationinfo(AllWorks[l.WkUID])
	id := formatinscriptiondates(dt, l)
	cv = fmt.Sprintf(cv, au, ti, fc, pi, id)

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

	terminalhyph := regexp.MustCompile("\\s([^\\s]+-)$")

	for i, _ := range lines {
		// turn "abc def" into "<observed id="abc">abc</observed> <observed id="def">def</observed>"
		// the complication is that x.MarkedUp contains html; use x.Accented to find the words

		// further complications: hyphenated words & capitalized words

		wds := strings.Split(lines[i].Accented, " ")
		wds = unique(wds)

		newline := lines[i].MarkedUp
		lastword := len(wds) - 1
		for w, _ := range wds {
			cv := capsvariants(wds[w])
			pattern, e := regexp.Compile(fmt.Sprintf("(^|\\s|\\[|\\>|⟨|‘|;)(%s)(\\s|\\.|\\]|\\<|⟩|’|\\!|,|;|\\?|·|$)", cv))
			if e == nil && w != lastword {
				// you will barf if wds[w] = *
				newline = pattern.ReplaceAllString(newline, `$1<observed id="$2">$2</observed>$3`)
			} else if e == nil && w == lastword {
				if terminalhyph.MatchString(wds[w]) {
					r := fmt.Sprintf(` <observed id="%s">$1</observed>`, wds[len(wds)-1])
					newline = terminalhyph.ReplaceAllString(newline, r)
				} else {
					newline = pattern.ReplaceAllString(newline, `$1<observed id="$2">$2</observed>$3`)
				}
			} else {
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

		trr[i] = fmt.Sprintf(tr, lines[i].Annotations, bl, cit)
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
