//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"runtime"
	"text/template"
	"time"
)

//
// ROUTING
//

// RtFrontpage - send the html for "/"
func RtFrontpage(c echo.Context) error {
	// will set if missing
	user := readUUIDCookie(c)
	s := SafeSessionRead(user)

	env := fmt.Sprintf("%s: %s - %s (%d workers)", runtime.Version(), runtime.GOOS, runtime.GOARCH, Config.WorkerCount)

	ahtm := AUTHHTML
	if !Config.Authenticate {
		ahtm = ""
	}

	subs := map[string]interface{}{
		"version":       VERSION,
		"authhtm":       ahtm,
		"env":           env,
		"user":          "Anonymous",
		"resultcontext": s.HitContext,
		"browsecontext": s.BrowseCtx,
		"proxval":       s.Proximity}

	f, e := efs.ReadFile("emb/frontpage.html")
	chke(e)

	tmpl, e := template.New("fp").Parse(string(f))
	chke(e)

	var b bytes.Buffer
	err := tmpl.Execute(&b, subs)
	chke(err)

	return c.HTML(http.StatusOK, b.String())
}

// readUUIDCookie - find the ID of the client
func readUUIDCookie(c echo.Context) string {
	cookie, err := c.Cookie("ID")
	if err != nil {
		id := writeUUIDCookie(c)
		return id
	}
	id := cookie.Value

	SessionLocker.Lock()
	if _, t := SessionMap[id]; !t {
		SessionMap[id] = MakeDefaultSession(id)
	}
	SessionLocker.Unlock()

	return id
}

// writeUUIDCookie - set the ID of the client
func writeUUIDCookie(c echo.Context) string {
	cookie := new(http.Cookie)
	cookie.Name = "ID"
	cookie.Path = "/"
	cookie.Value = uuid.New().String()
	cookie.Expires = time.Now().Add(4800 * time.Hour)
	c.SetCookie(cookie)
	msg(fmt.Sprintf("writeUUIDCookie() - new ID set: %s", cookie.Value), MSGPEEK)
	return cookie.Value
}
