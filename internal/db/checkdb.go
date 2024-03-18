package db

import (
	"bytes"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/launch"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"os"
	"os/exec"
	"strings"
)

var Msg = launch.NewMessageMakerWithDefaults()

// HipparchiaDBexists - does psql have hipparchiaDB in it yet?
func HipparchiaDBexists(pgpw string) bool {
	const (
		Q = `SELECT datname FROM pg_database WHERE datname='%s';`
	)

	// WARNING: passwords will be visible to `ps`, etc.; there are very few scenarios in which this might matter
	// people in a multi-user server setting are not really supposed to be using this mechanism to do admin

	binary := GetBinaryPath("psql")
	url := GetPostgresURI(pgpw)

	exists := false

	// windows wants "-c, url"; not "url, -c"
	cmd := exec.Command(binary, "-c", fmt.Sprintf(Q, vv.DEFAULTPSQLDB), url)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	Msg.EC(err)
	if strings.Contains(string(out), vv.DEFAULTPSQLDB) {
		exists = true
	}

	return exists
}

// HipparchiaDBHasData - true if an exec of `psql` finds `authors` in `pg_tables`
func HipparchiaDBHasData(userpw string) bool {
	const (
		Q        = `SELECT COUNT(universalid) FROM authors;`
		AUTH     = `Authentication failed: this likely means that neither the user nor the database exists. Consider deleting any configuration file inside '%s'`
		CHKFAIL  = "authentication failed"
		CHKEXIST = "does not exist"
	)

	// WARNING: passwords will be visible to `ps`, etc.

	stderr := new(bytes.Buffer)

	binary := GetBinaryPath("psql")
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
