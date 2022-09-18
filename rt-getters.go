//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

// JSStruct - this is for generating a specific ultra-boring brand of JSON
type JSStruct struct {
	V string `json:"value"`
}

func RtGetJSSession(c echo.Context) error {
	// see hipparchiajs/coreinterfaceclicks_go.js

	user := readUUIDCookie(c)
	if _, exists := sessions[user]; !exists {
		sessions[user] = makedefaultsession(user)
	}
	s := sessions[user]

	type JSO struct {
		// what the JS is looking for; note that vector stuff, etc is being skipped vs the python session dump
		Browsercontext    string `json:"browsercontext"`
		Christiancorpus   string `json:"christiancorpus"`
		Earliestdate      string `json:"earliestdate"`
		Greekcorpus       string `json:"greekcorpus"`
		Headwordindexing  string `json:"headwordindexing"`
		Incerta           string `json:"incerta"`
		Indexbyfrequency  string `json:"indexbyfrequency"`
		Inscriptioncorpus string `json:"inscriptioncorpus"`
		Latestdate        string `json:"latestdate"`
		Latincorpus       string `json:"latincorpus"`
		Linesofcontext    string `json:"linesofcontext"`
		Maxresults        string `json:"maxresults"`
		Nearornot         string `json:"nearornot"`
		Onehit            string `json:"onehit"`
		Papyruscorpus     string `json:"papyruscorpus"`
		Proximity         string `json:"proximity"`
		Rawinputstyle     string `json:"rawinputstyle"`
		Searchscope       string `json:"searchscope"`
		Sortorder         string `json:"sortorder"`
		Spuria            string `json:"spuria"`
		Varia             string `json:"varia"`
	}

	t2y := func(b bool) string {
		if b {
			return "yes"
		} else {
			return "no"
		}
	}
	i64s := func(i int64) string { return fmt.Sprintf("%d", i) }
	is := func(i int) string { return fmt.Sprintf("%d", i) }

	var jso JSO
	jso.Browsercontext = i64s(s.UI.BrowseCtx)
	jso.Christiancorpus = t2y(s.ActiveCorp["ch"])
	jso.Earliestdate = s.Earliest
	jso.Greekcorpus = t2y(s.ActiveCorp["gr"])
	jso.Headwordindexing = t2y(s.HeadwordIdx)
	jso.Incerta = t2y(s.IncertaOK)
	jso.Indexbyfrequency = t2y(s.FrqIdx)
	jso.Inscriptioncorpus = t2y(s.ActiveCorp["in"])
	jso.Latestdate = s.Latest
	jso.Latincorpus = t2y(s.ActiveCorp["lt"])
	jso.Linesofcontext = is(s.HitContext)
	jso.Maxresults = i64s(s.HitLimit)
	jso.Nearornot = s.NearOrNot
	jso.Papyruscorpus = t2y(s.ActiveCorp["dp"])
	jso.Proximity = is(s.Proximity)
	jso.Rawinputstyle = t2y(s.RawInput)
	jso.Searchscope = s.SearchScope
	jso.Sortorder = s.SortHitsBy
	jso.Spuria = t2y(s.SpuriaOK)
	jso.Varia = t2y(s.VariaOK)

	o, e := json.Marshal(jso)
	chke(e)
	return c.String(http.StatusOK, string(o))
}

func RtGetJSWorksOf(c echo.Context) error {
	// curl localhost:5000/get/json/worksof/lt0972
	//[{"value": "Satyrica (w001)"}, {"value": "Satyrica, fragmenta (w002)"}]
	id := c.Param("id")
	wl := AllAuthors[id].WorkList
	tp := "%s (%s)"
	var titles []JSStruct
	for _, w := range wl {
		n := fmt.Sprintf(tp, AllWorks[w].Title, w[6:10])
		titles = append(titles, JSStruct{n})
	}

	// send
	b, e := json.Marshal(titles)
	chke(e)
	// fmt.Printf("RtGetJSWorksOf():\n\t%s\n", b)
	return c.String(http.StatusOK, string(b))
}

