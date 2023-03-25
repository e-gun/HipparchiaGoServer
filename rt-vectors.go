//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/ynqa/wego/pkg/embedding"
	"github.com/ynqa/wego/pkg/model/modelutil/vector"
	"github.com/ynqa/wego/pkg/model/word2vec"
	"github.com/ynqa/wego/pkg/search"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
)

// VectorSearch - a special case for RtSearch() where you requested vectorization of the results
func VectorSearch(c echo.Context, srch SearchStruct) error {
	const (
		FAIL1 = `err: search.New(embs...)`
		FAIL2 = `err: searcher.SearchInternal(word, rank)`
	)

	fp := fingerprintvectorsearch(srch)

	isstored := vectordbcheck(fp)

	var embs embedding.Embeddings

	if isstored {
		embs = vectordbfetch(fp)
	} else {
		embs = generateembeddings(c, srch)
		vectordbadd(fp, embs)
	}

	fail := func(f string) error {
		soj := SearchOutputJSON{
			Title:         "VECTORS",
			Searchsummary: f,
			Found:         "[failed]",
			Image:         "",
			JS:            "",
		}
		AllSearches.Delete(srch.ID)
		return c.JSONPretty(http.StatusOK, soj, JSONINDENT)
	}

	// [b] make a query against the model
	searcher, err := search.New(embs...)
	if err != nil {
		return fail(FAIL1)
	}

	rank := VECTORNEIGHBORS // how many neighbors to output; min is 1
	word := srch.Seeking

	neighbors, err := searcher.SearchInternal(word, rank)
	if err != nil {
		return fail(FAIL2)
	}

	// neighbors.Describe()

	// "dextra" in big chunks of Lucan...

	//
	//   RANK |    WORD    | SIMILARITY
	//-------+------------+-------------
	//     1 | serpentum  |   0.962808
	//     2 | putat      |   0.954310
	//     3 | praecipiti |   0.947326
	//     4 | mors       |   0.942508
	//     5 | lux        |   0.940325
	//     6 | modum      |   0.938747
	//     7 | quisquis   |   0.938089
	//     8 | animae     |   0.936332
	//     9 | uiros      |   0.928818
	//    10 | etiam      |   0.927048

	// [c] prepare output

	table := make([][]string, len(neighbors))
	for i, n := range neighbors {
		table[i] = []string{
			fmt.Sprintf("%d", n.Rank),
			n.Word,
			fmt.Sprintf("%.4f", n.Similarity),
		}
	}

	out := "<pre>"
	for t := range table {
		out += fmt.Sprintf("%s\t%s\t\t\t%s\n", table[t][0], table[t][1], table[t][2])
	}
	out += "</pre>"

	// PYTHON GRAPHING TO MIMIC

	// want some day to build a graph as per matplotgraphmatches() in vectorgraphing.py

	//  FROM generatenearestneighbordata() in gensimnearestneighbors.py
	//  mostsimilar = findapproximatenearestneighbors(termone, vectorspace, vv)
	//  [('εὕρηϲιϲ', 1.0), ('εὑρίϲκω', 0.6673248708248138), ('φυϲιάω', 0.5833806097507477), ('νόμοϲ', 0.5505017340183258), ...]

	// FROM findapproximatenearestneighbors() in gensimnearestneighbors.py
	// 	explore = max(2500, vectorvalues.neighborscap)
	//
	//	try:
	//		mostsimilar = mymodel.wv.most_similar(query, topn=explore)
	//		mostsimilar = [s for s in mostsimilar if s[1] > vectorvalues.nearestneighborcutoffdistance]

	//  FROM matplotgraphmatches() in vectorgraphing.py:
	// 	edgelist = list()
	//	for t in mostsimilartuples:
	//		edgelist.append((searchterm, t[0], round(t[1]*10, 2)))
	//
	//	for r in relevantconnections:
	//		for c in relevantconnections[r]:
	//			edgelist.append((r, c[0], round(c[1]*10, 2)))
	//
	//	graph.add_weighted_edges_from(edgelist)
	//	edgelabels = {(u, v): d['weight'] for u, v, d in graph.edges(data=True)}

	// ? https://github.com/yourbasic/graph

	// ? https://github.com/go-echarts/go-echarts

	// look at what is possible and links: https://blog.gopheracademy.com/advent-2018/go-webgl/

	soj := SearchOutputJSON{
		Title:         "VECTORS",
		Searchsummary: "[no summary]",
		Found:         out,
		Image:         "",
		JS:            "",
	}

	AllSearches.Delete(srch.ID)

	return c.JSONPretty(http.StatusOK, soj, JSONINDENT)
}

