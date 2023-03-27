//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"bytes"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/labstack/echo/v4"
	"github.com/ynqa/wego/pkg/embedding"
	"github.com/ynqa/wego/pkg/search"
	"html/template"
	"io"
	"regexp"
)

//
// GRAPHING
//

var graphLabels = []opts.EdgeLabel{
	{Show: true, Formatter: "{c}"},
	{Show: true, Formatter: "{b}"},
}

func generategraphdata(c echo.Context, srch SearchStruct) map[string]search.Neighbors {
	const (
		FAIL1 = "generategraphdata() could not find neighbors of a neighbor: '%s' neighbors (via '%s')"
	)

	msg("generategraphdata()", MSGFYI)
	fp := fingerprintvectorsearch(srch)
	isstored := vectordbcheck(fp)
	var embs embedding.Embeddings
	if isstored {
		msg("generategraphdata(): fetching stored embeddings", MSGFYI)
		embs = vectordbfetch(fp)
	} else {
		embs = generateembeddings(c, srch)
		vectordbadd(fp, embs)
	}

	// [b] make a query against the model
	searcher, err := search.New(embs...)
	chke(err)

	ncount := VECTORNEIGHBORS // how many neighbors to output; min is 1
	word := srch.Seeking

	nn := make(map[string]search.Neighbors)
	neighbors, err := searcher.SearchInternal(word, ncount)
	chke(err)

	nn[word] = neighbors
	for _, n := range neighbors {
		meta, e := searcher.SearchInternal(n.Word, ncount)
		if e != nil {
			msg(fmt.Sprintf(FAIL1, n.Word, word), 3)
		} else {
			nn[n.Word] = meta
		}
	}

	return nn
}

func buildgraph(coreword string, nn map[string]search.Neighbors) string {

	// go-echarts is "too clever" and opaque about how to not do things its way
	// we override their page.Render() to yield html+js that gets injected to the "vectorgraphing" div on frontpage.html

	// [a] build a charts.Graph

	g := generategraph(coreword, nn)
	g.Validate()

	// [b] we are building a page with only one chart and doing it by hand
	p := components.NewPage()
	p.Renderer = NewCustomPageRender(p, p.Validate)

	// [c] add assets to the page
	assets := g.GetAssets()
	for _, v := range assets.JSAssets.Values {
		p.JSAssets.Add(v)
	}

	for _, v := range assets.CSSAssets.Values {
		p.CSSAssets.Add(v)
	}

	// [d] add the chart to the page
	p.Charts = append(p.Charts, g)
	p.Validate()

	// [e] render the chart and get the html+js for it
	var buf bytes.Buffer
	err := p.Render(&buf)
	chke(err)

	htmlandjs := string(buf.Bytes())

	return htmlandjs
}

// see also: https://echarts.apache.org/en/option.html#series-graph

// type SingleSeries struct {
//	Name string `json:"name,omitempty"`
//	Type string `json:"type,omitempty"`
//
//	// Graph
//	Links              interface{} `json:"links,omitempty"`
//	Layout             string      `json:"layout,omitempty"`
//	Force              interface{} `json:"force,omitempty"`
//	Categories         interface{} `json:"categories,omitempty"`
//	Roam               bool        `json:"roam,omitempty"`
//	EdgeSymbol         interface{} `json:"edgeSymbol,omitempty"`
//	EdgeSymbolSize     interface{} `json:"edgeSymbolSize,omitempty"`
//	EdgeLabel          interface{} `json:"edgeLabel,omitempty"`
//	Draggable          bool        `json:"draggable,omitempty"`
//	FocusNodeAdjacency bool        `json:"focusNodeAdjacency,omitempty"`

//	// series data
//	Data interface{} `json:"data"`
//
//	// series options
//	*opts.Encode        `json:"encode,omitempty"`
//	*opts.ItemStyle     `json:"itemStyle,omitempty"`
//	*opts.Label         `json:"label,omitempty"`
//	*opts.LabelLine     `json:"labelLine,omitempty"`
//	*opts.Emphasis      `json:"emphasis,omitempty"`
//	*opts.MarkLines     `json:"markLine,omitempty"`
//	*opts.MarkAreas     `json:"markArea,omitempty"`
//	*opts.MarkPoints    `json:"markPoint,omitempty"`
//	*opts.RippleEffect  `json:"rippleEffect,omitempty"`
//	*opts.LineStyle     `json:"lineStyle,omitempty"`
//	*opts.AreaStyle     `json:"areaStyle,omitempty"`
//	*opts.TextStyle     `json:"textStyle,omitempty"`
//	*opts.CircularStyle `json:"circular,omitempty"`
//}

