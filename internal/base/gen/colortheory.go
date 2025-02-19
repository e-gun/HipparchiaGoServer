package gen

import (
	"bytes"
	"fmt"
	"html/template"
)

const (
	// TETR - a tatradic + a mono scheme
	// tetradic:
	//
	// {{.ColorM7}}
	// {{.ColorT2}}
	// {{.ColorT3}}
	// {{.ColorT4}}
	//
	// + mono:
	//
	// {{.ColorM1}}
	// {{.ColorM2}}
	// {{.ColorM3}}
	// {{.ColorM4}}
	// {{.ColorM5}}
	// {{.ColorM6}}
	// {{.ColorM7}}
	TETRADTMPL = `
	--main-body-color: {{.ColorM2}};
	--main-font-color: {{.ColorM7}};
	--input-border-color: {{.ColorM3}};
	
	--buttoncolor: {{.ColorM4}};
	--button-hover: {{.ColorM6}};
	
	--fieldset-background: {{.ColorM3}};
	--focus-shadow: rgba(0, 0, 0, 0.5);
	--icons-color: {{.ColorM5}};
	
	--black: {{.ColorM1}};
	--blue: {{.ColorT3}}; 
	--brown: {{.ColorT4}};
	--brtblue: {{.ColorT3}};
	--copper: {{.ColorM7}};
	--deepblue: {{.ColorT2}}; 
	--dkbabyblue:  {{.ColorT3}};  
	--dkgreen: {{.ColorM7}};
	--dkgrey: {{.ColorT3}};
	--dkteal: {{.ColorM7}};
	--huedgrey: {{.ColorT4}};
	--invisible: hsla(0, 100%, 0%, 0);
	--lessoffwhite: {{.ColorM3}};
	--ltbabyblue: {{.ColorM4}};
	--ltgrey: {{.ColorM3}};
	--ltgrey2: {{.ColorT3}};
	--midgrey: {{.ColorT3}};
	--offwhite: {{.ColorM7}};
	--orange: {{.ColorT4}};
	--pink: {{.ColorT4}};
	--pinker: {{.ColorM4}};
	--plum: {{.ColorM2}};
	--pukegreen: {{.ColorT3}};
	--red: {{.ColorT4}}; 
	--dkgreen: {{.ColorM7}};
	--rustedorange: {{.ColorM5}}; 
	--sicklyyellow: {{.ColorT4}};
	--skyblue: {{.ColorT4}};
	--teal: {{.ColorT4}};
	--transparentgrey: hsla(0, 0%, 10%, .7);
	--vdkteal: {{.ColorT2}};
	--vdkgrey: {{.ColorT4}};
	--vltgrey: {{.ColorT2}};
	--white: {{.ColorM7}};`
)

//  What Are Tetradic Colors?
// A tetradic (or double-complementary) scheme involves four colors forming two complementary pairs.
// There are a couple of common approaches, but a standard tetradic relationship often starts with one hue and
// then includes its complement, plus another pair of complementary hues 90° away from each.

type hsl struct {
	Hue int
	Sat int
	Lum int
}

func (h *hsl) ToString() string {
	return fmt.Sprintf("hsla(%d,%d%%, %d%%, 1)", h.Hue, h.Sat, h.Lum)
}

type tetradiccolors struct {
	ColorM1 string
	ColorM2 string
	ColorM3 string
	ColorM4 string
	ColorM5 string
	ColorM6 string
	ColorM7 string
	ColorT1 string
	ColorT2 string
	ColorT3 string
	ColorT4 string
}

type monochrome struct {
	ColorM1 string
	ColorM2 string
	ColorM3 string
	ColorM4 string
	ColorM5 string
	ColorM6 string
	ColorM7 string
	Sat     int
	Lums    [7]int
	Hue     int
}

func (m *monochrome) ToString(l int) string {
	return fmt.Sprintf("hsla(%d,%d%%, %d%%, 1)", m.Hue, m.Sat, l)
}

func (m *monochrome) ReverseLums() {
	for i, j := 0, len(m.Lums)-1; i < j; i, j = i+1, j-1 {
		m.Lums[i], m.Lums[j] = m.Lums[j], m.Lums[i]
	}
}

