//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vlt

var (
	AllSessions   = MakeSessionVault()
	AllAuthorized = MakeAuthorizedVault()
	WebsocketPool = WSFillNewPool()
	WSInfo        = BuildWSInfoHubIf()
)
