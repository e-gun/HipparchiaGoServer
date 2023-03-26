//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"html/template"
	"io"
	"os"
	"regexp"
)

//
// GRAPHING
//

func buildgraph() string {
	msg("DEBUGGING: buildgraph()", 0)
	// go-echarts is "too clever" and opaque about how to not do things its way
	// we override their page.Render() to yield html+js that gets injected to the "vectorgraphing" div on frontpage.html

	// [a] build a charts.Graph
	g := graphtest()
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

var graphNodes = []opts.GraphNode{
	{Name: "Node1"},
	{Name: "Node2"},
	{Name: "Node3"},
	{Name: "Node4"},
	{Name: "Node5"},
	{Name: "Node6"},
	{Name: "Node7"},
	{Name: "Node8"},
}

func genLinks() []opts.GraphLink {
	links := make([]opts.GraphLink, 0)
	for i := 0; i < len(graphNodes)-1; i++ {
		for j := len(graphNodes) - 1; j > 1; j-- {
			links = append(links, opts.GraphLink{Source: graphNodes[i].Name, Target: graphNodes[j].Name, Value: float32(j * 2)})
		}
	}
	return links
}

func graphtest() *charts.Graph {
	graph := charts.NewGraph()
	graph.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "basic graph example"}),
	)
	graph.AddSeries("graph", graphNodes, genLinks(),
		charts.WithGraphChartOpts(
			// opts.GraphChart{Force: &opts.GraphForce{Repulsion: 8000}},
			opts.GraphChart{Force: &opts.GraphForce{Repulsion: 2000}},
		),
	)
	return graph

}

func graphNpmDep() *charts.Graph {
	graph := charts.NewGraph()
	//graph.SetGlobalOptions(
	//	charts.WithTitleOpts(opts.Title{
	//		Title: "dependencies demo",
	//	}))

	f, err := os.ReadFile("npmdepgraph.json")
	if err != nil {
		panic(err)
	}

	type Data struct {
		Nodes []opts.GraphNode
		Links []opts.GraphLink
	}

	var data Data
	if e := json.Unmarshal(f, &data); e != nil {
		fmt.Println(e)
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
