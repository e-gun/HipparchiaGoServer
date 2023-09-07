//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

// StartEchoServer - start serving; this blocks and does not return while the program remains alive
func StartEchoServer() {
	const (
		LLOGFMT = "r: ${status}\tt: ${latency_human}\tu: ${uri}\n"
		RLOGFMT = "i: ${remote_ip}\t r: ${status}\tt: ${latency_human}\tu: ${uri}\n"
	)

	//
	// SETUP
	//

	e := echo.New()

	if Config.Authenticate {
		// assume that anyone who is using authentication is serving via the internet and so set timeouts
		e.Server.ReadTimeout = TIMEOUTRD
		e.Server.WriteTimeout = TIMEOUTWR

		// also assume that this server is now exposed to scanning attempts that will spam 404s; block IPs that do this
		go IPBlacklistRW()
		go ResponseStatsKeeper()
		e.Use(PoliceResponse)
	}

	switch Config.EchoLog {
	case 3:
		e.Use(middleware.Logger())
	case 2:
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{Format: RLOGFMT}))
	case 1:
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{Format: LLOGFMT}))
	default:
		// do nothing
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
	// [a] authentication ("rt-authentication.go")
	//

	e.POST("/auth/login", RtAuthLogin)
	e.GET("/auth/logout", RtAuthLogout)
	e.GET("/auth/check", RtAuthChkuser)

	//
	// [b] browser ("rt-browsing.go")
	//

	e.GET("/browse/index/:locus", RtBrowseLine)      // '/browse/index/lt0550/001/1855'
	e.GET("/browse/locus/:locus", RtBrowseLocus)     // '/browse/locus/lt0550/001/3|100'
	e.GET("/browse/perseus/:locus", RtBrowsePerseus) // '/browse/perseus/lt0550/001/2:717'
	e.GET("/browse/rawlocus/:locus", RtBrowseRaw)    // '/browse/rawlocus/lt0474/037/2.10.4'

	e.GET("/browse/", RtGetEmptyGet) // dictionary can send an empty string instead of a value

	// [c] css ("rt-embedding.go")

	e.GET("/emb/css/hipparchiastyles.css", RtEmbHCSS)

	// [d] debugging

	// empty

	//
	// [e] frontpage ("rt-frontpage.go")
	//

	e.GET("/", RtFrontpage)

	//
	// [f] getters ("rt-getters.go")
	//

	e.GET("/get/json/sessionvariables", RtGetJSSession)
	e.GET("/get/json/worksof/:id", RtGetJSWorksOf)
	e.GET("/get/json/workstructure/:locus", RtGetJSWorksStruct)
	e.GET("/get/json/samplecitation/:locus", RtGetJSSampCit)
	e.GET("/get/json/authorinfo/:id", RtGetJSAuthorinfo)
	e.GET("/get/json/searchlistcontents", RtGetJSSearchlist)
	e.GET("/get/json/helpdata", RtGetJSHelpdata)

	e.GET("/get/json/worksof/", RtGetEmptyGet) // dictionary can send an empty string instead of a value: "/worksof/:id"
	e.GET("/get/json/workstructure/", RtGetEmptyGet)
	e.GET("/get/json/samplecitation/", RtGetEmptyGet)
	e.GET("/get/json/authorinfo/", RtGetEmptyGet)

	//
	// [g] hinters ("rt-hinters.go")
	//

	e.GET("/hints/author/:null", RtAuthorHints)      // "u: /hints/author/_?term=cic"
	e.GET("/hints/authgenre/:null", RtAuGenreHints)  // "u: /hints/authgenre/_?term=ep"
	e.GET("/hints/workgenre/:null", RtWkGenreHints)  //
	e.GET("/hints/authlocation/:null", RtAuLocHints) //
	e.GET("/hints/worklocation/:null", RtWkLocHints) //
	e.GET("/hints/lemmata/:null", RtLemmaHints)      // "u: /hints/lemmata/_?term=dol"

	//
	// [h] lexical ("rt-lexica.go")
	//

	e.GET("/lex/lookup/:wd", RtLexLookup)         // "u: /lex/lookup/dolor"
	e.GET("/lex/findbyform/:wd", RtLexFindByForm) // "u: /lex/findbyform/sanguis/lt0836"
	e.GET("/lex/reverselookup/:wd", RtLexReverse) // "u: /lex/reverselookup/0ae94619/sorrow"
	e.GET("/lex/idlookup/:wd", RtLexId)           // "u: /lex/idlookup/latin/42534.0"
	e.GET("/lex/chart/:wd", RtMorphchart)         // "u: /lex/chart/latin/14669.0/24366171/dolor"

	//
	// [i] resets ("rt-session.go")
	//

	e.GET("/reset/session", RtResetSession) // "u: /reset/session"

	//
	// [j] searching ("rt-search.go")
	//

	e.GET("/srch/conf/:id", RtSearchConfirm) // "GET /srch/conf/1f8f1d22 HTTP/1.1"
	e.GET("/srch/exec/:id", RtSearch)        // "GET /srch/exec/1f8f1d22?skg=dolor HTTP/1.1"

	//
	// [k] selection ("rt-selection.go")
	//

	e.GET("/selection/make/:locus", RtSelectionMake)   // "GET /selection/make/_?auth=gr7000 HTTP/1.1"
	e.GET("/selection/clear/:locus", RtSelectionClear) // "GET /selection/clear/auselections/0 HTTP/1.1"
	e.GET("/selection/fetch", RtSelectionFetch)        // "GET /selection/fetch HTTP/1.1"

	//
	// [l] set options ("rt-setoptions.go")
	//

	e.GET("/setoption/:opt", RtSetOption) // "u: /setoption/onehit/yes"

	//
	// [m] text and index ("rt-textindixesandvocab.go")
	//

	e.GET("/text/make/:null", RtTextMaker) // "u: /text/make/_"
	e.GET("/text/index/:id", RtIndexMaker) // "u: /text/index/a26ec16c"
	e.GET("/text/vocab/:id", RtVocabMaker) // "u: /text/vocab/ee068d29"

	//
	// [n] websocket ("rt-websocket.go")
	//

	e.GET("/ws", RtWebsocket)

	//
	// [o] serve via the embedded FS ("rt-embedding.go")
	//

	e.GET("/emb/echarts/:file", RtEmbEcharts)
	e.GET("/emb/jq/:file", RtEmbJQuery)
	e.GET("/emb/jq/images/:file", RtEmbJQueryImg)
	e.GET("/emb/js/:file", RtEmbJS)
	e.GET("/emb/otf/:file", RtEmbOTF)
	e.GET("/emb/ttf/:file", RtEmbTTF)
	e.GET("/emb/wof/:file", RtEmbWOF)
	e.GET("/favicon.ico", RtEbmFavicon)
	e.GET("/apple-touch-icon-precomposed.png", RtEbmTouchIcon)
	e.GET("/emb/pdf/:file", RtEmbPDFHelp)

	// [p] cookies ("rt-session.go")

	e.GET("/sc/set/:num", RtSessionSetsCookie)
	e.GET("/sc/get/:num", RtSessionGetCookie)

	// [q] vectors ("vectorqueryneighbors.go")
	// pseudo-route RtVectors in vectorqueryneighbors.go is called by RtSearch() if the current session has VecNNSearch set to true

	e.GET("/vbot/:typeandselection", RtVectorBot) // only the goroutine running the vectorbot is supposed to request this

	// next will do nothing if Config is not requesting these
	go runselftests()
	go activatevectorbot()

	e.HideBanner = true
	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%d", Config.HostIP, Config.HostPort)))
}

