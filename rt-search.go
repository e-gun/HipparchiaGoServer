//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	hasAccent = regexp.MustCompile("[äëïöüâêîôûàèìòùáéíóúᾂᾒᾢᾃᾓᾣᾄᾔᾤᾅᾕᾥᾆᾖᾦᾇᾗᾧἂἒἲὂὒἢὢἃἓἳὃὓἣὣἄἔἴὄὔἤὤἅἕἵὅὕἥὥἆἶὖἦὦἇἷὗἧὧᾲῂῲᾴῄῴᾷῇῷᾀᾐᾠᾁᾑᾡῒῢΐΰῧἀἐἰὀὐἠὠῤἁἑἱὁὑἡὡῥὰὲὶὸὺὴὼάέίόύήώᾶῖῦῆῶϊϋ]")
)

//
// ROUTING
//

// RtSearchConfirm - just tells the client JS where to find the poll
func RtSearchConfirm(c echo.Context) error {
	pt := fmt.Sprintf("%d", Config.HostPort)
	return c.String(http.StatusOK, pt)
}

// RtSearch - find X (derived from boxes on page) in Y (derived from the session)
func RtSearch(c echo.Context) error {
	// "OneBox"
	// [1] single word
	// [2] phrase
	// [3] lemma
	// "TwoBox"
	// [4] single + single
	// [5] lemma + single
	// [6] lemma + lemma
	// [7] phrase + single
	// [8] phrase + lemma
	// [9] phrase + phrase

	c.Response().After(func() { GCStats("RtSearch()") })

	user := readUUIDCookie(c)
	if !SafeAuthenticationCheck(user) {
		return c.JSONPretty(http.StatusOK, SearchOutputJSON{JS: VALIDATIONBOX}, JSONINDENT)
	}

	id := c.Param("id")
	srch := BuildDefaultSearch(c)
	srch.User = user

	srch.Seeking = c.QueryParam("skg")
	srch.Proximate = c.QueryParam("prx")
	srch.LemmaOne = c.QueryParam("lem")
	srch.LemmaTwo = c.QueryParam("plm")
	srch.ID = c.Param("id")
	srch.IsVector = false
	// HasPhrase makes us use a fake limit temporarily
	reallimit := srch.CurrentLimit

	srch.CleanInput()
	srch.SetType() // must happen before SSBuildQueries()
	srch.FormatInitialSummary()

	// now safe to rewrite skg oj that "^|\s", etc. can be added
	srch.Seeking = whitespacer(srch.Seeking, &srch)
	srch.Proximate = whitespacer(srch.Proximate, &srch)

	se := SafeSessionRead(user)
	sl := SessionIntoSearchlist(se)

	srch.SearchIn = sl.Inc
	srch.SearchEx = sl.Excl
	srch.SearchSize = sl.Size

	if srch.Twobox {
		srch.CurrentLimit = FIRSTSEARCHLIM
	}
	SSBuildQueries(&srch)

	srch.TableSize = len(srch.Queries)
	srch.IsActive = true
	SafeSearchMapInsert(srch)

	var completed SearchStruct
	if SearchMap[id].Twobox {
		if SearchMap[id].ProxScope == "words" {
			completed = WithinXWordsSearch(SearchMap[id])
		} else {
			completed = WithinXLinesSearch(SearchMap[id])
		}
	} else {
		completed = HGoSrch(SearchMap[id])
		if completed.HasPhrase {
			findphrasesacrosslines(&completed)
		}
	}

	if len(completed.Results) > reallimit {
		completed.Results = completed.Results[0:reallimit]
	}

	completed.SortResults()

	soj := SearchOutputJSON{}
	if se.HitContext == 0 {
		soj = FormatNoContextResults(&completed)
	} else {
		soj = FormatWithContextResults(&completed)
	}

	SafeSearchMapDelete(id)

	return c.JSONPretty(http.StatusOK, soj, JSONINDENT)
}

//
// INITIAL SETUP
//

