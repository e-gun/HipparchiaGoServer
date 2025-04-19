//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package str

import (
	"bytes"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"strings"
	"text/template"
)

// hipparchiaDB-# \d latin_morphology
//                           Table "public.latin_morphology"
//          Column           |          Type          | Collation | Nullable | Default
//---------------------------+------------------------+-----------+----------+---------
// observed_form             | character varying(64)  |           |          |
// xrefs                     | character varying(128) |           |          |
// prefixrefs                | character varying(128) |           |          |
// possible_dictionary_forms | jsonb                  |           |          |
// related_headwords         | character varying(256) |           |          |
//Indexes:
//    "latin_analysis_trgm_idx" gin (related_headwords gin_trgm_ops)
//    "latin_morphology_idx" btree (observed_form)

// hgdb=> \d greek_dictionary
//                     Table "public.greek_dictionary"
//    Column    |          Type           | Collation | Nullable | Default
//--------------+-------------------------+-----------+----------+---------
// idval        | double precision        |           |          |
// entry_name   | character varying(256)  |           |          |
// entry_metr   | character varying(256)  |           |          |
// idname       | character varying(16)   |           |          |
// entry_type   | character varying(16)   |           |          |
// translations | text                    |           |          |
// usedby       | character varying(1500) |           |          |
// prelim_info  | text                    |           |          |
// sense_ids    | text                    |           |          |
// senses       | jsonb                   |           |          |
//Indexes:
//    "greek_dictionary_idval_key" UNIQUE CONSTRAINT, btree (idval)

type DbLexicon struct {
	IdFloat    float32
	EntryName  string
	EntryMetr  string
	IdString   string // string like "n12345a" in the original data
	EntryType  string
	Transl     string
	Usedby     string
	PrelimInfo string
	SenseIDs   []string // HGB: row[7] = strings.Join(entry.SenseIDs, " ")
	Senses     []LexicalSenses
	// not part of the table...
	lang string // must be lower-case because of the call to pgx.RowToStructByPos[DbLexicon]
}

func (dbl *DbLexicon) SetLang(l string) {
	dbl.lang = l
}

func (dbl *DbLexicon) GetLang() string {
	if gen.IsLatin.MatchString(dbl.EntryName) {
		return "latin"
	} else {
		return "greek"
	}
}

func (dbl *DbLexicon) IsLatin() bool {
	if gen.IsLatin.MatchString(dbl.EntryName) {
		return true
	} else {
		return false
	}
}

func (ent *DbLexicon) PrintOut() {
	const (
		TMPL = `
	EntryName     {{.EntryName}}
	EntryMetr     {{.EntryMetr}}
	IDVal         {{.IdString}}
	Transl        {{.Transl}}
	Usedby        {{.Usedby}}
	SenseIDs      {{.SenseIDs}}`
	)
	m := map[string]string{
		"EntryName": ent.EntryName,
		"EntryMetr": ent.EntryMetr,
		"IDVal":     ent.IdString,
		"Transl":    ent.Transl,
		"Usedby":    ent.Usedby,
		"SenseIDs":  strings.Join(ent.SenseIDs, ";"),
	}

	t := template.Must(template.New("").Parse(TMPL))

	var b bytes.Buffer
	if ee := t.Execute(&b, m); ee != nil {
		fmt.Println(ee)
	}
	fmt.Println(b.String())
}

type LexicalSenses struct {
	ID       string `json:"id"`
	N        string `json:"n"`
	LVL      string `json:"lvl"`
	Contents string `json:"contents"`
}
