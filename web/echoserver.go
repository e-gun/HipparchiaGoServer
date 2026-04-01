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

	"github.com/e-gun/HipparchiaGoServer/internal/debug"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	pr "github.com/e-gun/policeresponses"
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

	// change as of echo 5.1.0; and see https://echo.labstack.com/docs/ip-address
	switch lnch.Config.RealIPMeth {
	case 1:
		e.IPExtractor = echo.ExtractIPFromXFFHeader()
	case 2:
		e.IPExtractor = echo.ExtractIPFromRealIPHeader()
	default:
		e.IPExtractor = echo.ExtractIPDirect()
	}

}

func policing(e *echo.Echo) {
	// pr has variables:
	// LockoutResponseFnc: either `BlacklistAndRedirect` or `BlacklistAndSendError`
	// LockoutResponseCode: either `403` or `418`
	// RedirectURL: "https://127.0.0.1" (default)

	// the defaults for the first two are:
	// BlacklistAndRedirect and 418; but the use of the former means that the latter is not going to be relevant

	pr.RedirectURL = "https://127.0.0.1/i_was_blacklisted_because_of_too_many_404s"

	e.Use(pr.PoliceRequestAndResponseV5)
	go pr.ResponseStatsKeeper()
	go pr.IPBlacklistKeeper()
	if !lnch.Config.BlackAndWhite {
		pr.Emit.ColorOn()
	}
	if lnch.Config.TickerActive || lnch.Config.LogToFile {
		pr.Emit.E = Msg.EmitToFile
	} else {
		pr.Emit.E = Msg.AlwaysEmit
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
		failuretolaunch(saddr, err)
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
		failuretolaunch(saddr, err)
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

	styled := Msg.ColStyle(fmt.Sprintf(Msg.Color("C3server started atC0 C2http%s://%sC0C3%sC0"), adds, saddr, addmsg))
	Msg.Emit(styled, -1)
	Msg.SNm = oldname
}

// failuretolaunch - exit the program and tell why
func failuretolaunch(saddr string, err error) {
	// the most common failure to launch is handled here: you are likely already running another copy...
	if strings.Contains(err.Error(), "address already in use") {
		// in full: `listen tcp 127.0.0.1:8001: bind: address already in use`
		Msg.SNm = vv.SHORTNAME

		Msg.Emit(Msg.ColStyle(fmt.Sprintf("C5%sC0 C7failed to startC0", vv.PROJNAME)), -1)
		Msg.Emit(Msg.ColStyle(fmt.Sprintf("C8-->C0 C3%sC0 C2is already useC0 <--", saddr)), -1)
		Msg.Emit("exiting...", -1)
		os.Exit(1)
	}

	// if that was not the error, we still want to know what killed us...
	log.Fatal(err)
}
