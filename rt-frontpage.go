//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
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
	"sort"
	"strings"
	"text/template"
	"time"
)

var (
	// have the option to return/generate some sort of fail message...
	emptyjsreturn = func(c echo.Context) error { return c.JSONPretty(http.StatusOK, "", JSONINDENT) }
)

//
// ROUTING
//

// RtFrontpage - send the html for "/"
func RtFrontpage(c echo.Context) error {
	const (
		UPSTR    = "[%v] HGS uptime: %v [%s]"
		PADDING  = " ----------------- "
		STATTMPL = "%s: %d"
		SPACER   = "    "
		VECTORS  = `
        <span id="vectorsearchcheckbox">
            <span class="rarechars small">v⃗</span><input type="checkbox" id="isvectorsearch" value="yes">
        </span>
       <span id="ldasearches">
            <span class="rarechars small">τ⃗</span><input type="checkbox" id="isldasearch" value="yes">
        </span>`
	)
	// will set if missing
	user := readUUIDCookie(c)
	s := AllSessions.GetSess(user)

	ahtm := AUTHHTML
	if !Config.Authenticate {
		ahtm = ""
	}

	gc := GitCommit
	if gc == "" {
		gc = "UNKNOWN"
	}
	ver := fmt.Sprintf("Version: %s [git: %s]", VERSION+VersSuppl, gc)

	env := fmt.Sprintf("%s: %s - %s (%d workers)", runtime.Version(), runtime.GOOS, runtime.GOARCH, Config.WorkerCount)

	// t() will give the uptime
	var mem runtime.MemStats

	t := func(up time.Duration) string {
		runtime.ReadMemStats(&mem)
		heap := fmt.Sprintf("%dM", mem.HeapAlloc/1024/1024)
		tick := fmt.Sprintf(UPSTR, time.Now().Format(time.TimeOnly), up.Truncate(time.Minute), heap)
		return PADDING + tick + PADDING
	}

	// svd() will report what requests have been made
	svd := func() string {
		responder := PIReply{req: true, response: make(chan map[string]int)}
		PIRequest <- responder
		ctr := <-responder.response

		exclude := []string{"main() post-initialization"}
		keys := StringMapKeysIntoSlice(ctr)
		keys = SetSubtraction(keys, exclude)
		sort.Strings(keys)

		var pairs []string
		for k := range keys {
			this := strings.TrimPrefix(keys[k], "Rt")
			this = strings.TrimSuffix(this, "()")
			pairs = append(pairs, fmt.Sprintf(SPACER+STATTMPL, this, ctr[keys[k]]))
		}
		return strings.Join(pairs, "\n")
	}

	vec := ""
	if !Config.VectorsDisabled {
		vec = VECTORS
	}

	// sample ticker output

	//      ----------------- [13:29:41] HGS uptime: 1m0s -----------------
	//
	//    BrowseLine: 5
	//    LexFindByForm: 2
	//    Search: 4

	subs := map[string]interface{}{
		"version":       VERSION + VersSuppl,
		"longver":       ver,
		"authhtm":       ahtm,
		"vec":           vec,
		"env":           env,
		"ticker":        t(time.Since(LaunchTime)) + "\n\n" + svd(),
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

	// .GetSess() will make a new session if id is not found
	_ = AllSessions.GetSess(id)

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
	msg(fmt.Sprintf("writeUUIDCookie() - new ID set: %s", cookie.Value), MSGTMI)
	return cookie.Value
}
