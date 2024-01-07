//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"runtime"
	"strings"
	"time"
)

var (
	Upgrader = websocket.Upgrader{}
)

//
// THE ROUTE
//

// RtWebsocket - progress info for a search (multiple clients client at a time)
func RtWebsocket(c echo.Context) error {
	const (
		FAILCON = "RtWebsocket(): ws connection failed"
	)

	user := readUUIDCookie(c)
	if !AllAuthorized.Check(user) {
		return nil
	}

	ws, err := Upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		msg(FAILCON, MSGNOTE)
		return nil
	}

	progresspoll := &WSClient{
		Conn: ws,
		Pool: WebsocketPool,
	}

	WebsocketPool.Add <- progresspoll
	progresspoll.ReceiveID()
	progresspoll.WSMessageLoop()
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
		GF  = `Formatting the results...&nbsp;`
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

		if pd.IsVect {
			m = ""
		}
		return fmt.Sprintf(EL1, pd.Elapsed, m)
	}()

	it := func() string {
		var m string
		switch pd.Iteration {
		case 2:
			m = MS
		case 3:
			m = GF
		default:
			// no change to m
		}
		return m
	}()

	if pctd != 0 && pd.Remain != 0 && pd.TotalWrk != 0 {
		// normal in progress
		htm += fmt.Sprintf(PCT, pcts, pd.Elapsed)
		htm += it
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
// WEBSOCKET INFRASTRUCTURE: see https://tutorialedge.net/projects/chat-system-in-go-and-react/part-4-handling-multiple-clients/
//

type PollData struct {
	TotalWrk  int    `json:"Poolofwork"`
	Remain    int    `json:"Remaining"`
	Hits      int    `json:"Hitcount"`
	Msg       string `json:"Statusmessage"`
	Elapsed   string `json:"Elapsed"`
	Extra     string `json:"Notes"`
	ID        string `json:"ID"`
	Iteration int
	IsVect    bool
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
	JSO       chan *WSJSOut
	ReadID    chan string
}

type WSJSOut struct {
	V     string `json:"value"`
	ID    string `json:"ID"`
	Close string `json:"close"`
}

// ReceiveID - get the searchID from the client; record it; then exit
func (c *WSClient) ReceiveID() {
	const (
		FAIL1 = `WSClient.ReceiveID() failed`
		FAIL2 = `WSClient.ReceiveID() never received the search id`
	)

	quit := time.Now().Add(time.Second * 1)

	for {
		_, m, err := c.Conn.ReadMessage()
		if err != nil {
			msg(FAIL1, MSGFYI)
			return
		}

		if len(m) != 0 {
			id := string(m)
			id = strings.Replace(id, `"`, "", -1)
			c.ID = id
			c.Pool.ReadID <- id
			break
		}

		if time.Now().After(quit) {
			msg(FAIL2, MSGFYI)
			break
		}
	}
}

// WSMessageLoop - output the constantly updated search progress to the websocket; then exit
func (c *WSClient) WSMessageLoop() {
	const (
		FAIL      = `WSClient.WSMessageLoop() never found '%s' in the SearchMap`
		SUCCESS   = `WSClient.WSMessageLoop() found '%s' in the SearchMap`
		VECAPPEND = `<br><span class="smallerthannormal">%s</span>`
	)

	// wait for the search to exist
	quit := time.Now().Add(time.Second * 1)

	for {
		responder := SIReply{key: c.ID, response: make(chan SrchInfo)}
		SIRequest <- responder
		srchinfo := <-responder.response
		if srchinfo.SrchCount != 0 && srchinfo.Exists {
			msg(fmt.Sprintf(SUCCESS, c.ID), MSGTMI)
			break
		}

		if time.Now().After(quit) {
			msg(fmt.Sprintf(FAIL, c.ID), MSGFYI)
			break
		}
	}

	srch := AllSearches.SimpleGetSS(c.ID)

	var pd PollData
	pd.IsVect = srch.IsVector

	// loop until search finishes
	for {
		responder := SIReply{key: c.ID, response: make(chan SrchInfo)}
		SIRequest <- responder
		srchinfo := <-responder.response

		if srchinfo.Exists {
			pd.Remain = srchinfo.Remain
			pd.Hits = srchinfo.Hits
			pd.TotalWrk = srchinfo.TableCt
			pd.Msg = strings.Replace(srchinfo.Summary, "Sought", "Seeking", -1)
		} else {
			break
		}

		pd.Elapsed = fmt.Sprintf("%.1fs", time.Now().Sub(srch.Launched).Seconds())

		pd.Iteration = srchinfo.Iteration

		if srchinfo.VProgStrg != "" {
			pd.Extra = fmt.Sprintf(VECAPPEND, srchinfo.VProgStrg)
		}

		jso := &WSJSOut{
			V:     formatpoll(pd),
			ID:    c.ID,
			Close: "open",
		}

		c.Pool.JSO <- jso
		time.Sleep(WSPOLLINGPAUSE)
	}
	WebsocketPool.Remove <- c
}

// WSPoolStartListening - the WSPool will listen for activity on its various channels (only called once at app launch)
func (pool *WSPool) WSPoolStartListening() {
	const (
		MSG1 = "Starting polling loop for %s"
		MSG2 = "WSPool client failed on WriteMessage()"
	)

	writemsg := func(jso *WSJSOut) {
		for cl := range pool.ClientMap {
			if cl.ID == jso.ID {
				js, y := json.Marshal(jso)
				chke(y)
				e := cl.Conn.WriteMessage(websocket.TextMessage, js)
				if e != nil {
					msg(MSG2, MSGWARN)
					delete(pool.ClientMap, cl)
				}
			}
		}
	}

	for {
		select {
		case id := <-pool.Add:
			pool.ClientMap[id] = true
		case id := <-pool.Remove:
			delete(pool.ClientMap, id)
		case id := <-pool.ReadID:
			msg(fmt.Sprintf(MSG1, id), MSGPEEK)
		case wrt := <-pool.JSO:
			writemsg(wrt)
		}
	}
}

// WSFillNewPool - build a new WSPool (one and only one built at app startup)
func WSFillNewPool() *WSPool {
	return &WSPool{
		Add:       make(chan *WSClient),
		Remove:    make(chan *WSClient),
		ClientMap: make(map[*WSClient]bool),
		JSO:       make(chan *WSJSOut),
		ReadID:    make(chan string),
	}
}

//
// CHANNEL-BASED SEARCHINFO REPORTING TO COMMUNICATE RESULTS BETWEEN ROUTINES
//

// SrchInfo - struct used to deliver info about searches in progress
type SrchInfo struct {
	ID        string
	Exists    bool
	Hits      int
	Remain    int
	TableCt   int
	SrchCount int
	VProgStrg string
	Summary   string
	Iteration int
}

// SIKVi - SearchInfoHub helper struct for setting an int val on the item at map[key]
type SIKVi struct {
	key string
	val int
}

// SIKVs - SearchInfoHub helper struct for setting a string val on the item at map[key]
type SIKVs struct {
	key string
	val string
}

// SIReply - SearchInfoHub helper struct for returning the SrchInfo stored at map[key]
type SIReply struct {
	key      string
	response chan SrchInfo
}

var (
	SIUpdateHits      = make(chan SIKVi, 2*runtime.NumCPU())
	SIUpdateRemain    = make(chan SIKVi, 2*runtime.NumCPU())
	SIUpdateVProgMsg  = make(chan SIKVs, 2*runtime.NumCPU())
	SIUpdateSummMsg   = make(chan SIKVs, 2*runtime.NumCPU())
	SIUpdateIteration = make(chan SIKVi, 2*runtime.NumCPU())
	SIUpdateTW        = make(chan SIKVi)
	SIRequest         = make(chan SIReply)
	SIDel             = make(chan string)
)

// SearchInfoHub - the loop that lets you read/write from/to the searchinfo channels
func SearchInfoHub() {
	var (
		Allinfo  = make(map[string]SrchInfo)
		Finished = make(map[string]bool)
	)

	reporter := func(r SIReply) {
		if _, ok := Allinfo[r.key]; ok {
			r.response <- Allinfo[r.key]
		} else {
			// "false" triggers a break in rt-websocket.go
			r.response <- SrchInfo{Exists: false}
		}
	}

	fetchifexists := func(id string) SrchInfo {
		if _, ok := Allinfo[id]; ok {
			return Allinfo[id]
		} else {
			// any non-zero value for SrchCount is fine; the test in re-websocket.go is just for 0
			return SrchInfo{ID: id, Exists: true, SrchCount: 1}
		}
	}

	// this silly mechanism because selftest had 2nd round of nn vector tests respawning after deletion; rare, but...
	storeunlessfinished := func(si SrchInfo) {
		if _, ok := Finished[si.ID]; !ok {
			Allinfo[si.ID] = si
		}
	}

	// the main loop; it will never exit
	for {
		select {
		case rq := <-SIRequest:
			reporter(rq)
		case tw := <-SIUpdateTW:
			x := fetchifexists(tw.key)
			x.TableCt = tw.val
			storeunlessfinished(x)
		case wr := <-SIUpdateHits:
			x := fetchifexists(wr.key)
			x.Hits = wr.val
			storeunlessfinished(x)
		case wr := <-SIUpdateRemain:
			x := fetchifexists(wr.key)
			x.Remain = wr.val
			storeunlessfinished(x)
		case wr := <-SIUpdateVProgMsg:
			x := fetchifexists(wr.key)
			x.VProgStrg = wr.val
			storeunlessfinished(x)
		case wr := <-SIUpdateSummMsg:
			x := fetchifexists(wr.key)
			x.Summary = wr.val
			storeunlessfinished(x)
		case wr := <-SIUpdateIteration:
			x := fetchifexists(wr.key)
			x.Iteration = wr.val
			storeunlessfinished(x)
		case del := <-SIDel:
			Finished[del] = true
			delete(Allinfo, del)
		}
	}
}

//
// FOR DEBUGGING ONLY
//

// wsclientreport - report the # and names of the active wsclients every N seconds
func wsclientreport(d time.Duration) {
	// add the following to main.go: "go wsclientreport()"
	for {
		cl := WebsocketPool.ClientMap
		var cc []string
		for k := range cl {
			cc = append(cc, k.ID)
		}
		msg(fmt.Sprintf("%d WebsocketPool clients: %s", len(cl), strings.Join(cc, ", ")), MSGNOTE)
		time.Sleep(d)
	}
}
