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
	"gonum.org/v1/gonum/mat"
	"html/template"
	"io"
	"math"
	"math/rand"
	"regexp"
	"strings"
)

const (
	CHRTWIDTH  = "1500px"
	CHRTHEIGHT = "1200px"
	DOTHUE     = 236
	DOTSAT     = 33
	DOTLUM     = 45
	DOTLUMPER  = DOTLUM + 25
	DOTSHIFT   = 0
)

//
// NEAREST NEIGHBORS FORCE GRAPHS
//

// buildblanknngraph - return a pre-formatted charts.Graph
func buildblanknngraph(settings string, coreword string, incl string) *charts.Graph {
	const (
		TITLESTR = "Nearest neighbors of »%s« in %s"
		SAVEFILE = "nearest_neighbors_of_%s"
	)

	// FYI: https://echarts.apache.org/en/theme-builder.html, but there is not much room for "theming" ATM

	t := fmt.Sprintf(TITLESTR, coreword, incl)
	sf := fmt.Sprintf(SAVEFILE, StripaccentsSTR(coreword))

	graph := charts.NewGraph()
	graph.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Width: CHRTWIDTH, Height: CHRTHEIGHT}),
		charts.WithTitleOpts(getcharttitleopts(t, settings)),
		charts.WithToolboxOpts(getcharttoolboxopts(sf)),
		// charts.WithLegendOpts(opts.Legend{}),
	)

	// on using a legend see also https://echarts.apache.org/en/option.html#legend.data
	// nb legend users will want to look at series and/or categories too
	// not clear that we can/will gain anything with legends unless/until the graphing is rethought/expanded

	return graph
}

