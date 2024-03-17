package structs

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
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
	return dbw.UID[vv.LENGTHOFAUTHORID+1:]
}

// AuID - ex: gr2017w068 --> gr2017
func (dbw *DbWork) AuID() string {
	if len(dbw.UID) < vv.LENGTHOFAUTHORID {
		return ""
	} else {
		return dbw.UID[:vv.LENGTHOFAUTHORID]
	}
}

// todo: refactor to avoid circularity
// MyAu - return the work's DbAuthor
func (dbw *DbWork) MyAu() *DbAuthor {
	a, ok := AllAuthors[dbw.AuID()]
	if !ok {
		msg.WARN(fmt.Sprintf("DbWork.MyAu() failed to find '%s'", dbw.AuID()))
		a = &DbAuthor{}
	}
	return a
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
