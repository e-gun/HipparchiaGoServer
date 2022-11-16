//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"bytes"
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

// JSStruct - this is for generating a specific ultra-boring brand of JSON that jQuery loves
type JSStruct struct {
	V string `json:"value"`
}

// RtGetJSSession - return the JSON for the session values for parsing by client JS
func RtGetJSSession(c echo.Context) error {
	// see hipparchiajs/coreinterfaceclicks_go.js

	user := readUUIDCookie(c)
	s := SafeSessionRead(user)

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
	jso.Browsercontext = i64s(s.BrowseCtx)
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

	return c.JSONPretty(http.StatusOK, jso, JSONINDENT)
}

// RtGetJSWorksOf - /get/json/worksof/lt0972 --> [{"value": "Satyrica (w001)"}, {"value": "Satyrica, fragmenta (w002)"}]
func RtGetJSWorksOf(c echo.Context) error {
	id := c.Param("id")
	wl := AllAuthors[id].WorkList
	tp := "%s (%s)"
	var titles []JSStruct
	for _, w := range wl {
		n := fmt.Sprintf(tp, AllWorks[w].Title, w[6:10])
		titles = append(titles, JSStruct{n})
	}

	return c.JSONPretty(http.StatusOK, titles, JSONINDENT)
}

// RtGetJSWorksStruct - lt0474/058 --> {"totallevels": 4, "level": 3, "label": "book", "low": "1", "high": "3", "range": ["1", "2", "3"]}
func RtGetJSWorksStruct(c echo.Context) error {
	// that is a top: interiors look like "1|3" for "book one", "subheading_val 3"

	locus := c.Param("locus")
	parsed := strings.Split(locus, "/")

	if len(parsed) < 2 || len(parsed) > 3 {
		return emptyjsreturn(c)
	}
	wkid := parsed[0] + "w" + parsed[1]

	if len(parsed) == 2 {
		parsed = append(parsed, "")
	}

	if _, ok := AllWorks[wkid]; !ok {
		return emptyjsreturn(c)
	}

	locc := strings.Split(parsed[2], "|")
	lvls := findvalidlevelvalues(wkid, locc)

	return c.JSONPretty(http.StatusOK, lvls, JSONINDENT)
}

// RtGetJSHelpdata - populate <div id="helptabs"> on frontpage.html via $('#helpbutton').click in documentready_go.js
func RtGetJSHelpdata(c echo.Context) error {
	cat := []string{"Interface", "Browsing", "Dictionaries", "MakingSearchLists", "BasicSyntax", "RegexSearching",
		"LemmaSearching", "Oddities", "Extending", "IncludedMaterials"}

	fm := make(map[string]string)
	fm["Browsing"] = "helpbrowsing.html"
	fm["Dictionaries"] = "helpdictionaries.html"
	fm["MakingSearchLists"] = "helpsearchlists.html"
	fm["BasicSyntax"] = "helpbasicsyntax.html"
	fm["RegexSearching"] = "helpregex.html"
	// fm["SpeedSearching"] = "helpspeed.html"
	fm["LemmaSearching"] = "helplemmata.html"
	// fm["VectorSearching"] = "helpvectors.html"
	fm["Oddities"] = "helpoddities.html"
	fm["Extending"] = "helpextending.html"
	fm["IncludedMaterials"] = "includedmaterials.html"
	// fm["Openness"] = "helpopenness.html"
	fm["Interface"] = "helpinterface.html"

	type JSOut struct {
		HC []string `json:"helpcategories"`
		HT map[string]string
	}

	hc := make(map[string]string)

	for k, v := range fm {
		b, e := efs.ReadFile("emb/h/" + v)
		chke(e)
		hc[k] = string(b)
	}

	var j JSOut
	j.HC = cat
	j.HT = hc

	return c.JSONPretty(http.StatusOK, j, JSONINDENT)

}

