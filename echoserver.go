//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"html/template"
	"io"
	"net/http"
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

// StartEchoServer - start serving; this blocks and does not return while the program remains alive
func StartEchoServer() {
	// https://echo.labstack.com/guide/
	// cf https://medium.com/cuddle-ai/building-microservice-using-golang-echo-framework-ff10ba06d508

	//
	// SETUP
	//

	e := echo.New()

	fp, err := efs.ReadFile("emb/frontpage.html")
	chke(err)

	fpt, err := template.New("fp").Parse(string(fp))
	chke(err)

	renderer := &TemplateRenderer{
		templates: fpt,
	}
	e.Renderer = renderer

	if Config.EchoLog == 2 {
		e.Use(middleware.Logger())
	} else if Config.EchoLog == 1 {
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{Format: "r: ${status}\tt: ${latency_human}\tu: ${uri}\n"}))
	}

	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(MAXECHOREQPERSECONDPERIP)))

	e.Use(middleware.Recover())

	if Config.Gzip {
		e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 5}))
	}

	//
	// HIPPARCHIA ROUTES
	//

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

	// [b1] sample input: '/browse/index/lt0550/001/1855'
	e.GET("/browse/index/:locus", RtBrowseline)

	// [b2] sample input: '/browse/locus/lt0550/001/3|100'
	e.GET("/browse/locus/:locus", RtBrowseLocus)

	// [b3] sample input: '/browse/perseus/lt0550/001/2:717'
	e.GET("/browse/perseus/:locus", RtBrowsePerseus)

	// [b4] sample input: '/browse/rawlocus/lt0474/037/2.10.4'
	e.GET("/browse/rawlocus/:locus", RtBrowseRaw)

	// [c] css

	e.GET("/emb/css/hipparchiastyles.css", RtEmbHCSS)

	// [d] debugging

	// empty

	//
	// [e] frontpage
	//

	e.GET("/", RtFrontpage)

	//
	// [f] getters
	//

	// [f1a] /get/response/cookie [unneeded/unimplemented ATM]

	// [f1b] /get/response/vectorfigure
	// [f2a] /get/json/sessionvariables
	e.GET("/get/json/sessionvariables", RtGetJSSession)

	// [f2b] /get/json/worksof
	e.GET("/get/json/worksof/:id", RtGetJSWorksOf)

	// [f2c] /get/json/workstructure
	e.GET("/get/json/workstructure/:locus", RtGetJSWorksStruct)

	// [f2d] /get/json/samplecitation
	e.GET("/get/json/samplecitation/:locus", RtGetJSSampCit)

	// [f2e] /get/json/authorinfo

	e.GET("/get/json/authorinfo/:id", RtGetJSAuthorinfo)

	// [f2f] /get/json/searchlistcontents

	e.GET("/get/json/searchlistcontents", RtGetJSSearchlist)

	// [f2e] /get/json/genrelistcontents [unneeded/unimplemented ATM]
	// [f2f] /get/json/vectorranges
	// [f2g] /get/json/helpdata
	e.GET("/get/json/helpdata", RtGetJSHelpdata)

	//
	// [g] hinters
	//

	// [g1] "GET /hints/author/_?term=au HTTP/1.1"
	e.GET("/hints/author/:null", RtAuthorHints)

	// [g2] authorgenre
	e.GET("/hints/authgenre/:null", RtAuGenreHints)
	// [g3] workgenre
	e.GET("/hints/workgenre/:null", RtWkGenreHints)

	// [g4] authorlocation
	e.GET("/hints/authlocation/:null", RtAuLocHints)

	// [g5] worklocation
	e.GET("/hints/worklocation/:null", RtWkLocHints)

	// [g6] lemmata: "GET http://localhost:8000/hints/lemmata/_?term=dol"
	e.GET("/hints/lemmata/:null", RtLemmaHints)

	//
	// [h] lexical
	//

	// [h1] uri: /lex/lookup/dolor
	e.GET("/lex/lookup/:wd", RtLexLookup)

	// [h2] GET http://localhost:8000/lex/findbyform/sapientem/lt0474
	e.GET("/lex/findbyform/:wd", RtLexFindByForm)

	// [h3] uri: /lex/reverselookup/0ae94619/sorrow
	e.GET("/lex/reverselookup/:wd", RtLexReverse)

	// [h4] http://127.0.0.1:8000/lex/idlookup/latin/24236.0
	e.GET("/lex/idlookup/:wd", RtLexId)

	// [h5] /lex/morphologychart/greek/39046.0/37925260/ἐπιγιγνώϲκω
	e.GET("/lex/chart/:wd", RtMorphchart)

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
	e.GET("/srch/conf/:id", RtSearchConfirm)

	// [j2] standard: "GET /search/standard/1f8f1d22?skg=dolor HTTP/1.1"
	e.GET("/srch/exec/:id", RtSearch)

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

	e.GET("/setoption/:opt", RtSetOption) // located below

	//
	// [m] text and index
	//

	// [m1] "/text/make/_"

	e.GET("/text/make/:null", RtTextMaker)

	// [m2] "/text/index/a26ec16c"

	e.GET("/text/index/:id", RtIndexMaker)

	// [m3] "http://localhost:5000/text/vocab_rawloc/9f9a0e80/lt0474/002/20
	e.GET("/text/vocab/:id", RtVocabMaker)

	//
	// [n] vectors [unneeded/unimplemented ATM]
	//

	//
	// [o] websocket
	//

	e.GET("/ws", RtWebsocket)

	//
	// [p] serve via the embedded FS
	//

	e.GET("/emb/jq/:file", RtEmbJQuery)
	e.GET("/emb/jq/images/:file", RtEmbJQueryImg)
	e.GET("/emb/js/:file", RtEmbJS)
	e.GET("/emb/ttf/:file", RtEmbTTF)
	e.GET("/favicon.ico", RtEbmFavicon)
	e.GET("/apple-touch-icon-precomposed.png", RtEbmTouchIcon)

	// [q] cookies

	// [q1] set
	e.GET("/sc/set/:num", RtSessionSetsCookie)

	// [q2] get
	e.GET("/sc/get/:num", RtSessionGetCookie)

	e.HideBanner = true
	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%d", Config.HostIP, Config.HostPort)))
}

