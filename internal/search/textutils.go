package search

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/generic"
	"github.com/e-gun/HipparchiaGoServer/internal/launch"
	"github.com/e-gun/HipparchiaGoServer/internal/mps"
	"github.com/e-gun/HipparchiaGoServer/internal/structs"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"regexp"
	"slices"
	"sort"
	"strings"
)

func LemmaIntoRegexSlice(hdwd string) []string {
	// rather than do one word per query, bundle things up: some words have >100 forms
	// ...(^|\\s)ἐδηλώϲαντο(\\s|$)|(^|\\s)δεδηλωμένοϲ(\\s|$)|(^|\\s)δήλουϲ(\\s|$)|(^|\\s)δηλούϲαϲ(\\s|$)...
	const (
		FAILMSG = "lemmaintoregexslice() could not find '%s'"
		FAILSLC = "FIND_NOTHING"
	)

	var qq []string
	if _, ok := mps.AllLemm[hdwd]; !ok {
		msg.FYI(fmt.Sprintf(FAILMSG, hdwd))
		return []string{FAILSLC}
	}

	tp := `(^|\s)%s(\s|$)`

	// there is a problem: unless you do something, "(^|\s)ἁλιεύϲ(\s|$)" will be a search term but this will not find "ἁλιεὺϲ"
	var lemm []string
	for _, l := range mps.AllLemm[hdwd].Deriv {
		lemm = append(lemm, generic.FindAcuteOrGrave(l))
	}

	ct := 0
	for true {
		var bnd []string
		for i := 0; i < vv.MAXLEMMACHUNKSIZE; i++ {
			if ct > len(lemm)-1 {
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

// FindPhrasesAcrossLines - "one two$" + "^three four" makes a hit if you want "one two three four"
func FindPhrasesAcrossLines(ss *structs.SearchStruct) {
	// modify ss in place

	const (
		FAIL = "<code>SEARCH FAILED: term sent to FindPhrasesAcrossLines() yielded error inside regexp.Compile()</code><br><br>"
	)

	recordfailure := func() {
		ss.Results = structs.WorkLineBundle{}
		ss.ExtraMsg = FAIL
	}

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
		for i := range ww {
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

	var valid = make(map[string]structs.DbWorkline, ss.Results.Len())

	skg := ss.Seeking
	if ss.SkgRewritten {
		skg = RestoreWhiteSpace(ss.Seeking)
	}

	find := regexp.MustCompile(`^ `)
	re := find.ReplaceAllString(skg, "(^|\\s)")
	find = regexp.MustCompile(` $`)
	re = find.ReplaceAllString(re, "(\\s|$)")
	fp, e := regexp.Compile(re)
	if e != nil {
		// Καῖϲα[ρ can be requested, but it will cause big problems
		// this mechanism likely needs to be inserted in more locations...
		recordfailure()
		return
	}

	altfp, e := regexp.Compile(ss.Seeking)
	if e != nil {
		recordfailure()
		return
	}

	rr := ss.Results.YieldAll()
	i := 0
	for r := range rr {
		// do the "it's all on this line" case separately
		li := ColumnPicker(ss.SrchColumn, r)
		f := fp.MatchString(li)
		if f {
			valid[r.BuildHyperlink()] = r
		} else if ss.SkgRewritten && altfp.MatchString(li) {
			// i.e. "it's all on this line" (second try)
			valid[r.BuildHyperlink()] = r
		} else if ss.OneHit {
			// problem with onehit + phrase: unless you do something the next yields ZERO results if OneHit is turned on
			// Sought »αρετα παϲα«
			// Searched 7,461 works and found 2 passages (1.80s)

			// the first pass will find the wrong lines. OneHit means FinalResultCollation() will register only the first
			// of a yet untested pair; but the real hit is in [b], not [a]; this is a 'linebundle' windowing problem from PSQL

			// [1.a] διανοίαϲ καὶ ὀρέξιοϲ ὧν ἁ μὲν διάνοια τῶ λόγον ἔχοντόϲ
			// [1.b] ἐντι ἁ δ ὄρεξιϲ τῶ ἀλόγω διὸ καὶ ἀρετὰ πᾶϲα ἐν

			// [2.a] καὶ τῷ ἀλόγῳ ϲύγκειται γὰρ ἐκ διανοίαϲ καὶ ὀρέξιοϲ ὧν ἁ μὲν διάνοια
			// [2.b] τῶ λόγον ἔχοντόϲ ἐντι ἁ δ ὄρεξιϲ τῶ ἀλόγω διὸ καὶ ἀρετὰ πᾶϲα ἐν
			nxt := GrabOneLine(r.AuID(), r.TbIndex+1)
			if nxt.WkUID == r.WkUID {
				n := ColumnPicker(ss.SrchColumn, nxt)
				if fp.MatchString(n) {
					valid[nxt.BuildHyperlink()] = nxt
				}
			}
		} else {
			var nxt structs.DbWorkline
			if i+1 < ss.Results.Len() {
				nxt = ss.Results.Lines[i+1]
				if r.TbIndex+1 > mps.AllWorks[r.WkUID].LastLine {
					nxt = structs.DbWorkline{}
				} else if r.WkUID != nxt.WkUID || r.TbIndex+1 != nxt.TbIndex {
					// grab the actual next line (i.e. index = 101)
					nxt = GrabOneLine(r.AuID(), r.TbIndex+1)
				}
			} else {
				// grab the actual next line (i.e. index = 101)
				nxt = GrabOneLine(r.AuID(), r.TbIndex+1)
				if r.WkUID != nxt.WkUID {
					nxt = structs.DbWorkline{}
				}
			}

			// combinator dodges double-register of hits
			nl := ColumnPicker(ss.SrchColumn, nxt)
			comb := getcombinations(re)
			for _, c := range comb {
				fp2, e1 := regexp.Compile(c[0])
				if e1 != nil {
					recordfailure()
					return
				}
				sp, e2 := regexp.Compile(c[1])
				if e2 != nil {
					recordfailure()
					return
				}
				f = fp2.MatchString(li)
				s := sp.MatchString(nl)
				if f && s && r.WkUID == nxt.WkUID {
					// yes! actually record a valid hit...
					valid[r.BuildHyperlink()] = r
				}
			}
		}
		i++
	}

	slc := make([]structs.DbWorkline, len(valid))
	counter := 0
	for _, r := range valid {
		slc[counter] = r
		counter += 1
	}

	ss.Results.Lines = slc
}

// WhiteSpacer - massage search string to let regex accept start/end of a line as whitespace
func WhiteSpacer(skg string, ss *structs.SearchStruct) string {
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

// RestoreWhiteSpace - undo WhiteSpacer() modifications
func RestoreWhiteSpace(skg string) string {
	// will have a problem rewriting regex inside phrasecombinations() if you don't clear WhiteSpacer() products out
	// even though we are about to put exactly this back in again...
	skg = strings.Replace(skg, "(^|\\s)", " ", 1)
	skg = strings.Replace(skg, "(\\s|$)", " ", -1)
	return skg
}

// ColumnPicker - convert from db column name into struct name
func ColumnPicker(c string, r structs.DbWorkline) string {
	const (
		MSG = "second.SrchColumn was not set; defaulting to 'stripped_line'"
	)

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
		msg.NOTE(MSG)
	}
	return li
}

// SearchTermFinder - find the universal regex equivalent of the search term
func SearchTermFinder(term string) *regexp.Regexp {
	//	you need to convert:
	//		ποταμον
	//	into:
	//		([πΠ][οὀὁὂὃὄὅόὸΟὈὉὊὋὌὍ][τΤ][αἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάᾈᾉᾊᾋᾌᾍᾎᾏἈἉἊἋἌἍἎἏΑ][μΜ][οὀὁὂὃὄὅόὸΟὈὉὊὋὌὍ][νΝ])

	const (
		MSG = "SearchTermFinder() could not compile the following: %s"
	)

	stre := generic.UniversalPatternMaker(term)
	pattern, e := regexp.Compile(stre)
	if e != nil {
		msg.WARN(fmt.Sprintf(MSG, stre))
		pattern = regexp.MustCompile("FAILED_FIND_NOTHING")
	}
	return pattern
}

//
// ALL of the following used to be a method on the struct but that yielded import problems
//

// CleanInput - remove bad chars, etc. from the submitted data
func CleanInput(s *structs.SearchStruct) {
	// address uv issues; lunate issues; ...
	// no need to parse a lemma: this bounces if there is not a key match to a map
	dropping := vv.USELESSINPUT + launch.Config.BadChars
	s.ID = generic.Purgechars(dropping, s.ID)
	s.Seeking = strings.ToLower(s.Seeking)
	s.Proximate = strings.ToLower(s.Proximate)

	if structs.HasAccent.MatchString(s.Seeking) || structs.HasAccent.MatchString(s.Proximate) {
		// lemma search will select accented automatically
		s.SrchColumn = "accented_line"
	}

	rs := []rune(s.Seeking)
	if len(rs) > vv.MAXINPUTLEN {
		s.Seeking = string(rs[0:vv.MAXINPUTLEN])
	}

	rp := []rune(s.Proximate)
	if len(rp) > vv.MAXINPUTLEN {
		s.Proximate = string(rs[0:vv.MAXINPUTLEN])
	}

	s.Seeking = generic.UVσςϲ(s.Seeking)
	s.Proximate = generic.UVσςϲ(s.Proximate)

	s.Seeking = generic.Purgechars(dropping, s.Seeking)
	s.Proximate = generic.Purgechars(dropping, s.Proximate)

	// don't let BoxA be blank if BoxB is not
	BoxA := s.Seeking == "" && s.LemmaOne == ""
	NotBoxB := s.Proximate != "" || s.LemmaTwo != ""

	if BoxA && NotBoxB {
		if s.Proximate != "" {
			s.Seeking = s.Proximate
			s.Proximate = ""
		}
		if s.LemmaTwo != "" {
			s.LemmaOne = s.LemmaTwo
			s.LemmaTwo = ""
		}
	}
}

// FormatInitialSummary - build HTML for the search summary
func FormatInitialSummary(s *structs.SearchStruct) {
	// ex:
	// Sought <span class="sought">»ἡμέρα«</span> within 2 lines of all 79 forms of <span class="sought">»ἀγαθόϲ«</span>
	const (
		TPM = `Sought %s<span class="sought">»%s«</span>%s`
		WIN = `%s within %d %s of %s<span class="sought">»%s«</span>`
		ADF = "all %d forms of "
		INF = "Grabbing all relevant lines..."
	)

	yn := ""
	if s.NotNear {
		yn = " not "
	}

	af1 := ""
	sk := s.Seeking
	if len(s.LemmaOne) != 0 {
		sk = s.LemmaOne
		if _, ok := mps.AllLemm[sk]; ok {
			af1 = fmt.Sprintf(ADF, len(mps.AllLemm[sk].Deriv))
		}
	}

	two := ""
	if s.Twobox {
		sk2 := s.Proximate
		af2 := ""
		if len(s.LemmaTwo) != 0 {
			sk2 = s.LemmaTwo
			if _, ok := mps.AllLemm[sk]; ok {
				af2 = fmt.Sprintf(ADF, len(mps.AllLemm[sk].Deriv))
			}
		}
		two = fmt.Sprintf(WIN, yn, s.ProxDist, s.ProxScope, af2, sk2)
	}

	sum := INF
	if sk != "" {
		sum = fmt.Sprintf(TPM, af1, sk, two)
	}
	s.InitSum = sum
}

// InclusionOverview - yield a summary of the inclusions; NeighborsSearch will use this when calling buildblanknngraph()
func InclusionOverview(s *structs.SearchStruct, sessincl structs.SearchIncExl) string {
	// possible to get burned, but this cheat is "good enough"
	// hipparchiaDB=# SELECT COUNT(universalid) FROM authors WHERE universalid LIKE 'gr%';
	// gr: 1823
	// lt: 362
	// in: 463
	// ch: 291
	// dp: 516

	const (
		MAXITEMS = 4
		GRCT     = 1823
		LTCT     = 362
		INCT     = 463
		CHCT     = 291
		DPCT     = 516
		FULL     = "all %d of the %s tables"
	)

	in := s.SearchIn
	BuildAuByName(&in)
	BuildWkByName(&in)

	// the named passages are available to a SeverSession, not a SearchStruct
	namemap := sessincl.MappedPsgByName
	var nameslc []string
	for _, v := range namemap {
		nameslc = append(nameslc, v)
	}
	sort.Strings(nameslc)

	var ov []string
	ov = append(ov, in.AuGenres...)
	ov = append(ov, in.WkGenres...)
	ov = append(ov, in.ListedABN...)
	ov = append(ov, in.ListedWBN...)
	ov = append(ov, nameslc...)

	notall := func() string {
		sort.Strings(ov)

		var enum []string

		if len(ov) != 1 {
			for i, p := range ov {
				enum = append(enum, fmt.Sprintf("(%d) %s", i+1, p))
			}
		} else {
			enum = append(enum, fmt.Sprintf("%s", ov[0]))
		}

		if len(enum) > MAXITEMS {
			diff := len(enum) - MAXITEMS
			enum = enum[0:MAXITEMS]
			enum = append(enum, fmt.Sprintf("and %d others", diff))
		}

		o := strings.Join(enum, "; ")
		nomarkup := strings.NewReplacer("<i>", "", "</i>", "")
		return nomarkup.Replace(o)
	}

	tt := len(ov)
	if tt != len(in.Authors) {
		tt = -1
	}

	r := ""
	switch tt {
	case GRCT:
		r = fmt.Sprintf(FULL, GRCT, "Greek author")
	case LTCT:
		r = fmt.Sprintf(FULL, LTCT, "Latin author")
	case INCT:
		r = fmt.Sprintf(FULL, INCT, "classical inscriptions")
	case DPCT:
		r = fmt.Sprintf(FULL, DPCT, "documentary papyri")
	case CHCT:
		r = fmt.Sprintf(FULL, CHCT, "christian era inscriptions")
	default:
		r = notall()
	}

	return r
}

func BuildAuByName(i *structs.SearchIncExl) {
	bn := make(map[string]string, len(i.MappedAuthByName))
	for _, a := range i.Authors {
		bn[a] = mps.AllAuthors[a].Cleaname
	}
	i.MappedAuthByName = bn

	var nn []string
	for _, v := range bn {
		nn = append(nn, v)
	}

	slices.Sort(nn)
	i.ListedABN = nn
}

func BuildWkByName(i *structs.SearchIncExl) {
	const (
		TMPL = `%s, <i>%s</i>`
	)
	bn := make(map[string]string, len(i.MappedWkByName))
	for _, w := range i.Works {
		ws := mps.AllWorks[w]
		au := mps.MyAu(ws).Name
		ti := ws.Title
		bn[w] = fmt.Sprintf(TMPL, au, ti)
	}
	i.MappedWkByName = bn

	var nn []string
	for _, v := range bn {
		nn = append(nn, v)
	}

	slices.Sort(nn)
	i.ListedWBN = nn
}

//
// SORTING: https://pkg.go.dev/sort#example__sortMultiKeys
//

type WLLessFunc func(p1, p2 *structs.DbWorkline) bool

// WLMultiSorter implements the Sort interface, sorting the changes within.
type WLMultiSorter struct {
	changes []structs.DbWorkline
	less    []WLLessFunc
}

// Sort sorts the argument slice according to the less functions passed to WLOrderedBy.
func (ms *WLMultiSorter) Sort(changes []structs.DbWorkline) {
	ms.changes = changes
	sort.Sort(ms)
}

// WLOrderedBy returns a Sorter that sorts using the less functions, in order.
// Call its Sort method to sort the data.
func WLOrderedBy(less ...WLLessFunc) *WLMultiSorter {
	return &WLMultiSorter{
		less: less,
	}
}

// Len is part of sort.Interface.
func (ms *WLMultiSorter) Len() int {
	return len(ms.changes)
}

// Swap is part of sort.Interface.
func (ms *WLMultiSorter) Swap(i, j int) {
	ms.changes[i], ms.changes[j] = ms.changes[j], ms.changes[i]
}

func (ms *WLMultiSorter) Less(i, j int) bool {
	p, q := &ms.changes[i], &ms.changes[j]
	// Try all but the last comparison.
	var k int
	for k = 0; k < len(ms.less)-1; k++ {
		less := ms.less[k]
		switch {
		case less(p, q):
			// p < q, so we have a decision.
			return true
		case less(q, p):
			// p > q, so we have a decision.
			return false
		}
		// p == q; try the next comparison.
	}
	// All comparisons to here said "equal", so just return whatever
	// the final comparison reports.
	return ms.less[k](p, q)
}

// SortResults - sort the search results by the session's registerselection criterion
func SortResults(s *structs.SearchStruct) {
	// Closures that order the DbWorkline structure:
	// see setsandslices.go and https://pkg.go.dev/sort#example__sortMultiKeys

	const (
		NULL = `Unavailable`
	)

	nameIncreasing := func(one, two *structs.DbWorkline) bool {
		a1 := MyAu(one).Shortname
		a2 := MyAu(two).Shortname
		return a1 < a2
	}

	titleIncreasing := func(one, two *structs.DbWorkline) bool {
		return MyWk(one).Title < MyWk(two).Title
	}

	dateIncreasing := func(one, two *structs.DbWorkline) bool {
		d1 := MyWk(one).RecDate
		d2 := MyWk(two).RecDate
		if d1 != NULL && d2 != NULL {
			return MyWk(one).ConvDate < MyWk(two).ConvDate
		} else if d1 == NULL && d2 != NULL {
			return MyAu(one).ConvDate < MyAu(two).ConvDate
		} else if d1 != NULL && d2 == NULL {
			return MyAu(one).ConvDate < MyAu(two).ConvDate
		} else {
			return MyAu(one).ConvDate < MyAu(two).ConvDate
		}
	}

	increasingLines := func(one, two *structs.DbWorkline) bool {
		return one.TbIndex < two.TbIndex
	}

	increasingID := func(one, two *structs.DbWorkline) bool {
		return one.BuildHyperlink() < two.BuildHyperlink()
	}

	increasingWLOC := func(one, two *structs.DbWorkline) bool {
		return MyWk(one).Prov < MyWk(two).Prov
	}

	sortby := s.StoredSession.SortHitsBy

	switch {
	case sortby == "shortname":
		WLOrderedBy(nameIncreasing, titleIncreasing, increasingLines).Sort(s.Results.Lines)
	case sortby == "converted_date":
		WLOrderedBy(dateIncreasing, nameIncreasing, titleIncreasing, increasingLines).Sort(s.Results.Lines)
	case sortby == "universalid":
		WLOrderedBy(increasingID).Sort(s.Results.Lines)
	case sortby == "provenance":
		// as this is likely an inscription search, why not sort next by date?
		WLOrderedBy(increasingWLOC, dateIncreasing).Sort(s.Results.Lines)
	default:
		// author nameIncreasing
		WLOrderedBy(nameIncreasing, increasingLines).Sort(s.Results.Lines)
	}
}

// MyWk - get the DbWork for this line
func MyWk(dbw *structs.DbWorkline) *structs.DbWork {
	w, ok := mps.AllWorks[dbw.WkUID]
	if !ok {
		msg.WARN(fmt.Sprintf("MyAu() failed to find '%s'", dbw.AuID()))
		w = &structs.DbWork{}
	}
	return w
}

// MyAu - get the DbAuthor for this line
func MyAu(dbw *structs.DbWorkline) *structs.DbAuthor {
	a, ok := mps.AllAuthors[dbw.AuID()]
	if !ok {
		msg.WARN(fmt.Sprintf("DbWorkline.MyAu() failed to find '%s'", dbw.AuID()))
		a = &structs.DbAuthor{}
	}
	return a
}
