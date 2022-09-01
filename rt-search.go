package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type SearchStruct struct {
	User       string
	ID         string
	Seeking    string
	Proximate  string
	LemmaOne   string
	LemmaTwo   string
	InitSum    string
	Summary    string
	ProxScope  string // "lines" or "words"
	ProxType   string // "near" or "not near"
	ProxVal    int64
	HasLemma   bool
	HasPhrase  bool
	IsVector   bool
	Twobox     bool
	NotNear    bool
	SrchColumn string // usually "stripped_line", sometimes "accented_line"
	SrchSyntax string // almost always "~="
	OrderBy    string // almost always "index" + ASC
	Limit      int64
	SkgSlice   []string // either just Seeking or a decomposed version of a Lemma's possibilities
	PrxSlice   []string
	SearchIn   SearchIncExl
	SearchEx   SearchIncExl
	Queries    []PrerolledQuery
	Results    []DbWorkline
	Launched   time.Time
	TTName     string
	SearchSize int // # of works
}

func (s SearchStruct) FmtOrderBy() string {
	var ob string
	a := `ORDER BY %s ASC %s`
	b := `LIMIT %d`
	if s.Limit > 0 {
		c := fmt.Sprintf(b, s.Limit)
		ob = fmt.Sprintf(a, s.OrderBy, c)
	} else {
		ob = fmt.Sprintf(a, s.OrderBy, "")
	}
	return ob
}

//
// ROUTING
//

func RtSearchConfirm(c echo.Context) error {
	// not going to be needed?
	// "test the activity of a poll so you don't start conjuring a bunch of key errors if you use wscheckpoll() prematurely"
	return c.String(http.StatusOK, "")
}

func RtSearchStandard(c echo.Context) error {
	start := time.Now()
	previous := time.Now()
	// "GET /search/standard/5446b840?skg=sine%20dolore HTTP/1.1"
	// "GET /search/standard/c2fba8e8?skg=%20dolore&prx=manif HTTP/1.1"
	// "GET /search/standard/2ad866e2?prx=manif&lem=dolor HTTP/1.1"
	// "GET /search/standard/02f3610f?lem=dolor&plm=manifesta HTTP/1.1"

	user := readUUIDCookie(c)

	id := c.Param("id")
	skg := c.QueryParam("skg")
	prx := c.QueryParam("prx")
	lem := c.QueryParam("lem")
	plm := c.QueryParam("plm")

	srch := builddefaultsearch(c)

	// HasPhrase makes us use a fake limit temporarily
	reallimit := srch.Limit

	timetracker("A", "builddefaultsearch()", start, previous)
	previous = time.Now()

	srch.Seeking = skg
	srch.Proximate = prx
	srch.LemmaOne = lem
	srch.LemmaTwo = plm
	srch.User = user
	srch.ID = id
	srch.IsVector = false

	parsesearchinput(&srch)

	// must happen before searchlistintoqueries()
	setsearchtype(&srch)

	srch.InitSum = formatinitialsummary(srch)

	sl := sessionintosearchlist(sessions[user])
	srch.SearchIn = sl.Inc
	srch.SearchEx = sl.Excl
	srch.SearchSize = sl.Size

	timetracker("B", "sessionintosearchlist()", start, previous)
	previous = time.Now()

	prq := searchlistintoqueries(&srch)
	timetracker("C", "searchlistintoqueries()", start, previous)
	previous = time.Now()

	srch.Queries = prq
	searches[id] = srch

	// return results via searches[id].Results
	if searches[id].Twobox {
		// todo: triple-check results against python
		// todo: "not near" syntax
		searches[id] = withinxlinessearch(searches[id])
	} else {
		searches[id] = HGoSrch(searches[id])
	}

	if searches[id].HasPhrase {
		// you did HGoSrch() and need to check the windowed lines
		// withinxlinessearch() has already done the checking
		// the cannot assign problem...
		mod := searches[id]
		mod.Results = findphrasesacrosslines(searches[id])
		if int64(len(mod.Results)) > reallimit {
			mod.Results = mod.Results[0:reallimit]
		}
		searches[id] = mod
	}

	previous = time.Now()

	//hits := searches[id].Results
	//for i, h := range hits {
	//	t := fmt.Sprintf("%d - %srch : %srch", i, h.FindLocus(), h.MarkedUp)
	//	fmt.Println(t)
	//}

	timetracker("D", fmt.Sprintf("search executed: %d hits", len(searches[id].Results)), start, previous)

	var js string
	if sessions[readUUIDCookie(c)].HitContext == 0 {
		js = string(formatnocontextresults(searches[id]))
	} else {
		js = string(formatwithcontextresults(searches[id]))
	}

	srchsumm[id] = SearchSummary{start, searches[id].Summary}
	msg(fmt.Sprintf("search count is %d", len(srchsumm)), 5)

	delete(searches, id)

	return c.String(http.StatusOK, js)
}

