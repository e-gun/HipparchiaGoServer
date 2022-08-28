//    HipparchiaGoDBHelper: search and vector helper app and functions for HipparchiaServer
//    Copyright: E Gunderson 2021
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

type DbWordCount struct {
	Word  string
	Total int64
	Gr    int64
	Lt    int64
	Dp    int64
	In    int64
	Ch    int64
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

// JSStruct - this is really just for generating JSON
type JSStruct struct {
	V string `json:"value"`
}
