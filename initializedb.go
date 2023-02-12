//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
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

func initializeHDB(pw string) {
	bindir := findpsql()
	eof := fmt.Sprintf(NEWDB, DEFAULTPSQLUSER, pw, DEFAULTPSQLDB, DEFAULTPSQLDB, DEFAULTPSQLDB, DEFAULTPSQLUSER)

	msg(bindir+"psql --dbname=postgres "+eof, MSGCRIT)
	cmd := exec.Command("bash", "-c", bindir+"psql --dbname=postgres "+eof)
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	chke(err)

	msg("initialized the database framework", MSGCRIT)

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
		msg("Cannot find PostgreSQL binaries: aborting", MSGCRIT)
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

func hipparchiaDBhasdata(bindir string) bool {
	query := `<<EOF
SELECT EXISTS (
    SELECT FROM 
        pg_tables
    WHERE 
        schemaname = 'public' AND 
        tablename  = 'authors'
    );
EOF`
	exists := false

	cmd := exec.Command("bash", "-c", bindir+"psql --dbname="+DEFAULTPSQLDB+" "+query)
	out, err := cmd.Output()
	chke(err)

	if strings.Contains(string(out), "f") {
		// val is already false
	} else {
		exists = true
	}
	return exists

}

func loadhDB() {
	const (
		FAIL = `Aborting Could not find database data. Make sure it is named 'C3%sC0' and resides 
in either the same directory as the application or at 'C3%sC0'`
		RESTORE = `pg_restore -v --format=directory --username=%s --dbname=%s %s`
		WARN    = "the database will start loading in %d seconds; this will take a while"
		DELAY   = 3
		ERR     = "there were errors when reloading the data; they are usually safe to ignore, especially if they involve 'hippa_rd'"
		OK      = "The data was loaded into the database. %s has finished setting itself up and can henceforth run normally."
	)
	var a error
	var b error

	_, a = os.Stat(HDBFOLDER)

	h, e := os.UserHomeDir()
	if e != nil {
		// how likely is this...?
		b = errors.New("cannot find UserHomeDir")
	} else {
		_, b = os.Stat(h + "/" + HDBFOLDER)
	}

	var fn string

	notfound := (a != nil) && (b != nil)
	if notfound {
		fmt.Println(coloroutput(fmt.Sprintf(FAIL, HDBFOLDER, h+"/"+HDBFOLDER)))
		os.Exit(0)
	} else {
		if a != nil {
			fn = HDBFOLDER
		} else {
			fn = h + "/" + HDBFOLDER
		}
	}

	msg(fmt.Sprintf(WARN, DELAY), MSGCRIT)
	time.Sleep(DELAY * time.Second)

	bindir := findpsql()
	rest := fmt.Sprintf(RESTORE, DEFAULTPSQLUSER, DEFAULTPSQLDB, fn)
	cmd := exec.Command("bash", "-c", bindir+rest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		msg(ERR, MSGCRIT)
	}

	msg(fmt.Sprintf(OK, MYNAME), MSGCRIT)
}
