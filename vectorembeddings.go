//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/e-gun/wego/pkg/embedding"
	"github.com/e-gun/wego/pkg/model"
	"github.com/e-gun/wego/pkg/model/glove"
	"github.com/e-gun/wego/pkg/model/lexvec"
	"github.com/e-gun/wego/pkg/model/modelutil/vector"
	"github.com/e-gun/wego/pkg/model/word2vec"
	"github.com/e-gun/wego/pkg/search"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"io"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"
)

//
// BAGGING
//

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
	LatExtra = []string{"at", "o", "tum", "tunc", "dum", "illic", "quia", "sive", "num", "adhuc", "tam", "ibi", "cur",
		"usquam", "quoque", "duo", "talis", "simul", "igitur", "utique²", "aliqui", "apud", "sic", "umquam", "ergo",
		"ob", "xu", "x", "iii", "u", "post", "ac", "ut"}
	LatStop = append(Latin100, LatExtra...)
	// LatinKeep - members of LatStop we will not toss
	LatinKeep = []string{"facio", "possum", "habeo", "video", "magnus", "bonus", "volo¹", "primus", "venio", "ago",
		"deus", "annus", "locus", "pater", "fero"}
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
	GreekExtra = []string{"ἀεί", "ὡϲαύτωϲ", "μηδέποτε", "μηδέ", "μηδ", "μηδέ", "ταὐτόϲ", "νυνί", "μεθ", "ἀντ", "μέχρι",
		"ἄνωθεν", "ὀκτώ", "ἓξ", "μετ", "τ", "μ", "αὐτόθ", "οὐδ", "εἵνεκ", "νόϲφι", "ἐκεῖ", "οὔκουν", "θ", "μάλιϲτ", "ὧδε",
		"πη", "τῇδ", "δι", "πρό", "ἀλλ", "ἕνεκα", "δ", "ἀλλά", "ἔπειτα", "καθ", "ταῦθ", "μήποτ", "ἀπ", "κ", "μήτ",
		"εὖτ", "αὖθιϲ", "∙∙∙", "∙∙", "∙", "∙∙∙∙", "oxy", "col", "fr", "*"}
	GreekStop = append(Greek150, GreekExtra...)
	// GreekKeep - members of GreekStop we will not toss
	GreekKeep = []string{"ἔχω", "λέγω¹", "θεόϲ", "φημί", "ποιέω", "ἵημι", "μόνοϲ", "κύριοϲ", "πόλιϲ", "θεάομαι", "δοκέω", "λαμβάνω",
		"δίδωμι", "βαϲιλεύϲ", "φύϲιϲ", "ἔτοϲ", "πατήρ", "ϲῶμα", "καλέω", "ἐρῶ", "υἱόϲ", "γαῖα", "ἀνήρ", "ὁράω",
		"ψυχή", "δύναμαι", "ἀρχή", "καλόϲ", "δύναμιϲ", "ἀγαθόϲ", "οἶδα", "δείκνυμι", "χρόνοϲ", "γράφω", "δραχμή",
		"μέροϲ", "λόγοϲ"}
	DefaultW2VVectors = word2vec.Options{
		BatchSize:          1024,
		Dim:                125,
		DocInMemory:        true,
		Goroutines:         20,
		Initlr:             0.025,
		Iter:               15,
		LogBatch:           100000,
		MaxCount:           -1,
		MaxDepth:           150,
		MinCount:           10,
		MinLR:              0.0000025,
		ModelType:          "skipgram",
		NegativeSampleSize: 5,
		OptimizerType:      "hs",
		SubsampleThreshold: 0.001,
		ToLower:            false,
		UpdateLRBatch:      100000,
		Verbose:            true,
		Window:             8,
	}
	// DefaultGloveVectors - wego's default: {0.75 10000 inc 10 false 20 0.025 15 100000 -1 5 sgd 0.001 false false 5 100}
	DefaultGloveVectors = glove.Options{
		// see also: https://nlp.stanford.edu/projects/glove/
		Alpha:              0.55,
		BatchSize:          1024,
		CountType:          "inc", // "inc", "prox" available; but we panic on "prox"
		Dim:                75,
		DocInMemory:        true,
		Goroutines:         20,
		Initlr:             0.025,
		Iter:               25,
		LogBatch:           100000,
		MaxCount:           -1,
		MinCount:           10,
		SolverType:         "adagrad", // "sdg", "adagrad" available
		SubsampleThreshold: 0.001,
		ToLower:            false,
		Verbose:            true,
		Window:             8,
		Xmax:               90,
	}
	DefaultLexVecVectors = lexvec.Options{
		BatchSize:          1024,
		Dim:                125,
		DocInMemory:        true,
		Goroutines:         20,
		Initlr:             0.025,
		Iter:               15,
		LogBatch:           100000,
		MaxCount:           -1,
		MinCount:           10,
		MinLR:              0.025 * 1.0e-4,
		NegativeSampleSize: 5,
		RelationType:       "ppmi", // "ppmi", "pmi", "co", "logco" are available
		Smooth:             0.75,
		SubsampleThreshold: 1.0e-3,
		ToLower:            false,
		UpdateLRBatch:      100000,
		Verbose:            true,
		Window:             8,
	}
)

