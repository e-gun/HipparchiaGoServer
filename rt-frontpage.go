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

//
// POOLED SERVERSESSIONS
//

type ServerSessionPool struct {
	ClientMap    map[string]*ServerSession
	AddSess      chan *ServerSession
	RemoveSess   chan string
	Update       chan *ServerSession
	FetchSS      chan string
	ReturnSess   chan *ServerSession
	RequestExist chan string
	Exists       chan SrchMsg
}

func SessionsFillNewPool() *ServerSessionPool {
	return &ServerSessionPool{
		ClientMap:    make(map[string]*ServerSession),
		AddSess:      make(chan *ServerSession),
		RemoveSess:   make(chan string),
		Update:       make(chan *ServerSession),
		FetchSS:      make(chan string),
		ReturnSess:   make(chan *ServerSession),
		RequestExist: make(chan string),
		Exists:       make(chan SrchMsg),
	}
}

func (p *ServerSessionPool) ServerSessionPoolStartListening() {
	for {
		select {
		case ss := <-p.AddSess:
			p.ClientMap[ss.ID] = ss
			msg(fmt.Sprintf("Session added; SessionPool size is %d", len(p.ClientMap)), 4)
			break
		case ss := <-p.RemoveSess:
			if _, ok := p.ClientMap[ss]; ok {
				delete(p.ClientMap, ss)
			}
			break
		case up := <-p.Update:
			p.ClientMap[up.ID] = up
			break
		case get := <-p.FetchSS:
			if s, ok := p.ClientMap[get]; ok {
				p.ReturnSess <- s
			}
			break
		case f := <-p.FetchSS:
			var ss *ServerSession
			if s, ok := p.ClientMap[f]; ok {
				ss = s
			}
			p.ReturnSess <- ss
		case exists := <-p.RequestExist:
			found := false
			if _, ok := p.ClientMap[exists]; ok {
				found = true
			}
			p.Exists <- SrchMsg{ID: exists, Yes: found}
			break
		}
	}
}

type ServerSession struct {
	ID          string
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
	BrowseCtx   int
	InputStyle  string
	HitLimit    int
	HitContext  int
	Earliest    string
	Latest      string
	TmpInt      int
	TmpStr      string
	LoginName   string
}

// FetchSession - get a session from the session pool
func FetchSession(u string) ServerSession {
	SessionPool.FetchSS <- u
	s := <-SessionPool.ReturnSess
	if s.ID != u {
		ds := makedefaultsession(u)
		s = &ds
	}
	return *s
}

// SessionInsert - insert a session into the session pool
func SessionInsert(ns ServerSession) {
	SessionPool.AddSess <- &ns
}

// SessionRemove - remove a session from the session pool
func SessionRemove(u string) {
	SessionPool.RemoveSess <- u
}

//
// ROUTING
//

// RtFrontpage - send the html for "/"
func RtFrontpage(c echo.Context) error {
	// will set if missing
	user := readUUIDCookie(c)
	s := FetchSession(user)

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

// makedefaultsession - fill in the blanks when setting up a new session
func makedefaultsession(id string) ServerSession {
	// note that SessionMap clears every time the server restarts

	var s ServerSession
	s.ID = id
	s.ActiveCorp = Config.DefCorp
	s.VariaOK = true
	s.IncertaOK = true
	s.SpuriaOK = true
	s.NearOrNot = "near"
	s.HitLimit = DEFAULTHITLIMIT
	s.Earliest = MINDATESTR
	s.Latest = MAXDATESTR
	s.SortHitsBy = SORTBY
	s.HitContext = DEFAULTLINESOFCONTEXT
	s.BrowseCtx = Config.BrowserCtx
	s.SearchScope = DEFAULTPROXIMITYSCOPE
	s.Proximity = DEFAULTPROXIMITY
	s.LoginName = "Anonymous"

	// readUUIDCookie() called this function, and it already holds a lock
	if Config.Authenticate {
		AuthorizedMap[id] = false
	} else {
		AuthorizedMap[id] = true
	}

	//msg("makedefaultsession() in non-default state for testing; this is not a release build of HGS", 0)
	//m := make(map[string]string)
	//m["lt0917_FROM_1431_TO_2193"] = "Lucanus, Marcus Annaeus, Bellum Civile, 3"
	//m["lt0917_FROM_2_TO_692"] = "Lucanus, Marcus Annaeus, Bellum Civile, 1"
	//m["lt0917_FROM_5539_TO_6410"] = "Lucanus, Marcus Annaeus, Bellum Civile, 8"
	//m["lt0917_FROM_6411_TO_7520"] = "Lucanus, Marcus Annaeus, Bellum Civile, 9"
	//m["lt0917_FROM_4666_TO_5538"] = "Lucanus, Marcus Annaeus, Bellum Civile, 7"
	//m["lt0917_FROM_3019_TO_3835"] = "Lucanus, Marcus Annaeus, Bellum Civile, 5"
	//s.Inclusions.Passages = []string{"lt0917_FROM_6411_TO_7520", "lt0917_FROM_4666_TO_5538", "lt0917_FROM_3019_TO_3835",
	//	"lt0917_FROM_1431_TO_2193", "lt0917_FROM_2_TO_692", "lt0917_FROM_5539_TO_6410"}
	//s.Inclusions.MappedPsgByName = m
	//s.Proximity = 4
	//s.SearchScope = "words"
	// s.Inclusions.BuildPsgByName()
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

	SessionPool.RequestExist <- id
	exists := <-SessionPool.Exists
	if !exists.Yes && exists.ID == id {
		s := makedefaultsession(id)
		SessionPool.AddSess <- &s
	}

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
	msg(fmt.Sprintf("new ID set: %s", cookie.Value), 4)
	return cookie.Value
}
