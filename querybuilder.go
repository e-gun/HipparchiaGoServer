package main

import (
	"fmt"
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
	CONCATSELFROM = `
		( SELECT * FROM
			( SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations,
				concat(stripped_line, ' ', lead(stripped_line) OVER (ORDER BY index ASC) ) 
			AS linebundle
				FROM %s 
					) first 
				) second`
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

	if ss.HasLemma {
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

	// TODO: anything that needs an unnest temptable can not pass through here yet...

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

	// TODO: here is where you could check for unnest temptable candidates: e.g., len(boundedincl[x]) > 80

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

	seltempl := SELECTFROM
	if ss.HasPhrase {
		seltempl = PRFXSELFRM + CONCATSELFROM
	}

	for _, a := range alltables {
		var qb QueryBuilder
		var prq PrerolledQuery
		qb.SelFrom = fmt.Sprintf(seltempl, a)

		// [b2] check to see if bounded
		if vv, found := boundedincl[a]; found {
			var in []string
			for _, v := range vv {
				i := fmt.Sprintf(idxtmpl, "", v.Start, v.Stop)
				in = append(in, i)
			}
			qb.WhrIdxInc = strings.Join(in, " OR ")
		}

		if vv, found := boundedexcl[a]; found {
			var in []string
			for _, v := range vv {
				i := fmt.Sprintf(idxtmpl, "NOT ", v.Start, v.Stop)
				in = append(in, i)
			}
			qb.WhrIdxExc = strings.Join(in, " AND ")
		}

		// [b3] search term might be lemmatized, hence the range

		// map to the %s items in the qtmpl below:
		// SELECTFROM + WHERETERM + WHEREINDEXINCL + WHEREINDEXEXCL + (either) ORDERBY&LIMIT (or) SECOND

		for _, skg := range ss.SkgSlice {
			var tail string
			if !ss.HasPhrase {
				// there is SECOND element
				qb.WhrTrm = fmt.Sprintf(`%s %s '%s' `, ss.SrchColumn, ss.SrchSyntax, skg)
				tail = ss.FmtOrderBy()
			} else {
				// in subqueryphrasesearch WHERETERM = "" since the search term comes after the "second" clause
				// in subqueryphrasesearch not ORDERBY&LIMIT but SECOND
				qb.WhrTrm = ""
				tail = subqueryphrasesearchtail(a, needstempt[a], skg, ss)
			}

			var qtmpl string
			if len(qb.WhrIdxInc) == 0 && len(qb.WhrIdxExc) == 0 {
				qtmpl = `%s WHERE %s %s%s%s`
			} else if len(qb.WhrIdxInc) != 0 && len(qb.WhrIdxExc) == 0 {
				// qtmpl = `%s WHERE %s AND ( %s ) %s%s`
				qtmpl = `%s WHERE %s ( %s ) AND %s%s`
			} else if len(qb.WhrIdxInc) == 0 && len(qb.WhrIdxExc) != 0 {
				qtmpl = `%s WHERE %s AND%s ( %s ) %s`
			} else if len(qb.WhrIdxInc) != 0 && len(qb.WhrIdxExc) != 0 {
				qtmpl = `%s WHERE %s AND ( %s ) AND ( %s ) %s`
			}

			prq.PsqlQuery = fmt.Sprintf(qtmpl, qb.SelFrom, qb.WhrTrm, qb.WhrIdxInc, qb.WhrIdxExc, tail)

			// fmt.Println(prq.PsqlQuery)
			prqq = append(prqq, prq)
		}
	}

	return prqq
}

func requiresindextemptable() {
	// test to see if there are too many "in0001wXXX" type entries
	// if there are, mimic wholeworktemptablecontents() in whereclauses.py
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

	// top bits provided above via: " qb.SelFrom = fmt.Sprintf(seltempl, a)"
	// baseq := PRFXSELFRM
	// bundq := CONCATSELFROM

	// so we need either [a] whr or [b] ttfirst + whr depending on hastt

	ttfirst := ""

	if hastt {
		tp := `
			EXISTS
				(SELECT 1 FROM %s_includelist_%s incl WHERE incl.includeindex = %s.index)`
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