//
// TWO-PART SEARCHES
//

func withinxlinessearch(originalsrch SearchStruct) SearchStruct {
	// after finding x, look for y within n lines of x
	// can do:
	// [1] single + single
	// [2] lemma + single
	// [3] lemma + lemma
	// [4] phrase + single
	// [5] phrase + lemma
	// [6] phrase + phrase

	// (part 1)
	//		HGoSrch(first)
	//
	// (part 2)
	// 		populate a new search list with a ton of passages via the first results
	//		HGoSrch(second)

	previous := time.Now()
	first := originalsrch

	first.Limit = FIRSTSEARCHLIM
	first = HGoSrch(first)
	d := fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf("%s withinxlinessearch(): %d initial hits", d, len(first.Results)), 4)
	previous = time.Now()

	if first.HasPhrase {
		mod := first
		// this will cut the hits by c. 50%
		mod.Results = findphrasesacrosslines(first)
		first = mod
	}
	d = fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf("%s findphrasesacrosslines(): %d trimmed hits", d, len(first.Results)), 4)
	previous = time.Now()

	second := first
	second.Results = []DbWorkline{}
	second.Queries = []PrerolledQuery{}
	second.SearchIn = SearchIncExl{}
	second.SearchEx = SearchIncExl{}
	second.TTName = strings.Replace(uuid.New().String(), "-", "", -1)
	second.SkgSlice = []string{}
	second.Seeking = second.Proximate
	second.LemmaOne = second.LemmaTwo
	second.Proximate = ""
	second.PrxSlice = []string{}
	second.LemmaTwo = ""

	setsearchtype(&second)

	pt := `%s_FROM_%d_TO_%d`

	var newpsg []string
	for _, r := range first.Results {
		np := fmt.Sprintf(pt, r.FindAuthor(), r.TbIndex-first.ProxVal, r.TbIndex+first.ProxVal)
		newpsg = append(newpsg, np)
	}

	// todo: not near logic

	second.Limit = originalsrch.Limit
	second.SearchIn.Passages = newpsg

	prq := searchlistintoqueries(&second)

	d = fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf("%s searchlistintoqueries() rerun", d), 4)
	previous = time.Now()

	second.Queries = prq
	searches[originalsrch.ID] = HGoSrch(second)

	d = fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf("%s withinxlinessearch(): %d subsequent hits", d, len(first.Results)), 4)

	// findphrasesacrosslines() check happens just after you exit this function

	return searches[originalsrch.ID]
}

//
// SETUP
//

// builddefaultsearch - fill out the basic values for a new search
func builddefaultsearch(c echo.Context) SearchStruct {
	user := readUUIDCookie(c)

	var s SearchStruct
	s.User = user
	s.Launched = time.Now()
	s.Limit = sessions[user].HitLimit
	s.SrchColumn = DEFAULTCOLUMN
	s.SrchSyntax = DEFAULTSYNTAX
	s.OrderBy = ORDERBY
	s.SearchIn = sessions[user].Inclusions
	s.SearchEx = sessions[user].Exclusions
	s.ProxVal = DEFAULTPROXIMITY
	s.ProxScope = DEFAULTPROXIMITYSCOPE
	s.NotNear = false
	s.Twobox = false
	s.HasPhrase = false
	s.HasLemma = false
	s.TTName = strings.Replace(uuid.New().String(), "-", "", -1)
	return s
}

