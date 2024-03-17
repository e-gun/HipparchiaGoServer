package structs

import (
	"fmt"
	"slices"
)

type SearchIncExl struct {
	// the first are for internal use
	AuGenres    []string
	WkGenres    []string
	AuLocations []string
	WkLocations []string
	Authors     []string
	Works       []string
	Passages    []string // "lt0474_FROM_36136_TO_36151"
	// the next are for output to the browser
	MappedPsgByName  map[string]string // "lt0474_FROM_36136_TO_36151": "Cicero, Pro Caelio, section 1
	MappedAuthByName map[string]string
	MappedWkByName   map[string]string
	// "val.Interface().([]string)" assertion in makeselections.go means we have to insist on the slices
	ListedPBN []string
	ListedABN []string
	ListedWBN []string
}

func (i *SearchIncExl) IsEmpty() bool {
	l := len(i.AuGenres) + len(i.WkGenres) + len(i.AuLocations) + len(i.WkLocations) + len(i.Authors)
	l += len(i.Works) + len(i.Passages)
	if l > 0 {
		return false
	} else {
		return true
	}
}

func (i *SearchIncExl) CountItems() int {
	l := len(i.AuGenres) + len(i.WkGenres) + len(i.AuLocations) + len(i.WkLocations) + len(i.Authors)
	l += len(i.Works) + len(i.Passages)
	return l
}

// todo: refactor to elim circle; fnc used only once
func (i *SearchIncExl) BuildAuByName() {
	bn := make(map[string]string, len(i.MappedAuthByName))
	for _, a := range i.Authors {
		bn[a] = AllAuthors[a].Cleaname
	}
	i.MappedAuthByName = bn

	var nn []string
	for _, v := range bn {
		nn = append(nn, v)
	}

	slices.Sort(nn)
	i.ListedABN = nn
}

func (i *SearchIncExl) BuildWkByName() {
	const (
		TMPL = `%s, <i>%s</i>`
	)
	bn := make(map[string]string, len(i.MappedWkByName))
	for _, w := range i.Works {
		ws := AllWorks[w]
		au := ws.MyAu().Name
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

func (i *SearchIncExl) BuildPsgByName() {
	var nn []string
	for _, v := range i.MappedPsgByName {
		nn = append(nn, v)
	}

	slices.Sort(nn)
	i.ListedPBN = nn
}
