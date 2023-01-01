//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
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

const (
	MUREPLACE   = `<span class="match">$0</span>` // note $0 vs $1
	HYPHREPLACE = `&nbsp;&nbsp;(&nbsp;match:&nbsp;<span class="match">%s</span>&nbsp;)`
)

// FormatNoContextResults - build zero context search results table
func FormatNoContextResults(ss *SearchStruct) SearchOutputJSON {
	const (
		TABLEROW = `
		<tr class="{{.TRClass}}">
			<td>
				<span class="findnumber">[{{.FindNumber}}]</span>&nbsp;&nbsp;{{.FindDate}}{{.FindCity}}
				{{.FindLocus}}
			</td>
			<td class="leftpad">
				<span class="foundtext">{{.TheLine}}</span>
			</td>
		</tr>`

		DATES    = `[<span class="date">%s</span>]`
		SPSUBBER = `<spcauthor">%s</span>,&nbsp;<spcwork">%s</span>: <browser_id="%s"><spclocus">%s</span></browser>`
	)

	type TRTempl struct {
		TRClass    string
		FindNumber int
		FindDate   string
		FindCity   string
		FindLocus  string
		TheLine    string
	}

	searchterm := gethighlighter(ss)

	trt, e := template.New("trt").Parse(TABLEROW)
	chke(e)

	var b bytes.Buffer

	for i, r := range ss.Results {
		r.PurgeMetadata()
		// highlight search term; should be folded into a single function w/ highlightsearchterm() below [type problem now]
		if searchterm.MatchString(r.MarkedUp) {
			r.MarkedUp = searchterm.ReplaceAllString(r.MarkedUp, MUREPLACE)
		} else {
			// might be in the hyphenated line
			if searchterm.MatchString(r.Hyphenated) {
				// needs more fiddling
				r.MarkedUp += fmt.Sprintf(HYPHREPLACE, r.Hyphenated)
			}
		}

		mu := formateditorialbrackets(r.MarkedUp)

		rc := ""
		if i%3 == 2 {
			rc = "nthrow"
		} else {
			rc = "regular"
		}

		au := r.MyAu().Shortname
		wk := r.MyWk().Title
		lk := r.BuildHyperlink()
		lc := strings.Join(r.FindLocus(), ".")

		// <span class="foundauthor">%s</span>,&nbsp;<span class="foundwork">%s</span>: <browser id="%s"><span class="foundlocus">%s</span></browser>
		ci := fmt.Sprintf(SPSUBBER, au, wk, lk, lc)
		ci = avoidlonglines(ci, MAXTITLELENGTH)
		ci = strings.Replace(ci, "<spc", `<span class="found`, -1)
		ci = strings.Replace(ci, `browser_id`, `browser id`, -1)

		tr := TRTempl{
			TRClass:    rc,
			FindNumber: i + 1,
			FindDate:   formatinscriptiondates(DATES, &r),
			FindCity:   formatinscriptionplaces(&r),
			FindLocus:  ci,
			TheLine:    mu,
		}

		err := trt.Execute(&b, tr)
		chke(err)

	}

	var out SearchOutputJSON
	out.JS = fmt.Sprintf(BROWSERJS, "browser")
	out.Title = ss.Seeking
	out.Image = ""
	out.Searchsummary = formatfinalsearchsummary(ss)

	out.Found = "<tbody>" + b.String() + "</tbody>"
	if Config.ZapLunates {
		out.Found = DeLunate(out.Found)
	}

	return out
}

type ResultPassageLine struct {
	Locus           string
	Contents        string
	Hyphenated      string
	ContinuingStyle string
	IsHighlight     bool
}

