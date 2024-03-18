package lnch

import (
	"github.com/e-gun/HipparchiaGoServer/internal/structs"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
)

// todo: import issue; put this function elsewhere
// BuildSelectionOverview will call the relevant SearchIncExl functions: see buildesearchlist.go
//func (s *ServerSession) BuildSelectionOverview() {
//	s.Inclusions.BuildAuByName()
//	s.Exclusions.BuildAuByName()
//	s.Inclusions.BuildWkByName()
//	s.Exclusions.BuildWkByName()
//	s.Inclusions.BuildPsgByName()
//	s.Exclusions.BuildPsgByName()
//}

// MakeDefaultSession - fill in the blanks when setting up a new session
func MakeDefaultSession(id string) structs.ServerSession {
	// note that SessionMap clears every time the server restarts

	var s structs.ServerSession
	s.ID = id
	s.ActiveCorp = Config.DefCorp
	s.VariaOK = true
	s.IncertaOK = true
	s.SpuriaOK = true
	s.NearOrNot = "near"
	s.HitLimit = vv.DEFAULTHITLIMIT
	s.Earliest = vv.MINDATESTR
	s.Latest = vv.MAXDATESTR
	s.SortHitsBy = vv.SORTBY
	s.HitContext = vv.DEFAULTLINESOFCONTEXT
	s.BrowseCtx = Config.BrowserCtx
	s.SearchScope = vv.DEFAULTPROXIMITYSCOPE
	s.Proximity = vv.DEFAULTPROXIMITY
	s.LoginName = "Anonymous"
	s.VocScansion = Config.VocabScans
	s.VocByCount = Config.VocabByCt
	s.VecGraphExt = Config.VectorWebExt
	s.VecNeighbCt = Config.VectorNeighb
	s.VecNNSearch = false
	s.VecModeler = Config.VectorModel
	s.VecTextPrep = Config.VectorTextPrep
	s.VecLDASearch = false
	s.LDA2D = true

	// todo: refactor to avoid circularity
	//if Config.Authenticate {
	//	AllAuthorized.Register(id, false)
	//} else {
	//	AllAuthorized.Register(id, true)
	//}

	//m("MakeDefaultSession() in non-default lnch for testing; this is not a release build of HGS", 0)
	//
	//s.VecLDASearch = true
	//s.VecNNSearch = true

	//m := make(map[string]string)
	//m["lt0917_FROM_1431_TO_2193"] = "Lucanus, Marcus Annaeus, Bellum Civile, 3"
	//m["lt0917_FROM_2_TO_692"] = "Lucanus, Marcus Annaeus, Bellum Civile, 1"
	//m["lt0917_FROM_5539_TO_6410"] = "Lucanus, Marcus Annaeus, Bellum Civile, 8"
	//m["lt0917_FROM_6411_TO_7520"] = "Lucanus, Marcus Annaeus, Bellum Civile, 9"
	//m["lt0917_FROM_4666_TO_5538"] = "Lucanus, Marcus Annaeus, Bellum Civile, 7"
	//m["lt0917_FROM_3019_TO_3835"] = "Lucanus, Marcus Annaeus, Bellum Civile, 5"
	//s.Inclusions.Passages = []string{"lt0917_FROM_6411_TO_7520", "lt0917_FROM_4666_TO_5538", "lt0917_FROM_3019_TO_3835",
	//	"lt0917_FROM_1431_TO_2193", "lt0917_FROM_2_TO_692", "lt0917_FROM_5539_TO_6410"}
	//s.Inclusions.MappedPsgByName = m
	//s.Proximity = 4
	//s.SearchScope = "words"
	//s.Inclusions.BuildPsgByName()

	return s
}
