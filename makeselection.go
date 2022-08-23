package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type SelectValues struct {
	Auth   string
	Work   string
	Start  string
	End    string
	AGenre string
	WGenre string
	ALoc   string
	WLoc   string
	Excl   bool
}

// WUID - return work universalid
func (s SelectValues) WUID() string {
	return s.Auth + "w" + s.Work
}

// AWPR - author, work, passage start, passage end
func (s SelectValues) AWPR() bool {
	if len(s.Auth) > 0 && len(s.Work) > 0 && len(s.Start) > 0 && len(s.End) > 0 {
		return true
	} else {
		return false
	}
}

// AWP - author, work, and passage
func (s SelectValues) AWP() bool {
	if len(s.Auth) > 0 && len(s.Work) > 0 && len(s.Start) > 0 && len(s.End) == 0 {
		return true
	} else {
		return false
	}
}

// AW - author, work, and not passage
func (s SelectValues) AW() bool {
	if len(s.Auth) > 0 && len(s.Work) > 0 && len(s.Start) == 0 && len(s.End) == 0 {
		return true
	} else {
		return false
	}
}

// A - author, not work, and not passage
func (s SelectValues) A() bool {
	if len(s.Auth) > 0 && len(s.Work) == 0 && len(s.Start) == 0 && len(s.End) == 0 {
		return true
	} else {
		return false
	}
}

func selected(sv SelectValues, s Session) Session {

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

	if s.Inclusions.PassagesByName == nil {
		s.Inclusions.PassagesByName = make(map[string]string)
	}
	if s.Exclusions.PassagesByName == nil {
		s.Exclusions.PassagesByName = make(map[string]string)
	}

	if sv.A() {
		if !sv.Excl {
			s.Inclusions.Authors = unique(append(s.Inclusions.Authors, sv.Auth))
		} else {
			s.Exclusions.Authors = unique(append(s.Exclusions.Authors, sv.Auth))
		}
	}

	if sv.AW() {
		if !sv.Excl {
			s.Inclusions.Works = unique(append(s.Inclusions.Works, fmt.Sprintf("%sw%s", sv.Auth, sv.Work)))
		} else {
			s.Exclusions.Works = unique(append(s.Exclusions.Works, fmt.Sprintf("%sw%s", sv.Auth, sv.Work)))
		}
	}

	if sv.AWP() {
		// [2]int64 comes back: first and last lines found via the query
		b := findendpointsfromlocus(sv.WUID(), sv.Start)
		r := strings.Replace(sv.Start, "|", ".", -1)
		ra := AllAuthors[sv.Auth].Shortname
		rw := AllWorks[sv.WUID()].Title
		cs := fmt.Sprintf("%s, %s, %s", ra, rw, r)
		t := `%s_FROM_%d_TO_%d`
		i := fmt.Sprintf(t, sv.Auth, b[0], b[1])
		if !sv.Excl {
			s.Inclusions.Passages = unique(append(s.Inclusions.Passages, i))
			s.Inclusions.PassagesByName[i] = cs
		} else {
			s.Exclusions.Passages = unique(append(s.Exclusions.Passages, i))
			s.Exclusions.PassagesByName[i] = cs
		}
	}

	if sv.AWPR() {
		// [2]int64 comes back: first and last lines found via the query
		b := findendpointsfromlocus(sv.WUID(), sv.Start)
		e := findendpointsfromlocus(sv.WUID(), sv.End)
		ra := AllAuthors[sv.Auth].Shortname
		rw := AllWorks[sv.WUID()].Title
		rs := strings.Replace(sv.Start, "|", ".", -1)
		re := strings.Replace(sv.End, "|", ".", -1)
		cs := fmt.Sprintf("%s, %s, %s - %s", ra, rw, rs, re)
		t := `%s_FROM_%d_TO_%d`
		i := fmt.Sprintf(t, sv.Auth, b[0], e[1])
		if !sv.Excl {
			s.Inclusions.Passages = unique(append(s.Inclusions.Passages, i))
			s.Inclusions.PassagesByName[i] = cs
		} else {
			s.Exclusions.Passages = unique(append(s.Exclusions.Passages, i))
			s.Exclusions.PassagesByName[i] = cs
		}
	}

	// TODO: genres, etc

	return s
}

