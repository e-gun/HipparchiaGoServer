//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//

package launch

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/generic"
	"github.com/e-gun/HipparchiaGoServer/internal/m"
	"github.com/e-gun/HipparchiaGoServer/internal/structs"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"os"
	"runtime"
	"strconv"
	"strings"
	"text/template"
)

//go:embed efs
var efs embed.FS

var (
	Config *structs.CurrentConfiguration
	// TODO: this is hollow
	msg = m.NewMessageMaker(BuildDefaultConfig(), m.LaunchStruct{})
	EPD = "emb/pdf/"
)

//type Poem struct {
//	content []byte
//	storage PoemStorage
//}
//type PoemStorage interface {
//	Type() string        // Return a string describing the storage type.
//	Load(string) []byte  // Load a poem by name.
//	Save(string, []byte) // Save a poem by name.
//}
//
//func NewPoem(ps PoemStorage) *Poem {
//	return &Poem{
//		content: []byte("I am a poem from a " + ps.Type() + "."),
//		storage: ps,
//	}
//}
//func (p *Poem) Save(name string) {
//	p.storage.Save(name, p.content)
//}
//
//type DBFncInjector struct {
//	F DBFncInjectorIf
//}
//type DBFncInjectorIf interface {
//	SetPostgresAdminPW() string
//	HipparchiaDBexists(s string) bool
//}
//
//func NewDBFncInj() *DBFncInjector {
//	return &DBFncInjector{}
//}
//
//func (i *DBFncInjector) iSetPostgresAdminPW() string {
//	return i.F.SetPostgresAdminPW()
//}
//func (i *DBFncInjector) iHipparchiaDBexists(s string) bool {
//	return i.F.HipparchiaDBexists(s)
//}

