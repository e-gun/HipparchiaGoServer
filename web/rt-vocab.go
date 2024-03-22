package web

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/search"
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/labstack/echo/v4"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

// RtVocabMaker - get the vocabulary for whatever collection of lines you would be searching
func RtVocabMaker(c echo.Context) error {
	c.Response().After(func() { Msg.LogPaths("RtVocabMaker()") })

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
	user := vlt.ReadUUIDCookie(c)
	if !vlt.AllAuthorized.Check(user) {
		return c.JSONPretty(http.StatusOK, JSFeeder{NJ: vv.JSVALIDATION}, vv.JSONINDENT)
	}

	start := time.Now()
	se := vlt.AllSessions.GetSess(user)

	id := c.Param("id")
	id = gen.Purgechars(lnch.Config.BadChars, id)

	// "si" is a blank search struct used for progress reporting
	si := search.BuildDefaultSearch(c)
	si.Type = "vocab"

	vlt.WSInfo.UpdateSummMsg <- vlt.WSSIKVs{si.WSID, MSG1}
	vlt.WSInfo.UpdateRemain <- vlt.WSSIKVi{si.WSID, 1}

	// [a] get all the lines you need and turn them into []WordInfo; Headwords to be filled in later
	mx := lnch.Config.MaxText * vv.MAXVOCABLINEGENERATION
	vocabsrch := search.SessionIntoBulkSearch(c, mx) // allow vocab lists to ingest more lines that text & index makers

	if vocabsrch.Results.Len() == 0 {
		return emptyjsreturn(c)
	}

	var slicedwords []str.WordInfo
	rr := vocabsrch.Results.YieldAll()
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
	hwct := db.ArrayToGetTeadwordCounts(morphslice)

	// [c1] get and map all the DbMorphology
	morphmap := db.ArrayToGetRequiredMorphObjects(morphslice)

	vlt.WSInfo.UpdateSummMsg <- vlt.WSSIKVs{id, MSG2}

	// [c2] map observed words to possibilities
	poss := make(map[string][]str.MorphPossib)
	for k, v := range morphmap {
		poss[k] = extractmorphpossibilities(v.RawPossib)
	}

	morphmap = make(map[string]str.DbMorphology) // clear after use

	// [c3] build a new slice of seen words with headwords attached
	var parsedwords []str.WordInfo
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
		scansion = db.ArrayToGetScansion(gen.StringMapKeysIntoSlice(vit))
	}

	// [f1] consolidate the information

	pat := regexp.MustCompile("^(.{1,3}\\.)\\s")

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
			TR:    polishtrans(vit[k], pat),
			Strip: strings.Replace(gen.StripaccentsSTR(k), "ϲ", "σ", -1),
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
	onlyhere = gen.Unique(onlyhere)
	onlyhere = gen.PolytonicSort(onlyhere)

	vis := make([]str.VocInfo, len(vim))
	ct := 0
	for _, v := range vim {
		vis[ct] = v
		ct += 1
	}

	vlt.WSInfo.UpdateSummMsg <- vlt.WSSIKVs{id, MSG3}

	// [f2] sort the results
	if se.VocByCount {
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

	vlt.WSInfo.UpdateSummMsg <- vlt.WSSIKVs{id, MSG4}

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

	el := fmt.Sprintf("%.2f", time.Now().Sub(start).Seconds())

	ky := multiworkkeymaker(mp, &vocabsrch)

	cp := ""
	if vocabsrch.Results.Len() == mx {
		cp = m.Sprintf(HITCAP, mx)
	}

	u := len(onlyhere)
	uw := `<p class="indented smallerthannormal">` + strings.Join(onlyhere, ", ") + `</p>`

	sum := fmt.Sprintf(SUMM, an, wn, cit, wf, u, uw, el, cp, ky)

	if lnch.Config.ZapLunates {
		htm = gen.DeLunate(htm)
	}

	var jso JSFeeder
	jso.SU = sum
	jso.HT = htm

	j := fmt.Sprintf(vv.LEXFINDJS, "vocabobserved")
	jso.NJ = fmt.Sprintf("<script>%s</script>", j)

	vlt.WSInfo.Del <- si.WSID
	vlt.WSInfo.Del <- vocabsrch.WSID

	return gen.JSONresponse(c, jso)
}
