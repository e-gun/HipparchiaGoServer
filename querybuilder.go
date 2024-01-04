//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
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
)

type PrerolledQuery struct {
	TempTable string
	PsqlQuery string
}

type QueryBounds struct {
	Start int
	Stop  int
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
	TAILBASIC  = `{{ .AU }} WHERE {{ .COL }} {{ .SYN }} '{{ .SK }}' ORDER BY index ASC LIMIT {{ .LIM }}`
	TAILBASIDX = `{{ .AU }} WHERE {{ .COL }} {{ .SYN }} '{{ .SK }}' AND ({{ .IDX }}) ORDER BY index ASC LIMIT {{ .LIM }}`
	TAILBASWIN = ` FROM {{ .AU }} ) first
			) second WHERE second.linebundle {{ .SYN }} '{{ .SK }}' ORDER BY index ASC LIMIT {{ .LIM }}`
	TAILWINIDX = ` FROM {{ .AU }} WHERE {{ .IDX }} ) first 
			) second WHERE second.linebundle {{ .SYN }} '{{ .SK }}' ORDER BY index ASC LIMIT {{ .LIM }}`
	TAILTT = ` {{ .AU }} WHERE EXISTS
		(SELECT 1 FROM {{ .AU }}_includelist_{{ .TTN }} incl WHERE incl.includeindex = {{ .AU }}.index AND {{ .COL }} {{ .SYN }} '{{ .SK }}') LIMIT {{ .LIM }}`
	TAILWINTT = ` FROM {{ .AU }} WHERE EXISTS
			(SELECT 1 FROM {{ .AU }}_includelist_{{ .TTN }} incl WHERE incl.includeindex = {{ .AU }}.index ) 
			) first
		) second WHERE second.linebundle {{ .SYN }} '{{ .SK }}' LIMIT {{ .LIM }}`
)

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

