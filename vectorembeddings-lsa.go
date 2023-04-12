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
	"strings"
)

// "github.com/james-bowman/nlp" also contains some interesting possibilities: LatentDirichletAllocation, etc.
// bagging would need to go as per the old HipparchiaGoDBHelper code: sentence by sentence
// then you can model the topics; it is not immediately clear how to graph and interpret them, though
// would be possible to model; then grab the N sentences that most resemble the N topics

// this requires retaining locus info: the sentences get heavily rewritten, after all
// HipparchiaGoDBHelper knows how to do all of this...

type BagWithLocus struct {
	Loc         string
	Bag         string
	ModifiedBag string
}

func lsatest(c echo.Context) {
	vs := sessionintobulksearch(c, Config.VectorMaxlines)
	lsa(vs.Results)
}

func lsa(dblines []DbWorkline) {
	const (
		SENTENCESPERBAG = 1
		NUMBEROFTOPICS  = 4
	)

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
	pipeline := nlp.NewPipeline(vectoriser, lda)

	docsOverTopics, err := pipeline.FitTransform(corpus...)
	if err != nil {
		fmt.Printf("Failed to model topics for documents because %v", err)
		return
	}

	// Examine Document over topic probability distribution
	dr, dc := docsOverTopics.Dims()
	for doc := 0; doc < dc; doc++ {
		fmt.Printf("\nTopic distribution for document: '%s' -", corpus[doc])
		for topic := 0; topic < dr; topic++ {
			if topic > 0 {
				fmt.Printf(",")
			}
			fmt.Printf(" Topic #%d=%f", topic, docsOverTopics.At(topic, doc))
		}
	}

	// Examine Topic over word probability distribution
	topicsOverWords := lda.Components()
	tr, tc := topicsOverWords.Dims()

	vocab := make([]string, len(vectoriser.Vocabulary))
	for k, v := range vectoriser.Vocabulary {
		vocab[v] = k
	}
	for topic := 0; topic < tr; topic++ {
		fmt.Printf("\nWord distribution for Topic #%d -", topic)
		for word := 0; word < tc; word++ {
			if word > 0 {
				fmt.Printf(",")
			}
			fmt.Printf(" '%s'=%f", vocab[word], topicsOverWords.At(topic, word))
		}
	}
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
