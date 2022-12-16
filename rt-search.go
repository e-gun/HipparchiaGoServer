//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
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

	c.Response().After(func() { gcstats("RtSearch()") })

	user := readUUIDCookie(c)
	if !SafeAuthenticationCheck(user) {
		return c.JSONPretty(http.StatusOK, SearchOutputJSON{JS: VALIDATIONBOX}, JSONINDENT)
	}

	id := c.Param("id")
	srch := builddefaultsearch(c)
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
// TWO-PART SEARCHES
//

// WithinXLinesSearch - find A within N lines of B
func WithinXLinesSearch(originalsrch SearchStruct) SearchStruct {
	// after finding A, look for B within N lines of A

	// (part 1)
	//		HGoSrch(first)
	//
	// (part 2)
	// 		populate a new search list with a ton of passages via the first results
	//		HGoSrch(second)

	const (
		PSGT = `%s_FROM_%d_TO_%d`
		MSG1 = "%s WithinXLinesSearch(): %d initial hits"
		MSG2 = "%s SSBuildQueries() rerun"
		MSG3 = "%s WithinXLinesSearch(): %d subsequent hits"
	)

	previous := time.Now()
	first := generateinitialhits(originalsrch)

	d := fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf(MSG1, d, len(first.Results)), 4)
	previous = time.Now()

	second := clonesearch(first, 2)
	second.Seeking = second.Proximate
	second.LemmaOne = second.LemmaTwo
	second.Proximate = first.Seeking
	second.LemmaTwo = first.LemmaOne

	second.SetType()

	newpsg := make([]string, len(first.Results))
	for i := 0; i < len(first.Results); i++ {
		// avoid "gr0028_FROM_-1_TO_5"
		low := first.Results[i].TbIndex - first.ProxDist
		if low < 1 {
			low = 1
		}
		np := fmt.Sprintf(PSGT, first.Results[i].AuID(), low, first.Results[i].TbIndex+first.ProxDist)
		newpsg[i] = np
	}

	second.CurrentLimit = originalsrch.OriginalLimit
	second.SearchIn.Passages = newpsg
	second.NotNear = false

	SSBuildQueries(&second)

	d = fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf(MSG2, d), 4)
	previous = time.Now()

	second = HGoSrch(second)
	if second.HasPhrase {
		findphrasesacrosslines(&second)
	}

	if first.NotNear {
		hitmapper := make(map[string]DbWorkline)

		// all the original hits start as "good"
		for i := 0; i < len(first.Results); i++ {
			hitmapper[first.Results[i].BuildHyperlink()] = first.Results[i]
		}

		// delete any hit that is within N-lines of any second hit
		// hence "second.NotNear = false" above vs "first.NotNear" to get here: need matches, not misses
		for i := 0; i < len(second.Results); i++ {
			low := second.Results[i].TbIndex - first.ProxDist
			high := second.Results[i].TbIndex + first.ProxDist
			for j := low; j <= high; j++ {
				hlk := fmt.Sprintf(WKLNHYPERLNKTEMPL, second.Results[i].AuID(), second.Results[i].WkID(), j)
				if _, ok := hitmapper[hlk]; ok {
					delete(hitmapper, hlk)
				}
			}
		}
		second.Results = stringmapintoslice(hitmapper)
	}

	d = fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf(MSG3, d, len(first.Results)), 4)

	return second
}

