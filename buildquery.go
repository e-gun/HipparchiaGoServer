//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type PrerolledQuery struct {
	TempTable string
	PsqlQuery string
	PsqlData  string
}

type QueryBuilder struct {
	SelFrom   string
	WhrTrm    string
	WhrIdxInc string
	WhrIdxExc string
}

type Boundaries struct {
	Start int64
	Stop  int64
}

type PRQTemplate struct {
	AU    string
	COL   string
	SYN   string
	SK    string
	LIM   string
	IDX   string
	TTN   string
	Tail  *template.Template
	PSCol string
}

const (
	SELECTFROM = `
		SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, 
			marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM %s`
	PRFXSELFRM = `
		SELECT second.wkuniversalid, second.index, second.level_05_value, second.level_04_value, second.level_03_value, second.level_02_value, second.level_01_value, second.level_00_value, 
			second.marked_up_line, second.accented_line, second.stripped_line, second.hyphenated_words, second.annotations FROM`
	ASLINEBUNDLE = `
		( SELECT * FROM
			( SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations,
				concat(%s, ' ', lead(%s) OVER (ORDER BY index ASC) ) AS linebundle`
)

// findselectionboundaries() stuff should all be handled when making the selections, not here
// and everything should be in the "gr0032w002_FROM_11313_TO_11843" format

// take a searchlist and its exceptionlist and convert them into a collection of sql queries

// the collection of work will be either map[string]PrerolledQuery []PrerolledQuery

// onehit should be checked early: "ORDER BY index ASC LIMIT 1"

// search types
// [a] simple
// [b] simplelemma
// [c] proximate by words
// [d] proximate by lines
// [e] simple phrase
// [f] phrase and proximity

//
// CORE LOGIC
//

