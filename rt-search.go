//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	hasAccent = regexp.MustCompile("[äëïöüâêîôûàèìòùáéíóúᾂᾒᾢᾃᾓᾣᾄᾔᾤᾅᾕᾥᾆᾖᾦᾇᾗᾧἂἒἲὂὒἢὢἃἓἳὃὓἣὣἄἔἴὄὔἤὤἅἕἵὅὕἥὥἆἶὖἦὦἇἷὗἧὧᾲῂῲᾴῄῴᾷῇῷᾀᾐᾠᾁᾑᾡῒῢΐΰῧἀἐἰὀὐἠὠῤἁἑἱὁὑἡὡῥὰὲὶὸὺὴὼάέίόύήώᾶῖῦῆῶϊϋ]")
)

type SearchStruct struct {
	User         string
	ID           string
	Seeking      string
	Proximate    string
	LemmaOne     string
	LemmaTwo     string
	InitSum      string
	Summary      string
	ProxScope    string // "lines" or "words"
	ProxType     string // "near" or "not near"
	ProxVal      int64
	HasLemma     bool
	HasPhrase    bool
	IsVector     bool
	IsActive     bool
	OneHit       bool
	Twobox       bool
	NotNear      bool
	SkgRewritten bool
	PhaseNum     int
	SrchColumn   string // usually "stripped_line", sometimes "accented_line"
	SrchSyntax   string // almost always "~"
	OrderBy      string // almost always "index" + ASC
	Limit        int64
	SkgSlice     []string // either just Seeking or a decomposed version of a Lemma's possibilities
	PrxSlice     []string
	SearchIn     SearchIncExl
	SearchEx     SearchIncExl
	Queries      []PrerolledQuery
	Results      []DbWorkline
	Launched     time.Time
	TTName       string
	SearchSize   int // # of works searched
	TableSize    int // # of tables searched
}

//
// ROUTING
//

func RtSearchConfirm(c echo.Context) error {
	return c.String(http.StatusOK, "8000")
}