// LookForConfigFile - test to see if we can find a config file; if not build one and check to see if the DB needs loading
func LookForConfigFile() {
	const (
		WRN      = "Warning: unable to vv: Cannot find a configuration file."
		FYI      = "\tC1Creating configuration directory: 'C3%sC1'C0"
		FNF      = "\tC1Generating a simple 'C3%sC1'C0"
		FWR      = "\tC1Wrote configuration to 'C3%sC1'C0\n"
		PWD1     = "\tchoose a password for the database user 'hippa_wr' ->C0 "
		NODB     = "hipparchiaDB does not exist: executing InitializeHDB()"
		FOUND    = "Found 'authors': skipping database loading.\n\tIf there are problems going forward you might need to reset the database: '-00'\n\n"
		NOTFOUND = "The database exists but seems to be empty. Need to reload the data."
	)
	_, a := os.Stat(vv.CONFIGBASIC)

	var b error
	var c error

	h, e := os.UserHomeDir()
	if e != nil {
		// how likely is this...?
		b = errors.New("cannot find UserHomeDir")
		c = errors.New("cannot find UserHomeDir")
	} else {
		_, b = os.Stat(fmt.Sprintf(vv.CONFIGALTAPTH, h) + vv.CONFIGBASIC)
		_, c = os.Stat(fmt.Sprintf(vv.CONFIGALTAPTH, h) + vv.CONFIGPROLIX)
	}

	notfound := (a != nil) && (b != nil) && (c != nil)

	if notfound {
		msg.CRIT(WRN)
		CopyInstructions()
		_, e = os.Stat(fmt.Sprintf(vv.CONFIGALTAPTH, h))
		if e != nil {
			fmt.Println(msg.Color(fmt.Sprintf(FYI, fmt.Sprintf(vv.CONFIGALTAPTH, h))))
			ee := os.MkdirAll(fmt.Sprintf(vv.CONFIGALTAPTH, h), os.FileMode(0700))
			msg.EC(ee)
		}

		fmt.Println(msg.Color(fmt.Sprintf(FNF, vv.CONFIGPROLIX)))
		fmt.Printf(msg.Color(PWD1))

		var hwrpw string
		_, err := fmt.Scan(&hwrpw)
		msg.EC(err)

		pgpw := SetPostgresAdminPW()

		cfg := BuildDefaultConfig()
		cfg.PGLogin.Pass = hwrpw

		content, err := json.MarshalIndent(cfg, vv.JSONINDENT, vv.JSONINDENT)
		msg.EC(err)

		err = os.WriteFile(fmt.Sprintf(vv.CONFIGALTAPTH, h)+vv.CONFIGPROLIX, content, 0644)
		msg.EC(err)

		fmt.Println(msg.Color(fmt.Sprintf(FWR, fmt.Sprintf(vv.CONFIGALTAPTH, h)+vv.CONFIGPROLIX)))

		// do we need to head over to selfinstaller.go and to initialize the database?

		if HipparchiaDBexists(pgpw) {
			msg.CRIT(NODB)
			InitializeHDB(pgpw, hwrpw)
		}

		if HipparchiaDBHasData(hwrpw) {
			msg.CRIT(FOUND)
		} else {
			msg.CRIT(NOTFOUND)
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
		FAIL5 = "Refusing to set a workercount greater than NumCPU: %d > %d ---> setting workercount value to NumCPU: %d"
		FAIL6 = "Could not open '%s'"
		FAIL7 = "ConfigAtLaunch() failed to execute help text template"
		FAIL8 = "Cannot find current working directory"
	)

	Config = BuildDefaultConfig()

	uh, _ := os.UserHomeDir()
	h := fmt.Sprintf(vv.CONFIGALTAPTH, uh)
	prolixcfg := fmt.Sprintf("%s/%s", h, vv.CONFIGPROLIX)

	loadedcfg, e := os.Open(prolixcfg)
	if e != nil {
		msg.CRIT(fmt.Sprintf(FAIL6, prolixcfg))
	}

	decoderc := json.NewDecoder(loadedcfg)
	confc := structs.CurrentConfiguration{}
	errc := decoderc.Decode(&confc)
	_ = loadedcfg.Close()

	if errc == nil {
		Config = &confc
	} else {
		msg.CRIT(fmt.Sprintf(FAIL3, prolixcfg))
	}

	// on old CONFIGPROLIX might mean you set the following to zero; that is very bad...
	if Config.MaxSrchTot == 0 {
		// "HipparchiaGoServer -ms 1" is a perfectly sensible setting...
		Config.MaxSrchTot = vv.MAXSEARCHTOTAL
	}

	if Config.MaxSrchIP == 0 {
		Config.MaxSrchIP = vv.MAXSEARCHPERIPADDR
	}

	var cf string

	args := os.Args[1:len(os.Args)]

	help := func() {
		PrintVersion(*Config)
		printbuildinfo(*Config)
		cwd, err := os.Getwd()
		if err != nil {
			msg.CRIT(FAIL8)
			cwd = "(unknown)"
		}

		kff := generic.StringMapKeysIntoSlice(vv.ServableFonts)

		m := map[string]interface{}{
			"badchars":   Config.BadChars,
			"confauth":   vv.CONFIGAUTH,
			"conffile":   vv.CONFIGPROLIX,
			"cpus":       runtime.NumCPU(),
			"css":        vv.CUSTOMCSSFILENAME,
			"cwd":        cwd,
			"ctxlines":   Config.BrowserCtx,
			"dbf":        vv.HDBFOLDER,
			"echoll":     Config.EchoLog,
			"hdbf":       vv.HDBFOLDER,
			"hgsll":      Config.LogLevel,
			"home":       h,
			"host":       Config.HostIP,
			"maxipsrch":  Config.MaxSrchIP,
			"maxtotscrh": Config.MaxSrchTot,
			"port":       Config.HostPort,
			"projurl":    vv.PROJURL,
			"vmodel":     Config.VectorModel,
			"workers":    Config.WorkerCount,
			"knownfnts":  strings.Join(kff, "C0, C3"),
			"deffnt":     Config.Font}

		t := template.Must(template.New("").Parse(vv.HELPTEXTTEMPLATE))

		var b bytes.Buffer
		if ee := t.Execute(&b, m); ee != nil {
			msg.CRIT(FAIL7)
		}
		fmt.Println(msg.Styled(msg.Color(b.String())))

		os.Exit(0)
	}

	for i, a := range args {
		switch a {
		case "-vv":
			PrintVersion(*Config)
			printbuildinfo(*Config)
			os.Exit(1)
		case "-v":
			fmt.Println(vv.VERSION + VersSuppl)
			os.Exit(1)
		case "-au":
			Config.Authenticate = true
		case "-av":
			Config.VectorBot = true
		case "-bc":
			bc, err := strconv.Atoi(args[i+1])
			msg.EC(err)
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
			msg.EC(err)
			Config.EchoLog = ll
		case "-ft":
			Config.Font = args[i+1]
		case "-gl":
			ll, err := strconv.Atoi(args[i+1])
			msg.EC(err)
			Config.LogLevel = ll
		case "-gz":
			Config.Gzip = true
		case "-h":
			help()
		case "-md":
			Config.VectorModel = args[i+1]
		case "-mi":
			mi, err := strconv.Atoi(args[i+1])
			msg.EC(err)
			Config.MaxSrchIP = mi
		case "-ms":
			ms, err := strconv.Atoi(args[i+1])
			msg.EC(err)
			Config.MaxSrchTot = ms
		case "-pc":
			Config.ProfileCPU = true
		case "-pd":
			CopyInstructions()
		case "-pg":
			js := args[i+1]
			var pl structs.PostgresLogin
			err := json.Unmarshal([]byte(js), &pl)
			if err != nil {
				msg.MAND(FAIL1)
				msg.CRIT(FAIL2)
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
			msg.EC(err)
			Config.HostPort = p
		case "-st":
			Config.SelfTest += 1
		case "-tk":
			Config.TickerActive = true
		case "-ui":
			Config.BadChars = args[i+1]
		case "-wc":
			wc, err := strconv.Atoi(args[i+1])
			msg.EC(err)
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
	msg.TMI(fmt.Sprintf("'%s%s'%s loaded", h, vv.CONFIGPROLIX, y))

	SetConfigPass(&confc, cf)

	if Config.VectorMaxlines == 0 {
		Config.VectorMaxlines = vv.VECTORMAXLINES
	}

	if Config.WorkerCount > runtime.NumCPU() {
		msg.CRIT(fmt.Sprintf(FAIL5, Config.WorkerCount, runtime.NumCPU(), runtime.NumCPU()))
		Config.WorkerCount = runtime.NumCPU()
	}
}

// BuildDefaultConfig - return a CurrentConfiguration filled out with various default values
func BuildDefaultConfig() *structs.CurrentConfiguration {
	var c structs.CurrentConfiguration
	c.Authenticate = false
	c.BadChars = vv.UNACCEPTABLEINPUT
	c.BlackAndWhite = vv.BLACKANDWHITE
	c.BrowserCtx = vv.DEFAULTBROWSERCTX
	c.CustomCSS = false
	c.DbDebug = false
	c.Font = vv.FONTSETTING
	c.Gzip = vv.USEGZIP
	c.HostIP = vv.SERVEDFROMHOST
	c.HostPort = vv.SERVEDFROMPORT
	c.LdaTopics = vv.LDATOPICS
	c.LdaGraph = false
	c.LogLevel = vv.DEFAULTGOLOGLEVEL
	c.EchoLog = vv.DEFAULTECHOLOGLEVEL
	c.ManualGC = false
	c.MaxText = vv.MAXTEXTLINEGENERATION
	c.MaxSrchIP = vv.MAXSEARCHPERIPADDR
	c.MaxSrchTot = vv.MAXSEARCHTOTAL
	c.ProfileCPU = false
	c.ProfileMEM = false
	c.QuietStart = false
	c.ResetVectors = false
	c.SelfTest = 0
	c.TickerActive = vv.TICKERISACTIVE
	c.VectorBot = false
	c.VectorChtHt = vv.DEFAULTCHRTHEIGHT
	c.VectorChtWd = vv.DEFAULTCHRTWIDTH
	c.VectorMaxlines = vv.VECTORMAXLINES
	c.VectorModel = vv.VECTORMODELDEFAULT
	c.VectorNeighb = vv.VECTORNEIGHBORS
	c.VectorsDisabled = false
	c.VectorTextPrep = vv.VECTORTEXTPREPDEFAULT
	c.VectorWebExt = vv.VECTROWEBEXTDEFAULT
	c.VocabByCt = vv.VOCABBYCOUNT
	c.VocabScans = vv.VOCABSCANSION
	c.WorkerCount = runtime.NumCPU()
	c.ZapLunates = false
	e := json.Unmarshal([]byte(vv.DEFAULTCORPORA), &c.DefCorp)
	if e != nil {
		fmt.Println("BuildDefaultConfig() could not json.Unmarshal DEFAULTCORPORA: " + vv.DEFAULTCORPORA)
	}

	pl := structs.PostgresLogin{
		Host:   vv.DEFAULTPSQLHOST,
		Port:   vv.DEFAULTPSQLPORT,
		User:   vv.DEFAULTPSQLUSER,
		Pass:   "",
		DBName: vv.DEFAULTPSQLDB,
	}

	c.PGLogin = pl

	return &c
}

// SetConfigPass - make sure that Config.PGLogin.Pass != ""
func SetConfigPass(cfg *structs.CurrentConfiguration, cf string) {
	const (
		FAIL3     = "FAILED to load database credentials from any of '%s', '%s' or '%s'"
		FAIL4     = "At a minimum be sure that a 'hgs-vv.json' file exists and that it has the following format:"
		FAIL6     = "Could not open '%s'"
		BLANKPASS = "PostgreSQLPassword is blank. Check your 'hgs-vv.json' file. NB: 'PostgreSQLPassword ≠ 'PosgreSQLPassword'.\n"
	)
	type ConfigFile struct {
		PostgreSQLPassword string
	}

	uh, _ := os.UserHomeDir()
	h := fmt.Sprintf(vv.CONFIGALTAPTH, uh)

	if cf == "" {
		cf = fmt.Sprintf("%s/%s", vv.CONFIGLOCATION, vv.CONFIGBASIC)
	}

	acf := fmt.Sprintf("%s/%s", h, vv.CONFIGBASIC)

	if Config.PGLogin.Pass == "" {
		Config.PGLogin = structs.PostgresLogin{}
		cfa, ee := os.Open(cf)
		if ee != nil {
			msg.TMI(fmt.Sprintf(FAIL6, cf))
		}
		cfb, ee := os.Open(acf)
		if ee != nil {
			msg.TMI(fmt.Sprintf(FAIL6, acf))
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
			msg.CRIT(fmt.Sprintf(FAIL3, cf, acf, fmt.Sprintf("%s/%s", h, vv.CONFIGPROLIX)))
			msg.CRIT(fmt.Sprintf(FAIL4))
			fmt.Printf(vv.MINCONFIG)
			msg.ExitOrHang(0)
		}

		thecfg := ConfigFile{}
		if erra == nil {
			thecfg = confa
		} else {
			thecfg = confb
		}

		if thecfg.PostgreSQLPassword == "" {
			msg.MAND(BLANKPASS)
		}

		Config.PGLogin = structs.PostgresLogin{
			Host:   vv.DEFAULTPSQLHOST,
			Port:   vv.DEFAULTPSQLPORT,
			User:   vv.DEFAULTPSQLUSER,
			DBName: vv.DEFAULTPSQLDB,
			Pass:   thecfg.PostgreSQLPassword,
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
			msg.WARN(FNF)
			return
		}

		msg.CRIT(FYI)

		err = os.WriteFile(f, data, vv.WRITEPERMS)
		if err != nil {
			msg.WARN(FNF)
			return
		}
		msg.CRIT(fmt.Sprintf("\t\tWrote:\t'%s'", f))
	}

	for _, info := range []string{CUST, FYIF, SEMV, BASF} {
		data, err := efs.ReadFile(EPD + info)
		if err != nil {
			return
		}
		err = os.WriteFile(info, data, vv.WRITEPERMS)
		if err != nil {
			msg.WARN(FNF)
			return
		}
		msg.CRIT(fmt.Sprintf("\t\tWrote:\t'%s'", info))
	}
}
