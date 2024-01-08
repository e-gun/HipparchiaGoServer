//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"strings"
)

//
// Note that SQLite will not really work. See the "devel-sqlite" branch for the brutal details. Way too slow...
//

type PostgresLogin struct {
	Host   string
	Port   int
	User   string
	Pass   string
	DBName string
}

// FillDBConnectionPool - build the pgxpool that the whole program will Acquire() from
func FillDBConnectionPool() *pgxpool.Pool {
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

	mn := Config.WorkerCount
	mx := SIMULTANEOUSSEARCHES * Config.WorkerCount

	pl := Config.PGLogin
	url := fmt.Sprintf(UTPL, pl.User, pl.Pass, pl.Host, pl.Port, pl.DBName, mn, mx)

	config, e := pgxpool.ParseConfig(url)
	if e != nil {
		msg(fmt.Sprintf(FAIL1, url), MSGMAND)
		os.Exit(0)
	}

	thepool, e := pgxpool.NewWithConfig(context.Background(), config)
	if e != nil {
		msg(fmt.Sprintf(FAIL2), MSGMAND)
		if strings.Contains(e.Error(), ERRRUN) {
			msg(fmt.Sprintf(FAILRUN, ERRRUN, Config.PGLogin.Port), MSGMAND)
		}
		if strings.Contains(e.Error(), ERRSRV) {
			msg(fmt.Sprintf(FAILSRV, ERRSRV), MSGMAND)
			parts := strings.Split(e.Error(), ERRSRV)
			msg(parts[1], MSGCRIT)
		}
		messenger.ExitOrHang(0)
	}
	return thepool
}

// GetDBConnection - Acquire() a connection from the main pgxpool
func GetDBConnection() *pgxpool.Conn {
	const (
		FAIL1   = "GetDBConnection() could not Acquire() from the DBConnectionPool."
		FAIL2   = `Your password in '%s' is incorrect? Too many connections to the server?`
		FAIL3   = `The database is empty. Deleting any configuration files so you can reset the server.`
		FAIL4   = `Failed to delete %s`
		ERRRUN  = `dial error`
		FAILRUN = `'%s': the PostgreSQL server cannot be found; check that it is running and serving on port %d`
	)

	dbc, e := SQLPool.Acquire(context.Background())
	if e != nil {
		if !HipparchiaDBHasData(Config.PGLogin.Pass) {
			// you need to reset the whole application...
			msg(coloroutput(fmt.Sprintf(FAIL3)), MSGMAND)
			h, err := os.UserHomeDir()
			chke(err)
			err = os.Remove(fmt.Sprintf(CONFIGALTAPTH, h) + CONFIGBASIC)
			if err != nil {
				msg(fmt.Sprintf(FAIL4, CONFIGBASIC), MSGCRIT)
			}
			err = os.Remove(fmt.Sprintf(CONFIGALTAPTH, h) + CONFIGPROLIX)
			if err != nil {
				msg(fmt.Sprintf(FAIL4, CONFIGPROLIX), MSGCRIT)
			}
			messenger.ExitOrHang(0)
		}

		msg(fmt.Sprintf(FAIL1), MSGMAND)
		if strings.Contains(e.Error(), ERRRUN) {
			msg(fmt.Sprintf(FAILRUN, ERRRUN, Config.PGLogin.Port), MSGCRIT)
		} else {
			msg(fmt.Sprintf(FAIL2, CONFIGBASIC), MSGMAND)
		}
		messenger.ExitOrHang(0)
	}
	return dbc
}
