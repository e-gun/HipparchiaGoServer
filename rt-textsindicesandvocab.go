package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"net/http"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type WordInfo struct {
	HW         string
	Wd         string
	Loc        string
	Cit        string
	IsHomonymn bool
	Trans      string
}

func RtVocabMaker(c echo.Context) error {
	// diverging from the way the python works
	// build not via the selection boxes but via the actual selection made and stored in the session
	start := time.Now()

	// [a] get all the lines you need and turn them into []WordInfo; Headwords to be filled in later
	srch := sessionintobulksearch(c)

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
		if len(hww) > 1 {
			s.IsHomonymn = true
		}
		for _, h := range hww {
			newwd := s
			newwd.HW = h.Headwd
			newwd.Trans = h.Transl
			parsedwords = append(parsedwords, newwd)
		}
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
		W  string
		C  int
		TR string
	}

	pat := regexp.MustCompile("^(.{1,3}\\.)\\s")
	var polishtrans = func(x string, pat *regexp.Regexp) string {
		x = strings.Replace(x, "<tr>", "", 1)
		x = strings.Replace(x, "</tr>", "", 1)
		elem := strings.Split(x, "; ")
		for i, e := range elem {
			elem[i] = pat.ReplaceAllString(e, `<span class="transtree">$1</span> `)
		}
		return strings.Join(elem, "; ")
	}

	vim := make(map[string]VocInf)
	for k, v := range vic {
		vim[k] = VocInf{
			W:  k,
			C:  v,
			TR: polishtrans(vit[k], pat),
		}
	}

	var vis []VocInf
	for _, v := range vim {
		vis = append(vis, v)
	}

	sort.Slice(vis, func(i, j int) bool { return stripaccentsSTR(vis[i].W) < stripaccentsSTR(vis[j].W) })

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
		nt := fmt.Sprintf(tr, v.W, v.W, v.C, v.TR)
		trr[i+1] = nt
	}

	trr[len(trr)-1] = tf

	thehtml := strings.Join(trr, "")

	type JSFeeder struct {
		Au string `json:"authorname"`
		Ti string `json:"title"`
		ST string `json:"structure"`
		WS string `json:"worksegment"`
		HT string `json:"texthtml"`
		EL string `json:"elapsed"`
		WF int    `json:"wordsfound"`
		KY string `json:"keytoworks"`
		NJ string `json:"newjs"`
	}

	var jso JSFeeder
	jso.Au = AllAuthors[srch.Results[0].FindAuthor()].Cleaname
	if srch.TableSize > 1 {
		jso.Au = jso.Au + fmt.Sprintf(" and %d more author(s)", srch.TableSize-1)
	}

	jso.Ti = AllWorks[srch.Results[0].WkUID].Title
	if srch.SearchSize > 1 {
		jso.Ti = jso.Ti + fmt.Sprintf(" and %d more works(s)", srch.SearchSize-1)
	}

	cf := AllWorks[srch.Results[0].WkUID].CitationFormat()
	var tc []string
	for _, x := range cf {
		if len(x) != 0 {
			tc = append(tc, x)
		}
	}

	jso.ST = strings.Join(tc, ", ")
	jso.HT = thehtml
	jso.EL = fmt.Sprintf("%.2f", time.Now().Sub(start).Seconds())
	jso.WF = srch.SearchSize
	jso.KY = "(TODO)"

	j := fmt.Sprintf(LEXFINDJS, "vocabobserved") + fmt.Sprintf(BROWSERJS, "vocabobserved")
	jso.NJ = fmt.Sprintf("<script>%s</script>", j)

	js, e := json.Marshal(jso)
	chke(e)

	return c.String(http.StatusOK, string(js))
}

