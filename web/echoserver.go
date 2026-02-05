//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/e-gun/HipparchiaGoServer/internal/debug"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/rs/zerolog"
)

var (
	Msg = lnch.NewMessageMakerWithDefaults()
)

// StartEchoServer - start serving; this blocks and does not return while the program remains alive
func StartEchoServer() {
	e := echo.New()
	configureecho(e)

	//
	// the next long block is all about logging
	// all logging config has to be kept inside StartEchoServer(); else "file already closed"
	//

	output := os.Stderr
	var err error

	if lnch.Config.LogToFile {
		uh, _ := os.UserHomeDir()
		output, err = os.Create(uh + "/" + vv.LOGFILEEL)
		if err != nil {
			os.Exit(1)
		}
		defer func(output *os.File) {
			e2 := output.Close()
			if e2 != nil {
				fmt.Println(e2.Error())
			}
		}(output)
	}

	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        output,
		TimeFormat: time.DateTime,
	})

	lvl1 := func(c *echo.Context, v middleware.RequestLoggerValues) error {
		// 2026-02-02 17:40:08 INF 200 URI=/emb/echarts/echarts.min.js
		if v.Status < 400 {
			logger.Info().
				Timestamp().
				Str("URI", v.URI).
				Msg(fmt.Sprintf("%d", v.Status))
		} else {
			logger.Warn().
				Timestamp().
				Str("URI", v.URI).
				Msg(fmt.Sprintf("%d", v.Status))
		}
		return nil
	}

	lvl2 := func(c *echo.Context, v middleware.RequestLoggerValues) error {
		// 2026-02-02 17:40:08 INF 200 IP=127.0.0.1:63190 URI=/emb/jq/jquery.min.js
		if v.Status < 400 {
			logger.Info().
				Timestamp().
				Str("IP", c.Request().RemoteAddr).
				Str("URI", v.URI).
				Msg(fmt.Sprintf("%d", v.Status))
		} else {
			logger.Warn().
				Timestamp().
				Str("IP", c.Request().RemoteAddr).
				Str("URI", v.URI).
				Msg(fmt.Sprintf("%d", v.Status))
		}
		return nil
	}

	lvl3 := func(c *echo.Context, v middleware.RequestLoggerValues) error {
		// 2026-02-02 17:40:08 INF 200 IP=127.0.0.1:64246 SZ=109 UA=Firefox/147.0 URI=/selection/fetch
		ua := strings.Split(v.UserAgent, " ")
		agent := ua[len(ua)-1]

		if v.Status < 400 {
			logger.Info().
				Timestamp().
				Str("SZ", fmt.Sprintf("%d", v.ResponseSize)).
				Str("IP", c.Request().RemoteAddr).
				Str("URI", v.URI).
				Str("UA", agent).
				Msg(fmt.Sprintf("%d", v.Status))
		} else {
			logger.Warn().
				Timestamp().
				Str("SZ", fmt.Sprintf("%d", v.ResponseSize)).
				Str("IP", c.Request().RemoteAddr).
				Str("URI", v.URI).
				Str("UA", agent).
				Msg(fmt.Sprintf("%d", v.Status))
		}
		return nil
	}

	lvlog := lvl3

	if lnch.Config.EchoLog > 0 {
		switch lnch.Config.EchoLog {
		case 3:
			lvlog = lvl3
		case 2:
			lvlog = lvl2
		case 1:
			lvlog = lvl1
		default:
			// do nothing; but this is effectively "3"
		}

		rqlogmiddleware := middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogContentLength: true,
			LogRemoteIP:      true,
			LogHost:          true,
			LogURI:           true,
			LogUserAgent:     true,
			LogStatus:        true,
			LogResponseSize:  true,
			LogValuesFunc:    lvlog,
		})

		e.Use(rqlogmiddleware)
	}

	//
	// now that logging is set up we can do the real business of launching...
	//

	buildroutes(e)

	// next will do nothing if Config is not activating these features
	go debug.RunSelfTests()
	go activatevectorbot()

	if candossl() {
		starttlsserver(e)
	} else {
		starthttpserver(e)
	}
}

func configureecho(e *echo.Echo) {
	if lnch.Config.Authenticate {
		// also assume that internet exposure yields scanning attempts that will spam 404s & 500s; block IPs that do this
		policing(e)
	}

	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(vv.MAXECHOREQPERSECONDPERIP)))

	e.Use(middleware.Recover())

	if lnch.Config.Gzip {
		e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 5}))
	}
}

func policing(e *echo.Echo) {
	e.Use(PoliceRequestAndResponseV5)
	go ResponseStatsKeeper()
	go IPBlacklistKeeper()
	if !lnch.Config.BlackAndWhite {
		Emit.ColorOn()
	}
	if lnch.Config.TickerActive || lnch.Config.LogToFile {
		Emit.E = Msg.EmitToFile
	} else {
		Emit.E = Msg.AlwaysEmit
	}
}

func candossl() bool {
	ok1 := false
	if _, err := os.Stat(lnch.Config.SSLCertDir + vv.SSLCPEM); err == nil {
		ok1 = true
	}
	ok2 := false
	if _, err := os.Stat(lnch.Config.SSLCertDir + vv.SSLPPEM); err == nil {
		ok2 = true
	}
	ok := ok1 && ok2
	return ok
}

func starthttpserver(e *echo.Echo) {
	Msg.WARN("(tls unavailable)")

	saddr := fmt.Sprintf("%s:%d", lnch.Config.HostIP, lnch.Config.HostPort)
	fmt.Printf("⇨ http server started at %s\n", saddr)

	s := http.Server{
		Addr:    saddr,
		Handler: e,
	}

	// assume that anyone who is using authentication is serving via the internet and so set timeouts
	if lnch.Config.Authenticate {
		s.ReadTimeout = vv.TIMEOUTRD
		s.WriteTimeout = vv.TIMEOUTWR
	}

	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func starttlsserver(e *echo.Echo) {
	saddr := fmt.Sprintf("%s:%d", lnch.Config.HostIP, lnch.Config.HostSSLPort)
	fmt.Printf("⇨ https server started at %s\n", saddr)

	s := http.Server{
		Addr:      saddr,
		Handler:   e, // set Echo as handler
		TLSConfig: &tls.Config{
			//Certificates: nil, // <-- s.ListenAndServeTLS will populate this field
		},
	}

	if lnch.Config.Authenticate {
		s.ReadTimeout = vv.TIMEOUTRD
		s.WriteTimeout = vv.TIMEOUTWR
	}

	if err := s.ListenAndServeTLS(lnch.Config.SSLCertDir+vv.SSLCPEM, lnch.Config.SSLCertDir+vv.SSLPPEM); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
