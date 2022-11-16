//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// the following routes will need authentication checks:
//	/browse/...
//	/lex/...
//	/srch/...
//  /text/...

// RtAuthLogin - accept and validate login info sent from <form id="hipparchiauserlogin"...>
func RtAuthLogin(c echo.Context) error {
	cid := readUUIDCookie(c)
	s := SafeSessionRead(cid)
	u := c.FormValue("user")
	p := c.FormValue("pw")

	MapLocker.Lock()
	if UserPassPairs[u] == p {
		SafeAuthenticationWrite(cid, true)
		s.LoginName = u
	} else {
		SafeAuthenticationWrite(cid, false)
		s.LoginName = "Anonymous"
	}
	MapLocker.Unlock()
	SafeSessionMapInsert(s)
	e := c.Redirect(http.StatusFound, "/")
	chke(e)
	return nil
}

// RtAuthLogout - log this session out
func RtAuthLogout(c echo.Context) error {
	u := readUUIDCookie(c)
	s := SafeSessionRead(u)
	s.LoginName = "Anonymous"
	SafeSessionMapInsert(s)
	SafeAuthenticationWrite(u, true)
	return c.JSONPretty(http.StatusOK, "Anonymous", JSONINDENT)
}

// RtAuthChkuser - report who this session is logged in as
func RtAuthChkuser(c echo.Context) error {
	user := readUUIDCookie(c)
	s := SafeSessionRead(user)
	a := SafeAuthenticationRead(s.ID)

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

// SafeAuthenticationRead - use a lock to safely read from AuthorizedMap
func SafeAuthenticationRead(u string) bool {
	MapLocker.RLock()
	defer MapLocker.RUnlock()
	s, e := AuthorizedMap[u]
	if e != true {
		AuthorizedMap[u] = false
		s = false
	}
	return s
}

// SafeAuthenticationWrite - use a lock to safely write to AuthorizedMap
func SafeAuthenticationWrite(u string, b bool) {
	MapLocker.RLock()
	defer MapLocker.RUnlock()
	AuthorizedMap[u] = b
	return
}
