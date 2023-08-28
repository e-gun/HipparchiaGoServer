//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"text/template"
)

type CurrentConfiguration struct {
	Authenticate    bool
	BadChars        string
	BlackAndWhite   bool
	BrowserCtx      int
	CustomCSS       bool
	DbDebug         bool
	DefCorp         map[string]bool
	EchoLog         int // 0: "none", 1: "terse", 2: "prolix", 3: "prolix+remoteip"
	Font            string
	Gzip            bool
	HostIP          string
	HostPort        int
	LdaGraph        bool
	LdaTopics       int
	LogLevel        int
	ManualGC        bool // see messenger.GCStats()
	MaxText         int
	MaxSrchIP       int
	MaxSrchTot      int
	PGLogin         PostgresLogin
	ProfileCPU      bool
	ProfileMEM      bool
	ResetVectors    bool
	QuietStart      bool
	SelfTest        int
	TickerActive    bool
	VectorsDisabled bool
	VectorBot       bool
	VectorMaxlines  int
	VectorModel     string
	VectorNeighb    int
	VectorTextPrep  string
	VectorWebExt    bool // "simple" when false; "expanded" when true
	VocabByCt       bool
	VocabScans      bool
	WorkerCount     int
	ZapLunates      bool
}

// LookForConfigFile - test to see if we can find a config file; if not build one and check to see if the DB needs loading
func LookForConfigFile() {
	const (
		WRN      = "Warning: unable to launch: Cannot find a configuration file."
		FYI      = "\tC1Creating configuration directory: 'C3%sC1'C0"
		FNF      = "\tC1Generating a simple 'C3%sC1'C0"
		FWR      = "\tC1Wrote configuration to 'C3%sC1'C0\n"
		PWD1     = "\tchoose a password for the database user 'hippa_wr' ->C0 "
		NODB     = "hipparchiaDB does not exist: executing InitializeHDB()"
		FOUND    = "Found 'authors': skipping database loading.\n\tIf there are problems going forward you might need to reset the database: '-00'\n\n"
		NOTFOUND = "The database exists but seems to be empty. Need to reload the data."
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
		CopyInstructions()
		_, e = os.Stat(fmt.Sprintf(CONFIGALTAPTH, h))
		if e != nil {
			fmt.Println(coloroutput(fmt.Sprintf(FYI, fmt.Sprintf(CONFIGALTAPTH, h))))
			ee := os.MkdirAll(fmt.Sprintf(CONFIGALTAPTH, h), os.FileMode(0700))
			chke(ee)
		}

		fmt.Println(coloroutput(fmt.Sprintf(FNF, CONFIGPROLIX)))
		fmt.Printf(coloroutput(PWD1))

		var hwrpw string
		_, err := fmt.Scan(&hwrpw)
		chke(err)

		pgpw := SetPostgresAdminPW()

		cfg := BuildDefaultConfig()
		cfg.PGLogin.Pass = hwrpw

		content, err := json.MarshalIndent(cfg, JSONINDENT, JSONINDENT)
		chke(err)

		err = os.WriteFile(fmt.Sprintf(CONFIGALTAPTH, h)+CONFIGPROLIX, content, 0644)
		chke(err)

		fmt.Println(coloroutput(fmt.Sprintf(FWR, fmt.Sprintf(CONFIGALTAPTH, h)+CONFIGPROLIX)))

		// do we need to head over to selfinstaller.go and to initialize the database?

		if !HipparchiaDBexists(pgpw) {
			msg(NODB, MSGCRIT)
			InitializeHDB(pgpw, hwrpw)
		}

		if HipparchiaDBHasData(hwrpw) {
			msg(FOUND, MSGCRIT)
		} else {
			msg(NOTFOUND, MSGCRIT)
			LoadhDBfolder(hwrpw)
		}
	}
}

