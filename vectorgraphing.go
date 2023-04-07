//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"bytes"
	"fmt"
	"github.com/e-gun/wego/pkg/search"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/labstack/echo/v4"
	"html/template"
	"io"
	"math"
	"regexp"
)

//
// GRAPHING
//

// buildgraph - generate the html and js for a nearest neighbors search
func buildgraph(c echo.Context, coreword string, settings string, nn map[string]search.Neighbors) string {
	// go-echarts is "too clever" and opaque about how to not do things its way
	// we override their page.Render() to yield html+js (see the ModX and CustomX code below)
	// this gets injected to the "vectorgraphing" div on frontpage.html

	// see also: https://echarts.apache.org/en/option.html#series-graph

	// [a] acquire a charts.Graph
	g := generategraph(c, coreword, settings, nn)
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

//	// series data
//	Data interface{} `json:"data"`
//

//
//	// The categories of node, which is optional. If there is a classification of nodes,
//	// the category of each node can be assigned through data[i].category.
//	// And the style of category will also be applied to the style of nodes. categories can also be used in legend.
//	Categories []*GraphCategory
//
//	// EdgeLabel is the properties of an label of edge.
//	EdgeLabel *EdgeLabel `json:"edgeLabel"`
//}

func generategraph(c echo.Context, coreword string, settings string, nn map[string]search.Neighbors) *charts.Graph {
	const (
		SYMSIZE       = 25
		PERIPHSYMSZ   = 15
		SIZEDISTORT   = 2.25
		PRECISON      = 4
		REPULSION     = 6000
		GRAVITY       = .15
		EDGELEN       = 40
		EDGEFNTSZ     = 8
		SERIESNAME    = ""
		LAYOUTTYPE    = "force"
		LABELPOSITON  = "right"
		DOTHUE        = 236
		DOTSL         = ", 33%, 40%, 1)"
		LINECURVINESS = 0       // from 0 to 1, but non-zero will double-up the lines...
		LINETYPE      = "solid" // "solid", "dashed", "dotted"
	)

	se := AllSessions.GetSess(readUUIDCookie(c))

	ft := Config.Font
	if ft == "Noto" {
		ft = "'hipparchiasemiboldstatic', sans-serif"
	}

	graph := newsvgraph(ft, settings, coreword)

	var gnn []opts.GraphNode
	var gll []opts.GraphLink
	valuelabel := opts.EdgeLabel{Show: true, FontSize: EDGEFNTSZ, Formatter: "{c}"}

	round := func(val float64) float32 {
		ratio := math.Pow(10, float64(PRECISON))
		return float32(math.Round(val*ratio) / ratio)
	}

	// find the average similarity: this will let you adjust bubble size so that most similar are biggest
	var maxsim float64
	for _, w := range nn[coreword] {
		if w.Similarity > maxsim {
			maxsim = w.Similarity
		}
	}

	// dotstyle := opts.ItemStyle{Color: DOTCOLOR}
	vardot := func(i int) *opts.ItemStyle {
		// dv := DOTHUE + (i * 3) + 1
		dv := DOTHUE
		vd := "hsla(" + fmt.Sprintf("%d", dv) + DOTSL
		return &opts.ItemStyle{Color: vd}
	}

	// periphdot := opts.ItemStyle{Color: DOTCOLPERIPH}
	periphvardot := func(i int) *opts.ItemStyle {
		// dv := DOTHUE + (i * 2) + 50
		dv := DOTHUE
		vd := "hsla(" + fmt.Sprintf("%d", dv) + DOTSL
		return &opts.ItemStyle{Color: vd}
	}

	used := make(map[string]bool)

	// the center point
	gnn = append(gnn, opts.GraphNode{Name: coreword, Value: 0, SymbolSize: fmt.Sprintf("%.4f", SYMSIZE*SIZEDISTORT), ItemStyle: vardot(-1)})
	used[coreword] = true

	// the words directly related to this word
	for i, w := range nn[coreword] {
		sizemod := fmt.Sprintf("%.4f", ((w.Similarity/maxsim)*SIZEDISTORT)*SYMSIZE)
		gnn = append(gnn, opts.GraphNode{Name: w.Word, Value: round(w.Similarity), SymbolSize: sizemod, ItemStyle: vardot(i)})
		gll = append(gll, opts.GraphLink{Source: coreword, Target: w.Word, Value: round(w.Similarity), Label: &valuelabel})
		used[w.Word] = true
	}

	// the relationships between the other words
	coreterms := ToSet(StringMapKeysIntoSlice(nn))

	// populate the nodes with just the core collection of terms
	simpleweb := func() {
		for t := range coreterms {
			if t == coreword {
				continue
			}
			for _, w := range nn[t] {
				if _, ok := coreterms[w.Word]; ok {
					gll = append(gll, opts.GraphLink{Source: t, Target: w.Word, Value: round(w.Similarity), Label: &valuelabel})
				}
			}
		}
	}

	// populate the nodes with both the core terms and the neighbors of those terms as well
	expandedweb := func() {
		i := -1
		for t := range coreterms {
			i += 1
			if t == coreword {
				continue
			}
			for _, w := range nn[t] {
				if _, ok := coreterms[w.Word]; ok {
					gll = append(gll, opts.GraphLink{Source: t, Target: w.Word, Value: round(w.Similarity), Label: &valuelabel})
				}
				if _, ok := used[w.Word]; !ok {
					gnn = append(gnn, opts.GraphNode{Name: w.Word, Value: round(w.Similarity), SymbolSize: PERIPHSYMSZ, ItemStyle: periphvardot(i)})
					used[w.Word] = true
				}
				gll = append(gll, opts.GraphLink{Source: t, Target: w.Word, Value: round(w.Similarity), Label: &valuelabel})
			}
		}
	}

	if se.VecGraphExt {
		expandedweb()
	} else {
		simpleweb()
	}

	graph.AddSeries(SERIESNAME, gnn, gll,
		charts.WithLabelOpts(
			opts.Label{
				Show:       true,
				Position:   LABELPOSITON,
				FontFamily: ft,
			},
		),
		charts.WithLineStyleOpts(
			opts.LineStyle{
				Curveness: LINECURVINESS,
				Type:      LINETYPE,
			}),
		charts.WithGraphChartOpts(
			// https://github.com/go-echarts/go-echarts/opts/charts.go
			// cf. https://echarts.apache.org/en/option.html#series-graph
			opts.GraphChart{
				Layout: LAYOUTTYPE,
				Force: &opts.GraphForce{
					Repulsion:  REPULSION,
					Gravity:    GRAVITY,
					EdgeLength: EDGELEN,
				},
				Roam:               true,
				FocusNodeAdjacency: true,
			},
		),
	)
	return graph
}

// newsvgraph - return a pre-formatted charts.Graph
func newsvgraph(ft string, settings string, coreword string) *charts.Graph {
	const (
		CHRTWIDTH  = "1600px"
		CHRTHEIGHT = "1200px"
		FONTSTYLE  = "normal"
		TITLESTR   = "Nearest neighbors of »%s«"
		LEFTALIGN  = "20"
		BOTTALIGN  = "3%"
		SAVETYPE   = "svg"
		SAVESTR    = "Save to file..."
		TEXTCOLOR  = "" // "black"
	)

	tst := opts.TextStyle{
		Color:      TEXTCOLOR,
		FontStyle:  FONTSTYLE,
		FontSize:   16,
		FontFamily: ft,
		Padding:    "15",
		Normal:     nil,
	}

	sst := opts.TextStyle{
		Color:      TEXTCOLOR,
		FontStyle:  FONTSTYLE,
		FontSize:   10,
		FontFamily: ft,
	}

	tit := opts.Title{
		Title:         fmt.Sprintf(TITLESTR, coreword),
		TitleStyle:    &tst,
		Subtitle:      settings, // can not see this if you put the title on the bottom of the image
		SubtitleStyle: &sst,
		Top:           "",
		Bottom:        BOTTALIGN,
		Left:          LEFTALIGN,
		Right:         "",
	}

	tbs := opts.ToolBoxFeatureSaveAsImage{
		Show:  true,
		Type:  SAVETYPE, // svg, jpeg, png; svg requires specific chart initialization
		Name:  fmt.Sprintf(TITLESTR, StripaccentsSTR(coreword)),
		Title: SAVESTR, // get chinese if ""
	}

	tbf := opts.ToolBoxFeature{
		SaveAsImage: &tbs,
	}

	tbo := opts.Toolbox{
		Show:    true,
		Orient:  "vertical",
		Left:    LEFTALIGN,
		Top:     "",
		Right:   "",
		Bottom:  "",
		Feature: &tbf,
	}

	graph := charts.NewGraph()
	graph.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Width: CHRTWIDTH, Height: CHRTHEIGHT}),
		charts.WithTitleOpts(tit),
		charts.WithToolboxOpts(tbo),
		// charts.WithLegendOpts(opts.Legend{}),
	)

	// on using a legend see also https://echarts.apache.org/en/option.html#legend.data
	// nb legend users will want to look at series and/or categories too
	// not clear that we can/will gain anything with legends unless/until the graphing is rethought/expanded

	return graph
}

