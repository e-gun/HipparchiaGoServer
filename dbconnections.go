//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
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

type PostgresLogin struct {
	Host   string
	Port   int
	User   string
	Pass   string
	DBName string
}

// FillPSQLPoool - build the pgxpool that the whole program will Acquire() from
func FillPSQLPoool() *pgxpool.Pool {
	// it is not clear that the casual user gains much from a pool; this mechanism mattered more for python

	// if min < WorkerCount the search will be slowed significantly
	// and remember that idle connections close, so you can have 20 workers fighting for one connection: very bad news

	// max should cap a networked server's resource allocation to the equivalent of N simultaneous users
	// after that point there should be a steep drop-off in responsiveness

	const (
		UTPL    = "postgres://%s:%s@%s:%d/%s?pool_min_conns=%d&pool_max_conns=%d"
		FAIL1   = "Configuration error. Could not execute ParseConfig(url) via '%s'"
		FAIL2   = "Could not connect to PostgreSQL"
		ERRRUN  = `dial error`
		FAILRUN = `'%s': the PostgreSQL server cannot be found; check that it is running and serving on port %d`
		ERRSRV  = `server error`
		FAILSRV = `'%s': there is configuration problem; see the following response from PostgreSQL:`
	)

	min := Config.WorkerCount
	max := SIMULTANEOUSSEARCHES * Config.WorkerCount

	pl := Config.PGLogin
	url := fmt.Sprintf(UTPL, pl.User, pl.Pass, pl.Host, pl.Port, pl.DBName, min, max)

	config, e := pgxpool.ParseConfig(url)
	if e != nil {
		msg(fmt.Sprintf(FAIL1, url), -1)
		os.Exit(0)
	}

	thepool, e := pgxpool.NewWithConfig(context.Background(), config)
	if e != nil {
		msg(fmt.Sprintf(FAIL2), -1)
		if strings.Contains(e.Error(), ERRRUN) {
			msg(fmt.Sprintf(FAILRUN, ERRRUN, Config.PGLogin.Port), -1)
		}
		if strings.Contains(e.Error(), ERRSRV) {
			msg(fmt.Sprintf(FAILSRV, ERRSRV), -1)
			parts := strings.Split(e.Error(), ERRSRV)
			msg(parts[1], 0)
		}
		os.Exit(0)
	}
	return thepool
}

// GetPSQLconnection - Acquire() a connection from the main pgxpool
func GetPSQLconnection() *pgxpool.Conn {
	const (
		FAIL = "GetPSQLconnection() could not Acquire() from SQLPool"
	)

	dbc, e := SQLPool.Acquire(context.Background())
	if e != nil {
		msg(fmt.Sprintf(FAIL), -1)
		panic(e)
	}
	return dbc
}