// BuildDefaultSearch - fill out the basic values for a new search
func BuildDefaultSearch(c echo.Context) SearchStruct {
	user := readUUIDCookie(c)
	sess := SafeSessionRead(user)

	var srch SearchStruct
	srch.User = user
	srch.Launched = time.Now()
	srch.CurrentLimit = sess.HitLimit
	srch.OriginalLimit = sess.HitLimit
	srch.SrchColumn = DEFAULTCOLUMN
	srch.SrchSyntax = DEFAULTSYNTAX
	srch.OrderBy = ORDERBY
	srch.SearchIn = sess.Inclusions
	srch.SearchEx = sess.Exclusions
	srch.ProxDist = sess.Proximity
	srch.ProxScope = sess.SearchScope
	srch.NotNear = false
	srch.Twobox = false
	srch.HasPhrase = false
	srch.HasLemma = false
	srch.SkgRewritten = false
	srch.OneHit = sess.OneHit
	srch.PhaseNum = 1
	srch.TTName = strings.Replace(uuid.New().String(), "-", "", -1)
	srch.AcqHitCounter()
	srch.AcqRemainCounter()

	if sess.NearOrNot == "notnear" {
		srch.NotNear = true
	}

	// msg("nonstandard BuildDefaultSearch() for testing", MSGCRIT)

	return srch
}

// BuildHollowSearch - is really a way to grab line collections via synthetic searchlists
func BuildHollowSearch() SearchStruct {
	s := SearchStruct{
		User:          "",
		ID:            strings.Replace(uuid.New().String(), "-", "", -1),
		Seeking:       "",
		Proximate:     "",
		LemmaOne:      "",
		LemmaTwo:      "",
		InitSum:       "",
		Summary:       "",
		ProxScope:     "",
		ProxType:      "",
		ProxDist:      0,
		HasLemma:      false,
		HasPhrase:     false,
		IsVector:      false,
		IsActive:      false,
		OneHit:        false,
		Twobox:        false,
		NotNear:       false,
		SkgRewritten:  false,
		PhaseNum:      0,
		SrchColumn:    DEFAULTCOLUMN,
		SrchSyntax:    DEFAULTSYNTAX,
		OrderBy:       ORDERBY,
		CurrentLimit:  FIRSTSEARCHLIM,
		OriginalLimit: FIRSTSEARCHLIM,
		SkgSlice:      nil,
		PrxSlice:      nil,
		SearchIn:      SearchIncExl{},
		SearchEx:      SearchIncExl{},
		Queries:       nil,
		Results:       nil,
		Launched:      time.Now(),
		TTName:        strings.Replace(uuid.New().String(), "-", "", -1),
		SearchSize:    0,
		TableSize:     0,
	}
	s.AcqHitCounter()
	s.AcqRemainCounter()
	return s
}

// whitespacer - massage search string to let regex accept start/end of a line as whitespace
func whitespacer(skg string, ss *SearchStruct) string {
	// whitespace issue: " ἐν Ὀρέϲτῃ " cannot be found at the start of a line where it is "ἐν Ὀρέϲτῃ "
	// do not run this before formatinitialsummary()
	// also used by searchformatting.go

	if strings.Contains(skg, " ") {
		ss.SkgRewritten = true
		rs := []rune(skg)
		a := ""
		if rs[0] == ' ' {
			a = "(^|\\s)"
		}
		z := ""
		if rs[len(rs)-1] == ' ' {
			z = "(\\s|$)"
		}
		skg = strings.TrimSpace(skg)
		skg = a + skg + z
	}
	return skg
}

// restorewhitespace - undo whitespacer() modifications
func restorewhitespace(skg string) string {
	// will have a problem rewriting regex inside phrasecombinations() if you don't clear whitespacer() products out
	// even though we are about to put exactly this back in again...
	skg = strings.Replace(skg, "(^|\\s)", " ", 1)
	skg = strings.Replace(skg, "(\\s|$)", " ", -1)
	return skg
}

//
// HELPERS
//

// badsearch - something went wrong, return a blank SearchStruct
func badsearch(msg string) SearchStruct {
	var s SearchStruct
	var l DbWorkline
	l.MarkedUp = msg
	s.Results = append(s.Results, l)
	return s
}

