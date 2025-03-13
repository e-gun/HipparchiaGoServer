//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vv

const (
	DEFAULTCOLUMN      = "stripped_line"
	DEFAULTPSQLHOST    = "127.0.0.1"
	DEFAULTPSQLUSER    = "hippa_wr"
	DEFAULTPSQLPORT    = 5432
	DEFAULTPSQLDB      = "hipparchiaDB"
	DEFAULTQUERYSYNTAX = "~"
	SERVEDFROMHOST     = "127.0.0.1"
	SERVEDFROMPORT     = 8000
	SERVEDFROMSSLPORT  = 4443
	TEMPTABLETHRESHOLD = 100 // if a table requires N "between" clauses, build a temptable instead to gather the needed lines
)
