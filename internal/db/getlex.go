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

// hipparchiaDB=# \d wordcounts_a
//                     Table "public.wordcounts_a"
//   Column    |         Type          | Collation | Nullable | Default
//-------------+-----------------------+-----------+----------+---------
// entry_name  | character varying(64) |           |          |
// total_count | integer               |           |          | 0
// gr_count    | integer               |           |          | 0
// lt_count    | integer               |           |          | 0
// dp_count    | integer               |           |          | 0
// in_count    | integer               |           |          | 0
// ch_count    | integer               |           |          | 0
//Indexes:
//    "wcindex_a" UNIQUE, btree (entry_name)

// hipparchiaDB=# \d dictionary_headword_wordcounts
//                   Table "public.dictionary_headword_wordcounts"
//          Column          |         Type          | Collation | Nullable | Default
//--------------------------+-----------------------+-----------+----------+---------
// entry_name               | character varying(64) |           |          |
// total_count              | integer               |           |          | 0
// gr_count                 | integer               |           |          | 0
// lt_count                 | integer               |           |          | 0
// dp_count                 | integer               |           |          | 0
// in_count                 | integer               |           |          | 0
// ch_count                 | integer               |           |          | 0
// frequency_classification | character varying(64) |           |          |
// early_occurrences        | integer               |           |          | 0
// middle_occurrences       | integer               |           |          | 0
// late_occurrences         | integer               |           |          | 0
// acta                     | integer               |           |          | 0
// agric                    | integer               |           |          | 0
// ...

// see CALCULATEWORDWEIGHTS in HipparchiaServer's startup.py on where these really come from
// alternate chars: "🄶", "🄻", "🄸", "🄳", "🄲"; but these align awkwardly on the page

func GetHeadwordWordCount(word string) str.DbHeadwordCount {
	// scan a headwordcount into the corresponding struct
	// note that if you reassign a genre, this is one of the place you have to edit
	const (
		QTP = `
		SELECT
			entry_name , total_count, gr_count, lt_count, dp_count, in_count, ch_count,
			frequency_classification, early_occurrences, middle_occurrences ,late_occurrences,
			acta, agric, alchem, anthol, apocalyp, apocryph, apol, astrol, astron, biogr, bucol,
			caten, chronogr, comic, comm, concil, coq, dialog, docu, doxogr, eccl, eleg, encom, epic,
			epigr, epist, evangel, exeget, fab, geogr, gnom, gramm, hagiogr, hexametr, hist, homilet,
			hymn, hypoth, iamb, ignotum, invectiv, inscr, jurisprud, lexicogr, liturg, lyr, magica, 
			math, mech, med, metrolog, mim, mus, myth, narrfict, nathist, onir, orac, orat,
			paradox, parod, paroem, perieg, phil, physiognom, poem, polyhist, prophet, pseudepigr, rhet,
			satura, satyr, schol, tact, test, theol, trag
		FROM dictionary_headword_wordcounts WHERE entry_name='%s'`

		FAIL = "headwordlookup() returned 'nil' when looking for '%s'"
		INFO = "headwordlookup() for '%s' returned %d finds"
	)

	dbconn := GetDBConnection()
	defer dbconn.Release()

	q := fmt.Sprintf(QTP, word)

	foundrows, err := dbconn.Query(context.Background(), q)
	Msg.EC(err)

	var thesefinds []str.DbHeadwordCount
	var co str.DbHeadwordCorpusCounts
	var chr str.DbHeadwordTimeCounts
	var g str.DbHeadwordGenreCounts
	var th str.DbHeadwordTheologyCounts
	var rh str.DbHeadwordRhetoricaCounts

	defer foundrows.Close()
	for foundrows.Next() {
		var thehit str.DbHeadwordCount
		e := foundrows.Scan(&thehit.Entry, &thehit.Total, &co.TGrk, &co.TLat, &co.TDP, &co.TIN, &co.TCh,
			&thehit.FrqCla, &chr.Early, &chr.Middle, &chr.Late,
			&th.Acta, &g.Agric, &g.Alchem, &g.Anthol, &th.Apocal, &th.Apocr, &th.Apol, &g.Astrol, &g.Astron, &g.Biogr, &g.Bucol,
			&th.Caten, &g.Chron, &g.Comic, &g.Comm, &g.Concil, &g.Coq, &g.Dial, &g.Docu, &g.Doxog, &th.Eccl, &g.Eleg, &rh.Encom, &g.Epic,
			&g.Epigr, &g.Epist, &th.Evang, &g.Exeg, &g.Fab, &g.Geog, &g.Gnom, &g.Gram, &th.Hagiog, &g.Hexam, &g.Hist, &th.Homil,
			&g.Hymn, &g.Hypoth, &g.Iamb, &g.Ignot, &rh.Invect, &g.Inscr, &g.Juris, &g.Lexic, &th.Litur, &g.Lyr, &g.Magica,
			&g.Math, &g.Mech, &g.Med, &g.Meteor, &g.Mim, &g.Mus, &g.Myth, &g.NarrFic, &g.NatHis, &g.Onir, &g.Orac, &rh.Orat,
			&g.Paradox, &g.Parod, &g.Paroem, &g.Perig, &g.Phil, &g.Physiog, &g.Poem, &g.Polyhist, &th.Proph, &g.Pseud, &rh.Rhet,
			&g.Satura, &g.Satyr, &g.Schol, &g.Tact, &g.Test, &th.Theol, &g.Trag)
		if e != nil {
			Msg.TMI(fmt.Sprintf(FAIL, word))
		}
		thehit.Corpus = co
		thehit.Chron = chr
		thehit.Genre = g
		thehit.Rhetorica = rh
		thehit.Theology = th
		thesefinds = append(thesefinds, thehit)
	}

	var thefind str.DbHeadwordCount
	if len(thesefinds) == 1 {
		thefind = thesefinds[0]
	} else {
		Msg.TMI(fmt.Sprintf(INFO, word, len(thesefinds)))
	}

	thefind.LoadCorpVals()
	thefind.LoadTimeVals()
	thefind.LoadGenreVals()

	return thefind
}

