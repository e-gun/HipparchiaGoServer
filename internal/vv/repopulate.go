package vv

// RePopulateGlobalMaps - full up WkCorpusMap, AuCorpusMap, ...
func RePopulateGlobalMaps() {
	WkCorpusMap = Buildwkcorpusmap()
	AuCorpusMap = Buildaucorpusmap()
	AuGenres = Buildaugenresmap()
	WkGenres = Buildwkgenresmap()
	AuLocs = Buildaulocationmap()
	WkLocs = Buildwklocationmap()
}
