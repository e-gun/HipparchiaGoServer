package main

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {

	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

// JSStruct - this is for generating a specific brand of JSON
type JSStruct struct {
	V string `json:"value"`
}

func StartEchoServer() {
	// https://echo.labstack.com/guide/
	// cf https://medium.com/cuddle-ai/building-microservice-using-golang-echo-framework-ff10ba06d508

	e := echo.New()

	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("static/html/frontpage.html")),
	}
	e.Renderer = renderer

	if cfg.EchoLog == 2 {
		e.Use(middleware.Logger())
	} else if cfg.EchoLog == 1 {
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{Format: "status: ${status}\turi: ${uri}\n"}))
	}

	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 5}))

	e.File("/favicon.ico", "static/images/hipparchia_favicon.ico")
	e.Static("/static", "static")

	// hipparchia routes

	//
	// [a] authentication
	//

	// [a1] '/authentication/attemptlogin'
	// [a2] '/authentication/logout'
	// [a3] '/authentication/checkuser'
	e.GET("/authentication/checkuser", RtAuthChkuser)

	//
	// [b] browser
	//

	// [b1] sample input: '/browse/linenumber/lt0550/001/1855'
	e.GET("/browse/linenumber/:locus", RtBrowseline)

	// [b2] sample input: '/browse/locus/lt0550/001/3|100'
	e.GET("/browse/locus/:locus", RtBrowseLocus)

	// [b3] sample input: '/browse/perseus/lt0550/001/2:717'
	e.GET("/browse/perseus/:locus", RtBrowsePerseus)

	// [b4] sample input: '/browse/rawlocus/lt0474/037/2.10.4'

	// [c] css
	// [d] debugging

	//
	// [e] frontpage
	//

	e.GET("/", RtFrontpage)

	//
	// [f] getters
	//

	// [f1a] /get/response/cookie
	// [f1b] /get/response/vectorfigure
	// [f2a] /get/json/sessionvariables
	e.GET("/get/json/sessionvariables", RtGetJSSession)

	// [f2b] /get/json/worksof
	e.GET("/get/json/worksof/:id", RtGetJSWorksOf)

	// [f2c] /get/json/workstructure
	e.GET("/get/json/workstructure/:locus", RtGetJSWorksStruct)

	// [f2d] /get/json/samplecitation
	// [f2e] /get/json/authorinfo
	// [f2f] /get/json/searchlistcontents
	// [f2e] /get/json/genrelistcontents
	// [f2f] /get/json/vectorranges
	// [f2g] /get/json/helpdata

	//
	// [g] hinters
	//

	// [g1] "GET /hints/author/_?term=au HTTP/1.1"
	e.GET("/hints/author/:id", RtAuthorHints)

	// [g2] authorgenre
	// [g3] workgenre
	// [g4] authorlocation
	// [g5] worklocation
	// [g6] lemmata: "GET http://localhost:8000/hints/lemmata/_?term=dol"
	e.GET("/hints/lemmata/:id", RtLemmaHints)

	//
	// [h] lexical
	//

	// [h1]
	// [h2] GET http://localhost:8000/lexica/findbyform/sapientem/lt0474
	e.GET("/lexica/findbyform/:id", RtLexFindByForm)

	// [h3]
	// [h4]

	//
	// [i] resets
	//

	// [i1] /reset/session
	e.GET("/reset/session", RtResetSession)
	// [i2] /reset/vectors
	// [i3] /reset/vectorimages

	//
	// [j] searching
	//

	// [j1] confirm: "GET /search/confirm/1f8f1d22 HTTP/1.1"
	e.GET("/search/confirm/:id", RtSearchConfirm)

	// [j2] standard: "GET /search/standard/1f8f1d22?skg=dolor HTTP/1.1"
	e.GET("/search/standard/:id", RtSearchStandard)

	// [j3] singleword
	// [j4] lemmatized

	//
	// [k] selection
	//

	// [k1] "GET /selection/make/_?auth=gr7000 HTTP/1.1"
	e.GET("/selection/make/:locus", RtSelectionMake)

	// [k2] "GET /selection/clear/auselections/0 HTTP/1.1"
	e.GET("/selection/clear/:locus", RtSelectionClear)

	// [k3] "GET /selection/fetch HTTP/1.1"
	e.GET("/selection/fetch", RtSelectionFetch)

	// [l] text and index
	// [m] vectors

	// [z] testing
	e.GET("/t", RtTest)

	e.Logger.Fatal(e.Start(":8000"))
}

