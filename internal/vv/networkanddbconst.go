//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vv

const (
	DEFAULTCOLUMN      = "stripped_line"
	DEFAULTPSQLHOST    = "127.0.0.1"
	DEFAULTPSQLUSER    = "hgdbuser"
	DEFAULTPSQLPORT    = 5432
	DEFAULTPSQLDB      = "hgdb"
	DEFAULTQUERYSYNTAX = "~"
	SERVEDFROMHOST     = "127.0.0.1"
	SERVEDFROMPORT     = 8001
	SERVEDFROMSSLPORT  = 4443
	TEMPTABLETHRESHOLD = 100 // if a table requires N "between" clauses, build a temptable instead to gather the needed lines
)
