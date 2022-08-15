package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"strconv"
)

// findtherows - use a redis.Conn to acquire []DbWorkline
func findtherows(thequery string, thecaller string, searchkey string, clientnumber int, rc redis.Conn, dbpool *pgxpool.Pool) []DbWorkline {
	// called by both linegrabber() and HipparchiaBagger()
	// this version contains polling data
	// it also assumes that thequery arrived via popping redis

	// [ii] update the polling data
	if thecaller != "bagger" {
		remain, err := redis.Int64(rc.Do("SCARD", searchkey))
		checkerror(err)

		k := fmt.Sprintf("%s_remaining", searchkey)
		_, e := rc.Do("SET", k, remain)
		checkerror(e)
		msg(fmt.Sprintf("%s #%d says that %d items remain", thecaller, clientnumber, remain), 3)
	}

	// [iii] decode the query
	var prq PrerolledQuery
	err := json.Unmarshal([]byte(thequery), &prq)
	checkerror(err)

	foundlines := worklinequery(prq, dbpool)

	return foundlines
}

// worklinequery - use a PrerolledQuery to acquire []DbWorkline
func worklinequery(prq PrerolledQuery, dbpool *pgxpool.Pool) []DbWorkline {
	// we omit keeping polling data...

	// [iv] build a temp table if needed
	if prq.TempTable != "" {
		_, err := dbpool.Exec(context.Background(), prq.TempTable)
		checkerror(err)
	}

	// [v] execute the main query
	var foundrows pgx.Rows
	var err error

	if prq.PsqlData != "" {
		foundrows, err = dbpool.Query(context.Background(), prq.PsqlQuery, prq.PsqlData)
		checkerror(err)
	} else {
		foundrows, err = dbpool.Query(context.Background(), prq.PsqlQuery)
		checkerror(err)
	}

	// [vi] iterate through the finds
	// don't check-and-load find-by-find because some searches are effectively uncapped
	// faster to test only after you finish each query
	// can over-stuff redis because HipparchaServer should only display hitcap results no matter how many you push

	var thesefinds []DbWorkline

	defer foundrows.Close()
	for foundrows.Next() {
		// [vi.1] convert the finds into DbWorklines
		var thehit DbWorkline
		err := foundrows.Scan(&thehit.WkUID, &thehit.TbIndex, &thehit.Lvl5Value, &thehit.Lvl4Value, &thehit.Lvl3Value,
			&thehit.Lvl2Value, &thehit.Lvl1Value, &thehit.Lvl0Value, &thehit.MarkedUp, &thehit.Accented,
			&thehit.Stripped, &thehit.Hypenated, &thehit.Annotations)
		checkerror(err)
		thesefinds = append(thesefinds, thehit)
	}

	return thesefinds
}

// simplecontextgrabber - grab a pile of lines centered around the focusline
func simplecontextgrabber(table string, focus int64, context int64) []DbWorkline {
	dbpool := grabpgsqlconnection()

	qt := "SELECT %s FROM %s WHERE (index BETWEEN %s AND %s) ORDER by index"

	low := focus - (context / 2)
	high := focus + (context / 2)

	var prq PrerolledQuery

	prq.TempTable = ""
	prq.PsqlData = ""
	prq.PsqlQuery = fmt.Sprintf(qt, WORLINETEMPLATE, table, strconv.FormatInt(low, 10), strconv.FormatInt(high, 10))

	foundlines := worklinequery(prq, dbpool)

	return foundlines
}
