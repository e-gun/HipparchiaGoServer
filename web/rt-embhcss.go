//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"bytes"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
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

type FontSwap struct {
	familiy string
	weight  string
	style   string
	stretch string
}

var (
	// used by cssmanualfontstyling
	cssswaps = map[string]FontSwap{
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
		Msg.WARN(fmt.Sprintf("RtEmbHCSS() can't find %s", ECSS))
		return c.String(http.StatusNotFound, "")
	}

	subs := map[string]interface{}{
		"fontname":     fsub,
		"sdf":          sdf,
		"fontfaceinfo": cssfontfacedirectives(lnch.Config.Font),
	}

	tmpl, e := template.New("fp").Parse(string(j))
	Msg.EC(e)

	var b bytes.Buffer
	err := tmpl.Execute(&b, subs)
	Msg.EC(err)

	css := b.String()

	// if the font is not being served, then replace font names with explicit style directives
	if !slices.Contains(gen.StringMapKeysIntoSlice(vv.ServableFonts), lnch.Config.Font) {
		css = cssmanualfontstyling(css)
	}

	// if the font is being served, but it is relatively hollow when it comes to styles, patch things
	if slices.Contains(gen.StringMapKeysIntoSlice(vv.ServableFonts), lnch.Config.Font) {
		f := vv.ServableFonts[lnch.Config.Font]
		if len(f.NeedsManualStyle) != 0 {
			css = fleshoutcss(f, css)
		}
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
		Msg.CRIT(fmt.Sprintf(FAIL1, h, vv.CUSTOMCSSFILENAME))
		lnch.Config.CustomCSS = false
		return RtEmbHCSS(c)
	}

	b, err := io.ReadAll(csf)
	if err != nil {
		Msg.CRIT(fmt.Sprintf(FAIL2, h, vv.CUSTOMCSSFILENAME))
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
		src: url('/emb/{{.ShrtType}}/{{.SubFolder}}/{{.Regular}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchiamonostatic';
		src: url('/emb/{{.ShrtType}}/{{.SubFolder}}/{{.Mono}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchialightstatic';
		src: url('/emb/{{.ShrtType}}/{{.SubFolder}}/{{.Light}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchiaboldstatic';
		src: url('/emb/{{.ShrtType}}/{{.SubFolder}}/{{.Bold}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchiaobliquestatic';
		src: url('/emb/{{.ShrtType}}/{{.SubFolder}}/{{.Italic}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchiabolditalicstatic';
		src: url('/emb/{{.ShrtType}}/{{.SubFolder}}/{{.BoldItalic}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchiasemicondensedstatic';
		src: url('/emb/{{.ShrtType}}/{{.SubFolder}}/{{.SemiCondRegular}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchiasemicondenseditalicstatic';
		src: url('/emb/{{.ShrtType}}/{{.SubFolder}}/{{.SemiCondItalic}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchiacondensedstatic';
		src: url('/emb/{{.ShrtType}}/{{.SubFolder}}/{{.CondensedRegular}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchiacondensedboldstatic';
		src: url('/emb/{{.ShrtType}}/{{.SubFolder}}/{{.CondensedBold}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchiacondenseditalicstatic';
		src: url('/emb/{{.ShrtType}}/{{.SubFolder}}/{{.CondensedItalic}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchiasemiboldstatic';
		src: url('/emb/{{.ShrtType}}/{{.SubFolder}}/{{.SemiBold}}') format('{{.Type}}');
		}

	@font-face {
		font-family: 'hipparchiathinstatic';
		src: url('/emb/{{.ShrtType}}/{{.SubFolder}}/{{.Thin}}') format('{{.Type}}');
		}`
	)

	css := ""
	if _, ok := vv.ServableFonts[f]; ok {
		fft, e := template.New("mt").Parse(FFS)
		Msg.EC(e)
		var b bytes.Buffer
		err := fft.Execute(&b, vv.ServableFonts[f])
		Msg.EC(err)
		css = b.String()
	}

	// note that incomplete fonts like Brill will have incomplete @font-face information
	// this can and will be checked

	return css
}

// cssmanualfontstyling - swap out: "font-family: 'hipparchiabolditalicstatic', sans-serif;" for explicit style directives
func cssmanualfontstyling(css string) string {

	// swap out: "font-family: 'hipparchiabolditalicstatic', sans-serif;" for explicit style directives
	outtmpl := "font-family: '%s', sans-serif;"
	intempl := "font-family: %s;\n\tfont-weight: %s;\n\tfont-style: %s;\n\tfont-stretch: %s;"
	for n, fs := range cssswaps {
		i := fmt.Sprintf(intempl, fs.familiy, fs.weight, fs.style, fs.stretch)
		o := fmt.Sprintf(outtmpl, n)
		css = strings.ReplaceAll(css, o, i)
	}

	// the above will have missed hipparchiamonostatic
	fs := cssswaps["hipparchiamonostatic"]
	i := fmt.Sprintf(intempl, fs.familiy, fs.weight, fs.style, fs.stretch)
	o := fmt.Sprintf("font-family: '%s', monospace;", "hipparchiamonostatic")
	css = strings.ReplaceAll(css, o, i)

	return css
}

func fleshoutcss(f str.FontTempl, css string) string {
	// blank looks like:
	//                font-family: 'hipparchialightstatic';
	//                src: url('/emb/ttf/') format('truetype');

	outtmpl := "font-family: '%s', sans-serif;"
	intempl := "font-family: %s;\n\tfont-weight: %s;\n\tfont-style: %s;\n\tfont-stretch: %s;"

	for _, n := range f.NeedsManualStyle {
		fs := cssswaps[n]
		i := fmt.Sprintf(intempl, fs.familiy, fs.weight, fs.style, fs.stretch)
		o := fmt.Sprintf(outtmpl, n)
		css = strings.ReplaceAll(css, o, i)
	}

	bad := "font-family: var(--systemdefaultfont), sans-serif;"
	good := "font-family: 'hipparchiasansstatic';"
	css = strings.ReplaceAll(css, bad, good)
	return css
}
