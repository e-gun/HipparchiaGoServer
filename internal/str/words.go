package str

import "sort"

type WordInfo struct {
	HeadWd     string
	HWdCount   int
	Word       string
	WdCount    int
	Loc        string
	Cit        string
	IsHomonymn bool
	Trans      string
	Wk         string
}

type VocInfo struct {
	Word         string
	C            int
	TR           string
	Strip        string
	Metr         string
	HWIsOnlyHere bool
	WdIsOnlyHere bool
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

type WeightedHeadword struct {
	Word  string
	Count int
}

type WHWList []WeightedHeadword

func (w WHWList) Len() int {
	return len(w)
}

func (w WHWList) Less(i, j int) bool {
	return w[i].Count > w[j].Count
}

func (w WHWList) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
}
