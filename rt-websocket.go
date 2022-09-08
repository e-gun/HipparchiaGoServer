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

func RtWebsocket(c echo.Context) error {
	// 	the client sends the name of a poll and this will output
	//	the status of the poll continuously while the poll remains active
	//
	//	example:
	//		progress {'active': 1, 'total': 20, 'remaining': 20, 'hits': 48, 'message': 'Putting the results in context',
	//		'elapsed': 14.0, 'extrainfo': '<span class="small"></span>'}

	// see also /static/hipparchiajs/progressindicator_go.js

	// https://echo.labstack.com/cookbook/websocket/

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	type ReplyJS struct {
		Active   string `json:"active"`
		TotalWrk int    `json:"Poolofwork"`
		Remain   int    `json:"Remaining"`
		Hits     int    `json:"Hitcount"`
		Msg      string `json:"Statusmessage"`
		Elapsed  string `json:"Elapsed"`
		Extra    string `json:"Notes"`
		ID       string `json:"ID"`
	}

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
		// the bug-trap is the quotes around that string
		bs := string(m)
		bs = strings.Replace(bs, `"`, "", -1)
		mm := strings.Replace(searches[bs].InitSum, "Sought", "Seeking", -1)

		_, found := searches[bs]

		if found && searches[bs].IsActive {
			for {
				var r ReplyJS

				// [a] the easy info to report
				r.Active = "is_active"
				r.ID = bs
				r.TotalWrk = searches[bs].TableSize
				r.Elapsed = fmt.Sprintf("%.1fs", time.Now().Sub(searches[bs].Launched).Seconds())
				if searches[bs].PhaseNum == 2 {
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

				if r.Remain != 0 {
					r.Msg = mm
				}

				// Write
				js, y := json.Marshal(r)
				chke(y)

				er := ws.WriteMessage(websocket.TextMessage, js)

				if er != nil {
					c.Logger().Error(er)
					msg("RtWebsocket(): ws failed to write: breaking", 1)
					break
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