// SSBuildQueries - populate a SearchStruct with []PrerolledQuery
func SSBuildQueries(s *SearchStruct) {
	const (
		REG = `(?P<auth>......)_FROM_(?P<start>\d+)_TO_(?P<stop>\d+)`
		IDX = `(index %sBETWEEN %d AND %d)` // %s is "" or "NOT "
	)
	// modifies the SearchStruct in place
	inc := s.SearchIn
	exc := s.SearchEx

	if len(s.LemmaOne) != 0 {
		s.SkgSlice = lemmaintoregexslice(s.LemmaOne)
	} else {
		s.SkgSlice = append(s.SkgSlice, s.Seeking)
	}

	syn := s.SrchSyntax
	if s.PhaseNum == 2 && s.NotNear {
		syn = "!~"
	}

	// if there are too many "in0001wXXX" type entries: requiresindextemptable()

	// au query looks like: SELECTFROM + WHERETERM + WHEREINDEX + ORDERBY&LIMIT

	// note that you can't use prepared statements: "SELECT ... FROM $1 WHERE stripped_line ~ $2" yields a syntax error.
	// that is, PREPARE prepares on a table and so cannot have a table name as a variable; accordingly the following
	// is valid, but not useful for our purposes outside lemmata searches (which are scattered across clients anyway):
	// "SELECT ... FROM gr0432 WHERE stripped_line ~ $1"

	// [au] figure out all bounded selections

	boundedincl := make(map[string][]QueryBounds)
	boundedexcl := make(map[string][]QueryBounds)

	// [a1] individual works included/excluded
	for _, w := range inc.Works {
		wk := AllWorks[w]
		b := QueryBounds{wk.FirstLine, wk.LastLine}
		boundedincl[wk.AuID()] = append(boundedincl[wk.AuID()], b)
	}

	for _, w := range exc.Works {
		wk := AllWorks[w]
		b := QueryBounds{wk.FirstLine, wk.LastLine}
		boundedexcl[wk.AuID()] = append(boundedexcl[wk.AuID()], b)
	}
	// fmt.Println(boundedincl) --> map[gr0545:[{13717 19042}]]

	// [a2] individual passages included/excluded

	pattern := regexp.MustCompile(REG)
	for _, p := range inc.Passages {
		// "gr0032_FROM_11313_TO_11843"
		// there is an "index out of range" panic you will see in here if "gr0028_FROM_-1_TO_5" arrives
		subs := pattern.FindStringSubmatch(p)
		au := subs[pattern.SubexpIndex("auth")]
		st, _ := strconv.Atoi(subs[pattern.SubexpIndex("start")])
		sp, _ := strconv.Atoi(subs[pattern.SubexpIndex("stop")])
		b := QueryBounds{st, sp}
		boundedincl[au] = append(boundedincl[au], b)
		// fmt.Printf("%s: %d - %d\n", au, st, sp)
	}

	for _, p := range exc.Passages {
		subs := pattern.FindStringSubmatch(p)
		au := subs[pattern.SubexpIndex("auth")]
		st, _ := strconv.Atoi(subs[pattern.SubexpIndex("start")])
		sp, _ := strconv.Atoi(subs[pattern.SubexpIndex("stop")])
		b := QueryBounds{st, sp}
		boundedexcl[au] = append(boundedexcl[au], b)
	}

	// [b] build the queries for the author tables

	// [b1] collapse inc.Authors, inc.Works, incl.Passages to find all tables in use
	// but the keys to boundedincl in fact gives you the answer to the latter two

	alltables := inc.Authors
	for t, _ := range boundedincl {
		alltables = append(alltables, t)
	}

	tails := acquiretails()

	prqq := make([]PrerolledQuery, len(alltables)*len(s.SkgSlice))
	count := 0

	type QueryBuilder struct {
		SelFrom   string
		WhrTrm    string
		WhrIdxInc string
		WhrIdxExc string
	}

	for _, au := range alltables {
		var qb QueryBuilder
		var prq PrerolledQuery

		// [b2a] check to see if bounded by inclusions
		if bb, found := boundedincl[au]; found {
			if len(bb) > TEMPTABLETHRESHOLD {
				prq.TempTable = requiresindextemptable(au, bb, s)
			} else {
				qb.WhrIdxInc = andorwhereclause(bb, IDX, "", " OR ")
			}
		}

		// [b2b] check to see if bounded by exclusions
		if bb, found := boundedexcl[au]; found {
			if len(bb) > TEMPTABLETHRESHOLD {
				// note that 200 incl + 200 excl will produce garbage; in practice you have only au ton of one of them
				prq.TempTable = requiresindextemptable(au, bb, s)
			} else {
				qb.WhrIdxExc = andorwhereclause(bb, IDX, "NOT ", " AND ")
			}
		}

		// [b3] search term might be lemmatized, hence the range

		for i, skg := range s.SkgSlice {
			sprq := prq
			// there are fancier ways to do this, but debugging and maintaining become overwhelming...

			// map to the %s items in the qtmpl below:
			// SELECTFROM + WHERETERM + WHEREINDEXINCL + WHEREINDEXEXCL + (either) ORDERBY&LIMIT (or) SECOND

			nott := len(prq.TempTable) == 0
			yestt := len(prq.TempTable) != 0
			noph := !s.HasPhraseBoxA
			yesphr := s.HasPhraseBoxA
			noidx := len(qb.WhrIdxExc) == 0 && len(qb.WhrIdxInc) == 0
			yesidx := len(qb.WhrIdxExc) != 0 || len(qb.WhrIdxInc) != 0

			// lemmata need unique tt names otherwise "ERROR: relation "gr5002_includelist_e83674d70344428bbb1feab0919bc2c6" already exists"
			// cbf6f9746f2a46d080aa988c8c6bfd16_0, cbf6f9746f2a46d080aa988c8c6bfd16_1, ...
			ntt := fmt.Sprintf("%s_%d", s.TTName, i)
			sprq.TempTable = strings.Replace(prq.TempTable, s.TTName, ntt, -1)

			var t PRQTemplate
			t.AU = au
			t.COL = s.SrchColumn
			t.SYN = syn
			t.LIM = fmt.Sprintf("%d", s.CurrentLimit)
			t.TTN = ntt
			t.PSCol = s.SrchColumn
			t.SK = skg

			if len(qb.WhrIdxExc) != 0 && len(qb.WhrIdxInc) != 0 {
				t.IDX = fmt.Sprintf("%s AND %s", qb.WhrIdxInc, qb.WhrIdxExc)
			} else {
				// safe because at least one of these is ""
				t.IDX = qb.WhrIdxInc + qb.WhrIdxExc
			}

			if nott && noph && noidx {
				t.Tail = tails["basic"]
				sprq = basicprq(t, sprq)
			} else if nott && noph && yesidx {
				// word in work(s)/passage(s): AND ( (index BETWEEN 481 AND 483) OR (index BETWEEN 501 AND 503) ... )
				t.Tail = tails["basic_and_indices"]
				sprq = basicidxprq(t, sprq)
			} else if nott && yesphr && noidx {
				t.Tail = tails["basic_window"]
				sprq = basicwindowprq(t, sprq)
			} else if nott && yesphr && yesidx {
				t.Tail = tails["window_with_indices"]
				sprq = windandidxprq(t, sprq)
			} else if yestt && noph {
				t.Tail = tails["simple_tt"]
				sprq = simplettprq(t, sprq)
			} else {
				t.Tail = tails["window_with_tt"]
				sprq = windowandttprq(t, sprq)
			}
			prqq[count] = sprq
			count += 1
		}
	}
	s.Queries = prqq
	SIUpdateTW <- SIKVi{s.ID, len(prqq)}
}

