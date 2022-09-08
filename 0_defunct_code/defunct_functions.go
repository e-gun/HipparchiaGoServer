package main

import (
	"bufio"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// buildquery.go

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

func blankconfig() CurrentConfiguration {
	// need a non-commandline config
	var thecfg CurrentConfiguration
	thecfg.PGLogin.Port = PSDefaultPort
	// cfg.PGLogin.Pass = cfg.PSQP
	thecfg.PGLogin.Pass = ""
	thecfg.PGLogin.User = PSDefaultUser
	thecfg.PGLogin.DBName = PSDefaultDB
	thecfg.PGLogin.Host = PSDefaultHost
	return thecfg
}

func parsesleectvals(r *http.Request) SelectionValues {
	// https://golangcode.com/get-a-url-parameter-from-a-request/
	// https://stackoverflow.com/questions/41279297/how-to-get-all-query-parameters-from-go-gin-context-object
	// gin: You should be able to do c.Request.URL.Query() which will return a Values which is a map[string][]string

	// TODO: check this stuff for bad characters
	// but 'auth', etc. can be parsed just by checking them against known author lists

	var sv SelectionValues

	kvp := r.URL.Query() // map[string][]string

	if _, ok := kvp["auth"]; ok {
		sv.Auth = kvp["auth"][0]
	}

	if _, ok := kvp["work"]; ok {
		sv.Work = kvp["work"][0]
	}

	if _, ok := kvp["locus"]; ok {
		sv.Start = kvp["locus"][0]
	}

	if _, ok := kvp["endpoint"]; ok {
		sv.End = kvp["endpoint"][0]
	}

	if _, ok := kvp["genre"]; ok {
		sv.AGenre = kvp["genre"][0]
	}

	if _, ok := kvp["wkgenre"]; ok {
		sv.WGenre = kvp["wkgenre"][0]
	}

	if _, ok := kvp["auloc"]; ok {
		sv.ALoc = kvp["auloc"][0]
	}

	if _, ok := kvp["wkprov"]; ok {
		sv.WLoc = kvp["wkprov"][0]
	}

	if _, ok := kvp["exclude"]; ok {
		if kvp["exclude"][0] == "t" {
			sv.IsExcl = true
		} else {
			sv.IsExcl = false
		}
	}

	return sv
}

func cleanSrchFeeder(ctx context.Context, name string, qq []PrerolledQuery) (<-chan PrerolledQuery, error) {
	emitqueries := make(chan PrerolledQuery, cfg.WorkerCount)
	//remainder := -1
	//host := progresssocket("pp_" + name)

	// channel emitter: i.e., the actual work
	go func() {
		defer close(emitqueries)
		for _, q := range qq {
			// fmt.Println(q)
			select {
			case <-ctx.Done():
				return
			default:
				// remainder = len(qq) - i - 1
				emitqueries <- q
			}
		}
	}()

	// tcp remainder broadcaster: i.e., the fluff

	//go func() {
	//	// cf https://notes.shichao.io/gopl/ch8/
	//	// [a] open a tcp port to broadcast on
	//	if host == nil {
	//		msg("progresssocket() could not open any ports", 1)
	//		return
	//	}
	//
	//	for {
	//		// [b] wait for someone to listen
	//		guest, err := host.Accept()
	//		if err != nil {
	//			continue
	//		}
	//		go func() {
	//			// send remainder value to it
	//			defer guest.Close()
	//			for {
	//				if remainder == 0 {
	//					// https://stackoverflow.com/questions/61049648/getting-bind-address-already-in-use-even-after-closing-the-connection-in-golang
	//					// "This connection, which is in TIME_WAIT state, can block further use of the port, making it
	//					// impossible to create a new listener, unless you give the right underlying settings to the host OS..."
	//					// that's the issue here:
	//					_, err := io.WriteString(guest, fmt.Sprintf("%d\n", remainder))
	//					chke(err)
	//					guest.Close()
	//					host.Close()
	//					break
	//				} else if remainder > -1 {
	//					// msg(fmt.Sprintf("remain: %d", remainder), 1)
	//					_, err := io.WriteString(guest, fmt.Sprintf("%d\n", remainder))
	//					if err != nil {
	//						return // e.g., client disconnected
	//					}
	//					time.Sleep(300)
	//				}
	//			}
	//		}()
	//	}
	//}()

	return emitqueries, nil
}

func cleanResultCollation(ctx context.Context, name string, max int64, values <-chan []DbWorkline) []DbWorkline {
	var allhits []DbWorkline
	//done := false
	//host := progresssocket("rc_" + name)
	for {
		select {
		case <-ctx.Done():
			log.Print(ctx.Err().Error())
			return allhits
		case val, ok := <-values:
			if ok {
				// the progress poll should be attached here
				// fmt.Println(fmt.Sprintf("current count: %d", len(allhits)))
				allhits = append(allhits, val...)
				if int64(len(allhits)) > max {
					// you popped over the cap...: this does in fact save time and exit in the middle
					// προκατελαβον cap of one: [Δ: 0.112s] HGoSrch()
					// προκατελαβον uncapped:   [Δ: 1.489s] HGoSrch()
					return allhits
				}
			} else {
				// rudundant?
				return allhits
			}
		}

		// tcp hits broadcaster: i.e., the fluff
		//go func() {
		//	// cf https://notes.shichao.io/gopl/ch8/
		//	// [a] open a tcp port to broadcast on
		//
		//	for {
		//		// [b] wait for someone to listen
		//		guest, err := host.Accept()
		//		if err != nil {
		//			continue
		//		}
		//		go func() {
		//			// send remainder value to it
		//			for {
		//				_, err := io.WriteString(guest, fmt.Sprintf("%d\n", len(allhits)))
		//				if err != nil {
		//					guest.Close()
		//					break
		//				}
		//				if done == true {
		//					guest.Close()
		//					break
		//				}
		//			}
		//		}()
		//	}
		//}()

	}
}

func ODLSrchFeeder(ctx context.Context, host net.Listener, ss *SearchStruct) (<-chan PrerolledQuery, error) {
	emitqueries := make(chan PrerolledQuery, cfg.WorkerCount)
	remainder := -1
	ctxclosed := false

	// channel emitter: i.e., the actual work
	go func() {
		defer close(emitqueries)
		for i, q := range ss.Queries {
			// fmt.Println(q)
			select {
			case <-ctx.Done():
				ctxclosed = true
				break
			default:
				remainder = len(ss.Queries) - i - 1
				emitqueries <- q
			}
		}
	}()

	// unix socket remainder broadcaster: i.e., the fluff
	go func() {
		if host == nil {
			msg("progresssocket() has no access to a socket", 1)
			return
		}

		for {
			if ctxclosed || remainder == 0 {
				break
			}

			// [a] wait for someone to listen
			guest, err := host.Accept()
			if err != nil {
				continue
			}

			go func() {
				// send remainder value to it
				for {
					if ctxclosed || remainder == 0 {
						_, ioe := io.WriteString(guest, fmt.Sprintf("%d\n", remainder))
						chke(ioe)
						break
					} else if !ctxclosed && remainder > -1 {
						// msg(fmt.Sprintf("remain: %d", remainder), 1)
						_, ioe := io.WriteString(guest, fmt.Sprintf("%d\n", remainder))
						if ioe != nil {
							break // e.g., client disconnected
						}
					}
				}
			}()
		}
	}()

	return emitqueries, nil
}

func OLDResultCollation(ctx context.Context, host net.Listener, max int64, values <-chan []DbWorkline) []DbWorkline {
	var allhits []DbWorkline
	done := false
	for {
		// the actual substance of the thing
		if done {
			break
		}
		select {
		case <-ctx.Done():
			log.Print(ctx.Err().Error())
			done = true
		case val, ok := <-values:
			if ok {
				allhits = append(allhits, val...)
				if int64(len(allhits)) > max {
					// you popped over the cap...: this does in fact save time and exit in the middle
					// προκατελαβον cap of one: [Δ: 0.112s] HGoSrch()
					// προκατελαβον uncapped:   [Δ: 1.489s] HGoSrch()
					done = true
				}
			} else {
				done = true
			}
		}

		// unix socket hits broadcaster: i.e., the fluff
		go func() {
			// [a] open a tcp port to broadcast on

			for {
				if done == true {
					break
				}
				// [b] wait for someone to listen
				guest, err := host.Accept()
				if err != nil {
					continue
				}
				go func() {
					// send remainder value to it
					for {
						_, err := io.WriteString(guest, fmt.Sprintf("%d\n", len(allhits)))
						if err != nil {
							break
						}
						if done == true {
							break
						}
					}
				}()
			}
		}()

	}
	return allhits
}

func OLD_RtWebsocket(c echo.Context) error {
	// 	the client sends the name of a poll and this will output
	//	the status of the poll continuously while the poll remains active
	//
	//	example:
	//		progress {'active': 1, 'total': 20, 'remaining': 20, 'hits': 48, 'message': 'Putting the results in context',
	//		'elapsed': 14.0, 'extrainfo': '<span class="small"></span>'}

	// see also /static/hipparchiajs/progressindicator_go.js

	// https://echo.labstack.com/cookbook/websocket/

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	type ReplyJS struct {
		Active   string `json:"active"`
		TotalWrk int    `json:"Poolofwork"`
		Remain   int    `json:"Remaining"`
		Hits     int    `json:"Hitcount"`
		Msg      string `json:"Statusmessage"`
		Elapsed  string `json:"Elapsed"`
		Extra    string `json:"Notes"`
		ID       string `json:"ID"`
	}

	for {
		if len(searches) != 0 {
			break
		}
	}

	done := false
	for {
		if done {
			break
		}
		// Read
		m := []byte{}
		_, m, e := ws.ReadMessage()
		if e != nil {
			// c.Logger().Error(err)
			msg("RtWebsocket(): ws failed to read: breaking", 1)
			break
		}

		// will yield: websocket received: "205da19d"
		// the bug-trap is the quotes around that string
		bs := string(m)
		bs = strings.Replace(bs, `"`, "", -1)
		mm := strings.Replace(searches[bs].InitSum, "Sought", "Seeking", -1)

		_, found := searches[bs]

		if found && searches[bs].IsActive {
			// if you insist on a full poll; but the thing works without input from pp and rc
			for {
				_, pp := os.Stat(fmt.Sprintf("%s/hgs_pp_%s", UNIXSOCKETPATH, searches[bs].ID))
				_, rc := os.Stat(fmt.Sprintf("%s/hgs_rc_%s", UNIXSOCKETPATH, searches[bs].ID))
				if pp == nil && rc == nil {
					// msg("found both search activity sockets", 5)
					break
				}
				if _, ok := searches[bs]; !ok {
					// don't wait forever
					break
				}
			}

			// we will grab the remainder value via unix socket
			rsock := false
			rconn, err := net.Dial("unix", fmt.Sprintf("%s/hgs_pp_%s", UNIXSOCKETPATH, searches[bs].ID))
			if err != nil {
				msg(fmt.Sprintf("RtWebsocket() has no connection to the remainder reports: %s/hgs_pp_%s", UNIXSOCKETPATH, searches[bs].ID), 1)
			} else {
				rsock = true
				defer rconn.Close()
			}

			// we will grab the hits value via unix socket
			hsock := false
			hconn, err := net.Dial("unix", fmt.Sprintf("%s/hgs_rc_%s", UNIXSOCKETPATH, searches[bs].ID))
			if err != nil {
				msg(fmt.Sprintf("RtWebsocket() has no connection to the hits reports: %s/hgs_rc_%s", UNIXSOCKETPATH, searches[bs].ID), 1)
			} else {
				// if there is no connection you will get a null pointer dereference
				hsock = true
				defer hconn.Close()
			}

			for {
				var r ReplyJS

				// [a] the easy info to report
				r.Active = "is_active"
				r.ID = bs
				r.TotalWrk = searches[bs].TableSize
				r.Elapsed = fmt.Sprintf("%.1fs", time.Now().Sub(searches[bs].Launched).Seconds())
				if searches[bs].PhaseNum == 2 {
					r.Extra = "(second pass)"
				} else {
					r.Extra = ""
				}

				// [b] the tricky info
				// [b1] set r.Remain via unix socket connection to SrchFeeder()'s broadcaster
				if rsock {
					r.Remain = func() int {
						connbuf := bufio.NewReader(rconn)
						for {
							rs, err := connbuf.ReadString('\n')
							if err != nil {
								break
							} else {
								// fmt.Println([]byte(rs)) --> [49 10]
								// and stripping the newline via strings is not working
								rr := []rune(rs)
								rr = rr[0 : len(rr)-1]
								st, _ := strconv.Atoi(string(rr))
								return st
							}
						}
						return -1
					}()
				} else {
					// see the JS: this turns off progress displays
					r.TotalWrk = -1
				}

				if r.Remain != 0 {
					r.Msg = mm
				} else if rsock {
					// will be zero if you never made the connection
					r.Msg = "Formatting the finds..."
				}

				// [b2] set r.Hits via unix socket connection to ResultCollation()'s broadcaster
				if hsock {
					r.Hits = func() int {
						connbuf := bufio.NewReader(hconn)
						for {
							ht, err := connbuf.ReadString('\n')
							if err != nil {
								break
							} else {
								// fmt.Println([]byte(rs)) --> [49 10]
								// and stripping the newline via strings is not working
								hh := []rune(ht)
								hh = hh[0 : len(hh)-1]
								h, _ := strconv.Atoi(string(hh))
								return h
							}
						}
						return 0
					}()
				}

				// Write
				js, y := json.Marshal(r)
				chke(y)

				er := ws.WriteMessage(websocket.TextMessage, js)

				if er != nil {
					c.Logger().Error(er)
					msg("RtWebsocket(): ws failed to write: breaking", 1)
					if hsock {
						hconn.Close()
					}
					if rsock {
						rconn.Close()
					}
					break
				}

				if _, exists := searches[bs]; !exists {
					if hsock {
						hconn.Close()
					}
					if rsock {
						rconn.Close()
					}
					done = true
					break
				}
			}
		}
	}

	return nil
}

// progresssocket - from where should the progress info be served?
func progresssocket(name string) net.Listener {
	// return a listener and the value of the port selected
	// socket vs tcp: tcp connection ends up in TIME_WAIT state and will block the port

	host, err := net.Listen("unix", fmt.Sprintf("%s/hgs_%s", UNIXSOCKETPATH, name))

	if err != nil {
		msg(fmt.Sprintf("progresssocket() could not open '%s/hgs_%s'", UNIXSOCKETPATH, name), 1)
	} else {
		// msg(fmt.Sprintf("progresssocket() opened '%s/hgs_%s'", UNIXSOCKETPATH, name), 1)
		return host
	}
	return nil
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