// WithinXWordsSearch - find A within N words of B
func WithinXWordsSearch(originalsrch SearchStruct) SearchStruct {
	// profiling will show that all your time is spent on "if basicprxfinder.MatchString(str) && !first.NotNear"
	// as one would guess...

	const (
		PSGT = `%s_FROM_%d_TO_%d`
		LNK  = `index/%s/%s/%d`
		RGX  = `^(?P<head>.*?)%s(?P<tail>.*?)$`
		MSG1 = "%s WithinXWordsSearch(): %d initial hits"
		MSG2 = "%s WithinXWordsSearch(): %d subsequent hits"
		BAD1 = "WithinXWordsSearch() could not compile second pass regex term 'submatchsrchfinder': %s"
		BAD2 = "WithinXWordsSearch() could not compile second pass regex term 'basicprxfinder': %s"
	)
	previous := time.Now()
	first := generateinitialhits(originalsrch)

	d := fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf(MSG1, d, len(first.Results)), 4)
	previous = time.Now()

	// the trick is we are going to grab ALL lines near the initial hit; then build strings; then search those strings ourselves
	// so the second search is "anything nearby"

	// [a] build the second search
	second := clonesearch(first, 2)
	sskg := second.Proximate
	slem := second.LemmaTwo
	second.Seeking = ""
	second.LemmaOne = ""
	second.Proximate = first.Seeking
	second.LemmaTwo = first.LemmaOne
	// avoid "WHERE accented_line !~ ''" : force the type and make sure to check "first.NotNear" below
	second.NotNear = false

	second.SetType()

	// [a1] hard code a suspect assumption...
	need := 2 + (first.ProxDist / AVGWORDSPERLINE)

	resultmapper := make(map[string]int, len(first.Results))
	newpsg := make([]string, len(first.Results))

	// [a2] pick the lines to grab and associate them with the hits they go with
	// map[index/gr0007/018/15195:93 index/gr0007/018/15196:93 index/gr0007/018/15197:93 index/gr0007/018/15198:93 ...
	for i := 0; i < len(first.Results); i++ {
		low := first.Results[i].TbIndex - need
		if low < 1 {
			low = 1
		}
		np := fmt.Sprintf(PSGT, first.Results[i].AuID(), low, first.Results[i].TbIndex+need)
		newpsg[i] = np
		for j := first.Results[i].TbIndex - need; j <= first.Results[i].TbIndex+need; j++ {
			m := fmt.Sprintf(LNK, first.Results[i].AuID(), first.Results[i].WkID(), j)
			resultmapper[m] = i
		}
	}

	second.CurrentLimit = FIRSTSEARCHLIM
	second.SearchIn.Passages = newpsg
	SSBuildQueries(&second)

	// [b] run the second "search" for anything/everything: ""

	ss := HGoSrch(second)

	d = fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf(MSG2, d, len(first.Results)), 4)
	previous = time.Now()

	// [c] convert these finds into strings and then search those strings
	// [c1] build bundles of lines
	bundlemapper := make(map[int][]DbWorkline)
	for i := 0; i < len(ss.Results); i++ {
		url := ss.Results[i].BuildHyperlink()
		bun := resultmapper[url]
		bundlemapper[bun] = append(bundlemapper[bun], ss.Results[i])
	}

	for i, b := range bundlemapper {
		sort.Slice(b, func(i, j int) bool { return b[i].TbIndex < b[j].TbIndex })
		bundlemapper[i] = b
	}

	// [c2] decompose them into long strings
	stringmapper := make(map[int]string)
	for idx, lines := range bundlemapper {
		var bundle []string
		for i := 0; i < len(lines); i++ {
			bundle = append(bundle, columnpicker(first.SrchColumn, lines[i]))
		}
		stringmapper[idx] = strings.Join(bundle, " ")
	}

	// [c3] grab the head and tail of each
	// Sought »ἀδύνατον γὰρ« within 4 words of all 19 forms of »φύϲιϲ«...
	var re string
	if len(slem) != 0 {
		re = strings.Join(lemmaintoregexslice(slem), "|")
	} else {
		re = sskg
	}

	basicprxfinder, e := regexp.Compile(re)
	if e != nil {
		m := fmt.Sprintf(BAD2, re)
		msg(m, 1)
		return badsearch(m)
	}

	if len(first.LemmaOne) != 0 {
		re = "(" + strings.Join(AllLemm[first.LemmaOne].Deriv, " | ") + ")"

	} else {
		re = first.Seeking
	}

	submatchsrchfinder, e := regexp.Compile(fmt.Sprintf(RGX, re))
	if e != nil {
		m := fmt.Sprintf(BAD1, re)
		msg(m, 1)
		return badsearch(m)
	}

	// [c4] search head and tail for the second search term

	// the count is inclusive: the search term is one of the words
	// unless you do something "non solum" w/in 4 words of "sed etiam" is the non-obvious way to catch single-word sandwiches:
	// "non solum pecuniae sed etiam..."

	pd := first.ProxDist
	ph2 := len(strings.Split(strings.TrimSpace(first.Proximate), " "))

	if ph2 > 1 {
		pd = pd + ph2
	}

	// now we have a new problem: Sought all 19 forms of »φύϲιϲ« within 4 words of »ἀδύνατον γὰρ«
	// what if the str contains multiple valid values for term #1?
	// [291]	ϲτερεῶν ἅψηται ὁ πυρετόϲ ἐπειδὴ μὴ ὁμαλῶϲ θερμαίνεται ἀλλὰ ἀνωμάλωϲ εἰϲὶ γάρ τινα μόρια κατὰ φύϲιν ἔχοντα τινὰ δὲ παρὰ φύϲιν ϲυμβαίνει τὰ κατὰ φύϲιν ἔχοντα ἀντιλαμβάνεϲθαι τῶν παρὰ φύϲιν διακειμένων ἀδύνατον γὰρ ὁμαλὴν γενέϲθαι τὴν δυϲκραϲίαν οἱ δὲ ἑκτικῷ κατεϲχημένοι πυρετῷ τοῦτο δέ ἐϲτιν οἱ τὰ ϲτερεὰ πυρέττοντεϲ
	//

	// [c4a] quick prune of the finds by checking for the second term in the word bundle
	//possiblyvalid := make(map[int]string)
	//for idx, str := range stringmapper {
	//	if basicprxfinder.MatchString(str) && !first.NotNear {
	//		possiblyvalid[idx] = str
	//	} else if first.NotNear {
	//		possiblyvalid[idx] = str
	//	}
	//}
	//
	//// [c4b] now make sure the first term is near enough to the second term: zoom to termtwo and then look out from it
	//
	//var validresults []DbWorkline
	//for idx, str := range possiblyvalid {
	//	subs := submatchsrchfinder.FindStringSubmatch(str)
	//	head := ""
	//	tail := ""
	//	if len(subs) != 0 {
	//		head = subs[submatchsrchfinder.SubexpIndex("head")]
	//		tail = subs[submatchsrchfinder.SubexpIndex("tail")]
	//	}
	//
	//	hh := strings.Split(head, " ")
	//	start := 0
	//	if len(hh)-pd-1 > 0 {
	//		start = len(hh) - pd - 1
	//	}
	//	hh = hh[start:]
	//	head = " " + strings.Join(hh, " ")
	//
	//	tt := strings.Split(tail, " ")
	//	if len(tt) >= pd+1 {
	//		tt = tt[0 : pd+1]
	//	}
	//	tail = strings.Join(tt, " ") + " "
	//
	//	if first.NotNear {
	//		// toss hits
	//		if !basicprxfinder.MatchString(head) && !basicprxfinder.MatchString(tail) {
	//			validresults = append(validresults, first.Results[idx])
	//		}
	//	} else {
	//		// collect hits
	//		if basicprxfinder.MatchString(head) || basicprxfinder.MatchString(tail) {
	//			validresults = append(validresults, first.Results[idx])
	//		}
	//	}
	//}

	kvp := XWordsSecondSearch(&ss, stringmapper, basicprxfinder, submatchsrchfinder, pd)

	var res []DbWorkline
	for _, v := range kvp {
		res = append(res, first.Results[v.k])
	}

	second.Results = res
	second.Seeking = first.Seeking
	second.LemmaOne = first.LemmaOne

	return second
}

