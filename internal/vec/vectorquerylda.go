//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vec

import (
	"context"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/mps"
	"github.com/e-gun/HipparchiaGoServer/internal/search"
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/e-gun/nlp"
	"github.com/labstack/echo/v4"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gonum.org/v1/gonum/mat"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	tsnemp "github.com/e-gun/tsnemp/pkg"
)

// bagging as per the old HipparchiaGoDBHelper code: sentence by sentence; much of the code below from HipparchiaGoDBHelper
// bowman's package can also do nearest neighbour similarity searches: LinearScanIndex.Search(qv mat.Vector, k int) -> []Match
// with some (i.e., a lot of...) work the output could be fed to JS as per the python LDA visualizer

// see bottom of file for sample results

//see https://github.com/james-bowman/nlp/blob/26d441fa0ded/lda.go
//DefaultLDA = nlp.LatentDirichletAllocation{
//	Iterations:                    1000,
//	PerplexityTolerance:           1e-2,
//	PerplexityEvaluationFrequency: 30,
//	BatchSize:                     100,
//	K:                             k,
//	BurnInPasses:                  1,
//	TransformationPasses:          500,
//	MeanChangeTolerance:           1e-5,
//	ChangeEvaluationFrequency:     30,
//	Alpha:                         0.1,
//	Eta:                           0.01,
//	RhoPhi: LearningSchedule{
//		S:     10,
//		Tau:   1000,
//		Kappa: 0.9,
//	},
//	RhoTheta: LearningSchedule{
//		S:     1,
//		Tau:   10,
//		Kappa: 0.9,
//	},
//	rhoPhiT:   1,
//	rhoThetaT: 1,
//	Rnd:       rand.New(rand.NewSource(uint64(time.Now().UnixNano()))),
//	Processes: runtime.GOMAXPROCS(0),
//}

type BagWithLocus struct {
	Loc         string
	Bag         string
	ModifiedBag string
	LDAScore    float64
	Workline    str.DbWorkline
}

func (b *BagWithLocus) GetWL() {
	tb := strings.Split(b.Loc, "/")
	ln, e := strconv.Atoi(tb[2])
	if e != nil {
		Msg.NOTE("BagWithLocus.GetWL() failed to convert ascii to int")
	}
	b.Workline = db.GrabOneLine(tb[1][0:vv.LENGTHOFAUTHORID], ln)
}

// LDASearch - search via Latent Dirichlet Allocation
func LDASearch(c echo.Context, srch str.SearchStruct) error {
	const (
		LDAMSG = `Building LDA model for the current selections`
		ESM1   = "Preparing the text for modeling"
		ESM2   = "Building topic models"
		ESM3   = "Building the graph (please be patient this can be very slow...)"
	)
	c.Response().After(func() { Msg.LogPaths("LDASearch()") })

	se := srch.StoredSession
	ntopics := se.LDAtopics
	if ntopics < 1 {
		ntopics = vv.LDATOPICS
	}

	var vs str.SearchStruct
	if srch.ID != "ldamodelbot()" {
		vs = search.SessionIntoBulkSearch(c, lnch.Config.VectorMaxlines)
		vlt.WSInfo.UpdateRemain <- vlt.WSSIKVi{vs.WSID, 1}
		vlt.WSInfo.UpdateSummMsg <- vlt.WSSIKVs{vs.WSID, LDAMSG}
		vlt.WSInfo.UpdateVProgMsg <- vlt.WSSIKVs{vs.WSID, fmt.Sprintf(ESM1)}
	} else {
		vs = srch
	}

	bags := ldapreptext(se.VecTextPrep, &vs)

	corpus := make([]string, len(bags))
	for i := 0; i < len(bags); i++ {
		corpus[i] = bags[i].ModifiedBag
	}

	stops := gen.StringMapKeysIntoSlice(getstopset())
	vectoriser := nlp.NewCountVectoriser(stops...)

	vlt.WSInfo.UpdateSummMsg <- vlt.WSSIKVs{vs.WSID, fmt.Sprintf(ESM2)}

	// consider building TESTITERATIONS models and making a table for each
	var dot mat.Matrix
	var tables []string

	// a chance to bail if you hit RtResetSession() in time
	if lnch.Config.SelfTest == 0 && !lnch.Config.VectorBot && !vlt.AllSessions.IsInVault(vs.User) {
		// mm("LDASearch() aborting: RtResetSession switched user to "+vs.User, MSGFYI)
		return gen.JSONresponse(c, str.SearchOutputJSON{})
	}

	docsOverTopics, topicsOverWords, ok := ldamodel(ntopics, corpus, vectoriser, &vs)
	if !ok {
		return gen.JSONresponse(c, str.SearchOutputJSON{})
	}

	tables = append(tables, ldatopicsummary(ntopics, topicsOverWords, vectoriser, docsOverTopics))
	tables = append(tables, ldatopsentences(ntopics, bags, corpus, docsOverTopics))
	dot = docsOverTopics

	htmltables := strings.Join(tables, "")

	incl := search.InclusionOverview(&srch, se.Inclusions)

	var img string
	if se.LDAgraph || srch.ID == "ldamodelbot()" {
		vlt.WSInfo.UpdateSummMsg <- vlt.WSSIKVs{vs.WSID, fmt.Sprintf(ESM3)}
		img = ldaplot(vs.Context, se.LDA2D, ntopics, incl, se.VecTextPrep, dot, bags)
	}

	soj := str.SearchOutputJSON{
		Title:         "",
		Searchsummary: "",
		Found:         htmltables,
		Image:         img,
		JS:            vv.VECTORJS,
	}

	vlt.WSInfo.Del <- srch.ID
	vlt.WSInfo.Del <- vs.ID

	return gen.JSONresponse(c, soj)
}

