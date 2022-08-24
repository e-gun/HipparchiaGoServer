package main

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"html/template"
	"io"
	"net/http"
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

func StartEchoServer() {
	// https://echo.labstack.com/guide/
	// cf https://medium.com/cuddle-ai/building-microservice-using-golang-echo-framework-ff10ba06d508
	e := echo.New()
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("static/html/frontpage.html")),
	}
	e.Renderer = renderer

	// e.Use(middleware.Logger())
	// e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{Format: "method=${method}, uri=${uri}, status=${status}\n"}))
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
	// [b2] sample input: '/browse/locus/lt0550/001/3|100'
	// [b3] sample input: '/browse/perseus/lt0550/001/2:717'
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
	// [f2c] /get/json/workstructure
	// [f2d] /get/json/samplecitation
	// [f2e] /get/json/authorinfo
	// [f2f] /get/json/searchlistcontents
	// [f2e] /get/json/genrelistcontents
	// [f2f] /get/json/vectorranges
	// [f2g] /get/json/helpdata

	// [g] hinters
	// [h] lexical

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
	e.GET("/selection/make", RtSelectionMake)

	// [k2] "GET /selection/clear/auselections/0 HTTP/1.1"
	e.GET("/selection/clear", RtSelectionClear)

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
	// python sample: {"_fresh": "no", "agnexclusions": [], "agnselections": [], "alocexclusions": [], "alocselections": [], "analogyfinder": "no", "auexclusions": [], "auselections": ["gr7000"], "authorflagging": "yes", "authorssummary": "yes", "available": {"greek_dictionary": true, "greek_lemmata": true, "greek_morphology": true, "latin_dictionary": true, "latin_lemmata": true, "latin_morphology": true, "wordcounts_0": true}, "baggingmethod": "winnertakesall", "bracketangled": "yes", "bracketcurly": "yes", "bracketround": "no", "bracketsquare": "yes", "browsercontext": "24", "christiancorpus": "no", "collapseattic": "yes", "cosdistbylineorword": "no", "cosdistbysentence": "no", "debugdb": "no", "debughtml": "no", "debuglex": "no", "debugparse": "no", "earliestdate": "-850", "fontchoice": "Noto", "greekcorpus": "yes", "headwordindexing": "no", "incerta": "yes", "indexbyfrequency": "no", "indexskipsknownwords": "no", "inscriptioncorpus": "no", "latestdate": "1500", "latincorpus": "yes", "ldacomponents": 7, "ldaiterations": 12, "ldamaxfeatures": 2000, "ldamaxfreq": 35, "ldaminfreq": 5, "ldamustbelongerthan": 3, "linesofcontext": 4, "loggedin": "no", "maxresults": "200", "morphdialects": "no", "morphduals": "yes", "morphemptyrows": "yes", "morphfinite": "yes", "morphimper": "yes", "morphinfin": "yes", "morphpcpls": "yes", "morphtables": "yes", "nearestneighborsquery": "no", "nearornot": "near", "onehit": "no", "papyruscorpus": "no", "phrasesummary": "no", "principleparts": "yes", "proximity": "1", "psgexclusions": [], "psgselections": [], "quotesummary": "yes", "rawinputstyle": "no", "searchinsidemarkup": "no", "searchscope": "lines", "semanticvectorquery": "no", "sensesummary": "yes", "sentencesimilarity": "no", "showwordcounts": "yes", "simpletextoutput": "no", "sortorder": "shortname", "spuria": "yes", "suppresscolors": "no", "tensorflowgraph": "no", "topicmodel": "no", "trimvectoryby": "none", "userid": "Anonymous", "varia": "yes", "vcutlem": 50, "vcutloc": 33, "vcutneighb": 15, "vdim": 300, "vdsamp": 5, "viterat": 12, "vminpres": 10, "vnncap": 15, "vsentperdoc": 1, "vwindow": 10, "wkexclusions": [], "wkgnexclusions": [], "wkgnselections": [], "wkselections": [], "wlocexclusions": [], "wlocselections": [], "xmission": "Any", "zaplunates": "no", "zapvees": "no"}

	user := readUUIDCookie(c)
	if _, exists := sessions[user]; !exists {
		sessions[user] = makedefaultsession(user)
	}

	s := sessions[user]
	o, e := json.Marshal(s)
	checkerror(e)
	return c.String(http.StatusOK, string(o))
}

func RtResetSession(c echo.Context) error {
	// delete my session..
	delete(sessions, readUUIDCookie(c))
	return RtFrontpage(c)
}

func RtSelectionMake(c echo.Context) error {
	return c.String(http.StatusOK, "")
}

func RtSelectionClear(c echo.Context) error {
	return c.String(http.StatusOK, "")
}

func RtSelectionFetch(c echo.Context) error {
	return c.String(http.StatusOK, "")
}

func RtTest(c echo.Context) error {
	a := len(AllAuthors)
	s := fmt.Sprintf("%d authors present", a)
	return c.String(http.StatusOK, s)
}