func lemmaintoregexslice(hdwd string) []string {
	// rather than do one word per query, bundle things up: some words have >100 forms
	// ...(^|\\s)ἐδηλώϲαντο(\\s|$)|(^|\\s)δεδηλωμένοϲ(\\s|$)|(^|\\s)δήλουϲ(\\s|$)|(^|\\s)δηλούϲαϲ(\\s|$)...
	const (
		FAILMSG = "lemmaintoregexslice() could not find '%s'"
		FAILSLC = "FIND_NOTHING"
	)

	var qq []string
	if _, ok := AllLemm[hdwd]; !ok {
		msg(fmt.Sprintf(FAILMSG, hdwd), MSGWARN)
		return []string{FAILSLC}
	}

	tp := `(^|\s)%s(\s|$)`

	// there is a problem: unless you do something, "(^|\s)ἁλιεύϲ(\s|$)" will be a search term but this will not find "ἁλιεὺϲ"
	var lemm []string
	for _, l := range AllLemm[hdwd].Deriv {
		lemm = append(lemm, FindAcuteOrGrave(l))
	}

	ct := 0
	for true {
		var bnd []string
		for i := 0; i < MAXLEMMACHUNKSIZE; i++ {
			if ct > len(lemm)-1 {
				break
			}
			re := fmt.Sprintf(tp, lemm[ct])
			bnd = append(bnd, re)
			ct += 1
		}
		qq = append(qq, strings.Join(bnd, "|"))
		if ct >= len(lemm)-1 {
			break
		}
	}
	return qq
}

// findphrasesacrosslines - "one two$" + "^three four" makes a hit if you want "one two three four"
func findphrasesacrosslines(ss *SearchStruct) {
	// modify ss in place

	const (
		FAIL = "<code>SEARCH FAILED: term sent to findphrasesacrosslines() yielded error inside regexp.Compile()</code><br><br>"
	)

	recordfailure := func() {
		ss.Results = []DbWorkline{}
		ss.ExtraMsg = FAIL
	}

	getcombinations := func(phr string) [][2]string {
		// 'one two three four five' -->
		// [('one', 'two three four five'), ('one two', 'three four five'), ('one two three', 'four five'), ('one two three four', 'five')]

		gt := func(n int, wds []string) []string {
			return wds[n:]
		}

		gh := func(n int, wds []string) []string {
			return wds[:n]
		}

		ww := strings.Split(phr, " ")
		var comb [][2]string
		for i := range ww {
			h := strings.Join(gh(i, ww), " ")
			t := strings.Join(gt(i, ww), " ")
			h = h + "$"
			t = "^" + t
			comb = append(comb, [2]string{h, t})
		}

		var trimmed [][2]string
		for _, c := range comb {
			head := strings.TrimSpace(c[0]) != "" && strings.TrimSpace(c[0]) != "$"
			tail := strings.TrimSpace(c[1]) != "" && strings.TrimSpace(c[0]) != "^"
			if head && tail {
				trimmed = append(trimmed, c)
			}
		}

		return trimmed
	}

	var valid = make(map[string]DbWorkline, len(ss.Results))

	skg := ss.Seeking
	if ss.SkgRewritten {
		skg = restorewhitespace(ss.Seeking)
	}

	find := regexp.MustCompile(`^ `)
	re := find.ReplaceAllString(skg, "(^|\\s)")
	find = regexp.MustCompile(` $`)
	re = find.ReplaceAllString(re, "(\\s|$)")

	fp, e := regexp.Compile(re)
	if e != nil {
		// Καῖϲα[ρ can be requested, but it will cause big problems
		// this mechanism likely needs to be inserted in more locations...
		recordfailure()
		return
	}

	altfp, e := regexp.Compile(ss.Seeking)
	if e != nil {
		recordfailure()
		return
	}

	for i := 0; i < len(ss.Results); i++ {
		r := ss.Results[i]
		// do the "it's all on this line" case separately
		li := ColumnPicker(ss.SrchColumn, r)
		f := fp.MatchString(li)
		if f {
			valid[r.BuildHyperlink()] = r
		} else if ss.SkgRewritten && altfp.MatchString(li) {
			// i.e. "it's all on this line" (second try)
			valid[r.BuildHyperlink()] = r
		} else {
			var nxt DbWorkline
			if i+1 < len(ss.Results) {
				nxt = ss.Results[i+1]
				if r.TbIndex+1 > AllWorks[r.WkUID].LastLine {
					nxt = DbWorkline{}
				} else if r.WkUID != nxt.WkUID || r.TbIndex+1 != nxt.TbIndex {
					// grab the actual next line (i.e. index = 101)
					nxt = GrabOneLine(r.AuID(), r.TbIndex+1)
				}
			} else {
				// grab the actual next line (i.e. index = 101)
				nxt = GrabOneLine(r.AuID(), r.TbIndex+1)
				if r.WkUID != nxt.WkUID {
					nxt = DbWorkline{}
				}
			}

			// combinator dodges double-register of hits
			nl := ColumnPicker(ss.SrchColumn, nxt)
			comb := getcombinations(re)
			for _, c := range comb {
				fp2, e1 := regexp.Compile(c[0])
				if e1 != nil {
					recordfailure()
					return
				}
				sp, e2 := regexp.Compile(c[1])
				if e2 != nil {
					recordfailure()
					return
				}
				f = fp2.MatchString(li)
				s := sp.MatchString(nl)
				if f && s && r.WkUID == nxt.WkUID {
					valid[r.BuildHyperlink()] = r
				}
			}
		}
	}

	slc := make([]DbWorkline, len(valid))
	counter := 0
	for _, r := range valid {
		slc[counter] = r
		counter += 1
	}

	ss.Results = slc
}

