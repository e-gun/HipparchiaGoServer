//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"github.com/james-bowman/nlp"
	"github.com/labstack/echo/v4"
	"regexp"
	"strconv"
	"strings"
)

// "github.com/james-bowman/nlp" also contains some interesting possibilities: LatentDirichletAllocation, etc.
// bagging as per the old HipparchiaGoDBHelper code: sentence by sentence; much of the code below from HipparchiaGoDBHelper

// bowman's package can also do nearest neighbour similarity searches: LinearScanIndex.Search(qv mat.Vector, k int) -> []Match

// with some (i.e., a lot of...) work the output could be fed to JS as per the python LDA visualizer

// see bottom of file for sample results

const (
	SENTENCESPERBAG = 1
	NUMBEROFTOPICS  = 5
	LDAITERATIONS   = 12
)

type BagWithLocus struct {
	Loc         string
	Bag         string
	ModifiedBag string
	LDAScore    float64
	Workline    DbWorkline
}

func (b *BagWithLocus) GetWL() {
	tb := strings.Split(b.Loc, "/")
	ln, e := strconv.Atoi(tb[2])
	if e != nil {
		msg("BagWithLocus.GetWL() failed to convert ascii to int", 2)
	}
	b.Workline = GrabOneLine(tb[1][:6], ln)
}

// ldatest - testing LatentDirichletAllocation
func ldatest(c echo.Context) {
	// not for production...

	// there is no interface for this
	// the output goes to the terminal, not to the web page

	// force "s.VecLDA = true" in MakeDefaultSession(); build; run
	// then all vector searches will pass through here (and NN searches will be unavailable in this build)

	vs := sessionintobulksearch(c, Config.VectorMaxlines)
	lda(vs.Results)
}

