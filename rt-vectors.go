package main

import (
	"bytes"
	"github.com/labstack/echo/v4"
	"github.com/ynqa/wego/pkg/embedding"
	"github.com/ynqa/wego/pkg/model/modelutil/vector"
	"github.com/ynqa/wego/pkg/model/word2vec"
	"github.com/ynqa/wego/pkg/search"
	"os"
	"strings"
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

func RtVectors(c echo.Context) error {

	// test via:
	// curl 127.0.0.1:8000/vect/exec/1

	// "testing one two three" will yield:
	// testing 0.011758 0.035748 0.022445 -0.048132 0.070287 -0.007568 0.089921 -0.003403 -0.002596 -0.036771
	//one -0.023640 0.004734 -0.022185 -0.035968 0.049385 -0.038856 -0.049740 0.065040 -0.025132 0.035743
	//two 0.022721 0.016821 -0.066850 -0.092613 0.001636 -0.009990 0.039538 0.011934 0.051826 -0.008637
	//three -0.040263 -0.056876 -0.010872 -0.032923 0.038590 0.065175 -0.041002 -0.009709 -0.037445 -0.025513
	//0.0.0

	// ultimately this is *a lot* like an indexmaker search
	// grab all lines
	// figure out all the headwords
	// swap headwords in...
	// merge code later

	srch := sessionintobulksearch(c, MAXTEXTLINEGENERATION)

	// turn results into unified text block

	// string addition will us a huge amount of time: 120s to concatinate Cicero: txt = txt + newtxt...
	// with strings.Builder we only need .1s to build the text...
	var sb strings.Builder
	preallocate := CHARSPERLINE * len(srch.Results) // NB: a long line has 60 chars
	sb.Grow(preallocate)

	for i := 0; i < len(srch.Results); i++ {
		s := strings.TrimSpace(srch.Results[i].Accented)
		sb.WriteString(s + " ")
	}

	// ArrayToGetRequiredMorphobjects(wordlist)

	// fyi
	opts := word2vec.Options{
		BatchSize:          0,
		Dim:                0,
		DocInMemory:        false,
		Goroutines:         0,
		Initlr:             0,
		Iter:               0,
		LogBatch:           0,
		MaxCount:           0,
		MaxDepth:           0,
		MinCount:           0,
		MinLR:              0,
		ModelType:          "cbow",
		NegativeSampleSize: 0,
		OptimizerType:      "ns",
		SubsampleThreshold: 0,
		ToLower:            false,
		UpdateLRBatch:      0,
		Verbose:            false,
		Window:             0,
	}

	opts = word2vec.DefaultOptions()

	model, err := word2vec.NewForOptions(opts)
	if err != nil {
		// problem
	}

	// input for  word2vec.Train() is 'io.ReadSeeker'
	b := bytes.NewReader([]byte(sb.String()))
	if err = model.Train(b); err != nil {
		// failed to train.
	}

	// write word vector.

	vfile := "/Users/erik/tmp/vect.out"
	rank := 10
	word := "dextra"

	f, err := os.Create(vfile)
	if err != nil {
		msg("failed to create vect.out", 0)
	}
	err = model.Save(f, vector.Agg)
	if err != nil {
		msg("failed to save vect.out", 0)
	}

	input, err := os.Open(vfile)
	if err != nil {
		return err
	}
	defer input.Close()
	embs, err := embedding.Load(input)
	if err != nil {
		return err
	}

	searcher, err := search.New(embs...)
	if err != nil {
		return err
	}
	neighbors, err := searcher.SearchInternal(word, rank)
	if err != nil {
		return err
	}
	neighbors.Describe()

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

	return emptyjsreturn(c)
}
