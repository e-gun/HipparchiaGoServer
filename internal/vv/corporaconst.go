//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vv

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	TLGABBREV = "tlg" // this and the following *must* match the corresponding HipparchiaGoBuilder values
	LATABBREV = "lat"
	DDPABREV  = "dpx"
	INSABBREV = "inx"
	CHRABBREV = "chx"

	AUNAMELEN = 4
	WKIDLEN   = 3
	ABBREVLEN = 3
	AUIDLEN   = ABBREVLEN + AUNAMELEN

	//any change to the above needs to be reflected in `web/emb/js/autocomplete.js` where on sees:
	//const WORKNAMELEN = 3;
	//const CORPPREFIXLEN = 3;
	//const AUTHNAMELEN = 4;
	//const AUIDLEN = CORPPREFIXLEN + AUTHNAMELEN;
	//const NUMBEROFLEVELS = 6;
)

var (
	TheCorpora = []string{TLGABBREV, LATABBREV, INSABBREV, CHRABBREV, DDPABREV}
	// GrCt - this and the next will be (re)set at launch time; but the supplied values are almost certainly right
	GrCt           = 1823
	LtCt           = 363
	InCt           = 463
	ChCt           = 291
	DpCt           = 516
	DefaultCorpora = fmt.Sprintf(`{"%s": true, "%s": true, "%s": false, "%s": false, "%s": false}`, TLGABBREV, LATABBREV, DDPABREV, INSABBREV, CHRABBREV)
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