//
// MISC SIMPLE ROUTES
//

// RtAuthChkuser - placeholder for authentication code [post v.1.0.0]
func RtAuthChkuser(c echo.Context) error {
	// currently unused
	return c.String(http.StatusOK, "")
}

// RtSessionSetsCookie - turn the session into a cookie
func RtSessionSetsCookie(c echo.Context) error {
	const (
		FAIL = "RtSessionSetsCookie() could not marshal the session"
	)
	num := c.Param("num")
	user := readUUIDCookie(c)
	s := SafeSessionRead(user)

	v, e := json.Marshal(s)
	if e != nil {
		v = []byte{}
		msg(FAIL, 1)
	}
	swap := strings.NewReplacer(`"`, "%22", ",", "%2C", " ", "%20")
	vs := swap.Replace(string(v))

	// note that cookie.Path = "/" is essential; otherwise different cookies for different contexts: "/browse" vs "/"
	cookie := new(http.Cookie)
	cookie.Name = "session" + num
	cookie.Path = "/"
	cookie.Value = vs
	cookie.Expires = time.Now().Add(4800 * time.Hour)
	c.SetCookie(cookie)

	return c.JSONPretty(http.StatusOK, "", JSONINDENT)
}

// RtSessionGetCookie - turn a stored cookie into a session
func RtSessionGetCookie(c echo.Context) error {
	// this code has input trust issues...
	const (
		FAIL1 = "RtSessionGetCookie failed to read cookie %s for %s"
		FAIL2 = "RtSessionGetCookie failed to unmarshal cookie %s for %s"
	)

	user := readUUIDCookie(c)
	num := c.Param("num")
	cookie, err := c.Cookie("session" + num)
	if err != nil {
		msg(fmt.Sprintf(FAIL1, num, user), 1)
		return c.String(http.StatusOK, "")
	}

	var s ServerSession
	// invalid character '%' looking for beginning of object key string:
	// {%22ID%22:%22723073ae-09a7-4b24-a5d6-7e20603d8c44%22%2C%22IsLoggedIn%22:true%2C...}
	swap := strings.NewReplacer("%22", `"`, "%2C", ",", "%20", " ")
	cv := swap.Replace(cookie.Value)

	err = json.Unmarshal([]byte(cv), &s)
	if err != nil {
		// invalid character '%' looking for beginning of object key string
		msg(fmt.Sprintf(FAIL2, num, user), 1)
		fmt.Println(err)
		return c.String(http.StatusOK, "")
	}

	SafeSessionMapInsert(s)

	e := c.Redirect(http.StatusFound, "/")
	chke(e)
	return nil
}

