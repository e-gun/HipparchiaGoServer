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
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

type WordInfo struct {
	HW         string
	Wd         string
	Loc        string
	Cit        string
	IsHomonymn bool
	Trans      string
	Wk         string
}

//
// ROUTES
//

// RtTextMaker - make a text of whatever collection of lines you would be searching
func RtTextMaker(c echo.Context) error {
	c.Response().After(func() { gcstats("RtTextMaker()") })
	// diverging from the way the python works
	// build not via the selection boxes but via the actual selection made and stored in the session

	// this has the downside of allowing for insanely large text generation
	// but, on the other hand, this now works like a simple search

	// then it gets output as a big "browser table"...

	user := readUUIDCookie(c)
	srch := sessionintobulksearch(c)

	if len(srch.Results) == 0 {
		return emptyjsreturn(c)
	}

	searches[srch.ID] = srch

	// now we have the lines we need....
	firstline := searches[srch.ID].Results[0]
	firstwork := AllWorks[firstline.WkUID]
	firstauth := AllAuthors[firstwork.FindAuthor()]

	tr := `
            <tr class="browser">
                <td class="browserembeddedannotations">%s</td>
                <td class="browsedline">%s</td>
                <td class="browsercite">%s</td>
            </tr>
		`

	lines := searches[srch.ID].Results
	block := make([]string, len(lines))
	for i, l := range lines {
		l.PurgeMetadata()
		block[i] = l.MarkedUp
	}

	whole := strings.Join(block, "✃✃✃")

	whole = textblockcleaner(whole)

	// reassemble
	block = strings.Split(whole, "✃✃✃")
	for i, b := range block {
		lines[i].MarkedUp = b
	}

	trr := make([]string, len(lines))
	previous := lines[0]
	workcount := 1
	for i, l := range lines {
		cit := selectivelydisplaycitations(lines[i], previous, -1)
		trr[i] = fmt.Sprintf(tr, lines[i].Annotations, lines[i].MarkedUp, cit)
		if l.WkUID != previous.WkUID {
			// you were doing multi-text generation
			workcount += 1
			aw := AllAuthors[AllWorks[l.WkUID].FindAuthor()].Name + fmt.Sprintf(`, <span class="italic">%s</span>`, AllWorks[l.WkUID].Title)
			aw = fmt.Sprintf(`<hr><span class="emph">[%d] %s</span>`, workcount, aw)
			extra := fmt.Sprintf(tr, "", aw, "")
			trr[i] = extra + trr[i]
		}
		previous = lines[i]
	}
	tab := strings.Join(trr, "")
	// that was the body, now do the head and tail
	top := fmt.Sprintf(`<div id="browsertableuid" uid="%s"></div>`, lines[0].FindAuthor())
	top += `<table><tbody>`
	top += `<tr class="spacing">` + strings.Repeat("&nbsp;", MINBROWSERWIDTH) + `</tr>`

	tab = top + tab + `</tbody></table>`

	// but we don't want/need "observed" tags

	// <div id="searchsummary">Cicero,&nbsp;<span class="foundwork">Philippicae</span><br><br>citation format:&nbsp;oration 3, section 13, line 1<br></div>
	st := `
	<div id="searchsummary">%s,&nbsp;<span class="foundwork">%s</span><br>
	citation format:&nbsp;%s<br></div>`

	sui := sessions[user].Inclusions

	au := firstauth.Shortname
	if len(sui.Authors) > 1 || len(sui.AuGenres) > 0 || len(sui.AuLocations) > 0 {
		au += " (and others)"
	}

	ti := firstwork.Title
	if len(sui.Works) > 1 || len(sui.WkGenres) > 0 || len(sui.WkLocations) > 0 {
		ti += " (and others)"
	}

	ct := basiccitation(firstwork, firstline)

	sum := fmt.Sprintf(st, au, ti, ct)

	cp := ""
	if len(srch.Results) == MAXTEXTLINEGENERATION {
		m := message.NewPrinter(language.English)
		cp = m.Sprintf(`<span class="small"><span class="red emph">text generation incomplete:</span> hit the cap of %d on allowed lines</span>`, MAXTEXTLINEGENERATION)
	}
	sum = sum + cp

	type JSFeeder struct {
		SU string `json:"searchsummary"`
		HT string `json:"thehtml"`
		JS string `json:"newjs"`
	}

	var jso JSFeeder
	jso.SU = sum
	jso.HT = tab
	jso.JS = ""

	return c.JSONPretty(http.StatusOK, jso, JSONINDENT)
}

