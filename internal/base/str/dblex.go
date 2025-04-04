//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package str

import (
	"bytes"
	"fmt"
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

// hipparchiaDB-# \d latin_dictionary
//                     Table "public.latin_dictionary"
//     Column     |          Type          | Collation | Nullable | Default
//----------------+------------------------+-----------+----------+---------
// entry_name     | character varying(256) |           |          |
// metrical_entry | character varying(256) |           |          |
// id_number      | real                   |           |          |
// entry_key      | character varying(64)  |           |          |
// pos            | character varying(64)  |           |          |
// translations   | text                   |           |          |
// entry_body     | text                   |           |          |
// html_body      | text                   |           |          |
//Indexes:
//    "latin_dictionary_idx" btree (entry_name)

//type DbLexicon struct {
//	// skipping 'unaccented_entry' from greek_dictionary
//	// skipping 'entry_key' from latin_dictionary
//	Word     string
//	Metrical string
//	ID       float32
//	POS      string
//	Transl   string
//	Entry    string
//	// not part of the table...
//	lang string // must be lower-case because of the call to pgx.RowToStructByPos[DbLexicon]
//}

type LSJEntry struct {
	PrelimInfo string
	EntryType  string
	IDStr      string
	IDVal      string
	Key        string
	SenseIDs   []string
	Senses     []LSJSense
}

type LSJSense struct {
	ID       string
	N        string
	LVL      string
	Contents string
}

type DbLexicon struct {
	EntryName  string
	EntryMetr  string
	IDVal      string // string like "n12345a" in the original data
	EntryType  string
	POS        string
	Transl     string
	Usedby     string
	PrelimInfo string
	SenseIDs   string // HGb: row[6] = strings.Join(entry.SenseIDs, " ")
	Senses     []LSJSense
	// not part of the table...
	lang string // must be lower-case because of the call to pgx.RowToStructByPos[DbLexicon]
}

func (dbl *DbLexicon) SetLang(l string) {
	dbl.lang = l
}

func (dbl *DbLexicon) GetLang() string {
	return dbl.lang
}

func (ent *DbLexicon) PrintOut() {
	const (
		TMPL = `
	EntryName     {{.EntryName}}
	EntryMetr     {{.EntryMetr}}
	IDVal         {{.IDVal}}
	Transl        {{.Transl}}
	Usedby        {{.Usedby}}
	SenseIDs      {{.SenseIDs}}`
	)
	m := map[string]string{
		"EntryName": ent.EntryName,
		"EntryMetr": ent.EntryMetr,
		"IDVal":     ent.IDVal,
		"Transl":    ent.Transl,
		"Usedby":    ent.Usedby,
		"SenseIDs":  ent.SenseIDs,
	}

	t := template.Must(template.New("").Parse(TMPL))

	var b bytes.Buffer
	if ee := t.Execute(&b, m); ee != nil {
		fmt.Println(ee)
	}
	fmt.Println(b.String())
}
