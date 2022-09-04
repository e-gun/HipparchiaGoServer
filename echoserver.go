package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {

	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

// JSStruct - this is for generating a specific brand of JSON
type JSStruct struct {
	V string `json:"value"`
}

func StartEchoServer() {
	// https://echo.labstack.com/guide/
	// cf https://medium.com/cuddle-ai/building-microservice-using-golang-echo-framework-ff10ba06d508

	e := echo.New()

	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("static/html/frontpage.html")),
	}
	e.Renderer = renderer

	if cfg.EchoLog == 2 {
		e.Use(middleware.Logger())
	} else if cfg.EchoLog == 1 {
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{Format: "status: ${status}\turi: ${uri}\n"}))
	}

	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 5}))

	e.File("/favicon.ico", "static/images/hipparchia_favicon.ico")
	e.Static("/static", "static")

	// hipparchia routes

	//
	// [a] authentication
	//

	// [a1] '/authentication/attemptlogin'
	// [a2] '/authentication/logout'
	// [a3] '/authentication/checkuser'
	e.GET("/authentication/checkuser", RtAuthChkuser)

	//
	// [b] browser
	//

	// [b1] sample input: '/browse/linenumber/lt0550/001/1855'
	e.GET("/browse/linenumber/:locus", RtBrowseline)

	// [b2] sample input: '/browse/locus/lt0550/001/3|100'
	e.GET("/browse/locus/:locus", RtBrowseLocus)

	// [b3] sample input: '/browse/perseus/lt0550/001/2:717'
	e.GET("/browse/perseus/:locus", RtBrowsePerseus)

	// [b4] sample input: '/browse/rawlocus/lt0474/037/2.10.4'

	// [c] css
	// [d] debugging

	//
	// [e] frontpage
	//

	e.GET("/", RtFrontpage)

	//
	// [f] getters
	//

	// [f1a] /get/response/cookie
	// [f1b] /get/response/vectorfigure
	// [f2a] /get/json/sessionvariables
	e.GET("/get/json/sessionvariables", RtGetJSSession)

	// [f2b] /get/json/worksof
	e.GET("/get/json/worksof/:id", RtGetJSWorksOf)

	// [f2c] /get/json/workstructure
	e.GET("/get/json/workstructure/:locus", RtGetJSWorksStruct)

	// [f2d] /get/json/samplecitation
	// [f2e] /get/json/authorinfo
	// [f2f] /get/json/searchlistcontents
	// [f2e] /get/json/genrelistcontents
	// [f2f] /get/json/vectorranges
	// [f2g] /get/json/helpdata
	e.GET("/get/json/helpdata", RtGetJSHelpdata)

	//
	// [g] hinters
	//

	// [g1] "GET /hints/author/_?term=au HTTP/1.1"
	e.GET("/hints/author/:id", RtAuthorHints)

	// [g2] authorgenre
	// [g3] workgenre
	// [g4] authorlocation
	// [g5] worklocation
	// [g6] lemmata: "GET http://localhost:8000/hints/lemmata/_?term=dol"
	e.GET("/hints/lemmata/:id", RtLemmaHints)

	//
	// [h] lexical
	//

	// [h1]
	// [h2] GET http://localhost:8000/lexica/findbyform/sapientem/lt0474
	e.GET("/lexica/findbyform/:id", RtLexFindByForm)

	// [h3]
	// [h4]

	//
	// [i] resets
	//

	// [i1] /reset/session
	e.GET("/reset/session", RtResetSession)
	// [i2] /reset/vectors
	// [i3] /reset/vectorimages

	//
	// [j] searching
	//

	// [j1] confirm: "GET /search/confirm/1f8f1d22 HTTP/1.1"
	e.GET("/search/confirm/:id", RtSearchConfirm)

	// [j2] standard: "GET /search/standard/1f8f1d22?skg=dolor HTTP/1.1"
	e.GET("/search/standard/:id", RtSearchStandard)

	// [j3] singleword
	// [j4] lemmatized

	//
	// [k] selection
	//

	// [k1] "GET /selection/make/_?auth=gr7000 HTTP/1.1"
	e.GET("/selection/make/:locus", RtSelectionMake)

	// [k2] "GET /selection/clear/auselections/0 HTTP/1.1"
	e.GET("/selection/clear/:locus", RtSelectionClear)

	// [k3] "GET /selection/fetch HTTP/1.1"
	e.GET("/selection/fetch", RtSelectionFetch)

	//
	// [l] setoption: http://localhost:8000/setoption/greekcorpus/yes
	//

	e.GET("/setoption/:opt", RtSetOption)

	//
	// [m] text and index
	//

	//
	// [n] vectors [unneeded/unimplemented ATM]
	//

	//
	// [o] websocket
	//

	e.GET("/ws", RtWebsocket)

	// [z] testing
	e.GET("/t", RtTest)

	e.Logger.Fatal(e.Start(":8000"))
}