//
// HELPERS
//

func acquiretails() map[string]*template.Template {
	// this avoids recompiling them a bunch of times in a loop

	mm := make(map[string]string)
	mm["basic"] = TAILBASIC
	mm["basic_and_indices"] = TAILBASIDX
	mm["basic_window"] = TAILBASWIN
	mm["window_with_indices"] = TAILWINIDX
	mm["simple_tt"] = TAILTT
	mm["window_with_tt"] = TAILWINTT

	t := make(map[string]*template.Template)
	for k, v := range mm {
		tmpl, e := template.New(k).Parse(v)
		chke(e)
		t[k] = tmpl
	}
	return t
}

// basicprq - PrerolledQuery for a string in an author table as a whole
func basicprq(t PRQTemplate, prq PrerolledQuery) PrerolledQuery {
	//
	//		SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value,
	//			marked_up_line, accented_line, stripped_line, hyphenated_words, annotations
	//			FROM lt0472 WHERE stripped_line ~* 'potest'  ORDER BY index ASC LIMIT 200

	// tail := `{{ .AU }} WHERE {{ .COL }} {{ .SYN }} '{{ .SK }}' ORDER BY index ASC LIMIT {{ .LIM }}`

	var b bytes.Buffer
	e := t.Tail.Execute(&b, t)
	chke(e)

	prq.PsqlQuery = fmt.Sprintf(SELECTFROM, b.String())
	return prq
}

// basicidxprq - PrerolledQuery for a string in a subsection of an author table (word in a work, e.g.)
func basicidxprq(t PRQTemplate, prq PrerolledQuery) PrerolledQuery {
	//
	//		SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value,
	//			marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt0472 WHERE stripped_line ~* 'nomen' AND (index BETWEEN 1 AND 2548) ORDER BY index ASC LIMIT 200

	// tail := `{{ .AU }} WHERE {{ .COL }} {{ .SYN }} '{{ .SK }}' AND ({{ .IDX }}) ORDER BY index ASC LIMIT {{ .LIM }}`

	var b bytes.Buffer
	e := t.Tail.Execute(&b, t)
	chke(e)

	prq.PsqlQuery = fmt.Sprintf(SELECTFROM, b.String())
	return prq

}

