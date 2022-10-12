//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	hasAccent = regexp.MustCompile("[äëïöüâêîôûàèìòùáéíóúᾂᾒᾢᾃᾓᾣᾄᾔᾤᾅᾕᾥᾆᾖᾦᾇᾗᾧἂἒἲὂὒἢὢἃἓἳὃὓἣὣἄἔἴὄὔἤὤἅἕἵὅὕἥὥἆἶὖἦὦἇἷὗἧὧᾲῂῲᾴῄῴᾷῇῷᾀᾐᾠᾁᾑᾡῒῢΐΰῧἀἐἰὀὐἠὠῤἁἑἱὁὑἡὡῥὰὲὶὸὺὴὼάέίόύήώᾶῖῦῆῶϊϋ]")
)

type SearchStruct struct {
	User         string
	ID           string
	Seeking      string
	Proximate    string
	LemmaOne     string
	LemmaTwo     string
	InitSum      string
	Summary      string
	ProxScope    string // "lines" or "words"
	ProxType     string // "near" or "not near"
	ProxDist     int64
	HasLemma     bool
	HasPhrase    bool
	IsVector     bool
	IsActive     bool
	OneHit       bool
	Twobox       bool
	NotNear      bool
	SkgRewritten bool
	PhaseNum     int
	SrchColumn   string // usually "stripped_line", sometimes "accented_line"
	SrchSyntax   string // almost always "~"
	OrderBy      string // almost always "index" + ASC
	Limit        int64
	SkgSlice     []string // either just Seeking or a decomposed version of a Lemma's possibilities
	PrxSlice     []string
	SearchIn     SearchIncExl
	SearchEx     SearchIncExl
	Queries      []PrerolledQuery
	Results      []DbWorkline
	Launched     time.Time
	TTName       string
	SearchSize   int // # of works searched
	TableSize    int // # of tables searched
}

func (s *SearchStruct) CleanInput() {
	// remove bad chars
	// address uv issues; lunate issues; ...
	// no need to parse a lemma: this bounces if there is not a key match to a map
	dropping := USELESSINPUT + cfg.BadChars
	s.ID = purgechars(dropping, s.ID)
	s.Seeking = strings.ToLower(s.Seeking)
	s.Proximate = strings.ToLower(s.Proximate)

	if hasAccent.MatchString(s.Seeking) || hasAccent.MatchString(s.Proximate) {
		// lemma search will select accented automatically
		s.SrchColumn = "accented_line"
	}

	rs := []rune(s.Seeking)
	if len(rs) > MAXINPUTLEN {
		s.Seeking = string(rs[0:MAXINPUTLEN])
	}

	rp := []rune(s.Proximate)
	if len(rp) > MAXINPUTLEN {
		s.Proximate = string(rs[0:MAXINPUTLEN])
	}

	s.Seeking = uvσςϲ(s.Seeking)
	s.Proximate = uvσςϲ(s.Proximate)

	s.Seeking = purgechars(dropping, s.Seeking)
	s.Proximate = purgechars(dropping, s.Proximate)
}

func (s *SearchStruct) SetType() {
	// skip detailed proximate checks because second pass search just feeds all of that into the primary fields
	const (
		ACC = `ϲῥἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάἐἑἒἓἔἕὲέἰἱἲἳἴἵἶἷὶίῐῑῒΐῖῗὀὁὂὃὄὅόὸὐὑὒὓὔὕὖὗϋῠῡῢΰῦῧύὺᾐᾑᾒᾓᾔᾕᾖᾗῂῃῄῆῇἤἢἥἣὴήἠἡἦἧὠὡὢὣὤὥὦὧᾠᾡᾢᾣᾤᾥᾦᾧῲῳῴῶῷώὼ`
		REG = `a-zα-ω`
	)
	ps := s.Proximate != ""
	psl := s.LemmaTwo != ""

	if ps || psl {
		s.Twobox = true
	}

	comp := fmt.Sprintf(`[%s%s]\s[%s%s]`, REG, ACC, REG, ACC)
	twowords := regexp.MustCompile(comp)

	if twowords.MatchString(s.Seeking) {
		s.HasPhrase = true
	}

	if len(s.LemmaOne) != 0 {
		s.HasLemma = true
		// accented line has "volat" in latin; and "uolo" will not find it
		if isGreek.MatchString(s.LemmaOne) {
			s.SrchColumn = "accented_line"
		}
	}
	return
}

