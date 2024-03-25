//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/mps"
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
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
	user := vlt.ReadUUIDCookie(c)
	optandval := c.Param("opt")
	parsed := strings.Split(optandval, "/")

	if len(parsed) != 2 {
		Msg.WARN(fmt.Sprintf(FAIL1, optandval))
		return c.String(http.StatusOK, "")
	}

	opt := parsed[0]
	val := parsed[1]

	ynoptionlist := []string{"greekcorpus", "latincorpus", "papyruscorpus", "inscriptioncorpus", "christiancorpus",
		"rawinputstyle", "onehit", "headwordindexing", "indexbyfrequency", "spuria", "incerta", "varia", "vocbycount",
		"vocscansion", "isvectorsearch", "extendedgraph", "ldagraph", "isldasearch", "ldagraph2dimensions"}

	s := vlt.AllSessions.GetSess(user)

	modifyglobalmapsifneeded := func(c string, y bool) {
		// this is a "laggy" click: something comparable to the vv initialization time
		// if you call it via "go modifyglobalmapsifneeded()" the lag vanishes: nobody will search <.5s later, right?
		if y && !mps.LoadedCorp[c] {
			start := time.Now()
			// append to the master work map
			mps.AllWorks = mps.MapNewWorkCorpus(c, mps.AllWorks)
			// append to the master author map
			mps.AllAuthors = mps.MapNewAuthorCorpus(c, mps.AllAuthors)
			// re-populateglobalmaps
			mps.RePopulateGlobalMaps()
			d := fmt.Sprintf("modifyglobalmapsifneeded(): %.3fs", time.Now().Sub(start).Seconds())
			Msg.PEEK(d)
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
				Msg.WARN(FAIL2)
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
			Msg.WARN(FAIL2)
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
				Msg.WARN(FAIL2)
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
				Msg.WARN(FAIL2)
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

	vlt.AllSessions.InsertSess(s)
	return c.String(http.StatusOK, "")
}
