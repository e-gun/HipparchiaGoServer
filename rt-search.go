//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"regexp"
	"strings"
	"text/template"
	"time"
)

var (
	// regex compiled here instead of inside of various loops
	isGreek   = regexp.MustCompile("[α-ωϲῥἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάἐἑἒἓἔἕὲέἰἱἲἳἴἵἶἷὶίῐῑῒΐῖῗὀὁὂὃὄὅόὸὐὑὒὓὔὕὖὗϋῠῡῢΰῦῧύὺᾐᾑᾒᾓᾔᾕᾖᾗῂῃῄῆῇἤἢἥἣὴήἠἡἦἧὠὡὢὣὤὥὦὧᾠᾡᾢᾣᾤᾥᾦᾧῲῳῴῶῷώὼ]")
	hasAccent = regexp.MustCompile("[äëïöüâêîôûàèìòùáéíóúᾂᾒᾢᾃᾓᾣᾄᾔᾤᾅᾕᾥᾆᾖᾦᾇᾗᾧἂἒἲὂὒἢὢἃἓἳὃὓἣὣἄἔἴὄὔἤὤἅἕἵὅὕἥὥἆἶὖἦὦἇἷὗἧὧᾲῂῲᾴῄῴᾷῇῷᾀᾐᾠᾁᾑᾡῒῢΐΰῧἀἐἰὀὐἠὠῤἁἑἱὁὑἡὡῥὰὲὶὸὺὴὼάέίόύήώᾶῖῦῆῶϊϋ]")
	esbboth   = regexp.MustCompile("\\[(.*?)\\]")
	erbboth   = regexp.MustCompile("\\((.*?)\\)")
	eabboth   = regexp.MustCompile("⟨(.*?)⟩")
	ecbboth   = regexp.MustCompile("\\{(.*?)\\}")
	// esbopens := regexp.MustCompile("\\[(.*?)(\\]|$)")
	// esbcloses := regexp.MustCompile("(^|\\[)(.*?)\\]")
	// erbopens := regexp.MustCompile("\\((.*?)(\\)|$)")
	// erbcloses := regexp.MustCompile("(^|\\()(.*?)\\)")
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
	IsActive   bool
	Twobox     bool
	NotNear    bool
	PhaseNum   int
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
	SearchSize int // # of works searched
	TableSize  int // # of tables searched
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

	return c.String(http.StatusOK, "8000")
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
	srch.TableSize = len(prq)

	timetracker("C", "searchlistintoqueries()", start, previous)
	previous = time.Now()

	srch.Queries = prq
	srch.IsActive = true
	searches[id] = srch

	// return results via searches[id].Results
	if searches[id].Twobox {
		// todo: triple-check results against python
		// todo: "not near" syntax
		if searches[id].ProxScope == "words" {
			searches[id] = withinxwordssearch(searches[id])
		} else {
			searches[id] = withinxlinessearch(searches[id])
		}
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

	timetracker("D", fmt.Sprintf("search executed: %d hits", len(searches[id].Results)), start, previous)
	previous = time.Now()

	var js string
	if sessions[readUUIDCookie(c)].HitContext == 0 {
		js = string(formatnocontextresults(searches[id]))
	} else {
		js = string(formatwithcontextresults(searches[id]))
	}

	timetracker("E", fmt.Sprintf("formatted %d hits", len(searches[id].Results)), start, previous)
	previous = time.Now()

	srchsumm[id] = SearchSummary{start, searches[id].Summary}
	msg(fmt.Sprintf("search count is %d", len(srchsumm)), 5)

	msg(fmt.Sprintf(`RtSearchStandard(): deleting searches["%s"]`, id), 5)
	delete(searches, id)

	return c.String(http.StatusOK, js)
}

//
// TWO-PART SEARCHES
//

// withinxlinessearch - find A within N lines of B
func withinxlinessearch(originalsrch SearchStruct) SearchStruct {
	// after finding A, look for B within N lines of A
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
	first := generateinitialhits(originalsrch)

	d := fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf("%s withinxlinessearch(): %d initial hits", d, len(first.Results)), 4)
	previous = time.Now()

	second := clonesearch(first, 2)
	second.Seeking = second.Proximate
	second.LemmaOne = second.LemmaTwo
	second.Proximate = first.Seeking
	second.LemmaTwo = first.LemmaOne

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

// withinxwordssearch - find A within N words of B
func withinxwordssearch(originalsrch SearchStruct) SearchStruct {
	// todo: not near logic

	previous := time.Now()
	first := generateinitialhits(originalsrch)

	d := fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf("%s withinxwordssearch(): %d initial hits", d, len(first.Results)), 4)
	previous = time.Now()

	// the trick is we are going to grab all lines near the initial hit; then build strings; then search those strings ourselves
	// so the second search is "anything nearby"

	// [a] build the second search
	second := clonesearch(first, 2)
	sskg := second.Proximate
	slem := second.LemmaTwo
	second.Seeking = ""
	second.LemmaOne = ""
	second.Proximate = first.Seeking
	second.LemmaTwo = first.LemmaOne

	setsearchtype(&second)

	// [a1] hard code a suspect assumption...
	AVERAGEWRDSPERLINE := 8
	need := 2 + (first.ProxVal / int64(AVERAGEWRDSPERLINE))

	pt := `%s_FROM_%d_TO_%d`
	t := `linenumber/%s/%s/%d`
	resultmapper := make(map[string]int)
	var newpsg []string

	// [a2] pick the lines to grab and associate them with the hits they go with
	for i, r := range first.Results {
		np := fmt.Sprintf(pt, r.FindAuthor(), r.TbIndex-need, r.TbIndex+need)
		newpsg = append(newpsg, np)
		for j := r.TbIndex - need; j <= r.TbIndex+need; j++ {
			m := fmt.Sprintf(t, r.FindAuthor(), r.FindWork(), j)
			resultmapper[m] = i
		}
	}

	second.SearchIn.Passages = newpsg
	prq := searchlistintoqueries(&second)

	// [b] run the second "search"
	second.Queries = prq
	searches[originalsrch.ID] = HGoSrch(second)

	d = fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf("%s withinxwordssearch(): %d subsequent hits", d, len(first.Results)), 4)
	previous = time.Now()

	// [c] convert these finds into strings and then search those strings
	// [c1] build bundles of lines
	bundlemapper := make(map[int][]DbWorkline)
	for _, r := range searches[originalsrch.ID].Results {
		url := r.BuildHyperlink()
		bun := resultmapper[url]
		bundlemapper[bun] = append(bundlemapper[bun], r)
	}

	// [c2] decompose them into long strings
	stringmapper := make(map[int]string)
	for idx, lines := range bundlemapper {
		var bundle []string
		for _, l := range lines {
			bundle = append(bundle, columnpicker(first.SrchColumn, l))
		}
		stringmapper[idx] = strings.Join(bundle, " ")
	}

	// [c3] grab the head and tail of each
	var re string
	if len(first.LemmaOne) != 0 {
		re = strings.Join(lemmaintoregexslice(first.LemmaOne), "|")
	} else {
		re = first.Seeking
	}

	rt := `^(?P<head>.*?)%s(?P<tail>.*?)$`

	patternone, e := regexp.Compile(fmt.Sprintf(rt, re))
	if e != nil {
		m := fmt.Sprintf("withinxwordssearch() could not compile second pass regex term: %s", re)
		msg(m, 1)
		return badsearch(m)
	}

	if len(slem) != 0 {
		re = strings.Join(lemmaintoregexslice(second.LemmaOne), "|")
	} else {
		re = sskg
	}

	patterntwo, e := regexp.Compile(re)
	if e != nil {
		m := fmt.Sprintf("withinxwordssearch() could not compile second pass regex term: %s", re)
		msg(m, 1)
		return badsearch(m)
	}

	// [c4] search head and tail for the second search term

	var validresults []DbWorkline
	for idx, str := range stringmapper {
		subs := patternone.FindStringSubmatch(str)
		head := ""
		tail := ""
		if len(subs) != 0 {
			head = subs[patternone.SubexpIndex("head")]
			tail = subs[patternone.SubexpIndex("tail")]
		}

		hh := strings.Split(head, " ")
		start := int64(0)
		if int64(len(hh))-first.ProxVal-1 > 0 {
			start = int64(len(hh)) - first.ProxVal - 1
		}
		hh = hh[start:]
		head = strings.Join(hh, " ")

		tt := strings.Split(tail, " ")
		if int64(len(tt)) >= first.ProxVal {
			tt = tt[0:first.ProxVal]
		}
		tail = strings.Join(tt, " ")

		if patterntwo.MatchString(head) || patterntwo.MatchString(tail) {
			validresults = append(validresults, first.Results[idx])
		}
	}

	second.Results = validresults

	return second
}