// ldapreptext - prepare the WorkLineBundle of a SearchStruct for lda analysis
func ldapreptext(bagger string, vs *str.SearchStruct) []BagWithLocus {

	var sb strings.Builder
	preallocate := vv.CHARSPERLINE * vs.Results.Len() // NB: a long line has 60 chars
	sb.Grow(preallocate)

	rr := vs.Results.YieldAll()
	for r := range rr {
		newtxt := fmt.Sprintf("⊏line/%s/%d⊐%s ", r.WkUID, r.TbIndex, r.MarkedUp)
		sb.WriteString(newtxt)
	}

	thetext := sb.String()
	sb.Reset()

	// do some preliminary cleanups

	strip := []string{`&nbsp;`, `- `, `<.*?>`}
	thetext = stripper(thetext, strip)

	// this would be a good place to deabbreviate, etc...
	thetext = makesubstitutions(thetext)
	thetext = gen.SwapAcuteForGrave(thetext)
	split := splitonpunctuaton(thetext)

	// empty sentences via "..."? not much of an issue: Cicero goes from 68790 to 68697
	// this will cost you c. .03s

	var ss []string
	for i := 0; i < len(split); i++ {
		if len(split[i]) > 0 {
			ss = append(ss, split[i])

			// fmt.Printf("(%d) %s\n", i, split[i])
			//(0) ⊏line/lt0959w014/34502⊐HALIEUTICA ⊏line/lt0959w014/34503⊐* * * ⊏line/lt0959w014/34504⊐Accepit mundus legem
			//(1)  dedit arma per omnes ⊏line/lt0959w014/34505⊐Admonuitque sui
			//(2)  uitulus sic namque minatur, ⊏line/lt0959w014/34506⊐Qui nondum gerit in tenera iam cornua fronte, ⊏line/lt0959w014/34507⊐Sic dammae fugiunt, pugnant uirtute leones ⊏line/lt0959w014/34508⊐Et morsu canis et caudae sic scorpios ictu ⊏line/lt0959w014/34509⊐Concussisque leuis pinnis sic euolat ales
		}
	}

	var thebags []BagWithLocus
	var first string
	var last string

	const tagger = `⊏(.*?)⊐`
	const notachar = `[^\sa-zα-ωϲῥἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάἐἑἒἓἔἕὲέἰἱἲἳἴἵἶἷὶίῐῑῒΐῖῗὀὁὂὃὄὅόὸὐὑὒὓὔὕὖὗϋῠῡῢΰῦῧύὺᾐᾑᾒᾓᾔᾕᾖᾗῂῃῄῆῇἤἢἥἣὴήἠἡἦἧὠὡὢὣὤὥὦὧᾠᾡᾢᾣᾤᾥᾦᾧῲῳῴῶῷώὼ]`
	re := regexp.MustCompile(tagger)

	cfg := ldavecconfig()

	iterations := len(ss) / cfg.SentencesPerBag
	index := 0
	for i := 0; i < iterations; i++ {
		parcel := strings.Join(ss[index:index+cfg.SentencesPerBag], " ")
		index = index + cfg.SentencesPerBag
		tags := re.FindAllStringSubmatch(parcel, -1)
		if len(tags) > 0 {
			first = tags[0][1]
			last = tags[len(tags)-1][1]
		} else {
			first = last
		}
		var sl BagWithLocus
		sl.Loc = first
		sl.Bag = strings.TrimSpace(strings.ToLower(parcel))
		sl.Bag = stripper(sl.Bag, []string{tagger, notachar})

		thebags = append(thebags, sl)

		// fmt.Println(sl)
		//{line/lt0959w014/34502 halieutica    accepit mundus legem }
		//{line/lt0959w014/34505  dedit arma per omnes admonuitque sui }
		//{line/lt0959w014/34506  uitulus sic namque minatur qui nondum gerit in tenera iam cornua fronte sic dammae fugiunt pugnant uirtute leones et morsu canis et caudae sic scorpios ictu concussisque leuis pinnis sic euolat ales }
	}

	allwords := make(map[string]bool, len(thebags))
	for i := 0; i < len(thebags); i++ {
		ww := strings.Split(thebags[i].Bag, " ")
		for j := 0; j < len(ww); j++ {
			allwords[ww[j]] = true
		}
	}

	slicedwords := gen.StringMapKeysIntoSlice(allwords)
	// catching resets
	if lnch.Config.SelfTest == 0 && !lnch.Config.VectorBot && !vlt.AllSessions.IsInVault(vs.User) {
		return []BagWithLocus{}
	}

	morphmapdbm := db.ArrayToGetRequiredMorphObjects(slicedwords) // map[string]DbMorphology
	morphmapstrslc := buildmorphmapstrslc(slicedwords, morphmapdbm)

	switch bagger {
	case "unparsed":
		thebags = ldaunmodifiedbagging(thebags)
	case "yoked":
		yokedmap := buildyokedparsemap(morphmapstrslc)
		thebags = ldayokedbagging(thebags, yokedmap)
	case "montecarlo":
		mcm := buildmontecarloparsemap(morphmapstrslc)
		thebags = ldamontecarlobagging(thebags, mcm)
	default:
		// winner
		winnermap := buildwinnertakesallparsemap(morphmapstrslc)
		thebags = ldawinnerbagging(thebags, winnermap)
	}

	// catching resets
	if lnch.Config.SelfTest == 0 && !lnch.Config.VectorBot && !vlt.AllSessions.IsInVault(vs.User) {
		return []BagWithLocus{}
	}

	return thebags
}

// ldaunmodifiedbagging - lda unmodified text bagger
func ldaunmodifiedbagging(thebags []BagWithLocus) []BagWithLocus {
	for i := 0; i < len(thebags); i++ {
		thebags[i].ModifiedBag = thebags[i].Bag
	}
	return thebags
}

