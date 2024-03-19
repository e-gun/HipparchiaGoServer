//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package mps

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/str"
)

// MapNewWorkCorpus - add a corpus to a workmap
func MapNewWorkCorpus(corpus string, workmap map[string]*str.DbWork) map[string]*str.DbWork {
	const (
		MSG = "MapNewWorkCorpus() added %d works from '%s'"
	)
	toadd := sliceworkcorpus(corpus)
	for i := 0; i < len(toadd); i++ {
		w := toadd[i]
		workmap[w.UID] = &w
	}

	LoadedCorp[corpus] = true

	Msg.PEEK(fmt.Sprintf(MSG, len(toadd), corpus))
	return workmap
}

//used to be a method on the struct but that yielded import problems

// DbWkMyAu - return the work's DbAuthor
func DbWkMyAu(dbw *str.DbWork) *str.DbAuthor {
	a, ok := AllAuthors[dbw.AuID()]
	if !ok {
		Msg.WARN(fmt.Sprintf("mps.DbWkMyAu() failed to find '%s'", dbw.AuID()))
		a = &str.DbAuthor{}
	}
	return a
}
