package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ynqa/wego/pkg/embedding"
	"github.com/ynqa/wego/pkg/model/modelutil/vector"
	"github.com/ynqa/wego/pkg/model/word2vec"
	"github.com/ynqa/wego/pkg/search"
	"net/http"
	"os"
	"sort"
	"strings"
)

var (
	// Latin100 - the 100 most common latin headwords
	Latin100 = []string{"qui¹", "et", "in", "edo¹", "is", "sum¹", "hic", "non", "ab", "ut", "Cos²", "si", "ad", "cum", "ex", "a", "eo¹",
		"ego", "quis¹", "tu", "Eos", "dico²", "ille", "sed", "de", "neque", "facio", "possum", "atque", "sui", "res",
		"quam", "aut", "ipse", "huc", "habeo", "do", "omne", "video", "ito", "magnus", "b", "alius²", "for", "idem",
		"suum", "etiam", "per", "enim", "omnes", "ita", "suus", "omnis", "autem", "vel", "vel", "Alius¹", "qui²", "quo",
		"nam", "bonus", "neo¹", "meus", "volo¹", "ne³", "ne¹", "suo", "verus", "pars", "reor", "sua", "vaco", "verum",
		"primus", "unus", "multus", "causa", "jam", "tamen", "Sue", "nos", "dies", "Ios", "modus", "tuus", "venio",
		"pro¹", "pro²", "ago", "deus", "annus", "locus", "homo", "pater", "eo²", "tantus", "fero", "quidem", "noster",
		"an", "locum"}
	// Greek150 - the 150 most common greek headwords
	Greek150 = []string{"ὁ", "καί", "τίϲ", "ἔδω", "δέ", "εἰμί", "δέω¹", "δεῖ", "δέομαι", "εἰϲ", "αὐτόϲ", "τιϲ", "οὗτοϲ", "ἐν",
		"γάροϲ", "γάρον", "γάρ", "οὐ", "μένω", "μέν", "τῷ", "ἐγώ", "ἡμόϲ", "κατά", "Ζεύϲ", "ἐπί", "ὡϲ", "διά",
		"πρόϲ", "προϲάμβ", "τε", "πᾶϲ", "ἐκ", "ἕ", "ϲύ", "Ἀλλά", "γίγνομαι", "ἁμόϲ", "ὅϲτιϲ", "ἤ¹", "ἤ²", "ἔχω",
		"ὅϲ", "μή", "ὅτι¹", "λέγω¹", "ὅτι²", "τῇ", "Τήιοϲ", "ἀπό", "εἰ", "περί", "ἐάν", "θεόϲ", "φημί", "ἐκάϲ",
		"ἄν¹", "ἄνω¹", "ἄλλοϲ", "qui¹", "πηρόϲ", "παρά", "ἀνά", "αὐτοῦ", "ποιέω", "ἄναξ", "ἄνα", "ἄν²", "πολύϲ",
		"οὖν", "λόγοϲ", "οὕτωϲ", "μετά", "ἔτι", "ὑπό", "ἑαυτοῦ", "ἐκεῖνοϲ", "εἶπον", "πρότεροϲ", "edo¹", "μέγαϲ",
		"ἵημι", "εἷϲ", "οὐδόϲ", "οὐδέ", "ἄνθρωποϲ", "ἠμί", "μόνοϲ", "κύριοϲ", "διό", "οὐδείϲ", "ἐπεί", "πόλιϲ",
		"τοιοῦτοϲ", "χάω", "καθά", "θεάομαι", "γε", "ἕτεροϲ", "δοκέω", "λαμβάνω", "δή", "δίδωμι", "ἵνα",
		"βαϲιλεύϲ", "φύϲιϲ", "ἔτοϲ", "πατήρ", "ϲῶμα", "καλέω", "ἐρῶ", "υἱόϲ", "ὅϲοϲ", "γαῖα", "οὔτε", "οἷοϲ",
		"ἀνήρ", "ὁράω", "ψυχή", "Ἔχιϲ", "ὥϲπερ", "αὐτόϲε", "χέω", "ὑπέρ", "ϲόϲ", "θεάω", "νῦν", "ἐμόϲ", "δύναμαι",
		"φύω", "πάλιν", "ὅλοξ", "ἀρχή", "καλόϲ", "δύναμιϲ", "πωϲ", "δύο", "ἀγαθόϲ", "οἶδα", "δείκνυμι", "χρόνοϲ",
		"ὅμοιοϲ", "ἕκαϲτοϲ", "ὁμοῖοϲ", "ὥϲτε", "ἡμέρα", "γράφω", "δραχμή", "μέροϲ"}
	// LatinKeep - members of Latin100 we will not toss
	LatinKeep = []string{"facio", "possum", "habeo", "video", "magnus", "bonus", "volo¹", "primus", "venio", "ago",
		"deus", "annus", "locus", "pater", "fero"}
	// GreekKeep - members of Greek150 we will not toss
	GreekKeep = []string{"ἔχω", "λέγω¹", "θεόϲ", "φημί", "ποιέω", "ἵημι", "μόνοϲ", "κύριοϲ", "πόλιϲ", "θεάομαι", "δοκέω", "λαμβάνω",
		"δίδωμι", "βαϲιλεύϲ", "φύϲιϲ", "ἔτοϲ", "πατήρ", "ϲῶμα", "καλέω", "ἐρῶ", "υἱόϲ", "γαῖα", "ἀνήρ", "ὁράω",
		"ψυχή", "δύναμαι", "ἀρχή", "καλόϲ", "δύναμιϲ", "ἀγαθόϲ", "οἶδα", "δείκνυμι", "χρόνοϲ", "γράφω", "δραχμή",
		"μέροϲ"}
	LatinStops = getlatinstops()
	GreekStops = getgreekstops()
)

