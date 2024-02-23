//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"cmp"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slices"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

type WordInfo struct {
	HeadWd     string
	HWdCount   int
	Word       string
	WdCount    int
	Loc        string
	Cit        string
	IsHomonymn bool
	Trans      string
	Wk         string
}

type VocInfo struct {
	Word         string
	C            int
	TR           string
	Strip        string
	Metr         string
	HWIsOnlyHere bool
	WdIsOnlyHere bool
}

//
// ROUTES
//

// RtTextMaker - make a text of whatever collection of lines you would be searching
func RtTextMaker(c echo.Context) error {
	c.Response().After(func() { messenger.LogPaths("RtTextMaker()") })
	// text generation works like a simple search for "anything" in each line of the selected texts
	// the results then gett output as a big "browser table"...

	const (
		TBLRW = `
            <tr class="browser">
                <td class="browserembeddedannotations">%s</td>
                <td class="browsedline">%s</td>
                <td class="browsercite">%s</td>
            </tr>
		`
		SUMM = `
		<div id="searchsummary">%s,&nbsp;<span class="foundwork">%s</span><br>
		citation format:&nbsp;%s<br></div>`

		SNIP   = `✃✃✃`
		HITCAP = `<span class="small"><span class="red emph">text generation incomplete:</span> hit the cap of %d on allowed lines</span>`
	)

	type JSFeeder struct {
		SU string `json:"searchsummary"`
		HT string `json:"thehtml"`
		JS string `json:"newjs"`
	}

	user := readUUIDCookie(c)
	if !AllAuthorized.Check(user) {
		return c.JSONPretty(http.StatusOK, JSFeeder{JS: JSVALIDATION}, JSONINDENT)
	}

	sess := AllSessions.GetSess(user)
	srch := sessionintobulksearch(c, MAXTEXTLINEGENERATION)

	if srch.Results.Len() == 0 {
		return emptyjsreturn(c)
	}

	// now we have the lines we need....
	firstline := srch.Results.FirstLine()
	firstwork := firstline.MyWk()
	firstauth := firstline.MyAu()

	lines := srch.Results.YieldAll()
	block := make([]string, srch.Results.Len())

	i := 0
	for l := range lines {
		l.PurgeMetadata()
		block[i] = l.MarkedUp
		i++
	}

	whole := strings.Join(block, SNIP)
	whole = textblockcleaner(whole)
	block = strings.Split(whole, SNIP)

	for i = 0; i < len(block); i++ {
		srch.Results.Lines[i].MarkedUp = block[i]
	}

	// delete after use...
	whole = ""
	block = []string{""}

	trr := make([]string, srch.Results.Len())
	previous := srch.Results.FirstLine()
	workcount := 1

	i = 0
	lines = srch.Results.YieldAll()
	for l := range lines {
		cit := selectivelydisplaycitations(l, previous, -1)
		trr[i] = fmt.Sprintf(TBLRW, l.Annotations, l.MarkedUp, cit)
		if l.WkUID != previous.WkUID {
			// you were doing multi-text generation
			workcount += 1
			aw := l.MyAu().Name + fmt.Sprintf(`, <span class="italic">%s</span>`, l.MyWk().Title)
			aw = fmt.Sprintf(`<hr><span class="emph">[%d] %s</span>`, workcount, aw)
			extra := fmt.Sprintf(TBLRW, "", aw, "")
			trr[i] = extra + trr[i]
		}
		previous = l
		i++
	}

	tab := strings.Join(trr, "")
	// that was the body, now do the head and tail
	top := fmt.Sprintf(`<div id="browsertableuid" uid="%s"></div>`, srch.Results.Lines[0].AuID())
	top += `<table><tbody>`
	top += `<tr class="spacing">` + strings.Repeat("&nbsp;", MINBROWSERWIDTH) + `</tr>`

	tab = top + tab + `</tbody></table>`

	// but we don't want/need "observed" tags

	// <div id="searchsummary">Cicero,&nbsp;<span class="foundwork">Philippicae</span><br><br>citation format:&nbsp;oration 3, section 13, line 1<br></div>

	sui := sess.Inclusions

	au := firstauth.Shortname
	if len(sui.Authors) > 1 || len(sui.AuGenres) > 0 || len(sui.AuLocations) > 0 {
		au += " (and others)"
	}

	ti := firstwork.Title
	if len(sui.Works) > 1 || len(sui.WkGenres) > 0 || len(sui.WkLocations) > 0 {
		ti += " (and others)"
	}

	ct := basiccitation(firstline)

	sum := fmt.Sprintf(SUMM, au, ti, ct)

	cp := ""
	if srch.Results.Len() == MAXTEXTLINEGENERATION {
		m := message.NewPrinter(language.English)
		cp = m.Sprintf(HITCAP, MAXTEXTLINEGENERATION)
	}
	sum = sum + cp

	if Config.ZapLunates {
		tab = DeLunate(tab)
	}

	var jso JSFeeder
	jso.SU = sum
	jso.HT = tab
	jso.JS = ""

	WSSIDel <- srch.WSID

	return JSONresponse(c, jso)
}

