//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vv

import "time"

const (
	GREEKCORP      = "gr"
	LATINCORP      = "lt"
	PAPYRUSCORP    = "dp"
	INSCRIPTCORP   = "in"
	CHRISTINSC     = "ch"
	DEFAULTCORPORA = "{\"gr\": true, \"lt\": true, \"in\": false, \"ch\": false, \"dp\": false}"

	AVGWORDSPERLINE   = 8 // hard coding a suspect assumption
	BLACKANDWHITE     = false
	CHARSPERLINE      = 60 // used by vector to preallocate memory: set it closer to a max than a real average
	CONFIGLOCATION    = "."
	CONFIGALTAPTH     = "%s/.config/" // %s = os.UserHomeDir()
	CONFIGAUTH        = "hgs-users.json"
	CONFIGBASIC       = "hgs-conf.json"
	CONFIGPROLIX      = "hgs-prolix-conf.json"
	CUSTOMCSSFILENAME = "custom-hipparchiastyles.css"
	// DBAUMAPSIZE              = 3455   //[HGS] [A2: 0.436s][Δ: 0.051s] 3455 authors built: map[string]DbAuthor
	DBLMMAPSIZE = 151701 //[HGS] [B1: 0.310s][Δ: 0.310s] unnested lemma map built (151701 items)
	// DBWKMAPSIZE              = 236835 //[HGS] [A1: 0.385s][Δ: 0.385s] 236835 works built: map[string]DbWork
	DEFAULTBROWSERCTX        = 14
	DEFAULTCOLUMN            = "stripped_line"
	DEFAULTECHOLOGLEVEL      = 0
	DEFAULTGOLOGLEVEL        = 0
	DEFAULTHITLIMIT          = 250
	DEFAULTLINESOFCONTEXT    = 4
	DEFAULTPROXIMITY         = 1
	DEFAULTPROXIMITYSCOPE    = "lines"
	DEFAULTPSQLHOST          = "127.0.0.1"
	DEFAULTPSQLUSER          = "hippa_wr"
	DEFAULTPSQLPORT          = 5432
	DEFAULTPSQLDB            = "hipparchiaDB"
	DEFAULTQUERYSYNTAX       = "~"
	FIRSTSEARCHLIM           = 750000 // 149570 lines in Cicero (lt0474); all 485 forms of »δείκνυμι« will pass 50k
	FONTSETTING              = "Noto"
	GENRESTOCOUNT            = 5
	HDBFOLDER                = "hDB"
	INCERTADATE              = 2500
	JSONINDENT               = "  "
	LENGTHOFAUTHORID         = 6
	LENGTHOFWORKID           = 3
	MAXBROWSERCONTEXT        = 60
	MAXDATE                  = 1500
	MAXDATESTR               = "1500"
	MAXDICTLOOKUP            = 125
	MAXDISTANCE              = 10
	MAXECHOREQPERSECONDPERIP = 60 // it takes c. 20 to load the front page for the first time; 40 lets you double-load; selftestsuite needs 60
	MAXHITLIMIT              = 2500
	MAXINPUTLEN              = 64
	MAXLEMMACHUNKSIZE        = 25
	MAXLINESHITCONTEXT       = 30
	MAXSEARCHINFOLISTLEN     = 100
	MAXSEARCHPERIPADDR       = 2
	MAXSEARCHTOTAL           = 4     // note that vectors and two-part searches generate subsearches and kick your total active search count over the number of "clicked" searches from RtSearch()
	MAXTEXTLINEGENERATION    = 40000 // euripides is 33517 lines, sophocles is 15729, cicero is 149570, e.g.; jQuery slows exponentially as lines increase
	MAXVOCABLINEGENERATION   = 1     // this is a multiplier for Config.MaxText; the browser does not get overwhelmed by these lists
	MAXTITLELENGTH           = 110
	MINBROWSERWIDTH          = 90
	MINDATE                  = -850
	MINDATESTR               = "-850"
	MINORGENREWTCAP          = 250
	NESTEDLEMMASIZE          = 543
	NUMBEROFCITATIONLEVELS   = 6
	ORDERBY                  = "index"
	POLLEVERYNTABLES         = 34 // 3455 is the max number of tables in a search...
	SERVEDFROMHOST           = "127.0.0.1"
	SERVEDFROMPORT           = 8000
	SIMULTANEOUSSEARCHES     = 3 // cap on the number of db connections at (S * Config.WorkerCount)
	SHOWCITATIONEVERYNLINES  = 10
	SORTBY                   = "shortname"
	TEMPTABLETHRESHOLD       = 100 // if a table requires N "between" clauses, build a temptable instead to gather the needed lines
	TERMINATIONS             = `(\s|\.|\]|\<|⟩|’|”|\!|,|:|;|\?|·|$)`
	TICKERISACTIVE           = false
	TICKERDELAY              = 30 * time.Second
	TIMEOUTRD                = 15 * time.Second  // only set if Config.Authenticate is true (and so in a "serve the net" situation)
	TIMEOUTWR                = 120 * time.Second // this is *very* generous, but some searches are slow/long
	USEGZIP                  = false
	VARIADATE                = 2000
	VOCABSCANSION            = false
	VOCABBYCOUNT             = false
	WRITEPERMS               = 0644
	WSPOLLINGPAUSE           = 10000000 * 10 // 10000000 * 10 = every .1s

	// UNACCEPTABLEINPUT       = `|"'!@:,=+_\/`
	UNACCEPTABLEINPUT = `"'!@:,=_/` // we want to be able to do regex...; echo+net/url means some can't even make it into a parser: #%&;
	USELESSINPUT      = `’“”̣`      // these can't be found and so should be dropped; note the subscript dot at the end

)
