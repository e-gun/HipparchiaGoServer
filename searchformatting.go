//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"bytes"
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"regexp"
	"strings"
	"text/template"
	"time"
)

type SearchOutputJSON struct {
	Title         string `json:"title"`
	Searchsummary string `json:"searchsummary"`
	Found         string `json:"found"`
	Image         string `json:"image"`
	JS            string `json:"js"`
}

func formatnocontextresults(ss SearchStruct) SearchOutputJSON {
	var out SearchOutputJSON
	out.JS = fmt.Sprintf(BROWSERJS, "browser")
	out.Title = ss.Seeking
	out.Image = ""
	out.Searchsummary = formatfinalsearchsummary(&ss)

	tablerow := `
	<tr class="%s">
		<td>
			<span class="findnumber">[%d]</span>&nbsp;&nbsp;%s%s
			%s
		</td>
		<td class="leftpad">
			<span class="foundtext">%s</span>
		</td>
	</tr>
	`
	dtt := `[<span class="date">%s</span>]`

	searchterm := gethighlighter(ss)

	rows := make([]string, len(ss.Results))
	for i, r := range ss.Results {
		r.PurgeMetadata()
		// highlight search term; should be folded into a single function w/ highlightsearchterm() below [type problem now]
		if searchterm.MatchString(r.MarkedUp) {
			// line.Contents = pattern.ReplaceAllString(line.Contents, `<span class="match">$1</span>`)
			r.MarkedUp = searchterm.ReplaceAllString(r.MarkedUp, `<span class="match">$0</span>`)
		} else {
			// might be in the hyphenated line
			if searchterm.MatchString(r.Hyphenated) {
				// needs more fiddling
				r.MarkedUp += fmt.Sprintf(`&nbsp;&nbsp;(&nbsp;match:&nbsp;<span class="match">%s</span>&nbsp;)`, r.Hyphenated)
			}
		}

		mu := formateditorialbrackets(r.MarkedUp)

		rc := ""
		if i%3 == 2 {
			rc = "nthrow"
		} else {
			rc = "regular"
		}

		au := AllAuthors[r.FindAuthor()].Shortname
		wk := AllWorks[r.WkUID].Title
		lk := r.BuildHyperlink()
		lc := strings.Join(r.FindLocus(), ".")
		wd := formatinscriptiondates(dtt, r)
		pl := formatinscriptionplaces(r)

		// <span class="foundauthor">%s</span>,&nbsp;<span class="foundwork">%s</span>: <browser id="%s"><span class="foundlocus">%s</span></browser>
		ct := `<spcauthor">%s</span>,&nbsp;<spcwork">%s</span>: <browser_id="%s"><spclocus">%s</span></browser>`
		ci := fmt.Sprintf(ct, au, wk, lk, lc)
		ci = avoidlonglines(ci, MAXTITLELENGTH)
		ci = strings.Replace(ci, "<spc", `<span class="found`, -1)
		ci = strings.Replace(ci, `browser_id`, `browser id`, -1)

		fm := fmt.Sprintf(tablerow, rc, i+1, wd, pl, ci, mu)
		rows[i] = fm
	}

	out.Found = "<tbody>" + strings.Join(rows, "") + "</tbody>"
	return out
}

type ResultPassageLine struct {
	Locus           string
	Contents        string
	Hyphenated      string
	ContinuingStyle string
	IsHighlight     bool
}

