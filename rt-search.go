//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"regexp"
	"strings"
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

	const (
		TOOMANYIP    = "<code>Cannot execute this search. Your ip address (%s) is already running the maximum number of simultaneous searches allowed: %d.</code>"
		TOOMANYTOTAL = "<code>Cannot execute this search. The server is already running the maximum number of simultaneous searches allowed: %d.</code>"
	)

	user := readUUIDCookie(c)

	// [A] ARE WE GOING TO DO THIS AT ALL?

	if !AllAuthorized.Check(user) {
		return JSONresponse(c, SearchOutputJSON{JS: VALIDATIONBOX})
	}

	getsrchcount := func(ip string) int {
		responder := WSSICount{key: ip, response: make(chan int)}
		WSInfo.IPSrchCount <- responder
		return <-responder.response
	}

	if getsrchcount(c.RealIP()) >= Config.MaxSrchIP {
		m := fmt.Sprintf(TOOMANYIP, c.RealIP(), getsrchcount(c.RealIP()))
		return JSONresponse(c, SearchOutputJSON{Searchsummary: m})
	}

	if len(WebsocketPool.ClientMap) >= Config.MaxSrchTot {
		m := fmt.Sprintf(TOOMANYTOTAL, len(WebsocketPool.ClientMap))
		return JSONresponse(c, SearchOutputJSON{Searchsummary: m})
	}

	// [B] OK, WE ARE DOING IT

	srch := BuildDefaultSearch(c)
	se := AllSessions.GetSess(user)

	// [C] BUT WHAT KIND OF SEARCH IS IT? MAYBE IT IS A VECTOR SEARCH...

	// note the racer says that there are *many* race candidates in the imported vector code...
	// "wego@v0.0.11/pkg/model/word2vec/optimizer.go:126"
	// "wego@v0.0.11/pkg/model/word2vec/model.go:75"
	// ...

	if se.VecNNSearch && !Config.VectorsDisabled {
		// not a normal search: jump to "vectorqueryneighbors.go" where we grab all lines; build a model; query against the model; return html
		return NeighborsSearch(c, srch)
	}

	if se.VecLDASearch && !Config.VectorsDisabled {
		// not a normal search: jump to "vectorquerylda.go"
		return LDASearch(c, srch)
	}

	// [D] OK, IT IS A SEARCH FOR A WORD OR PHRASE

	c.Response().After(func() { messenger.LogPaths("RtSearch()") })

	// HasPhraseBoxA makes us use a fake limit temporarily
	reallimit := srch.CurrentLimit

	var completed SearchStruct
	if srch.Twobox {
		if srch.ProxScope == "words" {
			completed = WithinXWordsSearch(srch)
		} else {
			completed = WithinXLinesSearch(srch)
		}
	} else {
		completed = srch
		SearchAndInsertResults(&completed)
		if completed.HasPhraseBoxA {
			findphrasesacrosslines(&completed)
		}
	}

	if completed.Results.Len() > reallimit {
		completed.Results.ResizeTo(reallimit)
	}

	// [E] DONE: TIME TO FORMAT

	completed.SortResults()
	soj := SearchOutputJSON{}
	if se.HitContext == 0 {
		soj = FormatNoContextResults(&completed)
	} else {
		soj = FormatWithContextResults(&completed)
	}

	WSInfo.Del <- srch.WSID
	return JSONresponse(c, soj)
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
	s.Results.Lines = append(s.Results.Lines, l)
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
		msg(fmt.Sprintf(FAILMSG, hdwd), MSGFYI)
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
		ss.Results = WorkLineBundle{}
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

	var valid = make(map[string]DbWorkline, ss.Results.Len())

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

	rr := ss.Results.YieldAll()
	i := 0
	for r := range rr {
		// do the "it's all on this line" case separately
		li := ColumnPicker(ss.SrchColumn, r)
		f := fp.MatchString(li)
		if f {
			valid[r.BuildHyperlink()] = r
		} else if ss.SkgRewritten && altfp.MatchString(li) {
			// i.e. "it's all on this line" (second try)
			valid[r.BuildHyperlink()] = r
		} else if ss.OneHit {
			// problem with onehit + phrase: unless you do something the next yields ZERO results if OneHit is turned on
			// Sought »αρετα παϲα«
			// Searched 7,461 works and found 2 passages (1.80s)

			// the first pass will find the wrong lines. OneHit means FinalResultCollation() will register only the first
			// of a yet untested pair; but the real hit is in [b], not [a]; this is a 'linebundle' windowing problem from PSQL

			// [1.a] διανοίαϲ καὶ ὀρέξιοϲ ὧν ἁ μὲν διάνοια τῶ λόγον ἔχοντόϲ
			// [1.b] ἐντι ἁ δ ὄρεξιϲ τῶ ἀλόγω διὸ καὶ ἀρετὰ πᾶϲα ἐν

			// [2.a] καὶ τῷ ἀλόγῳ ϲύγκειται γὰρ ἐκ διανοίαϲ καὶ ὀρέξιοϲ ὧν ἁ μὲν διάνοια
			// [2.b] τῶ λόγον ἔχοντόϲ ἐντι ἁ δ ὄρεξιϲ τῶ ἀλόγω διὸ καὶ ἀρετὰ πᾶϲα ἐν
			nxt := GrabOneLine(r.AuID(), r.TbIndex+1)
			if nxt.WkUID == r.WkUID {
				n := ColumnPicker(ss.SrchColumn, nxt)
				if fp.MatchString(n) {
					valid[nxt.BuildHyperlink()] = nxt
				}
			}
		} else {
			var nxt DbWorkline
			if i+1 < ss.Results.Len() {
				nxt = ss.Results.Lines[i+1]
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
					// yes! actually record a valid hit...
					valid[r.BuildHyperlink()] = r
				}
			}
		}
		i++
	}

	slc := make([]DbWorkline, len(valid))
	counter := 0
	for _, r := range valid {
		slc[counter] = r
		counter += 1
	}

	ss.Results.Lines = slc
}

// pruneresultsbylemma - take a collection of results and make sure some form of X is in them
func pruneresultsbylemma(hdwd string, ss *SearchStruct) {
	rgx := lemmaintoregexslice(hdwd)
	pat, e := regexp.Compile(strings.Join(rgx, "|"))
	if e != nil {
		pat = regexp.MustCompile("FAILED_FIND_NOTHING")
		msg(fmt.Sprintf("pruneresultsbylemma() could not compile the following: %s", strings.Join(rgx, "|")), MSGWARN)
	}

	var valid = make(map[string]DbWorkline, ss.Results.Len())

	rr := ss.Results.YieldAll()
	for r := range rr {
		// do the "it's all on this line" case separately
		li := ColumnPicker(ss.SrchColumn, r)
		if pat.MatchString(li) {
			valid[r.BuildHyperlink()] = r
		}
	}

	slc := make([]DbWorkline, len(valid))
	counter := 0
	for _, r := range valid {
		slc[counter] = r
		counter += 1
	}

	ss.Results.Lines = slc

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
