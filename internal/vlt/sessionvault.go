package vlt

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/launch"
	"github.com/e-gun/HipparchiaGoServer/internal/structs"
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
		SessionMap: make(map[string]structs.ServerSession),
		mutex:      sync.RWMutex{},
	}
}

// SessionVault - there should be only one of these; and it contains all the sessions
type SessionVault struct {
	SessionMap map[string]structs.ServerSession
	mutex      sync.RWMutex
}

func (sv *SessionVault) InsertSess(s structs.ServerSession) {
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

func (sv *SessionVault) GetSess(id string) structs.ServerSession {
	sv.mutex.Lock()
	defer sv.mutex.Unlock()
	s, e := sv.SessionMap[id]
	if e != true {
		s = launch.MakeDefaultSession(id)
	}
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
		AllSessions.InsertSess(launch.MakeDefaultSession(id))
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