func formatwithcontextresults(ss SearchStruct) SearchOutputJSON {
	thesession := sessions[ss.User]

	type PsgFormattingTemplate struct {
		Findnumber  int
		Foundauthor string
		Foundwork   string
		FindDate    string
		FindURL     string
		FindLocus   string
		FindCity    string
		RawCTX      []DbWorkline
		CookedCTX   []ResultPassageLine
		LocusBody   string
	}

	// gather all the lines you need: this is much faster than simplecontextgrabber() 200x in a single threaded loop
	// turn it into a new search where we accept any character as enough to yield a hit: ""
	res := clonesearch(ss, 3)
	res.Results = ss.Results
	res.Seeking = ""
	res.LemmaOne = ""
	res.Proximate = ""
	res.LemmaTwo = ""
	res.Limit = (ss.Limit * int64(thesession.HitContext)) * 3

	context := int64(thesession.HitContext / 2)
	t := `%s_FROM_%d_TO_%d`
	res.SearchIn.Passages = make([]string, len(res.Results))
	for i, r := range res.Results {
		low := r.TbIndex - context
		high := r.TbIndex + context
		if low < 1 {
			// avoid "gr0258_FROM_-1_TO_3"
			low = 1
		}
		res.SearchIn.Passages[i] = fmt.Sprintf(t, r.FindAuthor(), low, high)
	}

	res.Results = []DbWorkline{}
	BuildQueriesForSS(&res)
	res = HGoSrch(res)

	// now you have all the lines you will ever need
	linemap := make(map[string]DbWorkline)
	for _, r := range res.Results {
		linemap[r.BuildHyperlink()] = r
	}

	// iterate over the results to build the raw core data
	urt := `linenumber/%s/%s/%d`
	dtt := `[<span class="date">%s</span>]`

	allpassages := make([]PsgFormattingTemplate, len(ss.Results))
	for i, r := range ss.Results {
		var psg PsgFormattingTemplate
		psg.Findnumber = i + 1
		psg.Foundauthor = AllAuthors[r.FindAuthor()].Name
		psg.Foundwork = AllWorks[r.WkUID].Title
		psg.FindURL = r.BuildHyperlink()
		psg.FindLocus = strings.Join(r.FindLocus(), ".")
		psg.FindDate = formatinscriptiondates(dtt, r)
		psg.FindCity = formatinscriptionplaces(r)

		for j := r.TbIndex - context; j <= r.TbIndex+context; j++ {
			url := fmt.Sprintf(urt, r.FindAuthor(), r.FindWork(), j)
			psg.RawCTX = append(psg.RawCTX, linemap[url])
		}

		// if you want to do this the horrifyingly slow way...
		// psg.RawCTX = simplecontextgrabber(r.FindAuthor(), r.TbIndex, int64(thesession.HitContext/2))

		psg.CookedCTX = make([]ResultPassageLine, len(psg.RawCTX))
		for j := 0; j < len(psg.RawCTX); j++ {
			c := ResultPassageLine{}
			c.Locus = strings.Join(psg.RawCTX[j].FindLocus(), ".")

			if psg.RawCTX[j].BuildHyperlink() == psg.FindURL {
				c.IsHighlight = true
			} else {
				c.IsHighlight = false
			}
			psg.RawCTX[j].PurgeMetadata()
			c.Contents = psg.RawCTX[j].MarkedUp
			c.Hyphenated = psg.RawCTX[j].Hyphenated
			psg.CookedCTX[j] = c
		}
		allpassages[i] = psg
	}

	// fix the unmattched spans
	for _, p := range allpassages {
		// at the top
		p.CookedCTX[0].Contents = unbalancedspancleaner(p.CookedCTX[0].Contents)

		// across the whole
		block := make([]string, len(p.CookedCTX))
		for j, c := range p.CookedCTX {
			block[j] = c.Contents
		}
		whole := strings.Join(block, "✃✃✃")

		whole = textblockcleaner(whole)

		// reassemble
		block = strings.Split(whole, "✃✃✃")
		for i, b := range block {
			p.CookedCTX[i].Contents = b
		}
	}

	// highlight the search term: this includes the hyphenated_line issue
	searchterm := gethighlighter(ss)

	for _, p := range allpassages {
		for i, r := range p.CookedCTX {
			if r.IsHighlight && searchterm != nil {
				p.CookedCTX[i].Contents = fmt.Sprintf(`<span class="highlight">%s</span>`, p.CookedCTX[i].Contents)
				// highlightfocusline(&p.CookedCTX[i])
				highlightsearchterm(searchterm, &p.CookedCTX[i])
			}
			if len(ss.LemmaTwo) > 0 {
				// look for the proximate term
				re := lemmaintoregexslice(ss.LemmaTwo)
				pat, e := regexp.Compile(strings.Join(re, "|"))
				if e != nil {
					pat = regexp.MustCompile("FAILED_FIND_NOTHING")
					msg(fmt.Sprintf("searchtermfinder() could not compile the following: %s", strings.Join(re, "|")), 1)
				}
				highlightsearchterm(pat, &p.CookedCTX[i])
			}
			if len(ss.Proximate) > 0 {
				// look for the proximate term
				pat := searchtermfinder(ss.Proximate)
				highlightsearchterm(pat, &p.CookedCTX[i])
			}
		}
	}

	pht := `
	<locus>
		<span class="findnumber">[{{.Findnumber}}]</span>&nbsp;&nbsp;{{.FindDate}}{{.FindCity}}
		<span class="foundauthor">{{.Foundauthor}}</span>,&nbsp;<span class="foundwork">{{.Foundwork}}</span>
		<browser id="{{.FindURL}}"><span class="foundlocus">{{.FindLocus}}</span></browser>
	</locus>
	{{.LocusBody}}`

	tmpl, e := template.New("tr").Parse(pht)
	chke(e)

	plt := `<span class="locus">%s</span>&nbsp;<span class="foundtext">%s</span><br>
	`

	rows := make([]string, len(allpassages))
	for i, p := range allpassages {
		lines := make([]string, len(p.CookedCTX))
		for j, l := range p.CookedCTX {
			c := fmt.Sprintf(plt, l.Locus, l.Contents)
			lines[j] = c
		}
		p.LocusBody = strings.Join(lines, "")
		var b bytes.Buffer
		err := tmpl.Execute(&b, p)
		chke(err)

		rows[i] = b.String()
	}

	// ouput

	var out SearchOutputJSON
	out.JS = fmt.Sprintf(BROWSERJS, "browser")
	out.Title = restorewhitespace(ss.Seeking)
	out.Image = ""
	out.Searchsummary = formatfinalsearchsummary(&ss)
	out.Found = strings.Join(rows, "")

	return out
}

