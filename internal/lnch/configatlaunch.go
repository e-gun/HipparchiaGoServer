//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//

package lnch

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/base/mm"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/clr"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"os"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

//go:embed efs
var efs embed.FS

var (
	Config *str.CurrentConfiguration
	Msg    = mm.NewMessageMaker()
)

const (
	EPD = "emb/pdf/"
)

// LookForConfigFile - test to see if we can find a config file; if not build one and check to see if the DB needs loading
func LookForConfigFile() {
	var c error

	h, e := os.UserHomeDir()
	if e != nil {
		// how likely is this...?
		// b = errors.New("cannot find UserHomeDir")
	}
	cfp := fmt.Sprintf(vv.CONFIGALTAPTH, h) + "/" + vv.CONFIGOPTIONS
	_, c = os.Stat(cfp)

	notfound := c != nil

	if notfound {
		PGFSConfig(cfp, h)
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
		FAIL9 = "Could not find css color scheme '%s'. Falling back to '%s'"
	)

	Config = builddefaultconfig()

	uh, _ := os.UserHomeDir()
	configpath := fmt.Sprintf(vv.CONFIGALTAPTH, uh)
	options := fmt.Sprintf("%s/%s", configpath, vv.CONFIGOPTIONS)

	loadedcfg, e := os.Open(options)
	if e != nil {
		Msg.CRIT(fmt.Sprintf(FAIL6, options))
	}

	decoderc := json.NewDecoder(loadedcfg)
	confc := str.CurrentConfiguration{}
	errc := decoderc.Decode(&confc)
	_ = loadedcfg.Close()

	if errc == nil {
		Config = &confc
	} else {
		Msg.CRIT(fmt.Sprintf(FAIL3, options))
	}

	loadcolors(configpath)

	Msg.LLvl = Config.LogLevel

	// on old CONFIGOPTIONS might mean you set the following to zero; that is very bad...
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
		PrintBuildInfo(*Config)
		cwd, err := os.Getwd()
		if err != nil {
			Msg.CRIT(FAIL8)
			cwd = "(unknown)"
		}

		kff := gen.StringMapKeysIntoSlice(vv.ServableFonts)
		sort.Strings(kff)

		kcc := gen.StringMapKeysIntoSlice(clr.CssColorHSLs)
		sort.Strings(kcc)

		// see HELPTEXTTEMPLATE in vv.terminalconst.go
		m := map[string]interface{}{
			"authstatus":  Config.Authenticate,
			"badchars":    Config.BadChars,
			"confauth":    vv.CONFIGAUTH,
			"conffile":    vv.CONFIGOPTIONS,
			"cpus":        runtime.NumCPU(),
			"css":         vv.CUSTOMCSSFILENAME,
			"ctxlines":    Config.BrowserCtx,
			"cwd":         cwd,
			"dbf":         vv.HDBFOLDER,
			"defcol":      Config.CssColors,
			"deffnt":      Config.Font,
			"echoll":      Config.EchoLog,
			"hdbf":        vv.HDBFOLDER,
			"hgsll":       Config.LogLevel,
			"home":        configpath,
			"host":        Config.HostIP,
			"knowncolors": strings.Join(kcc, "C0, C3"),
			"knownfnts":   strings.Join(kff, "C0, C3"),
			"lfstatus":    Config.LogToFile,
			"loge":        vv.LOGFILEEL,
			"logm":        vv.LOGFILEML,
			"maxipsrch":   Config.MaxSrchIP,
			"maxtotscrh":  Config.MaxSrchTot,
			"port":        Config.HostPort,
			"projurl":     vv.PROJURL,
			"sslcert":     vv.SSLCPEM,
			"ssldir":      Config.SSLCertDir,
			"sslport":     Config.HostSSLPort,
			"sslpriv":     vv.SSLPPEM,
			"tlines":      Config.TickerLines,
			"uv":          Config.UVSub,
			"vmodel":      Config.VectorModel,
			"workers":     Config.WorkerCount,
		}

		t := template.Must(template.New("").Parse(vv.HELPTEXTTEMPLATE))

		var b bytes.Buffer
		if ee := t.Execute(&b, m); ee != nil {
			Msg.CRIT(FAIL7)
		}
		fmt.Println(Msg.Styled(Msg.Color(b.String())))

		os.Exit(0)
	}

	for i, a := range args {
		switch a {
		case "-vv":
			PrintVersion(*Config)
			PrintBuildInfo(*Config)
			os.Exit(1)
		case "-v":
			fmt.Println(vv.VERSION + VersSuppl)
			os.Exit(1)
		case "-au":
			// toggle...
			Config.Authenticate = !Config.Authenticate
		case "-av":
			Config.VectorBot = true
		case "-bc":
			bc, err := strconv.Atoi(args[i+1])
			Msg.EC(err)
			Config.BrowserCtx = bc
		case "-bw":
			Config.BlackAndWhite = true
		case "-cd":
			// redefine a color scheme
			// "-cd 220 15 95 0 Monochrome"
			modifycolorscheme(args[i+1], args[i+2], args[i+3], args[i+4], args[i+5])
		case "-cm":
			Config.CssColors = args[i+1]
		case "-cr":
			reportcolorschemes()
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
			Msg.EC(err)
			Config.EchoLog = ll
		case "-ft":
			Config.Font = args[i+1]
		case "-gl":
			ll, err := strconv.Atoi(args[i+1])
			Msg.EC(err)
			Config.LogLevel = ll
		case "-gz":
			Config.Gzip = true
		case "-h":
			help()
		case "-lf":
			// toggle...
			Config.LogToFile = !Config.LogToFile
		case "-md":
			Config.VectorModel = args[i+1]
		case "-mi":
			mi, err := strconv.Atoi(args[i+1])
			Msg.EC(err)
			Config.MaxSrchIP = mi
		case "-ms":
			ms, err := strconv.Atoi(args[i+1])
			Msg.EC(err)
			Config.MaxSrchTot = ms
		case "-pc":
			Config.ProfileCPU = true
		case "-pd":
			copyinstructions()
		case "-pg":
			js := args[i+1]
			var pl str.PostgresLogin
			err := json.Unmarshal([]byte(js), &pl)
			if err != nil {
				Msg.MAND(FAIL1)
				Msg.CRIT(FAIL2)
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
		case "-sd":
			Config.SSLCertDir = args[i+1]
		case "-sp":
			p, err := strconv.Atoi(args[i+1])
			Msg.EC(err)
			Config.HostPort = p
		case "-ss":
			p, err := strconv.Atoi(args[i+1])
			Msg.EC(err)
			Config.HostSSLPort = p
		case "-st":
			Config.SelfTest += 1
		case "-tk":
			Config.TickerActive = true
			Config.LogToFile = true
		case "-tl":
			ll, err := strconv.Atoi(args[i+1])
			if err != nil {
				fmt.Println("number not found after '-tl'")
				os.Exit(1)
			}
			Config.TickerLines = ll
		case "-ui":
			Config.BadChars = args[i+1]
		case "-uv":
			// toggle...
			Config.UVSub = !Config.UVSub
		case "-wc":
			wc, err := strconv.Atoi(args[i+1])
			Msg.EC(err)
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

	// the package needs an injection...
	str.UVSubs = Config.UVSub

	y := ""
	if errc != nil {
		y = " *not*"
	}
	Msg.TMI(fmt.Sprintf("'%s%s'%s loaded", configpath, vv.CONFIGOPTIONS, y))

	setconfigpass(&confc, cf)

	if Config.VectorMaxlines == 0 {
		Config.VectorMaxlines = vv.VECTORMAXLINES
	}

	if Config.WorkerCount > runtime.NumCPU() {
		Msg.CRIT(fmt.Sprintf(FAIL5, Config.WorkerCount, runtime.NumCPU(), runtime.NumCPU()))
		Config.WorkerCount = runtime.NumCPU()
	}

	if slices.Contains(gen.StringMapKeysIntoSlice(vv.ServableFonts), Config.Font) {
		f := vv.ServableFonts[Config.Font]
		if !f.HasLunateSigma {
			Config.ZapLunates = true
		}
	}

	if !slices.Contains(gen.StringMapKeysIntoSlice(clr.CssColorHSLs), Config.CssColors) {
		Msg.WARN(fmt.Sprintf(FAIL9, Config.CssColors, vv.DEFAULTCOLORSCHEME))
		Config.CssColors = vv.DEFAULTCOLORSCHEME
	}
}

// builddefaultconfig - return a CurrentConfiguration filled out with various default values
func builddefaultconfig() *str.CurrentConfiguration {
	var c str.CurrentConfiguration
	c.Authenticate = false
	c.BadChars = vv.UNACCEPTABLEINPUT
	c.BlackAndWhite = vv.BLACKANDWHITE
	c.BrowserCtx = vv.DEFAULTBROWSERCTX
	c.CustomCSS = false
	c.CssColors = vv.DEFAULTCOLORSCHEME
	c.DbDebug = false
	c.Font = vv.FONTSETTING
	c.Gzip = vv.USEGZIP
	c.HostIP = vv.SERVEDFROMHOST
	c.HostPort = vv.SERVEDFROMPORT
	c.HostSSLPort = vv.SERVEDFROMSSLPORT
	c.LdaTopics = vv.LDATOPICS
	c.LdaGraph = false
	c.LogLevel = vv.DEFAULTGOLOGLEVEL
	c.LogToFile = false
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
	c.SSLCertDir = vv.SSLCERTDIR
	c.TickerActive = vv.TICKERISACTIVE
	c.TickerLines = vv.TICKERLINES
	c.UVSub = vv.UVSUBSINDISPLAY
	c.VectorBot = false
	c.VectorChtHt = vv.DEFAULTCHRTHEIGHT
	c.VectorChtWd = vv.DEFAULTCHRTWIDTH
	c.VectorMaxlines = vv.VECTORMAXLINES
	c.VectorModel = vv.VECTORMODELDEFAULT
	c.VectorNeighb = vv.VECTORNEIGHBORS
	c.VectorsDisabled = false
	c.VectorTextPrep = vv.VECTORTEXTPREPDEFAULT
	c.VectorWebExt = vv.VECTORWEBEXTDEFAULT
	c.VocabByCt = vv.VOCABBYCOUNT
	c.VocabScans = vv.VOCABSCANSION
	c.WorkerCount = runtime.NumCPU()
	c.ZapLunates = false
	e := json.Unmarshal([]byte(vv.DefaultCorpora), &c.DefCorp)
	if e != nil {
		fmt.Println("builddefaultconfig() could not json.Unmarshal DefaultCorpora: " + vv.DefaultCorpora)
	}

	pl := str.PostgresLogin{
		Host:   vv.DEFAULTPSQLHOST,
		Port:   vv.DEFAULTPSQLPORT,
		User:   vv.DEFAULTPSQLUSER,
		Pass:   "",
		DBName: vv.DEFAULTPSQLDB,
	}

	c.PGLogin = pl

	return &c
}

// setconfigpass - make sure that Config.PGLogin.Pass != ""
func setconfigpass(cfg *str.CurrentConfiguration, cf string) {
	const (
		FAIL3     = "FAILED to load database credentials from '%s'"
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
		cf = fmt.Sprintf("%s/%s", vv.CONFIGLOCATION, vv.CONFIGOPTIONS)
	}

	if Config.PGLogin.Pass == "" {
		Config.PGLogin = str.PostgresLogin{}
		cfa, ee := os.Open(cf)
		if ee != nil {
			Msg.TMI(fmt.Sprintf(FAIL6, cf))
		}

		defer func(cfa *os.File) {
			err := cfa.Close()
			if err != nil {
			} // the file was almost certainly not found in the first place...
		}(cfa)

		decodera := json.NewDecoder(cfa)
		confa := ConfigFile{}
		erra := decodera.Decode(&confa)

		if erra != nil && cfg.PGLogin.DBName == "" {
			Msg.CRIT(fmt.Sprintf(FAIL3, fmt.Sprintf("%s/%s", h, vv.CONFIGOPTIONS)))
			Msg.CRIT(fmt.Sprintf(FAIL4))
			fmt.Printf(vv.MINCONFIG)
			Msg.ExitOrHang(0)
		}

		thecfg := ConfigFile{}
		if erra == nil {
			thecfg = confa
		}

		if thecfg.PostgreSQLPassword == "" {
			Msg.MAND(BLANKPASS)
		}

		Config.PGLogin = str.PostgresLogin{
			Host:   vv.DEFAULTPSQLHOST,
			Port:   vv.DEFAULTPSQLPORT,
			User:   vv.DEFAULTPSQLUSER,
			DBName: vv.DEFAULTPSQLDB,
			Pass:   thecfg.PostgreSQLPassword,
		}
	}
}

// copyinstructions - write the embedded PDF to the filesystem
func copyinstructions() {
	const (
		FYI  = "Writing instruction files to the current working directory."
		MACI = "HGS_INSTALLATION_MacOS.pdf"
		WINI = "HGS_INSTALLATION_Windows.pdf"
		NIXI = "HGS_INSTALLATION_Nix.pdf"
		CUST = "HGS_CUSTOMIZATION.pdf"
		SEMV = "HGS_SEMANTICVECTORS.pdf"
		FYIF = "HGS_FYI.pdf"
		BASF = "HGS_BASIC_USE.pdf"
		FNF  = "copyinstructions(): Embedded PDF not found. This function will now return."
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
			Msg.WARN(FNF)
			return
		}

		Msg.CRIT(FYI)

		err = os.WriteFile(f, data, vv.WRITEPERMS)
		if err != nil {
			Msg.WARN(FNF)
			return
		}
		Msg.CRIT(fmt.Sprintf("\t\tWrote:\t'%s'", f))
	}

	for _, info := range []string{CUST, FYIF, SEMV, BASF} {
		data, err := efs.ReadFile(EPD + info)
		if err != nil {
			return
		}
		err = os.WriteFile(info, data, vv.WRITEPERMS)
		if err != nil {
			Msg.WARN(FNF)
			return
		}
		Msg.CRIT(fmt.Sprintf("\t\tWrote:\t'%s'", info))
	}
}

func modifycolorscheme(hs string, ss string, ls string, lhls string, cs string) {
	const (
		ERR1 = `'%s' is not a valid color scheme.`
		ERR2 = `Known schemes are: %s.`
		ERR3 = `The scheme name needs to be followed by four integers: hue, sat, lum, high/low lum.`
		ERR4 = `Neither 'Light' nor 'Dark' can be modified.`
	)

	if hs == "Light" || hs == "Dark" {
		Msg.MAND(ERR4)
		return
	}

	if _, ok := clr.CssColorHSLs[cs]; !ok {
		Msg.MAND(fmt.Sprintf(ERR1, cs))
		kcc := gen.StringMapKeysIntoSlice(clr.CssColorModes)
		sort.Strings(kcc)
		Msg.MAND(fmt.Sprintf(ERR2, strings.Join(kcc, ",")))
		return
	}
	h, err1 := strconv.Atoi(hs)
	s, err2 := strconv.Atoi(ss)
	l, err3 := strconv.Atoi(ls)
	lh, err4 := strconv.Atoi(lhls)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		Msg.MAND(ERR3)
		fmt.Println(err1)
		fmt.Println(err2)
		fmt.Println(err3)
		fmt.Println(err4)
		return
	}
	clr.CssColorHSLs[cs] = []int{h, s, l, lh}
}

