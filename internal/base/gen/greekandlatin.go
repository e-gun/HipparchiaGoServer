//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package gen

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

const (
	TERMINATIONS = `(\s|\.|\]|\<|⟩|\)|’|”|\!|,|:|;|}|\?|⸥|«|—|·|$)` // circular imports means this is declared 2x... see also "vv.constants.go"
	ACCENTED     = `ἂἃἄἅἆἇᾂᾃᾄᾅᾆᾇᾶᾷὰάἒἓἔἕὲέἲἳἴἵἶἷῖίὶῒΐῗΐὂὃὄὅόὸὒὓὔὕὖὗῢΰῦῧύὺᾒᾓᾔᾕᾖᾗῂῄῆῇἤἢἥἣὴήἦἧὢὣὤὥὦὧᾢᾣᾤᾥᾦᾧῲῴῶῷώὼ`
	// GLC is derived from the HGB betacode converter and should be comprehensive
	//GLC      = `ΐάέήίΰαβγδεζηθικλμνξοπρτυφχψωϊϋόύώϝϲἀἁἂἃἄἅἆἇἐἑἒἓἔἕἠἡἢἣἤἥἦἧἰἱἲἳἴἵἶἷὀὁὂὃὄὅὐὑὒὓὔὕὖὗὠὡὢὣὤὥὦὧὰὲὴὶὸὺὼᾀᾁᾂᾃᾄᾅᾆᾇᾐᾑᾒᾓᾔᾕᾖᾗᾠᾡᾢᾣᾤᾥᾦᾧᾲᾳᾴᾶᾷῂῃῄῆῇῒῖῢῤῥῦῧῲῳῴῶῷ`
	//GUC      = `ΑΒΓΔΕΖΗΘΙΚΛΜΝΞΟΠΡΤΥΦΧΨΩϜϹἈἉἊἋἌἍἎἏἘἙἚἛἜἝἨἩἪἫἬἭἮἯἸἹἺἻἼἽἾἿὈὉὊὋὌὍὙὛὝὟὨὩὪὫὬὭὮὯᾊᾋᾌᾍᾎᾏᾚᾛᾜᾝᾞᾟᾪᾫᾬᾭᾮᾯᾼῌῬῼ⒣`
)

var (
	// runefd etc. to avoid looping this in hot code
	runefd         = getrunefeeder()
	exrunefd       = extendedrunefeeder()
	runereduce     = getrunereducer()
	uvreduce       = uvσςϲreducer()
	uvcaps         = uvcapsreducer()
	LunateSwap     = regexp.MustCompile("σ" + TERMINATIONS)
	LexSigmaAbbrev = regexp.MustCompile(`([\s>])ς\. `)     // delunated LSJ entries look weird: `κατὰ ς. `, etc.
	LexSigmaAbbMku = regexp.MustCompile(`\sς(</[^>]+>)\.`) // ` ς</quote>. `
	IsLatin        = regexp.MustCompile(`[a-zA-Z]`)
	FindAccent     = regexp.MustCompile(`[` + ACCENTED + `]`)

	//IsGreekLC  = regexp.MustCompile(`[` + GLC + `]`)
	//IsGreekUC  = regexp.MustCompile(`[` + GUC + `]`)
	//IsGreek    = regexp.MustCompile(`[` + GUC + GLC + `]`)
)

//
// THE EXPORTABLE FUNCTIONS
//

// StripaccentsSTR - ὀκνεῖϲ --> οκνειϲ, etc.
func StripaccentsSTR(u string) string {
	// reducer := getrunereducer()
	ru := []rune(u)
	stripped := make([]rune, len(ru))
	for i, x := range ru {
		stripped[i] = runereduce[x]
	}
	s := string(stripped)
	return s
}

// StripaccentsRUNE - ὀκνεῖϲ --> οκνειϲ, etc.
func StripaccentsRUNE(u []rune) []rune {
	// reducer := getrunereducer()
	stripped := make([]rune, len(u))
	for i, x := range u {
		stripped[i] = runereduce[x]
	}
	return stripped
}

