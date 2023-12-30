package main

// pg_dump hipparchiaDB --table=gr0007 --format plain > x.sql

// https://pkg.go.dev/modernc.org/sqlite

// to fully implement sqlite you need to rewrite "querybuilder.go"
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

	// "file::memory:?cache=shared" because next will close soon after first uses: sql.Open("sqlite3_with_regex", ":memory:")

	memdb, err := sql.Open("sqlite3_with_regex", "file::memory:?cache=shared")
	if err != nil {
		log.Fatal(err)
	}

	//memdb.SetConnMaxIdleTime(300 * time.Second)
	//memdb.SetConnMaxLifetime(300000 * time.Second)

	// breaks if you uncomment the next
	//memdb.SetMaxOpenConns(Config.WorkerCount * 4)
	//memdb.SetMaxIdleConns(Config.WorkerCount * 4)

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
		createandloadsqliteauthor(auu[i])
		if i%FRQ == 0 {
			msg(fmt.Sprintf(UPDATE, i, len(auu)), MSGFYI)
		}
	}
}

func auproducer() <-chan string {
	auu := StringMapKeysIntoSlice(AllAuthors)
	c := make(chan string)
	go func() {
		for i := 0; i < len(auu); i++ {
			c <- auu[i]
		}
		close(c)
	}()
	return c
}

func auconsumer(auin <-chan string) {
	for au := range auin {
		fmt.Println(au)
		createandloadsqliteauthor(au)
	}
}

func aufanOutUnbuffered(ch <-chan string) []chan string {
	cs := make([]chan string, Config.WorkerCount)
	for i, _ := range cs {
		// The size of the channels buffer controls how far behind the recievers
		// of the fanOut channels can lag the other channels.
		cs[i] = make(chan string)
	}
	go func() {
		for au := range ch {
			for _, c := range cs {
				c <- au
			}
		}
		for _, c := range cs {
			// close all our fanOut channels when the input channel is exhausted.
			close(c)
		}
	}()
	return cs
}

func fosqliteloadactiveauthors() {
	c := auproducer()
	chans := aufanOutUnbuffered(c)
	for i := 0; i < Config.WorkerCount; i++ {
		go auconsumer(chans[i])
	}
}

func createandloadsqliteauthor(au string) {
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
		FAIL1 = `failed to create author table "%s": %s`
		EMB   = "emb/db/%s/%s.csv.gz"
		QT    = `insert into %s("index", wkuniversalid, 
					   level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, 
					   marked_up_line, accented_line, stripped_line, hyphenated_words, annotations) 
					   values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		FAIL5    = `ReadFile failed: %s`
		FAIL2    = `missing header row(?): %s`
		FAIL3    = `insert prepare failed: %s`
		FAIL4    = `[%s] gzip.NewReader failed`
		SUCCESS2 = `loaded author table "%s"`
	)

	authfail := func(e error) {
		if e != nil {
			msg(fmt.Sprintf(FAIL1, au, e.Error()), MSGWARN)
		}
	}

	sqlfail := func(e error, m string) {
		if e != nil {
			msg(fmt.Sprintf(m, e.Error()), MSGWARN)
		}
	}

	// create the author

	ltconn := GetSQLiteConn()
	defer ltconn.Close()

	// tx, err := memdb.Begin()

	q := fmt.Sprintf(CREATE, au)
	_, err := ltconn.ExecContext(context.Background(), q)
	authfail(err)

	// load the table

	// [a] decompress the sql data
	dump := fmt.Sprintf(EMB, au[0:2], au)

	compressed, err := os.ReadFile(dump)
	sqlfail(err, FAIL5)
	decompressed, err := gzip.NewReader(bytes.NewReader(compressed))
	sqlfail(err, FAIL4)

	r := csv.NewReader(decompressed)
	_, err = r.Read()
	sqlfail(err, FAIL2)

	// [b] prepare the statement
	q = fmt.Sprintf(QT, au)
	stmt, e := ltconn.PrepareContext(context.Background(), q)
	sqlfail(e, FAIL3)

	// [c] iterate over the records and insert
	for {
		record, e3 := r.Read()
		if errors.Is(e3, io.EOF) {
			break
		}
		_, e2 := stmt.Exec(record[0], record[1], record[2], record[3], record[4], record[5], record[6], record[7], record[8], record[9], record[10], record[11], record[12])
		if e2 != nil {
			log.Fatalf("insert failed(%s): %s", record[0], e2)
		}
	}

	chke(err)

	msg(fmt.Sprintf(SUCCESS2, au), MSGPEEK)
}

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

func connsqlitetestquery(ltconn *sql.Conn, tq string) {
	rows, err := ltconn.QueryContext(context.Background(), tq)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("connsqlitetestquery()-ed")

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
	msg("postinitializationsqlitetest()", 2)
	au := "lt0016"
	tq := fmt.Sprintf(`select "index", wkuniversalid,
               level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, 
               marked_up_line, accented_line, stripped_line, hyphenated_words, annotations from %s where stripped_line regexp 'est'`, au)

	tx, err := SQLITEConn.Begin()
	if err != nil {
		log.Fatal(err)
	}

	txsqlitetestquery(tx, tq)
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	ltconn := GetSQLiteConn()
	defer ltconn.Close()
	tq = fmt.Sprintf(`select "index", wkuniversalid,
               level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, 
               marked_up_line, accented_line, stripped_line, hyphenated_words, annotations from %s where "index" BETWEEN 10 and 15`, au)
	connsqlitetestquery(ltconn, tq)

	chke(err)
}