func RtGetJSWorksStruct(c echo.Context) error {
	// curl localhost:5000/get/json/workstructure/lt0474/058
	//{"totallevels": 4, "level": 3, "label": "book", "low": "1", "high": "3", "range": ["1", "2", "3"]}
	// that is a top: interiors look like "1|3" for "book one", "subheading_val 3"

	locus := c.Param("locus")
	parsed := strings.Split(locus, "/")

	if len(parsed) < 2 || len(parsed) > 3 {
		return c.String(http.StatusOK, "")
	}
	wkid := parsed[0] + "w" + parsed[1]

	if len(parsed) == 2 {
		parsed = append(parsed, "")
	}

	if _, ok := AllWorks[wkid]; !ok {
		return c.String(http.StatusOK, "")
	}

	locc := strings.Split(parsed[2], "|")
	lvls := findvalidlevelvalues(wkid, locc)

	// send
	b, e := json.Marshal(lvls)
	chke(e)
	// fmt.Printf("RtGetJSWorksStruct():\n\t%s\n", b)
	return c.String(http.StatusOK, string(b))
}

func RtGetJSHelpdata(c echo.Context) error {
	// needs to return:
	//{"helpcategories": ["Interface", "Browsing", "Dictionaries", "MakingSearchLists", "BasicSyntax", "RegexSearching",
	// "SpeedSearching", "LemmaSearching", "VectorSearching", "Oddities", "Extending", "IncludedMaterials", "Openness"],
	// "Interface": "(the_html)", "Browsing": "(the_html)", ...}
	msg("called empty placeholder function: RtGetJSHelpdata()", 1)
	return c.String(http.StatusOK, "")
}

func RtGetJSAuthorinfo(c echo.Context) error {
	type AUTempl struct {
		Name     string
		ID       string
		Gen      string
		DateRec  string
		DateCalc string
		TotalWd  string
	}

	type WKTempl struct {
		ID      string
		Title   string
		Genre   string
		WdCount string
		PubInfo string
	}

	id := c.Param("id")
	au := AllAuthors[id]

	var at AUTempl
	at.Name = au.Name
	at.ID = au.UID
	at.Gen = au.Genres

	if len(at.Gen) != 0 {
		at.Gen = fmt.Sprintf("classified among: %s;", at.Gen)
	}

	if au.ConvDate != 2500 {
		at.DateCalc = fmt.Sprintf("assigned to approx date: %s ", i64tobce(au.ConvDate))
	} else {
		at.DateCalc = "(date is unavalable)"
	}

	if au.RecDate == "Unavailable" {
		at.DateRec = ""
	} else {
		at.DateRec = fmt.Sprintf(`(derived from "%s")`, au.RecDate)
	}

	var ww []WKTempl
	var twc int64
	p := message.NewPrinter(language.English)

	for _, w := range au.WorkList {
		ws := AllWorks[w]
		var wt WKTempl
		wt.ID = ws.UID[7:]
		wt.Title = ws.Title
		wt.Genre = ws.Genre
		wt.WdCount = p.Sprintf("%d", ws.WdCount)
		wt.PubInfo = formatpublicationinfo(ws)
		ww = append(ww, wt)
		twc += ws.WdCount
	}

	at.TotalWd = p.Sprintf("%d", twc)

	sort.Slice(ww, func(i, j int) bool { return ww[i].ID < ww[j].ID })

	mt := `
    <span class="emph"><span class="emph">{{.Name}}</span></span>&nbsp;
    [id: {{.ID}}]<br>&nbsp;
    {{.Gen}}
    {{.DateCalc}} {{.DateRec}}
	<br>
	Total words: {{.TotalWd}}
	<br><br><span class="italic">work numbers:</span><br>`

	wt := `({{.ID}})&nbsp;
		<span class="title">{{.Title}}</span>
		[{{.Genre}}]&nbsp;
		[{{.WdCount}} wds]
		{{.PubInfo}}<br>`

	mtt, e := template.New("mt").Parse(mt)
	chke(e)
	wtt, e := template.New("wt").Parse(wt)
	chke(e)

	var b bytes.Buffer
	err := mtt.Execute(&b, at)
	chke(err)
	for _, w := range ww {
		err = wtt.Execute(&b, w)
		chke(err)
	}

	info := b.String()

	v := JSStruct{info}
	j, e := json.Marshal(v)
	chke(e)

	return c.String(http.StatusOK, string(j))
}