func parsesearchinput(s *SearchStruct) {
	// remove bad chars
	// address uv issues; lunate issues; ...
	// no need to parse a lemma: this bounces if there is not a key match to a map

	s.Seeking = strings.ToLower(s.Seeking)
	s.Proximate = strings.ToLower(s.Proximate)

	if hasAccent.MatchString(s.Seeking) || hasAccent.MatchString(s.Proximate) {
		// lemma search will select accented automatically
		s.SrchColumn = "accented_line"
	}

	rs := []rune(s.Seeking)
	if len(rs) > MAXINPUTLEN {
		s.Seeking = string(rs[0:MAXINPUTLEN])
	}

	rp := []rune(s.Proximate)
	if len(rp) > MAXINPUTLEN {
		s.Proximate = string(rs[0:MAXINPUTLEN])
	}

	s.Seeking = uvσçϲ(s.Seeking)
	s.Proximate = uvσçϲ(s.Proximate)

	s.Seeking = purgechars(UNACCEPTABLEINPUT, s.Seeking)
	s.Proximate = purgechars(UNACCEPTABLEINPUT, s.Proximate)

}

func setsearchtype(srch *SearchStruct) {
	// skip detailed proximate checks because second pass search just feeds all of that into the primary fields

	ps := srch.Proximate != ""
	psl := srch.LemmaTwo != ""

	if ps || psl {
		srch.Twobox = true
	}

	// will not find greek...
	// pattern := regexp.MustCompile(`\w\s\w`)

	pattern := regexp.MustCompile(`[A-Za-zΑ-ΩϹα-ωϲ]\s[A-Za-zΑ-ΩϹα-ωϲ]`)

	if pattern.MatchString(srch.Seeking) {
		srch.HasPhrase = true
	}

	if len(srch.LemmaOne) != 0 {
		srch.HasLemma = true
		srch.SrchColumn = "accented_line"
	}

	return
}

//
// HELPERS
//

func lemmaintoregexslice(hdwd string) []string {
	// rather than do one word per query, bundle things up: some words have >100 forms
	// ...(^|\\s)ἐδηλώϲαντο(\\s|$)|(^|\\s)δεδηλωμένοϲ(\\s|$)|(^|\\s)δήλουϲ(\\s|$)|(^|\\s)δηλούϲαϲ(\\s|$)...
	var qq []string
	if _, ok := AllLemm[hdwd]; !ok {
		msg(fmt.Sprintf("lemmaintoregexslice() could not find '%s'", hdwd), 1)
		return qq
	}

	tp := `(^|\s)%s(\s|$)`
	lemm := AllLemm[hdwd].Deriv
	ct := 0
	for true {
		var bnd []string
		for i := 0; i < MAXLEMMACHUNKSIZE; i++ {
			if ct > len(lemm)-1 {
				//re := fmt.Sprintf(tp, lemm[ct])
				//bnd = append(bnd, re)
				//qq = append(qq, strings.Join(bnd, "|"))
				break
			}
			re := fmt.Sprintf(tp, lemm[ct])
			bnd = append(bnd, re)
			ct += 1
		}
		qq = append(qq, strings.Join(bnd, "|"))
		if ct >= len(lemm)-1 {
			break
		}
	}
	return qq
}

func findphrasesacrosslines(ss SearchStruct) []DbWorkline {
	// "one two$" + "^three four" makes a hit if you want "one two three four"
	// super slow...:
	// [HGS] [Δ: 1.474s]  withinxlinessearch(): 1631 initial hits
	// [HGS] [Δ: 7.433s]  findphrasesacrosslines(): 855 trimmed hits

	var valid = make(map[string]DbWorkline)

	find := regexp.MustCompile(`^\s`)
	re := find.ReplaceAllString(ss.Seeking, "(^|\\s)")
	find = regexp.MustCompile(`\s$`)
	re = find.ReplaceAllString(ss.Seeking, "(\\s|$)")

	for i, r := range ss.Results {
		// do the "it's all on this line" case separately
		li := columnpicker(ss.SrchColumn, r)
		fp := regexp.MustCompile(re)
		f := fp.MatchString(li)
		if f {
			valid[r.BuildHyperlink()] = r
		} else {
			// msg("'else'", 4)
			var nxt DbWorkline
			if i+1 < len(ss.Results) {
				nxt = ss.Results[i+1]
				if r.TbIndex+1 > AllWorks[r.WkUID].LastLine {
					nxt = DbWorkline{}
				} else if r.WkUID != nxt.WkUID || r.TbIndex+1 != nxt.TbIndex {
					// grab the actual next line (i.e. index = 101)
					nxt = graboneline(r.FindAuthor(), r.TbIndex+1)
				}
			} else {
				// grab the actual next line (i.e. index = 101)
				nxt = graboneline(r.FindAuthor(), r.TbIndex+1)
				if r.WkUID != nxt.WkUID {
					nxt = DbWorkline{}
				}
			}

			// combinator dodges double-register of hits
			nl := columnpicker(ss.SrchColumn, nxt)
			comb := phrasecombinations(re)
			for _, c := range comb {
				fp = regexp.MustCompile(c[0])
				sp := regexp.MustCompile(c[1])
				f = fp.MatchString(li)
				s := sp.MatchString(nl)
				if f && s && r.WkUID == nxt.WkUID {
					valid[r.BuildHyperlink()] = r
				}
			}
		}
	}

	var slc []DbWorkline
	for _, r := range valid {
		slc = append(slc, r)
	}
	slc = sortresults(slc, ss)
	return slc
}

