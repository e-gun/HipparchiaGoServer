package clr

import (
	"bytes"
	"fmt"
	"html/template"
	"slices"
)

func main() {
	// make the following '--main-font-color'
	h := 133
	s := 12
	lum := 10
	minmax := 85
	// fmt.Println(GenerateMonoScheme(h, s, l))
	// fmt.Println(GenerateTriadScheme(h, s, l))
	fmt.Println(GenerateMonoScheme(h, s, lum, minmax))
}

func GenerateMonoScheme(hue int, sat int, lum int, minormaxlum int) string {
	return generatecolorscheme("mono", hue, sat, lum, minormaxlum)
}

func GenerateTriadScheme(hue int, sat int, lum int, minormaxlum int) string {
	return generatecolorscheme("triad", hue, sat, lum, minormaxlum)
}

func GenerateTetradScheme(hue int, sat int, lum int, minormaxlum int) string {
	return generatecolorscheme("tetrad", hue, sat, lum, minormaxlum)
}

func GenerateSplitcompScheme(hue int, sat int, lum int, minormaxlum int) string {
	return generatecolorscheme("splitcomp", hue, sat, lum, minormaxlum)
}

func GenerateSquareScheme(hue int, sat int, lum int, minormaxlum int) string {
	return generatecolorscheme("square", hue, sat, lum, minormaxlum)
}

// generatecolorscheme - the white/black is hslA; black/white is hslB; use scheme 'name'
func generatecolorscheme(name string, hue int, sat int, lum int, minormaxlum int) string {
	pan := buildpancolor(name, hue, sat, lum, minormaxlum)

	csA := pan.m0.Allstrings() // the base
	csB := pan.m1.Allstrings() // the maybe present secondary+ colors
	csC := pan.m2.Allstrings()
	csD := pan.m3.Allstrings()
	csE := pan.m4.Allstrings()

	// not all of these are populated; and not all are useful if they are
	// the midpoints of csB, etc. are likely to be muddy/hard to see
	subs := map[string]interface{}{
		"colorA0": csA[0],
		"colorA1": csA[1],
		"colorA2": csA[2],
		"colorA3": csA[3],
		"colorA4": csA[4],
		"colorA5": csA[5],
		"colorA6": csA[6],
		"colorB0": csB[0],
		"colorB1": csB[1],
		"colorB2": csB[2],
		"colorB3": csB[3],
		"colorB4": csB[4],
		"colorB5": csB[5],
		"colorB6": csB[6],
		"colorC0": csC[0],
		"colorC1": csC[1],
		"colorC2": csC[2],
		"colorC3": csC[3],
		"colorC4": csC[4],
		"colorC5": csC[5],
		"colorC6": csC[6],
		"colorD0": csD[0],
		"colorD1": csD[1],
		"colorD2": csD[2],
		"colorD3": csD[3],
		"colorD4": csD[4],
		"colorD5": csD[5],
		"colorD6": csD[6],
		"colorE0": csE[0],
		"colorE1": csE[1],
		"colorE2": csE[2],
		"colorE3": csE[3],
		"colorE4": csE[4],
		"colorE5": csE[5],
		"colorE6": csE[6],
	}

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

func buildpancolor(name string, hue int, sat int, lum int, minormaxlum int) pancolor {
	c := hsl{
		Hue: hue,
		Sat: sat,
		Lum: lum,
	}

	var pan pancolor
	pan.color = c
	pan.m0 = generatemonostruct(c, minormaxlum)
	m1c := hsl{pan.hues.Hue1, sat, lum}
	m2c := hsl{pan.hues.Hue2, sat, lum}
	m3c := hsl{pan.hues.Hue3, sat, lum}
	m4c := hsl{pan.hues.Hue4, sat, lum}

	switch name {
	case "mono":
		pan.hues = findmonohues(hue)
		pan.tmpl = MONO
	case "triad":
		pan.hues = findtriad(hue)
		pan.m1 = generatemonostruct(m1c, minormaxlum)
		pan.m2 = generatemonostruct(m2c, minormaxlum)
		pan.tmpl = TRIADIC
	case "tetrad":
		pan.hues = findtetradic(hue)
		pan.m1 = generatemonostruct(m1c, minormaxlum)
		pan.m2 = generatemonostruct(m2c, minormaxlum)
		pan.m3 = generatemonostruct(m3c, minormaxlum)
		pan.tmpl = TETRADTMPL
	case "square":
		pan.hues = findsquare(hue)
		pan.m1 = generatemonostruct(m1c, minormaxlum)
		pan.m2 = generatemonostruct(m2c, minormaxlum)
		pan.m3 = generatemonostruct(m3c, minormaxlum)
		pan.tmpl = TETRADTMPL
	case "splitcomp":
		pan.hues = findsplitcomp(hue)
		pan.m1 = generatemonostruct(m1c, minormaxlum)
		pan.m2 = generatemonostruct(m2c, minormaxlum)
		pan.m3 = generatemonostruct(m3c, minormaxlum)
		pan.m4 = generatemonostruct(m4c, minormaxlum)
		pan.tmpl = SPLITCOMP
	default:
		pan.hues = findmonohues(hue)
		pan.tmpl = MONO
	}

	return pan
}

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
		newhue = newhue - 360
	}

	if newhue < 0 {
		newhue = 360 - newhue
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

type pentad struct {
	ColorT1 string
	ColorT2 string
	ColorT3 string
	ColorT4 string
	Sat     int
	Lum     int
	Hues    hueholder
}

func (t *pentad) ToString(h int) string {
	return fmt.Sprintf("hsla(%d, %d%%, %d%%, 1)", h, t.Sat, t.Lum)
}

func (t *pentad) FleshOut() {
	t.ColorT1 = t.ToString(t.Hues.Hue0)
	t.ColorT2 = t.ToString(t.Hues.Hue1)
	t.ColorT3 = t.ToString(t.Hues.Hue2)
	t.ColorT4 = t.ToString(t.Hues.Hue3)
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
