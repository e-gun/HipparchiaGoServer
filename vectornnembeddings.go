//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/e-gun/wego/pkg/embedding"
	"github.com/e-gun/wego/pkg/model"
	"github.com/e-gun/wego/pkg/model/glove"
	"github.com/e-gun/wego/pkg/model/lexvec"
	"github.com/e-gun/wego/pkg/model/modelutil/vector"
	"github.com/e-gun/wego/pkg/model/word2vec"
	"github.com/e-gun/wego/pkg/search"
	"github.com/labstack/echo/v4"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"io"
	"strings"
	"time"
)

//
// FLOW:
// 	generateneighborsdata() which means you need to...
//  	generateembeddings() which relies upon...
//		buildtextblock() with help of ...
//		buildparsemap() data
//

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
func generateneighborsdata(c echo.Context, s SearchStruct) map[string]search.Neighbors {
	const (
		FMSG  = `Fetching a stored model`
		GMSG  = `Generating a model`
		FAIL1 = "generateneighborsdata() could not find neighbors of a neighbor: '%s' neighbors (via '%s')"
		FAIL2 = "generateneighborsdata() failed to produce a Searcher"
		FAIL3 = "generateneighborsdata() failed to yield Neighbors"
		MQMEG = `Querying the model`
	)

	fp := fingerprintnnvectorsearch(s)
	isstored := vectordbchecknn(fp)
	var embs embedding.Embeddings
	if isstored {
		s.ExtraMsg = FMSG
		AllSearches.UpdateSS(s)
		embs = vectordbfetchnn(fp)
	} else {
		s.ExtraMsg = GMSG
		AllSearches.UpdateSS(s)
		embs = generateembeddings(c, s.VecModeler, s)
		vectordbaddnn(fp, embs)
		vectordbsizenn(MSGPEEK)
	}

	// [b] make a query against the model

	// len(s.Results) is zero, so it is OK to UpdateSS() without copying 500k lines
	s.ExtraMsg = MQMEG
	AllSearches.UpdateSS(s)

	searcher, err := search.New(embs...)
	if err != nil {
		msg(FAIL2, MSGFYI)
		searcher = func() *search.Searcher { return &search.Searcher{} }()
	}

	se := s.StoredSession
	ncount := se.VecNeighbCt // how many neighbors to output; min is 1
	if ncount < VECTORNEIGHBORSMIN || ncount > VECTORNEIGHBORSMAX {
		ncount = VECTORNEIGHBORS
	}

	word := s.LemmaOne
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

	AllSearches.Purge(s.ID)
	return nn
}

