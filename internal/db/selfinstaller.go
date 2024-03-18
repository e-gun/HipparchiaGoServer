//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package db

import (
	"errors"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"os"
	"os/exec"
	"runtime"
	"time"
)

//
// if LookForConfigFile() does not find a hgs-config.json or hgs-prolix-vv.json,
// it will generate a basic hgs-config.json and then call the functions below:
// is there a database? does it have data in it? are we able to load data into an empty database?
//

// InitializeHDB - insert the hipparchiaDB table and its user into postgres
func InitializeHDB(pgpw string, hdbpw string) {
	const (
		// C1   = `CREATE ROLE %s LOGIN ENCRYPTED PASSWORD '%s' NOSUPERUSER INHERIT CREATEDB NOCREATEROLE NOREPLICATION;`
		C1 = `
			DO
			$do$
			BEGIN
			   IF EXISTS (
				  SELECT FROM pg_catalog.pg_roles
				  WHERE  rolname = '%s') THEN
			
				  RAISE NOTICE 'Role "%s" already exists. Skipping.';
			   ELSE
				  CREATE ROLE %s LOGIN ENCRYPTED PASSWORD '%s' NOSUPERUSER INHERIT CREATEDB NOCREATEROLE NOREPLICATION;
			   END IF;
			END
			$do$;`
		C2 = `CREATE DATABASE "%s" WITH OWNER = %s ENCODING = 'UTF8';`
		// next is not allowed: "ERROR:  CREATE DATABASE cannot be executed from a function"
		//		C2 = `
		//DO
		//$do$
		//BEGIN
		//   IF EXISTS (
		//	  SELECT FROM pg_database
		//      WHERE datname = '%s' ) THEN
		//
		//      RAISE NOTICE 'DB "%s" already exists. Skipping.';
		//   ELSE
		//      CREATE DATABASE "%s" WITH OWNER = %s ENCODING = 'UTF8';
		//   END IF;
		//END
		//$do$;
		//`
		C3   = `CREATE EXTENSION IF NOT EXISTS pg_trgm;`
		DONE = "Initialized the database framework"
	)

	queries := []string{
		fmt.Sprintf(C1, vv.DEFAULTPSQLUSER, vv.DEFAULTPSQLUSER, vv.DEFAULTPSQLUSER, hdbpw),
		fmt.Sprintf(C2, vv.DEFAULTPSQLDB, vv.DEFAULTPSQLUSER),
		fmt.Sprintf(C3),
	}

	binary := GetBinaryPath("psql")
	url := GetPostgresURI(pgpw)

	for q := range queries {
		// this has to be looped because "CREATE DATABASE cannot run inside a transaction block"
		cmd := exec.Command(binary, "-c", queries[q], url)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		Msg.EC(err)
	}

	Msg.CRIT(DONE)
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
		WARN  = "The database will start loading in %d seconds. C7This will take several minutesC0"
		DELAY = 8
		ERR   = "There were errors when reloading the data.\n\tIt is safe to ignore errors that involve 'hippa_rd'"
		OK    = "The data was loaded into the database.\n\t%s has finished setting itself up\n\tand can henceforth run normally."
	)
	var a error
	var b error

	_, a = os.Stat(vv.HDBFOLDER)

	h, e := os.UserHomeDir()
	if e != nil {
		// how likely is this...?
		b = errors.New("cannot find UserHomeDir")
	} else {
		_, b = os.Stat(h + "/" + vv.HDBFOLDER)
	}

	var fn string

	notfound := (a != nil) && (b != nil)
	if notfound {
		fmt.Println(Msg.Color(fmt.Sprintf(FAIL, vv.HDBFOLDER, vv.MYNAME, h+"/"+vv.HDBFOLDER)))
		hd, err := os.UserHomeDir()
		Msg.EC(err)

		fp := fmt.Sprintf(vv.CONFIGALTAPTH, hd) + vv.CONFIGPROLIX
		_ = os.Remove(fp)
		fp = fmt.Sprintf(vv.CONFIGALTAPTH, hd) + vv.CONFIGBASIC
		_ = os.Remove(fp)

		fmt.Println()
		fmt.Println(Msg.Color(fmt.Sprintf(FAIL2, fp, pw)))
		os.Exit(0)
	} else {
		if a != nil {
			fn = vv.HDBFOLDER
		} else {
			fn = h + "/" + vv.HDBFOLDER
		}
	}

	fmt.Println(Msg.Color(fmt.Sprintf(WARN, DELAY)))
	time.Sleep(DELAY * time.Second)

	binary := GetBinaryPath("pg_restore")
	url := GetHippaWRURI(pw)

	// https://stackoverflow.com/questions/28324711/in-pg-restore-how-can-you-use-a-postgres-connection-string-to-specify-the-host
	// this shows you the non-parallel syntax for calling pg_restore
	cmd := exec.Command(binary, "-d", url, "-v", "-F", "directory", fn)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		Msg.CRIT(ERR)
	}

	Msg.CRIT(fmt.Sprintf(OK, vv.MYNAME))
	fmt.Println()
}

