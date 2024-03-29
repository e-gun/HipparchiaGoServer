//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"encoding/json"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"time"
)

//
// ROUTING
//

// RtSessionSetsCookie - turn the session into a cookie
func RtSessionSetsCookie(c echo.Context) error {
	const (
		FAIL = "RtSessionSetsCookie() could not marshal the session"
	)
	num := c.Param("num")
	user := vlt.ReadUUIDCookie(c)
	s := vlt.AllSessions.GetSess(user)

	v, e := json.Marshal(s)
	if e != nil {
		v = []byte{}
		Msg.WARN(FAIL)
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

	return c.JSONPretty(http.StatusOK, "", vv.JSONINDENT)
}

// RtSessionGetCookie - turn a stored cookie into a session
func RtSessionGetCookie(c echo.Context) error {
	// this code has input trust issues...
	const (
		FAIL1 = "RtSessionGetCookie failed to read cookie %s for %s"
		FAIL2 = "RtSessionGetCookie failed to unmarshal cookie %s for %s"
	)

	user := vlt.ReadUUIDCookie(c)
	num := c.Param("num")
	cookie, err := c.Cookie("session" + num)
	if err != nil {
		Msg.WARN(fmt.Sprintf(FAIL1, num, user))
		return c.String(http.StatusOK, "")
	}

	var s str.ServerSession
	// invalid character '%' looking for beginning of object key string:
	// {%22ID%22:%22723073ae-09a7-4b24-a5d6-7e20603d8c44%22%2C%22IsLoggedIn%22:true%2C...}
	swap := strings.NewReplacer("%22", `"`, "%2C", ",", "%20", " ")
	cv := swap.Replace(cookie.Value)

	err = json.Unmarshal([]byte(cv), &s)
	if err != nil {
		// invalid character '%' looking for beginning of object key string
		Msg.WARN(fmt.Sprintf(FAIL2, num, user))
		fmt.Println(err)
		return c.String(http.StatusOK, "")
	}

	vlt.AllSessions.InsertSess(s)

	e := c.Redirect(http.StatusFound, "/")
	Msg.EC(e)
	return nil
}

// RtResetSession - delete and then reset the session
func RtResetSession(c echo.Context) error {
	id := vlt.ReadUUIDCookie(c)

	vlt.AllSessions.Delete(id)

	// cancel any searches in progress: you are about to do a .CancelFnc()
	vlt.WSInfo.Reset <- id

	// [a] two-part searches are not canceled yet; and the incomplete results will be handed to the next function
	// canceling the subsequent parts happens via SSBuildQueries()
	// if !vlt.AllSessions.IsInVault(s.User) no actual queries will be loaded into the ss so the search ends instantly

	// [b] a different mechanism is used to halt a nn vector search once it starts training and the wego code has taken over
	// but the supplied context can cancel a training loop, yield empty embeddings, and then skip storage

	// [c] lda uses a similar mechanism: context inserted into nlp.LatentDirichletAllocation in the nlp code

	// reset the user ID and session
	newid := vlt.WriteUUIDCookie(c)
	vlt.AllSessions.InsertSess(vlt.MakeDefaultSession(newid))

	e := c.Redirect(http.StatusFound, "/")
	Msg.EC(e)
	return nil
}
