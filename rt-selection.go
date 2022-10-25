//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
	"fmt"
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

	user := readUUIDCookie(c)

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
	SafeSessionMapInsert(ns)

	cs := reportcurrentselections(c)

	return c.JSONPretty(http.StatusOK, cs, JSONINDENT)
}

// RtSelectionClear - remove a selection from the session
func RtSelectionClear(c echo.Context) error {
	// sample item in the "selectionstable"; the js that activates it is inside the "selectionscriptholder" div
	// <span class="wkselections selection" id="wkselections_00" title="Double-click to remove this item">Antigonus, <i>Historiarum mirabilium collectio</i></span>

	const (
		FAIL1 = "RtSelectionClear() was given bad input: %s"
		FAIL2 = "RtSelectionClear() was given bad category: %s"
		REG   = `(?P<auth>......)_FROM_(?P<start>\d+)_TO_(?P<stop>\d+)`
	)

	user := readUUIDCookie(c)

	locus := c.Param("locus")
	which := strings.Split(locus, "/")

	if len(which) != 2 {
		msg(fmt.Sprintf(FAIL1, locus), 1)
		return emptyjsreturn(c)
	}

	cat := which[0]
	id, e := strconv.Atoi(which[1])
	if e != nil {
		msg(fmt.Sprintf(FAIL1, locus), 1)
		return emptyjsreturn(c)
	}

	// cat := []string{"agn", "wgn", "aloc", "wloc", "au", "wk", "psg"}

	newsess := SafeSessionRead(user)
	newincl := newsess.Inclusions
	newexcl := newsess.Exclusions

	// sliceprinter("ListedPBN", newincl.ListedPBN)

	findkey := func(n int, ixl SearchIncExl) string {
		// issue: newincl.Passages is a []string that is built by order of arrival, but the display is ordered by AU & LINE
		// this means that if "book 1" and "book 3" are on the list, but you added "book 3" first, clicking "book 1" will
		// remove item #0 from the list, and drop... "book 3"

		// ListedPBN is the page order and name
		// its index value is the click id value
		// so, fetch the string from ListedPBN[id]
		// then search for that string in the values of MappedPsgByName

		ixl.BuildPsgByName()
		citation := ixl.ListedPBN[n]
		thekey := ""
		for k, v := range ixl.MappedPsgByName {
			if v == citation {
				thekey = k
				break
			}
		}
		return thekey
	}

	switch cat {
	case "agnselections":
		newincl.AuGenres = RemoveIndex(newincl.AuGenres, id)
	case "wgnselections":
		newincl.WkGenres = RemoveIndex(newincl.WkGenres, id)
	case "alocselections":
		newincl.AuLocations = RemoveIndex(newincl.AuLocations, id)
	case "wlocselections":
		newincl.WkLocations = RemoveIndex(newincl.WkLocations, id)
	case "auselections":
		newincl.Authors = RemoveIndex(newincl.Authors, id)
	case "wkselections":
		newincl.Works = RemoveIndex(newincl.Works, id)
	case "psgselections":
		// NB: restarting the server with an open browser can leave an impossible click; not really a bug, but...
		del := findkey(id, newincl)
		newincl.Passages = setsubtraction(newincl.Passages, []string{del})
		delete(newincl.MappedPsgByName, del)
	case "agnexclusions":
		newexcl.AuGenres = RemoveIndex(newexcl.AuGenres, id)
	case "wgnexclusions":
		newexcl.WkGenres = RemoveIndex(newexcl.WkGenres, id)
	case "alocexclusions":
		newexcl.AuLocations = RemoveIndex(newexcl.AuLocations, id)
	case "wlocexclusions":
		newexcl.WkLocations = RemoveIndex(newexcl.WkLocations, id)
	case "auexclusions":
		newexcl.Authors = RemoveIndex(newexcl.Authors, id)
	case "wkexclusions":
		newexcl.Works = RemoveIndex(newexcl.Works, id)
	case "psgexclusions":
		del := findkey(id, newexcl)
		newexcl.Passages = setsubtraction(newexcl.Passages, []string{del})
		delete(newexcl.MappedPsgByName, del)
	default:
		msg(fmt.Sprintf(FAIL2, cat), 1)
	}

	newsess.Inclusions = newincl
	newsess.Exclusions = newexcl

	SafeSessionMapInsert(newsess)
	r := RtSelectionFetch(c)

	//sliceprinter("newincl.Passages", newincl.Passages)
	//stringmapprinter("newincl.MappedPsgByName", newincl.MappedPsgByName)
	return r
}

