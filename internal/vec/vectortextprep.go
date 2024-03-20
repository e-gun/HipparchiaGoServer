//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vec

import (
	"encoding/json"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/str"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"math/rand"
	"os"
	"sort"
	"strings"
)

//
// PARSEMAPPERS
//

//
// this code uses many large maps; it is possible that https://github.com/dolthub/swiss might one day speed this up
//

// note that building these can be quite slow when you send 1m lines into the modeler
// it might be possible to speed them up, but they are still <25% of the model building time so...

type hwguesser struct {
	total int
	words map[int]string
}

// buildmontecarloparsemap
func buildmontecarloparsemap(parsemap map[string]map[string]bool) map[string]hwguesser {
	// turn a list of sentences into a list of headwords; here we figure out the chances of any given homonym
	// then we set ourselves up to do a weighted guess of which one is in use

	// if a word might be A, B, or C and A appears 50 times, B appears 25 times, and C appears 5 times, then you
	// want to randomly assign the word to A 5/8 of the time, etc.
	// this can be achieved by a cumulative weight: [A -> 1-50, B -> 51-75, C -> 76-80]; a guess of 66 is a "B"

	// [a] figure out all headwords in use

	allheadwords := make(map[string]bool)
	for i := range parsemap {
		for k, _ := range parsemap[i] {
			allheadwords[k] = true
		}
	}

	// [b] generate scoremap and assign scores to each of the headwords

	scoremap := db.FetchHeadwordCounts(allheadwords)

	// [c] note that there are capital words in the parsemap that need lowering

	// [c1] lower the internal values first
	for i := range parsemap {
		newmap := make(map[string]bool)
		for k, _ := range parsemap[i] {
			newmap[strings.ToLower(k)] = true
		}
		parsemap[i] = newmap
	}

	// [c2] lower the parsemap keys; how worried should we be about the collisions...
	lcparsemap := make(map[string]map[string]bool)
	for i := range parsemap {
		lcparsemap[strings.ToLower(i)] = parsemap[i]
	}

	// [d] run through the parsemap convert to a hwguesser map

	guessermap := make(map[string]hwguesser)
	for i := range lcparsemap {
		var g hwguesser
		g.words = make(map[int]string)
		t := 0
		for j, _ := range parsemap[i] {
			t += scoremap[j]
			g.words[t] = j
		}
		g.total = t
		guessermap[i] = g
	}

	return guessermap
}

// buildwinnertakesallparsemap - figure out which is the most common of the possible headwords for any given word
func buildwinnertakesallparsemap(parsemap map[string]map[string]bool) map[string]string {
	// turn a list of sentences into a list of headwords; here we figure out which headword is the dominant homonym
	// then we just use that term; "esse" always comes from "sum" and never "edo", etc.

	// [a] figure out all headwords in use

	allheadwords := make(map[string]bool)
	for i := range parsemap {
		for k, _ := range parsemap[i] {
			allheadwords[k] = true
		}
	}

	// [b] generate scoremap and assign scores to each of the headwords

	scoremap := db.FetchHeadwordCounts(allheadwords)

	// [c] note that there are capital words in the parsemap that need lowering

	// [c1] lower the internal values first
	for i := range parsemap {
		newmap := make(map[string]bool)
		for k, _ := range parsemap[i] {
			newmap[strings.ToLower(k)] = true
		}
		parsemap[i] = newmap
	}

	// [c2] lower the parsemap keys; how worried should we be about the collisions...
	lcparsemap := make(map[string]map[string]bool)
	for i := range parsemap {
		lcparsemap[strings.ToLower(i)] = parsemap[i]
	}

	// [d] run through the parsemap and kill off the losers

	winnermap := make(map[string]string)
	for i := range lcparsemap {
		var hwl str.WHWList
		for j, _ := range parsemap[i] {
			var thishw str.WeightedHeadword
			thishw.Word = j
			thishw.Count = scoremap[j]
			hwl = append(hwl, thishw)
		}
		sort.Sort(hwl)
		winnermap[i] = hwl[0].Word
	}

	return winnermap
}

