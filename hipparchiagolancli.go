//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"C"
	"fmt"
	"os"
	"strconv"
)

const (
	MYNAME        = "Hipparchia Golang Server"
	SHORTNAME     = "HGS"
	VERSION       = "0.0.9"
	PSQ           = `{"Host": "localhost", "Port": 5432, "User": "hippa_wr", "Pass": "", "DBName": "hipparchiaDB"}`
	PSDefaultHost = "localhost"
	PSDefaultUser = "hippa_wr"
	PSDefaultPort = 5432
	PSDefaultDB   = "hipparchiaDB"
	TwoPassThresh = 100 // cicero has >70 works
)

func main() {
	// the command line arguments get lost after the invocation of makeconfig() via the first grabpgsqlconnection()
	// and you are not allowed to reset them

	// main() instead has a cfg with the defaults burned into it
	// so we do this the stupic/bound to fail way...

	// fmt.Println(os.Args[1:len(os.Args)])

	args := os.Args[1:len(os.Args)]

	for i, a := range args {
		fmt.Println(a)
		if a == "-gl" {
			ll, e := strconv.Atoi(args[i+1])
			checkerror(e)
			cfg.LogLevel = ll
		}
		if a == "-el" {
			ll, e := strconv.Atoi(args[i+1])
			checkerror(e)
			cfg.EchoLog = ll
		}
	}

	versioninfo := fmt.Sprintf("%s CLI Debugging Interface (v.%s)", MYNAME, VERSION)
	versioninfo = versioninfo + fmt.Sprintf(" [loglevel=%d]", cfg.LogLevel)

	if cfg.LogLevel > 5 {
		cfg.LogLevel = 5
	}

	if cfg.LogLevel < 0 {
		cfg.LogLevel = 0
	}

	cfg.PGLogin = decodepsqllogin([]byte(cfg.PosgresInfo))

	fmt.Println("cfg.EchoLog")
	fmt.Println(cfg.EchoLog)

	fmt.Println(versioninfo)

	StartEchoServer()
}
