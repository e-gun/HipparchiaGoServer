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

// RtWebsocket - progress info for a search
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
		MapLocker.RLock()
		ls := len(SearchMap)
		MapLocker.RUnlock()
		if ls != 0 {
			break
		}
	}

	done := false
	id := ""

	for {
		if done {
			break
		}

		// Read
		var m []byte
		_, m, e := ws.ReadMessage() // will yield: websocket received: "205da19d"; bug-trap: the quotes around that string
		if e != nil {
			msg(FAILRD, 5)
			break
		}

		id = string(m)
		id = strings.Replace(id, `"`, "", -1)

		srch := SafeSearchMapRead(id) // but you still have to use the map's version for some things...

		if srch.IsActive {
			var r PollData
			r.TwoBox = srch.Twobox
			r.ID = id
			r.TotalWrk = srch.TableSize

			for {
				r.Elapsed = fmt.Sprintf("%.1fs", time.Now().Sub(srch.Launched).Seconds())

				if srch.PhaseNum > 1 {
					r.Extra = "(second pass)"
				} else {
					r.Extra = ""
				}

				// mutex protected gets
				r.Remain = SearchMap[id].Remain.Get()
				r.Hits = SearchMap[id].Hits.Get()

				// inside the loop because indexing modifies InitSum to send simple progress messages
				MapLocker.RLock()
				r.Msg = strings.Replace(SearchMap[id].InitSum, "Sought", "Seeking", -1)
				MapLocker.RUnlock()

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

				MapLocker.RLock()
				_, exists := SearchMap[id]
				MapLocker.RUnlock()
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
		ID:    id,
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

	const (
		FU  = `Finishing up...&nbsp;`
		MS  = `Searching for matches among the initial finds...&nbsp;`
		PCT = `: <span class="progress">%s</span> completed&nbsp;(%s)<br>`
		EL1 = `&nbsp;(%s)<br>%s`
		EL2 = `&nbsp;(%s)`
		HIT = `(<span class="progress">%d</span> found)<br>`
	)

	pctd := ((float32(pd.TotalWrk) - float32(pd.Remain)) / float32(pd.TotalWrk)) * 100
	pcts := fmt.Sprintf("%.0f", pctd) + "%"

	htm := pd.Msg

	tp := func() string {
		m := FU
		if pd.TwoBox {
			m = MS
		}
		return fmt.Sprintf(EL1, pd.Elapsed, m)
	}()

	if pctd != 0 && pd.Remain != 0 && pd.TotalWrk != 0 {
		// normal in progress
		htm += fmt.Sprintf(PCT, pcts, pd.Elapsed)
	} else if pd.Remain == 0 && pd.TotalWrk != 0 {
		// finished, mostly
		htm += tp
	} else if pd.TotalWrk == 0 {
		// vocab or index run have no "total work"
		htm += fmt.Sprintf(EL2, pd.Elapsed)
	} else {
		// fallback
		htm += fmt.Sprintf(EL2, pd.Elapsed)
	}

	if pd.Hits > 0 {
		htm += fmt.Sprintf(HIT, pd.Hits)
	}

	if len(pd.Extra) != 0 {
		htm += pd.Extra
	}

	return htm
}
