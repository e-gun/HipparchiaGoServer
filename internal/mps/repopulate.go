package mps

import (
	"github.com/e-gun/HipparchiaGoServer/internal/str"
)

var (
	AllWorks    = make(map[string]*str.DbWork)
	AllAuthors  = make(map[string]*str.DbAuthor) // populated by authormap.go
	AllLemm     = make(map[string]*str.DbLemma)
	NestedLemm  = make(map[string]map[string]*str.DbLemma)
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