// RtVocabMaker - get the vocabulary for whatever collection of lines you would be searching
func RtVocabMaker(c echo.Context) error {
	c.Response().After(func() { messenger.LogPaths("RtVocabMaker()") })

	// grab lines via a simple search for "anything" in each line of the selection made and stored in the session
	// todo: worry about γ' for γε

	const (
		SUMM = `
		<div id="searchsummary">Vocabulary for %s,&nbsp;<span class="foundwork">%s</span><br>
			citation format:&nbsp;%s<br>
			%s words found<br>
			Headwords that can be found exclusively in this selection: %d%s<br>
			<span class="small">(%ss)</span><br>
			%s
			%s
		</div>
		`
		THH = `
		<table>
		<tr>
				<th class="vocabtable">word</th>
				<th class="vocabtable">count</th>
				<th class="vocabtable">definitions</th>
		</tr>`

		TRR = `
		<tr>
			<td class="word"><vocabobserved id="%s">%s</vocabobserved></td>
			<td class="count">%d</td>
			<td class="trans">%s</td>
		</tr>`

		THHS = `
		<table>
		<tr>
				<th class="vocabtable">word</th>
				<th class="vocabtable">scansion</th>
				<th class="vocabtable">count</th>
				<th class="vocabtable">definitions</th>
		</tr>`

		TRRS = `
		<tr>
			<td class="word"><vocabobserved id="%s">%s</vocabobserved></td>
			<td class="scansion">%s</td>
			<td class="count">%d</td>
			<td class="trans">%s</td>
		</tr>`

		TCL    = `</table>`
		MSG1   = "Grabbing the lines... (part 1 of 4)"
		MSG2   = "Parsing the vocabulary...(part 2 of 4)"
		MSG3   = "Sifting the vocabulary...(part 3 of 4)"
		MSG4   = "Building the HTML...(part 4 of 4)"
		HITCAP = `<span class="small"><span class="red emph">vocabulary generation incomplete:</span>: hit the cap of %d on allowed lines</span>`
	)

	type JSFeeder struct {
		SU string `json:"searchsummary"`
		HT string `json:"thehtml"`
		NJ string `json:"newjs"`
	}
	user := readUUIDCookie(c)
	if !AllAuthorized.Check(user) {
		return c.JSONPretty(http.StatusOK, JSFeeder{NJ: JSVALIDATION}, JSONINDENT)
	}

	start := time.Now()
	se := AllSessions.GetSess(user)

	id := c.Param("id")
	id = Purgechars(Config.BadChars, id)

	// "si" is a blank search struct used for progress reporting
	si := BuildDefaultSearch(c)
	si.Type = "vocab"

	WSSIUpdateSummMsg <- WSSIKVs{si.WSID, MSG1}
	WSSIUpdateRemain <- WSSIKVi{si.WSID, 1}

	// [a] get all the lines you need and turn them into []WordInfo; Headwords to be filled in later
	mx := Config.MaxText * MAXVOCABLINEGENERATION
	vocabsrch := sessionintobulksearch(c, mx) // allow vocab lists to ingest more lines that text & index makers

	if vocabsrch.Results.Len() == 0 {
		return emptyjsreturn(c)
	}

	var slicedwords []WordInfo
	rr := vocabsrch.Results.YieldAll()
	for r := range rr {
		wds := r.AccentedSlice()
		for _, w := range wds {
			this := WordInfo{
				HeadWd:     "",
				Word:       UVσςϲ(SwapAcuteForGrave(w)),
				Loc:        r.BuildHyperlink(),
				Cit:        r.Citation(),
				IsHomonymn: false,
				Wk:         r.WkUID,
			}
			slicedwords = append(slicedwords, this)
		}
	}

	// [b] find the Unique values we are working with
	distinct := make(map[string]bool, len(slicedwords))
	for _, w := range slicedwords {
		distinct[w.Word] = true
	}

	// [c] prepare to find the headwords for all of these distinct words
	morphslice := make([]string, len(distinct))
	count := 0
	for w := range distinct {
		morphslice[count] = w
		count += 1
	}

	// for flagging words that appear only in this selection
	hwct := arraytogetheadwordcounts(morphslice)

	// [c1] get and map all the DbMorphology
	morphmap := arraytogetrequiredmorphobjects(morphslice)

	WSSIUpdateSummMsg <- WSSIKVs{id, MSG2}

	// [c2] map observed words to possibilities
	poss := make(map[string][]MorphPossib)
	for k, v := range morphmap {
		poss[k] = extractmorphpossibilities(v.RawPossib)
	}

	morphmap = make(map[string]DbMorphology) // clear after use

	// [c3] build a new slice of seen words with headwords attached
	var parsedwords []WordInfo
	for _, s := range slicedwords {
		hww := poss[s.Word]
		for _, h := range hww {
			newwd := s
			newwd.HeadWd = h.Headwd
			newwd.Trans = h.Transl
			newwd.HWdCount = hwct[h.Headwd]
			parsedwords = append(parsedwords, newwd)
		}
	}

	mp := make(map[string]rune)
	if vocabsrch.SearchSize > 1 {
		parsedwords, mp = addkeystowordinfo(parsedwords)
	}

	// [d] get the counts
	vic := make(map[string]int)
	for _, p := range parsedwords {
		vic[p.HeadWd]++
	}

	// [e] get the translations
	vit := make(map[string]string)
	for i := 0; i < len(parsedwords); i++ {
		vit[parsedwords[i].HeadWd] = parsedwords[i].Trans
	}

	scansion := make(map[string]string)
	if se.VocScansion {
		scansion = arraytogetscansion(StringMapKeysIntoSlice(vit))
	}

	// [f1] consolidate the information

	pat := regexp.MustCompile("^(.{1,3}\\.)\\s")

	vim := make(map[string]VocInfo)
	for k, v := range vic {
		m := scansion[k]
		if len(m) == 0 {
			// still might return "", of course...
			// but will do "aegyptius" --> "Aegyptĭus"
			m = scansion[cases.Title(language.Und).String(k)]
		}

		vim[k] = VocInfo{
			Word:  k,
			C:     v,
			TR:    polishtrans(vit[k], pat),
			Strip: strings.Replace(StripaccentsSTR(k), "ϲ", "σ", -1),
			Metr:  quantityfixer.Replace(m),
		}
	}

	// flag words that appear only in this selection
	var onlyhere []string
	for i := 0; i < len(parsedwords); i++ {
		if parsedwords[i].HWdCount > 0 && parsedwords[i].HWdCount == vim[parsedwords[i].Word].C {
			onlyhere = append(onlyhere, parsedwords[i].HeadWd)
		}
	}
	onlyhere = Unique(onlyhere)
	onlyhere = PolytonicSort(onlyhere)

	vis := make([]VocInfo, len(vim))
	ct := 0
	for _, v := range vim {
		vis[ct] = v
		ct += 1
	}

	WSSIUpdateSummMsg <- WSSIKVs{id, MSG3}

	// [f2] sort the results
	if se.VocByCount {
		countDecreasing := func(one, two *VocInfo) bool {
			return one.C > two.C
		}
		wordIncreasing := func(one, two *VocInfo) bool {
			return one.Strip < two.Strip
		}
		VIOrderedBy(countDecreasing, wordIncreasing).Sort(vis)
	} else {
		sort.Slice(vis, func(i, j int) bool { return vis[i].Strip < vis[j].Strip })
	}

	WSSIUpdateSummMsg <- WSSIKVs{id, MSG4}

	// [g] format the output

	headtempl := THH
	if se.VocScansion {
		headtempl = THHS
	}

	trr := make([]string, len(vis)+2)
	trr[0] = headtempl
	for i, v := range vis {
		var nt string
		if se.VocScansion {
			nt = fmt.Sprintf(TRRS, v.Word, v.Word, v.Metr, v.C, v.TR)
		} else {
			nt = fmt.Sprintf(TRR, v.Word, v.Word, v.C, v.TR)
		}
		trr[i+1] = nt
	}
	trr[len(trr)-1] = TCL

	// [g1] build the core: jso.HT

	htm := strings.Join(trr, "")

	// [g2] build the summary: jso.SU

	an := vocabsrch.Results.Lines[0].MyAu().Cleaname
	if vocabsrch.TableSize > 1 {
		an = an + fmt.Sprintf(" and %d more author(s)", vocabsrch.TableSize-1)
	}

	wn := vocabsrch.Results.Lines[0].MyWk().Title
	if vocabsrch.SearchSize > 1 {
		wn = wn + fmt.Sprintf(" and %d more works(s)", vocabsrch.SearchSize-1)
	}

	cf := vocabsrch.Results.Lines[0].MyWk().CitationFormat()
	var tc []string
	for _, x := range cf {
		if len(x) != 0 {
			tc = append(tc, x)
		}
	}

	cit := strings.Join(tc, ", ")

	m := message.NewPrinter(language.English)
	wf := m.Sprintf("%d", len(parsedwords))

	el := fmt.Sprintf("%.2f", time.Now().Sub(start).Seconds())

	ky := multiworkkeymaker(mp, &vocabsrch)

	cp := ""
	if vocabsrch.Results.Len() == mx {
		cp = m.Sprintf(HITCAP, mx)
	}

	u := len(onlyhere)
	uw := `<p class="indented smallerthannormal">` + strings.Join(onlyhere, ", ") + `</p>`

	sum := fmt.Sprintf(SUMM, an, wn, cit, wf, u, uw, el, cp, ky)

	if Config.ZapLunates {
		htm = DeLunate(htm)
	}

	var jso JSFeeder
	jso.SU = sum
	jso.HT = htm

	j := fmt.Sprintf(LEXFINDJS, "vocabobserved")
	jso.NJ = fmt.Sprintf("<script>%s</script>", j)

	WSSIDel <- si.WSID
	WSSIDel <- vocabsrch.WSID

	return JSONresponse(c, jso)
}

