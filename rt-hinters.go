//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"slices"
	"strings"
)

// RtAuthorHints - /hints/author/_?term=auf --> [{'value': 'Aufidius Bassus [lt0809]'}, {'value': 'Aufustius [lt0401]'}]
func RtAuthorHints(c echo.Context) error {
	const (
		TEMPL = "%s [%s]"
	)

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

	// trim by active corpora
	user := readUUIDCookie(c)
	s := AllSessions.GetSess(user)

	var trimmed [][2]string
	for _, a := range auu {
		co := a[1][0:2]
		if s.ActiveCorp[co] {
			trimmed = append(trimmed, a)
		}
	}

	auf := make([]string, len(trimmed))
	for i := 0; i < len(trimmed); i++ {
		t := trimmed[i]
		auf[i] = fmt.Sprintf(TEMPL, t[0], t[1])
	}

	slices.Sort(auf)
	out := tojsstructslice(auf)

	return c.JSONPretty(http.StatusOK, out, JSONINDENT)
}

// RtLemmaHints - /hints/lemmata/_?term=dol --> [{"value": "dolabella\u00b9"}, {"value": "dolabra"}, {"value": "dolamen"}, ... ]
func RtLemmaHints(c echo.Context) error {
	// note that this hates "προ" and "προτ": so many come back that you will lag the system
	term := c.QueryParam("term")
	// can't slice a unicode string...
	skg := []rune(term)

	if len(skg) < 2 {
		return emptyjsreturn(c)
	}

	skg = StripaccentsRUNE(skg)
	nl := string(skg[0:2])

	var matches []string
	if _, ok := NestedLemm[nl]; ok {
		for _, l := range NestedLemm[nl] {
			er := l.EntryRune()

			// do not overshoot "er"...: "slice bounds out of range"
			lim := len(skg)
			if lim > len(er) {
				lim = len(er)
			}

			potential := StripaccentsRUNE(er[0:lim])
			if len(er) >= len(skg) && string(potential) == string(skg) {
				// need to filter ab-cedo¹ --> abcedo
				matches = append(matches, l.Entry)
			}
		}
	}

	matches = PolytonicSort(matches)
	jss := tojsstructslice(matches)

	return c.JSONPretty(http.StatusOK, jss, JSONINDENT)
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

// basichinter - which substrings of the request are members of the master map?
func basichinter(c echo.Context, mastermap map[string]bool) error {
	skg := c.QueryParam("term")
	if len(skg) < 2 {
		return emptyjsreturn(c)
	}
	skg = strings.ToLower(skg)
	skg = strings.Title(skg)

	// is what we have a match?
	var ff []string
	for f := range mastermap {
		if strings.Contains(f, skg) {
			ff = append(ff, f)
		}
	}

	// not really needed in practice?
	ff = PolytonicSort(ff)
	fs := tojsstructslice(ff)

	return c.JSONPretty(http.StatusOK, fs, JSONINDENT)
}

// tojsstructslice - []string -> []JSStruct for web output
func tojsstructslice(ss []string) []JSStruct {
	jss := make([]JSStruct, len(ss))
	for i := 0; i < len(ss); i++ {
		jss[i] = JSStruct{ss[i]}
	}
	return jss
}
