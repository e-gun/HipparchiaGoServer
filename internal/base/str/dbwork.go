//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package str

var (
	TLGAbbrev = "gr" // should inject vv.TLGABBREV value at launch
	LATAbbrev = "lt" // should inject vv.LATABBREV value at launch vv.LATABBREV
)

type DbWork struct {
	UID       string
	Title     string
	Language  string
	Pub       string
	LL0       string
	LL1       string
	LL2       string
	LL3       string
	LL4       string
	LL5       string
	Genre     string
	Xmit      string
	Type      string
	Prov      string
	RecDate   string
	ConvDate  int
	WdCount   int
	FirstLine int
	LastLine  int
	Authentic bool
}

// WkID - ex: gr2017w068 --> 068
func (dbw *DbWork) WkID() string {
	return dbw.UID[LengthOfAuthorID+1:]
}

// AuID - ex: gr2017w068 --> gr2017
func (dbw *DbWork) AuID() string {
	if len(dbw.UID) < LengthOfAuthorID {
		return ""
	} else {
		return dbw.UID[:LengthOfAuthorID]
	}
}

// Corpus() - ex: gr2017w068 --> gr
func (dbw *DbWork) Corpus() string {
	return dbw.UID[0:LenthOfCorpusAbbrev]
}

func (dbw *DbWork) IsPHI() bool {
	// can't use vv: circular imports
	// so this could bite you some day...
	if dbw.UID[0:LenthOfCorpusAbbrev] != TLGAbbrev && dbw.UID[0:LenthOfCorpusAbbrev] != LATAbbrev {
		return true
	} else {
		return false
	}
}

func (dbw *DbWork) CitationFormat() []string {
	cf := []string{
		dbw.LL5, dbw.LL4, dbw.LL3, dbw.LL2, dbw.LL1, dbw.LL0,
	}
	return cf
}

// CountLevels - the work structure employs how many levels?
func (dbw *DbWork) CountLevels() int {
	ll := 0
	for _, l := range []string{dbw.LL5, dbw.LL4, dbw.LL3, dbw.LL2, dbw.LL1, dbw.LL0} {
		if len(l) > 0 {
			ll += 1
		}
	}
	return ll
}

// DateInRange - is the work dated between X and Y?
func (dbw DbWork) DateInRange(earliest int, latest int) bool {
	if earliest >= dbw.ConvDate && dbw.ConvDate <= latest {
		return true
	} else {
		return false
	}
}

// Length - how many db lines?
func (dbw DbWork) Length() int {
	if dbw.LastLine == 0 {
		// this would be a PHI build error in RemapInscriptionAuthorsAndWorks(): should be gone by now but for three
		// one-line zero-word stragglers that need checking...

		// hgdb=> select universalid,title,transmission,wordcount,firstline,lastline from works where lastline = 0 and universalid ~* '^in';;
		// universalid |                    title                     |  transmission   | wordcount | firstline | lastline
		//-------------+----------------------------------------------+-----------------+-----------+-----------+----------
		// inz00aw00a  |  (C. Crete, Aigaion Antron)                  | direct (in0130) |         0 |         0 |        0
		// inz093w00a  | SEG 40.426; SEG 19.375 (Phokis, Delphi)      | direct (in0040) |         0 |         0 |        0
		// inz05lw00a  | EG 1.343,2[LSAG 265,3] (Sikelia, Syrakousai) | direct (in0190) |         0 |         0 |        0
		//(3 rows)
		return 1
	} else {
		return dbw.LastLine - dbw.FirstLine
	}
}