// ldayokedbagging - lda yoked headwords text bagger
func ldayokedbagging(thebags []BagWithLocus, yokermap map[string]string) []BagWithLocus {
	stops := getstopset()
	for i := 0; i < len(thebags); i++ {
		var b strings.Builder
		yokedstring(&b, strings.Split(thebags[i].Bag, " "), yokermap, stops)
		thebags[i].ModifiedBag = b.String()
	}
	return thebags
}

// ldawinnerbagging - lda winner takes all headwords text bagger
func ldawinnerbagging(thebags []BagWithLocus, winnermap map[string]string) []BagWithLocus {
	stops := getstopset()
	for i := 0; i < len(thebags); i++ {
		var b strings.Builder
		winnerstring(&b, strings.Split(thebags[i].Bag, " "), winnermap, stops)
		thebags[i].ModifiedBag = b.String()

		// fmt.Printf("%s\t%s\n", thebags[i].Loc, thebags[i].ModifiedBag)
		//line/lt0959w014/34502	halieuticus    accipio mundus lego¹
		//line/lt0959w014/34505	 arma admoneo
		//line/lt0959w014/34506	 vitulus mino nondum gero¹ tener cornu frons² damma fugio pugno virtus leo² mordeo canae cauda scorpius ictus² concutio levis¹ pinnis evolo alo
	}
	return thebags
}

// ldamontecarlobagging - lda monte carlo headwords text bagger
func ldamontecarlobagging(thebags []BagWithLocus, montecarlo map[string]hwguesser) []BagWithLocus {
	stops := getstopset()
	for i := 0; i < len(thebags); i++ {
		var b strings.Builder
		montecarlostring(&b, strings.Split(thebags[i].Bag, " "), montecarlo, stops)
		thebags[i].ModifiedBag = b.String()
	}
	return thebags
}

// ldamodel - build the lda model for the corpus
func ldamodel(topics int, corpus []string, vectoriser *nlp.CountVectoriser, s *str.SearchStruct) (mat.Matrix, mat.Matrix, bool) {
	const (
		FAIL = "Failed to model topics for documents"
	)

	// cancellation is not a big deal: models are fairly fast; ldaplot() is where you can get stuck
	enablecancellation := func(l *nlp.LatentDirichletAllocation) {
		search.InsertNewContextIntoSS(s)
		l.Ctx = s.Context
		si := vlt.WSFetchSrchInfo(s.WSID)
		si.CancelFnc = s.CancelFnc
		vlt.WSInfo.InsertInfo <- si
	}

	cfg := ldavecconfig()
	lda := nlp.NewLatentDirichletAllocation(topics)
	enablecancellation(lda)

	lda.Processes = cfg.Goroutines
	lda.Iterations = cfg.LDAIterations
	lda.TransformationPasses = cfg.LDAXformPasses
	lda.BurnInPasses = cfg.BurnInPasses
	lda.ChangeEvaluationFrequency = cfg.ChangeEvalFrq
	lda.PerplexityEvaluationFrequency = cfg.PerplexEvalFrq

	// WARNING: the corpus and the vocab look different if you have the wrong separator set in yokedbags
	// nlp.NewPipeline(vectoriser, lda) will split yokedbags: spondeum•spondeus --> spondeum:25 spondeus:26
	pipeline := nlp.NewPipeline(vectoriser, lda)

	docsOverTopics, err := pipeline.FitTransform(corpus...)
	if err != nil {
		// you probably cancelled in the middle...
		Msg.FYI(FAIL)
		blank := mat.NewDense(1, 1, nil)
		return blank, blank, false
	}

	topicsOverWords := lda.Components()
	return docsOverTopics, topicsOverWords, true
}

// ldatopsentences - generate html table reporting sentences most associated with each topic
func ldatopsentences(ntopics int, thebags []BagWithLocus, corpus []string, docsOverTopics mat.Matrix) string {
	const (
		NTH = 2

		FULLTABLE = `
	<table class="ldasentences"><tbody>
	%s
	</tbody></table>
	<hr>`

		TABLETOP = `
    <tr class="vectorrow">
        <td class="vectorrank" colspan = "4">Sentences most associated with each topic</td>
    </tr>
	<tr class="vectorrow">
		<td class="vectorrank">Topic</td>
		<td class="vectorrank">Distance</td>
		<td class="vectorrank">Locus</td>
		<td class="vectorrank">Sentence</td>
	</tr>
    %s`

		TABLEROW = `
	<tr class="%s">%s
	</tr>`

		TABLEELEM = `
		<td class="vectorrank">%d</td>
		<td class="vectorscore">%.4f</td>
		<td class="vectorloc">%s</td>
		<td class="vectorsent">%s</td>`
	)

	// Examine Document over topic probability distribution

	rows, columns := docsOverTopics.Dims() // rows = NUMBEROFTOPICS; columns = len(thedocs)

	type DocRanker struct {
		d  string
		ff []float64 // would be nice if we could say [ntopics]float64
	}

	thedocs := make([]DocRanker, len(corpus))
	// need to fill the array with zeros to avoid "index out of range" error in next loop
	for doc := 0; doc < columns; doc++ {
		for i := 0; i < ntopics; i++ {
			thedocs[doc].ff = append(thedocs[doc].ff, float64(0))
		}
	}

	for doc := 0; doc < columns; doc++ {
		thedocs[doc].d = corpus[doc]
		for topic := 0; topic < rows; topic++ {
			f := docsOverTopics.At(topic, doc)
			thedocs[doc].ff[topic] = f
		}
	}

	// note that "i" is referring to the same item across slices; need this to be true...
	winners := make([]BagWithLocus, ntopics)
	for topic := 0; topic < rows; topic++ {
		mx := float64(0)
		winner := 0
		for i := 0; i < len(thedocs); i++ {
			ff := thedocs[i].ff
			if ff[topic] > mx {
				winner = i
				mx = ff[topic]
			}
			// fmt.Printf("(Topic #%d)(max=%f) Sentence #%d:\t%f - %s\n", topic, max, i, ff[topic], thedocs[i].d)
		}
		winners[topic] = thebags[winner]
		winners[topic].LDAScore = mx
		winners[topic].GetWL()
	}

	// [b] prepare text output
	var tablecolumn []string

	tp := `%s, %s %s`

	stripbold := strings.NewReplacer("&1", "", "&", "")

	for i, w := range winners {
		wl := w.Workline
		au := stripbold.Replace(mps.AllAuthors[wl.AuID()].IDXname)
		cit := fmt.Sprintf(tp, au, mps.AllWorks[wl.WkUID].Title, wl.Citation())
		r := fmt.Sprintf(TABLEELEM, i+1, w.LDAScore, cit, w.Bag)
		tablecolumn = append(tablecolumn, r)
	}

	var tablerows []string
	for i := range tablecolumn {
		rn := "vectorrow"
		if i%NTH == 0 {
			rn = "nthrow"
		}
		tablerows = append(tablerows, fmt.Sprintf(TABLEROW, rn, tablecolumn[i]))
	}

	tableout := fmt.Sprintf(TABLETOP, strings.Join(tablerows, "\n"))
	tableout = fmt.Sprintf(FULLTABLE, tableout)

	return tableout
}

