//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
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

type ServerSession struct {
	ID          string
	IsLoggedIn  bool
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
	BrowseCtx   int64
	InputStyle  string
	HitLimit    int64
	HitContext  int
	Earliest    string
	Latest      string
	TmpInt      int
	TmpStr      string
}

func RtFrontpage(c echo.Context) error {
	// will set if missing
	s := sessions[readUUIDCookie(c)]

	env := fmt.Sprintf("%s: %s - %s (%d workers)", runtime.Version(), runtime.GOOS, runtime.GOARCH, cfg.WorkerCount)

	subs := map[string]interface{}{
		"version":       VERSION,
		"env":           env,
		"authentic":     "",
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

// makedefaultsession - fill in the blanks when setting up a new session
func makedefaultsession(id string) ServerSession {
	// note that sessions clears every time the server restarts

	var s ServerSession
	s.ID = id
	s.ActiveCorp = cfg.DefCorp
	s.VariaOK = true
	s.IncertaOK = true
	s.SpuriaOK = true
	s.NearOrNot = "near"
	s.HitLimit = DEFAULTHITLIMIT
	s.Earliest = MINDATESTR
	s.Latest = MAXDATESTR
	s.SortHitsBy = SORTBY
	s.HitContext = DEFAULTLINESOFCONTEXT
	s.BrowseCtx = cfg.BrowserCtx
	s.SearchScope = DEFAULTPROXIMITYSCOPE
	s.Proximity = DEFAULTPROXIMITY

	if AUTHENTICATIONREQUIRED {
		s.IsLoggedIn = false
	} else {
		s.IsLoggedIn = true
	}

	//msg("makedefaultsession() in non-default state for testing", 1)
	//s.Inclusions.Passages = []string{"gr5030_FROM_10106_TO_10122"}

	return s
}

// readUUIDCookie - find the ID of the client
func readUUIDCookie(c echo.Context) string {
	cookie, err := c.Cookie("ID")
	if err != nil {
		id := writeUUIDCookie(c)
		return id
	}
	id := cookie.Value

	if _, t := sessions[id]; !t {
		sessions[id] = makedefaultsession(id)
	}

	if !sessions[id].IsLoggedIn {
		// go to authentication code
		// at the moment everyone should always be marked as logged in
	}

	return id
}

// writeUUIDCookie - set the ID of the client
func writeUUIDCookie(c echo.Context) string {
	// note that cookie.Path = "/" is essential; otherwise different cookies for different contexts: "/browse" vs "/"
	cookie := new(http.Cookie)
	cookie.Name = "ID"
	cookie.Path = "/"
	cookie.Value = uuid.New().String()
	cookie.Expires = time.Now().Add(4800 * time.Hour)
	c.SetCookie(cookie)
	msg(fmt.Sprintf("new ID set: %s", cookie.Value), 4)
	return cookie.Value
}
