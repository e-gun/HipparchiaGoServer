package main

import (
	"strconv"
	"strings"
)

// searchlistmanagement.py has:
// compilesearchlist(), sortsearchlist(), sortresultslist(),
// calculatewholeauthorsearches(), flagexclusions(), prunebydate(), removespuria()

type Session struct {
	Inclusions SearchInclusions
	Exclusions SearchExclusions
	ActiveCorp map[string]bool
	VariaOK    bool
	IncertaOK  bool
	// unimplemented for now
	Querytype      string
	AvailDBs       map[string]bool
	VectorVals     bool
	UISettings     bool
	OutPutSettings bool
}

type SearchInclusions struct {
	AuGenres    []string
	WkGenres    []string
	AuLocations []string
	WkLocations []string
	Authors     []string
	Works       []string
	Passages    []string
	DateRange   [2]string
}

type SearchExclusions struct {
	AuGenres    []string
	WkGenres    []string
	AuLocations []string
	WkLocations []string
	Authors     []string
	Works       []string
	Passages    []string
}

func compilesearchlist(s Session, aa map[string]DbAuthor, ww map[string]DbWork) {

	// note that we do all the initial stuff by adding WORKS to the list individually
	var searchlist []string

	// [a] trim mappers by active corpora
	auu := make(map[string]DbAuthor)
	wkk := make(map[string]DbWork)

	for k, v := range s.ActiveCorp {
		for _, a := range aa {
			if a.UID[0:3] == k && v == true {
				auu[a.UID] = a
			}
		}
		for _, w := range ww {
			if w.UID[0:3] == k && v == true {
				wkk[w.UID] = w
			}
		}
	}

	// [b] build the inclusion list
	incl := s.Inclusions
	excl := s.Exclusions

	// [b1] author genres to include
	for _, g := range incl.AuGenres {
		for _, a := range auu {
			if strings.Contains(a.Genres, g) {
				searchlist = append(searchlist, a.WorkList...)
			}
		}
	}

	// [b2] work genres to include
	for _, g := range incl.WkGenres {
		for _, w := range wkk {
			if w.Genre == g {
				searchlist = append(searchlist, w.UID)
			}
		}
	}

	// [b3] author locations to include
	for _, l := range incl.AuLocations {
		for _, a := range auu {
			if a.Location == l {
				searchlist = append(searchlist, a.WorkList...)
			}
		}
	}

	// [b4] work locations to include
	for _, l := range incl.WkLocations {
		for _, w := range wkk {
			if w.Prov == l {
				searchlist = append(searchlist, w.UID)
			}
		}
	}

	// 		a tricky spot: when/how to apply prunebydate()
	//		if you want to be able to seek 5th BCE oratory and Plutarch, then you need to let auselections take precedence
	//		accordingly we will do classes and genres first, then trim by date, then add in individual choices

	// [b5] prune by date
	// 'varia' and 'incerta' have special dates: incerta = 2500; varia = 2000
	before, _ := strconv.Atoi(incl.DateRange[0])
	after, _ := strconv.Atoi(incl.DateRange[1])
	b := int64(before)
	a := int64(after)

	if b != MINDATE || a != MAXDATE {
		// should have already been validated elsewhere...
		if b > a {
			b = a
		}

		// [b5a] first prune the bad dates
		var trimmed []string
		for _, uid := range searchlist {
			if wkk[uid].DateInRange(b, a) {
				trimmed = append(trimmed, uid)
			}
		}

		// [b5b] add back in any varia and/or incerta as needed
		if s.VariaOK {
			for _, uid := range searchlist {
				if wkk[uid].ConvDate == VARIADATE {
					trimmed = append(trimmed, uid)
				}
			}
		}

		if s.IncertaOK {
			for _, uid := range searchlist {
				if wkk[uid].ConvDate == INCERTADATE {
					trimmed = append(trimmed, uid)
				}
			}
		}

		searchlist = trimmed
	}

	// [b6] add all works of the authors selected

	// [b7] add the individual works selected

	// [b8] add the individual passages selected

	// [c] subtract from the inclusions
}

