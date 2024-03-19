//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/mps"
	"github.com/e-gun/HipparchiaGoServer/internal/search"
	"github.com/e-gun/HipparchiaGoServer/internal/str"
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type JSData struct {
	Pound string
	Url   string
}

// SelectionValues - what was registerselection?
type SelectionValues struct {
	Auth   string
	Work   string
	AGenre string
	WGenre string
	ALoc   string
	WLoc   string
	IsExcl bool
	IsRaw  bool
	Start  string
	End    string
}

// WUID - return work universalid
func (s SelectionValues) WUID() string {
	return s.Auth + "w" + s.Work
}

// AWPR - author, work, passage start, passage end
func (s SelectionValues) AWPR() bool {
	if len(s.Auth) > 0 && len(s.Work) > 0 && len(s.Start) > 0 && len(s.End) > 0 {
		return true
	} else {
		return false
	}
}

// AWP - author, work, and passage
func (s SelectionValues) AWP() bool {
	if len(s.Auth) > 0 && len(s.Work) > 0 && len(s.Start) > 0 && len(s.End) == 0 {
		return true
	} else {
		return false
	}
}

// AW - author, work, and not passage
func (s SelectionValues) AW() bool {
	if len(s.Auth) > 0 && len(s.Work) > 0 && len(s.Start) == 0 && len(s.End) == 0 {
		return true
	} else {
		return false
	}
}

// A - author, not work, and not passage
func (s SelectionValues) A() bool {
	if len(s.Auth) > 0 && len(s.Work) == 0 && len(s.Start) == 0 && len(s.End) == 0 {
		return true
	} else {
		return false
	}
}

// RtSelectionMake - register a selection and modify the session accordingly
func RtSelectionMake(c echo.Context) error {
	// GET http://localhost:8000/selection/make/_?auth=lt0474&work=073&locus=3|10&endpoint=

	// note that you need to return JSON: reportcurrentselections() to fill #selectionstable on the page
	// see bottom of this file for a sample list and the html of the table that goes with it

	user := vlt.ReadUUIDCookie(c)

	var sel SelectionValues
	sel.Auth = c.QueryParam("auth")
	sel.Work = c.QueryParam("work")
	sel.Start = c.QueryParam("locus")
	sel.End = c.QueryParam("endpoint")
	sel.AGenre = c.QueryParam("genre")
	sel.WGenre = c.QueryParam("wkgenre")
	sel.ALoc = c.QueryParam("auloc")
	sel.WLoc = c.QueryParam("wkprov")

	if c.QueryParam("raw") == "t" {
		sel.IsRaw = true
	} else {
		sel.IsRaw = false
	}

	if c.QueryParam("exclude") == "t" {
		sel.IsExcl = true
	} else {
		sel.IsExcl = false
	}

	ns := registerselection(user, sel)
	vlt.AllSessions.InsertSess(ns)

	cs := reportcurrentselections(c)

	return c.JSONPretty(http.StatusOK, cs, vv.JSONINDENT)
}

