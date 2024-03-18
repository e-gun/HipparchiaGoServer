//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package search

import (
	"context"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/mps"
	"github.com/e-gun/HipparchiaGoServer/internal/str"
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"strings"
	"time"
)

//
// INITIAL SETUP
//

func GenerateSrchInfo(srch *str.SearchStruct) vlt.WSSrchInfo {
	return vlt.WSSrchInfo{
		ID:        srch.WSID,
		User:      srch.User,
		Exists:    true,
		Hits:      0,
		Remain:    srch.TableSize,
		TableCt:   srch.TableSize,
		SrchCount: 1,
		VProgStrg: "",
		Summary:   srch.InitSum,
		Iteration: 0,
		SType:     srch.Type,
		Launched:  srch.Launched,
		CancelFnc: srch.CancelFnc,
	}
}

// BuildDefaultSearch - fill out the basic values for a new search
func BuildDefaultSearch(c echo.Context) str.SearchStruct {
	const (
		VECTORSEARCHSUMMARY = "Acquiring a model for the selected texts"
	)

	user := vlt.ReadUUIDCookie(c)
	sess := vlt.AllSessions.GetSess(user)

	// mm("nonstandard BuildDefaultSearch() for testing", MSGCRIT)

	var s str.SearchStruct
	s.User = user
	s.Launched = time.Now()
	s.CurrentLimit = sess.HitLimit
	s.OriginalLimit = sess.HitLimit
	s.SrchColumn = vv.DEFAULTCOLUMN
	s.SrchSyntax = vv.DEFAULTQUERYSYNTAX
	s.OrderBy = vv.ORDERBY
	s.SearchIn = sess.Inclusions
	s.SearchEx = sess.Exclusions
	s.ProxDist = sess.Proximity
	s.ProxScope = sess.SearchScope
	s.NotNear = false
	s.Twobox = false
	s.HasPhraseBoxA = false
	s.HasLemmaBoxA = false
	s.SkgRewritten = false
	s.OneHit = sess.OneHit
	s.PhaseNum = 1
	s.VecTextPrep = sess.VecTextPrep
	s.VecModeler = sess.VecModeler
	s.TTName = strings.Replace(uuid.New().String(), "-", "", -1)
	s.StoredSession = sess
	s.RealIP = c.RealIP()

	if sess.NearOrNot == "notnear" {
		s.NotNear = true
	}

	if sess.VecNNSearch {
		s.Type = "vector"
	}

	s.ID = c.Param("id")
	s.WSID = s.ID

	InsertNewContextIntoSS(&s)

	s.User = user

	s.Seeking = c.QueryParam("skg")
	s.Proximate = c.QueryParam("prx")
	s.LemmaOne = c.QueryParam("lem")
	s.LemmaTwo = c.QueryParam("plm")
	s.IPAddr = c.RealIP()

	CleanInput(&s)
	s.SetType()         // must happen before SSBuildQueries()
	OptimizeSrearch(&s) // maybe rewrite the search to make it faster
	FormatInitialSummary(&s)

	if s.Type == "vector" {
		s.InitSum = VECTORSEARCHSUMMARY
	}

	// now safe to rewrite skg oj that "^|\s", etc. can be added
	s.Seeking = WhiteSpacer(s.Seeking, &s)
	s.Proximate = WhiteSpacer(s.Proximate, &s)

	se := vlt.AllSessions.GetSess(user)
	s.StoredSession = se
	sl := SessionIntoSearchlist(se)

	s.SearchIn = sl.Inc
	s.SearchEx = sl.Excl
	s.SearchSize = sl.Size

	if s.Twobox {
		s.CurrentLimit = vv.FIRSTSEARCHLIM
	}

	// rewrite these might be "bad" if you are doing a bulk search since search terms will be registered
	// SessionIntoBulkSearch() will blank this out and call SSBuildQueries() all over again
	SSBuildQueries(&s)

	s.TableSize = len(s.Queries)
	s.IsActive = true

	vlt.WSInfo.InsertInfo <- GenerateSrchInfo(&s)
	return s
}