// ArrayToGetScansion - grab all scansions for a slice of words and return as a map
func ArrayToGetScansion(wordlist []string) map[string]string {
	const (
		TT = `CREATE TEMPORARY TABLE ttw_%s AS SELECT words AS w FROM unnest(ARRAY[%s]) words`
		QT = `SELECT entry_name, metrical_entry FROM %s_dictionary WHERE EXISTS 
				(SELECT 1 FROM ttw_%s temptable WHERE temptable.w = %s_dictionary.entry_name)`
	)

	type entryandmeter struct {
		Entry string
		Meter string
	}

	// look for the upper case matches too: Ϲωκράτηϲ and not just ϲωκρατέω (!)
	uppers := make([]string, len(wordlist))
	for i := 0; i < len(wordlist); i++ {
		uppers[i] = strings.Title(wordlist[i])
	}

	wordlist = append(wordlist, uppers...)

	dbconn := GetDBConnection()
	defer dbconn.Release()

	foundmetrics := make(map[string]string)
	var thehit entryandmeter

	foreach := []any{&thehit.Entry, &thehit.Meter}

	rwfnc := func() error {
		foundmetrics[thehit.Entry] = thehit.Meter
		return nil
	}

	// a waste of time to check the language on every word; just flail/fail once
	for _, uselang := range vv.TheLanguages {
		u := strings.Replace(uuid.New().String(), "-", "", -1)
		id := fmt.Sprintf("%s_%s_mw", u, uselang)
		a := fmt.Sprintf("'%s'", strings.Join(wordlist, "', '"))
		t := fmt.Sprintf(TT, id, a)

		_, err := dbconn.Exec(context.Background(), t)
		Msg.EC(err)

		foundrows, e := dbconn.Query(context.Background(), fmt.Sprintf(QT, uselang, id, uselang))
		Msg.EC(e)

		_, ee := pgx.ForEachRow(foundrows, foreach, rwfnc)
		Msg.EC(ee)
	}
	return foundmetrics
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

func ArrayToGetHeadwordCounts(wordlist []string) map[string]int {
	const (
		TT = `CREATE TEMPORARY TABLE ttw_%s AS SELECT words AS w FROM unnest(ARRAY[%s]) words`
		QT = `SELECT entry_name , total_count FROM dictionary_headword_wordcounts WHERE EXISTS 
				(SELECT 1 FROM ttw_%s temptable WHERE temptable.w = dictionary_headword_wordcounts.entry_name)`
	)

	dbconn := GetDBConnection()
	defer dbconn.Release()

	countmap := make(map[string]int)

	type tempstruct struct {
		w string
		c int
	}

	var thehit tempstruct

	foreach := []any{&thehit.w, &thehit.c}

	rwfnc := func() error {
		countmap[thehit.w] = thehit.c
		return nil
	}

	u := strings.Replace(uuid.New().String(), "-", "", -1)
	a := fmt.Sprintf("'%s'", strings.Join(wordlist, "', '"))

	t := fmt.Sprintf(TT, u, a)
	_, err := dbconn.Exec(context.Background(), t)
	Msg.EC(err)

	foundrows, e := dbconn.Query(context.Background(), fmt.Sprintf(QT, u))
	Msg.EC(e)

	_, ee := pgx.ForEachRow(foundrows, foreach, rwfnc)
	Msg.EC(ee)

	return countmap
}

// FetchHeadwordCounts - map a list of headwords to their corpus counts
func FetchHeadwordCounts(headwordset map[string]bool) map[string]int {
	const (
		MSG1 = "FetchHeadwordCounts() will search for %d headwords"
	)
	if len(headwordset) == 0 {
		return make(map[string]int)
	}

	tt := "CREATE TEMPORARY TABLE ttw_%s AS SELECT words AS w FROM unnest(ARRAY[%s]) words"
	qt := "SELECT entry_name, total_count FROM dictionary_headword_wordcounts WHERE EXISTS " +
		"(SELECT 1 FROM ttw_%s temptable WHERE temptable.w = dictionary_headword_wordcounts.entry_name)"

	rndid := strings.Replace(uuid.New().String(), "-", "", -1)

	hw := make([]string, 0, len(headwordset))
	for h := range headwordset {
		hw = append(hw, h)
	}

	Msg.PEEK(fmt.Sprintf(MSG1, len(headwordset)))

	dbconn := GetDBConnection()
	defer dbconn.Release()

	arr := strings.Join(hw, "', '")
	arr = fmt.Sprintf("'%s'", arr)

	tt = fmt.Sprintf(tt, rndid, arr)
	_, err := dbconn.Exec(context.Background(), tt)
	Msg.EC(err)

	qt = fmt.Sprintf(qt, rndid)
	foundrows, e := dbconn.Query(context.Background(), qt)
	Msg.EC(e)

	returnmap := make(map[string]int)
	defer foundrows.Close()
	for foundrows.Next() {
		var thehit str.WeightedHeadword
		err = foundrows.Scan(&thehit.Word, &thehit.Count)
		Msg.EC(err)
		returnmap[thehit.Word] = thehit.Count
	}

	// don't kill off unfound terms
	for i := range hw {
		if _, t := returnmap[hw[i]]; t {
			continue
		} else {
			returnmap[hw[i]] = 0
		}
	}

	// "returnmap" for Albinus , poet. [lt2002]
	// map[abscondo:213 apte:168 aptus:1423 capitolium:0 celsus¹:1050 concludo:353 dactylus:167 de:42695 deus:14899 eo¹:58129 fio:12305 fretum:746 fretus¹:761 ille:44214 jungo:2275 liber¹:7550 liber⁴:13403 libo¹:3996 metrum:383 moenia¹:1308 non:96475 nullus:11785 pateo:1828 patesco:46 possum:41631 quis²:0 quis¹:52619 qui²:19812 qui¹:251744 re-pono:47 res:38669 romanus:0 sed:44131 sinus¹:1223 spondeum:158 spondeus:205 sponte:841 terni:591 totus²:0 totus¹:9166 triumphus:1058 tueor:3734 urbs:8564 verro:3843 versum:435 versus³:3390 verto:1471 †uilem:0]

	return returnmap
}