// RtSelectionClear - remove a selection from the session
func RtSelectionClear(c echo.Context) error {
	const (
		FAIL1 = "RtSelectionClear() was given bad input: %s"
		FAIL2 = "RtSelectionClear() was given bad category: %s"
	)

	// NB: restarting the server with an open browser can leave an impossible JS click; not really a bug, but...

	user := vlt.ReadUUIDCookie(c)

	locus := c.Param("locus")
	which := strings.Split(locus, "/")

	if len(which) != 2 {
		Msg.WARN(fmt.Sprintf(FAIL1, locus))
		return emptyjsreturn(c)
	}

	cat := which[0]
	id := which[1]

	newsess := vlt.AllSessions.GetSess(user)
	search.BuildSelectionOverview(&newsess)
	newincl := newsess.Inclusions
	newexcl := newsess.Exclusions

	// sliceprinter("ListedPBN", newincl.ListedPBN)

	// kvpairmdkey - if the md5 of a v is in a k,v map, return the k
	kvpairmdkey := func(m string, kvp map[string]string) string {
		targetkey := ""
		for k, v := range kvp {
			mv := fmt.Sprintf("%x", md5.Sum([]byte(v)))
			if mv == m {
				targetkey = k
			}
		}
		return targetkey
	}

	// kvpairmd - if the md5 of a v is in a k,v map, return the v {
	kvpairmdval := func(m string, kvp map[string]string) string {
		targetval := ""
		for _, v := range kvp {
			mv := fmt.Sprintf("%x", md5.Sum([]byte(v)))
			if mv == m {
				targetval = v
			}
		}
		return targetval
	}

	// removemd - remove item from a []string if its md5sum matches
	removemd := func(ss []string, drop string) []string {
		ret := make([]string, 0)
		for i := 0; i < len(ss); i++ {
			md := fmt.Sprintf("%x", md5.Sum([]byte(ss[i])))
			if md != drop {
				ret = append(ret, ss[i])
			}
		}
		return ret
	}

	// u: /selection/clear/agnselections/a825ca922ccd8c8427edd2dfa4ac8aa6
	// 51b9ec70f5659ee042d6ee610b74887e is the md5sum of 'Epici'

	// u: /selection/clear/auselections/f6de296b2941374db110d2d3683e8ca9
	// f6de296b2941374db110d2d3683e8ca9 is the md5sum of 'Cicero - Cicero, Marcus Tullius'

	// u: /selection/clear/wkselections/73c1352631d6293a867de7dc77eb2360
	// 73c1352631d6293a867de7dc77eb2360: 'Cicero, Cato Maior de Senectute'

	// u: /selection/clear/psgselections/7a27ab72793be0bd038c77661cc606d9
	// 7a27ab72793be0bd038c77661cc606d9: 'Cicero, Cato Maior de Senectute, 13 - 20'

	switch cat {
	// PART ONE: INCLUSIONS
	case "agnselections":
		newincl.AuGenres = removemd(newincl.AuGenres, id)
	case "wgnselections":
		newincl.WkGenres = removemd(newincl.WkGenres, id)
	case "alocselections":
		newincl.AuLocations = removemd(newincl.AuLocations, id)
	case "wlocselections":
		newincl.WkLocations = removemd(newincl.WkLocations, id)
	case "auselections":
		// auselections + wkselections + psgselections + ...
		// direction: MappedAuthByName --> Authors
		// MappedAuthByName: lt0474:Cicero - Cicero, Marcus Tullius
		// Authors: lt0474
		// id is the md5 for "Cicero - Cicero, Marcus Tullius"
		foundkey := kvpairmdkey(id, newincl.MappedAuthByName)
		newincl.Authors = gen.SetSubtraction(newincl.Authors, []string{foundkey})
	case "wkselections":
		foundkey := kvpairmdkey(id, newincl.MappedWkByName)
		newincl.Works = gen.SetSubtraction(newincl.Works, []string{foundkey})
	case "psgselections":
		foundkey := kvpairmdkey(id, newincl.MappedPsgByName)
		newincl.Passages = gen.SetSubtraction(newincl.Passages, []string{foundkey})
		foundbval := kvpairmdval(id, newincl.MappedPsgByName)
		newincl.ListedPBN = gen.SetSubtraction(newincl.ListedPBN, []string{foundbval})
		delete(newincl.MappedPsgByName, foundkey)
	// PART TWO: EXCLUSIONS
	case "agnexclusions":
		newexcl.AuGenres = removemd(newexcl.AuGenres, id)
	case "wgnexclusions":
		newexcl.WkGenres = removemd(newexcl.WkGenres, id)
	case "alocexclusions":
		newexcl.AuLocations = removemd(newexcl.AuLocations, id)
	case "wlocexclusions":
		newexcl.WkLocations = removemd(newexcl.WkLocations, id)
	case "auexclusions":
		foundkey := kvpairmdkey(id, newexcl.MappedAuthByName)
		newexcl.Authors = gen.SetSubtraction(newexcl.Authors, []string{foundkey})
	case "wkexclusions":
		foundkey := kvpairmdkey(id, newexcl.MappedWkByName)
		newexcl.Works = gen.SetSubtraction(newexcl.Works, []string{foundkey})
	case "psgexclusions":
		foundkey := kvpairmdkey(id, newexcl.MappedPsgByName)
		newexcl.Passages = gen.SetSubtraction(newexcl.Passages, []string{foundkey})
		foundbval := kvpairmdval(id, newexcl.MappedPsgByName)
		newexcl.ListedPBN = gen.SetSubtraction(newexcl.ListedPBN, []string{foundbval})
		delete(newexcl.MappedPsgByName, foundkey)
	default:
		Msg.WARN(fmt.Sprintf(FAIL2, cat))
	}

	newsess.Inclusions = newincl
	newsess.Exclusions = newexcl

	vlt.AllSessions.InsertSess(newsess)
	r := RtSelectionFetch(c)

	//sliceprinter("newincl.Passages", newincl.Passages)
	//stringmapprinter("newincl.MappedPsgByName", newincl.MappedPsgByName)
	return r
}

func RtSelectionFetch(c echo.Context) error {
	sd := reportcurrentselections(c)
	return c.JSONPretty(http.StatusOK, sd, vv.JSONINDENT)
}