var (
	a4gswap = strings.NewReplacer("ὰ", "ά", "ὲ", "έ", "ὶ", "ί", "ὸ", "ό", "ὺ", "ύ", "ὴ", "ή", "ὼ", "ώ",
		"ἂ", "ἄ", "ἃ", "ἅ", "ᾲ", "ᾴ", "ᾂ", "ᾄ", "ᾃ", "ᾅ", "ἒ", "ἔ", "ἲ", "ἴ", "ὂ", "ὄ", "ὃ", "ὅ", "ὒ", "ὔ", "ὓ", "ὕ",
		"ἢ", "ἤ", "ἣ", "ἥ", "ᾓ", "ᾕ", "ᾒ", "ᾔ", "ὢ", "ὤ", "ὣ", "ὥ", "ᾣ", "ᾥ", "ᾢ", "ᾤ", "á", "a", "é", "e",
		"í", "i", "ó", "o", "ú", "u")
)

// SwapAcuteForGrave - ὰ --> ά
func SwapAcuteForGrave(thetext string) string {
	return a4gswap.Replace(thetext)
}

// SwapGraveForAcute - ά --> ὰ
func SwapGraveForAcute(thetext string) string {
	swap := strings.NewReplacer("ά", "ὰ", "έ", "ὲ", "ί", "ὶ", "ό", "ὸ", "ύ", "ὺ", "ή", "ὴ", "ώ", "ὼ",
		"ἄ", "ἂ", "ἅ", "ἃ", "ᾴ", "ᾲ", "ᾄ", "ᾂ", "ᾅ", "ᾃ", "ἔ", "ἒ", "ἴ", "ἲ", "ὄ", "ὂ", "ὅ", "ὃ", "ὔ", "ὒ", "ὕ", "ὓ",
		"ἤ", "ἢ", "ἥ", "ἣ", "ᾕ", "ᾓ", "ᾔ", "ᾒ", "ὤ", "ὢ", "ὥ", "ὣ", "ᾥ", "ᾣ", "ᾤ", "ᾢ", "a", "á", "e", "é",
		"i", "í", "o", "ó", "u", "ú")
	return swap.Replace(thetext)
}

// CapsVariants - build regex compilation template for a word and its capitalized variant: [aA][bB][cC]
func CapsVariants(word string) string {
	cv := ""
	rr := []rune(word)
	for _, r := range rr {
		rs := string(r)
		c := strings.ToUpper(rs)
		if c != rs {
			cv += fmt.Sprintf("[%s%s]", rs, c)
		} else {
			// to prevent `·` from becoming `··` if you somehow sent something that was not a pure `word` to this fnc
			cv += c
		}
	}
	return cv
}

// UVσςϲCapsVariants - UV aware version of CapsVariants: [aA][bB][cC][uUvV]
func UVσςϲCapsVariants(word string) string {
	cv := CapsVariants(word)
	cv = strings.Replace(cv, "[uU]", "[uUvV]", -1)
	cv = strings.Replace(cv, "[ϲϹ]", "[ϲϹσςΣ]", -1)
	return cv
}

// UVσςϲ - v to u, etc
func UVσςϲ(u string) string {
	ru := []rune(u)
	stripped := make([]rune, len(ru))
	for i, x := range ru {
		if _, ok := uvreduce[x]; ok {
			stripped[i] = uvreduce[x]
		} else {
			stripped[i] = x
		}
	}
	s := string(stripped)
	return s
}

// UVcaps - v to u, etc
func UVcaps(u string) string {
	ru := []rune(u)
	stripped := make([]rune, len(ru))
	for i, x := range ru {
		if _, ok := uvcaps[x]; ok {
			stripped[i] = uvcaps[x]
		} else {
			stripped[i] = x
		}
	}
	s := string(stripped)
	return s
}

// RestoreInitialVJ - semantic vectors need v and j; the hinter does not have them; note that this is too stupid to do non-initial letters
func RestoreInitialVJ(vj string) string {
	vv := regexp.MustCompile(`^u([aeiou])`)
	jj := regexp.MustCompile(`^i([aeiou])`)
	vj = vv.ReplaceAllString(vj, "v$1")
	vj = jj.ReplaceAllString(vj, "j$1")
	return vj
}

