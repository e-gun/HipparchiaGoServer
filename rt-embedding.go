package main

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"text/template"
)

//go:embed emb
var efs embed.FS

func RtEmbJQuery(c echo.Context) error {
	d := "emb/jq/"
	return pathembedder(c, d)
}

func RtEmbExtraJS(c echo.Context) error {
	d := "emb/extrajs/"
	return pathembedder(c, d)
}

func RtEmbJS(c echo.Context) error {
	d := "emb/js/"
	return pathembedder(c, d)
}

func RtEmbHCSS(c echo.Context) error {
	f := "emb/hipparchiastyles.css"
	j, e := efs.ReadFile(f)
	if e != nil {
		msg(fmt.Sprintf("RtEmbHCSS() can't find %s", f), 1)
		return c.String(http.StatusNotFound, "")
	}

	subs := map[string]interface{}{
		"fontname": cfg.Font,
	}

	tmpl, e := template.New("fp").Parse(string(j))
	chke(e)

	var b bytes.Buffer
	err := tmpl.Execute(&b, subs)
	chke(err)

	c.Response().Header().Add("Content-Type", "text/css")
	return c.String(http.StatusOK, b.String())
}

func RtEmbJQueryImg(c echo.Context) error {
	d := "emb/jq/images/"
	return pathembedder(c, d)
}

func RtEmbTTF(c echo.Context) error {
	d := "emb/ttf/"
	return pathembedder(c, d)
}

func pathembedder(c echo.Context, d string) error {
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

func RtEbmFavicon(c echo.Context) error {
	f := "emb/images/hipparchia_favicon.ico"
	return fileembedder(c, f)
}

func RtEbmTouchIcon(c echo.Context) error {
	f := "emb/images/hipparchia_apple-touch-icon-precomposed.png"
	return fileembedder(c, f)
}

func fileembedder(c echo.Context, f string) error {
	j, e := efs.ReadFile(f)
	if e != nil {
		msg(fmt.Sprintf("can't find %s", f), 1)
		return c.String(http.StatusNotFound, "")
	}
	return c.String(http.StatusOK, string(j))
}

/*
HipparchiaGoServer/emb/ % tree
.
├── extrajs
│   ├── js.cookie.js
│   ├── jsd3.js
│   └── jsforldavis.js
├── frontpage.html
├── hipparchiastyles.css
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
│   ├── authentication.js
│   ├── autocomplete.js
│   ├── browser.js
│   ├── coreinterfaceclicks_go.js
│   ├── documentready_go.js
│   ├── nonvectorspinners.js
│   ├── progressindicator_go.js
│   ├── radioclicks.js
│   ├── uielementlists_go.js
│   └── vectorclicks.js
└── ttf
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
    ├── NotoSansMono-Regular.ttf
    └── license_for_noto_fonts.txt
*/