func (s *SearchStruct) FormatInitialSummary() {
	// ex:
	// Sought <span class="sought">»ἡμέρα«</span> within 2 lines of all 79 forms of <span class="sought">»ἀγαθόϲ«</span>
	const (
		TPM = `Sought %s<span class="sought">»%s«</span>%s`
		WIN = `%s within %d %s of %s<span class="sought">»%s«</span>`
	)

	yn := ""
	if s.NotNear {
		yn = " not "
	}

	af1 := ""
	sk := s.Seeking
	if len(s.LemmaOne) != 0 {
		af := "all %d forms of "
		sk = s.LemmaOne
		af1 = fmt.Sprintf(af, len(AllLemm[sk].Deriv))
	}

	two := ""
	if s.Twobox {
		sk2 := s.Proximate
		af2 := ""
		if len(s.LemmaTwo) != 0 {
			af3 := "all %d forms of "
			sk2 = s.LemmaTwo
			af2 = fmt.Sprintf(af3, len(AllLemm[sk2].Deriv))
		}
		two = fmt.Sprintf(WIN, yn, s.ProxDist, s.ProxScope, af2, sk2)
	}
	sum := fmt.Sprintf(TPM, af1, sk, two)
	s.InitSum = sum
}

func (s *SearchStruct) SortResults() {
	// Closures that order the DbWorkline structure:
	// see generichelpers.go and https://pkg.go.dev/sort#example__sortMultiKeys

	const (
		NULL = `Unavailable`
	)

	nameIncreasing := func(one, two *DbWorkline) bool {
		a1 := one.MyAu().Shortname
		a2 := two.MyAu().Shortname
		return a1 < a2
	}

	titleIncreasing := func(one, two *DbWorkline) bool {
		return one.MyWk().Title < two.MyWk().Title
	}

	dateIncreasing := func(one, two *DbWorkline) bool {
		d1 := one.MyWk().RecDate
		d2 := two.MyWk().RecDate
		if d1 != NULL && d2 != NULL {
			return one.MyWk().ConvDate < two.MyWk().ConvDate
		} else if d1 == NULL && d2 != NULL {
			return one.MyAu().ConvDate < two.MyAu().ConvDate
		} else if d1 != NULL && d2 == NULL {
			return one.MyAu().ConvDate < two.MyAu().ConvDate
		} else {
			return one.MyAu().ConvDate < two.MyAu().ConvDate
		}
	}

	//dateDecreasing := func(one, two *DbWorkline) bool {
	//	return AllWorks[one.WkID()].ConvDate > AllWorks[two.WkID()].ConvDate
	//}

	increasingLines := func(one, two *DbWorkline) bool {
		return one.TbIndex < two.TbIndex
	}

	//decreasingLines := func(one, two *DbWorkline) bool {
	//	return one.TbIndex > two.TbIndex // Note: > orders downwards.
	//}

	increasingID := func(one, two *DbWorkline) bool {
		return one.BuildHyperlink() < two.BuildHyperlink()
	}

	//increasingALOC := func(one, two *DbWorkline) bool {
	//	return AllAuthors[one.AuID()].Location < AllAuthors[two.AuID()].Location
	//}

	increasingWLOC := func(one, two *DbWorkline) bool {
		return one.MyWk().Prov < two.MyWk().Prov
	}

	crit := sessions[s.User].SortHitsBy

	switch {
	// unhandled are "location" & "provenance"
	case crit == "shortname":
		OrderedBy(nameIncreasing, titleIncreasing, increasingLines).Sort(s.Results)
	case crit == "converted_date":
		OrderedBy(dateIncreasing, nameIncreasing, titleIncreasing, increasingLines).Sort(s.Results)
	case crit == "universalid":
		OrderedBy(increasingID).Sort(s.Results)
	case crit == "provenance":
		// as this is likely an inscription search, why not sort next by date?
		OrderedBy(increasingWLOC, dateIncreasing).Sort(s.Results)
	default:
		// author nameIncreasing
		OrderedBy(nameIncreasing, increasingLines).Sort(s.Results)
	}
}

