//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"time"
)

var (
	// have the option to return/generate some sort of fail message...
	emptyjsreturn = func(c echo.Context) error { return c.JSONPretty(http.StatusOK, "", JSONINDENT) }
)

//
// DEBUGGING
//

// see https://www.ditig.com/256-colors-cheat-sheet

const (
	RESET   = "\033[0m"
	BLUE1   = "\033[38;5;38m"  // DeepSkyBlue2
	BLUE2   = "\033[38;5;68m"  // SteelBlue3
	CYAN1   = "\033[38;5;109m" // LightSkyBlue3
	CYAN2   = "\033[38;5;152m" // LightCyan3
	GREEN   = "\033[38;5;108m" // DarkSeaGreen
	RED1    = "\033[38;5;160m" // Red3
	RED2    = "\033[38;5;168m" // HotPink3
	YELLOW1 = "\033[38;5;187m" // LightYellow3
	YELLOW2 = "\033[38;5;229m" // Wheat1
	GREY1   = "\033[38;5;254m" // Grey89
	GREY2   = "\033[38;5;247m" // Grey62
	GREY3   = "\033[38;5;242m" // Grey42
	WHITE   = "\033[38;5;255m" // Grey93
	PANIC   = "[%s%s v.%s%s] %sUNRECOVERABLE ERROR: PLEASE TAKE NOTE OF THE FOLLOWING PANIC MESSAGE%s\n"
)

// chke - send a generic message and panic on error
func chke(err error) {
	if err != nil {
		fmt.Printf(PANIC, YELLOW2, MYNAME, VERSION, RESET, RED2, RESET)
		panic(err)
	}
}

// msg - send a color-coded message; will not be seen unless threshold <= go log level
func msg(message string, threshold int) {
	if Config.LogLevel < threshold {
		return
	}

	var color string

	switch threshold {
	case MSGMAND:
		color = GREEN
	case MSGCRIT:
		color = RED2
	case MSGWARN:
		color = YELLOW1
	case MSGNOTE:
		color = YELLOW2
	case MSGFYI:
		color = CYAN2
	case MSGPEEK:
		color = BLUE2
	case MSGTMI:
		color = GREY3
	default:
		color = WHITE
	}

	if runtime.GOOS != "windows" && !Config.BlackAndWhite {
		fmt.Printf("[%s%s%s] %s%s%s\n", YELLOW1, SHORTNAME, RESET, color, message, RESET)
	} else {
		// terminal color codes not w's friend
		fmt.Printf("[%s] %s\n", SHORTNAME, message)
	}
}

// coloroutput - colorize output via a collection of escape substitutions; quick and dirty; not especially robust
func coloroutput(tagged string) string {
	// "[git: C4%sC0]" ==> green text for the %s
	swap := strings.NewReplacer("C1", "", "C2", "", "C3", "", "C4", "", "C5", "", "C6", "", "C0", "")

	if runtime.GOOS != "windows" && !Config.BlackAndWhite {
		swap = strings.NewReplacer("C1", YELLOW1, "C2", CYAN2, "C3", BLUE1, "C4", GREEN, "C5", RED2,
			"C6", GREY3, "C0", RESET)
	}
	tagged = swap.Replace(tagged)
	return tagged
}

// TimeTracker - report time elapsed since last checkpoint
func TimeTracker(letter string, m string, start time.Time, previous time.Time) {
	d := fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	m = fmt.Sprintf("[%s: %.3fs]", letter, time.Now().Sub(start).Seconds()) + d + m
	msg(m, TIMETRACKERMSGTHRESH)
}