// RtResetSession - delete and then reset the session
func RtResetSession(c echo.Context) error {
	user := readUUIDCookie(c)

	MapLocker.Lock()
	delete(SessionMap, user)
	MapLocker.Unlock()

	// then reset it
	readUUIDCookie(c)
	e := c.Redirect(http.StatusFound, "/")
	chke(e)
	return nil
}

// RtSetOption - modify the session in light of the selection made
func RtSetOption(c echo.Context) error {
	const (
		FAIL1 = "RtSetOption() was given bad input: %s"
		FAIL2 = "RtSetOption() hit an impossible case"
	)
	user := readUUIDCookie(c)
	optandval := c.Param("opt")
	parsed := strings.Split(optandval, "/")

	if len(parsed) != 2 {
		msg(fmt.Sprintf(FAIL1, optandval), 1)
		return c.String(http.StatusOK, "")
	}

	opt := parsed[0]
	val := parsed[1]

	ynoptionlist := []string{"greekcorpus", "latincorpus", "papyruscorpus", "inscriptioncorpus", "christiancorpus",
		"rawinputstyle", "onehit", "headwordindexing", "indexbyfrequency", "spuria", "incerta", "varia"}

	s := SafeSessionRead(user)

	if isinslice(ynoptionlist, opt) {
		valid := []string{"yes", "no"}
		if isinslice(valid, val) {
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
				msg(FAIL2, 1)
			}
		}
	}

	valoptionlist := []string{"nearornot", "searchscope", "sortorder"}
	if isinslice(valoptionlist, opt) {
		switch opt {
		case "nearornot":
			valid := []string{"near", "notnear"}
			if isinslice(valid, val) {
				s.NearOrNot = val
			}
		case "searchscope":
			valid := []string{"lines", "words"}
			if isinslice(valid, val) {
				s.SearchScope = val
			}
		case "sortorder":
			valid := []string{"shortname", "converted_date", "provenance", "universalid"}
			if isinslice(valid, val) {
				s.SortHitsBy = val
			}
		default:
			msg(FAIL2, 1)
		}
	}

	spinoptionlist := []string{"maxresults", "linesofcontext", "browsercontext", "proximity"}
	if isinslice(spinoptionlist, opt) {
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
					s.BrowseCtx = int64(intval)
				} else {
					s.BrowseCtx = MAXBROWSERCONTEXT
				}
			case "proximity":
				if intval <= MAXDISTANCE {
					s.Proximity = intval
				} else {
					s.HitLimit = MAXHITLIMIT
				}
			default:
				msg(FAIL2, 1)
			}
		}
	}

	dateoptionlist := []string{"earliestdate", "latestdate"}
	if isinslice(dateoptionlist, opt) {
		intval, e := strconv.Atoi(val)
		if e == nil {
			switch opt {
			case "earliestdate":
				if intval > MAXDATE {
					s.Earliest = fmt.Sprintf("%d", MAXDATE)
				} else if intval < MINDATE {
					s.Earliest = fmt.Sprintf("%d", MINDATE)
				} else {
					s.Earliest = val
				}
			case "latestdate":
				if intval > MAXDATE {
					s.Latest = fmt.Sprintf("%d", MAXDATE)
				} else if intval < MINDATE {
					s.Latest = fmt.Sprintf("%d", MINDATE)
				} else {
					s.Latest = val
				}
			default:
				msg(FAIL2, 1)
			}
		}

		ee, e1 := strconv.Atoi(s.Earliest)
		ll, e2 := strconv.Atoi(s.Latest)
		if e1 != nil {
			s.Earliest = MINDATESTR
		}
		if e2 != nil {
			s.Latest = MAXDATESTR
		}
		if e1 == nil && e2 == nil {
			if ee > ll {
				s.Earliest = s.Latest
			}
		}
	}

	MapLocker.Lock()
	delete(SessionMap, user)
	SessionMap[user] = s
	MapLocker.Unlock()

	return c.String(http.StatusOK, "")
}
