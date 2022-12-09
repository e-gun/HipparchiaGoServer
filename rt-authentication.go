//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// RtAuthLogin - accept and validate login info sent from <form id="hipparchiauserlogin"...>
func RtAuthLogin(c echo.Context) error {
	cid := readUUIDCookie(c)
	s := FetchSession(cid)
	u := c.FormValue("user")
	p := c.FormValue("pw")

	if UserPassPairs[u] == p {
		AuthenticationWrite(cid, true)
		s.LoginName = u
	} else {
		AuthenticationWrite(cid, false)
		s.LoginName = "Anonymous"
	}

	SessionInsert(s)
	e := c.Redirect(http.StatusFound, "/")
	chke(e)
	return nil
}

// RtAuthLogout - log this session out
func RtAuthLogout(c echo.Context) error {
	u := readUUIDCookie(c)
	s := FetchSession(u)
	s.LoginName = "Anonymous"
	SessionInsert(s)
	AuthenticationWrite(u, false)
	return c.JSONPretty(http.StatusOK, "Anonymous", JSONINDENT)
}

// RtAuthChkuser - report who this session is logged in as
func RtAuthChkuser(c echo.Context) error {
	user := readUUIDCookie(c)
	s := FetchSession(user)
	a := AuthenticationCheck(s.ID)

	type JSO struct {
		ID   string `json:"userid"`
		Auth bool   `json:"authorized"`
	}

	o := JSO{
		ID:   s.LoginName,
		Auth: a,
	}

	return c.JSONPretty(http.StatusOK, o, JSONINDENT)
}

// AuthenticationCheck - use a lock to safely read from AuthorizedMap; "true" if you have access
func AuthenticationCheck(u string) bool {
	if !Config.Authenticate {
		return true
	}

	MapLocker.RLock()
	defer MapLocker.RUnlock()
	s, e := AuthorizedMap[u]
	if e != true {
		AuthorizedMap[u] = false
		s = false
	}
	return s
}

// AuthenticationWrite - use a lock to safely write to AuthorizedMap
func AuthenticationWrite(u string, b bool) {
	MapLocker.RLock()
	defer MapLocker.RUnlock()
	AuthorizedMap[u] = b
	return
}