// basicwindowprq - PrerolledQuery for a phrase in an author table as a whole (i.e., a string with a whitespace)
func basicwindowprq(t PRQTemplate, prq PrerolledQuery) PrerolledQuery {
	//
	//		SELECT second.wkuniversalid, second.index, second.level_05_value, second.level_04_value, second.level_03_value, second.level_02_value, second.level_01_value, second.level_00_value,
	//			second.marked_up_line, second.accented_line, second.stripped_line, second.hyphenated_words, second.annotations FROM
	//		( SELECT * FROM
	//			( SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations,
	//				concat(stripped_line, ' ', lead(stripped_line) OVER (ORDER BY index ASC) ) AS linebundle FROM lt0472 ) first
	//		) second WHERE second.linebundle ~* 'nomen esse' ORDER BY index ASC LIMIT 200

	//tail := ` FROM {{ .AU }} ) first
	//		) second WHERE second.linebundle {{ .SYN }} '{{ .SK }}' ORDER BY index ASC LIMIT {{ .LIM }}`

	var b bytes.Buffer
	e := t.Tail.Execute(&b, t)
	chke(e)

	alb := fmt.Sprintf(ASLINEBUNDLE, t.PSCol, t.PSCol)

	prq.PsqlQuery = fmt.Sprintf(PRFXSELFRM + alb + b.String())
	return prq
}

// windandidxprq - PrerolledQuery for a phrase within selections of an author table
func windandidxprq(t PRQTemplate, prq PrerolledQuery) PrerolledQuery {
	//
	// 		SELECT second.wkuniversalid, second.index, second.level_05_value, second.level_04_value, second.level_03_value, second.level_02_value, second.level_01_value, second.level_00_value,
	//			second.marked_up_line, second.accented_line, second.stripped_line, second.hyphenated_words, second.annotations FROM
	//		( SELECT * FROM
	//			( SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations,
	//				concat(stripped_line, ' ', lead(stripped_line) OVER (ORDER BY index ASC) ) AS linebundle FROM lt0474 WHERE (index BETWEEN 104798 AND 109397) OR (index BETWEEN 67552 AND 70014) ) first
	//			) second WHERE second.linebundle ~* 'causa esse' ORDER BY index ASC LIMIT 200

	// tail := ` FROM {{ .AU }} WHERE {{ .IDX }} ) first
	//		) second WHERE second.linebundle {{ .SYN }} '{{ .SK }}' ORDER BY index ASC LIMIT {{ .LIM }}`

	var b bytes.Buffer
	e := t.Tail.Execute(&b, t)
	chke(e)

	alb := fmt.Sprintf(ASLINEBUNDLE, t.PSCol, t.PSCol)

	prq.PsqlQuery = fmt.Sprintf(PRFXSELFRM + alb + b.String())

	return prq
}

// simplettprq - PrerolledQuery that involves a temporary table to generate author table selections (but not a phrase search)
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

	var b bytes.Buffer
	e := t.Tail.Execute(&b, t)
	chke(e)

	prq.PsqlQuery = fmt.Sprintf(SELECTFROM, b.String())
	return prq
}

// windowandttprq - PrerolledQuery that involves a temporary table to generate author table selections and is a phrase search
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

	var b bytes.Buffer
	e := t.Tail.Execute(&b, t)
	chke(e)

	alb := fmt.Sprintf(ASLINEBUNDLE, t.PSCol, t.PSCol)

	prq.PsqlQuery = fmt.Sprintf(PRFXSELFRM + alb + b.String())
	return prq
}

func requiresindextemptable(au string, bb []QueryBounds, ss *SearchStruct) string {
	const (
		MSG = "%s requiresindextemptable(): %d []QueryBounds"
		CTT = `
		CREATE TEMPORARY TABLE %s_includelist_%s AS 
			SELECT values AS includeindex FROM 
				unnest(ARRAY[%s])
			values`
	)

	msg(fmt.Sprintf(MSG, au, len(bb)), MSGTMI)
	var required []int
	for _, b := range bb {
		for i := b.Start; i <= b.Stop; i++ {
			required = append(required, i)
		}
	}

	var arr []string
	for _, r := range required {
		arr = append(arr, fmt.Sprintf("%d", r))
	}
	a := strings.Join(arr, ",")
	ttsq := fmt.Sprintf(CTT, au, ss.TTName, a)

	return ttsq
}

func andorwhereclause(bounds []QueryBounds, templ string, negation string, syntax string) string {
	// idxtmpl := `(index %sBETWEEN %d AND %d)` // %s is "" or "NOT "
	var in []string
	for _, v := range bounds {
		i := fmt.Sprintf(templ, negation, v.Start, v.Stop)
		in = append(in, i)
	}

	return strings.Join(in, syntax)
}
