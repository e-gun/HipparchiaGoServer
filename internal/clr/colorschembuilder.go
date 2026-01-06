//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package clr

import (
	"bytes"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"text/template"
)

func main() {
	// make the following '--main-font-color'
	h := 133
	s := 12
	lum := 10
	minmax := 85
	fmt.Println(generatemonoscheme([]int{h, s, lum, minmax}))
}

func generatemonoscheme(hslm []int) string {
	return generatecolorscheme("mono", hslm[0], hslm[1], hslm[2], hslm[3])
}

func generatetriatscheme(hslm []int) string {
	return generatecolorscheme("triad", hslm[0], hslm[1], hslm[2], hslm[3])
}

func generatetetradscheme(hslm []int) string {
	return generatecolorscheme("tetrad", hslm[0], hslm[1], hslm[2], hslm[3])
}

func generatesplitcompscheme(hslm []int) string {
	return generatecolorscheme("splitcomp", hslm[0], hslm[1], hslm[2], hslm[3])
}

func generatesquarescheme(hslm []int) string {
	return generatecolorscheme("square", hslm[0], hslm[1], hslm[2], hslm[3])
}

func FindVectorDotHueAndLums(schemename string, hslm []int) (int, int, int) {
	name := convertschemname(schemename)
	pan := buildpancolor(name, hslm[0], hslm[1], hslm[2], hslm[3])

	var newhue int
	var newlum int
	var newperiphlum int

	if name == "mono" {
		newhue = pan.m0.Hue
		newlum = pan.m0.Lums[1]
		newperiphlum = pan.m0.Lums[2]
	} else {
		newhue = pan.m1.Hue
		newlum = pan.m1.Lums[5]
		newperiphlum = pan.m1.Lums[4]
	}
	return newhue, newlum, newperiphlum
}

func FindVectorLineHueAndLums(schemename string, hslm []int) (int, int) {
	name := convertschemname(schemename)
	pan := buildpancolor(name, hslm[0], hslm[1], hslm[2], hslm[3])

	var newhue int
	var newlum int

	if name == "mono" {
		newhue = pan.m0.Hue
		newlum = pan.m0.Lums[1]
	} else {
		// there are always three+ colors available...
		newhue = pan.m2.Hue
		newlum = pan.m2.Lums[4]
	}
	return newhue, newlum
}

func FindVectorFontHueAndLums(schemename string, hslm []int) (int, int) {
	name := convertschemname(schemename)
	pan := buildpancolor(name, hslm[0], hslm[1], hslm[2], hslm[3])

	var newhue int
	var newlum int

	if name == "mono" {
		newhue = pan.m0.Hue
		newlum = pan.m0.Lums[0]
	} else {
		// there are always three+ colors available...
		newhue = pan.m0.Hue
		newlum = pan.m0.Lums[6]
	}
	return newhue, newlum
}

func convertschemname(schemename string) string {
	var name string
	switch schemename {
	case "Light":
		name = "mono"
	case "Dark":
		name = "mono"
	case "Monochrome":
		name = "mono"
	case "Triadic":
		name = "triad"
	case "Tetradic":
		name = "tetrad"
	case "SplitComp":
		name = "splitcomp"
	case "Square":
		name = "square"
	default:
		name = "mono"
	}
	return name
}

// generatecolorscheme - the white/black is hslA; black/white is hslB; use scheme 'name'
func generatecolorscheme(name string, hue int, sat int, lum int, minormaxlum int) string {
	pan := buildpancolor(name, hue, sat, lum, minormaxlum)
	subs := preparestringsubs(name, hue, sat, lum, minormaxlum)

	css, e := template.New("mt").Parse(pan.tmpl)
	if e != nil {
		panic(e)
	}
	var b bytes.Buffer
	e = css.Execute(&b, subs)
	if e != nil {
		panic(e)
	}

	return b.String()
}

