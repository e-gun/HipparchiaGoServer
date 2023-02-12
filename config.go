//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
)

type CurrentConfiguration struct {
	Authenticate  bool
	BadChars      string
	BlackAndWhite bool
	BrowserCtx    int
	DbDebug       bool
	DefCorp       map[string]bool
	EchoLog       int // "none", "terse", "verbose"
	Font          string
	Gzip          bool
	HostIP        string
	HostPort      int
	LogLevel      int
	ManualGC      bool // see GCStats()
	MaxText       int
	PGLogin       PostgresLogin
	QuietStart    bool
	SelfTest      bool
	VocabByCt     bool
	VocabScans    bool
	WorkerCount   int
	ZapLunates    bool
}

// configatlaunch - read the configuration values from JSON and/or command line
func configatlaunch() {
	const (
		FAIL1     = "Could not parse your information as a valid collection of credentials. Use the following template:"
		FAIL2     = `"{\"Pass\": \"YOURPASSWORDHERE\" ,\"Host\": \"127.0.0.1\", \"Port\": 5432, \"DBName\": \"hipparchiaDB\" ,\"User\": \"hippa_wr\"}"`
		FAIL3     = "FAILED to load database credentials from any of '%s', '%s' or '%s'"
		FAIL4     = "Ata a minimum sure that a 'hgs-conf.json' file exists and that it has the following format:"
		FAIL5     = "Improperly formatted corpus list. Using:\n\t%s"
		FAIL6     = "Could not open '%s'"
		BLANKPASS = "PostgreSQLPassword is blank. Check your 'hgs-conf.json' file. NB: 'PostgreSQLPassword ≠ 'PosgreSQLPassword'.\n"
	)

	Config.Authenticate = false
	Config.BadChars = UNACCEPTABLEINPUT
	Config.BlackAndWhite = BLACKANDWHITE
	Config.BrowserCtx = DEFAULTBROWSERCTX
	Config.DbDebug = false
	Config.Font = FONTSETTING
	Config.Gzip = USEGZIP
	Config.HostIP = SERVEDFROMHOST
	Config.HostPort = SERVEDFROMPORT
	Config.ManualGC = true
	Config.MaxText = MAXTEXTLINEGENERATION
	Config.QuietStart = false
	Config.SelfTest = false
	Config.VocabByCt = VOCABBYCOUNT
	Config.VocabScans = VOCABSCANSION
	Config.WorkerCount = runtime.NumCPU()
	Config.ZapLunates = false
	e := json.Unmarshal([]byte(DEFAULTCORPORA), &Config.DefCorp)
	chke(e)

	cf := fmt.Sprintf("%s/%s", CONFIGLOCATION, CONFIGBASIC)

	uh, _ := os.UserHomeDir()
	h := fmt.Sprintf(CONFIGALTAPTH, uh)
	acf := fmt.Sprintf("%s/%s", h, CONFIGBASIC)
	pcf := fmt.Sprintf("%s/%s", h, CONFIGPROLIX)
	pwf := fmt.Sprintf("%s%s", h, CONFIGAUTH)

	cfc, e := os.Open(pcf)
	if e != nil {
		msg(fmt.Sprintf(FAIL6, pcf), MSGPEEK)
	}
	defer func(cfc *os.File) {
		err := cfc.Close()
		if err != nil {
		} // the file was almost certainly not found in the first place...
	}(cfc)

	decoderc := json.NewDecoder(cfc)
	confc := CurrentConfiguration{}
	errc := decoderc.Decode(&confc)

	if errc == nil {
		Config = confc
	}

	var pl PostgresLogin

	args := os.Args[1:len(os.Args)]

	for i, a := range args {
		switch a {
		case "-vv":
			printversion()
			printbuidldate()
			os.Exit(1)
		case "-v":
			fmt.Println(VERSION)
			os.Exit(1)
		case "-ac":
			err := json.Unmarshal([]byte(args[i+1]), &Config.DefCorp)
			if err != nil {
				msg(fmt.Sprintf(FAIL5, DEFAULTCORPORA), MSGCRIT)
			}
		case "-au":
			Config.Authenticate = true
		case "-bc":
			bc, err := strconv.Atoi(args[i+1])
			chke(err)
			Config.BrowserCtx = bc
		case "-bw":
			Config.BlackAndWhite = true
		case "-cf":
			cf = args[i+1]
		case "-db":
			Config.DbDebug = true
		case "-ft":
			Config.Font = args[i+1]
		case "-el":
			ll, err := strconv.Atoi(args[i+1])
			chke(err)
			Config.EchoLog = ll
		case "-gl":
			ll, err := strconv.Atoi(args[i+1])
			chke(err)
			Config.LogLevel = ll
		case "-gz":
			Config.Gzip = true
		case "-h":
			printversion()
			ht := coloroutput(HELPTEXT)
			fmt.Println(fmt.Sprintf(ht, pwf, DEFAULTBROWSERCTX, CONFIGLOCATION, CONFIGBASIC, h, CONFIGBASIC,
				DEFAULTECHOLOGLEVEL, DEFAULTGOLOGLEVEL, SERVEDFROMHOST, SERVEDFROMPORT,
				UNACCEPTABLEINPUT, runtime.NumCPU(), CONFIGPROLIX, h, PROJURL))
			os.Exit(1)
		case "-pg":
			js := args[i+1]
			err := json.Unmarshal([]byte(js), &pl)
			if err != nil {
				msg(FAIL1, MSGMAND)
				msg(FAIL2, MSGCRIT)
			}
		case "-q":
			Config.QuietStart = true
		case "-sa":
			Config.HostIP = args[i+1]
		case "-sp":
			p, err := strconv.Atoi(args[i+1])
			chke(err)
			Config.HostPort = p
		case "-st":
			Config.SelfTest = true
		case "-ui":
			Config.BadChars = args[i+1]
		case "-wc":
			wc, err := strconv.Atoi(args[i+1])
			chke(err)
			Config.WorkerCount = wc
		case "-zl":
			Config.ZapLunates = true
		default:
			// do nothing
		}
	}

	y := ""
	if errc != nil {
		y = " *not*"
	}
	msg(fmt.Sprintf("'%s%s'%s loaded", h, CONFIGPROLIX, y), MSGTMI)

	type ConfigFile struct {
		PostgreSQLPassword string
	}

	if Config.PGLogin.Pass == "" {
		Config.PGLogin = PostgresLogin{}
		cfa, ee := os.Open(cf)
		if ee != nil {
			msg(fmt.Sprintf(FAIL6, cf), MSGTMI)
		}
		cfb, ee := os.Open(acf)
		if ee != nil {
			msg(fmt.Sprintf(FAIL6, acf), MSGTMI)
		}

		defer func(cfa *os.File) {
			err := cfa.Close()
			if err != nil {
			} // the file was almost certainly not found in the first place...
		}(cfa)
		defer func(cfb *os.File) {
			err := cfb.Close()
			if err != nil {
			} // the file was almost certainly not found in the first place...
		}(cfb)

		decodera := json.NewDecoder(cfa)
		confa := ConfigFile{}
		erra := decodera.Decode(&confa)

		decoderb := json.NewDecoder(cfb)
		confb := ConfigFile{}
		errb := decoderb.Decode(&confb)
		if erra != nil && errb != nil && confc.PGLogin.DBName == "" {
			msg(fmt.Sprintf(FAIL3, cf, acf, pcf), MSGCRIT)
			msg(fmt.Sprintf(FAIL4), MSGCRIT)
			fmt.Printf(MINCONFIG)
			os.Exit(0)
		}
		conf := ConfigFile{}
		if erra == nil {
			conf = confa
		} else {
			conf = confb
		}

		if conf.PostgreSQLPassword == "" {
			msg(BLANKPASS, 0)
		}

		Config.PGLogin = PostgresLogin{
			Host:   DEFAULTPSQLHOST,
			Port:   DEFAULTPSQLPORT,
			User:   DEFAULTPSQLUSER,
			DBName: DEFAULTPSQLDB,
			Pass:   conf.PostgreSQLPassword,
		}
	}

	if Config.Authenticate {
		BuildUserPassPairs()
	}
}

