package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type SearchStruct struct {
	User       string
	Seeking    string
	Proximate  string
	LemmaOne   string
	LemmaTwo   string
	Summary    string
	QueryType  string
	ProxScope  string // "lines" or "words"
	ProxType   string // "near" or "not near"
	IsVector   bool
	NeedsWhere bool
	SrchColumn string // almost always "stripped_line"
	SrchSyntax string // almost always "~="
	OrderBy    string // almost always "index" + ASC
	Limit      int64
	SkgSlice   []string // either just Seeking or a decomposed version of a Lemma's possibilities
	PrxSlice   []string
	SearchIn   SearchIncExl
	SearchEx   SearchIncExl
	Queries    []PrerolledQuery
	Results    []DbWorkline
	Launched   time.Time
	SearchSize int
}

func (s SearchStruct) FmtOrderBy() string {
	var ob string
	a := `ORDER BY %s ASC %s`
	b := `LIMIT %d`
	if s.Limit > 0 {
		c := fmt.Sprintf(b, s.Limit)
		ob = fmt.Sprintf(a, s.OrderBy, c)
	} else {
		ob = fmt.Sprintf(a, s.OrderBy, "")
	}
	return ob
}

func (s SearchStruct) FmtWhereTerm(t string) string {
	a := `%s %s '%s' `
	wht := fmt.Sprintf(a, s.SrchColumn, s.SrchSyntax, t)
	return wht
}

func (s SearchStruct) HasLemma() bool {
	if len(s.LemmaOne) > 0 || len(s.LemmaTwo) > 0 {
		return true
	} else {
		return false
	}
}

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
	SELECTFROM = `SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM %s`
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
	// if there are too many "in0001wXXX" type entries: requiresindextemptable()

	// a query looks like: SELECTFROM + WHERETERM + WHEREINDEX + ORDERBY&LIMIT
	// FROM gr0308 WHERE ( (index BETWEEN 138 AND 175) OR (index BETWEEN 471 AND 510) ) AND ( stripped_line ~* $1 )
	// FROM lt0917 WHERE ( (index BETWEEN 1 AND 8069) OR (index BETWEEN 8070 AND 8092) ) AND ( (index NOT BETWEEN 1431 AND 2193) ) AND ( stripped_line ~* $1 )

	// TODO: anything that needs an unnest temptable can not pass through here yet...

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
		// fmt.Printf("%s: %d - %d", au, st, sp)
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
		qb.SelFrom = fmt.Sprintf(SELECTFROM, a)
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
		for _, s := range ss.SkgSlice {
			qb.WhrTrm = ss.FmtWhereTerm(s)
			ob := ss.FmtOrderBy()
			var qtmpl string
			if len(qb.WhrIdxInc) == 0 && len(qb.WhrIdxExc) == 0 {
				// SELECTFROM + WHERETERM + WHEREINDEXINCL + WHEREINDEXEXCL + ORDERBY&LIMIT
				qtmpl = `%s WHERE %s %s%s%s`
			} else if len(qb.WhrIdxInc) != 0 && len(qb.WhrIdxExc) == 0 {
				qtmpl = `%s WHERE %s AND ( %s ) %s%s`
			} else if len(qb.WhrIdxInc) == 0 && len(qb.WhrIdxExc) != 0 {
				qtmpl = `%s WHERE %s AND%s ( %s ) %s`
			} else if len(qb.WhrIdxInc) != 0 && len(qb.WhrIdxExc) != 0 {
				qtmpl = `%s WHERE %s AND ( %s ) AND ( %s ) %s`
			}
			prq.PsqlQuery = fmt.Sprintf(qtmpl, qb.SelFrom, qb.WhrTrm, qb.WhrIdxInc, qb.WhrIdxExc, ob)
			prqq = append(prqq, prq)
		}
	}

	return prqq
}

func recalculatesearchstruct(ss SearchStruct) SearchStruct {
	var newss SearchStruct
	return newss
}

func requiresindextemptable() {
	// test to see if there are too many "in0001wXXX" type entries
	// if there are, mimic wholeworktemptablecontents() in whereclauses.py
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
}
