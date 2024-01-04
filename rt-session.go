//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"sync"
	"time"
)

//
// SERVERSESSIONS
//

type ServerSession struct {
	ID           string
	Inclusions   SearchIncExl
	Exclusions   SearchIncExl
	ActiveCorp   map[string]bool
	VariaOK      bool   `json:"varia"`
	IncertaOK    bool   `json:"incerta"`
	SpuriaOK     bool   `json:"spuria"`
	RawInput     bool   `json:"rawinputstyle"`
	OneHit       bool   `json:"onehit"`
	HeadwordIdx  bool   `json:"headwordindexing"`
	FrqIdx       bool   `json:"indexbyfrequency"`
	VocByCount   bool   `json:"vocbycount"`
	VocScansion  bool   `json:"vocscansion"`
	NearOrNot    string `json:"nearornot"`
	SearchScope  string `json:"searchscope"`
	SortHitsBy   string `json:"sortorder"`
	Proximity    int    `json:"proximity"`
	BrowseCtx    int
	InputStyle   string
	HitLimit     int
	HitContext   int
	Earliest     string
	Latest       string
	TmpInt       int
	TmpStr       string
	LoginName    string
	VecGraphExt  bool
	VecModeler   string
	VecNeighbCt  int
	VecNNSearch  bool
	VecTextPrep  string
	VecLDASearch bool
	LDAgraph     bool
	LDAtopics    int
	LDA2D        bool
}

// BuildSelectionOverview will call the relevant SearchIncExl functions: see searchlistbuilder.go
func (s *ServerSession) BuildSelectionOverview() {
	s.Inclusions.BuildAuByName()
	s.Exclusions.BuildAuByName()
	s.Inclusions.BuildWkByName()
	s.Exclusions.BuildWkByName()
	s.Inclusions.BuildPsgByName()
	s.Exclusions.BuildPsgByName()
}

//
// THREAD SAFE INFRASTRUCTURE: MUTEX
//

// MakeSessionVault - called only once; yields the AllSessions vault
func MakeSessionVault() SessionVault {
	return SessionVault{
		SessionMap: make(map[string]ServerSession),
		mutex:      sync.RWMutex{},
	}
}

// SessionVault - there should be only one of these; and it contains all the sessions
type SessionVault struct {
	SessionMap map[string]ServerSession
	mutex      sync.RWMutex
}

func (sv *SessionVault) InsertSess(s ServerSession) {
	sv.mutex.Lock()
	defer sv.mutex.Unlock()
	sv.SessionMap[s.ID] = s
}

func (sv *SessionVault) Delete(id string) {
	sv.mutex.Lock()
	defer sv.mutex.Unlock()
	delete(sv.SessionMap, id)
}

func (sv *SessionVault) GetSess(id string) ServerSession {
	sv.mutex.Lock()
	defer sv.mutex.Unlock()
	s, e := sv.SessionMap[id]
	if e != true {
		s = MakeDefaultSession(id)
	}
	return s
}

// MakeDefaultSession - fill in the blanks when setting up a new session
func MakeDefaultSession(id string) ServerSession {
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
	s.VocScansion = Config.VocabScans
	s.VocByCount = Config.VocabByCt
	s.VecGraphExt = Config.VectorWebExt
	s.VecNeighbCt = Config.VectorNeighb
	s.VecNNSearch = false
	s.VecModeler = Config.VectorModel
	s.VecTextPrep = Config.VectorTextPrep
	s.VecLDASearch = false
	s.LDA2D = true

	if Config.Authenticate {
		AllAuthorized.Register(id, false)
	} else {
		AllAuthorized.Register(id, true)
	}

	//msg("MakeDefaultSession() in non-default state for testing; this is not a release build of HGS", 0)
	//
	//s.VecLDASearch = true
	//s.VecNNSearch = true

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
	//s.Inclusions.BuildPsgByName()

	return s
}

//
// ROUTING
//

// RtSessionSetsCookie - turn the session into a cookie
func RtSessionSetsCookie(c echo.Context) error {
	const (
		FAIL = "RtSessionSetsCookie() could not marshal the session"
	)
	num := c.Param("num")
	user := readUUIDCookie(c)
	s := AllSessions.GetSess(user)

	v, e := json.Marshal(s)
	if e != nil {
		v = []byte{}
		msg(FAIL, MSGWARN)
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

	return c.JSONPretty(http.StatusOK, "", JSONINDENT)
}

// RtSessionGetCookie - turn a stored cookie into a session
func RtSessionGetCookie(c echo.Context) error {
	// this code has input trust issues...
	const (
		FAIL1 = "RtSessionGetCookie failed to read cookie %s for %s"
		FAIL2 = "RtSessionGetCookie failed to unmarshal cookie %s for %s"
	)

	user := readUUIDCookie(c)
	num := c.Param("num")
	cookie, err := c.Cookie("session" + num)
	if err != nil {
		msg(fmt.Sprintf(FAIL1, num, user), MSGWARN)
		return c.String(http.StatusOK, "")
	}

	var s ServerSession
	// invalid character '%' looking for beginning of object key string:
	// {%22ID%22:%22723073ae-09a7-4b24-a5d6-7e20603d8c44%22%2C%22IsLoggedIn%22:true%2C...}
	swap := strings.NewReplacer("%22", `"`, "%2C", ",", "%20", " ")
	cv := swap.Replace(cookie.Value)

	err = json.Unmarshal([]byte(cv), &s)
	if err != nil {
		// invalid character '%' looking for beginning of object key string
		msg(fmt.Sprintf(FAIL2, num, user), MSGWARN)
		fmt.Println(err)
		return c.String(http.StatusOK, "")
	}

	AllSessions.InsertSess(s)

	e := c.Redirect(http.StatusFound, "/")
	chke(e)
	return nil
}

// RtResetSession - delete and then reset the session
func RtResetSession(c echo.Context) error {
	user := readUUIDCookie(c)
	AllSessions.Delete(user)

	// then reset it
	readUUIDCookie(c)
	e := c.Redirect(http.StatusFound, "/")
	chke(e)
	return nil
}