func main() {
	authormap := authormapper()
	workmap := workmapper()
	authormap = loadworksintoauthors(authormap, workmap)
	workmap = dateworksviaauthors(authormap, workmap)

}

// compilesearchlist() - searching[]: ['lt0474', 'lt0917', 'Apologetici', 'Doxographi', 'Eleg.', 'gr0032w002_FROM_11313_TO_11843', 'gr0062w001', 'Abdera']
// compilesearchlist() - excluding[]: ['lt0474w001', 'lt0917w001_AT_3']
// searchlist ['gr0244w001', 'lt0474w053', 'lt0474w057', 'gr0645w003', 'lt0474w042', 'gr0002w001', 'gr1184w002', 'gr1725w001', 'lt0474w056', 'lt0474w015', 'gr0319w002', 'lt0474w072', 'lt0524w002', 'lt0917w002', 'lt0474w022', 'gr0251w001', 'gr0533w005', 'gr0714w001', 'gr0267w006', 'lt0474w055', 'gr0243w001', 'gr0213w001', 'gr0242w001', 'gr0258w001', 'lt0474w068', 'lt0474w029', 'gr0319w004', 'gr0239w002', 'gr0266w001', 'gr1390w003', 'lt0474w024', 'lt0474w011', 'lt0474w006', 'lt0474w007', 'gr0217w001', 'gr0255w001', 'lt0474w047', 'lt0474w033', 'lt0474w016', 'gr1304w002', 'lt0959w008', 'gr1184w001', 'lt0474w059', 'gr0255w003', 'lt0474w038', 'lt0474w010', 'lt0474w005', 'gr0531w001', 'lt0474w021', 'lt0474w054', 'lt0959w001', 'gr1604w001', 'gr0231w001', 'lt0660w001', 'gr0212w004', 'lt0474w020', 'lt0474w061', 'lt0474w009', 'lt0474w070', 'lt0474w028', 'gr1766w001', 'lt0474w052', 'gr0261w001', 'lt0474w031', 'gr0234w001', 'gr0247w001', 'gr2606w001', 'gr0714w002', 'gr1495w003', 'lt0917w001', 'gr0216w001', 'lt0474w026', 'gr2647w001', 'gr0085w009', 'lt0474w018', 'gr0254w001', 'gr1390w002', 'lt0474w030', 'gr0239w001', 'gr1675w002', 'gr0263w001', 'lt0474w065', 'lt0474w039', 'gr0218w002', 'gr1632w004', 'lt0474w032', 'gr0086w049', 'lt0474w064', 'lt0474w036', 'lt0474w060', 'gr0214w001', 'lt0474w014', 'gr0218w001', 'gr0212w002', 'gr0245w001', 'gr1163w001', 'gr0284w057', 'gr1273w001', 'gr1461w001', 'lt0474w017', 'gr0237w001', 'gr0232w001', 'gr1635w001', 'lt0474w049', 'gr0533w006', 'gr0198w002', 'gr1304w001', 'gr0246w001', 'gr0212w001', 'lt0474w069', 'gr0280w001', 'gr1635w002', 'lt0474w035', 'lt0474w063', 'gr0888w001', 'gr0655w003', 'gr1390w004', 'lt0474w041', 'gr0645w001', 'lt0474w066', 'lt0474w003', 'gr1184w003', 'gr0308w002', 'gr1992w003', 'lt0474w023', 'gr0002w002', 'gr0262w001', 'gr2611w001', 'lt0474w046', 'gr2605w001', 'gr1205w001', 'lt0474w051', 'gr2022w062', 'lt0474w037', 'gr0533w010', 'gr0266w002', 'gr1461w002', 'lt0474w050', 'gr1635w003', 'lt0474w025', 'gr1735w003', 'gr0336w002', 'gr0308w009', 'lt0474w027', 'gr0257w001', 'gr0645w002', 'gr0706w001', 'lt0474w043', 'lt0474w074', 'gr2153w001', 'lt0456w001', 'gr0062w001', 'lt0692w006', 'lt0959w010', 'gr1390w001', 'gr0676w001', 'gr1766w002', 'gr0676w002', 'gr2646w001', 'lt0474w013', 'lt0474w019', 'lt0474w058', 'gr0002w003', 'gr0267w001', 'gr0655w002', 'gr0236w001', 'lt0474w012', 'lt0474w073', 'gr1495w002', 'gr0336w004', 'lt0524w001', 'lt0474w034', 'gr0529w002', 'lt0474w002', 'gr0529w001', 'lt0474w008', 'lt0474w040', 'lt0474w004', 'lt0474w067', 'lt0474w062', 'gr0032w002_FROM_11313_TO_11843', 'lt0474w044', 'lt0474w075', 'gr0645w004', 'lt0620w001', 'gr0528w002', 'gr0528w001', 'lt0474w071', 'lt0959w002', 'lt0474w048', 'gr1205w002', 'lt0959w009', 'gr2648w002', 'gr0011w009', 'gr1495w001', 'lt0474w045', 'gr2694w001']