//
// SERVERSTATS
//

type EchoResponseStats struct {
	TwoHundred  uint64
	FourOhThree uint64
	FourOhFour  uint64
	FiveHundred uint64
}

type BlackListRD struct {
	key  string
	resp chan bool
}

type BlackListWR struct {
	key  string
	resp chan bool
}

type StatListWR struct {
	key int
	ip  string
	uri string
}

var (
	BListWR         = make(chan BlackListWR)
	BListRD         = make(chan BlackListRD)
	SListWR         = make(chan StatListWR)
	EchoServerStats = NewEchoResponseStats()
)

// PoliceResponse - track response code counts and block repeat 404 offenders
func PoliceResponse(nextechohandler echo.HandlerFunc) echo.HandlerFunc {
	const (
		BLACK0 = `IP address %s was blacklisted: too many previous response code errors`
	)

	return func(c echo.Context) error {
		blcheck := BlackListRD{
			key:  c.RealIP(),
			resp: make(chan bool),
		}

		// presumed guilty
		rscheck := StatListWR{
			key: 403,
			ip:  c.RealIP(),
			uri: c.Request().RequestURI,
		}

		BListRD <- blcheck
		ok := <-blcheck.resp
		if !ok {
			SListWR <- rscheck
			e := echo.NewHTTPError(http.StatusForbidden, fmt.Sprintf(BLACK0, c.RealIP()))
			return e
		} else {
			// do this before setting c.Response().Status or you will always get "200"
			if err := nextechohandler(c); err != nil {
				c.Error(err)
			}
			rscheck.key = c.Response().Status
			SListWR <- rscheck
			return nil
		}
	}
}

