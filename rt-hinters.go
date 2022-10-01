//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"sort"
	"strings"
)

func RtAuthorHints(c echo.Context) error {
	// input is not validated

	// 127.0.0.1 - - [24/Aug/2022 19:57:47] "GET /hints/author/_?term=auf HTTP/1.1" 200 -
	// [{'value': 'Aufidius Bassus [lt0809]'}, {'value': 'Aufustius [lt0401]'}]

	skg := c.QueryParam("term")
	if len(skg) < 2 {
		return emptyjsreturn(c)
	}
	skg = strings.ToLower(skg)

	// is what we have a match?
	var auu [][2]string
	for _, a := range AllAuthors {
		var who string
		var an string

		// [sosthenes], et al. can be found via "sos" or "[so"
		if strings.Contains(skg, "[") {
			who = a.Cleaname
		} else {
			who = strings.Replace(a.Cleaname, "[", "", 1)
		}

		if len(who) >= len(skg) {
			an = strings.ToLower(who[0:len(skg)])
		}
		if an == skg {
			ai := [2]string{a.Cleaname, a.UID}
			auu = append(auu, ai)
		}
	}

	// 'hyg' -->
	// [[Hyginus , myth. lt1263] [Hyginus Astronomus lt0899] [Hyginus, Gaius Iulius lt0533] [Hyginus Gromaticus lt1266]]

	// trim by active corpora
	s := sessions[readUUIDCookie(c)]
	var trimmed [][2]string
	for _, a := range auu {
		co := a[1][0:2]
		if s.ActiveCorp[co] {
			trimmed = append(trimmed, a)
		}
	}

	var auf []JSStruct
	for _, t := range trimmed {
		st := fmt.Sprintf(`%s [%s]`, t[0], t[1])
		auf = append(auf, JSStruct{st})
	}

	// sort since we were working with a map
	sort.Slice(auf, func(i, j int) bool { return auf[i].V < auf[j].V })

	return c.JSONPretty(http.StatusOK, auf, JSONINDENT)
}

func RtLemmaHints(c echo.Context) error {
	// "GET http://localhost:8000/hints/lemmata/_?term=dol"
	// curl "http://localhost:5000/hints/lemmata/_?term=dol"
	//[{"value": "dolabella\u00b9"}, {"value": "dolabra"}, {"value": "dolamen"}, {"value": "dolatorium"}, ...]

	term := c.QueryParam("term")
	// can't slice a unicode string...
	skg := []rune(term)

	if len(skg) < 2 {
		return emptyjsreturn(c)
	}

	// we will be slicing unicode, so can't use string
	skg = stripaccentsRUNE(skg)
	nl := string(skg[0:2])

	var match []JSStruct
	if _, ok := NestedLemm[nl]; ok {
		for _, l := range NestedLemm[nl] {
			er := l.EntryRune()
			potential := stripaccentsRUNE(er[0:len(skg)])
			if len(er) >= len(skg) && string(potential) == string(skg) {
				// need to filter ab-cedo¹ --> abcedo
				match = append(match, JSStruct{l.Entry})
			}
		}
	}

	sort.Slice(match, func(i, j int) bool { return match[i].V < match[j].V })
	return c.JSONPretty(http.StatusOK, match, JSONINDENT)
}

func RtAuGenreHints(c echo.Context) error {
	return basichinter(c, AuGenres)
}

func RtWkGenreHints(c echo.Context) error {
	return basichinter(c, WkGenres)
}

func RtAuLocHints(c echo.Context) error {
	return basichinter(c, AuLocs)
}

func RtWkLocHints(c echo.Context) error {
	return basichinter(c, WkLocs)
}

func basichinter(c echo.Context, mastermap map[string]bool) error {
	skg := c.QueryParam("term")
	if len(skg) < 2 {
		return emptyjsreturn(c)
	}
	skg = strings.ToLower(skg)
	skg = strings.Title(skg)

	// is what we have a match?
	var ff []string
	for f, _ := range mastermap {
		if strings.Contains(f, skg) {
			ff = append(ff, f)
		}
	}

	fs := make([]JSStruct, len(ff))
	for i, f := range ff {
		fs[i] = JSStruct{f}
	}

	return c.JSONPretty(http.StatusOK, fs, JSONINDENT)
}