func searchlistintoqueries(ss *SearchStruct) []PrerolledQuery {
	var prqq []PrerolledQuery
	inc := ss.SearchIn
	exc := ss.SearchEx

	if len(ss.LemmaOne) != 0 {
		ss.SkgSlice = lemmaintoregexslice(ss.LemmaOne)
	} else {
		ss.SkgSlice = append(ss.SkgSlice, ss.Seeking)
	}

	// if there are too many "in0001wXXX" type entries: requiresindextemptable()

	// au query looks like: SELECTFROM + WHERETERM + WHEREINDEX + ORDERBY&LIMIT
	// FROM gr0308 WHERE ( (index BETWEEN 138 AND 175) OR (index BETWEEN 471 AND 510) ) AND ( stripped_line ~* $1 )
	// FROM lt0917 WHERE ( (index BETWEEN 1 AND 8069) OR (index BETWEEN 8070 AND 8092) ) AND ( (index NOT BETWEEN 1431 AND 2193) ) AND ( stripped_line ~* $1 )

	// [au] figure out all bounded selections

	boundedincl := make(map[string][]Boundaries)
	boundedexcl := make(map[string][]Boundaries)

	// [a1] individual works included/excluded
	for _, w := range inc.Works {
		wk := AllWorks[w]
		b := Boundaries{wk.FirstLine, wk.LastLine}
		boundedincl[wk.FindAuthor()] = append(boundedincl[wk.FindAuthor()], b)
	}

	for _, w := range exc.Works {
		wk := AllWorks[w]
		b := Boundaries{wk.FirstLine, wk.LastLine}
		boundedexcl[wk.FindAuthor()] = append(boundedexcl[wk.FindAuthor()], b)
	}

	// [a2] individual passages included/excluded

	pattern := regexp.MustCompile(`(?P<auth>......)_FROM_(?P<start>\d+)_TO_(?P<stop>\d+)`)
	for _, p := range inc.Passages {
		// "gr0032_FROM_11313_TO_11843"
		subs := pattern.FindStringSubmatch(p)
		au := subs[pattern.SubexpIndex("auth")]
		st, _ := strconv.Atoi(subs[pattern.SubexpIndex("start")])
		sp, _ := strconv.Atoi(subs[pattern.SubexpIndex("stop")])
		b := Boundaries{int64(st), int64(sp)}
		boundedincl[au] = append(boundedincl[au], b)
		// fmt.Printf("%s: %d - %d\n", au, st, sp)
	}

	for _, p := range exc.Passages {
		subs := pattern.FindStringSubmatch(p)
		au := subs[pattern.SubexpIndex("auth")]
		st, _ := strconv.Atoi(subs[pattern.SubexpIndex("start")])
		sp, _ := strconv.Atoi(subs[pattern.SubexpIndex("stop")])
		b := Boundaries{int64(st), int64(sp)}
		boundedexcl[au] = append(boundedexcl[au], b)
	}

	// [b] build the queries for the author tables
	idxtmpl := `(index %sBETWEEN %d AND %d)` // %s is "" or "NOT "

	// [b1] collapse inc.Authors, inc.Works, incl.Passages to find all tables in use
	// but the keys to boundedincl in fact gives you the answer to the latter two

	alltables := inc.Authors
	for t, _ := range boundedincl {
		alltables = append(alltables, t)
	}

	tails := acquiretails()

	for _, au := range alltables {
		var qb QueryBuilder
		var prq PrerolledQuery

		// [b2a] check to see if bounded by inclusions
		if bb, found := boundedincl[au]; found {
			if len(bb) > TEMPTABLETHRESHOLD {
				prq.TempTable = requiresindextemptable(au, bb, ss)
			} else {
				qb.WhrIdxInc = andorwhereclause(bb, idxtmpl, "", " OR ")
			}
		}

		// [b2b] check to see if bounded by exclusions
		if bb, found := boundedexcl[au]; found {
			if len(bb) > TEMPTABLETHRESHOLD {
				// note that 200 incl + 200 excl will produce garbage; in practice you have only au ton of one of them
				prq.TempTable = requiresindextemptable(au, bb, ss)
			} else {
				qb.WhrIdxExc = andorwhereclause(bb, idxtmpl, "NOT ", " AND ")
			}
		}

		// [b3] search term might be lemmatized, hence the range

		for _, skg := range ss.SkgSlice {
			// there are fancier ways to do this, but debugging and maintaining become overwhelming...

			// map to the %s items in the qtmpl below:
			// SELECTFROM + WHERETERM + WHEREINDEXINCL + WHEREINDEXEXCL + (either) ORDERBY&LIMIT (or) SECOND

			nott := len(prq.TempTable) == 0
			yestt := len(prq.TempTable) != 0
			noph := !ss.HasPhrase
			yesphr := ss.HasPhrase
			noidx := len(qb.WhrIdxExc) == 0 && len(qb.WhrIdxInc) == 0
			yesidx := len(qb.WhrIdxExc) != 0 || len(qb.WhrIdxInc) != 0

			var t PRQTemplate
			t.AU = au
			t.COL = ss.SrchColumn
			t.SYN = ss.SrchSyntax
			t.SK = skg
			t.LIM = fmt.Sprintf("%d", ss.Limit)
			if ss.NotNear {
				t.IDX = qb.WhrIdxExc
			} else {
				t.IDX = qb.WhrIdxInc
			}
			t.TTN = ss.TTName
			t.PSCol = ss.SrchColumn

			if nott && noph && noidx {
				t.Tail = tails["basic"]
				prq = basicprq(t, prq)
			} else if nott && noph && yesidx {
				// word in work(s)/passage(s): AND ( (index BETWEEN 481 AND 483) OR (index BETWEEN 501 AND 503) ... )
				t.Tail = tails["basic_and_indices"]
				prq = basicidxprq(t, prq)
			} else if nott && yesphr && noidx {
				t.Tail = tails["basic_window"]
				prq = basicwindowprq(t, prq)
			} else if nott && yesphr && yesidx {
				t.Tail = tails["window_with_indices"]
				prq = windandidxprq(t, prq)
			} else if yestt && noph {
				t.Tail = tails["simple_tt"]
				prq = simplettprq(t, prq)
			} else {
				t.Tail = tails["window_with_tt"]
				prq = windowandttprq(t, prq)
			}
			prqq = append(prqq, prq)
		}
	}
	return prqq
}

