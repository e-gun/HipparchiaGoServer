package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

func RtSearchConfirm(c echo.Context) error {
	// not going to be needed?
	// "test the activity of a poll so you don't start conjuring a bunch of key errors if you use wscheckpoll() prematurely"
	return c.String(http.StatusOK, "")
}

func RtSearchStandard(c echo.Context) error {
	// "GET /search/standard/5446b840?skg=sine%20dolore HTTP/1.1"
	// "GET /search/standard/c2fba8e8?skg=%20dolore&prx=manif HTTP/1.1"
	// "GET /search/standard/2ad866e2?prx=manif&lem=dolor HTTP/1.1"
	// "GET /search/standard/02f3610f?lem=dolor&plm=manifesta HTTP/1.1"
	fmt.Println("RtSearchStandard")
	user := readUUIDCookie(c)

	skg := c.QueryParam("skg")
	prx := c.QueryParam("prx")
	id := c.Param("id")
	lem := c.Param("lem")
	plm := c.Param("plm")

	s := builddefaultsearch(c)
	s.Seeking = skg
	s.Proximate = prx
	s.LemmaOne = lem
	s.LemmaTwo = plm
	s.IsVector = false
	sl := sessionintosearchlist(sessions[user])
	s.SearchIn = sl[0]
	s.SearchEx = sl[1]
	fmt.Println(s.SearchIn)

	// only true if not lemmatized
	s.SkgSlice = append(s.SkgSlice, s.Seeking)

	searches[id] = s

	fmt.Println(searches[id])

	prq := searchlistintoqueries(searches[id])
	fmt.Println(prq)
	con := grabpgsqlconnection()
	var hits []DbWorkline
	for _, q := range prq {
		r := worklinequery(q, con)
		hits = append(hits, r...)
	}

	for i, h := range hits {
		t := fmt.Sprintf("%d - %s : %s", i, h.FindLocus(), h.MarkedUp)
		fmt.Println(t)
	}

	return c.String(http.StatusOK, "")
}

func builddefaultsearch(c echo.Context) SearchStruct {
	var s SearchStruct
	user := readUUIDCookie(c)
	if _, exists := sessions[user]; !exists {
		sessions[user] = makedefaultsession(user)
	}

	s.User = user
	s.Limit = sessions[user].HitLimit
	s.SrchColumn = "stripped_line"
	s.SrchSyntax = "~*"
	s.OrderBy = "index"
	s.SearchIn = sessions[user].Inclusions
	s.SearchEx = sessions[user].Exclusions
	return s
}