// BuildHollowSearch - is really a way to grab line collections via synthetic searchlists
func BuildHollowSearch() str.SearchStruct {
	s := str.SearchStruct{
		User:          "",
		ID:            strings.Replace(uuid.New().String(), "-", "", -1),
		Seeking:       "",
		Proximate:     "",
		LemmaOne:      "",
		LemmaTwo:      "",
		InitSum:       "",
		Summary:       "",
		ProxScope:     "",
		ProxType:      "",
		ProxDist:      0,
		HasLemmaBoxA:  false,
		HasPhraseBoxA: false,
		IsActive:      false,
		OneHit:        false,
		Twobox:        false,
		NotNear:       false,
		SkgRewritten:  false,
		Type:          "",
		PhaseNum:      0,
		SrchColumn:    vv.DEFAULTCOLUMN,
		SrchSyntax:    vv.DEFAULTQUERYSYNTAX,
		OrderBy:       vv.ORDERBY,
		VecTextPrep:   vv.VECTORTEXTPREPDEFAULT,
		VecModeler:    vv.VECTORMODELDEFAULT,
		CurrentLimit:  vv.FIRSTSEARCHLIM,
		OriginalLimit: vv.FIRSTSEARCHLIM,
		SkgSlice:      nil,
		PrxSlice:      nil,
		SearchIn:      str.SearchIncExl{},
		SearchEx:      str.SearchIncExl{},
		Queries:       nil,
		Results:       str.WorkLineBundle{},
		Launched:      time.Now(),
		TTName:        strings.Replace(uuid.New().String(), "-", "", -1),
		SearchSize:    0,
		TableSize:     0,
	}

	InsertNewContextIntoSS(&s)

	s.WSID = s.ID
	s.StoredSession = lnch.MakeDefaultSession(s.ID)
	return s
}

// CloneSearch - make a copy of a search with results and queries, inter alia, ripped out
func CloneSearch(f *str.SearchStruct, iteration int) str.SearchStruct {
	// note that the clone is not accessible to RtWebsocket() because it never gets registered in the global SearchMap
	// this means no progress for second pass SearchMap; this can be achieved, but it is not currently a priority

	oid := strings.Replace(f.ID, "_pt2", "", -1) // so a pt3 does not look like "_pt2_pt3"
	id := fmt.Sprintf("%s_pt%d", oid, iteration)

	// THE DIVERGENCES
	//s.Results = WorkLineBundle{}
	//s.Queries = []PrerolledQuery{}
	//s.SearchIn = SearchIncExl{}
	//s.SearchEx = SearchIncExl{}
	//s.TTName = strings.Replace(uuid.New().String(), "-", "", -1)
	//s.SkgSlice = []string{}
	//s.PrxSlice = []string{}
	//s.PhaseNum = iteration
	//s.ID = id
	//s.Context
	//s.CancelFnc

	clone := str.SearchStruct{
		User:          f.User,
		IPAddr:        f.IPAddr,
		ID:            id,
		WSID:          f.ID,
		Seeking:       f.Seeking,
		Proximate:     f.Proximate,
		LemmaOne:      f.LemmaOne,
		LemmaTwo:      f.LemmaTwo,
		InitSum:       f.InitSum,
		Summary:       f.Summary,
		ProxScope:     f.ProxScope,
		ProxType:      f.ProxType,
		ProxDist:      f.ProxDist,
		HasLemmaBoxA:  f.HasLemmaBoxA,
		HasLemmaBoxB:  f.HasLemmaBoxB,
		HasPhraseBoxA: f.HasPhraseBoxA,
		HasPhraseBoxB: f.HasLemmaBoxA,
		IsLemmAndPhr:  f.IsLemmAndPhr,
		OneHit:        f.OneHit,
		Twobox:        f.Twobox,
		NotNear:       f.NotNear,
		SkgRewritten:  f.SkgRewritten,
		Type:          f.Type,
		PhaseNum:      iteration,
		SrchColumn:    f.SrchColumn,
		SrchSyntax:    f.SrchSyntax,
		OrderBy:       f.OrderBy,
		VecTextPrep:   f.VecTextPrep,
		VecModeler:    f.VecModeler,
		CurrentLimit:  f.CurrentLimit,
		OriginalLimit: f.OriginalLimit,
		SkgSlice:      []string{},
		PrxSlice:      []string{},
		SearchIn:      str.SearchIncExl{},
		SearchEx:      str.SearchIncExl{},
		Queries:       []str.PrerolledQuery{},
		Results:       str.WorkLineBundle{},
		Launched:      f.Launched,
		TTName:        strings.Replace(uuid.New().String(), "-", "", -1),
		SearchSize:    f.SearchSize,
		TableSize:     f.TableSize,
		ExtraMsg:      f.ExtraMsg,
		StoredSession: f.StoredSession,
		IsActive:      f.IsActive,
	}

	InsertNewContextIntoSS(&clone)

	vlt.WSInfo.UpdateIteration <- vlt.WSSIKVi{clone.WSID, clone.PhaseNum}

	return clone
}

func InsertNewContextIntoSS(ss *str.SearchStruct) {
	ss.Context, ss.CancelFnc = context.WithCancel(context.Background())
}