func RtGetJSSampCit(c echo.Context) error {
	// in: http://localhost:5000/get/json/samplecitation/lt0474/001
	// out: {"firstline": "1.1", "lastline": "99.9"}
	locus := c.Param("locus")
	parsed := strings.Split(locus, "/")

	if len(parsed) < 2 || len(parsed) > 3 {
		return c.String(http.StatusOK, "")
	}
	wkid := parsed[0] + "w" + parsed[1]

	if _, ok := AllWorks[wkid]; !ok {
		return c.String(http.StatusOK, "")
	}

	w := AllWorks[wkid]
	// because "t" is going to be the first line's citation you have to hunt for the real place where the text starts
	ff := simplecontextgrabber(w.FindAuthor(), w.FirstLine, 2)
	var actualfirst DbWorkline
	for i := len(ff) - 1; i > 0; i-- {
		loc := strings.Join(ff[i].FindLocus(), ".")
		if loc[0] != 't' && ff[i].TbIndex >= w.FirstLine {
			actualfirst = ff[i]
		}
	}
	l := graboneline(w.FindAuthor(), w.LastLine)

	cf := strings.Join(actualfirst.FindLocus(), ".")
	cl := strings.Join(l.FindLocus(), ".")

	type JSO struct {
		F string `json:"firstline"`
		L string `json:"lastline"`
	}
	j := JSO{cf, cl}
	b, e := json.Marshal(j)
	chke(e)
	return c.String(http.StatusOK, string(b))
}

func RtGetJSSearchlist(c echo.Context) error {
	m := message.NewPrinter(language.English)
	sl := sessionintosearchlist(sessions[readUUIDCookie(c)])
	tw := int64(0)

	var wkk []string
	for _, a := range sl.Inc.Authors {
		for _, w := range AllAuthors[a].WorkList {
			ct := `%s, <span class="italic">%s</span> [%d words]`
			cf := m.Sprintf(ct, AllAuthors[a].Cleaname, AllWorks[w].Title, AllWorks[w].WdCount)
			wkk = append(wkk, cf)
			tw += AllWorks[w].WdCount
		}
	}

	for _, w := range sl.Inc.Works {
		ct := `%s, <span class="italic">%s</span> [%d words]`
		cf := m.Sprintf(ct, AllAuthors[AllWorks[w].FindAuthor()].Cleaname, AllWorks[w].Title, AllWorks[w].WdCount)
		wkk = append(wkk, cf)
		tw += AllWorks[w].WdCount
	}

	pattern := regexp.MustCompile(`(?P<auth>......)_FROM_(?P<start>\d+)_TO_(?P<stop>\d+)`)
	for _, p := range sl.Inc.Passages {
		cit, count := searchlistpassages(pattern, p)
		wkk = append(wkk, cit)
		tw += int64(count)
	}

	for _, p := range sl.Excl.Passages {
		cit, count := searchlistpassages(pattern, p)
		wkk = append(wkk, cit+"[EXCLUDED]")
		tw -= int64(count)
	}

	if len(wkk) > MAXSEARCHINFOLISTLEN {
		diff := len(wkk) - MAXSEARCHINFOLISTLEN
		wkk = wkk[0:MAXSEARCHINFOLISTLEN]
		wkk = append(wkk, m.Sprintf(`<br>(and <span class="emph">%d</span> additional works)`, diff))
	}

	wkk = append(wkk, m.Sprintf(`<br><span class="emph">%d</span> total words`, tw))

	ht := strings.Join(wkk, "<br>\n")
	var j JSStruct
	j.V = ht

	// send
	b, e := json.Marshal(j)
	chke(e)
	return c.String(http.StatusOK, string(b))
}

func searchlistpassages(pattern *regexp.Regexp, p string) (string, int) {
	// "gr0032_FROM_11313_TO_11843"
	m := message.NewPrinter(language.English)
	subs := pattern.FindStringSubmatch(p)
	au := subs[pattern.SubexpIndex("auth")]
	st, _ := strconv.Atoi(subs[pattern.SubexpIndex("start")])
	sp, _ := strconv.Atoi(subs[pattern.SubexpIndex("stop")])
	f := graboneline(au, int64(st))
	l := graboneline(au, int64(sp))
	s := buildhollowsearch()
	s.SearchIn.Passages = []string{p}
	s.Queries = searchlistintoqueries(&s)
	lines := HGoSrch(s)
	count := 0
	for _, ln := range lines.Results {
		count += len(strings.Split(ln.Stripped, " "))
	}
	ct := m.Sprintf(`%s, <span class="italic">%s</span> %s - %s [%d words]`, AllAuthors[au].Cleaname, AllWorks[f.WkUID].Title, f.Citation(), l.Citation(), count)
	return ct, count
}
