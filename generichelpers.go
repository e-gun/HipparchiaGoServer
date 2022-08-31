package main

import (
	"fmt"
	"github.com/fatih/color"
	"sort"
	"time"
)

//
// DEBUGGING
//

func chke(err error) {
	if err != nil {
		red := color.New(color.FgHiRed).PrintfFunc()
		red("UNRECOVERABLE ERROR: PLEASE TAKE NOTE OF THE FOLLOWING PANIC MESSAGE [%s v.%s]\n", MYNAME, VERSION)
		panic(err)
	}
}

func msg(message string, threshold int) {
	hgc := color.New(color.FgYellow).SprintFunc()
	c := color.FgRed
	switch threshold {
	case -1:
		c = color.FgHiRed
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
	msg(m, 3)
}

//
// misc generic functions
//

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

// https://stackoverflow.com/questions/59579121/how-to-flatten-a-2d-slice-into-1d-slice

// flatten - turn a slice of slices into a slice
func flatten[T any](lists [][]T) []T {
	var res []T
	for _, list := range lists {
		res = append(res, list...)
	}
	return res
}

//
// https://pkg.go.dev/sort#example__sortMultiKeys
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

// https://stackoverflow.com/questions/37334119/how-to-delete-an-element-from-a-slice-in-golang

func RemoveIndex[T any](s []T, index int) []T {
	if len(s) < index {
		msg("RemoveIndex() tried to drop an out of range element", 1)
		return s
	}

	ret := make([]T, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}

//
// Geek and Latin functions
//

// stripaccents - ὀκνεῖϲ --> οκνειϲ, etc.
func stripaccents(u string) string {
	// because we don't have access to python's transtable function
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

	reducer := make(map[rune]rune)
	for f, _ := range feeder {
		for _, r := range feeder[f] {
			reducer[r] = f
		}
	}

	var stripped []rune
	for _, x := range []rune(u) {
		stripped = append(stripped, reducer[x])
	}

	s := string(stripped)
	return s
}

//func main() {
//	a := []int{1, 1, 1, 2, 3, 4, 4, 5}
//	b := 4
//	c := containsN(a, b)
//	fmt.Printf("# of %d found is %d", b, c)
//
//}