//
// FAN-OUT AND FAN-IN SECOND HALF OF WithinXWordsSearch()
//

func XWordsCheckFinds(p kvpair, basicprxfinder *regexp.Regexp, submatchsrchfinder *regexp.Regexp, pd int, notnear bool) kvpair {
	subs := submatchsrchfinder.FindStringSubmatch(p.v)
	head := ""
	tail := ""
	if len(subs) != 0 {
		head = subs[submatchsrchfinder.SubexpIndex("head")]
		tail = subs[submatchsrchfinder.SubexpIndex("tail")]
	}

	hh := strings.Split(head, " ")
	start := 0
	if len(hh)-pd-1 > 0 {
		start = len(hh) - pd - 1
	}
	hh = hh[start:]
	head = " " + strings.Join(hh, " ")

	tt := strings.Split(tail, " ")
	if len(tt) >= pd+1 {
		tt = tt[0 : pd+1]
	}
	tail = strings.Join(tt, " ") + " "

	var valid kvpair
	if notnear {
		// toss hits
		if !basicprxfinder.MatchString(head) && !basicprxfinder.MatchString(tail) {
			valid = p
		}
	} else {
		// collect hits
		if basicprxfinder.MatchString(head) || basicprxfinder.MatchString(tail) {
			valid = p
		}
	}
	return valid
}