//type ModelType = string
//
//const (
//	Cbow     ModelType = "cbow"
//	SkipGram ModelType = "skipgram"
//)

//type OptimizerType = string
//
//const (
//	NegativeSampling    OptimizerType = "ns"
//	HierarchicalSoftmax OptimizerType = "hs"
//)
//
//type Options struct {
//	BatchSize          int
//	Dim                int
//	DocInMemory        bool
//	Goroutines         int
//	Initlr             float64
//	Iter               int
//	LogBatch           int
//	MaxCount           int
//	MaxDepth           int
//	MinCount           int
//	MinLR              float64
//	ModelType          ModelType
//	NegativeSampleSize int
//	OptimizerType      OptimizerType
//	SubsampleThreshold float64
//	ToLower            bool
//	UpdateLRBatch      int
//	Verbose            bool
//	Window             int
//}

func VectorSearch(c echo.Context, srch SearchStruct) error {
	vs := sessionintobulksearch(c, MAXTEXTLINEGENERATION)
	srch.Results = vs.Results
	vs.Results = []DbWorkline{}

	var slicedwords []string
	for i := 0; i < len(srch.Results); i++ {
		wds := srch.Results[i].AccentedSlice()
		for _, w := range wds {
			slicedwords = append(slicedwords, UVσςϲ(SwapAcuteForGrave(w)))
		}
	}

	morphmapdbm := arraytogetrequiredmorphobjects(slicedwords) // map[string]DbMorphology

	// figure out which headwords to associate with the collection of words
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
			fmt.Printf("failed to unmarshal %s into objmap\n", morphmapdbm[m].Observed)
		}
		// second pass: : {"1": possib1, "2": possib2, ...}
		newmap := make(map[string]possib)
		for key, v := range objmap {
			var pp possib
			e := json.Unmarshal(v, &pp)
			if e != nil {
				fmt.Printf("failed second pass unmrashal of %s into newmap\n", morphmapdbm[m].Observed)
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

	winnermap := buildwinnertakesallparsemap(morphmapstrslc)

	// "winnermap" for Albinus , poet. [lt2002]
	// map[abscondere:[abscondo] apte:[aptus] capitolia:[capitolium] celsa:[celsus¹] concludere:[concludo] cui:[qui¹] dactylum:[dactylus] de:[de] deum:[deus] fieri:[fio] freta:[fretus¹] i:[eo¹] ille:[ille] iungens:[jungo] liber:[liber⁴] metris:[metrum] moenibus:[moenia¹] non:[non] nulla:[nullus] patuere:[pateo] posse:[possum] repostos:[re-pono] rerum:[res] romanarum:[romanus] sed:[sed] sinus:[sinus¹] spondeum:[spondeus] sponte:[sponte] ternis:[terni] totum:[totus¹] triumphis:[triumphus] tutae:[tueor] uersum:[verro] urbes:[urbs] †uilem:[†uilem]]

	// turn results into unified text block

	// string addition will use a huge amount of time: 120s to concatinate Cicero: txt = txt + newtxt...
	// with strings.Builder we only need .1s to build the text...

	var sb strings.Builder
	preallocate := CHARSPERLINE * len(srch.Results) // NB: a long line has 60 chars
	sb.Grow(preallocate)

	winner := true

	if winner {
		winnerstring(&sb, slicedwords, winnermap)
	} else {
		flatstring(&sb, slicedwords)
	}

	thetext := strings.TrimSpace(sb.String())

	// "thetext" for Albinus , poet. [lt2002]
	// res romanus liber⁴ eo¹ ille qui¹ terni capitolium celsus¹ triumphus sponte deus pateo qui¹ fretus¹ nullus re-pono abscondo sinus¹ non tueor moenia¹ urbs de metrum †uilem spondeus totus¹ concludo verro possum fio jungo sed dactylus aptus
	// vs. "RERUM ROMANARUM LIBER I
	//	Ille cui ternis Capitolia celsa triumphis..."

	// fyi
	//opts := word2vec.Options{
	//	BatchSize:          0,
	//	Dim:                0,
	//	DocInMemory:        false,
	//	Goroutines:         0,
	//	Initlr:             0,
	//	Iter:               0,
	//	LogBatch:           0,
	//	MaxCount:           0,
	//	MaxDepth:           0,
	//	MinCount:           0,
	//	MinLR:              0,
	//	ModelType:          "cbow",
	//	NegativeSampleSize: 0,
	//	OptimizerType:      "ns",
	//	SubsampleThreshold: 0,
	//	ToLower:            false,
	//	UpdateLRBatch:      0,
	//	Verbose:            false,
	//	Window:             0,
	//}

	//const (
	//	NegativeSampling    OptimizerType = "ns"
	//	HierarchicalSoftmax OptimizerType = "hs"
	//)

	//const (
	//	Cbow     ModelType = "cbow"
	//	SkipGram ModelType = "skipgram"
	//)

	// var (
	//	defaultBatchSize          = 10000
	//	defaultDim                = 10
	//	defaultDocInMemory        = false
	//	defaultGoroutines         = runtime.NumCPU()
	//	defaultInitlr             = 0.025
	//	defaultIter               = 15
	//	defaultLogBatch           = 100000
	//	defaultMaxCount           = -1
	//	defaultMaxDepth           = 100
	//	defaultMinCount           = 5
	//	defaultMinLR              = defaultInitlr * 1.0e-4
	//	defaultModelType          = Cbow
	//	defaultNegativeSampleSize = 5
	//	defaultOptimizerType      = NegativeSampling
	//	defaultSubsampleThreshold = 1.0e-3
	//	defaultToLower            = false
	//	defaultUpdateLRBatch      = 100000
	//	defaultVerbose            = false
	//	defaultWindow             = 5
	//)

	// results do not repeat because word2vec.Train() in pkg/model/word2vec/word2vec.go has
	// "vec[i] = (rand.Float64() - 0.5) / float64(dim)"

	// modelbuilders.py
	// 	negative (int, optional) – If > 0, negative sampling will be used, the int for negative specifies how many “noise words” should be drawn (usually between 5-20). If set to 0, no negative sampling is used.
	//	seed (int, optional) – Seed for the random number generator. Initial vectors for each word are seeded with a hash of the concatenation of word + str(seed). Note that for a fully deterministically-reproducible run, you must also limit the model to a single worker thread (workers=1), to eliminate ordering jitter from OS thread scheduling. (In Python 3, reproducibility between interpreter launches also requires use of the PYTHONHASHSEED environment variable to control hash randomization).
	// 	compute_loss (bool, optional) – If True, computes and stores loss value which can be retrieved using get_latest_training_loss()
	//  window (int, optional) – Maximum distance between the current and predicted word within a sentence
	//                gensimmodel = Word2Vec(bagsofwords,
	//                                       min_count=vv.minimumpresence,
	//                                       seed=1,
	//                                       epochs=vv.trainingiterations,
	//                                       vector_size=vv.dimensions,
	//                                       sample=vv.downsample,
	//                                       sg=1,  # the results seem terrible if you say sg=0
	//                                       window=vv.window,
	//                                       workers=workers,
	//                                       compute_loss=computeloss)

	//opts = word2vec.DefaultOptions()
	//opts.OptimizerType = "hs"
	//opts.Dim = 200
	//opts.DocInMemory = true
	//opts.Iter = 50 // need to crank this up well past the default if you are going to get separation instead of .9998, .9997, .9997, ...
	//// opts.MaxCount = 35
	//opts.MinCount = 8
	//// opts.ModelType = "skipgram"
	//opts.Window = 8

	opts := vectorconfig()

	// TODO
	// PROBLEM: 10 dimensional output no matter how many dimensions requested

	vmodel, err := word2vec.NewForOptions(opts)
	if err != nil {
		// problem
	}

	// input for  word2vec.Train() is 'io.ReadSeeker'
	b := bytes.NewReader([]byte(thetext))
	if err = vmodel.Train(b); err != nil {
		// failed to train.
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

	// write word vector.

	vfile := "/Users/erik/tmp/vect.out"
	rank := 10 // how many neighbors to output; min is 1
	word := srch.Seeking

	f, err := os.Create(vfile)
	if err != nil {
		return fail("failed to create vect.out")
	}
	err = vmodel.Save(f, vector.Agg)
	if err != nil {
		return fail("failed to save vect.out")
	}

	input, err := os.Open(vfile)
	if err != nil {
		return fail("err: os.Open(vfile)")
	}
	defer input.Close()
	embs, err := embedding.Load(input)
	if err != nil {
		return fail("err: embedding.Load(input)")
	}

	searcher, err := search.New(embs...)
	if err != nil {
		return fail("err: search.New(embs...)")
	}

	neighbors, err := searcher.SearchInternal(word, rank)
	if err != nil {
		return fail("err: searcher.SearchInternal(word, rank)")
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

	table := make([][]string, len(neighbors))
	for i, n := range neighbors {
		table[i] = []string{
			fmt.Sprintf("%d", n.Rank),
			n.Word,
			fmt.Sprintf("%f", n.Similarity),
		}
	}

	out := "<pre>"
	for t := range table {
		out += fmt.Sprintf("%s\t%s\t\t%s\n", table[t][0], table[t][1], table[t][2])
	}
	out += "</pre>"

	// want some day to build a graph as per matplotgraphmatches() in vectorgraphing.py

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

func flatstring(sb *strings.Builder, slicedwords []string) {
	for i := 0; i < len(slicedwords); i++ {
		sb.WriteString(slicedwords[i] + " ")
	}
}

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

func buildwinnertakesallparsemap(parsemap map[string]map[string]bool) map[string][]string {
	// turn a list of sentences into a list of list of headwords; here we figure out which headword is the dominant homonym
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

type WeightedHeadword struct {
	Word  string
	Count int
}

type WHWList []WeightedHeadword

func (w WHWList) Len() int {
	return len(w)
}

func (w WHWList) Less(i, j int) bool {
	return w[i].Count > w[j].Count
}

func (w WHWList) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
}

func fetchheadwordcounts(headwordset map[string]bool) map[string]int {
	if len(headwordset) == 0 {
		return make(map[string]int)
	}

	tt := "CREATE TEMPORARY TABLE ttw_%s AS SELECT words AS w FROM unnest(ARRAY[%s]) words"
	qt := "SELECT entry_name, total_count FROM dictionary_headword_wordcounts WHERE EXISTS " +
		"(SELECT 1 FROM ttw_%s temptable WHERE temptable.w = dictionary_headword_wordcounts.entry_name)"

	rndid := strings.Replace(uuid.New().String(), "-", "", -1)

	hw := make([]string, 0, len(headwordset))
	for h := range headwordset {
		hw = append(hw, h)
	}

	dbconn := GetPSQLconnection()

	arr := strings.Join(hw, "', '")
	arr = fmt.Sprintf("'%s'", arr)

	tt = fmt.Sprintf(tt, rndid, arr)
	_, err := dbconn.Exec(context.Background(), tt)
	chke(err)

	qt = fmt.Sprintf(qt, rndid)
	foundrows, e := dbconn.Query(context.Background(), qt)
	chke(e)

	returnmap := make(map[string]int)
	defer foundrows.Close()
	for foundrows.Next() {
		var thehit WeightedHeadword
		err = foundrows.Scan(&thehit.Word, &thehit.Count)
		chke(err)
		returnmap[thehit.Word] = thehit.Count
	}

	// don't kill off unfound terms
	for i := range hw {
		if _, t := returnmap[hw[i]]; t {
			continue
		} else {
			returnmap[hw[i]] = 0
		}
	}

	// "returnmap" for Albinus , poet. [lt2002]
	// map[abscondo:213 apte:168 aptus:1423 capitolium:0 celsus¹:1050 concludo:353 dactylus:167 de:42695 deus:14899 eo¹:58129 fio:12305 fretum:746 fretus¹:761 ille:44214 jungo:2275 liber¹:7550 liber⁴:13403 libo¹:3996 metrum:383 moenia¹:1308 non:96475 nullus:11785 pateo:1828 patesco:46 possum:41631 quis²:0 quis¹:52619 qui²:19812 qui¹:251744 re-pono:47 res:38669 romanus:0 sed:44131 sinus¹:1223 spondeum:158 spondeus:205 sponte:841 terni:591 totus²:0 totus¹:9166 triumphus:1058 tueor:3734 urbs:8564 verro:3843 versum:435 versus³:3390 verto:1471 †uilem:0]

	return returnmap
}

func getgreekstops() map[string]struct{} {
	gs := SetSubtraction(Greek150, GreekKeep)
	return ToSet(gs)
}

func getlatinstops() map[string]struct{} {
	ls := SetSubtraction(Latin100, LatinKeep)
	return ToSet(ls)
}

func vectorconfig() word2vec.Options {
	const (
		ERR1 = "vectorconfig() cannot find UserHomeDir"
		ERR2 = "vectorconfig() failed to parse "
		MSG1 = "wrote default vector configuration file "
		MSG2 = "read vector configuration from "
	)
	h, e := os.UserHomeDir()
	cfg := word2vec.DefaultOptions()
	if e != nil {
		msg(ERR1, 0)
		return cfg
	}
	_, yes := os.Stat(fmt.Sprintf(CONFIGALTAPTH, h) + CONFIGVECTOR)

	if yes != nil {
		content, err := json.MarshalIndent(cfg, JSONINDENT, JSONINDENT)
		chke(err)

		err = os.WriteFile(fmt.Sprintf(CONFIGALTAPTH, h)+CONFIGVECTOR, content, 0644)
		chke(err)
		msg(MSG1+CONFIGVECTOR, 1)
	} else {
		loadedcfg, _ := os.Open(fmt.Sprintf(CONFIGALTAPTH, h) + CONFIGVECTOR)
		decoderc := json.NewDecoder(loadedcfg)
		confc := word2vec.Options{}
		errc := decoderc.Decode(&confc)
		_ = loadedcfg.Close()
		if errc != nil {
			msg(ERR2+CONFIGVECTOR, 0)
			cfg = word2vec.DefaultOptions()
		}
		msg(MSG2+CONFIGVECTOR, 2)
	}

	return cfg
}