func generateinitialhits(first SearchStruct) SearchStruct {
	// part one of a two-part search

	first.Limit = FIRSTSEARCHLIM
	first = HGoSrch(first)

	if first.HasPhrase {
		mod := first
		// this will cut the hits by c. 50%
		mod.Results = findphrasesacrosslines(first)
		first = mod
	}
	return first
}

//
// INITIAL SETUP
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
	s.ProxVal = int64(sessions[user].Proximity)
	s.ProxScope = sessions[user].SearchScope
	s.NotNear = false
	s.Twobox = false
	s.HasPhrase = false
	s.HasLemma = false
	s.PhaseNum = 1
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

	s.Seeking = uvσςϲ(s.Seeking)
	s.Proximate = uvσςϲ(s.Proximate)

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

func badsearch(msg string) SearchStruct {
	var s SearchStruct
	var l DbWorkline
	l.MarkedUp = msg
	s.Results = append(s.Results, l)
	return s
}

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

func clonesearch(first SearchStruct, iteration int) SearchStruct {
	second := first
	second.Results = []DbWorkline{}
	second.Queries = []PrerolledQuery{}
	second.SearchIn = SearchIncExl{}
	second.SearchEx = SearchIncExl{}
	second.TTName = strings.Replace(uuid.New().String(), "-", "", -1)
	second.SkgSlice = []string{}
	second.PrxSlice = []string{}
	second.PhaseNum = iteration

	id := fmt.Sprintf("%s_pt%d", first.ID, iteration)
	second.ID = id // progresssocket() needs a new name
	return second
}

