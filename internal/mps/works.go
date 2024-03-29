//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package mps

import (
	"context"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/jackc/pgx/v5"
	"strings"
)

const (
	WORKTEMPLATE = ` universalid, title, language, publication_info,
		levellabels_00, levellabels_01, levellabels_02, levellabels_03, levellabels_04, levellabels_05,
		workgenre, transmission, worktype, provenance, recorded_date, converted_date, wordcount,
		firstline, lastline, authentic`
)

// ActiveWorkMapper - build a map of all works in the *active* corpora; keyed to the authorUID: map[string]*DbWork
func ActiveWorkMapper() map[string]*str.DbWork {
	// note that you are still on the hook for adding to the global workmap when someone cals "/setoption/papyruscorpus/yes"
	// AND you should never drop from the map because that is only session-specific: "no" is only "no for me"
	// so the memory footprint can only grow: but G&L is an 82M vv vs an 189M vv for everything

	// the bookkeeping is handled by modifyglobalmapsifneeded() inside of RtSetOption()

	workmap := make(map[string]*str.DbWork)

	for k, b := range lnch.Config.DefCorp {
		if b {
			workmap = MapNewWorkCorpus(k, workmap)
		}
	}
	return workmap
}

// MapNewWorkCorpus - add a corpus to a workmap
func MapNewWorkCorpus(corpus string, workmap map[string]*str.DbWork) map[string]*str.DbWork {
	const (
		MSG = "MapNewWorkCorpus() added %d works from '%s'"
	)
	toadd := sliceworkcorpus(corpus)
	for i := 0; i < len(toadd); i++ {
		w := toadd[i]
		workmap[w.UID] = &w
	}

	LoadedCorp[corpus] = true

	Msg.PEEK(fmt.Sprintf(MSG, len(toadd), corpus))
	return workmap
}

// sliceworkcorpus - fetch all relevant works from the db as a DbWork slice
func sliceworkcorpus(corpus string) []str.DbWork {
	// this is far and away the "heaviest" bit of the whole program if you grab every known work
	// Total: 204MB
	// 65.35MB (flat, cum) 32.03% of Total

	// hipparchiaDB-# \d works
	//                            Table "public.works"
	//      Column      |          Type          | Collation | Nullable | Default
	//------------------+------------------------+-----------+----------+---------
	// universalid      | character(10)          |           |          |
	// title            | character varying(512) |           |          |
	// language         | character varying(10)  |           |          |
	// publication_info | text                   |           |          |
	// levellabels_00   | character varying(64)  |           |          |
	// levellabels_01   | character varying(64)  |           |          |
	// levellabels_02   | character varying(64)  |           |          |
	// levellabels_03   | character varying(64)  |           |          |
	// levellabels_04   | character varying(64)  |           |          |
	// levellabels_05   | character varying(64)  |           |          |
	// workgenre        | character varying(32)  |           |          |
	// transmission     | character varying(32)  |           |          |
	// worktype         | character varying(32)  |           |          |
	// provenance       | character varying(64)  |           |          |
	// recorded_date    | character varying(64)  |           |          |
	// converted_date   | integer                |           |          |
	// wordcount        | integer                |           |          |
	// firstline        | integer                |           |          |
	// lastline         | integer                |           |          |
	// authentic        | boolean                |           |          |
	const (
		CT = `SELECT count(*) FROM works WHERE universalid ~* '^%s'`
		QT = `SELECT %s FROM works WHERE universalid ~* '^%s'`
	)

	var cc int
	cq := fmt.Sprintf(CT, corpus)
	qq := fmt.Sprintf(QT, WORKTEMPLATE, corpus)

	countrow := db.SQLPool.QueryRow(context.Background(), cq)
	err := countrow.Scan(&cc)

	foundrows, err := db.SQLPool.Query(context.Background(), qq)
	Msg.EC(err)

	workslice := make([]str.DbWork, cc)
	var w str.DbWork

	foreach := []any{&w.UID, &w.Title, &w.Language, &w.Pub, &w.LL0, &w.LL1, &w.LL2, &w.LL3, &w.LL4, &w.LL5, &w.Genre,
		&w.Xmit, &w.Type, &w.Prov, &w.RecDate, &w.ConvDate, &w.WdCount, &w.FirstLine, &w.LastLine, &w.Authentic}

	index := 0
	rwfnc := func() error {
		workslice[index] = w
		index++
		return nil
	}

	_, e := pgx.ForEachRow(foundrows, foreach, rwfnc)
	Msg.EC(e)

	return workslice
}

// Buildaucorpusmap - populate global variable used by SessionIntoSearchlist()
func Buildwkcorpusmap() map[string][]string {
	// SessionIntoSearchlist() could just grab a pre-rolled list instead of calculating every time...
	wkcorpusmap := make(map[string][]string)
	for _, w := range AllWorks {
		for _, c := range vv.TheCorpora {
			if w.UID[0:2] == c {
				wkcorpusmap[c] = append(wkcorpusmap[c], w.UID)
			}
		}
	}
	return wkcorpusmap
}

// Buildwkgenresmap - populate global variable used by hinter
func Buildwkgenresmap() map[string]bool {
	genres := make(map[string]bool)
	for _, w := range AllWorks {
		genres[w.Genre] = true
	}
	return genres
}

// Buildaulocationmap - populate global variable used by hinter
func Buildaulocationmap() map[string]bool {
	locations := make(map[string]bool)
	for _, a := range AllAuthors {
		ll := strings.Split(a.Location, ",")
		for _, l := range ll {
			locations[l] = true
		}
	}
	return locations
}

// Buildwklocationmap - populate global variable used by hinter
func Buildwklocationmap() map[string]bool {
	locations := make(map[string]bool)
	for _, w := range AllWorks {
		locations[w.Prov] = true
	}
	return locations
}
