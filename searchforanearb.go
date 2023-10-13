//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

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
	first := HGoSrch(originalsrch)

	if first.HasPhraseBoxA {
		findphrasesacrosslines(&first)
	}
	// this was toggled just before the queries were written; it needs to be reset now
	first.CurrentLimit = first.OriginalLimit

	d := fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf(MSG1, d, len(first.Results)), MSGPEEK)
	previous = time.Now()

	second := CloneSearch(first, 2)
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
	msg(fmt.Sprintf(MSG2, d), MSGPEEK)
	previous = time.Now()

	second = HGoSrch(second)
	if second.HasPhraseBoxA && !second.IsLemmAndPhr {
		findphrasesacrosslines(&second)
	} else if second.IsLemmAndPhr {
		pruneresultsbylemma(second.LemmaOne, &second)
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
		second.Results = StringMapIntoSlice(hitmapper)
	}

	d = fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf(MSG3, d, len(second.Results)), MSGPEEK)

	SIDel <- second.ID

	return second
}

// WithinXWordsSearch - find A within N words of B
func WithinXWordsSearch(originalsrch SearchStruct) SearchStruct {
	// after finding A, look for B within N words of A

	// (part 1)
	//		HGoSrch(first)
	//
	// (part 2)
	// 		grab the neighborhoods of these hits
	//		build long strings from the neighborhoods
	//		(fan out to XWordsCheckFinds())
	//		center self on A and then trim strings to "within N words"
	//      look for B in that zone

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
	first := HGoSrch(originalsrch)

	if first.HasPhraseBoxA {
		findphrasesacrosslines(&first)
	}

	// showinterimresults(&first)

	// this was toggled just before the queries were written; it needs to be reset now
	first.CurrentLimit = first.OriginalLimit

	d := fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf(MSG1, d, len(first.Results)), MSGPEEK)
	previous = time.Now()

	// the trick is we are going to grab ALL lines near the initial hit; then build strings; then search those strings ourselves
	// so the second search is "anything nearby"

	// [a] build the second search
	second := CloneSearch(first, 2)
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

	// showinterimresults(&ss)

	d = fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf(MSG2, d, len(first.Results)), MSGPEEK)
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

	// [c2] decompose them into long strings and assign to a KVPair (K will let you get back to first.Results[i])

	kvp := make([]KVPair, len(bundlemapper))
	count := 0
	for idx, lines := range bundlemapper {
		var bundle []string
		for i := 0; i < len(lines); i++ {
			bundle = append(bundle, ColumnPicker(first.SrchColumn, lines[i]))
		}
		kvp[count] = KVPair{K: idx, V: strings.Join(bundle, " ")}
		count += 1
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
		msg(m, MSGWARN)
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
		msg(m, MSGWARN)
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

	// parallelize when doing the second pass: only really matters if doing "all forms of X" near "all forms of Y"
	// if X is a very common word like "πόλιϲ" (125,274 forms)

	// [PARALLEL]
	// Sought all 50 forms of »πόλιϲ« within 5 words of all 16 forms of »τοξότηϲ«
	// Searched 7,461 works and found 14 passages (8.89s)

	// [MONO]
	// Sought all 50 forms of »πόλιϲ« within 5 words of all 16 forms of »τοξότηϲ«
	// Searched 7,461 works and found 14 passages (20.19s)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	emit, err := XWordsFeeder(ctx, &kvp, &ss)
	chke(err)

	workers := Config.WorkerCount
	findchannels := make([]<-chan int, workers)

	for i := 0; i < workers; i++ {
		fc, ee := XWordsConsumer(ctx, emit, basicprxfinder, submatchsrchfinder, pd, originalsrch.NotNear)
		chke(ee)
		findchannels[i] = fc
	}

	results := XWordsCollation(ctx, &ss, XWordsAggregator(ctx, findchannels...))
	if len(results) > ss.CurrentLimit {
		results = results[0:ss.CurrentLimit]
	}

	res := make([]DbWorkline, len(results))
	for i := 0; i < len(results); i++ {
		res[i] = first.Results[results[i]]
	}

	second.Results = res
	second.Seeking = first.Seeking
	second.LemmaOne = first.LemmaOne
	second.CurrentLimit = first.OriginalLimit

	SIDel <- second.ID

	return second
}

//
// FAN-OUT AND FAN-IN SECOND HALF OF WithinXWordsSearch()
//

type KVPair struct {
	K int
	V string
}

// XWordsFeeder - emit items to a channel from the []KVPair that will be consumed by the XWordsConsumer
func XWordsFeeder(ctx context.Context, kvp *[]KVPair, ss *SearchStruct) (<-chan KVPair, error) {
	emit := make(chan KVPair, Config.WorkerCount)
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
					// ss.Remain.Set(remainder)
					SIUpdateRemain <- SIKVi{ss.ID, remainder}
				}
				emit <- (*kvp)[i]
			}
		}
	}()

	return emit, nil
}

