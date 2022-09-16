package main

import (
	"embed"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

// https://www.abilityrush.com/how-to-embed-any-file-in-golang-binary/

//go:embed emb
var efs embed.FS

func RtEmbJQuery(c echo.Context) error {
	d := "emb/jq/"
	return generecembstring(c, d)
}

func RtEmbExtraJS(c echo.Context) error {
	d := "emb/extrajs/"
	return generecembstring(c, d)
}

func RtEmbJS(c echo.Context) error {
	d := "emb/js/"
	return generecembstring(c, d)
}

func RtEmbHCSS(c echo.Context) error {
	d := "emb/"
	return generecembstring(c, d)
}

func RtEmbJQueryImg(c echo.Context) error {
	d := "emb/jq/images/"
	return generecembstring(c, d)
}

func generecembstring(c echo.Context, d string) error {
	f := c.Param("file")
	j, e := efs.ReadFile(d + f)
	if e != nil {
		msg(fmt.Sprintf("can't find %s", d+f), 1)
		return c.String(http.StatusNotFound, "")
	}

	if strings.Contains(f, ".css") {
		// jquery-ui.css
		c.Response().Header().Add("Content-Type", "text/css")
	}

	return c.String(http.StatusOK, string(j))
}

/*
HipparchiaGoServer/static/ % tree
.
├── css
│   └── hipparchiastyles.css
├── hipparchiajs
│   ├── authentication.js
│   ├── autocomplete.js
│   ├── browser.js
│   ├── cookieclicks.js
│   ├── coreinterfaceclicks_go.js
│   ├── documentready_go.js
│   ├── indexandtextmaker_go.js
│   ├── nonvectorspinners.js
│   ├── progressindicator_go.js
│   ├── radioclicks.js
│   ├── uielementlists_go.js
│   └── vectorclicks.js
├── html
│   └── frontpage.html
├── images
│   ├── hipparchia_apple-touch-icon-precomposed.png
│   ├── hipparchia_favicon.ico
│   ├── hipparchia_logo.png
│   ├── ui-icons_444444_256x240.png
│   ├── ui-icons_555555_256x240.png
│   ├── ui-icons_777620_256x240.png
│   ├── ui-icons_777777_256x240.png
│   ├── ui-icons_cc0000_256x240.png
│   └── ui-icons_ffffff_256x240.png
├── jquery-ui.css
├── jquery-ui.js
├── jquery-ui.min.css
├── jquery-ui.min.js
├── jquery-ui.structure.css
├── jquery-ui.structure.min.css
├── jquery-ui.theme.css
├── jquery-ui.theme.min.css
├── jquery.min.js
├── js.cookie.js
├── jsd3.js
├── jsforldavis.js
└── ttf
    ├── 0_served_fonts_go_here
    ├── NotoSans-Bold.ttf
    ├── NotoSans-BoldItalic.ttf
    ├── NotoSans-ExtraLight.ttf
    ├── NotoSans-Italic.ttf
    ├── NotoSans-Regular.ttf
    ├── NotoSans-SemiBold.ttf
    ├── NotoSans-Thin.ttf
    ├── NotoSansDisplay_ExtraCondensed-Italic.ttf
    ├── NotoSansDisplay_ExtraCondensed-Regular.ttf
    ├── NotoSansDisplay_ExtraCondensed-SemiBold.ttf
    └── NotoSansMono-Regular.ttf
*/
