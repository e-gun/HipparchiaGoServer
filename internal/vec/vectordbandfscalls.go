//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vec

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/e-gun/wego/pkg/embedding"
	"github.com/e-gun/wego/pkg/model/glove"
	"github.com/e-gun/wego/pkg/model/lexvec"
	"github.com/e-gun/wego/pkg/model/word2vec"
	"github.com/jackc/pgx/v5"
	"io"
	"os"
	"runtime"
	"strings"
)

var (
	dbi = lnch.NewMessageMakerWithDefaults()
	Msg = lnch.NewMessageMakerWithDefaults()
)

var (
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
		ModelType:          "skipgram", // "cbow" and "skipgram" available; "cbow" results are not so hot
		NegativeSampleSize: 5,
		OptimizerType:      "hs",
		SubsampleThreshold: 0.001,
		ToLower:            false,
		UpdateLRBatch:      100000,
		Verbose:            false,
		Window:             8,
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
		RelationType:       "ppmi", // "ppmi", "pmi", "co", "logco" are available; "co" will fail to model
		Smooth:             0.75,
		SubsampleThreshold: 1.0e-3,
		ToLower:            false,
		UpdateLRBatch:      100000,
		Verbose:            false,
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
		Verbose:            false,
		Window:             8,
		Xmax:               90,
	}

	DefaultLDAVectors = LDAConfig{
		SentencesPerBag: vv.LDASENTPERBAG,
		LDAIterations:   vv.LDAITER,
		LDAXformPasses:  vv.LDAXFORMPASSES,
		BurnInPasses:    vv.LDABURNINPASSES,
		ChangeEvalFrq:   vv.LDACHGEVALFRQ,
		PerplexEvalFrq:  vv.LDAPERPEVALFRQ,
		PerplexTol:      vv.LDAPERPTOL,
		MaxLDAGraphSize: vv.LDAMAXGRAPHLINES,
	}
)

//
// DB INTERACTION
//

// fetchheadwordcounts - map a list of headwords to their corpus counts
//func fetchheadwordcounts(headwordset map[string]bool) map[string]int {
//	const (
//		MSG1 = "fetchheadwordcounts() will search for %d headwords"
//	)
//	if len(headwordset) == 0 {
//		return make(map[string]int)
//	}
//
//	tt := "CREATE TEMPORARY TABLE ttw_%s AS SELECT words AS w FROM unnest(ARRAY[%s]) words"
//	qt := "SELECT entry_name, total_count FROM dictionary_headword_wordcounts WHERE EXISTS " +
//		"(SELECT 1 FROM ttw_%s temptable WHERE temptable.w = dictionary_headword_wordcounts.entry_name)"
//
//	rndid := strings.Replace(uuid.New().String(), "-", "", -1)
//
//	hw := make([]string, 0, len(headwordset))
//	for h := range headwordset {
//		hw = append(hw, h)
//	}
//
//	Msg.PEEK(fmt.Sprintf(MSG1, len(headwordset)))
//
//	dbconn := db.GetDBConnection()
//	defer dbconn.Release()
//
//	arr := strings.Join(hw, "', '")
//	arr = fmt.Sprintf("'%s'", arr)
//
//	tt = fmt.Sprintf(tt, rndid, arr)
//	_, err := dbconn.Exec(context.Background(), tt)
//	dbi.EC(err)
//
//	qt = fmt.Sprintf(qt, rndid)
//	foundrows, e := dbconn.Query(context.Background(), qt)
//	dbi.EC(e)
//
//	returnmap := make(map[string]int)
//	defer foundrows.Close()
//	for foundrows.Next() {
//		var thehit WeightedHeadword
//		err = foundrows.Scan(&thehit.Word, &thehit.Count)
//		dbi.EC(err)
//		returnmap[thehit.Word] = thehit.Count
//	}
//
//	// don't kill off unfound terms
//	for i := range hw {
//		if _, t := returnmap[hw[i]]; t {
//			continue
//		} else {
//			returnmap[hw[i]] = 0
//		}
//	}
//
//	// "returnmap" for Albinus , poet. [lt2002]
//	// map[abscondo:213 apte:168 aptus:1423 capitolium:0 celsus¹:1050 concludo:353 dactylus:167 de:42695 deus:14899 eo¹:58129 fio:12305 fretum:746 fretus¹:761 ille:44214 jungo:2275 liber¹:7550 liber⁴:13403 libo¹:3996 metrum:383 moenia¹:1308 non:96475 nullus:11785 pateo:1828 patesco:46 possum:41631 quis²:0 quis¹:52619 qui²:19812 qui¹:251744 re-pono:47 res:38669 romanus:0 sed:44131 sinus¹:1223 spondeum:158 spondeus:205 sponte:841 terni:591 totus²:0 totus¹:9166 triumphus:1058 tueor:3734 urbs:8564 verro:3843 versum:435 versus³:3390 verto:1471 †uilem:0]
//
//	return returnmap
//}