//
// OVERRIDE GO-ECHARTS [original code at https://github.com/go-echarts/go-echarts]
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
	const (
		TEMPLNAME = "chart"
		PATTERN   = `(__f__")|("__f__)|(__f__)`
	)

	for _, fn := range r.before {
		fn()
	}

	contents := []string{CustomHeaderTpl, CustomBaseTpl, CustomPageTpl}
	tpl := ModMustTemplate(TEMPLNAME, contents)

	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, TEMPLNAME, r.c); err != nil {
		return err
	}

	pat := regexp.MustCompile(PATTERN)
	content := pat.ReplaceAll(buf.Bytes(), []byte(""))

	_, err := w.Write(content)
	return err
}

// ModMustTemplate creates a new template with the given name and parsed contents.
func ModMustTemplate(name string, contents []string) *template.Template {
	const (
		JSNAME = "safeJS"
	)

	tpl := template.Must(template.New(name).Parse(contents[0])).Funcs(template.FuncMap{
		JSNAME: func(s interface{}) template.JS {
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
	<!-- "style" overridden because it is set in hgs.css -->
	<!-- CustomPageTpl -->
	{{ if eq .Layout "none" }}
		{{- range .Charts }} {{ template "base" . }} {{- end }}
	{{ end }}
	
	{{ if eq .Layout "center" }}
		<!-- <style> .container {display: flex;justify-content: center;align-items: center; } .item {margin: auto;} </style> -->
		{{- range .Charts }} {{ template "base" . }} {{- end }}
	{{ end }}
	
	{{ if eq .Layout "flex" }}
		<!--  <style> .box { justify-content:center; display:flex; flex-wrap:wrap } </style> -->
		<div class="box"> {{- range .Charts }} {{ template "base" . }} {{- end }} </div>
	{{ end }}
{{ end }}
`