//
// ROUTING
//

// RtSearchConfirm - just tells the client JS where to find the poll
func RtSearchConfirm(c echo.Context) error {
	pt := fmt.Sprintf("%d", cfg.HostPort)
	return c.String(http.StatusOK, pt)
}

// RtSearch - find X (derived from boxes on page) in Y (derived from the session)
func RtSearch(c echo.Context) error {
	// "OneBox"
	// [1] single word
	// [2] phrase
	// [3] lemma
	// "TwoBox"
	// [4] single + single
	// [5] lemma + single
	// [6] lemma + lemma
	// [7] phrase + single
	// [8] phrase + lemma
	// [9] phrase + phrase

	c.Response().After(func() { gcstats("RtSearch()") })

	user := readUUIDCookie(c)
	id := c.Param("id")
	srch := builddefaultsearch(c)
	srch.User = user

	srch.Seeking = c.QueryParam("skg")
	srch.Proximate = c.QueryParam("prx")
	srch.LemmaOne = c.QueryParam("lem")
	srch.LemmaTwo = c.QueryParam("plm")
	srch.ID = c.Param("id")
	srch.IsVector = false
	// HasPhrase makes us use a fake limit temporarily
	reallimit := srch.Limit

	srch.CleanInput()
	srch.SetType() // must happen before SSBuildQueries()
	srch.FormatInitialSummary()

	// now safe to rewrite skg oj that "^|\s", etc. can be added
	srch.Seeking = whitespacer(srch.Seeking, &srch)
	srch.Proximate = whitespacer(srch.Proximate, &srch)

	sl := SessionIntoSearchlist(sessions[user])
	srch.SearchIn = sl.Inc
	srch.SearchEx = sl.Excl
	srch.SearchSize = sl.Size

	SSBuildQueries(&srch)

	srch.TableSize = len(srch.Queries)
	srch.IsActive = true
	searches[id] = srch

	var completed SearchStruct
	if searches[id].Twobox {
		// todo: triple-check results against python
		if searches[id].ProxScope == "words" {
			completed = WithinXWordsSearch(searches[id])
		} else {
			completed = WithinXLinesSearch(searches[id])
		}
	} else {
		completed = HGoSrch(searches[id])
	}

	if completed.HasPhrase {
		// you did HGoSrch() and need to check the windowed lines
		// WithinXLinesSearch() has already done the checking
		findphrasesacrosslines(&completed)
		if int64(len(completed.Results)) > reallimit {
			completed.Results = completed.Results[0:reallimit]
		}
	}
	completed.SortResults()

	soj := SearchOutputJSON{}
	if sessions[readUUIDCookie(c)].HitContext == 0 {
		soj = FormatNoContextResults(completed)
	} else {
		soj = FormatWithContextResults(completed)
	}

	delete(searches, id)
	progremain.Delete(id)

	return c.JSONPretty(http.StatusOK, soj, JSONINDENT)
}

//
// TWO-PART SEARCHES
//