// RtIndexMaker - build an index for whatever collection of lines you would be searching
func RtIndexMaker(c echo.Context) error {
	c.Response().After(func() { messenger.LogPaths("RtIndexMaker()") })

	// note that templates + bytes.Buffer is more legible than '%s' time and again BUT it is also slightly slower
	// this was tested via a rewrite of RtIndexMaker() and other rt-textindicesandvocab functions
	// Ar., Acharnians will index via template in 0.35s vs 0.28s without the templates

	// for the bytes.Buffer pattern see FormatNoContextResults() and FormatWithContextResults()

	// a lot of code duplication with RtVocabMaker() but consolidation is not as direct a matter as one might guess

	// THIS HOGS MEMORY DURING SELFTEST(): runtime.GC() does not catch jso data which is still "around" after the function
	// exits (it seems) textindexvocab and vectors are the places where one sees this; anything with a big JSON payload
	// seems to be a problem but a lot of this is hard to reproduce outside of the selftest()

	//[HGS] main() post-initialization runtime.GC() 249M --> 207M
	//[HGS] arraytogetrequiredmorphobjects() will search among 86067 words
	//[HGS] RtIndexMaker() runtime.GC() 394M --> 245M
	// ... [wait] ...
	//[HGS] Starting polling loop for b045a683
	//[HGS] RtSearch() runtime.GC() 240M --> 208M

	const (
		TBLTMP = `        
		<table>
		<tbody><tr>
			<th class="indextable">headword</th>
			<th class="indextable">word</th>
			<th class="indextable">count</th>
			<th class="indextable">passages</th>
		</tr>
		%s
		</table>`

		SUMM = `
		<div id="searchsummary">Index to %s,&nbsp;<span class="foundwork">%s</span><br>
			citation format:&nbsp;%s<br>
			%s words found<br>
			Forms that can be found exclusively in this selection: %d%s<br>
			<span class="small">(%ss)</span><br>
			%s
			%s
			<br>
			(NB: <span class="homonym">homonymns</span> will appear under more than one headword)
		</div>
	`
		UPW    = "ϙϙϙϙϙϙϙϙ<br>unparsed words"
		MSG1   = "Grabbing the lines...&nbsp;(part 1 of 4)"
		MSG2   = "Parsing the vocabulary...&nbsp;(part 2 of 4)"
		MSG3   = "Sifting the index...&nbsp;(part 3 of 4)"
		MSG4   = "Building the HTML...&nbsp;(part 4 of 4)"
		HITCAP = `<span class="small"><span class="red emph">indexing incomplete:</span>: hit the cap of %d on allowed lines</span>`
		WLMSG  = `&nbsp; <span class="smallerthannormal">(a list these words appears after the index to the whole)</span>`
		WLHTM  = `<p class="emph">Words that appear only here in the whole database:</p><p class="indented smallerthannormal">`
	)

	type JSFeeder struct {
		SU string `json:"searchsummary"`
		HT string `json:"thehtml"`
		NJ string `json:"newjs"`
	}

	user := readUUIDCookie(c)
	if !AllAuthorized.Check(user) {
		return c.JSONPretty(http.StatusOK, JSFeeder{NJ: JSVALIDATION}, JSONINDENT)
	}

	start := time.Now()

	id := c.Param("id")
	id = Purgechars(Config.BadChars, id)

	// "si" is a blank search struct used for progress reporting
	si := BuildDefaultSearch(c)
	si.Type = "index"

	WSSIUpdateSummMsg <- WSSIKVs{si.WSID, MSG1}
	WSSIUpdateRemain <- WSSIKVi{si.WSID, 1}

	// [a] gather the lines...

	srch := sessionintobulksearch(c, MAXTEXTLINEGENERATION)

	if srch.Results.IsEmpty() {
		return emptyjsreturn(c)
	}

	var slicedwords []WordInfo

	rr := srch.Results.YieldAll()
	for r := range rr {
		wds := r.AccentedSlice()
		for _, w := range wds {
			this := WordInfo{
				HeadWd:     "",
				Word:       UVσςϲ(SwapAcuteForGrave(w)),
				Loc:        r.BuildHyperlink(),
				Cit:        r.Citation(),
				IsHomonymn: false,
				Wk:         r.WkUID,
			}
			slicedwords = append(slicedwords, this)
		}
	}

	firstresult := srch.Results.FirstLine()
	linesingested := srch.Results.Len()
	srch.Results.Lines = make([]DbWorkline, 1) // clearing after use

	// [b] find the Unique values
	distinct := make(map[string]bool, len(slicedwords))
	for _, w := range slicedwords {
		distinct[w.Word] = true
	}

	// [c] find the headwords for all of these distinct words
	morphslice := make([]string, len(distinct))
	count := 0
	for w := range distinct {
		morphslice[count] = w
		count += 1
	}

	// [c1] map words to a dbMorphology

	morphmap := arraytogetrequiredmorphobjects(morphslice)

	WSSIUpdateSummMsg <- WSSIKVs{si.ID, MSG2}

	var slicedlookups []WordInfo
	for _, w := range slicedwords {
		emm := false
		mme := w.Word
		if _, ok := morphmap[w.Word]; !ok {
			// here is where you check to see if the word + an apostrophe can be found: γ is really γ' (i.e. γε)
			// this also means that you had to grab all of those extra forms in the first place
			if _, y := morphmap[w.Word+"'"]; y {
				emm = true
				w.Word = w.Word + "'"
				mme = w.Word
			} else {
				w.HeadWd = UPW
				slicedlookups = append(slicedlookups, w)
			}
		} else {
			emm = true
		}

		if emm {
			mps := extractmorphpossibilities(morphmap[mme].RawPossib)
			if len(mps) > 1 {
				for i := 0; i < len(mps); i++ {
					var additionalword WordInfo
					additionalword = w
					additionalword.HeadWd = mps[i].Headwd
					slicedlookups = append(slicedlookups, additionalword)
				}
			}
		}
	}

	morphmap = make(map[string]DbMorphology) // drop after use

	// keep track of unique values
	globalwordcounts := getwordcounts(StringMapKeysIntoSlice(distinct))
	localwordcounts := make(map[string]int)
	for _, k := range slicedwords {
		localwordcounts[k.Word] += 1
	}

	// flag words that appear only in this selection
	var onlyhere []string
	for w, lc := range localwordcounts {
		if globalwordcounts[w].Total == lc {
			onlyhere = append(onlyhere, w)
		}
	}
	onlyhere = Unique(onlyhere)
	onlyhere = PolytonicSort(onlyhere)

	slicedwords = []WordInfo{} // drop after use

	var trimslices []WordInfo
	for _, w := range slicedlookups {
		if len(w.HeadWd) != 0 {
			trimslices = append(trimslices, w)
		}
	}

	slicedlookups = []WordInfo{} // drop after use

	// pseudocode:

	//calculate homonyms: two maps
	// [a] map ishom: [string]bool
	// [b] map tester: [string]string: [word]headword
	// iterate
	// 	if word not in map: add
	// 	if word in map: is assoc w/ this headword?
	// 	if not: w is homonym

	ishom := make(map[string]bool)
	htest := make(map[string]string)
	for _, t := range trimslices {
		if _, ok := htest[t.Word]; !ok {
			htest[t.Word] = t.HeadWd
		} else {
			if htest[t.Word] != t.HeadWd {
				ishom[t.Word] = true
			}
		}
	}

	for i, t := range trimslices {
		if ishom[t.Word] {
			trimslices[i].IsHomonymn = true
		}
	}

	// last chance to add in keys for multiple work indices
	mp := make(map[string]rune)
	if srch.SearchSize > 1 {
		trimslices, mp = addkeystowordinfo(trimslices)
	}

	// [d] the final map
	// [d1] build it

	WSSIUpdateSummMsg <- WSSIKVs{si.ID, MSG3}

	indexmap := make(map[PolytonicSorterStruct][]WordInfo, len(trimslices))
	for _, w := range trimslices {
		// lunate sigma sorts after omega
		sigma := strings.Replace(StripaccentsSTR(w.HeadWd), "ϲ", "σ", -1)
		ss := PolytonicSorterStruct{
			sortstring:     sigma + w.HeadWd,
			originalstring: w.HeadWd,
		}
		indexmap[ss] = append(indexmap[ss], w)
	}

	m := message.NewPrinter(language.English)
	wf := m.Sprintf("%d", len(trimslices))
	trimslices = []WordInfo{} // drop after use

	// [d2] sort the keys

	keys := make([]PolytonicSorterStruct, len(indexmap))
	counter := 0
	for k, v := range indexmap {
		k.count = len(v)
		keys[counter] = k
		counter += 1
	}

	slices.SortFunc(keys, func(a, b PolytonicSorterStruct) int { return cmp.Compare(a.sortstring, b.sortstring) })

	// now you have a sorted index...; but a PolytonicSorterStruct does not make for a usable map key...
	plainkeys := make([]string, len(keys))
	for i, k := range keys {
		plainkeys[i] = k.originalstring
	}

	// example keys: [ἀβαϲάνιϲτοϲ ἀβουλία ἄβουλοϲ ἁβροδίαιτοϲ ἀγαθόϲ ἀγαθόω ἄγαν ...]

	plainmap := make(map[string][]WordInfo, len(indexmap))
	for k := range indexmap {
		plainmap[k.originalstring] = indexmap[k]
	}

	indexmap = make(map[PolytonicSorterStruct][]WordInfo, 1) // drop after use

	WSSIUpdateSummMsg <- WSSIKVs{si.ID, MSG4}

	trr := make([]string, len(plainkeys))
	for i, k := range plainkeys {
		// example
		// k: ἀδικέω; plainmap[k]: []WordInfo -> ἀδικεῖτε, ἀδικηϲάντων, ἀδικούμεθα, ...
		trr[i] = convertwordinfototablerow(plainmap[k])
	}

	htm := fmt.Sprintf(TBLTMP, strings.Join(trr, ""))

	// build the summary info: jso.SU

	an := firstresult.MyAu().Cleaname
	if srch.TableSize > 1 {
		an = an + fmt.Sprintf(" and %d more author(s)", srch.TableSize-1)
	}

	wn := firstresult.MyWk().Title
	if srch.SearchSize > 1 {
		wn = wn + fmt.Sprintf(" and %d more works(s)", srch.SearchSize-1)
	}

	cf := firstresult.MyWk().CitationFormat()
	var tc []string
	for _, x := range cf {
		if len(x) != 0 {
			tc = append(tc, x)
		}
	}

	cit := strings.Join(tc, ", ")

	el := fmt.Sprintf("%.2f", time.Now().Sub(start).Seconds())

	ky := multiworkkeymaker(mp, &srch)

	cp := ""
	if linesingested == MAXTEXTLINEGENERATION {
		cp = m.Sprintf(HITCAP, MAXTEXTLINEGENERATION)
	}

	u := len(onlyhere)
	uw := ""
	if u > 0 {
		uw = WLMSG
	}

	oh := WLHTM + strings.Join(onlyhere, ", ") + `</p>`

	sum := fmt.Sprintf(SUMM, an, wn, cit, wf, u, uw, el, cp, ky)

	htm += oh

	if Config.ZapLunates {
		htm = DeLunate(htm)
	}

	var jso JSFeeder
	jso.SU = sum
	jso.HT = htm

	j := fmt.Sprintf(LEXFINDJS, "indexobserved") + fmt.Sprintf(BROWSERJS, "indexedlocation")

	jso.NJ = fmt.Sprintf("<script>%s</script>", j)

	WSSIDel <- si.WSID
	WSSIDel <- srch.WSID

	return JSONresponse(c, jso)
}