func formatfinalsearchsummary(s *SearchStruct) string {
	// ex:
	//        Sought <span class="sought">»ἡμέρα«</span>
	//        <br>
	//        Searched 49,230 works and found 200 passages (0.12s)
	//        <br>
	//        Sorted by author name
	//        <!-- unlimited hits per author -->
	//        <br>
	//        <!-- dates did not matter -->
	//        [Search suspended: result cap reached.]

	t := `
		%s
		<br>
		Searched %d works and found %d passages (%ss)
		<br>
		Sorted by %s
		%s
		<br>
		%s
		%s
	`
	m := message.NewPrinter(language.English)

	var dr string
	if sessions[s.User].Earliest != MINDATESTR || sessions[s.User].Latest != MAXDATESTR {
		a := formatbcedate(sessions[s.User].Earliest)
		b := formatbcedate(sessions[s.User].Latest)
		dr = fmt.Sprintf("Searched between %s and %s<br>", a, b)
	} else {
		dr = "<!-- dates did not matter -->"
	}

	var hitcap string
	if int64(len(s.Results)) == s.Limit {
		hitcap = "[Search suspended: result cap reached.]"
	} else {
		hitcap = "<!-- did not hit the results cap -->"
	}

	oh := "<!-- unlimited hits per author -->"
	if s.OneHit {
		oh = `<br><span class="small">(only one hit allowed per author table)</span>`
	}

	so := sessions[s.User].SortHitsBy
	// shortname, converted_date, provenance, universalid
	switch so {
	case "shortname":
		so = "author name"
	case "converted_date":
		so = "date"
	case "provenance":
		so = "work location"
	case "universalid":
		so = "ID"
	}

	el := fmt.Sprintf("%.2f", time.Now().Sub(s.Launched).Seconds())
	// need to record # of works and not # of tables somewhere & at the right moment...
	sum := m.Sprintf(t, s.InitSum, s.SearchSize, len(s.Results), el, so, oh, dr, hitcap)
	return sum
}

func formatinitialsummary(s SearchStruct) string {
	// ex:
	// Sought <span class="sought">»ἡμέρα«</span> within 2 lines of all 79 forms of <span class="sought">»ἀγαθόϲ«</span>

	tmp := `Sought %s<span class="sought">»%s«</span>%s`
	win := `%s within %d %s of %s<span class="sought">»%s«</span>`

	yn := ""
	if s.NotNear {
		yn = " not "
	}

	af1 := ""
	sk := s.Seeking
	if len(s.LemmaOne) != 0 {
		af := "all %d forms of "
		sk = s.LemmaOne
		af1 = fmt.Sprintf(af, len(AllLemm[sk].Deriv))
	}

	two := ""
	if s.Twobox {
		sk2 := s.Proximate
		af2 := ""
		if len(s.LemmaTwo) != 0 {
			af3 := "all %d forms of "
			sk2 = s.LemmaTwo
			af2 = fmt.Sprintf(af3, len(AllLemm[sk2].Deriv))
		}
		two = fmt.Sprintf(win, yn, s.ProxVal, s.ProxScope, af2, sk2)
	}
	sum := fmt.Sprintf(tmp, af1, sk, two)
	return sum
}

func highlightsearchterm(pattern *regexp.Regexp, line *ResultPassageLine) {
	// 	html markup for the search term in the line so it can jump out at you
	//
	//	regexequivalent is compiled via searchtermfinder() in rt-search.go
	//

	// see the warnings and caveats at highlightsearchterm() in searchformatting.py
	if pattern.MatchString(line.Contents) {
		// line.Contents = pattern.ReplaceAllString(line.Contents, `<span class="match">$1</span>`)
		line.Contents = pattern.ReplaceAllString(line.Contents, `<span class="match">$0</span>`)
	} else {
		// might be in the hyphenated line
		if pattern.MatchString(line.Hyphenated) {
			// needs more fiddling
			line.Contents += fmt.Sprintf(`&nbsp;&nbsp;(&nbsp;match:&nbsp;<span class="match">%s</span>&nbsp;)`, line.Hyphenated)
		}
	}
}

