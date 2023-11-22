//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"embed"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

//go:embed emb
var efs embed.FS

//
// ROUTES
//

const (
	EEC  = "emb/echarts/"
	EJQ  = "emb/jq/"
	EJQI = "emb/jq/images/"
	EJS  = "emb/js/"
	ETT  = "emb/ttf/"
	EOT  = "emb/otf/"
	EWF  = "emb/wof/"
	EPD  = "emb/pdf/"
	EICO = "emb/images/hipparchia_favicon.ico"
	EITC = "emb/images/hipparchia_apple-touch-icon-precomposed.png"
)

func RtEmbJQuery(c echo.Context) error {
	return pathembedder(c, EJQ)
}

func RtEmbJQueryImg(c echo.Context) error {
	return pathembedder(c, EJQI)
}

func RtEmbJS(c echo.Context) error {
	return pathembedder(c, EJS)
}

func RtEmbTTF(c echo.Context) error {
	return pathembedder(c, ETT)
}

func RtEmbOTF(c echo.Context) error {
	return pathembedder(c, EOT)
}

func RtEmbWOF(c echo.Context) error {
	return pathembedder(c, EWF)
}

func RtEbmFavicon(c echo.Context) error {
	return fileembedder(c, EICO)
}

func RtEbmTouchIcon(c echo.Context) error {
	return fileembedder(c, EITC)
}

// RtEmbPDFHelp - send one of the embedded PDF files
func RtEmbPDFHelp(c echo.Context) error {
	return pathembedder(c, EPD)
}

// RtEmbEcharts - send one of the embedded graphing JS files
func RtEmbEcharts(c echo.Context) error {
	return pathembedder(c, EEC)
}

//
// HELPERS
//

// pathembedder - read and send file at path
func pathembedder(c echo.Context, d string) error {
	const (
		FNF      = "pathembedder() can't find '%s'"
		OOPSFILE = "emb/pdf/oops.pdf"
	)

	f := c.Param("file")
	j, e := efs.ReadFile(d + f)
	if e != nil {
		msg(fmt.Sprintf(FNF, d+f), MSGWARN)
		if !strings.HasSuffix(f, ".pdf") {
			// a normal 404 error
			return c.String(http.StatusNotFound, "")
		} else {
			// the documentation was not build in...
			// omit checking the error if OOPSFILE itself is not found: an empty string will be sent; no harm in that
			k, _ := efs.ReadFile(OOPSFILE)
			c.Response().Header().Add("Content-Type", "application/pdf")
			return c.String(http.StatusOK, string(k))
		}
	}

	add := addresponsehead(f)
	if len(add) != 0 {
		c.Response().Header().Add("Content-Type", add)
	}

	return c.String(http.StatusOK, string(j))
}

// fileembedder - read and send file
func fileembedder(c echo.Context, f string) error {
	j, e := efs.ReadFile(f)
	if e != nil {
		msg(fmt.Sprintf("can't find %s", f), MSGWARN)
		return c.String(http.StatusNotFound, "")
	}

	add := addresponsehead(f)
	if len(add) != 0 {
		c.Response().Header().Add("Content-Type", add)
	}

	return c.String(http.StatusOK, string(j))
}

// addresponsehead - set the response header for various file types
func addresponsehead(f string) string {
	// c.Response().Header().Add("Content-Type", "text/css")
	add := ""

	kvp := map[string]string{
		".css":   "text/css",
		".ico":   "image/vnd.microsoft.icon",
		".js":    "text/javascript",
		".otf":   "font/opentype",
		".png":   "image/png",
		".ttf":   "font/ttf",
		".woff2": "font/woff2",
		".pdf":   "application/pdf",
	}

	for k, v := range kvp {
		if strings.HasSuffix(f, k) {
			add = v
		}
	}

	return add
}
