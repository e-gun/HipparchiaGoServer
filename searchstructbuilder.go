//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"strings"
	"time"
)

//
// INITIAL SETUP
//

// InitializeSearch - set up a search; this is DRY code needed by both plain searches and vector searches
func InitializeSearch(c echo.Context, user string) SearchStruct {
	const (
		VECTORSEARCHSUMMARY = "Acquiring a model for the selected texts"
	)

	srch := BuildDefaultSearch(c)
	srch.User = user

	srch.Seeking = c.QueryParam("skg")
	srch.Proximate = c.QueryParam("prx")
	srch.LemmaOne = c.QueryParam("lem")
	srch.LemmaTwo = c.QueryParam("plm")
	srch.IPAddr = c.RealIP()

	srch.CleanInput()
	srch.SetType()  // must happen before SSBuildQueries()
	srch.Optimize() // maybe rewrite the search to make it faster
	srch.FormatInitialSummary()

	if srch.Type == "vector" {
		srch.InitSum = VECTORSEARCHSUMMARY
	}

	// now safe to rewrite skg oj that "^|\s", etc. can be added
	srch.Seeking = whitespacer(srch.Seeking, &srch)
	srch.Proximate = whitespacer(srch.Proximate, &srch)

	se := AllSessions.GetSess(user)
	srch.StoredSession = se
	sl := SessionIntoSearchlist(se)

	srch.SearchIn = sl.Inc
	srch.SearchEx = sl.Excl
	srch.SearchSize = sl.Size

	if srch.Twobox {
		srch.CurrentLimit = FIRSTSEARCHLIM
	}

	SSBuildQueries(&srch)

	srch.TableSize = len(srch.Queries)
	srch.IsActive = true

	SIUpdateRemain <- SIKVi{srch.WSID, srch.TableSize}
	return srch
}

// BuildDefaultSearch - fill out the basic values for a new search
func BuildDefaultSearch(c echo.Context) SearchStruct {
	user := readUUIDCookie(c)
	sess := AllSessions.GetSess(user)

	var s SearchStruct
	s.User = user
	s.Launched = time.Now()
	s.CurrentLimit = sess.HitLimit
	s.OriginalLimit = sess.HitLimit
	s.SrchColumn = DEFAULTCOLUMN
	s.SrchSyntax = DEFAULTQUERYSYNTAX
	s.OrderBy = ORDERBY
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

	if sess.NearOrNot == "notnear" {
		s.NotNear = true
	}

	if sess.VecNNSearch {
		s.Type = "vector"
	}

	s.ID = c.Param("id")
	s.WSID = s.ID

	// msg("nonstandard BuildDefaultSearch() for testing", MSGCRIT)

	return s
}

// BuildHollowSearch - is really a way to grab line collections via synthetic searchlists
func BuildHollowSearch() SearchStruct {
	s := SearchStruct{
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
		SrchColumn:    DEFAULTCOLUMN,
		SrchSyntax:    DEFAULTQUERYSYNTAX,
		OrderBy:       ORDERBY,
		VecTextPrep:   VECTORTEXTPREPDEFAULT,
		VecModeler:    VECTORMODELDEFAULT,
		CurrentLimit:  FIRSTSEARCHLIM,
		OriginalLimit: FIRSTSEARCHLIM,
		SkgSlice:      nil,
		PrxSlice:      nil,
		SearchIn:      SearchIncExl{},
		SearchEx:      SearchIncExl{},
		Queries:       nil,
		Results:       WorkLineBundle{},
		Launched:      time.Now(),
		TTName:        strings.Replace(uuid.New().String(), "-", "", -1),
		SearchSize:    0,
		TableSize:     0,
	}

	s.WSID = s.ID
	s.StoredSession = MakeDefaultSession(s.ID)
	return s
}

// CloneSearch - make a copy of a search with results and queries, inter alia, ripped out
func CloneSearch(f *SearchStruct, iteration int) SearchStruct {
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

	clone := SearchStruct{
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
		IsActive:      f.IsActive,
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
		SearchIn:      SearchIncExl{},
		SearchEx:      SearchIncExl{},
		Queries:       []PrerolledQuery{},
		Results:       WorkLineBundle{},
		Launched:      f.Launched,
		TTName:        strings.Replace(uuid.New().String(), "-", "", -1),
		SearchSize:    f.SearchSize,
		TableSize:     f.TableSize,
		ExtraMsg:      f.ExtraMsg,
		StoredSession: f.StoredSession,
	}

	SIUpdateIteration <- SIKVi{clone.WSID, clone.PhaseNum}
	return clone
}
