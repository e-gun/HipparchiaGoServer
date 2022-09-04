package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"net"
	"os"
	"strconv"
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

	for {
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
			// if you insist on a full poll; but the thing works without input from pp and rc
			for {
				_, pp := os.Stat(fmt.Sprintf("%s/hgs_pp_%s", UNIXSOCKETPATH, searches[bs].ID))
				_, rc := os.Stat(fmt.Sprintf("%s/hgs_rc_%s", UNIXSOCKETPATH, searches[bs].ID))
				if pp == nil && rc == nil {
					// msg("found both search activity sockets", 5)
					break
				}
				if _, ok := searches[bs]; !ok {
					// don't wait forever
					break
				}
			}

			// we will grab the remainder value via unix socket
			rsock := false
			rconn, err := net.Dial("unix", fmt.Sprintf("%s/hgs_pp_%s", UNIXSOCKETPATH, searches[bs].ID))
			if err != nil {
				msg(fmt.Sprintf("RtWebsocket() has no connection to the remainder reports: %s/hgs_pp_%s", UNIXSOCKETPATH, searches[bs].ID), 1)
			} else {
				rsock = true
				defer rconn.Close()
			}

			// we will grab the hits value via unix socket
			hsock := false
			hconn, err := net.Dial("unix", fmt.Sprintf("%s/hgs_rc_%s", UNIXSOCKETPATH, searches[bs].ID))
			if err != nil {
				msg(fmt.Sprintf("RtWebsocket() has no connection to the hits reports: %s/hgs_rc_%s", UNIXSOCKETPATH, searches[bs].ID), 1)
			} else {
				// if there is no connection you will get a null pointer dereference
				hsock = true
				defer hconn.Close()
			}

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

				// [b] the tricky info
				// [b1] set r.Remain via TCP connection to SrchFeeder()'s broadcaster
				if rsock {
					r.Remain = func() int {
						connbuf := bufio.NewReader(rconn)
						for {
							rs, err := connbuf.ReadString('\n')
							if err != nil {
								break
							} else {
								// fmt.Println([]byte(rs)) --> [49 10]
								// and stripping the newline via strings is not working
								rr := []rune(rs)
								rr = rr[0 : len(rr)-1]
								rs, _ := strconv.Atoi(string(rr))
								return rs
							}
						}
						return -1
					}()
				} else {
					// see the JS: this turns off progress displays
					r.TotalWrk = -1
				}

				if r.Remain != 0 {
					r.Msg = mm
				} else if rsock {
					// will be zero if you never made the connection
					r.Msg = "Formatting the finds..."
				}

				// [b2] set r.Hits via TCP connection to ResultCollation()'s broadcaster
				if hsock {
					r.Hits = func() int {
						connbuf := bufio.NewReader(hconn)
						for {
							ht, err := connbuf.ReadString('\n')
							if err != nil {
								break
							} else {
								// fmt.Println([]byte(rs)) --> [49 10]
								// and stripping the newline via strings is not working
								hh := []rune(ht)
								hh = hh[0 : len(hh)-1]
								h, _ := strconv.Atoi(string(hh))
								return h
							}
						}
						return 0
					}()
				}

				// Write
				js, y := json.Marshal(r)
				chke(y)

				er := ws.WriteMessage(websocket.TextMessage, js)

				if er != nil {
					c.Logger().Error(er)
					msg("RtWebsocket(): ws failed to write: breaking", 1)
					if hsock {
						hconn.Close()
					}
					if rsock {
						rconn.Close()
					}
					break
				}

				if _, exists := searches[bs]; !exists {
					if hsock {
						hconn.Close()
					}
					if rsock {
						rconn.Close()
					}
					break
				}
			}
		}
	}
	return nil
}
