package main

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

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
	jso.Proximity = is(s.HitContext)
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
		new := fmt.Sprintf(tp, AllWorks[w].Title, w[6:10])
		titles = append(titles, JSStruct{new})
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

	// TODO: input sanitization

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
