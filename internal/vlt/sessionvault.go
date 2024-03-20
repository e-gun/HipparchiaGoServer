//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vlt

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/str"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"sync"
	"time"
)

//
// THREAD SAFE INFRASTRUCTURE: MUTEX
//

// MakeSessionVault - called only once; yields the AllSessions vault
func MakeSessionVault() SessionVault {
	return SessionVault{
		SessionMap: make(map[string]str.ServerSession),
		mutex:      sync.RWMutex{},
	}
}

// SessionVault - there should be only one of these; and it contains all the sessions
type SessionVault struct {
	SessionMap map[string]str.ServerSession
	mutex      sync.RWMutex
}

func (sv *SessionVault) InsertSess(s str.ServerSession) {
	sv.mutex.Lock()
	defer sv.mutex.Unlock()
	sv.SessionMap[s.ID] = s
}

func (sv *SessionVault) Delete(id string) {
	sv.mutex.Lock()
	defer sv.mutex.Unlock()
	delete(sv.SessionMap, id)
}

func (sv *SessionVault) IsInVault(id string) bool {
	sv.mutex.Lock()
	defer sv.mutex.Unlock()
	_, b := sv.SessionMap[id]
	// fmt.Println(StringMapKeysIntoSlice(sv.SessionMap))
	return b
}

func (sv *SessionVault) GetSess(id string) str.ServerSession {
	sv.mutex.Lock()
	defer sv.mutex.Unlock()
	s, e := sv.SessionMap[id]
	if e != true {
		s = MakeDefaultSession(id)
	}
	return s
}

// MakeDefaultSession - fill in the blanks when setting up a new session
func MakeDefaultSession(id string) str.ServerSession {
	// note that SessionMap clears every time the server restarts

	var s str.ServerSession
	s.ID = id
	s.ActiveCorp = lnch.Config.DefCorp
	s.VariaOK = true
	s.IncertaOK = true
	s.SpuriaOK = true
	s.NearOrNot = "near"
	s.HitLimit = vv.DEFAULTHITLIMIT
	s.Earliest = vv.MINDATESTR
	s.Latest = vv.MAXDATESTR
	s.SortHitsBy = vv.SORTBY
	s.HitContext = vv.DEFAULTLINESOFCONTEXT
	s.BrowseCtx = lnch.Config.BrowserCtx
	s.SearchScope = vv.DEFAULTPROXIMITYSCOPE
	s.Proximity = vv.DEFAULTPROXIMITY
	s.LoginName = "Anonymous"
	s.VocScansion = lnch.Config.VocabScans
	s.VocByCount = lnch.Config.VocabByCt
	s.VecGraphExt = lnch.Config.VectorWebExt
	s.VecNeighbCt = lnch.Config.VectorNeighb
	s.VecNNSearch = false
	s.VecModeler = lnch.Config.VectorModel
	s.VecTextPrep = lnch.Config.VectorTextPrep
	s.VecLDASearch = false
	s.LDA2D = true

	if lnch.Config.Authenticate {
		AllAuthorized.Register(id, false)
	} else {
		AllAuthorized.Register(id, true)
	}

	//mm("MakeDefaultSession() in non-default lnch for testing; this is not a release build of HGS", 0)
	//
	//s.VecLDASearch = true
	//s.VecNNSearch = true

	//mm := make(map[string]string)
	//mm["lt0917_FROM_1431_TO_2193"] = "Lucanus, Marcus Annaeus, Bellum Civile, 3"
	//mm["lt0917_FROM_2_TO_692"] = "Lucanus, Marcus Annaeus, Bellum Civile, 1"
	//mm["lt0917_FROM_5539_TO_6410"] = "Lucanus, Marcus Annaeus, Bellum Civile, 8"
	//mm["lt0917_FROM_6411_TO_7520"] = "Lucanus, Marcus Annaeus, Bellum Civile, 9"
	//mm["lt0917_FROM_4666_TO_5538"] = "Lucanus, Marcus Annaeus, Bellum Civile, 7"
	//mm["lt0917_FROM_3019_TO_3835"] = "Lucanus, Marcus Annaeus, Bellum Civile, 5"
	//s.Inclusions.Passages = []string{"lt0917_FROM_6411_TO_7520", "lt0917_FROM_4666_TO_5538", "lt0917_FROM_3019_TO_3835",
	//	"lt0917_FROM_1431_TO_2193", "lt0917_FROM_2_TO_692", "lt0917_FROM_5539_TO_6410"}
	//s.Inclusions.MappedPsgByName = mm
	//s.Proximity = 4
	//s.SearchScope = "words"
	//s.Inclusions.BuildPsgByName()

	return s
}

// cookies here for import issues

// ReadUUIDCookie - find the ID of the client
func ReadUUIDCookie(c echo.Context) string {
	cookie, err := c.Cookie("ID")
	if err != nil {
		id := WriteUUIDCookie(c)
		return id
	}
	id := cookie.Value

	if !AllSessions.IsInVault(id) {
		AllSessions.InsertSess(MakeDefaultSession(id))
	}

	return id
}

// WriteUUIDCookie - set the ID of the client
func WriteUUIDCookie(c echo.Context) string {
	cookie := new(http.Cookie)
	cookie.Name = "ID"
	cookie.Path = "/"
	cookie.Value = uuid.New().String()
	cookie.Expires = time.Now().Add(4800 * time.Hour)
	c.SetCookie(cookie)
	Msg.TMI(fmt.Sprintf("WriteUUIDCookie() - new ID set: %s", cookie.Value))
	return cookie.Value
}
