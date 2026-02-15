//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/rs/zerolog"
)

func GetZLogger(output *os.File) *zerolog.Logger {
	return new(zerolog.New(zerolog.ConsoleWriter{
		Out:        output,
		TimeFormat: time.DateTime,
	}))
}

// GetMyLvlLogger - return a logging function; different degrees of explicitness as per `thelevel`
func GetMyLvlLogger(thelevel int, logger *zerolog.Logger) func(c *echo.Context, v middleware.RequestLoggerValues) error {
	// 1: 2026-02-02 17:40:08 INF 200 URI=/emb/echarts/echarts.min.js
	// 2: 2026-02-02 17:40:08 INF 200 IP=127.0.0.1 URI=/emb/jq/jquery.min.js
	// 3: 2026-02-02 17:40:08 INF 200 IP=127.0.0.1 SZ=109 UA=Firefox/147.0 URI=/selection/fetch

	lvl1 := func(c *echo.Context, v middleware.RequestLoggerValues) error {
		// 2026-02-02 17:40:08 INF 200 URI=/emb/echarts/echarts.min.js
		st := fmt.Sprintf("%d", v.Status)
		if v.Status < 400 {
			logger.Info().
				Timestamp().
				Str("URI", v.URI).
				Msg(st)
		} else {
			logger.Warn().
				Timestamp().
				Str("URI", v.URI).
				Msg(st)
		}
		return nil
	}

	lvl2 := func(c *echo.Context, v middleware.RequestLoggerValues) error {
		// 2026-02-02 17:40:08 INF 200 IP=127.0.0.1 URI=/emb/jq/jquery.min.js
		ipsplit := strings.Split(c.Request().RemoteAddr, ":")
		if len(ipsplit) != 2 {
			return c.String(http.StatusForbidden, "IP format error")
		}

		st := fmt.Sprintf("%d", v.Status)
		if v.Status < 400 {
			logger.Info().
				Timestamp().
				Str("IP", ipsplit[0]).
				Str("URI", v.URI).
				Msg(st)
		} else {
			logger.Warn().
				Timestamp().
				Str("IP", ipsplit[0]).
				Str("URI", v.URI).
				Msg(st)
		}
		return nil
	}

	lvl3 := func(c *echo.Context, v middleware.RequestLoggerValues) error {
		// 2026-02-02 17:40:08 INF 200 IP=127.0.0.1 SZ=109 UA=Firefox/147.0 URI=/selection/fetch
		ipsplit := strings.Split(c.Request().RemoteAddr, ":")
		if len(ipsplit) != 2 {
			return c.String(http.StatusForbidden, "IP format error")
		}

		ua := strings.Split(v.UserAgent, " ")
		agent := ua[len(ua)-1]
		st := fmt.Sprintf("%d", v.Status)

		if v.Status < 400 {
			logger.Info().
				Timestamp().
				Str("SZ", fmt.Sprintf("%d", v.ResponseSize)).
				Str("IP", ipsplit[0]).
				Str("URI", v.URI).
				Str("UA", agent).
				Msg(st)
		} else {
			logger.Warn().
				Timestamp().
				Str("SZ", fmt.Sprintf("%d", v.ResponseSize)).
				Str("IP", ipsplit[0]).
				Str("URI", v.URI).
				Str("UA", agent).
				Msg(st)
		}
		return nil
	}

	lvlog := lvl3

	switch thelevel {
	case 3:
		lvlog = lvl3
	case 2:
		lvlog = lvl2
	case 1:
		lvlog = lvl1
	default:
		// do nothing; but this is "3" as per the above
	}

	return lvlog
}
