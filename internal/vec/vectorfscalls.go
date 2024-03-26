//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vec

import (
	"encoding/json"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/e-gun/wego/pkg/model/glove"
	"github.com/e-gun/wego/pkg/model/lexvec"
	"github.com/e-gun/wego/pkg/model/word2vec"
	"os"
	"runtime"
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
		// mm(MSG2+vv.CONFIGVECTORLDA, MSGTMI)
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