func getgreekstops() map[string]struct{} {
	gs := SetSubtraction(GreekStop, GreekKeep)
	return ToSet(gs)
}

func getlatinstops() map[string]struct{} {
	ls := SetSubtraction(LatStop, LatinKeep)
	return ToSet(ls)
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

// generateneighborsdata - generate the Neighbors data for a headword within a search
func generateneighborsdata(c echo.Context, srch SearchStruct) map[string]search.Neighbors {
	const (
		FMSG  = `Fetching a stored model`
		GMSG  = `Generating a model`
		FAIL1 = "generateneighborsdata() could not find neighbors of a neighbor: '%s' neighbors (via '%s')"
		FAIL2 = "generateneighborsdata() failed to produce a Searcher"
		FAIL3 = "generateneighborsdata() failed to yield Neighbors"
		MQMEG = `Querying the model`
	)

	fp := fingerprintvectorsearch(srch, srch.VecModeler, srch.VecTextPrep)
	isstored := vectordbcheck(fp)
	var embs embedding.Embeddings
	if isstored {
		srch.ExtraMsg = FMSG
		AllSearches.InsertSS(srch)
		embs = vectordbfetch(fp)
	} else {
		srch.ExtraMsg = GMSG
		AllSearches.InsertSS(srch)
		embs = generateembeddings(c, srch.VecModeler, srch)
		vectordbadd(fp, embs)
	}

	// [b] make a query against the model

	srch.ExtraMsg = MQMEG
	AllSearches.InsertSS(srch)

	searcher, err := search.New(embs...)
	if err != nil {
		msg(FAIL2, MSGFYI)
		searcher = func() *search.Searcher { return &search.Searcher{} }()
	}

	se := AllSessions.GetSess(readUUIDCookie(c))
	ncount := se.VecNeighbCt // how many neighbors to output; min is 1
	if ncount < VECTORNEIGHBORSMIN || ncount > VECTORNEIGHBORSMAX {
		ncount = VECTORNEIGHBORS
	}

	word := srch.LemmaOne

	nn := make(map[string]search.Neighbors)
	neighbors, err := searcher.SearchInternal(word, ncount)
	if err != nil {
		msg(FAIL3, MSGFYI)
		neighbors = search.Neighbors{}
	}

	nn[word] = neighbors
	for _, n := range neighbors {
		meta, e := searcher.SearchInternal(n.Word, ncount)
		if e != nil {
			msg(fmt.Sprintf(FAIL1, n.Word, word), MSGFYI)
		} else {
			nn[n.Word] = meta
		}
	}

	return nn
}

// generateembeddings - turn a search into a collection of semantic vector embeddings
func generateembeddings(c echo.Context, modeltype string, srch SearchStruct) embedding.Embeddings {
	const (
		FAIL1  = "model initialization failed"
		FAIL2  = "generateembeddings() failed to train vector embeddings"
		MSG1   = "generateembeddings() gathered %d lines"
		MSG2   = "generateembeddings() successfuly trained a %s model"
		PRLMSG = `Acquiring the raw data`
		TBMSG  = `Turning %d lines into a unified text block`
		VMSG   = `Training run <code>#%d</code> out of <code>%d</code> total iterations.`
		DBMSG  = `Storing the model in the database. Then fetching it again.`
	)

	// vectorbot sends a search with pre-generated results:
	// lack of a real session means we can't call readUUIDCookie() repeatedly
	// this also means we need the "modeltype" parameter as well (bot: configtype; surfer: sessiontype)

	srch.ExtraMsg = PRLMSG
	AllSearches.InsertSS(srch)

	var vs SearchStruct
	if len(srch.Results) == 0 {
		vs = sessionintobulksearch(c, Config.VectorMaxlines)
	}

	msg(fmt.Sprintf(MSG1, len(vs.Results)), MSGPEEK)

	p := message.NewPrinter(language.English)
	srch.ExtraMsg = p.Sprintf(TBMSG, len(vs.Results))
	AllSearches.InsertSS(srch)

	srch.Results = vs.Results
	vs.Results = []DbWorkline{}

	thetext := buildtextblock(srch.VecTextPrep, srch.Results)
	srch.Results = []DbWorkline{}

	// "thetext" for Albinus , poet. [lt2002]
	// res romanus liber⁴ eo¹ ille qui¹ terni capitolium celsus¹ triumphus sponte deus pateo qui¹ fretus¹ nullus re-pono abscondo sinus¹ non tueor moenia¹ urbs de metrum †uilem spondeus totus¹ concludo verro possum fio jungo sed dactylus aptus

	// vs. "RERUM ROMANARUM LIBER I
	//	Ille cui ternis Capitolia celsa triumphis..."

	// [a] vectorize the text block

	var vmodel model.Model
	var ti int

	switch modeltype {
	case "glove":
		cfg := glovevectorconfig()
		m, err := glove.NewForOptions(cfg)
		if err != nil {
			msg(FAIL1, 1)
		}
		vmodel = m
		ti = cfg.Iter
	case "lexvec":
		cfg := lexvecvectorconfig()
		m, err := lexvec.NewForOptions(cfg)
		if err != nil {
			msg(FAIL1, 1)
		}
		vmodel = m
		ti = cfg.Iter
	default:
		cfg := w2vvectorconfig()
		m, err := word2vec.NewForOptions(cfg)
		if err != nil {
			msg(FAIL1, 1)
		}
		vmodel = m
		ti = cfg.Iter
	}

	// input for  word2vec.Train() is 'io.ReadSeeker'
	b := bytes.NewReader([]byte(thetext))

	finished := make(chan bool)

	// .Train() but do not block; so we can also .Reporter()
	go func() {
		if err := vmodel.Train(b); err != nil {
			msg(FAIL2, 1)
		} else {
			msg(fmt.Sprintf(MSG2, Config.VectorModel), MSGTMI)
		}
		finished <- true
	}()

	ct := make(chan int)
	rep := make(chan string)
	go vmodel.Reporter(ct, rep)

	getreport := func() {
		// wd := "unk"
		// tm := "n/a"
		in := 0
		for {
			select {
			case m := <-ct:
				in = m
			case m := <-rep:
				// msg(m, 2)
				// [HGS] trained 100062 words 529.0315ms
				coll := strings.Split(m, " ")
				if len(coll) == 4 {
					// wd = coll[1]
					// tm = coll[3]
				}
			}
			srch.ExtraMsg = fmt.Sprintf(VMSG, in, ti)
			AllSearches.InsertSS(srch)
			time.Sleep(WSPOLLINGPAUSE)
		}
	}

	go getreport()

	_ = <-finished

	srch.ExtraMsg = DBMSG
	AllSearches.InsertSS(srch)

	// use buffers; skip the disk; psql used for storage: vectordbadd() & vectordbfetch()
	var buf bytes.Buffer
	w := io.Writer(&buf)
	err := vmodel.Save(w, vector.Agg)

	r := io.Reader(&buf)
	embs, err := embedding.Load(r)
	chke(err)

	return embs
}

// buildtextblock - turn []DbWorkline into a single long string
func buildtextblock(method string, lines []DbWorkline) string {
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

	// [e] turn results into unified text block

	// string addition will use a huge amount of time: 120s to concatinate Cicero: txt = txt + newtxt...
	// with strings.Builder we only need .1s to build the text...

	var sb strings.Builder
	preallocate := CHARSPERLINE * len(lines) // NB: a long line has 60 chars
	sb.Grow(preallocate)

	switch method {
	case "unparsed":
		flatstring(&sb, slicedwords)
	case "montecarlo":
		mcm := buildmontecarloparsemap(morphmapstrslc)

		// "mcm" for Albinus , poet. [lt2002]
		// map[abscondere:{213 map[213:abscondo]} apte:{1591 map[168:apte 1591:aptus]} capitolia:{0 map[0:capitolium]} celsa:{1050 map[1050:celsus¹]} concludere:{353 map[353:concludo]} cui:{324175 map[0:quis² 251744:qui¹ 271556:qui² 324175:quis¹]} dactylum:{167 map[167:dactylus]} de:{42695 map[42695:de]} deum:{14899 map[14899:deus]} fieri:{12305 map[12305:fio]} freta:{1507 map[746:fretum 1507:fretus¹]} i:{58129 map[58129:eo¹]} ille:{44214 map[44214:ille]} iungens:{2275 map[2275:jungo]} liber:{24949 map[7550:liber¹ 20953:liber⁴ 24949:libo¹]} metris:{383 map[383:metrum]} moenibus:{1308 map[1308:moenia¹]} non:{96475 map[96475:non]} nulla:{11785 map[11785:nullus]} patuere:{1874 map[1828:pateo 1874:patesco]} posse:{41631 map[41631:possum]} repostos:{47 map[47:re-pono]} rerum:{38669 map[38669:res]} romanarum:{0 map[0:romanus]} sed:{44131 map[44131:sed]} sinus:{1223 map[1223:sinus¹]} spondeum:{363 map[158:spondeum 363:spondeus]} sponte:{841 map[841:sponte]} ternis:{591 map[591:terni]} totum:{9166 map[0:totus² 9166:totus¹]} triumphis:{1058 map[1058:triumphus]} tutae:{3734 map[3734:tueor]} uersum:{9139 map[1471:verto 5314:verro 5749:versum 9139:versus³]} urbes:{8564 map[8564:urbs]} †uilem:{0 map[0:†uilem]}]

		montecarlostring(&sb, slicedwords, mcm)
	case "yoked":
		yokedmap := buildyokedparsemap(morphmapstrslc)

		// "yokedmap" for Albinus , poet. [lt2002]
		// map[abscondere:abscondo apte:apte•aptus capitolia:capitolium celsa:celsus¹ concludere:concludo cui:quis²•quis¹•qui²•qui¹ dactylum:dactylus de:de deum:deus fieri:fio freta:fretum•fretus¹ i:eo¹ ille:ille iungens:jungo liber:liber¹•liber⁴•libo¹ metris:metrum moenibus:moenia¹ non:non nulla:nullus patuere:pateo•patesco posse:possum repostos:re-pono rerum:res romanarum:romanus sed:sed sinus:sinus¹ spondeum:spondeum•spondeus sponte:sponte ternis:terni totum:totus²•totus¹ triumphis:triumphus tutae:tueor uersum:verro•versum•versus³•verto urbes:urbs †uilem:†uilem]

		yokedstring(&sb, slicedwords, yokedmap)
	default: // "winner"
		winnermap := buildwinnertakesallparsemap(morphmapstrslc)

		// "winnermap" for Albinus , poet. [lt2002]
		// map[abscondere:[abscondo] apte:[aptus] capitolia:[capitolium] celsa:[celsus¹] concludere:[concludo] cui:[qui¹] dactylum:[dactylus] de:[de] deum:[deus] fieri:[fio] freta:[fretus¹] i:[eo¹] ille:[ille] iungens:[jungo] liber:[liber⁴] metris:[metrum] moenibus:[moenia¹] non:[non] nulla:[nullus] patuere:[pateo] posse:[possum] repostos:[re-pono] rerum:[res] romanarum:[romanus] sed:[sed] sinus:[sinus¹] spondeum:[spondeus] sponte:[sponte] ternis:[terni] totum:[totus¹] triumphis:[triumphus] tutae:[tueor] uersum:[verro] urbes:[urbs] †uilem:[†uilem]]

		winnerstring(&sb, slicedwords, winnermap)
	}

	return strings.TrimSpace(sb.String())
}

// flatstring - helper for buildtextblock() to generate unmodified text
func flatstring(sb *strings.Builder, slicedwords []string) {
	ls := readstopconfig("latin")
	gs := readstopconfig("greek")
	ss := append(gs, ls...)
	stops := ToSet(ss)

	for i := 0; i < len(slicedwords); i++ {
		// drop skipwords
		_, s := stops[slicedwords[i]]
		if s {
			continue
		} else {
			sb.WriteString(slicedwords[i] + " ")
		}
	}

	for i := 0; i < len(slicedwords); i++ {
		sb.WriteString(slicedwords[i] + " ")
	}
}

// montecarlostring - helper for buildtextblock() to generate lucky-ducky substitutions
func montecarlostring(sb *strings.Builder, slicedwords []string, guessermap map[string]hwguesser) {
	ls := readstopconfig("latin")
	gs := readstopconfig("greek")
	ss := append(gs, ls...)
	stops := ToSet(ss)
	var w string
	for i := 0; i < len(slicedwords); i++ {
		// pick a word...
		mc := guessermap[slicedwords[i]]
		if mc.total > 0 {
			g := rand.Intn(mc.total)
			for k, v := range mc.words {
				if k < g {
					w = v
					break
				}
			}
		} else {
			// just grab the first one
			for _, v := range mc.words {
				w = v
				break
			}
		}

		_, s := stops[w]
		if s {
			continue
		} else {
			sb.WriteString(w + " ")
		}
	}

}

// yokedstring - helper for buildtextblock() to generate conjoined string substitutions
func yokedstring(sb *strings.Builder, slicedwords []string, yokedmap map[string]string) {
	// exact same logic as winnerstring()
	winnerstring(sb, slicedwords, yokedmap)
}

// winnerstring - helper for buildtextblock() to generate winner takes all substitutions
func winnerstring(sb *strings.Builder, slicedwords []string, winnermap map[string]string) {
	ls := readstopconfig("latin")
	gs := readstopconfig("greek")
	ss := append(gs, ls...)
	stops := ToSet(ss)

	for i := 0; i < len(slicedwords); i++ {
		// drop skipwords
		w := winnermap[slicedwords[i]]
		_, s := stops[w]
		if s {
			continue
		} else {
			sb.WriteString(w + " ")
		}
	}
}

// buildyokedparsemap
func buildyokedparsemap(parsemap map[string]map[string]bool) map[string]string {
	const (
		SEPARATOR = `•`
	)
	// turn a list of sentences into a list of headwords; here we accept all headwords and yoke them
	// "esse" is "sum" + "edo", etc.

	// [a] figure out all headwords in use

	allheadwords := make(map[string]bool)
	for i := range parsemap {
		for k, _ := range parsemap[i] {
			allheadwords[k] = true
		}
	}

	// [b] note that there are capital words in the parsemap that need lowering

	// [b1] lower the internal values first
	for i := range parsemap {
		newmap := make(map[string]bool)
		for k, _ := range parsemap[i] {
			newmap[strings.ToLower(k)] = true
		}
		parsemap[i] = newmap
	}

	// [b2] lower the parsemap keys; how worried should we be about the collisions...
	lcparsemap := make(map[string]map[string]bool)
	for i := range parsemap {
		lcparsemap[strings.ToLower(i)] = parsemap[i]
	}

	// [c] build the yoked map

	yoked := make(map[string]string)
	for i := range lcparsemap {
		var ww []string
		// for j := 0; j < len(lcparsemap[i]); j++ {
		for j, _ := range parsemap[i] {
			ww = append(ww, j)
		}
		sort.Strings(ww)

		yoked[i] = strings.Join(ww, SEPARATOR)
	}

	return yoked
}

type hwguesser struct {
	total int
	words map[int]string
}

// buildmontecarloparsemap
func buildmontecarloparsemap(parsemap map[string]map[string]bool) map[string]hwguesser {
	// turn a list of sentences into a list of headwords; here we figure out the chances of any given homonym
	// then we set ourselves up to do a weighted guess of which one is in use
	// if a word might be A, B, or C and A appears 50 times, B appears 25 times, and C appears 5 times, then you
	// want to randomly assign the word to A 5/8 of the time, etc.
	// this can be acheived by a cumulative weight: [A -> 1-50, B -> 51-75, C -> 76-80]; a guess of 66 is a "B"

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

	guessermap := make(map[string]hwguesser)
	for i := range lcparsemap {
		var g hwguesser
		g.words = make(map[int]string)
		t := 0
		for j, _ := range parsemap[i] {
			t += scoremap[j]
			g.words[t] = j
		}
		g.total = t
		guessermap[i] = g
	}

	return guessermap
}

// buildwinnertakesallparsemap - figure out which is the most common of the possible headwords for any given word
func buildwinnertakesallparsemap(parsemap map[string]map[string]bool) map[string]string {
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

	winnermap := make(map[string]string)
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
		winnermap[i] = hwl[0].Word
	}

	return winnermap
}

