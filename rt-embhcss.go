package main

import (
	"bytes"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"text/template"
)

//const (
//	TEST = `
//dictionaryentry{
//	color: var(--rustedorange);
//	font-family: 'hipparchiasemiboldstatic', sans-serif;
//}
//
//dictionaryentry:hover{
//	border-bottom: 1px dotted black;
//	border-top: 1px dotted black;
//	color: var(--dkbabyblue);
//	font-family: 'hipparchiasemiboldstatic', sans-serif;
//	text-shadow: 1px 1px 3px var(--transparentgrey);
//}
//`
//)

//func main() {
//	fmt.Println(TEST)
//	fmt.Println(cssmanualfontstyling(TEST))
//}

// RtEmbHCSS - send "hipparchiastyles.css" after building it as per the configured font settings
func RtEmbHCSS(c echo.Context) error {
	const (
		ECSS = "emb/css/hgs.css"
	)

	// if you asked for a font on the command line, the next two lines will do something about that
	fsub := Config.Font
	sdf := "var(--systemdefaultfont), "

	// if the font is local, then blank out "--systemdefaultfont" and get ready to map the font files into the CSS
	if IsInSlice(Config.Font, StringMapKeysIntoSlice(ServableFonts)) {
		fsub = ""
		sdf = ""
	}

	j, e := efs.ReadFile(ECSS)
	if e != nil {
		msg(fmt.Sprintf("RtEmbHCSS() can't find %s", ECSS), MSGWARN)
		return c.String(http.StatusNotFound, "")
	}

	subs := map[string]interface{}{
		"fontname":     fsub,
		"sdf":          sdf,
		"fontfaceinfo": cssfontfacedirectives(Config.Font),
	}

	tmpl, e := template.New("fp").Parse(string(j))
	chke(e)

	var b bytes.Buffer
	err := tmpl.Execute(&b, subs)
	chke(err)

	css := b.String()
	if !IsInSlice(Config.Font, StringMapKeysIntoSlice(ServableFonts)) {
		css = cssmanualfontstyling(css)
	}

	c.Response().Header().Add("Content-Type", "text/css")
	return c.String(http.StatusOK, css)
}

//
// FONTS
//

type FontTempl struct {
	Type             string
	ShrtType         string
	Bold             string
	BoldItalic       string
	CondensedBold    string
	CondensedItalic  string
	CondensedRegular string
	Italic           string
	Light            string
	Mono             string
	Regular          string
	SemiBold         string
	Thin             string
}

// the fonts we know how to serve: only one installed ATM
// NB: Fira, Inter, and Ubuntu have all been toyed with: none are really as good as NotoSans...

var (
	NotoFont = FontTempl{
		Type:             "truetype",
		ShrtType:         "ttf",
		Bold:             "NotoSans-Bold.ttf",
		BoldItalic:       "NotoSans-BoldItalic.ttf",
		CondensedBold:    "NotoSans-CondensedSemiBold.ttf",
		CondensedItalic:  "NotoSans-CondensedItalic.ttf",
		CondensedRegular: "NotoSans-CondensedMedium.ttf",
		Italic:           "NotoSans-Italic.ttf",
		Light:            "NotoSans-ExtraLight.ttf",
		Mono:             "NotoSansMono-Regular.ttf",
		Regular:          "NotoSans-Regular.ttf",
		SemiBold:         "NotoSans-SemiBold.ttf",
		Thin:             "NotoSans-Thin.ttf",
	}
)

