//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vv

import "time"

const (
	DEFAULTECHOLOGLEVEL = 0
	DEFAULTGOLOGLEVEL   = 0
	POLLEVERYNTABLES    = 34 // 3455 is the max number of tables in a search...
	TICKERISACTIVE      = false
	TICKERDELAY         = 3 * time.Second
	TICKERLINES         = 25
	TIMEOUTRD           = 15 * time.Second  // only set if Config.Authenticate is true (and so in a "serve the net" situation)
	TIMEOUTWR           = 120 * time.Second // this is *very* generous, but some searches are slow/long
	USEGZIP             = false
	WSPOLLINGPAUSE      = 10000000 * 10 // 10000000 * 10 = every .1s
)
