//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vv

const (
	CONFIGVECTORW2V       = "hgs-vector-conf-w2v.json"
	CONFIGVECTORGLOVE     = "hgs-vector-conf-glove.json"
	CONFIGVECTORLEXVEC    = "hgs-vector-conf-lexvec.json"
	CONFIGVECTORLDA       = "hgs-vector-conf-lda.json"
	CONFIGVECTORSTOPSLAT  = "hgs-vector-stops-latin.json"
	CONFIGVECTORSTOPSGRK  = "hgs-vector-stops-greek.json"
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
	VECTORTABLENAMELDA    = "semantic_vectors_lda"
	VECTORMAXLINES        = 1000000 // 964403 lines will get you all of Latin
	VECTORMODELDEFAULT    = "w2v"
	VECTORTEXTPREPDEFAULT = "winner"
	VECTORWEBEXTDEFAULT   = false
)