// ldatopicsummary - html table that reports on top words and topic weights in the model
func ldatopicsummary(ntopics int, topicsOverWords mat.Matrix, vectoriser *nlp.CountVectoriser, docsOverTopics mat.Matrix) string {
	const (
		TOPN = 8
		NTH  = 2

		FULLTABLE = `
	<table class="ldawords"><tbody>
	%s
	</tbody></table>
	`

		TABLETOP = `
    <tr class="vectorrow">
        <td class="vectorrank" colspan = "4">Topic model of selection via Latent Dirichlet Allocation</td>
    </tr>
	<tr class="vectorrow">
		<td class="vectorrank">Topic</td>
		<td class="vectorrank">Top %d words associated with each topic</td>
		<td class="vectorrank"># of sentences with topic N as their dominant topic</td>
		<td class="vectorrank">scaled total accumulated weight of each topic</td>
	</tr>
    %s`

		TABLEROW = `
	<tr class="%s">%s
	</tr>`

		TABLEELEM = `
		<td class="vectorrank">%d</td>
		<td class="vectorsent">%s</td>
		<td class="vectorsent">%d (%.2f%%)</td>
		<td class="vectorsent">%.2f%%</td>`
	)

	tops := ldasortedtopics(ntopics, topicsOverWords, vectoriser)
	docspertopic := ldadocpertopic(ntopics, docsOverTopics)
	docsbyweight := ldadocbyweight(ntopics, docsOverTopics)

	tr, _ := topicsOverWords.Dims()
	_, dc := docsOverTopics.Dims()

	topn := TOPN
	if topn > ntopics {
		topn = ntopics
	}

	var tablecolumn []string
	for topic := 0; topic < tr; topic++ {
		ts := tops[topic]
		ww := make([]string, topn)
		for i := 0; i < topn; i++ {
			// ww[i] = fmt.Sprintf("%s (%.4f)", ts[i].W, ts[i].V)
			ww[i] = ts[i].W
		}
		data := strings.Join(ww, ", ")
		r := fmt.Sprintf(TABLEELEM, topic+1, data, docspertopic[topic], float64(docspertopic[topic])/float64(dc)*100, docsbyweight[topic]*100)
		tablecolumn = append(tablecolumn, r)
	}

	var tablerows []string
	for i := range tablecolumn {
		rn := "vectorrow"
		if i%NTH == 0 {
			rn = "nthrow"
		}
		tablerows = append(tablerows, fmt.Sprintf(TABLEROW, rn, tablecolumn[i]))
	}

	tableout := fmt.Sprintf(TABLETOP, topn, strings.Join(tablerows, "\n"))
	tableout = fmt.Sprintf(FULLTABLE, tableout)
	return tableout
}

type topicsorter struct {
	W string
	V float64
}

// ldasortedtopics - sorted most significant words for each topic
func ldasortedtopics(ntopics int, topicsOverWords mat.Matrix, vectoriser *nlp.CountVectoriser) map[int][]topicsorter {
	const (
		TOPN = 8
	)

	top := TOPN
	if top > ntopics {
		top = ntopics
	}

	tr, tc := topicsOverWords.Dims()

	vocab := make([]string, len(vectoriser.Vocabulary))
	for k, v := range vectoriser.Vocabulary {
		vocab[v] = k
	}

	tops := make(map[int][]topicsorter)
	for topic := 0; topic < tr; topic++ {
		tss := make([]topicsorter, tc)
		for word := 0; word < tc; word++ {
			tss[word] = topicsorter{
				W: vocab[word],
				V: topicsOverWords.At(topic, word),
			}
		}
		sort.Slice(tss, func(i, j int) bool {
			return tss[i].V > tss[j].V
		})
		tops[topic] = tss[0:top]
	}
	return tops
}

// ldadocpertopic - N sentences have topic X as their dominant topic
func ldadocpertopic(ntopics int, docsOverTopics mat.Matrix) []int {
	counter := make([]int, ntopics)
	dr, dc := docsOverTopics.Dims()
	for doc := 0; doc < dc; doc++ {
		mx := float64(0)
		winner := 0
		for topic := 0; topic < dr; topic++ {
			// any given corpus[doc] will look like
			// Topic #0=0.006009, Topic #1=0.006915, Topic #2=0.000688, Topic #3=0.449514, Topic #4=0.536875
			if docsOverTopics.At(topic, doc) > mx {
				winner = topic
				mx = docsOverTopics.At(topic, doc)
			}
		}
		counter[winner] += 1
	}
	return counter
}

