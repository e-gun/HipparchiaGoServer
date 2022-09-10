package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"sync"
)

type WordInfo struct {
	HW         string
	Wd         string
	Loc        string
	Cit        string
	IsHomonymn bool
}

func RtIndexMaker(c echo.Context) error {
	// diverging from the way the python works
	// build not via the selection boxes but via the actual selection made and stored in the session

	// user := readUUIDCookie(c)
	srch := sessionintobulksearch(c)

	// [a] take every word and build WordInfo for it [can parallelize]
	var slicedwords []WordInfo
	for _, r := range srch.Results {
		wds := r.AccentedSlice()
		for _, w := range wds {
			this := WordInfo{
				HW:         "",
				Wd:         uvσςϲ(acuteforgrave(w)),
				Loc:        r.BuildHyperlink(),
				Cit:        r.Citation(),
				IsHomonymn: false,
			}
			slicedwords = append(slicedwords, this)
		}
	}

	// [b] find the unique values
	distinct := make(map[string]bool)
	for _, w := range slicedwords {
		distinct[w.Wd] = true
	}

	// [c] find the headwords for all of these distinct words
	var morphslice []string
	for w := range distinct {
		morphslice = append(morphslice, w)
	}

	// delete after use
	distinct = make(map[string]bool)

	morphmap := arraytogetrequiredmorphobjects(morphslice)

	var slicedlookups []WordInfo
	for _, w := range slicedwords {
		if m, ok := morphmap[w.Wd]; !ok {
			w.HW = "(unparsed)"
			slicedlookups = append(slicedlookups, w)
		} else {
			mps := m.PossibSlice()
			if len(mps) > 1 {
				w.IsHomonymn = true
				for i := 0; i < len(mps); i++ {
					w.HW = mps[i]
					slicedlookups = append(slicedlookups, w)
				}
			}
		}
	}

	// delete after use
	morphmap = make(map[string]DbMorphology)
	slicedwords = []WordInfo{}

	// [d] the final map
	// [d1] build it
	indexmap := make(map[string][]WordInfo)
	for _, w := range slicedlookups {
		indexmap[w.HW] = append(indexmap[w.HW], w)
	}

	// [d2] sort the keys
	var keys []string
	for k, _ := range indexmap {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	// now you have a sorted index...

	//// [d2] sort the []WordInfo entries
	//for k, v := range indexmap {
	//	sort.Slice(indexmap[k], func(i, j int) bool { return v[i].Loc < v[j].Loc })
	//}

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
	}

	var jso JSFeeder
	jso.Au = "(au todo)"
	jso.Ti = "(ti todo)"
	jso.WS = "(ws todo)"
	jso.HT = htm
	jso.EL = "(el todo)"
	jso.WF = len(slicedlookups)
	jso.KY = "(ky todo)"

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
	qt := `SELECT observed_form, xrefs, prefixrefs, related_headwords FROM %s_morphology WHERE EXISTS 
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
			err = foundrows.Scan(&thehit.Observed, &thehit.Xrefs, &thehit.PrefixXrefs, &thehit.RawPossib)
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
	// tph := `<span class="homonym"><indexobserved id="%s">%s</indexobserved></span>`

	//<tr>
	//<td class="headword">&nbsp;</td>
	//<td class="word"><indexobserved id="possedisse">possedisse</indexobserved></td>
	//<td class="count">5</td>
	//<td class="passages"><indexedlocation id="linenumber/lt0474/001/314">36.1</indexedlocation>, <indexedlocation id="linenumber/lt0474/001/318">36.5</indexedlocation>, <indexedlocation id="linenumber/lt0474/001/324">36.11</indexedlocation>, <indexedlocation id="linenumber/lt0474/001/739">73.6</indexedlocation>, <indexedlocation id="linenumber/lt0474/001/954">89.3</indexedlocation></td>
	//</tr>

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

// 	var trr []string
//	for _, k := range keys {
//		ww := indexmap[k]
//		if len(ww) == 1 {
//
//		}
//
//		for _, w := range ww {
//			r := fmt.Sprintf(tr, w.HW, w.HW)
//		}
//		// first the summary
//		ww := indexmap[k]
//		il := fmt.Sprintf(tp, ww[0].HW, ww[0].HW)
//		s := fmt.Sprintf(tr, ww[0].HW, il, len(ww))
//
//	}