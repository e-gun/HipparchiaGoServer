package format

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/mps"
	"github.com/e-gun/HipparchiaGoServer/internal/search"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"regexp"
	"strings"
)

// ProlixBrowswerCitations - the prolix bibliographic info for a line/work
func ProlixBrowswerCitations(f str.DbWorkline, l str.DbWorkline) string {
	const (
		CVBROWSER = `
		<p class="currentlyviewing">
		%s<br>
		<span class="currentlyviewingcitation">%s — %s</span>
		<br>
		<span class="publicationinfo">%s</span>
		%s</p>`
		CVTMAKER = `
		<p class="currentlyviewing">%s<br><span class="publicationinfo">(%s)</span>%s</p>`
		CT     = `<cvauthor">%s</span>, <cvwork">%s</span>`
		HIDDEN = `<!-- transmission: %s -->`
	)

	w := search.DbWlnMyWk(&f)

	au := mps.DbWkMyAu(w).Name
	ti := w.Title

	ci := fmt.Sprintf(CT, au, ti)
	ci = gen.AvoidLongLines(ci, vv.MINBROWSERWIDTH)
	ci = strings.Replace(ci, "<cv", `<span class="currentlyviewing`, -1)

	dt := `<br>(Assigned date of %s)`
	beg := BasicCitation(f)
	end := BasicCitation(l)
	pi := gen.AvoidLongLines(w.Pub, 2*vv.MINBROWSERWIDTH) // lots of hidden markup + we are small
	id := search.FormatInscriptionDates(dt, &f)

	cv := fmt.Sprintf(CVBROWSER, ci, beg, end, pi, id)

	// to let the textmaker use this code...
	if f.TbIndex == l.TbIndex {
		cv = fmt.Sprintf(CVTMAKER, ci, pi, id)
	}

	if w.Xmit != "" {
		// helps with debugging: this lets you find your way back to the original phi file
		cv += fmt.Sprintf(HIDDEN, w.Xmit)
	}

	return cv
}

// BuildBrowserTable - where the actual HTML gets generated; this table is better for cut-and-paste...
func BuildBrowserTable(focus int, lines []str.DbWorkline, zaplunates bool, regularizewidth bool) string {
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
	)

	// the builder has failed to parse some notes; do something about that
	// longnotes := false // cap chars per line in notes
	// lines = reannotatelines(lines, longnotes)

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
			r := fmt.Sprintf(OBSREGTEMPL, gen.UVσςϲCapsVariants(w))
			pattern, e := regexp.Compile(r)
			if e != nil {
				// you will barf if w = *
				// Msg.PEEK(fmt.Sprintf(FAIL, w))
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
				np, e := regexp.Compile(fmt.Sprintf(OBSREGTEMPL, gen.UVσςϲCapsVariants(lmw)))
				if e != nil {
					// web.Msg.PEEK(fmt.Sprintf(FAIL, lmw))
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

		cit := SelectivelyDisplayCitations(lines[i], previous, focus)

		blines[i] = bl
		bcites[i] = fmt.Sprintf("<span class=\"eighty\">%s</span>&nbsp;", cit) // the "normal" sized space is to maintain vertical alignment
		bnotes[i] = fmt.Sprintf("<span class=\"eighty\">%s</span>&nbsp;", FormatAnnotations(lines[i]))
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
	if regularizewidth {
		top += stabilizebrowserwidth(longestnote)
	}
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
