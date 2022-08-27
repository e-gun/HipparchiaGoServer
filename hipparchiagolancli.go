//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"C"
	"flag"
	"fmt"
	"os"
	"time"
)

const (
	myname          = "Hipparchia Golang Server"
	shortname       = "HGS"
	version         = "0.0.9"
	tesquery        = "SELECT * FROM %s WHERE index BETWEEN %d and %d"
	testdb          = "lt0448"
	teststart       = 1
	testend         = 26
	linelength      = 72
	pollinginterval = 333 * time.Millisecond
	skipheadwords   = "unus verum omne sum¹ ab δύο πρότεροϲ ἄνθρωποϲ τίϲ δέω¹ ὅϲτιϲ homo πᾶϲ οὖν εἶπον ἠμί ἄν² tantus μένω μέγαϲ οὐ verus neque eo¹ nam μέν ἡμόϲ aut Sue διό reor ut ἐγώ is πωϲ ἐκάϲ enim ὅτι² παρά ἐν Ἔχιϲ sed ἐμόϲ οὐδόϲ ad de ita πηρόϲ οὗτοϲ an ἐπεί a γάρ αὐτοῦ ἐκεῖνοϲ ἀνά ἑαυτοῦ quam αὐτόϲε et ὑπό quidem Alius¹ οἷοϲ noster γίγνομαι ἄνα προϲάμβ ἄν¹ οὕτωϲ pro² tamen ἐάν atque τε qui² si multus idem οὐδέ ἐκ omnes γε δεῖ πολύϲ in ἔδω ὅτι¹ μή Ios ἕτεροϲ cum meus ὅλοξ suus omnis ὡϲ sua μετά Ἀλλά ne¹ jam εἰϲ ἤ² ἄναξ ἕ ὅϲοϲ dies ipse ὁ hic οὐδείϲ suo ἔτι ἄνω¹ ὅϲ νῦν ὁμοῖοϲ edo¹ εἰ qui¹ πάλιν ὥϲπερ ne³ ἵνα τιϲ διά φύω per τοιοῦτοϲ for eo² huc locum neo¹ sui non ἤ¹ χάω ex κατά δή ἁμόϲ ὅμοιοϲ αὐτόϲ etiam vaco πρόϲ Ζεύϲ ϲύ quis¹ tuus b εἷϲ Eos οὔτε τῇ καθά ego tu ille pro¹ ἀπό suum εἰμί ἄλλοϲ δέ alius² pars vel ὥϲτε χέω res ἡμέρα quo δέομαι modus ὑπέρ ϲόϲ ito τῷ περί Τήιοϲ ἕκαϲτοϲ autem καί ἐπί nos θεάω γάρον γάροϲ Cos²"
	skipinflected   = "ἀρ ita a inquit ego die nunc nos quid πάντων ἤ με θεόν δεῖ for igitur ϲύν b uers p ϲου τῷ εἰϲ ergo ἐπ ὥϲτε sua me πρό sic aut nisi rem πάλιν ἡμῶν φηϲί παρά ἔϲτι αὐτῆϲ τότε eos αὐτούϲ λέγει cum τόν quidem ἐϲτιν posse αὐτόϲ post αὐτῶν libro m hanc οὐδέ fr πρῶτον μέν res ἐϲτι αὐτῷ οὐχ non ἐϲτί modo αὐτοῦ sine ad uero fuit τοῦ ἀπό ea ὅτι parte ἔχει οὔτε ὅταν αὐτήν esse sub τοῦτο i omnes break μή ἤδη ϲοι sibi at mihi τήν in de τούτου ab omnia ὃ ἦν γάρ οὐδέν quam per α autem eius item ὡϲ sint length οὗ eum ἀντί ex uel ἐπειδή re ei quo ἐξ δραχμαί αὐτό ἄρα ἔτουϲ ἀλλ οὐκ τά ὑπέρ τάϲ μάλιϲτα etiam haec nihil οὕτω siue nobis si itaque uac erat uestig εἶπεν ἔϲτιν tantum tam nec unde qua hoc quis iii ὥϲπερ semper εἶναι e ½ is quem τῆϲ ἐγώ καθ his θεοῦ tibi ubi pro ἄν πολλά τῇ πρόϲ l ἔϲται οὕτωϲ τό ἐφ ἡμῖν οἷϲ inter idem illa n se εἰ μόνον ac ἵνα ipse erit μετά μοι δι γε enim ille an sunt esset γίνεται omnibus ne ἐπί τούτοιϲ ὁμοίωϲ παρ causa neque cr ἐάν quos ταῦτα h ante ἐϲτίν ἣν αὐτόν eo ὧν ἐπεί οἷον sed ἀλλά ii ἡ t te ταῖϲ est sit cuius καί quasi ἀεί o τούτων ἐϲ quae τούϲ minus quia tamen iam d διά primum r τιϲ νῦν illud u apud c ἐκ δ quod f quoque tr τί ipsa rei hic οἱ illi et πῶϲ φηϲίν τοίνυν s magis unknown οὖν dum text μᾶλλον habet τοῖϲ qui αὐτοῖϲ suo πάντα uacat τίϲ pace ἔχειν οὐ κατά contra δύο ἔτι αἱ uet οὗτοϲ deinde id ut ὑπό τι lin ἄλλων τε tu ὁ cf δή potest ἐν eam tum μου nam θεόϲ κατ ὦ cui nomine περί atque δέ quibus ἡμᾶϲ τῶν eorum"
	memoutputfile   = "mem_profiler_output.bin"
	cpuoutputfile   = "cpu_profiler_output.bin"
	browseauthor    = "gr0062"
	browsework      = "028"
	browseline      = 14672
	browsecontext   = 4
	lexword         = "καρποῦ"
	lexauthor       = "gr0062"
	RP              = `{"Addr": "localhost:6379", "Password": "", "DB": 0}`
	PSQ             = `{"Host": "localhost", "Port": 5432, "User": "hippa_wr", "Pass": "", "DBName": "hipparchiaDB"}`
	PSDefaultHost   = "localhost"
	PSDefaultUser   = "hippa_wr"
	PSDefaultPort   = 5432
	PSDefaultDB     = "hipparchiaDB"
	TwoPassThresh   = 100 // cicero has >70 works
	NumWorkers      = 10
)

