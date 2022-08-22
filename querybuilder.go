package main

type WhereClause struct {
	Table  string
	Type   string
	Bounds []Boundaries
	Omit   []Boundaries
	Temp   TempSQL
}

func (w WhereClause) HasWhere() bool {
	if len(w.Bounds) > 0 || len(w.Omit) > 0 || len(w.Temp.Query) > 0 {
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

// findselectionboundaries() stuff should all be handled when making the selections, not here
// and everything should be in the "gr0032w002_FROM_11313_TO_11843" format

// take a searchlist and its exceptionlist and convert them into a collection of sql queries

// the collection of work will be either map[string]PrerolledQuery []PrerolledQuery

// onehit should be checked early: "ORDER BY index ASC LIMIT 1"

// complications: temporary tables
// these are needed
// [a] to unnest an array of lines to search inside an inscription author
// [b] subqueryphrasesearching

func searchlistintoqueries(sl []string) []PrerolledQuery {

}

func configurewhereclausedata(sl []string, ww map[string]DbWork) {
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