// XWordsConsumer - grab a KVPair; check to see if it is a hit; emit the valid hits to a channel
func XWordsConsumer(ctx context.Context, kvp <-chan KVPair, bf *regexp.Regexp, sf *regexp.Regexp, dist int, notnear bool) (<-chan int, error) {
	emitfinds := make(chan int)
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

// XWordsAggregator - gather all hits from the findchannels into one place and then feed them to XWordsCollation
func XWordsAggregator(ctx context.Context, findchannels ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	emitaggregate := make(chan int)
	broadcast := func(hits <-chan int) {
		defer wg.Done()
		for h := range hits {
			select {
			case emitaggregate <- h:
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

// XWordsCollation - return the actual []KVPair results after pulling them from the XWordsAggregator channel
func XWordsCollation(ctx context.Context, ss *SearchStruct, hits <-chan int) []int {
	var allhits []int
	done := false
	for {
		if done {
			break
		}
		select {
		case <-ctx.Done():
			done = true
		case h, ok := <-hits:
			if ok {
				if h != -1 {
					// *something* came back from XWordsCheckFinds; a negative result is {-1, ""}
					allhits = append(allhits, h)
					// ss.Hits.Set(len(allhits))
					SIUpdateHits <- SIKVi{ss.ID, len(allhits)}
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

// XWordsCheckFinds - parallel hit checker logic for WithinXWordsSearch
func XWordsCheckFinds(p KVPair, basicprxfinder *regexp.Regexp, submatchsrchfinder *regexp.Regexp, proximity int, notnear bool) int {
	const (
		CUTPRE = `^(?P<head>.*?)`
		CUTSUF = `(?P<tail>.*?)$`
	)

	// the default return is "not a hit"
	result := -1

	// now we have a new problem: Sought all 19 forms of »φύϲιϲ« within 4 words of »ἀδύνατον γὰρ«
	// what if the string contains multiple valid values for term #1?
	// [291]	ϲτερεῶν ἅψηται ὁ πυρετόϲ ἐπειδὴ μὴ ὁμαλῶϲ θερμαίνεται ἀλλὰ ἀνωμάλωϲ εἰϲὶ γάρ τινα μόρια κατὰ φύϲιν ἔχοντα τινὰ δὲ παρὰ φύϲιν ϲυμβαίνει τὰ κατὰ φύϲιν ἔχοντα ἀντιλαμβάνεϲθαι τῶν παρὰ φύϲιν διακειμένων ἀδύνατον γὰρ ὁμαλὴν γενέϲθαι τὴν δυϲκραϲίαν οἱ δὲ ἑκτικῷ κατεϲχημένοι πυρετῷ τοῦτο δέ ἐϲτιν οἱ τὰ ϲτερεὰ πυρέττοντεϲ
	//

	// quick preliminary test (which does seem to shave 5-10% from your time...)
	possible := false
	if basicprxfinder.MatchString(p.V) && !notnear {
		possible = true
	} else if notnear {
		possible = true
	}
	if !possible {
		return result
	}

	subs := submatchsrchfinder.FindStringSubmatch(p.V)
	head := ""
	tail := ""
	if len(subs) != 0 {
		head = subs[submatchsrchfinder.SubexpIndex("head")]
		tail = subs[submatchsrchfinder.SubexpIndex("tail")]
	}

	hh := strings.Split(head, " ")
	start := 0
	if len(hh)-proximity-1 > 0 {
		// "len(hh) - proximity" is wrong; "within 5" will only find Ἔχειϲ within 6 words of λανθάνει in S. Ant.; it comes 5 words before
		start = len(hh) - proximity - 1
	}
	hh = hh[start:]
	head = " " + strings.Join(hh, " ")

	tt := strings.Split(tail, " ")

	// but we can't build the tail without making another check...

	// Sought »ἐϲχάτη χθονόϲ« within 9 words of all 41 forms of »γαῖα«
	// in the following we pick up the first »ἐϲχάτη χθονόϲ« of two copies of and set it as the border, but miss a hit if you do not look after the second...
	// [9]     ὁ ποιητὴϲ ἐνταῦθά φηϲιν οὐ τὰ πρὸϲ ὠκεανὸν ἀλλὰ τὰ ἐκεῖ πρὸϲ τῇ κατὰ νεῖλον θαλάϲϲῃ καθὰ καὶ αἰϲχύλοϲ εἰπών ἔϲτιν πόλιϲ κάνωβοϲ ἐϲχάτη χθονόϲ πᾶϲα γὰρ ἀγχίαλοϲ ἐϲχάτη χθονόϲ διὸ καὶ μενελαϊ/τηϲ νομὸϲ ἐκεῖ ὡϲ τοιαύτηϲ γῆϲ ὑπὸ μενελάῳ ποτὲ γενομένηϲ  steph byz ἀπόλλωνοϲ πόλιϲ ἐν αἰγύπτῳ πρὸϲ
	//        h	false    νεῖλον θαλάϲϲῃ καθὰ καὶ αἰϲχύλοϲ εἰπών ἔϲτιν πόλιϲ κάνωβοϲ
	//        t	false    πᾶϲα γὰρ ἀγχίαλοϲ ἐϲχάτη χθονόϲ διὸ καὶ μενελαϊ/τηϲ
	// this split is baked in via RGX above: `^(?P<head>.*?)%s(?P<tail>.*?)$`

	checkfordupes := submatchsrchfinder.FindStringSubmatch(tail)

	if len(checkfordupes) == 0 {
		// only one copy of the search term is in here; just build the tail...
		if len(tt) >= proximity {
			tt = tt[0:proximity]
		}
	} else {
		// no, you have two+ copies of the initial search item in here; recalculate the tail...

		// recover the word/phrase we were looking for
		srchtrm, _ := strings.CutPrefix(submatchsrchfinder.String(), CUTPRE)
		srchtrm, _ = strings.CutSuffix(srchtrm, CUTSUF)

		// calculate the new tail
		tt = IterativeProxWordsMatching(tail, srchtrm, proximity)
	}

	tail = strings.Join(tt, " ") + " "

	if notnear {
		// toss hits
		if !basicprxfinder.MatchString(head) && !basicprxfinder.MatchString(tail) {
			result = p.K
		}
	} else {
		// collect hits

		// pf := fmt.Sprintf("\n\treg\t%s", basicprxfinder.String())
		// pf := ""
		// htv := "[%d]\t%s%s\n\t%t\t%s\n\t%t\t%s"
		// msg(fmt.Sprintf(htv, p.K, p.V, pf, basicprxfinder.MatchString(head), head, basicprxfinder.MatchString(tail), tail), MSGNOTE)

		if basicprxfinder.MatchString(head) || basicprxfinder.MatchString(tail) {
			result = p.K
		}
	}
	return result
}

// IterativeProxWordsMatching - multiple hits for a search term are right on top of one another...
func IterativeProxWordsMatching(text string, sought string, proximity int) []string {
	// [HGS] phr:       ἐϲχάτη χθονόϲ
	// [HGS] headinsidethetail:     πᾶϲα γὰρ ἀγχίαλοϲ
	// [HGS] tail2:     διὸ καὶ μενελαϊ/τηϲ νομὸϲ ἐκεῖ ὡϲ τοιαύτηϲ γῆϲ ὑπὸ μενελάῳ ποτὲ γενομένηϲ  steph byz ἀπόλλωνοϲ πόλιϲ ἐν αἰγύπτῳ πρὸϲ
	// "tail" needs to be longer: proximity = proximity + len(headinsidethetail) + len(searchphrase)

	// the caveat: imagine you have »ἐϲχάτη χθονόϲ« + word1 + word2 + word3 + word4 + word5 + »ἐϲχάτη χθονόϲ« + tail1 + tail2 AND your distance is 2
	// you can't just add everything after  »ἐϲχάτη χθονόϲ« #1 since word3 is not within range of either copy of »ἐϲχάτη χθονόϲ«
	// the rewritten tail should be "word1 word2 word4 word5 ἐϲχάτη χθονόϲ tail1 tail2" [i.e., skip word3]

	// for testing...
	// text := `zero one two ἐϲχάτη χθονόϲ word0 word1 word2 word3 word4 word5 word6 word7 word8 word9 ἐϲχάτη χθονόϲ tail0 tail1 tail2 tail3 tail4 tail5 ἐϲχάτη χθονόϲ rec0 rec1 rec2 rec3`

	var tail []string

	// if "sought" is fancy regex, strings.Split() is not going to work right...
	// segments := strings.Split(text, sought)

	// this is the right way to split; it should be hard to get a compile error since this is a recompilation
	re, e := regexp.Compile(sought)
	chke(e)

	segments := re.Split(text, -1)

	// helper functions

	appenduptoxitems := func(first []string, second []string, items int) []string {
		if len(second) >= items {
			return append(first, second[0:items]...)
		} else {
			return append(first, second[0:]...)
		}
	}

	selectivebuilder := func(split []string) []string {
		embeddedheadsize := len(split)
		if embeddedheadsize <= proximity {
			// caveat is irrelevant; free to grab anything in the tail as we have it
			tail = append(tail, split...)
			tail = append(tail, sought)
		} else {
			// caveat in play; need to bracket off some material
			var pt1 []string
			var pt2 []string
			if embeddedheadsize-proximity > proximity {
				// add prx from the start and prx from the end of the series
				pt1 = split[0:proximity]
				pt2 = split[embeddedheadsize-proximity : embeddedheadsize]
				tail = append(tail, pt1...)
				tail = append(tail, pt2...)
			} else {
				// "zero one two" and not "zero one one two" @ proximity = 2
				tail = append(tail, split...)
			}
			tail = append(tail, sought)
		}
		return tail
	}

	// main loop

	last := len(segments) - 1

	for i := 0; i < len(segments); i++ {
		splt := strings.Split(strings.TrimSpace(segments[i]), " ")
		if i != last {
			tail = selectivebuilder(splt)
		} else {
			tail = appenduptoxitems(tail, splt, proximity)
		}
	}

	return tail
}