// WithinXLinesSearch - find A within N lines of B
func WithinXLinesSearch(originalsrch SearchStruct) SearchStruct {
	// after finding A, look for B within N lines of A

	// (part 1)
	//		HGoSrch(first)
	//
	// (part 2)
	// 		populate a new search list with a ton of passages via the first results
	//		HGoSrch(second)

	previous := time.Now()
	first := generateinitialhits(originalsrch)

	d := fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf("%s WithinXLinesSearch(): %d initial hits", d, len(first.Results)), 4)
	previous = time.Now()

	second := clonesearch(first, 2)
	second.Seeking = second.Proximate
	second.LemmaOne = second.LemmaTwo
	second.Proximate = first.Seeking
	second.LemmaTwo = first.LemmaOne

	second.SetType()

	pt := `%s_FROM_%d_TO_%d`

	newpsg := make([]string, len(first.Results))
	for i, r := range first.Results {
		// avoid "gr0028_FROM_-1_TO_5"
		low := r.TbIndex - first.ProxDist
		if low < 1 {
			low = 1
		}
		np := fmt.Sprintf(pt, r.AuID(), low, r.TbIndex+first.ProxDist)
		newpsg[i] = np
	}

	second.Limit = originalsrch.Limit
	second.SearchIn.Passages = newpsg

	SSBuildQueries(&second)

	d = fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf("%s SSBuildQueries() rerun", d), 4)
	previous = time.Now()

	second = HGoSrch(second)

	// was this a "notnear" search?
	if second.NotNear {
		var actualhits []DbWorkline
		// any original hits that match lines from pt2 are the "real" hits
		mapper := make(map[string]bool)
		for _, r := range first.Results {
			mapper[r.Citation()] = true
		}
		for _, r := range second.Results {
			if _, ok := mapper[r.Citation()]; ok {
				actualhits = append(actualhits, r)
			}
		}
		second.Results = actualhits
	}

	d = fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf("%s WithinXLinesSearch(): %d subsequent hits", d, len(first.Results)), 4)

	// findphrasesacrosslines() check happens just after you exit this function
	return second
}