func (m *monochrome) FleshOut() {
	m.ColorM1 = m.ToString(m.Lums[0])
	m.ColorM2 = m.ToString(m.Lums[1])
	m.ColorM3 = m.ToString(m.Lums[2])
	m.ColorM4 = m.ToString(m.Lums[3])
	m.ColorM5 = m.ToString(m.Lums[4])
	m.ColorM6 = m.ToString(m.Lums[5])
	m.ColorM7 = m.ToString(m.Lums[6])
}

type tetrad struct {
	ColorT1 string
	ColorT2 string
	ColorT3 string
	ColorT4 string
	Sat     int
	Lum     int
	Hues    tetradhues
}

func (t *tetrad) ToString(h int) string {
	return fmt.Sprintf("hsla(%d,%d%%, %d%%, 1)", h, t.Sat, t.Lum)
}

func (t *tetrad) FleshOut() {
	t.ColorT1 = t.ToString(t.Hues.Hue1)
	t.ColorT2 = t.ToString(t.Hues.Hue2)
	t.ColorT3 = t.ToString(t.Hues.Hue3)
	t.ColorT4 = t.ToString(t.Hues.Hue4)
}

type tetradhues struct {
	Hue1 int
	Hue2 int
	Hue3 int
	Hue4 int
}

func rotate(hue int, degr int) int {
	newhue := hue + degr
	if newhue > 360 {
		newhue = newhue - 360
	}
	return newhue
}

// findtetrad - hue should be 0-355 (degrees...)
func findtetrad(hue int) tetradhues {
	if hue > 360 {
		hue = 0
	}

	var rt tetradhues
	rt.Hue1 = hue
	rt.Hue3 = rotate(hue, 180)
	rt.Hue2 = rotate(hue, 90)
	rt.Hue4 = rotate(rt.Hue2, 180)
	return rt
}

func findmonos(h hsl) [7]int {
	// seven evenly spaced items...
	// compress them into a h/l range
	// only lum matters

	var low int
	var high int
	var dark bool

	if h.Lum < 50 {
		low = h.Lum
		dark = true
	} else {
		high = h.Lum
		dark = false
	}

	var distance int
	if dark {
		distance = low
		high = 100 - distance
	} else {
		distance = 100 - high
		low = distance
	}

	rng := high - low
	gap := float32(rng) / 6

	var monos [7]int
	for i := 0; i < 7; i++ {
		monos[i] = int(float32(low) + (gap * float32(i)))
	}
	return monos
}

func GenerateTetradColors(hue int, sat int, lum int) string {
	h := hsl{
		Hue: hue,
		Sat: sat,
		Lum: lum,
	}

	monos := findmonos(h)
	var temphsl hsl
	temphsl.Sat = h.Sat
	temphsl.Lum = h.Lum

	//for _, mono := range monos {
	//	temphsl.Lum = mono
	//	fmt.Println(temphsl.ToString())
	//}

	var mc monochrome
	mc.Sat = h.Sat
	mc.Hue = h.Hue
	mc.Lums = monos

	if h.Lum < 50 {
		mc.ReverseLums()
	}
	mc.FleshOut()

	var teth tetradhues
	teth = findtetrad(h.Hue)

	var tet tetrad
	tet.Hues = teth
	tet.Sat = h.Sat
	tet.Lum = h.Lum

	tet.FleshOut()

	var tc tetradiccolors
	tc.ColorM1 = mc.ColorM1
	tc.ColorM2 = mc.ColorM2
	tc.ColorM3 = mc.ColorM3
	tc.ColorM4 = mc.ColorM4
	tc.ColorM5 = mc.ColorM5
	tc.ColorM6 = mc.ColorM6
	tc.ColorM7 = mc.ColorM7
	tc.ColorT1 = tet.ColorT1
	tc.ColorT2 = tet.ColorT2
	tc.ColorT3 = tet.ColorT3
	tc.ColorT4 = tet.ColorT4

	css, e := template.New("mt").Parse(TETRADTMPL)
	if e != nil {
		panic(e)
	}
	var b bytes.Buffer
	e = css.Execute(&b, tc)
	if e != nil {
		panic(e)
	}
	return b.String()
}
