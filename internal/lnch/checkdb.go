//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package lnch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

//
// if LookForConfigFile() does not find a hgs-config.json or hgs-prolix-conf.json,
// it will generate a basic hgs-config.json and then call the functions below:
// is there a database? does it have data in it? are we able to load data into an empty database?
//

func PGFSConfig(h string) {
	Msg.TMI("PGFSConfig()")
	const (
		WRN      = "Warning: unable to lnch: Cannot find a configuration file."
		FYI      = "\tC1Creating configuration directory: 'C3%sC1'C0"
		FNF      = "\tC1Generating a simple 'C3%sC1'C0"
		FWR      = "\tC1Wrote configuration to 'C3%sC1'C0\n"
		PWD1     = "\tchoose a password for the database user 'hippa_wr' ->C0 "
		NODB     = "hipparchiaDB does not exist: executing InitializeHDB()"
		YESDB    = "hipparchiaDB already exists"
		FOUND    = "Found 'authors': skipping database loading.\n\tIf there are problems going forward you might need to reset the database: '-00'\n\n"
		NOTFOUND = "The database exists but seems to be empty. Need to reload the data."
	)

	Msg.CRIT(WRN)
	CopyInstructions()
	_, e := os.Stat(fmt.Sprintf(vv.CONFIGALTAPTH, h))
	if e != nil {
		fmt.Println(Msg.Color(fmt.Sprintf(FYI, fmt.Sprintf(vv.CONFIGALTAPTH, h))))
		ee := os.MkdirAll(fmt.Sprintf(vv.CONFIGALTAPTH, h), os.FileMode(0700))
		Msg.EC(ee)
	}

	fmt.Println(Msg.Color(fmt.Sprintf(FNF, vv.CONFIGPROLIX)))
	fmt.Printf(Msg.Color(PWD1))

	var hwrpw string
	_, err := fmt.Scan(&hwrpw)
	Msg.EC(err)

	pgpw := RequestPostgresAdminPW()

	cfg := BuildDefaultConfig()
	cfg.PGLogin.Pass = hwrpw

	content, err := json.MarshalIndent(cfg, vv.JSONINDENT, vv.JSONINDENT)
	Msg.EC(err)

	err = os.WriteFile(fmt.Sprintf(vv.CONFIGALTAPTH, h)+vv.CONFIGPROLIX, content, 0644)
	Msg.EC(err)

	fmt.Println(Msg.Color(fmt.Sprintf(FWR, fmt.Sprintf(vv.CONFIGALTAPTH, h)+vv.CONFIGPROLIX)))

	// do we need to head over to selfinstaller.go and to initialize the database?

	dbe := HipparchiaDBexists(pgpw)
	if !dbe {
		Msg.CRIT(NODB)
		InitializeHDB(pgpw, hwrpw)
	} else {
		Msg.CRIT(YESDB)
	}

	dbd := HipparchiaDBHasData(hwrpw)
	if dbd {
		Msg.CRIT(FOUND)
	} else {
		Msg.CRIT(NOTFOUND)
		LoadhDBfolder(hwrpw)
	}
}

// LoadhDBfolder - take a psql dump and `pg_restore` it by exec-ing the binary
func LoadhDBfolder(pw string) {
	Msg.TMI("LoadhDBfolder()")
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

	binary := GetPGBinaryPath("pg_restore")
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
	Msg.TMI("ReLoadDBfolder()")
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
		pgpw := RequestPostgresAdminPW()
		binary := GetPGBinaryPath("psql")
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

// InitializeHDB - insert the hipparchiaDB table and its user into postgres
func InitializeHDB(pgpw string, hdbpw string) {
	Msg.TMI("InitializeHDB()")
	const (
		C0 = `CREATE ROLE %s LOGIN ENCRYPTED PASSWORD '%s' NOSUPERUSER INHERIT CREATEDB NOCREATEROLE NOREPLICATION;`
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

	binary := GetPGBinaryPath("psql")
	url := GetPostgresURI(pgpw)

	for q := range queries {
		// this has to be looped because "CREATE DATABASE cannot run inside a transaction block"
		// Msg.TMI(queries[q])
		cmd := exec.Command(binary, "-c", queries[q], url)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		Msg.EC(err)
	}

	Msg.CRIT(DONE)
}

// HipparchiaDBexists - does psql have hipparchiaDB in it yet?
func HipparchiaDBexists(pgpw string) bool {
	Msg.TMI("HipparchiaDBexists()")
	const (
		Q = `SELECT datname FROM pg_database WHERE datname='%s';`
	)

	// WARNING: passwords will be visible to `ps`, etc.; there are very few scenarios in which this might matter
	// people in a multi-user server setting are not really supposed to be using this mechanism to do admin

	binary := GetPGBinaryPath("psql")
	url := GetPostgresURI(pgpw)

	exists := false

	// windows wants "-c, url"; not "url, -c"
	cmd := exec.Command(binary, "-c", fmt.Sprintf(Q, vv.DEFAULTPSQLDB), url)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	Msg.EC(err)

	// empty looks like:
	// "(0 rows)"
	if strings.Contains(string(out), vv.DEFAULTPSQLDB) {
		exists = true
	}

	return exists
}

// HipparchiaDBHasData - true if an exec of `psql` finds `authors` in `pg_tables`
func HipparchiaDBHasData(userpw string) bool {
	Msg.MAND("HipparchiaDBHasData")
	const (
		Q        = `SELECT COUNT(universalid) FROM authors;`
		AUTH     = `Authentication failed: this likely means that neither the user nor the database exists. Consider deleting any configuration file inside '%s'`
		CHKFAIL  = "authentication failed"
		CHKEXIST = "does not exist"
	)

	// WARNING: passwords will be visible to `ps`, etc.

	stderr := new(bytes.Buffer)

	binary := GetPGBinaryPath("psql")
	url := GetHippaWRURI(userpw)
	cmd := exec.Command(binary, "-c", Q, url)
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		// we actually expect the error "exit status 1" when the query looks for a table that is not there
	}

	check := stderr.String()
	var found bool

	if strings.Contains(check, CHKEXIST) {
		found = false
	} else {
		found = true
	}

	if strings.Contains(check, CHKFAIL) {
		hd, e := os.UserHomeDir()
		Msg.EC(e)
		Msg.CRIT(fmt.Sprintf(AUTH, fmt.Sprintf(vv.CONFIGALTAPTH, hd)))
	}
	return found
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
	SetConfigPass(Config, "")

	binary := GetPGBinaryPath("pg_dump")
	url := GetHippaWRURI(Config.PGLogin.Pass)

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

	pgpw := RequestPostgresAdminPW()
	binary := GetPGBinaryPath("psql")
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

// RequestPostgresAdminPW - ask for the password for the postgres admin user
func RequestPostgresAdminPW() string {
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

// GetPGBinaryPath - return the path of a psql or pg_restore binary
func GetPGBinaryPath(command string) string {
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
