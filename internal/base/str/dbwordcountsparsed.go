//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package str

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
)

// hipparchiaDB=# \d wordcounts_a
//                     Table "public.wordcounts_a"
//   Column    |         Type          | Collation | Nullable | Default
//-------------+-----------------------+-----------+----------+---------
// entry_name  | character varying(64) |           |          |
// total_count | integer               |           |          | 0
// gr_count    | integer               |           |          | 0
// lt_count    | integer               |           |          | 0
// dp_count    | integer               |           |          | 0
// in_count    | integer               |           |          | 0
// ch_count    | integer               |           |          | 0
//Indexes:
//    "wcindex_a" UNIQUE, btree (entry_name)

// hipparchiaDB=# \d dictionary_headword_wordcounts
//                   Table "public.dictionary_headword_wordcounts"
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
// ...

// see CALCULATEWORDWEIGHTS in HipparchiaServer's startup.py on where these really come from
// alternate chars: "🄶", "🄻", "🄸", "🄳", "🄲"; but these align awkwardly on the page

// these values are valid for 1.4.0- but not 2.0.0+

var (
	CORPUSWEIGTING = map[string]float32{"Ⓖ": 1.0, "Ⓛ": 12.7, "Ⓘ": 15.19, "Ⓓ": 18.14, "Ⓒ": 85.78}
	ERAWEIGHTING   = map[string]float32{"e": 6.93, "m": 1.87, "l": 1}
	GKGENREWEIGHT  = map[string]float32{"Acta": 85.38, "Alchem": 72.13, "Anthol": 17.68, "Apocal": 117.69, "Apocr": 89.77,
		"Apol": 7.0, "Astrol": 20.68, "Astron": 44.72, "Biogr": 6.39, "Bucol": 416.66, "Caten": 5.21,
		"Chron": 4.55, "Comic": 29.61, "Comm": 1.0, "Concil": 16.75, "Coq": 532.74, "Dial": 7.1,
		"Docu": 2.66, "Doxogr": 130.84, "Eccl": 7.57, "Eleg": 188.08, "Encom": 13.17, "Epic": 19.36,
		"Epigr": 10.87, "Epist": 4.7, "Evang": 118.66, "Exeg": 1.24, "Fab": 140.87,
		"Geog": 10.74, "Gnom": 88.54, "Gram": 8.65, "Hagiog": 22.83, "Hexam": 110.78,
		"Hist": 1.44, "Homil": 6.87, "Hymn": 48.18, "Hypoth": 12.95, "Iamb": 122.22,
		"Ignot": 122914.2, "Invect": 238.54, "Inscr": 1.91, "Juris": 51.42, "Lexic": 4.14,
		"Litur": 531.5, "Lyr": 213.43, "Magica": 85.38, "Math": 9.91, "Mech": 103.44, "Med": 2.25,
		"Metro": 276.78, "Mim": 2183.94, "Mus": 96.32, "Myth": 201.78, "NarrFic": 14.62,
		"NatHis": 9.67, "Onir": 145.15, "Orac": 240.47, "Orat": 6.67, "Paradox": 267.32,
		"Parod": 831.51, "Paroem": 65.58, "Perig": 220.38, "Phil": 3.69, "Physiog": 628.77,
		"Poem": 62.82, "Polyhist": 24.91, "Proph": 95.51, "Pseud": 611.65, "Rhet": 8.67,
		"Satura": 291.58, "Satyr": 96.78, "Schol": 5.56, "Tact": 52.01, "Test": 66.53, "Theol": 6.28,
		"Trag": 35.8, "AllRelig": 0.58, "AllRhet": 2.9}
	LATGENREWEIGHT = map[string]float32{"Agric": 5.27, "Astron": 17.15, "Biogr": 9.88, "Bucol": 40.39, "Bomic": 4.21, "Comm": 2.25,
		"Coq": 60.0, "Dial": 1134.73, "Docu": 6.19, "Eleg": 8.35, "Encom": 404.6, "Epic": 2.37,
		"Epigr": 669.3, "Epist": 2.06, "Fab": 25.4, "Gnom": 147.23, "Gramm": 5.74, "Hexam": 20.06,
		"Hist": 1.0, "Hypoth": 762.59, "Ignotum": 586.58, "Inscr": 1.29, "Juris": 1.11,
		"Lexic": 27.71, "Lyr": 24.76, "Med": 7.26, "Mim": 1045.69, "NarrFic": 11.7,
		"Nathist": 1.94, "Orat": 1.81, "Parod": 339.23, "Phil": 2.3, "Poem": 14.34,
		"Polyhist": 4.75, "Rhet": 2.71, "Satura": 23.0, "Tact": 37.6, "Trag": 13.29, "Allrelig": 0,
		"Allrhet": 1.08}
	lookforlatinchars = regexp.MustCompile(`[a-z]`)
)