func preparestringsubs(name string, hue int, sat int, lum int, minormaxlum int) map[string]interface{} {
	pan := buildpancolor(name, hue, sat, lum, minormaxlum)

	transp := gettextshadowcolor(name, lum, minormaxlum, pan)

	csA := pan.m0.Allstrings() // the base
	csB := pan.m1.Allstrings() // the maybe present secondary+ colors
	csC := pan.m2.Allstrings()
	csD := pan.m3.Allstrings()
	csE := pan.m4.Allstrings()

	// not all of these are populated; and not all are useful if they are
	// the midpoints of csB, etc. are likely to be muddy/hard to see

	subs := map[string]interface{}{
		"colorA0":         csA[0],
		"colorA1":         csA[1],
		"colorA2":         csA[2],
		"colorA3":         csA[3],
		"colorA4":         csA[4],
		"colorA5":         csA[5],
		"colorA6":         csA[6],
		"colorB0":         csB[0],
		"colorB1":         csB[1],
		"colorB2":         csB[2],
		"colorB3":         csB[3],
		"colorB4":         csB[4],
		"colorB5":         csB[5],
		"colorB6":         csB[6],
		"colorC0":         csC[0],
		"colorC1":         csC[1],
		"colorC2":         csC[2],
		"colorC3":         csC[3],
		"colorC4":         csC[4],
		"colorC5":         csC[5],
		"colorC6":         csC[6],
		"colorD0":         csD[0],
		"colorD1":         csD[1],
		"colorD2":         csD[2],
		"colorD3":         csD[3],
		"colorD4":         csD[4],
		"colorD5":         csD[5],
		"colorD6":         csD[6],
		"colorE0":         csE[0],
		"colorE1":         csE[1],
		"colorE2":         csE[2],
		"colorE3":         csE[3],
		"colorE4":         csE[4],
		"colorE5":         csE[5],
		"colorE6":         csE[6],
		"textshadowColor": transp,
	}
	return subs
}

// gettextshadowcolor - this depends on whether we have a dark font on a light background or not
func gettextshadowcolor(name string, lum int, minormaxlum int, pan pancolor) string {
	// bleh: the whole thing is a kludge and brittle...

	// assume light text on a dark background and so a "bright" text shadow
	transp := fmt.Sprintf("hsla(%d, %d%%, 80%%, .80)", pan.m0.Hue, pan.m0.Sat)

	// usually a0/a1 is the main font color and a5/a6 is the background...
	// but this is not true for "mono"
	a0fontona6background := map[string]bool{
		"mono":      false,
		"splitcomp": true,
		"square":    true,
		"tetrad":    true,
		"triad":     true,
	}

	// is lum the high or the low in the set of lums?
	// if so, then a0 is a small (i.e dark) value and a6 is a high (i.e. light) value
	minormaxismin := true
	if minormaxlum > lum {
		minormaxismin = false
	}

	// now do we want that bright text shadow or not?
	fliptodarkshadow := false
	if a0fontona6background[name] {
		if !minormaxismin {
			fliptodarkshadow = true
		}
	} else if !a0fontona6background[name] {
		// yes, the logic could be collapsed; but this is more readable
		if minormaxismin {
			fliptodarkshadow = true
		}
	}

	if fliptodarkshadow {
		transp = fmt.Sprintf("hsla(%d, %d%%, 20%%, .66)", pan.m0.Hue, pan.m0.Sat)
	}

	return transp
}

func buildpancolor(name string, hue int, sat int, lum int, minormaxlum int) pancolor {
	c := hsl{
		Hue: hue,
		Sat: sat,
		Lum: lum,
	}

	var pan pancolor
	pan.color = c
	pan.m0 = generatemonostruct(c, minormaxlum)

	buildhsl := func(p pancolor) [4]hsl {
		var hhs [4]hsl
		hhs[0] = hsl{p.hues.Hue1, sat, lum}
		hhs[1] = hsl{p.hues.Hue2, sat, lum}
		hhs[2] = hsl{p.hues.Hue3, sat, lum}
		hhs[3] = hsl{p.hues.Hue4, sat, lum}
		return hhs
	}

	switch name {
	case "mono":
		pan.hues = findmonohues(hue)
		pan.tmpl = MONO
	case "splitcomp":
		pan.hues = findsplitcomp(hue)
		hhs := buildhsl(pan)
		pan.m1 = generatemonostruct(hhs[0], minormaxlum)
		pan.m2 = generatemonostruct(hhs[1], minormaxlum)
		pan.tmpl = SPLITCOMP
	case "square":
		pan.hues = findsquare(hue)
		hhs := buildhsl(pan)
		pan.m1 = generatemonostruct(hhs[0], minormaxlum)
		pan.m2 = generatemonostruct(hhs[1], minormaxlum)
		pan.m3 = generatemonostruct(hhs[2], minormaxlum)
		pan.tmpl = SQUARE
	case "tetrad":
		pan.hues = findtetradic(hue)
		hhs := buildhsl(pan)
		pan.m1 = generatemonostruct(hhs[0], minormaxlum)
		pan.m2 = generatemonostruct(hhs[1], minormaxlum)
		pan.m3 = generatemonostruct(hhs[2], minormaxlum)
		pan.m4 = generatemonostruct(hhs[3], minormaxlum)
		pan.tmpl = TETRADTMPL
	case "triad":
		pan.hues = findtriad(hue)
		hhs := buildhsl(pan)
		pan.m1 = generatemonostruct(hhs[0], minormaxlum)
		pan.m2 = generatemonostruct(hhs[1], minormaxlum)
		pan.tmpl = TRIADIC
	default:
		pan.hues = findmonohues(hue)
		pan.tmpl = MONO
	}

	return pan
}

