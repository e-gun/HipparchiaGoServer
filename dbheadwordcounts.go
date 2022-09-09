package main

import (
	"context"
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"sort"
	"strings"
)

var (
	CORPUSWEIGTING = map[string]float32{"Ⓖ": 1.0, "Ⓛ": 12.7, "Ⓘ": 15.19, "Ⓓ": 18.14, "Ⓒ": 85.78}
)

//
// HEADWORDS
//

func headwordinfo(wc DbHeadwordCount) {

}

func headwordprevalence(wc DbHeadwordCount) string {
	// Prevalence (all forms): Ⓖ 95,843 / Ⓛ 10 / Ⓘ 151 / Ⓓ 751 / Ⓒ 64 / Ⓣ 96,819
	m := message.NewPrinter(language.English)

	cv := wc.CorpVal

	// sort.Slice(cv, func(i, j int) bool { return cv[i].count < cv[j].count })

	var pd []string

	for _, c := range cv {
		pd = append(pd, m.Sprintf("%s %d", c.name, c.count))
	}
	pd = append(pd, m.Sprintf("%s %d", "Ⓣ", wc.Total))

	p := "Prevalence (all forms): " + strings.Join(pd, " / ")

	return p
}

func headworddistrib(wc DbHeadwordCount) string {
	// Weighted distribution by corpus: Ⓖ 100 / Ⓓ 14 / Ⓒ 6 / Ⓘ 2 / Ⓛ 0
	cv := wc.CorpVal

	for i, c := range cv {
		cv[i].count = int(float32(c.count) * CORPUSWEIGTING[c.name])
	}

	sort.Slice(cv, func(i, j int) bool { return cv[i].count > cv[j].count })

	max := cv[0].count
	var pd []string
	for _, c := range cv {
		cpt := (float32(c.count) / float32(max)) * 100
		pd = append(pd, fmt.Sprintf("%s %d", c.name, int(cpt)))
	}

	p := "<br>Weighted distribution by corpus: " + strings.Join(pd, " / ")
	return p
}

func headwordchronology(wc DbHeadwordCount) {
	// Weighted chronological distribution: ⓔ 100 / ⓛ 84 / ⓜ 62
}

func headwordgenres(wc DbHeadwordCount) {
	// Predominant genres: comm (100), mech (97), jurisprud (93), med (84), mus (75), nathist (61), paroem (60), allrelig (57)
}

// HWData - to help sort values inside DbHeadwordCount
type HWData struct {
	name  string
	count int
}

type DbHeadwordTimeCounts struct {
	Early  int
	Middle int
	Late   int
}

type DbHeadwordCorpusCounts struct {
	TGrk int
	TLat int
	TDP  int
	TIN  int
	TCh  int
}

type DbHeadwordGenreCounts struct {
	Agric    int
	Alchem   int
	Anthol   int
	Astrol   int
	Astron   int
	Biogr    int
	Bucol    int
	Chron    int
	Comic    int
	Comm     int
	Concil   int
	Coq      int
	Dial     int
	Docu     int
	Doxog    int
	Eleg     int
	Epic     int
	Epigr    int
	Epist    int
	Exeg     int
	Fab      int
	Geog     int
	Gnom     int
	Gram     int
	Hexam    int
	Hist     int
	Hymn     int
	Hypoth   int
	Iamb     int
	Ignot    int
	Inscr    int
	Juris    int
	Lexic    int
	Lyr      int
	Magica   int
	Math     int
	Mech     int
	Med      int
	Meteor   int
	Mim      int
	Mus      int
	Myth     int
	NarrFic  int
	NatHis   int
	Onir     int
	Orac     int
	Paradox  int
	Parod    int
	Paroem   int
	Perig    int
	Phil     int
	Physiog  int
	Poem     int
	Polyhist int
	Pseud    int
	Satura   int
	Satyr    int
	Schol    int
	Tact     int
	Test     int
	Trag     int
	AllRhet  int
	AllRelig int
}

type DbHeadwordRhetoricaCounts struct {
	Encom  int
	Invect int
	Orat   int
	Rhet   int
}

type DbHeadwordTheologyCounts struct {
	Acta   int
	Apocal int
	Apocr  int
	Apol   int
	Caten  int
	Eccl   int
	Evang  int
	Hagiog int
	Homil  int
	Litur  int
	Proph  int
	Theol  int
}

type DbHeadwordCount struct {
	Entry     string
	Total     int
	FrqCla    string
	Chron     DbHeadwordTimeCounts
	Genre     DbHeadwordGenreCounts
	Corpus    DbHeadwordCorpusCounts
	Rhetorica DbHeadwordRhetoricaCounts
	Theology  DbHeadwordTheologyCounts
	CorpVal   []HWData
	TimeVal   []HWData
	TagVal    []HWData
}

