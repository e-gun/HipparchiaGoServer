//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vv

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

// if you alter the db contents, you need to adjust this...

const (
	DEFAULTCORPORA = "{\"gr\": true, \"lt\": true, \"in\": false, \"ch\": false, \"dp\": false}"
	GREEKCORP      = "gr"
	LATINCORP      = "lt"
	PAPYRUSCORP    = "dp"
	INSCRIPTCORP   = "in"
	CHRISTINSC     = "ch"
)

var (
	TheCorpora = []string{GREEKCORP, LATINCORP, INSCRIPTCORP, CHRISTINSC, PAPYRUSCORP}
	// GrCt - this and the next will be (re)set at launch time; but the supplied values are almost certainly right
	GrCt = 1823
	LtCt = 363
	InCt = 463
	ChCt = 291
	DpCt = 516
)

func SetCorpusCount(corp string, p *pgxpool.Pool) int {
	const (
		Q = `SELECT count(*) FROM authors WHERE universalid ~* '^%s'`
	)

	theresult, err := p.Query(context.Background(), fmt.Sprintf(Q, corp))
	if err != nil {
		panic(err)
	}

	ct := 0

	for theresult.Next() {
		theval := 0
		e := theresult.Scan(&theval)
		if e != nil {
			panic(e)
		}
		ct = theval
	}
	theresult.Close()
	return ct
}