// registerselection - do the hard work of parsing a selection
func registerselection(user string, sv SelectionValues) str.ServerSession {
	// have to deal with all sorts of possibilities
	// [a] author: "GET /selection/make/_?auth=gr7000 HTTP/1.1"
	// [b] work: "GET /selection/make/_?auth=lt0474&work=001 HTTP/1.1"
	// [b2] work excluded: "GET /selection/make/_?auth=lt0474&work=037&exclude=t HTTP/1.1"
	// [c] section of work: "GET /selection/make/_?auth=lt0474&work=043&locus=1&endpoint= HTTP/1.1"
	// [c2] subsection of work: "GET /selection/make/_?auth=lt0474&work=037&locus=3|13&endpoint= HTTP/1.1"
	// [c3] span of a work: "GET /selection/make/_?auth=lt0474&work=037&locus=2|100&endpoint=3|20 HTTP/1.1"
	// [d] author genre: "GET /selection/make/_?genre=Alchemistae HTTP/1.1"
	// [e] work genre: "GET /selection/make/_?wkgenre=Apocalyp. HTTP/1.1"
	// [f] author location: "GET /selection/make/_?auloc=Abdera HTTP/1.1"
	// [g] work proven: "GET /selection/make/_?wkprov=Abdera%20(Thrace) HTTP/1.1"

	const (
		PSGT = `%s_FROM_%d_TO_%d`
	)

	s := vlt.AllSessions.GetSess(user)

	sep := "|"
	if s.RawInput {
		sep = "."
	}

	if s.Inclusions.MappedPsgByName == nil {
		s.Inclusions.MappedPsgByName = make(map[string]string)
	}
	if s.Exclusions.MappedPsgByName == nil {
		s.Exclusions.MappedPsgByName = make(map[string]string)
	}

	if sv.A() {
		if !sv.IsExcl {
			s.Inclusions.Authors = gen.Unique(append(s.Inclusions.Authors, sv.Auth))
		} else {
			s.Exclusions.Authors = gen.Unique(append(s.Exclusions.Authors, sv.Auth))
		}
	}

	if sv.AW() {
		if !sv.IsExcl {
			s.Inclusions.Works = gen.Unique(append(s.Inclusions.Works, fmt.Sprintf("%sw%s", sv.Auth, sv.Work)))
		} else {
			s.Exclusions.Works = gen.Unique(append(s.Exclusions.Works, fmt.Sprintf("%sw%s", sv.Auth, sv.Work)))
		}
	}

	if sv.AWP() {
		// [2]int comes back: first and last lines found via the query
		b := findendpointsfromlocus(sv.WUID(), sv.Start, sep)
		r := strings.Replace(sv.Start, "|", ".", -1)
		ra := mps.AllAuthors[sv.Auth].Shortname
		rw := mps.AllWorks[sv.WUID()].Title
		cs := fmt.Sprintf("%s, %s, %s", ra, rw, r)
		i := fmt.Sprintf(PSGT, sv.Auth, b[0], b[1])
		if !sv.IsExcl {
			s.Inclusions.Passages = gen.Unique(append(s.Inclusions.Passages, i))
			s.Inclusions.MappedPsgByName[i] = cs
		} else {
			s.Exclusions.Passages = gen.Unique(append(s.Exclusions.Passages, i))
			s.Exclusions.MappedPsgByName[i] = cs
		}
	}

	if sv.AWPR() {
		// [2]int comes back: first and last lines found via the query
		b := findendpointsfromlocus(sv.WUID(), sv.Start, sep)
		e := findendpointsfromlocus(sv.WUID(), sv.End, sep)
		ra := mps.AllAuthors[sv.Auth].Shortname
		rw := mps.AllWorks[sv.WUID()].Title
		rs := strings.Replace(sv.Start, "|", ".", -1)
		re := strings.Replace(sv.End, "|", ".", -1)
		cs := fmt.Sprintf("%s, %s, %s - %s", ra, rw, rs, re)
		i := fmt.Sprintf(PSGT, sv.Auth, b[0], e[1])
		if !sv.IsExcl {
			s.Inclusions.Passages = gen.Unique(append(s.Inclusions.Passages, i))
			s.Inclusions.MappedPsgByName[i] = cs
		} else {
			s.Exclusions.Passages = gen.Unique(append(s.Exclusions.Passages, i))
			s.Exclusions.MappedPsgByName[i] = cs
		}
	}

	if len(sv.AGenre) != 0 {
		if _, ok := mps.AuGenres[sv.AGenre]; ok {
			if !sv.IsExcl {
				s.Inclusions.AuGenres = gen.Unique(append(s.Inclusions.AuGenres, sv.AGenre))
			} else {
				s.Exclusions.AuGenres = gen.Unique(append(s.Exclusions.AuGenres, sv.AGenre))
			}
		}
	}

	if len(sv.ALoc) != 0 {
		if _, ok := mps.AuLocs[sv.ALoc]; ok {
			if !sv.IsExcl {
				s.Inclusions.AuLocations = gen.Unique(append(s.Inclusions.AuLocations, sv.ALoc))
			} else {
				s.Exclusions.AuLocations = gen.Unique(append(s.Exclusions.AuLocations, sv.ALoc))
			}
		}
	}

	if len(sv.WGenre) != 0 {
		if _, ok := mps.WkGenres[sv.WGenre]; ok {
			if !sv.IsExcl {
				s.Inclusions.WkGenres = gen.Unique(append(s.Inclusions.WkGenres, sv.WGenre))
			} else {
				s.Exclusions.WkGenres = gen.Unique(append(s.Exclusions.WkGenres, sv.WGenre))
			}
		}
	}

	if len(sv.WLoc) != 0 {
		if _, ok := mps.WkLocs[sv.WLoc]; ok {
			if !sv.IsExcl {
				s.Inclusions.WkLocations = gen.Unique(append(s.Inclusions.WkLocations, sv.WLoc))
			} else {
				s.Exclusions.WkLocations = gen.Unique(append(s.Exclusions.WkLocations, sv.WLoc))
			}
		}
	}

	s = rationalizeselections(s, sv)

	return s
}