// ldadocbyweight - scaled total accumulated weight of each topic
func ldadocbyweight(ntopics int, docsOverTopics mat.Matrix) []float64 {
	counter := make([]float64, ntopics)
	dr, dc := docsOverTopics.Dims()
	for doc := 0; doc < dc; doc++ {
		for topic := 0; topic < dr; topic++ {
			// any given corpus[doc] will look like
			// Topic #0=0.006009, Topic #1=0.006915, Topic #2=0.000688, Topic #3=0.449514, Topic #4=0.536875
			counter[topic] += docsOverTopics.At(topic, doc)
		}
	}

	mx := counter
	sort.Float64s(mx)
	high := mx[len(mx)-1]

	scaled := make([]float64, ntopics)
	for i := 0; i < ntopics; i++ {
		scaled[i] = counter[i] / high
	}
	return scaled
}

//
// LDA graphing prep
//

// see https://pkg.go.dev/gonum.org/v1/gonum/mat@v0.12.0#pkg-index

// ldaplot - plot the lda results
func ldaplot(ctx context.Context, graph2d bool, ntopics int, incl string, bagger string, docsOverTopics mat.Matrix, bags []BagWithLocus) string {
	const (
		PERPLEX = 150 // default 300
		LEARNRT = 100 // default 100
		MAXITER = 150 // default 300
		VERBOSE = false
		SKIPPED = `<span class="emph center">Skipped building a graph because the selection is larger than the cap: <code>%d</code> lines vs <code>%d</code> lines</span>`
	)

	dr, dc := docsOverTopics.Dims()
	doclabels := make([]float64, dc)
	for doc := 0; doc < dc; doc++ {
		mx := float64(0)
		winner := 0
		for topic := 0; topic < dr; topic++ {
			// any given corpus[doc] will look like
			// Topic #0=0.006009, Topic #1=0.006915, Topic #2=0.000688, Topic #3=0.449514, Topic #4=0.536875
			if docsOverTopics.At(topic, doc) > mx {
				winner = topic
				mx = docsOverTopics.At(topic, doc)
			}

		}
		doclabels[doc] = float64(winner)
	}

	var dd []float64
	for doc := 0; doc < dc; doc++ {
		for topic := 0; topic < dr; topic++ {
			dd = append(dd, docsOverTopics.At(topic, doc))
		}
	}

	// note that we flop r & c in the uncommented code; otherwise you get an 8x2 matrix later...
	// wv := mat.NewDense(dr, dc, dd)
	//fmt.Println(Y)
	// Y is the label for each row in the matrix

	// at the moment you get the following with Ovidius, Publius Naso, Halieutica [sp.]
	// &{{39 1 [6 6 2 6 0 6 6 6 4 7 1 0 7 6 4 3 1 5 2 1 1 4 4 5 2 5 7 5 3 1 4 2 6 5 2 1 0 7 6] 1} 39 1}
	//Computing P-values for point 0 of 8...
	//8
	//&{{8 2 [-2.808419429316656e-05 9.116052264392355e-06 -0.00010388007260382351 -1.5467174909099705e-05 5.304935449968505e-05 -5.228522386515839e-05 3.919385587076133e-06 -4.8151150810876286e-05 -0.00010957199636924793 -3.192900565781712e-05 6.719290262741853e-05 1.537943454450102e-05 1.1299109041087078e-05 1.4831064365075226e-05 0.00010607551151097121 0.00010850600406898292] 2} 8 2}
	// but what you want is a 39x2 matrix in that second slot

	wv := mat.NewDense(dc, dr, dd)
	Y := mat.NewDense(dc, 1, doclabels)

	graph := true
	cfg := ldavecconfig()
	if dc > cfg.MaxLDAGraphSize {
		graph = false
	}

	// this is the slow part...: you make it here in no time; then you can easily wait >1m
	// tsnemp needs to learn how to accept context
	vv.LaunchTime = time.Now()
	var htmlandjs string
	if graph2d && graph {
		// t := NewTSNE(2, PERPLEX, LEARNRT, MAXITER, VERBOSE)
		t := tsnemp.NewMPTSNE(lnch.Config.WorkerCount, 2, PERPLEX, LEARNRT, MAXITER, VERBOSE)
		t.EmbedDataWithCtx(ctx, wv, nil)
		htmlandjs = lda2dscatter(ntopics, incl, bagger, t.Y, Y, bags)
	} else if graph {
		// 3d
		nd := tsnemp.NewMPTSNE(lnch.Config.WorkerCount, 3, PERPLEX, LEARNRT, MAXITER, VERBOSE)
		nd.EmbedDataWithCtx(ctx, wv, nil)
		htmlandjs = lda3dscatter(ntopics, incl, bagger, nd.Y, Y, bags)
	} else {
		p := message.NewPrinter(language.English)
		htmlandjs = p.Sprintf(SKIPPED, dc, cfg.MaxLDAGraphSize)
	}
	Msg.PEEK(fmt.Sprintf("ldaplot() EmbedData required %.3fs", time.Now().Sub(vv.LaunchTime).Seconds()))

	return htmlandjs
}

//
// CLEANING
//

// stripper - delete each in a list of items from a string
func stripper(item string, purge []string) string {
	for i := 0; i < len(purge); i++ {
		re := regexp.MustCompile(purge[i])
		item = re.ReplaceAllString(item, "")
	}
	return item
}

