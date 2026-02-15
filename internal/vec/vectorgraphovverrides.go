//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vec

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-echarts/go-echarts/v2/render"
	"github.com/go-echarts/go-echarts/v2/types"
)

//
// OVERRIDE GO-ECHARTS [original code at https://github.com/go-echarts/go-echarts]
//

// this is kludgy and likely to be a source of ongoing pain: v2.4.0 broke the api of v2.3.3
// any/all of the following could break again; the code is scattered across the go-echarts repo too

// 9cde6a4b was the last time HipparchiaGoServer used v2.3.3
// the material below used to be provided by the bottom of vectorgraphing.go

// ModRenderer etc modified from https://github.com/go-echarts/go-echarts/blob/master/render/engine.go
type ModRenderer interface {
	Render(w io.Writer) error
	RenderContent() []byte
	RenderSnippet() render.ChartSnippet
}

type CustomPageRender struct {
	ModRenderer
	c      interface{}
	before []func()
}

func (r *CustomPageRender) RenderContent() []byte {
	// v2.4 added this to the api; v2.3.3 does not need RenderContent()
	//TODO implement me
	fmt.Println("vec.CustomPageRender.RenderContent is unimplemented")
	panic("implement me")
}

func (r *CustomPageRender) RenderSnippet() render.ChartSnippet {
	// v2.4 added this to the api; v2.3.3 does not need RenderSnippet()
	//TODO implement me
	fmt.Println("vec.CustomPageRender.RenderSnippet is unimplemented")
	panic("implement me")
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

// RenderisSet - see https://github.com/go-echarts/go-echarts/blob/master/render/engine.go
func RenderisSet(name string, data interface{}) bool {
	v := reflect.ValueOf(data)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return false
	}

	return v.FieldByName(name).IsValid()
}

// ModMustTemplate creates a new template with the given name and parsed contents.
func ModMustTemplate(name string, contents []string) *template.Template {
	tpl := template.New(name).Funcs(template.FuncMap{
		"safeJS": func(s interface{}) template.JS {
			return template.JS(fmt.Sprint(s))
		},
		"isSet": RenderisSet,
		"injectInstance": func(funcStr types.FuncStr, echartsInstancePlaceholder string, chartID string) string {
			instance := render.EchartsInstancePrefix + chartID
			return strings.Replace(string(funcStr), echartsInstancePlaceholder, instance, -1)
		},
	})
	tpl = template.Must(tpl.Parse(contents[0]))

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
{{- define "base_element" -}}
<div class="container">
    <div class="item" id="{{ .ChartID }}" style="width:{{ .Initialization.Width }};height:{{ .Initialization.Height }};"></div>
</div>
{{- end -}}

{{- define "base_script" -}}
<script type="text/javascript">
    "use strict";
    let goecharts_{{ .ChartID | safeJS }} = echarts.init(document.getElementById('{{ .ChartID | safeJS }}'), "{{ .Theme }}", { renderer: "{{  .Initialization.Renderer }}" });
    let option_{{ .ChartID | safeJS }} = {{ template "base_option" . }}
    goecharts_{{ .ChartID | safeJS }}.setOption(option_{{ .ChartID | safeJS }});

  {{- range  $listener := .EventListeners }}
    {{if .Query  }}
    goecharts_{{ $.ChartID | safeJS }}.on({{ $listener.EventName }}, {{ $listener.Query | safeJS }}, {{ injectInstance $listener.Handler "%MY_ECHARTS%"  $.ChartID | safeJS }});
    {{ else }}
    goecharts_{{ $.ChartID | safeJS }}.on({{ $listener.EventName }}, {{ injectInstance $listener.Handler "%MY_ECHARTS%"  $.ChartID | safeJS }})
    {{ end }}
  {{- end }}

    {{- range .JSFunctions.Fns }}
    {{ injectInstance . "%MY_ECHARTS%"  $.ChartID  | safeJS }}
    {{- end }}
</script>
{{- end -}}

{{- define "base_option" }}
    {{- .JSONNotEscaped | safeJS }}
{{- end }};

{{- define "base" }}
    {{- template "base_element" . }}
    {{- template "base_script" . }}
{{- end }}
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

// version 2.3.3

// unused
// ModMustTemplate233 creates a new template with the given name and parsed contents.
//func ModMustTemplate233(name string, contents []string) *template.Template {
//	const (
//		JSNAME = "safeJS"
//	)
//
//	tpl := template.Must(template.New(name).Parse(contents[0])).Funcs(template.FuncMap{
//		JSNAME: func(s interface{}) template.JS {
//			return template.JS(fmt.Sprint(s))
//		},
//	})
//
//	for _, cont := range contents[1:] {
//		tpl = template.Must(tpl.Parse(cont))
//	}
//	return tpl
//}

// unused
// CustomBaseTpl233 - to enable svg, add the following to "let goecharts_...": `, {renderer: "svg"}`; but the fonts will break
//var CustomBaseTpl233 = `
//{{- define "base" }}
//<!-- CustomBaseTpl -->
//<div class="container">
//    <div class="item" id="{{ .ChartID }}" style="width:{{ .Initialization.Width }};height:{{ .Initialization.Height }};"></div>
//</div>
//<script type="text/javascript">
//    "use strict";
//    let goecharts_{{ .ChartID | safeJS }} = echarts.init(document.getElementById('{{ .ChartID | safeJS }}'), "{{ .Theme }}");
//    let option_{{ .ChartID | safeJS }} = {{ .JSONNotEscaped | safeJS }};
//	let action_{{ .ChartID | safeJS }} = {{ .JSONNotEscapedAction | safeJS }};
//    goecharts_{{ .ChartID | safeJS }}.setOption(option_{{ .ChartID | safeJS }});
// 	goecharts_{{ .ChartID | safeJS }}.dispatchAction(action_{{ .ChartID | safeJS }});
//
//    {{- range .JSFunctions.Fns }}
//    {{ . | safeJS }}
//    {{- end }}
//</script>
//{{ end }}
//`