// buildyokedparsemap - set of sentences --> set of headwords; accept all headwords and yoke them: "esse" is "sum" + "edo", etc.
func buildyokedparsemap(parsemap map[string]map[string]bool) map[string]string {
	const (
		SEPARATOR = `ˣ` // wrong and nlp.NewPipeline(vectoriser, lda) will split yokedbags: spondeum•spondeus --> spondeum:25 spondeus:26
	)

	// [a] figure out all headwords in use

	allheadwords := make(map[string]bool)
	for i := range parsemap {
		for k, _ := range parsemap[i] {
			allheadwords[k] = true
		}
	}

	// [b] note that there are capital words in the parsemap that need lowering

	// [b1] lower the internal values first
	for i := range parsemap {
		newmap := make(map[string]bool)
		for k, _ := range parsemap[i] {
			newmap[strings.ToLower(k)] = true
		}
		parsemap[i] = newmap
	}

	// [b2] lower the parsemap keys; how worried should we be about the collisions...
	lcparsemap := make(map[string]map[string]bool)
	for i := range parsemap {
		lcparsemap[strings.ToLower(i)] = parsemap[i]
	}

	// [c] build the yoked map

	yoked := make(map[string]string)
	for i := range lcparsemap {
		var ww []string
		for j, _ := range parsemap[i] {
			ww = append(ww, j)
		}
		sort.Strings(ww)

		yoked[i] = strings.Join(ww, SEPARATOR)
	}

	return yoked
}

//
// buildtextblock() HELPERS
//

// flatstring - helper for buildtextblock() to generate unmodified text
func flatstring(sb *strings.Builder, slicedwords []string) {
	stops := getstopset()
	for i := 0; i < len(slicedwords); i++ {
		// drop skipwords
		_, s := stops[slicedwords[i]]
		if s {
			continue
		} else {
			sb.WriteString(slicedwords[i] + " ")
		}
	}

	for i := 0; i < len(slicedwords); i++ {
		sb.WriteString(slicedwords[i] + " ")
	}
}

// yokedstring - helper for buildtextblock() to generate conjoined string substitutions
func yokedstring(sb *strings.Builder, slicedwords []string, yokedmap map[string]string, stops map[string]struct{}) {
	// exact same logic as winnerstring()
	winnerstring(sb, slicedwords, yokedmap, stops)
}

// winnerstring - helper for buildtextblock() to generate winner takes all substitutions
func winnerstring(sb *strings.Builder, slicedwords []string, winnermap map[string]string, stops map[string]struct{}) {
	for i := 0; i < len(slicedwords); i++ {
		// drop skipwords
		w := winnermap[slicedwords[i]]
		_, s := stops[w]
		if s {
			continue
		} else {
			sb.WriteString(w + " ")
		}
	}
}

// montecarlostring - helper for buildtextblock() to generate lucky-ducky substitutions
func montecarlostring(sb *strings.Builder, slicedwords []string, guessermap map[string]hwguesser, stops map[string]struct{}) {
	var w string
	for i := 0; i < len(slicedwords); i++ {
		w = ""
		// pick a word...
		mc := guessermap[slicedwords[i]]
		if mc.total > 0 {
			g := rand.Intn(mc.total)
			for k, v := range mc.words {
				if g < k {
					w = v
					break
				}
			}
		} else {
			// just grab the first one
			for _, v := range mc.words {
				w = v
				break
			}
		}

		if w == "" {
			w = slicedwords[i]
		}

		_, s := stops[w]
		if s {
			continue
		} else {
			sb.WriteString(w + " ")
		}
	}
}

//
// STOPWORDS
//

// readstopconfig - read the vv.CONFIGVECTORSTOP file and return []stopwords; if it does not exist, generate it
func readstopconfig(fn string) []string {
	const (
		ERR1 = "readstopconfig() cannot find UserHomeDir"
		ERR2 = "readstopconfig() failed to parse "
		MSG1 = "readstopconfig() wrote vector stop configuration file: "
	)

	var stops []string
	var vcfg string

	switch fn {
	case "latin":
		vcfg = vv.CONFIGVECTORSTOPSLAT
		stops = gen.StringMapKeysIntoSlice(getlatinstops())
	case "greek":
		vcfg = vv.CONFIGVECTORSTOPSGRK
		stops = gen.StringMapKeysIntoSlice(getgreekstops())
	}

	h, e := os.UserHomeDir()
	if e != nil {
		Msg.MAND(ERR1)
		return stops
	}

	_, yes := os.Stat(fmt.Sprintf(vv.CONFIGALTAPTH, h) + vcfg)

	if yes != nil {
		sort.Strings(stops)
		content, err := json.MarshalIndent(stops, vv.JSONINDENT, vv.JSONINDENT)
		Msg.EC(err)

		err = os.WriteFile(fmt.Sprintf(vv.CONFIGALTAPTH, h)+vcfg, content, vv.WRITEPERMS)
		Msg.EC(err)
		Msg.PEEK(MSG1 + vcfg)
	} else {
		loadedcfg, _ := os.Open(fmt.Sprintf(vv.CONFIGALTAPTH, h) + vcfg)
		decoderc := json.NewDecoder(loadedcfg)
		var stp []string
		errc := decoderc.Decode(&stp)
		_ = loadedcfg.Close()
		if errc != nil {
			Msg.CRIT(ERR2 + vcfg)
		} else {
			stops = stp
		}
	}
	return stops
}

