//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vlt

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coder/websocket"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"strings"
	"time"
)

var Msg = lnch.NewMessageMakerWithDefaults()

//
// WEBSOCKET INFRASTRUCTURE: see https://tutorialedge.net/projects/chat-system-in-go-and-react/part-4-handling-multiple-clients/
//

type WSClient struct {
	ID   string
	Conn *websocket.Conn
	Pool *wspool
}

type polldata struct {
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

type wspool struct {
	Add       chan *WSClient
	Remove    chan *WSClient
	ClientMap map[*WSClient]bool
	JSO       chan *wsjsout
	ReadID    chan string
}

type wsjsout struct {
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

	ctx := context.Background()
	defer ctx.Done()

	for {
		_, m, err := c.Conn.Read(ctx)
		if err != nil {
			Msg.FYI(FAIL1)
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
			Msg.FYI(FAIL2)
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
		responder := WSSIReply{Key: c.ID, Response: make(chan WSSrchInfo)}
		WSInfo.RequestInfo <- responder
		return <-responder.Response
	}

	// wait for the search to exist; then run the message loop

	// [a] wait
	maxwait := time.Now().Add(time.Millisecond * 50)

	for {
		srchinfo := getsrchinfo()
		if srchinfo.SrchCount != 0 && srchinfo.Exists {
			Msg.PEEK(fmt.Sprintf(SUCCESS, c.ID))
			break
		}

		if time.Now().After(maxwait) {
			// only very fast searches that finish in a tiny fraction of a second  (c. 0.01s) should hit maxwait
			Msg.FYI(fmt.Sprintf(FAIL, c.ID))
			break
		}
	}

	si := getsrchinfo()

	var pd polldata
	pd.SType = si.SType

	// [b] loop until search finishes (but it might already be over...)
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

		jso := &wsjsout{
			V:     formatpoll(pd),
			ID:    c.ID,
			Close: "open",
		}

		c.Pool.JSO <- jso
		time.Sleep(vv.WSPOLLINGPAUSE)
	}
	WebsocketPool.Remove <- c
}

// WSPoolStartListening - the wspool will listen for activity on its various channels (only called once at app vv)
func (pool *wspool) WSPoolStartListening() {
	const (
		MSG1 = "Starting polling loop for %s"
		MSG2 = "WSPoolStartListening(): wspool client %s failed on WriteMessage()"
	)

	ctx := context.Background()
	defer ctx.Done()

	writemsg := func(jso *wsjsout) {
		for cl := range pool.ClientMap {
			if cl.ID == jso.ID {
				js, y := json.Marshal(jso)
				Msg.EC(y)
				e := cl.Conn.Write(ctx, websocket.MessageText, js)
				if e != nil {
					Msg.WARN(fmt.Sprintf(MSG2, jso.ID))
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
			Msg.PEEK(fmt.Sprintf(MSG1, id))
		case wrt := <-pool.JSO:
			writemsg(wrt)
		}
	}
}

// wsfillnewpool - build a new wspool (one and only one built at app startup)
func wsfillnewpool() *wspool {
	return &wspool{
		Add:       make(chan *WSClient),
		Remove:    make(chan *WSClient),
		ClientMap: make(map[*WSClient]bool),
		JSO:       make(chan *wsjsout),
		ReadID:    make(chan string),
	}
}

// formatpoll - build HTML to send to the JS on the other side
func formatpoll(pd polldata) string {
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

	htm := pd.Msg // see TPM, etc. in searchstruct.go; e.g.: Seeking <span class="sought">»μελιϲϲα«</span>

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
			// no change to mm
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
