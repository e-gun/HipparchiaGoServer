//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package search

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/mps"
	"github.com/e-gun/HipparchiaGoServer/internal/str"
)

//
// The following used to be a method on the struct but that yielded import problems
//

// DbWlnMyWk - get the DbWork for this line
func DbWlnMyWk(dbw *str.DbWorkline) *str.DbWork {
	w, ok := mps.AllWorks[dbw.WkUID]
	if !ok {
		Msg.WARN(fmt.Sprintf("search.DbWlnMyWk() failed to find '%s'", dbw.AuID()))
		w = &str.DbWork{}
	}
	return w
}

// DbWlnMyAu - get the DbAuthor for this line
func DbWlnMyAu(dbw *str.DbWorkline) *str.DbAuthor {
	a, ok := mps.AllAuthors[dbw.AuID()]
	if !ok {
		Msg.WARN(fmt.Sprintf("search.DbWkMyAu() failed to find '%s'", dbw.AuID()))
		a = &str.DbAuthor{}
	}
	return a
}