// ReLoadDBfolder - drop hipparchiaDB; then LoadhDBfolder
func ReLoadDBfolder(pw string) {
	const (
		C1 = `DROP DATABASE "%s";`
		C2 = `CREATE DATABASE "%s" WITH OWNER = %s ENCODING = 'UTF8';`
	)

	ok := youhavebeenwarned()
	if !ok {
		return
	}

	queries := []string{
		fmt.Sprintf(C1, vv.DEFAULTPSQLDB),
		fmt.Sprintf(C2, vv.DEFAULTPSQLDB, vv.DEFAULTPSQLUSER),
	}

	for q := range queries {
		pgpw := SetPostgresAdminPW()
		binary := GetBinaryPath("psql")
		url := GetPostgresURI(pgpw)
		cmd := exec.Command(binary, "-c", queries[q], url)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			// do nothing: "does not exist" error in all likelihood
		}
	}

	LoadhDBfolder(pw)
}

// SetPostgresAdminPW - ask for the password for the postgres admin user
func SetPostgresAdminPW() string {
	const (
		PWD2 = "C2I also need the database password for the postgres administrator ->C0 "
	)
	var pgpw string
	if runtime.GOOS != "darwin" {
		// macos users have admin access already (on their primary account...) and do not need a pg admin password
		fmt.Printf(Msg.Color(PWD2))
		_, ee := fmt.Scan(&pgpw)
		Msg.EC(ee)
	}
	return pgpw
}

// ArchiveDB - dump the database to the filesystem
func ArchiveDB() {
	const (
		MSG   = "Extracting the database.."
		ERR   = "ArchiveDB(): pg_dump failed. You should NOT trust this archive. Deleting it..."
		WRK   = 1 // problem (on virtualized machine): "server closed the connection unexpectedly" if WRK > 1
		WARN  = "The database will start archiving in %d seconds. C7This will take several minutesC0"
		DELAY = 5
	)

	fmt.Println(Msg.Color(fmt.Sprintf(WARN, DELAY)))
	time.Sleep(DELAY * time.Second)

	// pg_dump --clean "hipparchiaDB" --user hippa_wr | split -b 100m - out/hipparchiaDB-
	// pg_dump -U postgres -F d -j 5 db1 -f db1_backup

	// don't want an extra 1GB... should run with "-rv" flag before doing "-ex", but maybe you didn't
	// unable to call "vectordbreset()" at this juncture
	// panic: runtime error: invalid memory address or nil pointer dereference

	// highly likely that you do not have a value for Config.PGLogin.Pass yet, but you need one...
	lnch.SetConfigPass(lnch.Config, "")

	binary := GetBinaryPath("pg_dump")
	url := GetHippaWRURI(lnch.Config.PGLogin.Pass)

	workers := fmt.Sprintf("%d", WRK)

	cmd := exec.Command(binary, "-v", "-T", vv.VECTORTABLENAMENN, "-F", "d", "-j", workers, "-f", vv.HDBFOLDER, url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	Msg.CRIT(MSG)
	err := cmd.Run()
	if err != nil {
		Msg.CRIT(ERR)
		e := os.RemoveAll(vv.HDBFOLDER)
		Msg.EC(e)
	}
}

// DBSelfDestruct - purge all data and undo everything InitializeHDB and LoadhDBfolder did
func DBSelfDestruct() {
	const (
		C1    = `DROP DATABASE "%s";`
		C2    = `DROP ROLE %s;`
		C3    = `DROP EXTENSION pg_trgm;`
		DONE1 = "Deleted the database framework"
		DONE2 = "Deleted configuration files inside '%s'"
	)

	ok := youhavebeenwarned()
	if !ok {
		return
	}

	queries := []string{
		fmt.Sprintf(C1, vv.DEFAULTPSQLDB),
		fmt.Sprintf(C2, vv.DEFAULTPSQLUSER),
		C3,
	}

	pgpw := SetPostgresAdminPW()
	binary := GetBinaryPath("psql")
	url := GetPostgresURI(pgpw)

	for q := range queries {
		cmd := exec.Command(binary, "-c", queries[q], url)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			// do nothing: "does not exist" error in all likelihood
		}
	}

	Msg.CRIT(DONE1)

	hd, e := os.UserHomeDir()
	Msg.EC(e)
	cp := fmt.Sprintf(vv.CONFIGALTAPTH, hd)
	_ = os.Remove(cp + vv.CONFIGBASIC)
	_ = os.Remove(cp + vv.CONFIGPROLIX)
	Msg.CRIT(fmt.Sprintf(DONE2, cp))
}

