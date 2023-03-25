//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ynqa/wego/pkg/embedding"
	"github.com/ynqa/wego/pkg/model/word2vec"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

//
// BAGGING
//

var (
	// Latin100 - the 100 most common latin headwords
	Latin100 = []string{"qui¹", "et", "in", "edo¹", "is", "sum¹", "hic", "non", "ab", "ut", "Cos²", "si", "ad", "cum", "ex", "a", "eo¹",
		"ego", "quis¹", "tu", "Eos", "dico²", "ille", "sed", "de", "neque", "facio", "possum", "atque", "sui", "res",
		"quam", "aut", "ipse", "huc", "habeo", "do", "omne", "video", "ito", "magnus", "b", "alius²", "for", "idem",
		"suum", "etiam", "per", "enim", "omnes", "ita", "suus", "omnis", "autem", "vel", "vel", "Alius¹", "qui²", "quo",
		"nam", "bonus", "neo¹", "meus", "volo¹", "ne³", "ne¹", "suo", "verus", "pars", "reor", "sua", "vaco", "verum",
		"primus", "unus", "multus", "causa", "jam", "tamen", "Sue", "nos", "dies", "Ios", "modus", "tuus", "venio",
		"pro¹", "pro²", "ago", "deus", "annus", "locus", "homo", "pater", "eo²", "tantus", "fero", "quidem", "noster",
		"an", "locum"}
	LatExtra = []string{"at", "o", "tum", "tunc", "dum", "illic", "quia", "sive", "num", "adhuc"}
	LatStop  = append(Latin100, LatExtra...)
	// LatinKeep - members of LatStop we will not toss
	LatinKeep = []string{"facio", "possum", "habeo", "video", "magnus", "bonus", "volo¹", "primus", "venio", "ago",
		"deus", "annus", "locus", "pater", "fero"}
	// Greek150 - the 150 most common greek headwords
	Greek150 = []string{"ὁ", "καί", "τίϲ", "ἔδω", "δέ", "εἰμί", "δέω¹", "δεῖ", "δέομαι", "εἰϲ", "αὐτόϲ", "τιϲ", "οὗτοϲ", "ἐν",
		"γάροϲ", "γάρον", "γάρ", "οὐ", "μένω", "μέν", "τῷ", "ἐγώ", "ἡμόϲ", "κατά", "Ζεύϲ", "ἐπί", "ὡϲ", "διά",
		"πρόϲ", "προϲάμβ", "τε", "πᾶϲ", "ἐκ", "ἕ", "ϲύ", "Ἀλλά", "γίγνομαι", "ἁμόϲ", "ὅϲτιϲ", "ἤ¹", "ἤ²", "ἔχω",
		"ὅϲ", "μή", "ὅτι¹", "λέγω¹", "ὅτι²", "τῇ", "Τήιοϲ", "ἀπό", "εἰ", "περί", "ἐάν", "θεόϲ", "φημί", "ἐκάϲ",
		"ἄν¹", "ἄνω¹", "ἄλλοϲ", "qui¹", "πηρόϲ", "παρά", "ἀνά", "αὐτοῦ", "ποιέω", "ἄναξ", "ἄνα", "ἄν²", "πολύϲ",
		"οὖν", "λόγοϲ", "οὕτωϲ", "μετά", "ἔτι", "ὑπό", "ἑαυτοῦ", "ἐκεῖνοϲ", "εἶπον", "πρότεροϲ", "edo¹", "μέγαϲ",
		"ἵημι", "εἷϲ", "οὐδόϲ", "οὐδέ", "ἄνθρωποϲ", "ἠμί", "μόνοϲ", "κύριοϲ", "διό", "οὐδείϲ", "ἐπεί", "πόλιϲ",
		"τοιοῦτοϲ", "χάω", "καθά", "θεάομαι", "γε", "ἕτεροϲ", "δοκέω", "λαμβάνω", "δή", "δίδωμι", "ἵνα",
		"βαϲιλεύϲ", "φύϲιϲ", "ἔτοϲ", "πατήρ", "ϲῶμα", "καλέω", "ἐρῶ", "υἱόϲ", "ὅϲοϲ", "γαῖα", "οὔτε", "οἷοϲ",
		"ἀνήρ", "ὁράω", "ψυχή", "Ἔχιϲ", "ὥϲπερ", "αὐτόϲε", "χέω", "ὑπέρ", "ϲόϲ", "θεάω", "νῦν", "ἐμόϲ", "δύναμαι",
		"φύω", "πάλιν", "ὅλοξ", "ἀρχή", "καλόϲ", "δύναμιϲ", "πωϲ", "δύο", "ἀγαθόϲ", "οἶδα", "δείκνυμι", "χρόνοϲ",
		"ὅμοιοϲ", "ἕκαϲτοϲ", "ὁμοῖοϲ", "ὥϲτε", "ἡμέρα", "γράφω", "δραχμή", "μέροϲ"}
	GreekExtra = []string{}
	GreekStop  = append(Greek150, GreekExtra...)
	// GreekKeep - members of GreekStop we will not toss
	GreekKeep = []string{"ἔχω", "λέγω¹", "θεόϲ", "φημί", "ποιέω", "ἵημι", "μόνοϲ", "κύριοϲ", "πόλιϲ", "θεάομαι", "δοκέω", "λαμβάνω",
		"δίδωμι", "βαϲιλεύϲ", "φύϲιϲ", "ἔτοϲ", "πατήρ", "ϲῶμα", "καλέω", "ἐρῶ", "υἱόϲ", "γαῖα", "ἀνήρ", "ὁράω",
		"ψυχή", "δύναμαι", "ἀρχή", "καλόϲ", "δύναμιϲ", "ἀγαθόϲ", "οἶδα", "δείκνυμι", "χρόνοϲ", "γράφω", "δραχμή",
		"μέροϲ"}
	LatinStops     = getlatinstops()
	GreekStops     = getgreekstops()
	DefaultVectors = word2vec.Options{
		BatchSize:          1024,
		Dim:                125,
		DocInMemory:        true,
		Goroutines:         20,
		Initlr:             0.025,
		Iter:               15,
		LogBatch:           100000,
		MaxCount:           -1,
		MaxDepth:           150,
		MinCount:           10,
		MinLR:              0.0000025,
		ModelType:          "skipgram",
		NegativeSampleSize: 5,
		OptimizerType:      "hs",
		SubsampleThreshold: 0.001,
		ToLower:            false,
		UpdateLRBatch:      100000,
		Verbose:            true,
		Window:             8,
	}
)