// rationalizeselections - make sure that A, B, C, ... make sense as a collection of selections
func rationalizeselections(original str.ServerSession, sv SelectionValues) str.ServerSession {
	// if you select "book 2" after selecting the whole, select only book 2
	// if you select the whole after book 2, then the whole
	// etc...

	const (
		PSGT = `%s_FROM_%s_TO_%s`
	)

	rationalized := original

	si := rationalized.Inclusions
	se := rationalized.Exclusions

	// there are clever ways to do this with reflection, but they won't be readable

	if sv.A() && !sv.IsExcl {
		// [a] kick this author from the other column
		var clean []string
		for _, a := range se.Authors {
			if a != sv.Auth {
				clean = append(clean, a)
			}
		}
		se.Authors = clean

		// [b] remove the works from this column
		clean = []string{}
		for _, w := range si.Works {
			if w[0:vv.LENGTHOFAUTHORID] != sv.Auth {
				clean = append(clean, w)
			}
		}
		si.Works = clean

		// [c] remove the passages from this column
		clean = []string{}
		for _, p := range si.Passages {
			if p[0:vv.LENGTHOFAUTHORID] != sv.Auth {
				clean = append(clean, p)
			} else {
				delete(si.MappedPsgByName, p)
			}
		}
		si.Passages = clean
	} else if sv.A() && sv.IsExcl {
		// [a] kick this author from the other column
		var clean []string
		for _, a := range si.Authors {
			if a != sv.Auth {
				clean = append(clean, a)
			}
		}
		si.Authors = clean

		// [b] remove the works from both columns
		clean = []string{}
		for _, w := range si.Works {
			if w[0:vv.LENGTHOFAUTHORID] != sv.Auth {
				clean = append(clean, w)
			}
		}
		si.Works = clean

		clean = []string{}
		for _, w := range se.Works {
			if w[0:vv.LENGTHOFAUTHORID] != sv.Auth {
				clean = append(clean, w)
			}
		}
		se.Works = clean

		// [c] remove the passages from both columns
		clean = []string{}
		for _, p := range si.Passages {
			if p[0:vv.LENGTHOFAUTHORID] != sv.Auth {
				clean = append(clean, p)
			} else {
				delete(si.MappedPsgByName, p)
			}
		}
		si.Passages = clean

		clean = []string{}
		for _, p := range se.Passages {
			if p[0:vv.LENGTHOFAUTHORID] != sv.Auth {
				clean = append(clean, p)
			} else {
				delete(se.MappedPsgByName, p)
			}
		}
		se.Passages = clean
	} else if sv.AW() && !sv.IsExcl {
		// [a] kick this author from both columns
		var clean []string
		for _, a := range si.Authors {
			if a != sv.Auth {
				clean = append(clean, a)
			}
		}
		si.Authors = clean

		clean = []string{}
		for _, a := range se.Authors {
			if a != sv.Auth {
				clean = append(clean, a)
			}
		}
		se.Authors = clean

		// [b] kick this work from the other column
		clean = []string{}
		for _, w := range se.Works {
			if w != sv.Work {
				clean = append(clean, w)
			}
		}
		se.Works = clean

		// [c] remove the passages from this column
		clean = []string{}
		for _, p := range si.Passages {
			if workvalueofpassage(p) != sv.WUID() {
				clean = append(clean, p)
			} else {
				delete(si.MappedPsgByName, p)
			}
		}
		si.Passages = clean
	} else if sv.AW() && sv.IsExcl {
		// [a] kick this author from both columns
		var clean []string
		for _, a := range si.Authors {
			if a != sv.Auth {
				clean = append(clean, a)
			}
		}
		si.Authors = clean

		clean = []string{}
		for _, a := range se.Authors {
			if a != sv.WUID() {
				clean = append(clean, a)
			}
		}
		se.Authors = clean

		// [b] kick this work from the other column
		clean = []string{}
		for _, w := range si.Works {
			if w != sv.WUID() {
				clean = append(clean, w)
			}
		}
		si.Works = clean

		// [c] remove the passages from both columns
		clean = []string{}
		for _, p := range si.Passages {
			if workvalueofpassage(p) != sv.WUID() {
				clean = append(clean, p)
			} else {
				delete(si.MappedPsgByName, p)
			}
		}
		si.Passages = clean

		clean = []string{}
		for _, p := range se.Passages {
			if workvalueofpassage(p) != sv.WUID() {
				clean = append(clean, p)
			} else {
				delete(si.MappedPsgByName, p)
			}
		}
		se.Passages = clean
	} else if sv.AWP() && !sv.IsExcl {
		// [a] kick this author from both columns
		var clean []string
		for _, a := range si.Authors {
			if a != sv.Auth {
				clean = append(clean, a)
			}
		}
		si.Authors = clean

		clean = []string{}
		for _, a := range se.Authors {
			if a != sv.Auth {
				clean = append(clean, a)
			}
		}
		se.Authors = clean

		// [b] kick this work from both columns
		clean = []string{}
		for _, w := range si.Works {
			if w != sv.WUID() {
				clean = append(clean, w)
			}
		}
		si.Works = clean

		clean = []string{}
		for _, w := range se.Works {
			if w != sv.WUID() {
				clean = append(clean, w)
			}
		}
		se.Works = clean

		// [c] kick this passage from the other column
		clean = []string{}
		s := fmt.Sprintf(PSGT, sv.Auth, sv.Start, sv.End)
		for _, p := range se.Passages {
			if p != s {
				clean = append(clean, p)
			} else {
				delete(se.MappedPsgByName, p)
			}
		}
		se.Passages = clean
		// not going to sweat overlapping passages: hard to make them in the first place
	} else if sv.AWP() && sv.IsExcl {
		// [a] kick this author from both columns
		var clean []string
		for _, a := range si.Authors {
			if a != sv.Auth {
				clean = append(clean, a)
			}
		}
		si.Authors = clean

		clean = []string{}
		for _, a := range se.Authors {
			if a != sv.Auth {
				clean = append(clean, a)
			}
		}
		se.Authors = clean

		// [b] kick this work from this column
		clean = []string{}
		for _, w := range se.Works {
			if w != sv.WUID() {
				clean = append(clean, w)
			}
		}
		se.Works = clean

		// [c] kick this passage from the other column
		clean = []string{}
		s := fmt.Sprintf(PSGT, sv.Auth, sv.Start, sv.End)
		for _, p := range si.Passages {
			if p != s {
				clean = append(clean, p)
			} else {
				delete(si.MappedPsgByName, p)
			}
		}
		si.Passages = clean
	}

	rationalized.Inclusions = si
	rationalized.Exclusions = se

	return rationalized
}

