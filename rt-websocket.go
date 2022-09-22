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
	Active   string `json:"active"`
	TotalWrk int    `json:"Poolofwork"`
	Remain   int    `json:"Remaining"`
	Hits     int    `json:"Hitcount"`
	Msg      string `json:"Statusmessage"`
	Elapsed  string `json:"Elapsed"`
	Extra    string `json:"Notes"`
	ID       string `json:"ID"`
}

// RtWebsocket - progress info for searches
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

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		msg("RtWebsocket(): ws connection failed", 1)
		return nil
	}
	defer ws.Close()

	for {
		if len(searches) != 0 {
			break
		}
	}

	done := false

	for {
		if done {
			break
		}
		// Read

		m := []byte{}
		_, m, e := ws.ReadMessage()
		if e != nil {
			// c.Logger().Error(err)
			msg("RtWebsocket(): ws failed to read: breaking", 1)
			break
		}

		// will yield: websocket received: "205da19d"
		// bug-trap: the quotes around that string

		bs := string(m)
		bs = strings.Replace(bs, `"`, "", -1)
		_, found := searches[bs]

		if found && searches[bs].IsActive {
			for {
				var r PollData
				r.Active = "is_active"
				r.ID = bs
				r.TotalWrk = searches[bs].TableSize
				r.Elapsed = fmt.Sprintf("%.1fs", time.Now().Sub(searches[bs].Launched).Seconds())

				if searches[bs].PhaseNum > 1 {
					r.Extra = "(second pass)"
				} else {
					r.Extra = ""
				}

				result, ok := progremain.Load(searches[bs].ID)
				if ok {
					r.Remain = result.(int)
				}

				result, ok = proghits.Load(searches[bs].ID)
				if ok {
					r.Hits = result.(int)
				}

				// inside the loop because indexing modifies InitSum to send simple progress messages
				r.Msg = strings.Replace(searches[bs].InitSum, "Sought", "Seeking", -1)

				// Write
				pd := formatpoll(r)

				type JSOut struct {
					V  string `json:"value"`
					ID string `json:"ID"`
				}

				jso := JSOut{
					V:  pd,
					ID: r.ID,
				}

				js, y := json.Marshal(jso)
				chke(y)

				er := ws.WriteMessage(websocket.TextMessage, js)

				if er != nil {
					c.Logger().Error(er)
					msg("RtWebsocket(): ws failed to write: breaking", 1)
					done = true
					break
				} else {
					time.Sleep(WSPOLLINGPAUSE)
				}

				if _, exists := searches[bs]; !exists {
					done = true
					break
				}
			}
		}
	}
	return nil
}

// formatpoll - build HTML to send to the JS on the other side
func formatpoll(pd PollData) string {
	// example:
	// Seeking <span class="sought">»μελιϲϲα«</span>: <span class="progress">31%</span> completed&nbsp;(0.3s)<br>(<span class="progress">199</span> found)<br>
	pctd := ((float32(pd.TotalWrk) - float32(pd.Remain)) / float32(pd.TotalWrk)) * 100
	pcts := fmt.Sprintf("%.0f", pctd) + "%"

	htm := pd.Msg

	if pctd != 0 && pd.Remain != 0 && pd.TotalWrk != 0 {
		// normal in progress
		htm += fmt.Sprintf(`: <span class="progress">%s</span> completed&nbsp;(%s)<br>`, pcts, pd.Elapsed)
	} else if pd.Remain == 0 && pd.TotalWrk != 0 {
		// finished, mostly
		htm += fmt.Sprintf(`&nbsp;(%s)<br>Finishing up...&nbsp;`, pd.Elapsed)
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
