package main

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type Session struct {
	ID             string
	Inclusions     SearchIncExl
	Exclusions     SearchIncExl
	ActiveCorp     map[string]bool
	VariaOK        bool
	IncertaOK      bool
	SpuriaOK       bool
	AvailDBs       map[string]bool `json:"available"`
	VectorVals     bool
	UISettings     bool
	OutPutSettings bool
	Analogyfinder  bool   `json:"analogyfinder"`
	Authorflagging bool   `json:"authorflagging"`
	Authorssummary bool   `json:"authorssummary"`
	Baggingmethod  string `json:"baggingmethod"`
	HitLimit       int64
	Earliest       string
	Latest         string
}

func makedefaultsession(id string) Session {
	var s Session
	s.ID = id
	// this format is out of sync w/ the JS but necc. for the searching code ATM: lt vs latincorpus, etc
	s.ActiveCorp = map[string]bool{"gr": true, "lt": true, "in": false, "ch": false, "dp": false}
	s.VariaOK = true
	s.IncertaOK = true
	s.SpuriaOK = true
	s.AvailDBs = map[string]bool{"greek_dictionary": true, "greek_lemmata": true, "greek_morphology": true, "latin_dictionary": true, "latin_lemmata": true, "latin_morphology": true, "wordcounts_0": true}
	s.Analogyfinder = false
	s.HitLimit = 200
	s.Inclusions.DateRange = [2]string{"-850", "1500"}

	return s
}

func RtFrontpage(c echo.Context) error {
	id := readUUIDCookie(c)
	if _, t := sessions[id]; !t {
		sessions[id] = makedefaultsession(id)
	}
	err := c.File("static/html/frontpage.html")
	if err != nil {
		return nil
	}
	return nil
}

func readUUIDCookie(c echo.Context) string {
	cookie, err := c.Cookie("ID")
	if err != nil {
		id := writeUUIDCookie(c)
		return id
	}
	id := cookie.Value
	return id
}

func writeUUIDCookie(c echo.Context) string {
	cookie := new(http.Cookie)
	cookie.Name = "ID"
	cookie.Value = uuid.New().String()
	cookie.Expires = time.Now().Add(4800 * time.Hour)
	c.SetCookie(cookie)
	return cookie.Value
}

// sample python session:
// {"_fresh": "no", "agnexclusions": [], "agnselections": [], "alocexclusions": [], "alocselections": [], "analogyfinder": "no",
//"auexclusions": [], "auselections": [], "authorflagging": "yes", "authorssummary": "yes",
// "available": {"greek_dictionary": true, "greek_lemmata": true, "greek_morphology": true, "latin_dictionary": true, "latin_lemmata": true, "latin_morphology": true, "wordcounts_0": true},
// "baggingmethod": "winnertakesall", "bracketangled": "yes", "bracketcurly": "yes", "bracketround": "no", "bracketsquare": "yes",
// "browsercontext": "24", "christiancorpus": "no", "collapseattic": "yes", "cosdistbylineorword": "no", "cosdistbysentence": "no",
// "debugdb": "no", "debughtml": "no", "debuglex": "no", "debugparse": "no", "earliestdate": "-850", "fontchoice": "Noto",
// "greekcorpus": "yes", "headwordindexing": "no", "incerta": "yes", "indexbyfrequency": "no", "indexskipsknownwords": "no",
// "inscriptioncorpus": "no", "latestdate": "1500", "latincorpus": "yes", "ldacomponents": 7, "ldaiterations": 12,
// "ldamaxfeatures": 2000, "ldamaxfreq": 35, "ldaminfreq": 5, "ldamustbelongerthan": 3, "linesofcontext": 4,
// "loggedin": "no", "maxresults": "200", "morphdialects": "no", "morphduals": "yes", "morphemptyrows": "yes",
// "morphfinite": "yes", "morphimper": "yes", "morphinfin": "yes", "morphpcpls": "yes", "morphtables": "yes",
// "nearestneighborsquery": "no", "nearornot": "near", "onehit": "no", "papyruscorpus": "no", "phrasesummary": "no",
// "principleparts": "yes", "proximity": "1", "psgexclusions": [], "psgselections": [], "quotesummary": "yes",
// "rawinputstyle": "no", "searchinsidemarkup": "no", "searchscope": "lines", "semanticvectorquery": "no",
// "sensesummary": "yes", "sentencesimilarity": "no", "showwordcounts": "yes", "simpletextoutput": "no",
// "sortorder": "shortname", "spuria": "yes", "suppresscolors": "no", "tensorflowgraph": "no", "topicmodel": "no",
// "trimvectoryby": "none", "userid": "Anonymous", "varia": "yes", "vcutlem": 50, "vcutloc": 33, "vcutneighb": 15,
// "vdim": 300, "vdsamp": 5, "viterat": 12, "vminpres": 10, "vnncap": 15, "vsentperdoc": 1, "vwindow": 10,
// "wkexclusions": [], "wkgnexclusions": [], "wkgnselections": [], "wkselections": [], "wlocexclusions": [],
// "wlocselections": [], "xmission": "Any", "zaplunates": "no", "zapvees": "no"}