func colorsamplesheet(schemename string, hslm []int) string {
	return colornamesamplesheet(schemename, hslm) + "<br>" + colorcsssamplesheet(schemename, hslm)
}

func colornamesamplesheet(schemename string, hslm []int) string {
	// A0 A0   ▣▣▣▣▣    A0 A6     ▣▣▣▣▣       hsla(20, 17%, 22%, 1)
	const (
		TMPL = `
		%s%d A0 <span style="background-color: {{.colorA0}}; color: {{.color%s%d}}">&nbsp;&nbsp;▣▣▣▣▣&nbsp;&nbsp;</span>&nbsp;&nbsp;%s%d A6&nbsp;&nbsp;
		<span style="background-color: {{.colorA6}}; color: {{.color%s%d}}">&nbsp;&nbsp;▣▣▣▣▣&nbsp;&nbsp;</span> 
		&nbsp;&nbsp;&nbsp;&nbsp;{{.color%s%d}}`
	)

	name := convertschemname(schemename)

	subs := preparestringsubs(name, hslm[0], hslm[1], hslm[2], hslm[3])

	lett := []string{"A", "B", "C", "D", "E"}

	sheet := []string{}

	for _, l := range lett {
		for n := range 7 {
			sheet = append(sheet, fmt.Sprintf(TMPL, l, n, l, n, l, n, l, n, l, n))
		}
	}

	sample := strings.Join(sheet, "<br>")

	css, e := template.New("mt").Parse(sample)
	if e != nil {
		panic(e)
	}
	var b bytes.Buffer
	e = css.Execute(&b, subs)
	if e != nil {
		panic(e)
	}

	cs := b.String()

	var filtered []string
	tofilter := strings.Split(cs, "\n")
	for flt := range tofilter {
		if !strings.Contains(tofilter[flt], "hsla(0, 0%, 0%, 1)") {
			filtered = append(filtered, tofilter[flt])
		}
	}
	return strings.Join(filtered, "\n")
}

func colorcsssamplesheet(schemename string, hslm []int) string {
	// colorized version of:
	// A0 A6   --main-body-color

	const (
		TMPLLT = `$2$3 A6 <span style="background-color: {{.colorA6}}; color: {{.color$2$3}}">&nbsp;&nbsp;--$1&nbsp;&nbsp;</span> `
		TMPLDK = `$2$3 A0 <span style="background-color: {{.colorA0}}; color: {{.color$2$3}}">&nbsp;&nbsp;--$1&nbsp;&nbsp;</span> `
	)

	finder := regexp.MustCompile("--(.*?): ...color(.)(.)}};")

	name := convertschemname(schemename)
	pan := buildpancolor(name, hslm[0], hslm[1], hslm[2], hslm[3])
	csslist := strings.Split(pan.tmpl, "\n")

	subs := preparestringsubs(name, hslm[0], hslm[1], hslm[2], hslm[3])

	var accumulate []string
	for _, cs := range csslist {
		if strings.Contains(cs, "{{.color") {
			accumulate = append(accumulate, finder.ReplaceAllString(cs, TMPLLT))
		}
	}
	for _, cs := range csslist {
		if strings.Contains(cs, "{{.color") {
			accumulate = append(accumulate, finder.ReplaceAllString(cs, TMPLDK))
		}
	}

	sample := strings.Join(accumulate, "<br>")

	css, e := template.New("mt").Parse(sample)
	if e != nil {
		panic(e)
	}
	var b bytes.Buffer
	e = css.Execute(&b, subs)
	if e != nil {
		panic(e)
	}
	cs := b.String()
	return cs
}