func RtSearchStandard(c echo.Context) error {
	// "GET /search/standard/5446b840?skg=sine%20dolore HTTP/1.1"
	// "GET /search/standard/c2fba8e8?skg=%20dolore&prx=manif HTTP/1.1"
	// "GET /search/standard/2ad866e2?prx=manif&lem=dolor HTTP/1.1"
	// "GET /search/standard/02f3610f?lem=dolor&plm=manifesta HTTP/1.1"
	start := time.Now()
	previous := time.Now()
	user := readUUIDCookie(c)

	id := c.Param("id")
	skg := c.QueryParam("skg")
	prx := c.QueryParam("prx")
	lem := c.QueryParam("lem")
	plm := c.QueryParam("plm")

	srch := builddefaultsearch(c)

	// HasPhrase makes us use a fake limit temporarily
	reallimit := srch.Limit

	srch.Seeking = skg
	srch.Proximate = prx
	srch.LemmaOne = lem
	srch.LemmaTwo = plm
	srch.User = user
	srch.ID = purgechars(UNACCEPTABLEINPUT, id)
	srch.IsVector = false

	parsesearchinput(&srch)

	// must happen before searchlistintoqueries()
	setsearchtype(&srch)

	srch.InitSum = formatinitialsummary(srch)

	// now safe to rewrite skg so that "^|\s", etc. can be added
	srch.Seeking = whitespacer(srch.Seeking, &srch)
	srch.Proximate = whitespacer(srch.Proximate, &srch)

	sl := sessionintosearchlist(sessions[user])
	srch.SearchIn = sl.Inc
	srch.SearchEx = sl.Excl
	srch.SearchSize = sl.Size

	timetracker("A", "sessionintosearchlist()", start, previous)
	previous = time.Now()

	prq := searchlistintoqueries(&srch)
	srch.TableSize = len(prq)

	timetracker("B", "searchlistintoqueries()", start, previous)
	previous = time.Now()

	srch.Queries = prq
	srch.IsActive = true
	searches[id] = srch

	var completed SearchStruct
	if searches[id].Twobox {
		// todo: triple-check results against python
		if searches[id].ProxScope == "words" {
			completed = withinxwordssearch(searches[id])
		} else {
			completed = withinxlinessearch(searches[id])
		}
	} else {
		completed = HGoSrch(searches[id])
	}

	if completed.HasPhrase {
		// you did HGoSrch() and need to check the windowed lines
		// withinxlinessearch() has already done the checking
		// the cannot assign problem...
		completed.Results = findphrasesacrosslines(completed)
		if int64(len(completed.Results)) > reallimit {
			completed.Results = completed.Results[0:reallimit]
		}
	}

	timetracker("C", fmt.Sprintf("search executed: %d hits", len(searches[id].Results)), start, previous)
	previous = time.Now()

	resultsorter(&completed)

	searches[id] = completed

	var js string
	if sessions[readUUIDCookie(c)].HitContext == 0 {
		js = string(formatnocontextresults(searches[id]))
	} else {
		js = string(formatwithcontextresults(searches[id]))
	}

	timetracker("D", fmt.Sprintf("formatted %d hits", len(searches[id].Results)), start, previous)
	previous = time.Now()

	delete(searches, id)
	progremain.Delete(id)

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
		// avoid "gr0028_FROM_-1_TO_5"
		low := r.TbIndex - first.ProxVal
		if low < 1 {
			low = 1
		}
		np := fmt.Sprintf(pt, r.FindAuthor(), low, r.TbIndex+first.ProxVal)
		newpsg = append(newpsg, np)
	}

	second.Limit = originalsrch.Limit
	second.SearchIn.Passages = newpsg

	prq := searchlistintoqueries(&second)

	d = fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf("%s searchlistintoqueries() rerun", d), 4)
	previous = time.Now()

	second.Queries = prq
	second = HGoSrch(second)

	// was this a "notnear" search?
	if second.NotNear {
		var actualhits []DbWorkline
		// any original hits that match lines from pt2 are the "real" hits
		mapper := make(map[string]bool)
		for _, r := range first.Results {
			mapper[r.Citation()] = true
		}
		for _, r := range second.Results {
			if _, ok := mapper[r.Citation()]; ok {
				actualhits = append(actualhits, r)
			}
		}
		second.Results = actualhits
	}

	d = fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf("%s withinxlinessearch(): %d subsequent hits", d, len(first.Results)), 4)

	// findphrasesacrosslines() check happens just after you exit this function
	searches[originalsrch.ID] = second
	return searches[originalsrch.ID]
}

// withinxwordssearch - find A within N words of B
func withinxwordssearch(originalsrch SearchStruct) SearchStruct {

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
	// avoid "WHERE accented_line !~ ''" : force the type and make sure to check "first.NotNear" below
	second.NotNear = false

	setsearchtype(&second)

	// [a1] hard code a suspect assumption...
	AVERAGEWRDSPERLINE := 8
	need := 2 + (first.ProxVal / int64(AVERAGEWRDSPERLINE))

	pt := `%s_FROM_%d_TO_%d`
	t := `linenumber/%s/%s/%d`

	resultmapper := make(map[string]int, len(first.Results))
	var newpsg []string

	// [a2] pick the lines to grab and associate them with the hits they go with
	for i, r := range first.Results {
		low := r.TbIndex - need
		if low < 1 {
			low = 1
		}
		np := fmt.Sprintf(pt, r.FindAuthor(), low, r.TbIndex+need)
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

		if first.NotNear {
			// toss hits
			if !patterntwo.MatchString(head) && !patterntwo.MatchString(tail) {
				validresults = append(validresults, first.Results[idx])
			}
		} else {
			// collect hits
			if patterntwo.MatchString(head) || patterntwo.MatchString(tail) {
				validresults = append(validresults, first.Results[idx])
			}
		}
	}

	second.Results = validresults

	// restore deleted search info
	second.Seeking = sskg
	second.LemmaOne = slem

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
	s.SkgRewritten = false
	s.OneHit = sessions[user].OneHit
	s.PhaseNum = 1
	s.TTName = strings.Replace(uuid.New().String(), "-", "", -1)

	if sessions[user].NearOrNot == "notnear" {
		s.NotNear = true
	}

	// msg("nonstandard builddefaultsearch() for testing", 1)

	return s
}

