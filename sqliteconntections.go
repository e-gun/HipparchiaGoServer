package main

// pg_dump hipparchiaDB --table=gr0007 --format plain > x.sql

// https://pkg.go.dev/modernc.org/sqlite

// to fully implement sqlite you would need to rewrite "querybuilder.go"
// the chief problem with that is that sqlite uses "LIKE '%string%'" instead of "~ 'string'"
// the syntax swap is not simple; you need to build a SQLITE extension to recover regexp
// see https://pkg.go.dev/github.com/mattn/go-sqlite3#readme-extensions

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"io"
	"log"
	"os"
	"regexp"
)

func opensqlite() *sql.DB {
	// ultimately need a connection pool?
	// https://turriate.com/articles/making-sqlite-faster-in-go

	// for regex see:
	// https://pkg.go.dev/github.com/mattn/go-sqlite3#section-readme
	regex := func(re, s string) (bool, error) {
		return regexp.MatchString(re, s)
	}
	sql.Register("sqlite3_with_regex",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				return conn.RegisterFunc("regexp", regex, true)
			},
		})

	memdb, err := sql.Open("sqlite3_with_regex", ":memory:")
	if err != nil {
		log.Fatal(err)
	}

	memdb.SetMaxIdleConns(2 * Config.WorkerCount)
	memdb.SetMaxOpenConns(4 * Config.WorkerCount)
	return memdb
}

func sqliteloadactiveauthors() {
	const (
		UPDATE = `author #%d of %d loaded`
		FRQ    = 100
	)

	if SQLProvider == "pgsql" {
		return
	}

	auu := StringMapKeysIntoSlice(AllAuthors)
	for i := 0; i < len(auu); i++ {
		createandloadsqliteauthor(SQLITEConn, auu[i])
		if i%FRQ == 0 {
			msg(fmt.Sprintf(UPDATE, i, len(auu)), MSGFYI)
		}
	}
}

