package mps

import (
	"github.com/e-gun/HipparchiaGoServer/internal/db"
)

// set these in main.go at launch time

var (
	UnparsedWeights           = map[string]float32{}
	ParsedWeightsCorpora      = map[string]float32{}
	ParsedWeightsEras         = map[string]float32{}
	ParsedWeightsGenres       = map[string]float32{}
	ParsedGreekWeightsCorpora = map[string]float32{}
	ParsedGreekWeightsEras    = map[string]float32{}
	ParsedGreekWeightsGenres  = map[string]float32{}
	ParsedLatinWeightsCorpora = map[string]float32{}
	ParsedLatinWeightsEras    = map[string]float32{}
	ParsedLatinWeightsGenres  = map[string]float32{}
)

func LoadUnparsedWordCountWeights() {
	const (
		EN1 = `__wordcounttotals`
	)
	wct := db.GetIndividualUnparsedWordCount(EN1)
	swp := wct.SortedWeightedCorpusPairs()

	// fmt.Println("LoadUnparsedWordCountWeights:", swp)
	// LoadUnparsedWordCountWeights: [{TLG 1} {LAT 9.851012} {DDP 18.133312} {INS 20.769785}]

	for _, pair := range swp {
		UnparsedWeights[pair.Field] = pair.Value
	}
}

func LoadParsedWordCountWeights() {
	const (
		EN1 = `__wordcounttotals`
		EN2 = `__greekunparsedwordcounttotalsstoredamongheadwordcounts`
		EN3 = `__latinunparsedwordcounttotalsstoredamongheadwordcounts`
	)

	// see HGB CountGenreRawWords():
	// weighting calculations should be based off of the raw word count and not parsed word counts
	// this is *not* headword data; but it is being stored in a 'headword' table because only
	// this table knows about genres and TheEras

	// __wordcounttotals
	wct := db.GetIndividualHeadwordCount(EN1)
	swp := wct.SortedWeightedCorpusPairs()
	for _, pair := range swp {
		ParsedWeightsCorpora[pair.Field] = pair.Value
	}

	swp = wct.SortedWeightedEraPairs()
	for _, pair := range swp {
		ParsedWeightsEras[pair.Field] = pair.Value
	}

	swp = wct.SortedWeightedGenrePairs()
	for _, pair := range swp {
		ParsedWeightsGenres[pair.Field] = pair.Value
	}

	// __greekunparsedwordcounttotalsstoredamongheadwordcounts

	wct = db.GetIndividualHeadwordCount(EN2)
	swp = wct.SortedWeightedCorpusPairs()
	for _, pair := range swp {
		ParsedGreekWeightsCorpora[pair.Field] = pair.Value
	}

	swp = wct.SortedWeightedEraPairs()
	for _, pair := range swp {
		ParsedGreekWeightsEras[pair.Field] = pair.Value
	}

	swp = wct.SortedWeightedGenrePairs()
	for _, pair := range swp {
		ParsedGreekWeightsGenres[pair.Field] = pair.Value
	}

	// __latinunparsedwordcounttotalsstoredamongheadwordcounts
	wct = db.GetIndividualHeadwordCount(EN3)
	swp = wct.SortedWeightedCorpusPairs()
	for _, pair := range swp {
		ParsedLatinWeightsCorpora[pair.Field] = pair.Value
	}

	swp = wct.SortedWeightedEraPairs()
	for _, pair := range swp {
		ParsedLatinWeightsEras[pair.Field] = pair.Value
	}

	swp = wct.SortedWeightedGenrePairs()
	for _, pair := range swp {
		ParsedLatinWeightsGenres[pair.Field] = pair.Value
	}
}
