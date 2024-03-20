package debug

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"strings"
	"time"
)

//
// FOR DEBUGGING ONLY
//

// WSClientReport - report the # and names of the active wsclients every N seconds
func WSClientReport(d time.Duration) {
	// add the following to main.go: "go WSClientReport()"
	for {
		cl := vlt.WebsocketPool.ClientMap
		var cc []string
		for k := range cl {
			cc = append(cc, k.ID)
		}
		Msg.NOTE(fmt.Sprintf("%d WebsocketPool clients: %s", len(cl), strings.Join(cc, ", ")))
		time.Sleep(d)
	}
}