// ConfigAtLaunch - read the configuration values from JSON and/or command line
func ConfigAtLaunch() {
	const (
		FAIL1 = "Could not parse your information as a valid collection of credentials. Use the following template:"
		FAIL2 = `"{\"Pass\": \"YOURPASSWORDHERE\" ,\"Host\": \"127.0.0.1\", \"Port\": 5432, \"DBName\": \"hipparchiaDB\" ,\"User\": \"hippa_wr\"}"`
		FAIL3 = `Could not parse the information in '%s'. Skipping and attempting to use built-in defaults instead.`
		FAIL5 = "Improperly formatted corpus list. Using:\n\t%s"
		FAIL6 = "Could not open '%s'"
		FAIL7 = "ConfigAtLaunch() failed to execute help text template"
		FAIL8 = "Cannot find current working directory"
	)

	Config = BuildDefaultConfig()

	uh, _ := os.UserHomeDir()
	h := fmt.Sprintf(CONFIGALTAPTH, uh)
	prolixcfg := fmt.Sprintf("%s/%s", h, CONFIGPROLIX)

	loadedcfg, e := os.Open(prolixcfg)
	if e != nil {
		msg(fmt.Sprintf(FAIL6, prolixcfg), MSGPEEK)
	}

	decoderc := json.NewDecoder(loadedcfg)
	confc := CurrentConfiguration{}
	errc := decoderc.Decode(&confc)
	_ = loadedcfg.Close()

	if errc == nil {
		Config = confc
	} else {
		msg(fmt.Sprintf(FAIL3, prolixcfg), MSGCRIT)
	}

	// on old CONFIGPROLIX might mean you set the following to zero; that is very bad...
	if Config.MaxSrchTot == 0 {
		// "HipparchiaGoServer -ms 1" is a perfectly sensible setting...
		Config.MaxSrchTot = MAXSEARCHTOTAL
	}

	if Config.MaxSrchIP == 0 {
		Config.MaxSrchIP = MAXSEARCHPERIPADDR
	}

	var cf string

	args := os.Args[1:len(os.Args)]

	help := func() {
		printversion()
		printbuildinfo()
		cwd, err := os.Getwd()
		if err != nil {
			msg(FAIL8, MSGCRIT)
			cwd = "(unknown)"
		}

		m := map[string]interface{}{
			"badchars":   Config.BadChars,
			"confauth":   CONFIGAUTH,
			"conffile":   CONFIGPROLIX,
			"cpus":       runtime.NumCPU(),
			"css":        CUSTOMCSSFILENAME,
			"cwd":        cwd,
			"ctxlines":   Config.BrowserCtx,
			"dbf":        HDBFOLDER,
			"echoll":     Config.EchoLog,
			"hdbf":       HDBFOLDER,
			"hgsll":      Config.LogLevel,
			"home":       h,
			"host":       Config.HostIP,
			"maxipsrch":  Config.MaxSrchIP,
			"maxtotscrh": Config.MaxSrchTot,
			"port":       Config.HostPort,
			"projurl":    PROJURL,
			"vmodel":     Config.VectorModel,
			"workers":    Config.WorkerCount}
		t := template.Must(template.New("").Parse(HELPTEXTTEMPLATE))
		var b bytes.Buffer
		if err := t.Execute(&b, m); err != nil {
			msg(FAIL7, MSGCRIT)
		}
		fmt.Println(styleoutput(coloroutput(b.String())))

		os.Exit(0)
	}

	for i, a := range args {
		switch a {
		case "-vv":
			printversion()
			printbuildinfo()
			os.Exit(1)
		case "-v":
			fmt.Println(VERSION + VersSuppl)
			os.Exit(1)
		case "-au":
			Config.Authenticate = true
		case "-av":
			Config.VectorBot = true
		case "-bc":
			bc, err := strconv.Atoi(args[i+1])
			chke(err)
			Config.BrowserCtx = bc
		case "-bw":
			Config.BlackAndWhite = true
		case "-cs":
			Config.CustomCSS = true
		case "-db":
			Config.DbDebug = true
		case "-dv":
			Config.VectorsDisabled = true
		case "-ex":
			ArchiveDB()
			os.Exit(0)
		case "-el":
			ll, err := strconv.Atoi(args[i+1])
			chke(err)
			Config.EchoLog = ll
		case "-ft":
			Config.Font = args[i+1]
		case "-gl":
			ll, err := strconv.Atoi(args[i+1])
			chke(err)
			Config.LogLevel = ll
		case "-gz":
			Config.Gzip = true
		case "-h":
			help()
		case "-md":
			Config.VectorModel = args[i+1]
		case "-mi":
			mi, err := strconv.Atoi(args[i+1])
			chke(err)
			Config.MaxSrchIP = mi
		case "-ms":
			ms, err := strconv.Atoi(args[i+1])
			chke(err)
			Config.MaxSrchTot = ms
		case "-pc":
			Config.ProfileCPU = true
		case "-pd":
			CopyInstructions()
		case "-pg":
			js := args[i+1]
			var pl PostgresLogin
			err := json.Unmarshal([]byte(js), &pl)
			if err != nil {
				msg(FAIL1, MSGMAND)
				msg(FAIL2, MSGCRIT)
			}
			Config.PGLogin = pl
		case "-pm":
			Config.ProfileMEM = true
		case "-q":
			Config.QuietStart = true
		case "-rl":
			ReLoadDBfolder(Config.PGLogin.Pass)
		case "-rv":
			Config.ResetVectors = true
		case "-sa":
			Config.HostIP = args[i+1]
		case "-sp":
			p, err := strconv.Atoi(args[i+1])
			chke(err)
			Config.HostPort = p
		case "-st":
			Config.SelfTest += 1
		case "-tk":
			Config.TickerActive = true
		case "-ui":
			Config.BadChars = args[i+1]
		case "-wc":
			wc, err := strconv.Atoi(args[i+1])
			chke(err)
			Config.WorkerCount = wc
		case "-zl":
			Config.ZapLunates = true
		case "-00":
			DBSelfDestruct()
			os.Exit(0)
		default:
			// do nothing
		}
	}

	y := ""
	if errc != nil {
		y = " *not*"
	}
	msg(fmt.Sprintf("'%s%s'%s loaded", h, CONFIGPROLIX, y), MSGTMI)

	SetConfigPass(confc, cf)

	if Config.Authenticate {
		BuildUserPassPairs()
	}

	if Config.VectorMaxlines == 0 {
		Config.VectorMaxlines = VECTORMAXLINES
	}
}

