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

//
// HELPERS
//

// pathembedder - read and send file at path
func pathembedder(c echo.Context, d string) error {
	f := c.Param("file")
	j, e := efs.ReadFile(d + f)
	if e != nil {
		msg(fmt.Sprintf("can't find %s", d+f), MSGWARN)
		return c.String(http.StatusNotFound, "")
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

	if strings.HasSuffix(f, ".css") {
		// jquery-ui.css
		add = "text/css"
	}

	if strings.HasSuffix(f, ".ico") {
		add = "image/vnd.microsoft.icon"
	}

	if strings.HasSuffix(f, ".js") {
		add = "text/javascript"
	}

	if strings.HasSuffix(f, ".png") {
		add = "image/png"
	}

	if strings.HasSuffix(f, ".ttf") {
		add = "font/ttf"
	}

	if strings.HasSuffix(f, ".otf") {
		add = "font/opentype"
	}

	if strings.HasSuffix(f, ".woff2") {
		add = "font/woff2"
	}

	if strings.HasSuffix(f, ".pdf") {
		add = "application/pdf"
	}

	return add
}
