//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package str

import (
	"fmt"
	"reflect"
	"sort"
)

type DbUnparsedWordCounts struct {
	Word  string
	Total int
	TLG   int
	LAT   int
	DDP   int
	INS   int
	CHR   int
}

func (wc DbUnparsedWordCounts) PrintOut() {
	const (
		TMPL = "%s\tt: %d\tg: %d\tl: %d\ti: %d\td: %d\tc: %d\n"
	)
	fmt.Printf(TMPL, wc.Word, wc.Total, wc.TLG, wc.LAT, wc.INS, wc.DDP, wc.CHR)
}

func (wc DbUnparsedWordCounts) SortedFVPairs(startfield int, stopfield int) []FieldValuePair {
	var pairs []FieldValuePair
	v := reflect.ValueOf(wc)
	for i := startfield; i < stopfield; i++ {
		fieldName := v.Type().Field(i).Name
		fieldValue := v.Field(i).Interface().(int) // Assuming all fields are int
		pairs = append(pairs, FieldValuePair{fieldName, fieldValue})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Value > pairs[j].Value
	})
	return pairs
}

func (wc DbUnparsedWordCounts) SortedCorpusPairs() []FieldValuePair {
	startindex := 2
	endindex := 6
	return wc.SortedFVPairs(startindex, endindex)
}

func (wc DbUnparsedWordCounts) SortedWeightedPairs(fvp []FieldValuePair) []WeightedFieldValuePair {
	maxval := float32(fvp[0].Value)
	wfvp := make([]WeightedFieldValuePair, len(fvp))
	for i, pair := range fvp {
		wfvp[i] = WeightedFieldValuePair{
			Field: pair.Field,
			Value: maxval / float32(pair.Value),
		}
	}
	return wfvp
}

func (wc DbUnparsedWordCounts) SortedWeightedCorpusPairs() []WeightedFieldValuePair {
	fvp := wc.SortedCorpusPairs()
	return wc.SortedWeightedPairs(fvp)
}