// BuildUserPassPairs - set up authentication map via CONFIGAUTH
func BuildUserPassPairs() {
	const (
		FAIL1 = `failed to unmarshall authorization config file`
		FAIL2 = `You are requiring authentication but there are no UserPassPairs: aborting launch`
		FAIL3 = "Could not open '%s'"
	)

	uh, _ := os.UserHomeDir()
	h := fmt.Sprintf(CONFIGALTAPTH, uh)
	pwf := fmt.Sprintf("%s%s", h, CONFIGAUTH)

	pwc, e := os.Open(pwf)
	if e != nil {
		msg(fmt.Sprintf(FAIL3, pwf), MSGCRIT)
	}
	defer func(pwc *os.File) {
		err := pwc.Close()
		if err != nil {
		} // the file was almost certainly not found in the first place...
	}(pwc)

	filebytes, _ := io.ReadAll(pwc)

	type UserPass struct {
		User string
		Pass string
	}

	var upp []UserPass
	err := json.Unmarshal(filebytes, &upp)
	if err != nil {
		msg(FAIL1, MSGNOTE)
	}

	for _, u := range upp {
		UserPassPairs[u.User] = u.Pass
	}

	if Config.Authenticate && len(UserPassPairs) == 0 {
		msg(FAIL2, MSGCRIT)
		os.Exit(1)
	}
}