// WithinXWordsSearch - find A within N words of B
func WithinXWordsSearch(originalsrch SearchStruct) SearchStruct {
	previous := time.Now()
	first := generateinitialhits(originalsrch)

	d := fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf("%s WithinXWordsSearch(): %d initial hits", d, len(first.Results)), 4)
	previous = time.Now()

	// the trick is we are going to grab ALL lines near the initial hit; then build strings; then search those strings ourselves
	// so the second search is "anything nearby"

	// [a] build the second search
	second := clonesearch(first, 2)
	sskg := second.Proximate
	slem := second.LemmaTwo
	second.Seeking = ""
	second.LemmaOne = ""
	second.Proximate = first.Seeking
	second.LemmaTwo = first.LemmaOne
	// avoid "WHERE accented_line !~ ''" : force the type and make sure to check "first.NotNear" below
	second.NotNear = false

	second.SetType()

	// [a1] hard code a suspect assumption...
	need := 2 + (first.ProxDist / int64(AVGWORDSPERLINE))

	pt := `%s_FROM_%d_TO_%d`
	t := `index/%s/%s/%d`

	resultmapper := make(map[string]int, len(first.Results))
	newpsg := make([]string, len(first.Results))

	// [a2] pick the lines to grab and associate them with the hits they go with
	for i, r := range first.Results {
		low := r.TbIndex - need
		if low < 1 {
			low = 1
		}
		np := fmt.Sprintf(pt, r.AuID(), low, r.TbIndex+need)
		newpsg[i] = np
		for j := r.TbIndex - need; j <= r.TbIndex+need; j++ {
			m := fmt.Sprintf(t, r.AuID(), r.WkID(), j)
			resultmapper[m] = i
		}
	}

	second.SearchIn.Passages = newpsg
	SSBuildQueries(&second)

	// [b] run the second "search" for anything/everything: ""

	searches[originalsrch.ID] = HGoSrch(second)

	d = fmt.Sprintf("[Δ: %.3fs] ", time.Now().Sub(previous).Seconds())
	msg(fmt.Sprintf("%s WithinXWordsSearch(): %d subsequent hits", d, len(first.Results)), 4)
	previous = time.Now()

	// [c] convert these finds into strings and then search those strings
	// [c1] build bundles of lines
	bundlemapper := make(map[int][]DbWorkline)
	for _, r := range searches[originalsrch.ID].Results {
		url := r.BuildHyperlink()
		bun := resultmapper[url]
		bundlemapper[bun] = append(bundlemapper[bun], r)
	}

	// [c2] decompose them into long strings
	stringmapper := make(map[int]string)
	for idx, lines := range bundlemapper {
		var bundle []string
		for _, l := range lines {
			bundle = append(bundle, columnpicker(first.SrchColumn, l))
		}
		stringmapper[idx] = strings.Join(bundle, " ")
	}

	// [c3] grab the head and tail of each
	var re string
	if len(first.LemmaOne) != 0 {
		// this next makes things insanely slow
		// re = "(" + strings.Join(lemmaintoregexslice(first.LemmaOne), "|") + ")"

		// but the next will not find all "tribuo" near all "beneficium" in Hyginus
		// [169a.2.3] compressit. pro quo beneficium ei tribuit, iussitque eius fuscinam
		// re = strings.Join(lemmaintoregexslice(first.LemmaOne), "|")

		// this is going to catch and slice incomplete words? The cost is a miscalculation of distance, no?
		// this also will misfind  »uinco« within 3 words of all forms of »libero« at VP, Historia Romana 2.33.1.7
		// it will hit the 'victor' in 'liberarat uictoria'
		// re = "(" + strings.Join(AllLemm[first.LemmaOne].Deriv, "|") + ")"

		// the risk on this one is ^ or $. But is it possible for LemmaOne to be at the edge of a stringmapper string?
		re = "(" + strings.Join(AllLemm[first.LemmaOne].Deriv, " | ") + ")"

	} else {
		re = first.Seeking
	}

	rt := `^(?P<head>.*?)%s(?P<tail>.*?)$`

	patternone, e := regexp.Compile(fmt.Sprintf(rt, re))
	if e != nil {
		m := fmt.Sprintf("WithinXWordsSearch() could not compile second pass regex term 'patternone': %s", re)
		msg(m, 1)
		return badsearch(m)
	}

	if len(slem) != 0 {
		re = strings.Join(lemmaintoregexslice(slem), "|")
	} else {
		re = sskg
	}

	patterntwo, e := regexp.Compile(re)
	if e != nil {
		m := fmt.Sprintf("WithinXWordsSearch() could not compile second pass regex term 'patterntwo': %s", re)
		msg(m, 1)
		return badsearch(m)
	}

	// [c4] search head and tail for the second search term

	// the count is inclusive: the search term is one of the words
	// unless you do something "non solum" w/in 4 words of "sed etiam" is the non-obvious way to catch single-word sandwiches:
	// "non solum pecuniae sed etiam..."

	pd := first.ProxDist

	ph1 := int64(len(strings.Split(strings.TrimSpace(first.Seeking), " ")))
	ph2 := int64(len(strings.Split(strings.TrimSpace(first.Proximate), " ")))

	if ph1 > 1 {
		pd = pd + ph1
	}

	if ph2 > 1 {
		pd = pd + ph2
	}

	var validresults []DbWorkline
	for idx, str := range stringmapper {
		subs := patternone.FindStringSubmatch(str)
		head := ""
		tail := ""
		if len(subs) != 0 {
			head = subs[patternone.SubexpIndex("head")]
			tail = subs[patternone.SubexpIndex("tail")]
		}

		hh := strings.Split(head, " ")
		start := int64(0)
		if int64(len(hh))-pd-1 > 0 {
			start = int64(len(hh)) - pd - 1
		}
		hh = hh[start:]
		head = strings.Join(hh, " ")

		tt := strings.Split(tail, " ")
		if int64(len(tt)) >= pd+1 {
			tt = tt[0 : pd+1]
		}
		tail = strings.Join(tt, " ")

		//msg("WithinXWordsSearch(): head, tail, patterntwo",3)
		//fmt.Println(head)
		//fmt.Println(tail)
		//fmt.Println(patterntwo)

		if first.NotNear {
			// toss hits
			if !patterntwo.MatchString(head) && !patterntwo.MatchString(tail) {
				validresults = append(validresults, first.Results[idx])
			}
		} else {
			// collect hits
			if patterntwo.MatchString(head) || patterntwo.MatchString(tail) {
				validresults = append(validresults, first.Results[idx])
			}
		}
	}

	second.Results = validresults

	// the next will drop a hit in findphrasesacrosslines()
	// 	Sought » non tribuit« within 5 words of »mihi autem« in Vitr. Arch 2 will have its hit pruned
	//	this is because "non tribuit" is there but "mihi autem" is not.

	//second.Seeking = sskg
	//second.LemmaOne = slem

	// so do this instead...
	second.Seeking = first.Seeking
	second.LemmaOne = first.LemmaOne

	return second
}

