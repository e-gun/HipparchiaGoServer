package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type JSData struct {
	Pound string
	Url   string
}

// SelectionValues - what was selected?
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

func RtSelectionMake(c echo.Context) error {
	// GET http://localhost:8000/selection/make/_?auth=lt0474&work=073&locus=3|10&endpoint=

	// note that you need to return JSON: reportcurrentselections() so as to fill #selectionstable on the page

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

	sessions[user] = selected(user, sel)
	jsbytes := reportcurrentselections(c)

	return c.String(http.StatusOK, string(jsbytes))
}

func RtSelectionClear(c echo.Context) error {
	// GET http://localhost:8000/selection/clear/wkselections/0
	user := readUUIDCookie(c)

	locus := c.Param("locus")
	which := strings.Split(locus, "/")

	if len(which) != 2 {
		msg(fmt.Sprintf("RtSelectionClear() was given bad input: %s", locus), 1)
		return c.String(http.StatusOK, "")
	}

	cat := which[0]
	id, e := strconv.Atoi(which[1])
	if e != nil {
		msg(fmt.Sprintf("RtSelectionClear() was given bad input: %s", locus), 1)
		return c.String(http.StatusOK, "")
	}

	// cat := []string{"agn", "wgn", "aloc", "wloc", "au", "wk", "psg"}

	mod := sessions[user]
	modi := mod.Inclusions
	mode := mod.Exclusions

	switch cat {
	case "agnselections":
		modi.AuGenres = RemoveIndex(modi.AuGenres, id)
	case "wgnselections":
		modi.WkGenres = RemoveIndex(modi.WkGenres, id)
	case "alocselections":
		modi.AuLocations = RemoveIndex(modi.AuLocations, id)
	case "wlocselections":
		modi.WkLocations = RemoveIndex(modi.WkLocations, id)
	case "auselections":
		//key := modi.Passages[id]
		//delete(modi.MappedAuthByName, key)
		modi.Authors = RemoveIndex(modi.Authors, id)
	case "wkselections":
		//key := modi.Passages[id]
		//delete(modi.MappedWkByName, key)
		modi.Works = RemoveIndex(modi.Works, id)
	case "psgselections":
		key := modi.Passages[id]
		delete(modi.MappedPsgByName, key)
		modi.Passages = RemoveIndex(modi.Passages, id)
	case "agnexclusions":
		mode.AuGenres = RemoveIndex(mode.AuGenres, id)
	case "wgnexclusions":
		mode.WkGenres = RemoveIndex(mode.WkGenres, id)
	case "alocexclusions":
		mode.AuLocations = RemoveIndex(mode.AuLocations, id)
	case "wlocexclusions":
		mode.WkLocations = RemoveIndex(mode.WkLocations, id)
	case "auexclusions":
		//key := mode.Passages[id]
		//delete(mode.MappedAuthByName, key)
		mode.Authors = RemoveIndex(mode.Authors, id)
	case "wkexclusions":
		//key := mode.Passages[id]
		//delete(mode.MappedPsgByName, key)
		mode.Works = RemoveIndex(mode.Works, id)
	case "psgexclusions":
		key := mode.Passages[id]
		delete(mode.MappedAuthByName, key)
		mode.Passages = RemoveIndex(mode.Passages, id)
	default:
		msg(fmt.Sprintf("RtSelectionClear() was given bad category: %s", cat), 1)
	}

	mod.Inclusions = modi
	mod.Exclusions = mode

	sessions[user] = mod

	r := RtSelectionFetch(c)

	return r
}

func RtSelectionFetch(c echo.Context) error {
	jsbytes := reportcurrentselections(c)
	msg(string(jsbytes), 1)
	return c.String(http.StatusOK, string(jsbytes))
}

func selected(user string, sv SelectionValues) Session {
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

	s := sessions[user]
	sep := "|"

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
		t := `%s_FROM_%d_TO_%d`
		i := fmt.Sprintf(t, sv.Auth, b[0], b[1])
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
		t := `%s_FROM_%d_TO_%d`
		i := fmt.Sprintf(t, sv.Auth, b[0], e[1])
		if !sv.IsExcl {
			s.Inclusions.Passages = unique(append(s.Inclusions.Passages, i))
			s.Inclusions.MappedPsgByName[i] = cs
		} else {
			s.Exclusions.Passages = unique(append(s.Exclusions.Passages, i))
			s.Exclusions.MappedPsgByName[i] = cs
		}
	}

	// TODO: genres, etc

	return s
}

func rationalizeselections() {
	// if you select "book 2" after selecting the whole, select only book 2
	// if you select the whole after book 2, then the whole
	// etc...

}

func findendpointsfromlocus(wuid string, locus string, sep string) [2]int64 {
	fl := [2]int64{0, 0}
	wk := AllWorks[wuid]

	wl := wk.CountLevels()
	ll := strings.Split(locus, sep)
	if len(ll) > wl {
		ll = ll[0:wl]
	}

	if len(ll) == 0 || ll[0] == "_0" {
		fl = [2]int64{wk.FirstLine, wk.LastLine}
		return fl
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

	tb := wk.FindAuthor()

	dbpool := GetPSQLconnection()
	defer dbpool.Close()
	qt := `SELECT index FROM %s WHERE wkuniversalid='%s' AND %s ORDER BY index ASC`

	a := strings.Join(use, " AND ")
	q := fmt.Sprintf(qt, tb, wuid, a)

	foundrows, err := dbpool.Query(context.Background(), q)
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
		msg(fmt.Sprintf("findendpointsfromlocus() failed to find the following inside of %s: %s", wuid, locus), -1)
		fl = [2]int64{1, 1}
	} else {
		fl = [2]int64{idx[0], idx[len(idx)-1]}
	}

	return fl
}

