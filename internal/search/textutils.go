package search

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/generic"
	"github.com/e-gun/HipparchiaGoServer/internal/structs"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"regexp"
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
	if _, ok := vv.AllLemm[hdwd]; !ok {
		msg.FYI(fmt.Sprintf(FAILMSG, hdwd))
		return []string{FAILSLC}
	}

	tp := `(^|\s)%s(\s|$)`

	// there is a problem: unless you do something, "(^|\s)ἁλιεύϲ(\s|$)" will be a search term but this will not find "ἁλιεὺϲ"
	var lemm []string
	for _, l := range vv.AllLemm[hdwd].Deriv {
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
				if r.TbIndex+1 > vv.AllWorks[r.WkUID].LastLine {
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