// GetBinaryPath - return the path of a psql or pg_restore binary
func GetBinaryPath(command string) string {
	const (
		MACPGAPP = "/Applications/Postgres.app/Contents/Versions/%d/bin/"
		MACPGFD  = "/Applications/Postgres.app"
		MACBREW  = "/opt/homebrew/opt/postgresql@%d/bin/"
		WINPGEXE = `C:\Program Files\PostgreSQL\%d\bin\`
		LNXBIN   = `/usr/bin/`
		LNXLBIN  = `/usr/local/bin/`
		FAIL     = "Cannot find PostgreSQL binaries: aborting"
	)

	bindir := ""
	suffix := ""

	// linux and freebsd need fewer checks
	if runtime.GOOS == "linux" || runtime.GOOS == "freebsd" {
		_, y := os.Stat(LNXBIN + command)
		if y == nil {
			// != nil will trigger a fail later
			return LNXBIN + command
		}
		_, y = os.Stat(LNXLBIN + command)
		if y == nil {
			// != nil will trigger a fail later
			return LNXLBIN + command
		}
	}

	// mac and windows are entangled with versioning issues
	if runtime.GOOS == "darwin" {
		_, y := os.Stat(MACPGFD)
		if y == nil {
			bindir = MACPGAPP
		} else {
			bindir = MACBREW
		}
	} else if runtime.GOOS == "windows" {
		bindir = WINPGEXE
		suffix = ".exe"
	}

	vers := 0

	for i := 21; i > 12; i-- {
		_, y := os.Stat(fmt.Sprintf(bindir, i) + command + suffix)
		if y == nil {
			vers = i
			break
		}
	}

	if vers == 0 {
		Msg.CRIT(FAIL)
		os.Exit(0)
	}

	bindir = fmt.Sprintf(bindir, vers)
	return bindir + command + suffix
}

// GetPostgresURI - return a URI to connect to postgres as an administrator; different URI for macOS vs others
func GetPostgresURI(pgpw string) string {
	const (
		UPWD = `postgresql://%s:%s@%s:%d/%s`
		UBLK = `postgresql://%s:%d/%s`
	)
	var url string
	if runtime.GOOS == "darwin" {
		// macos users have admin access already (on their primary account...) and do not need a pg admin password
		// postgresql://localhost:5432/postgres
		url = fmt.Sprintf(UBLK, vv.DEFAULTPSQLHOST, vv.DEFAULTPSQLPORT, "postgres")
	} else {
		// postgresql://postgres:password@localhost:5432/postgres
		url = fmt.Sprintf(UPWD, "postgres", pgpw, vv.DEFAULTPSQLHOST, vv.DEFAULTPSQLPORT, "postgres")
	}
	return url
}

// GetHippaWRURI - return a URI to connect to postgres as DEFAULTPSQLUSER
func GetHippaWRURI(pw string) string {
	const (
		U = `postgresql://%s:%s@%s:%d/%s`
	)
	return fmt.Sprintf(U, vv.DEFAULTPSQLUSER, pw, vv.DEFAULTPSQLHOST, vv.DEFAULTPSQLPORT, vv.DEFAULTPSQLDB)
}

func youhavebeenwarned() bool {
	const (
		CONF = `You are about to C5RESETC0 the database this program uses.
The application will be C7NON-FUNCTIONALC0 after this unless/until you reload 
this data. 

In short, this very dangerous. 

Type C6YESC0 to confirm that you want to proceed. --> `
		NOPE = "Did not receive confirmation. Aborting..."
	)

	yes := true

	var ok string
	fmt.Printf(Msg.Color(CONF))
	_, ee := fmt.Scan(&ok)
	Msg.EC(ee)
	if ok != "YES" {
		fmt.Println(NOPE)
		yes = false
	}

	return yes
}
