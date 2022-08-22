package main

import "fmt"

type SearchStruct struct {
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

type WhereClause struct {
	Table        string
	Type         string
	Bounds       []Boundaries
	Omit         []Boundaries
	WhindowWhere string
	IndexWhere   string
	FirstPass    string
	SecPass      string
	Temp         TempSQL
}

func (w WhereClause) HasWhere() bool {
	if len(w.Bounds) > 0 || len(w.Omit) > 0 || len(w.Temp.Query) > 0 || len(w.FirstPass) > 0 {
		return true
	} else {
		return false
	}
}

type Boundaries struct {
	SS [2]int64
}

type TempSQL struct {
	Data  []int64
	Query string
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

func searchlistintoqueries(sl []string, ss SearchStruct) []PrerolledQuery {
	var prq []PrerolledQuery

	// if there are too many "in0001wXXX" type entries: requiresindextemptable()

	// a query looks like: SELECTFROM + WHERETERM + WHEREINDEX
	// FROM gr0308 WHERE ( (index BETWEEN 138 AND 175) OR (index BETWEEN 471 AND 510) ) AND ( stripped_line ~* $1 )
	// FROM lt0917 WHERE ( (index BETWEEN 1 AND 8069) OR (index BETWEEN 8070 AND 8092) ) AND ( (index NOT BETWEEN 1431 AND 2193) ) AND ( stripped_line ~* $1 )

	return prq
}

func configureindexwhereclausedata(srch SearchStruct, sl []string) {
	// + wholeworktemptablecontents() and wholeworkbetweenclausecontents()

	// [a] whole author table

	// [b] no exclusion, just "between"
	// "between" example
	// Ultimately you need this to search Aristophanes' Birds and Clouds:
	//
	//		WHERE (index BETWEEN 2885 AND 4633) AND (index BETWEEN 7921 AND 9913)'
	// template = '(index BETWEEN {min} AND {max})'
	// whereclause = 'WHERE ' + ' AND '.join(wheres)

	// [c] an exclusion
}

func requiresindextemptable() {
	// test to see if there are too many "in0001wXXX" type entries
	// if there are, mimic wholeworktemptablecontents() in whereclauses.py
}