type kvpair struct {
	k int
	v string
}

func XWordsSecondSearch(ss *SearchStruct, stringmapper map[int]string, bf *regexp.Regexp, sf *regexp.Regexp, dist int) []kvpair {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kvp := make([]kvpair, len(stringmapper))
	count := 0
	for k, v := range stringmapper {
		kvp[count] = kvpair{k, v}
		count += 1
	}

	emit, err := XWordsFeeder(ctx, &kvp, ss)
	chke(err)

	workers := Config.WorkerCount
	findchannels := make([]<-chan kvpair, workers)

	for i := 0; i < workers; i++ {
		fc, e := XWordsConsumer(ctx, emit, bf, sf, dist, ss.NotNear)
		chke(e)
		findchannels[i] = fc
	}

	results := XWordsCollation(ctx, ss, XWordsAggregator(ctx, findchannels...))
	if len(results) > ss.CurrentLimit {
		results = results[0:ss.CurrentLimit]
	}

	return results

}

func XWordsFeeder(ctx context.Context, kvp *[]kvpair, ss *SearchStruct) (<-chan kvpair, error) {
	emit := make(chan kvpair, Config.WorkerCount)
	remainder := -1

	go func() {
		defer close(emit)
		for i := 0; i < len(*kvp); i++ {
			select {
			case <-ctx.Done():
				break
			default:
				remainder = len(ss.Queries) - i - 1
				if remainder%POLLEVERYNTABLES == 0 {
					ss.Remain.Set(remainder)
				}
				emit <- (*kvp)[i]
			}
		}
	}()

	return emit, nil
}

func XWordsConsumer(ctx context.Context, kvp <-chan kvpair, bf *regexp.Regexp, sf *regexp.Regexp, dist int, notnear bool) (<-chan kvpair, error) {
	emitfinds := make(chan kvpair)
	go func() {
		defer close(emitfinds)
		for p := range kvp {
			select {
			case <-ctx.Done():
				return
			default:
				emitfinds <- XWordsCheckFinds(p, bf, sf, dist, notnear)
			}
		}
	}()
	return emitfinds, nil
}

func XWordsAggregator(ctx context.Context, findchannels ...<-chan kvpair) <-chan kvpair {
	var wg sync.WaitGroup
	emitaggregate := make(chan kvpair)
	broadcast := func(ll <-chan kvpair) {
		defer wg.Done()
		for l := range ll {
			select {
			case emitaggregate <- l:
			case <-ctx.Done():
				return
			}
		}
	}
	wg.Add(len(findchannels))
	for _, fc := range findchannels {
		go broadcast(fc)
	}

	go func() {
		wg.Wait()
		close(emitaggregate)
	}()

	return emitaggregate
}

