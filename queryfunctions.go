//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//
// This file should contain the *exhaustive* collection of functions that execute searches
// that return either a WorkLineBundle or a DbWorkline
//

// AcquireWorkLineBundle - use a PrerolledQuery to acquire a *WorkLineBundle
func AcquireWorkLineBundle(prq PrerolledQuery, dbconn *pgxpool.Conn) *WorkLineBundle {
	// NB: you have to use a dbconn.Exec() and can't use SQLPool.Exex() because with the latter the temp table will
	// get separated from the main query:
	// ERROR: relation "{ttname}" does not exist (SQLSTATE 42P01)

	// [a] build a temp table if needed

	if prq.TempTable != "" {
		_, err := dbconn.Exec(context.Background(), prq.TempTable)
		chke(err)
	}

	// [b] execute the main query (nb: query needs to satisfy needs of RowToStructByPos in [c])

	foundrows, err := dbconn.Query(context.Background(), prq.PsqlQuery)
	chke(err)

	// [c] convert the finds into []DbWorkline

	thesefinds, err := pgx.CollectRows(foundrows, pgx.RowToStructByPos[DbWorkline])
	chke(err)

	return &WorkLineBundle{Lines: thesefinds}
}

// SimpleContextGrabber - grab a *WorkLineBundle centered around the focusline (only called by GenerateBrowsedPassage)
func SimpleContextGrabber(table string, focus int, context int) *WorkLineBundle {
	const (
		QTMPL = "SELECT %s FROM %s WHERE (index BETWEEN %d AND %d) ORDER by index"
	)

	dbconn := GetDBConnection()
	defer dbconn.Release()

	low := focus - context
	high := focus + context

	var prq PrerolledQuery
	prq.TempTable = ""
	prq.PsqlQuery = fmt.Sprintf(QTMPL, WORLINETEMPLATE, table, low, high)

	foundlines := AcquireWorkLineBundle(prq, dbconn)

	return foundlines
}

// GrabOneLine - return a single DbWorkline from a table
func GrabOneLine(table string, line int) DbWorkline {
	const (
		QTMPL = "SELECT %s FROM %s WHERE index = %d"
	)

	dbconn := GetDBConnection()
	defer dbconn.Release()

	var prq PrerolledQuery
	prq.TempTable = ""
	prq.PsqlQuery = fmt.Sprintf(QTMPL, WORLINETEMPLATE, table, line)
	foundlines := AcquireWorkLineBundle(prq, dbconn)
	if foundlines.Len() != 0 {
		// "index = %d" in QTMPL ought to mean you can never have len(foundlines) > 1 because index values are unique
		return foundlines.FirstLine()
	} else {
		return DbWorkline{}
	}
}
