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
	NeedsWhere bool
	TwoPart    bool
	SrchColumn string // usually "stripped_line", sometimes "accented_line"
	SrchSyntax string // almost always "~="
	OrderBy    string // almost always "index" + ASC
	Limit      int64
	SkgSlice   []string // either just Seeking or a decomposed VERSION of a Lemma's possibilities
	PrxSlice   []string
	SearchIn   SearchIncExl
	SearchEx   SearchIncExl
	Queries    []PrerolledQuery
	Results    []DbWorkline
	Launched   time.Time
	TTName     string
	SearchSize int
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

//func (s SearchStruct) FmtWhereTerm(t string) string {
//	a := `%s %s '%s' `
//	wht := fmt.Sprintf(a, s.SrchColumn, s.SrchSyntax, t)
//	return wht
//}

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
	if searches[id].QueryType == "proximity" {
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
	s.TTName = uuid.New().String()
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
	//	CREATE TEMPORARY TABLE in110f_includelist_UNIQUENAME AS
	//        SELECT values
	//            AS includeindex FROM unnest(ARRAY[16197,16198,16199,16200,16201,16202,16203,16204,16205,16206,16207]) values
	// (and then)
	//     SELECT second.wkuniversalid, second.index, second.level_05_value, second.level_04_value, second.level_03_value, second.level_02_value, second.level_01_value, second.level_00_value, second.marked_up_line, second.accented_line, second.stripped_line, second.hyphenated_words, second.annotations FROM
	//        ( SELECT * FROM
	//            ( SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations, concat(accented_line, ' ', lead(accented_line) OVER (ORDER BY index ASC)
	//            ) AS linebundle
	//                FROM in110f WHERE EXISTS
	//                ( SELECT 1 FROM in110f_includelist_UNIQUENAME incl WHERE incl.includeindex = in110f.index ) ) first
	//        ) second
	//    WHERE second.linebundle ~ 'ἀνέθηκεν ὑπὲρ'  LIMIT 200;

	first := s
	first.Limit = FIRSTSEARCHLIM
	first = HGoSrch(first)

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

	// prepare new searchlists
	second := first
	second.Results = []DbWorkline{}
	second.Queries = []PrerolledQuery{}
	second.TTName = uuid.New().String()
	second.SkgSlice = second.PrxSlice
	second.PrxSlice = first.SkgSlice

	var ttsq = make(map[string]string)
	ctt := `
	CREATE TEMPORARY TABLE %s_includelist_%s AS 
		SELECT values AS includeindex FROM 
			unnest(ARRAY[%s])
		values`

	for r, vv := range required {
		var arr []string
		for v := range vv {
			arr = append(arr, strconv.FormatInt(int64(v), 10))
		}
		a := strings.Join(arr, ",")
		ttsq[r] = fmt.Sprintf(ctt, r, second.TTName, a)
	}

	seltempl := PRFXSELFRM + CONCATSELFROM

	wha := `
	WHERE EXISTS 
		(SELECT 1 FROM %s_includelist_%s incl 
			WHERE incl.includeindex = %s.index AND %s.accented_line %s '%s')`
	whb := `WHERE second.linebundle ~ '%s' LIMIT %d;`
	var prqq = make(map[string][]PrerolledQuery)
	for _, q := range second.SkgSlice {
		for r, _ := range required {
			var prq PrerolledQuery
			prq.TempTable = ttsq[r]
			whc := fmt.Sprintf(wha, r, second.TTName, r, r, second.SrchSyntax, q)
			whd := fmt.Sprintf(whb, q, second.Limit)
			prq.PsqlQuery = fmt.Sprintf(seltempl, whc) + whd
			prqq[r] = append(prqq[r], prq)
		}
	}

	second = HGoSrch(second)
	return second
}
