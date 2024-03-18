package mps

import (
	"github.com/e-gun/HipparchiaGoServer/internal/structs"
)

var (
	AllWorks    = make(map[string]*structs.DbWork)
	AllAuthors  = make(map[string]*structs.DbAuthor) // populated by authormap.go
	AllLemm     = make(map[string]*structs.DbLemma)
	NestedLemm  = make(map[string]map[string]*structs.DbLemma)
	WkCorpusMap = make(map[string][]string)
	AuCorpusMap = make(map[string][]string)
)

// RePopulateGlobalMaps - full up WkCorpusMap, AuCorpusMap, ...
func RePopulateGlobalMaps() {
	WkCorpusMap = Buildwkcorpusmap()
	AuCorpusMap = Buildaucorpusmap()
	//AuGenres = Buildaugenresmap()
	//WkGenres = Buildwkgenresmap()
	//AuLocs = Buildaulocationmap()
	//WkLocs = Buildwklocationmap()
}
