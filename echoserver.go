package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"html/template"
	"io"
	"net/http"
)

// TemplateRenderer is a custom html/template renderer for Echo framework
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
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{Format: "method=${method}, uri=${uri}, status=${status}\n"}))
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

	e.File("/", "static/html/frontpage.html")

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
	// [i] resets
	// [j] searching

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
	// see hipparchiajs/coreinterfaceclicks.js
	return c.String(http.StatusOK, "")
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
