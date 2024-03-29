//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package str

import "html/template"

type PrerolledQuery struct {
	TempTable string
	PsqlQuery string
}

type QueryBounds struct {
	Start int
	Stop  int
}

type PRQTemplate struct {
	AU    string
	COL   string
	SYN   string
	SK    string
	LIM   string
	IDX   string
	TTN   string
	Tail  *template.Template
	PSCol string
}
