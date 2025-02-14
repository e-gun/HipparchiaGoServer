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
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/labstack/echo/v4"
	"io"
	"io/fs"
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

const (
	LIGHTCOLORS = `
	--main-body-color: hsla(0, 0%, 98%, 1);
	--main-font-color: hsla(0, 0%, 6%, 1);
	--input-border-color: hsla(0, 0%, 92%, 1);
	
	--buttoncolor: hsla(0, 0%, 93%, 1);
	--button-hover: hsla(0, 0%, 90%, 1);
	--fieldset-background: hsla(0, 0%, 98%, 1);
	--focus-shadow: rgba(0, 0, 0, 0.5);
	--icons-color: rgba(0, 0, 0, 0.54);
	
	--black: hsla(0, 0%, 0%, 1);
	--blue: hsla(240, 100%, 27%, 1);
	--brown: hsla(22, 22%, 26%, 1);
	--brtblue: hsla(236, 44%, 40%, 1);
	--copper: hsla(11, 53%, 30%);
	--deepblue: hsla(233, 77%, 26%, 1);
	--dkbabyblue: hsla(237, 43%, 57%, 1);
	--dkgreen: hsla(120, 80%, 20%, 1);
	--dkgrey: hsla(0, 0%, 33%, 1);
	--dkteal: hsla(203, 22%, 26%, 1);
	--huedgrey: hsl(240, 10%, 61%);
	--invisible: hsla(0, 100%, 100%, 0);
	--lessoffwhite: hsla(0, 0%, 98%, 1);
	--ltbabyblue: hsla(200, 33%, 95%, 1);
	--ltgrey: hsla(0, 0%, 90%, 1);
	--ltgrey2: hsl(206, 7%, 81%);
	--midgrey: hsla(0, 0%, 67%, 1);
	--offwhite: hsla(0, 0%, 99%, 1);
	--orange: hsla(47, 100%, 30%, 1);
	--pink: hsla(0, 33%, 96%, 1);
	--pinker: hsl(0, 73%, 80%);
	--plum: hsla(291, 15%, 38%);
	--pukegreen: hsl(71, 95%, 22%);
	--red: hsla(346, 77%, 26%, 1);
	--rustedorange: hsla(23, 37%, 39%, 1);
	--sicklyyellow: hsl(45, 16%, 53%);
	--skyblue: hsla(205, 92%, 37%, 1);
	--teal: hsla(196, 14%, 39%, 1);
	--transparentgrey: hsla(0, 0%, 67%, .8);
	--vdkteal: hsla(196, 27%, 20%, 1);
	--vdkgrey: hsla(0, 0%, 20%, 1);
	--vltgrey: hsla(0, 0%, 96%, 1);
	--white: hsla(0, 0%, 100%, 1);`

	DARKCOLORS = `
	--main-body-color: hsla(0, 0%, 10%, 1);
	--main-font-color: hsla(0, 0%, 95%, 1);
	--input-border-color: hsla(0, 0%, 20%, 1);
	
	--buttoncolor: hsla(0, 0%, 20%, 1);
	--button-hover: hsla(0, 0%, 10%, 1);
	
	--fieldset-background: hsla(0, 0%, 2%, 1);
	--focus-shadow: rgba(0, 0, 0, 0.5);
	--icons-color: rgba(255, 255, 255, 0.54);
	
	--black: hsla(0, 0%, 95%, 1);
	--blue: hsl(167, 35%, 77%);  /* pea green... */
	--brown: hsla(22, 22%, 74%, 1);
	--brtblue: hsla(236, 44%, 85%, 1);
	--copper: hsla(11, 53%, 83%);
	--deepblue: hsla(64, 88%, 84%, 1);  /* yellow... */
	--dkbabyblue:  hsla(64, 52%, 84%, 1);  /* yellow... */
	--dkgreen: hsla(120, 80%, 80%, 1);
	--dkgrey: hsla(0, 0%, 82%, 1);
	--dkteal: hsla(203, 22%, 82%, 1);
	--huedgrey: hsl(113, 35%, 79%);
	--invisible: hsla(0, 100%, 0%, 0);
	--lessoffwhite: hsla(0, 0%, 2%, 1);
	--ltbabyblue: hsla(200, 33%, 15%, 1);
	--ltgrey: hsla(0, 0%, 20%, 1);
	--ltgrey2: hsl(206, 7%, 11%);
	--midgrey: hsla(0, 0%, 53%, 1);
	--offwhite: hsla(0, 0%, 1%, 1);
	--orange: hsla(47, 100%, 75%, 1);
	--pink: hsla(0, 33%, 15%, 1);
	--pinker: hsl(0, 73%, 20%);
	--plum: hsla(291, 15%, 72%);
	--pukegreen: hsl(71, 95%, 78%);
	--red: hsl(258, 35%, 83%); /* lt purple... */
	--dkgreen: hsla(120, 80%, 80%, 1);
	--rustedorange: hsla(23, 37%, 71%, 1);
	--sicklyyellow: hsl(45, 16%, 67%);
	--skyblue: hsla(205, 92%, 73%, 1);
	--teal: hsla(196, 14%, 71%, 1);
	--transparentgrey: hsla(0, 0%, 23%, .8);
	--vdkteal: hsla(196, 27%, 80%, 1);
	--vdkgrey: hsla(0, 0%, 80%, 1);
	--vltgrey: hsla(0, 0%, 4%, 1);
	--white: hsla(0, 0%, 0%, 1);`
)