// generateembeddings - turn a search into a collection of semantic vector embeddings
func generateembeddings(c echo.Context, modeltype string, s SearchStruct) embedding.Embeddings {
	const (
		FAIL1  = "model initialization failed"
		FAIL2  = "generateembeddings() failed to train vector embeddings"
		FAIL3  = "generateembeddings() failed to save vector embeddings"
		FAIL4  = "generateembeddings() failed to load vector embeddings"
		MSG1   = "generateembeddings() gathered %d lines"
		MSG2   = "generateembeddings() successfuly trained a %s model (%ss)"
		PRLMSG = `Acquiring the raw data`
		TBMSG  = `Turning %d lines into a unified text block`
		VMSG   = `Training run <code>#%d</code> out of <code>%d</code> total iterations.`
		DBMSG  = `Storing the model in the database. Then fetching it again.`
	)

	// vectorbot sends a search with pre-generated results:
	// lack of a real session means we can't call readUUIDCookie() repeatedly
	// this also means we need the "modeltype" parameter as well (bot: configtype; surfer: sessiontype)
	start := time.Now()

	s.ExtraMsg = PRLMSG
	AllSearches.UpdateSS(s)
	var vs SearchStruct
	p := message.NewPrinter(language.English)

	// vectorbot already has s.Results vs normal user who does not
	if len(s.Results) == 0 {
		vs = sessionintobulksearch(c, Config.VectorMaxlines)
		msg(fmt.Sprintf(MSG1, len(vs.Results)), MSGPEEK)
		s.Results = vs.Results
		s.ExtraMsg = p.Sprintf(TBMSG, len(vs.Results))
		vs.Results = []DbWorkline{}
	}

	AllSearches.UpdateSS(s)

	thetext := buildtextblock(s.VecTextPrep, s.Results)
	s.Results = []DbWorkline{}

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
			t := fmt.Sprintf("%.3f", time.Now().Sub(start).Seconds())
			msg(fmt.Sprintf(MSG2, modeltype, t), MSGTMI)
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
			s.ExtraMsg = fmt.Sprintf(VMSG, in, ti)
			AllSearches.UpdateSS(s)
			time.Sleep(WSPOLLINGPAUSE)
			if !s.IsActive {
				break
			}
		}
	}

	go getreport()

	_ = <-finished

	s.ExtraMsg = DBMSG
	s.IsActive = false
	AllSearches.UpdateSS(s)

	// use buffers; skip the disk; psql used for storage: vectordbaddnn() & vectordbfetchnn()
	var buf bytes.Buffer
	w := io.Writer(&buf)
	err := vmodel.Save(w, vector.Agg)
	if err != nil {
		msg(FAIL3, MSGNOTE)

	}

	r := io.Reader(&buf)
	var embs embedding.Embeddings
	embs, err = embedding.Load(r)
	if err != nil {
		msg(FAIL4, MSGNOTE)
		embs = embedding.Embeddings{}
	}

	// msg(MSG3, MSGTMI)

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

	morphmapstrslc := buildmorphmapstrslc(slicedwords, morphmapdbm)

	// "morphmapstrslc" for Albinus , poet. [lt2002]
	// map[abscondere:map[abscondo:true] apte:map[apte:true aptus:true] capitolia:map[capitolium:true] celsa:map[celsus¹:true] concludere:map[concludo:true] cui:map[quis²:true quis¹:true qui²:true qui¹:true] dactylum:map[dactylus:true] de:map[de:true] deum:map[deus:true] fieri:map[fio:true] freta:map[fretum:true fretus¹:true] i:map[eo¹:true] ille:map[ille:true] iungens:map[jungo:true] liber:map[liber¹:true liber⁴:true libo¹:true] metris:map[metrum:true] moenibus:map[moenia¹:true] non:map[non:true] nulla:map[nullus:true] patuere:map[pateo:true patesco:true] posse:map[possum:true] repostos:map[re-pono:true] rerum:map[res:true] romanarum:map[romanus:true] sed:map[sed:true] sinus:map[sinus¹:true] spondeum:map[spondeum:true spondeus:true] sponte:map[sponte:true] ternis:map[terni:true] totum:map[totus²:true totus¹:true] triumphis:map[triumphus:true] tutae:map[tueor:true] uersum:map[verro:true versum:true versus³:true verto:true] urbes:map[urbs:true] †uilem:map[†uilem:true]]
	//

	// [d] turn results into unified text block

	// string addition will use a huge amount of time: 120s to concatinate Cicero: txt = txt + newtxt...
	// with strings.Builder we only need .1s to build the text...

	var sb strings.Builder
	preallocate := CHARSPERLINE * len(lines) // NB: a long line has 60 chars
	sb.Grow(preallocate)
	stops := getstopset()
	switch method {
	case "unparsed":
		flatstring(&sb, slicedwords)
	case "montecarlo":
		mcm := buildmontecarloparsemap(morphmapstrslc)

		// "mcm" for Albinus , poet. [lt2002]
		// map[abscondere:{213 map[213:abscondo]} apte:{1591 map[168:apte 1591:aptus]} capitolia:{0 map[0:capitolium]} celsa:{1050 map[1050:celsus¹]} concludere:{353 map[353:concludo]} cui:{324175 map[0:quis² 251744:qui¹ 271556:qui² 324175:quis¹]} dactylum:{167 map[167:dactylus]} de:{42695 map[42695:de]} deum:{14899 map[14899:deus]} fieri:{12305 map[12305:fio]} freta:{1507 map[746:fretum 1507:fretus¹]} i:{58129 map[58129:eo¹]} ille:{44214 map[44214:ille]} iungens:{2275 map[2275:jungo]} liber:{24949 map[7550:liber¹ 20953:liber⁴ 24949:libo¹]} metris:{383 map[383:metrum]} moenibus:{1308 map[1308:moenia¹]} non:{96475 map[96475:non]} nulla:{11785 map[11785:nullus]} patuere:{1874 map[1828:pateo 1874:patesco]} posse:{41631 map[41631:possum]} repostos:{47 map[47:re-pono]} rerum:{38669 map[38669:res]} romanarum:{0 map[0:romanus]} sed:{44131 map[44131:sed]} sinus:{1223 map[1223:sinus¹]} spondeum:{363 map[158:spondeum 363:spondeus]} sponte:{841 map[841:sponte]} ternis:{591 map[591:terni]} totum:{9166 map[0:totus² 9166:totus¹]} triumphis:{1058 map[1058:triumphus]} tutae:{3734 map[3734:tueor]} uersum:{9139 map[1471:verto 5314:verro 5749:versum 9139:versus³]} urbes:{8564 map[8564:urbs]} †uilem:{0 map[0:†uilem]}]

		montecarlostring(&sb, slicedwords, mcm, stops)
	case "yoked":
		yokedmap := buildyokedparsemap(morphmapstrslc)

		// "yokedmap" for Albinus , poet. [lt2002]
		// map[abscondere:abscondo apte:apte•aptus capitolia:capitolium celsa:celsus¹ concludere:concludo cui:quis²•quis¹•qui²•qui¹ dactylum:dactylus de:de deum:deus fieri:fio freta:fretum•fretus¹ i:eo¹ ille:ille iungens:jungo liber:liber¹•liber⁴•libo¹ metris:metrum moenibus:moenia¹ non:non nulla:nullus patuere:pateo•patesco posse:possum repostos:re-pono rerum:res romanarum:romanus sed:sed sinus:sinus¹ spondeum:spondeum•spondeus sponte:sponte ternis:terni totum:totus²•totus¹ triumphis:triumphus tutae:tueor uersum:verro•versum•versus³•verto urbes:urbs †uilem:†uilem]

		yokedstring(&sb, slicedwords, yokedmap, stops)
	default: // "winner"
		winnermap := buildwinnertakesallparsemap(morphmapstrslc)

		// "winnermap" for Albinus , poet. [lt2002]
		// map[abscondere:[abscondo] apte:[aptus] capitolia:[capitolium] celsa:[celsus¹] concludere:[concludo] cui:[qui¹] dactylum:[dactylus] de:[de] deum:[deus] fieri:[fio] freta:[fretus¹] i:[eo¹] ille:[ille] iungens:[jungo] liber:[liber⁴] metris:[metrum] moenibus:[moenia¹] non:[non] nulla:[nullus] patuere:[pateo] posse:[possum] repostos:[re-pono] rerum:[res] romanarum:[romanus] sed:[sed] sinus:[sinus¹] spondeum:[spondeus] sponte:[sponte] ternis:[terni] totum:[totus¹] triumphis:[triumphus] tutae:[tueor] uersum:[verro] urbes:[urbs] †uilem:[†uilem]]

		winnerstring(&sb, slicedwords, winnermap, stops)
	}

	return strings.TrimSpace(sb.String())
}

// buildmorphmapstrslc - a map that lets you convert words into headwords
func buildmorphmapstrslc(slicedwords []string, morphmapdbm map[string]DbMorphology) map[string]map[string]bool {
	// figure out which headwords to associate with the collection of words
	//
	//	// this information is inside DbMorphology.RawPossib
	//	// but it needs to be parsed
	//	// example: `{"1": {"transl": "A. nom. plur; II. a guardian god", "analysis": "masc gen pl", "headword": "deus", "scansion": "deūm", "xref_kind": "9", "xref_value": "22568216"}, "2": {"transl": "A. nom. plur; II. a guardian god", "analysis": "masc acc sg", "headword": "deus", "scansion": "", "xref_kind": "9", "xref_value": "22568216"}}`

	const (
		FAIL1 = "failed to unmarshal %s into objmap\n"
		FAIL2 = "failed second pass unmarshal of %s into newmap\n"
	)

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

	return morphmapstrslc
}
