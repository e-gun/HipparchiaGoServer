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
	"github.com/e-gun/wego/pkg/model/glove"
	"github.com/e-gun/wego/pkg/model/lexvec"
	"github.com/e-gun/wego/pkg/model/word2vec"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"os"
	"runtime"
	"strings"
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
		Verbose:            true,
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
)

//
// DB INTERACTION
//

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
		MSG2 = "%s compression: %dM -> %dM (%.1f percent)"
		FAIL = "vectordbadd() failed when calling json.Marshal(embs): nothing stored"
		INS  = `
			INSERT INTO %s
				(fingerprint, vectorsize, vectordata)
			VALUES ('%s', $1, $2)`
		GZ = gzip.BestSpeed
	)

	eb, err := json.Marshal(embs)
	if err != nil {
		msg(FAIL, MSGNOTE)
		eb = []byte{}
	}

	l1 := len(eb)

	// https://stackoverflow.com/questions/61077668/how-to-gzip-string-and-return-byte-array-in-golang
	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, GZ)
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
	msg(fmt.Sprintf(MSG2, fp, l1/1024/1024, l2/1024/1024, (float32(l2)/float32(l1))*100), MSGPEEK)
}

// vectordbfetch - get a set of embeddings from VECTORTABLENAME
func vectordbfetch(fp string) embedding.Embeddings {
	const (
		MSG1 = "vectordbfetch(): "
		MSG2 = "vectordbfetch() pulled empty set of embeddings for %s"
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

	// the data in the tables is zipped and needs unzipping
	zr, err := gzip.NewReader(&buf)
	chke(err)
	err = zr.Close()
	chke(err)
	decompr, err := io.ReadAll(zr)
	chke(err)

	var emb embedding.Embeddings
	err = json.Unmarshal(decompr, &emb)
	chke(err)

	if emb.Empty() {
		msg(fmt.Sprintf(MSG2, fp), MSGNOTE)
	}

	msg(MSG1+fp, MSGPEEK)

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

// vectordbsize - how much space is the vectordb using?
func vectordbsize(priority int) {
	const (
		SZQ  = "SELECT SUM(vectorsize) AS total FROM semantic_vectors"
		MSG4 = "Disk space used by stored vectors is currently %dMB"
	)
	var size int64
	dbconn := GetPSQLconnection()
	defer dbconn.Release()
	err := dbconn.QueryRow(context.Background(), SZQ).Scan(&size)
	chke(err)
	msg(fmt.Sprintf(MSG4, size/1024/1024), priority)
}

func vectordbcount(priority int) {
	const (
		SZQ  = "SELECT COUNT(vectorsize) AS total FROM semantic_vectors"
		MSG4 = "Number of stored vector models: %d"
	)
	var size int64
	dbconn := GetPSQLconnection()
	defer dbconn.Release()
	err := dbconn.QueryRow(context.Background(), SZQ).Scan(&size)
	chke(err)
	msg(fmt.Sprintf(MSG4, size), priority)
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
