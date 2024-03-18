//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"bytes"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"text/template"
)

// RtEmbHCSS - send "hipparchiastyles.css" after building it as per the configured font settings
func RtEmbHCSS(c echo.Context) error {
	const (
		ECSS = "emb/css/hgs.css"
	)

	if lnch.Config.CustomCSS {
		return CustomCSS(c)
	}

	// if you asked for a font on the command line, the next two lines will do something about that
	fsub := lnch.Config.Font
	sdf := "var(--systemdefaultfont), "

	// if the font is being served, then blank out "--systemdefaultfont" and get ready to map the font files into the CSS
	if slices.Contains(gen.StringMapKeysIntoSlice(vv.ServableFonts), lnch.Config.Font) {
		fsub = ""
		sdf = ""
	}

	j, e := efs.ReadFile(ECSS)
	if e != nil {
		msg.WARN(fmt.Sprintf("RtEmbHCSS() can't find %s", ECSS))
		return c.String(http.StatusNotFound, "")
	}

	subs := map[string]interface{}{
		"fontname":     fsub,
		"sdf":          sdf,
		"fontfaceinfo": cssfontfacedirectives(lnch.Config.Font),
	}

	tmpl, e := template.New("fp").Parse(string(j))
	msg.EC(e)

	var b bytes.Buffer
	err := tmpl.Execute(&b, subs)
	msg.EC(err)

	css := b.String()

	// if the font is not being served, then replace font names with explicit style directives
	if !slices.Contains(gen.StringMapKeysIntoSlice(vv.ServableFonts), lnch.Config.Font) {
		css = cssmanualfontstyling(css)
	}

	c.Response().Header().Add("Content-Type", "text/css")
	return c.String(http.StatusOK, css)
}

func CustomCSS(c echo.Context) error {
	const (
		FAIL1 = "could not open CSS file '%s%s'; using default instead"
		FAIL2 = "could not read CSS file '%s%s'; using default instead"
	)

	uh, _ := os.UserHomeDir()
	h := fmt.Sprintf(vv.CONFIGALTAPTH, uh)
	f := fmt.Sprintf("%s/%s", h, vv.CUSTOMCSSFILENAME)

	csf, ee := os.Open(f)
	if ee != nil {
		msg.CRIT(fmt.Sprintf(FAIL1, h, vv.CUSTOMCSSFILENAME))
		lnch.Config.CustomCSS = false
		return RtEmbHCSS(c)
	}

	b, err := io.ReadAll(csf)
	if err != nil {
		msg.CRIT(fmt.Sprintf(FAIL2, h, vv.CUSTOMCSSFILENAME))
		lnch.Config.CustomCSS = false
		return RtEmbHCSS(c)
	}

	c.Response().Header().Add("Content-Type", "text/css")
	return c.String(http.StatusOK, string(b))
}

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
		font-family: 'hipparchiasemicondensedstatic';
		src: url('/emb/{{.ShrtType}}/{{.SemiCondRegular}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchiasemicondenseditalicstatic';
		src: url('/emb/{{.ShrtType}}/{{.SemiCondItalic}}') format('{{.Type}}');
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
	if _, ok := vv.ServableFonts[f]; ok {
		fft, e := template.New("mt").Parse(FFS)
		msg.EC(e)
		var b bytes.Buffer
		err := fft.Execute(&b, vv.ServableFonts[f])
		msg.EC(err)
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
		"hipparchiasansstatic":                {"var(--systemdefaultfont), sans-serif", "normal", "normal", "normal"},
		"hipparchiamonostatic":                {"monospace", "normal", "normal", "normal"},
		"hipparchialightstatic":               {"var(--systemdefaultfont), sans-serif", "200", "normal", "normal"},
		"hipparchiaboldstatic":                {"var(--systemdefaultfont), sans-serif", "bold", "normal", "normal"},
		"hipparchiaobliquestatic":             {"var(--systemdefaultfont), sans-serif", "normal", "oblique", "normal"},
		"hipparchiabolditalicstatic":          {"var(--systemdefaultfont), sans-serif", "bold", "oblique", "normal"},
		"hipparchiasemicondensedstatic":       {"var(--systemdefaultfont), sans-serif", "normal", "normal", "condensed"},
		"hipparchiasemicondenseditalicstatic": {"var(--systemdefaultfont), sans-serif", "normal", "oblique", "condensed"},
		"hipparchiacondensedstatic":           {"var(--systemdefaultfont), sans-serif", "normal", "normal", "condensed"},
		"hipparchiacondensedboldstatic":       {"var(--systemdefaultfont), sans-serif", "bold", "normal", "condensed"},
		"hipparchiacondenseditalicstatic":     {"var(--systemdefaultfont), sans-serif", "normal", "oblique", "condensed"},
		"hipparchiasemiboldstatic":            {"var(--systemdefaultfont), sans-serif", "600", "normal", "normal"},
		"hipparchiathinstatic":                {"var(--systemdefaultfont), sans-serif", "100", "normal", "normal"},
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
