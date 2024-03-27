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

// DictEntryGrabber - search postgres tables and return []DbLexicon
func DictEntryGrabber(seeking string, dict string, col string, syntax string) []str.DbLexicon {
	const (
		FLDS = `entry_name, metrical_entry, id_number, pos, translations, html_body`
		PSQQ = `SELECT %s FROM %s_dictionary WHERE %s %s '%s' ORDER BY id_number ASC LIMIT %d`
	)

	// note that "html_body" is only available via HipparchiaBuilder 1.6.0+
	q := fmt.Sprintf(PSQQ, FLDS, dict, col, syntax, seeking, vv.MAXDICTLOOKUP)

	var lexicalfinds []str.DbLexicon
	var thehit str.DbLexicon
	dedup := make(map[float32]bool)

	foreach := []any{&thehit.Word, &thehit.Metrical, &thehit.ID, &thehit.POS, &thehit.Transl, &thehit.Entry}
	rwfnc := func() error {
		thehit.SetLang(dict)
		if _, dup := dedup[thehit.ID]; !dup {
			// use ID and not Lex because καρπόϲ.53442 is not καρπόϲ.53443
			dedup[thehit.ID] = true
			lexicalfinds = append(lexicalfinds, thehit)
		}
		return nil
	}

	foundrows, err := SQLPool.Query(context.Background(), q)
	Msg.EC(err)

	_, e := pgx.ForEachRow(foundrows, foreach, rwfnc)
	Msg.EC(e)

	return lexicalfinds
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

// ArrayToGetHeadwordCounts - get the int counts for a slice of headwords
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

// MorphPossibIntoLexPossib - []MorphPossib into []DbLexicon
func MorphPossibIntoLexPossib(d string, mpp []str.MorphPossib) []str.DbLexicon {
	const (
		FLDS = `entry_name, metrical_entry, id_number, pos, translations, html_body`
		PSQQ = `SELECT %s FROM %s_dictionary WHERE %s ~* '^%s(|¹|²|³|⁴|1|2)$' ORDER BY id_number ASC`
		COLM = "entry_name"
	)
	var hwm []string
	for _, p := range mpp {
		if strings.TrimSpace(p.Headwd) != "" {
			hwm = append(hwm, p.Headwd)
		}
	}

	// the next is primed to produce problems: see καρποῦ which will turn καρπόϲ1 and καρπόϲ2 into just καρπόϲ; need xref_value?
	// but we have probably taken care of this below: see the comments
	hwm = gen.Unique(hwm)

	// [d] get the wordobjects for each Unique headword: probedictionary()

	// note that "html_body" is only available via HipparchiaBuilder 1.6.0+

	var lexicalfinds []str.DbLexicon
	var thehit str.DbLexicon
	dedup := make(map[float32]bool)

	foreach := []any{&thehit.Word, &thehit.Metrical, &thehit.ID, &thehit.POS, &thehit.Transl, &thehit.Entry}

	rwfnc := func() error {
		thehit.SetLang(d)
		if _, dup := dedup[thehit.ID]; !dup {
			// use ID and not Lex because καρπόϲ.53442 is not καρπόϲ.53443
			dedup[thehit.ID] = true
			lexicalfinds = append(lexicalfinds, thehit)
		}
		return nil
	}

	for _, w := range hwm {
		q := fmt.Sprintf(PSQQ, FLDS, d, COLM, w)
		foundrows, err := SQLPool.Query(context.Background(), q)
		Msg.EC(err)

		_, e := pgx.ForEachRow(foundrows, foreach, rwfnc)
		Msg.EC(e)

	}
	return lexicalfinds
}

// FindProximateEntry - what is the name and id of the entry next to this entry?
func FindProximateEntry(w str.DbLexicon, nxt string) str.DbLexicon {
	const (
		PROXENTRYQUERY = `SELECT entry_name, id_number from %s_dictionary WHERE id_number %s %.0f ORDER BY id_number %s LIMIT 1`
		NOTH           = `formatlexicaloutput() found no entry %s '%s'`
	)

	oper := ">"
	ord := "ASC"
	em := "after"

	if nxt != "next" {
		oper = "<"
		ord = "DESC"
		em = "before"
	}

	var prx str.DbLexicon
	p := SQLPool.QueryRow(context.Background(), fmt.Sprintf(PROXENTRYQUERY, w.GetLang(), oper, w.ID, ord))
	e := p.Scan(&prx.Entry, &prx.ID)
	if e != nil {
		Msg.FYI(fmt.Sprintf(NOTH, em, w.Entry))
	}

	return prx
}
