//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package str

import (
	"fmt"
	"reflect"
	"sort"
)

// hgdb=> \d unparsed_wordcounts
//                  Table "public.unparsed_wordcounts"
//   Column    |         Type          | Collation | Nullable | Default
//-------------+-----------------------+-----------+----------+---------
// entry_name  | character varying(96) |           | not null |
// total_count | integer               |           |          | 0
// gr_count    | integer               |           |          | 0
// lt_count    | integer               |           |          | 0
// dp_count    | integer               |           |          | 0
// in_count    | integer               |           |          | 0
// ch_count    | integer               |           |          | 0
//Indexes:
//    "unparsed_wordcounts_entry_name_key" UNIQUE CONSTRAINT, btree (entry_name)

// hgdb=> \d headword_wordcounts
//                        Table "public.headword_wordcounts"
//          Column          |         Type          | Collation | Nullable | Default
//--------------------------+-----------------------+-----------+----------+---------
// entry_name               | character varying(64) |           |          |
// total_count              | integer               |           |          | 0
// gr_count                 | integer               |           |          | 0
// lt_count                 | integer               |           |          | 0
// dp_count                 | integer               |           |          | 0
// in_count                 | integer               |           |          | 0
// ch_count                 | integer               |           |          | 0
// frequency_classification | character varying(64) |           |          |
// early_occurrences        | integer               |           |          | 0
// middle_occurrences       | integer               |           |          | 0
// late_occurrences         | integer               |           |          | 0
// acta                     | integer               |           |          | 0
// agric                    | integer               |           |          | 0
// alchem                   | integer               |           |          | 0
// ...

// see CALCULATEWORDWEIGHTS in HipparchiaServer's startup.py on where these really come from
// alternate chars: "🄶", "🄻", "🄸", "🄳", "🄲"; but these align awkwardly on the page

type DbHeadwordCounts struct {
	Word     string
	Total    int
	TLG      int
	LAT      int
	DDP      int
	INS      int
	CHR      int
	FrqClas  string
	Early    int
	Middle   int
	Late     int
	Acta     int
	Agric    int
	Alchem   int
	Anthol   int
	Apocal   int
	Apocry   int
	Apol     int
	Astrol   int
	Astron   int
	Biogr    int
	Bucol    int
	Caten    int
	Chron    int
	Comic    int
	Comm     int
	Concil   int
	Coq      int
	Dial     int
	Docu     int
	Doxog    int
	Eccl     int
	Eleg     int
	Encom    int
	Epic     int
	Epigr    int
	Epist    int
	Evang    int
	Exeg     int
	Fab      int
	Geog     int
	Gnom     int
	Gram     int
	Hagiog   int
	Hexam    int
	Hist     int
	Homil    int
	Hymn     int
	Hypoth   int
	Iamb     int
	Ignot    int
	Inscr    int
	Invectiv int
	Juris    int
	Lexic    int
	Liturg   int
	Lyr      int
	Magica   int
	Math     int
	Mech     int
	Med      int
	Meteor   int
	Mim      int
	Mus      int
	Myth     int
	NarrFic  int
	NatHis   int
	Onir     int
	Orac     int
	Orat     int
	Paradox  int
	Papyrus  int
	Parod    int
	Paroem   int
	Perig    int
	Phil     int
	Physiog  int
	Poem     int
	Polyhist int
	Proph    int
	Pseud    int
	Rhet     int
	Satura   int
	Satyr    int
	Schol    int
	Tact     int
	Test     int
	Theol    int
	Trag     int
	AllRhet  int
	AllRelig int
}

func (wc DbHeadwordCounts) PrintOut(w string) {
	fmt.Println("DbHeadwordCounts:", w)
	fmt.Println("AllRhet:", wc.AllRhet)
	fmt.Println("AllRelig:", wc.AllRelig)
	fmt.Println("Trag:", wc.Trag)
	fmt.Println("Epic:", wc.Epic)
	fmt.Println("Phil:", wc.Phil)
}

func (wc DbHeadwordCounts) SortedFVPairs(startfield int, stopfield int) []FieldValuePair {
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

func (wc DbHeadwordCounts) SortedEraPairs() []FieldValuePair {
	startindex := 8
	endindex := 11
	return wc.SortedFVPairs(startindex, endindex)
}

func (wc DbHeadwordCounts) SortedCorpusPairs() []FieldValuePair {
	startindex := 2
	endindex := 7
	return wc.SortedFVPairs(startindex, endindex)
}

func (wc DbHeadwordCounts) SortedGenrePairs() []FieldValuePair {
	startindex := 11
	endindex := 91
	return wc.SortedFVPairs(startindex, endindex)
}

func (wc DbHeadwordCounts) SortedWeightedPairs(fvp []FieldValuePair) []WeightedFieldValuePair {
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

func (wc DbHeadwordCounts) SortedWeightedEraPairs() []WeightedFieldValuePair {
	fvp := wc.SortedEraPairs()
	return wc.SortedWeightedPairs(fvp)
}

func (wc DbHeadwordCounts) SortedWeightedGenrePairs() []WeightedFieldValuePair {
	fvp := wc.SortedGenrePairs()
	return wc.SortedWeightedPairs(fvp)
}

func (wc DbHeadwordCounts) SortedWeightedCorpusPairs() []WeightedFieldValuePair {
	fvp := wc.SortedCorpusPairs()
	return wc.SortedWeightedPairs(fvp)
}
