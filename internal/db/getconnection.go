//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package db

import (
	"context"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"strings"
)

var (
	SQLPool *pgxpool.Pool
)

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
		if !lnch.HipparchiaDBHasData(lnch.Config.PGLogin.Pass) {
			// you need to reset the whole application...
			Msg.MAND(Msg.Color(fmt.Sprintf(FAIL3)))
			h, err := os.UserHomeDir()
			Msg.EC(err)
			err = os.Remove(fmt.Sprintf(vv.CONFIGALTAPTH, h) + vv.CONFIGBASIC)
			if err != nil {
				Msg.CRIT(fmt.Sprintf(FAIL4, vv.CONFIGBASIC))
			}
			err = os.Remove(fmt.Sprintf(vv.CONFIGALTAPTH, h) + vv.CONFIGPROLIX)
			if err != nil {
				Msg.CRIT(fmt.Sprintf(FAIL4, vv.CONFIGPROLIX))
			}
			Msg.ExitOrHang(0)
		}

		Msg.MAND(fmt.Sprintf(FAIL1))
		if strings.Contains(e.Error(), ERRRUN) {
			Msg.CRIT(fmt.Sprintf(FAILRUN, ERRRUN, lnch.Config.PGLogin.Port))
		} else {
			Msg.MAND(fmt.Sprintf(FAIL2, vv.CONFIGBASIC))
		}
		Msg.ExitOrHang(0)
	}
	return dbc
}
