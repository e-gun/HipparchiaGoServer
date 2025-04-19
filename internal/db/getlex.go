//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package db

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"strings"
)

// DictEntryGrabber - search postgres tables and return []DbLexicon
func DictEntryGrabber(seeking string, dict string, col string, syntax string) []str.DbLexicon {
	const (
		FLDS = `idval, entry_name, entry_metr, idname, entry_type, translations, usedby, prelim_info, sense_ids, senses`
		PSQQ = `SELECT %s FROM %s_dictionary WHERE %s %s '%s' ORDER BY idval ASC LIMIT %d`
	)

	dbconn := getdbconnection()
	defer dbconn.Release()

	q := fmt.Sprintf(PSQQ, FLDS, dict, col, syntax, seeking, vv.MAXDICTLOOKUP)

	var lexicalfinds []str.DbLexicon
	var thehit str.DbLexicon
	var jsonsenses []byte
	var unsplitids string
	dedup := make(map[string]bool)

	foreach := []any{
		&thehit.IdFloat,
		&thehit.EntryName,
		&thehit.EntryMetr,
		&thehit.IdString,
		&thehit.EntryType,
		&thehit.Transl,
		&thehit.Usedby,
		&thehit.PrelimInfo,
		&unsplitids,
		&jsonsenses,
	}

	rwfnc := func() error {
		thehit.SetLang(dict)
		sensemap := make(map[int]str.LexicalSenses)
		err := json.Unmarshal(jsonsenses, &sensemap)
		if err != nil {
			fmt.Println("Error decoding JSON:", err)
		}

		// need to store these in order
		thehit.Senses = make([]str.LexicalSenses, len(sensemap))
		keys := gen.IntMapKeysIntoSlice(sensemap) // comes back sorted: 0, 1, 2, ...
		for i, k := range keys {
			thehit.Senses[i] = sensemap[k]
		}

		thehit.SenseIDs = strings.Split(unsplitids, " ")

		if _, dup := dedup[thehit.IdString]; !dup {
			// use ID and not Lex because καρπόϲ.53442 is not καρπόϲ.53443
			dedup[thehit.IdString] = true
			lexicalfinds = append(lexicalfinds, thehit)
		}
		return nil
	}

	foundrows, err := dbconn.Query(context.Background(), q)
	Msg.EC(err)

	_, e := pgx.ForEachRow(foundrows, foreach, rwfnc)
	Msg.EC(e)

	return lexicalfinds
}

