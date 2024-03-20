//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package mps

import (
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
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
