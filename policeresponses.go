package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

//
// RESPONSEPOLICING is only active if Config.Authenticate is "true"
//

type EchoResponseStats struct {
	TwoHundred  uint64
	FourOhThree uint64
	FourOhFour  uint64
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

// global variables to manage the RESPONSEPOLICING infrastructure
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
		checkblacklist := BlackListRD{ip: c.RealIP(), resp: make(chan bool)}
		BListRD <- checkblacklist
		ok := <-checkblacklist.resp

		// presumed guilty: 403
		registerresult := StatListWR{
			code: 403,
			ip:   c.RealIP(),
			uri:  c.Request().RequestURI,
		}

		if !ok {
			// register a 403
			SListWR <- registerresult
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
		BLACK0       = `IP address %s was blacklisted: too many previous response code errors; %d address(es) on the blacklist`
	)

	strikecount := make(map[string]int)
	blacklist := make(map[string]struct{})

	// NB: this loop will never exit
	for {
		select {
		case rcv := <-BListRD:
			valid := true
			if _, ok := blacklist[rcv.ip]; ok {
				// you are on the blacklist...
				valid = false
			}
			rcv.resp <- valid
		case snd := <-BListWR:
			ret := false
			if _, ok := strikecount[snd.ip]; !ok {
				strikecount[snd.ip] = 1
			} else if strikecount[snd.ip] >= FAILSALLOWED {
				blacklist[snd.ip] = struct{}{}
				msg(fmt.Sprintf(BLACK0, snd.ip, len(blacklist)), MSGNOTE)
				ret = true
			} else {
				strikecount[snd.ip]++
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
		FRQ403 = 100
		FYI404 = `StatusNotFound count is %d`
		FRQ404 = 100
		FYI500 = `StatusInternalServerError count is %d.`
		FRQ500 = 1
	)

	// NB: this loop will never exit
	for {
		status := <-SListWR
		switch status.code {
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
			EchoServerStats.FourOhFour++
			if EchoServerStats.FourOhFour%FRQ404 == 0 {
				msg(fmt.Sprintf(FYI404, EchoServerStats.FourOhFour), MSGNOTE)
			}

			// you need to be logged on the blacklist...
			wr := BlackListWR{ip: status.ip, resp: make(chan bool)}
			BListWR <- wr
			ok := <-wr.resp
			if !ok {
				msg(fmt.Sprintf(BLACK1, status.ip, status.uri), MSGWARN)
			}

		case 500:
			EchoServerStats.FiveHundred++
			if EchoServerStats.FiveHundred%FRQ500 == 0 {
				msg(fmt.Sprintf(FYI500, EchoServerStats.FiveHundred), MSGWARN)
			}

			// you need to be logged on the blacklist...
			wr := BlackListWR{ip: status.ip, resp: make(chan bool)}
			BListWR <- wr
			ok := <-wr.resp
			if !ok {
				msg(fmt.Sprintf(BLACK2, status.ip, status.uri), MSGWARN)
			}
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
		FiveHundred: 0,
	}
}