// lda - report the N sentences that most fit the N topics you are modeling
func lda(dblines []DbWorkline) {

	var sb strings.Builder
	preallocate := CHARSPERLINE * len(dblines) // NB: a long line has 60 chars
	sb.Grow(preallocate)

	for i := 0; i < len(dblines); i++ {
		newtxt := fmt.Sprintf("⊏line/%s/%d⊐%s ", dblines[i].WkUID, dblines[i].TbIndex, dblines[i].MarkedUp)
		sb.WriteString(newtxt)
	}

	thetext := sb.String()
	sb.Reset()

	// do some preliminary cleanups

	strip := []string{`&nbsp;`, `- `, `<.*?>`}
	thetext = stripper(thetext, strip)

	// this would be a good place to deabbreviate, etc...
	thetext = makesubstitutions(thetext)
	thetext = acuteforgrave(thetext)
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

	// SentPerBag = number of sentences per bag

	iterations := len(ss) / SENTENCESPERBAG
	index := 0
	for i := 0; i < iterations; i++ {
		parcel := strings.Join(ss[index:index+SENTENCESPERBAG], " ")
		index = index + SENTENCESPERBAG
		tags := re.FindAllStringSubmatch(parcel, -1)
		if len(tags) > 0 {
			first = tags[0][1]
			last = tags[len(tags)-1][1]
		} else {
			first = last
		}
		var sl BagWithLocus
		sl.Loc = first
		sl.Bag = strings.ToLower(parcel)
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

	slicedwords := StringMapKeysIntoSlice(allwords)
	morphmapdbm := arraytogetrequiredmorphobjects(slicedwords) // map[string]DbMorphology
	morphmapstrslc := buildmorphmapstrslc(slicedwords, morphmapdbm)
	winnermap := buildwinnertakesallparsemap(morphmapstrslc)

	for i := 0; i < len(thebags); i++ {
		var b strings.Builder
		winnerstring(&b, strings.Split(thebags[i].Bag, " "), winnermap)
		thebags[i].ModifiedBag = b.String()

		// fmt.Printf("%s\t%s\n", thebags[i].Loc, thebags[i].ModifiedBag)
		//line/lt0959w014/34502	halieuticus    accipio mundus lego¹
		//line/lt0959w014/34505	 arma admoneo
		//line/lt0959w014/34506	 vitulus mino nondum gero¹ tener cornu frons² damma fugio pugno virtus leo² mordeo canae cauda scorpius ictus² concutio levis¹ pinnis evolo alo
	}

	corpus := make([]string, len(thebags))
	for i := 0; i < len(thebags); i++ {
		corpus[i] = thebags[i].ModifiedBag
	}

	stops := StringMapKeysIntoSlice(getstopset())
	vectoriser := nlp.NewCountVectoriser(stops...)

	lda := nlp.NewLatentDirichletAllocation(NUMBEROFTOPICS)
	lda.Processes = Config.WorkerCount
	lda.Iterations = LDAITERATIONS

	pipeline := nlp.NewPipeline(vectoriser, lda)

	docsOverTopics, err := pipeline.FitTransform(corpus...)
	if err != nil {
		fmt.Printf("Failed to model topics for documents because %v", err)
		return
	}

	// Examine Document over topic probability distribution
	type DocRanker struct {
		d  string
		ff [NUMBEROFTOPICS]float64
	}

	thedocs := make([]DocRanker, len(corpus))
	rows, columns := docsOverTopics.Dims() // rows = NUMBEROFTOPICS; columns = len(thedocs)

	for doc := 0; doc < columns; doc++ {
		thedocs[doc].d = corpus[doc]
		for topic := 0; topic < rows; topic++ {
			f := docsOverTopics.At(topic, doc)
			thedocs[doc].ff[topic] = f
		}
	}

	// note that "i" is referring to the same item across slices; need this to be true...
	winners := make([]BagWithLocus, NUMBEROFTOPICS)
	for topic := 0; topic < rows; topic++ {
		max := float64(0)
		winner := 0
		for i := 0; i < len(thedocs); i++ {
			ff := thedocs[i].ff
			if ff[topic] > max {
				winner = i
				max = ff[topic]
			}
			// fmt.Printf("(Topic #%d)(max=%f) Sentence #%d:\t%f - %s\n", topic, max, i, ff[topic], thedocs[i].d)
		}
		winners[topic] = thebags[winner]
		winners[topic].LDAScore = max
		winners[topic].GetWL()
	}

	for i := 0; i < NUMBEROFTOPICS; i++ {
		w := winners[i]
		wl := w.Workline
		tp := `%s, %s %s`
		cit := fmt.Sprintf(tp, AllAuthors[wl.AuID()].Cleaname, AllWorks[wl.WkUID].Title, wl.Citation())
		fmt.Printf("topic %d:\t%f.3\t%s\t%s\n\n", i+1, w.LDAScore, cit, w.Bag)
	}

	// Examine Topic over word probability distribution
	//topicsOverWords := lda.Components()
	//tr, tc := topicsOverWords.Dims()
	//
	//vocab := make([]string, len(vectoriser.Vocabulary))
	//for k, v := range vectoriser.Vocabulary {
	//	vocab[v] = k
	//}
	//for topic := 0; topic < tr; topic++ {
	//	fmt.Printf("\nWord distribution for Topic #%d -", topic)
	//	for word := 0; word < tc; word++ {
	//		if word > 0 {
	//			fmt.Printf(",")
	//		}
	//		fmt.Printf(" '%s'=%f", vocab[word], topicsOverWords.At(topic, word))
	//	}
	//}
}

//
// CLEANING
//

func stripper(item string, purge []string) string {
	// delete each in a list of items from a string
	for i := 0; i < len(purge); i++ {
		re := regexp.MustCompile(purge[i])
		item = re.ReplaceAllString(item, "")
	}
	return item
}

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

func splitonpunctuaton(thetext string) []string {
	// replacement for recursivesplitter(): perhaps very slightly faster, but definitely much more direct and legible
	// swap all punctuation for one item; then split on it...
	swap := strings.NewReplacer("?", ".", "!", ".", "·", ".", ";", ".")
	thetext = swap.Replace(thetext)
	split := strings.Split(thetext, ".")

	// slower way of doing the same...

	//re := regexp.MustCompile("[?!;·]")
	//thetext = re.ReplaceAllString(thetext, ".")
	//split := strings.Split(thetext, ".")

	return split
}

func acuteforgrave(thetext string) string {
	swap := strings.NewReplacer("ὰ", "ά", "ὲ", "έ", "ὶ", "ί", "ὸ", "ό", "ὺ", "ύ", "ὴ", "ή", "ὼ", "ώ",
		"ἂ", "ἄ", "ἒ", "ἔ", "ἲ", "ἴ", "ὂ", "ὄ", "ὒ", "ὔ", "ἢ", "ἤ", "ὢ", "ὤ", "ᾃ", "ᾅ", "ᾓ", "ᾕ", "ᾣ", "ᾥ",
		"ᾂ", "ᾄ", "ᾒ", "ᾔ", "ᾢ", "ᾤ")
	return swap.Replace(thetext)
}

//
// SAMPLE RESULTS: 3 queries in a row of Ap., Met. w/ 5 topics & 12 iterations
//

// the third is interesting: beginnings and endings are being found...

// [HGS] TESTING: VectorSearch rerouting to ldatest()
//topic 1:	0.993753.3	Apuleius Madaurensis, Metamorphoses 9.12.11	 dii boni quales illic homunculi uibicibus liuidis totam cutem depicti dorsumque plagosum scissili centunculo magis inumbrati quam obtecti nonnulli exiguo tegili tantum modo pubem iniecti cuncti tamen sic tunicati ut essent per pannulos manifesti frontes litterati et capillum semirasi et pedes anulati tum lurore deformes et fumosis tenebris uaporosae caliginis palpebras adesi atque adeo male luminati et in modum pugilum qui puluisculo perspersi dimicant farinulenta cinere sordide candidati
//
//topic 2:	0.994870.3	Apuleius Madaurensis, Metamorphoses 11.3.15	 corona multiformis uariis floribus sublimem destrinxerat uerticem cuius media quidem super frontem plana rutunditas in modum speculi uel immo argumentum lunae candidum lumen emicabat dextra laeuaque sulcis insurgentium uiperarum cohibita spicis etiam cerialibus desuper porrectis conspicuante tunica multicolor bysso tenui pertexta nunc albo candore lucida nunc croceo flore lutea nunc roseo rubore flammida et quae longe longeque etiam meum confutabat optutum palla nigerrima splendescens atro nitore quae circumcirca remeans et sub dexterum latus ad umerum laeuum recurrens umbonis uicem deiecta parte laciniae multiplici contabulatione dependula ad ultimas oras nodulis fimbriarum decoriter confluctuabat
//
//topic 3:	0.994591.3	Apuleius Madaurensis, Metamorphoses 10.20.5	 quattuor eunuchi confestim puluillis compluribus uentose tumentibus pluma delicata terrestrem nobis cubitum praestruunt sed et stragula ueste auro ac murice tyrio depicta probe consternunt ac desuper breuibus admodum sed satis copiosis puluillis aliis nimis modicis quis maxillas et ceruices delicatae mulieres suffulcire consuerunt superstruunt
//
//topic 4:	0.993883.3	Apuleius Madaurensis, Metamorphoses 11.30.3	 nec deinceps postposito uel in supinam procrastinationem reiecto negotio statim sacerdoti meo relatis quae uideram inanimae protinus castimoniae iugum subeo et lege perpetua praescriptis illis decem diebus spontali sobrietate multiplicatis instructum teletae comparo largitus omnibus ex studio pietatis magis quam mensura rerum mearum collatis
//
//topic 5:	0.992361.3	Apuleius Madaurensis, Metamorphoses 6.12.1	 perrexit psyche uolenter non obsequium quidem illa functura sed requiem malorum praecipitio fluuialis rupis habiturante sed inde de fluuio musicae suauis nutricula leni crepitu dulcis aurae diuinitus inspirata sic uaticinatur harundo uiridis psyche tantis aerumnis exercita neque tua miserrima morte meas sanctas aquas polluas nec uero istud horae con tra formidabiles oues feras aditum quoad de solis fraglantia mutuatae calorem truci rabie solent efferri cornuque acuto et fronte saxea et non nunquam uenenatis morsibus in exitium saeuire mortalium
//
//[HGS] TESTING: VectorSearch rerouting to ldatest()
//topic 1:	0.993477.3	Apuleius Madaurensis, Metamorphoses 4.8.9	 estur ac potatur incondite pulmentis aceruatim panibus aggeratim poculis agminatim ingestis
//
//topic 2:	0.996230.3	Apuleius Madaurensis, Metamorphoses 5.20.6	 nouaculam praeacutam adpulsu etiam palmulae lenientis exasperatam tori qua parte cubare consuesti latenter absconde lucernamque concinnem completam oleo claro lumine praemicantem subde aliquo claudentis aululae tegmine omnique isto apparatu tenacissime dissimulato postquam sulcatum trahens gressum cubile solitum conscenderit iamque porrectus et exordio somni prementis implicitus altum soporem flare coeperit toro delapsa nudoque uestigio pensilem gradum paullulatim minuens cae cae tenebrae custodia liberata lucerna praeclari tui facinoris opportunitatem de luminis consilio mutuare et ancipiti telo illo audaciter prius dextera sursum elata nisu quam ualido noxii serpentis nodum ceruicis et capitis abscide
//
//topic 3:	0.993673.3	Apuleius Madaurensis, Metamorphoses 9.32.9	 sed ecce siderum ordinatis ambagibus per numeros dierum ac mensuum remeans annus post mustulentas autumni delicias ad hibernas capricorni pruinas deflexerat et adsiduis pluuiis noctur nisque rorationibus sub dio et intecto conclusus stabulo continuo discruciabar frigore quippe cum meus dominus prae nimia paupertate ne sibi quidem nedum mihi posset stramen aliquod uel exiguum tegimen parare sed frondoso casulae contentus umbraculo degeret
//
//topic 4:	0.996601.3	Apuleius Madaurensis, Metamorphoses 11.28.3	 nam et uiriculas patrimonii peregrinationis adtriuerant impensae et erogationes urbicae pristinis illis prouincialibus antistabant plurimum
//
//topic 5:	0.993569.3	Apuleius Madaurensis, Metamorphoses 8.27.1	 die sequenti uariis coloribus indusiati et deformiter quisque formati facie caenoso pigmento delita et oculis obunctis graphice prodeunt mitellis et crocotis et carbasinis et bombycinis iniecti quidam tunicas albas in modum lanciolarum quoquouersum fluente purpura depictas cingulo subligati pedes luteis induti calceis
//
//[HGS] TESTING: VectorSearch rerouting to ldatest()
//topic 1:	0.994495.3	Apuleius Madaurensis, Metamorphoses 11.25.13	 tiberiusbi respondent sidera redeunt tempora gaudent numina seruiunt elementante tuo nutu spirant flamina nutriunt nubila germinant semina crescunt germinante tuam maiestatem perhorrescunt aues caelo meantes ferae montibus errantes serpentes solo latentes beluae ponto natantes
//
//topic 2:	0.995036.3	Apuleius Madaurensis, Metamorphoses 1.2.5	 postquam ardua montium et lubrica uallium et roscida cespitum et glebosa camporum emensus emersi in equo indigena peralbo uehens iam eo quoque admodum fesso ut ipse etiam fatigationem sedentariam incessus uegetatione discuterem in pedes desilio equi sudorem fronde detergeo frontem curiose exfrico auris remulceo frenos detraho in gradum lenem sensim proueho quoad lassitudinis incommodum alui solitum ac naturale praesidium eliquaret
//
//topic 3:	0.992108.3	Apuleius Madaurensis, Metamorphoses 1.6.9	 at uero domi tuae iam defletus et conclamatus es liberis tuis tutores iuridici prouincialis decreto dati uxor persolutis feralibus officiis luctu et maerore diuturno deformata diffletis paene ad extremam captiuitatem oculis suis domus infortunium nouarum nuptiarum gaudiis a suis sibi parentibus hilarare compellitur
//
//topic 4:	0.996159.3	Apuleius Madaurensis, Metamorphoses 10.20.5	 quattuor eunuchi confestim puluillis compluribus uentose tumentibus pluma delicata terrestrem nobis cubitum praestruunt sed et stragula ueste auro ac murice tyrio depicta probe consternunt ac desuper breuibus admodum sed satis copiosis puluillis aliis nimis modicis quis maxillas et ceruices delicatae mulieres suffulcire consuerunt superstruunt
//
//topic 5:	0.992793.3	Apuleius Madaurensis, Metamorphoses 11.16.28	 tunc cuncti populi tam religiosi quam profani uannos onustas aromatis et huiusce modi suppliciis certatim congerunt et insuper fluctus libant intritum lacte confectum donec muneribus largis et deuotionibus faustis completa nauis absoluta strophiis ancoralibus peculiari serenoque flatu pelago redderetur
