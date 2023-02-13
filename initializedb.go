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
	WINPGEXE  = `C:\Program Files\PostgreSQL\%d\bin\`
	HDBFOLDER = "hDB"
	DONE      = "Initialized the database framework"
)

// initializeHDB - insert the hipparchiaDB table and its user into postgres
func initializeHDB(pw string) {
	const (
		C1 = `CREATE USER %s WITH PASSWORD '%s';`
		C2 = `CREATE DATABASE "%s";`
		C3 = `ALTER DATABASE "%s" OWNER TO %s;`
		C4 = `CREATE EXTENSION pg_trgm;`
	)

	queries := []string{
		fmt.Sprintf(C1, DEFAULTPSQLUSER, pw),
		fmt.Sprintf(C2, DEFAULTPSQLDB),
		fmt.Sprintf(C3, DEFAULTPSQLDB, DEFAULTPSQLUSER),
		fmt.Sprintf(C4),
	}

	for q := range queries {
		// this has to be looped because "CREATE DATABASE cannot run inisde a transaction block"
		cmd := exec.Command(FindpPSQL()+"psql", "-d", "postgres", "-c", queries[q])
		// cmd := exec.Command("bash", "-c", bindir+"psql --dbname=postgres "+eof)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		chke(err)
	}

	msg(DONE, MSGCRIT)
}

// FindpPSQL - return the path of the psql executable
func FindpPSQL() string {
	const (
		FAIL = "Cannot find PostgreSQL binaries: aborting"
	)

	bindir := ""
	suffix := ""
	if runtime.GOOS == "darwin" {
		bindir = MACPGAPP
	} else if runtime.GOOS == "windows" {
		bindir = WINPGEXE
		suffix = ".exe"
	}

	vers := 0

	for i := 21; i > 12; i-- {
		_, y := os.Stat(fmt.Sprintf(bindir, i) + "psql" + suffix)
		if y == nil {
			vers = i
			break
		}
	}

	if vers == 0 {
		msg(FAIL, MSGCRIT)
		os.Exit(0)
	}

	bindir = fmt.Sprintf(bindir, vers)
	return bindir
}

func hipparchiaDBexists() bool {
	const (
		Q = `SELECT datname FROM pg_database WHERE datname='%s';`
	)

	binary := GetBinaryPath("psql")

	exists := false

	cmd := exec.Command(binary, "-d", "postgres", "-c", fmt.Sprintf(Q, DEFAULTPSQLDB))
	out, err := cmd.Output()
	chke(err)
	if strings.Contains(string(out), DEFAULTPSQLDB) {
		exists = true
	}

	// fmt.Printf("hipparchiaDBexists(): %t\n", exists)
	return exists
}

// HipparchiaDBHasData - true if an exec of `psql` finds `authors` in `pg_tables`
func HipparchiaDBHasData() bool {
	const (
		Q = `SELECT EXISTS ( SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename  = 'authors');`
	)

	exists := false

	binary := GetBinaryPath("psql")

	cmd := exec.Command(binary, "-d", DEFAULTPSQLDB, "-c", Q)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(out)
	}

	if strings.Contains(string(out), "f") {
		// val is already false
	} else {
		exists = true
	}

	// fmt.Printf("HipparchiaDBHasData(): %t\n", exists)
	return exists

}

// LoadhDBfolder - take a psql dump and `pg_restore` it by exec-ing the binary
func LoadhDBfolder(pw string) {
	const (
		FAIL = `ABORTING initialization: C7Could not find the folder with the database dataC0. 
Make sure that data folder is named 'C3%sC0' and resides 
EITHER in the same directory as %s 
OR at 'C3%sC0'`
		FAIL2 = `Deleting 'C3%sC0'
[You will need to reset your password when asked. Currently: 'C3%sC0']`
		RESTORE = `pg_restore -v --format=directory --username=%s --dbname=%s %s`
		WARN    = "The database will start loading in %d seconds. C7This will take several minutesC0"
		DELAY   = 8
		ERR     = "There were errors when reloading the data. It is safe to ignore errors that involve 'hippa_rd'"
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
		fmt.Println(coloroutput(fmt.Sprintf(FAIL, HDBFOLDER, MYNAME, h+"/"+HDBFOLDER)))
		hd, err := os.UserHomeDir()
		chke(err)
		fp := fmt.Sprintf(CONFIGALTAPTH, hd) + CONFIGBASIC
		_ = os.Remove(fp)
		fmt.Println()
		fmt.Println(coloroutput(fmt.Sprintf(FAIL2, fp, pw)))
		os.Exit(0)
	} else {
		if a != nil {
			fn = HDBFOLDER
		} else {
			fn = h + "/" + HDBFOLDER
		}
	}

	fmt.Println(coloroutput(fmt.Sprintf(WARN, DELAY)))
	time.Sleep(DELAY * time.Second)

	binary := GetBinaryPath("pg_restore")

	// rest := fmt.Sprintf(RESTORE, DEFAULTPSQLUSER, DEFAULTPSQLDB, fn)
	// cmd := exec.Command("bash", "-c", bindir+rest)

	// RESTORE = `pg_restore -v --format=directory --username=%s --dbname=%s %s`
	cmd := exec.Command(binary, "-v", "-F", "directory", "-U", DEFAULTPSQLUSER, "-d", DEFAULTPSQLDB, fn)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		msg(ERR, MSGCRIT)
	}

	msg(fmt.Sprintf(OK, MYNAME), MSGCRIT)
	fmt.Println()
}

func GetBinaryPath(command string) string {
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	return FindpPSQL() + command + suffix
}