// BuildDefaultConfig - return a CurrentConfiguration filled out with various default values
func BuildDefaultConfig() CurrentConfiguration {
	var c CurrentConfiguration
	c.Authenticate = false
	c.BadChars = UNACCEPTABLEINPUT
	c.BlackAndWhite = BLACKANDWHITE
	c.BrowserCtx = DEFAULTBROWSERCTX
	c.CustomCSS = false
	c.DbDebug = false
	c.Font = FONTSETTING
	c.Gzip = USEGZIP
	c.HostIP = SERVEDFROMHOST
	c.HostPort = SERVEDFROMPORT
	c.LdaTopics = LDATOPICS
	c.LdaGraph = false
	c.ManualGC = true
	c.MaxText = MAXTEXTLINEGENERATION
	c.MaxSrchIP = MAXSEARCHPERIPADDR
	c.MaxSrchTot = MAXSEARCHTOTAL
	c.ProfileCPU = false
	c.ProfileMEM = false
	c.QuietStart = false
	c.ResetVectors = false
	c.SelfTest = 0
	c.TickerActive = TICKERISACTIVE
	c.VectorBot = false
	c.VectorMaxlines = VECTORMAXLINES
	c.VectorModel = VECTORMODELDEFAULT
	c.VectorNeighb = VECTORNEIGHBORS
	c.VectorsDisabled = false
	c.VectorTextPrep = VECTORTEXTPREPDEFAULT
	c.VectorWebExt = VECTROWEBEXTDEFAULT
	c.VocabByCt = VOCABBYCOUNT
	c.VocabScans = VOCABSCANSION
	c.WorkerCount = runtime.NumCPU()
	c.ZapLunates = false
	e := json.Unmarshal([]byte(DEFAULTCORPORA), &c.DefCorp)
	if e != nil {
		fmt.Println("BuildDefaultConfig() could not json.Unmarshal DEFAULTCORPORA: " + DEFAULTCORPORA)
	}

	pl := PostgresLogin{
		Host:   DEFAULTPSQLHOST,
		Port:   DEFAULTPSQLPORT,
		User:   DEFAULTPSQLUSER,
		Pass:   "",
		DBName: DEFAULTPSQLDB,
	}

	c.PGLogin = pl

	return c
}