func RtAuthChkuser(c echo.Context) error {
	return c.String(http.StatusOK, "")
}

func RtResetSession(c echo.Context) error {
	// delete my session
	delete(sessions, readUUIDCookie(c))

	// then reset it
	readUUIDCookie(c)
	return RtFrontpage(c)
}

func RtSetOption(c echo.Context) error {
	optandval := c.Param("opt")
	parsed := strings.Split(optandval, "/")

	if len(parsed) != 2 {
		msg(fmt.Sprintf("RtSetOption() was given bad input: %s", optandval), 1)
		return c.String(http.StatusOK, "")
	}

	opt := parsed[0]
	val := parsed[1]

	ynoptionlist := []string{"greekcorpus", "latincorpus", "papyruscorpus", "inscriptioncorpus", "christiancorpus",
		"rawinputstyle", "onehit", "headwordindexing", "indexbyfrequency", "spuria", "incerta", "varia"}

	s := sessions[readUUIDCookie(c)]

	if contains(ynoptionlist, opt) {
		valid := []string{"yes", "no"}
		if contains(valid, val) {
			var b bool
			if val == "yes" {
				b = true
			} else {
				b = false
			}
			switch opt {
			case "greekcorpus":
				s.ActiveCorp["gr"] = b
			case "latincorpus":
				s.ActiveCorp["lt"] = b
			case "papyruscorpus":
				s.ActiveCorp["dp"] = b
			case "inscriptioncorpus":
				s.ActiveCorp["in"] = b
			case "christiancorpus":
				s.ActiveCorp["ch"] = b
			case "rawinputstyle":
				s.RawInput = b
			case "onehit":
				s.OneHit = b
			case "indexbyfrequency":
				s.FrqIdx = b
			case "headwordindexing":
				s.HeadwordIdx = b
			case "spuria":
				s.SpuriaOK = b
			case "incerta":
				s.IncertaOK = b
			case "varia":
				s.VariaOK = b
			default:
				msg("RtSetOption() hit an impossible case", 1)
			}
		}
	}

	valoptionlist := []string{"nearornot", "searchscope", "sortorder"}
	if contains(valoptionlist, opt) {
		switch opt {
		case "nearornot":
			valid := []string{"near", "notnear"}
			if contains(valid, val) {
				s.NearOrNot = val
			}
		case "searchscope":
			valid := []string{"lines", "words"}
			if contains(valid, val) {
				s.SearchScope = val
			}
		case "sortorder":
			// unhandled are "location" & "provenance": see goroutinesearcher.go
			valid := []string{"shortname", "converted_date", "location", "provenance", "universalid"}
			if contains(valid, val) {
				s.SortHitsBy = val
			}
		default:
			msg("RtSetOption() hit an impossible case", 1)
		}
	}

	spinoptionlist := []string{"maxresults", "linesofcontext", "browsercontext"}
	if contains(spinoptionlist, opt) {
		intval, e := strconv.Atoi(val)
		if e == nil {
			switch opt {
			case "maxresults":
				if intval < MAXHITLIMIT {
					s.HitLimit = int64(intval)
				} else {
					s.HitLimit = MAXHITLIMIT
				}
			case "linesofcontext":
				if intval < MAXLINESHITCONTEXT {
					s.HitContext = intval
				} else {
					s.HitContext = intval
				}
			case "browsercontext":
				if intval < MAXBROWSERCONTEXT {
					s.UI.BrowseCtx = int64(intval)
				} else {
					s.UI.BrowseCtx = MAXBROWSERCONTEXT
				}
			default:
				msg("RtSetOption() hit an impossible case", 1)
			}
		}
	}

	delete(sessions, readUUIDCookie(c))
	sessions[readUUIDCookie(c)] = s

	return c.String(http.StatusOK, "")
}

func RtTest(c echo.Context) error {
	a := len(AllAuthors)
	s := fmt.Sprintf("%d authors present", a)
	return c.String(http.StatusOK, s)
}

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