// SessionIntoBulkSearch - grab every line of text in the currently selected set of authors, works, and passages
func SessionIntoBulkSearch(c echo.Context, lim int) str.SearchStruct {
	user := vlt.ReadUUIDCookie(c)
	sess := vlt.AllSessions.GetSess(user)

	ss := BuildDefaultSearch(c)
	ss.Seeking = ""
	ss.Proximate = ""
	ss.LemmaOne = ""
	ss.LemmaTwo = ""
	ss.SkgSlice = []string{}
	ss.CurrentLimit = lim
	ss.InitSum = "Gathering and formatting the text..."
	ss.ID = strings.Replace(uuid.New().String(), "-", "", -1)

	// BuildDefaultSearch() set some things that need resetting
	ss.SetType()

	sl := SessionIntoSearchlist(sess)
	ss.SearchIn = sl.Inc
	ss.SearchEx = sl.Excl
	ss.SearchSize = sl.Size

	SSBuildQueries(&ss)

	ss.TableSize = len(ss.Queries)
	ss.IsActive = true

	SearchAndInsertResults(&ss)
	return ss
}

// OptimizeSrearch - think about rewriting the search to make it faster
func OptimizeSrearch(s *str.SearchStruct) {
	// only zero or one of the following should be true

	// if BoxA has a lemma and BoxB has a phrase, it is almost certainly faster to search B, then A...
	if s.HasLemmaBoxA && s.HasPhraseBoxB {
		s.SwapPhraseAndLemma()
		return
	}

	// all forms of an uncommon word should (usually) be sought before all forms of a common word...
	if s.HasLemmaBoxA && s.HasLemmaBoxB {
		PickFastestLemma(s)
		return
	}

	// a single word should be faster than a lemma; but do not swap an empty string
	if s.HasLemmaBoxA && !s.HasPhraseBoxB && s.Proximate != "" {
		s.SwapWordAndLemma()
		return
	}

	// consider looking for the string with more characters in it first
	if len(s.Seeking) > 0 && len(s.Proximate) > 0 {
		s.SearchQuickestFirst()
		return
	}
}

// PickFastestLemma - all forms of an uncommon word should (usually) be sought before all forms of a common word
func PickFastestLemma(s *str.SearchStruct) {
	// Sought all 65 forms of »δημηγορέω« within 1 lines of all 386 forms of »γιγνώϲκω«
	// swapped: 20s vs 80s

	// Sought all 68 forms of »διαμάχομαι« within 1 lines of all 644 forms of »ποιέω«
	// similar to previous: 20s vs forever...

	// Sought all 12 forms of »αὐτοκράτωρ« within 1 lines of all 50 forms of »πόλιϲ«
	// swapped: 4.17s vs 10.09s

	// it does not *always* save time to just pick the uncommon word:

	// Sought all 50 forms of »πόλιϲ« within 1 lines of all 191 forms of »ὁπλίζω«
	// this fnc will COST you 10s when you swap 33s instead of 23s.

	// the "191 forms" take longer to find than the "50 forms"; that is, the fast first pass of πόλιϲ is fast enough
	// to offset the cost of looking for ὁπλίζω among the 125274 initial hits (vs 2547 initial hits w/ ὁπλίζω run first)

	// note that it is *usually* the case that words with more forms also have more hits
	// the penalty for being wrong is relatively low; the savings when you get this right can be significant

	const (
		NOTE1 = "PickFastestLemma() is swapping %s for %s: possible hits %d < %d; known forms %d < %d"
		NOTE2 = "PickFastestLemma() is NOT swapping %s for %s: possible hits %d vs %d; known forms %d vs %d"
	)

	hw1 := HeadwordLookup(s.LemmaOne)
	hw2 := HeadwordLookup(s.LemmaTwo)

	// how many forms to look up?
	fc1 := len(mps.AllLemm[s.LemmaOne].Deriv)
	fc2 := len(mps.AllLemm[s.LemmaTwo].Deriv)

	// the "&&" tries to address the »πόλιϲ« vs »ὁπλίζω« problem: see the notes above
	if (hw1.Total > hw2.Total) && (fc1 > fc2) {
		s.LemmaTwo = hw1.Entry
		s.LemmaOne = hw2.Entry
		Msg.PEEK(fmt.Sprintf(NOTE1, hw2.Entry, hw1.Entry, hw2.Total, hw1.Total, fc2, fc1))
	} else {
		Msg.PEEK(fmt.Sprintf(NOTE2, hw1.Entry, hw2.Entry, hw1.Total, hw2.Total, fc1, fc2))
	}
}
