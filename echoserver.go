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
	"sync"
)

var (
	EchoServerStats = NewEchoResponseStats()
)

// StartEchoServer - start serving; this blocks and does not return while the program remains alive
func StartEchoServer() {
	// https://echo.labstack.com/guide/
	// cf https://medium.com/cuddle-ai/building-microservice-using-golang-echo-framework-ff10ba06d508

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
		e.Use(EchoServerStats.PoliceResponse)
	}

	if Config.EchoLog == 3 {
		e.Use(middleware.Logger())
	} else if Config.EchoLog == 2 {
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{Format: RLOGFMT}))
	} else if Config.EchoLog == 1 {
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{Format: LLOGFMT}))
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

	// [f1b] /get/response/vectorfigure
	// [f2e] /get/json/genrelistcontents [unneeded/unimplemented ATM]
	// [f2f] /get/json/vectorranges
	e.GET("/get/json/sessionvariables", RtGetJSSession)
	e.GET("/get/json/worksof/:id", RtGetJSWorksOf)
	e.GET("/get/json/workstructure/:locus", RtGetJSWorksStruct)
	e.GET("/get/json/samplecitation/:locus", RtGetJSSampCit)
	e.GET("/get/json/authorinfo/:id", RtGetJSAuthorinfo)
	e.GET("/get/json/searchlistcontents", RtGetJSSearchlist)
	e.GET("/get/json/helpdata", RtGetJSHelpdata)

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
	// [i2] /reset/vectors
	// [i3] /reset/vectorimages

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

	if Config.SelfTest > 0 {
		go func() {
			for i := 0; i < Config.SelfTest; i++ {
				msg(fmt.Sprintf("Running Selftest %d of %d", i+1, Config.SelfTest), 0)
				selftest()
			}
		}()
	}

	if Config.VectorBot {
		go activatevectorbot()
	}

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
	Scanners    map[string]int
	Hackers     map[string]int
	Blacklist   map[string]struct{}
	// Whitelist  map[string]struct{}
	mutex sync.RWMutex
}

func NewEchoResponseStats() *EchoResponseStats {
	return &EchoResponseStats{
		TwoHundred:  0,
		FourOhThree: 0,
		FourOhFour:  0,
		FiveHundred: 0,
		Scanners:    make(map[string]int),
		Hackers:     make(map[string]int),
		Blacklist:   make(map[string]struct{}),
		mutex:       sync.RWMutex{},
	}
}

// PoliceResponse - track response code counts and block repeat 404 offenders
func (ers *EchoResponseStats) PoliceResponse(nextechohandler echo.HandlerFunc) echo.HandlerFunc {
	const (
		BLACK0 = `IP address %s was blacklisted: too many previous response code errors`
		BLACK1 = `IP address %s was blacklisted: %d StatusNotFound errors`
		BLACK2 = `IP address %s was blacklisted: %d StatusInternalServerError errors`
		FYI200 = `StatusOK count is %d`
		FRQ200 = 1000
		FYI403 = `StatusForbidden count is %d. There are %d IPs currently on the blacklist.`
		FRQ403 = 100
		FYI404 = `StatusNotFound count is %d`
		FRQ404 = 50
		FYI500 = `StatusInternalServerError count is %d`
		FRQ500 = 25
	)

	return func(c echo.Context) error {
		ip := c.RealIP()

		ers.mutex.Lock()
		defer ers.mutex.Unlock()

		// see https://echo.labstack.com/docs/error-handling
		if _, yes := ers.Blacklist[ip]; yes {
			ers.FourOhThree++
			if ers.FourOhThree%FRQ403 == 0 {
				msg(fmt.Sprintf(FYI403, ers.FourOhThree, len(ers.Blacklist)), MSGNOTE)
			}

			e := echo.NewHTTPError(http.StatusForbidden, fmt.Sprintf(BLACK0, c.RealIP()))
			return e
		}

		if err := nextechohandler(c); err != nil {
			c.Error(err)
		}

		switch c.Response().Status {
		case 200:
			ers.TwoHundred++
			if ers.TwoHundred%FRQ200 == 0 {
				msg(fmt.Sprintf(FYI200, ers.TwoHundred), MSGNOTE)
			}

		case 404:
			ers.FourOhFour++

			if _, ok := ers.Scanners[ip]; !ok {
				ers.Scanners[ip] = 1
			} else {
				ers.Scanners[ip]++
			}

			if ers.Scanners[ip] >= MAXFOUROHFOUR {
				ers.Blacklist[ip] = struct{}{}
				msg(fmt.Sprintf(BLACK1, c.RealIP(), ers.Scanners[ip]), MSGWARN)
			}

			if ers.FourOhFour%FRQ404 == 0 {
				msg(fmt.Sprintf(FYI404, ers.FourOhFour), MSGNOTE)
			}

		case 500:
			ers.FiveHundred++

			if _, ok := ers.Hackers[ip]; !ok {
				ers.Hackers[ip] = 1
			} else {
				ers.Hackers[ip]++
			}

			if ers.Hackers[ip] >= MAXFIVEHUNDRED {
				ers.Blacklist[ip] = struct{}{}
				msg(fmt.Sprintf(BLACK2, c.RealIP(), ers.Scanners[ip]), MSGWARN)
			}

			if ers.FiveHundred%FRQ500 == 0 {
				msg(fmt.Sprintf(FYI500, ers.FourOhFour), MSGWARN)
			}

		default:
			// do nothing
			// 302 from "/reset/session" is about the only other code one sees
		}
		return nil
	}
}
