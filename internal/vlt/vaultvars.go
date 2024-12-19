//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vlt

var (
	AllSessions   = makesessionvault()
	AllAuthorized = makeauthorizedvault()
	WebsocketPool = wsfillnewpool()
	WSInfo        = buildwsinfohubif()
)