// VectorDBInitNN - initialize vv.VECTORTABLENAMENN
func VectorDBInitNN() {
	const (
		CREATE = `
			CREATE TABLE %s
			(
			  fingerprint character(32),
			  vectorsize  int,
			  vectordata  bytea
			)`
		EXISTS = "already exists"
	)
	ex := fmt.Sprintf(CREATE, vv.VECTORTABLENAMENN)
	_, err := db.SQLPool.Exec(context.Background(), ex)
	if err != nil {
		m := err.Error()
		if !strings.Contains(m, EXISTS) {
			dbi.EC(err)
		}
	} else {
		Msg.FYI("VectorDBInitNN(): success")
	}
}

// VectorDBCheckNN - has a search with this fingerprint already been stored?
func VectorDBCheckNN(fp string) bool {
	const (
		Q   = `SELECT fingerprint FROM %s WHERE fingerprint = '%s' LIMIT 1`
		F   = `VectorDBCheckNN() found %s`
		DNE = "does not exist"
	)

	q := fmt.Sprintf(Q, vv.VECTORTABLENAMENN, fp)
	foundrow, err := db.SQLPool.Query(context.Background(), q)
	if err != nil {
		m := err.Error()
		if strings.Contains(m, DNE) {
			VectorDBInitNN()
		}
		return false
	}

	type simplestring struct {
		S string
	}

	ss, err := pgx.CollectOneRow(foundrow, pgx.RowToStructByPos[simplestring])
	if err != nil {
		// m := err.Error()
		// m will be "no rows in result set" if you did not find the fingerprint
		return false
	} else {
		Msg.TMI(fmt.Sprintf(F, ss.S))
		return true
	}
}

// VectorDBAddNN - add a set of embeddings to vv.VECTORTABLENAMENN
func VectorDBAddNN(fp string, embs embedding.Embeddings) {
	const (
		MSG1 = "VectorDBAddNN(): "
		MSG2 = "%s compression: %dM -> %dM (-> %.1f%%)"
		MSG3 = "VectorDBAddNN() was sent empty embeddings"
		FAIL = "VectorDBAddNN() failed when calling json.Marshal(embs): nothing stored"
		INS  = `
			INSERT INTO %s
				(fingerprint, vectorsize, vectordata)
			VALUES ('%s', $1, $2)`
		GZ = gzip.BestSpeed
	)

	if embs.Empty() {
		Msg.PEEK(MSG3)
		return
	}

	// json vs jsi: jsoniter.ConfigFastest, this will marshal the float with 6 digits precision (lossy)
	eb, err := json.Marshal(embs)
	if err != nil {
		Msg.NOTE(FAIL)
		eb = []byte{}
	}

	// https://stackoverflow.com/questions/61077668/how-to-gzip-string-and-return-byte-array-in-golang
	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, GZ)
	dbi.EC(err)
	_, err = zw.Write(eb)
	dbi.EC(err)
	err = zw.Close()
	dbi.EC(err)

	b := buf.Bytes()
	l2 := len(b)

	ex := fmt.Sprintf(INS, vv.VECTORTABLENAMENN, fp)

	_, err = db.SQLPool.Exec(context.Background(), ex, l2, b)
	dbi.EC(err)
	Msg.TMI(MSG1 + fp)

	// compressed is c. 33% of original
	// l1 := len(eb)
	// m(fmt.Sprintf(MSG2, fp, l1/1024/1024, l2/1024/1024, (float32(l2)/float32(l1))*100), MSGTMI)
	buf.Reset()
}

// VectorDBFetchNN - get a set of embeddings from vv.VECTORTABLENAMENN
func VectorDBFetchNN(fp string) embedding.Embeddings {
	const (
		MSG1 = "VectorDBFetchNN(): "
		MSG2 = "VectorDBFetchNN() pulled empty set of embeddings for %s"
		Q    = `SELECT vectordata FROM %s WHERE fingerprint = '%s' LIMIT 1`
	)

	q := fmt.Sprintf(Q, vv.VECTORTABLENAMENN, fp)
	var vect []byte
	foundrow, err := db.SQLPool.Query(context.Background(), q)
	dbi.EC(err)

	defer foundrow.Close()
	for foundrow.Next() {
		err = foundrow.Scan(&vect)
		dbi.EC(err)
	}

	var buf bytes.Buffer
	buf.Write(vect)

	// the data in the tables is zipped and needs unzipping
	zr, err := gzip.NewReader(&buf)
	dbi.EC(err)
	err = zr.Close()
	dbi.EC(err)
	decompr, err := io.ReadAll(zr)
	dbi.EC(err)

	var emb embedding.Embeddings
	err = json.Unmarshal(decompr, &emb)
	dbi.EC(err)
	buf.Reset()

	if emb.Empty() {
		Msg.NOTE(fmt.Sprintf(MSG2, fp))
	}

	// m(MSG1+fp, MSGPEEK)

	return emb
}

