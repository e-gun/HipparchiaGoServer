//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"time"
)

//
// SERVERSESSIONS
//

type ServerSession struct {
	ID          string
	Inclusions  SearchIncExl
	Exclusions  SearchIncExl
	ActiveCorp  map[string]bool
	VariaOK     bool   `json:"varia"`
	IncertaOK   bool   `json:"incerta"`
	SpuriaOK    bool   `json:"spuria"`
	RawInput    bool   `json:"rawinputstyle"`
	OneHit      bool   `json:"onehit"`
	HeadwordIdx bool   `json:"headwordindexing"`
	FrqIdx      bool   `json:"indexbyfrequency"`
	NearOrNot   string `json:"nearornot"`
	SearchScope string `json:"searchscope"`
	SortHitsBy  string `json:"sortorder"`
	Proximity   int    `json:"proximity"`
	BrowseCtx   int
	InputStyle  string
	HitLimit    int
	HitContext  int
	Earliest    string
	Latest      string
	TmpInt      int
	TmpStr      string
	LoginName   string
}

// SafeSessionRead - use a lock to safely read a ServerSession from the SessionMap
func SafeSessionRead(u string) ServerSession {
	SessionLocker.RLock()
	defer SessionLocker.RUnlock()
	s, e := SessionMap[u]
	if e != true {
		s = makedefaultsession(u)
	}
	return s
}

// SafeSessionMapInsert - use a lock to safely swap a ServerSession into the SessionMap
func SafeSessionMapInsert(ns ServerSession) {
	SessionLocker.Lock()
	defer SessionLocker.Unlock()
	SessionMap[ns.ID] = ns
}

// SafeSessionMapDelete - use a lock to safely delete a ServerSession from the SessionMap
func SafeSessionMapDelete(u string) {
	SessionLocker.Lock()
	defer SessionLocker.Unlock()
	delete(SessionMap, u)
}

// makedefaultsession - fill in the blanks when setting up a new session
func makedefaultsession(id string) ServerSession {
	// note that SessionMap clears every time the server restarts

	var s ServerSession
	s.ID = id
	s.ActiveCorp = Config.DefCorp
	s.VariaOK = true
	s.IncertaOK = true
	s.SpuriaOK = true
	s.NearOrNot = "near"
	s.HitLimit = DEFAULTHITLIMIT
	s.Earliest = MINDATESTR
	s.Latest = MAXDATESTR
	s.SortHitsBy = SORTBY
	s.HitContext = DEFAULTLINESOFCONTEXT
	s.BrowseCtx = Config.BrowserCtx
	s.SearchScope = DEFAULTPROXIMITYSCOPE
	s.Proximity = DEFAULTPROXIMITY
	s.LoginName = "Anonymous"

	// readUUIDCookie() called this function, and it already holds a lock
	if Config.Authenticate {
		AuthorizedMap[id] = false
	} else {
		AuthorizedMap[id] = true
	}

	//msg("makedefaultsession() in non-default state for testing; this is not a release build of HGS", 0)
	//m := make(map[string]string)
	//m["lt0917_FROM_1431_TO_2193"] = "Lucanus, Marcus Annaeus, Bellum Civile, 3"
	//m["lt0917_FROM_2_TO_692"] = "Lucanus, Marcus Annaeus, Bellum Civile, 1"
	//m["lt0917_FROM_5539_TO_6410"] = "Lucanus, Marcus Annaeus, Bellum Civile, 8"
	//m["lt0917_FROM_6411_TO_7520"] = "Lucanus, Marcus Annaeus, Bellum Civile, 9"
	//m["lt0917_FROM_4666_TO_5538"] = "Lucanus, Marcus Annaeus, Bellum Civile, 7"
	//m["lt0917_FROM_3019_TO_3835"] = "Lucanus, Marcus Annaeus, Bellum Civile, 5"
	//s.Inclusions.Passages = []string{"lt0917_FROM_6411_TO_7520", "lt0917_FROM_4666_TO_5538", "lt0917_FROM_3019_TO_3835",
	//	"lt0917_FROM_1431_TO_2193", "lt0917_FROM_2_TO_692", "lt0917_FROM_5539_TO_6410"}
	//s.Inclusions.MappedPsgByName = m
	//s.Proximity = 4
	//s.SearchScope = "words"
	// s.Inclusions.BuildPsgByName()
	return s
}

//
// ROUTING
//

// RtSessionSetsCookie - turn the session into a cookie
func RtSessionSetsCookie(c echo.Context) error {
	const (
		FAIL = "RtSessionSetsCookie() could not marshal the session"
	)
	num := c.Param("num")
	user := readUUIDCookie(c)
	s := SafeSessionRead(user)

	v, e := json.Marshal(s)
	if e != nil {
		v = []byte{}
		msg(FAIL, MSGWARN)
	}
	swap := strings.NewReplacer(`"`, "%22", ",", "%2C", " ", "%20")
	vs := swap.Replace(string(v))

	// note that cookie.Path = "/" is essential; otherwise different cookies for different contexts: "/browse" vs "/"
	cookie := new(http.Cookie)
	cookie.Name = "session" + num
	cookie.Path = "/"
	cookie.Value = vs
	cookie.Expires = time.Now().Add(4800 * time.Hour)
	c.SetCookie(cookie)

	return c.JSONPretty(http.StatusOK, "", JSONINDENT)
}

// RtSessionGetCookie - turn a stored cookie into a session
func RtSessionGetCookie(c echo.Context) error {
	// this code has input trust issues...
	const (
		FAIL1 = "RtSessionGetCookie failed to read cookie %s for %s"
		FAIL2 = "RtSessionGetCookie failed to unmarshal cookie %s for %s"
	)

	user := readUUIDCookie(c)
	num := c.Param("num")
	cookie, err := c.Cookie("session" + num)
	if err != nil {
		msg(fmt.Sprintf(FAIL1, num, user), MSGWARN)
		return c.String(http.StatusOK, "")
	}

	var s ServerSession
	// invalid character '%' looking for beginning of object key string:
	// {%22ID%22:%22723073ae-09a7-4b24-a5d6-7e20603d8c44%22%2C%22IsLoggedIn%22:true%2C...}
	swap := strings.NewReplacer("%22", `"`, "%2C", ",", "%20", " ")
	cv := swap.Replace(cookie.Value)

	err = json.Unmarshal([]byte(cv), &s)
	if err != nil {
		// invalid character '%' looking for beginning of object key string
		msg(fmt.Sprintf(FAIL2, num, user), MSGWARN)
		fmt.Println(err)
		return c.String(http.StatusOK, "")
	}

	SafeSessionMapInsert(s)

	e := c.Redirect(http.StatusFound, "/")
	chke(e)
	return nil
}

// RtResetSession - delete and then reset the session
func RtResetSession(c echo.Context) error {
	user := readUUIDCookie(c)
	SafeSessionMapDelete(user)

	// then reset it
	readUUIDCookie(c)
	e := c.Redirect(http.StatusFound, "/")
	chke(e)
	return nil
}
