//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package search

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/mps"
	"slices"
)

// BuildSelectionOverview will call the relevant SearchIncExl functions: see buildesearchlist.go
func BuildSelectionOverview(s *str.ServerSession) {
	BuildAuByName(&s.Inclusions)
	BuildAuByName(&s.Exclusions)
	BuildWkByName(&s.Inclusions)
	BuildWkByName(&s.Exclusions)
	s.Inclusions.BuildPsgByName()
	s.Exclusions.BuildPsgByName()
}

func BuildAuByName(i *str.SearchIncExl) {
	bn := make(map[string]string, len(i.MappedAuthByName))
	for _, a := range i.Authors {
		bn[a] = mps.AllAuthors[a].Cleaname
	}
	i.MappedAuthByName = bn

	var nn []string
	for _, v := range bn {
		nn = append(nn, v)
	}

	slices.Sort(nn)
	i.ListedABN = nn
}

func BuildWkByName(i *str.SearchIncExl) {
	const (
		TMPL = `%s, <i>%s</i>`
	)
	bn := make(map[string]string, len(i.MappedWkByName))
	for _, w := range i.Works {
		ws := mps.AllWorks[w]
		au := mps.DbWkMyAu(ws).Name
		ti := ws.Title
		bn[w] = fmt.Sprintf(TMPL, au, ti)
	}
	i.MappedWkByName = bn

	var nn []string
	for _, v := range bn {
		nn = append(nn, v)
	}

	slices.Sort(nn)
	i.ListedWBN = nn
}
