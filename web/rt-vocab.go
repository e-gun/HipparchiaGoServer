//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/format"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/search"
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/labstack/echo/v5"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// RtVocabMaker - get the vocabulary for whatever collection of lines you would be searching
func RtVocabMaker(c *echo.Context) error {
	// todo: worry about γ' for γε

	// item example: <indexedlocation id="index/lt1351/004/4180">⒟ 1.82.18</indexedlocation>

	// big lists become unclickable: "RangeError: Maximum call stack size exceeded" in jquery.min.js
	// see "https://codewithleo.com/rangeerror-maximum-call-stack-size-exceeded/"
	//
	// looking at jquery-3.7.1.js: the error is flagged at line 865
	// RangeError: Maximum call stack size exceeded
	//    at Function.find (jquery.js:865:11) ["push.apply( results, context.getElementsByTagName( selector ) );"]
	//    at jQuery.fn.init.find (jquery.js:2822:11) ["jQuery.find( selector, self[ i ], ret );"]
	//    at jQuery.fn.init (jquery.js:2932:32) ["return ( context || root ).find( selector );"]
	//    at jQuery (jquery.js:159:10)
	//    at <anonymous>:26:2
	//    at DOMEval (jquery.js:130:12)
	//    at domManip (jquery.js:5951:8) ["DOMEval( node.textContent.replace( rcleanScript, "" ), node, doc );"]
	//    at jQuery.fn.init.append (jquery.js:6088:10)
	//    at jQuery.fn.init.<anonymous> (jquery.js:6182:18) ["this.empty().append( value )"]
	//    at access (jquery.js:3905:8) [a bulk caller; "fn.call( elems, value );"]
	//
	// only a major rewrite of jquery would work? pure js in INDEXANDVOCLISTCLICKTOLOOKUP is the other alternative...

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
		<tr class="%s">
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
		<tr class="%s">
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

	user := vlt.ReadUUIDCookie(c)
	if !vlt.AllAuthorized.Check(user) {
		return c.JSONPretty(http.StatusOK, str.CommonJSONOutput{JS: vv.JSVALIDATION}, vv.JSONINDENT)
	}

	defer vlt.LogPaths("RtVocabMaker()")

	start := time.Now()
	s := vlt.AllSessions.GetSess(user)

	id := c.Param("id")
	id = gen.Purgechars(lnch.Config.BadChars, id)

	// "si" is a blank search struct used for progress reporting
	si := search.BuildDefaultSearch(c)
	si.Type = "vocab"

	vlt.WSInfo.UpdateSummMsg <- vlt.WSSIKVs{Key: si.WSID, Val: MSG1}
	vlt.WSInfo.UpdateRemain <- vlt.WSSIKVi{Key: si.WSID, Val: 1}

	// [a] get all the lines you need and turn them into []WordInfo; Headwords to be filled in later
	mx := lnch.Config.MaxText * vv.MAXVOCABLINEGENERATION
	vocabsrch := search.SessionIntoBulkSearch(c, mx) // allow vocab lists to ingest more lines that text & index makers

	if vocabsrch.Results.Len() == 0 {
		return emptyjsreturn(c)
	}

	var slicedwords []str.WordInfo
	rr := vocabsrch.Results.Yield()
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
	hwct := db.ArrayToGetHeadwordCounts(morphslice)

	// [c1] get and map all the DbMorphology
	morphmap := db.ArrayToGetRequiredMorphObjects(morphslice)

	vlt.WSInfo.UpdateSummMsg <- vlt.WSSIKVs{Key: id, Val: MSG2}

	// [c2] map observed words to possibilities
	poss := make(map[string][]str.MorphPossib)

	for k, v := range morphmap {
		poss[k] = extractmorphpossibilities(v.RawPossib)
	}

	morphmap = make(map[string]str.DbMorphology) // clear after use

	// [c3] build a new slice of seen words with headwords attached
	var parsedwords []str.WordInfo
	for _, sw := range slicedwords {
		hww := poss[sw.Word]
		for _, h := range hww {
			newwd := sw
			newwd.HeadWd = h.Headwd
			newwd.Trans = h.Transl
			newwd.HWdCount = hwct[h.Headwd]
			parsedwords = append(parsedwords, newwd)
		}
	}

	// patch the definitions in `parsedwords` with lexical info
	parsedwords = patchparsedwords(parsedwords)

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
	if s.VocScansion {
		scansion = db.ArrayToGetScansion(gen.StringMapKeysIntoSlice(vit))
	}

	// [f1] consolidate the information

	pat := regexp.MustCompile(`^(.{1,3}\.)\s`)

	vim := make(map[string]str.VocInfo)
	for k, v := range vic {
		m := scansion[k]
		if len(m) == 0 {
			// still might return "", of course...
			// but will do "aegyptius" --> "Aegyptĭus"
			m = scansion[cases.Title(language.Und).String(k)]
		}

		vim[k] = str.VocInfo{
			Word:  k,
			C:     v,
			TR:    format.PolishTrans(vit[k], pat),
			Strip: strings.Replace(gen.StripaccentsSTR(k), "ϲ", "σ", -1),
			Metr:  gen.QuantityFixer.Replace(m),
		}
	}

	// flag words that appear only in this selection
	var onlyhere []string
	for i := 0; i < len(parsedwords); i++ {
		if parsedwords[i].HWdCount > 0 && parsedwords[i].HWdCount == vim[parsedwords[i].Word].C {
			onlyhere = append(onlyhere, parsedwords[i].HeadWd)
		}
	}
	onlyhere = gen.Unique(onlyhere)
	onlyhere = gen.PolytonicSort(onlyhere)

	vis := make([]str.VocInfo, len(vim))
	ct := 0
	for _, v := range vim {
		vis[ct] = v
		ct += 1
	}

	vlt.WSInfo.UpdateSummMsg <- vlt.WSSIKVs{Key: id, Val: MSG3}

	// [f2] sort the results
	if s.VocByCount {
		countDecreasing := func(one, two *str.VocInfo) bool {
			return one.C > two.C
		}
		wordIncreasing := func(one, two *str.VocInfo) bool {
			return one.Strip < two.Strip
		}
		str.VIOrderedBy(countDecreasing, wordIncreasing).Sort(vis)
	} else {
		sort.Slice(vis, func(i, j int) bool { return vis[i].Strip < vis[j].Strip })
	}

	vlt.WSInfo.UpdateSummMsg <- vlt.WSSIKVs{Key: id, Val: MSG4}

	// [g] format the output

	headtempl := THH
	if s.VocScansion {
		headtempl = THHS
	}

	trr := make([]string, len(vis)+2)
	trr[0] = headtempl

	uniqueid := 0
	for i, v := range vis {
		uniqueid++
		rc := ""
		if i%2 == 0 {
			rc = "nthrow"
		} else {
			rc = "regular"
		}

		var nt string
		wdid := fmt.Sprintf("%s--%d", v.Word, uniqueid)
		if s.VocScansion {
			nt = fmt.Sprintf(TRRS, rc, wdid, v.Word, v.Metr, v.C, v.TR)
		} else {
			nt = fmt.Sprintf(TRR, rc, wdid, v.Word, v.C, v.TR)
		}
		trr[i+1] = nt
	}
	trr[len(trr)-1] = TCL

	// [g1] build the core: jso.Htm

	htm := strings.Join(trr, "")

	// [g2] build the summary: jso.Sum

	an := search.DbWlnMyAu(&vocabsrch.Results.Lines[0]).Cleaname
	if vocabsrch.TableSize > 1 {
		an = an + fmt.Sprintf(" and %d more author(s)", vocabsrch.TableSize-1)
	}

	wn := search.DbWlnMyWk(&vocabsrch.Results.Lines[0]).Title
	if vocabsrch.SearchSize > 1 {
		wn = wn + fmt.Sprintf(" and %d more works(s)", vocabsrch.SearchSize-1)
	}

	cf := search.DbWlnMyWk(&vocabsrch.Results.Lines[0]).CitationFormat()
	var tc []string
	for _, x := range cf {
		if len(x) != 0 {
			tc = append(tc, x)
		}
	}

	cit := strings.Join(tc, ", ")

	m := message.NewPrinter(language.English)
	wf := m.Sprintf("%d", len(parsedwords))

	el := fmt.Sprintf("%.2f", time.Since(start).Seconds())

	ky := multiworkkeymaker(mp, &vocabsrch)

	cp := ""
	if vocabsrch.Results.Len() == mx {
		cp = m.Sprintf(HITCAP, mx)
	}

	u := len(onlyhere)
	uw := `<p class="indented smallerthannormal">` + strings.Join(onlyhere, ", ") + `</p>`

	sum := fmt.Sprintf(SUMM, an, wn, cit, wf, u, uw, el, cp, ky)

	if s.ZapLunates {
		htm = gen.DeLunate(htm)
	}

	var jso str.CommonJSONOutput
	jso.Sum = sum
	jso.Htm = htm
	jso.Tit = fmt.Sprintf("Vocabulary for %s, %s", an, wn)

	ctl := strings.Replace(vv.INDEXANDVOCLISTCLICKTOLOOKUP, "**REPLACEME**", "vocabobserved", 1)
	jso.JS = "<script>" + ctl + "</script>"

	vlt.WSInfo.Del <- si.WSID
	vlt.WSInfo.Del <- vocabsrch.WSID

	return jsonresponse(c, jso)
}

// patchparsedwords - patch the definitions in `parsedwords` with lexical db definitions
func patchparsedwords(parsedwords []str.WordInfo) []str.WordInfo {
	const (
		SEPARATOR = ` ‖ `
		MAXDFNS   = 20
	)

	var tofind []string
	for _, w := range parsedwords {
		tofind = append(tofind, w.HeadWd)
	}
	tofind = gen.Unique(tofind)
	retranslator := db.BulkEntryTranslations(tofind)

	for i, w := range parsedwords {
		nt, ok := retranslator[w.HeadWd]
		if ok && len(nt) > len(w.Trans) {
			nte := strings.Split(nt, SEPARATOR)
			var newnte []string
			for j, d := range nte {
				if j < MAXDFNS {
					newe := fmt.Sprintf("(%d) %s", j+1, d)
					newnte = append(newnte, newe)
				}
			}
			parsedwords[i].Trans = strings.Join(newnte, "; ")
			if len(nte) > MAXDFNS {
				parsedwords[i].Trans += fmt.Sprintf(" [+ %d additional meanings]", len(nte)-MAXDFNS)
			}
		}
	}
	return parsedwords
}