func columnpicker(c string, r DbWorkline) string {
	var li string
	switch c {
	case "stripped_line":
		li = r.Stripped
	case "accented_line":
		li = r.Accented
	case "marked_up_line": // only a maniac tries to search via marked_up_line
		li = r.MarkedUp
	default:
		li = r.Stripped
		msg("second.SrchColumn was not set; defaulting to 'stripped_line'", 2)
	}
	return li
}

func phrasecombinations(phr string) [][2]string {
	// 'one two three four five' -->
	// [('one', 'two three four five'), ('one two', 'three four five'), ('one two three', 'four five'), ('one two three four', 'five')]

	gt := func(n int, wds []string) []string {
		return wds[n:]
	}

	gh := func(n int, wds []string) []string {
		return wds[:n]
	}

	ww := strings.Split(phr, " ")
	var comb [][2]string
	for i, _ := range ww {
		h := strings.Join(gh(i, ww), " ")
		t := strings.Join(gt(i, ww), " ")
		h = h + "$"
		t = "^" + t
		comb = append(comb, [2]string{h, t})
	}

	var trimmed [][2]string
	for _, c := range comb {
		head := strings.TrimSpace(c[0]) != "" && strings.TrimSpace(c[0]) != "$"
		tail := strings.TrimSpace(c[1]) != "" && strings.TrimSpace(c[0]) != "^"
		if head && tail {
			trimmed = append(trimmed, c)
		}
	}

	//for i, c := range trimmed {
	//	fmt.Printf("%d:\n\t0: %s\n\t1: %s\n", i, c[0], c[1])
	//}

	return trimmed
}

//
// FORMATTING
//

type SearchOutputJSON struct {
	Title         string `json:"title"`
	Searchsummary string `json:"searchsummary"`
	Found         string `json:"found"`
	Image         string `json:"image"`
	JS            string `json:"js"`
}

func formatnocontextresults(s SearchStruct) []byte {
	var out SearchOutputJSON
	out.JS = BROWSERJS
	out.Title = s.Seeking
	out.Image = ""
	out.Searchsummary = formatfinalsearchsummary(&s)

	TABLEROW := `
	<tr class="%s">
		<td>
			<span class="findnumber">[%d]</span>&nbsp;&nbsp;
			<span class="foundauthor">%s</span>,&nbsp;<span class="foundwork">%s</span>:
			<browser id="%s"><span class="foundlocus">%s</span></browser>
		</td>
		<td class="leftpad">
			<span class="foundtext">%s</span>
		</td>
	</tr>
	`

	var rows []string
	for i, r := range s.Results {
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
		fm := fmt.Sprintf(TABLEROW, rc, i+1, au, wk, lk, lc, r.MarkedUp)
		rows = append(rows, fm)
	}

	out.Found = "<tbody>" + strings.Join(rows, "") + "</tbody>"

	js, e := json.Marshal(out)
	chke(e)

	return js
}

