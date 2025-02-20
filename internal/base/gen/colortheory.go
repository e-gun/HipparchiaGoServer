package gen

import (
	"bytes"
	"fmt"
	"html/template"
)

const (
	CHEATENDPOINTS = 7
	CHEATINNER     = 3

	SPLITCOMP = `
    --main-body-color: {{.ColorM2}};
    --main-font-color: {{.ColorM7}};
    --input-border-color: {{.ColorM3}};
    
    --buttoncolor: {{.ColorM4}};
    --button-hover: {{.ColorM5}};
    
    --fieldset-background: {{.ColorM3}};
    --focus-shadow: rgba(0, 0, 0, 0.5);
    --icons-color: {{.ColorM5}};

    --black: {{.ColorM1}};
    --blue: {{.ColorT3}}; 
    --brown: {{.ColorT3}};
    --brtblue: {{.ColorT3}};
    --copper: {{.ColorM7}};
    --deepblue: {{.ColorT2}}; 
    --dkbabyblue:  {{.ColorT3}};  
    --dkgreen: {{.ColorM7}};
    --dkgrey: {{.ColorT3}};
    --dkteal: {{.ColorM7}};
    --huedgrey: {{.ColorM7}};
    --invisible: hsla(0, 100%, 0%, 0);
    --lessoffwhite: {{.ColorM3}};
    --ltbabyblue: {{.ColorM4}};
    --ltgrey: {{.ColorM3}};
    --ltgrey2: {{.ColorT3}};
    --midgrey: {{.ColorT3}};
    --offwhite: {{.ColorM7}};
    --orange: {{.ColorT3}};
    --pink: {{.ColorT3}};
    --pinker: {{.ColorM4}};
    --plum: {{.ColorM2}};
    --pukegreen: {{.ColorT3}};
    --red: {{.ColorT3}}; 
    --dkgreen: {{.ColorM7}};
    --rustedorange: {{.ColorM5}}; 
    --sicklyyellow: {{.ColorT3}};
    --skyblue: {{.ColorT3}};
    --teal: {{.ColorT3}};
    --transparentgrey: hsla(0, 0%, 10%, .7);
    --vdkteal: {{.ColorT2}};
    --vdkgrey: {{.ColorM7}};
    --vltgrey: {{.ColorT2}};
    --white: {{.ColorM7}};`

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

	TRIADIC = `
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
	--brown: {{.ColorT3}};
	--brtblue: {{.ColorM7}};
	--copper: {{.ColorM7}};
	--deepblue: {{.ColorT2}}; 
	--dkbabyblue:  {{.ColorM7}};  
	--dkgreen: {{.ColorM7}};
	--dkgrey: {{.ColorM7}};
	--dkteal: {{.ColorM7}};
	--huedgrey: {{.ColorT2}};
	--invisible: hsla(0, 100%, 0%, 0);
	--lessoffwhite: {{.ColorM3}};
	--ltbabyblue: {{.ColorM4}};
	--ltgrey: {{.ColorM3}};
	--ltgrey2: {{.ColorM7}};
	--midgrey: {{.ColorM7}};
	--offwhite: {{.ColorM7}};
	--orange: {{.ColorM7}};
	--pink: {{.ColorM7}};
	--pinker: {{.ColorM4}};
	--plum: {{.ColorM2}};
	--pukegreen: {{.ColorM7}};
	--red: {{.ColorM7}}; 
	--dkgreen: {{.ColorM7}};
	--rustedorange: {{.ColorM5}}; 
	--sicklyyellow: {{.ColorM7}};
	--skyblue: {{.ColorM7}};
	--teal: {{.ColorM7}};
	--transparentgrey: hsla(0, 0%, 10%, .7);
	--vdkteal: {{.ColorT2}};
	--vdkgrey: {{.ColorT2}};
	--vltgrey: {{.ColorT2}};
	--white: {{.ColorM7}};`

	MONO = `
	--main-body-color: {{.ColorM6}};
	--main-font-color: {{.ColorM1}};
	--input-border-color: {{.ColorM5}};
	
	--buttoncolor: {{.ColorM7}};
	--button-hover: {{.ColorM5}};
	
	--fieldset-background: {{.ColorM5}};
	--focus-shadow: rgba(0, 0, 0, 0.5);
	--icons-color: {{.ColorM2}};
	
	--black: {{.ColorM7}};
	--blue: {{.ColorM7}};
	--brown: {{.ColorM7}};
	--brtblue: {{.ColorM3}};
	--copper: {{.ColorM1}};
	--deepblue: {{.ColorM7}};
	--dkbabyblue: {{.ColorM7}};
	--dkgreen: {{.ColorM1}};
	--dkgrey: {{.ColorM1}};
	--dkteal: {{.ColorM1}};
	--huedgrey: {{.ColorM7}};
	--invisible: hsla(0, 100%, 0%, 0);
	--lessoffwhite: {{.ColorM5}};
	--ltbabyblue: {{.ColorM4}};
	--ltgrey: {{.ColorM7}};
	--ltgrey2: {{.ColorM7}};
	--midgrey: {{.ColorM7}};
	--offwhite: {{.ColorM7}};
	--orange: {{.ColorM4}};
	--pink: {{.ColorM4}};
	--pinker: {{.ColorM4}};
	--plum: {{.ColorM6}};
	--pukegreen: {{.ColorM6}};
	--red: {{.ColorM3}};
	--dkgreen: {{.ColorM6}};
	--rustedorange: {{.ColorM1}};
	--sicklyyellow: {{.ColorM4}};
	--skyblue: {{.ColorM6}};
	--teal: {{.ColorM4}};
	--transparentgrey: hsla(0, 0%, 10%, .7);
	--vdkteal: {{.ColorM2}};
	--vdkgrey: {{.ColorM7}};
	--vltgrey: {{.ColorM7}};
	--white: {{.ColorM1}};`
)