func reportcolorschemes() {
	const (
		HEAD = "\thue\tsat\tlum1\tlum2"
		TMPL = "\t%d\t%d\t%d\t%d\t%s"
	)
	fmt.Println("Color scheme values (modify via '-cd' flag)")
	fmt.Println(HEAD)

	kcc := gen.StringMapKeysIntoSlice(clr.CssColorHSLs)
	sort.Strings(kcc)

	for _, n := range kcc {
		cv := clr.CssColorHSLs[n]
		fmt.Println(fmt.Sprintf(TMPL, cv[0], cv[1], cv[2], cv[3], n))
	}
}

func loadcolors(configpath string) {
	clrs := map[string][]int{}
	clroptfile := fmt.Sprintf("%s/%s", configpath, vv.CONFIGCOLORS)
	loadedcfg, e := os.Open(clroptfile)
	if e != nil {
		writedefaultcolors(clroptfile)
	} else {
		decoder := json.NewDecoder(loadedcfg)
		errc := decoder.Decode(&clrs)
		if errc != nil {
			fmt.Println("could not parse color data: overwriting " + vv.CONFIGCOLORS)
			writedefaultcolors(clroptfile)
		} else {
			clr.CssColorHSLs = clrs
		}
	}
}

func writedefaultcolors(clroptfilepath string) {
	const (
		COLORNOTE = "\n// format: hue, saturation, minormaxlum1, minormaxlum2; 'Dark' and 'Light' cannot be modified."
	)
	// built-in clr.CssColorHSLs will be used
	jsonclr, err := json.MarshalIndent(clr.CssColorHSLs, "", "\t")
	if err != nil {
		fmt.Println("Error marshaling data:", err)
	}
	note := []byte(COLORNOTE)
	err = os.WriteFile(clroptfilepath, append(jsonclr, note...), 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
	}
}
