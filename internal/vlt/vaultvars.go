package vlt

var (
	AllSessions   = MakeSessionVault()
	AllAuthorized = MakeAuthorizedVault()
	WebsocketPool = WSFillNewPool()
	WSInfo        = BuildWSInfoHubIf()
)