func RtSelectionFetch(c echo.Context) error {
	sd := reportcurrentselections(c)
	return c.JSONPretty(http.StatusOK, sd, JSONINDENT)
}

// registerselection - do the hard work of parsing a selection
func registerselection(user string, sv SelectionValues) ServerSession {
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

	s := SafeSessionRead(user)

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
			s.Inclusions.Authors = unique(append(s.Inclusions.Authors, sv.Auth))
		} else {
			s.Exclusions.Authors = unique(append(s.Exclusions.Authors, sv.Auth))
		}
	}

	if sv.AW() {
		if !sv.IsExcl {
			s.Inclusions.Works = unique(append(s.Inclusions.Works, fmt.Sprintf("%sw%s", sv.Auth, sv.Work)))
		} else {
			s.Exclusions.Works = unique(append(s.Exclusions.Works, fmt.Sprintf("%sw%s", sv.Auth, sv.Work)))
		}
	}

	if sv.AWP() {
		// [2]int64 comes back: first and last lines found via the query
		b := findendpointsfromlocus(sv.WUID(), sv.Start, sep)
		r := strings.Replace(sv.Start, "|", ".", -1)
		ra := AllAuthors[sv.Auth].Shortname
		rw := AllWorks[sv.WUID()].Title
		cs := fmt.Sprintf("%s, %s, %s", ra, rw, r)
		i := fmt.Sprintf(PSGT, sv.Auth, b[0], b[1])
		if !sv.IsExcl {
			s.Inclusions.Passages = unique(append(s.Inclusions.Passages, i))
			s.Inclusions.MappedPsgByName[i] = cs
		} else {
			s.Exclusions.Passages = unique(append(s.Exclusions.Passages, i))
			s.Exclusions.MappedPsgByName[i] = cs
		}
	}

	if sv.AWPR() {
		// [2]int64 comes back: first and last lines found via the query
		b := findendpointsfromlocus(sv.WUID(), sv.Start, sep)
		e := findendpointsfromlocus(sv.WUID(), sv.End, sep)
		ra := AllAuthors[sv.Auth].Shortname
		rw := AllWorks[sv.WUID()].Title
		rs := strings.Replace(sv.Start, "|", ".", -1)
		re := strings.Replace(sv.End, "|", ".", -1)
		cs := fmt.Sprintf("%s, %s, %s - %s", ra, rw, rs, re)
		i := fmt.Sprintf(PSGT, sv.Auth, b[0], e[1])
		if !sv.IsExcl {
			s.Inclusions.Passages = unique(append(s.Inclusions.Passages, i))
			s.Inclusions.MappedPsgByName[i] = cs
		} else {
			s.Exclusions.Passages = unique(append(s.Exclusions.Passages, i))
			s.Exclusions.MappedPsgByName[i] = cs
		}
	}

	if len(sv.AGenre) != 0 {
		if _, ok := AuGenres[sv.AGenre]; ok {
			if !sv.IsExcl {
				s.Inclusions.AuGenres = unique(append(s.Inclusions.AuGenres, sv.AGenre))
			} else {
				s.Exclusions.AuGenres = unique(append(s.Exclusions.AuGenres, sv.AGenre))
			}
		}
	}

	if len(sv.ALoc) != 0 {
		if _, ok := AuLocs[sv.ALoc]; ok {
			if !sv.IsExcl {
				s.Inclusions.AuLocations = unique(append(s.Inclusions.AuLocations, sv.ALoc))
			} else {
				s.Exclusions.AuLocations = unique(append(s.Exclusions.AuLocations, sv.ALoc))
			}
		}
	}

	if len(sv.WGenre) != 0 {
		if _, ok := WkGenres[sv.WGenre]; ok {
			if !sv.IsExcl {
				s.Inclusions.WkGenres = unique(append(s.Inclusions.WkGenres, sv.WGenre))
			} else {
				s.Exclusions.WkGenres = unique(append(s.Exclusions.WkGenres, sv.WGenre))
			}
		}
	}

	if len(sv.WLoc) != 0 {
		if _, ok := WkLocs[sv.WLoc]; ok {
			if !sv.IsExcl {
				s.Inclusions.WkLocations = unique(append(s.Inclusions.WkLocations, sv.WLoc))
			} else {
				s.Exclusions.WkLocations = unique(append(s.Exclusions.WkLocations, sv.WLoc))
			}
		}
	}

	s = rationalizeselections(s, sv)

	return s
}