// workvalueofpassage - what work does "lt0474_FROM_58578_TO_61085" come from?
func workvalueofpassage(psg string) string {
	pattern := regexp.MustCompile(`(?P<auth>......)_FROM_(?P<start>\d+)_TO_(?P<stop>\d+)`)
	// "gr0032_FROM_11313_TO_11843"
	subs := pattern.FindStringSubmatch(psg)
	au := subs[pattern.SubexpIndex("auth")]
	st, _ := strconv.Atoi(subs[pattern.SubexpIndex("start")])
	sp, _ := strconv.Atoi(subs[pattern.SubexpIndex("stop")])
	thework := ""
	for _, w := range mps.AllAuthors[au].WorkList {
		ws := mps.AllWorks[w].FirstLine
		we := mps.AllWorks[w].LastLine
		if st >= ws && sp <= we {
			thework = w
		}
	}

	return thework
}

// findendpointsfromlocus - given a locus, what index values correspond to the start and end of that text segment?
func findendpointsfromlocus(wuid string, locus string, sep string) [2]int {
	// we are wrapping endpointer() to give us a couple of bites at perseus citaiton problems

	const (
		MSG = "findendpointsfromlocus() retrying endpointer(): '%s' --> '%s'"
	)

	fl, success := endpointer(wuid, locus, sep)
	if success || sep != ":" {
		return fl
	}

	dc := regexp.MustCompile("(\\d+)([a-f])$")
	if dc.MatchString(locus) {
		// plato, et al
		// [HGS] endpointer() failed to find the following inside of gr0059w030: '407e'
		// [HGS] findendpointsfromlocus() retrying endpointer(): '407e' --> '407:e'
		r := fmt.Sprintf("$1%s$2", sep)
		newlocus := dc.ReplaceAllString(locus, r)
		Msg.PEEK(fmt.Sprintf(MSG, locus, newlocus))
		fl, success = endpointer(wuid, newlocus, sep)
	} else {
		// cicero, et.al
		// [a] [HGS] findendpointsfromlocus() failed to find the following inside of lt0474w049: 4:8:18
		// this should in fact be "4.18"
		// [b] BUT in lt0474w024 "10:24" you want "24"
		ll := strings.Split(locus, sep)
		if len(ll) > 2 {
			newlocus := strings.Join(gen.RemoveIndex(ll, 1), ":")
			Msg.PEEK(fmt.Sprintf(MSG, locus, newlocus))
			fl, success = endpointer(wuid, newlocus, sep)
		} else if len(ll) == 2 {
			newlocus := strings.Join(gen.RemoveIndex(ll, 0), ":")
			Msg.PEEK(fmt.Sprintf(MSG, locus, newlocus))
			fl, success = endpointer(wuid, newlocus, sep)
		}
	}

	return fl
}