// VectorDBReset - drop vv.VECTORTABLENAMENN
func VectorDBReset() {
	const (
		MSG1 = "VectorDBReset() dropped "
		MSG2 = "VectorDBReset(): 'DROP TABLE %s' returned an (ignored) error: \n\t%s"
		E    = `DROP TABLE %s`
	)
	ex := fmt.Sprintf(E, vv.VECTORTABLENAMENN)

	_, err := db.SQLPool.Exec(context.Background(), ex)
	if err != nil {
		ms := err.Error()
		Msg.TMI(fmt.Sprintf(MSG2, vv.VECTORTABLENAMENN, ms))
	} else {
		Msg.NOTE(MSG1 + vv.VECTORTABLENAMENN)
	}
}

// VectorDBSizeNN - how much space is the vectordb using?
func VectorDBSizeNN(priority int) {
	const (
		SZQ  = "SELECT SUM(vectorsize) AS total FROM " + vv.VECTORTABLENAMENN
		MSG4 = "Disk space used by stored vectors is currently %dMB"
	)
	var size int64

	err := db.SQLPool.QueryRow(context.Background(), SZQ).Scan(&size)
	dbi.EC(err)
	Msg.Emit(fmt.Sprintf(MSG4, size/1024/1024), priority)
}

func VectorDBCountNN(priority int) {
	const (
		SZQ  = "SELECT COUNT(vectorsize) AS total FROM " + vv.VECTORTABLENAMENN
		MSG4 = "Number of stored vector models: %d"
		DNE  = "does not exist"
	)
	var size int64

	err := db.SQLPool.QueryRow(context.Background(), SZQ).Scan(&size)
	if err != nil {
		m := err.Error()
		if strings.Contains(m, DNE) {
			VectorDBInitNN()
		}
		size = 0
	}
	Msg.Emit(fmt.Sprintf(MSG4, size), priority)
}

//
// LDA vv.CONFIGURATION
//

type LDAConfig struct {
	SentencesPerBag int
	LDAIterations   int
	LDAXformPasses  int
	BurnInPasses    int
	ChangeEvalFrq   int
	PerplexEvalFrq  int
	PerplexTol      float64
	Goroutines      int
	MaxLDAGraphSize int
}

func ldavecconfig() LDAConfig {
	const (
		ERR1 = "ldavecconfig() cannot find UserHomeDir"
		ERR2 = "ldavecconfig() failed to parse "
		MSG1 = "wrote default vector configuration file "
		MSG2 = "read vector configuration from "
	)

	cfg := DefaultLDAVectors
	cfg.Goroutines = runtime.NumCPU()

	h, e := os.UserHomeDir()
	if e != nil {
		Msg.MAND(ERR1)
		return cfg
	}

	_, yes := os.Stat(fmt.Sprintf(vv.CONFIGALTAPTH, h) + vv.CONFIGVECTORLDA)

	if yes != nil {
		content, err := json.MarshalIndent(cfg, vv.JSONINDENT, vv.JSONINDENT)
		dbi.EC(err)

		err = os.WriteFile(fmt.Sprintf(vv.CONFIGALTAPTH, h)+vv.CONFIGVECTORLDA, content, vv.WRITEPERMS)
		dbi.EC(err)
		Msg.PEEK(MSG1 + vv.CONFIGVECTORLDA)
	} else {
		loadedcfg, _ := os.Open(fmt.Sprintf(vv.CONFIGALTAPTH, h) + vv.CONFIGVECTORLDA)
		decoderc := json.NewDecoder(loadedcfg)
		vc := LDAConfig{}
		errc := decoderc.Decode(&vc)
		_ = loadedcfg.Close()
		if errc != nil {
			Msg.CRIT(ERR2 + vv.CONFIGVECTORLDA)
			vc = cfg
		}
		// m(MSG2+vv.CONFIGVECTORLDA, MSGTMI)
		cfg = vc
	}

	if cfg.MaxLDAGraphSize == 0 {
		cfg.MaxLDAGraphSize = vv.LDAMAXGRAPHLINES
	}
	return cfg
}

//
// WEGO NOTES AND DEFAULTS
//

