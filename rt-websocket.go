//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
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

	htm := pd.Msg // see TPM, etc. in searchstructs.go; e.g.: Seeking <span class="sought">»μελιϲϲα«</span>

	// conditionally add "finishing" message
	tp := func() string {
		// textandindex or vectors
		m := ""

		// regular searches
		if pd.SType == "" {
			m = FU
		}

		return fmt.Sprintf(EL1, pd.Elapsed, m)
	}()

	// conditionally add message based on iteration #
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

	if pd.Hits > 0 && pd.SType == "" {
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
	SType     string
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

	getsrchinfo := func() WSSrchInfo {
		responder := WSSIReply{key: c.ID, response: make(chan WSSrchInfo)}
		WSInfo.RequestInfo <- responder
		return <-responder.response
	}

	// wait for the search to exist
	quit := time.Now().Add(time.Second * 1)

	for {
		srchinfo := getsrchinfo()
		if srchinfo.SrchCount != 0 && srchinfo.Exists {
			msg(fmt.Sprintf(SUCCESS, c.ID), MSGTMI)
			break
		}

		if time.Now().After(quit) {
			msg(fmt.Sprintf(FAIL, c.ID), MSGFYI)
			break
		}
	}

	si := getsrchinfo()

	var pd PollData
	pd.SType = si.SType

	// loop until search finishes
	for {
		srchinfo := getsrchinfo()
		if srchinfo.Exists {
			pd.Remain = srchinfo.Remain
			pd.Hits = srchinfo.Hits
			pd.TotalWrk = srchinfo.TableCt
			pd.Msg = strings.Replace(srchinfo.Summary, "Sought", "Seeking", -1)
		} else {
			break
		}

		pd.Elapsed = fmt.Sprintf("%.1fs", time.Now().Sub(srchinfo.Launched).Seconds())

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
// CHANNEL-BASED SEARCHINFO REPORTING TO COMMUNICATE RESULTS BETWEEN ROUTINES: search routes write; websocket reads
//

// WSSrchInfo - struct used to deliver info about searches in progress
type WSSrchInfo struct {
	ID        string
	User      string
	Exists    bool
	Hits      int
	Remain    int
	TableCt   int
	SrchCount int
	VProgStrg string
	Summary   string
	Iteration int
	SType     string
	Launched  time.Time
	RealIP    string
	CancelFnc context.CancelFunc
}

// WSSIKVi - WSSearchInfoHub helper struct for setting an int val on the item at map[key]
type WSSIKVi struct {
	key string
	val int
}

// WSSIKVs - WSSearchInfoHub helper struct for setting a string val on the item at map[key]
type WSSIKVs struct {
	key string
	val string
}

// WSSIReply - WSSearchInfoHub helper struct for returning the WSSrchInfo stored at map[key]
type WSSIReply struct {
	key      string
	response chan WSSrchInfo
}

type WSSICount struct {
	key      string
	response chan int
}

type WSInfoHubInterface struct {
	UpdateHits      chan WSSIKVi
	UpdateRemain    chan WSSIKVi
	UpdateVProgMsg  chan WSSIKVs
	UpdateSummMsg   chan WSSIKVs
	UpdateIteration chan WSSIKVi
	UpdateTW        chan WSSIKVi
	RequestInfo     chan WSSIReply
	InsertInfo      chan WSSrchInfo
	IPSrchCount     chan WSSICount
	Del             chan string
	Reset           chan string
}

// BuildWSInfoHubIf - build the WSInfoHubInterface that will interact with WSSearchInfoHub (one and only one built at app startup)
func BuildWSInfoHubIf() *WSInfoHubInterface {
	return &WSInfoHubInterface{
		UpdateHits:      make(chan WSSIKVi, 2*runtime.NumCPU()),
		UpdateRemain:    make(chan WSSIKVi, 2*runtime.NumCPU()),
		UpdateVProgMsg:  make(chan WSSIKVs, 2*runtime.NumCPU()),
		UpdateSummMsg:   make(chan WSSIKVs, 2*runtime.NumCPU()),
		UpdateIteration: make(chan WSSIKVi, 2*runtime.NumCPU()),
		UpdateTW:        make(chan WSSIKVi),
		RequestInfo:     make(chan WSSIReply),
		InsertInfo:      make(chan WSSrchInfo),
		IPSrchCount:     make(chan WSSICount),
		Del:             make(chan string),
		Reset:           make(chan string),
	}
}

// WSSearchInfoHub - the loop that lets you read/write from/to the various WSSrchInfo channels via the WSInfo global (a *WSInfoHubInterface)
func WSSearchInfoHub() {
	const (
		CANC    = "WSSearchInfoHub() reports that '%s' was cancelled"
		FINWAIT = 10
		FINCHK  = 60
	)

	var (
		Allinfo  = make(map[string]WSSrchInfo)
		Finished = make(map[string]time.Time)
	)

	reporter := func(r WSSIReply) {
		if _, ok := Allinfo[r.key]; ok {
			r.response <- Allinfo[r.key]
		} else {
			// "false" triggers a break in rt-websocket.go
			r.response <- WSSrchInfo{Exists: false}
		}
		// msg(fmt.Sprintf("%d WSSearchInfoHub searches: %s", len(Allinfo), strings.Join(StringMapKeysIntoSlice(Allinfo), ", ")), MSGNOTE)
	}

	fetchifexists := func(id string) WSSrchInfo {
		if _, ok := Allinfo[id]; ok {
			return Allinfo[id]
		} else {
			// any non-zero value for SrchCount is fine; the test in rt-websocket.go is just for 0
			return WSSrchInfo{ID: id, Exists: true, SrchCount: 1}
		}
	}

	ipcount := func(id string) int {
		count := 0
		for _, v := range Allinfo {
			if v.RealIP == id {
				count++
			}
		}
		return count
	}

	// see also the notes at RtResetSession()
	cancelall := func(u string) {
		for _, v := range Allinfo {
			if v.User == u {
				v.CancelFnc()
				msg(fmt.Sprintf(CANC, v.ID), MSGPEEK)
			}
		}
	}

	// this silly mechanism because selftest had 2nd round of nn vector tests respawning after deletion; rare, but...
	storeunlessfinished := func(si WSSrchInfo) {
		if _, ok := Finished[si.ID]; !ok {
			Allinfo[si.ID] = si
		}
	}

	// storeunlessfinished() requires a cleanup function too...
	cleanfinished := func() {
		for {
			for f := range Finished {
				ft := Finished[f]
				later := ft.Add(time.Second * FINWAIT)
				if time.Now().After(later) {
					delete(Finished, f)
				} else {
					fmt.Println(later)
				}
			}
			time.Sleep(time.Second * FINCHK)
		}
	}

	go cleanfinished()

	//UNCOMMENT FOR DEBUGGING BUILDS
	//allinfo := func() {
	//	for {
	//		ai := StringMapKeysIntoSlice(Allinfo)
	//		msg("ai: "+strings.Join(ai, ", "), 2)
	//		for f := range Finished {
	//			msg(f+" is in finished", 2)
	//		}
	//		time.Sleep(1 * time.Second)
	//	}
	//}
	//go allinfo()

	// the main loop; it will never exit
	for {
		select {
		case rq := <-WSInfo.RequestInfo:
			reporter(rq)
		case tw := <-WSInfo.UpdateTW:
			x := fetchifexists(tw.key)
			x.TableCt = tw.val
			storeunlessfinished(x)
		case wr := <-WSInfo.UpdateHits:
			x := fetchifexists(wr.key)
			x.Hits = wr.val
			storeunlessfinished(x)
		case wr := <-WSInfo.UpdateRemain:
			x := fetchifexists(wr.key)
			x.Remain = wr.val
			storeunlessfinished(x)
		case wr := <-WSInfo.UpdateVProgMsg:
			x := fetchifexists(wr.key)
			x.VProgStrg = wr.val
			storeunlessfinished(x)
		case wr := <-WSInfo.UpdateSummMsg:
			x := fetchifexists(wr.key)
			x.Summary = wr.val
			storeunlessfinished(x)
		case wr := <-WSInfo.UpdateIteration:
			x := fetchifexists(wr.key)
			x.Iteration = wr.val
			storeunlessfinished(x)
		case si := <-WSInfo.InsertInfo:
			storeunlessfinished(si)
		case ipc := <-WSInfo.IPSrchCount:
			ipc.response <- ipcount(ipc.key)
		case reset := <-WSInfo.Reset:
			cancelall(reset)
		case del := <-WSInfo.Del:
			Finished[del] = time.Now()
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