func formatwithcontextresults(ss SearchStruct) []byte {
	// things to worry about: formateditorialbrackets(); unbalancedspancleaner()

	// unbalancedspancleaner() has to be run on the first line & after the whole block has been built

	// how/when to do <span class="highlight">

	thesession := sessions[ss.User]

	type ResultPassageLine struct {
		Locus           string
		ShowLocus       bool
		Contents        string
		Hyphenated      string
		ContinuingStyle string
		IsHighlight     bool
	}

	type PsgFormattingTemplate struct {
		Findnumber  int
		Foundauthor string
		Foundwork   string
		FindURL     string
		FindLocus   string
		RawCTX      []DbWorkline
		CookedCTX   []ResultPassageLine
		LocusBody   string
	}

	// iterate over the results to build the raw core data
	var allpassages []PsgFormattingTemplate
	for i, r := range ss.Results {
		var psg PsgFormattingTemplate
		psg.Findnumber = i + 1
		psg.Foundauthor = AllAuthors[r.FindAuthor()].Name
		psg.Foundwork = AllWorks[r.WkUID].Title
		psg.FindURL = r.BuildHyperlink()
		psg.FindLocus = basiccitation(AllWorks[r.FindAuthor()], r)
		psg.RawCTX = simplecontextgrabber(r.FindAuthor(), r.TbIndex, int64(thesession.HitContext/2))

		for j := 0; j < len(psg.RawCTX); j++ {
			c := ResultPassageLine{}
			c.Locus = strings.Join(psg.RawCTX[j].FindLocus(), ".")

			if psg.RawCTX[j].BuildHyperlink() == psg.FindURL {
				c.IsHighlight = true
			} else {
				c.IsHighlight = false
			}
			c.Contents = psg.RawCTX[j].MarkedUp
			c.Hyphenated = psg.RawCTX[j].Hypenated
			psg.CookedCTX = append(psg.CookedCTX, c)
		}
		allpassages = append(allpassages, psg)
	}

	// fix the unmattched spans
	for _, p := range allpassages {
		// at the top
		p.CookedCTX[0].Contents = unbalancedspancleaner(p.CookedCTX[0].Contents)

		// across the whole
		var block []string
		for _, c := range p.CookedCTX {
			block = append(block, c.Contents)
		}
		whole := strings.Join(block, "✃✃✃")
		whole = unbalancedspancleaner(whole)

		// reassemble
		block = strings.Split(whole, "✃✃✃")
		for i, b := range block {
			p.CookedCTX[i].Contents = b
		}
	}

	// highlight the search term: this includes the hyphenated_line issue

	// search for brackets
	// TODO: it's fiddly

	// search for span inheretance
	// TODO: it's fiddly

	// decide which lines need to display their citation info
	// TODO: set to "all on" atm
	for _, p := range allpassages {
		for _, c := range p.CookedCTX {
			c.ShowLocus = true
		}
	}

	// build a passage bundle

	// aggregate the bundles

	pht := `
	<locus>
		<span class="findnumber">[{{.Findnumber}}]</span>&nbsp;&nbsp;
		<span class="foundauthor">{{.Foundauthor}}</span>,&nbsp;<span class="foundwork">{{.Foundwork}}</span>:
		<browser id="{{.FindURL}}"><span class="foundlocus">{{.FindLocus}}</span></browser>
	</locus>
	{{.LocusBody}}`

	tmpl, e := template.New("tr").Parse(pht)
	chke(e)

	plt := `<span class="locus">%s</span>&nbsp;<span class="foundtext">%s</span><br>
	`

	var rows []string
	for _, p := range allpassages {
		var lines []string
		for _, l := range p.CookedCTX {
			c := fmt.Sprintf(plt, l.Locus, l.Contents)
			lines = append(lines, c)
		}
		p.LocusBody = strings.Join(lines, "")
		var b bytes.Buffer
		err := tmpl.Execute(&b, p)
		chke(err)

		// fmt.Println(b.String())
		rows = append(rows, b.String())
	}

	// ouput

	var out SearchOutputJSON
	out.JS = BROWSERJS
	out.Title = ss.Seeking
	out.Image = ""
	out.Searchsummary = formatfinalsearchsummary(&ss)
	out.Found = strings.Join(rows, "")

	js, e := json.Marshal(out)
	chke(e)

	return js
}

func formatfinalsearchsummary(s *SearchStruct) string {

	t := `
	<div id="searchsummary">
		%s
		<br>
		Searched %d works and found %d passages (%ss)
		<br>
		<!-- unlimited hits per author -->
		Sorted by %s
		<br>
		%s
		%s
	</div>
	`

	var dr string
	if sessions[s.User].Inclusions.DateRange != [2]string{"-850", "1500"} {
		a := formatbcedate(sessions[s.User].Inclusions.DateRange[0])
		b := formatbcedate(sessions[s.User].Inclusions.DateRange[1])
		dr = fmt.Sprintf("<br>Searched between %s and %s", a, b)
	} else {
		dr = "<!-- dates did not matter -->"
	}

	var hitcap string
	if int64(len(s.Results)) == s.Limit {
		hitcap = "[Search suspended: result cap reached.]"
	} else {
		hitcap = "<!-- did not hit the results cap -->"
	}

	so := sessions[s.User].SortHitsBy
	el := fmt.Sprintf("%.3f", time.Now().Sub(s.Launched).Seconds())
	// need to record # of works and not # of tables somewhere & at the right moment...
	sum := fmt.Sprintf(t, s.InitSum, s.SearchSize, len(s.Results), el, so, dr, hitcap)
	return sum
}

