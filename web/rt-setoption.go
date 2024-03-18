//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"context"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/mps"
	"github.com/e-gun/HipparchiaGoServer/internal/structs"
	"github.com/e-gun/HipparchiaGoServer/internal/vaults"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

// RtSetOption - modify the session in light of the selection made
func RtSetOption(c echo.Context) error {
	const (
		FAIL1 = "RtSetOption() was given bad input: %s"
		FAIL2 = "RtSetOption() hit an impossible case"
	)
	user := vaults.ReadUUIDCookie(c)
	optandval := c.Param("opt")
	parsed := strings.Split(optandval, "/")

	if len(parsed) != 2 {
		msg.WARN(fmt.Sprintf(FAIL1, optandval))
		return c.String(http.StatusOK, "")
	}

	opt := parsed[0]
	val := parsed[1]

	ynoptionlist := []string{"greekcorpus", "latincorpus", "papyruscorpus", "inscriptioncorpus", "christiancorpus",
		"rawinputstyle", "onehit", "headwordindexing", "indexbyfrequency", "spuria", "incerta", "varia", "vocbycount",
		"vocscansion", "isvectorsearch", "extendedgraph", "ldagraph", "isldasearch", "ldagraph2dimensions"}

	s := vaults.AllSessions.GetSess(user)

	modifyglobalmapsifneeded := func(c string, y bool) {
		// this is a "laggy" click: something comparable to the vv initialization time
		// if you call it via "go modifyglobalmapsifneeded()" the lag vanishes: nobody will search <.5s later, right?
		if y && !vv.LoadedCorp[c] {
			start := time.Now()
			// append to the master work map
			mps.AllWorks = mapnewworkcorpus(c, mps.AllWorks)
			// append to the master author map
			mps.AllAuthors = mapnewauthorcorpus(c, mps.AllAuthors)
			// re-populateglobalmaps
			mps.RePopulateGlobalMaps()
			d := fmt.Sprintf("modifyglobalmapsifneeded(): %.3fs", time.Now().Sub(start).Seconds())
			msg.PEEK(d)
		}
	}

	if slices.Contains(ynoptionlist, opt) {
		valid := []string{"yes", "no"}
		if slices.Contains(valid, val) {
			var b bool
			if val == "yes" {
				b = true
			} else {
				b = false
			}
			switch opt {
			case "greekcorpus":
				s.ActiveCorp[vv.GREEKCORP] = b
				go modifyglobalmapsifneeded(vv.GREEKCORP, b)
			case "latincorpus":
				s.ActiveCorp[vv.LATINCORP] = b
				go modifyglobalmapsifneeded(vv.LATINCORP, b)
			case "papyruscorpus":
				s.ActiveCorp[vv.PAPYRUSCORP] = b
				go modifyglobalmapsifneeded(vv.PAPYRUSCORP, b)
			case "inscriptioncorpus":
				s.ActiveCorp[vv.INSCRIPTCORP] = b
				go modifyglobalmapsifneeded(vv.INSCRIPTCORP, b)
			case "christiancorpus":
				s.ActiveCorp[vv.CHRISTINSC] = b
				go modifyglobalmapsifneeded(vv.CHRISTINSC, b)
			case "rawinputstyle":
				s.RawInput = b
			case "onehit":
				s.OneHit = b
			case "indexbyfrequency":
				s.FrqIdx = b
			case "headwordindexing":
				s.HeadwordIdx = b
			case "spuria":
				s.SpuriaOK = b
			case "incerta":
				s.IncertaOK = b
			case "varia":
				s.VariaOK = b
			case "vocbycount":
				s.VocByCount = b
			case "vocscansion":
				s.VocScansion = b
			case "isvectorsearch":
				s.VecNNSearch = b
			case "isldasearch":
				s.VecLDASearch = b
			case "extendedgraph":
				s.VecGraphExt = b
			case "ldagraph":
				s.LDAgraph = b
			case "ldagraph2dimensions":
				s.LDA2D = b
			default:
				msg.WARN(FAIL2)
			}
		}
	}

	valoptionlist := []string{"nearornot", "searchscope", "sortorder", "modeler", "vtextprep"}
	if slices.Contains(valoptionlist, opt) {
		switch opt {
		case "nearornot":
			valid := []string{"near", "notnear"}
			if slices.Contains(valid, val) {
				s.NearOrNot = val
			}
		case "searchscope":
			valid := []string{"lines", "words"}
			if slices.Contains(valid, val) {
				s.SearchScope = val
			}
		case "sortorder":
			valid := []string{"shortname", "converted_date", "provenance", "universalid"}
			if slices.Contains(valid, val) {
				s.SortHitsBy = val
			}
		case "modeler":
			valid := []string{"w2v", "glove", "lexvec"}
			if slices.Contains(valid, val) {
				s.VecModeler = val
			}
		case "vtextprep":
			valid := []string{"winner", "unparsed", "yoked", "montecarlo"}
			if slices.Contains(valid, val) {
				s.VecTextPrep = val
			}
		default:
			msg.WARN(FAIL2)
		}
	}

	spinoptionlist := []string{"maxresults", "linesofcontext", "browsercontext", "proximity", "neighborcount", "ldatopiccount"}
	if slices.Contains(spinoptionlist, opt) {
		intval, e := strconv.Atoi(val)
		if e == nil {
			switch opt {
			case "maxresults":
				if intval < vv.MAXHITLIMIT {
					s.HitLimit = intval
				} else {
					s.HitLimit = vv.MAXHITLIMIT
				}
			case "linesofcontext":
				if intval < vv.MAXLINESHITCONTEXT {
					s.HitContext = intval
				} else {
					s.HitContext = intval
				}
			case "browsercontext":
				if intval < vv.MAXBROWSERCONTEXT {
					s.BrowseCtx = intval
				} else {
					s.BrowseCtx = vv.MAXBROWSERCONTEXT
				}
			case "proximity":
				if 1 <= intval || intval <= vv.MAXDISTANCE {
					s.Proximity = intval
				} else if intval < 1 {
					s.Proximity = 1
				} else {
					s.Proximity = vv.MAXDISTANCE
				}
			case "neighborcount":
				if vv.VECTORNEIGHBORSMIN <= intval || intval <= vv.VECTORNEIGHBORSMAX {
					s.VecNeighbCt = intval
				} else if intval < vv.VECTORNEIGHBORSMIN {
					s.VecNeighbCt = vv.VECTORNEIGHBORSMIN
				} else {
					s.VecNeighbCt = vv.VECTORNEIGHBORSMAX
				}
			case "ldatopiccount":
				if 1 <= intval || intval <= vv.LDAMAXTOPICS {
					s.LDAtopics = intval
				} else if intval < 1 {
					s.LDAtopics = 1
				} else {
					s.LDAtopics = vv.LDAMAXTOPICS
				}
			default:
				msg.WARN(FAIL2)
			}
		}
	}

	dateoptionlist := []string{"earliestdate", "latestdate"}
	if slices.Contains(dateoptionlist, opt) {
		intval, e := strconv.Atoi(val)
		if e == nil {
			switch opt {
			case "earliestdate":
				if intval > vv.MAXDATE {
					s.Earliest = fmt.Sprintf("%d", vv.MAXDATE)
				} else if intval < vv.MINDATE {
					s.Earliest = fmt.Sprintf("%d", vv.MINDATE)
				} else {
					s.Earliest = val
				}
			case "latestdate":
				if intval > vv.MAXDATE {
					s.Latest = fmt.Sprintf("%d", vv.MAXDATE)
				} else if intval < vv.MINDATE {
					s.Latest = fmt.Sprintf("%d", vv.MINDATE)
				} else {
					s.Latest = val
				}
			default:
				msg.WARN(FAIL2)
			}
		}

		ee, e1 := strconv.Atoi(s.Earliest)
		ll, e2 := strconv.Atoi(s.Latest)
		if e1 != nil {
			s.Earliest = vv.MINDATESTR
		}
		if e2 != nil {
			s.Latest = vv.MAXDATESTR
		}
		if e1 == nil && e2 == nil {
			if ee > ll {
				s.Earliest = s.Latest
			}
		}
	}

	vaults.AllSessions.InsertSess(s)
	return c.String(http.StatusOK, "")
}

