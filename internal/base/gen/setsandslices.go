//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package gen

import (
	"fmt"
	"slices"
)

//
// SETS AND SLICES
//

// RemoveIndex - remove item #N from a slice
func RemoveIndex[T any](s []T, index int) []T {
	if len(s) == 0 || len(s) < index {
		// messaging.go imports gen, so you can't 'mm' this
		fmt.Println("RemoveIndex() tried to drop an out of range element")
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
