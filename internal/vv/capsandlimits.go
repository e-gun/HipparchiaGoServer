//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
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
	SIMULTANEOUSSEARCHES     = 3               // cap on the number of db connections at (S * Config.WorkerCount)
	UNACCEPTABLEINPUT        = `"'!@:,=_/#%&;` // we want to be able to do regex...; echo+net/url means some can't even make it into a parser: #%&;
	USELESSINPUT             = `’“”̣`          // these can't be found and so should be dropped; note the subscript dot at the end
)

// hgdb=> select converted_date from authors order by converted_date asc limit 2;
// converted_date
//----------------
//           -750
//           -750

// hgdb=> select universalid,title,recorded_date,converted_date from works order by converted_date asc limit 10;
// universalid |                          title                          |         recorded_date          | converted_date
//-------------+---------------------------------------------------------+--------------------------------+----------------
// inz08sw13b  | #478: IG XII,Suppl. p. 87 (Thera)                       | sh.aft.247／6 ﹠ aft. c.235 bc |          -2461
// inz06nw46a  | #137:  (Aitolia, Phistyon)                              | 213／2 bc ﹠ 200-150 bc        |          -2132
// inz03mw1d0  | #21: ICS(2) 408,18g (Kypros, Palaipaphos)               | XIa                            |          -1050
// chz06ow3yf  | #230:  (Asturic. [Gal.], S. Martin de Salas)            | 866-910 bc                     |           -866
// inz06qw1nf  | #409 titulus a sinistra (384a) deest:  (Rhodos, Lindos) | c 9a                           |           -850
// inz0c8w5dx  | #1375:  (Att.)                                          | c.9 a                          |           -850
// inz06qw1ni  | #412:  (Rhodos, Lindos)                                 | c 9a                           |           -850
// inz0c8w5dw  | #1374:  (Att.)                                          | c.9 a                          |           -850
// inz0bww7j0  | #1312:  (Phryg., Apameia (Dinar))                       | c.9 bc                         |           -850
// inz0ciwexn  | #603:  (Att.)                                           | 9 bc                           |           -850
