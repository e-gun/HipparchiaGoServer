package structs

//
// SERVERSESSIONS
//

type ServerSession struct {
	ID           string
	Inclusions   SearchIncExl
	Exclusions   SearchIncExl
	ActiveCorp   map[string]bool
	VariaOK      bool   `json:"varia"`
	IncertaOK    bool   `json:"incerta"`
	SpuriaOK     bool   `json:"spuria"`
	RawInput     bool   `json:"rawinputstyle"`
	OneHit       bool   `json:"onehit"`
	HeadwordIdx  bool   `json:"headwordindexing"`
	FrqIdx       bool   `json:"indexbyfrequency"`
	VocByCount   bool   `json:"vocbycount"`
	VocScansion  bool   `json:"vocscansion"`
	NearOrNot    string `json:"nearornot"`
	SearchScope  string `json:"searchscope"`
	SortHitsBy   string `json:"sortorder"`
	Proximity    int    `json:"proximity"`
	BrowseCtx    int
	InputStyle   string
	HitLimit     int
	HitContext   int
	Earliest     string
	Latest       string
	TmpInt       int
	TmpStr       string
	LoginName    string
	VecGraphExt  bool
	VecModeler   string
	VecNeighbCt  int
	VecNNSearch  bool
	VecTextPrep  string
	VecLDASearch bool
	LDAgraph     bool
	LDAtopics    int
	LDA2D        bool
}

// BuildSelectionOverview will call the relevant SearchIncExl functions: see buildesearchlist.go
func (s *ServerSession) BuildSelectionOverview() {
	//s.Inclusions.BuildAuByName()
	//s.Exclusions.BuildAuByName()
	//s.Inclusions.BuildWkByName()
	//s.Exclusions.BuildWkByName()
	s.Inclusions.BuildPsgByName()
	s.Exclusions.BuildPsgByName()
}
