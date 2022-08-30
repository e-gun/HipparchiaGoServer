package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type SearchStruct struct {
	User       string
	ID         string
	Seeking    string
	Proximate  string
	LemmaOne   string
	LemmaTwo   string
	Summary    string
	QueryType  string
	ProxScope  string // "lines" or "words"
	ProxType   string // "near" or "not near"
	ProxVal    int64
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

func (s SearchStruct) HasLemma() bool {
	if len(s.LemmaOne) > 0 || len(s.LemmaTwo) > 0 {
		return true
	} else {
		return false
	}
}

type SearchOutput struct {
	// meant to turn into JSON
	Title         string `json:"title"`
	Searchsummary string `json:"searchsummary"`
	Found         string `json:"found"`
	Image         string `json:"image"`
	JS            string `json:"js"`
}

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
	timetracker("A", "builddefaultsearch()", start, previous)
	previous = time.Now()

	srch.Seeking = skg
	srch.Proximate = prx
	srch.LemmaOne = lem
	srch.LemmaTwo = plm
	srch.User = user
	srch.ID = id
	srch.IsVector = false

	sl := sessionintosearchlist(sessions[user])
	srch.SearchIn = sl.Inc
	srch.SearchEx = sl.Excl
	srch.SearchSize = sl.Size

	timetracker("B", "sessionintosearchlist()", start, previous)
	previous = time.Now()

	// must happen before searchlistintoqueries()
	srch = setsearchtype(srch)

	//if srch.LemmaTwo != "" {
	//	srch.PrxSlice = lemmaintoregexslice(srch.LemmaTwo)
	//} else {
	//	srch.PrxSlice = append(srch.PrxSlice, srch.Proximate)
	//}

	prq := searchlistintoqueries(srch)
	timetracker("C", "searchlistintoqueries()", start, previous)
	previous = time.Now()

	srch.Queries = prq
	searches[id] = srch

	// return results via searches[id].Results
	if searches[id].Twobox {
		// this needs to be able to do word + lemma; double lemmata; phrase + phrase...
		// can do:
		// [1] single + single
		// [2]
		// todo: "not near" syntax
		msg("twobox", 4)
		searches[id] = withinxlinessearch(searches[id])
	} else {
		searches[id] = HGoSrch(searches[id])
	}

	if searches[id].QueryType == "phrase" || searches[id].QueryType == "phrase_and_proximity" {
		// you did HGoSrch() and need to check the windowed lines
		// withinxlinessearch() has already done the checking
		// the cannot assign problem...
		mod := searches[id]
		mod.Results = findphrasesacrosslines(searches[id])
		searches[id] = mod
	}

	previous = time.Now()

	//hits := searches[id].Results
	//for i, h := range hits {
	//	t := fmt.Sprintf("%d - %srch : %srch", i, h.FindLocus(), h.MarkedUp)
	//	fmt.Println(t)
	//}

	timetracker("D", fmt.Sprintf("search executed: %d hits", len(searches[id].Results)), start, previous)

	js := string(formatnocontextresults(searches[id]))

	return c.String(http.StatusOK, js)
}

func builddefaultsearch(c echo.Context) SearchStruct {
	var s SearchStruct

	user := readUUIDCookie(c)

	s.User = user
	s.Launched = time.Now()
	s.Limit = sessions[user].HitLimit
	s.SrchColumn = "stripped_line"
	s.SrchSyntax = "~*"
	s.OrderBy = "index"
	s.SearchIn = sessions[user].Inclusions
	s.SearchEx = sessions[user].Exclusions
	s.ProxVal = DEFAULTPROXIMITY
	s.ProxScope = DEFAULTPROXIMITYSCOPE
	s.NotNear = false
	s.TTName = strings.Replace(uuid.New().String(), "-", "", -1)
	return s
}

func setsearchtype(srch SearchStruct) SearchStruct {
	containsphrase := false
	containslemma := false
	twobox := false

	// will not find greek...
	// pattern := regexp.MustCompile(`\w\s\w`)

	pattern := regexp.MustCompile(`[A-Za-zΑ-ΩϹα-ωϲ]\s[A-Za-zΑ-ΩϹα-ωϲ]`)

	if pattern.MatchString(srch.Seeking) || pattern.MatchString(srch.Proximate) {
		containsphrase = true
	}
	if srch.LemmaOne != "" || srch.LemmaTwo != "" {
		containslemma = true
	}
	if srch.LemmaOne != "" && srch.LemmaTwo != "" {
		twobox = true
	}

	if srch.Seeking != "" && srch.Proximate != "" {
		twobox = true
	}

	if srch.Seeking != "" && srch.LemmaTwo != "" {
		twobox = true
	}

	if srch.LemmaOne != "" && srch.Proximate != "" {
		twobox = true
	}

	srch.Twobox = twobox

	if containsphrase && !twobox {
		srch.QueryType = "phrase"
	} else if containsphrase && twobox {
		srch.QueryType = "phrase_and_proximity"
	} else if containslemma && !twobox {
		srch.QueryType = "simplelemma"
		srch.SrchColumn = "accented_line"
	} else if twobox {
		srch.QueryType = "proximity"
	} else {
		srch.QueryType = "simple"
	}

	return srch
}