//
// HELPERS
//

func acquiretails() map[string]*template.Template {
	// this avoids recompiling them a bunch of times in a loop

	mm := make(map[string]string)
	mm["basic"] = `{{ .AU }} WHERE {{ .COL }} {{ .SYN }} '{{ .SK }}' ORDER BY index ASC LIMIT {{ .LIM }}`
	mm["basic_and_indices"] = `{{ .AU }} WHERE {{ .COL }} {{ .SYN }} '{{ .SK }}' AND ({{ .IDX }}) ORDER BY index ASC LIMIT {{ .LIM }}`
	mm["basic_window"] = ` FROM {{ .AU }} ) first
			) second WHERE second.linebundle {{ .SYN }} '{{ .SK }}' ORDER BY index ASC LIMIT {{ .LIM }}`
	mm["window_with_indices"] = ` FROM {{ .AU }} WHERE {{ .IDX }} ) first 
			) second WHERE second.linebundle {{ .SYN }} '{{ .SK }}' ORDER BY index ASC LIMIT {{ .LIM }}`
	mm["simple_tt"] = ` {{ .AU }} WHERE EXISTS
		(SELECT 1 FROM {{ .AU }}_includelist_{{ .TTN }} incl WHERE incl.includeindex = {{ .AU }}.index AND {{ .COL }} {{ .SYN }} '{{ .SK }}') LIMIT {{ .LIM }}`
	mm["window_with_tt"] = ` FROM {{ .AU }} WHERE EXISTS
			(SELECT 1 FROM {{ .AU }}_includelist_{{ .TTN }} incl WHERE incl.includeindex = {{ .AU }}.index ) 
			) first
		) second WHERE second.linebundle {{ .SYN }} '{{ .SK }}' LIMIT {{ .LIM }}`

	t := make(map[string]*template.Template)
	for k, v := range mm {
		tmpl, e := template.New(k).Parse(v)
		chke(e)
		t[k] = tmpl
	}
	return t
}

func basicprq(t PRQTemplate, prq PrerolledQuery) PrerolledQuery {
	// word in an author
	//
	//		SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value,
	//			marked_up_line, accented_line, stripped_line, hyphenated_words, annotations
	//			FROM lt0472 WHERE stripped_line ~* 'potest'  ORDER BY index ASC LIMIT 200

	// tail := `{{ .AU }} WHERE {{ .COL }} {{ .SYN }} '{{ .SK }}' ORDER BY index ASC LIMIT {{ .LIM }}`

	msg(t.Tail.Name(), 5)
	var b bytes.Buffer
	e := t.Tail.Execute(&b, t)
	chke(e)

	prq.PsqlQuery = fmt.Sprintf(SELECTFROM, b.String())
	return prq
}

func basicidxprq(t PRQTemplate, prq PrerolledQuery) PrerolledQuery {
	// word in a work
	//		SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value,
	//			marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt0472 WHERE stripped_line ~* 'nomen' AND (index BETWEEN 1 AND 2548) ORDER BY index ASC LIMIT 200

	// tail := `{{ .AU }} WHERE {{ .COL }} {{ .SYN }} '{{ .SK }}' AND ({{ .IDX }}) ORDER BY index ASC LIMIT {{ .LIM }}`

	msg(t.Tail.Name(), 5)
	var b bytes.Buffer
	e := t.Tail.Execute(&b, t)
	chke(e)

	prq.PsqlQuery = fmt.Sprintf(SELECTFROM, b.String())
	return prq

}