// ColumnPicker - convert from db column name into struct name
func ColumnPicker(c string, r DbWorkline) string {
	const (
		MSG = "second.SrchColumn was not set; defaulting to 'stripped_line'"
	)

	var li string
	switch c {
	case "stripped_line":
		li = r.Stripped
	case "accented_line":
		li = r.Accented
	case "marked_up_line": // only a maniac tries to search via marked_up_line
		li = r.MarkedUp
	default:
		li = r.Stripped
		msg(MSG, MSGNOTE)
	}
	return li
}

// CloneSearch - make a copy of a search with results and queries, inter alia, ripped out
func CloneSearch(f SearchStruct, iteration int) SearchStruct {
	// note that the clone is not accessible to RtWebsocket() because it never gets registered in the global SearchMap
	// this means no progress for second pass SearchMap; this can be achieved, but it is not currently a priority
	s := f
	s.Results = []DbWorkline{}
	s.Queries = []PrerolledQuery{}
	s.SearchIn = SearchIncExl{}
	s.SearchEx = SearchIncExl{}
	s.TTName = strings.Replace(uuid.New().String(), "-", "", -1)
	s.SkgSlice = []string{}
	s.PrxSlice = []string{}
	s.PhaseNum = iteration

	oid := strings.Replace(f.ID, "_pt2", "", -1) // so a pt3 does not look like "_pt2_pt3"
	id := fmt.Sprintf("%s_pt%d", oid, iteration)
	s.ID = id
	return s
}

// UniversalPatternMaker - feeder for SearchTermFinder()
func UniversalPatternMaker(term string) string {
	// also used by searchformatting.go
	// converter := extendedrunefeeder()
	converter := ERuneFd // see top of generichelpers.go
	st := []rune(term)
	var stre string
	for _, r := range st {
		if _, ok := converter[r]; ok {
			re := fmt.Sprintf("[%s]", string(converter[r]))
			stre += re
		} else {
			stre += string(r)
		}
	}
	stre = fmt.Sprintf("(%s)", stre)
	return stre
}

// SearchTermFinder - find the universal regex equivalent of the search term
func SearchTermFinder(term string) *regexp.Regexp {
	//	you need to convert:
	//		ποταμον
	//	into:
	//		([πΠ][οὀὁὂὃὄὅόὸΟὈὉὊὋὌὍ][τΤ][αἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάᾈᾉᾊᾋᾌᾍᾎᾏἈἉἊἋἌἍἎἏΑ][μΜ][οὀὁὂὃὄὅόὸΟὈὉὊὋὌὍ][νΝ])

	const (
		MSG = "SearchTermFinder() could not compile the following: %s"
	)

	stre := UniversalPatternMaker(term)
	pattern, e := regexp.Compile(stre)
	if e != nil {
		msg(fmt.Sprintf(MSG, stre), MSGWARN)
		pattern = regexp.MustCompile("FAILED_FIND_NOTHING")
	}
	return pattern
}

// SafeSearchMapInsert - use a lock to safely swap a SearchStruct into the SearchMap
func SafeSearchMapInsert(ns SearchStruct) {
	SearchLocker.Lock()
	defer SearchLocker.Unlock()
	SearchMap[ns.ID] = ns
}

// SafeSearchMapRead - use a lock to safely read a SearchStruct from the SearchMap
func SafeSearchMapRead(id string) SearchStruct {
	SearchLocker.RLock()
	defer SearchLocker.RUnlock()
	s, e := SearchMap[id]
	if e != true {
		s = BuildHollowSearch()
		s.ID = id
		s.IsActive = false
	}
	return s
}

func SafeSearchMapDelete(id string) {
	SearchLocker.RLock()
	defer SearchLocker.RUnlock()
	delete(SearchMap, id)
}