//
// CONFIGURATION NOT FOUND?
//

func checkforconfiguration() {
	const (
		WRN      = "Warning: unable to launch: Cannot find a configuration file."
		FYI      = "\tC1Creating configuration directory: 'C3%sC1'C0"
		FNF      = "\tC1Generating a simple 'C3%sC1'C0"
		FWR      = "\tC1Wrote a configuration file to 'C3%sC1'C0\n"
		PWD      = "\tC2enter the password you wish to use ->C0 "
		NODB     = "hipparchiaDB does not exist: executing initializeHDB()"
		FOUND    = "Found 'authors': skipping database loading"
		NOTFOUND = "Could not find 'authors' table. Need to reload the data."
	)
	_, a := os.Stat(CONFIGBASIC)

	var b error
	var c error

	h, e := os.UserHomeDir()
	if e != nil {
		// how likely is this...?
		b = errors.New("cannot find UserHomeDir")
		c = errors.New("cannot find UserHomeDir")
	} else {
		_, b = os.Stat(fmt.Sprintf(CONFIGALTAPTH, h) + CONFIGBASIC)
		_, c = os.Stat(fmt.Sprintf(CONFIGALTAPTH, h) + CONFIGPROLIX)
	}

	notfound := (a != nil) && (b != nil) && (c != nil)

	if notfound {
		msg(WRN, MSGCRIT)

		_, e = os.Stat(fmt.Sprintf(CONFIGALTAPTH, h))
		if e != nil {
			fmt.Println(coloroutput(fmt.Sprintf(FYI, fmt.Sprintf(CONFIGALTAPTH, h))))
			ee := os.MkdirAll(fmt.Sprintf(CONFIGALTAPTH, h), os.FileMode(0700))
			chke(ee)
		}

		fmt.Println(coloroutput(fmt.Sprintf(FNF, CONFIGBASIC)))
		fmt.Printf(coloroutput(PWD))

		var pw string
		_, err := fmt.Scan(&pw)
		chke(err)

		type ConfOut struct {
			PostgreSQLPassword string
		}

		content, err := json.Marshal(ConfOut{pw})
		chke(err)

		err = os.WriteFile(fmt.Sprintf(CONFIGALTAPTH, h)+CONFIGBASIC, content, 0644)
		chke(err)

		fmt.Println(coloroutput(fmt.Sprintf(FWR, fmt.Sprintf(CONFIGALTAPTH, h)+CONFIGBASIC)))

		if hipparchiaDBexists(findpsql()) {
			// msg("hipparchiaDB exists: skipping initializeHDB()", MSGCRIT)
		} else {
			msg(NODB, MSGCRIT)
			initializeHDB(pw)
		}

		if hipparchiaDBhasdata(findpsql()) {
			msg(FOUND, MSGCRIT)
		} else {
			msg(NOTFOUND, MSGCRIT)
			loadhDB(pw)
		}
	}
}
