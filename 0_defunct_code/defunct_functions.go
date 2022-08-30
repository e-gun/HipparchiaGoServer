package main

import (
	"fmt"
	"github.com/google/uuid"
	"regexp"
	"strconv"
	"strings"
)

// rt-search.go

func old_withinxlinessearch(s SearchStruct) SearchStruct {
	// after finding x, look for y within n lines of x

	// "decessionis" near "spem" in Cicero...

	// (part 1)
	//		HGoSrch(first)
	//
	// (part 2.1)
	//		CREATE TEMPORARY TABLE lt0474_includelist_24bfe76dc1124f07becabb389a4f393d AS
	//		SELECT values AS includeindex FROM
	//			unnest(ARRAY[39844,39845,39846,39847,39848,39849,39850,39851,39852,39853,128858,128859,128860,128861,128862,128863,128864,128865,128866,128867,138278,138279,138280,138281,138282,138283,138284,138285,138286,138287])
	//		values

	// (part 2.2)
	// 		SELECT second.wkuniversalid, second.index, second.level_05_value, second.level_04_value, second.level_03_value, second.level_02_value, second.level_01_value, second.level_00_value,
	//			second.marked_up_line, second.accented_line, second.stripped_line, second.hyphenated_words, second.annotations FROM
	//		( SELECT * FROM
	//			( SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations,
	//				concat(stripped_line, ' ', lead(stripped_line) OVER (ORDER BY index ASC) ) AS linebundle
	//				FROM
	//	lt0474 WHERE EXISTS
	//		(SELECT 1 FROM lt0474_includelist_24bfe76dc1124f07becabb389a4f393d incl
	//			WHERE incl.includeindex = lt0474.index)
	//					) first
	//				) second WHERE second.linebundle ~ 'spem' LIMIT 200;

	// alternate strategy, but not a universal solution to the various types of search linebundles can handle:
	// ... FROM lt0474 WHERE ( (index BETWEEN 128860 AND 128866) OR (index BETWEEN 39846 AND 39852) OR ... )

	// todo: it looks like we can't do "within 5 lines" this way since the bundle is too small; grablinebundles() instead

	first := s
	first.Limit = FIRSTSEARCHLIM
	first = HGoSrch(first)

	msg(fmt.Sprintf("withinxlinessearch(): %d initial hits", len(first.Results)), 4)

	// convert the hits into new selections:
	// a temptable will be built once you know which lines do you need from which works

	var required = make(map[string][]int64)
	for _, r := range first.Results {
		w := AllWorks[r.WkUID]
		var idx []int64
		for i := r.TbIndex - s.ProxVal; i < r.TbIndex+s.ProxVal; i++ {
			if i >= w.FirstLine && i <= w.LastLine {
				idx = append(idx, i)
			}
		}
		required[w.FindAuthor()] = append(required[w.FindAuthor()], idx...)
	}

	// prepare new search
	fss := first.SkgSlice

	second := first
	second.Results = []DbWorkline{}
	second.Queries = []PrerolledQuery{}
	second.TTName = strings.Replace(uuid.New().String(), "-", "", -1)
	second.SkgSlice = second.PrxSlice
	second.PrxSlice = fss
	second.Limit = s.Limit

	var ttsq = make(map[string]string)
	ctt := `
	CREATE TEMPORARY TABLE %s_includelist_%s AS 
		SELECT values AS includeindex FROM 
			unnest(ARRAY[%s])
		values`

	for r, vv := range required {
		var arr []string
		for _, v := range vv {
			arr = append(arr, strconv.FormatInt(v, 10))
		}
		a := strings.Join(arr, ",")
		ttsq[r] = fmt.Sprintf(ctt, r, second.TTName, a)
	}

	seltempl := PRFXSELFRM + CONCATSELFROM

	wha := `
	%s WHERE EXISTS 
		(SELECT 1 FROM %s_includelist_%s incl 
			WHERE incl.includeindex = %s.index)`
	whb := ` WHERE second.linebundle ~ '%s' LIMIT %d;`
	var prqq = make(map[string][]PrerolledQuery)
	for i, q := range second.SkgSlice {
		for r, _ := range required {
			var prq PrerolledQuery
			modname := second.TTName + fmt.Sprintf("_%d", i)
			prq.TempTable = strings.Replace(ttsq[r], second.TTName, modname, -1)
			whc := fmt.Sprintf(wha, r, r, modname, r)
			whd := fmt.Sprintf(whb, q, second.Limit)
			prq.PsqlQuery = fmt.Sprintf(seltempl, whc) + whd
			prqq[r] = append(prqq[r], prq)
		}
	}

	for _, q := range prqq {
		second.Queries = append(second.Queries, q...)
	}

	second = HGoSrch(second)

	// windows of lines come back: e.g., three lines that look like they match when only one matches [3131, 3132, 3133]
	// figure out which line is really the line with the goods

	phrasefinder := regexp.MustCompile(`[A-Za-zΑ-ΩϹα-ωϲ]\s[A-Za-zΑ-ΩϹα-ωϲ]`)

	if phrasefinder.MatchString(second.Seeking) {
		second.Results = findphrasesacrosslines(second)
	} else {
		second.Results = validatebundledhits(second)
	}

	return second
}

func validatebundledhits(ss SearchStruct) []DbWorkline {
	// if the second search term available in the window of lines?
	re := ss.Proximate
	if ss.LemmaTwo != "" {
		re = lemmaintoflatregex(ss.LemmaTwo)
	}

	find := regexp.MustCompile(re)

	var valid []DbWorkline
	for _, r := range ss.Results {
		li := columnpicker(ss.SrchColumn, r)
		if find.MatchString(li) {
			valid = append(valid, r)
		}
	}

	return valid
}

func lemmaintoflatregex(hdwd string) string {
	// a single regex string for all forms
	var re string
	if _, ok := AllLemm[hdwd]; !ok {
		msg(fmt.Sprintf("lemmaintoregexslice() could not find '%s'", hdwd), 1)
		return re
	}

	tp := `(^|\s)%s(\s|$)`
	lemm := AllLemm[hdwd].Deriv

	var bnd []string
	for _, l := range lemm {
		bnd = append(bnd, fmt.Sprintf(tp, l))
	}

	re = strings.Join(bnd, "|")

	return re
}