// RtVocabMaker - get the vocabulary for whatever collection of lines you would be searching
func RtVocabMaker(c echo.Context) error {
	c.Response().After(func() { gcstats("RtVocabMaker()") })

	// diverging from the way the python works
	// build not via the selection boxes but via the actual selection made and stored in the session
	// todo: worry about γ' for γε
	start := time.Now()

	id := c.Param("id")
	id = purgechars(cfg.BadChars, id)

	// for progress reporting
	si := builddefaultsearch(c)
	si.ID = id
	si.InitSum = "Grabbing the lines... (part 1 of 4)"
	si.IsActive = true
	searches[si.ID] = si
	progremain.Store(si.ID, 1)

	// [a] get all the lines you need and turn them into []WordInfo; Headwords to be filled in later
	srch := sessionintobulksearch(c)

	if len(srch.Results) == 0 {
		return emptyjsreturn(c)
	}

	var slicedwords []WordInfo
	for _, r := range srch.Results {
		wds := r.AccentedSlice()
		for _, w := range wds {
			this := WordInfo{
				HW:         "",
				Wd:         uvσςϲ(swapacuteforgrave(w)),
				Loc:        r.BuildHyperlink(),
				Cit:        r.Citation(),
				IsHomonymn: false,
				Wk:         r.WkUID,
			}
			slicedwords = append(slicedwords, this)
		}
	}

	// [b] find the unique values we are working with
	distinct := make(map[string]bool, len(slicedwords))
	for _, w := range slicedwords {
		distinct[w.Wd] = true
	}

	// [c] prepare to find the headwords for all of these distinct words
	morphslice := make([]string, len(distinct))
	count := 0
	for w := range distinct {
		morphslice[count] = w
		count += 1
	}

	// [c1] get and map all the DbMorphology
	morphmap := arraytogetrequiredmorphobjects(morphslice)

	si.InitSum = "Parsing the vocabulary...(part 2 of 4)"
	searches[si.ID] = si

	boundary := regexp.MustCompile(`(\{|, )"\d": `)
	// [c2] map observed words to possibilities
	poss := make(map[string][]MorphPossib)
	for k, v := range morphmap {
		poss[k] = extractmorphpossibilities(v.RawPossib, boundary)
	}

	// [c3] build a new slice of seen words with headwords attached
	var parsedwords []WordInfo
	for _, s := range slicedwords {
		hww := poss[s.Wd]
		for _, h := range hww {
			newwd := s
			newwd.HW = h.Headwd
			newwd.Trans = h.Transl
			parsedwords = append(parsedwords, newwd)
		}
	}

	mp := make(map[string]rune)
	if srch.SearchSize > 1 {
		parsedwords, mp = addkeystowordinfo(parsedwords)
	}

	// [d] get the counts
	vic := make(map[string]int)
	for _, p := range parsedwords {
		vic[p.HW]++
	}

	// [e] get the translations
	vit := make(map[string]string)
	for _, p := range parsedwords {
		vit[p.HW] = p.Trans
	}

	// [f] consolidate the information

	type VocInf struct {
		Wd    string
		C     int
		TR    string
		Strip string
	}

	pat := regexp.MustCompile("^(.{1,3}\\.)\\s")

	vim := make(map[string]VocInf)
	for k, v := range vic {
		vim[k] = VocInf{
			Wd:    k,
			C:     v,
			TR:    polishtrans(vit[k], pat),
			Strip: strings.Replace(stripaccentsSTR(k), "ϲ", "σ", -1),
		}
	}

	vis := make([]VocInf, len(vim))
	ct := 0
	for _, v := range vim {
		vis[ct] = v
		ct += 1
	}

	si.InitSum = "Sifting the vocabulary...(part 3 of 4)"
	searches[si.ID] = si

	sort.Slice(vis, func(i, j int) bool { return vis[i].Strip < vis[j].Strip })

	si.InitSum = "Building the HTML...(part 4 of 4)"
	searches[si.ID] = si

	// [g] format

	th := `
	<table>
	<tr>
			<th class="vocabtable">word</th>
			<th class="vocabtable">count</th>
			<th class="vocabtable">definitions</th>
	</tr>`

	tr := `
		<tr>
			<td class="word"><vocabobserved id="%s">%s</vocabobserved></td>
			<td class="count">%d</td>
			<td class="trans">%s</td>
		</tr>`

	tf := `</table>`

	// preallocation means assign to index vs append
	trr := make([]string, len(vis)+2)
	trr[0] = th

	for i, v := range vis {
		nt := fmt.Sprintf(tr, v.Wd, v.Wd, v.C, v.TR)
		trr[i+1] = nt
	}

	trr[len(trr)-1] = tf

	htm := strings.Join(trr, "")

	st := `
	<div id="searchsummary">Vocabulary for %s,&nbsp;<span class="foundwork">%s</span><br>
		citation format:&nbsp;%s<br>
		%s words found<br>
		<span class="small">(%ss)</span><br>
		%s
		%s
	</div>
	`

	an := AllAuthors[srch.Results[0].FindAuthor()].Cleaname
	if srch.TableSize > 1 {
		an = an + fmt.Sprintf(" and %d more author(s)", srch.TableSize-1)
	}

	wn := AllWorks[srch.Results[0].WkUID].Title
	if srch.SearchSize > 1 {
		wn = wn + fmt.Sprintf(" and %d more works(s)", srch.SearchSize-1)
	}

	cf := AllWorks[srch.Results[0].WkUID].CitationFormat()
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

	ky := multiworkkeymaker(mp, &srch)

	cp := ""
	if len(srch.Results) == MAXTEXTLINEGENERATION {
		cp = m.Sprintf(`<span class="small"><span class="red emph">vocabulary generation incomplete:</span>: hit the cap of %d on allowed lines</span>`, MAXTEXTLINEGENERATION)
	}

	sum := fmt.Sprintf(st, an, wn, cit, wf, el, cp, ky)

	type JSFeeder struct {
		SU string `json:"searchsummary"`
		HT string `json:"thehtml"`
		NJ string `json:"newjs"`
	}

	var jso JSFeeder
	jso.SU = sum
	jso.HT = htm

	j := fmt.Sprintf(LEXFINDJS, "vocabobserved") + fmt.Sprintf(BROWSERJS, "vocabobserved")
	jso.NJ = fmt.Sprintf("<script>%s</script>", j)

	// clean up progress reporting
	delete(searches, si.ID)
	progremain.Delete(si.ID)

	return c.JSONPretty(http.StatusOK, jso, JSONINDENT)
}