func RtAuthChkuser(c echo.Context) error {
	return c.String(http.StatusOK, "")
}

func RtGetJSSession(c echo.Context) error {
	// see hipparchiajs/coreinterfaceclicks_go.js
	// python sample: {"_fresh": "no", "agnexclusions": [], "agnselections": [], "alocexclusions": [], "alocselections": [], "analogyfinder": "no", "auexclusions": [], "auselections": ["gr7000"], "authorflagging": "yes", "authorssummary": "yes", "available": {"greek_dictionary": true, "greek_lemmata": true, "greek_morphology": true, "latin_dictionary": true, "latin_lemmata": true, "latin_morphology": true, "wordcounts_0": true}, "baggingmethod": "winnertakesall", "bracketangled": "yes", "bracketcurly": "yes", "bracketround": "no", "bracketsquare": "yes", "browsercontext": "24", "christiancorpus": "no", "collapseattic": "yes", "cosdistbylineorword": "no", "cosdistbysentence": "no", "debugdb": "no", "debughtml": "no", "debuglex": "no", "debugparse": "no", "earliestdate": "-850", "fontchoice": "Noto", "greekcorpus": "yes", "headwordindexing": "no", "incerta": "yes", "indexbyfrequency": "no", "indexskipsknownwords": "no", "inscriptioncorpus": "no", "latestdate": "1500", "latincorpus": "yes", "ldacomponents": 7, "ldaiterations": 12, "ldamaxfeatures": 2000, "ldamaxfreq": 35, "ldaminfreq": 5, "ldamustbelongerthan": 3, "linesofcontext": 4, "loggedin": "no", "maxresults": "200", "morphdialects": "no", "morphduals": "yes", "morphemptyrows": "yes", "morphfinite": "yes", "morphimper": "yes", "morphinfin": "yes", "morphpcpls": "yes", "morphtables": "yes", "nearestneighborsquery": "no", "nearornot": "near", "onehit": "no", "papyruscorpus": "no", "phrasesummary": "no", "principleparts": "yes", "proximity": "1", "psgexclusions": [], "psgselections": [], "quotesummary": "yes", "rawinputstyle": "no", "searchinsidemarkup": "no", "searchscope": "lines", "semanticvectorquery": "no", "sensesummary": "yes", "sentencesimilarity": "no", "showwordcounts": "yes", "simpletextoutput": "no", "sortorder": "SHORTNAME", "spuria": "yes", "suppresscolors": "no", "tensorflowgraph": "no", "topicmodel": "no", "trimvectoryby": "none", "userid": "Anonymous", "varia": "yes", "vcutlem": 50, "vcutloc": 33, "vcutneighb": 15, "vdim": 300, "vdsamp": 5, "viterat": 12, "vminpres": 10, "vnncap": 15, "vsentperdoc": 1, "vwindow": 10, "wkexclusions": [], "wkgnexclusions": [], "wkgnselections": [], "wkselections": [], "wlocexclusions": [], "wlocselections": [], "xmission": "Any", "zaplunates": "no", "zapvees": "no"}

	user := readUUIDCookie(c)
	if _, exists := sessions[user]; !exists {
		sessions[user] = makedefaultsession(user)
	}

	type JSO struct {
	}

	s := sessions[user]
	o, e := json.Marshal(s)
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

	locc := strings.Split(parsed[2], "|")
	lvls := findvalidlevelvalues(wkid, locc)

	// send
	b, e := json.Marshal(lvls)
	chke(e)
	// fmt.Printf("RtGetJSWorksStruct():\n\t%s\n", b)
	return c.String(http.StatusOK, string(b))
}

