//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package str

type PostgresLogin struct {
	Host   string
	Port   int
	User   string
	Pass   string
	DBName string
}