func formatinitialsummary(s SearchStruct) string {
	tmp := `Sought %s<span class="sought">»%s«</span>%s`
	win := ` within %d %s of %s<span class="sought">»%s«</span>`

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
		two = fmt.Sprintf(win, s.ProxVal, s.ProxScope, af2, sk2)
	}
	sum := fmt.Sprintf(tmp, af1, sk, two)

	return sum
}

func formatbcedate(d string) string {
	s, e := strconv.Atoi(d)
	if e != nil {
		s = 9999
	}
	if s > 0 {
		d += " C.E."
	} else {
		d = strings.Replace(d, "-", "", -1) + " B.C.E."
	}
	return d
}

func formateditorialbrackets(html string) string {
	// sample:
	// [<span class="editorialmarker_squarebrackets">ἔδοχϲεν τε͂ι βολε͂ι καὶ το͂ι</span>]

	// special cases:
	// [a] no "open" or "close" bracket at the head/tail of a line: ^τε͂ι βολε͂ι καὶ] το͂ι...$ / ^...ἔδοχϲεν τε͂ι βολε͂ι [καὶ το͂ι$
	// [b] we are continuing from a previous state: no brackets here, but should insert a span; the previous line will need to notify the subsequent...

	// types: editorialmarker_angledbrackets; editorialmarker_curlybrackets, editorialmarker_roundbrackets, editorialmarker_squarebrackets
	//

	return ""
}

func highlightsearchterm(html, ss *SearchStruct) string {
	// 	html markup for the search term in the line so it can jump out at you
	//
	//	regexequivalent is compiled via compilesearchtermequivalent()
	//
	//	in order to properly highlight a polytonic word that you found via a unaccented search you need to convert:
	//		ποταμον
	//	into:
	//		([πΠ][οὀὁὂὃὄὅόὸΟὈὉὊὋὌὍ][τΤ][αἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάᾈᾉᾊᾋᾌᾍᾎᾏἈἉἊἋἌἍἎἏΑ][μΜ][οὀὁὂὃὄὅόὸΟὈὉὊὋὌὍ][νΝ])
	return ""
}

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

/*

the following yields a strange problem: "&nbsp;" will render literally rather than as a space in the output. why?
templating makes the formatting code a lot more readable...

func formatnocontextresults(s SearchStruct) []byte {
	var out SearchOutputJSON
	out.JS = BROWSERJS
	out.Title = s.Seeking
	out.Image = ""
	out.Searchsummary = formatfinalsearchsummary(s)

	type TR struct {
		RC string
		NU int
		AU string
		WK string
		LK string
		LC string
		MU string
	}

	TABLEROW := `
	<tr class="{{.RC}}">
		<td>
			<span class="findnumber">[{{.NU}}]</span>&nbsp;&nbsp;
			<span class="foundauthor">{{.AU}}</span>,&nbsp;<span class="foundwork">{{.WK}}</span>:
			<browser id="{{.LK}}"><span class="foundlocus">{{.LC}}</span></browser>
		</td>
		<td class="leftpad">
			<span class="foundtext">{{.MU}}</span>
		</td>
	</tr>
	`

	tmpl, e := template.New("tr").Parse(TABLEROW)
	chke(e)

	var rows []string
	for i, r := range s.Results {
		rc := ""
		if i%3 == 2 {
			rc = "nthrow"
		} else {
			rc = "regular"
		}

		var tr TR
		tr.RC = rc
		tr.AU = AllAuthors[r.FindAuthor()].Shortname
		tr.WK = AllWorks[r.WkUID].Title
		tr.LK = r.BuildHyperlink()
		tr.LC = strings.Join(r.FindLocus(), ".")
		tr.MU = r.MarkedUp

		var b bytes.Buffer
		err := tmpl.Execute(&b, tr)
		chke(err)

		fmt.Println(b.String())
		rows = append(rows, b.String())
	}

	out.Found = "<tbody>" + strings.Join(rows, "") + "</tbody>"

	js, e := json.Marshal(out)
	chke(e)

	return js
}

*/
