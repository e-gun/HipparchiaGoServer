//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vv

const (
	CONFIGVECTORW2V       = "vector-conf-w2v.json"
	CONFIGVECTORGLOVE     = "vector-conf-glove.json"
	CONFIGVECTORLEXVEC    = "vector-conf-lexvec.json"
	CONFIGVECTORLDA       = "vector-conf-lda.json"
	CONFIGVECTORSTOPSLAT  = "vector-stops-latin.json"
	CONFIGVECTORSTOPSGRK  = "vector-stops-greek.json"
	DEFAULTCHRTWIDTH      = "1500px"
	DEFAULTCHRTHEIGHT     = "1200px"
	LDATOPICS             = 8
	LDAMAXTOPICS          = 30
	LDASENTPERBAG         = 1
	LDAITER               = 200
	LDAXFORMPASSES        = 100
	LDABURNINPASSES       = 2
	LDACHGEVALFRQ         = 10
	LDAPERPEVALFRQ        = 10
	LDAPERPTOL            = 1e-2
	LDAMAXGRAPHLINES      = 30000
	VECTORNEIGHBORS       = 16
	VECTORNEIGHBORSMAX    = 40
	VECTORNEIGHBORSMIN    = 4
	VECTORTABLENAMENN     = "semantic_vectors_nn"
	VECTORMAXLINES        = 2000000 // 964403 lines will get you all of Latin
	VECTORMODELDEFAULT    = "w2v"
	VECTORTEXTPREPDEFAULT = "winner"
	VECTORWEBEXTDEFAULT   = false
)
