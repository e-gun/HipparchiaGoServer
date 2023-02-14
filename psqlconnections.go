//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"runtime"
	"strings"
)

//
// Note that SQLite will not really work. It is 10% as fast on a single word search, does not like concat(),
// cannot readily do regex, etc. This makes SQLite way too costly even if the vision of a serverless solution with only
// a "hgs_sqlite.db" file is enticing.
//

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
		exitorhang(0)
	}
	return thepool
}

// GetPSQLconnection - Acquire() a connection from the main pgxpool
func GetPSQLconnection() *pgxpool.Conn {
	const (
		FAIL1   = "GetPSQLconnection() could not Acquire() from SQLPool."
		FAIL2   = `Your password in '%s' is incorrect?`
		FAIL3   = `The database is empty. Deleting any 'C3%sC0' so you can reset the server.`
		ERRRUN  = `dial error`
		FAILRUN = `'%s': the PostgreSQL server cannot be found; check that it is running and serving on port %d`
	)

	dbc, e := SQLPool.Acquire(context.Background())
	if e != nil {
		if !HipparchiaDBHasData(Config.PGLogin.Pass) {
			// you need to reset the whole application...
			msg(coloroutput(fmt.Sprintf(FAIL3, CONFIGBASIC)), MSGMAND)
			h, err := os.UserHomeDir()
			chke(err)
			err = os.Remove(fmt.Sprintf(CONFIGALTAPTH, h) + CONFIGBASIC)
			chke(err)
			exitorhang(0)
		}

		msg(fmt.Sprintf(FAIL1), MSGMAND)
		if strings.Contains(e.Error(), ERRRUN) {
			msg(fmt.Sprintf(FAILRUN, ERRRUN, Config.PGLogin.Port), MSGCRIT)
		} else {
			msg(fmt.Sprintf(FAIL2, CONFIGBASIC), MSGMAND)
		}
		exitorhang(0)
	}
	return dbc
}

// exitorhang - windows need to hang to keep the error window open
func exitorhang(e int) {
	const (
		HANG = `Execution suspended. %s is now frozen. Read any errors above. Then close this window.`
	)
	if runtime.GOOS != "windows" {
		os.Exit(e)
	} else {
		msg(fmt.Sprintf(HANG, MYNAME), -1)
		for {
			// you are now hung
		}
	}
}
