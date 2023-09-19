//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"sync"
)

// RtAuthLogin - accept and validate login info sent from <form id="hipparchiauserlogin"...>
func RtAuthLogin(c echo.Context) error {
	cid := readUUIDCookie(c)
	s := AllSessions.GetSess(cid)
	u := c.FormValue("user")
	p := c.FormValue("pw")

	if UserPassPairs[u] == p {
		AllAuthorized.Register(cid, true)
		s.LoginName = u
	} else {
		AllAuthorized.Register(cid, false)
		s.LoginName = "Anonymous"
	}

	AllSessions.InsertSess(s)
	e := c.Redirect(http.StatusFound, "/")
	chke(e)
	return nil
}

// RtAuthLogout - log this session out
func RtAuthLogout(c echo.Context) error {
	u := readUUIDCookie(c)
	s := AllSessions.GetSess(u)
	s.LoginName = "Anonymous"
	AllSessions.InsertSess(s)
	AllAuthorized.Register(u, false)
	return c.JSONPretty(http.StatusOK, "Anonymous", JSONINDENT)
}

// RtAuthChkuser - report who this session is logged in as
func RtAuthChkuser(c echo.Context) error {
	user := readUUIDCookie(c)
	s := AllSessions.GetSess(user)
	a := AllAuthorized.Check(s.ID)

	type JSO struct {
		ID   string `json:"userid"`
		Auth bool   `json:"authorized"`
	}

	o := JSO{
		ID:   s.LoginName,
		Auth: a,
	}
	return JSONresponse(c, o)
}

//
// THREAD SAFE INFRASTRUCTURE: MUTEX
//

// MakeAuthorizedVault - called only once; yields the AllAuthorized vault
func MakeAuthorizedVault() AuthVault {
	return AuthVault{
		UserMap: make(map[string]bool),
		mutex:   sync.RWMutex{},
	}
}

// AuthVault - there should be only one of these; and it contains all the authorization info
type AuthVault struct {
	UserMap map[string]bool
	mutex   sync.RWMutex
}

func (av *AuthVault) Check(u string) bool {
	if !Config.Authenticate {
		return true
	}
	av.mutex.Lock()
	defer av.mutex.Unlock()
	s, e := av.UserMap[u]
	if e != true {
		av.UserMap[u] = false
		s = false
	}
	return s
}

func (av *AuthVault) Register(u string, b bool) {
	av.mutex.Lock()
	defer av.mutex.Unlock()
	av.UserMap[u] = b
	return
}