//
// HELPERS
//

// sessionintobulksearch - grab every line of text in the currently registerselection set of authors, works, and passages
func sessionintobulksearch(c echo.Context, lim int) SearchStruct {
	user := readUUIDCookie(c)
	sess := AllSessions.GetSess(user)

	srch := BuildDefaultSearch(c)
	srch.Seeking = ""
	srch.CurrentLimit = lim
	srch.InitSum = "(gathering and formatting lines of text)"
	srch.ID = strings.Replace(uuid.New().String(), "-", "", -1)

	srch.CleanInput()

	sl := SessionIntoSearchlist(sess)
	srch.SearchIn = sl.Inc
	srch.SearchEx = sl.Excl
	srch.SearchSize = sl.Size
	SSBuildQueries(&srch)
	srch.IsActive = true
	srch.TableSize = len(srch.Queries)
	SearchAndInsertResults(&srch)
	return srch
}

// arraytogetscansion - grab all scansions for a slice of words and return as a map
func arraytogetscansion(wordlist []string) map[string]string {
	const (
		TT = `CREATE TEMPORARY TABLE ttw_%s AS SELECT words AS w FROM unnest(ARRAY[%s]) words`
		QT = `SELECT entry_name, metrical_entry FROM %s_dictionary WHERE EXISTS 
				(SELECT 1 FROM ttw_%s temptable WHERE temptable.w = %s_dictionary.entry_name)`
	)

	type entryandmeter struct {
		Entry string
		Meter string
	}

	// look for the upper case matches too: Ϲωκράτηϲ and not just ϲωκρατέω (!)
	uppers := make([]string, len(wordlist))
	for i := 0; i < len(wordlist); i++ {
		uppers[i] = strings.Title(wordlist[i])
	}

	wordlist = append(wordlist, uppers...)

	dbconn := GetDBConnection()
	defer dbconn.Release()

	foundmetrics := make(map[string]string)
	var thehit entryandmeter

	foreach := []any{&thehit.Entry, &thehit.Meter}

	rwfnc := func() error {
		foundmetrics[thehit.Entry] = thehit.Meter
		return nil
	}

	// a waste of time to check the language on every word; just flail/fail once
	for _, uselang := range TheLanguages {
		u := strings.Replace(uuid.New().String(), "-", "", -1)
		id := fmt.Sprintf("%s_%s_mw", u, uselang)
		a := fmt.Sprintf("'%s'", strings.Join(wordlist, "', '"))
		t := fmt.Sprintf(TT, id, a)

		_, err := dbconn.Exec(context.Background(), t)
		chke(err)

		foundrows, e := dbconn.Query(context.Background(), fmt.Sprintf(QT, uselang, id, uselang))
		chke(e)

		_, ee := pgx.ForEachRow(foundrows, foreach, rwfnc)
		chke(ee)
	}
	return foundmetrics
}