func (hw *DbHeadwordCount) LoadCorpVals() {
	// Prevalence (all forms): Ⓖ 95,843 / Ⓛ 10 / Ⓘ 151 / Ⓓ 751 / Ⓒ 64 / Ⓣ 96,819
	var vv []HWData
	vv = append(vv, HWData{"Ⓖ", hw.Corpus.TGrk})
	vv = append(vv, HWData{"Ⓛ", hw.Corpus.TLat})
	vv = append(vv, HWData{"Ⓘ", hw.Corpus.TIN})
	vv = append(vv, HWData{"Ⓓ", hw.Corpus.TDP})
	vv = append(vv, HWData{"Ⓒ", hw.Corpus.TCh})
	hw.CorpVal = vv
}

func (hw *DbHeadwordCount) TimeVals() {
	// Weighted chronological distribution: ⓔ 100 / ⓛ 84 / ⓜ 62
	var vv []HWData
	vv = append(vv, HWData{"ⓔ", hw.Chron.Early})
	vv = append(vv, HWData{"ⓛ", hw.Chron.Late})
	vv = append(vv, HWData{"ⓜ", hw.Chron.Middle})
	hw.TimeVal = vv
}

func headwordlookup(word string) DbHeadwordCount {
	// scan a headwordcount into the corresponding struct
	// note that if you reassign a genre, this is one of the place you have to edit

	qt := `
	SELECT
		entry_name , total_count, gr_count, lt_count, dp_count, in_count, ch_count,
		frequency_classification, early_occurrences, middle_occurrences ,late_occurrences,
		acta, agric, alchem, anthol, apocalyp, apocryph, apol, astrol, astron, biogr, bucol,
		caten, chronogr, comic, comm, concil, coq, dialog, docu, doxogr, eccl, eleg, encom, epic,
		epigr, epist, evangel, exeget, fab, geogr, gnom, gramm, hagiogr, hexametr, hist, homilet,
		hymn, hypoth, iamb, ignotum, invectiv, inscr, jurisprud, lexicogr, liturg, lyr, magica, 
		math, mech, med, metrolog, mim, mus, myth, narrfict, nathist, onir, orac, orat,
		paradox, parod, paroem, perieg, phil, physiognom, poem, polyhist, prophet, pseudepigr, rhet,
		satura, satyr, schol, tact, test, theol, trag
	FROM dictionary_headword_wordcounts WHERE entry_name='%s'`

	dbpool := GetPSQLconnection()
	defer dbpool.Close()

	q := fmt.Sprintf(qt, word)

	foundrows, err := dbpool.Query(context.Background(), q)
	chke(err)

	var thesefinds []DbHeadwordCount
	defer foundrows.Close()
	for foundrows.Next() {
		var thehit DbHeadwordCount
		// "cannot assign 16380 into *main.DbHeadwordTimeCounts"
		var co DbHeadwordCorpusCounts
		var chr DbHeadwordTimeCounts
		var g DbHeadwordGenreCounts
		var th DbHeadwordTheologyCounts
		var rh DbHeadwordRhetoricaCounts

		err := foundrows.Scan(&thehit.Entry, &thehit.Total, &co.TGrk, &co.TLat, &co.TDP, &co.TIN, &co.TCh,
			&thehit.FrqCla, &chr.Early, &chr.Middle, &chr.Late,
			&th.Acta, &g.Agric, &g.Alchem, &g.Anthol, &th.Apocal, &th.Apocr, &th.Apol, &g.Astrol, &g.Astron, &g.Biogr, &g.Bucol,
			&th.Caten, &g.Chron, &g.Comic, &g.Comm, &g.Concil, &g.Coq, &g.Dial, &g.Docu, &g.Doxog, &th.Eccl, &g.Eleg, &rh.Encom, &g.Epic,
			&g.Epigr, &g.Epist, &th.Evang, &g.Exeg, &g.Fab, &g.Geog, &g.Gnom, &g.Gram, &th.Hagiog, &g.Hexam, &g.Hist, &th.Homil,
			&g.Hymn, &g.Hypoth, &g.Iamb, &g.Ignot, &rh.Invect, &g.Inscr, &g.Juris, &g.Lexic, &th.Litur, &g.Lyr, &g.Magica,
			&g.Math, &g.Mech, &g.Med, &g.Meteor, &g.Mim, &g.Mus, &g.Myth, &g.NarrFic, &g.NatHis, &g.Onir, &g.Orac, &rh.Orat,
			&g.Paradox, &g.Parod, &g.Paroem, &g.Perig, &g.Phil, &g.Physiog, &g.Poem, &g.Polyhist, &th.Proph, &g.Pseud, &rh.Rhet,
			&g.Satura, &g.Satyr, &g.Schol, &g.Tact, &g.Test, &th.Theol, &g.Trag)
		chke(err)
		thehit.Corpus = co
		thehit.Chron = chr
		thehit.Genre = g
		thehit.Rhetorica = rh
		thehit.Theology = th
		thesefinds = append(thesefinds, thehit)
	}

	var thefind DbHeadwordCount
	if len(thesefinds) == 1 {
		thefind = thesefinds[0]
	} else {
		msg(fmt.Sprintf("headwordlookup() for %s returned %d finds: this is wrong", word, len(thesefinds)), 1)
	}

	// fmt.Println(thefind)
	thefind.LoadCorpVals()
	return thefind
}