// cssfontfacedirectives - swap the served font file info into the CSS
func cssfontfacedirectives(f string) string {
	const (
		FFS = `
	@font-face {
		font-family: 'hipparchiasansstatic';
		src: url('/emb/{{.ShrtType}}/{{.Regular}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchiamonostatic';
		src: url('/emb/{{.ShrtType}}/{{.Mono}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchialightstatic';
		src: url('/emb/{{.ShrtType}}/{{.Light}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchiaboldstatic';
		src: url('/emb/{{.ShrtType}}/{{.Bold}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchiaobliquestatic';
		src: url('/emb/{{.ShrtType}}/{{.Italic}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchiabolditalicstatic';
		src: url('/emb/{{.ShrtType}}/{{.BoldItalic}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchiacondensedstatic';
		src: url('/emb/{{.ShrtType}}/{{.CondensedRegular}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchiacondensedboldstatic';
		src: url('/emb/{{.ShrtType}}/{{.CondensedBold}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchiacondenseditalicstatic';
		src: url('/emb/{{.ShrtType}}/{{.CondensedItalic}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchiasemiboldstatic';
		src: url('/emb/{{.ShrtType}}/{{.SemiBold}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchiathinstatic';
		src: url('/emb/{{.ShrtType}}/{{.Thin}}') format('{{.Type}}');
		}`
	)

	css := ""
	if _, ok := ServableFonts[f]; ok {
		fft, e := template.New("mt").Parse(FFS)
		chke(e)
		var b bytes.Buffer
		err := fft.Execute(&b, ServableFonts[f])
		chke(err)
		css = b.String()
	}

	return css
}

// cssmanualfontstyling - swap out: "font-family: 'hipparchiabolditalicstatic', sans-serif;" for explicit style directives
func cssmanualfontstyling(css string) string {
	type FontSwap struct {
		familiy string
		weight  string
		style   string
		stretch string
	}

	swaps := map[string]FontSwap{
		"hipparchiasansstatic":            FontSwap{"var(--systemdefaultfont), sans-serif", "normal", "normal", "normal"},
		"hipparchiamonostatic":            FontSwap{"monospace", "normal", "normal", "normal"},
		"hipparchialightstatic":           FontSwap{"var(--systemdefaultfont), sans-serif", "200", "normal", "normal"},
		"hipparchiaboldstatic":            FontSwap{"var(--systemdefaultfont), sans-serif", "bold", "normal", "normal"},
		"hipparchiaobliquestatic":         FontSwap{"var(--systemdefaultfont), sans-serif", "normal", "oblique", "normal"},
		"hipparchiabolditalicstatic":      FontSwap{"var(--systemdefaultfont), sans-serif", "bold", "oblique", "normal"},
		"hipparchiacondensedstatic":       FontSwap{"var(--systemdefaultfont), sans-serif", "normal", "normal", "condensed"},
		"hipparchiacondensedboldstatic":   FontSwap{"var(--systemdefaultfont), sans-serif", "bold", "normal", "condensed"},
		"hipparchiacondenseditalicstatic": FontSwap{"var(--systemdefaultfont), sans-serif", "normal", "oblique", "condensed"},
		"hipparchiasemiboldstatic":        FontSwap{"var(--systemdefaultfont), sans-serif", "600", "normal", "normal"},
		"hipparchiathinstatic":            FontSwap{"var(--systemdefaultfont), sans-serif", "100", "normal", "normal"},
	}

	// swap out: "font-family: 'hipparchiabolditalicstatic', sans-serif;" for explicit style directives
	outtmpl := "font-family: '%s', sans-serif;"
	intempl := "font-family: %s;\n\tfont-weight: %s;\n\tfont-style: %s;\n\tfont-stretch: %s;"
	for n, fs := range swaps {
		i := fmt.Sprintf(intempl, fs.familiy, fs.weight, fs.style, fs.stretch)
		o := fmt.Sprintf(outtmpl, n)
		css = strings.ReplaceAll(css, o, i)
	}

	// the above will have missed hipparchiamonostatic
	fs := swaps["hipparchiamonostatic"]
	i := fmt.Sprintf(intempl, fs.familiy, fs.weight, fs.style, fs.stretch)
	o := fmt.Sprintf("font-family: '%s', monospace;", "hipparchiamonostatic")
	css = strings.ReplaceAll(css, o, i)

	return css
}
