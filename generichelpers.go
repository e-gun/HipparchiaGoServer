//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"slices"
	"sort"
)

//
// DEBUGGING
//

// stringkeyprinter - print out the keys of a map
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

// showinterimresults - print out the current results
func showinterimresults(s *SearchStruct) {
	const (
		NB  = "showinterimresults()"
		FMT = "[%d] %s\t%s\t%s"
	)

	msg(NB, MSGWARN)

	mp := make(map[string]DbWorkline, s.Results.Len())
	kk := make([]string, s.Results.Len())

	for i := 0; i < s.Results.Len(); i++ {
		r := s.Results.Lines[i]
		mp[r.BuildHyperlink()] = r
		kk[i] = r.BuildHyperlink()
	}

	slices.Sort(kk)

	for i, k := range kk {
		r := mp[k]
		v := fmt.Sprintf(FMT, i, r.BuildHyperlink(), s.Seeking, r.MarkedUp)
		msg(v, MSGNOTE)
	}
}

//
// SETS AND SLICES
//

// RemoveIndex - remove item #N from a slice
func RemoveIndex[T any](s []T, index int) []T {
	if len(s) == 0 || len(s) < index {
		msg("RemoveIndex() tried to drop an out of range element", MSGFYI)
		return s
	}
	return slices.Delete(s, index, index+1)
}

// ToSet - returns a blank map of a slice
func ToSet[T comparable](sl []T) map[T]struct{} {
	m := make(map[T]struct{})
	for i := 0; i < len(sl); i++ {
		m[sl[i]] = struct{}{}
	}
	return m
}

// Unique - return only the unique items from a slice
func Unique[T comparable](s []T) []T {
	// can't use slices.Compact because that only looks as consecutive repeats: [a, a, b, a] -> [a, b, a]

	set := ToSet(s)

	var result []T
	for k := range set {
		result = append(result, k)
	}

	return result
}

func SetSubtraction[T comparable](aa []T, bb []T) []T {
	//  NB this is likely SLOW: be careful looping it 10k times
	// 	aa := []string{"a", "b", "c", "d", "g", "h"}
	//	bb := []string{"a", "b", "e", "f", "g"}
	//	dd := SetSubtraction(aa, bb)
	//  [c d h]

	// this makes more sense in some other context where bb is big and amorphous...
	bb = Unique(bb)

	aa = slices.DeleteFunc(aa, func(c T) bool {
		return slices.Contains(bb, c)
	})

	return aa
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

// ChunkSlice - turn a slice into a slice of slices of size N; thanks to https://stackoverflow.com/questions/35179656/slice-chunking-in-go
func ChunkSlice[T any](items []T, size int) (chunks [][]T) {
	for size < len(items) {
		items, chunks = items[size:], append(chunks, items[0:size:size])
	}
	return append(chunks, items)
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

type VILessFunc func(p1, p2 *VocInfo) bool

type VIMultiSorter struct {
	changes []VocInfo
	less    []VILessFunc
}

func VIOrderedBy(less ...VILessFunc) *VIMultiSorter {
	return &VIMultiSorter{
		less: less,
	}
}

func (ms *VIMultiSorter) Sort(changes []VocInfo) {
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
