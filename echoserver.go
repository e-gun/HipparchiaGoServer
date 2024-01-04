//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"bytes"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"strings"
)

// StartEchoServer - start serving; this blocks and does not return while the program remains alive
func StartEchoServer() {
	const (
		LLOGFMT = "r: ${status}\tt: ${latency_human}\tu: ${uri}\n"
		RLOGFMT = "${remote_ip}\t${custom}\t${status}\t${bytes_out}\t${uri}\n"
	)

	// ctf - a CustomTagFunc return a short user agent
	ctf := func(c echo.Context, buf *bytes.Buffer) (int, error) {
		ua := strings.Split(c.Request().UserAgent(), " ")
		if len(ua) == 0 {
			return 0, nil
		} else {
			last := ua[len(ua)-1]
			buf.Write([]byte(last))
			return 1, nil
		}
	}

	//
	// SETUP
	//

	e := echo.New()

	if Config.Authenticate {
		// assume that anyone who is using authentication is serving via the internet and so set timeouts
		e.Server.ReadTimeout = TIMEOUTRD
		e.Server.WriteTimeout = TIMEOUTWR

		// also assume that internet exposure yields scanning attempts that will spam 404s && 500s; block IPs that do this
		// see "policeresponses.go" for these functions
		go IPBlacklistKeeper()
		go ResponseStatsKeeper()
		e.Use(PoliceResponse)
	}

	switch Config.EchoLog {
	case 3:
		e.Use(middleware.Logger())
	case 2:
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{Format: RLOGFMT, CustomTagFunc: ctf}))
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
	e.HidePort = false
	e.Debug = false
	e.DisableHTTP2 = true
	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%d", Config.HostIP, Config.HostPort)))
}

// JSONresponse - send the JSON; this function lets one test and document different strategies; jsr should be a json-ready struct
func JSONresponse(c echo.Context, jsr any) error {
	const (
		RESPONDER = 1
	)

	// note that JSONPretty will end up strikingly prominent on the profiler: a waste of memory and cycles unless
	// you are debugging and want to be able to inspect the json manually

	// [1] "vanilla"; and it turns out there is nothing wrong with vanilla; seems like the best choice
	opt1 := func() error { return c.JSON(http.StatusOK, jsr) }

	// [2] "costs a lot of RAM in return for what?"
	//opt2 := func() error { return c.JSONPretty(http.StatusOK, jsr, JSONINDENT) }

	// [3] "maybe streaming makes sense..." but this uses slightly more memory than [a] and is slightly slower?
	//opt3 := func() error {
	//	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	//	c.Response().WriteHeader(http.StatusOK)
	//	return json.NewEncoder(c.Response()).Encode(jsr)
	//}

	// [4] jsoniter: purportedly faster json, but we are one-big and not many-small...
	// requires: import jsoniter "github.com/json-iterator/go"
	// nb: not fully "ConfigCompatibleWithStandardLibrary" as it cannot do "JSONPretty"
	//

	//opt4 := func() error {
	//	b, e := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(&jsr)
	//	chke(e)
	//	return c.JSONBlob(http.StatusOK, b)
	//}

	//switch RESPONDER {
	//case 1:
	//	return opt1()
	//case 2:
	//	return opt2()
	//case 3:
	//	return opt3()
	//case 4:
	//	return opt4()
	//default:
	//	return opt1()
	//}

	return opt1()
}