func formatinscriptiondates(template string, dbw DbWorkline) string {
	// show the years for inscriptions
	datestring := ""
	fc := dbw.FindCorpus()
	dated := fc == "in" || fc == "ch" || fc == "dp"
	if dated {
		cd := i64tobce(AllWorks[dbw.WkUID].ConvDate)
		if cd == "2500 C.E." {
			cd = "??? BCE/CE"
		}
		datestring = fmt.Sprintf(template, strings.Replace(cd, ".", "", -1))
	}
	return datestring
}

func formatinscriptionplaces(dbw DbWorkline) string {
	// show the places for inscriptions
	placestring := ""
	fc := dbw.FindCorpus()
	placed := fc == "in" || fc == "ch" || fc == "dp"
	if placed {
		placestring = fmt.Sprintf(` [<span class="rust">%s</span>] `, AllWorks[dbw.WkUID].Prov)
	}
	return placestring
}

// textblockcleaner - address multi-line formatting challenges by running a suite of clean-ups
func textblockcleaner(html string) string {
	// do it early and in this order
	// presupposes the snippers are in there: "✃✃✃"
	html = unbalancedspancleaner(html)
	html = formateditorialbrackets(html)
	html = formatmultilinebrackets(html)

	return html
}

// unbalancedspancleaner - helper for textblockcleaner()
func unbalancedspancleaner(html string) string {
	// 	unbalanced spans inside of result chunks: ask for 4 lines of context and search for »ἀδύνατον γ[άὰ]ρ«
	//	this will cough up two examples of the problem in Alexander, In Aristotelis analyticorum priorum librum i commentarium
	//
	//	the first line of context shows spans closing here that were opened in a previous line
	//
	//		<span class="locus">98.14</span>&nbsp;<span class="foundtext">ὅρων ὄντων πρὸϲ τὸ μέϲον.</span></span></span><br />
	//
	//	the last line of the context is opening a span that runs into the next line of the text where it will close
	//	but since the next line does not appear, the span remains open. This will make the next results bold + italic + ...
	//
	//		<span class="locus">98.18</span>&nbsp;<span class="foundtext"><hmu_roman_in_a_greek_text>p. 28a18 </hmu_roman_in_a_greek_text><span class="title"><span class="expanded">Καθόλου μὲν οὖν ὄντων, ὅταν καὶ τὸ Π καὶ τὸ Ρ παντὶ</span><br />
	//
	//	the solution:
	//		open anything that needs opening: this needs to be done with the first line
	//		close anything left hanging: this needs to be done with the whole passage
	//
	//	return the html with these supplemental tags

	xopen := `<span class="htmlbalancingsupplement">`
	xclose := `</span>`

	op := regexp.MustCompile("<span")
	cl := regexp.MustCompile("</span>")

	opened := len(op.FindAllString(html, -1))
	closed := len(cl.FindAllString(html, -1))

	if closed > opened {
		for i := 0; i < closed-opened; i++ {
			html = xopen + html
		}
	}

	if opened > closed {
		for i := 0; i < opened-closed; i++ {
			html = html + xclose
		}
	}
	return html
}

// don't let regex compliation get looped...
var (
	esbboth = regexp.MustCompile("\\[(.*?)\\]")
	erbboth = regexp.MustCompile("\\((.*?)\\)")
	eabboth = regexp.MustCompile("⟨(.*?)⟩")
	ecbboth = regexp.MustCompile("\\{(.*?)\\}")
)

// formateditorialbrackets - helper for textblockcleaner()
func formateditorialbrackets(html string) string {
	// sample:
	// [<span class="editorialmarker_squarebrackets">ἔδοχϲεν τε͂ι βολε͂ι καὶ το͂ι</span>]

	// special cases:
	// [a] no "open" or "close" bracket at the head/tail of a line: ^τε͂ι βολε͂ι καὶ] το͂ι...$ / ^...ἔδοχϲεν τε͂ι βολε͂ι [καὶ το͂ι$
	// [b] we are continuing from a previous state: no brackets here, but should insert a span; the previous line will need to notify the subsequent...

	// types: editorialmarker_angledbrackets; editorialmarker_curlybrackets, editorialmarker_roundbrackets, editorialmarker_squarebrackets
	//

	// try running this against text blocks only: it probably saves plenty of trouble later

	// see buildtext() in textbuilder.py for some regex recipies

	html = esbboth.ReplaceAllString(html, `[<span class="editorialmarker_squarebrackets">$1</span>]`)
	html = erbboth.ReplaceAllString(html, `(<span class="editorialmarker_roundbrackets">$1</span>)`)
	html = eabboth.ReplaceAllString(html, `⟨<span class="editorialmarker_angledbrackets">$1</span>⟩`)
	html = ecbboth.ReplaceAllString(html, `{<span class="editorialmarker_curlybrackets">$1</span>}`)

	return html
}