// fetchheadwordcounts - map a list of headwords to their corpus counts
func fetchheadwordcounts(headwordset map[string]bool) map[string]int {
	const (
		MSG1 = "fetchheadwordcounts() will search for %d headwords"
	)
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

	msg(fmt.Sprintf(MSG1, len(headwordset)), MSGPEEK)

	dbconn := GetPSQLconnection()
	defer dbconn.Release()

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

//
// WEGO NOTES AND DEFAULTS
//

// w2vvectorconfig - read the CONFIGVECTORW2V file and return word2vec.Options
func w2vvectorconfig() word2vec.Options {
	const (
		ERR1 = "w2vvectorconfig() cannot find UserHomeDir"
		ERR2 = "w2vvectorconfig() failed to parse "
		MSG1 = "wrote default vector configuration file "
		MSG2 = "read vector configuration from "
	)

	// cfg := word2vec.DefaultOptions()
	cfg := DefaultW2VVectors
	cfg.Goroutines = runtime.NumCPU()

	h, e := os.UserHomeDir()
	if e != nil {
		msg(ERR1, 0)
		return cfg
	}

	_, yes := os.Stat(fmt.Sprintf(CONFIGALTAPTH, h) + CONFIGVECTORW2V)

	if yes != nil {
		content, err := json.MarshalIndent(cfg, JSONINDENT, JSONINDENT)
		chke(err)

		err = os.WriteFile(fmt.Sprintf(CONFIGALTAPTH, h)+CONFIGVECTORW2V, content, WRITEPERMS)
		chke(err)
		msg(MSG1+CONFIGVECTORW2V, MSGPEEK)
	} else {
		loadedcfg, _ := os.Open(fmt.Sprintf(CONFIGALTAPTH, h) + CONFIGVECTORW2V)
		decoderc := json.NewDecoder(loadedcfg)
		vc := word2vec.Options{}
		errc := decoderc.Decode(&vc)
		_ = loadedcfg.Close()
		if errc != nil {
			msg(ERR2+CONFIGVECTORW2V, MSGCRIT)
			cfg = DefaultW2VVectors
		}
		msg(MSG2+CONFIGVECTORW2V, MSGTMI)
		cfg = vc
	}

	return cfg
}

// lexvecvectorconfig() - read the CONFIGVECTORW2V file and return word2vec.Options
func lexvecvectorconfig() lexvec.Options {
	const (
		ERR1 = "w2vvectorconfig() cannot find UserHomeDir"
		ERR2 = "w2vvectorconfig() failed to parse "
		MSG1 = "wrote default vector configuration file "
		MSG2 = "read vector configuration from "
	)

	// cfg := lexvec.DefaultOptions()
	cfg := DefaultLexVecVectors
	cfg.Goroutines = runtime.NumCPU()

	h, e := os.UserHomeDir()
	if e != nil {
		msg(ERR1, 0)
		return cfg
	}

	_, yes := os.Stat(fmt.Sprintf(CONFIGALTAPTH, h) + CONFIGVECTORLEXVEX)

	if yes != nil {
		content, err := json.MarshalIndent(cfg, JSONINDENT, JSONINDENT)
		chke(err)

		err = os.WriteFile(fmt.Sprintf(CONFIGALTAPTH, h)+CONFIGVECTORLEXVEX, content, WRITEPERMS)
		chke(err)
		msg(MSG1+CONFIGVECTORLEXVEX, MSGPEEK)
	} else {
		loadedcfg, _ := os.Open(fmt.Sprintf(CONFIGALTAPTH, h) + CONFIGVECTORLEXVEX)
		decoderc := json.NewDecoder(loadedcfg)
		vc := lexvec.Options{}
		errc := decoderc.Decode(&vc)
		_ = loadedcfg.Close()
		if errc != nil {
			msg(ERR2+CONFIGVECTORLEXVEX, MSGCRIT)
			cfg = DefaultLexVecVectors
		}
		msg(MSG2+CONFIGVECTORLEXVEX, MSGTMI)
		cfg = vc
	}

	return cfg
}

// glovevectorconfig() - read the CONFIGVECTORW2V file and return word2vec.Options
func glovevectorconfig() glove.Options {
	const (
		ERR1 = "w2vvectorconfig() cannot find UserHomeDir"
		ERR2 = "w2vvectorconfig() failed to parse "
		MSG1 = "wrote default vector configuration file "
		MSG2 = "read vector configuration from "
	)

	// cfg := glove.DefaultOptions()
	cfg := DefaultGloveVectors
	cfg.Goroutines = runtime.NumCPU()

	h, e := os.UserHomeDir()
	if e != nil {
		msg(ERR1, 0)
		return cfg
	}

	_, yes := os.Stat(fmt.Sprintf(CONFIGALTAPTH, h) + CONFIGVECTORGLOVE)

	if yes != nil {
		content, err := json.MarshalIndent(cfg, JSONINDENT, JSONINDENT)
		chke(err)

		err = os.WriteFile(fmt.Sprintf(CONFIGALTAPTH, h)+CONFIGVECTORGLOVE, content, WRITEPERMS)
		chke(err)
		msg(MSG1+CONFIGVECTORGLOVE, MSGPEEK)
	} else {
		loadedcfg, _ := os.Open(fmt.Sprintf(CONFIGALTAPTH, h) + CONFIGVECTORGLOVE)
		decoderc := json.NewDecoder(loadedcfg)
		vc := glove.Options{}
		errc := decoderc.Decode(&vc)
		_ = loadedcfg.Close()
		if errc != nil {
			msg(ERR2+CONFIGVECTORGLOVE, MSGCRIT)
			cfg = DefaultGloveVectors
		}
		msg(MSG2+CONFIGVECTORGLOVE, MSGTMI)
		cfg = vc
	}

	return cfg
}

// readstopconfig - read the CONFIGVECTORSTOP file and return []stopwords; if it does not exist, generate it
func readstopconfig(fn string) []string {
	const (
		ERR1 = "readstopconfig() cannot find UserHomeDir"
		ERR2 = "readstopconfig() failed to parse "
		MSG1 = "readstopconfig() wrote vector stop configuration file: "
		MSG2 = "readstopconfig() read vector stop configuration from: "
	)

	var stops []string
	var vcfg string

	switch fn {
	case "latin":
		vcfg = CONFIGVECTORSTOPSLAT
		stops = StringMapKeysIntoSlice(getlatinstops())
	case "greek":
		vcfg = CONFIGVECTORSTOPSGRK
		stops = StringMapKeysIntoSlice(getgreekstops())
	}

	h, e := os.UserHomeDir()
	if e != nil {
		msg(ERR1, 0)
		return stops
	}

	_, yes := os.Stat(fmt.Sprintf(CONFIGALTAPTH, h) + vcfg)

	if yes != nil {
		sort.Strings(stops)
		content, err := json.MarshalIndent(stops, JSONINDENT, JSONINDENT)
		chke(err)

		err = os.WriteFile(fmt.Sprintf(CONFIGALTAPTH, h)+vcfg, content, WRITEPERMS)
		chke(err)
		msg(MSG1+vcfg, MSGPEEK)
	} else {
		loadedcfg, _ := os.Open(fmt.Sprintf(CONFIGALTAPTH, h) + vcfg)
		decoderc := json.NewDecoder(loadedcfg)
		var stp []string
		errc := decoderc.Decode(&stp)
		_ = loadedcfg.Close()
		if errc != nil {
			msg(ERR2+vcfg, MSGCRIT)
		} else {
			stops = stp
		}
		msg(MSG2+vcfg, MSGTMI)
	}
	return stops
}

//
// DB INTERACTION
//

// vectordbinit - initialize VECTORTABLENAME
func vectordbinit(dbconn *pgxpool.Conn) {
	const (
		CREATE = `
			CREATE TABLE %s
			(
			  fingerprint character(32),
			  vectorsize  int,
			  vectordata  bytea
			)`
	)
	ex := fmt.Sprintf(CREATE, VECTORTABLENAME)
	_, err := dbconn.Exec(context.Background(), ex)
	chke(err)
	msg("vectordbinit(): success", 3)
}

// vectordbcheck - has a search with this fingerprint already been stored?
func vectordbcheck(fp string) bool {
	const (
		Q = `SELECT fingerprint FROM %s WHERE fingerprint = '%s' LIMIT 1`
	)
	dbconn := GetPSQLconnection()
	defer dbconn.Release()

	q := fmt.Sprintf(Q, VECTORTABLENAME, fp)
	foundrow, err := dbconn.Query(context.Background(), q)
	if err != nil {
		m := err.Error()
		if strings.Contains(m, "does not exist") {
			vectordbinit(dbconn)
		}
	}
	return foundrow.Next()
}

// vectordbadd - add a set of embeddings to VECTORTABLENAME
func vectordbadd(fp string, embs embedding.Embeddings) {
	const (
		MSG1 = "vectordbadd(): "
		MSG2 = "%s compression: %dk -> %dk (%.1f percent)"
		INS  = `
			INSERT INTO %s
				(fingerprint, vectorsize, vectordata)
			VALUES ('%s', $1, $2)`
	)

	eb, err := json.Marshal(embs)
	chke(err)

	l1 := len(eb)

	// https://stackoverflow.com/questions/61077668/how-to-gzip-string-and-return-byte-array-in-golang
	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, gzip.DefaultCompression)
	chke(err)
	_, err = zw.Write(eb)
	chke(err)
	err = zw.Close()
	chke(err)

	b := buf.Bytes()
	l2 := len(b)

	ex := fmt.Sprintf(INS, VECTORTABLENAME, fp)

	dbconn := GetPSQLconnection()
	defer dbconn.Release()

	_, err = dbconn.Exec(context.Background(), ex, l2, b)
	chke(err)
	msg(MSG1+fp, MSGPEEK)

	// compressed is c. 28% of original
	msg(fmt.Sprintf(MSG2, fp, l1/1024, l2/1024, (float32(l2)/float32(l1))*100), MSGPEEK)
}