// arraytogetrequiredmorphobjects - map a slice of words to the corresponding DbMorphology
func arraytogetrequiredmorphobjects(wordlist []string) map[string]DbMorphology {
	// hipparchiaDB=# \d greek_morphology
	//                           Table "public.greek_morphology"
	//          Column           |          Type          | Collation | Nullable | Default
	//---------------------------+------------------------+-----------+----------+---------
	// observed_form             | character varying(64)  |           |          |
	// xrefs                     | character varying(128) |           |          |
	// prefixrefs                | character varying(128) |           |          |
	// possible_dictionary_forms | jsonb                  |           |          |
	// related_headwords         | character varying(256) |           |          |
	//Indexes:
	//    "greek_analysis_trgm_idx" gin (related_headwords gin_trgm_ops)
	//    "greek_morphology_idx" btree (observed_form)

	const (
		TT = `CREATE TEMPORARY TABLE ttw_%s AS SELECT words AS w FROM unnest(ARRAY[%s]) words`
		QT = `SELECT observed_form, xrefs, prefixrefs, possible_dictionary_forms, related_headwords FROM %s_morphology WHERE EXISTS 
				(SELECT 1 FROM ttw_%s temptable WHERE temptable.w = %s_morphology.observed_form)`
		MSG1      = "arraytogetrequiredmorphobjects() will search among %d words"
		CHUNKSIZE = 999999
	)

	dbconn := GetDBConnection()
	defer dbconn.Release()

	// look for the upper case matches too: Ϲωκράτηϲ and not just ϲωκρατέω (!)
	uppers := make([]string, len(wordlist))
	for i := 0; i < len(wordlist); i++ {
		uppers[i] = strings.Title(wordlist[i])
	}

	// γ': a lot of cycles looking for a small number of words...
	apo := make([]string, len(wordlist))
	for i := 0; i < len(wordlist); i++ {
		// need to escape the single quote
		// hipparchiaDB=# select * from greek_morphology where observed_form = 'οὑφ'''
		apo[i] = wordlist[i] + "''"
	}

	wordlist = append(wordlist, uppers...)
	wordlist = append(wordlist, apo...)

	msg(fmt.Sprintf(MSG1, len(wordlist)), MSGPEEK)

	foundmorph := make(map[string]DbMorphology)
	var thehit DbMorphology

	foreach := []any{&thehit.Observed, &thehit.Xrefs, &thehit.PrefixXrefs, &thehit.RawPossib, &thehit.RelatedHW}

	rwfnc := func() error {
		foundmorph[thehit.Observed] = thehit
		return nil
	}

	// vectorization revealed that 10m words is too much for this function
	// [HGS] arraytogetrequiredmorphobjects() will search for 10708941 words
	// [Hipparchia Golang Server v.1.2.0a] UNRECOVERABLE ERROR: PLEASE TAKE NOTE OF THE FOLLOWING PANIC MESSAGE
	// ERROR: invalid memory alloc request size 1073741824 (SQLSTATE XX000)

	// this could be parallelized...

	chunkedlist := ChunkSlice(wordlist, CHUNKSIZE)
	for _, cl := range chunkedlist {
		// a waste of time to check the language on every word; just flail/fail once
		for _, uselang := range TheLanguages {
			u := strings.Replace(uuid.New().String(), "-", "", -1)
			id := fmt.Sprintf("%s_%s_mw", u, uselang)
			a := fmt.Sprintf("'%s'", strings.Join(cl, "', '"))
			t := fmt.Sprintf(TT, id, a)

			_, err := dbconn.Exec(context.Background(), t)
			chke(err)

			foundrows, e := dbconn.Query(context.Background(), fmt.Sprintf(QT, uselang, id, uselang))
			chke(e)

			_, ee := pgx.ForEachRow(foundrows, foreach, rwfnc)
			chke(ee)
		}
	}
	return foundmorph
}