// formatnngraph - fill out a blank graph
func formatnngraph(c echo.Context, graph *charts.Graph, coreword string, nn map[string]search.Neighbors) *charts.Graph {
	const (
		SYMSIZE       = 25
		PERIPHSYMSZ   = 15
		SIZEDISTORT   = 2.25
		PRECISON      = 4
		REPULSION     = 6000
		GRAVITY       = .15
		EDGELEN       = 40
		EDGEFNTSZ     = 8.0
		SERIESNAME    = ""
		LAYOUTTYPE    = "force"
		LABELPOSITON  = "right"
		LABELFTSIZE   = 12.0
		LINECURVINESS = 0       // from 0 to 1, but non-zero will double-up the lines...
		LINETYPE      = "solid" // "solid", "dashed", "dotted"
	)

	se := AllSessions.GetSess(readUUIDCookie(c))

	var gnn []opts.GraphNode
	var gll []opts.GraphLink
	valuelabel := opts.EdgeLabel{Show: true, FontSize: EDGEFNTSZ, Formatter: "{c}"}
	hiddenvals := opts.EdgeLabel{Show: false, FontSize: EDGEFNTSZ, Formatter: "{c}"}

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
		dv := DOTHUE + (i * DOTSHIFT)
		vd := fmthsl(dv, DOTSAT, DOTLUM)
		return &opts.ItemStyle{Color: vd}
	}

	// periphdot := opts.ItemStyle{Color: DOTCOLPERIPH}
	periphvardot := func(i int) *opts.ItemStyle {
		dv := DOTHUE + (i * DOTSHIFT)
		vd := fmthsl(dv, DOTSAT, DOTLUMPER)
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
				gll = append(gll, opts.GraphLink{Source: t, Target: w.Word, Value: round(w.Similarity), Label: &hiddenvals})
			}
		}
	}

	if se.VecGraphExt {
		expandedweb()
	} else {
		simpleweb()
	}

	ft := Config.Font
	if ft == "Noto" {
		ft = "'hipparchiasemiboldstatic', sans-serif"
	}

	graph.AddSeries(SERIESNAME, gnn, gll,
		charts.WithLabelOpts(
			opts.Label{
				Show:       true,
				Position:   LABELPOSITON,
				FontFamily: ft,
				FontSize:   LABELFTSIZE,
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

// customnngraphhtmlandjs - generate the html and js for a nearest neighbors search
func customnngraphhtmlandjs(g *charts.Graph) string {
	// go-echarts is "too clever" and opaque about how to not do things its way
	// we override their page.Render() to yield html+js (see the ModX and CustomX code below)
	// this gets injected to the "vectorgraphing" div on frontpage.html

	// see also: https://echarts.apache.org/en/option.html#series-graph

	g.Validate()

	// [a] we are building a page with only one chart and doing it by hand
	p := components.NewPage()
	p.Renderer = NewCustomPageRender(p, p.Validate)

	// [b] add assets to the page
	assets := g.GetAssets()
	for _, v := range assets.JSAssets.Values {
		p.JSAssets.Add(v)
	}

	for _, v := range assets.CSSAssets.Values {
		p.CSSAssets.Add(v)
	}

	// [c] add the chart to the page
	p.Charts = append(p.Charts, g)
	p.Validate()

	// [d] render the chart and get the html+js for it
	var buf bytes.Buffer
	err := p.Render(&buf)
	chke(err)

	htmlandjs := string(buf.Bytes())

	return htmlandjs
}

// fmthsl - turn hsl integers into an html hsl string
func fmthsl(h int, s int, l int) string {
	// 0, 0, 0 --> hsla(0, 0%, 0%, 1);
	st := func(i int) string { return fmt.Sprintf("%d", i) }
	return "hsla(" + st(h) + ", " + st(s) + "%, " + st(l) + "%, 1)"
}

//
// SHARED CHART FEATURES
//

func getcharttooltip() opts.Tooltip {
	const (
		TTF = ""
	)
	// https://echarts.apache.org/en/option.html#tooltip

	// the fancy TTF used at https://echarts.apache.org/examples/en/editor.html?c=scatter-aqi-color
	// 		TTF      = `
	//		function (param) {
	//			  var value = param.value;
	//			  // prettier-ignore
	//			  return '<div style="border-bottom: 1px solid rgba(255,255,255,.3); font-size: 18px;padding-bottom: 7px;margin-bottom: 7px">'
	//						+ param.seriesName + ' ' + value[0] + '日：'
	//						+ value[7]
	//						+ '</div>'
	//						+ schema[1].text + '：' + value[1] + '<br>'
	//						+ schema[2].text + '：' + value[2] + '<br>'
	//						+ schema[3].text + '：' + value[3] + '<br>'
	//						+ schema[4].text + '：' + value[4] + '<br>'
	//						+ schema[5].text + '：' + value[5] + '<br>'
	//						+ schema[6].text + '：' + value[6] + '<br>';
	//			}`
	// could use this to actually give a real location, etc. via a get()

	return opts.Tooltip{
		Show:        true,
		Trigger:     "item",      // item, axis, none
		TriggerOn:   "mousemove", // mousemove, click, mousemove|click, none
		Enterable:   false,
		Formatter:   TTF,
		AxisPointer: nil,
	}
}

func getcharttoolboxopts(sfn string) opts.Toolbox {
	const (
		SAVETYPE  = "png" // svg, jpeg, or png
		SAVESTR   = "Save to file..."
		LEFTALIGN = "20"
	)

	// A note on SAVETYPE: svg requires a chart initialization option: {renderer: 'svg'}
	// see https://echarts.apache.org/en/api.html#echarts and see CustomBaseTpl below
	// BUT, then the fonts turn into a problem since SVG has its own way of handling them
	// see: https://vecta.io/blog/how-to-use-fonts-in-svg
	// SO, at the end of the day, you do not want to use SVG

	tbs := opts.ToolBoxFeatureSaveAsImage{
		Show:  true,
		Type:  SAVETYPE,
		Name:  sfn,
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
	return tbo
}

func getcharttitleopts(t string, st string) opts.Title {
	const (
		LEFTALIGN = "20"
		BOTTALIGN = "0%"
		FONTSTYLE = "normal"
		FONTSIZE  = 14
		FONTDIFF  = 6
		TEXTPAD   = "10"
	)
	ft := Config.Font
	if ft == "Noto" {
		ft = "'hipparchiasemiboldstatic', sans-serif"
	}

	tst := opts.TextStyle{
		Color:      fmthsl(DOTHUE, DOTSAT, DOTLUM),
		FontStyle:  FONTSTYLE,
		FontSize:   FONTSIZE,
		FontFamily: ft,
		Padding:    TEXTPAD,
	}

	sst := opts.TextStyle{
		Color:      fmthsl(DOTHUE, DOTSAT, DOTLUMPER),
		FontStyle:  FONTSTYLE,
		FontSize:   FONTSIZE - FONTDIFF,
		FontFamily: ft,
	}

	tit := opts.Title{
		Title:         t,
		TitleStyle:    &tst,
		Subtitle:      st, // can not see this if you put the title on the very bottom of the image
		SubtitleStyle: &sst,
		Top:           "",
		Bottom:        BOTTALIGN,
		Left:          LEFTALIGN,
		Right:         "",
	}
	return tit
}

func getchartseriesstyle(top int) charts.SeriesOpts {
	so := charts.WithItemStyleOpts(getchartitemstyle(top))
	return so
}

func getchartitemstyle(top int) opts.ItemStyle {
	const (
		STARTHUE  = 0
		HUEOFFSET = 40
	)
	is := opts.ItemStyle{
		Color: fmthsl(STARTHUE+(HUEOFFSET*top), DOTSAT, DOTLUM),
	}
	return is
}

func getdotcitation(idx int, bags []BagWithLocus) string {
	const (
		NAMETMPL = "%s: %s"
		SAMPSIZE = 7
	)
	loc := strings.TrimPrefix(bags[idx].Loc, "line/")
	init := strings.Split(bags[idx].Bag, " ")
	samp := ""
	if len(init) > SAMPSIZE {
		samp = strings.Join(init[0:SAMPSIZE], " ") + "..."
		return fmt.Sprintf(NAMETMPL, loc, samp)
	} else {
		samp = strings.Join(init[0:], " ") + "..."
		return fmt.Sprintf(NAMETMPL, loc, samp)
	}
}

//
// LDA SCATTER GRAPHS
//

func lda2dscatter(ntopics int, incl string, bagger string, Y, labels mat.Matrix, bags []BagWithLocus) string {
	const (
		DOTSIZE  = 8
		DOTSTYLE = "triangle"
		TITLE    = "t-SNE scattergraph of %s"
		SAVEFILE = "lda_tsne_2d_scattergraph"
	)

	t := fmt.Sprintf(TITLE, incl)
	st := "text prep: " + bagger

	scatter := charts.NewScatter()
	scatter.SetGlobalOptions(
		charts.WithTitleOpts(getcharttitleopts(t, st)),
		charts.WithInitializationOpts(opts.Initialization{Width: CHRTHEIGHT, Height: CHRTHEIGHT}), // square
		charts.WithTooltipOpts(getcharttooltip()),
		charts.WithToolboxOpts(getcharttoolboxopts(SAVEFILE)),
	)

	dr, _ := Y.Dims()

	generateseries := func(topic int) []opts.ScatterData {
		// Value        interface{} `json:"value,omitempty"`
		items := make([]opts.ScatterData, 0)
		for i := 0; i < dr; i++ {
			label := int(labels.At(i, 0))
			if label == topic {
				x := Y.At(i, 0)
				y := Y.At(i, 1)
				items = append(items, opts.ScatterData{
					Value:        []float64{x, y},
					Symbol:       DOTSTYLE,
					SymbolSize:   DOTSIZE,
					SymbolRotate: rand.Intn(360),
					Name:         getdotcitation(i, bags),
				})
			}
		}
		return items
	}

	for i := 0; i < ntopics; i++ {
		scatter.AddSeries(fmt.Sprintf("Topic %d", i+1), generateseries(i), getchartseriesstyle(i))
	}

	page := components.NewPage()
	page.AddCharts(
		scatter,
	)

	// output to a file...
	//f, err := os.Create("lda2dscatter.html")
	//if err != nil {
	//	panic(err)
	//}
	//err = page.Render(io.MultiWriter(f))
	//chke(err)

	htmlandjs := customscatterhtmlandjs(scatter)
	return htmlandjs
}

func lda3dscatter(ntopics int, incl string, bagger string, Y, labels mat.Matrix, bags []BagWithLocus) string {
	const (
		TITLE    = "t-SNE scattergraph of %s"
		SAVEFILE = "lda_tsne_3d_scattergraph"
	)

	scatter := charts.NewScatter3D()
	scatter.SetGlobalOptions(
		charts.WithXAxis3DOpts(opts.XAxis3D{Name: "X-AXIS", Show: true}),
		charts.WithYAxis3DOpts(opts.YAxis3D{Name: "Y-AXIS"}),
		charts.WithZAxis3DOpts(opts.ZAxis3D{Name: "Z-AXIS"}),
	)

	t := fmt.Sprintf(TITLE, incl)
	st := "text prep: " + bagger

	scatter.SetGlobalOptions(
		charts.WithTitleOpts(getcharttitleopts(t, st)),
		charts.WithInitializationOpts(opts.Initialization{Width: CHRTHEIGHT, Height: CHRTHEIGHT}), // square
		charts.WithTooltipOpts(getcharttooltip()),
		charts.WithToolboxOpts(getcharttoolboxopts(SAVEFILE)),
	)

	dr, _ := Y.Dims()

	col := func(top int) *opts.ItemStyle {
		is := getchartitemstyle(top)
		return &is
	}

	generateseries := func(topic int) []opts.Chart3DData {
		// Value        interface{} `json:"value,omitempty"`
		items := make([]opts.Chart3DData, 0)
		for i := 0; i < dr; i++ {
			label := int(labels.At(i, 0))
			if label == topic {
				x := Y.At(i, 0)
				y := Y.At(i, 1)
				z := Y.At(i, 2)
				items = append(items, opts.Chart3DData{
					Value:     []interface{}{x, y, z},
					ItemStyle: col(topic),
					Name:      getdotcitation(i, bags),
				})
			}
		}
		return items
	}

	charts.WithEmphasisOpts(opts.Emphasis{
		Label:     nil,
		ItemStyle: nil,
	})

	for i := 0; i < ntopics; i++ {
		scatter.AddSeries(fmt.Sprintf("Topic %d", i+1), generateseries(i), getchartseriesstyle(i))
	}

	page := components.NewPage()
	page.AddCharts(
		scatter,
	)

	// output to a file...
	//f, err := os.Create("lda2dscatter.html")
	//if err != nil {
	//	panic(err)
	//}
	//err = page.Render(io.MultiWriter(f))
	//chke(err)

	// TODO? - custom3dscatterhtmlandjs() will panic...
	// htmlandjs := custom3dscatterhtmlandjs(scatter)

	// KLUDGE
	// use buffers; skip the disk
	htmlandjs := pagerendertojscriptblock(page)

	return htmlandjs
}

// pagerendertojscriptblock - kludge to use buffers to render a page and then trim the surrounding html
func pagerendertojscriptblock(page *components.Page) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	err := page.Render(io.MultiWriter(w))
	chke(err)

	r := io.Reader(&buf)
	sb := new(strings.Builder)
	_, err = io.Copy(sb, r)
	chke(err)

	// now you have " <html><head>...</head><body>STUFF</body></html>" ; you want to grab STUFF
	one := strings.Split(sb.String(), "<body>")
	two := one[1]
	three := strings.Split(two, "</body>")
	four := three[0]
	return four
}

func customscatterhtmlandjs(s *charts.Scatter) string {
	// go-echarts is "too clever" and opaque about how to not do things its way
	// we override their page.Render() to yield html+js (see the ModX and CustomX code below)
	// this gets injected to the "vectorgraphing" div on frontpage.html

	s.Validate()

	// [a] we are building a page with only one chart and doing it by hand
	p := components.NewPage()
	p.Renderer = NewCustomPageRender(p, p.Validate)

	// [b] add assets to the page
	assets := s.GetAssets()
	for _, v := range assets.JSAssets.Values {
		p.JSAssets.Add(v)
	}

	for _, v := range assets.CSSAssets.Values {
		p.CSSAssets.Add(v)
	}

	// [c] add the chart to the page
	p.Charts = append(p.Charts, s)
	p.Validate()

	// [d] render the chart and get the html+js for it
	var buf bytes.Buffer
	err := p.Render(&buf)
	if err != nil {
		msg("customscatterhtmlandjs() failed to render the page template", 1)
	}

	htmlandjs := string(buf.Bytes())

	return htmlandjs
}

func custom3dscatterhtmlandjs(s *charts.Scatter3D) string {
	// WARNING: this will not produce a chart right now

	s.Validate()

	// [a] we are building a page with only one chart and doing it by hand
	p := components.NewPage()
	p.Renderer = NewCustomPageRender(p, p.Validate)

	// [b] add assets to the page
	assets := s.GetAssets()
	for _, v := range assets.JSAssets.Values {
		p.JSAssets.Add(v)
	}

	for _, v := range assets.CSSAssets.Values {
		p.CSSAssets.Add(v)
	}

	// [c] add the chart to the page
	p.Charts = append(p.Charts, s)
	p.Validate()

	// [d] render the chart and get the html+js for it
	var buf bytes.Buffer
	err := p.Render(&buf)
	if err != nil {
		msg("custom3dscatterhtmlandjs() failed to render the page template", 1)
		//[Hipparchia Golang Server v.1.2.6] UNRECOVERABLE ERROR
		//template: chart:11:41: executing "base" at <.JSONNotEscapedAction>: can't evaluate field JSONNotEscapedAction in type *charts.Scatter3D
	}

	htmlandjs := string(buf.Bytes())

	return htmlandjs
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
    <!-- Note that all of these comments get nuked and will not be sent out to the page. Alas... -->
    <meta charset="utf-8">
    <title>{{ .PageTitle }} [CustomHeaderTpl]</title>
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

// note that BaseTpl has since changed at https://github.com/go-echarts/go-echarts/blob/master/charts/base.go

// CustomBaseTpl - to enable svg, add the following to "let goecharts_...": `, {renderer: "svg"}`; but the fonts will break
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