/*
	greekworderaweights = {'early': 6.93, 'middle': 1.87, 'late': 1}

	corporaweights = {'gr': 1.0, 'lt': 12.7, 'in': 15.19, 'dp': 18.14, 'ch': 85.78}

	greekgenreweights = {'acta': 85.38, 'alchem': 72.13, 'anthol': 17.68, 'apocalyp': 117.69, 'apocryph': 89.77,
	                     'apol': 7.0, 'astrol': 20.68, 'astron': 44.72, 'biogr': 6.39, 'bucol': 416.66, 'caten': 5.21,
	                     'chronogr': 4.55, 'comic': 29.61, 'comm': 1.0, 'concil': 16.75, 'coq': 532.74, 'dialog': 7.1,
	                     'docu': 2.66, 'doxogr': 130.84, 'eccl': 7.57, 'eleg': 188.08, 'encom': 13.17, 'epic': 19.36,
	                     'epigr': 10.87, 'epist': 4.7, 'evangel': 118.66, 'exeget': 1.24, 'fab': 140.87,
	                     'geogr': 10.74, 'gnom': 88.54, 'gramm': 8.65, 'hagiogr': 22.83, 'hexametr': 110.78,
	                     'hist': 1.44, 'homilet': 6.87, 'hymn': 48.18, 'hypoth': 12.95, 'iamb': 122.22,
	                     'ignotum': 122914.2, 'invectiv': 238.54, 'inscr': 1.91, 'jurisprud': 51.42, 'lexicogr': 4.14,
	                     'liturg': 531.5, 'lyr': 213.43, 'magica': 85.38, 'math': 9.91, 'mech': 103.44, 'med': 2.25,
	                     'metrolog': 276.78, 'mim': 2183.94, 'mus': 96.32, 'myth': 201.78, 'narrfict': 14.62,
	                     'nathist': 9.67, 'onir': 145.15, 'orac': 240.47, 'orat': 6.67, 'paradox': 267.32,
	                     'parod': 831.51, 'paroem': 65.58, 'perieg': 220.38, 'phil': 3.69, 'physiognom': 628.77,
	                     'poem': 62.82, 'polyhist': 24.91, 'prophet': 95.51, 'pseudepigr': 611.65, 'rhet': 8.67,
	                     'satura': 291.58, 'satyr': 96.78, 'schol': 5.56, 'tact': 52.01, 'test': 66.53, 'theol': 6.28,
	                     'trag': 35.8, 'allrelig': 0.58, 'allrhet': 2.9}

	latingenreweights = {'agric': 5.27, 'astron': 17.15, 'biogr': 9.88, 'bucol': 40.39, 'comic': 4.21, 'comm': 2.25,
	                     'coq': 60.0, 'dialog': 1134.73, 'docu': 6.19, 'eleg': 8.35, 'encom': 404.6, 'epic': 2.37,
	                     'epigr': 669.3, 'epist': 2.06, 'fab': 25.4, 'gnom': 147.23, 'gramm': 5.74, 'hexametr': 20.06,
	                     'hist': 1.0, 'hypoth': 762.59, 'ignotum': 586.58, 'inscr': 1.29, 'jurisprud': 1.11,
	                     'lexicogr': 27.71, 'lyr': 24.76, 'med': 7.26, 'mim': 1045.69, 'narrfict': 11.7,
	                     'nathist': 1.94, 'orat': 1.81, 'parod': 339.23, 'phil': 2.3, 'poem': 14.34,
	                     'polyhist': 4.75, 'rhet': 2.71, 'satura': 23.0, 'tact': 37.6, 'trag': 13.29, 'allrelig': 0,
	                     'allrhet': 1.08}
*/