//
// UTILS
//

// findmonohues - blank, but for parity
func findmonohues(hue int) hueholder {
	return hueholder{
		Hue0: hue,
		Hue1: -1,
		Hue2: -1,
		Hue3: -1,
		Hue4: -1,
	}
}

// findsquare - hue should be 0-360 (degrees...)
func findsquare(hue int) hueholder {
	if hue > 360 {
		hue = 0
	}
	var rt hueholder
	rt.Hue0 = hue
	rt.Hue2 = rotate(hue, 180)
	rt.Hue1 = rotate(hue, 90)
	rt.Hue3 = rotate(rt.Hue1, 180)
	rt.Hue4 = -1
	return rt
}

func findtetradic(hue int) hueholder {
	// 5 hues; three cluster on one side of circle; two on other
	var rt hueholder
	rt.Hue0 = hue
	rt.Hue2 = rotate(hue, -30)
	rt.Hue1 = rotate(hue, 30)
	rt.Hue3 = rotate(hue, -150)
	rt.Hue4 = rotate(hue, 150)
	return rt
}

func findsplitcomp(hue int) hueholder {
	// +/- 150 degrees from original
	var rt hueholder
	rt.Hue0 = hue
	rt.Hue2 = rotate(hue, -150)
	rt.Hue1 = rotate(hue, 150)
	rt.Hue3 = -1
	rt.Hue4 = -1
	return rt
}

func findtriad(hue int) hueholder {
	if hue > 360 {
		hue = 360 - hue
	}

	var rt hueholder
	rt.Hue0 = hue
	rt.Hue1 = rotate(hue, 120)
	rt.Hue2 = rotate(hue, 240)
	rt.Hue3 = -1
	rt.Hue4 = -1
	return rt
}

func generatemonostruct(h hsl, lim int) monochrome {
	// seven evenly spaced items...
	// compress them into a h/l range
	// only lum matters

	var low int
	var high int

	if h.Lum < 50 {
		low = h.Lum
		high = lim
	} else {
		low = lim
		high = h.Lum
	}

	rng := high - low
	gap := float32(rng) / 6

	var monos [7]int

	for i := 0; i < 7; i++ {
		monos[i] = int(float32(low) + (gap * float32(i)))
	}

	if h.Lum < lim {
		slices.Reverse(monos[:])
	}

	return monochrome{
		Hue:  h.Hue,
		Sat:  h.Sat,
		Lums: monos,
	}
}

func rotate(hue int, degr int) int {
	newhue := hue + degr
	if newhue > 360 {
		return newhue - 360
	}

	if newhue < 0 {
		return 360 + newhue
	}

	return newhue
}

// STRUCTS
type pancolor struct {
	color hsl
	hues  hueholder
	m0    monochrome
	m1    monochrome
	m2    monochrome
	m3    monochrome
	m4    monochrome
	tmpl  string
}

type hueholder struct {
	Hue0 int
	Hue1 int
	Hue2 int
	Hue3 int
	Hue4 int
}

type hsl struct {
	Hue int
	Sat int
	Lum int
}

func (h *hsl) ToString() string {
	return fmt.Sprintf("hsla(%d, %d%%, %d%%, 1)", h.Hue, h.Sat, h.Lum)
}

type monochrome struct {
	Hue  int
	Sat  int
	Lums [7]int
}

func (m *monochrome) ToString(l int) string {
	return fmt.Sprintf("hsla(%d, %d%%, %d%%, 1)", m.Hue, m.Sat, l)
}

func (m *monochrome) Allstrings() [7]string {
	var allstrings [7]string
	for i := 0; i < 7; i++ {
		allstrings[i] = m.ToString(m.Lums[i])
	}
	return allstrings
}

func (m *monochrome) ReverseLums() {
	for i, j := 0, len(m.Lums)-1; i < j; i, j = i+1, j-1 {
		m.Lums[i], m.Lums[j] = m.Lums[j], m.Lums[i]
	}
}

func (m *monochrome) Empty() {
	m.Sat = -1
	m.Hue = -1
	m.Lums = [7]int{-1, -1, -1, -1, -1, -1, -1}
}