// endpinter - given a locus, what index values correspond to the start and end of that text segment?
func endpointer(wuid string, locus string, sep string) ([2]int, bool) {
	// mm(fmt.Sprintf("wuid: '%s'; locus: '%s'; sep: '%s'", wuid, locus, sep), 1)
	// [HGS] wuid: 'lt0474w049'; locus: '3|14|_0'; sep: '|'
	// [HGS] wuid: 'lt0474w049'; locus: '4:8:18'; sep: ':'

	const (
		QTMP = `SELECT index FROM %s WHERE wkuniversalid='%s' AND %s ORDER BY index ASC`
		FAIL = "endpointer() failed to find the following inside of %s: '%s'"
		WNFD = "endpointer() failed to find a work: %s"
	)

	fl := [2]int{0, 0}
	success := false

	// dictionary click inside 'τάλαντον' at end of first segment: "...δίκαϲ ῥέπει τάλαντον Bacchylides 17.25."
	// error 500: /browse/perseus/gr0199/002/17:25
	// but there is no work 002; the numbers start at 010

	wk := validateworkselection(wuid)
	if wk.UID == "work_not_found" {
		Msg.FYI(fmt.Sprintf(WNFD, wuid))
		return fl, false
	}

	wl := wk.CountLevels()
	ll := strings.Split(locus, sep)
	if len(ll) > wl {
		ll = ll[0:wl]
	}

	if len(ll) == 0 || ll[0] == "_0" {
		fl = [2]int{wk.FirstLine, wk.LastLine}
		return fl, true
	}

	if ll[len(ll)-1] == "_0" {
		ll = ll[0 : len(ll)-1]
	}

	col := []string{"level_00_value", "level_01_value", "level_02_value", "level_03_value", "level_04_value", "level_05_value"}
	tem := `%s='%s'`
	var use []string
	for i, l := range ll {
		s := fmt.Sprintf(tem, col[wl-i-1], l)
		use = append(use, s)
	}

	tb := wk.AuID()

	a := strings.Join(use, " AND ")
	q := fmt.Sprintf(QTMP, tb, wuid, a)

	foundrows, err := db.SQLPool.Query(context.Background(), q)
	Msg.EC(err)

	idx, err := pgx.CollectRows(foundrows, pgx.RowTo[int])
	Msg.EC(err)

	if len(idx) == 0 {
		// bogus input
		Msg.PEEK(fmt.Sprintf(FAIL, wuid, locus))
		fl = [2]int{1, 1}
	} else {
		fl = [2]int{idx[0], idx[len(idx)-1]}
		success = true
	}

	return fl, success
}

// SelectionData - JS output struct
type SelectionData struct {
	TimeRestr string `json:"timeexclusions"`
	Select    string `json:"selections"`
	Exclude   string `json:"exclusions"`
	NewJS     string `json:"newjs"`
	Count     int    `json:"numberofselections"`
}

