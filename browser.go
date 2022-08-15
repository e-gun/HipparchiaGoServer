package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// HipparchiaBrowser - browse Author A at line X with a context of Y lines
func HipparchiaBrowser(au string, wk string, fc int64, ctx int64) []byte {
	// build a response to "GET /browse/linenumber/gr0062/028/14672 HTTP/1.1"
	// note the high initialization costs of doing this as CLI vs a proper server

	authormap := authormapper()
	workmap := workmapper()

	// [b] acquire the lines we need to display in the body

	lines := simplecontextgrabber(au, fc, ctx)
	k := fmt.Sprintf("%sw%s", au, wk)
	w := workmap[k]

	// [c] format the lines

	// not yet implemented
	// lines = paragraphformatting(lines)

	// [d] acquire and format the HTML

	ci := formatcitationinfo(authormap, w, lines[0])
	pi := formatpublicationinfo(workmap[k])
	tr := buildbrowsertable(fc, lines)

	// [e] fill out the JSON-ready struct

	fl := fmt.Sprintf(`linenumber/%s/%s/%d`, au, wk, lines[0].TbIndex)
	ll := fmt.Sprintf(`linenumber/%s/%s/%d`, au, wk, lines[len(lines)-1].TbIndex)
	ab := fmt.Sprintf(`%s [%s]`, authormap[au].Cleaname, au)
	wb := fmt.Sprintf(`%s (w%s)`, w.Title, w.WorkNum)

	var bp BrowsedPassage
	bp.Browseforwards = ll
	bp.Browseback = fl
	bp.Authornumber = au
	bp.Workid = lines[0].WkUID
	bp.Authorboxcontents = ab
	bp.Workboxcontents = wb
	bp.Browserhtml = ci + pi + tr
	bp.Worknumber = wk

	// debugging
	// fmt.Println(ci)

	js, e := json.Marshal(bp)
	checkerror(e)

	return js
}

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
		{"pages", "pp. ", ". "},
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
	// INCOMPLETE
	cv := `
		<p class="currentlyviewing">
		<span class="currentlyviewingauthor">%s</span>, <br />
		<span class="currentlyviewingwork">%s</span><br />
		<span class="currentlyviewingcitation">%s</span></p>`
	au := authormap[w.FindAuthor()].Name
	ti := w.Title
	cf := w.CitationFormat()
	loc := l.FindLocus()
	cf = cf[6-(len(loc)) : 6]

	var cit []string
	for i, _ := range loc {
		cit = append(cit, fmt.Sprintf("%s %s", cf[i], loc[i]))
	}
	fc := strings.Join(cit, ", ")
	cv = fmt.Sprintf(cv, au, ti, fc)

	return cv
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
			pattern := regexp.MustCompile(fmt.Sprintf("(^|\\s)(%s)(\\s|\\.|,|;|·|$)", wds[w]))
			newline = pattern.ReplaceAllString(newline, `$1<observed id="$2">$2</observed>$3`)
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
	MINWIDTH := 80
	top := fmt.Sprintf(`<div id="browsertableuid" uid="%s"></div>\n`, lines[0].FindAuthor())
	top += `<table>\n<tbody>\n`
	top += `<tr class="spacing">` + strings.Repeat("&nbsp;", MINWIDTH) + `</tr>\n`

	tab = top + tab + `</tbody>\n</table>\n`

	return tab
}
