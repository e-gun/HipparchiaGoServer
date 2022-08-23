package main

import (
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

func main() {
	// https://echo.labstack.com/guide/
	// cf https://medium.com/cuddle-ai/building-microservice-using-golang-echo-framework-ff10ba06d508
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.File("/favicon.ico", "static/images/hipparchia_favicon.ico")
	e.Static("/static", "static")

	// hipparchia routes
	// [a] authentication
	// [b] browser
	// [c] css
	// [d] debugging
	// [e] frontpage

	e.File("/", "static/html/frontpage.html")

	// [f] getters
	// [g] hinters
	// [h] lexical
	// [i] resets
	// [j] searching
	// [k] selection

	// [k1] "GET /selection/make/_?auth=gr7000 HTTP/1.1"
	e.GET("/selection/make", RtSelectionMake)

	// [k2] "GET /selection/clear/auselections/0 HTTP/1.1"
	e.GET("/selection/clear", RtSelectionClear)

	// [k3] "GET /selection/fetch HTTP/1.1"
	e.GET("/selection/fetch", RtSelectionFetch)

	// [l] text and index
	// [m] vectors

	e.Logger.Fatal(e.Start(":8000"))
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