// FindAcuteOrGrave - prepare regex equiv: ά -> [άὰ]
func FindAcuteOrGrave(s string) string {
	feeder := make(map[rune][]rune)
	feeder['ά'] = []rune("ὰά")
	feeder['έ'] = []rune("ὲέ")
	feeder['ή'] = []rune("ὴή")
	feeder['ί'] = []rune("ὶί")
	feeder['ό'] = []rune("όὸ")
	feeder['ύ'] = []rune("ύὺ")
	feeder['ώ'] = []rune("ώὼ")
	feeder['ἂ'] = []rune("ἂἄ")
	feeder['ἒ'] = []rune("ἒἔ")
	feeder['ἢ'] = []rune("ἢἤ")
	feeder['ἲ'] = []rune("ἲἴ")
	feeder['ὂ'] = []rune("ὂὄ")
	feeder['ὒ'] = []rune("ὒὔ")
	feeder['ὓ'] = []rune("ὓὕ")
	feeder['ὢ'] = []rune("ὢὤ")
	feeder['ὣ'] = []rune("ὣὥ")
	feeder['ἃ'] = []rune("ἅἃ")
	feeder['ᾲ'] = []rune("ᾲᾴ")
	feeder['ᾂ'] = []rune("ᾂᾄ")
	feeder['ἣ'] = []rune("ἣἥ")
	feeder['ᾒ'] = []rune("ᾒᾔ")
	feeder['ᾓ'] = []rune("ᾓᾕ")
	feeder['ὃ'] = []rune("ὃὅ")
	feeder['ὂ'] = []rune("ὂὄ")
	feeder['ὒ'] = []rune("ὒὔ")
	feeder['ᾂ'] = []rune("ᾂᾄ")
	feeder['ᾃ'] = []rune("ᾃᾅ")
	feeder['ᾢ'] = []rune("ᾢᾤ")
	feeder['ᾣ'] = []rune("ᾣᾥ")

	rr := []rune(s)
	var mod []rune
	for _, r := range rr {
		if _, ok := feeder[r]; ok {
			st := fmt.Sprintf("[%s]", string(feeder[r]))
			mod = append(mod, []rune(st)...)
		} else {
			mod = append(mod, r)
		}
	}
	return string(mod)
}

// DeLunate - Τὴν οὖν τῶν ϲωμάτων ϲύνταξιν ϲκεψαμένουϲ πρὸϲ --> Τὴν οὖν τῶν σωμάτων σύνταξιν σκεψαμένους πρὸς
func DeLunate(txt string) string {
	txt = strings.Replace(txt, "ϲ", "σ", -1)
	txt = strings.Replace(txt, "Ϲ", "Σ", -1)
	txt = LunateSwap.ReplaceAllString(txt, "ς${1}")
	return txt
}

// LexDeLunate - same as DeLunate but add checks for the abbreviation of examples in sigma-entries:
func LexDeLunate(txt string) string {
	// example is σιτίον and `τἀν Πρυτανείῳ ς. public maintenance` --> `τἀν Πρυτανείῳ σ. public maintenance`
	txt = DeLunate(txt)
	txt = LexSigmaAbbrev.ReplaceAllString(txt, "${1}σ. ")
	// txt = strings.Replace(txt, "ς</hb-lx-lg-grc>.", "σ</hb-lx-lg-grc>.", -1) // not worth building a regex
	txt = LexSigmaAbbMku.ReplaceAllString(txt, " σ${1}. ")
	return txt
}

// ReLunate -  Τὴν οὖν τῶν σωμάτων σύνταξιν σκεψαμένους πρὸς --> Τὴν οὖν τῶν ϲωμάτων ϲύνταξιν ϲκεψαμένουϲ πρὸϲ
func ReLunate(txt string) string {
	txt = strings.Replace(txt, "σ", "ϲ", -1)
	txt = strings.Replace(txt, "ς", "ϲ", -1)
	txt = strings.Replace(txt, "Σ", "Ϲ", -1)
	return txt
}

// FormatBCEDate - turn "-300" into "300 B.C.E."
func FormatBCEDate(d string) string {
	s, e := strconv.Atoi(d)
	if e != nil {
		s = 9999
	}
	if s > 0 {
		d += " C.E."
	} else {
		d = strings.Replace(d, "-", "", -1) + " B.C.E."
	}
	return d
}

// IntToBCE - turn an int into something like "300 B.C.E."
func IntToBCE(i int) string {
	return FormatBCEDate(fmt.Sprintf("%d", i))
}

type PolytonicSorterStruct struct {
	Sortstring     string
	Originalstring string
	Count          int // Count is needed by RtIndexMaker()
}