// RtIndexMaker - build an index for whatever collection of lines you would be searching
func RtIndexMaker(c echo.Context) error {
	c.Response().After(func() { gcstats("RtIndexMaker()") })

	// diverging from the way the python works
	// build not via the selection boxes but via the actual selection made and stored in the session

	// a lot of code duplication with RtVocabMaker() but consolidation is not as direct a matter as one might guess

	start := time.Now()

	id := c.Param("id")
	id = purgechars(cfg.BadChars, id)

	// for progress reporting
	si := builddefaultsearch(c)
	si.ID = id
	si.InitSum = "Grabbing the lines...&nbsp;(part 1 of 4)"
	si.IsActive = true
	searches[si.ID] = si
	progremain.Store(si.ID, 1)

	srch := sessionintobulksearch(c)

	if len(srch.Results) == 0 {
		return emptyjsreturn(c)
	}

	var slicedwords []WordInfo
	for _, r := range srch.Results {
		wds := r.AccentedSlice()
		for _, w := range wds {
			this := WordInfo{
				HW:         "",
				Wd:         uvσςϲ(swapacuteforgrave(w)),
				Loc:        r.BuildHyperlink(),
				Cit:        r.Citation(),
				IsHomonymn: false,
				Wk:         r.WkUID,
			}
			slicedwords = append(slicedwords, this)
		}
	}

	// [b] find the unique values
	distinct := make(map[string]bool, len(slicedwords))
	for _, w := range slicedwords {
		distinct[w.Wd] = true
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

	si.InitSum = "Parsing the vocabulary...&nbsp;(part 2 of 4)"
	searches[si.ID] = si

	boundary := regexp.MustCompile(`(\{|, )"\d": `)
	var slicedlookups []WordInfo
	for _, w := range slicedwords {
		emm := false
		mme := w.Wd
		if _, ok := morphmap[w.Wd]; !ok {
			// here is where you check to see if the word + an apostrophe can be found: γ is really γ' (i.e. γε)
			// this also means that you had to grab all of those extra forms in the first plac
			if _, y := morphmap[w.Wd+"'"]; y {
				emm = true
				w.Wd = w.Wd + "'"
				mme = w.Wd
			} else {
				w.HW = "ϙϙϙϙϙϙϙϙ<br>unparsed words"
				slicedlookups = append(slicedlookups, w)
			}
		} else {
			emm = true
		}

		if emm {
			mps := extractmorphpossibilities(morphmap[mme].RawPossib, boundary)
			if len(mps) > 1 {
				for i := 0; i < len(mps); i++ {
					var additionalword WordInfo
					additionalword = w
					additionalword.HW = mps[i].Headwd
					// additionalword.Stripped = stripaccentsSTR(additionalword.HW)
					slicedlookups = append(slicedlookups, additionalword)
				}
			}
		}
	}

	slicedwords = []WordInfo{} // drop after use

	var trimslices []WordInfo
	for _, w := range slicedlookups {
		if len(w.HW) != 0 {
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
		if _, ok := htest[t.Wd]; !ok {
			htest[t.Wd] = t.HW
		} else {
			if htest[t.Wd] != t.HW {
				ishom[t.Wd] = true
			}
		}
	}

	for i, t := range trimslices {
		if ishom[t.Wd] {
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

	type SorterStruct struct {
		sorter string
		value  string
		count  int
	}

	si.InitSum = "Sifting the index...&nbsp;(part 3 of 4)"
	searches[si.ID] = si

	indexmap := make(map[SorterStruct][]WordInfo, len(trimslices))
	for _, w := range trimslices {
		// lunate sigma sorts after omega
		sigma := strings.Replace(stripaccentsSTR(w.HW), "ϲ", "σ", -1)
		ss := SorterStruct{
			sorter: sigma + w.HW,
			value:  w.HW,
		}
		indexmap[ss] = append(indexmap[ss], w)
	}

	m := message.NewPrinter(language.English)
	wf := m.Sprintf("%d", len(trimslices))
	trimslices = []WordInfo{} // drop after use

	// [d2] sort the keys

	keys := make([]SorterStruct, len(indexmap))
	counter := 0
	for k, v := range indexmap {
		k.count = len(v)
		keys[counter] = k
		counter += 1
	}

	// sort can't do polytonic greek: so there is a lot of (slow) extra stuff that has to happen
	sort.Slice(keys, func(i, j int) bool { return keys[i].sorter < keys[j].sorter })

	// now you have a sorted index...; but a SorterStruct does not make for a usable map key...
	plainkeys := make([]string, len(keys))
	for i, k := range keys {
		plainkeys[i] = k.value
	}

	plainmap := make(map[string][]WordInfo, len(indexmap))
	for k, _ := range indexmap {
		plainmap[k.value] = indexmap[k]
	}

	indexmap = make(map[SorterStruct][]WordInfo, 1) // drop after use

	si.InitSum = "Building the HTML...&nbsp;(part 4 of 4)"
	searches[si.ID] = si

	trr := make([]string, len(plainkeys))
	for i, k := range plainkeys {
		trr[i] = convertwordinfototablerow(plainmap[k])
	}

	tb := `        
		<table>
        <tbody><tr>
            <th class="indextable">headword</th>
            <th class="indextable">word</th>
            <th class="indextable">count</th>
            <th class="indextable">passages</th>
        </tr>
		%s
		</table>`

	htm := fmt.Sprintf(tb, strings.Join(trr, ""))

	// <div id="searchsummary">Index to Cicero - Cicero, Marcus Tullius,&nbsp;<span class="foundwork">Philippicae</span>
	// <br>citation format:&nbsp;oration, section, line<br>236 words found<br><span class="small">(0.10s)</span><br></div>
	st := `
	<div id="searchsummary">Index to %s,&nbsp;<span class="foundwork">%s</span><br>
		citation format:&nbsp;%s<br>
		%s words found<br>
		<span class="small">(%ss)</span><br>
		%s
		%s
		<br>
		(NB: <span class="homonym">homonymns</span> will appear under more than one headword)
	</div>
	`

	an := AllAuthors[srch.Results[0].FindAuthor()].Cleaname
	if srch.TableSize > 1 {
		an = an + fmt.Sprintf(" and %d more author(s)", srch.TableSize-1)
	}

	wn := AllWorks[srch.Results[0].WkUID].Title
	if srch.SearchSize > 1 {
		wn = wn + fmt.Sprintf(" and %d more works(s)", srch.SearchSize-1)
	}

	cf := AllWorks[srch.Results[0].WkUID].CitationFormat()
	var tc []string
	for _, x := range cf {
		if len(x) != 0 {
			tc = append(tc, x)
		}
	}

	cit := strings.Join(tc, ", ")

	// wf := m.Sprintf("%d", len(trimslices))

	el := fmt.Sprintf("%.2f", time.Now().Sub(start).Seconds())

	ky := multiworkkeymaker(mp, &srch)

	cp := ""
	if len(srch.Results) == MAXTEXTLINEGENERATION {
		cp = m.Sprintf(`<span class="small"><span class="red emph">indexing incomplete:</span>: hit the cap of %d on allowed lines</span>`, MAXTEXTLINEGENERATION)
	}

	sum := fmt.Sprintf(st, an, wn, cit, wf, el, cp, ky)

	type JSFeeder struct {
		SU string `json:"searchsummary"`
		HT string `json:"thehtml"`
		NJ string `json:"newjs"`
	}

	var jso JSFeeder
	jso.SU = sum
	jso.HT = htm

	j := fmt.Sprintf(LEXFINDJS, "indexobserved") + fmt.Sprintf(BROWSERJS, "indexedlocation")
	jso.NJ = fmt.Sprintf("<script>%s</script>", j)

	// clean up progress reporting
	delete(searches, si.ID)
	progremain.Delete(si.ID)

	return c.JSONPretty(http.StatusOK, jso, JSONINDENT)
}

//
// HELPERS
//

// sessionintobulksearch - grab every line of text in the currently selected set of authors, works, and passages
func sessionintobulksearch(c echo.Context) SearchStruct {
	user := readUUIDCookie(c)

	srch := builddefaultsearch(c)
	srch.Seeking = ""
	srch.Limit = MAXTEXTLINEGENERATION
	srch.InitSum = "(gathering and formatting lines of text)"
	srch.ID = strings.Replace(uuid.New().String(), "-", "", -1)

	srch.CleanInput()

	sl := SessionIntoSearchlist(sessions[user])
	srch.SearchIn = sl.Inc
	srch.SearchEx = sl.Excl
	srch.SearchSize = sl.Size
	SSBuildQueries(&srch)
	srch.IsActive = true
	srch.TableSize = len(srch.Queries)
	srch = HGoSrch(srch)
	return srch
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

	dbconn := GetPSQLconnection()
	defer dbconn.Release()

	tt := `CREATE TEMPORARY TABLE ttw_%s AS SELECT words AS w FROM unnest(ARRAY[%s]) words`
	qt := `SELECT observed_form, xrefs, prefixrefs, possible_dictionary_forms, related_headwords FROM %s_morphology WHERE EXISTS 
		(SELECT 1 FROM ttw_%s temptable WHERE temptable.w = %s_morphology.observed_form)`

	foundmorph := make(map[string]DbMorphology)

	// a waste of time to check the language on every word; just flail/fail once
	for _, uselang := range []string{"greek", "latin"} {
		u := strings.Replace(uuid.New().String(), "-", "", -1)
		id := fmt.Sprintf("%s_%s_mw", u, uselang)
		a := fmt.Sprintf("'%s'", strings.Join(wordlist, "', '"))
		t := fmt.Sprintf(tt, id, a)

		_, err := dbconn.Exec(context.Background(), t)
		chke(err)

		foundrows, e := dbconn.Query(context.Background(), fmt.Sprintf(qt, uselang, id, uselang))
		chke(e)

		defer foundrows.Close()
		count := 0
		for foundrows.Next() {
			count += 1
			var thehit DbMorphology
			err = foundrows.Scan(&thehit.Observed, &thehit.Xrefs, &thehit.PrefixXrefs, &thehit.RawPossib, &thehit.RelatedHW)
			chke(err)
			foundmorph[thehit.Observed] = thehit
		}
	}
	return foundmorph
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
	uu = unique(uu)
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

	ky := ""
	wkk := srch.SearchSize > 1
	auu := srch.TableSize > 1

	if auu || wkk {
		var out []string
		for k, v := range mapper {
			t := fmt.Sprintf(`<span class="italic">%s</span>`, AllWorks[k].Title)
			if auu {
				t = AllAuthors[AllWorks[k].FindAuthor()].Name + ", " + t
			}
			out = append(out, fmt.Sprintf("%s: %s\n", string(v), t))
		}
		sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
		ky = strings.Join(out, "; ")
		ky = `<br><span class="emph">Works:</span> ` + ky
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

	// build it
	indexmap := make(map[string][]WordInfo, len(ww))
	for _, w := range ww {
		indexmap[w.Wd] = append(indexmap[w.Wd], w)
	}

	// sort the keys
	keys := make([]string, len(indexmap))
	count := 0
	for k, _ := range indexmap {
		keys[count] = k
		count += 1
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	// standard TR
	tr := `
	<tr>
		<td class="headword"><indexobserved id="%s">%s</indexobserved></td>
		<td class="word"><indexobserved id="%s">%s</indexobserved></td>
		<td class="count">%d</td>
		<td class="passages">%s</td>
	</tr>`

	// homonymn TR
	trh := `
	<tr>
		<td class="headword"><indexobserved id="%s">%s</indexobserved></td>
		<td class="word"><span class="homonym"><indexobserved id="%s">%s</indexobserved></span></td>
		<td class="count">%d</td>
		<td class="passages">%s</td>
	</tr>`

	tp := `<indexedlocation id="%s">%s</indexedlocation>`

	trr := make([]string, len(keys))
	used := make(map[string]bool)
	for i, k := range keys {
		wii := indexmap[k]
		hw := ""
		if used[wii[0].HW] {
			// skip
		} else {
			hw = wii[0].HW
		}

		tem := tp

		sort.Slice(wii, func(i, j int) bool { return wii[i].Loc < wii[j].Loc })

		// get all passages related to this word
		var pp []string
		dedup := make(map[string]bool) // this is hacky: why are their duplicates to begin with?
		for j := 0; j < len(wii); j++ {
			if _, ok := dedup[wii[j].Loc]; !ok {
				pp = append(pp, fmt.Sprintf(tem, wii[j].Loc, wii[j].Cit))
				dedup[wii[j].Loc] = true
			}
		}
		p := strings.Join(pp, ", ")

		templ := tr
		if wii[0].IsHomonymn {
			templ = trh
		}

		t := fmt.Sprintf(templ, hw, hw, wii[0].Wd, wii[0].Wd, len(pp), p)
		trr[i] = t
		used[wii[0].HW] = true
	}

	out := strings.Join(trr, "")
	return out
}

// polishtrans - add "transtree" spans to the mini-translation lists to highlight structure
func polishtrans(x string, pat *regexp.Regexp) string {
	// don't loop "pat". it's not really a variable. here it is:
	// pat := regexp.MustCompile("^(.{1,3}\\.)\\s")

	// sample:
	// <span class="transtree">A.</span> as Adv., bearing the same relation to ὡϲ as ὅϲτε to ὅϲ, and used by Hom.
	// more freq. than ὡϲ in similes, when it is commonly written divisim, and is relat. to a demonstr. ὥϲ: sts. c. pres. Indic;
	// <span class="transtree">B.</span> the actual

	x = nohtml.ReplaceAllString(x, "")
	elem := strings.Split(x, "; ")
	for i, e := range elem {
		elem[i] = pat.ReplaceAllString(e, `<span class="transtree">$1</span> `)
	}
	return strings.Join(elem, "; ")
}
