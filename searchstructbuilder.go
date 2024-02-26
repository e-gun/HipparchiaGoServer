//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"strings"
	"time"
)

//
// INITIAL SETUP
//

func GenerateSrchInfo(srch *SearchStruct) WSSrchInfo {
	return WSSrchInfo{
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
func BuildDefaultSearch(c echo.Context) SearchStruct {
	const (
		VECTORSEARCHSUMMARY = "Acquiring a model for the selected texts"
	)

	user := readUUIDCookie(c)
	sess := AllSessions.GetSess(user)

	// msg("nonstandard BuildDefaultSearch() for testing", MSGCRIT)

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
	s.RealIP = c.RealIP()

	if sess.NearOrNot == "notnear" {
		s.NotNear = true
	}

	if sess.VecNNSearch {
		s.Type = "vector"
	}

	s.ID = c.Param("id")
	s.WSID = s.ID

	InsertContextIntoSS(&s)

	s.User = user

	s.Seeking = c.QueryParam("skg")
	s.Proximate = c.QueryParam("prx")
	s.LemmaOne = c.QueryParam("lem")
	s.LemmaTwo = c.QueryParam("plm")
	s.IPAddr = c.RealIP()

	s.CleanInput()
	s.SetType()  // must happen before SSBuildQueries()
	s.Optimize() // maybe rewrite the search to make it faster
	s.FormatInitialSummary()

	if s.Type == "vector" {
		s.InitSum = VECTORSEARCHSUMMARY
	}

	// now safe to rewrite skg oj that "^|\s", etc. can be added
	s.Seeking = whitespacer(s.Seeking, &s)
	s.Proximate = whitespacer(s.Proximate, &s)

	se := AllSessions.GetSess(user)
	s.StoredSession = se
	sl := SessionIntoSearchlist(se)

	s.SearchIn = sl.Inc
	s.SearchEx = sl.Excl
	s.SearchSize = sl.Size

	if s.Twobox {
		s.CurrentLimit = FIRSTSEARCHLIM
	}

	SSBuildQueries(&s)

	s.TableSize = len(s.Queries)
	s.IsActive = true

	// but some fields are not set up quite yet
	WSInfo.InsertInfo <- GenerateSrchInfo(&s)
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

	InsertContextIntoSS(&s)

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
	//s.Context
	//s.CancelFnc

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
		IsActive:      f.IsActive,
	}

	InsertContextIntoSS(&clone)

	WSInfo.UpdateIteration <- WSSIKVi{clone.WSID, clone.PhaseNum}

	return clone
}

func InsertContextIntoSS(ss *SearchStruct) {
	ss.Context, ss.CancelFnc = context.WithCancel(context.Background())
}
