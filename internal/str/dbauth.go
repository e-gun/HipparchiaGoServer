package str

type DbAuthor struct {
	UID       string
	Language  string
	IDXname   string
	Name      string
	Shortname string
	Cleaname  string
	Genres    string
	RecDate   string
	ConvDate  int
	Location  string
	// beyond the DB starts here
	WorkList []string
}