func createandloadsqliteauthor(memdb *sql.DB, au string) {
	const (
		CREATE = `
	CREATE TABLE %s (
	"index" integer,
    wkuniversalid character varying(10),
    level_05_value character varying(64),
    level_04_value character varying(64),
    level_03_value character varying(64),
    level_02_value character varying(64),
    level_01_value character varying(64),
    level_00_value character varying(64),
    marked_up_line text,
    accented_line text,
    stripped_line text,
    hyphenated_words character varying(128),
    annotations character varying(256)
);`
		FAIL1    = `failed to create author table "%s": %s`
		SUCCESS1 = `created author table "%s"`
		EMB      = "emb/db/%s/%s.csv.gz"
		QT       = `insert into %s(wkuniversalid, "index", 
               level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, 
               marked_up_line, accented_line, stripped_line, hyphenated_words, annotations) 
               values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		FAIL5    = `ReadFile failed: %s`
		FAIL2    = `missing header row(?): %s`
		FAIL3    = `insert prepare failed: %s`
		FAIL4    = `[%s] gzip.NewReader failed`
		FAIL6    = `createandloadsqliteauthor() insert failed(%s): %s`
		SUCCESS2 = `loaded author table "%s"`
	)

	// create the author

	checkfail := func(e error) {
		if e != nil {
			msg(fmt.Sprintf(FAIL1, au, e.Error()), MSGWARN)
		} else {
			msg(fmt.Sprintf(SUCCESS1, au), MSGPEEK)
		}
	}

	tx, err := memdb.Begin()

	q := fmt.Sprintf(CREATE, au)
	_, err = tx.Exec(q)
	checkfail(err)

	// load the table
	fail := func(e error, m string) {
		if e != nil {
			msg(fmt.Sprintf(m, e.Error()), MSGWARN)
		}
	}

	dump := fmt.Sprintf(EMB, au[0:2], au)

	compressed, err := os.ReadFile(dump)
	fail(err, FAIL5)
	decompressed, err := gzip.NewReader(bytes.NewReader(compressed))
	fail(err, FAIL4)

	r := csv.NewReader(decompressed)
	_, err = r.Read()
	fail(err, FAIL2)

	q = fmt.Sprintf(QT, au)
	stmt, e := tx.Prepare(q)
	fail(e, FAIL3)

	for {
		record, e3 := r.Read()
		if errors.Is(e3, io.EOF) {
			break
		}
		// fmt.Println(record)
		_, e2 := stmt.Exec(record[0], record[1], record[2], record[3], record[4], record[5], record[6], record[7], record[8], record[9], record[10], record[11], record[12])
		if e2 != nil {
			log.Fatalf("insert failed(%s): %s", record[0], e2)
		}
	}

	q = fmt.Sprintf("SELECT COUNT(*) FROM %s", au)
	row := tx.QueryRow(q)
	var ct int
	err = row.Scan(&ct)
	if err != nil {
		log.Fatalf(e.Error())
	}
	msg(fmt.Sprintf("%s: %d rows", au, ct), 1)

	//tq := fmt.Sprintf(`select wkuniversalid, "index",
	//           level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value,
	//           marked_up_line, accented_line, stripped_line, hyphenated_words, annotations from %s where stripped_line regexp 'est'`, au)

	// txsqlitetestquery(tx, tq, au)

	err = tx.Commit()
	chke(err)

	msg(fmt.Sprintf(SUCCESS2, au), MSGPEEK)
}

//func mdbcreatesqliteauthor(memdb *sql.DB, au string) {
//	const (
//		CREATE = `
//	CREATE TABLE %s (
//	"index" integer,
//    wkuniversalid character varying(10),
//    level_05_value character varying(64),
//    level_04_value character varying(64),
//    level_03_value character varying(64),
//    level_02_value character varying(64),
//    level_01_value character varying(64),
//    level_00_value character varying(64),
//    marked_up_line text,
//    accented_line text,
//    stripped_line text,
//    hyphenated_words character varying(128),
//    annotations character varying(256)
//);`
//		FAIL1   = `failed to create author table "%s": %s`
//		SUCCESS = `created author table "%s"`
//	)
//
//	checkfail := func(e error) {
//		if e != nil {
//			msg(fmt.Sprintf(FAIL1, au, e.Error()), MSGWARN)
//		} else {
//			msg(fmt.Sprintf(SUCCESS, au), MSGPEEK)
//		}
//	}
//
//	q := fmt.Sprintf(CREATE, au)
//	_, err := memdb.Exec(q)
//	checkfail(err)
//
//}
//
//func memdbloadaqliteauthor(memdb *sql.DB, au string) {
//	const (
//		EMB = "emb/db/%s/%s.csv.bz2"
//		QT  = `insert into %s(wkuniversalid, "index",
//               level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value,
//               marked_up_line, accented_line, stripped_line, hyphenated_words, annotations)
//               values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
//		FAIL1   = `ReadFile failed: %s`
//		FAIL2   = `missing header row(?): %s`
//		FAIL3   = `insert prepare failed: %s`
//		FAIL4   = `[%s] gzip.NewReader failed`
//		SUCCESS = `loaded author table "%s"`
//		BT      = "BEGIN TRANSACTION"
//		COMM    = "COMMIT"
//	)
//
//	fail := func(e error, m string) {
//		if e != nil {
//			msg(fmt.Sprintf(m, e.Error()), MSGWARN)
//		}
//	}
//
//	dump := fmt.Sprintf(EMB, au[0:2], au)
//
//	compressed, err := os.ReadFile(dump)
//	fail(err, FAIL1)
//	decompressed := bzip2.NewReader(bytes.NewReader(compressed))
//
//	r := csv.NewReader(decompressed)
//	_, err = r.Read()
//	fail(err, FAIL2)
//
//	q := fmt.Sprintf(QT, au)
//	stmt, e := memdb.Prepare(q)
//	fail(e, FAIL3)
//
//	for {
//		record, e3 := r.Read()
//		if errors.Is(e3, io.EOF) {
//			break
//		}
//		_, e2 := stmt.Exec(record[0], record[1], record[2], record[3], record[4], record[5], record[6], record[7], record[8], record[9], record[10], record[11], record[12])
//		if e2 != nil {
//			log.Fatalf("insert failed(%s): %s", record[0], e2)
//		}
//	}
//
//	msg(fmt.Sprintf(SUCCESS, au), MSGPEEK)
//}
//
//func createsqliteauthor(au string) {
//	const (
//		CREATE = `
//	CREATE TABLE %s (
//	"index" integer,
//    wkuniversalid character varying(10),
//    level_05_value character varying(64),
//    level_04_value character varying(64),
//    level_03_value character varying(64),
//    level_02_value character varying(64),
//    level_01_value character varying(64),
//    level_00_value character varying(64),
//    marked_up_line text,
//    accented_line text,
//    stripped_line text,
//    hyphenated_words character varying(128),
//    annotations character varying(256)
//);`
//		FAIL1   = `failed to create author table "%s": %s`
//		SUCCESS = `created author table "%s"`
//	)
//
//	conn := GetSQLiteConn()
//	defer conn.Close()
//
//	checkfail := func(e error) {
//		if e != nil {
//			msg(fmt.Sprintf(FAIL1, au, e.Error()), MSGWARN)
//		} else {
//			msg(fmt.Sprintf(SUCCESS, au), MSGPEEK)
//		}
//	}
//
//	q := fmt.Sprintf(CREATE, au)
//	_, err := conn.ExecContext(context.Background(), q)
//	checkfail(err)
//}
//
//func loadaqliteauthor(au string) {
//	const (
//		EMB = "emb/db/%s/%s.csv.gz"
//		QT  = `insert into %s(wkuniversalid, "index",
//               level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value,
//               marked_up_line, accented_line, stripped_line, hyphenated_words, annotations)
//               values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
//		FAIL1   = `[%s] ReadFile failed: %s`
//		FAIL2   = `[%s] missing header row(?): %s`
//		FAIL3   = `[%s] insert prepare failed: %s`
//		FAIL4   = `[%s] gzip.NewReader failed`
//		SUCCESS = `loaded author table "%s"`
//		TQ      = `SELECT accented_line from %s WHERE "index" = 1`
//	)
//
//	conn := GetSQLiteConn()
//	defer conn.Close()
//
//	fail := func(e error, m string) {
//		if e != nil {
//			msg(fmt.Sprintf(m, au, e.Error()), MSGWARN)
//		}
//	}
//
//	dump := fmt.Sprintf(EMB, au[0:2], au)
//
//	compressed, err := os.ReadFile(dump)
//	fail(err, FAIL1)
//	decompressed, err := gzip.NewReader(bytes.NewReader(compressed))
//	fail(err, FAIL4)
//
//	r := csv.NewReader(decompressed)
//	_, err = r.Read()
//	fail(err, FAIL2)
//
//	q := fmt.Sprintf(QT, au)
//
//	stmt, e := conn.PrepareContext(context.Background(), q)
//	fail(e, FAIL3)
//
//	for {
//		record, er := r.Read()
//		if errors.Is(er, io.EOF) {
//			break
//		}
//
//		// q := fmt.Sprintf(QT, au)
//		//stmt, e := memdb.Prepare(q)
//		//fail(e, FAIL3)
//		//
//		//_, e = stmt.Exec(record[0], record[1], record[2], record[3], record[4], record[5], record[6], record[7], record[8], record[9], record[10], record[11], record[12])
//
//		_, e2 := stmt.Exec(record[0], record[1], record[2], record[3], record[4], record[5], record[6], record[7], record[8], record[9], record[10], record[11], record[12])
//		if e2 != nil {
//			log.Fatalf("insert failed(%s): %s", record[0], e2)
//		}
//	}
//
//	wl := GrabOneLine(au, 1)
//	msg(wl.MarkedUp, 5)
//
//	msg(fmt.Sprintf(SUCCESS, au), MSGPEEK)
//}

//func sqlitetestquery(memdb *sql.DB, tq string) {
//	rows, err := memdb.Query(tq)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println("TQ-ed")
//
//	defer rows.Close()
//	var rr []DbWorkline
//	for rows.Next() {
//		var rw DbWorkline
//		if e := rows.Scan(&rw.TbIndex, &rw.WkUID, &rw.Lvl5Value, &rw.Lvl4Value, &rw.Lvl3Value, &rw.Lvl2Value, &rw.Lvl1Value, &rw.Lvl0Value, &rw.MarkedUp, &rw.Accented, &rw.Stripped, &rw.Hyphenated, &rw.Annotations); err != nil {
//			log.Fatal(e)
//		}
//		rr = append(rr, rw)
//	}
//	msg("sqlitetestquery() results:", 1)
//	for i := 0; i < len(rr); i++ {
//		fmt.Println(rr[i].BuildHyperlink() + ": " + rr[i].MarkedUp)
//	}
//}

func txsqlitetestquery(tx *sql.Tx, tq string) {
	rows, err := tx.Query(tq)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("TQ-ed")

	defer rows.Close()
	var rr []DbWorkline
	for rows.Next() {
		var rw DbWorkline
		if e := rows.Scan(&rw.TbIndex, &rw.WkUID, &rw.Lvl5Value, &rw.Lvl4Value, &rw.Lvl3Value, &rw.Lvl2Value, &rw.Lvl1Value, &rw.Lvl0Value, &rw.MarkedUp, &rw.Accented, &rw.Stripped, &rw.Hyphenated, &rw.Annotations); err != nil {
			log.Fatal(e)
		}
		rr = append(rr, rw)
	}

	for i := 0; i < len(rr); i++ {
		fmt.Println(rr[i].BuildHyperlink() + ": " + rr[i].Stripped)
	}
}

func GetSQLiteConn() *sql.Conn {
	conn, e := SQLITEConn.Conn(context.Background())
	chke(e)
	return conn
}

func postinitializationsqlitetest() {
	au := "lt0016"
	tq := fmt.Sprintf(`select wkuniversalid, "index", 
               level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, 
               marked_up_line, accented_line, stripped_line, hyphenated_words, annotations from %s where stripped_line regexp 'est'`, au)

	tx, err := SQLITEConn.Begin()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("new tx...")
	txsqlitetestquery(tx, tq)
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}
