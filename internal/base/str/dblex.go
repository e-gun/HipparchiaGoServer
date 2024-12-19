//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package str

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
