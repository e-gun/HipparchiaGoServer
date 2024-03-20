//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package mps

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
)

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