// makesubstitutions - swap full words for abbreviations, etc
func makesubstitutions(thetext string) string {
	// https://golang.org/pkg/strings/#NewReplacer
	// cf cleanvectortext() in vectorhelpers.py
	swap := strings.NewReplacer("v", "u", "j", "i", "σ", "ϲ", "ς", "ϲ", "A.", "Aulus", "App.", "Appius",
		"C.", "Caius", "G.", "Gaius", "Cn.", "Cnaius", "Gn.", "Gnaius", "D.", "Decimus", "L.", "Lucius", "M.", "Marcus",
		"M.’", "Manius", "N.", "Numerius", "P.", "Publius", "Q.", "Quintus", "S.", "Spurius", "Sp.", "Spurius",
		"Ser.", "Servius", "Sex.", "Sextus", "T.", "Titus", "Ti", "Tiberius", "V.", "Vibius", "a.", "ante",
		"d.", "dies", "Id.", "Idibus", "Kal.", "Kalendas", "Non.", "Nonas", "prid.", "pridie", "Ian.", "Ianuarias",
		"Feb.", "Februarias", "Mart.", "Martias", "Apr.", "Aprilis", "Mai.", "Maias", "Iun.", "Iunias",
		"Quint.", "Quintilis", "Sext.", "Sextilis", "Sept.", "Septembris", "Oct.", "Octobris", "Nov.", "Novembris",
		"Dec.", "Decembris")

	return swap.Replace(thetext)
}

// splitonpunctuaton - swap all punctuation for one item; then split on it...
func splitonpunctuaton(thetext string) []string {
	swap := strings.NewReplacer("?", ".", "!", ".", "·", ".", ";", ".", ":", ".")
	thetext = swap.Replace(thetext)
	return strings.Split(thetext, ".")
}

//
// SAMPLE RESULTS: 3 queries in a row of Ap., Met. w/ 5 topics & 12 iterations
//

// the third is interesting: beginnings and endings are being found...

// [HGS] TESTING: NeighborsSearch rerouting to LDASearch()
//topic 1:	0.993753.3	Apuleius Madaurensis, Metamorphoses 9.12.11	 dii boni quales illic homunculi uibicibus liuidis totam cutem depicti dorsumque plagosum scissili centunculo magis inumbrati quam obtecti nonnulli exiguo tegili tantum modo pubem iniecti cuncti tamen sic tunicati ut essent per pannulos manifesti frontes litterati et capillum semirasi et pedes anulati tum lurore deformes et fumosis tenebris uaporosae caliginis palpebras adesi atque adeo male luminati et in modum pugilum qui puluisculo perspersi dimicant farinulenta cinere sordide candidati
//topic 2:	0.994870.3	Apuleius Madaurensis, Metamorphoses 11.3.15	 corona multiformis uariis floribus sublimem destrinxerat uerticem cuius media quidem super frontem plana rutunditas in modum speculi uel immo argumentum lunae candidum lumen emicabat dextra laeuaque sulcis insurgentium uiperarum cohibita spicis etiam cerialibus desuper porrectis conspicuante tunica multicolor bysso tenui pertexta nunc albo candore lucida nunc croceo flore lutea nunc roseo rubore flammida et quae longe longeque etiam meum confutabat optutum palla nigerrima splendescens atro nitore quae circumcirca remeans et sub dexterum latus ad umerum laeuum recurrens umbonis uicem deiecta parte laciniae multiplici contabulatione dependula ad ultimas oras nodulis fimbriarum decoriter confluctuabat
//topic 3:	0.994591.3	Apuleius Madaurensis, Metamorphoses 10.20.5	 quattuor eunuchi confestim puluillis compluribus uentose tumentibus pluma delicata terrestrem nobis cubitum praestruunt sed et stragula ueste auro ac murice tyrio depicta probe consternunt ac desuper breuibus admodum sed satis copiosis puluillis aliis nimis modicis quis maxillas et ceruices delicatae mulieres suffulcire consuerunt superstruunt
//topic 4:	0.993883.3	Apuleius Madaurensis, Metamorphoses 11.30.3	 nec deinceps postposito uel in supinam procrastinationem reiecto negotio statim sacerdoti meo relatis quae uideram inanimae protinus castimoniae iugum subeo et lege perpetua praescriptis illis decem diebus spontali sobrietate multiplicatis instructum teletae comparo largitus omnibus ex studio pietatis magis quam mensura rerum mearum collatis
//topic 5:	0.992361.3	Apuleius Madaurensis, Metamorphoses 6.12.1	 perrexit psyche uolenter non obsequium quidem illa functura sed requiem malorum praecipitio fluuialis rupis habiturante sed inde de fluuio musicae suauis nutricula leni crepitu dulcis aurae diuinitus inspirata sic uaticinatur harundo uiridis psyche tantis aerumnis exercita neque tua miserrima morte meas sanctas aquas polluas nec uero istud horae con tra formidabiles oues feras aditum quoad de solis fraglantia mutuatae calorem truci rabie solent efferri cornuque acuto et fronte saxea et non nunquam uenenatis morsibus in exitium saeuire mortalium
//
//[HGS] TESTING: NeighborsSearch rerouting to LDASearch()
//topic 1:	0.993477.3	Apuleius Madaurensis, Metamorphoses 4.8.9	 estur ac potatur incondite pulmentis aceruatim panibus aggeratim poculis agminatim ingestis
//topic 2:	0.996230.3	Apuleius Madaurensis, Metamorphoses 5.20.6	 nouaculam praeacutam adpulsu etiam palmulae lenientis exasperatam tori qua parte cubare consuesti latenter absconde lucernamque concinnem completam oleo claro lumine praemicantem subde aliquo claudentis aululae tegmine omnique isto apparatu tenacissime dissimulato postquam sulcatum trahens gressum cubile solitum conscenderit iamque porrectus et exordio somni prementis implicitus altum soporem flare coeperit toro delapsa nudoque uestigio pensilem gradum paullulatim minuens cae cae tenebrae custodia liberata lucerna praeclari tui facinoris opportunitatem de luminis consilio mutuare et ancipiti telo illo audaciter prius dextera sursum elata nisu quam ualido noxii serpentis nodum ceruicis et capitis abscide
//topic 3:	0.993673.3	Apuleius Madaurensis, Metamorphoses 9.32.9	 sed ecce siderum ordinatis ambagibus per numeros dierum ac mensuum remeans annus post mustulentas autumni delicias ad hibernas capricorni pruinas deflexerat et adsiduis pluuiis noctur nisque rorationibus sub dio et intecto conclusus stabulo continuo discruciabar frigore quippe cum meus dominus prae nimia paupertate ne sibi quidem nedum mihi posset stramen aliquod uel exiguum tegimen parare sed frondoso casulae contentus umbraculo degeret
//topic 4:	0.996601.3	Apuleius Madaurensis, Metamorphoses 11.28.3	 nam et uiriculas patrimonii peregrinationis adtriuerant impensae et erogationes urbicae pristinis illis prouincialibus antistabant plurimum
//topic 5:	0.993569.3	Apuleius Madaurensis, Metamorphoses 8.27.1	 die sequenti uariis coloribus indusiati et deformiter quisque formati facie caenoso pigmento delita et oculis obunctis graphice prodeunt mitellis et crocotis et carbasinis et bombycinis iniecti quidam tunicas albas in modum lanciolarum quoquouersum fluente purpura depictas cingulo subligati pedes luteis induti calceis
//
//[HGS] TESTING: NeighborsSearch rerouting to LDASearch()
//topic 1:	0.994495.3	Apuleius Madaurensis, Metamorphoses 11.25.13	 tiberiusbi respondent sidera redeunt tempora gaudent numina seruiunt elementante tuo nutu spirant flamina nutriunt nubila germinant semina crescunt germinante tuam maiestatem perhorrescunt aues caelo meantes ferae montibus errantes serpentes solo latentes beluae ponto natantes
//topic 2:	0.995036.3	Apuleius Madaurensis, Metamorphoses 1.2.5	 postquam ardua montium et lubrica uallium et roscida cespitum et glebosa camporum emensus emersi in equo indigena peralbo uehens iam eo quoque admodum fesso ut ipse etiam fatigationem sedentariam incessus uegetatione discuterem in pedes desilio equi sudorem fronde detergeo frontem curiose exfrico auris remulceo frenos detraho in gradum lenem sensim proueho quoad lassitudinis incommodum alui solitum ac naturale praesidium eliquaret
//topic 3:	0.992108.3	Apuleius Madaurensis, Metamorphoses 1.6.9	 at uero domi tuae iam defletus et conclamatus es liberis tuis tutores iuridici prouincialis decreto dati uxor persolutis feralibus officiis luctu et maerore diuturno deformata diffletis paene ad extremam captiuitatem oculis suis domus infortunium nouarum nuptiarum gaudiis a suis sibi parentibus hilarare compellitur
//topic 4:	0.996159.3	Apuleius Madaurensis, Metamorphoses 10.20.5	 quattuor eunuchi confestim puluillis compluribus uentose tumentibus pluma delicata terrestrem nobis cubitum praestruunt sed et stragula ueste auro ac murice tyrio depicta probe consternunt ac desuper breuibus admodum sed satis copiosis puluillis aliis nimis modicis quis maxillas et ceruices delicatae mulieres suffulcire consuerunt superstruunt
//topic 5:	0.992793.3	Apuleius Madaurensis, Metamorphoses 11.16.28	 tunc cuncti populi tam religiosi quam profani uannos onustas aromatis et huiusce modi suppliciis certatim congerunt et insuper fluctus libant intritum lacte confectum donec muneribus largis et deuotionibus faustis completa nauis absoluta strophiis ancoralibus peculiari serenoque flatu pelago redderetur