// ArrayToGetScansion - grab all scansions for a slice of words and return as a map
func ArrayToGetScansion(wordlist []string) map[string]string {
	const (
		TT = `CREATE TEMPORARY TABLE ttw_%s AS SELECT words AS w FROM unnest(ARRAY[%s]) words`
		QT = `SELECT entry_name, entry_metr FROM %s_dictionary WHERE EXISTS 
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

	dbconn := getdbconnection()
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

// ArrayToGetHeadwordCounts - get the int counts for a slice of headwords
func ArrayToGetHeadwordCounts(wordlist []string) map[string]int {
	const (
		TT = `CREATE TEMPORARY TABLE ttw_%s AS SELECT words AS w FROM unnest(ARRAY[%s]) words`
		QT = `SELECT entry_name , total_count FROM dictionary_headword_wordcounts WHERE EXISTS 
				(SELECT 1 FROM ttw_%s temptable WHERE temptable.w = dictionary_headword_wordcounts.entry_name)`
	)

	dbconn := getdbconnection()
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

// MorphPossibIntoLexPossib - []MorphPossib into []DbLexicon
func MorphPossibIntoLexPossib(d string, mpp []str.MorphPossib) []str.DbLexicon {
	const (
		FLDS = `idval, entry_name, entry_metr, idname, entry_type, translations, usedby, prelim_info, sense_ids, senses`
		PSQQ = `SELECT %s FROM %s_dictionary WHERE %s ~* '^%s(|¹|²|³|⁴|1|2)$' ORDER BY idval ASC`
		COLM = "entry_name"
		SQLE = "MorphPossibIntoLexPossib() failed on query: \n\t%s"
	)

	var hwm []string
	for _, p := range mpp {
		if strings.TrimSpace(p.Headwd) != "" {
			hwm = append(hwm, p.Headwd)
		}
	}

	dbconn := getdbconnection()
	defer dbconn.Release()

	// the next is primed to produce problems: see καρποῦ which will turn καρπόϲ1 and καρπόϲ2 into just καρπόϲ; need xref_value?
	// but we have probably taken care of this below: see the comments

	// todo: HGB has fixed this sort of thing?

	hwm = gen.Unique(hwm)

	// [d] get the wordobjects for each Unique headword: probedictionary()

	// note that "html_body" is only available via HipparchiaBuilder 1.6.0+

	var lexicalfinds []str.DbLexicon
	var thehit str.DbLexicon
	var jsonsenses []byte
	var unsplitids string
	dedup := make(map[string]bool)

	foreach := []any{
		&thehit.IdFloat,
		&thehit.EntryName,
		&thehit.EntryMetr,
		&thehit.IdString,
		&thehit.EntryType,
		&thehit.Transl,
		&thehit.Usedby,
		&thehit.PrelimInfo,
		&unsplitids,
		&jsonsenses}

	rwfnc := func() error {
		sensemap := make(map[int]str.LexicalSenses)
		err := json.Unmarshal(jsonsenses, &sensemap)
		if err != nil {
			fmt.Println("Error decoding JSON:", err)
		}

		// need to store these in order
		thehit.Senses = make([]str.LexicalSenses, len(sensemap))
		for i, sense := range sensemap {
			thehit.Senses[i] = sense
		}

		thehit.SenseIDs = strings.Split(unsplitids, " ")

		thehit.SetLang(d)
		if _, dup := dedup[thehit.IdString]; !dup {
			// use ID and not Lex because καρπόϲ.53442 is not καρπόϲ.53443
			dedup[thehit.IdString] = true
			lexicalfinds = append(lexicalfinds, thehit)
		}
		return nil
	}

	numberstripper := strings.NewReplacer("-", "", "¹", "", "²", "", "³", "")

	for _, w := range hwm {
		w = numberstripper.Replace(w)
		q := fmt.Sprintf(PSQQ, FLDS, d, COLM, w)
		foundrows, err := dbconn.Query(context.Background(), q)
		Msg.EC(err)

		// fmt.Println("MorphPossibIntoLexPossib()", q)
		// is this no longer true as of HGB?

		// nb: there is some wonky data in the morph possibilities because of some corner cases not caught by the builder
		// [HGS-DBI] SELECT entry_name, metrical_entry, idstring, pos, translations, html_body FROM greek_dictionary WHERE entry_name ~* '^ὀμβρόω(|¹|²|³|⁴|1|2)$' ORDER BY idval ASC
		// [HGS-DBI] SELECT entry_name, metrical_entry, idstring, pos, translations, html_body FROM greek_dictionary WHERE entry_name ~* '^ό)μβροϲ(|¹|²|³|⁴|1|2)$' ORDER BY idval ASC
		// the second has a ')' that yields an error
		// "ERROR: invalid regular expression: parentheses () not balanced (SQLSTATE 2201B)"

		_, e := pgx.ForEachRow(foundrows, foreach, rwfnc)
		if e != nil {
			// you can survive this error... log i
			// number of field descriptions must equal number of destinations, got 9 and 8
			fmt.Println(e)
			Msg.FYI(fmt.Sprintf(SQLE, q))

		}
	}
	return lexicalfinds
}

// FindProximateEntry - what is the name and id of the entry next to this entry? not obvious because n30004, n30004a, n30005
func FindProximateEntry(w str.DbLexicon, nxt string) str.DbLexicon {
	const (
		PROXENTRYQUERY = `SELECT entry_name, idval from %s_dictionary WHERE idval %s $1 ORDER BY idval %s LIMIT 1`
		NOTH           = `FindProximateEntry() found no entry %s '%f'`
	)

	dbconn := getdbconnection()
	defer dbconn.Release()

	oper := ">"
	ord := "ASC"
	em := "after"

	if nxt != "next" {
		oper = "<"
		ord = "DESC"
		em = "before"
	}

	var prx str.DbLexicon
	q := fmt.Sprintf(PROXENTRYQUERY, w.GetLang(), oper, ord)

	p := dbconn.QueryRow(context.Background(), q, w.IdFloat)
	e := p.Scan(&prx.EntryName, &prx.IdString)
	if e != nil {
		Msg.FYI(fmt.Sprintf(NOTH, em, w.IdFloat))
	}

	return prx
}
