//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"github.com/fatih/color"
	"sort"
	"strconv"
	"strings"
	"time"
)

//
// DEBUGGING
//

func chke(err error) {
	if err != nil {
		red := color.New(color.FgHiRed).PrintfFunc()
		red("[%s v.%s] UNRECOVERABLE ERROR: PLEASE TAKE NOTE OF THE FOLLOWING PANIC MESSAGE\n", MYNAME, VERSION)
		panic(err)
	}
}

func msg(message string, threshold int) {
	hgc := color.New(color.FgYellow).SprintFunc()
	c := color.FgRed
	switch threshold {
	case -1:
		// c = color.FgHiRed
		c = color.FgGreen
	case 0:
		c = color.FgRed
	case 1:
		c = color.FgHiYellow
	case 2:
		c = color.FgYellow
	case 3:
		c = color.FgWhite
	case 4:
		c = color.FgHiBlack
	case 5:
		c = color.FgHiBlack
	default:
		c = color.FgWhite
	}
	mc := color.New(c).SprintFunc()

	if cfg.LogLevel >= threshold {
		fmt.Printf("[%s] %s\n", hgc(SHORTNAME), mc(message))
	}
}

func timetracker(letter string, m string, start time.Time, previous time.Time) {
	d := fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	m = fmt.Sprintf("[%s: %.3fs]", letter, time.Now().Sub(start).Seconds()) + d + m
	msg(m, TIMETRACKERMSGTHRESH)
}

//
// SETS AND SLICES
//