// {'_fresh': False, 'agnexclusions': [], 'agnselections': ['Apologetici', 'Doxographi'], 'alocexclusions': [], 'alocselections': ['Abdera'], 'analogyfinder': False, 'auexclusions': [], 'auselections': ['lt0474', 'lt0917'], 'authorflagging': True, 'authorssummary': True, 'available': {'greek_dictionary': True, 'greek_lemmata': True, 'greek_morphology': True, 'latin_dictionary': True, 'latin_lemmata': True, 'latin_morphology': True, 'wordcounts_0': True}, 'baggingmethod': 'winnertakesall', 'bracketangled': True, 'bracketcurly': True, 'bracketround': False, 'bracketsquare': True, 'browsercontext': '24', 'christiancorpus': True, 'collapseattic': True, 'cosdistbylineorword': False, 'cosdistbysentence': False, 'debugdb': False, 'debughtml': False, 'debuglex': False, 'debugparse': False, 'earliestdate': '-850', 'fontchoice': 'Noto', 'greekcorpus': True, 'headwordindexing': False, 'incerta': True, 'indexbyfrequency': False, 'indexskipsknownwords': False, 'inscriptioncorpus': True, 'latestdate': '1500', 'latincorpus': True, 'ldacomponents': 7, 'ldaiterations': 12, 'ldamaxfeatures': 2000, 'ldamaxfreq': 35, 'ldaminfreq': 5, 'ldamustbelongerthan': 3, 'linesofcontext': 4, 'loggedin': False, 'maxresults': '200', 'morphdialects': False, 'morphduals': True, 'morphemptyrows': True, 'morphfinite': True, 'morphimper': True, 'morphinfin': True, 'morphpcpls': True, 'morphtables': True, 'nearestneighborsquery': False, 'nearornot': 'near', 'onehit': False, 'papyruscorpus': True, 'phrasesummary': False, 'principleparts': True, 'proximity': '1', 'psgexclusions': ['lt0917w001_AT_3'], 'psgselections': ['gr0032w002_FROM_11313_TO_11843'], 'quotesummary': True, 'rawinputstyle': False, 'searchinsidemarkup': False, 'searchscope': 'lines', 'semanticvectorquery': False, 'sensesummary': True, 'sentencesimilarity': False, 'showwordcounts': True, 'simpletextoutput': False, 'sortorder': 'shortname', 'spuria': True, 'suppresscolors': False, 'tensorflowgraph': False, 'topicmodel': False, 'trimvectoryby': 'none', 'userid': 'Anonymous', 'varia': True, 'vcutlem': 50, 'vcutloc': 33, 'vcutneighb': 15, 'vdim': 300, 'vdsamp': 5, 'viterat': 12, 'vminpres': 10, 'vnncap': 15, 'vsentperdoc': 1, 'vwindow': 10, 'wkexclusions': ['lt0474w001'], 'wkgnexclusions': [], 'wkgnselections': ['Eleg.'], 'wkselections': ['gr0062w001'], 'wlocexclusions': [], 'wlocselections': [], 'xmission': 'Any', 'zaplunates': False, 'zapvees': False}

