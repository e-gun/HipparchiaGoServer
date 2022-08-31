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
				concat(stripped_line, ' ', lead(stripped_line) OVER (ORDER BY index ASC) ) AS linebundle`

	CONCATSELFROM = `
		( SELECT * FROM
			( SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations,
				concat(stripped_line, ' ', lead(stripped_line) OVER (ORDER BY index ASC) ) 
			AS linebundle
				FROM %s 
					) first 
				) second`
	WHEREXISTSSELECT = `
		%s WHERE EXISTS 
			(SELECT 1 FROM %s_includelist_%s incl WHERE incl.includeindex = %s.index AND %s ~ '%s') LIMIT %s;`
)

// findselectionboundaries() stuff should all be handled when making the selections, not here
// and everything should be in the "gr0032w002_FROM_11313_TO_11843" format

// take a searchlist and its exceptionlist and convert them into a collection of sql queries

// the collection of work will be either map[string]PrerolledQuery []PrerolledQuery

// onehit should be checked early: "ORDER BY index ASC LIMIT 1"

// complications: temporary tables
// these are needed
// [a] to unnest an array of lines to search inside an inscription author
// [b] subqueryphrasesearching

// search types
// [a] simple - basicprecomposedsqlsearcher()
// [b] simplelemma - basicprecomposedsqlsearcher()
// [c] proximate by words - generatepreliminaryhitlist() [via basicprecomposedsqlsearcher()] + basicprecomposedsqlsearcher()
// [d] proximate by lines - generatepreliminaryhitlist() [via basicprecomposedsqlsearcher()] + paredowntowithinxwords()
// [e] simple phrase - precomposedsqlsubqueryphrasesearch()
// [f] phrase and proximity - precomposedphraseandproximitysearch() + either precomposedsqlsubqueryphrasesearch() or basicprecomposedsqlsearcher()

// basicprecomposedsqlsearcher

// searchlistintosqldict()

func searchlistintoqueries(ss SearchStruct) []PrerolledQuery {
	var prqq []PrerolledQuery
	inc := ss.SearchIn
	exc := ss.SearchEx

	if len(ss.LemmaOne) != 0 {
		ss.SkgSlice = lemmaintoregexslice(ss.LemmaOne)
	} else {
		ss.SkgSlice = append(ss.SkgSlice, ss.Seeking)
	}

	// fmt.Println(inc)
	// fmt.Println(ss.QueryType)

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
		t.SK = ss.Seeking
		t.LIM = fmt.Sprintf("%d", ss.Limit)
		if ss.NotNear {
			t.IDX = qb.WhrIdxExc
		} else {
			t.IDX = qb.WhrIdxInc
		}
		t.TTN = ss.TTName

		// todo
		// problem remains with tt and lemma: WHERE accented_line ~* ''

		if nott && noph && noidx {
			msg("basic", 5)
			prq = basicprq(t, prq)
		} else if nott && noph && yesidx {
			// word in work(s)/passage(s): AND ( (index BETWEEN 481 AND 483) OR (index BETWEEN 501 AND 503) ... )
			msg("basic_and_indices", 5)
			prq = basicidxprq(t, prq)
		} else if nott && yesphr && noidx {
			msg("basic_window", 5)
			prq = basicwindowprq(t, prq)
		} else if nott && yesphr && yesidx {
			msg("window_with_indices", 5)
			prq = windandidxprq(t, prq)
		} else if yestt && noph {
			msg("simple_tt", 5)
			prq = simplettprq(t, prq)
		} else {
			msg("window_with_tt", 5)
			prq = windowandttprq(t, prq)
		}
		prqq = append(prqq, prq)
	}
	return prqq
}

type PRQTemplate struct {
	AU  string
	COL string
	SYN string
	SK  string
	LIM string
	IDX string
	TTN string
}

func basicprq(t PRQTemplate, prq PrerolledQuery) PrerolledQuery {
	// word in an author
	//
	//		SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value,
	//			marked_up_line, accented_line, stripped_line, hyphenated_words, annotations
	//			FROM lt0472 WHERE stripped_line ~* 'potest'  ORDER BY index ASC LIMIT 200

	tail := `{{ .AU }} WHERE {{ .COL }} {{ .SYN }} '{{ .SK }}' ORDER BY index ASC LIMIT {{ .LIM }}`

	tmpl, e := template.New("b").Parse(tail)
	chke(e)
	var b bytes.Buffer
	e = tmpl.Execute(&b, t)
	chke(e)

	prq.PsqlQuery = fmt.Sprintf(SELECTFROM, b.String())
	return prq
}

func basicidxprq(t PRQTemplate, prq PrerolledQuery) PrerolledQuery {
	// word in a work
	//		SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value,
	//			marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt0472 WHERE stripped_line ~* 'nomen' AND (index BETWEEN 1 AND 2548) ORDER BY index ASC LIMIT 200

	tail := `{{ .AU }} WHERE {{ .COL }} {{ .SYN }} '{{ .SK }}' AND ({{ .IDX }}) ORDER BY index ASC LIMIT {{ .LIM }}`

	tmpl, e := template.New("bi").Parse(tail)
	chke(e)
	var b bytes.Buffer
	e = tmpl.Execute(&b, t)
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

	tail := ` FROM {{ .AU }} ) first 
			) second WHERE second.linebundle {{ .SYN }} '{{ .SK }}' ORDER BY index ASC LIMIT {{ .LIM }}`

	tmpl, e := template.New("bw").Parse(tail)
	chke(e)
	var b bytes.Buffer
	e = tmpl.Execute(&b, t)
	chke(e)

	prq.PsqlQuery = fmt.Sprintf(PRFXSELFRM + ASLINEBUNDLE + b.String())
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

	tail := ` FROM {{ .AU }} WHERE {{ .IDX }} ) first 
			) second WHERE second.linebundle {{ .SYN }} '{{ .SK }}' ORDER BY index ASC LIMIT {{ .LIM }}`

	tmpl, e := template.New("wdx").Parse(tail)
	chke(e)
	var b bytes.Buffer
	e = tmpl.Execute(&b, t)
	chke(e)

	prq.PsqlQuery = fmt.Sprintf(PRFXSELFRM + ASLINEBUNDLE + b.String())

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

	tail := ` {{ .AU }} WHERE EXISTS
		(SELECT 1 FROM {{ .AU }}_includelist_{{ .TTN }} incl WHERE incl.includeindex = {{ .AU }}.index AND {{ .COL }} {{ .SYN }} '{{ .SK }}') LIMIT {{ .LIM }}`

	tmpl, e := template.New("stt").Parse(tail)
	chke(e)
	var b bytes.Buffer
	e = tmpl.Execute(&b, t)
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

	tail := ` FROM {{ .AU }} WHERE EXISTS
			(SELECT 1 FROM {{ .AU }}_includelist_{{ .TTN }} incl WHERE incl.includeindex = {{ .AU }}.index ) 
			) first
		) second WHERE second.linebundle {{ .SYN }} '{{ .SK }}' LIMIT {{ .LIM }}`

	tmpl, e := template.New("wtt").Parse(tail)
	chke(e)
	var b bytes.Buffer
	e = tmpl.Execute(&b, t)
	chke(e)

	prq.PsqlQuery = fmt.Sprintf(PRFXSELFRM + ASLINEBUNDLE + b.String())
	return prq
}

func OLD_searchlistintoqueries(ss SearchStruct) []PrerolledQuery {
	var prqq []PrerolledQuery
	inc := ss.SearchIn
	exc := ss.SearchEx

	if len(ss.LemmaOne) != 0 {
		ss.SkgSlice = lemmaintoregexslice(ss.LemmaOne)
	} else {
		ss.SkgSlice = append(ss.SkgSlice, ss.Seeking)
	}

	// fmt.Println(inc)
	// fmt.Println(ss.QueryType)

	// if there are too many "in0001wXXX" type entries: requiresindextemptable()

	// a query looks like: SELECTFROM + WHERETERM + WHEREINDEX + ORDERBY&LIMIT
	// FROM gr0308 WHERE ( (index BETWEEN 138 AND 175) OR (index BETWEEN 471 AND 510) ) AND ( stripped_line ~* $1 )
	// FROM lt0917 WHERE ( (index BETWEEN 1 AND 8069) OR (index BETWEEN 8070 AND 8092) ) AND ( (index NOT BETWEEN 1431 AND 2193) ) AND ( stripped_line ~* $1 )

	needstempt := make(map[string]bool)

	for _, a := range AllAuthors {
		needstempt[a.UID] = false
	}

	// [a] figure out all bounded selections

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

	for _, a := range alltables {
		var qb QueryBuilder
		var prq PrerolledQuery

		// [b2] check to see if bounded
		if bb, found := boundedincl[a]; found {
			if len(bb) > TEMPTABLETHRESHOLD {
				needstempt[a] = true
				prq.TempTable = requiresindextemptable(a, bb, ss)
			} else {
				qb.WhrIdxInc = andorwhereclause(bb, idxtmpl, "", " OR ")
			}
		}

		if bb, found := boundedexcl[a]; found {
			if len(bb) > TEMPTABLETHRESHOLD {
				// note that 200 incl + 200 excl will produce garbage; in practice you have only a ton of one of them
				needstempt[a] = true
				prq.TempTable = requiresindextemptable(a, bb, ss)
			} else {
				qb.WhrIdxExc = andorwhereclause(bb, idxtmpl, "NOT ", " AND ")
			}
		}

		// [b3] search term might be lemmatized, hence the range

		// map to the %s items in the qtmpl below:
		// SELECTFROM + WHERETERM + WHEREINDEXINCL + WHEREINDEXEXCL + (either) ORDERBY&LIMIT (or) SECOND

		for _, skg := range ss.SkgSlice {
			var tail string
			if ss.HasPhrase {
				// in subqueryphrasesearch WHERETERM = "" since the search term comes after the "second" clause
				// in subqueryphrasesearch not ORDERBY&LIMIT but SECOND
				qb.WhrTrm = ""
				tail = subqueryphrasesearchtail(a, needstempt[a], skg, ss)
			} else {
				// there is SECOND element
				qb.WhrTrm = fmt.Sprintf(`%s %s '%s' `, ss.SrchColumn, ss.SrchSyntax, skg)
				tail = ss.FmtOrderBy()
			}

			var qtmpl string
			if len(qb.WhrIdxInc) == 0 && len(qb.WhrIdxExc) == 0 {
				qtmpl = `%s WHERE %s %s%s%s`
			} else if len(qb.WhrIdxInc) != 0 && len(qb.WhrIdxExc) == 0 {
				// kludgy and should be regularized some day...
				if ss.HasPhrase {
					qtmpl = `%s WHERE %s ( %s ) AND %s%s`
				} else {
					qtmpl = `%s WHERE %s AND ( %s ) %s%s`
				}
			} else if len(qb.WhrIdxInc) == 0 && len(qb.WhrIdxExc) != 0 {
				qtmpl = `%s WHERE %s AND%s ( %s ) %s`
			} else if len(qb.WhrIdxInc) != 0 && len(qb.WhrIdxExc) != 0 {
				qtmpl = `%s WHERE %s AND ( %s ) AND ( %s ) %s`
			}

			seltempl := SELECTFROM
			if len(prq.TempTable) > 0 && !ss.HasPhrase {
				// SELECT * lt0448  WHERE EXISTS (SELECT 1 from lt0448_includelist incl WHERE .. ) LIMIT 200;
				seltempl = SELECTFROM
				qtmpl = WHEREXISTSSELECT
			} else if ss.HasPhrase {
				seltempl = PRFXSELFRM + CONCATSELFROM
			}

			qb.SelFrom = fmt.Sprintf(seltempl, a)
			prq.PsqlQuery = fmt.Sprintf(qtmpl, qb.SelFrom, qb.WhrTrm, qb.WhrIdxInc, qb.WhrIdxExc, tail)

			//fmt.Println(prq)
			prqq = append(prqq, prq)
		}
	}
	return prqq
}

func requiresindextemptable(au string, bb []Boundaries, ss SearchStruct) string {
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

func subqueryphrasesearchtail(au string, hastt bool, skg string, ss SearchStruct) string {
	// a more general note...
	//
	// we use subquery syntax to grab multi-line windows of text for phrase searching
	//
	//    line ends and line beginning issues can be overcome this way, but then you have plenty of
	//    bookkeeping to do to get the proper results focussed on the right line (TODO: bookkeeping)
	//
	//    these searches take linear time: same basic time for any given scope regardless of the query

	// "dolore omni " in all of Lucretius:

	// 	SELECT second.wkuniversalid, second.index, second.level_05_value, second.level_04_value, second.level_03_value, second.level_02_value, second.level_01_value, second.level_00_value,
	//			second.marked_up_line, second.accented_line, second.stripped_line, second.hyphenated_words, second.annotations FROM
	//		( SELECT * FROM
	//			( SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations,
	//				concat(stripped_line, ' ', lead(stripped_line) OVER (ORDER BY index ASC) ) AS linebundle
	//				FROM lt0550
	//					) first
	//				) second WHERE  second.linebundle ~ 'dolore omni ' LIMIT 200
	//
	// in 3 works of Cicero is the same but the WHERE clause has been added
	//     SELECT second.wkuniversalid, second.index, second.level_05_value, second.level_04_value, second.level_03_value, second.level_02_value, second.level_01_value, second.level_00_value, second.marked_up_line, second.accented_line, second.stripped_line, second.hyphenated_words, second.annotations FROM
	//        ( SELECT * FROM
	//            ( SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations, concat(stripped_line, ' ', lead(stripped_line) OVER (ORDER BY index ASC) ) AS linebundle
	//                FROM lt0474 WHERE ( (index BETWEEN 1 AND 1052) OR (index BETWEEN 1053 AND 2631) OR (index BETWEEN 3911 AND 16459) ) ) first
	//        ) second
	//    WHERE second.linebundle ~ $1  LIMIT 200

	// »ἐν τῆι βουλῆι« in inscriptions from 1CE to 150CE: the TT has been added & the where clause is complicated by this

	// [a] 	CREATE TEMPORARY TABLE in0c17_includelist_bfc1d910ba3e4f6d8670e530f89ecdda AS
	//		SELECT values
	//			AS includeindex FROM unnest(ARRAY[34903,34904,34905,34906,34907,34908,34909,34910]) values
	// [b]      SELECT second.wkuniversalid, second.index, second.level_05_value, second.level_04_value, second.level_03_value, second.level_02_value, second.level_01_value, second.level_00_value, second.marked_up_line, second.accented_line, second.stripped_line, second.hyphenated_words, second.annotations FROM
	//        ( SELECT * FROM
	//            ( SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations, concat(accented_line, ' ', lead(accented_line) OVER (ORDER BY index ASC) ) AS linebundle
	//                FROM in0c17 WHERE
	//            EXISTS
	//                (SELECT 1 FROM in0c17_includelist_bfc1d910ba3e4f6d8670e530f89ecdda incl WHERE incl.includeindex = in0c17.index
	//             ) ) first
	//        ) second
	//    WHERE second.linebundle ~ $1  LIMIT 200

	//	lt0474 WHERE EXISTS
	//		(SELECT 1 FROM lt0474_includelist_24bfe76dc1124f07becabb389a4f393d incl
	//			WHERE incl.includeindex = lt0474.index)
	//					) first
	//				) second WHERE second.linebundle ~ 'spem' LIMIT 200;

	// top bits provided above via: " qb.SelFrom = fmt.Sprintf(seltempl, a)"
	// baseq := PRFXSELFRM
	// bundq := CONCATSELFROM

	// so we need either [a] whr or [b] ttfirst + whr depending on hastt

	ttfirst := ""

	if hastt {
		tp := `
			EXISTS
				(SELECT 1 FROM %s_includelist_%s incl WHERE incl.includeindex = %s.index) 
				) first 
			) second WHERE `
		ttfirst = fmt.Sprintf(tp, au, ss.TTName, au)
	}

	w := `second.linebundle ~ '%s' LIMIT %d`

	whr := fmt.Sprintf(w, skg, ss.Limit)

	return ttfirst + whr
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
	prq := searchlistintoqueries(ss)
	fmt.Println(prq)

	c := grabpgsqlconnection()
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

/*
[a] word in an author

		SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value,
			marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt0472 WHERE stripped_line ~* 'potest'  ORDER BY index ASC LIMIT 200

[b] word in a work

		SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value,
			marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt0448 WHERE stripped_line ~* 'potest'  AND ( (index BETWEEN 1 AND 6192) ) ORDER BY index ASC LIMIT 200

[c] word near word [no temptable]
		SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value,
			marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt0472 WHERE stripped_line ~* 'erat'  AND ( (index BETWEEN 481 AND 483) OR (index BETWEEN 501 AND 503) ... ) ORDER BY index ASC LIMIT 200

[d] phrase in an author

		SELECT second.wkuniversalid, second.index, second.level_05_value, second.level_04_value, second.level_03_value, second.level_02_value, second.level_01_value, second.level_00_value,
			second.marked_up_line, second.accented_line, second.stripped_line, second.hyphenated_words, second.annotations FROM
		( SELECT * FROM
			( SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations,
				concat(stripped_line, ' ', lead(stripped_line) OVER (ORDER BY index ASC) )
			AS linebundle
				FROM lt0472
					) first
				) second WHERE  second.linebundle ~ 'aut quid ' LIMIT 200

[e] phrase near word
	= [d] +
	[WRONG][todo: the right way is a simple "WHERE stripped_line ~* 'potest'  AND ( (index BETWEEN 501 AND 503) ) LIMIT 200"]

	SELECT second.wkuniversalid, second.index, second.level_05_value, second.level_04_value, second.level_03_value, second.level_02_value, second.level_01_value, second.level_00_value,
			second.marked_up_line, second.accented_line, second.stripped_line, second.hyphenated_words, second.annotations FROM
		( SELECT * FROM
			( SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations,
				concat(stripped_line, ' ', lead(stripped_line) OVER (ORDER BY index ASC) )
			AS linebundle
				FROM lt0472
					) first
				) second WHERE  ( (index BETWEEN 501 AND 503) ) AND second.linebundle ~ 'uncta' LIMIT 200


[f] phrase near phrase
	= [d] +
			SELECT second.wkuniversalid, second.index, second.level_05_value, second.level_04_value, second.level_03_value, second.level_02_value, second.level_01_value, second.level_00_value,
			second.marked_up_line, second.accented_line, second.stripped_line, second.hyphenated_words, second.annotations FROM
		( SELECT * FROM
			( SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations,
				concat(stripped_line, ' ', lead(stripped_line) OVER (ORDER BY index ASC) )
			AS linebundle
				FROM lt0472
					) first
				) second WHERE  ( (index BETWEEN 501 AND 503) ) AND second.linebundle ~ 'nisi uncta' LIMIT 200

[g] lemma
SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value,
			marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt0472 WHERE accented_line ~* '(^|\s)nominisque(\s|$)|(^|\s)nominum(\s|$)|...'  ORDER BY index ASC LIMIT 200


[h] lemma near word - tt
		SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value,
			marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt0472 WHERE accented_line ~* 'unice'  AND ( (index BETWEEN 491 AND 493) OR (index BETWEEN 503 AND 505) OR... ) ORDER BY index ASC LIMIT 200

[i] lemma near word + tt
	= [g] +
	[WRONG][todo: (SELECT 1 FROM lt0472_includelist_f79d14b56d1c481c8bbdeadc4d2e23c0 incl WHERE incl.includeindex = lt0472.index AND stripped_line ~ 'unice') LIMIT 200;
	CREATE TEMPORARY TABLE lt0472_includelist_755d8184cf6840dd8e6afab0bd34f535 AS
		SELECT values AS includeindex FROM
			unnest(ARRAY[491,492,493,503,504,505,576,577,578,1150,1151,1152,1214,1215,1216,2006,2007,2008,2009,2010,2011,2063,2064,2065,2135,2136,2137,2167,2168,2169])
		values

		SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value,
			marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt0472 WHERE EXISTS
			(SELECT 1 FROM accented_line ~* 'unice' _includelist_ incl WHERE incl.includeindex = .index AND ORDER BY index ASC LIMIT 200 ~ '%!s(MISSING)') LIMIT %!s(MISSING);

[j] lemma near phrase + tt
	= [g] +
	[WRONG]
	CREATE TEMPORARY TABLE lt0472_includelist_380505330f364d9fa40024d16d3191bf AS
		SELECT values AS includeindex FROM
			unnest(ARRAY[491,492,493,503,504,505,576,577,578,1150,1151,1152,1214,1215,1216,2006,2007,2008,2009,2010,2011,2063,2064,2065,2135,2136,2137,2167,2168,2169])
		values
	+

		SELECT second.wkuniversalid, second.index, second.level_05_value, second.level_04_value, second.level_03_value, second.level_02_value, second.level_01_value, second.level_00_value,
			second.marked_up_line, second.accented_line, second.stripped_line, second.hyphenated_words, second.annotations FROM
		( SELECT * FROM
			( SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations,
				concat(stripped_line, ' ', lead(stripped_line) OVER (ORDER BY index ASC) )
			AS linebundle
				FROM lt0472
					) first
				) second WHERE
			EXISTS
				(SELECT 1 FROM lt0472_includelist_380505330f364d9fa40024d16d3191bf incl WHERE incl.includeindex = lt0472.index)
				) first
			) second WHERE second.linebundle ~ 'in tabulas' LIMIT 200

	[todo: make it look like this...]
		SELECT second.wkuniversalid, second.index, second.level_05_value, second.level_04_value, second.level_03_value, second.level_02_value, second.level_01_value, second.level_00_value,
			second.marked_up_line, second.accented_line, second.stripped_line, second.hyphenated_words, second.annotations FROM
		( SELECT * FROM
			( SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations,
				concat(stripped_line, ' ', lead(stripped_line) OVER (ORDER BY index ASC) )
			AS linebundle
				FROM lt0472 WHERE
			EXISTS
				(SELECT 1 FROM lt0472_includelist_380505330f364d9fa40024d16d3191bf incl WHERE incl.includeindex = lt0472.index)
				) first
			) second WHERE second.linebundle ~ 'in tabulas' LIMIT 200

*/