func getgreekstops() map[string]struct{} {
	gs := SetSubtraction(GreekStop, GreekKeep)
	return ToSet(gs)
}

func getlatinstops() map[string]struct{} {
	ls := SetSubtraction(LatStop, LatinKeep)
	return ToSet(ls)
}

type WeightedHeadword struct {
	Word  string
	Count int
}

type WHWList []WeightedHeadword

func (w WHWList) Len() int {
	return len(w)
}

func (w WHWList) Less(i, j int) bool {
	return w[i].Count > w[j].Count
}

func (w WHWList) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
}

func fetchheadwordcounts(headwordset map[string]bool) map[string]int {
	if len(headwordset) == 0 {
		return make(map[string]int)
	}

	tt := "CREATE TEMPORARY TABLE ttw_%s AS SELECT words AS w FROM unnest(ARRAY[%s]) words"
	qt := "SELECT entry_name, total_count FROM dictionary_headword_wordcounts WHERE EXISTS " +
		"(SELECT 1 FROM ttw_%s temptable WHERE temptable.w = dictionary_headword_wordcounts.entry_name)"

	rndid := strings.Replace(uuid.New().String(), "-", "", -1)

	hw := make([]string, 0, len(headwordset))
	for h := range headwordset {
		hw = append(hw, h)
	}

	dbconn := GetPSQLconnection()
	defer dbconn.Release()

	arr := strings.Join(hw, "', '")
	arr = fmt.Sprintf("'%s'", arr)

	tt = fmt.Sprintf(tt, rndid, arr)
	_, err := dbconn.Exec(context.Background(), tt)
	chke(err)

	qt = fmt.Sprintf(qt, rndid)
	foundrows, e := dbconn.Query(context.Background(), qt)
	chke(e)

	returnmap := make(map[string]int)
	defer foundrows.Close()
	for foundrows.Next() {
		var thehit WeightedHeadword
		err = foundrows.Scan(&thehit.Word, &thehit.Count)
		chke(err)
		returnmap[thehit.Word] = thehit.Count
	}

	// don't kill off unfound terms
	for i := range hw {
		if _, t := returnmap[hw[i]]; t {
			continue
		} else {
			returnmap[hw[i]] = 0
		}
	}

	// "returnmap" for Albinus , poet. [lt2002]
	// map[abscondo:213 apte:168 aptus:1423 capitolium:0 celsus¹:1050 concludo:353 dactylus:167 de:42695 deus:14899 eo¹:58129 fio:12305 fretum:746 fretus¹:761 ille:44214 jungo:2275 liber¹:7550 liber⁴:13403 libo¹:3996 metrum:383 moenia¹:1308 non:96475 nullus:11785 pateo:1828 patesco:46 possum:41631 quis²:0 quis¹:52619 qui²:19812 qui¹:251744 re-pono:47 res:38669 romanus:0 sed:44131 sinus¹:1223 spondeum:158 spondeus:205 sponte:841 terni:591 totus²:0 totus¹:9166 triumphus:1058 tueor:3734 urbs:8564 verro:3843 versum:435 versus³:3390 verto:1471 †uilem:0]

	return returnmap
}