func main() {
	// make the following '--main-font-color' (if it is light)
	h := 250
	s := 12
	l := 75
	// fmt.Println(GenerateMonoScheme(h, s, l))
	// fmt.Println(GenerateTriadScheme(h, s, l))
	fmt.Println(GenerateSplitcompScheme(h, s, l))
}

func GenerateMonoScheme(hue int, sat int, lum int) string {
	var blank hueholder
	return generatescheme(hue, sat, lum, blank, MONO)
}

func GenerateTriadScheme(hue int, sat int, lum int) string {
	teth := findtriad(hue)
	return generatescheme(hue, sat, lum, teth, TRIADIC)
}

func GenerateTetradScheme(hue int, sat int, lum int) string {
	teth := findtetradic(hue)
	return generatescheme(hue, sat, lum, teth, TRIADIC)
}

func GenerateSplitcompScheme(hue int, sat int, lum int) string {
	teth := findsplitcomp(hue)
	return generatescheme(hue, sat, lum, teth, SPLITCOMP)
}

func GenerateSquareScheme(hue int, sat int, lum int) string {
	teth := findsquare(hue)
	return generatescheme(hue, sat, lum, teth, TETRADTMPL)
}

func generatescheme(hue int, sat int, lum int, hhld hueholder, tmpl string) string {
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

	var tet tetrad
	tet.Hues = hhld
	tet.Sat = h.Sat
	tet.Lum = h.Lum

	if tet.Lum < 100-CHEATENDPOINTS {
		tet.Lum = tet.Lum + CHEATENDPOINTS
	}

	tet.FleshOut()

	var tc colortemplate
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

	css, e := template.New("mt").Parse(tmpl)
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

// findsquare - hue should be 0-360 (degrees...)
func findsquare(hue int) hueholder {
	if hue > 360 {
		hue = 0
	}

	var rt hueholder
	rt.Hue1 = hue
	rt.Hue3 = rotate(hue, 180)
	rt.Hue2 = rotate(hue, 90)
	rt.Hue4 = rotate(rt.Hue2, 180)
	return rt
}

func findtetradic(hue int) hueholder {
	// 5 hues; three cluster on one side of circle; two on other
	var rt hueholder
	rt.Hue1 = hue
	rt.Hue3 = rotate(hue, -30)
	rt.Hue2 = rotate(hue, 30)
	rt.Hue4 = rotate(hue, -150)
	rt.Hue5 = rotate(hue, 150)
	return rt
}

func findsplitcomp(hue int) hueholder {
	// +/- 150 degrees from original
	var rt hueholder
	rt.Hue1 = hue
	rt.Hue3 = rotate(hue, -150)
	rt.Hue2 = rotate(hue, 150)
	rt.Hue4 = -1
	rt.Hue5 = -1
	return rt
}

func findtriad(hue int) hueholder {
	if hue > 360 {
		hue = 0
	}

	var rt hueholder
	rt.Hue1 = hue
	rt.Hue2 = rotate(hue, 120)
	rt.Hue3 = rotate(hue, 240)
	rt.Hue4 = -1
	rt.Hue5 = -1
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

//
// STRUCTS
//

type tetrad struct {
	ColorT1 string
	ColorT2 string
	ColorT3 string
	ColorT4 string
	Sat     int
	Lum     int
	Hues    hueholder
}

func (t *tetrad) ToString(h int) string {
	return fmt.Sprintf("hsla(%d, %d%%, %d%%, 1)", h, t.Sat, t.Lum)
}

func (t *tetrad) FleshOut() {
	t.ColorT1 = t.ToString(t.Hues.Hue1)
	t.ColorT2 = t.ToString(t.Hues.Hue2)
	t.ColorT3 = t.ToString(t.Hues.Hue3)
	t.ColorT4 = t.ToString(t.Hues.Hue4)
}

type hueholder struct {
	Hue1 int
	Hue2 int
	Hue3 int
	Hue4 int
	Hue5 int
}

type hsl struct {
	Hue int
	Sat int
	Lum int
}

func (h *hsl) ToString() string {
	return fmt.Sprintf("hsla(%d, %d%%, %d%%, 1)", h.Hue, h.Sat, h.Lum)
}

type colortemplate struct {
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
	return fmt.Sprintf("hsla(%d, %d%%, %d%%, 1)", m.Hue, m.Sat, l)
}

func (m *monochrome) ReverseLums() {
	for i, j := 0, len(m.Lums)-1; i < j; i, j = i+1, j-1 {
		m.Lums[i], m.Lums[j] = m.Lums[j], m.Lums[i]
	}
}

func (m *monochrome) FleshOut() {
	m.CheatHighLow()

	m.ColorM1 = m.ToString(m.Lums[0])
	m.ColorM2 = m.ToString(m.Lums[1])
	m.ColorM3 = m.ToString(m.Lums[2])
	m.ColorM4 = m.ToString(m.Lums[3])
	m.ColorM5 = m.ToString(m.Lums[4])
	m.ColorM6 = m.ToString(m.Lums[5])
	m.ColorM7 = m.ToString(m.Lums[6])
}

func (m *monochrome) CheatHighLow() {
	if m.Lums[0] > CHEATENDPOINTS {
		m.Lums[0] = m.Lums[0] - CHEATENDPOINTS
	}

	if m.Lums[1] > CHEATINNER+CHEATENDPOINTS {
		m.Lums[1] = m.Lums[1] - CHEATINNER
	}

	if m.Lums[5] < 100-(CHEATINNER+CHEATENDPOINTS) {
		m.Lums[5] = m.Lums[5] + CHEATINNER
	}

	if m.Lums[6] < 100-CHEATENDPOINTS {
		m.Lums[6] = m.Lums[6] + CHEATENDPOINTS
	}
}
