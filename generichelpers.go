//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/labstack/echo/v4"
	"net/http"
	"runtime"
	"sort"
	"time"
)

var (
	// have the option to return/generate some sort of fail message...
	emptyjsreturn = func(c echo.Context) error { return c.JSONPretty(http.StatusOK, "", JSONINDENT) }
)

//
// DEBUGGING
//

// chke - send a generic message and panic on error
func chke(err error) {
	if err != nil {
		red := color.New(color.FgHiRed).PrintfFunc()
		red("[%s v.%s] UNRECOVERABLE ERROR: PLEASE TAKE NOTE OF THE FOLLOWING PANIC MESSAGE\n", MYNAME, VERSION)
		panic(err)
	}
}

// msg - send a color-coded message; will not be seen unless threshold <= go log level
func msg(message string, threshold int) {
	if Config.LogLevel < threshold {
		return
	}

	hgc := color.New(color.FgYellow).SprintFunc()
	c := color.FgRed
	switch threshold {
	case -1:
		// c = color.FgHiRed
		c = color.FgCyan
	case 0:
		c = color.FgRed
	case 1:
		c = color.FgMagenta
	case 2:
		c = color.FgYellow
	case 3:
		c = color.FgWhite
	case 4:
		c = color.FgHiBlue
	case 5:
		c = color.FgHiBlack
	default:
		c = color.FgWhite
	}
	mc := color.New(c).SprintFunc()

	switch runtime.GOOS {
	case "windows":
		// terminal color codes not w's friend
		fmt.Printf("[%s] %s\n", SHORTNAME, message)
	default:
		fmt.Printf("[%s] %s\n", hgc(SHORTNAME), mc(message))
	}

}

func timetracker(letter string, m string, start time.Time, previous time.Time) {
	d := fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	m = fmt.Sprintf("[%s: %.3fs]", letter, time.Now().Sub(start).Seconds()) + d + m
	msg(m, TIMETRACKERMSGTHRESH)
}

// gcstats - force garbage collection and report on the results
func gcstats(fn string) {
	// NB: this could potentially backfire
	// "GC runs a garbage collection and blocks the caller until the garbage collection is complete.
	// It may also block the entire program." (https://pkg.go.dev/runtime#GC)
	// nevertheless it turns HGS into a 380MB program instead of a 540MB program
	const (
		MSG = "%s runtime.GC() %s --> %s"
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
	msg(fmt.Sprintf(MSG, fn, b, a), 4)
}

//
// SETS AND SLICES
//

// RemoveIndex - remove item #N from a slice
func RemoveIndex[T any](s []T, index int) []T {
	// https://stackoverflow.com/questions/37334119/how-to-delete-an-element-from-a-slice-in-golang
	if len(s) == 0 || len(s) < index {
		msg("RemoveIndex() tried to drop an out of range element", 3)
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
	//  NB this is SLOW: be careful looping it 10k times
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

// isinslice - is item X an element of slice A?
func isinslice[T comparable](sl []T, seek T) bool {
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

// flatten - turn a slice of slices into a slice: [][]T --> []T
func flatten[T any](lists [][]T) []T {
	// https://stackoverflow.com/questions/59579121/how-to-flatten-a-2d-slice-into-1d-slice
	var res []T
	for _, list := range lists {
		res = append(res, list...)
	}
	return res
}

// stringmapintoslice - convert map[string]T to []T
func stringmapintoslice[T any](mp map[string]T) []T {
	sl := make([]T, len(mp))
	i := 0
	for _, v := range mp {
		sl[i] = v
		i += 1
	}
	return sl
}

// stringmapkeysintoslice - convert map[string]T to []string
func stringmapkeysintoslice[T any](mp map[string]T) []string {
	sl := make([]string, len(mp))
	i := 0
	for k, _ := range mp {
		sl[i] = k
		i += 1
	}
	return sl
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
		}
	}
	s := string(stripped)
	return s
}