// mapnewworkcorpus - add a corpus to a workmap
func mapnewworkcorpus(corpus string, workmap map[string]*structs.DbWork) map[string]*structs.DbWork {
	const (
		MSG = "mapnewworkcorpus() added %d works from '%s'"
	)
	toadd := sliceworkcorpus(corpus)
	for i := 0; i < len(toadd); i++ {
		w := toadd[i]
		workmap[w.UID] = &w
	}

	vv.LoadedCorp[corpus] = true

	msg.PEEK(fmt.Sprintf(MSG, len(toadd), corpus))
	return workmap
}

// sliceworkcorpus - fetch all relevant works from the db as a DbWork slice
func sliceworkcorpus(corpus string) []structs.DbWork {
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
	qq := fmt.Sprintf(QT, mps.WORKTEMPLATE, corpus)

	countrow := db.SQLPool.QueryRow(context.Background(), cq)
	err := countrow.Scan(&cc)

	foundrows, err := db.SQLPool.Query(context.Background(), qq)
	msg.EC(err)

	workslice := make([]structs.DbWork, cc)
	var w structs.DbWork

	foreach := []any{&w.UID, &w.Title, &w.Language, &w.Pub, &w.LL0, &w.LL1, &w.LL2, &w.LL3, &w.LL4, &w.LL5, &w.Genre,
		&w.Xmit, &w.Type, &w.Prov, &w.RecDate, &w.ConvDate, &w.WdCount, &w.FirstLine, &w.LastLine, &w.Authentic}

	index := 0
	rwfnc := func() error {
		workslice[index] = w
		index++
		return nil
	}

	_, e := pgx.ForEachRow(foundrows, foreach, rwfnc)
	msg.EC(e)

	return workslice
}