// fingerprintvectorsearch - derive a unique md5 for any given mix of search items & vector settings
func fingerprintvectorsearch(srch SearchStruct) string {
	const (
		MSG1 = "VectorSearch() fingerprint: "
		FAIL = "fingerprintvectorsearch() failed to Marshal"
	)
	// unless you sort, you do not get repeatable results with a md5sum of srch.SearchIn if you look at "all latin"
	var inc []string
	sort.Strings(srch.SearchIn.AuGenres)
	sort.Strings(srch.SearchIn.WkGenres)
	sort.Strings(srch.SearchIn.AuLocations)
	sort.Strings(srch.SearchIn.WkLocations)
	sort.Strings(srch.SearchIn.Authors)
	sort.Strings(srch.SearchIn.Works)
	sort.Strings(srch.SearchIn.Passages)
	inc = append(inc, srch.SearchIn.AuGenres...)
	inc = append(inc, srch.SearchIn.WkGenres...)
	inc = append(inc, srch.SearchIn.AuLocations...)
	inc = append(inc, srch.SearchIn.WkLocations...)
	inc = append(inc, srch.SearchIn.Authors...)
	inc = append(inc, srch.SearchIn.Works...)
	inc = append(inc, srch.SearchIn.Passages...)

	var exc []string
	sort.Strings(srch.SearchEx.AuGenres)
	sort.Strings(srch.SearchEx.WkGenres)
	sort.Strings(srch.SearchEx.AuLocations)
	sort.Strings(srch.SearchEx.WkLocations)
	sort.Strings(srch.SearchEx.Authors)
	sort.Strings(srch.SearchEx.Works)
	sort.Strings(srch.SearchEx.Passages)
	exc = append(exc, srch.SearchEx.AuGenres...)
	exc = append(exc, srch.SearchEx.WkGenres...)
	exc = append(exc, srch.SearchEx.AuLocations...)
	exc = append(exc, srch.SearchEx.WkLocations...)
	exc = append(exc, srch.SearchEx.Authors...)
	exc = append(exc, srch.SearchEx.Works...)
	exc = append(exc, srch.SearchEx.Passages...)

	f1, e1 := json.Marshal(inc)
	f2, e2 := json.Marshal(exc)
	f3, e3 := json.Marshal(vectorconfig())
	if e1 != nil || e2 != nil || e3 != nil {
		msg(FAIL, 0)
		os.Exit(1)
	}
	f1 = append(f1, f2...)
	f1 = append(f1, f3...)
	m := fmt.Sprintf("%x", md5.Sum(f1))
	msg(MSG1+m, MSGTMI)

	return m
}