// 8 topic sends the modeler to book 11 hard; and will multi-hit 11.30!!

// if the iterations goes way, way up, the topic sentences get very short
// 1000 iterations
//topic 1:        0.999957.3      Apuleius Madaurensis, Metamorphoses 2.30.27      aures pertracto deruunt
//topic 2:        0.999925.3      Apuleius Madaurensis, Metamorphoses 1.25.10      sed non impune
//topic 3:        0.999881.3      Apuleius Madaurensis, Metamorphoses 3.9.20       quod monstrum
//topic 4:        0.997353.3      Apuleius Madaurensis, Metamorphoses 4.34.13      quid canitiem scinditis
//topic 5:        0.999912.3      Apuleius Madaurensis, Metamorphoses 2.18.18      nec tamen incomitatus ibo

// 250 iterations
//topic 1:        0.997896.3      Apuleius Madaurensis, Metamorphoses 1.21.6       adnuit
//topic 2:        0.999632.3      Apuleius Madaurensis, Metamorphoses 1.24.20      quae autem tibi causa peregrinationis huius
//topic 3:        0.997320.3      Apuleius Madaurensis, Metamorphoses 2.7.8        ipsa linea tunica mundule amicta et russea fasceola praenitente altiuscule sub ipsas papillas succinctula illud cibarium uasculum floridis palmulis rotabat in circulum et in orbis flexibus crebra succutiens et simul membra sua leniter inlubricans lumbis sensim  uibrantibus spinam mobilem quatiens placide decenter undabat
//topic 4:        0.999935.3      Apuleius Madaurensis, Metamorphoses 9.17.8       aretem meam condiscipulam memoras
//topic 5:        0.999668.3      Apuleius Madaurensis, Metamorphoses 2.18.18      nec tamen incomitatus ibo