func searchtermfinder(term string) *regexp.Regexp {
	// find the universal regex equivalent of the search term
	//	you need to convert:
	//		ποταμον
	//	into:
	//		([πΠ][οὀὁὂὃὄὅόὸΟὈὉὊὋὌὍ][τΤ][αἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάᾈᾉᾊᾋᾌᾍᾎᾏἈἉἊἋἌἍἎἏΑ][μΜ][οὀὁὂὃὄὅόὸΟὈὉὊὋὌὍ][νΝ])

	converter := getrunefeeder()
	st := []rune(term)
	var stre string
	for _, r := range st {
		if _, ok := converter[r]; ok {
			re := fmt.Sprintf("[%s]", string(converter[r]))
			stre += re
		} else {
			stre += string(r)
		}
	}
	stre = fmt.Sprintf("(%s)", stre)

	pattern, e := regexp.Compile(stre)
	if e != nil {
		msg(fmt.Sprintf("searchtermfinder() could not compile the following: %s", stre), 1)
		pattern = regexp.MustCompile("FAILED_FIND_NOTHING")
	}
	return pattern
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
	out.JS = fmt.Sprintf(BROWSERJS, "browser")
	out.Title = s.Seeking
	out.Image = ""
	out.Searchsummary = formatfinalsearchsummary(&s)

	TABLEROW := `
	<tr class="%s">
		<td>
			<span class="findnumber">[%d]</span>&nbsp;&nbsp;%s
			<span class="foundauthor">%s</span>,&nbsp;<span class="foundwork">%s</span>:
			<browser id="%s"><span class="foundlocus">%s</span></browser>
		</td>
		<td class="leftpad">
			<span class="foundtext">%s</span>
		</td>
	</tr>
	`
	dtt := `[<span class="date">%s</span>]`

	pat := searchtermfinder(s.Seeking)

	var rows []string
	for i, r := range s.Results {
		// highlight search term; should be folded into a single function w/ highlightsearchterm() below
		var mu string
		if pat.MatchString(r.MarkedUp) {
			mu = pat.ReplaceAllString(r.MarkedUp, `<span class="match">$1</span>`)
		} else {
			// might be in the hyphenated line
			if pat.MatchString(r.Hyphenated) {
				// todo: needs more fiddling
				mu = r.MarkedUp + fmt.Sprintf(`&nbsp;&nbsp;(&nbsp;match:&nbsp;<span class="match">%s</span>&nbsp;)`, r.Hyphenated)
			} else {
				mu = r.MarkedUp
			}
		}

		mu = formateditorialbrackets(mu)

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
		fm := fmt.Sprintf(TABLEROW, rc, i+1, wd, au, wk, lk, lc, mu)
		rows = append(rows, fm)
	}

	out.Found = "<tbody>" + strings.Join(rows, "") + "</tbody>"

	js, e := json.Marshal(out)
	chke(e)

	return js
}

type ResultPassageLine struct {
	Locus           string
	Contents        string
	Hyphenated      string
	ContinuingStyle string
	IsHighlight     bool
}