// generateinitialhits - part one of a two-part search
func generateinitialhits(first SearchStruct) SearchStruct {
	first.Limit = FIRSTSEARCHLIM
	first = HGoSrch(first)

	if first.HasPhrase {
		findphrasesacrosslines(&first)
	}
	return first
}

//
// INITIAL SETUP
//

// builddefaultsearch - fill out the basic values for a new search
func builddefaultsearch(c echo.Context) SearchStruct {
	user := readUUIDCookie(c)

	var s SearchStruct
	s.User = user
	s.Launched = time.Now()
	s.Limit = sessions[user].HitLimit
	s.SrchColumn = DEFAULTCOLUMN
	s.SrchSyntax = DEFAULTSYNTAX
	s.OrderBy = ORDERBY
	s.SearchIn = sessions[user].Inclusions
	s.SearchEx = sessions[user].Exclusions
	s.ProxDist = int64(sessions[user].Proximity)
	s.ProxScope = sessions[user].SearchScope
	s.NotNear = false
	s.Twobox = false
	s.HasPhrase = false
	s.HasLemma = false
	s.SkgRewritten = false
	s.OneHit = sessions[user].OneHit
	s.PhaseNum = 1
	s.TTName = strings.Replace(uuid.New().String(), "-", "", -1)

	if sessions[user].NearOrNot == "notnear" {
		s.NotNear = true
	}

	// msg("nonstandard builddefaultsearch() for testing", 1)

	return s
}

// buildhollowsearch - is really a way to grab line collections via synthetic searchlists
func buildhollowsearch() SearchStruct {
	s := SearchStruct{
		User:         "",
		ID:           strings.Replace(uuid.New().String(), "-", "", -1),
		Seeking:      "",
		Proximate:    "",
		LemmaOne:     "",
		LemmaTwo:     "",
		InitSum:      "",
		Summary:      "",
		ProxScope:    "",
		ProxType:     "",
		ProxDist:     0,
		HasLemma:     false,
		HasPhrase:    false,
		IsVector:     false,
		IsActive:     false,
		OneHit:       false,
		Twobox:       false,
		NotNear:      false,
		SkgRewritten: false,
		PhaseNum:     0,
		SrchColumn:   DEFAULTCOLUMN,
		SrchSyntax:   DEFAULTSYNTAX,
		OrderBy:      ORDERBY,
		Limit:        FIRSTSEARCHLIM,
		SkgSlice:     nil,
		PrxSlice:     nil,
		SearchIn:     SearchIncExl{},
		SearchEx:     SearchIncExl{},
		Queries:      nil,
		Results:      nil,
		Launched:     time.Now(),
		TTName:       strings.Replace(uuid.New().String(), "-", "", -1),
		SearchSize:   0,
		TableSize:    0,
	}
	return s
}

// whitespacer - massage search string to let regex accept start/end of a line as whitespace
func whitespacer(skg string, ss *SearchStruct) string {
	// whitespace issue: " ἐν Ὀρέϲτῃ " cannot be found at the start of a line where it is "ἐν Ὀρέϲτῃ "
	// do not run this before formatinitialsummary()
	// also used by searchformatting.go

	if strings.Contains(skg, " ") {
		ss.SkgRewritten = true
		rs := []rune(skg)
		a := ""
		if rs[0] == ' ' {
			a = "(^|\\s)"
		}
		z := ""
		if rs[len(rs)-1] == ' ' {
			z = "(\\s|$)"
		}
		skg = strings.TrimSpace(skg)
		skg = a + skg + z
	}
	return skg
}

// restorewhitespace - undo whitespacer() modifications
func restorewhitespace(skg string) string {
	// will have a problem rewriting regex inside phrasecombinations() if you don't clear whitespacer() products out
	// even though we are about to put exactly this back in again...
	skg = strings.Replace(skg, "(^|\\s)", " ", 1)
	skg = strings.Replace(skg, "(\\s|$)", " ", -1)
	return skg
}