// reportcurrentselections - prepare JSON for the page re. current selections
func reportcurrentselections(c echo.Context) SelectionData {
	// ultimately feeding autocomplete.js
	//    $('#timerestrictions').html(selectiondata.timeexclusions);
	//    $('#selectioninfocell').html(selectiondata.selections);
	//    $('#exclusioninfocell').html(selectiondata.exclusions);
	//    $('#selectionscriptholder').html(selectiondata['newjs']);

	const (
		PL    = "<span class=\"picklabel\">%s</span><br>\n"
		SL    = "\t\t\t\t\t<span class=\"%sselections selection\" id=\"%sselections_%s\" title=\"Double-click to remove this item\">%s</span><br>\n"
		EL    = "\t\t\t\t\t<span class=\"%ssexclusions selection\" id=\"%sexclusions_%s\" title=\"Double-click to remove this item\">%s</span><br>\n"
		TL    = `Unless specifically listed, authors/works must come from %s to %s`
		JSIN  = `%sselections_%s`
		JSINU = `/selection/clear/%sselections/%s`
		JSEX  = `%sexclusions_%s`
		JSEXU = `/selection/clear/%sexclusions/%s`
	)

	user := vlt.ReadUUIDCookie(c)
	s := vlt.AllSessions.GetSess(user)
	search.BuildSelectionOverview(&s)

	i := s.Inclusions
	e := s.Exclusions

	// need to do it in this order: don't walk through the map keys
	cat := []string{"agn", "wgn", "aloc", "wloc", "au", "wk", "psg"}
	catmap := map[string][2]string{
		"agn":  {"Author categories", "AuGenres"},
		"wgn":  {"Work genres", "WkGenres"},
		"aloc": {"Author location", "AuLocations"},
		"wloc": {"Work provenance", "WkLocations"},
		"au":   {"Authors", "ListedABN"},
		"wk":   {"Works", "ListedWBN"},
		"psg":  {"Passages", "ListedPBN"},
	}

	var jsinfo []JSData

	var sd SelectionData

	// note that there can be a problem coordinating the selections with the add/drop clicks
	// this is a list + order-of-arrival problem at its core: was #1 on the reported list #1 to be selected?
	// use md5 fingerprints to circumvent this sort of bookkeeping (by adding a different kind...)

	// problem: is it possible to get two matching values on these lists: 'A... A... B...'?
	// probably: there are multiple 'Diodorus' entries; by AU you will get 'Diodours (Eleg.)'; by WK you will get 'Diodorus'
	// there is the possibility of a collision between 'Diodorus, Fragmenta' and 'Diodorus, Fragmenta' (with somebody or other...)
	// kind of hard to produce this; and the code will let you click-kill one of them, but maybe the wrong one...

	// just cicero selected:
	// {[] [] [] [] [lt0474] [] [] map[] map[lt0474:Cicero - Cicero, Marcus Tullius] map[] [] [Cicero - Cicero, Marcus Tullius] []}
	// 'Cicero - Cicero, Marcus Tullius' --> f6de296b2941374db110d2d3683e8ca9

	// one passage:
	// {[] [] [] [] [] [] [lt0959_FROM_26820_TO_27569] map[lt0959_FROM_26820_TO_27569:Ovidius, Publius Naso, Tristia, 1] map[] map[] [Ovidius, Publius Naso, Tristia, 1] [] []}
	// 'Ovidius, Publius Naso, Tristia, 1' --> 118ecd1993e1d5316cc3ed2bfb3d486a

	// run inclusions, then exclusions
	var rows [2][]string
	swap := [2]string{SL, EL}
	for idx, v := range [2]str.SearchIncExl{i, e} {
		for _, ct := range cat {
			label := catmap[ct][0]
			using := catmap[ct][1]
			val := reflect.ValueOf(&v).Elem().FieldByName(using)
			// PITA to cast a Value to its type: https://stackoverflow.com/questions/17262238/how-to-cast-reflect-value-to-its-type
			// note that the next is terrible if we are not 100% sure of the interface
			slc := val.Interface().([]string)
			if len(slc) > 0 {
				rows[idx] = append(rows[idx], fmt.Sprintf(PL, label))
				for _, g := range slc {
					// fingerprint the selection so the JS can click to drop "f6de296b2941374db110d2d3683e8ca9", vel sim
					md := fmt.Sprintf("%x", md5.Sum([]byte(g)))
					// mm(fmt.Sprintf("g: %s\tmd: %s", g, md), 1)
					st := fmt.Sprintf(swap[idx], ct, ct, md, g)
					rows[idx] = append(rows[idx], st)
					if swap[idx] == SL {
						a := fmt.Sprintf(JSIN, ct, md)
						b := fmt.Sprintf(JSINU, ct, md)
						jsinfo = append(jsinfo, JSData{a, b})
					} else {
						a := fmt.Sprintf(JSEX, ct, md)
						b := fmt.Sprintf(JSEXU, ct, md)
						jsinfo = append(jsinfo, JSData{a, b})
					}
				}
			}
		}
	}

	mustshow := false

	if s.Earliest != vv.MINDATESTR || s.Latest != vv.MAXDATESTR {
		mustshow = true
		sd.TimeRestr = fmt.Sprintf(TL, gen.FormatBCEDate(s.Earliest), gen.FormatBCEDate(s.Latest))
	}

	sd.Select = strings.Join(rows[0], "")
	sd.Exclude = strings.Join(rows[1], "")
	sd.Count = i.CountItems() + e.CountItems()
	if sd.Count == 0 && mustshow {
		// no author selections, but date selections
		// force the JS to do: $('#selectionstable').show();
		sd.Count = 1
	}
	sd.NewJS = formatnewselectionjs(jsinfo)

	return sd
}