func XWordsCollation(ctx context.Context, ss *SearchStruct, values <-chan kvpair) []kvpair {
	var allhits []kvpair
	done := false
	for {
		if done {
			break
		}
		select {
		case <-ctx.Done():
			done = true
		case val, ok := <-values:
			if ok {
				if len(val.v) != 0 {
					// *something* came back, but a negative result is {0, ""}
					allhits = append(allhits, val)
					ss.Hits.Set(len(allhits))
				}
				if len(allhits) > ss.OriginalLimit {
					done = true
				}
			} else {
				done = true
			}
		}
	}
	return allhits
}

// generateinitialhits - part one of a two-part search
func generateinitialhits(first SearchStruct) SearchStruct {
	first = HGoSrch(first)

	if first.HasPhrase {
		findphrasesacrosslines(&first)
	}
	// this was toggled just before the queries were written; it needs to be reset now
	first.CurrentLimit = first.OriginalLimit
	return first
}

//
// INITIAL SETUP
//

// builddefaultsearch - fill out the basic values for a new search
func builddefaultsearch(c echo.Context) SearchStruct {
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

	// msg("nonstandard builddefaultsearch() for testing", 1)

	return srch
}

// buildhollowsearch - is really a way to grab line collections via synthetic searchlists
func buildhollowsearch() SearchStruct {
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
		msg(fmt.Sprintf(FAILMSG, hdwd), 1)
		return []string{FAILSLC}
	}

	tp := `(^|\s)%s(\s|$)`

	// there is a problem: unless you do something, "(^|\s)ἁλιεύϲ(\s|$)" will be a search term but this will not find "ἁλιεὺϲ"
	var lemm []string
	for _, l := range AllLemm[hdwd].Deriv {
		lemm = append(lemm, findacuteorgrave(l))
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
		li := columnpicker(ss.SrchColumn, r)
		f := fp.MatchString(li)
		if f {
			valid[r.BuildHyperlink()] = r
		} else if ss.SkgRewritten && altfp.MatchString(li) {
			// i.e. "it's all on this line" (second try)
			// msg("althit", 1)
			valid[r.BuildHyperlink()] = r
		} else {
			// msg("'else'", 4)
			var nxt DbWorkline
			if i+1 < len(ss.Results) {
				nxt = ss.Results[i+1]
				if r.TbIndex+1 > AllWorks[r.WkUID].LastLine {
					nxt = DbWorkline{}
				} else if r.WkUID != nxt.WkUID || r.TbIndex+1 != nxt.TbIndex {
					// grab the actual next line (i.e. index = 101)
					nxt = graboneline(r.AuID(), r.TbIndex+1)
				}
			} else {
				// grab the actual next line (i.e. index = 101)
				nxt = graboneline(r.AuID(), r.TbIndex+1)
				if r.WkUID != nxt.WkUID {
					nxt = DbWorkline{}
				}
			}

			// combinator dodges double-register of hits
			nl := columnpicker(ss.SrchColumn, nxt)
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

// columnpicker - convert from db column name into struct name
func columnpicker(c string, r DbWorkline) string {
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
		msg(MSG, 2)
	}
	return li
}

// clonesearch - make a copy of a search with results and queries, inter alia, ripped out
func clonesearch(f SearchStruct, iteration int) SearchStruct {
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

// universalpatternmaker - feeder for searchtermfinder()
func universalpatternmaker(term string) string {
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

// searchtermfinder - find the universal regex equivalent of the search term
func searchtermfinder(term string) *regexp.Regexp {
	//	you need to convert:
	//		ποταμον
	//	into:
	//		([πΠ][οὀὁὂὃὄὅόὸΟὈὉὊὋὌὍ][τΤ][αἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάᾈᾉᾊᾋᾌᾍᾎᾏἈἉἊἋἌἍἎἏΑ][μΜ][οὀὁὂὃὄὅόὸΟὈὉὊὋὌὍ][νΝ])

	const (
		MSG = "searchtermfinder() could not compile the following: %s"
	)

	stre := universalpatternmaker(term)
	pattern, e := regexp.Compile(stre)
	if e != nil {
		msg(fmt.Sprintf(MSG, stre), 1)
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
		s = buildhollowsearch()
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
