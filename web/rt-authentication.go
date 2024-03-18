//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"github.com/e-gun/HipparchiaGoServer/internal/generic"
	"github.com/e-gun/HipparchiaGoServer/internal/vaults"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/labstack/echo/v4"
	"net/http"
)

// RtAuthLogin - accept and validate login info sent from <form id="hipparchiauserlogin"...>
func RtAuthLogin(c echo.Context) error {
	cid := vaults.ReadUUIDCookie(c)
	s := vaults.AllSessions.GetSess(cid)
	u := c.FormValue("user")
	p := c.FormValue("pw")

	if vaults.UserPassPairs[u] == p {
		vaults.AllAuthorized.Register(cid, true)
		s.LoginName = u
	} else {
		vaults.AllAuthorized.Register(cid, false)
		s.LoginName = "Anonymous"
	}

	vaults.AllSessions.InsertSess(s)
	e := c.Redirect(http.StatusFound, "/")
	msg.EC(e)
	return nil
}

// RtAuthLogout - log this session out
func RtAuthLogout(c echo.Context) error {
	u := vaults.ReadUUIDCookie(c)
	s := vaults.AllSessions.GetSess(u)
	s.LoginName = "Anonymous"
	vaults.AllSessions.InsertSess(s)
	vaults.AllAuthorized.Register(u, false)
	return c.JSONPretty(http.StatusOK, "Anonymous", vv.JSONINDENT)
}

// RtAuthChkuser - report who this session is logged in as
func RtAuthChkuser(c echo.Context) error {
	user := vaults.ReadUUIDCookie(c)
	s := vaults.AllSessions.GetSess(user)
	a := vaults.AllAuthorized.Check(s.ID)

	type JSO struct {
		ID   string `json:"userid"`
		Auth bool   `json:"authorized"`
	}

	o := JSO{
		ID:   s.LoginName,
		Auth: a,
	}
	return generic.JSONresponse(c, o)
}