var (
	cfg      CurrentConfiguration
	sessions = make(map[string]Session)
	searches = make(map[string]SearchStruct)
)

type CurrentConfiguration struct {
	RedisKey        string
	MaxHits         int64
	WorkerCount     int
	LogLevel        int
	RedisInfo       string
	PosgresInfo     string
	BagMethod       string
	SentPerBag      int
	VectTestDB      string
	VectStart       int
	VectEnd         int
	VSkipHW         string
	VSkipInf        string
	TestV1          string
	TestV2          string
	TestV3          string
	PSQP            string
	BrowseFoundline int64
	BrowseContext   int64
	IsVectPtr       *bool
	IsWSPtr         *bool
	IsTestPtr       *bool
	WSPort          int
	WSFail          int
	WSSave          int
	ProfCPUPtr      *bool
	ProfMemPtr      *bool
	SendVersPtr     *bool
	RLogin          RedisLogin
	PGLogin         PostgresLogin
}

func main() {

	versioninfo := fmt.Sprintf("%s CLI Debugging Interface (v.%s)", myname, version)

	// grabpgsqlconnection() needs cfg; it is run before main(); so it calls configatstartup()
	// and re-config will produce "flag redefined" errors

	// configatstartup()

	if *cfg.SendVersPtr {
		fmt.Println(versioninfo)
		os.Exit(1)
	}

	versioninfo = versioninfo + fmt.Sprintf(" [loglevel=%d]", cfg.LogLevel)

	if cfg.LogLevel > 5 {
		cfg.LogLevel = 5
	}

	if cfg.LogLevel < 0 {
		cfg.LogLevel = 0
	}

	cfg.PGLogin = decodepsqllogin([]byte(cfg.PosgresInfo))

	fmt.Println(versioninfo)
	msg("cfg.IsTestPtr was set: Testing Run", 1)
	// test_selection()
	// test_compilesearchlist()
	// test_searchlistintoqueries()
	StartEchoServer()
	return

}

//
// CLI STARTUP CONFIGURATION
//

func configatstartup() {
	// WARNING: a password might get hard-coded into the binary. It is easy to use the binary in HipparchiaServer
	// without providing valid credentials to the binary, but if you do you must pass them and your credentials will be
	// visible to a "ps aux | grep ..."; but a hard-coded binary is not so good if you are going to share it...

	// testing flags
	cfg.IsTestPtr = flag.Bool("tt", false, "[testing] assert that this is a testing run")
	flag.StringVar(&cfg.TestV1, "t1", "", "[testing] parameter 1")
	flag.StringVar(&cfg.TestV2, "t2", "", "[testing] parameter 2")
	flag.StringVar(&cfg.TestV3, "t3", "", "[testing] parameter 3")
	flag.StringVar(&cfg.PSQP, "psqp", "", "[testing] PSQL Password")

	cfg.SendVersPtr = flag.Bool("v", false, "[common] print version and exit")

	flag.Int64Var(&cfg.MaxHits, "c", 200, "[searches] max hit count")
	flag.IntVar(&cfg.LogLevel, "l", 3, "[common] logging level: 0 is silent; 5 is very noisy")
	flag.StringVar(&cfg.PosgresInfo, "p", PSQ, "[common] psql logon information (as a JSON string)")

}
