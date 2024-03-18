package mps

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/structs"
)

// MapNewWorkCorpus - add a corpus to a workmap
func MapNewWorkCorpus(corpus string, workmap map[string]*structs.DbWork) map[string]*structs.DbWork {
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
func DbWkMyAu(dbw *structs.DbWork) *structs.DbAuthor {
	a, ok := AllAuthors[dbw.AuID()]
	if !ok {
		Msg.WARN(fmt.Sprintf("mps.DbWkMyAu() failed to find '%s'", dbw.AuID()))
		a = &structs.DbAuthor{}
	}
	return a
}