func basicwindowprq(t PRQTemplate, prq PrerolledQuery) PrerolledQuery {
	// phrase in an author
	//		SELECT second.wkuniversalid, second.index, second.level_05_value, second.level_04_value, second.level_03_value, second.level_02_value, second.level_01_value, second.level_00_value,
	//			second.marked_up_line, second.accented_line, second.stripped_line, second.hyphenated_words, second.annotations FROM
	//		( SELECT * FROM
	//			( SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations,
	//				concat(stripped_line, ' ', lead(stripped_line) OVER (ORDER BY index ASC) ) AS linebundle FROM lt0472 ) first
	//		) second WHERE second.linebundle ~* 'nomen esse' ORDER BY index ASC LIMIT 200

	//tail := ` FROM {{ .AU }} ) first
	//		) second WHERE second.linebundle {{ .SYN }} '{{ .SK }}' ORDER BY index ASC LIMIT {{ .LIM }}`

	msg(t.Tail.Name(), 5)
	var b bytes.Buffer
	e := t.Tail.Execute(&b, t)
	chke(e)

	alb := fmt.Sprintf(ASLINEBUNDLE, t.PSCol, t.PSCol)

	prq.PsqlQuery = fmt.Sprintf(PRFXSELFRM + alb + b.String())
	return prq
}

func windandidxprq(t PRQTemplate, prq PrerolledQuery) PrerolledQuery {
	// phrase within selections from the author
	// 		SELECT second.wkuniversalid, second.index, second.level_05_value, second.level_04_value, second.level_03_value, second.level_02_value, second.level_01_value, second.level_00_value,
	//			second.marked_up_line, second.accented_line, second.stripped_line, second.hyphenated_words, second.annotations FROM
	//		( SELECT * FROM
	//			( SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations,
	//				concat(stripped_line, ' ', lead(stripped_line) OVER (ORDER BY index ASC) ) AS linebundle FROM lt0474 WHERE (index BETWEEN 104798 AND 109397) OR (index BETWEEN 67552 AND 70014) ) first
	//			) second WHERE second.linebundle ~* 'causa esse' ORDER BY index ASC LIMIT 200

	// tail := ` FROM {{ .AU }} WHERE {{ .IDX }} ) first
	//		) second WHERE second.linebundle {{ .SYN }} '{{ .SK }}' ORDER BY index ASC LIMIT {{ .LIM }}`

	msg(t.Tail.Name(), 5)
	var b bytes.Buffer
	e := t.Tail.Execute(&b, t)
	chke(e)

	alb := fmt.Sprintf(ASLINEBUNDLE, t.PSCol, t.PSCol)

	prq.PsqlQuery = fmt.Sprintf(PRFXSELFRM + alb + b.String())

	return prq
}

func simplettprq(t PRQTemplate, prq PrerolledQuery) PrerolledQuery {
	// 	CREATE TEMPORARY TABLE lt0472_includelist_f5d653cfcdab44c6bfb662f688d47e73 AS
	//		SELECT values AS includeindex FROM
	//			unnest(ARRAY[2,3,4,5,6,7,8,9,...])
	//		values
	// (and then)
	//		SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value,
	//			marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM  lt0472 WHERE EXISTS
	//		(SELECT 1 FROM lt0472_includelist_f5d653cfcdab44c6bfb662f688d47e73 incl WHERE incl.includeindex = lt0472.index AND stripped_line ~* 'carm') LIMIT 200

	//tail := ` {{ .AU }} WHERE EXISTS
	//	(SELECT 1 FROM {{ .AU }}_includelist_{{ .TTN }} incl WHERE incl.includeindex = {{ .AU }}.index AND {{ .COL }} {{ .SYN }} '{{ .SK }}') LIMIT {{ .LIM }}`

	msg(t.Tail.Name(), 5)
	var b bytes.Buffer
	e := t.Tail.Execute(&b, t)
	chke(e)

	prq.PsqlQuery = fmt.Sprintf(SELECTFROM, b.String())
	return prq
}

