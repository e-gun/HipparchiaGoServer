//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// attempt to initialize hipparchiaDB on first launch

const (
	MACPGAPP  = "/Applications/Postgres.app/Contents/Versions/%d/bin/"
	MACBREWA  = "/opt/homebrew/opt/postgresql@%d/bin/"
	MACBREWB  = "/usr/local/bin/"
	WINPGEXE  = `C:\Program Files\PostgreSQL\%d\bin\`
	DBRESTORE = "pg_restore -v --format=directory --username=hippa_wr --dbname=hipparchiaDB %s"
	HDBFOLDER = "hDB"

	NEWDB = `<<EOF
	CREATE USER %s WITH PASSWORD '%s';
	SELECT 'CREATE DATABASE "%s"' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname='"%s"')\gexec
	ALTER DATABASE "%s" OWNER TO %s;
	CREATE EXTENSION pg_trgm;
EOF`
)

func initializeHDB() {
	bindir := findpsql()
	pl := Config.PGLogin
	eof := fmt.Sprintf(NEWDB, pl.User, pl.Pass, pl.DBName, pl.DBName, pl.DBName, pl.User)

	cmd := exec.Command("bash", "-c", bindir+"psql --dbname=postgres "+eof)
	_, err := cmd.Output()
	chke(err)

	msg("Loading the database framework", MSGCRIT)

}

func findpsql() string {
	bindir := ""
	if runtime.GOOS == "darwin" {
		bindir = MACPGAPP
	}

	vers := 0

	for i := 14; i < 20; i++ {
		_, y := os.Stat(fmt.Sprintf(bindir, i) + "psql")
		if y == nil {
			vers = i
			break
		}
	}

	if vers == 0 {
		msg("Cannot find PostgreSQL binaries: aborting", 0)
		os.Exit(0)
	}

	bindir = fmt.Sprintf(bindir, vers)
	return bindir
}

func hipparchiaDBexists(bindir string) bool {
	query := `<<EOF
SELECT datname FROM pg_database WHERE datname='%s';
EOF
`
	exists := false

	eof := fmt.Sprintf(query, DEFAULTPSQLDB)

	cmd := exec.Command("bash", "-c", bindir+"psql --dbname=postgres "+eof)
	out, err := cmd.Output()
	chke(err)
	if strings.Contains(string(out), DEFAULTPSQLDB) {
		exists = true
	}
	return exists
}
