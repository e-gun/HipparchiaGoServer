//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"strings"
	"time"
)

type PollData struct {
	TotalWrk int    `json:"Poolofwork"`
	Remain   int    `json:"Remaining"`
	Hits     int    `json:"Hitcount"`
	Msg      string `json:"Statusmessage"`
	Elapsed  string `json:"Elapsed"`
	Extra    string `json:"Notes"`
	ID       string `json:"ID"`
	TwoBox   bool
}

// RtWebsocket - progress info for SearchMap
func RtWebsocket(c echo.Context) error {
	// 	the client sends the name of a poll and this will output
	//	the status of the poll continuously while the poll remains active
	//
	//	example:
	//		progress {'active': 1, 'total': 20, 'remaining': 20, 'hits': 48, 'message': 'Putting the results in context',
	//		'elapsed': 14.0, 'extrainfo': '<span class="small"></span>'}

	// see also /static/hipparchiajs/progressindicator_go.js

	// https://echo.labstack.com/cookbook/websocket/

	// you can spend 3.5s on a search vs 2.0 seconds if you poll as fast as possible
	// POLLEVERYNTABLES in SrchFeeder() and WSPOLLINGPAUSE here make a huge difference

	// does not look like you can run two wss at once; not obvious why this is the case...

	const (
		FAILCON   = "RtWebsocket(): ws connection failed"
		FAILRD    = "RtWebsocket(): ws failed to read: breaking"
		FAILWR    = "RtWebsocket(): ws failed to write: breaking"
		FAILSND   = "RtWebsocket() could not send stop message"
		FAILCLOSE = "RtWebsocket() failed to close"
	)

	type JSOut struct {
		V     string `json:"value"`
		ID    string `json:"ID"`
		Close string `json:"close"`
	}

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		msg(FAILCON, 2)
		return nil
	}

	for {
		maplocker.RLock()
		ls := len(SearchMap)
		maplocker.RUnlock()
		if ls != 0 {
			break
		}
	}

	done := false
	bs := ""

	for {
		if done {
			break
		}
		// Read

		var m []byte
		_, m, e := ws.ReadMessage()
		if e != nil {
			msg(FAILRD, 5)
			break
		}

		// will yield: websocket received: "205da19d"
		// bug-trap: the quotes around that string

		bs = string(m)
		bs = strings.Replace(bs, `"`, "", -1)
		_, found := SearchMap[bs]

		if found && SearchMap[bs].IsActive {
			var r PollData
			r.TwoBox = SearchMap[bs].Twobox

			for {
				r.ID = bs
				r.TotalWrk = SearchMap[bs].TableSize
				r.Elapsed = fmt.Sprintf("%.1fs", time.Now().Sub(SearchMap[bs].Launched).Seconds())

				if SearchMap[bs].PhaseNum > 1 {
					r.Extra = "(second pass)"
				} else {
					r.Extra = ""
				}

				// is lock/unlock of the relevant mutex in fact unneeded: only this loop will read this value from that map
				r.Remain = SearchMap[bs].Remain.Get()
				r.Hits = SearchMap[bs].Hits.Get()

				// inside the loop because indexing modifies InitSum to send simple progress messages
				r.Msg = strings.Replace(SearchMap[bs].InitSum, "Sought", "Seeking", -1)

				// Write
				pd := formatpoll(r)

				jso := JSOut{
					V:     pd,
					ID:    r.ID,
					Close: "open",
				}

				js, y := json.Marshal(jso)
				chke(y)

				er := ws.WriteMessage(websocket.TextMessage, js)

				if er != nil {
					msg(FAILWR, 5)
					done = true
					break
				} else {
					time.Sleep(WSPOLLINGPAUSE)
				}

				maplocker.RLock()
				_, exists := SearchMap[bs]
				maplocker.RUnlock()
				if !exists {
					done = true
					break
				}
			}
		}
	}

	// tell the websocket on the other end to close
	// this is not supposed to be strictly necessary, but there have been problems reconnecting after multiple SearchMap
	end := JSOut{
		V:     "",
		ID:    bs,
		Close: "close",
	}

	stop, y := json.Marshal(end)
	chke(y)

	er := ws.WriteMessage(websocket.TextMessage, stop)
	if er != nil {
		msg(FAILSND, 5)
	}

	err = ws.Close()
	if err != nil {
		msg(FAILCLOSE, 5)
	}
	return nil
}

// formatpoll - build HTML to send to the JS on the other side
func formatpoll(pd PollData) string {
	// example:
	// Seeking <span class="sought">»μελιϲϲα«</span>: <span class="progress">31%</span> completed&nbsp;(0.3s)<br>
	// (<span class="progress">199</span> found)<br>
	pctd := ((float32(pd.TotalWrk) - float32(pd.Remain)) / float32(pd.TotalWrk)) * 100
	pcts := fmt.Sprintf("%.0f", pctd) + "%"

	htm := pd.Msg

	tp := func() string {
		m := `Finishing up...&nbsp;`
		if pd.TwoBox {
			m = `Searching for matches among the initial finds...&nbsp;`
		}
		return fmt.Sprintf(`&nbsp;(%s)<br>%s`, pd.Elapsed, m)
	}()

	if pctd != 0 && pd.Remain != 0 && pd.TotalWrk != 0 {
		// normal in progress
		htm += fmt.Sprintf(`: <span class="progress">%s</span> completed&nbsp;(%s)<br>`, pcts, pd.Elapsed)
	} else if pd.Remain == 0 && pd.TotalWrk != 0 {
		// finished, mostly
		htm += tp
	} else if pd.TotalWrk == 0 {
		// vocab or index run have no "total work"
		htm += fmt.Sprintf(`&nbsp;(%s)`, pd.Elapsed)
	} else {
		// fallback
		htm += fmt.Sprintf(`&nbsp;(%s)`, pd.Elapsed)
	}

	if pd.Hits > 0 {
		htm += fmt.Sprintf(`(<span class="progress">%d</span> found)<br>`, pd.Hits)
	}

	if len(pd.Extra) != 0 {
		htm += pd.Extra
	}

	return htm
}
