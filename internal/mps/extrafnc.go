package mps

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/structs"
)

//used to be a method on the struct but that yielded import problems

// MyAu - return the work's DbAuthor
func MyAu(dbw *structs.DbWork) *structs.DbAuthor {
	a, ok := AllAuthors[dbw.AuID()]
	if !ok {
		msg.WARN(fmt.Sprintf("DbWork.MyAu() failed to find '%s'", dbw.AuID()))
		a = &structs.DbAuthor{}
	}
	return a
}
