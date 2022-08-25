package main

import (
	"encoding/json"
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
		return c.String(http.StatusOK, "")
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

	// send
	b, e := json.Marshal(auf)
	checkerror(e)
	return c.String(http.StatusOK, string(b))
}