//
// HELPERS
//

// badsearch - something went wrong, return a blank SearchStruct
func badsearch(msg string) SearchStruct {
	var s SearchStruct
	var l DbWorkline
	l.MarkedUp = msg
	s.Results = append(s.Results, l)
	return s
}

func lemmaintoregexslice(hdwd string) []string {
	// rather than do one word per query, bundle things up: some words have >100 forms
	// ...(^|\\s)ἐδηλώϲαντο(\\s|$)|(^|\\s)δεδηλωμένοϲ(\\s|$)|(^|\\s)δήλουϲ(\\s|$)|(^|\\s)δηλούϲαϲ(\\s|$)...

	var qq []string
	if _, ok := AllLemm[hdwd]; !ok {
		msg(fmt.Sprintf("lemmaintoregexslice() could not find '%s'", hdwd), 1)
		return qq
	}

	tp := `(^|\s)%s(\s|$)`

	// there is a problem: unless you do something, "(^|\s)ἁλιεύϲ(\s|$)" will be a search term but this will not find "ἁλιεὺϲ"
	var lemm []string
	for _, l := range AllLemm[hdwd].Deriv {
		lemm = append(lemm, findacuteorgrave(l))
	}

	ct := 0
	for true {
		var bnd []string
		for i := 0; i < MAXLEMMACHUNKSIZE; i++ {
			if ct > len(lemm)-1 {
				//re := fmt.Sprintf(tp, lemm[ct])
				//bnd = append(bnd, re)
				//qq = append(qq, strings.Join(bnd, "|"))
				break
			}
			re := fmt.Sprintf(tp, lemm[ct])
			bnd = append(bnd, re)
			ct += 1
		}
		qq = append(qq, strings.Join(bnd, "|"))
		if ct >= len(lemm)-1 {
			break
		}
	}
	return qq
}

// findphrasesacrosslines - "one two$" + "^three four" makes a hit if you want "one two three four"
func findphrasesacrosslines(ss *SearchStruct) {
	// modify ss in place

	getcombinations := func(phr string) [][2]string {
		// 'one two three four five' -->
		// [('one', 'two three four five'), ('one two', 'three four five'), ('one two three', 'four five'), ('one two three four', 'five')]

		gt := func(n int, wds []string) []string {
			return wds[n:]
		}

		gh := func(n int, wds []string) []string {
			return wds[:n]
		}

		ww := strings.Split(phr, " ")
		var comb [][2]string
		for i, _ := range ww {
			h := strings.Join(gh(i, ww), " ")
			t := strings.Join(gt(i, ww), " ")
			h = h + "$"
			t = "^" + t
			comb = append(comb, [2]string{h, t})
		}

		var trimmed [][2]string
		for _, c := range comb {
			head := strings.TrimSpace(c[0]) != "" && strings.TrimSpace(c[0]) != "$"
			tail := strings.TrimSpace(c[1]) != "" && strings.TrimSpace(c[0]) != "^"
			if head && tail {
				trimmed = append(trimmed, c)
			}
		}

		return trimmed
	}

	var valid = make(map[string]DbWorkline, len(ss.Results))

	skg := ss.Seeking
	if ss.SkgRewritten {
		skg = restorewhitespace(ss.Seeking)
	}

	find := regexp.MustCompile(`^ `)
	re := find.ReplaceAllString(skg, "(^|\\s)")
	find = regexp.MustCompile(` $`)
	re = find.ReplaceAllString(re, "(\\s|$)")
	fp := regexp.MustCompile(re)
	altfp := regexp.MustCompile(ss.Seeking)

	for i, r := range ss.Results {
		// do the "it's all on this line" case separately
		li := columnpicker(ss.SrchColumn, r)
		f := fp.MatchString(li)
		if f {
			valid[r.BuildHyperlink()] = r
		} else if ss.SkgRewritten && altfp.MatchString(li) {
			// i.e. "it's all on this line" (second try)
			// msg("althit", 1)
			valid[r.BuildHyperlink()] = r
		} else {
			// msg("'else'", 4)
			var nxt DbWorkline
			if i+1 < len(ss.Results) {
				nxt = ss.Results[i+1]
				if r.TbIndex+1 > AllWorks[r.WkUID].LastLine {
					nxt = DbWorkline{}
				} else if r.WkUID != nxt.WkUID || r.TbIndex+1 != nxt.TbIndex {
					// grab the actual next line (i.e. index = 101)
					nxt = graboneline(r.AuID(), r.TbIndex+1)
				}
			} else {
				// grab the actual next line (i.e. index = 101)
				nxt = graboneline(r.AuID(), r.TbIndex+1)
				if r.WkUID != nxt.WkUID {
					nxt = DbWorkline{}
				}
			}

			// combinator dodges double-register of hits
			nl := columnpicker(ss.SrchColumn, nxt)
			comb := getcombinations(re)
			for _, c := range comb {
				fp = regexp.MustCompile(c[0])
				sp := regexp.MustCompile(c[1])
				f = fp.MatchString(li)
				s := sp.MatchString(nl)
				if f && s && r.WkUID == nxt.WkUID {
					valid[r.BuildHyperlink()] = r
				}
			}
		}
	}

	slc := make([]DbWorkline, len(valid))
	counter := 0
	for _, r := range valid {
		slc[counter] = r
		counter += 1
	}

	ss.Results = slc
}

