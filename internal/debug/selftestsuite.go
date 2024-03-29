//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//

package debug

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/mm"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/vec"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strings"
	"time"
)

// time tests and profiling tests

type SrchTest struct {
	id string
	t1 string
	t2 string
	s  string
	m  string
}

func (t *SrchTest) Url() string {
	sid := "selftest-" + strings.Replace(uuid.New().String(), "-", "", -1)
	uri := fmt.Sprintf(t.s, sid, t.t1, t.t2)
	return fmt.Sprintf("http://%s:%d/%s", lnch.Config.HostIP, lnch.Config.HostPort, uri)
}

func (t *SrchTest) Msg() string {
	return fmt.Sprintf(t.m, strings.ReplaceAll(t.t1, "%20", " "), strings.ReplaceAll(t.t2, "%20", " "))
}

// RunSelfTests - loop selftestsuite()
func RunSelfTests() {
	if lnch.Config.SelfTest > 0 {
		go func() {
			Msg.SNm = vv.SHORTNAME + "-SELFTEST"
			for i := 0; i < lnch.Config.SelfTest; i++ {
				Msg.MAND(fmt.Sprintf("Running Selftest %d of %d", i+1, lnch.Config.SelfTest))
				selftestsuite()
			}
		}()
	}
}

// selftestsuite - iterate through a list of tests
func selftestsuite() {
	const (
		SKG1  = "srch/exec/%s?skg=%s%s"
		SKG2  = "srch/exec/%s?skg=%s&prx=%s"
		LEM1  = "srch/exec/%s?lem=%s%s"
		LEM2  = "srch/exec/%s?lem=%s&prx=%s"
		LEM3  = "srch/exec/%s?lem=%s&plm=%s"
		TXT   = "text/make/_"
		IDX   = "text/index/testing"
		VOC   = "text/vocab/testing"
		MSG1  = "single word in corpus: '%s'%s"
		MSG2  = "phrase in corpus: '%s'%s"
		MSG6  = "phrase near phrase: '%s' near '%s'"
		MSG3  = "lemma in corpus: '%s'%s"
		MSG4  = "lemma near phrase: '%s' near '%s'"
		MSG5  = "lemma near lemma in corpus: '%s' near '%s'"
		MSG7  = "build a text for %d arbitrary lines"
		MSG8  = "build an index to %d arbitrary lines"
		MSG9  = "build vocabulary list for %d arbitrary lines"
		MSG10 = "browse 50 passages"
		MSG11 = "look up %d specific words"
		MSG12 = "look up %d word substrings"
		MSG13 = "reverse lookup for %d word substrings"
		MSG14 = "semantic vector model test: %s - %d author(s) with %d text preparation modes per author"
		MSG15 = "lda vector model test - %d author(s) with %d text preparation modes per author"
		URL   = "http://%s:%d/vbot/%s/%s"
	)

	// NOTES ON SELFTEST MEMORY USE

	// [a] currently a non-vector selftest will go from its post-initialization value to something larger in the end;
	// this final figure will drop after a while: garbage collected; you will return to the base memory footprint.

	// [b] a vector selftest will go from its post-initialization value to something *much, much* larger in the end;
	// this final figure will drop *somewhat* after a while; you will never return to the base memory footprint.

	// [c] a series of non-selftest vector searches will consume a lot of RAM; this final figure will drop after a
	// while; you will return to the base memory footprint. So [b] is producing (matrix) objects that elude GC.
	// This is a bug. It has proven to be very elusive. But it is not a bug that affects real world use unless one
	// always runs in selftest mode...

	//go tool pprof heap.4.pprof
	//Type: inuse_space
	//Time: Sep 22, 2023 at 3:59pm (EDT)
	//Entering interactive mode (type "help" for commands, "o" for options)
	//(pprof) top20
	//Showing nodes accounting for 491.94MB, 97.80% of 503.01MB total
	//Dropped 87 nodes (cum <= 2.52MB)
	//Showing top 20 nodes out of 72
	//      flat  flat%   sum%        cum   cum%
	//  131.08MB 26.06% 26.06%   131.08MB 26.06%  github.com/e-gun/wego/pkg/model/modelutil/matrix.New

	stm := lnch.NewMessageMakerConfigured()
	stm.SNm = vv.SHORTNAME + "-SELFTEST"
	stm.LLvl = mm.MSGFYI

	st := []SrchTest{
		{
			id: "A1",
			t1: "vervex",
			t2: "",
			s:  SKG1,
			m:  MSG1,
		},
		{
			id: "A2",
			t1: "plato%20omnem",
			t2: "",
			s:  SKG1,
			m:  MSG2,
		},
		{
			id: "A3",
			t1: "καὶ%20δὴ%20καὶ",
			t2: "εἴ%20που%20καὶ",
			s:  SKG2,
			m:  MSG6,
		},
		{
			id: "B1",
			t1: "φθορώδηϲ",
			t2: "",
			s:  LEM1,
			m:  MSG3,
		},
		{
			id: "B2",
			t1: "γαῖα",
			t2: "ἐϲχάτη%20χθονόϲ",
			s:  LEM2,
			m:  MSG4,
		},
		{
			id: "B3",
			t1: "πόλιϲ",
			t2: "ὁπλίζω",
			s:  LEM3,
			m:  MSG5,
		},
	}

	start := time.Now()
	previous := time.Now()

	stm.Emit("entering selftestsuite mode (4 segments)", mm.MSGMAND)

	u := fmt.Sprintf("http://%s:%d/", lnch.Config.HostIP, lnch.Config.HostPort)

	tt := [5]bool{true, true, true, true, true}
	// tt := [5]bool{false, false, false, false, true}

	getter := func(u string) {
		res, e := http.Get(u)
		Msg.EC(e)
		// want to get rid of pprof: "54.13MB 19.12% 38.54%    55.87MB 19.73%  main.JSONresponse.func4"
		_, e = io.ReadAll(res.Body)
		e = res.Body.Close()
		Msg.EC(e)
	}

	// [I] 6 search tests
	if tt[0] {
		stm.Emit("[I] 6 search tests", mm.MSGWARN)
		for i := 0; i < len(st); i++ {
			getter(st[i].Url())
			stm.Timer(st[i].id, st[i].Msg(), start, previous)
			previous = time.Now()
		}
	}

	// [II] 3 text, index, and vocab maker tests
	if tt[1] {
		stm.Emit("[II] 3 text, index, and vocab maker tests", mm.MSGWARN)

		getter(u + TXT)
		stm.Timer("C1", fmt.Sprintf(MSG7, lnch.Config.MaxText), start, previous)
		previous = time.Now()

		getter(u + IDX)
		stm.Timer("C2", fmt.Sprintf(MSG8, lnch.Config.MaxText), start, previous)
		previous = time.Now()

		getter(u + VOC)
		stm.Timer("C3", fmt.Sprintf(MSG9, lnch.Config.MaxText), start, previous)
		previous = time.Now()
	}

	// [III] 4 browsing and lexical tests
	if tt[2] {
		stm.Emit("[III] 4 browsing and lexical tests", mm.MSGWARN)

		br := "browse/index/gr00%d/001/%d"
		for i := 0; i < 50; i++ {
			getter(u + fmt.Sprintf(br, i+10, 100))
		}
		stm.Timer("D1", MSG10, start, previous)
		previous = time.Now()

		wds := "ob eiusdem hominis consulatum una cum salute obtinendum et ut vestrae mentes atque sententiae cum populi "
		wds += "Romani voluntatibus suffragiisque consentiant eaque res vobis populoque"
		wds += "Περὶ μὲν τῶν κατηγορημένων ὦ ἄνδρεϲ δικαϲταί ἱκανῶϲ ὑμῖν ἀποδέδεικται ἀκοῦϲαι δὲ καὶ περὶ τῶν ἄλλων ὑμᾶϲ ἀξιῶ"
		wds += "ἐνίκηϲα καὶ ἀνήλωϲα ϲὺν τῇ τοῦ τρίποδοϲ ἀναθέϲει"

		lex := strings.Split(wds, " ")
		for i := 0; i < len(lex); i++ {
			getter(u + "lex/findbyform/" + lex[i] + "/test")
		}
		stm.Timer("D2", fmt.Sprintf(MSG11, len(lex)), start, previous)
		previous = time.Now()

		wds = "pud sud obse αφροδ γραμ ποικιλ pud sud obse αφροδ γραμ ποικιλ pud sud obse αφροδ γραμ ποικιλ"

		lex = strings.Split(wds, " ")
		for i := 0; i < len(lex); i++ {
			getter(u + "lex/lookup/" + lex[i])
		}
		stm.Timer("D3", fmt.Sprintf(MSG12, len(lex)), start, previous)
		previous = time.Now()

		wds = "love hate plague desire soldier horse"

		lex = strings.Split(wds, " ")
		for i := 0; i < len(lex); i++ {
			getter(u + "lex/reverselookup/testing/" + lex[i])
		}

		stm.Timer("D4", fmt.Sprintf(MSG13, len(lex)), start, previous)
		previous = time.Now()
	}

	if lnch.Config.VectorsDisabled {
		stm.Emit("exiting selftestsuite mode", mm.MSGMAND)
		return
	}

	// vector selftestsuite
	vec.VectorDBReset()
	ovm := lnch.Config.VectorModel
	otx := lnch.Config.VectorTextPrep

	// glove seizes scads of memory and never releases it; need to fix wego, though, it seems
	vmod := []string{"w2v", "lexvec", "glove"}
	vtxp := []string{"winner", "unparsed", "yoked", "montecarlo"}
	vauu := []string{"gr0011"} // sophocles

	// fnc for [iv.3]
	httpgetauthor := func(v string) {
		for _, a := range vauu {
			url := fmt.Sprintf(URL, lnch.Config.HostIP, lnch.Config.HostPort, v, a)
			getter(url)
		}
	}

	// fnc for [iv.2]
	preptext := func(v string) {
		for _, t := range vtxp {
			lnch.Config.VectorTextPrep = t
			httpgetauthor(v)
		}
	}

	// [IV] nearest neighbor vectorization tests
	if tt[3] {
		stm.Emit("[IV] nearest neighbor vectorization tests", mm.MSGWARN)

		// [iv.1]
		buildmodel := func(m string, count int) {
			lnch.Config.VectorModel = m
			preptext("nn")
			nb := fmt.Sprintf(MSG14, m, len(vauu), len(vtxp))
			stm.Timer(fmt.Sprintf("E%d", count), nb, start, previous)
			previous = time.Now()
		}

		// loop 1 -> 2 -> 3
		count := 1
		for _, m := range vmod {
			buildmodel(m, count)
			count++
		}
	}

	// [V] lda vectorization tests
	if tt[4] {
		stm.Emit("[V] lda vectorization tests", mm.MSGWARN)
		vauu = []string{"lt0472"} // catullus

		preptext("lda")
		nb := fmt.Sprintf(MSG15, len(vauu), len(vtxp))
		stm.Timer("F", nb, start, previous)
	}

	previous = time.Now()
	stm.MAND("exiting selftestsuite mode")

	lnch.Config.VectorModel = ovm
	lnch.Config.VectorTextPrep = otx
	lnch.Config.SelfTest = 0
}