func generategraph(coreword string, nn map[string]search.Neighbors) *charts.Graph {

	graph := charts.NewGraph()
	graph.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Width: "1500px", Height: "1000px"}),
		charts.WithTitleOpts(opts.Title{Title: "generategraph()"}),
	)

	var gnn []opts.GraphNode
	var gll []opts.GraphLink

	coreterms := ToSet(StringMapKeysIntoSlice(nn))
	stringkeyprinter("coreterms", coreterms)

	// the center point
	gnn = append(gnn, opts.GraphNode{Name: coreword, Value: 0})

	// the words directly related to this word
	for _, w := range nn[coreword] {
		gnn = append(gnn, opts.GraphNode{Name: w.Word, Value: float32(w.Similarity) * 1000})
		gll = append(gll, opts.GraphLink{Source: coreword, Target: w.Word, Value: float32(w.Similarity) * 1000, Label: &graphLabels[0]})
	}

	// the relationships between the other words [fancier would be to have each word center its own cluster]
	for t := range coreterms {
		if t == coreword {
			continue
		}
		for _, w := range nn[t] {
			if _, ok := coreterms[w.Word]; ok {
				gll = append(gll, opts.GraphLink{Source: t, Target: w.Word, Value: float32(w.Similarity) * 1000, Label: &graphLabels[0]})
			}
		}
	}

	graph.AddSeries("graph", gnn, gll,
		charts.WithGraphChartOpts(
			// opts.GraphChart{Force: &opts.GraphForce{Repulsion: 8000}},
			opts.GraphChart{
				Layout: "force",
				Force: &opts.GraphForce{
					Repulsion:  8000,
					Gravity:    .1,
					EdgeLength: 40,
				},
				Roam:               true,
				FocusNodeAdjacency: true,
				EdgeLabel:          &graphLabels[1],
			},
		),
	)
	return graph
}

//
// OVERRIDE GO-ECHARTS
//

// ModRenderer etc modified from https://github.com/go-echarts/go-echarts/render/engine.go
type ModRenderer interface {
	Render(w io.Writer) error
}

type CustomPageRender struct {
	c      interface{}
	before []func()
}

// NewCustomPageRender returns a render implementation for Page.
func NewCustomPageRender(c interface{}, before ...func()) ModRenderer {
	return &CustomPageRender{c: c, before: before}
}

// Render renders the page into the given io.Writer.
func (r *CustomPageRender) Render(w io.Writer) error {
	for _, fn := range r.before {
		fn()
	}

	contents := []string{CustomHeaderTpl, CustomBaseTpl, CustomPageTpl}
	tpl := ModMustTemplate("chart", contents)

	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, "chart", r.c); err != nil {
		return err
	}

	pat := regexp.MustCompile(`(__f__")|("__f__)|(__f__)`)
	content := pat.ReplaceAll(buf.Bytes(), []byte(""))

	_, err := w.Write(content)
	return err
}

// ModMustTemplate creates a new template with the given name and parsed contents.
func ModMustTemplate(name string, contents []string) *template.Template {
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

// CustomHeaderTpl etc. adapted from https://github.com/go-echarts/go-echarts/templates/
var CustomHeaderTpl = `
{{ define "header" }}
<head>
	<!-- CustomHeaderTpl -->
    <meta charset="utf-8">
    <title>{{ .PageTitle }}</title>
{{- range .JSAssets.Values }}
    <script src="{{ . }}"></script>
{{- end }}
{{- range .CustomizedJSAssets.Values }}
    <script src="{{ . }}"></script>
{{- end }}
{{- range .CSSAssets.Values }}
    <link href="{{ . }}" rel="stylesheet">
{{- end }}
{{- range .CustomizedCSSAssets.Values }}
    <link href="{{ . }}" rel="stylesheet">
{{- end }}
</head>
{{ end }}
`

var CustomBaseTpl = `
{{- define "base" }}
<!-- CustomBaseTpl -->
<div class="container">
    <div class="item" id="{{ .ChartID }}" style="width:{{ .Initialization.Width }};height:{{ .Initialization.Height }};"></div>
</div>
<script type="text/javascript">
    "use strict";
    let goecharts_{{ .ChartID | safeJS }} = echarts.init(document.getElementById('{{ .ChartID | safeJS }}'), "{{ .Theme }}");
    let option_{{ .ChartID | safeJS }} = {{ .JSONNotEscaped | safeJS }};
	let action_{{ .ChartID | safeJS }} = {{ .JSONNotEscapedAction | safeJS }};
    goecharts_{{ .ChartID | safeJS }}.setOption(option_{{ .ChartID | safeJS }});
 	goecharts_{{ .ChartID | safeJS }}.dispatchAction(action_{{ .ChartID | safeJS }});

    {{- range .JSFunctions.Fns }}
    {{ . | safeJS }}
    {{- end }}
</script>
{{ end }}
`

var CustomPageTpl = `
{{- define "chart" }}
	<!-- CustomPageTpl -->
	{{ if eq .Layout "none" }}
		{{- range .Charts }} {{ template "base" . }} {{- end }}
	{{ end }}
	
	{{ if eq .Layout "center" }}
		<style> .container {display: flex;justify-content: center;align-items: center;} .item {margin: auto;} </style>
		{{- range .Charts }} {{ template "base" . }} {{- end }}
	{{ end }}
	
	{{ if eq .Layout "flex" }}
		<style> .box { justify-content:center; display:flex; flex-wrap:wrap } </style>
		<div class="box"> {{- range .Charts }} {{ template "base" . }} {{- end }} </div>
	{{ end }}
{{ end }}
`

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

// ? https://github.com/yourbasic/graph

// ? https://github.com/go-echarts/go-echarts

// look at what is possible and links: https://blog.gopheracademy.com/advent-2018/go-webgl/