// buildtextblock - turn []DbWorkline into a single long string
func buildtextblock(lines []DbWorkline) string {
	const (
		FAIL1 = "failed to unmarshal %s into objmap\n"
		FAIL2 = "failed second pass unmarshal of %s into newmap\n"
	)
	// [a] get all the words we need
	var slicedwords []string
	for i := 0; i < len(lines); i++ {
		wds := lines[i].AccentedSlice()
		for _, w := range wds {
			slicedwords = append(slicedwords, UVσςϲ(SwapAcuteForGrave(w)))
		}
	}

	// [b] get basic morphology info for those words
	morphmapdbm := arraytogetrequiredmorphobjects(slicedwords) // map[string]DbMorphology

	// [c] figure out which headwords to associate with the collection of words

	// this information is inside DbMorphology.RawPossib
	// but it needs to be parsed
	// example: `{"1": {"transl": "A. nom. plur; II. a guardian god", "analysis": "masc gen pl", "headword": "deus", "scansion": "deūm", "xref_kind": "9", "xref_value": "22568216"}, "2": {"transl": "A. nom. plur; II. a guardian god", "analysis": "masc acc sg", "headword": "deus", "scansion": "", "xref_kind": "9", "xref_value": "22568216"}}`

	type possib struct {
		Trans string `json:"found"`
		Anal  string `json:"analysis"`
		Head  string `json:"headword"`
		Scans string `json:"scansion"`
		Xrefk string `json:"xref_kind"`
		Xrefv string `json:"xref_value"`
	}

	morphmapstrslc := make(map[string]map[string]bool, len(morphmapdbm))
	for m := range morphmapdbm {
		morphmapstrslc[m] = make(map[string]bool)
		// first pass: {"1": bytes1, "2": bytes2, ...}
		var objmap map[string]json.RawMessage
		err := json.Unmarshal([]byte(morphmapdbm[m].RawPossib), &objmap)
		if err != nil {
			fmt.Printf(FAIL1, morphmapdbm[m].Observed)
		}
		// second pass: : {"1": possib1, "2": possib2, ...}
		newmap := make(map[string]possib)
		for key, v := range objmap {
			var pp possib
			e := json.Unmarshal(v, &pp)
			if e != nil {
				fmt.Printf(FAIL2, morphmapdbm[m].Observed)
			}
			newmap[key] = pp
		}

		for _, v := range newmap {
			morphmapstrslc[m][v.Head] = true
		}
	}

	// if you just iterate over morphmapstrslc, you drop unparsed terms: the next will retain them
	for _, w := range slicedwords {
		if _, t := morphmapstrslc[w]; t {
			continue
		} else {
			morphmapstrslc[w] = make(map[string]bool)
			morphmapstrslc[w][w] = true
		}
	}

	// "morphmapstrslc" for Albinus , poet. [lt2002]
	// map[abscondere:map[abscondo:true] apte:map[apte:true aptus:true] capitolia:map[capitolium:true] celsa:map[celsus¹:true] concludere:map[concludo:true] cui:map[quis²:true quis¹:true qui²:true qui¹:true] dactylum:map[dactylus:true] de:map[de:true] deum:map[deus:true] fieri:map[fio:true] freta:map[fretum:true fretus¹:true] i:map[eo¹:true] ille:map[ille:true] iungens:map[jungo:true] liber:map[liber¹:true liber⁴:true libo¹:true] metris:map[metrum:true] moenibus:map[moenia¹:true] non:map[non:true] nulla:map[nullus:true] patuere:map[pateo:true patesco:true] posse:map[possum:true] repostos:map[re-pono:true] rerum:map[res:true] romanarum:map[romanus:true] sed:map[sed:true] sinus:map[sinus¹:true] spondeum:map[spondeum:true spondeus:true] sponte:map[sponte:true] ternis:map[terni:true] totum:map[totus²:true totus¹:true] triumphis:map[triumphus:true] tutae:map[tueor:true] uersum:map[verro:true versum:true versus³:true verto:true] urbes:map[urbs:true] †uilem:map[†uilem:true]]
	//

	// [d] swap out words for headwords
	winnermap := buildwinnertakesallparsemap(morphmapstrslc)

	// "winnermap" for Albinus , poet. [lt2002]
	// map[abscondere:[abscondo] apte:[aptus] capitolia:[capitolium] celsa:[celsus¹] concludere:[concludo] cui:[qui¹] dactylum:[dactylus] de:[de] deum:[deus] fieri:[fio] freta:[fretus¹] i:[eo¹] ille:[ille] iungens:[jungo] liber:[liber⁴] metris:[metrum] moenibus:[moenia¹] non:[non] nulla:[nullus] patuere:[pateo] posse:[possum] repostos:[re-pono] rerum:[res] romanarum:[romanus] sed:[sed] sinus:[sinus¹] spondeum:[spondeus] sponte:[sponte] ternis:[terni] totum:[totus¹] triumphis:[triumphus] tutae:[tueor] uersum:[verro] urbes:[urbs] †uilem:[†uilem]]

	// [e] turn results into unified text block

	// string addition will use a huge amount of time: 120s to concatinate Cicero: txt = txt + newtxt...
	// with strings.Builder we only need .1s to build the text...

	var sb strings.Builder
	preallocate := CHARSPERLINE * len(lines) // NB: a long line has 60 chars
	sb.Grow(preallocate)

	winner := true

	if winner {
		winnerstring(&sb, slicedwords, winnermap)
	} else {
		flatstring(&sb, slicedwords)
	}

	return strings.TrimSpace(sb.String())
}

