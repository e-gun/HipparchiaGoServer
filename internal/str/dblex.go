//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package str

import "strings"

type DbLexicon struct {
	// skipping 'unaccented_entry' from greek_dictionary
	// skipping 'entry_key' from latin_dictionary
	Word     string
	Metrical string
	ID       float32
	POS      string
	Transl   string
	Entry    string
	// not part of the table...
	lang string // must be lower-case because of the call to pgx.RowToStructByPos[DbLexicon]
}

func (dbl *DbLexicon) SetLang(l string) {
	dbl.lang = l
}

func (dbl *DbLexicon) GetLang() string {
	return dbl.lang
}

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