// PolytonicSort - sort a slice of polytonic words; note that this is going to be relatively costly to execute
func PolytonicSort(pt []string) []string {
	ss := make([]PolytonicSorterStruct, len(pt))
	for i := 0; i < len(pt); i++ {
		ss[i] = PolytonicSorterStruct{
			Sortstring:     strings.Replace(StripaccentsSTR(pt[i]), "ϲ", "σ", -1) + pt[i],
			Originalstring: pt[i],
			Count:          0,
		}
	}

	slices.SortFunc(ss, func(a, b PolytonicSorterStruct) int { return cmp.Compare(a.Sortstring, b.Sortstring) })

	for i := 0; i < len(pt); i++ {
		pt[i] = ss[i].Originalstring
	}

	return pt
}

//
// THE HELPERS/FEEDERS
//

func getrunereducer() map[rune]rune {
	// because we don't have access to python's transtable function
	// runefd := getrunefeeder()
	// runefd now a var at top of file

	reducer := make(map[rune]rune)
	for f, _ := range runefd {
		for _, r := range runefd[f] {
			reducer[r] = f
		}
	}
	return reducer
}

// getrunefeeder - this one will de-capitalize and de-accentuate (needed for various strippers)
func getrunefeeder() map[rune][]rune {
	feeder := make(map[rune][]rune)
	feeder['α'] = []rune("αἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάᾈᾉᾊᾋᾌᾍᾎᾏἈἉἊἋἌἍἎἏΑ")
	feeder['ε'] = []rune("εἐἑἒἓἔἕὲέἘἙἚἛἜἝΕ")
	feeder['ι'] = []rune("ιἰἱἲἳἴἵἶἷὶίῐῑῒΐῖῗΐἸἹἺἻἼἽἾἿΙ")
	feeder['ο'] = []rune("οὀὁὂὃὄὅόὸὈὉὊὋὌὍΟ")
	feeder['υ'] = []rune("υὐὑὒὓὔὕὖὗϋῠῡῢΰῦῧύὺὙὛὝὟΥ")
	feeder['η'] = []rune("ηᾐᾑᾒᾓᾔᾕᾖᾗῂῃῄῆῇἤἢἥἣὴήἠἡἦἧᾘᾙᾚᾛᾜᾝᾞᾟἨἩἪἫἬἭἮἯΗ")
	feeder['ω'] = []rune("ωὠὡὢὣὤὥὦὧᾠᾡᾢᾣᾤᾥᾦᾧῲῳῴῶῷώὼᾨᾩᾪᾫᾬᾭᾮᾯὨὩὪὫὬὭὮὯ")
	feeder['ρ'] = []rune("ρῤῥῬ")
	feeder['β'] = []rune("βΒ")
	feeder['ψ'] = []rune("ψΨ")
	feeder['δ'] = []rune("δΔ")
	feeder['φ'] = []rune("φΦ")
	feeder['γ'] = []rune("γΓ")
	feeder['ξ'] = []rune("ξΞ")
	feeder['κ'] = []rune("κΚ")
	feeder['λ'] = []rune("λΛ")
	feeder['μ'] = []rune("μΜ")
	feeder['ν'] = []rune("νΝ")
	feeder['π'] = []rune("πΠ")
	feeder['ϙ'] = []rune("ϙϘ")
	feeder['ϲ'] = []rune("ϲσΣςϹ")
	feeder['τ'] = []rune("τΤ")
	feeder['χ'] = []rune("χΧ")
	feeder['θ'] = []rune("θΘ")
	feeder['ζ'] = []rune("ζΖ")
	feeder['a'] = []rune("aAÁÄáäă")
	feeder['b'] = []rune("bB")
	feeder['c'] = []rune("cC")
	feeder['d'] = []rune("dD")
	feeder['e'] = []rune("eEÉËéëāĕē")
	feeder['f'] = []rune("fF")
	feeder['g'] = []rune("gG")
	feeder['h'] = []rune("hH")
	feeder['i'] = []rune("iIÍÏíïJj")
	feeder['k'] = []rune("kK")
	feeder['l'] = []rune("lL")
	feeder['m'] = []rune("mM")
	feeder['n'] = []rune("nN")
	feeder['o'] = []rune("oOÓÖóöŏō")
	feeder['p'] = []rune("pP")
	feeder['q'] = []rune("qQ")
	feeder['r'] = []rune("rR")
	feeder['s'] = []rune("sS")
	feeder['t'] = []rune("tT")
	feeder['u'] = []rune("uUvVÜÚüú")
	feeder['w'] = []rune("wW")
	feeder['x'] = []rune("xX")
	feeder['y'] = []rune("yY")
	feeder['z'] = []rune("zZ")
	return feeder
}

