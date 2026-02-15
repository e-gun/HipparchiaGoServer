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

	"github.com/e-gun/HipparchiaGoServer/internal/debug"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

var (
	Msg = lnch.NewMessageMakerWithDefaults()
)

// StartEchoServer - start serving; this blocks and does not return while the program remains alive
func StartEchoServer() {
	e := echo.New()
	configureecho(e)

	//
	// the next block is all about logging
	// "file already closed" danger if `output` is not handled properly (i.e., opened and closed in this function)
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

	if lnch.Config.EchoLog > 0 {
		rqlogmiddleware := middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogContentLength: true,
			LogRemoteIP:      true,
			LogHost:          true,
			LogURI:           true,
			LogUserAgent:     true,
			LogStatus:        true,
			LogResponseSize:  true,
			LogValuesFunc:    GetMyLvlLogger(lnch.Config.EchoLog, GetZLogger(output)),
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
	saddr := fmt.Sprintf("%s:%d", lnch.Config.HostIP, lnch.Config.HostPort)

	serverlaunchmessage(saddr, false)

	s := http.Server{
		Addr:    saddr,
		Handler: e,
	}

	// assume that anyone who is using authentication is serving via the internet and so set timeouts
	// some searches can be very long so localhost should be given every opportunity to run
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

	serverlaunchmessage(saddr, true)

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

	if err := s.ListenAndServeTLS(lnch.Config.SSLCertDir+vv.SSLCPEM, lnch.Config.SSLCertDir+vv.SSLPPEM); errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func serverlaunchmessage(saddr string, isssl bool) {
	adds := ""
	addmsg := " (tls unavailable)"
	if isssl {
		adds = "s"
		addmsg = ""
	}

	oldname := Msg.SNm
	Msg.SNm = vv.SHORTNAME

	styled := Msg.ColStyle(fmt.Sprintf(Msg.Color("C3server listening atC0 C2http%s://%sC0C3%sC0"), adds, saddr, addmsg))
	Msg.Emit(styled, -1)
	Msg.SNm = oldname
}