// columnpicker - convert from db column name into struct name
func columnpicker(c string, r DbWorkline) string {
	var li string
	switch c {
	case "stripped_line":
		li = r.Stripped
	case "accented_line":
		li = r.Accented
	case "marked_up_line": // only a maniac tries to search via marked_up_line
		li = r.MarkedUp
	default:
		li = r.Stripped
		msg("second.SrchColumn was not set; defaulting to 'stripped_line'", 2)
	}
	return li
}

// clonesearch - make a copy of a search with results and queries, inter alia, ripped out
func clonesearch(f SearchStruct, iteration int) SearchStruct {
	// note that the clone is not accessible to RtWebsocket() because it never gets registered in the global searches
	// this means no progress for second pass searches; this can be achieved, but it is not currently a priority
	s := f
	s.Results = []DbWorkline{}
	s.Queries = []PrerolledQuery{}
	s.SearchIn = SearchIncExl{}
	s.SearchEx = SearchIncExl{}
	s.TTName = strings.Replace(uuid.New().String(), "-", "", -1)
	s.SkgSlice = []string{}
	s.PrxSlice = []string{}
	s.PhaseNum = iteration

	oid := strings.Replace(f.ID, "_pt2", "", -1) // so a pt3 does not look like "_pt2_pt3"
	id := fmt.Sprintf("%s_pt%d", oid, iteration)
	s.ID = id
	return s
}

// universalpatternmaker - feeder for searchtermfinder()
func universalpatternmaker(term string) string {
	// also used by searchformatting.go
	// converter := extendedrunefeeder()
	converter := erunef // see top of generichelpers.go
	st := []rune(term)
	var stre string
	for _, r := range st {
		if _, ok := converter[r]; ok {
			re := fmt.Sprintf("[%s]", string(converter[r]))
			stre += re
		} else {
			stre += string(r)
		}
	}
	stre = fmt.Sprintf("(%s)", stre)
	return stre
}

// searchtermfinder - find the universal regex equivalent of the search term
func searchtermfinder(term string) *regexp.Regexp {
	//	you need to convert:
	//		ποταμον
	//	into:
	//		([πΠ][οὀὁὂὃὄὅόὸΟὈὉὊὋὌὍ][τΤ][αἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάᾈᾉᾊᾋᾌᾍᾎᾏἈἉἊἋἌἍἎἏΑ][μΜ][οὀὁὂὃὄὅόὸΟὈὉὊὋὌὍ][νΝ])

	stre := universalpatternmaker(term)
	pattern, e := regexp.Compile(stre)
	if e != nil {
		msg(fmt.Sprintf("searchtermfinder() could not compile the following: %s", stre), 1)
		pattern = regexp.MustCompile("FAILED_FIND_NOTHING")
	}
	return pattern
}
