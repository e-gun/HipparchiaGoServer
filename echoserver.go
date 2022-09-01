package main

import (
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
	e.GET("/get/json/helpdata", RtGetJSHelpdata)

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

	//
	// [l] setoption: http://localhost:8000/setoption/greekcorpus/yes
	//

	e.GET("/setoption/:opt", RtSetOption)

	//
	// [m] text and index
	//

	//
	// [n] vectors [unneeded/unimplemented ATM]
	//

	// [z] testing
	e.GET("/t", RtTest)

	e.Logger.Fatal(e.Start(":8000"))
}

func RtAuthChkuser(c echo.Context) error {
	return c.String(http.StatusOK, "")
}

func RtResetSession(c echo.Context) error {
	// delete my session
	delete(sessions, readUUIDCookie(c))

	// then reset it
	readUUIDCookie(c)
	return RtFrontpage(c)
}

func RtSetOption(c echo.Context) error {
	optandval := c.Param("opt")
	parsed := strings.Split(optandval, "/")

	if len(parsed) != 2 {
		msg(fmt.Sprintf("RtSetOption() was given bad input: %s", optandval), 1)
		return c.String(http.StatusOK, "")
	}

	opt := parsed[0]
	val := parsed[1]

	ynoptionlist := []string{"greekcorpus", "latincorpus", "papyruscorpus", "inscriptioncorpus", "christiancorpus",
		"rawinputstyle", "onehit", "headwordindexing", "indexbyfrequency", "spuria", "incerta", "varia"}

	s := sessions[readUUIDCookie(c)]

	if contains(ynoptionlist, opt) {
		valid := []string{"yes", "no"}
		if contains(valid, val) {
			var b bool
			if val == "yes" {
				b = true
			} else {
				b = false
			}
			switch opt {
			case "greekcorpus":
				s.ActiveCorp["gr"] = b
			case "latincorpus":
				s.ActiveCorp["lt"] = b
			case "papyruscorpus":
				s.ActiveCorp["dp"] = b
			case "inscriptioncorpus":
				s.ActiveCorp["in"] = b
			case "christiancorpus":
				s.ActiveCorp["ch"] = b
			case "rawinputstyle":
				s.RawInput = b
			case "onehit":
				s.OneHit = b
			case "indexbyfrequency":
				s.FrqIdx = b
			case "headwordindexing":
				s.HeadwordIdx = b
			case "spuria":
				s.SpuriaOK = b
			case "incerta":
				s.IncertaOK = b
			case "varia":
				s.VariaOK = b
			default:
				msg("RtSetOption() hit an impossible case", 1)
			}
		}
	}

	valoptionlist := []string{"nearornot", "searchscope", "sortorder"}
	if contains(valoptionlist, opt) {
		switch opt {
		case "nearornot":
			valid := []string{"near", "notnear"}
			if contains(valid, val) {
				s.NearOrNot = val
			}
		case "searchscope":
			valid := []string{"lines", "words"}
			if contains(valid, val) {
				s.SearchScope = val
			}
		case "sortorder":
			// unhandled are "location" & "provenance": see goroutinesearcher.go
			valid := []string{"shortname", "converted_date", "location", "provenance", "universalid"}
			if contains(valid, val) {
				s.SortHitsBy = val
			}
		default:
			msg("RtSetOption() hit an impossible case", 1)
		}
	}

	spinoptionlist := []string{"maxresults", "linesofcontext", "browsercontext"}
	if contains(spinoptionlist, opt) {
		intval, e := strconv.Atoi(val)
		if e == nil {
			switch opt {
			case "maxresults":
				if intval > MAXHITLIMIT {
					s.HitLimit = int64(intval)
				} else {
					s.HitLimit = MAXHITLIMIT
				}
			case "linesofcontext":
				if intval > MAXLINESHITCONTEXT {
					s.HitContext = int(intval)
				} else {
					s.HitContext = MAXLINESHITCONTEXT
				}
			case "browsercontext":
				if intval > MAXBROWSERCONTEXT {
					s.UI.BrowseCtx = int64(intval)
				} else {
					s.UI.BrowseCtx = MAXBROWSERCONTEXT
				}
			default:
				msg("RtSetOption() hit an impossible case", 1)
			}
		}
	}

	st := fmt.Sprintf("set '%s' to '%s'", parsed[0], parsed[1])
	delete(sessions, readUUIDCookie(c))
	sessions[readUUIDCookie(c)] = s

	msg(st, 1)

	return c.String(http.StatusOK, "")
}

func RtTest(c echo.Context) error {
	a := len(AllAuthors)
	s := fmt.Sprintf("%d authors present", a)
	return c.String(http.StatusOK, s)
}
