//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package mps

import (
	"context"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/str"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/jackc/pgx/v5"
	"strings"
)

var (
	Msg = lnch.NewMessageMakerWithDefaults()
)

const (
	AUTHORTEMPLATE = ` universalid, language, idxname, akaname, shortname, cleanname, genres, recorded_date, converted_date, location `
)

// ActiveAuthorMapper - build a map of all authors in the *active* corpora; keyed to the authorUID: map[string]*DbAuthor
func ActiveAuthorMapper() map[string]*str.DbAuthor {
	// see comments at top of ActiveWorkMapper(): they apply here too
	authmap := make(map[string]*str.DbAuthor)
	for k, b := range lnch.Config.DefCorp {
		if b {
			authmap = MapNewAuthorCorpus(k, authmap)
		}
	}
	return authmap
}

// MapNewAuthorCorpus - add a corpus to an authormap
func MapNewAuthorCorpus(corpus string, authmap map[string]*str.DbAuthor) map[string]*str.DbAuthor {
	const (
		MSG = "MapNewAuthorCorpus() added %d authors from '%s'"
	)

	toadd := sliceauthorcorpus(corpus)
	for i := 0; i < len(toadd); i++ {
		a := toadd[i]
		authmap[a.UID] = &a
	}

	LoadedCorp[corpus] = true

	Msg.PEEK(fmt.Sprintf(MSG, len(toadd), corpus))

	return authmap
}

// sliceauthorcorpus - fetch all relevant works from the db as a DbAuthor slice
func sliceauthorcorpus(corpus string) []str.DbAuthor {
	// hipparchiaDB-# \d authors
	//                          Table "public.authors"
	//     Column     |          Type          | Collation | Nullable | Default
	//----------------+------------------------+-----------+----------+---------
	// universalid    | character(6)           |           |          |
	// language       | character varying(10)  |           |          |
	// idxname        | character varying(128) |           |          |
	// akaname        | character varying(128) |           |          |
	// shortname      | character varying(128) |           |          |
	// cleanname      | character varying(128) |           |          |
	// genres         | character varying(512) |           |          |
	// recorded_date  | character varying(64)  |           |          |
	// converted_date | integer                |           |          |
	// location       | character varying(128) |           |          |

	const (
		CT = `SELECT count(*) FROM authors WHERE universalid ~* '^%s'`
		QT = `SELECT %s FROM authors WHERE universalid ~* '^%s'`
	)

	// need to be ready to load the worklists into the authors
	// so: build a map of {UID: WORKLIST...}; map called by rfnc()

	worklists := make(map[string][]string)
	for _, w := range AllWorks {
		wk := w.UID
		au := wk[0:vv.LENGTHOFAUTHORID]
		if _, y := worklists[au]; !y {
			worklists[au] = []string{wk}
		} else {
			worklists[au] = append(worklists[au], wk)
		}
	}

	var cc int
	cq := fmt.Sprintf(CT, corpus)
	qq := fmt.Sprintf(QT, AUTHORTEMPLATE, corpus)

	countrow := db.SQLPool.QueryRow(context.Background(), cq)
	err := countrow.Scan(&cc)

	foundrows, err := db.SQLPool.Query(context.Background(), qq)
	Msg.EC(err)

	authslice := make([]str.DbAuthor, cc)
	var a str.DbAuthor
	foreach := []any{&a.UID, &a.Language, &a.IDXname, &a.Name, &a.Shortname, &a.Cleaname, &a.Genres, &a.RecDate, &a.ConvDate, &a.Location}

	index := 0
	rfnc := func() error {
		a.WorkList = worklists[a.UID]
		authslice[index] = a
		index++
		return nil
	}

	_, e := pgx.ForEachRow(foundrows, foreach, rfnc)
	Msg.EC(e)

	return authslice
}

// Buildaucorpusmap - populate global variable used by SessionIntoSearchlist()
func Buildaucorpusmap() map[string][]string {
	// SessionIntoSearchlist() could just grab a pre-rolled list instead of calculating every time...
	aucorpusmap := make(map[string][]string)
	for _, a := range AllAuthors {
		for _, c := range vv.TheCorpora {
			if a.UID[0:2] == c {
				aucorpusmap[c] = append(aucorpusmap[c], a.UID)
			}
		}
	}
	return aucorpusmap
}

// Buildaugenresmap - populate global variable used by hinter
func Buildaugenresmap() map[string]bool {
	genres := make(map[string]bool)
	for _, a := range AllAuthors {
		gg := strings.Split(a.Genres, ",")
		for _, g := range gg {
			genres[g] = true
		}
	}
	return genres
}