// RemoveIndex - remove item #N from a slice
func RemoveIndex[T any](s []T, index int) []T {
	// https://stackoverflow.com/questions/37334119/how-to-delete-an-element-from-a-slice-in-golang
	if len(s) < index {
		msg("RemoveIndex() tried to drop an out of range element", 1)
		return s
	}

	ret := make([]T, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}

// unique - return only the unique items from a slice
func unique[T comparable](s []T) []T {
	// https://gosamples.dev/generics-remove-duplicates-slice/
	inResult := make(map[T]bool)
	var result []T
	for _, str := range s {
		if _, ok := inResult[str]; !ok {
			inResult[str] = true
			result = append(result, str)
		}
	}
	return result
}

// setsubtraction - returns [](set(aa) - set(bb))
func setsubtraction[T comparable](aa []T, bb []T) []T {
	//  NB this seems to be SLOW: be careful looping it 10k times
	// 	aa := []string{"a", "b", "c", "d"}
	//	bb := []string{"a", "b", "e", "f"}
	//	dd := setsubtraction(aa, bb)
	//	fmt.Println(dd)
	//  [c d]

	pruner := make(map[T]bool)
	for _, b := range bb {
		pruner[b] = true
	}

	remain := make(map[T]bool)
	for _, a := range aa {
		if _, y := pruner[a]; !y {
			remain[a] = true
		}
	}

	result := make([]T, 0, len(remain))
	for r, _ := range remain {
		result = append(result, r)
	}
	return result
}

// contains - is item X an element of slice A?
func contains[T comparable](sl []T, seek T) bool {
	for _, v := range sl {
		if v == seek {
			return true
		}
	}
	return false
}

// containsN - how many Xs in slice A?
func containsN[T comparable](sl []T, seek T) int {
	count := 0
	for _, v := range sl {
		if v == seek {
			count += 1
		}
	}
	return count
}

// flatten - turn a slice of slices into a slice
func flatten[T any](lists [][]T) []T {
	// https://stackoverflow.com/questions/59579121/how-to-flatten-a-2d-slice-into-1d-slice
	var res []T
	for _, list := range lists {
		res = append(res, list...)
	}
	return res
}

//
// SORTING: https://pkg.go.dev/sort#example__sortMultiKeys
//

type lessFunc func(p1, p2 *DbWorkline) bool

// multiSorter implements the Sort interface, sorting the changes within.
type multiSorter struct {
	changes []DbWorkline
	less    []lessFunc
}

// Sort sorts the argument slice according to the less functions passed to OrderedBy.
func (ms *multiSorter) Sort(changes []DbWorkline) {
	ms.changes = changes
	sort.Sort(ms)
}

// OrderedBy returns a Sorter that sorts using the less functions, in order.
// Call its Sort method to sort the data.
func OrderedBy(less ...lessFunc) *multiSorter {
	return &multiSorter{
		less: less,
	}
}

// Len is part of sort.Interface.
func (ms *multiSorter) Len() int {
	return len(ms.changes)
}

// Swap is part of sort.Interface.
func (ms *multiSorter) Swap(i, j int) {
	ms.changes[i], ms.changes[j] = ms.changes[j], ms.changes[i]
}

func (ms *multiSorter) Less(i, j int) bool {
	p, q := &ms.changes[i], &ms.changes[j]
	// Try all but the last comparison.
	var k int
	for k = 0; k < len(ms.less)-1; k++ {
		less := ms.less[k]
		switch {
		case less(p, q):
			// p < q, so we have a decision.
			return true
		case less(q, p):
			// p > q, so we have a decision.
			return false
		}
		// p == q; try the next comparison.
	}
	// All comparisons to here said "equal", so just return whatever
	// the final comparison reports.
	return ms.less[k](p, q)
}

//
// STRINGS and []RUNE
//

// purgechars - drop any of the chars in the []byte from the string
func purgechars(bad string, checking string) string {
	reducer := make(map[rune]bool)
	for _, r := range []rune(bad) {
		reducer[r] = true
	}

	var stripped []rune
	for _, x := range []rune(checking) {
		if _, skip := reducer[x]; !skip {
			stripped = append(stripped, x)
		} else {
		}
	}
	s := string(stripped)
	return s
}

//
// Geek and Latin functions
//

// stripaccentsSTR - ὀκνεῖϲ --> οκνειϲ, etc.
func stripaccentsSTR(u string) string {
	reducer := getrunereducer()
	var stripped []rune
	for _, x := range []rune(u) {
		stripped = append(stripped, reducer[x])
	}
	s := string(stripped)
	return s
}

// stripaccentsRUNE - ὀκνεῖϲ --> οκνειϲ, etc.
func stripaccentsRUNE(u []rune) []rune {
	reducer := getrunereducer()
	var stripped []rune
	for _, x := range u {
		stripped = append(stripped, reducer[x])
	}
	return stripped
}

func getrunereducer() map[rune]rune {
	// because we don't have access to python's transtable function
	feeder := getrunefeeder()
	reducer := make(map[rune]rune)
	for f, _ := range feeder {
		for _, r := range feeder[f] {
			reducer[r] = f
		}
	}
	return reducer
}

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

func findacuteorgrave(s string) string {
	// prepare regex equiv: ά -> [άὰ]
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

func swapacuteforgrave(thetext string) string {
	swap := strings.NewReplacer("ὰ", "ά", "ὲ", "έ", "ὶ", "ί", "ὸ", "ό", "ὺ", "ύ", "ὴ", "ή", "ὼ", "ώ",
		"ἂ", "ἄ", "ἃ", "ἅ", "ᾲ", "ᾴ", "ᾂ", "ᾄ", "ᾃ", "ᾅ", "ἒ", "ἔ", "ἲ", "ἴ", "ὂ", "ὄ", "ὃ", "ὅ", "ὒ", "ὔ", "ὓ", "ὕ",
		"ἢ", "ἤ", "ἣ", "ἥ", "ᾓ", "ᾕ", "ᾒ", "ᾔ", "ὢ", "ὤ", "ὣ", "ὥ", "ᾣ", "ᾥ", "ᾢ", "ᾤ", "á", "a", "é", "e",
		"í", "i", "ó", "o", "ú", "u")
	return swap.Replace(thetext)
}

func capsvariants(word string) string {
	// build regex compilation template for a word and its capitalized variant
	cv := ""
	rr := []rune(word)
	for _, r := range rr {
		rs := string(r)
		c := strings.ToUpper(rs)
		cv += fmt.Sprintf("[%s%s]", rs, c)
	}
	return cv
}

func uvσςϲ(u string) string {
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

	var stripped []rune
	for _, x := range []rune(u) {
		if _, ok := reducer[x]; ok {
			stripped = append(stripped, reducer[x])
		} else {
			stripped = append(stripped, x)
		}
	}
	s := string(stripped)
	return s

}

func formatbcedate(d string) string {
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

func i64tobce(i int64) string {
	return formatbcedate(fmt.Sprintf("%d", i))
}
