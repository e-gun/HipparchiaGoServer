//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package format

import (
	"fmt"
	"maps"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/mps"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// FmtLexicalOutput - turn a DbLexicon word into HTML
func FmtLexicalOutput(w str.DbLexicon) string {
	const (
		HEADTEMPL = `<div id="%s_%s"><hr>
		<p class="dictionaryheading" id="%s_%s">%s &nbsp;%s</p>
	`
		FORMSUMM = `<formsummary parserxref="%d" lexicalid="%s" headword="%s" lang="%s">%d known forms</formsummary>`

		FRQSUM = `<p class="wordcounts">Relative frequency: <span class="blue">%s</span></p>`

		NAVTABLE = `
		<table class="navtable">
			<tbody>
			<tr>
				<td class="alignleft">
					<span class="label">Previous: </span>
					<dictionaryidsearch entryid="%s" language="%s">%s</dictionaryidsearch>
				</td>
				<td>&nbsp;</td>
				<td class="alignright">
					<span class="label">Next: </span>
					<dictionaryidsearch entryid="%s" language="%s">%s</dictionaryidsearch>
				</td>
			</tr>
			</tbody>
		</table>`
	)
	var elem []string

	// [h1] first part of a lexical entry:

	var met string
	if w.EntryMetr != "" {
		met = gen.QuantityFixer.Replace(w.EntryMetr)
		met = gen.NumberStripper.Replace(met)
		met = fmt.Sprintf("\n<span class=\"entryvowelquantities\">⸤%s⸥</span>", met)
		// but do not show if there are no long/short marks
		if !strings.ContainsAny(met, "̄̆̄ᾰᾱῐῑῠῡ") {
			met = ""
		}
	}

	elem = append(elem, fmt.Sprintf(HEADTEMPL, w.EntryName, w.IdString, w.EntryName, w.IdString, w.EntryName, met))

	// [h1a] known forms in use

	hwc := db.GetIndividualHeadwordCount(w.EntryName)
	elem = append(elem, fmt.Sprintf(FRQSUM, hwc.FrqClas))

	lw := gen.UVσςϲ(w.EntryName) // otherwise "venio" will hit AllLemm instead of "uenio"
	if _, ok := mps.AllLemm[lw]; ok {
		elem = append(elem, fmt.Sprintf(FORMSUMM, mps.AllLemm[lw].Xref, w.IdString, w.EntryName, w.GetLang(), len(mps.AllLemm[lw].Deriv)))
	}

	// [h1b] principle parts

	// TODO: but not at all a priority for v1.x

	// [h2] wordcounts data including weighted distributions

	elem = append(elem, `<div class="wordcounts">`)
	elem = append(elem, headwordprevalencebycorp(hwc))
	elem = append(elem, headworddistribbycorp(hwc))
	elem = append(elem, headworddistribbyera(hwc))
	elem = append(elem, headworddistribbygenre(hwc))
	elem = append(elem, `</div>`)

	// [h4] the actual body of the entry

	elem = append(elem, formatpreliminfo(w))

	if !w.IsLatin() {
		for _, s := range w.Senses {
			elem = append(elem, formatgksenseinfo(s))
		}
	} else {
		elem = append(elem, formatltsenseinfo(w.Senses))
	}

	// [h5] previous & next entry

	prev := db.FindProximateEntry(w, "prev")
	nxt := db.FindProximateEntry(w, "next")

	pn := fmt.Sprintf(NAVTABLE, prev.IdString, w.GetLang(), prev.EntryName, nxt.IdString, w.GetLang(), nxt.EntryName)
	elem = append(elem, pn)

	html := strings.Join(elem, "")

	// fmt.Println(html)
	return html
}

// FmtLexPrevalenceData - turn a wordcount into an HTML summary
func FmtLexPrevalenceData(w str.DbUnparsedWordCounts, s string) string {
	// <p class="wordcounts">Prevalence (all forms): <span class="prevalence">Ⓣ</span> 1482 / <span class="prevalence">Ⓖ</span> 1415 / <span class="prevalence">Ⓓ</span> 54 / <span class="prevalence">Ⓘ</span> 11 / <span class="prevalence">Ⓒ</span> 2</p>
	const (
		PDPAR = `<p class="wordcounts">Prevalence of <span class="emph">%s</span>: %s</p>`
		PDSPA = `<span class="rarechars prevalence">%s</span> %d`
	)

	m := message.NewPrinter(language.English)

	labels := map[string]string{"Total": "Ⓣ", "TLG": "Ⓖ", "LAT": "Ⓛ", "DDP": "Ⓓ", "INS": "Ⓘ", "CHR": "Ⓒ"}

	var pdd []string
	for _, l := range []string{"Total", "TLG", "LAT", "DDP", "INS", "CHR"} {
		v := reflect.ValueOf(w).FieldByName(l).Int()
		if v > 0 {
			pdd = append(pdd, m.Sprintf(PDSPA, labels[l], v))
		}
	}

	spans := strings.Join(pdd, " / ")
	html := fmt.Sprintf(PDPAR, s, spans)
	return html
}

// FmtLexParsingData - turn []MorphPossib into HTML
func FmtLexParsingData(mpp []str.MorphPossib) string {
	const (
		OBSERVED = `<span class="obsv"><span class="obsv"> from <span class="baseform"><a class="lex" href="#%s_%s">%s</a></span>
	`
		BFTRANS  = `<span class="baseformtranslation">&nbsp;(“%s”)</span></span></span>`
		MORPHTAB = `
		<table class="morphtable">
			<tbody>
			%s
			</tbody>
		</table>
	`
		MORPHTR = `<tr>%s</tr>`
		MORPHTD = `<td class="%s">%s</td>`
	)
	pat := regexp.MustCompile(`^(.{1,3}\.)\s`)

	mpmap := make(map[string]str.MorphPossib, len(mpp))
	for _, p := range mpp {
		k := p.Headwd + " - " + p.Analysis + " - " + p.Transl
		mpmap[k] = p
	}

	keys := gen.StringMapKeysIntoSlice(mpmap)
	sort.Strings(keys)

	var html string
	usecounter := false
	// on mpp is always empty: why?
	if len(mpp) > 2 {
		usecounter = true
	}
	ct := 0
	memo := ""
	// there are duplicates in the original parsing data
	dedup := make(map[string]bool)
	letter := 0

	for _, k := range keys {
		m := mpmap[k]

		if strings.TrimSpace(m.Analysis) == "" {
			continue
		}

		getlett := func() string {
			if len(mpp) > 2 {
				return fmt.Sprintf("[%s]", string(rune(letter+97)))
			}
			return ""
		}()

		if usecounter && m.Xrefval != memo {
			ct += 1
			html += fmt.Sprintf("(%d)&nbsp;", ct)
		}

		if m.Xrefval != memo {
			html += fmt.Sprintf(OBSERVED, m.Headwd, m.Xrefval, m.Headwd)
			if strings.TrimSpace(m.Transl) != "" {
				m.Transl = PolishTrans(m.Transl, pat)
				html += fmt.Sprintf(BFTRANS, m.Transl)
			}
		}

		dd := m.Headwd + " - " + m.Analysis
		if _, ok := dedup[dd]; !ok {
			pos := strings.Split(m.Analysis, " ")
			var tab string
			tab = fmt.Sprintf(MORPHTD, "morphcell", getlett)
			for _, p := range pos {
				tab += fmt.Sprintf(MORPHTD, "morphcell", p)
			}
			tab = fmt.Sprintf(MORPHTR, tab)
			tab = fmt.Sprintf(MORPHTAB, tab)
			html += tab
			memo = m.Xrefval
			dedup[dd] = true
		} else {
			letter -= 1
		}
		letter += 1
	}

	return html
}

// PolishTrans - add "transtree" spans to the mini-translation lists to highlight structure
func PolishTrans(tr string, pat *regexp.Regexp) string {
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

func formatpreliminfo(w str.DbLexicon) string {
	const (
		TMPL = `<div class="hb-lx-prelim">
	<hb-fs-l-bold>morphology:</hb-fs-l-bold> %s %s
</div><br>
`
	)
	return fmt.Sprintf(TMPL, w.EntryName, w.PrelimInfo)
}

func formatgksenseinfo(s str.LexicalSenses) string {
	const (
		TMPL = `
<hb-lx-sense>
	<hb-lx-sensecounter>(%d)</hb-lx-sensecounter>
	<br>
	<div class="hb-lx-sense-contents">%s</div><br>
</hb-lx-sense>
`
	)

	// ids start at 0
	id := strings.Split(s.ID, ".")[1]
	idv, _ := strconv.Atoi(id)
	idv += 1
	return fmt.Sprintf(TMPL, idv, s.Contents)
}

func formatltsenseinfo(ss []str.LexicalSenses) string {
	const (
		TEMPL2 = " <hb-lx-sensecounter>&nbsp;(%s)&nbsp;</hb-lx-sensecounter>"
	)
	// the latin dictionary senses are not tidy like the greek ones
	// specifically you see a lot of "micro-senses" that are part of a
	// hierarchy; so you can't just enumerate these as per LSJ; they need to be
	// parsed and organized

	// "iungo"
	// I n25350.0 1
	//I n25350.1 1
	//A n25350.2 2
	//1 n25350.3 3
	//2 n25350.4 3
	//3 n25350.5 3
	//4 n25350.6 3
	//B n25350.7 2
	//1 n25350.8 3
	//(a) n25350.9 5
	//(b) n25350.10 5
	//2 n25350.11 3
	//3 n25350.12 3
	//4 n25350.13 3
	//5 n25350.14 3
	//6 n25350.15 3
	//7 n25350.16 3
	//II n25350.17 1
	//A n25350.18 2
	//B n25350.19 2
	//1 n25350.20 3
	//2 n25350.21 3
	//3 n25350.22 3
	//(a) n25350.23 5
	//(b) n25350.24 5

	// 'clipeum'
	// I n8620.0 1
	//B n8620.1 2
	//II n8620.2 1
	//A n8620.3 2
	//B n8620.4 2
	//C n8620.5 2
	//D n8620.6 2
	//E n8620.7 2

	lhh := make([]lexhierchyholder, len(ss))
	prev := lexhierchyholder{lv: 5}

	for i, s := range ss {
		thislh := lexhierchyholder{}
		thislh.AssignHierarchyVals(s)
		lhh[i] = thislh
		prev.ReAssignValsViaNext(&thislh)
		prev = thislh
	}

	// note that the index of lhh matches the index of ss
	lhh = rationalizelhh(lhh)

	var all []string

	previousl1 := "I"
	for i, s := range ss {
		if lhh[i].l1 != previousl1 {
			all = append(all, "<br><br>")
			previousl1 = lhh[i].l1
		}
		// ids start at 0
		//id := strings.Split(s.ID, ".")[1]
		//idv, _ := strconv.Atoi(id)
		//idv += 1
		hs := fmt.Sprintf(TEMPL2, lhh[i].ReturnHierarchy())
		all = append(all, hs+s.Contents)
	}
	return strings.Join(all, "\n")
}

// DISTRIBUTION and PREVALENCE
// 1 word of Trag is equiv to N words of Phil...
// the logic for hinges on the use of mps.ParsedGreekWeightsCorpora, etc. and str.WeightedFieldValuePair
// the high count value is set to "1" and all others are "1/N" to give a multiplier that is applied to the value you see
// so if there is 5 more Phil than Epic, 3 instances of a word in Epic is worth 15 in Phil and so if there are 3 in
// Epic and 10 in Phil, the word will be scored as "more typical of epic" since 15 > 10.
// ParsedGreekWeightsCorpora loads the product of HGB CalculateParsedWordcountTotals(): `__greekwordcounttotals`, etc.
// that gives you the raw totals for every word in every genre, era, and corpus and loads it next to the individual totals

// DbHeadwordCounts: __latinwordcounttotals
//Epic: 1,074,868
//Phil: 4,209,536
//Trag: 201,647
//AllRhet: 3,642,682
//AllRelig: 3,237,452
//Total: 37,421,670

func headwordprevalencebycorp(wc str.DbHeadwordCounts) string {
	// Prevalence (all forms): Ⓖ 95,843 / Ⓛ 10 / Ⓘ 151 / Ⓓ 751 / Ⓒ 64 / Ⓣ 96,819
	const (
		PREVSPAN = `<span class="prevalence rarechars">%s</span>&nbsp;%d`
		PREVSUM  = `<br>Prevalence (all forms): `
	)

	var pd []string
	m := message.NewPrinter(language.English)

	cv := wc.SortedCorpusPairs()

	for _, c := range cv {
		if c.Value > 0 {
			pd = append(pd, m.Sprintf(PREVSPAN, c.Field, c.Value))
		}
	}
	pd = append(pd, m.Sprintf(PREVSPAN, "Ⓣ", wc.Total))

	p := PREVSUM + strings.Join(pd, " / ")

	return p
}

func headworddistribbycorp(wc str.DbHeadwordCounts) string {
	// Weighted distribution by corpus: Ⓖ 100 / Ⓓ 14 / Ⓒ 6 / Ⓘ 2 / Ⓛ 0
	const (
		DIST     = `<br>Distribution by corpus: `
		PREVSPAN = `<span class="rarechars prevalence">%s</span>&nbsp;%d`
	)
	var pd []string
	m := message.NewPrinter(language.English)

	// you will induce a race condition if you do not use clones of the maps
	mymap := maps.Clone(mps.ParsedGreekWeightsCorpora)

	if gen.IsLatin.MatchString(wc.Word) {
		mymap = mps.ParsedLatinWeightsCorpora
	} else {
		// refuse to allow a greek word to look prevalent in LAT
		mymap["LAT"] = 0
	}

	cv := wc.SortedCorpusPairs()
	// [{TLG 6822} {INS 104} {DDP 11} {CHR 8} {LAT 0}]

	wp := make([]str.WeightedFieldValuePair, len(cv))
	for i, v := range cv {
		recalc := mymap[v.Field] * float32(v.Value)
		wp[i].Value = recalc
		wp[i].Field = v.Field
	}

	sort.Slice(wp, func(i, j int) bool {
		return wp[i].Value > wp[j].Value
	})
	// [{TLG 6822} {INS 3076.8318} {CHR 772.97174} {DDP 279.2075} {LAT 0}]

	// now make it "out of 100"
	wmpax := wp[0].Value
	for i := range wp {
		wp[i].Value = (wp[i].Value / wmpax) * 100
	}
	// [{TLG 100} {INS 45.10161} {CHR 11.330574} {DDP 4.092751} {LAT 0}]

	for _, w := range wp {
		if w.Value > 0 {
			pd = append(pd, m.Sprintf(PREVSPAN, w.Field, int(w.Value)))
		}
	}

	out := DIST + strings.Join(pd, " / ")
	return out
}

func headworddistribbyera(wc str.DbHeadwordCounts) string {
	// Weighted chronological distribution: ℯ 100 / ℓ 84 / 𝓂 62
	const (
		DIST     = `<br>Distribution by time: `
		PREVSPAN = `<span class="rarechars prevalence">%s</span>&nbsp;%d`
	)

	// latin eras do not really work (yet?)
	if gen.IsLatin.MatchString(wc.Word) {
		return ""
	}

	var pd []string
	m := message.NewPrinter(language.English)

	mymap := maps.Clone(mps.ParsedGreekWeightsEras)

	//if gen.IsLatin.MatchString(wc.Word) {
	//	mymap = mps.ParsedLatinWeightsEras
	//}

	cv := wc.SortedEraPairs()
	// [{Late 7289} {Middle 2769} {Early 442}]

	wp := make([]str.WeightedFieldValuePair, len(cv))
	for i, v := range cv {
		recalc := mymap[v.Field] * float32(v.Value)
		wp[i].Value = recalc
		wp[i].Field = v.Field
	}

	sort.Slice(wp, func(i, j int) bool {
		return wp[i].Value > wp[j].Value
	})
	// [{Late 7289} {Early 3809.652} {Middle 3808.7402}]

	// now make it "out of 100"
	wmpax := wp[0].Value
	for i := range wp {
		wp[i].Value = (wp[i].Value / wmpax) * 100
	}
	// [{Late 100} {Early 52.26577} {Middle 52.253265}]

	for _, w := range wp {
		if w.Value > 0 {
			pd = append(pd, m.Sprintf(PREVSPAN, w.Field, int(w.Value)))
		}
	}

	return DIST + strings.Join(pd, " / ")
}

func headworddistribbygenre(wc str.DbHeadwordCounts) string {
	// Predominant genres: comm (100), mech (97), jurisprud (93), med (84), mus (75), nathist (61), paroem (60), allrelig (57)
	const (
		DIST          = `<br>Distribution by genre: `
		PREVSPAN      = `<span class="rarechars prevalence">%s</span>&nbsp;%d`
		GENREWTCUTOFF = 150
	)
	var pd []string
	m := message.NewPrinter(language.English)

	mymap := maps.Clone(mps.ParsedGreekWeightsGenres)
	if gen.IsLatin.MatchString(wc.Word) {
		mymap = maps.Clone(mps.ParsedGreekWeightsGenres)
	}

	cv := wc.SortedGenrePairs()
	// [{Hist 3212} {Comm 1155} {Phil 968} {Schol 340} {AllRhet 333} ...]

	wp := make([]str.WeightedFieldValuePair, len(cv))
	for i, v := range cv {
		var wv float32
		// do not let genres with few words to let their one hit count way, way too much
		if mymap[v.Field] > GENREWTCUTOFF {
			wv = 0
		} else {
			wv = mymap[v.Field]
		}

		recalc := wv * float32(v.Value)
		wp[i].Value = recalc
		wp[i].Field = v.Field
	}

	sort.Slice(wp, func(i, j int) bool {
		return wp[i].Value > wp[j].Value
	})
	// [{Tact 12906.822} {Hist 3285.3674} {Perig 3251.8145} {Test 2656.0415} ...]

	// now make it "out of 100"
	wmpax := wp[0].Value
	for i := range wp {
		wp[i].Value = (wp[i].Value / wmpax) * 100
	}

	for i, w := range wp {
		if i >= vv.GENRESTOCOUNT {
			break
		}
		if w.Value > 0 {
			pd = append(pd, m.Sprintf(PREVSPAN, w.Field, int(w.Value)))
		}
	}
	// [{Tact 100} {Hist 25.454504} {Perig 25.19454} {Test 20.578585} ...]

	return DIST + strings.Join(pd, "; ")

}

func rationalizelhh(lhs []lexhierchyholder) []lexhierchyholder {

	decode := map[string]int{
		"I":    1,
		"II":   2,
		"III":  3,
		"IV":   4,
		"V":    5,
		"VI":   6,
		"VII":  7,
		"VIII": 8,
		"IX":   9,
		"X":    10,
		"":     11,
	}

	// l1
	current := lexhierchyholder{l1: "I"}
	for i, lh := range lhs {
		if lh.l1 == "" {
			lh.l1 = current.l1
		} else if decode[lh.l1] > decode[current.l1] {
			current.l1 = lh.l1
			current.l2 = ""
			current.l3 = ""
			current.l4 = ""
			current.l5 = ""
			lh.l2 = ""
			lh.l3 = ""
			lh.l4 = ""
			lh.l5 = ""
		}
		lhs[i] = lh
	}

	// l2
	for i, lh := range lhs {
		tr := []rune{0}
		if len(lh.l2) != 0 {
			tr = []rune(lh.l2)
		}

		cr := []rune{0}
		if len(current.l2) != 0 {
			cr = []rune(current.l2)
		}

		if lh.l2 == "" && lh.l1 == current.l1 {
			lh.l2 = current.l2
		} else if tr[0] > cr[0] {
			current.l1 = lh.l1
			current.l2 = lh.l2
			current.l3 = ""
			current.l4 = ""
			current.l5 = ""
			lh.l3 = ""
			lh.l4 = ""
			lh.l5 = ""
		}
		lhs[i] = lh
	}

	// l3
	for i, lh := range lhs {
		tr := []rune{0}
		if len(lh.l3) != 0 {
			tr = []rune(lh.l3)
		}

		cr := []rune{0}
		if len(current.l3) != 0 {
			cr = []rune(current.l3)
		}

		if lh.l3 == "" && lh.l2 == current.l2 {
			lh.l3 = current.l3
		} else if tr[0] > cr[0] {
			current.l1 = lh.l1
			current.l2 = lh.l2
			current.l3 = lh.l3
			current.l4 = ""
			current.l5 = ""
			lh.l4 = ""
			lh.l5 = ""
		}

		lhs[i] = lh
	}

	// l4
	for i, lh := range lhs {
		tr := []rune{0}
		if len(lh.l4) != 0 {
			tr = []rune(lh.l4)
		}

		cr := []rune{0}
		if len(current.l4) != 0 {
			cr = []rune(current.l4)
		}

		if lh.l4 == "" && lh.l3 == current.l3 {
			lh.l4 = current.l4
		} else if tr[0] > cr[0] {
			current.l1 = lh.l1
			current.l2 = lh.l2
			current.l3 = lh.l3
			current.l4 = lh.l4
			current.l5 = ""
			lh.l5 = ""
		}
		lhs[i] = lh
	}

	// l5 quickfix: a, b, g, d... counters; see "furo"
	for i, lh := range lhs {
		if i > 0 {
			if lh.l5 == "g" && lhs[i-1].l5 == "b" {
				lh.l5 = "c"
				lhs[i] = lh
			}
		}
	}
	return lhs
}

type lexhierchyholder struct {
	l1 string
	l2 string
	l3 string
	l4 string
	l5 string
	lv int
	id string
}

func (lh *lexhierchyholder) AssignHierarchyVals(s str.LexicalSenses) {
	switch s.LVL {
	case "1":
		lh.l1 = s.N
	case "2":
		lh.l2 = s.N
	case "3":
		lh.l3 = s.N
	case "4":
		lh.l4 = s.N
	case "5":
		x := s.N
		x = strings.ReplaceAll(x, "(", "")
		x = strings.ReplaceAll(x, ")", "")
		lh.l5 = x
	default:
		fmt.Println("lexhierchyholder.AssignHierarchyVals() unknown lvl", s.LVL)
	}
	lh.lv, _ = strconv.Atoi(s.LVL)
	lh.id = s.ID
}

func (lh *lexhierchyholder) ReAssignValsViaNext(next *lexhierchyholder) {
	// set I.A by spotting I.B
	if next.lv == 2 && lh.lv == 1 && next.l2 == "B" {
		lh.l2 = "A"
	}
}

func (lh *lexhierchyholder) ReturnHierarchy() string {
	var hs []string
	if lh.l1 != "" {
		hs = append(hs, lh.l1)
	}
	if lh.l2 != "" {
		hs = append(hs, lh.l2)
	}
	if lh.l3 != "" {
		hs = append(hs, lh.l3)
	}
	if lh.l4 != "" {
		hs = append(hs, lh.l4)
	}
	if lh.l5 != "" {
		hs = append(hs, lh.l5)
	}
	return strings.Join(hs, ".")
}

func (lh *lexhierchyholder) PrintOut() {
	const (
		TMPL = `ID: %s
    1: %s
    2: %s
    3: %s
    4: %s
    5: %s
    %s
`
	)
	fmt.Printf(TMPL, lh.id, lh.l1, lh.l2, lh.l3, lh.l4, lh.l5, lh.ReturnHierarchy())
}