func reportcurrentselections(c echo.Context) []byte {
	// ultimately feeding autocomplete.js
	//    $('#timerestrictions').html(selectiondata.timeexclusions);
	//    $('#selectioninfocell').html(selectiondata.selections);
	//    $('#exclusioninfocell').html(selectiondata.exclusions);
	//    $('#selectionscriptholder').html(selectiondata['newjs']);

	s := sessions[readUUIDCookie(c)]
	s.Inclusions.BuildAuByName()
	s.Exclusions.BuildAuByName()
	s.Inclusions.BuildWkByName()
	s.Exclusions.BuildWkByName()
	s.Inclusions.BuildPsgByName()
	s.Exclusions.BuildPsgByName()

	i := s.Inclusions
	e := s.Exclusions

	//tb := `
	//<table id="selectionstable" style="">
	//<tbody>
	//{{.TimeRestr}}
	//<tr>
	//	<td class="infocells" id="selectioninfocell" title="Selection list" width="47%">
	//	{{.AuCatI}}{{.WkGenI}}{{.AuLocI}}{{.WkLocI}}{{.AuI}}{{.WksI}}{{.PsgI}}
	//	</td>
	//    <td style="text-align: center;" id="jscriptwigetcell" width="6%">
	//        <p id="searchinfo"><span class="ui-button-icon ui-icon ui-icon-info" title="Show/hide details of the current search list">&nbsp;</span></p>
	//    </td>
	//	<td class="infocellx" id="exclusioninfocell" title="Exclusion list" width="47%">
	//		{{.AuCatE}}{{.WkGenE}}{{.AuLocE}}{{.WkLocE}}{{.AuE}}{{.WksE}}{{.PsgE}}
	//	</td>
	//</tr>
	//</tbody>
	//</table>`

	pl := `<span class="picklabel">%s</span><br>`
	sl := `<span class="%sselections selection" id="%sselections_%02d" title="Double-click to remove this item">%s</span><br>`
	el := `<span class="%ssexclusions selection" id="%sexclusions_%02d" title="Double-click to remove this item">%s</span><br>`

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

	// JS output struct
	type SelectionData struct {
		TimeRestr string `json:"timeexclusions"`
		Select    string `json:"selections"`
		Exclude   string `json:"exclusions"`
		NewJS     string `json:"newjs"`
		Count     int    `json:"numberofselections"`
	}

	var jsinfo []JSData
	jsin := `%sselections_%02d`
	jsinu := `/selection/clear/%sselections/%d`
	jsex := `%sexclusions_%02d`
	jsexu := `/selection/clear/%sexclusions/%d`

	var sd SelectionData

	// run inclusions, then exclusions
	var rows [2][]string
	swap := [2]string{sl, el}
	for idx, v := range [2]SearchIncExl{i, e} {
		for _, ct := range cat {
			label := catmap[ct][0]
			using := catmap[ct][1]
			val := reflect.ValueOf(&v).Elem().FieldByName(using)
			// PITA to cast a Value to its type: https://stackoverflow.com/questions/17262238/how-to-cast-reflect-value-to-its-type
			// note that the next is terrible if we are not 100% sure of the interface
			slc := val.Interface().([]string)
			if len(slc) > 0 {
				rows[idx] = append(rows[idx], fmt.Sprintf(pl, label))
				for n, g := range slc {
					st := fmt.Sprintf(swap[idx], ct, ct, n, g)
					rows[idx] = append(rows[idx], st)
					if swap[idx] == sl {
						a := fmt.Sprintf(jsin, ct, n)
						b := fmt.Sprintf(jsinu, ct, n)
						jsinfo = append(jsinfo, JSData{a, b})
					} else {
						a := fmt.Sprintf(jsex, ct, n)
						b := fmt.Sprintf(jsexu, ct, n)
						jsinfo = append(jsinfo, JSData{a, b})
					}
				}
			}
		}
	}

	sd.Select = strings.Join(rows[0], "")
	sd.Exclude = strings.Join(rows[1], "")
	sd.Count = i.CountItems() + e.CountItems()
	sd.NewJS = formatnewselectionjs(jsinfo)

	js, err := json.Marshal(sd)
	chke(err)
	return js
}

func formatnewselectionjs(jsinfo []JSData) string {
	t := `
		$( '#%s' ).dblclick(function() {
			$.getJSON('%s', function (selectiondata) { 
				reloadselections(selectiondata); });
		});`

	s := `
	<script>%s
	</script>
	`

	var info []string
	for _, j := range jsinfo {
		info = append(info, fmt.Sprintf(t, j.Pound, j.Url))
	}

	script := fmt.Sprintf(s, strings.Join(info, ""))
	return script
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
	id := "testing"
	sessions[id] = s
	var sv SelectionValues
	//sv.Auth = "lt0474"
	//sv.Work = ""
	//sv.Start = ""
	// sv.Start = "2|100"
	// sv.End = "3|20"
	sv.IsExcl = false
	s = selected(id, sv)
	fmt.Println(s.Inclusions)
	return
}