func formatwithcontextresults(ss SearchStruct) []byte {
	// things to worry about: formateditorialbrackets(); unbalancedspancleaner()

	// unbalancedspancleaner() has to be run on the first line & after the whole block has been built

	// how/when to do <span class="highlight">

	thesession := sessions[ss.User]

	type PsgFormattingTemplate struct {
		Findnumber  int
		Foundauthor string
		Foundwork   string
		FindDate    string
		FindURL     string
		FindLocus   string
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
	for _, r := range res.Results {
		low := r.TbIndex - context
		high := r.TbIndex + context
		if low < 1 {
			// avoid "gr0258_FROM_-1_TO_3"
			low = 1
		}
		res.SearchIn.Passages = append(res.SearchIn.Passages, fmt.Sprintf(t, r.FindAuthor(), low, high))
	}

	res.Results = []DbWorkline{}

	res.Queries = searchlistintoqueries(&res)
	res = HGoSrch(res)

	// now you have all the lines you will ever need
	linemap := make(map[string]DbWorkline)
	for _, r := range res.Results {
		linemap[r.BuildHyperlink()] = r
	}

	// iterate over the results to build the raw core data
	urt := `linenumber/%s/%s/%d`
	dtt := `[<span class="date">%s</span>]`

	var allpassages []PsgFormattingTemplate
	for i, r := range ss.Results {
		var psg PsgFormattingTemplate
		psg.Findnumber = i + 1
		psg.Foundauthor = AllAuthors[r.FindAuthor()].Name
		psg.Foundwork = AllWorks[r.WkUID].Title
		psg.FindURL = r.BuildHyperlink()
		psg.FindLocus = strings.Join(r.FindLocus(), ".")
		psg.FindDate = formatinscriptiondates(dtt, r)

		for j := r.TbIndex - context; j <= r.TbIndex+context; j++ {
			url := fmt.Sprintf(urt, r.FindAuthor(), r.FindWork(), j)
			psg.RawCTX = append(psg.RawCTX, linemap[url])
		}

		// if you want to do this the horrifyingly slow way...
		// psg.RawCTX = simplecontextgrabber(r.FindAuthor(), r.TbIndex, int64(thesession.HitContext/2))

		for j := 0; j < len(psg.RawCTX); j++ {
			c := ResultPassageLine{}
			c.Locus = strings.Join(psg.RawCTX[j].FindLocus(), ".")

			if psg.RawCTX[j].BuildHyperlink() == psg.FindURL {
				c.IsHighlight = true
			} else {
				c.IsHighlight = false
			}
			c.Contents = psg.RawCTX[j].MarkedUp
			c.Hyphenated = psg.RawCTX[j].Hyphenated
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

		whole = textblockcleaner(whole)

		// reassemble
		block = strings.Split(whole, "✃✃✃")
		for i, b := range block {
			p.CookedCTX[i].Contents = b
		}
	}

	// highlight the search term: this includes the hyphenated_line issue

	for _, p := range allpassages {
		for i, r := range p.CookedCTX {
			if r.IsHighlight {
				highlightfocusline(&p.CookedCTX[i])
				pat := searchtermfinder(ss.Seeking)
				highlightsearchterm(pat, &p.CookedCTX[i])
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
		<span class="findnumber">[{{.Findnumber}}]</span>&nbsp;&nbsp;{{.FindDate}}
		<span class="foundauthor">{{.Foundauthor}}</span>,&nbsp;<span class="foundwork">{{.Foundwork}}</span>
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
	out.JS = fmt.Sprintf(BROWSERJS, "browser")
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
	if sessions[s.User].Earliest != MINDATESTR && sessions[s.User].Latest != MAXDATESTR {
		a := formatbcedate(sessions[s.User].Earliest)
		b := formatbcedate(sessions[s.User].Latest)
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

func highlightfocusline(line *ResultPassageLine) {
	line.Contents = fmt.Sprintf(`<span class="highlight">%s</span>`, line.Contents)
}

func highlightsearchterm(pattern *regexp.Regexp, line *ResultPassageLine) {
	// 	html markup for the search term in the line so it can jump out at you
	//
	//	regexequivalent is compiled via searchtermfinder()
	//

	// see the warnings and caveats at highlightsearchterm() in searchformatting.py
	if pattern.MatchString(line.Contents) {
		line.Contents = pattern.ReplaceAllString(line.Contents, `<span class="match">$1</span>`)
	} else {
		// might be in the hyphenated line
		if pattern.MatchString(line.Hyphenated) {
			// todo: needs more fiddling
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
