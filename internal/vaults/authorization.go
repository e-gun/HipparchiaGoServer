package vaults

import (
	"encoding/json"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/launch"
	"github.com/e-gun/HipparchiaGoServer/internal/structs"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"io"
	"os"
	"sync"
)

var (
	UserPassPairs = make(map[string]string)
)

//
// THREAD SAFE INFRASTRUCTURE: MUTEX
//

// MakeAuthorizedVault - called only once; yields the AllAuthorized vault
func MakeAuthorizedVault() AuthVault {
	return AuthVault{
		UserMap: make(map[string]bool),
		mutex:   sync.RWMutex{},
	}
}

// AuthVault - there should be only one of these; and it contains all the authorization info
type AuthVault struct {
	UserMap map[string]bool
	mutex   sync.RWMutex
}

func (av *AuthVault) Check(u string) bool {
	if !launch.Config.Authenticate {
		return true
	}
	av.mutex.Lock()
	defer av.mutex.Unlock()
	s, e := av.UserMap[u]
	if e != true {
		av.UserMap[u] = false
		s = false
	}
	return s
}

func (av *AuthVault) Register(u string, b bool) {
	av.mutex.Lock()
	defer av.mutex.Unlock()
	av.UserMap[u] = b
	return
}

// BuildUserPassPairs - set up authentication map via CONFIGAUTH
func BuildUserPassPairs(cc structs.CurrentConfiguration) {
	const (
		FAIL1 = `failed to unmarshall authorization config file`
		FAIL2 = `You are requiring authentication but there are no UserPassPairs: aborting vv`
		FAIL3 = "Could not open '%s'"
	)

	uh, _ := os.UserHomeDir()
	h := fmt.Sprintf(vv.CONFIGALTAPTH, uh)
	pwf := fmt.Sprintf("%s%s", h, vv.CONFIGAUTH)

	pwc, e := os.Open(pwf)
	if e != nil {
		Msg.CRIT(fmt.Sprintf(FAIL3, pwf))
	}
	defer func(pwc *os.File) {
		err := pwc.Close()
		if err != nil {
		} // the file was almost certainly not found in the first place...
	}(pwc)

	filebytes, _ := io.ReadAll(pwc)

	type UserPass struct {
		User string
		Pass string
	}

	var upp []UserPass
	err := json.Unmarshal(filebytes, &upp)
	if err != nil {
		Msg.NOTE(FAIL1)
	}

	for _, u := range upp {
		UserPassPairs[u.User] = u.Pass
	}

	if cc.Authenticate && len(UserPassPairs) == 0 {
		Msg.CRIT(FAIL2)
		os.Exit(1)
	}
}