//
// DB INTERACTION
//

func vectordbinit(dbconn *pgxpool.Conn) {
	const (
		CREATE = `
			CREATE TABLE %s
			(
			  fingerprint character(32),
			  vectordata  bytea
			)`
	)
	ex := fmt.Sprintf(CREATE, VECTORTABLENAME)
	_, err := dbconn.Exec(context.Background(), ex)
	chke(err)
	msg("vectordbinit(): success", 3)
}

func vectordbcheck(fp string) bool {
	const (
		Q = `SELECT fingerprint FROM %s WHERE fingerprint = '%s' LIMIT 1`
	)
	dbconn := GetPSQLconnection()
	defer dbconn.Release()

	q := fmt.Sprintf(Q, VECTORTABLENAME, fp)
	foundrow, err := dbconn.Query(context.Background(), q)
	if err != nil {
		m := err.Error()
		if strings.Contains(m, "does not exist") {
			vectordbinit(dbconn)
		}
	}
	return foundrow.Next()
}

func vectordbadd(fp string, embs embedding.Embeddings) {
	const (
		MSG1 = "vectordbadd(): "
		INS  = `
			INSERT INTO %s
				(fingerprint, vectordata)
			VALUES ('%s', $1)`
	)

	eb, err := json.Marshal(embs)
	chke(err)

	l1 := len(eb)

	// https://stackoverflow.com/questions/61077668/how-to-gzip-string-and-return-byte-array-in-golang
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err = zw.Write(eb)
	chke(err)
	err = zw.Close()
	chke(err)

	b := buf.Bytes()
	l2 := len(b)

	ex := fmt.Sprintf(INS, VECTORTABLENAME, fp)

	dbconn := GetPSQLconnection()
	defer dbconn.Release()

	_, err = dbconn.Exec(context.Background(), ex, b)
	chke(err)
	msg(MSG1+fp, MSGFYI)

	// the savings is real: compressed is c. 27% of original
	msg(fmt.Sprintf("vector compression: %d -> %d (%.1f percent)", l1, l2, (float32(l2)/float32(l1))*100), 3)
}