// vectordbfetch - get a set of embeddings from VECTORTABLENAME
func vectordbfetch(fp string) embedding.Embeddings {
	const (
		MSG1 = "vectordbfetch(): "
		Q    = `SELECT vectordata FROM %s WHERE fingerprint = '%s' LIMIT 1`
	)
	dbconn := GetPSQLconnection()
	defer dbconn.Release()

	q := fmt.Sprintf(Q, VECTORTABLENAME, fp)
	var vect []byte
	foundrow, err := dbconn.Query(context.Background(), q)
	chke(err)

	defer foundrow.Close()
	for foundrow.Next() {
		err = foundrow.Scan(&vect)
		chke(err)
	}

	var buf bytes.Buffer
	buf.Write(vect)

	// unzip
	zr, err := gzip.NewReader(&buf)
	chke(err)
	err = zr.Close()
	chke(err)
	decompr, err := io.ReadAll(zr)
	chke(err)

	var emb embedding.Embeddings
	err = json.Unmarshal(decompr, &emb)
	chke(err)

	msg(MSG1+fp, MSGFYI)

	return emb
}

// vectordbreset - drop VECTORTABLENAME
func vectordbreset() {
	const (
		MSG1 = "vectordbreset() dropped "
		MSG2 = "vectordbreset(): 'DROP TABLE %s' returned an (ignored) error"
		E    = `DROP TABLE %s`
	)
	ex := fmt.Sprintf(E, VECTORTABLENAME)
	dbconn := GetPSQLconnection()
	defer dbconn.Release()

	_, err := dbconn.Exec(context.Background(), ex)
	if err != nil {
		msg(fmt.Sprintf(MSG2, VECTORTABLENAME), MSGFYI)
	} else {
		msg(MSG1+VECTORTABLENAME, MSGFYI)
	}
}