// buildhollowsearch - is really a way to grab line collections via synthetic searchlists
func buildhollowsearch() SearchStruct {
	s := SearchStruct{
		User:         "",
		ID:           strings.Replace(uuid.New().String(), "-", "", -1),
		Seeking:      "",
		Proximate:    "",
		LemmaOne:     "",
		LemmaTwo:     "",
		InitSum:      "",
		Summary:      "",
		ProxScope:    "",
		ProxType:     "",
		ProxVal:      0,
		HasLemma:     false,
		HasPhrase:    false,
		IsVector:     false,
		IsActive:     false,
		OneHit:       false,
		Twobox:       false,
		NotNear:      false,
		SkgRewritten: false,
		PhaseNum:     0,
		SrchColumn:   DEFAULTCOLUMN,
		SrchSyntax:   DEFAULTSYNTAX,
		OrderBy:      ORDERBY,
		Limit:        FIRSTSEARCHLIM,
		SkgSlice:     nil,
		PrxSlice:     nil,
		SearchIn:     SearchIncExl{},
		SearchEx:     SearchIncExl{},
		Queries:      nil,
		Results:      nil,
		Launched:     time.Now(),
		TTName:       strings.Replace(uuid.New().String(), "-", "", -1),
		SearchSize:   0,
		TableSize:    0,
	}
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
	// twowords := regexp.MustCompile(`\w\s\w`)
	// will not find `τῷ φίλῳ` or `πλάντᾳ ἵνα`
	// twowords := regexp.MustCompile(`[A-Za-zΑ-ΩϹα-ωϲ]\s[A-Za-zΑ-ΩϹα-ωϲ]`)

	acc := `ϲῥἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάἐἑἒἓἔἕὲέἰἱἲἳἴἵἶἷὶίῐῑῒΐῖῗὀὁὂὃὄὅόὸὐὑὒὓὔὕὖὗϋῠῡῢΰῦῧύὺᾐᾑᾒᾓᾔᾕᾖᾗῂῃῄῆῇἤἢἥἣὴήἠἡἦἧὠὡὢὣὤὥὦὧᾠᾡᾢᾣᾤᾥᾦᾧῲῳῴῶῷώὼ`
	reg := `a-zα-ω`
	comp := fmt.Sprintf(`[%s%s]\s[%s%s]`, reg, acc, reg, acc)
	twowords := regexp.MustCompile(comp)

	if twowords.MatchString(srch.Seeking) {
		srch.HasPhrase = true
	}

	if len(srch.LemmaOne) != 0 {
		srch.HasLemma = true
		srch.SrchColumn = "accented_line"
	}

	return
}

func restorewhitespace(skg string) string {
	// will have a problem rewriting regex inside phrasecombinations() if you don't clear whitespacer() products out
	// even though we are about to put exactly this back in again...
	skg = strings.Replace(skg, "(^|\\s)", " ", 1)
	skg = strings.Replace(skg, "(\\s|$)", " ", -1)
	return skg
}

