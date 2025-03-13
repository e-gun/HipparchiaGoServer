//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vv

const (
	FIRSTSEARCHLIM           = 750000 // 149570 lines in Cicero (lt0474); all 485 forms of »δείκνυμι« will pass 50k
	MAXBROWSERCONTEXT        = 60
	MAXDATE                  = 1500
	MAXDATESTR               = "1500"
	MAXDICTLOOKUP            = 150
	MAXDISTANCE              = 10
	MAXECHOREQPERSECONDPERIP = 60 // it takes c. 20 to load the front page for the first time; 40 lets you double-load; selftestsuite needs 60
	MAXHITLIMIT              = 4000
	MAXINPUTLEN              = 64
	MAXLEMMACHUNKSIZE        = 25
	MAXLINESHITCONTEXT       = 30
	MAXSEARCHINFOLISTLEN     = 125
	MAXSEARCHPERIPADDR       = 2
	MAXSEARCHTOTAL           = 4     // note that vectors and two-part searches generate subsearches and kick your total active search count over the number of "clicked" searches from RtSearch()
	MAXTEXTLINEGENERATION    = 40000 // euripides is 33517 lines, sophocles is 15729, cicero is 149570, e.g.; jQuery slows exponentially as lines increase
	MAXTITLELENGTH           = 110
	MAXVOCABLINEGENERATION   = 1 // this is a multiplier for Config.MaxText; the browser does not get overwhelmed by these lists
	MINDATE                  = -850
	MINDATESTR               = "-850"
	MINORGENREWTCAP          = 250
	SIMULTANEOUSSEARCHES     = 3               // cap on the number of db connections at (S * Config.WorkerCount)
	UNACCEPTABLEINPUT        = `"'!@:,=_/#%&;` // we want to be able to do regex...; echo+net/url means some can't even make it into a parser: #%&;
	USELESSINPUT             = `’“”̣`          // these can't be found and so should be dropped; note the subscript dot at the end
)
