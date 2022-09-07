package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

func RtIndexMaker(c echo.Context) error {
	// diverging from the way the python works
	// build not via the selection boxes but via the actual selection made and stored in the session

	// user := readUUIDCookie(c)
	srch := sessionintobulksearch(c)
	searches[srch.ID] = srch

	type JSFeeder struct {
		Au string `json:"authorname"`
		Ti string `json:"title"`
		WS string `json:"worksegment"`
		HT string `json:"texthtml"`
		EL string `json:"elapsed"`
		WF string `json:"wordsfound"`
		KY string `json:"keytoworks"`
	}
	var jso JSFeeder
	js, e := json.Marshal(jso)
	chke(e)

	return c.String(http.StatusOK, string(js))
}

func RtTextMaker(c echo.Context) error {
	// diverging from the way the python works
	// build not via the selection boxes but via the actual selection made and stored in the session

	// this has the downside of allowing for insanely large text generation
	// but, on the other hand, this now works like a simple search

	// then it gets output as a big browser table...

	user := readUUIDCookie(c)
	srch := sessionintobulksearch(c)
	searches[srch.ID] = srch

	// now we have the lines we need....
	firstline := searches[srch.ID].Results[0]
	firstwork := AllWorks[firstline.WkUID]
	firstauth := AllAuthors[firstwork.FindAuthor()]

	// ci := formatcitationinfo(firstwork, firstline)
	tr := buildbrowsertable(-1, searches[srch.ID].Results)

	type JSFeeder struct {
		Au string `json:"authorname"`
		Ti string `json:"title"`
		St string `json:"structure"`
		WS string `json:"worksegment"`
		HT string `json:"texthtml"`
	}

	sui := sessions[user].Inclusions
	var jso JSFeeder

	jso.Au = firstauth.Shortname

	if len(sui.Authors) > 1 || len(sui.AuGenres) > 0 || len(sui.AuLocations) > 0 {
		jso.Au += " (and others)"
	}

	jso.Ti = firstwork.Title
	if len(sui.Works) > 1 || len(sui.WkGenres) > 0 || len(sui.WkLocations) > 0 {
		jso.Ti += " (and others)"
	}

	jso.St = basiccitation(firstwork, firstline)
	jso.WS = "" // unused for now
	jso.HT = tr

	js, e := json.Marshal(jso)
	chke(e)

	return c.String(http.StatusOK, string(js))
}

func sessionintobulksearch(c echo.Context) SearchStruct {
	user := readUUIDCookie(c)

	srch := builddefaultsearch(c)
	srch.Seeking = ""
	srch.Limit = MAXTEXTLINEGENERATION
	srch.InitSum = "(gathering and formatting line of text)"
	srch.ID = strings.Replace(uuid.New().String(), "-", "", -1)

	parsesearchinput(&srch)
	sl := sessionintosearchlist(sessions[user])
	srch.SearchIn = sl.Inc
	srch.SearchEx = sl.Excl
	srch.SearchSize = sl.Size
	prq := searchlistintoqueries(&srch)
	srch.Queries = prq
	srch.IsActive = true
	searches[srch.ID] = srch
	searches[srch.ID] = HGoSrch(searches[srch.ID])

	return searches[srch.ID]
}