func arraytogetheadwordcounts(wordlist []string) map[string]int {
	const (
		TT = `CREATE TEMPORARY TABLE ttw_%s AS SELECT words AS w FROM unnest(ARRAY[%s]) words`
		QT = `SELECT entry_name , total_count FROM dictionary_headword_wordcounts WHERE EXISTS 
				(SELECT 1 FROM ttw_%s temptable WHERE temptable.w = dictionary_headword_wordcounts.entry_name)`
	)

	dbconn := GetDBConnection()
	defer dbconn.Release()

	countmap := make(map[string]int)

	type tempstruct struct {
		w string
		c int
	}

	var thehit tempstruct

	foreach := []any{&thehit.w, &thehit.c}

	rwfnc := func() error {
		countmap[thehit.w] = thehit.c
		return nil
	}

	u := strings.Replace(uuid.New().String(), "-", "", -1)
	a := fmt.Sprintf("'%s'", strings.Join(wordlist, "', '"))

	t := fmt.Sprintf(TT, u, a)
	_, err := dbconn.Exec(context.Background(), t)
	chke(err)

	foundrows, e := dbconn.Query(context.Background(), fmt.Sprintf(QT, u))
	chke(e)

	_, ee := pgx.ForEachRow(foundrows, foreach, rwfnc)
	chke(ee)

	return countmap
}

