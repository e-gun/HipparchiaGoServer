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

	if strings.Contains(f, ".css") {
		// jquery-ui.css
		add = "text/css"
	}

	if strings.Contains(f, ".ico") {
		add = "image/vnd.microsoft.icon"
	}

	if strings.Contains(f, ".js") {
		add = "text/javascript"
	}

	if strings.Contains(f, ".png") {
		add = "image/png"
	}

	if strings.Contains(f, ".ttf") {
		add = "font/ttf"
	}

	if strings.Contains(f, ".otf") {
		add = "font/opentype"
	}

	if strings.Contains(f, ".woff2") {
		add = "font/woff2"
	}

	return add
}

/*
HipparchiaGoServer/emb/ % tree
.
├── css
│   └── hgs.css
├── frontpage.html
├── h
│   ├── helpbasicsyntax.html
│   ├── helpbrowsing.html
│   ├── helpdictionaries.html
│   ├── helpextending.html
│   ├── helpinterface.html
│   ├── helplemmata.html
│   ├── helpoddities.html
│   ├── helpregex.html
│   ├── helpsearchlists.html
│   └── includedmaterials.html
├── images
│   ├── hipparchia_apple-touch-icon-precomposed.png
│   ├── hipparchia_favicon.ico
│   └── hipparchia_logo.png
├── jq
│   ├── images
│   │   ├── ui-icons_444444_256x240.png
│   │   ├── ui-icons_555555_256x240.png
│   │   ├── ui-icons_777620_256x240.png
│   │   ├── ui-icons_777777_256x240.png
│   │   ├── ui-icons_cc0000_256x240.png
│   │   └── ui-icons_ffffff_256x240.png
│   ├── jquery-ui.css
│   ├── jquery-ui.js
│   ├── jquery-ui.min.css
│   ├── jquery-ui.min.js
│   ├── jquery-ui.structure.css
│   ├── jquery-ui.structure.min.css
│   ├── jquery-ui.theme.css
│   ├── jquery-ui.theme.min.css
│   ├── jquery.min.js
│   └── license_for_jquery.txt
├── js
│   ├── autocomplete.js
│   ├── browser.js
│   ├── coreinterfaceclicks_go.js
│   ├── documentready_go.js
│   ├── radioclicks_go.js
│   ├── uielementlists_go.js
│   └── vectorclicks.js
├── otf
│   ├── MaterialIconsSharp-Regular.otf
│   └── license_for_materialicons.txt
├── ttf
│   ├── NotoSans-Bold.ttf
│   ├── NotoSans-BoldItalic.ttf
│   ├── NotoSans-CondensedItalic.ttf
│   ├── NotoSans-CondensedMedium.ttf
│   ├── NotoSans-CondensedSemiBold.ttf
│   ├── NotoSans-ExtraLight.ttf
│   ├── NotoSans-Italic.ttf
│   ├── NotoSans-Regular.ttf
│   ├── NotoSans-SemiBold.ttf
│   ├── NotoSans-Thin.ttf
│   ├── NotoSansMono-Regular.ttf
│   └── license_for_noto_fonts.txt
└── wof
    ├── iosevka-regular.woff2
    └── iosevka_license.md

10 directories, 54 files
*/
