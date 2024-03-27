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

// GetMultipleWordCounts - return total word count figures for each word in a slice of words
func GetMultipleWordCounts(ww []string) map[string]str.DbWordCount {
	const (
		TTT  = `CREATE TEMPORARY TABLE ttw_%s AS SELECT values AS wordforms FROM unnest(ARRAY[%s]) values`
		WCQT = `SELECT entry_name, total_count FROM wordcounts_%s WHERE EXISTS 
		(SELECT 1 FROM ttw_%s temptable WHERE temptable.wordforms = wordcounts_%s.entry_name)`
		CHARR = `abcdefghijklmnopqrstuvwxyzαβψδεφγηιξκλμνοπρτυωχθζϲ`
	)

	dbconn := GetDBConnection()
	defer dbconn.Release()

	byfirstlett := make(map[string][]string)

	// [c2] query the database

	// pgsql single quote escape: quote followed by a single quote to be escaped: κρυφθεῖϲ''
	// but they will in fact be stored less the apostrophe...

	for _, w := range ww {
		init := gen.StripaccentsRUNE([]rune(w))
		if len(init) == 0 {
			continue
		}
		i := string(init[0])
		if strings.Contains(CHARR, i) {
			byfirstlett[i] = append(byfirstlett[i], strings.Replace(w, "'", "", -1))
		} else {
			byfirstlett["0"] = append(byfirstlett["0"], strings.Replace(w, "'", "", -1))
		}
	}

	// arr := fmt.Sprintf("'%s'", strings.Join(esc, "', '"))

	// hipparchiaDB=# CREATE TEMPORARY TABLE ttw AS
	//    SELECT values AS wordforms FROM
	//      unnest(ARRAY['κόραϲ', 'κόραι', 'κῶραι', 'κούρῃϲιν', 'κούραϲ', 'κούραιϲιν', 'κόραν', 'κώρα', 'κόραιϲιν', 'κόραιϲι', 'κόρα', 'κόρᾳϲ'])
	//    values;
	//
	//SELECT entry_name, total_count FROM wordcounts_κ WHERE EXISTS (
	//  (SELECT 1 FROM ttw temptable WHERE temptable.wordforms = wordcounts_κ.entry_name )
	//);
	//SELECT 12
	// entry_name | total_count
	//------------+-------------
	// κόραν      |          59
	// κούραιϲιν  |           1
	// κῶραι      |           4
	// κόρᾳϲ      |           1
	// κούρῃϲιν   |           9
	// κόραι      |         363
	// κόραϲ      |         668
	// κόραιϲιν   |           2
	// κόραιϲι    |           8
	// κούραϲ     |          89
	// κόρα       |          72
	// κώρα       |           9
	//(12 rows)

	wcc := make(map[string]str.DbWordCount)
	var wc str.DbWordCount

	each := []any{&wc.Word, &wc.Total}

	rfnc := func() error {
		wcc[wc.Word] = wc
		return nil
	}

	// this bit could be parallelized...
	for l := range byfirstlett {
		arr := fmt.Sprintf("'%s'", strings.Join(byfirstlett[l], "', '"))
		rnd := strings.Replace(uuid.New().String(), "-", "", -1)
		_, ee := dbconn.Exec(context.Background(), fmt.Sprintf(TTT, rnd, arr))
		Msg.EC(ee)

		q := fmt.Sprintf(WCQT, l, rnd, l)
		rr, ee := dbconn.Query(context.Background(), q)
		Msg.EC(ee)

		// you just found »ἥρμοττ« which gives you »ἥρμοττ'«: see below for where this becomes an issue
		_, er := pgx.ForEachRow(rr, each, rfnc)
		Msg.EC(er)
	}

	return wcc
}

// GetIndividualWordCount - return total word count figures for one word
func GetIndividualWordCount(wd string) str.DbWordCount {
	const (
		FLDS = `entry_name, total_count, gr_count, lt_count, dp_count, in_count, ch_count`
		PSQQ = `SELECT %s FROM wordcounts_%s where entry_name = '%s'`
		NOTH = `findbyform() found no results for '%s'`
	)
	// golang hates indexing unicode strings: strings are bytes, and unicode chars take more than one byte
	c := []rune(wd)
	q := fmt.Sprintf(PSQQ, FLDS, gen.StripaccentsSTR(string(c[0])), wd)

	var wc str.DbWordCount
	ct := SQLPool.QueryRow(context.Background(), q)
	e := ct.Scan(&wc.Word, &wc.Total, &wc.Gr, &wc.Lt, &wc.Dp, &wc.In, &wc.Ch)
	if e != nil {
		Msg.FYI(fmt.Sprintf(NOTH, wd))
	}
	return wc
}

// GetIndividualHeadwordCount - get a DbHeadwordCount for a single word
func GetIndividualHeadwordCount(word string) str.DbHeadwordCount {
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

// MapHeadwordCounts - map a list of headwords to their corpus counts
func MapHeadwordCounts(headwordset map[string]bool) map[string]int {
	const (
		MSG1 = "MapHeadwordCounts() will search for %d headwords"
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
