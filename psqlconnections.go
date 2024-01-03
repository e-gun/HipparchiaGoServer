//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"os/exec"
	"strings"
	"time"
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

// GetPSQLconnection - Acquire() a connection from the main pgxpool; include tests for the status of the installation
func GetPSQLconnection() *pgxpool.Conn {
	const (
		FAIL1   = "GetPSQLconnection() could not Acquire() from SQLPool."
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

// PostgresDumpDB - dump the database to the filesystem as a pgdump
func PostgresDumpDB() {
	const (
		MSG   = "Extracting the database.."
		ERR   = "PostgresDumpDB(): pg_dump failed. You should NOT trust this archive. Deleting it..."
		WRK   = 1 // problem (on virtualized machine): "server closed the connection unexpectedly" if WRK > 1
		WARN  = "The database will start archiving in %d seconds. C7This will take several minutesC0"
		DELAY = 5
	)

	fmt.Println(coloroutput(fmt.Sprintf(WARN, DELAY)))
	time.Sleep(DELAY * time.Second)

	// pg_dump --clean "hipparchiaDB" --user hippa_wr | split -b 100m - out/hipparchiaDB-
	// pg_dump -U postgres -F d -j 5 db1 -f db1_backup

	// don't want an extra 1GB... should run with "-rv" flag before doing "-ex", but maybe you didn't
	// unable to call "vectordbreset()" at this juncture
	// panic: runtime error: invalid memory address or nil pointer dereference

	// highly likely that you do not have a value for Config.PGLogin.Pass yet, but you need one...
	SetConfigPass(Config, "")

	binary := GetBinaryPath("pg_dump")
	url := GetHippaWRURI(Config.PGLogin.Pass)

	workers := fmt.Sprintf("%d", WRK)

	cmd := exec.Command(binary, "-v", "-T", VECTORTABLENAMENN, "-F", "d", "-j", workers, "-f", HDBFOLDER, url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	msg(MSG, MSGCRIT)
	err := cmd.Run()
	if err != nil {
		msg(ERR, MSGCRIT)
		e := os.RemoveAll(HDBFOLDER)
		chke(e)
	}
}

// PostgresDBtoCSV - dump the database to the filesystem as CSV
func PostgresDBtoCSV() {
	const (
		DQ     = `\COPY %s TO '%s/%s/%s.csv' DELIMITER ',' CSV HEADER;` // COPY lt2000 TO '/Users/erik/tmp/lt2000.csv' DELIMITER ',' CSV HEADER;
		OUTDIR = `csv_db`
	)
	b := GetBinaryPath("psql")

	support := []string{"authors", "works", "latin_morphology", "greek_morphology", "latin_dictionary",
		"greek_dictionary", "greek_lemmata", "latin_lemmata", "dictionary_headword_wordcounts"}
	counts := strings.Split("abcdefghijklmnopqrstuvwxyz0αβψδεφγηιξκλμνοπρϲτυω", "")

	allauthortables := StringMapKeysIntoSlice(AllAuthors)

	writeout := func(q string) {
		cmd := exec.Command(b, "-d", "hipparchiaDB", "-c", q)
		// cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		chke(err)
		msg(q, MSGFYI)
	}

	h, e := os.UserHomeDir()
	chke(e)
	// h := "/tmp"

	e = os.Mkdir(h+"/"+OUTDIR, 0755)
	//if strings.Contains(e.Error(), "exists") {
	//	msg(h+"/"+OUTDIR+" already exists", MSGFYI)
	//} else {
	//	chke(e)
	//}

	// psql -d hipparchiaDB -c "\COPY lt0881 TO '/Users/erik/csv_db/lt0881.csv' DELIMITER ',' CSV HEADER;"
	for i := 0; i < len(allauthortables); i++ {
		q := fmt.Sprintf(DQ, allauthortables[i], h, OUTDIR, allauthortables[i])
		writeout(q)
	}

	for i := 0; i < len(support); i++ {
		q := fmt.Sprintf(DQ, support[i], h, OUTDIR, support[i])
		writeout(q)
	}

	for i := 0; i < len(counts); i++ {
		q := fmt.Sprintf(DQ, "wordcounts_"+counts[i], h, OUTDIR, "wordcounts_"+counts[i])
		writeout(q)
	}

}

//
// DBConnectionHolder, etc. YIELD ABILITY TO HANDLE HETEROGENEOUS DATABASE CONNECTIONS: postgres and sqlite
//

type DBConnectionHolder struct {
	Postgres *pgxpool.Conn
	Lite     *sql.Conn
}

func (ch *DBConnectionHolder) Release() {
	switch SQLProvider {
	case "sqlite":
		ch.Lite.Close()
	default:
		ch.Postgres.Release()
	}
}

func GrabDBConnection() *DBConnectionHolder {
	var lt *sql.Conn
	var pg *pgxpool.Conn

	switch SQLProvider {
	case "sqlite":
		lt = GetSQLiteConn()
	default:
		pg = GetPSQLconnection()
	}

	return &DBConnectionHolder{
		Postgres: pg,
		Lite:     lt,
	}
}