func parsesleectvals(r *http.Request) SelectValues {
	// https://golangcode.com/get-a-url-parameter-from-a-request/
	// https://stackoverflow.com/questions/41279297/how-to-get-all-query-parameters-from-go-gin-context-object
	// gin: You should be able to do c.Request.URL.Query() which will return a Values which is a map[string][]string

	// TODO: check this stuff for bad characters
	// but 'auth', etc. can be parsed just by checking them against known author lists

	var sv SelectValues

	kvp := r.URL.Query() // map[string][]string

	if _, ok := kvp["auth"]; ok {
		sv.Auth = kvp["auth"][0]
	}

	if _, ok := kvp["work"]; ok {
		sv.Work = kvp["work"][0]
	}

	if _, ok := kvp["locus"]; ok {
		sv.Start = kvp["locus"][0]
	}

	if _, ok := kvp["endpoint"]; ok {
		sv.End = kvp["endpoint"][0]
	}

	if _, ok := kvp["genre"]; ok {
		sv.AGenre = kvp["genre"][0]
	}

	if _, ok := kvp["wkgenre"]; ok {
		sv.WGenre = kvp["wkgenre"][0]
	}

	if _, ok := kvp["auloc"]; ok {
		sv.ALoc = kvp["auloc"][0]
	}

	if _, ok := kvp["wkprov"]; ok {
		sv.WLoc = kvp["wkprov"][0]
	}

	if _, ok := kvp["exclude"]; ok {
		if kvp["exclude"][0] == "t" {
			sv.Excl = true
		} else {
			sv.Excl = false
		}
	}

	return sv
}

func rationalizeselections() {
	// if you select "book 2" after selecting the whole, select only book 2
	// if you select the whole after book 2, then the whole
	// etc...

}

func findendpointsfromlocus(wuid string, locus string) [2]int64 {
	fl := [2]int64{0, 0}
	wk := AllWorks[wuid]

	wl := wk.CountLevels()
	ll := strings.Split(locus, "|")
	if len(ll) > wl {
		ll = ll[0:wl]
	}

	col := [6]string{"level_00_value", "level_01_value", "level_02_value", "level_03_value", "level_04_value", "level_05_value"}
	tem := `%s='%s'`
	var use []string
	for i, l := range ll {
		s := fmt.Sprintf(tem, col[wl-i-1], l)
		use = append(use, s)
	}

	tb := wk.FindAuthor()

	dbpool := grabpgsqlconnection()
	qt := `SELECT index FROM %s WHERE wkuniversalid='%s' AND %s ORDER BY index ASC`

	// if the last selection box was empty you are sent '_0' instead of a real value
	if ll[0] == "_0" {
		lu := len(use)
		use = use[1:lu]
	}

	if len(use) == 0 {
		fl = [2]int64{wk.FirstLine, wk.LastLine}
		return fl
	}

	a := strings.Join(use, " AND ")
	q := fmt.Sprintf(qt, tb, wuid, a)

	fmt.Println(q)
	foundrows, err := dbpool.Query(context.Background(), q)
	checkerror(err)

	var idx []int64

	defer foundrows.Close()
	for foundrows.Next() {
		var thehit int64
		err := foundrows.Scan(&thehit)
		checkerror(err)
		idx = append(idx, thehit)
	}

	fl = [2]int64{idx[0], idx[len(idx)-1]}
	return fl
}

func test_selection() {
	// t := AllAuthors["lt0474"].Cleaname
	// [c3] span of a work: "GET /selection/make/_?auth=lt0474&work=037&locus=2|100&endpoint=3|20 HTTP/1.1"
	// sv.Start = "2|100"
	// sv.End = "3|20"
	// --> [lt0474_FROM_58578_TO_61085]
	// --> [lt0474_FROM_57716_TO_60904]

	// ./HipparchiaGoServer -tt -psqp XXX -t1 lt0474 -t2 024 -t3 "1"
	// SELECT index FROM lt0474 WHERE wkuniversalid='lt0474w024' AND level_01_value='1' ORDER BY index ASC
	//{[] [] [] [] [] [] [lt0474_FROM_36136_TO_36151] [ ]}

	var s Session
	var sv SelectValues
	sv.Auth = cfg.TestV1
	sv.Work = cfg.TestV2
	sv.Start = cfg.TestV3
	// sv.Start = "2|100"
	// sv.End = "3|20"
	sv.Excl = false
	s = selected(sv, s)
	fmt.Println(s.Inclusions)
	return
}