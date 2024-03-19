//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"github.com/e-gun/HipparchiaGoServer/internal/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/labstack/echo/v4"
	"net/http"
)

// RtAuthLogin - accept and validate login info sent from <form id="hipparchiauserlogin"...>
func RtAuthLogin(c echo.Context) error {
	cid := vlt.ReadUUIDCookie(c)
	s := vlt.AllSessions.GetSess(cid)
	u := c.FormValue("user")
	p := c.FormValue("pw")

	if vlt.UserPassPairs[u] == p {
		vlt.AllAuthorized.Register(cid, true)
		s.LoginName = u
	} else {
		vlt.AllAuthorized.Register(cid, false)
		s.LoginName = "Anonymous"
	}

	vlt.AllSessions.InsertSess(s)
	e := c.Redirect(http.StatusFound, "/")
	Msg.EC(e)
	return nil
}

// RtAuthLogout - log this session out
func RtAuthLogout(c echo.Context) error {
	u := vlt.ReadUUIDCookie(c)
	s := vlt.AllSessions.GetSess(u)
	s.LoginName = "Anonymous"
	vlt.AllSessions.InsertSess(s)
	vlt.AllAuthorized.Register(u, false)
	return c.JSONPretty(http.StatusOK, "Anonymous", vv.JSONINDENT)
}

// RtAuthChkuser - report who this session is logged in as
func RtAuthChkuser(c echo.Context) error {
	user := vlt.ReadUUIDCookie(c)
	s := vlt.AllSessions.GetSess(user)
	a := vlt.AllAuthorized.Check(s.ID)

	type JSO struct {
		ID   string `json:"userid"`
		Auth bool   `json:"authorized"`
	}

	o := JSO{
		ID:   s.LoginName,
		Auth: a,
	}
	return gen.JSONresponse(c, o)
}
