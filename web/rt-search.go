//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/generic"
	"github.com/e-gun/HipparchiaGoServer/internal/launch"
	"github.com/e-gun/HipparchiaGoServer/internal/search"
	"github.com/e-gun/HipparchiaGoServer/internal/structs"
	"github.com/e-gun/HipparchiaGoServer/internal/vaults"
	"github.com/e-gun/HipparchiaGoServer/internal/vect"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/labstack/echo/v4"
	"net/http"
)

//
// ROUTING
//

// RtSearchConfirm - just tells the client JS where to find the poll
func RtSearchConfirm(c echo.Context) error {
	pt := fmt.Sprintf("%d", launch.Config.HostPort)
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

	const (
		TOOMANYIP    = "<code>Cannot execute this search. Your ip address (%s) is already running the maximum number of simultaneous searches allowed: %d.</code>"
		TOOMANYTOTAL = "<code>Cannot execute this search. The server is already running the maximum number of simultaneous searches allowed: %d.</code>"
	)

	user := vaults.ReadUUIDCookie(c)

	// [A] ARE WE GOING TO DO THIS AT ALL?

	if !vaults.AllAuthorized.Check(user) {
		return generic.JSONresponse(c, structs.SearchOutputJSON{JS: vv.VALIDATIONBOX})
	}

	getsrchcount := func(ip string) int {
		responder := vaults.WSSICount{Key: ip, Response: make(chan int)}
		vaults.WSInfo.IPSrchCount <- responder
		return <-responder.Response
	}

	if getsrchcount(c.RealIP()) >= launch.Config.MaxSrchIP {
		m := fmt.Sprintf(TOOMANYIP, c.RealIP(), getsrchcount(c.RealIP()))
		return generic.JSONresponse(c, structs.SearchOutputJSON{Searchsummary: m})
	}

	if len(vaults.WebsocketPool.ClientMap) >= launch.Config.MaxSrchTot {
		m := fmt.Sprintf(TOOMANYTOTAL, len(vaults.WebsocketPool.ClientMap))
		return generic.JSONresponse(c, structs.SearchOutputJSON{Searchsummary: m})
	}

	// [B] OK, WE ARE DOING IT

	srch := search.BuildDefaultSearch(c)
	se := vaults.AllSessions.GetSess(user)

	// [C] BUT WHAT KIND OF SEARCH IS IT? MAYBE IT IS A VECTOR SEARCH...

	// note the racer says that there are *many* race candidates in the imported vector code...
	// "wego@v0.0.11/pkg/model/word2vec/optimizer.go:126"
	// "wego@v0.0.11/pkg/model/word2vec/model.go:75"
	// ...

	if se.VecNNSearch && !launch.Config.VectorsDisabled {
		// not a normal search: jump to "vectorqueryneighbors.go" where we grab all lines; build a model; query against the model; return html
		return vect.NeighborsSearch(c, srch)
	}

	if se.VecLDASearch && !launch.Config.VectorsDisabled {
		// not a normal search: jump to "vectorquerylda.go"
		return vect.LDASearch(c, srch)
	}

	// [D] OK, IT IS A SEARCH FOR A WORD OR PHRASE

	c.Response().After(func() { msg.LogPaths("RtSearch()") })

	// HasPhraseBoxA makes us use a fake limit temporarily
	reallimit := srch.CurrentLimit

	var completed structs.SearchStruct
	if srch.Twobox {
		if srch.ProxScope == "words" {
			completed = search.WithinXWordsSearch(srch)
		} else {
			completed = search.WithinXLinesSearch(srch)
		}
	} else {
		completed = srch
		search.SearchAndInsertResults(&completed)
		if completed.HasPhraseBoxA {
			search.FindPhrasesAcrossLines(&completed)
		}
	}

	if completed.Results.Len() > reallimit {
		completed.Results.ResizeTo(reallimit)
	}

	// [E] DONE: TIME TO FORMAT

	search.SortResults(&completed)
	soj := structs.SearchOutputJSON{}
	if se.HitContext == 0 {
		soj = search.FormatNoContextResults(&completed)
	} else {
		soj = search.FormatWithContextResults(&completed)
	}

	vaults.WSInfo.Del <- srch.WSID
	return generic.JSONresponse(c, soj)
}