// w2vvectorconfig - read the vv.CONFIGVECTORW2V file and return word2vec.Options
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
		Msg.MAND(ERR1)
		return cfg
	}

	_, yes := os.Stat(fmt.Sprintf(vv.CONFIGALTAPTH, h) + vv.CONFIGVECTORW2V)

	if yes != nil {
		content, err := json.MarshalIndent(cfg, vv.JSONINDENT, vv.JSONINDENT)
		dbi.EC(err)

		err = os.WriteFile(fmt.Sprintf(vv.CONFIGALTAPTH, h)+vv.CONFIGVECTORW2V, content, vv.WRITEPERMS)
		dbi.EC(err)
		Msg.PEEK(MSG1 + vv.CONFIGVECTORW2V)
	} else {
		loadedcfg, _ := os.Open(fmt.Sprintf(vv.CONFIGALTAPTH, h) + vv.CONFIGVECTORW2V)
		decoderc := json.NewDecoder(loadedcfg)
		vc := word2vec.Options{}
		errc := decoderc.Decode(&vc)
		_ = loadedcfg.Close()
		if errc != nil {
			Msg.CRIT(ERR2 + vv.CONFIGVECTORW2V)
			vc = DefaultW2VVectors
		}
		Msg.TMI(MSG2 + vv.CONFIGVECTORW2V)
		cfg = vc
	}

	return cfg
}

// lexvecvectorconfig() - read the vv.CONFIGVECTORLEXVEC file and return word2vec.Options
func lexvecvectorconfig() lexvec.Options {
	const (
		ERR1 = "lexvecvectorconfig() cannot find UserHomeDir"
		ERR2 = "lexvecvectorconfig() failed to parse "
		MSG1 = "wrote default vector configuration file "
		MSG2 = "read vector configuration from "
	)

	// cfg := lexvec.DefaultOptions()
	cfg := DefaultLexVecVectors
	cfg.Goroutines = runtime.NumCPU()

	h, e := os.UserHomeDir()
	if e != nil {
		Msg.MAND(ERR1)
		return cfg
	}

	_, yes := os.Stat(fmt.Sprintf(vv.CONFIGALTAPTH, h) + vv.CONFIGVECTORLEXVEC)

	if yes != nil {
		content, err := json.MarshalIndent(cfg, vv.JSONINDENT, vv.JSONINDENT)
		dbi.EC(err)

		err = os.WriteFile(fmt.Sprintf(vv.CONFIGALTAPTH, h)+vv.CONFIGVECTORLEXVEC, content, vv.WRITEPERMS)
		dbi.EC(err)
		Msg.PEEK(MSG1 + vv.CONFIGVECTORLEXVEC)
	} else {
		loadedcfg, _ := os.Open(fmt.Sprintf(vv.CONFIGALTAPTH, h) + vv.CONFIGVECTORLEXVEC)
		decoderc := json.NewDecoder(loadedcfg)
		vc := lexvec.Options{}
		errc := decoderc.Decode(&vc)
		_ = loadedcfg.Close()
		if errc != nil {
			Msg.CRIT(ERR2 + vv.CONFIGVECTORLEXVEC)
			vc = DefaultLexVecVectors
		}
		Msg.TMI(MSG2 + vv.CONFIGVECTORLEXVEC)
		cfg = vc
	}
	return cfg
}

// glovevectorconfig() - read the vv.CONFIGVECTORGLOVE file and return word2vec.Options
func glovevectorconfig() glove.Options {
	const (
		ERR1 = "glovevectorconfig() cannot find UserHomeDir"
		ERR2 = "glovevectorconfig() failed to parse "
		MSG1 = "wrote default vector configuration file "
		MSG2 = "read vector configuration from "
	)

	// cfg := glove.DefaultOptions()
	cfg := DefaultGloveVectors
	cfg.Goroutines = runtime.NumCPU()

	h, e := os.UserHomeDir()
	if e != nil {
		Msg.MAND(ERR1)
		return cfg
	}

	_, yes := os.Stat(fmt.Sprintf(vv.CONFIGALTAPTH, h) + vv.CONFIGVECTORGLOVE)

	if yes != nil {
		content, err := json.MarshalIndent(cfg, vv.JSONINDENT, vv.JSONINDENT)
		dbi.EC(err)

		err = os.WriteFile(fmt.Sprintf(vv.CONFIGALTAPTH, h)+vv.CONFIGVECTORGLOVE, content, vv.WRITEPERMS)
		dbi.EC(err)
		Msg.PEEK(MSG1 + vv.CONFIGVECTORGLOVE)
	} else {
		loadedcfg, _ := os.Open(fmt.Sprintf(vv.CONFIGALTAPTH, h) + vv.CONFIGVECTORGLOVE)
		decoderc := json.NewDecoder(loadedcfg)
		vc := glove.Options{}
		errc := decoderc.Decode(&vc)
		_ = loadedcfg.Close()
		if errc != nil {
			Msg.CRIT(ERR2 + vv.CONFIGVECTORGLOVE)
			vc = DefaultGloveVectors
		}
		Msg.TMI(MSG2 + vv.CONFIGVECTORGLOVE)
		cfg = vc
	}

	return cfg
}
