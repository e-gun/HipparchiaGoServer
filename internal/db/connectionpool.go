//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package db

import (
	"context"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/str"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"strings"
)

var Msg = lnch.NewMessageMakerWithDefaults()

//
// Note that SQLite will not really work. See the "devel-sqlite" branch for the brutal details. Way too slow...
//

// FillDBConnectionPool - build the pgxpool that the whole program will Acquire() from
func FillDBConnectionPool(cfg str.CurrentConfiguration) *pgxpool.Pool {
	// it is not clear that the casual user gains much from a pool; this mechanism mattered more for python

	// if min < WorkerCount the search will be slowed significantly
	// and remember that idle connections close, so you can have 20 workers fighting for one connection: very bad news

	// max should cap a networked server's resource allocation to the equivalent of N simultaneous users
	// after that point there should be a steep drop-off in responsiveness

	// nb: macos users can send ANYTHING as a password for hippa_wr: admin access already (on their primary account...)

	const (
		UTPL    = "postgres://%s:%s@%s:%d/%s?pool_min_conns=%d&pool_max_conns=%d"
		FAIL1   = "Configuration error. Could not execute ParseConfig(url) via '%s'"
		FAIL2   = "Could not connect to PostgreSQL"
		ERRRUN  = `dial error`
		FAILRUN = `'%s': the PostgreSQL server cannot be found; check that it is running and serving on port %d`
		ERRSRV  = `server error`
		FAILSRV = `'%s': there is configuration problem; see the following response from PostgreSQL:`
	)

	mn := cfg.WorkerCount
	mx := vv.SIMULTANEOUSSEARCHES * cfg.WorkerCount

	pl := cfg.PGLogin
	url := fmt.Sprintf(UTPL, pl.User, pl.Pass, pl.Host, pl.Port, pl.DBName, mn, mx)

	config, e := pgxpool.ParseConfig(url)
	if e != nil {
		Msg.MAND(fmt.Sprintf(FAIL1, url))
		os.Exit(0)
	}

	thepool, e := pgxpool.NewWithConfig(context.Background(), config)
	if e != nil {
		Msg.MAND(fmt.Sprintf(FAIL2))
		if strings.Contains(e.Error(), ERRRUN) {
			Msg.MAND(fmt.Sprintf(FAILRUN, ERRRUN, cfg.PGLogin.Port))
		}
		if strings.Contains(e.Error(), ERRSRV) {
			Msg.MAND(fmt.Sprintf(FAILSRV, ERRSRV))
			parts := strings.Split(e.Error(), ERRSRV)
			Msg.CRIT(parts[1])
		}
		Msg.ExitOrHang(0)
	}
	return thepool
}