//
// FORMATTING
//

// addkeystowordinfo - index to more than one work needs to have a key attached to the citations
func addkeystowordinfo(wii []WordInfo) ([]WordInfo, map[string]rune) {
	// build the key: 9372 = ⒜
	uu := make([]string, len(wii))
	for i, w := range wii {
		uu[i] = w.Wk
	}
	uu = Unique(uu)
	sort.Slice(uu, func(i, j int) bool { return uu[i] < uu[j] })
	mp := make(map[string]rune)
	for i, u := range uu {
		mp[u] = rune(i + 9372)
	}

	for i, w := range wii {
		wii[i].Cit = fmt.Sprintf("%s %s", string(mp[w.Wk]), wii[i].Cit)
	}

	return wii, mp
}

// multiworkkeymaker - index to more than one work needs a key to the whole
func multiworkkeymaker(mapper map[string]rune, srch *SearchStruct) string {
	// <br><span class="emph">Works:</span> ⒜: <span class="italic">De caede Eratosthenis</span>
	//; ⒝: <span class="italic">Epitaphius [Sp.]</span>
	//; ⒞: <span class="italic">Contra Simonem</span> ...
	const (
		START = `<br><span class="emph">Works:</span> `
	)

	ky := ""
	wkk := srch.SearchSize > 1
	auu := srch.TableSize > 1

	if auu || wkk {
		var out []string
		for k, v := range mapper {
			t := fmt.Sprintf(`<span class="italic">%s</span>`, AllWorks[k].Title)
			if auu {
				t = AllWorks[k].MyAu().Name + ", " + t
			}
			out = append(out, fmt.Sprintf("%s: %s\n", string(v), t))
		}
		sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
		ky = strings.Join(out, "; ")
		ky = START + ky
	}
	return ky
}