func windowandttprq(t PRQTemplate, prq PrerolledQuery) PrerolledQuery {
	// 	CREATE TEMPORARY TABLE lt0893_includelist_fce25efdd0d4f4ecab77e636f8c512224 AS
	//		SELECT values AS includeindex FROM
	//			unnest(ARRAY[2,3,4,5,6,7,8,9,...])
	//		values
	// (and then)
	// 		SELECT second.wkuniversalid, second.index, second.level_05_value, second.level_04_value, second.level_03_value, second.level_02_value, second.level_01_value, second.level_00_value,
	//			second.marked_up_line, second.accented_line, second.stripped_line, second.hyphenated_words, second.annotations FROM
	//		( SELECT * FROM
	//			( SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations,
	//				concat(stripped_line, ' ', lead(stripped_line) OVER (ORDER BY index ASC) ) AS linebundle FROM lt0893 WHERE EXISTS
	//			(SELECT 1 FROM lt0893_includelist_ce25efdd0d4f4ecab77e636f8c512224 incl WHERE incl.includeindex = lt0893.index )
	//			) first
	//		) second WHERE second.linebundle ~* 'ad italos' LIMIT 200

	//tail := ` FROM {{ .AU }} WHERE EXISTS
	//		(SELECT 1 FROM {{ .AU }}_includelist_{{ .TTN }} incl WHERE incl.includeindex = {{ .AU }}.index )
	//		) first
	//	) second WHERE second.linebundle {{ .SYN }} '{{ .SK }}' LIMIT {{ .LIM }}`

	msg(t.Tail.Name(), 5)
	var b bytes.Buffer
	e := t.Tail.Execute(&b, t)
	chke(e)

	alb := fmt.Sprintf(ASLINEBUNDLE, t.PSCol, t.PSCol)

	prq.PsqlQuery = fmt.Sprintf(PRFXSELFRM + alb + b.String())
	return prq
}

func requiresindextemptable(au string, bb []Boundaries, ss *SearchStruct) string {
	// mimic wholeworktemptablecontents() in whereclauses.py
	m := fmt.Sprintf("%s requiresindextemptable(): %d []Boundaries", au, len(bb))
	msg(m, 4)
	var required []int64
	for _, b := range bb {
		for i := b.Start; i <= b.Stop; i++ {
			required = append(required, i)
		}
	}

	ctt := `
	CREATE TEMPORARY TABLE %s_includelist_%s AS 
		SELECT values AS includeindex FROM 
			unnest(ARRAY[%s])
		values`

	var arr []string
	for _, r := range required {
		arr = append(arr, strconv.FormatInt(r, 10))
	}
	a := strings.Join(arr, ",")
	ttsq := fmt.Sprintf(ctt, au, ss.TTName, a)

	return ttsq
}

func andorwhereclause(bounds []Boundaries, templ string, negation string, syntax string) string {
	// idxtmpl := `(index %sBETWEEN %d AND %d)` // %s is "" or "NOT "
	var in []string
	for _, v := range bounds {
		i := fmt.Sprintf(templ, negation, v.Start, v.Stop)
		in = append(in, i)
	}
	return strings.Join(in, syntax)
}

func test_searchlistintoqueries() {
	start := time.Now()
	previous := time.Now()

	var ss SearchStruct
	ss.Seeking = `dolore\s`
	ss.SrchColumn = "stripped_line"
	ss.SrchSyntax = "~*"
	ss.OrderBy = "index"
	ss.Limit = 200
	ss.SkgSlice = append(ss.SkgSlice, ss.Seeking)
	ss.SearchIn.Authors = []string{"lt0959", "lt0857"}
	ss.SearchIn.Works = []string{"lt0474w041", "lt0474w064"}
	ss.SearchIn.Passages = []string{"gr0032_FROM_11313_TO_11843", "lt0474_FROM_58578_TO_61085", "lt0474_FROM_36136_TO_36151"}
	prq := searchlistintoqueries(&ss)
	fmt.Println(prq)

	c := GetPSQLconnection()
	defer c.Close()
	var hits []DbWorkline
	for _, q := range prq {
		r := worklinequery(q, c)
		hits = append(hits, r...)
	}

	for i, h := range hits {
		t := fmt.Sprintf("%d - %s : %s", i, h.FindLocus(), h.MarkedUp)
		fmt.Println(t)
	}
	timetracker("-", "query built and executed", start, previous)
	c.Close()
}
