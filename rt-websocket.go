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

//
// THE ROUTE
//

// RtWebsocket - progress info for a search (multiple clients client at a time)
func RtWebsocket(c echo.Context) error {
	// see https://tutorialedge.net/projects/chat-system-in-go-and-react/part-4-handling-multiple-clients/

	const (
		FAILCON = "RtWebsocket(): ws connection failed"
	)

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		msg(FAILCON, 2)
		return nil
	}

	progresspoll := &WSClient{
		Conn: ws,
		Pool: WebsocketPool,
	}

	WebsocketPool.Add <- progresspoll
	progresspoll.ReadID()
	progresspoll.WSWriteJSON()
	WebsocketPool.Remove <- progresspoll
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

//
// WEBSOCKET INFRASTRUCTURE
//

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

type WSClient struct {
	ID   string
	Conn *websocket.Conn
	Pool *WSPool
}

type WSPool struct {
	Add       chan *WSClient
	Remove    chan *WSClient
	ClientMap map[*WSClient]bool
	JSO       chan WSJSOut
	ReadID    chan string
}

type WSJSOut struct {
	V     string `json:"value"`
	ID    string `json:"ID"`
	Close string `json:"close"`
}

func (c *WSClient) ReadID() {
	const (
		FAIL = `WSClient.ReadID() failed`
	)
	for {
		_, m, err := c.Conn.ReadMessage()
		if err != nil {
			msg(FAIL, 3)
			return
		}

		if len(m) != 0 {
			id := string(m)
			id = strings.Replace(id, `"`, "", -1)
			c.ID = id
			c.Pool.ReadID <- id
			break
		}
	}
}

func (c *WSClient) WSWriteJSON() {
	// wait for the search to exist
	for {
		MapLocker.RLock()
		ls := len(SearchMap)
		_, exists := SearchMap[c.ID]
		MapLocker.RUnlock()
		if ls != 0 && exists {
			break
		}
	}

	srch := SafeSearchMapRead(c.ID)

	var r PollData
	r.TwoBox = srch.Twobox
	r.TotalWrk = srch.TableSize

	// loop until search finishes
	for {
		r.Elapsed = fmt.Sprintf("%.1fs", time.Now().Sub(srch.Launched).Seconds())

		if srch.PhaseNum > 1 {
			r.Extra = "(second pass)"
		} else {
			r.Extra = ""
		}

		// mutex protected gets
		r.Remain = SearchMap[c.ID].Remain.Get()
		r.Hits = SearchMap[c.ID].Hits.Get()

		MapLocker.RLock()
		r.Msg = strings.Replace(SearchMap[c.ID].InitSum, "Sought", "Seeking", -1)
		MapLocker.RUnlock()

		pd := formatpoll(r)

		jso := WSJSOut{
			V:     pd,
			ID:    c.ID,
			Close: "open",
		}

		c.Pool.JSO <- jso
		time.Sleep(WSPOLLINGPAUSE)

		MapLocker.RLock()
		_, exists := SearchMap[c.ID]
		MapLocker.RUnlock()

		if !exists {
			break
		}
	}
}

func (pool *WSPool) WSPoolStartListening() {
	const (
		MGS1 = "WSPool add: size of connection pool is %d"
		MSG2 = "WSPool remove: size of connection pool is %d"
		MSG3 = "Starting polling loop for %s"
	)
	for {
		select {
		case id := <-pool.Add:
			pool.ClientMap[id] = true
			msg(fmt.Sprintf(MGS1, len(pool.ClientMap)), 4)
			break
		case id := <-pool.Remove:
			delete(pool.ClientMap, id)
			msg(fmt.Sprintf(MSG2, len(pool.ClientMap)), 4)
			break
		case m := <-pool.ReadID:
			msg(fmt.Sprintf(MSG3, m), 4)
		case jso := <-pool.JSO:
			for client, _ := range pool.ClientMap {
				if client.ID == jso.ID {
					js, y := json.Marshal(jso)
					chke(y)
					er := client.Conn.WriteMessage(websocket.TextMessage, js)
					if er != nil {
						fmt.Println(er)
						return
					}
				}
			}
		}
	}
}

func WSFillNewPool() *WSPool {
	return &WSPool{
		Add:       make(chan *WSClient),
		Remove:    make(chan *WSClient),
		ClientMap: make(map[*WSClient]bool),
		JSO:       make(chan WSJSOut),
		ReadID:    make(chan string),
	}
}