// formatmultilinebrackets - helper for textblockcleaner()
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

	// this won't dow 3+ lines, just 2...
	pattern := regexp.MustCompile("(?P<brktype><span class=\"editorialmarker_\\w+brackets\">)(?P<line_end>[^\\<]*?)✃✃✃(?P<line_start>[^\\]]*?</span>)")
	html = pattern.ReplaceAllString(html, "$1$2</span>✃✃✃$1$3")

	return html
}

// gethighlighter - set regex to highlight the search term
func gethighlighter(ss SearchStruct) *regexp.Regexp {
	var re *regexp.Regexp

	skg := ss.Seeking
	prx := ss.Proximate
	if ss.SkgRewritten {
		// quasi-bugged because of "\s" --> "\[sS]"; meanwhile whitespacer() can't use " " for its own reasons...
		// ((^|\[sS])[εἐἑἒἓἔἕὲέἘἙἚἛἜἝΕ][νΝ] [οὀὁὂὃὄὅόὸὈὉὊὋὌὍΟ][ρῤῥῬ][εἐἑἒἓἔἕὲέἘἙἚἛἜἝΕ][ϲσΣςϹ][τΤ][ηᾐᾑᾒᾓᾔᾕᾖᾗῂῃῄῆῇἤἢἥἣὴήἠἡἦἧᾘᾙᾚᾛᾜᾝᾞᾟἨἩἪἫἬἭἮἯΗ](\[sS]|$))
		skg = strings.Replace(whitespacer(skg, &ss), "(^|\\s)", "(^| )", 1)
		skg = strings.Replace(whitespacer(skg, &ss), "(\\s|$)", "( |$)", 1)
		prx = strings.Replace(whitespacer(prx, &ss), "(^|\\s)", "(^| )", 1)
		prx = strings.Replace(whitespacer(prx, &ss), "(\\s|$)", "( |$)", 1)
	}

	if len(ss.Seeking) != 0 {
		re = searchtermfinder(skg)
	} else if len(ss.LemmaOne) != 0 {
		re = lemmahighlighter(ss.LemmaOne)
	} else if len(ss.Proximate) != 0 {
		re = searchtermfinder(prx)
	} else if len(ss.LemmaTwo) != 0 {
		re = lemmahighlighter(ss.LemmaTwo)
	} else {
		msg(fmt.Sprintf("gethighlighter() cannot find anything to highlight\n\t%ss", ss.InitSum), 3)
		re = nil
	}
	return re
}

func lemmahighlighter(lm string) *regexp.Regexp {
	// don't let "(^|\s)τρεῖϲ(\s|$)|(^|\s)τρία(\s|$)|(^|\s)τριϲίν(\s|$)|(^|\s)τριῶν(\s|$)|(^|\s)τρί(\s|$)|(^|\s)τριϲί(\s|$)"
	// turn into "(^|\[sS])[τΤ][ρῤῥῬ][εἐἑἒἓἔἕὲέἘἙἚἛἜἝΕ]ῖ[ϲσΣςϹ](\[sS]|$)|(^|\[sS])..."
	// can't send "(^|\s)" through universalpatternmaker()

	// abutting markup is killing off some items, but adding "<" and ">" produces worse problems still

	// now you also need to worry about punctuation that abuts the find
	// tp := `[\^\s;]%s[\s\.,;·’$]`
	tp := `%s` // move from match $1 to $0 in highlightsearchterm() yielded this shift...

	lemm := AllLemm[lm].Deriv

	whole := strings.Join(lemm, ")✃✃✃(")
	st := universalpatternmaker(whole)
	lup := strings.Split(st, "✃✃✃")
	for i, l := range lup {
		lup[i] = fmt.Sprintf(tp, l)
	}
	rec := strings.Join(lup, "|")

	r, e := regexp.Compile(rec)
	if e != nil {
		msg("gethighlighter() could not compile LemmaOne into regex", 3)
	}
	return r
}
