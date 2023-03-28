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
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/ynqa/wego/pkg/embedding"
	"github.com/ynqa/wego/pkg/model/modelutil/vector"
	"github.com/ynqa/wego/pkg/model/word2vec"
	"github.com/ynqa/wego/pkg/search"
	"io"
	"os"
	"sort"
	"strings"
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
		"usquam", "quoque", "duo", "talis", "simul", "igitur", "utique²", "aliqui", "apud", "sic", "umquam"}
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
		"ἄνωθεν", "ὀκτώ", "ἓξ", "μετ", "τ", "μ", "αὐτόθ", "οὐδ", "εἵνεκ", "νόϲφι"}
	GreekStop = append(Greek150, GreekExtra...)
	// GreekKeep - members of GreekStop we will not toss
	GreekKeep = []string{"ἔχω", "λέγω¹", "θεόϲ", "φημί", "ποιέω", "ἵημι", "μόνοϲ", "κύριοϲ", "πόλιϲ", "θεάομαι", "δοκέω", "λαμβάνω",
		"δίδωμι", "βαϲιλεύϲ", "φύϲιϲ", "ἔτοϲ", "πατήρ", "ϲῶμα", "καλέω", "ἐρῶ", "υἱόϲ", "γαῖα", "ἀνήρ", "ὁράω",
		"ψυχή", "δύναμαι", "ἀρχή", "καλόϲ", "δύναμιϲ", "ἀγαθόϲ", "οἶδα", "δείκνυμι", "χρόνοϲ", "γράφω", "δραχμή",
		"μέροϲ"}
	LatinStops     = getlatinstops()
	GreekStops     = getgreekstops()
	DefaultVectors = word2vec.Options{
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
		MSG1  = "generateneighborsdata(): fetching stored embeddings"
		FAIL1 = "generateneighborsdata() could not find neighbors of a neighbor: '%s' neighbors (via '%s')"
		FAIL2 = "generateneighborsdata() failed to produce a Searcher"
		FAIL3 = "generateneighborsdata() failed to yield Neighbors"
	)

	fp := fingerprintvectorsearch(srch)
	isstored := vectordbcheck(fp)
	var embs embedding.Embeddings
	if isstored {
		msg(MSG1, MSGPEEK)
		embs = vectordbfetch(fp)
	} else {
		embs = generateembeddings(c, srch)
		vectordbadd(fp, embs)
	}

	// [b] make a query against the model
	searcher, err := search.New(embs...)
	if err != nil {
		msg(FAIL2, MSGFYI)
		searcher = func() *search.Searcher { return &search.Searcher{} }()
	}

	ncount := VECTORNEIGHBORS // how many neighbors to output; min is 1
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
func generateembeddings(c echo.Context, srch SearchStruct) embedding.Embeddings {
	const (
		FAIL1 = "word2vec model initialization failed"
		FAIL2 = "generateembeddings() failed to train vector embeddings"
		MSG1  = "generateembeddings() gathered %d lines"
	)

	vs := sessionintobulksearch(c, VECTORMAXLINES)
	srch.Results = vs.Results
	vs.Results = []DbWorkline{}
	msg(fmt.Sprintf(MSG1, len(srch.Results)), MSGFYI)

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

	msg(fmt.Sprintf(MSG1, len(headwordset)), MSGFYI)

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

// vectorconfig - read the CONFIGVECTOR file and return word2vec.Options
func vectorconfig() word2vec.Options {
	const (
		ERR1 = "vectorconfig() cannot find UserHomeDir"
		ERR2 = "vectorconfig() failed to parse "
		MSG1 = "wrote default vector configuration file "
		MSG2 = "read vector configuration from "
	)

	// cfg := word2vec.DefaultOptions()
	cfg := DefaultVectors

	h, e := os.UserHomeDir()
	if e != nil {
		msg(ERR1, 0)
		return cfg
	}

	_, yes := os.Stat(fmt.Sprintf(CONFIGALTAPTH, h) + CONFIGVECTOR)

	if yes != nil {
		content, err := json.MarshalIndent(cfg, JSONINDENT, JSONINDENT)
		chke(err)

		err = os.WriteFile(fmt.Sprintf(CONFIGALTAPTH, h)+CONFIGVECTOR, content, WRITEPERMS)
		chke(err)
		msg(MSG1+CONFIGVECTOR, MSGPEEK)
	} else {
		loadedcfg, _ := os.Open(fmt.Sprintf(CONFIGALTAPTH, h) + CONFIGVECTOR)
		decoderc := json.NewDecoder(loadedcfg)
		vc := word2vec.Options{}
		errc := decoderc.Decode(&vc)
		_ = loadedcfg.Close()
		if errc != nil {
			msg(ERR2+CONFIGVECTOR, MSGCRIT)
			cfg = DefaultVectors
		}
		msg(MSG2+CONFIGVECTOR, MSGTMI)
		cfg = vc
	}

	return cfg
}

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

// see also: https://link.springer.com/article/10.1007/s41019-019-0096-6

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
	msg(MSG1+fp, MSGFYI)

	// the savings is real: compressed is c. 27% of original
	msg(fmt.Sprintf("vector compression: %dk -> %dk (%.1f percent)", l1/1024, l2/1024, (float32(l2)/float32(l1))*100), 3)
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