func vectordbfetch(fp string) embedding.Embeddings {
	const (
		MSG1 = "vectordbfetch(): "
		Q    = `SELECT vectordata FROM %s WHERE fingerprint = '%s' LIMIT 1`
	)
	dbconn := GetPSQLconnection()
	defer dbconn.Release()

	q := fmt.Sprintf(Q, VECTORTABLENAME, fp)
	var vect []byte
	foundrow, err := dbconn.Query(context.Background(), q)
	chke(err)

	defer foundrow.Close()
	for foundrow.Next() {
		err = foundrow.Scan(&vect)
		chke(err)
	}

	// hipparchiaDB=# SELECT vectordata FROM vectors WHERE fingerprint = 'adb0ad4fe86ab27032cec006d0f68e6a' LIMIT 1;
	// it looks like:  \x3166386230383030303030303030303030306666....
	// or: 1f8b08000000000000ff849...

	var buf bytes.Buffer
	buf.Write(vect)

	// unzip
	zr, err := gzip.NewReader(&buf)
	chke(err)
	err = zr.Close()
	chke(err)
	decompr, err := io.ReadAll(zr)
	chke(err)

	var emb embedding.Embeddings
	err = json.Unmarshal(decompr, &emb)
	chke(err)

	msg(MSG1+fp, MSGFYI)

	return emb
}

func vectordbreset() {
	const (
		MSG1 = "vectordbreset()"
		E    = `DROP TABLE %s`
	)
	ex := fmt.Sprintf(E, VECTORTABLENAME)
	dbconn := GetPSQLconnection()
	defer dbconn.Release()

	_, err := dbconn.Exec(context.Background(), ex)
	chke(err)
	msg(MSG1, MSGFYI)
}

//
// GRAPHING
//

