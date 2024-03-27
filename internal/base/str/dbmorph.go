//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package str

import "strings"

// hipparchiaDB=# \d greek_morphology
//                           Table "public.greek_morphology"
//          Column           |          Type          | Collation | Nullable | Default
//---------------------------+------------------------+-----------+----------+---------
// observed_form             | character varying(64)  |           |          |
// xrefs                     | character varying(128) |           |          |
// prefixrefs                | character varying(128) |           |          |
// possible_dictionary_forms | jsonb                  |           |          |
// related_headwords         | character varying(256) |           |          |
//Indexes:
//    "greek_analysis_trgm_idx" gin (related_headwords gin_trgm_ops)
//    "greek_morphology_idx" btree (observed_form)

type DbMorphology struct {
	Observed    string
	Xrefs       string
	PrefixXrefs string
	RawPossib   string
	RelatedHW   string
}

func (dbm *DbMorphology) PossibSlice() []string {
	return strings.Split(dbm.RawPossib, " ")
}

type DbWordCount struct {
	Word  string
	Total int
	Gr    int
	Lt    int
	Dp    int
	In    int
	Ch    int
}

type MorphPossib struct {
	Transl   string `json:"transl"`
	Anal     string `json:"analysis"`
	Headwd   string `json:"headword"`
	Scansion string `json:"scansion"`
	Xrefkind string `json:"xref_kind"`
	Xrefval  string `json:"xref_value"`
}
