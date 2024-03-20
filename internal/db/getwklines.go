//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package db

import (
	"context"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	WORLINETEMPLATE = `wkuniversalid, index,
			level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value,
			marked_up_line, accented_line, stripped_line, hyphenated_words, annotations`
)

//
// This file should contain the *exhaustive* collection of functions that execute searches
// that return either a WorkLineBundle or a DbWorkline
//

// AcquireWorkLineBundle - use a PrerolledQuery to acquire a *WorkLineBundle
func AcquireWorkLineBundle(prq str.PrerolledQuery, dbconn *pgxpool.Conn) *str.WorkLineBundle {
	// NB: you have to use a dbconn.Exec() and can't use SQLPool.Exex() because with the latter the temp table will
	// get separated from the main query:
	// ERROR: relation "{ttname}" does not exist (SQLSTATE 42P01)

	// [a] build a temp table if needed

	if prq.TempTable != "" {
		_, err := dbconn.Exec(context.Background(), prq.TempTable)
		Msg.EC(err)
	}

	// [b] execute the main query (nb: query needs to satisfy needs of RowToStructByPos in [c])

	foundrows, err := dbconn.Query(context.Background(), prq.PsqlQuery)
	Msg.EC(err)

	// [c] convert the finds into []DbWorkline

	thesefinds, err := pgx.CollectRows(foundrows, pgx.RowToStructByPos[str.DbWorkline])
	Msg.EC(err)

	return &str.WorkLineBundle{Lines: thesefinds}
}

// SimpleContextGrabber - grab a *WorkLineBundle centered around the focusline (only called by GenerateBrowsedPassage)
func SimpleContextGrabber(table string, focus int, context int) *str.WorkLineBundle {
	const (
		QTMPL = "SELECT %s FROM %s WHERE (index BETWEEN %d AND %d) ORDER by index"
	)

	dbconn := GetDBConnection()
	defer dbconn.Release()

	low := focus - context
	high := focus + context

	var prq str.PrerolledQuery
	prq.TempTable = ""
	prq.PsqlQuery = fmt.Sprintf(QTMPL, WORLINETEMPLATE, table, low, high)

	foundlines := AcquireWorkLineBundle(prq, dbconn)

	return foundlines
}

// GrabOneLine - return a single DbWorkline from a table
func GrabOneLine(table string, line int) str.DbWorkline {
	const (
		QTMPL = "SELECT %s FROM %s WHERE index = %d"
	)

	dbconn := GetDBConnection()
	defer dbconn.Release()

	var prq str.PrerolledQuery
	prq.TempTable = ""
	prq.PsqlQuery = fmt.Sprintf(QTMPL, WORLINETEMPLATE, table, line)
	foundlines := AcquireWorkLineBundle(prq, dbconn)
	if foundlines.Len() != 0 {
		// "index = %d" in QTMPL ought to mean you can never have len(foundlines) > 1 because index values are unique
		return foundlines.FirstLine()
	} else {
		return str.DbWorkline{}
	}
}