// RtEmbHCSS - send "hipparchiastyles.css" after building it as per the configured font settings
func RtEmbHCSS(c echo.Context) error {
	const (
		ECSS = "emb/css/hgs.css"
	)

	if lnch.Config.CustomCSS {
		return CustomCSS(c)
	}

	user := vlt.ReadUUIDCookie(c)
	s := vlt.AllSessions.GetSess(user)

	// if you asked for a font on the command line, the next two lines will do something about that
	fsub := lnch.Config.Font
	sdf := "var(--systemdefaultfont), "

	// if the font is being served, then blank out "--systemdefaultfont" and get ready to map the font files into the CSS
	// note that that option will nullify the UI font selection
	if slices.Contains(gen.StringMapKeysIntoSlice(vv.ServableFonts), lnch.Config.Font) {
		fsub = ""
		sdf = ""
	}

	j, e := efs.ReadFile(ECSS)
	if e != nil {
		Msg.WARN(fmt.Sprintf("RtEmbHCSS() can't find %s", ECSS))
		return c.String(http.StatusNotFound, "")
	}

	colors := LIGHTCOLORS
	if s.DarkMode {
		colors = DARKCOLORS
	}
	subs := map[string]interface{}{
		"fontname":     fsub,
		"sdf":          sdf,
		"fontfaceinfo": cssfontfacedirectives(s.FontSel),
		"colorinfo":    colors,
	}

	tmpl, e := template.New("fp").Parse(string(j))
	Msg.EC(e)

	var b bytes.Buffer
	err := tmpl.Execute(&b, subs)
	Msg.EC(err)

	css := b.String()

	// if the font is not being served, then replace font names with explicit style directives
	if !slices.Contains(gen.StringMapKeysIntoSlice(vv.ServableFonts), s.FontSel) {
		css = cssmanualfontstyling(css)
	}

	// if the font is being served, but it is relatively hollow when it comes to styles, patch things
	if slices.Contains(gen.StringMapKeysIntoSlice(vv.ServableFonts), s.FontSel) {
		f := vv.ServableFonts[s.FontSel]
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
		src: url('/emb/fnt/{{.ShrtType}}/{{.SubFolder}}/{{.Regular}}') format('{{.Type}}');
		font-display: swap;
		}

	@font-face {
		font-family: 'hipparchiamonostatic';
		src: url('/emb/fnt/{{.ShrtType}}/{{.SubFolder}}/{{.Mono}}') format('{{.Type}}');
		font-display: swap;
		}

	@font-face {
		font-family: 'hipparchialightstatic';
		src: url('/emb/fnt/{{.ShrtType}}/{{.SubFolder}}/{{.Light}}') format('{{.Type}}');
		font-display: swap;
		}

	@font-face {
		font-family: 'hipparchiaboldstatic';
		src: url('/emb/fnt/{{.ShrtType}}/{{.SubFolder}}/{{.Bold}}') format('{{.Type}}');
		font-display: swap;
		}

	@font-face {
		font-family: 'hipparchiaobliquestatic';
		src: url('/emb/fnt/{{.ShrtType}}/{{.SubFolder}}/{{.Italic}}') format('{{.Type}}');
		font-display: swap;
		}

	@font-face {
		font-family: 'hipparchiabolditalicstatic';
		src: url('/emb/fnt/{{.ShrtType}}/{{.SubFolder}}/{{.BoldItalic}}') format('{{.Type}}');
		font-display: swap;
		}

	@font-face {
		font-family: 'hipparchiasemicondensedstatic';
		src: url('/emb/fnt/{{.ShrtType}}/{{.SubFolder}}/{{.SemiCondRegular}}') format('{{.Type}}');
		font-display: swap;
		}

	@font-face {
		font-family: 'hipparchiasemicondenseditalicstatic';
		src: url('/emb/fnt/{{.ShrtType}}/{{.SubFolder}}/{{.SemiCondItalic}}') format('{{.Type}}');
		font-display: swap;
		}

	@font-face {
		font-family: 'hipparchiacondensedstatic';
		src: url('/emb/fnt/{{.ShrtType}}/{{.SubFolder}}/{{.CondensedRegular}}') format('{{.Type}}');
		font-display: swap;
		}

	@font-face {
		font-family: 'hipparchiacondensedboldstatic';
		src: url('/emb/fnt/{{.ShrtType}}/{{.SubFolder}}/{{.CondensedBold}}') format('{{.Type}}');
		font-display: swap;
		}

	@font-face {
		font-family: 'hipparchiacondenseditalicstatic';
		src: url('/emb/fnt/{{.ShrtType}}/{{.SubFolder}}/{{.CondensedItalic}}') format('{{.Type}}');
		font-display: swap;
		}

	@font-face {
		font-family: 'hipparchiasemiboldstatic';
		src: url('/emb/fnt/{{.ShrtType}}/{{.SubFolder}}/{{.SemiBold}}') format('{{.Type}}');
		font-display: swap;
		}

	@font-face {
		font-family: 'hipparchiathinstatic';
		src: url('/emb/fnt/{{.ShrtType}}/{{.SubFolder}}/{{.Thin}}') format('{{.Type}}');
		font-display: swap;
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
	// fmt.Println(css)

	return css
}

// cssmanualfontstyling - swap out: "font-family: 'hipparchiabolditalicstatic', sans-serif;" for explicit style directives
func cssmanualfontstyling(css string) string {
	// swap out: "font-family: 'hipparchiabolditalicstatic', sans-serif;" for explicit style directives
	outtmpl := "font-family: '%s', sans-serif;"
	intempl := "font-family: %s;\n\tfont-weight: %s;\n\tfont-style: %s;\n\tfont-stretch: %s;"
	for n, sw := range cssswaps {
		i := fmt.Sprintf(intempl, sw.familiy, sw.weight, sw.style, sw.stretch)
		o := fmt.Sprintf(outtmpl, n)
		css = strings.ReplaceAll(css, o, i)
	}

	// the above will have missed hipparchiamonostatic
	fswap := cssswaps["hipparchiamonostatic"]
	i := fmt.Sprintf(intempl, fswap.familiy, fswap.weight, fswap.style, fswap.stretch)
	o := fmt.Sprintf("font-family: '%s', monospace;", "hipparchiamonostatic")
	css = strings.ReplaceAll(css, o, i)

	return css
}

func fleshoutcss(f str.FontTempl, css string) string {
	// blank looks like:
	//                font-family: 'hipparchialightstatic';
	//                src: url('/emb/ttf/') format('truetype');

	have := "font-family: '%s';"
	want := "font-family: '%s';\n\t\tfont-weight: %s;\n\t\tfont-style: %s;\n\t\tfont-stretch: %s;"

	for _, n := range f.NeedsManualStyle {
		sw := cssswaps[n]
		w := fmt.Sprintf(want, n, sw.weight, sw.style, sw.stretch)
		h := fmt.Sprintf(have, n)
		css = strings.ReplaceAll(css, h, w)
	}

	bad := "font-family: var(--systemdefaultfont), sans-serif;"
	good := "font-family: 'hipparchiasansstatic';"
	css = strings.ReplaceAll(css, bad, good)
	return css
}

func CheckFontsAvailable(tocheck map[string]str.FontTempl) map[string]str.FontTempl {
	var (
		embfontdir  = "emb/fnt/"
		mincontents = 3 // six font files min; anything less is likely just a comment that the folder was skipped...
		subfolders  = []string{"ttf", "otf"}
	)

	// this will purge available fonts atm: subfolder issue

	type fse struct {
		de fs.DirEntry
		fd string
	}

	var allentries []fse
	for _, sf := range subfolders {
		entries, err := efs.ReadDir(embfontdir + sf)
		if err != nil {
			Msg.CRIT(fmt.Sprintf("CheckFontsAvailable(): Embedded font folder '%s' does not exist", embfontdir+"/"+sf))
		}
		newentries := make([]fse, 0, len(entries))
		for _, entry := range entries {
			newentries = append(newentries, fse{entry, embfontdir + sf + "/"})
		}
		allentries = append(allentries, newentries...)
	}

	avail := make(map[string]str.FontTempl)
	emap := make(map[string]bool)

	for _, e := range allentries {
		ff, ee := efs.ReadDir(e.fd + e.de.Name())
		if ee != nil {
			Msg.CRIT(fmt.Sprintf("CheckFontsAvailable(): Cannot read contents of font folder '%s'", e.fd+e.de.Name()))
		}
		if len(ff) > mincontents {
			emap[e.de.Name()] = true
		}
	}

	for k, v := range tocheck {
		if emap[v.SubFolder] {
			avail[k] = v
		} else {
			Msg.MAND(fmt.Sprintf("CheckFontsAvailable(): Font folder '%s' unavailable", v.SubFolder))
		}
	}
	return avail
}
