//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vlt

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"time"
)

//
// RESPONSEPOLICING is only active if Config.Authenticate is "true"
//

type EchoResponseStats struct {
	TwoHundred  uint64
	FourOhThree uint64
	FourOhFour  uint64
	FourOhFive  uint64
	FiveHundred uint64
}

type BlackListRD struct {
	ip   string
	resp chan bool
}

type BlackListWR struct {
	ip   string
	resp chan bool
}

type StatListWR struct {
	code int
	ip   string
	uri  string
}

// variables to manage the RESPONSEPOLICING infrastructure
var (
	BListWR         = make(chan BlackListWR)
	BListRD         = make(chan BlackListRD)
	SListWR         = make(chan StatListWR)
	EchoServerStats = NewEchoResponseStats()
)

// PoliceRequestAndResponse - track Response code counts + block repeat 404 offenders; this is custom middleware for an *echo.Echo
func PoliceRequestAndResponse(nextechohandler echo.HandlerFunc) echo.HandlerFunc {
	const (
		BLACK0 = `IP address %s was blacklisted: too many previous Response code errors`
		SLOWDN = 3
		BLACK1 = `IP address %s received a strike: invalid request prefix in URI "%s"`
	)

	return func(c echo.Context) error {
		// presumed guilty: 403
		registerresult := StatListWR{
			code: 403,
			ip:   c.RealIP(),
			uri:  c.Request().RequestURI,
		}

		// already known to be bad?
		checkblacklist := BlackListRD{ip: c.RealIP(), resp: make(chan bool)}
		BListRD <- checkblacklist
		ok := <-checkblacklist.resp

		// is something like 'http://journalseek.net/' in the request?
		rq := c.Request().RequestURI
		if strings.HasPrefix(rq, "http:") || strings.HasPrefix(rq, "https:") {
			ok = false
			addtoblacklist := BlackListWR{ip: c.RealIP(), resp: make(chan bool)}
			BListWR <- addtoblacklist
			white := <-addtoblacklist.resp // are you over the limit?
			if !white {
				Msg.WARN(fmt.Sprintf(BLACK1, c.RealIP(), rq))
			}
		}

		if !ok {
			// register a 403
			SListWR <- registerresult
			time.Sleep(SLOWDN * time.Second)
			e := echo.NewHTTPError(http.StatusForbidden, fmt.Sprintf(BLACK0, c.RealIP()))
			return e
		} else {
			// do this before setting c.Response().Status or you will always get "200"
			if err := nextechohandler(c); err != nil {
				c.Error(err)
			}
			// register some other result code
			registerresult.code = c.Response().Status
			SListWR <- registerresult
			return nil
		}
	}
}

// IPBlacklistKeeper - blacklist read/write
func IPBlacklistKeeper() {
	const (
		FAILSALLOWED = 3
		BLACK0       = `IP address %s was blacklisted: too many previous Response code errors; %d address(es) on the blacklist`
	)

	strikecount := make(map[string]int)
	blacklist := make(map[string]struct{})

	// NB: this loop will never exit
	// the channels are returning 'bool'
	for {
		select {
		case rd := <-BListRD: // read from the blacklist
			valid := true
			if _, ok := blacklist[rd.ip]; ok {
				// you are on the blacklist...
				valid = false
			}
			rd.resp <- valid
		case wr := <-BListWR: // check strikes; maybe write to the blacklist
			ret := false
			if _, ok := strikecount[wr.ip]; !ok {
				strikecount[wr.ip] = 1
			} else if strikecount[wr.ip] >= FAILSALLOWED {
				blacklist[wr.ip] = struct{}{}
				Msg.NOTE(fmt.Sprintf(BLACK0, wr.ip, len(blacklist)))
				ret = true
			} else {
				strikecount[wr.ip]++
			}
			wr.resp <- ret
		}
	}
}

// ResponseStatsKeeper - log echo responses; should have exclusive r/w access to EchoServerStats
func ResponseStatsKeeper() {
	const (
		BLACK1 = `IP address %s received a strike: StatusNotFound error for URI "%s"`
		BLACK2 = `IP address %s received a strike: StatusInternalServerError for URI "%s"`
		BLACK3 = `IP address %s received a strike: MethodNotAllowed for URI "%s"`
		FYI200 = `StatusOK count is %d`
		FRQ200 = 1000
		FYI403 = `[%s] StatusForbidden count is %d. Last blocked was %s requesting "%s"`
		FRQ403 = 100
		FYI404 = `[%s] StatusNotFound count is %d`
		FRQ404 = 100
		FYI405 = `[%s] MethodNotAllowed count is %d`
		FRQ405 = 5
		FYI500 = `[%s] StatusInternalServerError count is %d.`
		FRQ500 = 1
	)

	warn := func(v uint64, frq uint64, fyi string) {
		if v%frq == 0 {
			Msg.NOTE(fmt.Sprintf(fyi, v))
		}
	}

	blacklist := func(status StatListWR, note string) {
		// you need to be logged on the blacklist...
		wr := BlackListWR{ip: status.ip, resp: make(chan bool)}
		BListWR <- wr
		ok := <-wr.resp
		if !ok {
			Msg.WARN(fmt.Sprintf(BLACK1, status.ip, status.uri))
		}
	}

	// NB: this loop will never exit
	for {
		status := <-SListWR
		when := time.Now().Format(time.RFC822)
		switch status.code {
		case 200:
			EchoServerStats.TwoHundred++
			warn(EchoServerStats.TwoHundred, FRQ200, FYI200)
		case 403:
			// you are already on the blacklist...
			EchoServerStats.FourOhThree++
			// use of 'when' makes this different...
			if EchoServerStats.FourOhThree%FRQ403 == 0 {
				Msg.NOTE(fmt.Sprintf(FYI403, when, EchoServerStats.FourOhThree, status.ip, status.uri))
			}
		case 404:
			EchoServerStats.FourOhFour++
			warn(EchoServerStats.FourOhFour, FRQ404, FYI404)
			blacklist(status, BLACK1)
		case 405:
			// these seem to come only from hostile scanners; it is a bug that needs fixing if a real user sees this
			EchoServerStats.FourOhFive++
			warn(EchoServerStats.FourOhFive, FRQ405, FYI405)
			blacklist(status, BLACK3)
		case 500:
			EchoServerStats.FiveHundred++
			warn(EchoServerStats.FiveHundred, FRQ500, FYI500)
			blacklist(status, BLACK2)
		default:
			// do nothing: not interested
			// 302 from "/reset/session"
			// 101 from "/ws"
		}
	}
}

// NewEchoResponseStats - return the one and only copy of EchoResponseStats, i.e. the EchoServerStats global variable
func NewEchoResponseStats() *EchoResponseStats {
	return &EchoResponseStats{
		TwoHundred:  0,
		FourOhThree: 0,
		FourOhFour:  0,
		FourOhFive:  0,
		FiveHundred: 0,
	}
}
