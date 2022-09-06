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

// RtBrowseline - open a browser if sent '/browse/linenumber/lt0550/001/1855'
func RtBrowseline(c echo.Context) error {
	// sample input: '/browse/linenumber/lt0550/001/1855'
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

	// [b] acquire the lines we need to display in the body

	lines := simplecontextgrabber(au, fc, ctx/2)
	k := fmt.Sprintf("%sw%s", au, wk)
	w := AllWorks[k]

	// [c] format the lines

	// not yet implemented
	// lines = paragraphformatting(lines)

	// [d] acquire and format the HTML

	// need to set lines[0] to the focus, ie the middle of the pile of lines
	ci := formatcitationinfo(AllAuthors, w, lines[0])
	pi := formatpublicationinfo(AllWorks[k])
	tr := buildbrowsertable(fc, lines)

	// [e] fill out the JSON-ready struct
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
		// marshal will not do lc names
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
	bp.Browserhtml = ci + pi + tr
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

func formatpublicationinfo(w DbWork) string {
	// 	in:
	//		<volumename>FHG </volumename>4 <press>Didot </press><city>Paris </city><year>1841–1870</year><pages>371 </pages><pagesintocitations>Frr. 1–2</pagesintocitations><editor>Müller, K. </editor>
	//	out:
	//		<span class="pubvolumename">FHG <br /></span><span class="pubpress">Didot , </span><span class="pubcity">Paris , </span><span class="pubyear">1841–1870. </span><span class="pubeditor"> (Müller, K. )</span>

	type Swapper struct {
		Name  string
		Left  string
		Right string
	}

	tags := []Swapper{
		{"volumename", "", ". "},
		{"press", "", ", "},
		{"city", "", ", "},
		{"year", "", ". "},
		{"yearreprinted", "[", "] "},
		{"series", "", ""},
		{"editor", "(", ")"},
		{"pages", " pp. ", ". "},
	}

	pubinfo := ""

	for _, t := range tags {
		tag := fmt.Sprintf("<%s>(?P<data>.*?)</%s>", t.Name, t.Name)
		pattern := regexp.MustCompile(tag)
		found := pattern.MatchString(w.Pub)
		if found {
			subs := pattern.FindStringSubmatch(w.Pub)
			data := strings.TrimRight(subs[pattern.SubexpIndex("data")], " ")
			pub := fmt.Sprintf(`<span class="pub%s">%s%s%s</span>`, t.Name, t.Left, data, t.Right)
			pubinfo += pub
		}
	}

	return pubinfo
}

func formatcitationinfo(authormap map[string]DbAuthor, w DbWork, l DbWorkline) string {
	cv := `
		<p class="currentlyviewing">
		<span class="currentlyviewingauthor">%s</span>, 
		<span class="currentlyviewingwork">%s</span><br />
		<span class="currentlyviewingcitation">%s</span>
		%s</p>`

	dt := `<br>(Assigned date of %s)`

	au := authormap[w.FindAuthor()].Name
	ti := w.Title
	fc := basiccitation(w, l)
	id := formatinscriptiondates(dt, l)
	cv = fmt.Sprintf(cv, au, ti, fc, id)

	return cv
}

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

	// no handling of 'lines every' yet
	// no handling of rollovers at new sections yet
	// no handling of 'issamework' (for papyri, etc) yet

	// try to fix some multi-line issues by building a text block...
	// across the whole
	var block []string
	for _, l := range lines {
		block = append(block, l.MarkedUp)
	}

	whole := strings.Join(block, "✃✃✃")
	whole = unbalancedspancleaner(whole)
	whole = formateditorialbrackets(whole)
	whole = formatmultilinebrackets(whole)
	// fmt.Println(whole)

	// reassemble
	block = strings.Split(whole, "✃✃✃")
	for i, b := range block {
		lines[i].MarkedUp = b
	}

	var trr []string
	for i, _ := range lines {
		cit := strings.Join(lines[i].FindLocus(), ".")

		// turn "abc def" into "<observed id="abc">abc</observed> <observed id="def">def</observed>"
		// the complication is that x.MarkedUp contains html; use x.Accented to find the words
		wds := strings.Split(lines[i].Accented, " ")
		wds = unique(wds)

		newline := lines[i].MarkedUp
		for w, _ := range wds {
			// this is going to have a problem if something already abuts markup...
			// will need to keep track of the complete list of terminating items.

			pattern, e := regexp.Compile(fmt.Sprintf("(^|\\s)(%s)(\\s|\\.|,|;|·|$)", wds[w]))
			if e == nil {
				// you will barf if wds[w] = *
				newline = pattern.ReplaceAllString(newline, `$1<observed id="$2">$2</observed>$3`)
			} else {
				msg(fmt.Sprintf("buildbrowsertable() could not regex compile %s", wds[w]), 4)
			}
		}

		var bl string
		if lines[i].TbIndex != focus {
			bl = newline
		} else {
			bl = fmt.Sprintf("%s%s%s", fla, newline, flb)
		}

		trr = append(trr, fmt.Sprintf(tr, lines[i].Annotations, bl, cit))
	}
	tab := strings.Join(trr, "")

	// that was the body, now do the head and tail
	top := fmt.Sprintf(`<div id="browsertableuid" uid="%s"></div>`, lines[0].FindAuthor())
	top += `<table><tbody>`
	top += `<tr class="spacing">` + strings.Repeat("&nbsp;", MINBROWSERWIDTH) + `</tr>`

	tab = top + tab + `</tbody></table>`

	return tab
}

func formatmultilinebrackets(html string) string {
	// try to get the spanning right in a browser table for the following:
	// porrigant; sunt qui non usque ad vitium accedant (necesse 	114.11.4
	// est enim hoc facere aliquid grande temptanti) sed qui ipsum 	114.11.5

	// we have already marked the opening w/ necesse... but it needs to close and reopen for a new table row
	// use the block delimiter ("✃✃✃") to help with this

	// sunt qui illos detineant et✃✃✃porrigant; sunt qui non usque ad vitium accedant (<span class="editorialmarker_roundbrackets">necesse✃✃✃est enim hoc facere aliquid grande temptanti</span>) sed qui ipsum✃✃✃vitium ament.✃✃✃

	// also want to do this before you have a lot of "span" spam in the line...

	// the next ovverruns; need to stop at "<"
	// pattern := regexp.MustCompile("(?P<brktype><span class=\"editorialmarker_\\w+brackets\">)(?P<line_end>.*?)✃✃✃(?P<line_start>.*?</span>)")

	pattern := regexp.MustCompile("(?P<brktype><span class=\"editorialmarker_\\w+brackets\">)(?P<line_end>[^\\<]*?)✃✃✃(?P<line_start>.*?</span>)")
	html = pattern.ReplaceAllString(html, "$1$2</span>✃✃✃$1$3")

	return html
}