// rationalizeselections - make sure that A, B, C, ... make sense as a collection of selections
func rationalizeselections(original ServerSession, sv SelectionValues) ServerSession {
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
			if w[0:6] != sv.Auth {
				clean = append(clean, w)
			}
		}
		si.Works = clean

		// [c] remove the passages from this column
		clean = []string{}
		for _, p := range si.Passages {
			if p[0:6] != sv.Auth {
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
			if w[0:6] != sv.Auth {
				clean = append(clean, w)
			}
		}
		si.Works = clean

		clean = []string{}
		for _, w := range se.Works {
			if w[0:6] != sv.Auth {
				clean = append(clean, w)
			}
		}
		se.Works = clean

		// [c] remove the passages from both columns
		clean = []string{}
		for _, p := range si.Passages {
			if p[0:6] != sv.Auth {
				clean = append(clean, p)
			} else {
				delete(si.MappedPsgByName, p)
			}
		}
		si.Passages = clean

		clean = []string{}
		for _, p := range se.Passages {
			if p[0:6] != sv.Auth {
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
	for _, w := range AllAuthors[au].WorkList {
		ws := AllWorks[w].FirstLine
		we := AllWorks[w].LastLine
		if int64(st) >= ws && int64(sp) <= we {
			thework = w
		}
	}

	return thework
}

// findendpointsfromlocus - given a locus, what index values correspond to the start and end of that text segment?
func findendpointsfromlocus(wuid string, locus string, sep string) [2]int64 {
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
		msg(fmt.Sprintf(MSG, locus, newlocus), 3)
		fl, success = endpointer(wuid, newlocus, sep)
	} else {
		// cicero, et.al
		// [HGS] findendpointsfromlocus() failed to find the following inside of lt0474w049: 4:8:18
		// this should in fact be "4.18"
		ll := strings.Split(locus, sep)
		if len(ll) >= 2 {
			newlocus := strings.Join(RemoveIndex(ll, 1), ":")
			msg(fmt.Sprintf(MSG, locus, newlocus), 3)
			fl, success = endpointer(wuid, newlocus, sep)
		}
	}

	return fl
}

// endpinter - given a locus, what index values correspond to the start and end of that text segment?
func endpointer(wuid string, locus string, sep string) ([2]int64, bool) {
	// msg(fmt.Sprintf("wuid: '%s'; locus: '%s'; sep: '%s'", wuid, locus, sep), 1)
	// [HGS] wuid: 'lt0474w049'; locus: '3|14|_0'; sep: '|'
	// [HGS] wuid: 'lt0474w049'; locus: '4:8:18'; sep: ':'

	const (
		QTMP = `SELECT index FROM %s WHERE wkuniversalid='%s' AND %s ORDER BY index ASC`
		FAIL = "endpointer() failed to find the following inside of %s: '%s'"
	)

	success := false
	fl := [2]int64{0, 0}
	wk := AllWorks[wuid]

	wl := wk.CountLevels()
	ll := strings.Split(locus, sep)
	if len(ll) > wl {
		ll = ll[0:wl]
	}

	if len(ll) == 0 || ll[0] == "_0" {
		fl = [2]int64{wk.FirstLine, wk.LastLine}
		return fl, true
	}

	if ll[len(ll)-1] == "_0" {
		ll = ll[0 : len(ll)-1]
	}

	col := [6]string{"level_00_value", "level_01_value", "level_02_value", "level_03_value", "level_04_value", "level_05_value"}
	tem := `%s='%s'`
	var use []string
	for i, l := range ll {
		s := fmt.Sprintf(tem, col[wl-i-1], l)
		use = append(use, s)
	}

	tb := wk.AuID()

	dbconn := GetPSQLconnection()
	defer dbconn.Release()

	a := strings.Join(use, " AND ")
	q := fmt.Sprintf(QTMP, tb, wuid, a)

	foundrows, err := dbconn.Query(context.Background(), q)
	chke(err)

	var idx []int64

	defer foundrows.Close()
	for foundrows.Next() {
		var thehit int64
		err := foundrows.Scan(&thehit)
		chke(err)
		idx = append(idx, thehit)
	}
	if len(idx) == 0 {
		// bogus input
		msg(fmt.Sprintf(FAIL, wuid, locus), 3)
		fl = [2]int64{1, 1}
	} else {
		fl = [2]int64{idx[0], idx[len(idx)-1]}
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
		PL    = `<span class="picklabel">%s</span><br>`
		SL    = `<span class="%sselections selection" id="%sselections_%02d" title="Double-click to remove this item">%s</span><br>`
		EL    = `<span class="%ssexclusions selection" id="%sexclusions_%02d" title="Double-click to remove this item">%s</span><br>`
		TL    = `Unless specifically listed, authors/works must come from %s to %s`
		JSIN  = `%sselections_%02d`
		JSINU = `/selection/clear/%sselections/%d`
		JSEX  = `%sexclusions_%02d`
		JSEXU = `/selection/clear/%sexclusions/%d`
	)

	user := readUUIDCookie(c)
	s := SafeSessionRead(user)
	s.Inclusions.BuildAuByName()
	s.Exclusions.BuildAuByName()
	s.Inclusions.BuildWkByName()
	s.Exclusions.BuildWkByName()
	s.Inclusions.BuildPsgByName()
	s.Exclusions.BuildPsgByName()

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

	// run inclusions, then exclusions
	var rows [2][]string
	swap := [2]string{SL, EL}
	for idx, v := range [2]SearchIncExl{i, e} {
		for _, ct := range cat {
			label := catmap[ct][0]
			using := catmap[ct][1]
			val := reflect.ValueOf(&v).Elem().FieldByName(using)
			// PITA to cast a Value to its type: https://stackoverflow.com/questions/17262238/how-to-cast-reflect-value-to-its-type
			// note that the next is terrible if we are not 100% sure of the interface
			slc := val.Interface().([]string)
			if len(slc) > 0 {
				rows[idx] = append(rows[idx], fmt.Sprintf(PL, label))
				for n, g := range slc {
					st := fmt.Sprintf(swap[idx], ct, ct, n, g)
					rows[idx] = append(rows[idx], st)
					if swap[idx] == SL {
						a := fmt.Sprintf(JSIN, ct, n)
						b := fmt.Sprintf(JSINU, ct, n)
						jsinfo = append(jsinfo, JSData{a, b})
					} else {
						a := fmt.Sprintf(JSEX, ct, n)
						b := fmt.Sprintf(JSEXU, ct, n)
						jsinfo = append(jsinfo, JSData{a, b})
					}
				}
			}
		}
	}

	mustshow := false

	if s.Earliest != MINDATESTR || s.Latest != MAXDATESTR {
		mustshow = true
		sd.TimeRestr = fmt.Sprintf(TL, formatbcedate(s.Earliest), formatbcedate(s.Latest))
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