// mapnewauthorcorpus - add a corpus to an authormap
func mapnewauthorcorpus(corpus string, authmap map[string]*structs.DbAuthor) map[string]*structs.DbAuthor {
	const (
		MSG = "mapnewauthorcorpus() added %d authors from '%s'"
	)

	toadd := sliceauthorcorpus(corpus)
	for i := 0; i < len(toadd); i++ {
		a := toadd[i]
		authmap[a.UID] = &a
	}

	vv.LoadedCorp[corpus] = true

	msg.PEEK(fmt.Sprintf(MSG, len(toadd), corpus))

	return authmap
}

// sliceauthorcorpus - fetch all relevant works from the db as a DbAuthor slice
func sliceauthorcorpus(corpus string) []structs.DbAuthor {
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
	for _, w := range mps.AllWorks {
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
	qq := fmt.Sprintf(QT, mps.AUTHORTEMPLATE, corpus)

	countrow := db.SQLPool.QueryRow(context.Background(), cq)
	err := countrow.Scan(&cc)

	foundrows, err := db.SQLPool.Query(context.Background(), qq)
	msg.EC(err)

	authslice := make([]structs.DbAuthor, cc)
	var a structs.DbAuthor
	foreach := []any{&a.UID, &a.Language, &a.IDXname, &a.Name, &a.Shortname, &a.Cleaname, &a.Genres, &a.RecDate, &a.ConvDate, &a.Location}

	index := 0
	rfnc := func() error {
		a.WorkList = worklists[a.UID]
		authslice[index] = a
		index++
		return nil
	}

	_, e := pgx.ForEachRow(foundrows, foreach, rfnc)
	msg.EC(e)

	return authslice
}