// generateembeddings - turn a search into a collection of semantic vector embeddings
func generateembeddings(c echo.Context, srch SearchStruct) embedding.Embeddings {
	const (
		FAIL1 = "word2vec model initialization failed"
		FAIL2 = "generateembeddings() failed to train vector embeddings"
	)

	// note that MAXTEXTLINEGENERATION will prevent a vectorization of the full corpus
	vs := sessionintobulksearch(c, VECTORMAXLINES)
	srch.Results = vs.Results
	vs.Results = []DbWorkline{}

	thetext := buildtextblock(srch.Results)

	// "thetext" for Albinus , poet. [lt2002]
	// res romanus liber⁴ eo¹ ille qui¹ terni capitolium celsus¹ triumphus sponte deus pateo qui¹ fretus¹ nullus re-pono abscondo sinus¹ non tueor moenia¹ urbs de metrum †uilem spondeus totus¹ concludo verro possum fio jungo sed dactylus aptus

	// vs. "RERUM ROMANARUM LIBER I
	//	Ille cui ternis Capitolia celsa triumphis..."

	// [a] vectorize the text block

	vmodel, err := word2vec.NewForOptions(vectorconfig())
	if err != nil {
		msg(FAIL1, 1)
	}

	// input for  word2vec.Train() is 'io.ReadSeeker'
	b := bytes.NewReader([]byte(thetext))
	if err = vmodel.Train(b); err != nil {
		msg(FAIL2, 1)
	}

	// write word vector to disk & then read it off the disk
	//vfile := "/Users/erik/tmp/vect.out"
	//f, err := os.Create(vfile)
	//chke(err)
	//err = vmodel.Save(f, vector.Agg)
	//chke(err)
	//input, err := os.Open(vfile)
	//chke(err)
	//defer input.Close()
	//embs, err := embedding.Load(input)
	//chke(err)

	// use buffers; skip the disk; psql used for storage: vectordbadd() & vectordbfetch()
	var buf bytes.Buffer
	w := io.Writer(&buf)
	err = vmodel.Save(w, vector.Agg)

	r := io.Reader(&buf)
	embs, err := embedding.Load(r)
	chke(err)

	return embs
}

// flatstring - helper for buildtextblock() to generate unmodified text
func flatstring(sb *strings.Builder, slicedwords []string) {
	for i := 0; i < len(slicedwords); i++ {
		sb.WriteString(slicedwords[i] + " ")
	}
}

// winnerstring - helper for buildtextblock() to generate winner takes all substitutions
func winnerstring(sb *strings.Builder, slicedwords []string, winnermap map[string][]string) {
	for i := 0; i < len(slicedwords); i++ {
		// drop skipwords
		w := winnermap[slicedwords[i]][0]
		_, s1 := LatinStops[w]
		_, s2 := GreekStops[w]
		if s1 || s2 {
			continue
		} else {
			sb.WriteString(w + " ")
		}
	}
}

// buildwinnertakesallparsemap - figure out which is the most common of the possible headwords for any given word
func buildwinnertakesallparsemap(parsemap map[string]map[string]bool) map[string][]string {
	// turn a list of sentences into a list of headwords; here we figure out which headword is the dominant homonym
	// then we just use that term; "esse" always comes from "sum" and never "edo", etc.

	// [a] figure out all headwords in use

	allheadwords := make(map[string]bool)
	for i := range parsemap {
		for k, _ := range parsemap[i] {
			allheadwords[k] = true
		}
	}

	// [b] generate scoremap and assign scores to each of the headwords

	scoremap := fetchheadwordcounts(allheadwords)

	// [c] note that there are capital words in the parsemap that need lowering

	// [c1] lower the internal values first
	for i := range parsemap {
		newmap := make(map[string]bool)
		for k, _ := range parsemap[i] {
			newmap[strings.ToLower(k)] = true
		}
		parsemap[i] = newmap
	}

	// [c2] lower the parsemap keys; how worried should we be about the collisions...
	lcparsemap := make(map[string]map[string]bool)
	for i := range parsemap {
		lcparsemap[strings.ToLower(i)] = parsemap[i]
	}

	// [d] run through the parsemap and kill off the losers

	newparsemap := make(map[string][]string)
	for i := range lcparsemap {
		var hwl WHWList
		// for j := 0; j < len(lcparsemap[i]); j++ {
		for j, _ := range parsemap[i] {
			var thishw WeightedHeadword
			thishw.Word = j
			thishw.Count = scoremap[j]
			hwl = append(hwl, thishw)
		}
		sort.Sort(hwl)

		newparsemap[i] = make([]string, 0, 1)
		newparsemap[i] = append(newparsemap[i], hwl[0].Word)
	}

	return newparsemap
}
