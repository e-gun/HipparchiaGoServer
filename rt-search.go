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
	srch.IsVector = false

	sl := sessionintosearchlist(sessions[user])
	srch.SearchIn = sl.Inc
	srch.SearchEx = sl.Excl
	srch.SearchSize = sl.Size

	timetracker("B", "sessionintosearchlist()", start, previous)
	previous = time.Now()

	// must happen before searchlistintoqueries()
	srch = setsearchtype(srch)

	if srch.LemmaOne != "" {
		srch.SkgSlice = lemmaintoregex(srch.LemmaOne)
	} else {
		srch.SkgSlice = append(srch.SkgSlice, srch.Seeking)
	}

	if srch.LemmaTwo != "" {
		srch.PrxSlice = lemmaintoregex(srch.LemmaTwo)
	} else {
		srch.PrxSlice = append(srch.PrxSlice, srch.Proximate)
	}

	prq := searchlistintoqueries(srch)
	timetracker("C", "searchlistintoqueries()", start, previous)
	previous = time.Now()

	srch.Queries = prq

	searches[id] = srch

	// return results via searches[id].Results
	if searches[id].Twobox {
		// this can do word + lemma; double lemmata; phrase + phrase...
		// todo: "not near" syntax
		searches[id] = withinxlinessearch(searches[id])
	} else {
		searches[id] = HGoSrch(searches[id])
	}

	timetracker("D", "HGoSrch()", start, previous)
	previous = time.Now()

	//hits := searches[id].Results
	//for i, h := range hits {
	//	t := fmt.Sprintf("%d - %srch : %srch", i, h.FindLocus(), h.MarkedUp)
	//	fmt.Println(t)
	//}

	timetracker("E", fmt.Sprintf("search executed: %d hits", len(searches[id].Results)), start, previous)

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
		Sought %s<span class="sought">»%s«</span>
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

	af := ""
	sk := s.Seeking
	if s.LemmaOne != "" {
		af = "all forms of "
		sk = s.LemmaOne
	}

	so := sessions[s.User].SrchOutSettings.SortHitsBy
	el := fmt.Sprintf("%.3f", time.Now().Sub(s.Launched).Seconds())
	// need to record # of works and not # of tables somewhere & at the right moment...
	sum := fmt.Sprintf(t, af, sk, s.SearchSize, len(s.Results), el, so, dr, hitcap)
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

func lemmaintoregex(hdwd string) []string {
	// rather than do one word per query, bundle things up: some words have >100 forms
	// ...(^|\\s)ἐδηλώϲαντο(\\s|$)|(^|\\s)δεδηλωμένοϲ(\\s|$)|(^|\\s)δήλουϲ(\\s|$)|(^|\\s)δηλούϲαϲ(\\s|$)...
	var qq []string
	if _, ok := AllLemm[hdwd]; !ok {
		msg(fmt.Sprintf("lemmaintoregex() could not find '%s'", hdwd), 1)
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

func withinxlinessearch(s SearchStruct) SearchStruct {
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

	// alternate: ... FROM lt0474 WHERE ( (index BETWEEN 128860 AND 128866) OR (index BETWEEN 39846 AND 39852) OR ... )

	first := s
	first.Limit = FIRSTSEARCHLIM
	first = HGoSrch(first)

	// fmt.Printf("%d initial hits\n", len(first.Results))

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

	var ttsq = make(map[string]string)
	ctt := `
	CREATE TEMPORARY TABLE %s_includelist_%s AS 
		SELECT values AS includeindex FROM 
			unnest(ARRAY[%s])
		values`

	for r, vv := range required {
		var arr []string
		for _, v := range vv {
			arr = append(arr, strconv.FormatInt(int64(v), 10))
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
	for _, q := range second.SkgSlice {
		for r, _ := range required {
			var prq PrerolledQuery
			prq.TempTable = ttsq[r]
			whc := fmt.Sprintf(wha, r, r, second.TTName, r)
			whd := fmt.Sprintf(whb, q, second.Limit)
			prq.PsqlQuery = fmt.Sprintf(seltempl, whc) + whd
			prqq[r] = append(prqq[r], prq)
		}
	}

	for _, q := range prqq {
		second.Queries = append(second.Queries, q...)
	}

	second = HGoSrch(second)
	return second
}
