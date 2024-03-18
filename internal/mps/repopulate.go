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
	AuGenres    = make(map[string]bool)
	WkGenres    = make(map[string]bool)
	AuLocs      = make(map[string]bool)
	WkLocs      = make(map[string]bool)
	LoadedCorp  = make(map[string]bool)
)

// RePopulateGlobalMaps - full up WkCorpusMap, AuCorpusMap, ...
func RePopulateGlobalMaps() {
	WkCorpusMap = Buildwkcorpusmap()
	AuCorpusMap = Buildaucorpusmap()
	AuGenres = Buildaugenresmap()
	WkGenres = Buildwkgenresmap()
	AuLocs = Buildaulocationmap()
	WkLocs = Buildwklocationmap()
}