func RtResetSession(c echo.Context) error {
	// delete my session
	delete(sessions, readUUIDCookie(c))

	// then reset it
	readUUIDCookie(c)
	return RtFrontpage(c)
}

func RtSelectionMake(c echo.Context) error {
	// GET http://localhost:8000/selection/make/_?auth=lt0474&work=073&locus=3|10&endpoint=

	// note that you need to return JSON: reportcurrentselections() so as to fill #selectionstable on the page

	user := readUUIDCookie(c)
	var sel SelectValues
	sel.Auth = c.QueryParam("auth")
	sel.Work = c.QueryParam("work")
	sel.Start = c.QueryParam("locus")
	sel.End = c.QueryParam("endpoint")
	sel.AGenre = c.QueryParam("genre")
	sel.WGenre = c.QueryParam("wkgenre")
	sel.ALoc = c.QueryParam("auloc")
	sel.WLoc = c.QueryParam("wkprov")

	if c.QueryParam("raw") == "t" {
		sel.IsRaw = true
	} else {
		sel.IsRaw = false
	}

	if c.QueryParam("exclude") == "t" {
		sel.IsExcl = true
	} else {
		sel.IsExcl = false
	}

	sessions[user] = selected(user, sel)
	jsbytes := reportcurrentselections(c)

	fmt.Println(string(jsbytes))

	return c.String(http.StatusOK, string(jsbytes))
}

func RtSelectionClear(c echo.Context) error {
	// GET http://localhost:8000/selection/clear/wkselections/0
	user := readUUIDCookie(c)

	locus := c.Param("locus")
	which := strings.Split(locus, "/")

	if len(which) != 2 {
		msg(fmt.Sprintf("RtSelectionClear() was given bad input: %s", locus), 1)
		return c.String(http.StatusOK, "")
	}

	cat := which[0]
	id, e := strconv.Atoi(which[1])
	if e != nil {
		msg(fmt.Sprintf("RtSelectionClear() was given bad input: %s", locus), 1)
		return c.String(http.StatusOK, "")
	}

	// cat := []string{"agn", "wgn", "aloc", "wloc", "au", "wk", "psg"}

	mod := sessions[user]
	modi := mod.Inclusions
	mode := mod.Exclusions

	switch cat {
	case "agnselections":
		modi.AuGenres = RemoveIndex(modi.AuGenres, id)
	case "wgnselections":
		modi.WkGenres = RemoveIndex(modi.WkGenres, id)
	case "alocselections":
		modi.AuLocations = RemoveIndex(modi.AuLocations, id)
	case "wlocselections":
		modi.WkLocations = RemoveIndex(modi.WkLocations, id)
	case "auselections":
		modi.Authors = RemoveIndex(modi.Authors, id)
	case "wkselections":
		modi.Works = RemoveIndex(modi.Works, id)
	case "psgselections":
		modi.Passages = RemoveIndex(modi.Passages, id)
	case "agnexclusions":
		mode.AuGenres = RemoveIndex(mode.AuGenres, id)
	case "wgnexclusions":
		mode.WkGenres = RemoveIndex(mode.WkGenres, id)
	case "alocexclusions":
		mode.AuLocations = RemoveIndex(mode.AuLocations, id)
	case "wlocexclusions":
		mode.WkLocations = RemoveIndex(mode.WkLocations, id)
	case "auexclusions":
		mode.Authors = RemoveIndex(mode.Authors, id)
	case "wkexclusions":
		mode.Works = RemoveIndex(mode.Works, id)
	case "psgexclusions":
		mode.Passages = RemoveIndex(mode.Passages, id)
	default:
		msg(fmt.Sprintf("RtSelectionClear() was given bad category: %s", cat), 1)
	}

	delete(sessions, user)
	sessions[user] = mod

	r := RtSelectionFetch(c)

	return r
}

func RtSelectionFetch(c echo.Context) error {
	jsbytes := reportcurrentselections(c)
	return c.String(http.StatusOK, string(jsbytes))
}

func RtTest(c echo.Context) error {
	a := len(AllAuthors)
	s := fmt.Sprintf("%d authors present", a)
	return c.String(http.StatusOK, s)
}