func formatnocontextresults(s SearchStruct) []byte {
	var out SearchOutput
	out.JS = BROWSERJS
	out.Title = s.Seeking
	out.Image = ""
	out.Searchsummary = formatsearchsummary(s)

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
		if i%3 == 0 {
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

func formatsearchsummary(s SearchStruct) string {

	t := `
	<div id="searchsummary">
		Sought %s<span class="sought">»%s«</span>%s
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

	win := ` within %d %s of %s<span class="sought">»%s«</span>`

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

	af := ""
	sk := s.Seeking
	if s.LemmaOne != "" {
		af = "all forms of "
		sk = s.LemmaOne
	}

	two := ""
	if s.Twobox {
		sk = s.Proximate
		af = ""
		if s.LemmaTwo != "" {
			af = "all forms of "
			sk = s.LemmaTwo
		}
		two = fmt.Sprintf(win, s.ProxVal, s.ProxScope, af, sk)
	}

	so := sessions[s.User].SrchOutSettings.SortHitsBy
	el := fmt.Sprintf("%.3f", time.Now().Sub(s.Launched).Seconds())
	// need to record # of works and not # of tables somewhere & at the right moment...
	sum := fmt.Sprintf(t, af, sk, two, s.SearchSize, len(s.Results), el, so, dr, hitcap)
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

func lemmaintoflatregex(hdwd string) string {
	// a single regex string for all forms
	var re string
	if _, ok := AllLemm[hdwd]; !ok {
		msg(fmt.Sprintf("lemmaintoregexslice() could not find '%s'", hdwd), 1)
		return re
	}

	tp := `(^|\s)%s(\s|$)`
	lemm := AllLemm[hdwd].Deriv

	var bnd []string
	for _, l := range lemm {
		bnd = append(bnd, fmt.Sprintf(tp, l))
	}

	re = strings.Join(bnd, "|")

	return re
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

func withinxlinessearch(originalsrch SearchStruct) SearchStruct {
	// after finding x, look for y within n lines of x

	// "decessionis" near "spem" in Cicero...

	// (part 1)
	//		HGoSrch(first)
	//
	// (part 2)
	// 		populate a new search list with a ton of passages via the first results
	//		HGoSrch(second)

	// todo: this won't work with phrases

	first := originalsrch
	first.Limit = FIRSTSEARCHLIM
	first = HGoSrch(first)

	// osk := originalsrch.Seeking
	// oss := originalsrch.SkgSlice
	osl := originalsrch.Limit
	// oslm := originalsrch.LemmaOne

	msg(fmt.Sprintf("withinxlinessearch(): %d initial hits", len(first.Results)), 4)

	if first.QueryType == "phrase_and_proximity" {
		mod := first
		mod.Results = findphrasesacrosslines(first)
		first = mod
	}

	msg(fmt.Sprintf("withinxlinessearch(): %d findphrasesacrosslines() hits", len(first.Results)), 4)

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
	second.Limit = osl

	second = setsearchtype(second)

	msg(fmt.Sprintf("ll1: %s\nll2:%s\n", second.QueryType, second.LemmaTwo), 4)

	pt := `%s_FROM_%d_TO_%d`

	var newpsg []string
	for _, r := range first.Results {
		np := fmt.Sprintf(pt, r.FindAuthor(), r.TbIndex-originalsrch.ProxVal, r.TbIndex+originalsrch.ProxVal)
		newpsg = append(newpsg, np)
	}

	// todo: not near logic
	second.SearchIn.Passages = newpsg

	prq := searchlistintoqueries(second)

	second.Queries = prq
	searches[originalsrch.ID] = second
	searches[originalsrch.ID] = HGoSrch(searches[originalsrch.ID])

	return searches[originalsrch.ID]
}

func old_withinxlinessearch(s SearchStruct) SearchStruct {
	// after finding x, look for y within n lines of x

	// "decessionis" near "spem" in Cicero...

	// (part 1)
	//		HGoSrch(first)
	//
	// (part 2.1)
	//		CREATE TEMPORARY TABLE lt0474_includelist_24bfe76dc1124f07becabb389a4f393d AS
	//		SELECT values AS includeindex FROM
	//			unnest(ARRAY[39844,39845,39846,39847,39848,39849,39850,39851,39852,39853,128858,128859,128860,128861,128862,128863,128864,128865,128866,128867,138278,138279,138280,138281,138282,138283,138284,138285,138286,138287])
	//		values

	// (part 2.2)
	// 		SELECT second.wkuniversalid, second.index, second.level_05_value, second.level_04_value, second.level_03_value, second.level_02_value, second.level_01_value, second.level_00_value,
	//			second.marked_up_line, second.accented_line, second.stripped_line, second.hyphenated_words, second.annotations FROM
	//		( SELECT * FROM
	//			( SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations,
	//				concat(stripped_line, ' ', lead(stripped_line) OVER (ORDER BY index ASC) ) AS linebundle
	//				FROM
	//	lt0474 WHERE EXISTS
	//		(SELECT 1 FROM lt0474_includelist_24bfe76dc1124f07becabb389a4f393d incl
	//			WHERE incl.includeindex = lt0474.index)
	//					) first
	//				) second WHERE second.linebundle ~ 'spem' LIMIT 200;

	// alternate strategy, but not a universal solution to the various types of search linebundles can handle:
	// ... FROM lt0474 WHERE ( (index BETWEEN 128860 AND 128866) OR (index BETWEEN 39846 AND 39852) OR ... )

	// todo: it looks like we can't do "within 5 lines" this way since the bundle is too small; grablinebundles() instead

	first := s
	first.Limit = FIRSTSEARCHLIM
	first = HGoSrch(first)

	msg(fmt.Sprintf("withinxlinessearch(): %d initial hits", len(first.Results)), 4)

	// convert the hits into new selections:
	// a temptable will be built once you know which lines do you need from which works

	var required = make(map[string][]int64)
	for _, r := range first.Results {
		w := AllWorks[r.WkUID]
		var idx []int64
		for i := r.TbIndex - s.ProxVal; i < r.TbIndex+s.ProxVal; i++ {
			if i >= w.FirstLine && i <= w.LastLine {
				idx = append(idx, i)
			}
		}
		required[w.FindAuthor()] = append(required[w.FindAuthor()], idx...)
	}

	// prepare new search
	fss := first.SkgSlice

	second := first
	second.Results = []DbWorkline{}
	second.Queries = []PrerolledQuery{}
	second.TTName = strings.Replace(uuid.New().String(), "-", "", -1)
	second.SkgSlice = second.PrxSlice
	second.PrxSlice = fss
	second.Limit = s.Limit

	var ttsq = make(map[string]string)
	ctt := `
	CREATE TEMPORARY TABLE %s_includelist_%s AS 
		SELECT values AS includeindex FROM 
			unnest(ARRAY[%s])
		values`

	for r, vv := range required {
		var arr []string
		for _, v := range vv {
			arr = append(arr, strconv.FormatInt(v, 10))
		}
		a := strings.Join(arr, ",")
		ttsq[r] = fmt.Sprintf(ctt, r, second.TTName, a)
	}

	seltempl := PRFXSELFRM + CONCATSELFROM

	wha := `
	%s WHERE EXISTS 
		(SELECT 1 FROM %s_includelist_%s incl 
			WHERE incl.includeindex = %s.index)`
	whb := ` WHERE second.linebundle ~ '%s' LIMIT %d;`
	var prqq = make(map[string][]PrerolledQuery)
	for i, q := range second.SkgSlice {
		for r, _ := range required {
			var prq PrerolledQuery
			modname := second.TTName + fmt.Sprintf("_%d", i)
			prq.TempTable = strings.Replace(ttsq[r], second.TTName, modname, -1)
			whc := fmt.Sprintf(wha, r, r, modname, r)
			whd := fmt.Sprintf(whb, q, second.Limit)
			prq.PsqlQuery = fmt.Sprintf(seltempl, whc) + whd
			prqq[r] = append(prqq[r], prq)
		}
	}

	for _, q := range prqq {
		second.Queries = append(second.Queries, q...)
	}

	second = HGoSrch(second)

	// windows of lines come back: e.g., three lines that look like they match when only one matches [3131, 3132, 3133]
	// figure out which line is really the line with the goods

	phrasefinder := regexp.MustCompile(`[A-Za-zΑ-ΩϹα-ωϲ]\s[A-Za-zΑ-ΩϹα-ωϲ]`)

	if phrasefinder.MatchString(second.Seeking) {
		second.Results = findphrasesacrosslines(second)
	} else {
		second.Results = validatebundledhits(second)
	}

	return second
}

func validatebundledhits(ss SearchStruct) []DbWorkline {
	// if the second search term available in the window of lines?
	re := ss.Proximate
	if ss.LemmaTwo != "" {
		re = lemmaintoflatregex(ss.LemmaTwo)
	}

	find := regexp.MustCompile(re)

	var valid []DbWorkline
	for _, r := range ss.Results {
		li := columnpicker(ss.SrchColumn, r)
		if find.MatchString(li) {
			valid = append(valid, r)
		}
	}

	return valid
}

func findphrasesacrosslines(ss SearchStruct) []DbWorkline {
	// "one two$" + "^three four" makes a hit if you want "one two three four"
	var valid = make(map[string]DbWorkline)

	find := regexp.MustCompile(`^\s`)
	re := find.ReplaceAllString(ss.Seeking, "(^|\\s)")
	find = regexp.MustCompile(`\s$`)
	re = find.ReplaceAllString(ss.Seeking, "(\\s|$)")

	for i, r := range ss.Results {
		// do the "it's all on this line" case separately
		li := columnpicker(ss.SrchColumn, r)
		//msg(li, 4)
		//msg(ss.Seeking, 4)
		fp := regexp.MustCompile(re)
		f := fp.MatchString(li)
		if f {
			// msg("initial match", 4)
			valid[r.BuildHyperlink()] = r
		} else {
			// msg("'else'", 4)
			var nxt DbWorkline
			if i+1 < len(ss.Results) {
				nxt = ss.Results[i+1]
				if r.WkUID != nxt.WkUID || r.TbIndex+1 != nxt.TbIndex {
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

			// the following version will double-register some hits; combinator dodges that

			//pattern := regexp.MustCompile(ss.Seeking)
			//nl := columnpicker(ss.SrchColumn, nxt)
			//targ := li + " " + nl
			//fmt.Printf("'%s'\n", targ)
			//targ = strings.Replace(targ, "  ", " ", -1)
			//if pattern.MatchString(targ) && r.WkUID == nxt.WkUID {
			//	valid[r.BuildHyperlink()] = r
			//}

			nl := columnpicker(ss.SrchColumn, nxt)
			// msg(nl, 4)
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

/* [python]

Sought »dolor« within 3 lines of »maer«
Searched 1 works and found 2 passages (0.17s)
Sorted by name
{'lt1038': {'temptable': '', 'query': 'SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt1038 WHERE ( (index BETWEEN 1474 AND 1480) OR (index BETWEEN 11397 AND 11403) OR (index BETWEEN 822 AND 828) OR (index BETWEEN 5039 AND 5045) OR (index BETWEEN 2539 AND 2545) OR (index BETWEEN 5129 AND 5135) OR (index BETWEEN 2832 AND 2838) OR (index BETWEEN 11252 AND 11258) OR (index BETWEEN 2885 AND 2891) OR (index BETWEEN 11827 AND 11833) OR (index BETWEEN 8396 AND 8402) OR (index BETWEEN 5383 AND 5389) OR (index BETWEEN 11364 AND 11370) OR (index BETWEEN 384 AND 390) OR (index BETWEEN 6951 AND 6957) OR (index BETWEEN 11826 AND 11832) OR (index BETWEEN 3342 AND 3348) OR (index BETWEEN 1834 AND 1840) OR (index BETWEEN 542 AND 548) OR (index BETWEEN 6980 AND 6986) OR (index BETWEEN 7009 AND 7015) OR (index BETWEEN 9790 AND 9796) OR (index BETWEEN 3305 AND 3311) OR (index BETWEEN 3277 AND 3283) OR (index BETWEEN 4310 AND 4316) OR (index BETWEEN 6297 AND 6303) OR (index BETWEEN 11256 AND 11262) OR (index BETWEEN 199 AND 205) OR (index BETWEEN 5070 AND 5076) OR (index BETWEEN 7140 AND 7146) OR (index BETWEEN 9802 AND 9808) OR (index BETWEEN 11385 AND 11391) OR (index BETWEEN 3262 AND 3268) OR (index BETWEEN 11867 AND 11873) OR (index BETWEEN 12022 AND 12028) OR (index BETWEEN 12104 AND 12110) ) AND ( stripped_line ~* %s )  ORDER BY index ASC LIMIT 200', 'data': ('maer',)}}


Sought all 12 known forms of »maeror« within 3 lines of »dolor«
Searched 1 works and found 2 passages (0.14s)
Sorted by name
{'lt1038_0': {'data': '(^|\\s)maerorem(\\s|$)', 'query': 'SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt1038 WHERE ( (index BETWEEN 1474 AND 1480) OR (index BETWEEN 11397 AND 11403) OR (index BETWEEN 822 AND 828) OR (index BETWEEN 5039 AND 5045) OR (index BETWEEN 2539 AND 2545) OR (index BETWEEN 5129 AND 5135) OR (index BETWEEN 2832 AND 2838) OR (index BETWEEN 11252 AND 11258) OR (index BETWEEN 2885 AND 2891) OR (index BETWEEN 11827 AND 11833) OR (index BETWEEN 8396 AND 8402) OR (index BETWEEN 5383 AND 5389) OR (index BETWEEN 11364 AND 11370) OR (index BETWEEN 384 AND 390) OR (index BETWEEN 6951 AND 6957) OR (index BETWEEN 11826 AND 11832) OR (index BETWEEN 3342 AND 3348) OR (index BETWEEN 1834 AND 1840) OR (index BETWEEN 542 AND 548) OR (index BETWEEN 6980 AND 6986) OR (index BETWEEN 7009 AND 7015) OR (index BETWEEN 9790 AND 9796) OR (index BETWEEN 3305 AND 3311) OR (index BETWEEN 3277 AND 3283) OR (index BETWEEN 4310 AND 4316) OR (index BETWEEN 6297 AND 6303) OR (index BETWEEN 11256 AND 11262) OR (index BETWEEN 199 AND 205) OR (index BETWEEN 5070 AND 5076) OR (index BETWEEN 7140 AND 7146) OR (index BETWEEN 9802 AND 9808) OR (index BETWEEN 11385 AND 11391) OR (index BETWEEN 3262 AND 3268) OR (index BETWEEN 11867 AND 11873) OR (index BETWEEN 12022 AND 12028) OR (index BETWEEN 12104 AND 12110) ) AND ( stripped_line ~* %s )  ORDER BY index ASC LIMIT 200', 'temptable': ''}, 'lt1038_1': {'data': '(^|\\s)maerore(\\s|$)', 'query': 'SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt1038 WHERE ( (index BETWEEN 1474 AND 1480) OR (index BETWEEN 11397 AND 11403) OR (index BETWEEN 822 AND 828) OR (index BETWEEN 5039 AND 5045) OR (index BETWEEN 2539 AND 2545) OR (index BETWEEN 5129 AND 5135) OR (index BETWEEN 2832 AND 2838) OR (index BETWEEN 11252 AND 11258) OR (index BETWEEN 2885 AND 2891) OR (index BETWEEN 11827 AND 11833) OR (index BETWEEN 8396 AND 8402) OR (index BETWEEN 5383 AND 5389) OR (index BETWEEN 11364 AND 11370) OR (index BETWEEN 384 AND 390) OR (index BETWEEN 6951 AND 6957) OR (index BETWEEN 11826 AND 11832) OR (index BETWEEN 3342 AND 3348) OR (index BETWEEN 1834 AND 1840) OR (index BETWEEN 542 AND 548) OR (index BETWEEN 6980 AND 6986) OR (index BETWEEN 7009 AND 7015) OR (index BETWEEN 9790 AND 9796) OR (index BETWEEN 3305 AND 3311) OR (index BETWEEN 3277 AND 3283) OR (index BETWEEN 4310 AND 4316) OR (index BETWEEN 6297 AND 6303) OR (index BETWEEN 11256 AND 11262) OR (index BETWEEN 199 AND 205) OR (index BETWEEN 5070 AND 5076) OR (index BETWEEN 7140 AND 7146) OR (index BETWEEN 9802 AND 9808) OR (index BETWEEN 11385 AND 11391) OR (index BETWEEN 3262 AND 3268) OR (index BETWEEN 11867 AND 11873) OR (index BETWEEN 12022 AND 12028) OR (index BETWEEN 12104 AND 12110) ) AND ( stripped_line ~* %s )  ORDER BY index ASC LIMIT 200', 'temptable': ''}, 'lt1038_2': {'data': '(^|\\s)maerores(\\s|$)', 'query': 'SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt1038 WHERE ( (index BETWEEN 1474 AND 1480) OR (index BETWEEN 11397 AND 11403) OR (index BETWEEN 822 AND 828) OR (index BETWEEN 5039 AND 5045) OR (index BETWEEN 2539 AND 2545) OR (index BETWEEN 5129 AND 5135) OR (index BETWEEN 2832 AND 2838) OR (index BETWEEN 11252 AND 11258) OR (index BETWEEN 2885 AND 2891) OR (index BETWEEN 11827 AND 11833) OR (index BETWEEN 8396 AND 8402) OR (index BETWEEN 5383 AND 5389) OR (index BETWEEN 11364 AND 11370) OR (index BETWEEN 384 AND 390) OR (index BETWEEN 6951 AND 6957) OR (index BETWEEN 11826 AND 11832) OR (index BETWEEN 3342 AND 3348) OR (index BETWEEN 1834 AND 1840) OR (index BETWEEN 542 AND 548) OR (index BETWEEN 6980 AND 6986) OR (index BETWEEN 7009 AND 7015) OR (index BETWEEN 9790 AND 9796) OR (index BETWEEN 3305 AND 3311) OR (index BETWEEN 3277 AND 3283) OR (index BETWEEN 4310 AND 4316) OR (index BETWEEN 6297 AND 6303) OR (index BETWEEN 11256 AND 11262) OR (index BETWEEN 199 AND 205) OR (index BETWEEN 5070 AND 5076) OR (index BETWEEN 7140 AND 7146) OR (index BETWEEN 9802 AND 9808) OR (index BETWEEN 11385 AND 11391) OR (index BETWEEN 3262 AND 3268) OR (index BETWEEN 11867 AND 11873) OR (index BETWEEN 12022 AND 12028) OR (index BETWEEN 12104 AND 12110) ) AND ( stripped_line ~* %s )  ORDER BY index ASC LIMIT 200', 'temptable': ''}, 'lt1038_3': {'data': '(^|\\s)maerori(\\s|$)', 'query': 'SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt1038 WHERE ( (index BETWEEN 1474 AND 1480) OR (index BETWEEN 11397 AND 11403) OR (index BETWEEN 822 AND 828) OR (index BETWEEN 5039 AND 5045) OR (index BETWEEN 2539 AND 2545) OR (index BETWEEN 5129 AND 5135) OR (index BETWEEN 2832 AND 2838) OR (index BETWEEN 11252 AND 11258) OR (index BETWEEN 2885 AND 2891) OR (index BETWEEN 11827 AND 11833) OR (index BETWEEN 8396 AND 8402) OR (index BETWEEN 5383 AND 5389) OR (index BETWEEN 11364 AND 11370) OR (index BETWEEN 384 AND 390) OR (index BETWEEN 6951 AND 6957) OR (index BETWEEN 11826 AND 11832) OR (index BETWEEN 3342 AND 3348) OR (index BETWEEN 1834 AND 1840) OR (index BETWEEN 542 AND 548) OR (index BETWEEN 6980 AND 6986) OR (index BETWEEN 7009 AND 7015) OR (index BETWEEN 9790 AND 9796) OR (index BETWEEN 3305 AND 3311) OR (index BETWEEN 3277 AND 3283) OR (index BETWEEN 4310 AND 4316) OR (index BETWEEN 6297 AND 6303) OR (index BETWEEN 11256 AND 11262) OR (index BETWEEN 199 AND 205) OR (index BETWEEN 5070 AND 5076) OR (index BETWEEN 7140 AND 7146) OR (index BETWEEN 9802 AND 9808) OR (index BETWEEN 11385 AND 11391) OR (index BETWEEN 3262 AND 3268) OR (index BETWEEN 11867 AND 11873) OR (index BETWEEN 12022 AND 12028) OR (index BETWEEN 12104 AND 12110) ) AND ( stripped_line ~* %s )  ORDER BY index ASC LIMIT 200', 'temptable': ''}, 'lt1038_4': {'data': '(^|\\s)maeroremq[uv]e(\\s|$)', 'query': 'SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt1038 WHERE ( (index BETWEEN 1474 AND 1480) OR (index BETWEEN 11397 AND 11403) OR (index BETWEEN 822 AND 828) OR (index BETWEEN 5039 AND 5045) OR (index BETWEEN 2539 AND 2545) OR (index BETWEEN 5129 AND 5135) OR (index BETWEEN 2832 AND 2838) OR (index BETWEEN 11252 AND 11258) OR (index BETWEEN 2885 AND 2891) OR (index BETWEEN 11827 AND 11833) OR (index BETWEEN 8396 AND 8402) OR (index BETWEEN 5383 AND 5389) OR (index BETWEEN 11364 AND 11370) OR (index BETWEEN 384 AND 390) OR (index BETWEEN 6951 AND 6957) OR (index BETWEEN 11826 AND 11832) OR (index BETWEEN 3342 AND 3348) OR (index BETWEEN 1834 AND 1840) OR (index BETWEEN 542 AND 548) OR (index BETWEEN 6980 AND 6986) OR (index BETWEEN 7009 AND 7015) OR (index BETWEEN 9790 AND 9796) OR (index BETWEEN 3305 AND 3311) OR (index BETWEEN 3277 AND 3283) OR (index BETWEEN 4310 AND 4316) OR (index BETWEEN 6297 AND 6303) OR (index BETWEEN 11256 AND 11262) OR (index BETWEEN 199 AND 205) OR (index BETWEEN 5070 AND 5076) OR (index BETWEEN 7140 AND 7146) OR (index BETWEEN 9802 AND 9808) OR (index BETWEEN 11385 AND 11391) OR (index BETWEEN 3262 AND 3268) OR (index BETWEEN 11867 AND 11873) OR (index BETWEEN 12022 AND 12028) OR (index BETWEEN 12104 AND 12110) ) AND ( stripped_line ~* %s )  ORDER BY index ASC LIMIT 200', 'temptable': ''}, 'lt1038_5': {'data': '(^|\\s)maerorib[uv]s(\\s|$)', 'query': 'SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt1038 WHERE ( (index BETWEEN 1474 AND 1480) OR (index BETWEEN 11397 AND 11403) OR (index BETWEEN 822 AND 828) OR (index BETWEEN 5039 AND 5045) OR (index BETWEEN 2539 AND 2545) OR (index BETWEEN 5129 AND 5135) OR (index BETWEEN 2832 AND 2838) OR (index BETWEEN 11252 AND 11258) OR (index BETWEEN 2885 AND 2891) OR (index BETWEEN 11827 AND 11833) OR (index BETWEEN 8396 AND 8402) OR (index BETWEEN 5383 AND 5389) OR (index BETWEEN 11364 AND 11370) OR (index BETWEEN 384 AND 390) OR (index BETWEEN 6951 AND 6957) OR (index BETWEEN 11826 AND 11832) OR (index BETWEEN 3342 AND 3348) OR (index BETWEEN 1834 AND 1840) OR (index BETWEEN 542 AND 548) OR (index BETWEEN 6980 AND 6986) OR (index BETWEEN 7009 AND 7015) OR (index BETWEEN 9790 AND 9796) OR (index BETWEEN 3305 AND 3311) OR (index BETWEEN 3277 AND 3283) OR (index BETWEEN 4310 AND 4316) OR (index BETWEEN 6297 AND 6303) OR (index BETWEEN 11256 AND 11262) OR (index BETWEEN 199 AND 205) OR (index BETWEEN 5070 AND 5076) OR (index BETWEEN 7140 AND 7146) OR (index BETWEEN 9802 AND 9808) OR (index BETWEEN 11385 AND 11391) OR (index BETWEEN 3262 AND 3268) OR (index BETWEEN 11867 AND 11873) OR (index BETWEEN 12022 AND 12028) OR (index BETWEEN 12104 AND 12110) ) AND ( stripped_line ~* %s )  ORDER BY index ASC LIMIT 200', 'temptable': ''}, 'lt1038_6': {'data': '(^|\\s)maeroris(\\s|$)', 'query': 'SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt1038 WHERE ( (index BETWEEN 1474 AND 1480) OR (index BETWEEN 11397 AND 11403) OR (index BETWEEN 822 AND 828) OR (index BETWEEN 5039 AND 5045) OR (index BETWEEN 2539 AND 2545) OR (index BETWEEN 5129 AND 5135) OR (index BETWEEN 2832 AND 2838) OR (index BETWEEN 11252 AND 11258) OR (index BETWEEN 2885 AND 2891) OR (index BETWEEN 11827 AND 11833) OR (index BETWEEN 8396 AND 8402) OR (index BETWEEN 5383 AND 5389) OR (index BETWEEN 11364 AND 11370) OR (index BETWEEN 384 AND 390) OR (index BETWEEN 6951 AND 6957) OR (index BETWEEN 11826 AND 11832) OR (index BETWEEN 3342 AND 3348) OR (index BETWEEN 1834 AND 1840) OR (index BETWEEN 542 AND 548) OR (index BETWEEN 6980 AND 6986) OR (index BETWEEN 7009 AND 7015) OR (index BETWEEN 9790 AND 9796) OR (index BETWEEN 3305 AND 3311) OR (index BETWEEN 3277 AND 3283) OR (index BETWEEN 4310 AND 4316) OR (index BETWEEN 6297 AND 6303) OR (index BETWEEN 11256 AND 11262) OR (index BETWEEN 199 AND 205) OR (index BETWEEN 5070 AND 5076) OR (index BETWEEN 7140 AND 7146) OR (index BETWEEN 9802 AND 9808) OR (index BETWEEN 11385 AND 11391) OR (index BETWEEN 3262 AND 3268) OR (index BETWEEN 11867 AND 11873) OR (index BETWEEN 12022 AND 12028) OR (index BETWEEN 12104 AND 12110) ) AND ( stripped_line ~* %s )  ORDER BY index ASC LIMIT 200', 'temptable': ''}, 'lt1038_7': {'data': '(^|\\s)maerorq[uv]e(\\s|$)', 'query': 'SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt1038 WHERE ( (index BETWEEN 1474 AND 1480) OR (index BETWEEN 11397 AND 11403) OR (index BETWEEN 822 AND 828) OR (index BETWEEN 5039 AND 5045) OR (index BETWEEN 2539 AND 2545) OR (index BETWEEN 5129 AND 5135) OR (index BETWEEN 2832 AND 2838) OR (index BETWEEN 11252 AND 11258) OR (index BETWEEN 2885 AND 2891) OR (index BETWEEN 11827 AND 11833) OR (index BETWEEN 8396 AND 8402) OR (index BETWEEN 5383 AND 5389) OR (index BETWEEN 11364 AND 11370) OR (index BETWEEN 384 AND 390) OR (index BETWEEN 6951 AND 6957) OR (index BETWEEN 11826 AND 11832) OR (index BETWEEN 3342 AND 3348) OR (index BETWEEN 1834 AND 1840) OR (index BETWEEN 542 AND 548) OR (index BETWEEN 6980 AND 6986) OR (index BETWEEN 7009 AND 7015) OR (index BETWEEN 9790 AND 9796) OR (index BETWEEN 3305 AND 3311) OR (index BETWEEN 3277 AND 3283) OR (index BETWEEN 4310 AND 4316) OR (index BETWEEN 6297 AND 6303) OR (index BETWEEN 11256 AND 11262) OR (index BETWEEN 199 AND 205) OR (index BETWEEN 5070 AND 5076) OR (index BETWEEN 7140 AND 7146) OR (index BETWEEN 9802 AND 9808) OR (index BETWEEN 11385 AND 11391) OR (index BETWEEN 3262 AND 3268) OR (index BETWEEN 11867 AND 11873) OR (index BETWEEN 12022 AND 12028) OR (index BETWEEN 12104 AND 12110) ) AND ( stripped_line ~* %s )  ORDER BY index ASC LIMIT 200', 'temptable': ''}, 'lt1038_8': {'data': '(^|\\s)maerorest(\\s|$)', 'query': 'SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt1038 WHERE ( (index BETWEEN 1474 AND 1480) OR (index BETWEEN 11397 AND 11403) OR (index BETWEEN 822 AND 828) OR (index BETWEEN 5039 AND 5045) OR (index BETWEEN 2539 AND 2545) OR (index BETWEEN 5129 AND 5135) OR (index BETWEEN 2832 AND 2838) OR (index BETWEEN 11252 AND 11258) OR (index BETWEEN 2885 AND 2891) OR (index BETWEEN 11827 AND 11833) OR (index BETWEEN 8396 AND 8402) OR (index BETWEEN 5383 AND 5389) OR (index BETWEEN 11364 AND 11370) OR (index BETWEEN 384 AND 390) OR (index BETWEEN 6951 AND 6957) OR (index BETWEEN 11826 AND 11832) OR (index BETWEEN 3342 AND 3348) OR (index BETWEEN 1834 AND 1840) OR (index BETWEEN 542 AND 548) OR (index BETWEEN 6980 AND 6986) OR (index BETWEEN 7009 AND 7015) OR (index BETWEEN 9790 AND 9796) OR (index BETWEEN 3305 AND 3311) OR (index BETWEEN 3277 AND 3283) OR (index BETWEEN 4310 AND 4316) OR (index BETWEEN 6297 AND 6303) OR (index BETWEEN 11256 AND 11262) OR (index BETWEEN 199 AND 205) OR (index BETWEEN 5070 AND 5076) OR (index BETWEEN 7140 AND 7146) OR (index BETWEEN 9802 AND 9808) OR (index BETWEEN 11385 AND 11391) OR (index BETWEEN 3262 AND 3268) OR (index BETWEEN 11867 AND 11873) OR (index BETWEEN 12022 AND 12028) OR (index BETWEEN 12104 AND 12110) ) AND ( stripped_line ~* %s )  ORDER BY index ASC LIMIT 200', 'temptable': ''}, 'lt1038_9': {'data': '(^|\\s)maeror[uv]m(\\s|$)', 'query': 'SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt1038 WHERE ( (index BETWEEN 1474 AND 1480) OR (index BETWEEN 11397 AND 11403) OR (index BETWEEN 822 AND 828) OR (index BETWEEN 5039 AND 5045) OR (index BETWEEN 2539 AND 2545) OR (index BETWEEN 5129 AND 5135) OR (index BETWEEN 2832 AND 2838) OR (index BETWEEN 11252 AND 11258) OR (index BETWEEN 2885 AND 2891) OR (index BETWEEN 11827 AND 11833) OR (index BETWEEN 8396 AND 8402) OR (index BETWEEN 5383 AND 5389) OR (index BETWEEN 11364 AND 11370) OR (index BETWEEN 384 AND 390) OR (index BETWEEN 6951 AND 6957) OR (index BETWEEN 11826 AND 11832) OR (index BETWEEN 3342 AND 3348) OR (index BETWEEN 1834 AND 1840) OR (index BETWEEN 542 AND 548) OR (index BETWEEN 6980 AND 6986) OR (index BETWEEN 7009 AND 7015) OR (index BETWEEN 9790 AND 9796) OR (index BETWEEN 3305 AND 3311) OR (index BETWEEN 3277 AND 3283) OR (index BETWEEN 4310 AND 4316) OR (index BETWEEN 6297 AND 6303) OR (index BETWEEN 11256 AND 11262) OR (index BETWEEN 199 AND 205) OR (index BETWEEN 5070 AND 5076) OR (index BETWEEN 7140 AND 7146) OR (index BETWEEN 9802 AND 9808) OR (index BETWEEN 11385 AND 11391) OR (index BETWEEN 3262 AND 3268) OR (index BETWEEN 11867 AND 11873) OR (index BETWEEN 12022 AND 12028) OR (index BETWEEN 12104 AND 12110) ) AND ( stripped_line ~* %s )  ORDER BY index ASC LIMIT 200', 'temptable': ''}, 'lt1038_10': {'data': '(^|\\s)maeroreq[uv]e(\\s|$)', 'query': 'SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt1038 WHERE ( (index BETWEEN 1474 AND 1480) OR (index BETWEEN 11397 AND 11403) OR (index BETWEEN 822 AND 828) OR (index BETWEEN 5039 AND 5045) OR (index BETWEEN 2539 AND 2545) OR (index BETWEEN 5129 AND 5135) OR (index BETWEEN 2832 AND 2838) OR (index BETWEEN 11252 AND 11258) OR (index BETWEEN 2885 AND 2891) OR (index BETWEEN 11827 AND 11833) OR (index BETWEEN 8396 AND 8402) OR (index BETWEEN 5383 AND 5389) OR (index BETWEEN 11364 AND 11370) OR (index BETWEEN 384 AND 390) OR (index BETWEEN 6951 AND 6957) OR (index BETWEEN 11826 AND 11832) OR (index BETWEEN 3342 AND 3348) OR (index BETWEEN 1834 AND 1840) OR (index BETWEEN 542 AND 548) OR (index BETWEEN 6980 AND 6986) OR (index BETWEEN 7009 AND 7015) OR (index BETWEEN 9790 AND 9796) OR (index BETWEEN 3305 AND 3311) OR (index BETWEEN 3277 AND 3283) OR (index BETWEEN 4310 AND 4316) OR (index BETWEEN 6297 AND 6303) OR (index BETWEEN 11256 AND 11262) OR (index BETWEEN 199 AND 205) OR (index BETWEEN 5070 AND 5076) OR (index BETWEEN 7140 AND 7146) OR (index BETWEEN 9802 AND 9808) OR (index BETWEEN 11385 AND 11391) OR (index BETWEEN 3262 AND 3268) OR (index BETWEEN 11867 AND 11873) OR (index BETWEEN 12022 AND 12028) OR (index BETWEEN 12104 AND 12110) ) AND ( stripped_line ~* %s )  ORDER BY index ASC LIMIT 200', 'temptable': ''}, 'lt1038_11': {'data': '(^|\\s)maeror(\\s|$)', 'query': 'SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt1038 WHERE ( (index BETWEEN 1474 AND 1480) OR (index BETWEEN 11397 AND 11403) OR (index BETWEEN 822 AND 828) OR (index BETWEEN 5039 AND 5045) OR (index BETWEEN 2539 AND 2545) OR (index BETWEEN 5129 AND 5135) OR (index BETWEEN 2832 AND 2838) OR (index BETWEEN 11252 AND 11258) OR (index BETWEEN 2885 AND 2891) OR (index BETWEEN 11827 AND 11833) OR (index BETWEEN 8396 AND 8402) OR (index BETWEEN 5383 AND 5389) OR (index BETWEEN 11364 AND 11370) OR (index BETWEEN 384 AND 390) OR (index BETWEEN 6951 AND 6957) OR (index BETWEEN 11826 AND 11832) OR (index BETWEEN 3342 AND 3348) OR (index BETWEEN 1834 AND 1840) OR (index BETWEEN 542 AND 548) OR (index BETWEEN 6980 AND 6986) OR (index BETWEEN 7009 AND 7015) OR (index BETWEEN 9790 AND 9796) OR (index BETWEEN 3305 AND 3311) OR (index BETWEEN 3277 AND 3283) OR (index BETWEEN 4310 AND 4316) OR (index BETWEEN 6297 AND 6303) OR (index BETWEEN 11256 AND 11262) OR (index BETWEEN 199 AND 205) OR (index BETWEEN 5070 AND 5076) OR (index BETWEEN 7140 AND 7146) OR (index BETWEEN 9802 AND 9808) OR (index BETWEEN 11385 AND 11391) OR (index BETWEEN 3262 AND 3268) OR (index BETWEEN 11867 AND 11873) OR (index BETWEEN 12022 AND 12028) OR (index BETWEEN 12104 AND 12110) ) AND ( stripped_line ~* %s )  ORDER BY index ASC LIMIT 200', 'temptable': ''}}


Sought »dolor« within 3 lines of »penates maeroris«
Searched 1 works and found 1 passage (0.17s)
Sorted by name
{'lt1038': {'temptable': '', 'query': "\n    SELECT second.wkuniversalid, second.index, second.level_05_value, second.level_04_value, second.level_03_value, second.level_02_value, second.level_01_value, second.level_00_value, second.marked_up_line, second.accented_line, second.stripped_line, second.hyphenated_words, second.annotations FROM\n        ( SELECT * FROM\n            ( SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations, concat(stripped_line, ' ', lead(stripped_line) OVER (ORDER BY index ASC) ) AS linebundle\n                FROM lt1038  ) first\n        ) second\n    WHERE second.linebundle ~ %s  LIMIT 2000000", 'data': ('penates maeroris',)}}
*/

/* debug
can't find "Amisit enim | filiam" @ pliny epp. book 5, letter 16, section 7
"enim filiam" is a great test because 3 of the 4 hits are at line ends
*/