// formatnewselectionjs - prepare the JS that the client needs in order to report the current selections
func formatnewselectionjs(jsinfo []JSData) string {
	const (
		JSFNC = `
		$( '#%s' ).dblclick(function() {
			$.getJSON('%s', function (selectiondata) { 
				reloadselections(selectiondata); });
		});`

		SCR = `
		<script>%s
		</script>
		`
	)
	if len(jsinfo) == 0 {
		return ""
	}

	var info []string
	for _, j := range jsinfo {
		info = append(info, fmt.Sprintf(JSFNC, j.Pound, j.Url))
	}

	script := fmt.Sprintf(SCR, strings.Join(info, ""))
	return script
}

// validateworkselection - what if you request a work that does not exist? return something...
func validateworkselection(uid string) *str.DbWork {
	w := &str.DbWork{}
	w.UID = "work_not_found"
	au := uid[0:vv.LENGTHOFAUTHORID]
	if _, ok := mps.AllWorks[uid]; ok {
		w = mps.AllWorks[uid]
	} else {
		if _, y := mps.AllAuthors[au]; y {
			// firstwork; otherwise we are still set to "null"
			w = mps.AllWorks[mps.AllAuthors[au].WorkList[0]]
		}
	}
	return w
}

// A LIST

//Author categories
//	Epici
//	Medici
//Work genres
//	Epist.
//	Eleg.
//Authors
//	Cicero - Cicero, Marcus Tullius
//	Ovidius, Publius Naso
//Passages
//	Caesar, De Bello Gallico, 3
//	Livy, Ab Urbe Condita, 6.10 - 8.15
//
//

// THE TABLE

//<table id="selectionstable" style="">
//            <tbody>
//            <tr>
//                <th colspan="5" id="timerestrictions"></th>
//            </tr>
//            <tr>
//                <td class="infocells" id="selectioninfocell" width="44%" title="Selection list"><span class="picklabel">Author categories</span><br>
//                    <span class="agnselections selection" id="agnselections_51b9ec70f5659ee042d6ee610b74887e" title="Double-click to remove this item">Epici</span><br>
//                    <span class="agnselections selection" id="agnselections_4c1401171d5425c227b8eb7f0ab46440" title="Double-click to remove this item">Medici</span><br>
//<span class="picklabel">Work genres</span><br>
//                    <span class="wgnselections selection" id="wgnselections_682242d3e04f7dc5d887171dfd4ba7cf" title="Double-click to remove this item">Epist.</span><br>
//                    <span class="wgnselections selection" id="wgnselections_b1e0ecef33ae5441b6c9047396ef48fb" title="Double-click to remove this item">Eleg.</span><br>
//<span class="picklabel">Authors</span><br>
//                    <span class="auselections selection" id="auselections_f6de296b2941374db110d2d3683e8ca9" title="Double-click to remove this item">Cicero - Cicero, Marcus Tullius</span><br>
//                    <span class="auselections selection" id="auselections_a3312f035444f5c879c0322519c9d094" title="Double-click to remove this item">Ovidius, Publius Naso</span><br>
//<span class="picklabel">Passages</span><br>
//                    <span class="psgselections selection" id="psgselections_6c55264a3d6e8c2fc715e92e61a5af4e" title="Double-click to remove this item">Caesar, De Bello Gallico, 3</span><br>
//                    <span class="psgselections selection" id="psgselections_558de47b1b32c56600041134858b48fc" title="Double-click to remove this item">Livy, Ab Urbe Condita, 6.10 - 8.15</span><br>
//</td>
//                <td style="text-align: center;" id="jscriptwigetcell" width="6%">
//                    <p id="searchinfo"><span class="material-icons md-mid" title="Show/hide details of the current search list">info</span></p>
//                </td>
//                <td class="infocellx" id="exclusioninfocell" width="44%" title="Exclusion list"></td>
//                <td style="text-align: center;" width="6%">
//                    <p id="textofthis"><span class="material-icons md-mid" title="Generate a simple text of this selection">library_books</span></p>
//                    <p id="makeanindex"><span class="material-icons md-mid" title="Build an index to this selection">subject</span></p>
//                    <p id="makevocablist"><span class="material-icons md-mid" title="Build a vocabulary list for this selection">format_list_numbered</span></p>
//                </td>
//            </tr>
//            </tbody>
//        </table>