var (
	// Latin100 - the 100 most common latin headwords
	Latin100 = []string{"qui¹", "et", "in", "edo¹", "is", "sum¹", "hic", "non", "ab", "ut", "Cos²", "si", "ad", "cum", "ex", "a", "eo¹",
		"ego", "quis¹", "tu", "Eos", "dico²", "ille", "sed", "de", "neque", "facio", "possum", "atque", "sui", "res",
		"quam", "aut", "ipse", "huc", "habeo", "do", "omne", "video", "ito", "magnus", "b", "alius²", "for", "idem",
		"suum", "etiam", "per", "enim", "omnes", "ita", "suus", "omnis", "autem", "vel", "vel", "Alius¹", "qui²", "quo",
		"nam", "bonus", "neo¹", "meus", "volo¹", "ne³", "ne¹", "suo", "verus", "pars", "reor", "sua", "vaco", "verum",
		"primus", "unus", "multus", "causa", "jam", "tamen", "Sue", "nos", "dies", "Ios", "modus", "tuus", "venio",
		"pro¹", "pro²", "ago", "deus", "annus", "locus", "homo", "pater", "eo²", "tantus", "fero", "quidem", "noster",
		"an", "locum"}
	LatExtra = []string{"at", "o", "tum", "tunc", "dum", "illic", "quia", "sive", "num", "adhuc", "tam", "ibi", "cur",
		"usquam", "quoque", "duo", "talis", "simul", "igitur", "utique²", "aliqui", "apud", "sic", "umquam", "ergo",
		"ob", "xu", "x", "iii", "u", "post", "ac", "ut", "totus", "iste", "sue", "ceter", "inter", "eos"}
	LatStop = append(Latin100, LatExtra...)
	// LatinKeep - members of LatStop we will not toss
	LatinKeep = []string{"facio", "possum", "habeo", "video", "magnus", "bonus", "volo¹", "primus", "venio", "ago",
		"deus", "annus", "locus", "pater", "fero"}
	// Greek150 - the 150 most common greek headwords
	Greek150 = []string{"ὁ", "καί", "τίϲ", "ἔδω", "δέ", "εἰμί", "δέω¹", "δεῖ", "δέομαι", "εἰϲ", "αὐτόϲ", "τιϲ", "οὗτοϲ", "ἐν",
		"γάροϲ", "γάρον", "γάρ", "οὐ", "μένω", "μέν", "τῷ", "ἐγώ", "ἡμόϲ", "κατά", "Ζεύϲ", "ἐπί", "ὡϲ", "διά",
		"πρόϲ", "προϲάμβ", "τε", "πᾶϲ", "ἐκ", "ἕ", "ϲύ", "Ἀλλά", "γίγνομαι", "ἁμόϲ", "ὅϲτιϲ", "ἤ¹", "ἤ²", "ἔχω",
		"ὅϲ", "μή", "ὅτι¹", "λέγω¹", "ὅτι²", "τῇ", "Τήιοϲ", "ἀπό", "εἰ", "περί", "ἐάν", "θεόϲ", "φημί", "ἐκάϲ",
		"ἄν¹", "ἄνω¹", "ἄλλοϲ", "qui¹", "πηρόϲ", "παρά", "ἀνά", "αὐτοῦ", "ποιέω", "ἄναξ", "ἄνα", "ἄν²", "πολύϲ",
		"οὖν", "λόγοϲ", "οὕτωϲ", "μετά", "ἔτι", "ὑπό", "ἑαυτοῦ", "ἐκεῖνοϲ", "εἶπον", "πρότεροϲ", "edo¹", "μέγαϲ",
		"ἵημι", "εἷϲ", "οὐδόϲ", "οὐδέ", "ἄνθρωποϲ", "ἠμί", "μόνοϲ", "κύριοϲ", "διό", "οὐδείϲ", "ἐπεί", "πόλιϲ",
		"τοιοῦτοϲ", "χάω", "καθά", "θεάομαι", "γε", "ἕτεροϲ", "δοκέω", "λαμβάνω", "δή", "δίδωμι", "ἵνα",
		"βαϲιλεύϲ", "φύϲιϲ", "ἔτοϲ", "πατήρ", "ϲῶμα", "καλέω", "ἐρῶ", "υἱόϲ", "ὅϲοϲ", "γαῖα", "οὔτε", "οἷοϲ",
		"ἀνήρ", "ὁράω", "ψυχή", "Ἔχιϲ", "ὥϲπερ", "αὐτόϲε", "χέω", "ὑπέρ", "ϲόϲ", "θεάω", "νῦν", "ἐμόϲ", "δύναμαι",
		"φύω", "πάλιν", "ὅλοξ", "ἀρχή", "καλόϲ", "δύναμιϲ", "πωϲ", "δύο", "ἀγαθόϲ", "οἶδα", "δείκνυμι", "χρόνοϲ",
		"ὅμοιοϲ", "ἕκαϲτοϲ", "ὁμοῖοϲ", "ὥϲτε", "ἡμέρα", "γράφω", "δραχμή", "μέροϲ"}
	GreekExtra = []string{"ἀεί", "ὡϲαύτωϲ", "μηδέποτε", "μηδέ", "μηδ", "μηδέ", "ταὐτόϲ", "νυνί", "μεθ", "ἀντ", "μέχρι",
		"ἄνωθεν", "ὀκτώ", "ἓξ", "μετ", "τ", "μ", "αὐτόθ", "οὐδ", "εἵνεκ", "νόϲφι", "ἐκεῖ", "οὔκουν", "θ", "μάλιϲτ", "ὧδε",
		"πη", "τῇδ", "δι", "πρό", "ἀλλ", "ἕνεκα", "δ", "ἀλλά", "ἔπειτα", "καθ", "ταῦθ", "μήποτ", "ἀπ", "κ", "μήτ",
		"εὖτ", "αὖθιϲ", "∙∙∙", "∙∙", "∙", "∙∙∙∙", "oxy", "col", "fr", "*", "ϲύν", "ὅδε", "γ", "μέντοι", "εἶμι", "τότε",
		"ποτέ", "ὅταν", "πάνυ", "ἐπ", "πού", "οὐκοῦν", "παρ", "ὅπωϲ", "μᾶλλον", "μηδείϲ", "νή", "μήτε", "ἅπαϲ", "τοίνυν",
		"τοίνυν", "ἄρα", "αὖ", "εἴτε", "ἅμα", "ἆρ", "εὖ", "ϲχεδόν"}
	GreekStop = append(Greek150, GreekExtra...)
	// GreekKeep - members of GreekStop we will not toss
	GreekKeep = []string{"ἔχω", "λέγω¹", "θεόϲ", "φημί", "ποιέω", "ἵημι", "μόνοϲ", "κύριοϲ", "πόλιϲ", "θεάομαι", "δοκέω", "λαμβάνω",
		"δίδωμι", "βαϲιλεύϲ", "φύϲιϲ", "ἔτοϲ", "πατήρ", "ϲῶμα", "καλέω", "ἐρῶ", "υἱόϲ", "γαῖα", "ἀνήρ", "ὁράω",
		"ψυχή", "δύναμαι", "ἀρχή", "καλόϲ", "δύναμιϲ", "ἀγαθόϲ", "οἶδα", "δείκνυμι", "χρόνοϲ", "γράφω", "δραχμή",
		"μέροϲ", "λόγοϲ"}
)

func getgreekstops() map[string]struct{} {
	gs := gen.SetSubtraction(GreekStop, GreekKeep)
	return gen.ToSet(gs)
}

func getlatinstops() map[string]struct{} {
	ls := gen.SetSubtraction(LatStop, LatinKeep)
	return gen.ToSet(ls)
}

func getstopset() map[string]struct{} {
	ls := readstopconfig("latin")
	gs := readstopconfig("greek")
	ss := append(gs, ls...)
	return gen.ToSet(ss)
}