// 100 iterations
// [HGS] TESTING: NeighborsSearch rerouting to LDASearch()
//topic 1:        0.995875.3      Apuleius Madaurensis, Metamorphoses 1.1.7        exordior
//topic 2:        0.994873.3      Apuleius Madaurensis, Metamorphoses 6.6.14       cedunt nubes et caelum filiae panditur et summus aether cum gaudio suscipit deam nec obuias aquilas uel accipitres rapaces pertimescit magnae veneris canora familiante tunc se protinus ad iouis regias arces dirigit et petitu superbo mercuri dei uocalis operae necessa riam usuram postulat
//topic 3:        0.995717.3      Apuleius Madaurensis, Metamorphoses 4.21.20      sic etiam thrasyleon nobis periuit sed a gloria non peribit
//topic 4:        0.995084.3      Apuleius Madaurensis, Metamorphoses 11.28.3      nam et uiriculas patrimonii peregrinationis adtriuerant impensae et erogationes urbicae pristinis illis prouincialibus antistabant plurimum
//topic 5:        0.997372.3      Apuleius Madaurensis, Metamorphoses 9.17.8       aretem meam condiscipulam memoras

// 50 iterations
//topic 1:        0.992799.3      Apuleius Madaurensis, Metamorphoses 7.16.4       equinis armentis namque me congregem pastor egregius mandati dominici serus auscultator aliquando permisit
//topic 2:        0.994830.3      Apuleius Madaurensis, Metamorphoses 11.26.1       diu denique gratiarum gerendarum sermone prolixo commoratus tandem digredior et recta patrium larem reuisurus meum post aliquam multum temporis contendo paucisque post diebus deae potentis instinctu raptim constrictis sarcinulis naue conscensa romam uersus profectionem dirigo tutusque prosperitate uentorum ferentium augusti portum celerrime peruenio ac dehinc carpento peruolaui uesperaque quam dies insequebatur iduum decembrium sacrosanctam istam ciuitatem accedo
//topic 3:        0.996028.3      Apuleius Madaurensis, Metamorphoses 11.11.21     eius orificium non altiuscule leuatum in canalem porrectum longo riuulo prominebat ex alia uero parte multum recedens spatiosa dilatione adhaerebat ansa quam contorto nodulo supersedebat aspis squameae ceruicis striato tumore sublimis
//topic 4:        0.999708.3      Apuleius Madaurensis, Metamorphoses 1.4.21       haec tibi merces deposita est
//topic 5:        0.994164.3      Apuleius Madaurensis, Metamorphoses 11.30.3      nec deinceps postposito uel in supinam procrastinationem reiecto negotio statim sacerdoti meo relatis quae uideram inanimae protinus castimoniae iugum subeo et lege perpetua praescriptis illis decem diebus spontali sobrietate multiplicatis instructum teletae comparo largitus omnibus ex studio pietatis magis quam mensura rerum mearum collatis

// 25 iterations
//[HGS] TESTING: NeighborsSearch rerouting to LDASearch()
//topic 1:        0.995270.3      Apuleius Madaurensis, Metamorphoses 10.18.15     spretis luculentis illis suis uehiculis ac posthabitis decoris raedarum carpentis quae partim contecta partim reuelata frustra nouissimis trahebantur consequiis equis etiam thessalicis et aliis iumentis gallicanis quibus generosa suboles perhibet pretiosam dignitatem me phaleris aureis et fucatis ephippiis et purpureis tapetis et frenis argenteis et pictilibus balteis et tintinnabulis perargutis exornatum ipse residens amantissime nonnunquam comissimis adfatur sermonibus atque inter alia pleraque summe se delectari  profitebatur quod haberet in me simul et conuiuam et uectorem
//topic 2:        0.994604.3      Apuleius Madaurensis, Metamorphoses 11.3.15      corona multiformis uariis floribus sublimem destrinxerat uerticem cuius media quidem super frontem plana rutunditas in modum speculi uel immo argumentum lunae candidum lumen emicabat dextra laeuaque sulcis insurgentium uiperarum cohibita spicis etiam cerialibus desuper porrectis conspicuante tunica multicolor bysso tenui pertexta nunc albo candore lucida nunc croceo flore lutea nunc roseo rubore flammida et quae longe longeque etiam meum confutabat optutum palla nigerrima splendescens atro nitore quae circumcirca remeans et sub dexterum latus ad umerum laeuum recurrens umbonis uicem deiecta parte laciniae multiplici contabulatione dependula ad ultimas oras nodulis fimbriarum decoriter confluctuabat
//topic 3:        0.995729.3      Apuleius Madaurensis, Metamorphoses 10.5.14      sed dira illa femina et malitiae nouercalis exemplar unicum non acerba filii morte non parricidii conscientia non infortunio domus non luctu mariti uel aerumna funeris commota cladem familiae in uindictae compendium traxit missoque protinus cursore qui uianti marito domus expugnationem nuntiaret ac mox eodem ocius ab itinere regresso personata nimia temeritate insimulat priuigni ueneno filium suum interceptum
//topic 4:        0.995073.3      Apuleius Madaurensis, Metamorphoses 6.24.10      tunc apollo cantauit ad citharam venus suaui musicae superingressa formonsa saltauit scaena sibi sic concinnata ut musae quidem chorum canerent tibias inflaret saturus et paniscus ad fistulam diceret
//topic 5:        0.993980.3      Apuleius Madaurensis, Metamorphoses 11.28.3      nam et uiriculas patrimonii peregrinationis adtriuerant impensae et erogationes urbicae pristinis illis prouincialibus antistabant plurimum

// notes from other experiments

// LDA for Ap. Met. really does gravitate to books 1 and 11. Somehow noting the open/close/programmatic quality of these books
// LDA for Cic, Ep. ad Att. will routinely grab Greek sentences. I.e., Greek is one of the consistent "topics".
// LDA for Plato is very interested in the spuria; it seems to note that the as-if plato is especially on the nose...
// LDA for Sophocles is extremely likely to turn up Philoctetes 739: ἆ ἆ ἆ ἆ; "lament" being a "topic"... [also v. interested in φεῦ φεῦ δύϲταν]