func RtGetJSAuthorinfo(c echo.Context) error {
	const (
		MTEMPL = `
		<span class="emph"><span class="emph">{{.Name}}</span></span>&nbsp;
		[id: {{.ID}}]<br>&nbsp;
		{{.Gen}}
		{{.DateCalc}} {{.DateRec}}
		<br>
		Total words: {{.TotalWd}}
		<br><br><span class="italic">work numbers:</span><br>`

		WTEMPL = `
		({{.ID}})&nbsp;
		<span class="title">{{.Title}}</span>
		[{{.Genre}}]&nbsp;
		[{{.WdCount}} wds]
		{{.PubInfo}}<br>`
	)

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

	mtt, e := template.New("mt").Parse(MTEMPL)
	chke(e)
	wtt, e := template.New("wt").Parse(WTEMPL)
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

	return c.JSONPretty(http.StatusOK, v, JSONINDENT)
}

func RtGetJSSampCit(c echo.Context) error {
	// in: http://localhost:5000/get/json/samplecitation/lt0474/001
	// out: {"firstline": "1.1", "lastline": "99.9"}
	locus := c.Param("locus")
	parsed := strings.Split(locus, "/")

	if len(parsed) < 2 || len(parsed) > 3 {
		return emptyjsreturn(c)
	}
	wkid := parsed[0] + "w" + parsed[1]

	if _, ok := AllWorks[wkid]; !ok {
		return emptyjsreturn(c)
	}

	w := AllWorks[wkid]
	// because "t" is going to be the first line's citation you have to hunt for the real place where the text starts
	ff := simplecontextgrabber(w.AuID(), w.FirstLine, 2)
	var actualfirst DbWorkline
	for i := len(ff) - 1; i > 0; i-- {
		loc := strings.Join(ff[i].FindLocus(), ".")
		if loc[0] != 't' && ff[i].TbIndex >= w.FirstLine {
			actualfirst = ff[i]
		}
	}
	l := graboneline(w.AuID(), w.LastLine)

	cf := strings.Join(actualfirst.FindLocus(), ".")
	cl := strings.Join(l.FindLocus(), ".")

	type JSO struct {
		F string `json:"firstline"`
		L string `json:"lastline"`
	}

	j := JSO{cf, cl}

	return c.JSONPretty(http.StatusOK, j, JSONINDENT)
}

// RtGetJSSearchlist - report the search list contents to the browser
func RtGetJSSearchlist(c echo.Context) error {
	const (
		WORKTMPL  = `%s, <span class="italic">%s</span> [%d words]`
		SPILLOVER = `<br>(and <span class="emph">%d</span> additional works)`
		SUMMARY   = `<br><span class="emph">%d</span> total words`
		REG       = `(?P<auth>......)_FROM_(?P<start>\d+)_TO_(?P<stop>\d+)`
	)

	user := readUUIDCookie(c)
	sess := SafeSessionRead(user)

	m := message.NewPrinter(language.English)
	sl := SessionIntoSearchlist(sess)
	tw := int64(0)

	var wkk []string
	for _, a := range sl.Inc.Authors {
		for _, w := range AllAuthors[a].WorkList {
			ct := WORKTMPL
			cf := m.Sprintf(ct, AllAuthors[a].Cleaname, AllWorks[w].Title, AllWorks[w].WdCount)
			wkk = append(wkk, cf)
			tw += AllWorks[w].WdCount
		}
	}

	for _, w := range sl.Inc.Works {
		thiswk := AllWorks[w]
		ct := WORKTMPL
		cf := m.Sprintf(ct, thiswk.MyAu().Cleaname, thiswk.Title, thiswk.WdCount)
		wkk = append(wkk, cf)
		tw += thiswk.WdCount
	}

	pattern := regexp.MustCompile(REG)
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
		wkk = append(wkk, m.Sprintf(SPILLOVER, diff))
	}

	wkk = append(wkk, m.Sprintf(SUMMARY, tw))

	ht := strings.Join(wkk, "<br>\n")
	var j JSStruct
	j.V = ht

	return c.JSONPretty(http.StatusOK, j, JSONINDENT)
}

func searchlistpassages(pattern *regexp.Regexp, p string) (string, int) {
	const (
		PSGTEMPL = `%s, <span class="italic">%s</span> %s - %s [%d words]`
	)
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
	SSBuildQueries(&s)
	lines := HGoSrch(s)
	count := 0
	for _, ln := range lines.Results {
		count += len(strings.Split(ln.Stripped, " "))
	}
	ct := m.Sprintf(PSGTEMPL, AllAuthors[au].Cleaname, AllWorks[f.WkUID].Title, f.Citation(), l.Citation(), count)
	return ct, count
}
