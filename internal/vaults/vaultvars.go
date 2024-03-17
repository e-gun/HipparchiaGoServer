package vaults

var (
	AllSessions   = MakeSessionVault()
	AllAuthorized = MakeAuthorizedVault()
	WebsocketPool = WSFillNewPool()
	WSInfo        = BuildWSInfoHubIf()
)