func NewEchoResponseStats() *EchoResponseStats {
	return &EchoResponseStats{
		TwoHundred:  0,
		FourOhThree: 0,
		FourOhFour:  0,
		FiveHundred: 0,
	}
}

// IPBlacklistRW - blacklist read/write; should have e
func IPBlacklistRW() {
	const (
		CAP    = 4
		BLACK0 = `IP address %s was blacklisted: too many previous response code errors; %d addresses on the blacklist`
	)

	strikecount := make(map[string]int)
	blacklist := make(map[string]struct{})

	for {
		select {
		case rcv := <-BListRD:
			valid := true
			if _, ok := blacklist[rcv.key]; ok {
				// you are on the blacklist...
				valid = false
			}
			rcv.resp <- valid
		case snd := <-BListWR:
			ret := false
			if _, ok := strikecount[snd.key]; !ok {
				strikecount[snd.key] = 1
			} else if strikecount[snd.key] >= CAP {
				blacklist[snd.key] = struct{}{}
				msg(fmt.Sprintf(BLACK0, snd.key, len(blacklist)), MSGNOTE)
				ret = true
			} else {
				strikecount[snd.key]++
			}
			snd.resp <- ret
		}
	}
}

// ResponseStatsKeeper - log echo responses; should have exclusive r/w access to EchoServerStats
func ResponseStatsKeeper() {
	const (
		BLACK1 = `IP address %s received a strike: StatusNotFound error for URI "%s"`
		BLACK2 = `IP address %s received a strike: StatusInternalServerError for URI "%s"`
		FYI200 = `StatusOK count is %d`
		FRQ200 = 1000
		FYI403 = `StatusForbidden count is %d. Last blocked was %s requesting "%s"`
		FRQ403 = 5
		FYI404 = `StatusNotFound count is %d`
		FRQ404 = 100
		FYI500 = `StatusInternalServerError count is %d.`
		FRQ500 = 10
	)

	for {
		status := <-SListWR
		switch status.key {
		case 200:
			EchoServerStats.TwoHundred++
			if EchoServerStats.TwoHundred%FRQ200 == 0 {
				msg(fmt.Sprintf(FYI200, EchoServerStats.TwoHundred), MSGNOTE)
			}
		case 403:
			// you are already on the blacklist...
			EchoServerStats.FourOhThree++
			if EchoServerStats.FourOhThree%FRQ403 == 0 {
				msg(fmt.Sprintf(FYI403, EchoServerStats.FourOhThree, status.ip, status.uri), MSGNOTE)
			}
		case 404:
			// you need to be registered for the blacklist...
			EchoServerStats.FourOhFour++
			wr := BlackListWR{
				key:  status.ip,
				resp: make(chan bool),
			}

			BListWR <- wr
			ok := <-wr.resp

			if !ok {
				msg(fmt.Sprintf(BLACK1, status.ip, status.uri), MSGWARN)
			}

			if EchoServerStats.FourOhFour%FRQ404 == 0 {
				msg(fmt.Sprintf(FYI404, EchoServerStats.FourOhFour), MSGNOTE)
			}
		case 500:
			EchoServerStats.FiveHundred++
			wr := BlackListWR{
				key:  status.ip,
				resp: make(chan bool),
			}

			BListWR <- wr
			ok := <-wr.resp

			if !ok {
				msg(fmt.Sprintf(BLACK2, status.ip, status.uri), MSGWARN)
			}

			if EchoServerStats.FiveHundred%FRQ500 == 0 {
				msg(fmt.Sprintf(FYI500, EchoServerStats.FiveHundred), MSGWARN)
			}
		default:
			// 302 from "reset" is the only other code...
		}
	}
}