func buildgraph() error {
	// TESTING

	// https://github.com/go-echarts/examples/blob/master/examples/graph.go
	// [start] page.Render(io.MultiWriter(f))

	// see: https://github.com/go-echarts/go-echarts/blob/master/components/page.go
	// a page is a struct that contains a render.Renderer

	// see: https://github.com/go-echarts/go-echarts/blob/master/render/engine.go
	// a render.Renderer is an interface that calls Render(w io.Writer)

	// [rendering]
	// [a] call func (r *pageRender) Render(w io.Writer) error
	// [b] run any "before" functions
	// [c] build a slice of templates
	// [d] call "MustTemplate" on these
	// [e] call tpl.ExecuteTemplate and fill a buffer

	// [templating]
	// [a] tpl.ExecuteTemplate will use "r.c", that is pageRender.c where c is "interface{}"
	// [b] in order to have something populating "c" you need to have called NewPageRender(c interface{}, before ...func())
	// [c] NewPage() in https://github.com/go-echarts/go-echarts/blob/master/components/page.go calls NewPageRender(page, page.Validate)

	// [page validation]
	// NewPage() calls page.Assets.InitAssets(); this is an opts.Assets: see https://github.com/go-echarts/go-echarts/blob/master/opts/global.go

	// [assets]
	// these are JSAssets, CSSAssets, CustomizedJSAssets, CustomizedCSSAssets
	// they belong to types.orderedset: see "github.com/go-echarts/go-echarts/v2/types"

	// JSAssets.Init("echarts.min.js"): // Init creates a new OrderedSet instance, and adds any given items into this set.

	g := graphNpmDep()

	// we are building a page with only one chart and doing it by hand
	page := components.NewPage()

	assets := g.GetAssets()
	for _, v := range assets.JSAssets.Values {
		page.JSAssets.Add(v)
	}

	for _, v := range assets.CSSAssets.Values {
		page.CSSAssets.Add(v)
	}

	g.Validate()

	page.Charts = append(page.Charts, g)

	//fmt.Println(page.Charts[0])

	var ctpl = `
	{{- define "chart" }}
	<!DOCTYPE html>
	<html>
		{{- template "header" . }}
	<body>
		{{- template "base" . }}
	<style>
		.container {margin-top:30px; display: flex;justify-content: center;align-items: center;}
		.item {margin: auto;}
	</style>
	</body>
	</html>
	{{ end }}
	`

	t := func(name string, contents []string) *template.Template {
		tpl := template.Must(template.New(name).Parse(contents[0])).Funcs(template.FuncMap{
			"safeJS": func(s interface{}) template.JS {
				return template.JS(fmt.Sprint(s))
			},
		})

		for _, cont := range contents[1:] {
			tpl = template.Must(tpl.Parse(cont))
		}
		return tpl
	}

	tpl := t("chart", []string{ctpl})
	msg("x", 1)
	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, "chart", ctpl); err != nil {
		return err
	}

	msg("y", 1)
	pat := regexp.MustCompile(`(__f__")|("__f__)|(__f__)`)
	content := pat.ReplaceAll(buf.Bytes(), []byte(""))
	msg("z", 1)
	fmt.Println(string(content))

	// {      }
	//{map[echarts.min.js:true] [echarts.min.js]}
	//{map[] []}
	//center
	// &{{{false      <nil>   map[]  <nil> 0 0      <nil>} {false   false  <nil>} {false      <nil>} {dependencies demo <nil>   <nil>      } {<nil>} { 0 0 [ ] [ ] {false   false  <nil>}} {{ 0 0  false 0 0 false 0 0 0 0 0 false false} false} {{ 0 0  false 0 0 false 0 0 0 0 0 false false}   {  0  <nil> <nil>} 0 0 false} {<nil> <nil> <nil>} <nil> 0x140043fae70 {Awesome go-echarts 900px 500px  fHptAJfNusSt https://go-echarts.github.io/go-echarts-assets/assets/ white} {{map[echarts.min.js:true] [https://go-echarts.github.io/go-echarts-assets/assets/echarts.min.js]} {map[] []} {map[] []} {map[] []}} {[]  0 <nil> <nil> <nil>} { <nil> false} {   } {[]} {<nil> <nil>     } [{graph graph  0 0   false false  [{jquery jsdom 0 <nil>} {jquery xmlhttprequest 0 <nil>} {jquery htmlparser 0 <nil>} {jquery contextify 0 <nil>} {backbone underscore 0 <nil>} {faye faye-websocket 0 <nil>} {faye cookiejar 0 <nil>} {socket.io redis 0 <nil>} {socket.io socket.io-client 0 <nil>} {mongoose mongodb 0 <nil>} {mongoose hooks 0 <nil>} {mongoose ms 0 <nil>} {cheerio underscore 0 <nil>} {cheerio htmlparser2 0 <nil>} {cheerio entities 0 <nil>} {express mkdirp 0 <nil>} {express connect 0 <nil>} {express commander 0 <nil>} {express debug 0 <nil>} {express cookie 0 <nil>} {express send 0 <nil>} {express methods 0 <nil>}] none <nil> [] true <nil> <nil> <nil> false true <nil> false false false false false   <nil> <nil> <nil> 0  false 0 <nil>     0 <nil> <nil>  [] []   false false false 0 0  0 0  0 [{jquery -739.36383 -404.26147 0 false <nil>  4.7252817 0x140009c56d0} {backbone -134.2215 -862.7517 0 false <nil>  6.1554675 0x140009c5720} {underscore -75.53079 -734.4221 0 false <nil>  100 0x140009c5770} {faye -818.97516 624.50604 0 false <nil>  0.67816025 0x140009c57c0} {socket.io -710.59204 120.37976 0 false <nil>  19.818306 0x140009c5810} {requirejs 71.52897 -612.5541 0 false <nil>  4.0862627 0x140009c5860} {amdefine 1202.1166 -556.3107 0 false <nil>  2.3822114 0x140009c58b0} {mongoose -1150.2018 378.15536 0 false <nil>  10.81118 0x140009c5900} {underscore.deferred -127.03764 477.03778 0 false <nil>  0.40429485 0x140009c5950}] <nil> <nil> <nil> <nil> 0x140084c1170 <nil> <nil> <nil> <nil> 0x140043fb560 <nil> <nil> <nil>}] {[{  false <nil> 0 false <nil> <nil> 0 0 0 <nil> <nil> <nil> <nil>}] [{  false <nil> 0 false <nil> <nil> 0 <nil> <nil> <nil> <nil>}]} {false  0  <nil> <nil> <nil>} {false  0  <nil> <nil> <nil>} {false  0  <nil> <nil> <nil>} {false 0 0 0 <nil>} {     false } [] [#5470c6 #91cc75 #fac858 #ee6666 #73c0de #3ba272 #fc8452 #9a60b4 #ea7ccc] [] [] [] [] false false false false false false false false []} { { [] <nil>}}}

	// type Page struct {
	//	render.Renderer  // "github.com/go-echarts/go-echarts/v2/render"
	//	opts.Initialization
	//	opts.Assets
	//
	//	Charts []interface{}
	//	Layout Layout
	//}

	// func (r *chartRender) Render(w io.Writer) error {
	//	for _, fn := range r.before {
	//		fn()
	//	}
	//
	//	contents := []string{tpls.HeaderTpl, tpls.BaseTpl, tpls.ChartTpl}
	//	tpl := MustTemplate(ModChart, contents)
	//
	//	var buf bytes.Buffer
	//	if err := tpl.ExecuteTemplate(&buf, ModChart, r.c); err != nil {
	//		return err
	//	}
	//
	//	content := pat.ReplaceAll(buf.Bytes(), []byte(""))
	//
	//	_, err := w.Write(content)
	//	return err
	//}

	// AddCharts adds new charts to the page.
	//func (page *Page) AddCharts(charts ...Charter) *Page {
	//	for i := 0; i < len(charts); i++ {
	//	assets := charts[i].GetAssets()
	//	for _, v := range assets.JSAssets.Values {
	//	page.JSAssets.Add(v)
	//}
	//
	//	for _, v := range assets.CSSAssets.Values {
	//	page.CSSAssets.Add(v)
	//}
	//	charts[i].Validate()
	//	page.Charts = append(page.Charts, charts[i])
	//}
	//	return page
	//}

	// TESTING
	return nil
}

