//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type ServerSession struct {
	ID             string
	Inclusions     SearchIncExl
	Exclusions     SearchIncExl
	ActiveCorp     map[string]bool
	VariaOK        bool            `json:"varia"`
	IncertaOK      bool            `json:"incerta"`
	SpuriaOK       bool            `json:"spuria"`
	AvailDBs       map[string]bool `json:"available"`
	RawInput       bool            `json:"rawinputstyle"`
	OneHit         bool            `json:"onehit"`
	HeadwordIdx    bool            `json:"headwordindexing"`
	FrqIdx         bool            `json:"indexbyfrequency"`
	NearOrNot      string          `json:"nearornot"`
	SearchScope    string          `json:"searchscope"`
	SortHitsBy     string          `json:"sortorder"`
	Proximity      int             `json:"proximity"`
	Analogyfinder  bool            `json:"analogyfinder"`
	Authorflagging bool            `json:"authorflagging"`
	Authorssummary bool            `json:"authorssummary"`
	Baggingmethod  string          `json:"baggingmethod"`
	HitLimit       int64
	HitContext     int
	Earliest       string
	Latest         string
	TmpInt         int
	TmpStr         string
	UI             UISettings
}

type UISettings struct {
	BrowseCtx   int64
	InputStyle  string
	LxFlagAu    bool
	WCShow      bool
	PptAndMorph bool
}

func RtFrontpage(c echo.Context) error {
	// will set if missing
	readUUIDCookie(c)

	subs := map[string]interface{}{"version": VERSION}
	err := c.Render(http.StatusOK, "frontpage.html", subs)
	return err
}

// makedefaultsession - fill in the blanks when setting up a new session
func makedefaultsession(id string) ServerSession {
	// note that sessions clears every time the server restarts
	var s ServerSession
	s.ID = id
	s.ActiveCorp = map[string]bool{"gr": true, "lt": true, "in": false, "ch": false, "dp": false}
	s.VariaOK = true
	s.IncertaOK = true
	s.SpuriaOK = true
	s.AvailDBs = map[string]bool{"greek_dictionary": true, "greek_lemmata": true, "greek_morphology": true, "latin_dictionary": true, "latin_lemmata": true, "latin_morphology": true, "wordcounts_0": true}
	s.Analogyfinder = false
	s.HitLimit = DEFAULTHITLIMIT
	s.Earliest = MINDATESTR
	s.Latest = MAXDATESTR
	s.SortHitsBy = "Name"
	s.HitContext = DEFAULTLINESOFCONTEXT
	s.UI.BrowseCtx = DEFAULTBROWSERCTX
	s.SearchScope = DEFAULTPROXIMITYSCOPE
	s.Proximity = DEFAULTPROXIMITY
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
