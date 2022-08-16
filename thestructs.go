//    HipparchiaGoDBHelper: search and vector helper app and functions for HipparchiaServer
//    Copyright: E Gunderson 2021
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

type PrerolledQuery struct {
	TempTable string
	PsqlQuery string
	PsqlData  string
}

type DbAuthor struct {
	UID       string
	Language  string
	IDXname   string
	Name      string
	Shortname string
	Cleaname  string
	Genres    string
	RecDate   string
	ConvDate  int64
	Location  string
}

type DbWork struct {
	UID       string
	Title     string
	Language  string
	Pub       string
	LL0       string
	LL1       string
	LL2       string
	LL3       string
	LL4       string
	LL5       string
	Genre     string
	Xmit      string
	Type      string
	Prov      string
	RecDate   string
	ConvDate  int64
	WdCount   int64
	FirstLine int64
	LastLine  int64
	Authentic bool
	// not in the DB, but derived: gr2017w068 --> 068
	WorkNum string
}

func (dbw DbWork) FindWorknumber() string {
	// ex: gr2017w068
	return dbw.UID[7:]
}

func (dbw DbWork) FindAuthor() string {
	// ex: gr2017w068
	return dbw.UID[:6]
}

func (dbw DbWork) CitationFormat() []string {
	cf := []string{
		dbw.LL5,
		dbw.LL4,
		dbw.LL3,
		dbw.LL2,
		dbw.LL1,
		dbw.LL0,
	}
	return cf
}

type DbWorkline struct {
	WkUID       string
	TbIndex     int64
	Lvl5Value   string
	Lvl4Value   string
	Lvl3Value   string
	Lvl2Value   string
	Lvl1Value   string
	Lvl0Value   string
	MarkedUp    string
	Accented    string
	Stripped    string
	Hypenated   string
	Annotations string
}

func (dbw DbWorkline) FindLocus() []string {
	loc := [6]string{
		dbw.Lvl5Value,
		dbw.Lvl4Value,
		dbw.Lvl3Value,
		dbw.Lvl2Value,
		dbw.Lvl1Value,
		dbw.Lvl0Value,
	}

	var trim []string
	for _, l := range loc {
		if l != "-1" {
			trim = append(trim, l)
		}
	}
	return trim
}

func (dbw DbWorkline) FindAuthor() string {
	return dbw.WkUID[:6]
}

type DbWordCount struct {
	Word  string
	Total int64
	Gr    int64
	Lt    int64
	Dp    int64
	In    int64
	Ch    int64
}

type DbLexicon struct {
	// skipping 'unaccented_entry' from greek_dictionary
	// skipping 'entry_key' from latin_dictionary
	Word     string
	Metrical string
	ID       int64
	POS      string
	Transl   string
	Entry    string
}

type RedisLogin struct {
	Addr     string
	Password string
	DB       int
}

type PostgresLogin struct {
	Host   string
	Port   int
	User   string
	Pass   string
	DBName string
}

// https://golangbyexample.com/sort-custom-struct-collection-golang/
type WeightedHeadword struct {
	Word  string
	Count int
}

type WHWList []WeightedHeadword

func (w WHWList) Len() int {
	return len(w)
}

func (w WHWList) Less(i, j int) bool {
	return w[i].Count > w[j].Count
}

func (w WHWList) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
}

type BagWithLocus struct {
	Loc string
	Bag string
}

type DbMorphology struct {
	Observed    string
	Xrefs       string
	PrefixXrefs string
	RawPossib   string
	RelatedHW   string
}

type MorphPossib struct {
	Transl   string `json:"transl"`
	Anal     string `json:"analysis"`
	Headwd   string `json:"headword"`
	Scansion string `json:"scansion"`
	Xrefkind string `json:"xref_kind"`
	Xrefval  string `json:"xref_value"`
}

type CompositePollingData struct {
	// this has to be kept in sync with rediskeys[8] and HipparchiaServer's interface
	Launchtime    float64
	Active        string // redis polls store 'yes' or 'no'; but the value is converted to T/F by .getactivity()
	Statusmessage string
	Remaining     int64
	Poolofwork    int64
	Hitcount      int64
	Portnumber    int64
	Notes         string
	ID            string // this is not stored in redis; it is asserted here
}

type BrowsedPassage struct {
	// marshal will not do lc names
	Browseforwards    string `json:"browseforwards"`
	Browseback        string `json:"browseback"`
	Authornumber      string `json:"authornumber"`
	Workid            string `json:"workid"`
	Worknumber        string `json:"worknumber"`
	Authorboxcontents string `json:"authorboxcontents"`
	Workboxcontents   string `json:"workboxcontents"`
	Browserhtml       string `json:"browserhtml"`
}

type CurrentConfiguration struct {
	RedisKey        string
	MaxHits         int64
	WorkerCount     int
	LogLevel        int
	RedisInfo       string
	PosgresInfo     string
	BagMethod       string
	SentPerBag      int
	VectTestDB      string
	VectStart       int
	VectEnd         int
	VSkipHW         string
	VSkipInf        string
	BrowseAuthor    string
	BrowseWork      string
	BrowseFoundline int64
	BrowseContext   int64
	IsVectPtr       *bool
	IsWSPtr         *bool
	IsBrPtr         *bool
	WSPort          int
	WSFail          int
	WSSave          int
	ProfCPUPtr      *bool
	ProfMemPtr      *bool
	SendVersPtr     *bool
	RLogin          RedisLogin
	PGLogin         PostgresLogin
}