// SetConfigPass - make sure that Config.PGLogin.Pass != ""
func SetConfigPass(cfg CurrentConfiguration, cf string) {
	const (
		FAIL3     = "FAILED to load database credentials from any of '%s', '%s' or '%s'"
		FAIL4     = "At a minimum be sure that a 'hgs-conf.json' file exists and that it has the following format:"
		FAIL6     = "Could not open '%s'"
		BLANKPASS = "PostgreSQLPassword is blank. Check your 'hgs-conf.json' file. NB: 'PostgreSQLPassword ≠ 'PosgreSQLPassword'.\n"
	)
	type ConfigFile struct {
		PostgreSQLPassword string
	}

	uh, _ := os.UserHomeDir()
	h := fmt.Sprintf(CONFIGALTAPTH, uh)

	if cf == "" {
		cf = fmt.Sprintf("%s/%s", CONFIGLOCATION, CONFIGBASIC)
	}

	acf := fmt.Sprintf("%s/%s", h, CONFIGBASIC)

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
		if erra != nil && errb != nil && cfg.PGLogin.DBName == "" {
			msg(fmt.Sprintf(FAIL3, cf, acf, fmt.Sprintf("%s/%s", h, CONFIGPROLIX)), MSGCRIT)
			msg(fmt.Sprintf(FAIL4), MSGCRIT)
			fmt.Printf(MINCONFIG)
			messenger.ExitOrHang(0)
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
}

// CopyInstructions - write the embedded PDF to the filesystem
func CopyInstructions() {
	const (
		FYI  = "Writing instruction files to the current working directory."
		MACI = "HGS_INSTALLATION_MacOS.pdf"
		WINI = "HGS_INSTALLATION_Windows.pdf"
		NIXI = "HGS_INSTALLATION_Nix.pdf"
		CUST = "HGS_CUSTOMIZATION.pdf"
		SEMV = "HGS_SEMANTICVECTORS.pdf"
		FYIF = "HGS_FYI.pdf"
		BASF = "HGS_BASIC_USE.pdf"
		FNF  = "CopyInstructions(): Embedded PDF not found. This function will now return."
	)

	var f string

	goos := runtime.GOOS
	switch goos {
	case "darwin":
		f = MACI
	case "windows":
		f = WINI
	case "linux":
		f = NIXI
	default:
		f = ""
	}

	if f != "" {
		data, err := efs.ReadFile(EPD + f)
		if err != nil {
			msg(FNF, MSGWARN)
			return
		}

		msg(FYI, MSGCRIT)

		err = os.WriteFile(f, data, WRITEPERMS)
		if err != nil {
			msg(FNF, MSGWARN)
			return
		}
		msg(fmt.Sprintf("\t\tWrote:\t'%s'", f), MSGCRIT)
	}

	for _, info := range []string{CUST, FYIF, SEMV, BASF} {
		data, err := efs.ReadFile(EPD + info)
		if err != nil {
			return
		}
		err = os.WriteFile(info, data, WRITEPERMS)
		if err != nil {
			msg(FNF, MSGWARN)
			return
		}
		msg(fmt.Sprintf("\t\tWrote:\t'%s'", info), MSGCRIT)
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
