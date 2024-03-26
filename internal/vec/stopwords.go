//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vec

import (
	"encoding/json"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"os"
	"sort"
)

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