func whitespacer(skg string, ss *SearchStruct) string {
	// whitespace issue: " ἐν Ὀρέϲτῃ " cannot be found at the start of a line where it is "ἐν Ὀρέϲτῃ "
	// do not run this before formatinitialsummary()
	if strings.Contains(skg, " ") {
		ss.SkgRewritten = true
		rs := []rune(skg)
		a := ""
		if rs[0] == ' ' {
			a = "(^|\\s)"
		}
		z := ""
		if rs[len(rs)-1] == ' ' {
			z = "(\\s|$)"
		}
		skg = strings.TrimSpace(skg)
		skg = a + skg + z
	}
	return skg
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

	// there is a problem: unless you do something, "(^|\s)ἁλιεύϲ(\s|$)" will be a search term but this will not find "ἁλιεὺϲ"
	var lemm []string
	for _, l := range AllLemm[hdwd].Deriv {
		lemm = append(lemm, findacuteorgrave(l))
	}

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

	var valid = make(map[string]DbWorkline, len(ss.Results))

	skg := ss.Seeking
	if ss.SkgRewritten {
		skg = restorewhitespace(ss.Seeking)
	}

	find := regexp.MustCompile(`^ `)
	re := find.ReplaceAllString(skg, "(^|\\s)")
	find = regexp.MustCompile(` $`)
	re = find.ReplaceAllString(re, "(\\s|$)")
	fp := regexp.MustCompile(re)
	altfp := regexp.MustCompile(ss.Seeking)

	for i, r := range ss.Results {
		// do the "it's all on this line" case separately
		li := columnpicker(ss.SrchColumn, r)
		f := fp.MatchString(li)
		if f {
			valid[r.BuildHyperlink()] = r
		} else if ss.SkgRewritten && altfp.MatchString(li) {
			// i.e. "it's all on this line" (second try)
			// msg("althit", 1)
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

func universalpatternmaker(term string) string {
	// feeder for searchtermfinder() also used by searchformatting.go
	// converter := extendedrunefeeder()
	converter := erunef // see top of generichelpers.go
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
	return stre
}

func searchtermfinder(term string) *regexp.Regexp {
	// find the universal regex equivalent of the search term
	//	you need to convert:
	//		ποταμον
	//	into:
	//		([πΠ][οὀὁὂὃὄὅόὸΟὈὉὊὋὌὍ][τΤ][αἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάᾈᾉᾊᾋᾌᾍᾎᾏἈἉἊἋἌἍἎἏΑ][μΜ][οὀὁὂὃὄὅόὸΟὈὉὊὋὌὍ][νΝ])

	stre := universalpatternmaker(term)
	pattern, e := regexp.Compile(stre)
	if e != nil {
		msg(fmt.Sprintf("searchtermfinder() could not compile the following: %s", stre), 1)
		pattern = regexp.MustCompile("FAILED_FIND_NOTHING")
	}
	return pattern
}

func resultsorter(ss *SearchStruct) {
	// Closures that order the DbWorkline structure:
	// see generichelpers.go and https://pkg.go.dev/sort#example__sortMultiKeys
	nameIncreasing := func(one, two *DbWorkline) bool {
		a1 := AllAuthors[one.FindAuthor()].Shortname
		a2 := AllAuthors[two.FindAuthor()].Shortname
		return a1 < a2
	}

	titleIncreasing := func(one, two *DbWorkline) bool {
		return AllWorks[one.WkUID].Title < AllWorks[two.WkUID].Title
	}

	dateIncreasing := func(one, two *DbWorkline) bool {
		d1 := AllWorks[one.WkUID].RecDate
		d2 := AllWorks[two.WkUID].RecDate
		if d1 != "Unavailable" && d2 != "Unavailable" {
			return AllWorks[one.WkUID].ConvDate < AllWorks[two.WkUID].ConvDate
		} else if d1 == "Unavailable" && d2 != "Unavailable" {
			return AllAuthors[one.FindAuthor()].ConvDate < AllWorks[two.WkUID].ConvDate
		} else if d1 != "Unavailable" && d2 == "Unavailable" {
			return AllWorks[one.WkUID].ConvDate < AllAuthors[two.FindAuthor()].ConvDate
		} else {
			return AllAuthors[one.FindAuthor()].ConvDate < AllAuthors[two.FindAuthor()].ConvDate
		}
	}

	//dateDecreasing := func(one, two *DbWorkline) bool {
	//	return AllWorks[one.FindWork()].ConvDate > AllWorks[two.FindWork()].ConvDate
	//}

	increasingLines := func(one, two *DbWorkline) bool {
		return one.TbIndex < two.TbIndex
	}

	//decreasingLines := func(one, two *DbWorkline) bool {
	//	return one.TbIndex > two.TbIndex // Note: > orders downwards.
	//}

	increasingID := func(one, two *DbWorkline) bool {
		return one.BuildHyperlink() < two.BuildHyperlink()
	}

	crit := sessions[ss.User].SortHitsBy

	switch {
	// unhandled are "location" & "provenance"
	case crit == "shortname":
		OrderedBy(nameIncreasing, titleIncreasing, increasingLines).Sort(ss.Results)
	case crit == "converted_date":
		OrderedBy(dateIncreasing, nameIncreasing, titleIncreasing, increasingLines).Sort(ss.Results)
	case crit == "universalid":
		OrderedBy(increasingID).Sort(ss.Results)
	default:
		// author nameIncreasing
		OrderedBy(nameIncreasing, increasingLines).Sort(ss.Results)
	}
}
