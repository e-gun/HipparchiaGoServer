package mps

import (
	"github.com/e-gun/HipparchiaGoServer/internal/db"
)

// set these in main.go at launch time

var (
	UnparsedWeights      = map[string]float32{}
	ParsedWeightsCorpora = map[string]float32{}
	ParsedWeightsEras    = map[string]float32{}
	ParsedWeightsGenres  = map[string]float32{}
)

func LoadUnparsedWordCountWeights() {
	const (
		EN1 = `__wordcounttotals`
	)
	wct := db.GetIndividualUnparsedWordCount(EN1)
	swp := wct.SortedWeightedCorpusPairs()

	// fmt.Println("LoadUnparsedWordCountWeights:", swp)
	// LoadUnparsedWordCountWeights: [{TGrk 1} {TLat 9.851012} {TDP 18.133312} {TIN 20.769785}]

	for _, pair := range swp {
		UnparsedWeights[pair.Field] = pair.Value
	}
}

func LoadParsedWordCountWeights() {
	const (
		EN1 = `__wordcounttotals`
	)
	wct := db.GetIndividualHeadwordCount(EN1)
	swp := wct.SortedWeightedCorpusPairs()
	for _, pair := range swp {
		ParsedWeightsCorpora[pair.Field] = pair.Value
	}

	swp = wct.SortedWeightedEraPairs()
	for _, pair := range swp {
		ParsedWeightsEras[pair.Field] = pair.Value
	}

	// fmt.Println("LoadParsedWordCountWeights", swp)
	// LoadParsedWordCountWeights [{Late 1} {Middle 1.375493} {Early 8.6191225}]

	swp = wct.SortedWeightedGenrePairs()
	for _, pair := range swp {
		ParsedWeightsGenres[pair.Field] = pair.Value
	}
}
