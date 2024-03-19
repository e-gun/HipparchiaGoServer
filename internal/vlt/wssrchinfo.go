//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vlt

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"
)

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

// WSSIKVi - WSSearchInfoHub helper struct for setting an int Val on the item at map[Key]
type WSSIKVi struct {
	Key string
	Val int
}

// WSSIKVs - WSSearchInfoHub helper struct for setting a string Val on the item at map[Key]
type WSSIKVs struct {
	Key string
	Val string
}

// WSSIReply - WSSearchInfoHub helper struct for returning the WSSrchInfo stored at map[Key]
type WSSIReply struct {
	Key      string
	Response chan WSSrchInfo
}

type WSSICount struct {
	Key      string
	Response chan int
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
		if _, ok := Allinfo[r.Key]; ok {
			r.Response <- Allinfo[r.Key]
		} else {
			// "false" triggers a break in rt-websocket.go
			r.Response <- WSSrchInfo{Exists: false}
		}
		// mm(fmt.Sprintf("%d WSSearchInfoHub searches: %s", len(Allinfo), strings.Join(StringMapKeysIntoSlice(Allinfo), ", ")), MSGNOTE)
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
				Msg.PEEK(fmt.Sprintf(CANC, v.ID))
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
	//		mm("ai: "+strings.Join(ai, ", "), 2)
	//		for f := range Finished {
	//			mm(f+" is in finished", 2)
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
			x := fetchifexists(tw.Key)
			x.TableCt = tw.Val
			storeunlessfinished(x)
		case wr := <-WSInfo.UpdateHits:
			x := fetchifexists(wr.Key)
			x.Hits = wr.Val
			storeunlessfinished(x)
		case wr := <-WSInfo.UpdateRemain:
			x := fetchifexists(wr.Key)
			x.Remain = wr.Val
			storeunlessfinished(x)
		case wr := <-WSInfo.UpdateVProgMsg:
			x := fetchifexists(wr.Key)
			x.VProgStrg = wr.Val
			storeunlessfinished(x)
		case wr := <-WSInfo.UpdateSummMsg:
			x := fetchifexists(wr.Key)
			x.Summary = wr.Val
			storeunlessfinished(x)
		case wr := <-WSInfo.UpdateIteration:
			x := fetchifexists(wr.Key)
			x.Iteration = wr.Val
			storeunlessfinished(x)
		case si := <-WSInfo.InsertInfo:
			storeunlessfinished(si)
		case ipc := <-WSInfo.IPSrchCount:
			ipc.Response <- ipcount(ipc.Key)
		case reset := <-WSInfo.Reset:
			cancelall(reset)
		case del := <-WSInfo.Del:
			Finished[del] = time.Now()
			delete(Allinfo, del)
		}
	}
}

func WSFetchSrchInfo(id string) WSSrchInfo {
	responder := WSSIReply{Key: id, Response: make(chan WSSrchInfo)}
	WSInfo.RequestInfo <- responder
	return <-responder.Response
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
		Msg.NOTE(fmt.Sprintf("%d WebsocketPool clients: %s", len(cl), strings.Join(cc, ", ")))
		time.Sleep(d)
	}
}