func graphNpmDep() *charts.Graph {
	graph := charts.NewGraph()
	graph.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "dependencies demo",
		}))

	f, err := ioutil.ReadFile("npmdepgraph.json")
	if err != nil {
		panic(err)
	}

	type Data struct {
		Nodes []opts.GraphNode
		Links []opts.GraphLink
	}

	var data Data
	if err := json.Unmarshal(f, &data); err != nil {
		fmt.Println(err)
	}

	graph.AddSeries("graph", data.Nodes, data.Links).
		SetSeriesOptions(
			charts.WithGraphChartOpts(opts.GraphChart{
				Layout:             "none",
				Roam:               true,
				FocusNodeAdjacency: true,
			}),
			charts.WithEmphasisOpts(opts.Emphasis{
				Label: &opts.Label{
					Show:     true,
					Color:    "black",
					Position: "left",
				},
			}),
			charts.WithLineStyleOpts(opts.LineStyle{
				Curveness: 0.3,
			}),
		)
	return graph
}

//
// WEGO NOTES AND DEFAULTS
//

func vectorconfig() word2vec.Options {
	const (
		ERR1 = "vectorconfig() cannot find UserHomeDir"
		ERR2 = "vectorconfig() failed to parse "
		MSG1 = "wrote default vector configuration file "
		MSG2 = "read vector configuration from "
	)

	// cfg := word2vec.DefaultOptions()
	cfg := DefaultVectors

	h, e := os.UserHomeDir()
	if e != nil {
		msg(ERR1, 0)
		return cfg
	}

	_, yes := os.Stat(fmt.Sprintf(CONFIGALTAPTH, h) + CONFIGVECTOR)

	if yes != nil {
		content, err := json.MarshalIndent(cfg, JSONINDENT, JSONINDENT)
		chke(err)

		err = os.WriteFile(fmt.Sprintf(CONFIGALTAPTH, h)+CONFIGVECTOR, content, 0644)
		chke(err)
		msg(MSG1+CONFIGVECTOR, 1)
	} else {
		loadedcfg, _ := os.Open(fmt.Sprintf(CONFIGALTAPTH, h) + CONFIGVECTOR)
		decoderc := json.NewDecoder(loadedcfg)
		vc := word2vec.Options{}
		errc := decoderc.Decode(&vc)
		_ = loadedcfg.Close()
		if errc != nil {
			msg(ERR2+CONFIGVECTOR, 0)
			cfg = DefaultVectors
		}
		msg(MSG2+CONFIGVECTOR, 2)
		cfg = vc
	}

	return cfg
}

//const (
//	NegativeSampling    OptimizerType = "ns"
//	HierarchicalSoftmax OptimizerType = "hs"
//)

//const (
//	Cbow     ModelType = "cbow"
//	SkipGram ModelType = "skipgram"
//)

