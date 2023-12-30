//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
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
	user := readUUIDCookie(c)
	optandval := c.Param("opt")
	parsed := strings.Split(optandval, "/")

	if len(parsed) != 2 {
		msg(fmt.Sprintf(FAIL1, optandval), MSGWARN)
		return c.String(http.StatusOK, "")
	}

	opt := parsed[0]
	val := parsed[1]

	ynoptionlist := []string{"greekcorpus", "latincorpus", "papyruscorpus", "inscriptioncorpus", "christiancorpus",
		"rawinputstyle", "onehit", "headwordindexing", "indexbyfrequency", "spuria", "incerta", "varia", "vocbycount",
		"vocscansion", "isvectorsearch", "extendedgraph", "ldagraph", "isldasearch", "ldagraph2dimensions"}

	s := AllSessions.GetSess(user)

	modifyglobalmapsifneeded := func(c string, y bool) {
		// this is a "laggy" click: something comparable to the launch initialization time
		// if you call it via "go modifyglobalmapsifneeded()" the lag vanishes: nobody will search <.5s later, right?
		if y && !LoadedCorp[c] && SQLProvider == "pgsql" {
			start := time.Now()
			// append to the master work map
			AllWorks = mapnewworkcorpus(c, AllWorks)
			// append to the master author map
			AllAuthors = mapnewauthorcorpus(c, AllAuthors)
			// re-populateglobalmaps
			populateglobalmaps()
			d := fmt.Sprintf("modifyglobalmapsifneeded(): %.3fs", time.Now().Sub(start).Seconds())
			msg(d, MSGPEEK)
		}
		if SQLProvider != "pgsql" {
			msg("modifyglobalmapsifneeded(): SQLite cannot toggle databases (yet...)", MSGMAND)
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
				s.ActiveCorp[GREEKCORP] = b
				go modifyglobalmapsifneeded(GREEKCORP, b)
			case "latincorpus":
				s.ActiveCorp[LATINCORP] = b
				go modifyglobalmapsifneeded(LATINCORP, b)
			case "papyruscorpus":
				s.ActiveCorp[PAPYRUSCORP] = b
				go modifyglobalmapsifneeded(PAPYRUSCORP, b)
			case "inscriptioncorpus":
				s.ActiveCorp[INSCRIPTCORP] = b
				go modifyglobalmapsifneeded(INSCRIPTCORP, b)
			case "christiancorpus":
				s.ActiveCorp[CHRISTINSC] = b
				go modifyglobalmapsifneeded(CHRISTINSC, b)
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
				msg(FAIL2, MSGWARN)
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
			msg(FAIL2, MSGWARN)
		}
	}

	spinoptionlist := []string{"maxresults", "linesofcontext", "browsercontext", "proximity", "neighborcount", "ldatopiccount"}
	if slices.Contains(spinoptionlist, opt) {
		intval, e := strconv.Atoi(val)
		if e == nil {
			switch opt {
			case "maxresults":
				if intval < MAXHITLIMIT {
					s.HitLimit = intval
				} else {
					s.HitLimit = MAXHITLIMIT
				}
			case "linesofcontext":
				if intval < MAXLINESHITCONTEXT {
					s.HitContext = intval
				} else {
					s.HitContext = intval
				}
			case "browsercontext":
				if intval < MAXBROWSERCONTEXT {
					s.BrowseCtx = intval
				} else {
					s.BrowseCtx = MAXBROWSERCONTEXT
				}
			case "proximity":
				if 1 <= intval || intval <= MAXDISTANCE {
					s.Proximity = intval
				} else if intval < 1 {
					s.Proximity = 1
				} else {
					s.Proximity = MAXDISTANCE
				}
			case "neighborcount":
				if VECTORNEIGHBORSMIN <= intval || intval <= VECTORNEIGHBORSMAX {
					s.VecNeighbCt = intval
				} else if intval < VECTORNEIGHBORSMIN {
					s.VecNeighbCt = VECTORNEIGHBORSMIN
				} else {
					s.VecNeighbCt = VECTORNEIGHBORSMAX
				}
			case "ldatopiccount":
				if 1 <= intval || intval <= LDAMAXTOPICS {
					s.LDAtopics = intval
				} else if intval < 1 {
					s.LDAtopics = 1
				} else {
					s.LDAtopics = LDAMAXTOPICS
				}
			default:
				msg(FAIL2, MSGWARN)
			}
		}
	}

	dateoptionlist := []string{"earliestdate", "latestdate"}
	if slices.Contains(dateoptionlist, opt) {
		intval, e := strconv.Atoi(val)
		if e == nil {
			switch opt {
			case "earliestdate":
				if intval > MAXDATE {
					s.Earliest = fmt.Sprintf("%d", MAXDATE)
				} else if intval < MINDATE {
					s.Earliest = fmt.Sprintf("%d", MINDATE)
				} else {
					s.Earliest = val
				}
			case "latestdate":
				if intval > MAXDATE {
					s.Latest = fmt.Sprintf("%d", MAXDATE)
				} else if intval < MINDATE {
					s.Latest = fmt.Sprintf("%d", MINDATE)
				} else {
					s.Latest = val
				}
			default:
				msg(FAIL2, MSGWARN)
			}
		}

		ee, e1 := strconv.Atoi(s.Earliest)
		ll, e2 := strconv.Atoi(s.Latest)
		if e1 != nil {
			s.Earliest = MINDATESTR
		}
		if e2 != nil {
			s.Latest = MAXDATESTR
		}
		if e1 == nil && e2 == nil {
			if ee > ll {
				s.Earliest = s.Latest
			}
		}
	}

	AllSessions.InsertSess(s)
	return c.String(http.StatusOK, "")
}