// [debugging] calling HipparchiaGoServer [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt0456 WHERE ( (index BETWEEN 1 AND 29) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt0524 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr2611 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr1184 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr1735 WHERE ( (index BETWEEN 6 AND 208) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0266 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr1635 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0336 WHERE ( (index BETWEEN 4 AND 77) OR (index BETWEEN 89 AND 164) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0645 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr2605 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0284 WHERE ( (index BETWEEN 38224 AND 38234) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0888 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0531 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0319 WHERE ( (index BETWEEN 443 AND 926) OR (index BETWEEN 206 AND 260) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0267 WHERE ( (index BETWEEN 601 AND 810) OR (index BETWEEN 1 AND 72) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0216 WHERE ( (index BETWEEN 1 AND 91) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0011 WHERE ( (index BETWEEN 15699 AND 15705) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0243 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr1992 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr1163 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt0620 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0218 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0002 WHERE ( (index BETWEEN 1432 AND 1439) OR (index BETWEEN 1440 AND 1467) OR (index BETWEEN 1 AND 1431) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0212 WHERE ( (index BETWEEN 1 AND 52) OR (index BETWEEN 211 AND 223) OR (index BETWEEN 53 AND 108) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr2606 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0242 WHERE ( (index BETWEEN 1 AND 4) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0706 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt0474 WHERE ( (index BETWEEN 86757 AND 93240) OR (index BETWEEN 44657 AND 51145) OR (index BETWEEN 28556 AND 29932) OR (index BETWEEN 98020 AND 99108) OR (index BETWEEN 149198 AND 149432) OR (index BETWEEN 19502 AND 22042) OR (index BETWEEN 42234 AND 43475) OR (index BETWEEN 38631 AND 40002) OR (index BETWEEN 99109 AND 100350) OR (index BETWEEN 33517 AND 35592) OR (index BETWEEN 74320 AND 76999) OR (index BETWEEN 142241 AND 142280) OR (index BETWEEN 149433 AND 149570) OR (index BETWEEN 3208 AND 3910) OR (index BETWEEN 143147 AND 143782) OR (index BETWEEN 138275 AND 140506) OR (index BETWEEN 37182 AND 37801) OR (index BETWEEN 141681 AND 142240) OR (index BETWEEN 142607 AND 142859) OR (index BETWEEN 93241 AND 98019) OR (index BETWEEN 143109 AND 143142) OR (index BETWEEN 23779 AND 24224) OR (index BETWEEN 3911 AND 16459) OR (index BETWEEN 24225 AND 25776) OR (index BETWEEN 51146 AND 55518) OR (index BETWEEN 79963 AND 80540) OR (index BETWEEN 142598 AND 142606) OR (index BETWEEN 30259 AND 30790) OR (index BETWEEN 36135 AND 37181) OR (index BETWEEN 77000 AND 77622) OR (index BETWEEN 142860 AND 143108) OR (index BETWEEN 70234 AND 71197) OR (index BETWEEN 143968 AND 143996) OR (index BETWEEN 22043 AND 23778) OR (index BETWEEN 109398 AND 123772) OR (index BETWEEN 40003 AND 41360) OR (index BETWEEN 64382 AND 67551) OR (index BETWEEN 44186 AND 44656) OR (index BETWEEN 143783 AND 143842) OR (index BETWEEN 30791 AND 32607) OR (index BETWEEN 32608 AND 33516) OR (index BETWEEN 100351 AND 104134) OR (index BETWEEN 27074 AND 28176) OR (index BETWEEN 144567 AND 149197) OR (index BETWEEN 71198 AND 74319) OR (index BETWEEN 143997 AND 144566) OR (index BETWEEN 143843 AND 143967) OR (index BETWEEN 104798 AND 109397) OR (index BETWEEN 25777 AND 27073) OR (index BETWEEN 41361 AND 41741) OR (index BETWEEN 43476 AND 43804) OR (index BETWEEN 104135 AND 104797) OR (index BETWEEN 143143 AND 143146) OR (index BETWEEN 55519 AND 62949) OR (index BETWEEN 35593 AND 36134) OR (index BETWEEN 142281 AND 142597) OR (index BETWEEN 67552 AND 70014) OR (index BETWEEN 140507 AND 141680) OR (index BETWEEN 16879 AND 17455) OR (index BETWEEN 29933 AND 30258) OR (index BETWEEN 37802 AND 38630) OR (index BETWEEN 2632 AND 3207) OR (index BETWEEN 70015 AND 70233) OR (index BETWEEN 123773 AND 138274) OR (index BETWEEN 28177 AND 28555) OR (index BETWEEN 80541 AND 86756) OR (index BETWEEN 17456 AND 18671) OR (index BETWEEN 1053 AND 2631) OR (index BETWEEN 62950 AND 64381) OR (index BETWEEN 43805 AND 44185) OR (index BETWEEN 16460 AND 16878) OR (index BETWEEN 77623 AND 79962) OR (index BETWEEN 18672 AND 19501) OR (index BETWEEN 41742 AND 42233) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr1632 WHERE ( (index BETWEEN 129 AND 165) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr2648 WHERE ( (index BETWEEN 1078 AND 1401) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0244 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr1766 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0263 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0262 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0198 WHERE ( (index BETWEEN 30 AND 37) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0280 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr1461 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0239 WHERE ( (index BETWEEN 24 AND 197) OR (index BETWEEN 1 AND 23) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt0660 WHERE ( (index BETWEEN 1 AND 1265) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0246 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0214 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0255 WHERE ( (index BETWEEN 1 AND 97) OR (index BETWEEN 101 AND 141) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0254 WHERE ( (index BETWEEN 1 AND 2) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0062 WHERE ( (index BETWEEN 1 AND 414) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0257 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0261 WHERE ( (index BETWEEN 1 AND 46) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr2646 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0232 WHERE ( (index BETWEEN 1 AND 1089) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0086 WHERE ( (index BETWEEN 108203 AND 108210) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt0917 WHERE ( (index BETWEEN 1 AND 8069) OR (index BETWEEN 8070 AND 8092) ) AND ( (index NOT BETWEEN 1431 AND 2193) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0213 WHERE ( (index BETWEEN 1 AND 100) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0676 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0231 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr1273 WHERE ( (index BETWEEN 1 AND 2) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0234 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0533 WHERE ( (index BETWEEN 657 AND 1448) OR (index BETWEEN 4101 AND 4243) OR (index BETWEEN 1449 AND 2942) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0308 WHERE ( (index BETWEEN 138 AND 175) OR (index BETWEEN 471 AND 510) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr2022 WHERE ( (index BETWEEN 55547 AND 57240) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr1304 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0032 WHERE ( (index BETWEEN 11313 AND 11843) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr1390 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr2694 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr1725 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr1675 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0251 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0217 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr1205 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr2647 WHERE  ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM gr0085 WHERE ( (index BETWEEN 14284 AND 14284) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] { SELECT wkuniversalid, index, level_05_value, level_04_value, level_03_value, level_02_value, level_01_value, level_00_value, marked_up_line, accented_line, stripped_line, hyphenated_words, annotations FROM lt0959 WHERE ( (index BETWEEN 26819 AND 30408) OR (index BETWEEN 30409 AND 33654) OR (index BETWEEN 1 AND 2514) OR (index BETWEEN 33655 AND 34297) OR (index BETWEEN 2515 AND 6484) ) AND ( stripped_line ~* $1 )  ORDER BY index ASC LIMIT 200 dolor} [debugging]
//[debugging] 324 hits have been stored at 55d3c68d_results [debugging]
