package web

import (
	"cmp"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/mps"
	"github.com/e-gun/HipparchiaGoServer/internal/search"
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/labstack/echo/v4"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"net/http"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"
)

// RtIndexMaker - build an index for whatever collection of lines you would be searching
func RtIndexMaker(c echo.Context) error {
	c.Response().After(func() { Msg.LogPaths("RtIndexMaker()") })

	// note that templates + bytes.Buffer is more legible than '%s' time and again BUT it is also slightly slower
	// this was tested via a rewrite of RtIndexMaker() and other rt-textindicesandvocab functions
	// Ar., Acharnians will index via template in 0.35s vs 0.28s without the templates

	// for the bytes.Buffer pattern see FormatNoContextResults() and FormatWithContextResults()

	// a lot of code duplication with RtVocabMaker() but consolidation is not as direct a matter as one might guess

	// THIS HOGS MEMORY DURING SELFTEST(): runtime.GC() does not catch jso data which is still "around" after the function
	// exits (it seems) textindexvocab and vectors are the places where one sees this; anything with a big JSON payload
	// seems to be a problem but a lot of this is hard to reproduce outside of the selftest()

	//[HGS] main() post-initialization runtime.GC() 249M --> 207M
	//[HGS] ArrayToGetRequiredMorphObjects() will search among 86067 words
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

	user := vlt.ReadUUIDCookie(c)
	if !vlt.AllAuthorized.Check(user) {
		return c.JSONPretty(http.StatusOK, JSFeeder{NJ: vv.JSVALIDATION}, vv.JSONINDENT)
	}

	start := time.Now()

	id := c.Param("id")
	id = gen.Purgechars(lnch.Config.BadChars, id)

	// "si" is a blank search struct used for progress reporting
	si := search.BuildDefaultSearch(c)
	si.Type = "index"

	vlt.WSInfo.UpdateSummMsg <- vlt.WSSIKVs{si.WSID, MSG1}
	vlt.WSInfo.UpdateRemain <- vlt.WSSIKVi{si.WSID, 1}

	// [a] gather the lines...

	srch := search.SessionIntoBulkSearch(c, vv.MAXTEXTLINEGENERATION)

	if srch.Results.IsEmpty() {
		return emptyjsreturn(c)
	}

	var slicedwords []str.WordInfo

	rr := srch.Results.YieldAll()
	for r := range rr {
		wds := r.AccentedSlice()
		for _, w := range wds {
			this := str.WordInfo{
				HeadWd:     "",
				Word:       gen.UVσςϲ(gen.SwapAcuteForGrave(w)),
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
	srch.Results.Lines = make([]str.DbWorkline, 1) // clearing after use

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

	// one of the places where you can catch a session reset
	if !vlt.AllSessions.IsInVault(user) {
		return gen.JSONresponse(c, JSFeeder{})
	}

	morphmap := db.ArrayToGetRequiredMorphObjects(morphslice)

	vlt.WSInfo.UpdateSummMsg <- vlt.WSSIKVs{si.ID, MSG2}

	var slicedlookups []str.WordInfo
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
					var additionalword str.WordInfo
					additionalword = w
					additionalword.HeadWd = mps[i].Headwd
					slicedlookups = append(slicedlookups, additionalword)
				}
			}
		}
	}

	morphmap = make(map[string]str.DbMorphology) // drop after use

	// one of the places where you can catch a session reset
	if !vlt.AllSessions.IsInVault(user) {
		return gen.JSONresponse(c, JSFeeder{})
	}

	// keep track of unique values
	globalwordcounts := getwordcounts(gen.StringMapKeysIntoSlice(distinct))
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
	onlyhere = gen.Unique(onlyhere)
	onlyhere = gen.PolytonicSort(onlyhere)

	slicedwords = []str.WordInfo{} // drop after use

	var trimslices []str.WordInfo
	for _, w := range slicedlookups {
		if len(w.HeadWd) != 0 {
			trimslices = append(trimslices, w)
		}
	}

	slicedlookups = []str.WordInfo{} // drop after use

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

	// one of the places where you can catch a session reset
	if !vlt.AllSessions.IsInVault(user) {
		return gen.JSONresponse(c, JSFeeder{})
	}

	// [d] the final map
	// [d1] build it

	vlt.WSInfo.UpdateSummMsg <- vlt.WSSIKVs{si.ID, MSG3}

	indexmap := make(map[gen.PolytonicSorterStruct][]str.WordInfo, len(trimslices))
	for _, w := range trimslices {
		// lunate sigma sorts after omega
		sigma := strings.Replace(gen.StripaccentsSTR(w.HeadWd), "ϲ", "σ", -1)
		ss := gen.PolytonicSorterStruct{
			Sortstring:     sigma + w.HeadWd,
			Originalstring: w.HeadWd,
		}
		indexmap[ss] = append(indexmap[ss], w)
	}

	m := message.NewPrinter(language.English)
	wf := m.Sprintf("%d", len(trimslices))
	trimslices = []str.WordInfo{} // drop after use

	// [d2] sort the keys

	keys := make([]gen.PolytonicSorterStruct, len(indexmap))
	counter := 0
	for k, v := range indexmap {
		k.Count = len(v)
		keys[counter] = k
		counter += 1
	}

	slices.SortFunc(keys, func(a, b gen.PolytonicSorterStruct) int { return cmp.Compare(a.Sortstring, b.Sortstring) })

	// now you have a sorted index...; but a PolytonicSorterStruct does not make for a usable map key...
	plainkeys := make([]string, len(keys))
	for i, k := range keys {
		plainkeys[i] = k.Originalstring
	}

	// example keys: [ἀβαϲάνιϲτοϲ ἀβουλία ἄβουλοϲ ἁβροδίαιτοϲ ἀγαθόϲ ἀγαθόω ἄγαν ...]

	plainmap := make(map[string][]str.WordInfo, len(indexmap))
	for k := range indexmap {
		plainmap[k.Originalstring] = indexmap[k]
	}

	indexmap = make(map[gen.PolytonicSorterStruct][]str.WordInfo, 1) // drop after use

	vlt.WSInfo.UpdateSummMsg <- vlt.WSSIKVs{si.ID, MSG4}

	trr := make([]string, len(plainkeys))
	for i, k := range plainkeys {
		// example
		// k: ἀδικέω; plainmap[k]: []WordInfo -> ἀδικεῖτε, ἀδικηϲάντων, ἀδικούμεθα, ...
		trr[i] = convertwordinfototablerow(plainmap[k])
	}

	htm := fmt.Sprintf(TBLTMP, strings.Join(trr, ""))

	// build the summary info: jso.SU

	an := search.DbWlnMyAu(&firstresult).Cleaname
	if srch.TableSize > 1 {
		an = an + fmt.Sprintf(" and %d more author(s)", srch.TableSize-1)
	}

	wn := search.DbWlnMyWk(&firstresult).Title
	if srch.SearchSize > 1 {
		wn = wn + fmt.Sprintf(" and %d more works(s)", srch.SearchSize-1)
	}

	cf := search.DbWlnMyWk(&firstresult).CitationFormat()
	var tc []string
	for _, x := range cf {
		if len(x) != 0 {
			tc = append(tc, x)
		}
	}

	cit := strings.Join(tc, ", ")

	el := fmt.Sprintf("%.2f", time.Now().Sub(start).Seconds())

	// one of the places where you can catch a session reset
	if !vlt.AllSessions.IsInVault(user) {
		return gen.JSONresponse(c, JSFeeder{})
	}

	ky := multiworkkeymaker(mp, &srch)

	cp := ""
	if linesingested == vv.MAXTEXTLINEGENERATION {
		cp = m.Sprintf(HITCAP, vv.MAXTEXTLINEGENERATION)
	}

	u := len(onlyhere)
	uw := ""
	if u > 0 {
		uw = WLMSG
	}

	oh := WLHTM + strings.Join(onlyhere, ", ") + `</p>`

	sum := fmt.Sprintf(SUMM, an, wn, cit, wf, u, uw, el, cp, ky)

	htm += oh

	if lnch.Config.ZapLunates {
		htm = gen.DeLunate(htm)
	}

	var jso JSFeeder
	jso.SU = sum
	jso.HT = htm

	j := fmt.Sprintf(vv.LEXFINDJS, "indexobserved") + fmt.Sprintf(vv.BROWSERJS, "indexedlocation")

	jso.NJ = fmt.Sprintf("<script>%s</script>", j)

	vlt.WSInfo.Del <- si.WSID
	vlt.WSInfo.Del <- srch.WSID

	return gen.JSONresponse(c, jso)
}

//
// FORMATTING: rt-index.go and rt-vocab.go share these functions
//

// addkeystowordinfo - index to more than one work needs to have a key attached to the citations
func addkeystowordinfo(wii []str.WordInfo) ([]str.WordInfo, map[string]rune) {
	// build the key: 9372 = ⒜
	uu := make([]string, len(wii))
	for i, w := range wii {
		uu[i] = w.Wk
	}
	uu = gen.Unique(uu)
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
func multiworkkeymaker(mapper map[string]rune, srch *str.SearchStruct) string {
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
			t := fmt.Sprintf(`<span class="italic">%s</span>`, mps.AllWorks[k].Title)
			if auu {
				t = mps.DbWkMyAu(mps.AllWorks[k]).Name + ", " + t
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
func convertwordinfototablerow(ww []str.WordInfo) string {
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
	indexmap := make(map[string][]str.WordInfo, len(ww))
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

	keys = gen.PolytonicSort(keys)

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

	tr = str.NoHTML.ReplaceAllString(tr, "")
	elem := strings.Split(tr, "; ")
	for i, e := range elem {
		elem[i] = pat.ReplaceAllString(e, TT)
	}
	return strings.Join(elem, "; ")
}
