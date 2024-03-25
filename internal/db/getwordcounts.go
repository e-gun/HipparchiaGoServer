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

// GetWordCounts - return total word count figures for each word in a slice of words
func GetWordCounts(ww []string) map[string]str.DbWordCount {
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