func RtIndexMaker(c echo.Context) error {
	// diverging from the way the python works
	// build not via the selection boxes but via the actual selection made and stored in the session
	start := time.Now()
	srch := sessionintobulksearch(c)

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

	boundary := regexp.MustCompile(`(\{|, )"\d": `)
	var slicedlookups []WordInfo
	for _, w := range slicedwords {
		if m, ok := morphmap[w.Wd]; !ok {
			w.HW = "﹙unparsed﹚"
			slicedlookups = append(slicedlookups, w)
		} else {
			mps := extractmorphpossibilities(m.RawPossib, boundary)
			if len(mps) > 1 {
				w.IsHomonymn = true
				for i := 0; i < len(mps); i++ {
					var additionalword WordInfo
					additionalword = w
					additionalword.HW = mps[i].Headwd
					slicedlookups = append(slicedlookups, additionalword)
				}
			}
		}
	}

	var trimslices []WordInfo
	for _, w := range slicedlookups {
		if len(w.HW) != 0 {
			trimslices = append(trimslices, w)
		}
	}

	// [d] the final map
	// [d1] build it
	indexmap := make(map[string][]WordInfo)
	for _, w := range trimslices {
		indexmap[w.HW] = append(indexmap[w.HW], w)
	}

	// [d2] sort the keys
	var keys []string
	for k, _ := range indexmap {
		keys = append(keys, k)
	}

	// sort can't do polytonic greek
	sort.Slice(keys, func(i, j int) bool { return stripaccentsSTR(keys[i]) < stripaccentsSTR(keys[j]) })

	// now you have a sorted index...

	var trr []string
	for _, k := range keys {
		wii := indexmap[k]
		trr = append(trr, convertwordinfototablerow(wii))
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

	type JSFeeder struct {
		Au string `json:"authorname"`
		Ti string `json:"title"`
		WS string `json:"worksegment"`
		HT string `json:"indexhtml"`
		EL string `json:"elapsed"`
		WF int    `json:"wordsfound"`
		KY string `json:"keytoworks"`
		ST string `json:"structure"`
		NJ string `json:"newjs"`
	}

	var jso JSFeeder
	jso.Au = AllAuthors[srch.Results[0].FindAuthor()].Cleaname
	if srch.TableSize > 1 {
		jso.Au = jso.Au + fmt.Sprintf(" and %d more author(s)", srch.TableSize-1)
	}

	jso.Ti = AllWorks[srch.Results[0].WkUID].Title
	if srch.SearchSize > 1 {
		jso.Ti = jso.Ti + fmt.Sprintf(" and %d more works(s)", srch.SearchSize-1)
	}

	if srch.SearchSize == 1 && srch.TableSize == 1 {
		jso.KY = ""
	} else {
		// todo: build the key to the works...
	}

	if len(searches[readUUIDCookie(c)].SearchIn.ListedPBN) == 0 {
		jso.WS = ""
	} else {
		jso.WS = strings.Join(searches[readUUIDCookie(c)].SearchIn.ListedPBN, "; ")
	}

	cf := AllWorks[srch.Results[0].WkUID].CitationFormat()
	var tc []string
	for _, x := range cf {
		if len(x) != 0 {
			tc = append(tc, x)
		}
	}

	jso.ST = strings.Join(tc, ", ")
	jso.HT = htm
	jso.EL = fmt.Sprintf("%.2f", time.Now().Sub(start).Seconds())
	jso.WF = len(trimslices)

	j := fmt.Sprintf(LEXFINDJS, "indexobserved") + fmt.Sprintf(BROWSERJS, "indexobserved")
	jso.NJ = fmt.Sprintf("<script>%s</script>", j)

	js, e := json.Marshal(jso)
	chke(e)

	return c.String(http.StatusOK, string(js))
}

func RtTextMaker(c echo.Context) error {
	// diverging from the way the python works
	// build not via the selection boxes but via the actual selection made and stored in the session

	// this has the downside of allowing for insanely large text generation
	// but, on the other hand, this now works like a simple search

	// then it gets output as a big browser table...

	user := readUUIDCookie(c)
	srch := sessionintobulksearch(c)
	searches[srch.ID] = srch

	// now we have the lines we need....
	firstline := searches[srch.ID].Results[0]
	firstwork := AllWorks[firstline.WkUID]
	firstauth := AllAuthors[firstwork.FindAuthor()]

	// ci := formatcitationinfo(firstwork, firstline)
	tr := buildbrowsertable(-1, searches[srch.ID].Results)

	type JSFeeder struct {
		Au string `json:"authorname"`
		Ti string `json:"title"`
		St string `json:"structure"`
		WS string `json:"worksegment"`
		HT string `json:"texthtml"`
	}

	sui := sessions[user].Inclusions
	var jso JSFeeder

	jso.Au = firstauth.Shortname

	if len(sui.Authors) > 1 || len(sui.AuGenres) > 0 || len(sui.AuLocations) > 0 {
		jso.Au += " (and others)"
	}

	jso.Ti = firstwork.Title
	if len(sui.Works) > 1 || len(sui.WkGenres) > 0 || len(sui.WkLocations) > 0 {
		jso.Ti += " (and others)"
	}

	jso.St = basiccitation(firstwork, firstline)
	jso.WS = "" // unused for now
	jso.HT = tr

	js, e := json.Marshal(jso)
	chke(e)

	return c.String(http.StatusOK, string(js))
}

func sessionintobulksearch(c echo.Context) SearchStruct {
	user := readUUIDCookie(c)

	srch := builddefaultsearch(c)
	srch.Seeking = ""
	srch.Limit = MAXTEXTLINEGENERATION
	srch.InitSum = "(gathering and formatting line of text)"
	srch.ID = strings.Replace(uuid.New().String(), "-", "", -1)

	parsesearchinput(&srch)
	sl := sessionintosearchlist(sessions[user])
	srch.SearchIn = sl.Inc
	srch.SearchEx = sl.Excl
	srch.SearchSize = sl.Size
	prq := searchlistintoqueries(&srch)
	srch.Queries = prq
	srch.IsActive = true
	searches[srch.ID] = srch
	searches[srch.ID] = HGoSrch(searches[srch.ID])
	srch.TableSize = len(prq)

	return searches[srch.ID]
}

func arraytogetrequiredmorphobjects(wordlist []string) map[string]DbMorphology {
	// NB: this goroutine version not in fact that much faster with Cicero than doing it without goroutines as one giant array
	// but the implementation pattern is likely useful for some place where it will make a difference

	// look for the upper case matches too: Ϲωκράτηϲ and not just ϲωκρατέω (!)
	var uppers []string
	for i := 0; i < len(wordlist); i++ {
		uppers = append(uppers, strings.Title(wordlist[i]))
	}

	wordlist = append(wordlist, uppers...)
	// note that we are hereby going to feed some of the workers huge lists of capitalized words that will return few hits
	workers := runtime.NumCPU()

	totalwork := len(wordlist)
	chunksize := totalwork / workers
	leftover := totalwork % workers
	wordmap := make(map[int][]string, workers)

	if totalwork <= workers {
		wordmap[0] = wordlist
	} else {
		thestart := 0
		for i := 0; i < workers; i++ {
			wordmap[i] = wordlist[thestart : thestart+chunksize]
			thestart = thestart + chunksize
		}

		// leave no sentence behind!
		if leftover > 0 {
			wordmap[workers-1] = append(wordmap[workers-1], wordlist[totalwork-leftover-1:totalwork-1]...)
		}
	}

	// https://stackoverflow.com/questions/46010836/using-goroutines-to-process-values-and-gather-results-into-a-slice
	// see the comments of Paul Hankin re. building an anonymous function

	var wg sync.WaitGroup
	var collector []map[string]DbMorphology
	outputchannels := make(chan map[string]DbMorphology, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		// "i" will be captured if sent into the function
		j := i
		go func(wordlist []string, workerid int) {
			defer wg.Done()
			dbp := GetPSQLconnection()
			defer dbp.Close()
			outputchannels <- morphologyworker(wordmap[j], j, dbp)
		}(wordmap[i], i)
	}

	go func() {
		wg.Wait()
		close(outputchannels)
	}()

	// merge the results
	for c := range outputchannels {
		collector = append(collector, c)
	}

	// map the results
	foundmorph := make(map[string]DbMorphology)
	for _, mmap := range collector {
		for w := range mmap {
			foundmorph[w] = mmap[w]
		}
	}

	return foundmorph
}

func morphologyworker(wordlist []string, workerid int, dbpool *pgxpool.Pool) map[string]DbMorphology {
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

	tt := `CREATE TEMPORARY TABLE ttw_%s AS SELECT words AS w FROM unnest(ARRAY[%s]) words`
	qt := `SELECT observed_form, xrefs, prefixrefs, possible_dictionary_forms, related_headwords FROM %s_morphology WHERE EXISTS 
		(SELECT 1 FROM ttw_%s temptable WHERE temptable.w = %s_morphology.observed_form)`

	foundmorph := make(map[string]DbMorphology)

	// a waste of time to check the language on every word; just flail/fail once
	for _, uselang := range []string{"greek", "latin"} {
		u := strings.Replace(uuid.New().String(), "-", "", -1)
		id := fmt.Sprintf("%s_%s_mw_%d", u, uselang, workerid)
		a := fmt.Sprintf("'%s'", strings.Join(wordlist, "', '"))
		t := fmt.Sprintf(tt, id, a)

		_, err := dbpool.Exec(context.Background(), t)
		chke(err)

		foundrows, e := dbpool.Query(context.Background(), fmt.Sprintf(qt, uselang, id, uselang))
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

func convertwordinfototablerow(ww []WordInfo) string {
	// every word has the same headword
	// now we build a sub-map after the pattern of the main map: but now the keys are the words, not the headwords

	// build it
	indexmap := make(map[string][]WordInfo)
	for _, w := range ww {
		indexmap[w.Wd] = append(indexmap[w.Wd], w)
	}

	// sort the keys
	var keys []string
	for k, _ := range indexmap {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	tr := `
	<tr>
	<td class="headword">%s</td>
	<td class="word"><indexobserved id="%s">%s</indexobserved></td>
	<td class="count">%d</td>
	<td class="passages">%s</td>
	</tr>`

	tp := `<indexedlocation id="%s">%s</indexedlocation>`

	var trr []string
	used := make(map[string]bool)
	for _, k := range keys {
		wii := indexmap[k]
		hw := ""
		if used[wii[0].HW] {
			// skip
		} else {
			hw = wii[0].HW
		}

		// todo: not working...
		tem := tp
		//if wii[0].IsHomonymn {
		//	tem = tph
		//}

		// get all passages related to this word
		var pp []string
		sort.Slice(wii, func(i, j int) bool { return wii[i].Loc < wii[j].Loc })
		for i := 0; i < len(wii); i++ {
			pp = append(pp, fmt.Sprintf(tem, wii[i].Loc, wii[i].Cit))
		}
		p := strings.Join(pp, ", ")
		t := fmt.Sprintf(tr, hw, wii[0].Wd, wii[0].Wd, len(wii), p)
		trr = append(trr, t)
		used[wii[0].HW] = true
	}

	out := strings.Join(trr, "")
	return out
}
