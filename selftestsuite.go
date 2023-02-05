package main

import (
	"fmt"
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

func (t *SrchTest) U() string {
	srch := fmt.Sprintf(t.s, "testing", t.t1, t.t2)
	return fmt.Sprintf("http://%s:%d/%s", Config.HostIP, Config.HostPort, srch)
}

func (t *SrchTest) M() string {
	return fmt.Sprintf(t.m, t.t1, t.t2)
}

func selftest() {
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
		MSG5  = "lemma near lemma in corpus: '%s' near '%s"
		MSG7  = "build a text for %d arbitrary lines"
		MSG8  = "build an index to %d arbitrary lines"
		MSG9  = "build vocabulary list for %d arbitrary lines"
		MSG10 = "browse 50 passages"
		MSG11 = "look up %d specific words"
		MSG12 = "look up %d word substrings"
		MSG13 = "reverse lookup for %d word substrings"
	)

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
			t1: "Πόλιϲ",
			t2: "ὁπλίζω",
			s:  LEM3,
			m:  MSG5,
		},
	}

	ll := Config.LogLevel
	Config.LogLevel = MSGFYI

	time.Sleep(WSPOLLINGPAUSE * 3)
	msg("entering selftest mode (3 segments)", MSGMAND)

	start := time.Now()
	previous := time.Now()

	msg("[I] 6 search tests", MSGNOTE)
	for i := 0; i < len(st); i++ {
		_, err := http.Get(st[i].U())
		chke(err)
		TimeTracker(st[i].id, st[i].M(), start, previous)
		previous = time.Now()
	}

	msg("[II] 3 text, index, and vocab maker tests", MSGNOTE)
	u := fmt.Sprintf("http://%s:%d/", Config.HostIP, Config.HostPort)
	_, err := http.Get(u + TXT)
	chke(err)
	TimeTracker("C1", fmt.Sprintf(MSG7, Config.MaxText), start, previous)
	previous = time.Now()

	_, err = http.Get(u + IDX)
	chke(err)
	TimeTracker("C2", fmt.Sprintf(MSG8, Config.MaxText), start, previous)
	previous = time.Now()

	_, err = http.Get(u + VOC)
	chke(err)
	TimeTracker("C3", fmt.Sprintf(MSG9, Config.MaxText), start, previous)
	previous = time.Now()

	msg("[III] 4 browsing and lexical tests", MSGNOTE)

	br := "browse/index/gr00%d/001/%d"
	for i := 0; i < 50; i++ {
		_, err = http.Get(u + fmt.Sprintf(br, i+10, 100))
	}
	TimeTracker("D1", MSG10, start, previous)
	previous = time.Now()

	wds := "ob eiusdem hominis consulatum una cum salute obtinendum et ut vestrae mentes atque sententiae cum populi "
	wds += "Romani voluntatibus suffragiisque consentiant eaque res vobis populoque"
	wds += "Περὶ μὲν τῶν κατηγορημένων ὦ ἄνδρεϲ δικαϲταί ἱκανῶϲ ὑμῖν ἀποδέδεικται ἀκοῦϲαι δὲ καὶ περὶ τῶν ἄλλων ὑμᾶϲ ἀξιῶ"
	wds += "ἐνίκηϲα καὶ ἀνήλωϲα ϲὺν τῇ τοῦ τρίποδοϲ ἀναθέϲει"

	lex := strings.Split(wds, " ")
	for i := 0; i < len(lex); i++ {
		_, err = http.Get(u + "lex/findbyform/" + lex[i] + "/test")
	}
	TimeTracker("D2", fmt.Sprintf(MSG11, len(lex)), start, previous)
	previous = time.Now()

	wds = "pud sud obse αφροδ γραμ ποικιλ"

	lex = strings.Split(wds, " ")
	for i := 0; i < len(lex); i++ {
		_, err = http.Get(u + "lex/lookup/" + lex[i])
	}
	TimeTracker("D3", fmt.Sprintf(MSG12, len(lex)), start, previous)
	previous = time.Now()

	wds = "love hate plague desire soldier horse"

	lex = strings.Split(wds, " ")
	for i := 0; i < len(lex); i++ {
		_, err = http.Get(u + "lex/reverselookup/testing/" + lex[i])
	}

	TimeTracker("D4", fmt.Sprintf(MSG13, len(lex)), start, previous)
	previous = time.Now()

	msg("exiting selftest mode", MSGMAND)
	Config.LogLevel = ll
}
