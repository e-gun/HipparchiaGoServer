package vaults

import (
	"github.com/e-gun/HipparchiaGoServer/internal/launch"
	"github.com/e-gun/HipparchiaGoServer/internal/structs"
	"sync"
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