// convertwordinfototablerow - []WordInfo --> "<tr>...</tr>"
func convertwordinfototablerow(ww []WordInfo) string {
	// every word has the same headword
	// now we build a sub-map after the pattern of the main map: but now the keys are the words, not the headwords

	// example:
	// 	<tr>
	//		<td class="headword"><indexobserved id=""></indexobserved></td>
	//		<td class="word"><indexobserved id="διδόντεϲ">διδόντεϲ</indexobserved></td>
	//		<td class="count">2</td>
	//		<td class="passages"><indexedlocation id="index/gr0540/015/3831">⒪ 2.4</indexedlocation>, <indexedlocation id="index/gr0540/025/5719">⒴ 32.5</indexedlocation></td>
	//	</tr>

	const (
		TBLRW = `
		<tr>
			<td class="headword"><indexobserved id="%s">%s</indexobserved></td>
			<td class="word"><indexobserved id="%s">%s</indexobserved></td>
			<td class="count">%d</td>
			<td class="passages">%s</td>
		</tr>`

		HMNTBLRW = `
		<tr>
			<td class="headword"><indexobserved id="%s">%s</indexobserved></td>
			<td class="word"><span class="homonym"><indexobserved id="%s">%s</indexobserved></span></td>
			<td class="count">%d</td>
			<td class="passages">%s</td>
		</tr>`

		IDXLOC = `<indexedlocation id="%s">%s</indexedlocation>`
	)

	// build it
	indexmap := make(map[string][]WordInfo, len(ww))
	for _, w := range ww {
		indexmap[w.Word] = append(indexmap[w.Word], w)
	}

	// sort the keys
	keys := make([]string, len(indexmap))
	count := 0
	for k := range indexmap {
		keys[count] = k
		count += 1
	}

	keys = PolytonicSort(keys)

	trr := make([]string, len(keys))
	used := make(map[string]bool)
	for i, k := range keys {
		wii := indexmap[k]
		hw := ""
		if used[wii[0].HeadWd] {
			// skip
		} else {
			hw = wii[0].HeadWd
		}

		sort.Slice(wii, func(i, j int) bool { return wii[i].Loc < wii[j].Loc })

		// get all passages related to this word
		var pp []string
		dedup := make(map[string]bool) // this is hacky: why duplicates to begin with?
		for j := 0; j < len(wii); j++ {
			if _, ok := dedup[wii[j].Loc]; !ok {
				pp = append(pp, fmt.Sprintf(IDXLOC, wii[j].Loc, wii[j].Cit))
				dedup[wii[j].Loc] = true
			}
		}
		p := strings.Join(pp, ", ")

		templ := TBLRW
		if wii[0].IsHomonymn {
			templ = HMNTBLRW
		}

		t := fmt.Sprintf(templ, hw, hw, wii[0].Word, wii[0].Word, len(pp), p)
		trr[i] = t
		used[wii[0].HeadWd] = true
	}

	out := strings.Join(trr, "")
	return out
}

// polishtrans - add "transtree" spans to the mini-translation lists to highlight structure
func polishtrans(tr string, pat *regexp.Regexp) string {
	// don't loop "pat". it's not really a variable. here it is:
	// pat := regexp.MustCompile("^(.{1,3}\\.)\\s")

	// sample:
	// <span class="transtree">A.</span> as Adv., bearing the same relation to ὡϲ as ὅϲτε to ὅϲ, and used by Hom.
	// more freq. than ὡϲ in similes, when it is commonly written divisim, and is relat. to a demonstr. ὥϲ: sts. c. pres. Indic;
	// <span class="transtree">B.</span> the actual

	const (
		TT = `<span class="transtree">$1</span> `
	)

	tr = NoHTML.ReplaceAllString(tr, "")
	elem := strings.Split(tr, "; ")
	for i, e := range elem {
		elem[i] = pat.ReplaceAllString(e, TT)
	}
	return strings.Join(elem, "; ")
}