type DbHeadwordCounts struct {
	Word     string
	Total    int
	TGrk     int
	TLat     int
	TDP      int
	TIN      int
	TCh      int
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

// HWData - to help sort values inside DbHeadwordCount
//type HWData struct {
//	Name  string
//	Count int
//}
//
//func (hw *DbHeadwordCounts) GetCorpVals() []HWData {
//	// Prevalence (all forms): Ⓖ 95,843 / Ⓛ 10 / Ⓘ 151 / Ⓓ 751 / Ⓒ 64 / Ⓣ 96,819
//	var vv []HWData
//	vv = append(vv, HWData{"Ⓖ", hw.TGrk})
//	vv = append(vv, HWData{"Ⓛ", hw.TLat})
//	vv = append(vv, HWData{"Ⓘ", hw.TIN})
//	vv = append(vv, HWData{"Ⓓ", hw.TDP})
//	vv = append(vv, HWData{"Ⓒ", hw.TCh})
//	return vv
//}
//
//func (hw *DbHeadwordCounts) GetWeightedCorpVals() []HWData {
//	vv := hw.GetCorpVals()
//	for i, v := range vv {
//		vv[i].Count = int(float32(v.Count) * CORPUSWEIGTING[v.Name])
//	}
//	return vv
//}
//
//func (hw *DbHeadwordCounts) GetSortedCorpVals() []HWData {
//	vv := hw.GetCorpVals()
//	sort.Slice(vv, func(i, j int) bool {
//		return vv[i].Count > vv[j].Count
//	})
//	return vv
//}
//
//func (hw *DbHeadwordCounts) GetWeightedSortedCorpVals() []HWData {
//	vv := hw.GetWeightedCorpVals()
//	sort.Slice(vv, func(i, j int) bool {
//		return vv[i].Count > vv[j].Count
//	})
//	return vv
//}
//
//func (hw *DbHeadwordCounts) GetSortedTrimmedCorpVals() []HWData {
//	vv := hw.GetSortedCorpVals()
//	var trimmed []HWData
//	for _, v := range vv {
//		if v.Count > 0 {
//			trimmed = append(trimmed, v)
//		}
//	}
//	return trimmed
//}
//
//func (hw *DbHeadwordCounts) GetWeightedSortedTrimmedCorpVals() []HWData {
//	vv := hw.GetWeightedSortedCorpVals()
//	var trimmed []HWData
//	for _, v := range vv {
//		if v.Count > 0 {
//			trimmed = append(trimmed, v)
//		}
//	}
//	return trimmed
//}
//
//func (hw *DbHeadwordCounts) GetTimeVals() []HWData {
//	// Weighted chronological distribution: ℯ 100 / ℓ 84 / 𝓂 62
//	var vv []HWData
//	vv = append(vv, HWData{"e", hw.Early})
//	vv = append(vv, HWData{"l", hw.Late})
//	vv = append(vv, HWData{"m", hw.Middle})
//	return vv
//}
//
//func (hw *DbHeadwordCounts) GetWeightedTimeVals() []HWData {
//	vv := hw.GetTimeVals()
//	for i, v := range vv {
//		vv[i].Count = int(float32(v.Count) * ERAWEIGHTING[v.Name])
//	}
//	return vv
//}
//
//func (hw *DbHeadwordCounts) GetSortedTimeVals() []HWData {
//	vv := hw.GetTimeVals()
//	sort.Slice(vv, func(i, j int) bool {
//		return vv[i].Count > vv[j].Count
//	})
//	return vv
//}
//
//func (hw *DbHeadwordCounts) GetWeightedSortedTimeVals() []HWData {
//	vv := hw.GetWeightedTimeVals()
//	sort.Slice(vv, func(i, j int) bool {
//		return vv[i].Count > vv[j].Count
//	})
//	return vv
//}
//
//func (hw *DbHeadwordCounts) GetSortedTrimmedTimeVals() []HWData {
//	vv := hw.GetSortedTimeVals()
//	var trimmed []HWData
//	for _, v := range vv {
//		if v.Count > 0 {
//			trimmed = append(trimmed, v)
//		}
//	}
//	return trimmed
//}
//
//func (hw *DbHeadwordCounts) GetWeightedSortedTrimmedTimeVals() []HWData {
//	vv := hw.GetWeightedSortedTimeVals()
//	var trimmed []HWData
//	for _, v := range vv {
//		if v.Count > 0 {
//			trimmed = append(trimmed, v)
//		}
//	}
//	return trimmed
//}
//
//func (hw DbHeadwordCounts) GetSortedGenreVals() []HWData {
//	const (
//		INDEXOFFIRSTGENRE = 11
//		INDEXOFLASTGENRE  = 88
//	)
//
//	// don't use reflect on a pointer: *DbHeadwordCounts
//	val := reflect.ValueOf(hw)
//	typ := reflect.TypeOf(hw)
//
//	type fieldinfo struct {
//		Name  string
//		Value int
//	}
//
//	var fields []fieldinfo
//	for i := INDEXOFFIRSTGENRE; i < INDEXOFLASTGENRE+1; i++ {
//		fields = append(fields, fieldinfo{
//			Name:  typ.Field(i).Name,
//			Value: int(val.Field(i).Int()),
//		})
//	}
//
//	// Sort by value descending
//	sort.Slice(fields, func(i, j int) bool {
//		return fields[i].Value > fields[j].Value
//	})
//
//	var newhwdata []HWData
//	for _, field := range fields {
//		newhwdata = append(newhwdata, HWData{
//			Name:  field.Name,
//			Count: field.Value,
//		})
//	}
//	return newhwdata
//}
//
//func (hw *DbHeadwordCounts) GetWeightedSortedGenreVals() []HWData {
//	if hw.IsGreek() {
//		return hw.GetGreekWeightedSortedGenreVals()
//	} else {
//		return hw.GetLatinWeightedSortedGenreVals()
//	}
//}
//
//func (hw *DbHeadwordCounts) GetGreekWeightedSortedGenreVals() []HWData {
//	const (
//		MINORGKGENRETHRESHOLD = 250
//	)
//	return hw.GenerateWeightedSortedGenreVals(GKGENREWEIGHT, MINORGKGENRETHRESHOLD)
//}
//
//func (hw *DbHeadwordCounts) GetLatinWeightedSortedGenreVals() []HWData {
//	const (
//		MINORGKGENRETHRESHOLD = 250
//	)
//	return hw.GenerateWeightedSortedGenreVals(LATGENREWEIGHT, MINORGKGENRETHRESHOLD)
//}
//
//// GenerateWeightedSortedGenreVals - this is the one that does the real work of ranking and sorting
//func (hw *DbHeadwordCounts) GenerateWeightedSortedGenreVals(weightmap map[string]float32, threshold float32) []HWData {
//
//	vv := hw.GetSortedGenreVals()
//	for i, hwc := range vv {
//		gwt := weightmap[hwc.Name]
//		if gwt > threshold {
//			// you will never get weights for these genres; but you will still be able to peek when testing
//			gwt = -1 * gwt
//		}
//		vv[i] = HWData{
//			Name:  hwc.Name,
//			Count: int(float32(hwc.Count) * gwt),
//		}
//	}
//
//	// the weighting just unsorted things, so...
//	sort.Slice(vv, func(i, j int) bool {
//		return vv[i].Count > vv[j].Count
//	})
//
//	// trim empties
//	var trimmed []HWData
//	for _, v := range vv {
//		if v.Count > 0 {
//			trimmed = append(trimmed, v)
//		}
//	}
//
//	if len(trimmed) == 0 {
//		return trimmed
//	}
//
//	// now make the top value into "100" and weigh the rest relative to it
//	top := trimmed[0].Count
//	for i := 0; i < len(trimmed); i++ {
//		trimmed[i].Count = int(100 * (float32(trimmed[i].Count) / float32(top)))
//	}
//
//	return trimmed
//}
//
//func (hw *DbHeadwordCounts) IsGreek() bool {
//	if lookforlatinchars.MatchString(hw.Word) {
//		return false
//	} else {
//		return true
//	}
//}

// new...

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