// FormatWithContextResults - build n-lines of context search results table
func FormatWithContextResults(thesearch *SearchStruct) SearchOutputJSON {
	// profiling will show that the bulk of your time is spent on (in descending order):
	// lemmaintoregexslice(), regexp.Compile(strings.Join(re, "|")), and highlightsearchterm()
	// the cost is not outlandish, but regex is fairly expensive

	const (
		FINDTEMPL = `
		<locus>
			<span class="findnumber">[{{.Findnumber}}]</span>&nbsp;&nbsp;{{.FindDate}}{{.FindCity}}
			<span class="foundauthor">{{.Foundauthor}}</span>,&nbsp;<span class="foundwork">{{.Foundwork}}</span>
			<browser id="{{.FindURL}}"><span class="foundlocus">{{.FindLocus}}</span></browser>
		</locus>
		{{.LocusBody}}`

		FOUNDLINE = `<span class="locus">%s</span>&nbsp;<span class="foundtext">%s</span><br>
		`
		PSGTEMPL    = `%s_FROM_%d_TO_%d`
		URT         = `index/%s/%s/%d`
		DTT         = `[<span class="date">%s</span>]`
		HIGHLIGHTER = `<span class="highlight">%s</span>`
		SNIP        = `✃✃✃`
	)
	thesession := SafeSessionRead(thesearch.User)

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

	// gather all the lines you need: this is much faster than SimpleContextGrabber() 200x in a single threaded loop
	// turn it into a new search where we accept any character as enough to yield a hit: ""
	res := CloneSearch(*thesearch, 3)
	res.Results = thesearch.Results
	res.Seeking = ""
	res.LemmaOne = ""
	res.Proximate = ""
	res.LemmaTwo = ""
	res.CurrentLimit = (thesearch.CurrentLimit * thesession.HitContext) * 3

	context := thesession.HitContext / 2

	res.SearchIn.Passages = make([]string, len(res.Results))
	for i, r := range res.Results {
		low := r.TbIndex - context
		high := r.TbIndex + context
		if low < 1 {
			// avoid "gr0258_FROM_-1_TO_3"
			low = 1
		}
		res.SearchIn.Passages[i] = fmt.Sprintf(PSGTEMPL, r.AuID(), low, high)
	}

	res.Results = []DbWorkline{}
	SSBuildQueries(&res)
	res = HGoSrch(res)

	// now you have all the lines you will ever need
	linemap := make(map[string]DbWorkline)
	for _, r := range res.Results {
		linemap[r.BuildHyperlink()] = r
	}

	// iterate over the results to build the raw core data

	allpassages := make([]PsgFormattingTemplate, len(thesearch.Results))
	for i, r := range thesearch.Results {
		var psg PsgFormattingTemplate
		psg.Findnumber = i + 1
		psg.Foundauthor = r.MyAu().Name
		psg.Foundwork = r.MyWk().Title
		psg.FindURL = r.BuildHyperlink()
		psg.FindLocus = strings.Join(r.FindLocus(), ".")
		psg.FindDate = formatinscriptiondates(DTT, &r)
		psg.FindCity = formatinscriptionplaces(&r)

		for j := r.TbIndex - context; j <= r.TbIndex+context; j++ {
			url := fmt.Sprintf(URT, r.AuID(), r.WkID(), j)
			psg.RawCTX = append(psg.RawCTX, linemap[url])
		}

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
		whole := strings.Join(block, SNIP)

		whole = textblockcleaner(whole)

		// reassemble
		block = strings.Split(whole, SNIP)
		for i, b := range block {
			p.CookedCTX[i].Contents = b
		}
	}

	// highlight the search term: this includes the hyphenated_line issue
	searchterm := gethighlighter(thesearch)

	for _, p := range allpassages {
		for i, r := range p.CookedCTX {
			if r.IsHighlight && searchterm != nil {
				p.CookedCTX[i].Contents = fmt.Sprintf(HIGHLIGHTER, p.CookedCTX[i].Contents)
				highlightsearchterm(searchterm, &p.CookedCTX[i])
			}
			if len(thesearch.LemmaTwo) > 0 {
				// look for the proximate term
				re := lemmaintoregexslice(thesearch.LemmaTwo)
				pat, e := regexp.Compile(strings.Join(re, "|"))
				if e != nil {
					pat = regexp.MustCompile("FAILED_FIND_NOTHING")
					msg(fmt.Sprintf("SearchTermFinder() could not compile the following: %s", strings.Join(re, "|")), MSGWARN)
				}
				highlightsearchterm(pat, &p.CookedCTX[i])
			}
			if len(thesearch.Proximate) > 0 {
				// look for the proximate term
				pat := SearchTermFinder(thesearch.Proximate)
				highlightsearchterm(pat, &p.CookedCTX[i])
			}
		}
	}

	tmpl, e := template.New("tr").Parse(FINDTEMPL)
	chke(e)

	var b bytes.Buffer
	for _, p := range allpassages {
		lines := make([]string, len(p.CookedCTX))
		for j, l := range p.CookedCTX {
			c := fmt.Sprintf(FOUNDLINE, l.Locus, l.Contents)
			lines[j] = c
		}
		p.LocusBody = strings.Join(lines, "")
		err := tmpl.Execute(&b, p)
		chke(err)
	}

	// ouput

	var out SearchOutputJSON
	out.JS = fmt.Sprintf(BROWSERJS, "browser")
	out.Title = restorewhitespace(thesearch.Seeking)
	out.Image = ""
	out.Searchsummary = formatfinalsearchsummary(thesearch)
	out.Found = b.String()

	if Config.ZapLunates {
		out.Found = DeLunate(out.Found)
	}

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

	const (
		TEMPL = `
		%s
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
		BETW   = "Searched between %s and %s<br>"
		DDM    = "<!-- dates did not matter -->"
		NOCAP  = "<!-- did not hit the results cap -->"
		YESCAP = "[Search suspended: result cap reached.]"
		INFAU  = "<!-- unlimited hits per author -->"
		ONEAU  = `<br><span class="small">(only one hit allowed per author table)</span>`
	)

	m := message.NewPrinter(language.English)
	sess := SafeSessionRead(s.User)
	var dr string
	if sess.Earliest != MINDATESTR || sess.Latest != MAXDATESTR {
		a := FormatBCEDate(sess.Earliest)
		b := FormatBCEDate(sess.Latest)
		dr = fmt.Sprintf(BETW, a, b)
	} else {
		dr = DDM
	}

	var hitcap string
	if len(s.Results) == s.CurrentLimit {
		hitcap = YESCAP
	} else {
		hitcap = NOCAP
	}

	oh := INFAU
	if s.OneHit {
		oh = ONEAU
	}

	var so string

	switch sess.SortHitsBy {
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
	sum := m.Sprintf(TEMPL, s.ExtraMsg, s.InitSum, s.SearchSize, len(s.Results), el, so, oh, dr, hitcap)
	return sum
}

// highlightsearchterm - html markup for the search term in the line so it can jump out at you
func highlightsearchterm(pattern *regexp.Regexp, line *ResultPassageLine) {
	//	regexequivalent is compiled via SearchTermFinder() in rt-search.go

	// see the warnings and caveats at highlightsearchterm() in searchformatting.py
	if pattern.MatchString(line.Contents) {
		line.Contents = pattern.ReplaceAllString(line.Contents, MUREPLACE)
	} else {
		// might be in the hyphenated line
		if pattern.MatchString(line.Hyphenated) {
			// needs more fiddling
			line.Contents += fmt.Sprintf(HYPHREPLACE, line.Hyphenated)
		}
	}
}

// formatinscriptiondates - show the years for inscriptions
func formatinscriptiondates(template string, dbw *DbWorkline) string {
	datestring := ""
	fc := dbw.FindCorpus()
	dated := fc == "in" || fc == "ch" || fc == "dp"
	if dated {
		cd := IntToBCE(AllWorks[dbw.WkUID].ConvDate)
		if cd == "2500 C.E." {
			cd = "??? BCE/CE"
		}
		datestring = fmt.Sprintf(template, strings.Replace(cd, ".", "", -1))
	}
	return datestring
}

// formatinscriptionplaces - show the places for inscriptions
func formatinscriptionplaces(dbw *DbWorkline) string {
	const (
		PLACER = ` [<span class="rust">%s</span>] `
	)

	placestring := ""
	fc := dbw.FindCorpus()
	placed := fc == "in" || fc == "ch" || fc == "dp"
	if placed {
		placestring = fmt.Sprintf(PLACER, AllWorks[dbw.WkUID].Prov)
	}
	return placestring
}

// textblockcleaner - address multi-line formatting challenges by running a suite of clean-ups
func textblockcleaner(html string) string {
	// do it early and in this order
	// presupposes the snippers are in there: "✃✃✃"
	// used by rt-browser and rt-texsindicesandvocab as well
	html = unbalancedspancleaner(html)
	html = formateditorialbrackets(html)
	html = formatmultilinespans(html)

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

	const (
		SPANOPEN  = `<span class="htmlbalancingsupplement">`
		SPANCLOSE = `</span>`
	)

	op := regexp.MustCompile("<span")
	cl := regexp.MustCompile("</span>")

	opened := len(op.FindAllString(html, -1))
	closed := len(cl.FindAllString(html, -1))

	if closed > opened {
		for i := 0; i < closed-opened; i++ {
			html = SPANOPEN + html
		}
	}

	if opened > closed {
		for i := 0; i < opened-closed; i++ {
			html = html + SPANCLOSE
		}
	}
	return html
}

// don't let regex compilation get looped...
var (
	spkr    = regexp.MustCompile("<speaker>\\[(.*?)</speaker>") // really just "[ϲτρ. α." problem in Aeschylus? fix in builder?
	esbboth = regexp.MustCompile("\\[(.*?)]")
	erbboth = regexp.MustCompile("\\((.*?)\\)")
	eabboth = regexp.MustCompile("⟨(.*?)⟩")
	ecbboth = regexp.MustCompile("\\{(.*?)}")
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

	const (
		SPEAK  = `<speaker>$1</speaker>`
		SQUARE = `[<span class="editorialmarker_squarebrackets">$1</span>]`
		ROUND  = `(<span class="editorialmarker_roundbrackets">$1</span>)`
		ANGLE  = `⟨<span class="editorialmarker_angledbrackets">$1</span>⟩`
		CURLY  = `{<span class="editorialmarker_curlybrackets">$1</span>}`
	)

	html = spkr.ReplaceAllString(html, SPEAK)
	html = esbboth.ReplaceAllString(html, SQUARE)
	html = erbboth.ReplaceAllString(html, ROUND)
	html = eabboth.ReplaceAllString(html, ANGLE)
	html = ecbboth.ReplaceAllString(html, CURLY)

	return html
}

func formatmultilinespans(html string) string {
	// good test zone follows; not, though, that the original data seems not to have been marked right
	// that makes seeing whether this code is doing its job a bit tougher...

	// hipparchiaDB=# select index,marked_up_line from gr0535 where index between 328 and 332;
	// index |                                                                                 marked_up_line
	//-------+---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
	//   328 | των οὐκ ἀξιοπίϲτουϲ εἶναι φαϲκόντων, <span class="expanded_text">‘τοιοῦτοϲ γάρ,</span> φηϲί, <span class="expanded_text">καὶ
	//   329 | ὁ τόποϲ ἦν ἐν ᾧ ἡ ὕβριϲ ἐπετελέϲθη· εἰ δὲ ἐν τῷ δου-
	//   330 | ρείῳ ἵππῳ ὑβρίϲθη, τοὺϲ ἀριϲτεῖϲ ἂν ὑμῖν παρει-
	//   331 | χόμην μάρτυραϲ Μενέλαον καὶ Διομήδη καὶ Ὀδυϲϲέα’</span>. <hmu_standalone_endofpage />
	//   332 | &nbsp;&nbsp;&nbsp;<span class="latin normal">MAXIM. CONF. </span><span class="latin italic">l. comm. </span><span class="latin normal">p. 586 Comb. (Migne, PG 91, 828):</span>
	//(5 rows)

	// square bracket run @ antiphon: οὖν τοῦτο καὶ ἐμοὶ γενέϲθω, εἴπερ ἐμοῦ θέλοντοϲ ἔλεγχον

	// this can get too "greedy" in the fragments of the tragedians where lines end "[ " and then the next is not " ]"
	// the irregularities in the original data make this basically insoluble as a problem; but formatmultilinespans()
	// in this form probably gets more things right than wrong; contrast [defunct] formatmultilinebrackets() which
	// prevented a lot of spillage but not all, and mostly because it was so naive

	const (
		SPLT = "✃✃✃"
	)

	type spantype struct {
		open  string
		close string
	}

	st1 := spantype{"<span class=\"expanded_text\">", "</span>"}
	st2 := spantype{"<hmu_serviusformatting>", "</hmu_serviusformatting>"}
	st3 := spantype{"<span class=\"editorialmarker_squarebrackets\">", "</span>"}
	st4 := spantype{"<span class=\"editorialmarker_roundbrackets\">", "</span>"}
	st5 := spantype{"<span class=\"editorialmarker_angledbrackets\">", "</span>"}
	st6 := spantype{"<span class=\"editorialmarker_curlybrackets\">", "</span>"}
	// st7 := spantype{"<span class=\"bold\">", "</span>"} // gr4089 has span problems via the build itself

	tocheck := []spantype{st1, st2, st3, st4, st5, st6}

	spanner := func(lines []string, st spantype) []string {
		add := ""
		newlines := make([]string, len(lines))
		for i, l := range lines {
			l = add + l
			back := strings.Split(l, st.open)
			if len(back) > 1 {
				if strings.Contains(back[len(back)-1], st.close) {
					add = ""
				} else {
					add = st.open
					l = l + st.close
				}
			}
			newlines[i] = l
		}
		return newlines
	}

	htmlslc := strings.Split(html, SPLT)
	for _, c := range tocheck {
		if strings.Contains(html, c.open) {
			htmlslc = spanner(htmlslc, c)
		}
	}
	html = strings.Join(htmlslc, SPLT)
	return html
}

// gethighlighter - set regex to highlight the search term
func gethighlighter(ss *SearchStruct) *regexp.Regexp {
	// "s", "sp", "spa", ... will mean html gets highlighting: `<span class="xyz" ...>`
	// there has to be a more clever way to do this...
	const (
		FAIL   = "gethighlighter() cannot find anything to highlight\n\t%ss"
		FAILRE = "MATCH_NOTHING"
		SKIP1  = "^s$|^sp$|^spa$|^span$|^hmu$"
		SKIP2  = "|^c$|^cl$|^cla$|^clas$|^class$"
		SKIP3  = "|^a$|^as$|^ass$"
		SKIP4  = "|^l$|^la$|^lat$|^lati$|^latin$"
		SKIP   = SKIP1 + SKIP2 + SKIP3 + SKIP4
	)

	var re *regexp.Regexp

	skg := ss.Seeking
	prx := ss.Proximate

	skip := regexp.MustCompile(SKIP)
	if skip.MatchString(skg) || skip.MatchString(prx) {
		return regexp.MustCompile(FAILRE)
	}

	if ss.SkgRewritten {
		// quasi-bugged because of "\s" --> "\[sS]"; meanwhile whitespacer() can't use " " for its own reasons...
		// ((^|\[sS])[εἐἑἒἓἔἕὲέἘἙἚἛἜἝΕ][νΝ] [οὀὁὂὃὄὅόὸὈὉὊὋὌὍΟ][ρῤῥῬ][εἐἑἒἓἔἕὲέἘἙἚἛἜἝΕ][ϲσΣςϹ][τΤ][ηᾐᾑᾒᾓᾔᾕᾖᾗῂῃῄῆῇἤἢἥἣὴήἠἡἦἧᾘᾙᾚᾛᾜᾝᾞᾟἨἩἪἫἬἭἮἯΗ](\[sS]|$))
		skg = strings.Replace(whitespacer(skg, ss), "(^|\\s)", "(^| )", 1)
		skg = strings.Replace(whitespacer(skg, ss), "(\\s|$)", "( |$)", 1)
		prx = strings.Replace(whitespacer(prx, ss), "(^|\\s)", "(^| )", 1)
		prx = strings.Replace(whitespacer(prx, ss), "(\\s|$)", "( |$)", 1)
	}

	if len(ss.Seeking) != 0 {
		re = SearchTermFinder(skg)
	} else if len(ss.LemmaOne) != 0 {
		re = lemmahighlighter(ss.LemmaOne)
	} else if len(ss.Proximate) != 0 {
		re = SearchTermFinder(prx)
	} else if len(ss.LemmaTwo) != 0 {
		re = lemmahighlighter(ss.LemmaTwo)
	} else {
		msg(fmt.Sprintf(FAIL, ss.InitSum), MSGFYI)
		re = regexp.MustCompile(FAILRE)
	}
	return re
}

// lemmahighlighter - set regex to highlight a lemmatized search term
func lemmahighlighter(lm string) *regexp.Regexp {
	// don't let "(^|\s)τρεῖϲ(\s|$)|(^|\s)τρία(\s|$)|(^|\s)τριϲίν(\s|$)|(^|\s)τριῶν(\s|$)|(^|\s)τρί(\s|$)|(^|\s)τριϲί(\s|$)"
	// turn into "(^|\[sS])[τΤ][ρῤῥῬ][εἐἑἒἓἔἕὲέἘἙἚἛἜἝΕ]ῖ[ϲσΣςϹ](\[sS]|$)|(^|\[sS])..."
	// can't send "(^|\s)" through UniversalPatternMaker()

	// abutting markup is killing off some items, but adding "<" and ">" produces worse problems still

	// now you also need to worry about punctuation that abuts the find
	// tp := `[\^\s;]%s[\s\.,;·’$]`

	const (
		FAIL   = "lemmahighlighter() could not compile lemma into regex"
		JOINER = ")✃✃✃("
		SNIP   = "✃✃✃"
		TP     = `%s` // move from match $1 to $0 in highlightsearchterm() yielded this shift...
	)

	lemm := AllLemm[lm].Deriv

	whole := strings.Join(lemm, JOINER)
	st := UniversalPatternMaker(whole)
	lup := strings.Split(st, SNIP)
	for i, l := range lup {
		lup[i] = fmt.Sprintf(TP, l)
	}
	rec := strings.Join(lup, "|")

	r, e := regexp.Compile(rec)
	if e != nil {
		msg(FAIL, MSGFYI)
	}
	return r
}