// extendedrunefeeder - this one will do acute for grave (needed for lemma highlighting)
func extendedrunefeeder() map[rune][]rune {
	feeder := getrunefeeder()
	feeder['ά'] = []rune("ὰά")
	feeder['έ'] = []rune("ὲέ")
	feeder['ή'] = []rune("ὴή")
	feeder['ί'] = []rune("ὶί")
	feeder['ό'] = []rune("όὸ")
	feeder['ύ'] = []rune("ύὺ")
	feeder['ώ'] = []rune("ώὼ")
	feeder['ἂ'] = []rune("ἂἄ")
	feeder['ἒ'] = []rune("ἒἔ")
	feeder['ἢ'] = []rune("ἢἤ")
	feeder['ἲ'] = []rune("ἲἴ")
	feeder['ὂ'] = []rune("ὂὄ")
	feeder['ὒ'] = []rune("ὒὔ")
	feeder['ὓ'] = []rune("ὓὕ")
	feeder['ὢ'] = []rune("ὢὤ")
	feeder['ὣ'] = []rune("ὣὥ")
	feeder['ἃ'] = []rune("ἅἃ")
	feeder['ᾲ'] = []rune("ᾲᾴ")
	feeder['ᾂ'] = []rune("ᾂᾄ")
	feeder['ἣ'] = []rune("ἣἥ")
	feeder['ᾒ'] = []rune("ᾒᾔ")
	feeder['ᾓ'] = []rune("ᾓᾕ")
	feeder['ὃ'] = []rune("ὃὅ")
	feeder['ὂ'] = []rune("ὂὄ")
	feeder['ὒ'] = []rune("ὒὔ")
	feeder['ᾂ'] = []rune("ᾂᾄ")
	feeder['ᾃ'] = []rune("ᾃᾅ")
	feeder['ᾢ'] = []rune("ᾢᾤ")
	feeder['ᾣ'] = []rune("ᾣᾥ")
	return feeder
}

// uvσςϲreducer - provide map to UVσςϲ
func uvσςϲreducer() map[rune]rune {
	// map[73:105 74:105 85:117 86:117 105:105 106:105 ...]
	feeder := make(map[rune][]rune)

	feeder['u'] = []rune("uUvVÜÚüú")
	feeder['ϲ'] = []rune("ϲσΣςϹ")
	feeder['i'] = []rune("iIÍÏíïJj")

	reducer := make(map[rune]rune)
	for f, _ := range feeder {
		for _, r := range feeder[f] {
			reducer[r] = f
		}
	}
	return reducer
}

// uvcapsreducer - provide map to UVcaps
func uvcapsreducer() map[rune]rune {
	// map[73:105 74:105 85:117 86:117 105:105 106:105 ...]
	feeder := make(map[rune][]rune)

	feeder['u'] = []rune("uv")
	feeder['i'] = []rune("ij")
	feeder['U'] = []rune("ÜÚ")
	feeder['V'] = []rune("U")
	feeder['I'] = []rune("IJ")

	reducer := make(map[rune]rune)
	for f, _ := range feeder {
		for _, r := range feeder[f] {
			reducer[r] = f
		}
	}
	return reducer
}

// StripExtraAccent - morphology lookups fail on these words; so οἷόν --> οἷον; θαυμάζεταί --> θαυμάζεται; μίμηϲίϲ --> μίμηϲιϲ
func StripExtraAccent(w string) string {
	acc := FindAccent.FindAllStringSubmatch(w, -1)

	if len(acc) < 2 {
		return w
	}

	chunks := FindAccent.Split(w, -1)

	var parts []string
	if len(chunks) == 3 {
		parts = []string{
			chunks[0],                  // μ
			acc[0][0],                  // ί
			chunks[1],                  // μηϲ
			StripaccentsSTR(acc[1][0]), // ί --> ι
			chunks[2],                  // ϲ
		}
	}

	if len(chunks) == 2 {
		parts = []string{
			acc[0][0],
			chunks[0],
			StripaccentsSTR(acc[1][0]),
			chunks[1],
		}
	}

	// fmt.Println(w, "multiple accents; -->", strings.Join(parts, ""))
	return strings.Join(parts, "")
}