// GCStats - force garbage collection and report on the results
func GCStats(fn string) {
	// NB: this could potentially backfire
	// "GC runs a garbage collection and blocks the caller until the garbage collection is complete.
	// It may also block the entire program." (https://pkg.go.dev/runtime#GC)
	// nevertheless it turns HGS into a 380MB program instead of a 540MB program
	const (
		MSG = "%s runtime.GC() %s --> %s"
		MPR = MSGPEEK
	)

	if !Config.ManualGC {
		return
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	b := fmt.Sprintf("%dM", m.HeapAlloc/1024/1024)
	runtime.GC()
	runtime.ReadMemStats(&m)
	a := fmt.Sprintf("%dM", m.HeapAlloc/1024/1024)
	msg(fmt.Sprintf(MSG, fn, b, a), MPR)
}

// stringmapprinter - print out the k/v pairs of a map
func stringkeyprinter[T any](n string, m map[string]T) {
	msg(n, MSGWARN)
	counter := 0
	for k, _ := range m {
		fmt.Printf("[%d] %s\n", counter, k)
		counter += 1
	}
}

// stringmapprinter - print out the k/v pairs of a map
func stringmapprinter[T any](n string, m map[string]T) {
	msg(n, MSGWARN)
	counter := 0
	for k, v := range m {
		fmt.Printf("[%d] %s\t", counter, k)
		fmt.Println(v)
		counter += 1
	}
}

// sliceprinter - print out the members of a slice
func sliceprinter[T any](n string, s []T) {
	msg(n, MSGWARN)
	for i, v := range s {
		fmt.Printf("[%d]\t", i)
		fmt.Println(v)
	}
}

//
// SETS AND SLICES
//

// RemoveIndex - remove item #N from a slice
func RemoveIndex[T any](s []T, index int) []T {
	// https://stackoverflow.com/questions/37334119/how-to-delete-an-element-from-a-slice-in-golang
	if len(s) == 0 || len(s) < index {
		msg("RemoveIndex() tried to drop an out of range element", MSGFYI)
		return s
	}

	ret := make([]T, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}

// Unique - return only the unique items from a slice
func Unique[T comparable](s []T) []T {
	// https://gosamples.dev/generics-remove-duplicates-slice/
	dedup := make(map[T]bool)
	var result []T
	for i := 0; i < len(s); i++ {
		if _, ok := dedup[s[i]]; !ok {
			dedup[s[i]] = true
			result = append(result, s[i])
		}
	}
	return result
}

// SetSubtraction - returns [](set(aa) - set(bb))
func SetSubtraction[T comparable](aa []T, bb []T) []T {
	//  NB this is SLOW: be careful looping it 10k times
	// 	aa := []string{"a", "b", "c", "d"}
	//	bb := []string{"a", "b", "e", "f"}
	//	dd := SetSubtraction(aa, bb)
	//	fmt.Println(dd)
	//  [c d]

	// might be faster: https://github.com/emirpasic/gods

	pruner := make(map[T]bool)
	for i := 0; i < len(bb); i++ {
		pruner[bb[i]] = true
	}

	remain := make(map[T]bool)
	for i := 0; i < len(aa); i++ {
		if _, y := pruner[aa[i]]; !y {
			remain[aa[i]] = true
		}
	}

	result := make([]T, 0, len(remain))
	for r := range remain {
		result = append(result, r)
	}
	return result
}

// IsInSlice - is item X an element of slice A?
func IsInSlice[T comparable](sl []T, seek T) bool {
	for _, v := range sl {
		if v == seek {
			return true
		}
	}
	return false
}

// ContainsN - how many Xs in slice A?
func ContainsN[T comparable](sl []T, seek T) int {
	count := 0
	for _, v := range sl {
		if v == seek {
			count += 1
		}
	}
	return count
}

// FlattenSlices - turn a slice of slices into a slice: [][]T --> []T
func FlattenSlices[T any](lists [][]T) []T {
	// https://stackoverflow.com/questions/59579121/how-to-flatten-a-2d-slice-into-1d-slice
	var res []T
	for _, list := range lists {
		res = append(res, list...)
	}
	return res
}

// StringMapIntoSlice - convert map[string]T to []T
func StringMapIntoSlice[T any](mp map[string]T) []T {
	sl := make([]T, len(mp))
	i := 0
	for _, v := range mp {
		sl[i] = v
		i += 1
	}
	return sl
}

// StringMapKeysIntoSlice - convert map[string]T to []string
func StringMapKeysIntoSlice[T any](mp map[string]T) []string {
	sl := make([]string, len(mp))
	i := 0
	for k := range mp {
		sl[i] = k
		i += 1
	}
	return sl
}

//
// SORTING: https://pkg.go.dev/sort#example__sortMultiKeys
//

type WLLessFunc func(p1, p2 *DbWorkline) bool

// WLMultiSorter implements the Sort interface, sorting the changes within.
type WLMultiSorter struct {
	changes []DbWorkline
	less    []WLLessFunc
}

// Sort sorts the argument slice according to the less functions passed to WLOrderedBy.
func (ms *WLMultiSorter) Sort(changes []DbWorkline) {
	ms.changes = changes
	sort.Sort(ms)
}

// WLOrderedBy returns a Sorter that sorts using the less functions, in order.
// Call its Sort method to sort the data.
func WLOrderedBy(less ...WLLessFunc) *WLMultiSorter {
	return &WLMultiSorter{
		less: less,
	}
}

// Len is part of sort.Interface.
func (ms *WLMultiSorter) Len() int {
	return len(ms.changes)
}

// Swap is part of sort.Interface.
func (ms *WLMultiSorter) Swap(i, j int) {
	ms.changes[i], ms.changes[j] = ms.changes[j], ms.changes[i]
}

func (ms *WLMultiSorter) Less(i, j int) bool {
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

type VILessFunc func(p1, p2 *VocInf) bool

type VIMultiSorter struct {
	changes []VocInf
	less    []VILessFunc
}

func VIOrderedBy(less ...VILessFunc) *VIMultiSorter {
	return &VIMultiSorter{
		less: less,
	}
}

func (ms *VIMultiSorter) Sort(changes []VocInf) {
	ms.changes = changes
	sort.Sort(ms)
}

// Len is part of sort.Interface.
func (ms *VIMultiSorter) Len() int {
	return len(ms.changes)
}

// Swap is part of sort.Interface.
func (ms *VIMultiSorter) Swap(i, j int) {
	ms.changes[i], ms.changes[j] = ms.changes[j], ms.changes[i]
}

func (ms *VIMultiSorter) Less(i, j int) bool {
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

// Purgechars - drop any of the chars in the bad-string from the check-string
func Purgechars(bad string, checking string) string {
	rb := []rune(bad)
	reducer := make(map[rune]bool, len(rb))
	for _, r := range rb {
		reducer[r] = true
	}

	var stripped []rune
	for _, x := range []rune(checking) {
		if _, skip := reducer[x]; !skip {
			stripped = append(stripped, x)
		}
	}
	s := string(stripped)
	return s
}
