//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package db

import (
	"context"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"strings"
)

// GetMorphMatch - word into []DbMorphology
func GetMorphMatch(word string, lang string) []str.DbMorphology {
	const (
		FLDS = `observed_form, xrefs, prefixrefs, possible_dictionary_forms, related_headwords`
		PSQQ = "SELECT %s FROM %s_morphology WHERE observed_form = '%s'"
	)

	psq := fmt.Sprintf(PSQQ, FLDS, lang, word)

	foundrows, err := SQLPool.Query(context.Background(), psq)
	Msg.EC(err)

	thesefinds, err := pgx.CollectRows(foundrows, pgx.RowToStructByPos[str.DbMorphology])
	Msg.EC(err)

	return thesefinds
}

// ArrayToGetRequiredMorphObjects - map a slice of words to the corresponding DbMorphology
func ArrayToGetRequiredMorphObjects(wordlist []string) map[string]str.DbMorphology {
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

	const (
		TT = `CREATE TEMPORARY TABLE ttw_%s AS SELECT words AS w FROM unnest(ARRAY[%s]) words`
		QT = `SELECT observed_form, xrefs, prefixrefs, possible_dictionary_forms, related_headwords FROM %s_morphology WHERE EXISTS 
				(SELECT 1 FROM ttw_%s temptable WHERE temptable.w = %s_morphology.observed_form)`
		MSG1      = "ArrayToGetRequiredMorphObjects() will search among %d words"
		CHUNKSIZE = 999999
	)

	dbconn := GetDBConnection()
	defer dbconn.Release()

	// look for the upper case matches too: Ϲωκράτηϲ and not just ϲωκρατέω (!)
	uppers := make([]string, len(wordlist))
	for i := 0; i < len(wordlist); i++ {
		uppers[i] = strings.Title(wordlist[i])
	}

	// γ': a lot of cycles looking for a small number of words...
	apo := make([]string, len(wordlist))
	for i := 0; i < len(wordlist); i++ {
		// need to escape the single quote
		// hipparchiaDB=# select * from greek_morphology where observed_form = 'οὑφ'''
		apo[i] = wordlist[i] + "''"
	}

	wordlist = append(wordlist, uppers...)
	wordlist = append(wordlist, apo...)

	Msg.PEEK(fmt.Sprintf(MSG1, len(wordlist)))

	foundmorph := make(map[string]str.DbMorphology)
	var thehit str.DbMorphology

	foreach := []any{&thehit.Observed, &thehit.Xrefs, &thehit.PrefixXrefs, &thehit.RawPossib, &thehit.RelatedHW}

	rwfnc := func() error {
		foundmorph[thehit.Observed] = thehit
		return nil
	}

	// vectorization revealed that 10m words is too much for this function
	// [HGS] ArrayToGetRequiredMorphObjects() will search for 10708941 words
	// [Hipparchia Golang Server v.1.2.0a] UNRECOVERABLE ERROR: PLEASE TAKE NOTE OF THE FOLLOWING PANIC MESSAGE
	// ERROR: invalid memory alloc request size 1073741824 (SQLSTATE XX000)

	// this could be parallelized...

	chunkedlist := gen.ChunkSlice(wordlist, CHUNKSIZE)
	for _, cl := range chunkedlist {
		// a waste of time to check the language on every word; just flail/fail once
		for _, uselang := range vv.TheLanguages {
			u := strings.Replace(uuid.New().String(), "-", "", -1)
			id := fmt.Sprintf("%s_%s_mw", u, uselang)
			a := fmt.Sprintf("'%s'", strings.Join(cl, "', '"))
			t := fmt.Sprintf(TT, id, a)

			_, err := dbconn.Exec(context.Background(), t)
			Msg.EC(err)

			foundrows, e := dbconn.Query(context.Background(), fmt.Sprintf(QT, uselang, id, uselang))
			Msg.EC(e)

			_, ee := pgx.ForEachRow(foundrows, foreach, rwfnc)
			Msg.EC(ee)
		}
	}
	return foundmorph
}

// GetAllFormsOf - convert an 'xref' into a map of all possible forms of that word
func GetAllFormsOf(lg string, xr string) map[string]str.DbMorphology {
	const (
		MFLD = `observed_form, xrefs, prefixrefs, possible_dictionary_forms, related_headwords`
		MQT  = `SELECT %s FROM %s_morphology WHERE xrefs ~ '%s' AND prefixrefs=''`
	)
	// for ἐπιγιγνώϲκω...
	// select * from greek_morphology where greek_morphology.xrefs='37925260';

	dbconn := GetDBConnection()
	defer dbconn.Release()

	// hipparchiaDB=# select observed_form, xrefs from latin_morphology where observed_form = 'crediti';
	// observed_form |       xrefs
	//---------------+--------------------
	// crediti       | 19078850, 19078631
	//
	// [this means you need '~' and not '=' as your syntax]

	// ISSUE: ὑφίϲτημι returns compound forms --> ὑφιϲτάμενοι (36) / παρυφιϲτάμενοι (1) / ϲυνυφιϲτάμενοι (2)
	// BUT: παρυφίϲτημι has a form prevalence of 0...
	// CHOICE: the "clean" version of ὑφίϲτημι OR recognizing the compounds at all

	// SQL: "AND prefixrefs=''" cleans things out...; and that is what was chosen

	psq := fmt.Sprintf(MQT, MFLD, lg, xr)

	foundrows, err := dbconn.Query(context.Background(), psq)
	Msg.EC(err)

	dbmmap := make(map[string]str.DbMorphology)
	var thehit str.DbMorphology

	foreach := []any{&thehit.Observed, &thehit.Xrefs, &thehit.PrefixXrefs, &thehit.RawPossib, &thehit.RelatedHW}
	rfnc := func() error {
		thehit.Observed = strings.ToLower(thehit.Observed)
		dbmmap[thehit.Observed] = thehit
		return nil
	}
	_, e := pgx.ForEachRow(foundrows, foreach, rfnc)
	Msg.EC(e)

	return dbmmap
}
