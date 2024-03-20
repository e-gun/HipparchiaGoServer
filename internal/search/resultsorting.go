//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package search

import (
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"sort"
)

//
// SORTING: https://pkg.go.dev/sort#example__sortMultiKeys
//

type WLLessFunc func(p1, p2 *str.DbWorkline) bool

// WLMultiSorter implements the Sort interface, sorting the changes within.
type WLMultiSorter struct {
	changes []str.DbWorkline
	less    []WLLessFunc
}

// Sort sorts the argument slice according to the less functions passed to WLOrderedBy.
func (ms *WLMultiSorter) Sort(changes []str.DbWorkline) {
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

// SortResults - sort the search results by the session's registerselection criterion
func SortResults(s *str.SearchStruct) {
	// Closures that order the DbWorkline structure:
	// see setsandslices.go and https://pkg.go.dev/sort#example__sortMultiKeys

	const (
		NULL = `Unavailable`
	)

	nameIncreasing := func(one, two *str.DbWorkline) bool {
		a1 := DbWlnMyAu(one).Shortname
		a2 := DbWlnMyAu(two).Shortname
		return a1 < a2
	}

	titleIncreasing := func(one, two *str.DbWorkline) bool {
		return DbWlnMyWk(one).Title < DbWlnMyWk(two).Title
	}

	dateIncreasing := func(one, two *str.DbWorkline) bool {
		d1 := DbWlnMyWk(one).RecDate
		d2 := DbWlnMyWk(two).RecDate
		if d1 != NULL && d2 != NULL {
			return DbWlnMyWk(one).ConvDate < DbWlnMyWk(two).ConvDate
		} else if d1 == NULL && d2 != NULL {
			return DbWlnMyAu(one).ConvDate < DbWlnMyAu(two).ConvDate
		} else if d1 != NULL && d2 == NULL {
			return DbWlnMyAu(one).ConvDate < DbWlnMyAu(two).ConvDate
		} else {
			return DbWlnMyAu(one).ConvDate < DbWlnMyAu(two).ConvDate
		}
	}

	increasingLines := func(one, two *str.DbWorkline) bool {
		return one.TbIndex < two.TbIndex
	}

	increasingID := func(one, two *str.DbWorkline) bool {
		return one.BuildHyperlink() < two.BuildHyperlink()
	}

	increasingWLOC := func(one, two *str.DbWorkline) bool {
		return DbWlnMyWk(one).Prov < DbWlnMyWk(two).Prov
	}

	sortby := s.StoredSession.SortHitsBy

	switch {
	case sortby == "shortname":
		WLOrderedBy(nameIncreasing, titleIncreasing, increasingLines).Sort(s.Results.Lines)
	case sortby == "converted_date":
		WLOrderedBy(dateIncreasing, nameIncreasing, titleIncreasing, increasingLines).Sort(s.Results.Lines)
	case sortby == "universalid":
		WLOrderedBy(increasingID).Sort(s.Results.Lines)
	case sortby == "provenance":
		// as this is likely an inscription search, why not sort next by date?
		WLOrderedBy(increasingWLOC, dateIncreasing).Sort(s.Results.Lines)
	default:
		// author nameIncreasing
		WLOrderedBy(nameIncreasing, increasingLines).Sort(s.Results.Lines)
	}
}