// var (
//	defaultBatchSize          = 10000
//	defaultDim                = 10
//	defaultDocInMemory        = false
//	defaultGoroutines         = runtime.NumCPU()
//	defaultInitlr             = 0.025
//	defaultIter               = 15
//	defaultLogBatch           = 100000
//	defaultMaxCount           = -1
//	defaultMaxDepth           = 100
//	defaultMinCount           = 5
//	defaultMinLR              = defaultInitlr * 1.0e-4
//	defaultModelType          = Cbow
//	defaultNegativeSampleSize = 5
//	defaultOptimizerType      = NegativeSampling
//	defaultSubsampleThreshold = 1.0e-3
//	defaultToLower            = false
//	defaultUpdateLRBatch      = 100000
//	defaultVerbose            = false
//	defaultWindow             = 5
//)

// results do not repeat because word2vec.Train() in pkg/model/word2vec/word2vec.go has
// "vec[i] = (rand.Float64() - 0.5) / float64(dim)"

// see also: https://link.springer.com/article/10.1007/s41019-019-0096-6

//
// GENSIM NOTES
//

// modelbuilders.py
// 	negative (int, optional) – If > 0, negative sampling will be used, the int for negative specifies how many “noise words” should be drawn (usually between 5-20). If set to 0, no negative sampling is used.
//	seed (int, optional) – Seed for the random number generator. Initial vectors for each word are seeded with a hash of the concatenation of word + str(seed). Note that for a fully deterministically-reproducible run, you must also limit the model to a single worker thread (workers=1), to eliminate ordering jitter from OS thread scheduling. (In Python 3, reproducibility between interpreter launches also requires use of the PYTHONHASHSEED environment variable to control hash randomization).
// 	compute_loss (bool, optional) – If True, computes and stores loss value which can be retrieved using get_latest_training_loss()
//  window (int, optional) – Maximum distance between the current and predicted word within a sentence
//                gensimmodel = Word2Vec(bagsofwords,
//                                       min_count=vv.minimumpresence,
//                                       seed=1,
//                                       epochs=vv.trainingiterations,
//                                       vector_size=vv.dimensions,
//                                       sample=vv.downsample,
//                                       sg=1,  # the results seem terrible if you say sg=0
//                                       window=vv.window,
//                                       workers=workers,
//                                       compute_loss=computeloss)

// https://github.com/go-echarts/go-echarts/blob/master/opts/charts.go
// // GraphChart is the option set for graph chart.
//// https://echarts.apache.org/en/option.html#series-graph
//type GraphChart struct {
//	// Graph layout.
//	// * 'none' No layout, use x, y provided in node as the position of node.
//	// * 'circular' Adopt circular layout, see the example Les Miserables.
//	// * 'force' Adopt force-directed layout, see the example Force, the
//	// detail about layout configurations are in graph.force
//	Layout string
//
//	// Force is the option set for graph force layout.
//	Force *GraphForce
//
//	// Whether to enable mouse zooming and translating. false by default.
//	// If either zooming or translating is wanted, it can be set to 'scale' or 'move'.
//	// Otherwise, set it to be true to enable both.
//	Roam bool
//
//	// EdgeSymbol is the symbols of two ends of edge line.
//	// * 'circle'
//	// * 'arrow'
//	// * 'none'
//	// example: ["circle", "arrow"] or "circle"
//	EdgeSymbol interface{}
//
//	// EdgeSymbolSize is size of symbol of two ends of edge line. Can be an array or a single number
//	// example: [5,10] or 5
//	EdgeSymbolSize interface{}
//
//	// Draggable allows you to move the nodes with the mouse if they are not fixed.
//	Draggable bool
//
//	// Whether to focus/highlight the hover node and it's adjacencies.
//	FocusNodeAdjacency bool
//
//	// The categories of node, which is optional. If there is a classification of nodes,
//	// the category of each node can be assigned through data[i].category.
//	// And the style of category will also be applied to the style of nodes. categories can also be used in legend.
//	Categories []*GraphCategory
//
//	// EdgeLabel is the properties of an label of edge.
//	EdgeLabel *EdgeLabel `json:"edgeLabel"`
//}
