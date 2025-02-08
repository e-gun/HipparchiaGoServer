//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"encoding/json"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/lnch"
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"time"
)

// RtSessionSetsCookie - turn the session into a cookie
func RtSessionSetsCookie(c echo.Context) error {
	const (
		FAIL = "RtSessionSetsCookie() could not marshal the session"
	)
	num := c.Param("num")
	user := vlt.ReadUUIDCookie(c)
	s := vlt.AllSessions.GetSess(user)

	v, e := json.Marshal(s)
	if e != nil {
		v = []byte{}
		Msg.WARN(FAIL)
	}
	swap := strings.NewReplacer(`"`, "%22", ",", "%2C", " ", "%20")
	vs := swap.Replace(string(v))

	// note that cookie.Path = "/" is essential; otherwise different cookies for different contexts: "/browse" vs "/"
	cookie := new(http.Cookie)
	cookie.Name = "session" + num
	cookie.Path = "/"
	cookie.Value = vs
	cookie.Expires = time.Now().Add(4800 * time.Hour)
	c.SetCookie(cookie)

	return c.JSONPretty(http.StatusOK, "", vv.JSONINDENT)
}

// RtSessionGetCookie - turn a stored cookie into a session
func RtSessionGetCookie(c echo.Context) error {
	const (
		FAIL1 = "RtSessionGetCookie failed to read cookie %s for %s"
		FAIL2 = "RtSessionGetCookie failed to unmarshal cookie %s for %s"
		FAIL3 = "RtSessionGetCookie aborted because user %s not logged in"
	)

	user := vlt.ReadUUIDCookie(c)
	num := c.Param("num")
	cookie, err := c.Cookie("session" + num)
	if err != nil {
		Msg.WARN(fmt.Sprintf(FAIL1, num, user))
		return c.String(http.StatusOK, "")
	}

	cookie.Value = cleancookie(cookie.Value)

	var s str.ServerSession
	// invalid character '%' looking for beginning of object key string:
	// {%22ID%22:%22723073ae-09a7-4b24-a5d6-7e20603d8c44%22%2C%22IsLoggedIn%22:true%2C...}
	swap := strings.NewReplacer("%22", `"`, "%2C", ",", "%20", " ")
	cv := swap.Replace(cookie.Value)

	cv = strings.Replace(cv, "%", "", -1)

	err = json.Unmarshal([]byte(cv), &s)
	if err != nil {
		// invalid character '%' looking for beginning of object key string
		Msg.WARN(fmt.Sprintf(FAIL2, num, user))
		fmt.Println(err)
		return c.String(http.StatusOK, "")
	}

	// this code has input trust issues...
	// specifically, if you read a cookie that says "IsLoggedIn%22:true", should you just believe that?
	currentuser := vlt.ReadUUIDCookie(c)
	if lnch.Config.Authenticate && !vlt.AllAuthorized.Check(currentuser) {
		Msg.WARN(fmt.Sprintf(FAIL3, user))
		return nil
	}

	cs := vlt.AllSessions.GetSess(currentuser)
	s.ID = cs.ID

	vlt.AllSessions.InsertSess(s)

	e := c.Redirect(http.StatusFound, "/")
	Msg.EC(e)
	return nil
}

// RtResetSession - delete and then reset the session
func RtResetSession(c echo.Context) error {

	user := vlt.ReadUUIDCookie(c)
	oldsess := vlt.AllSessions.GetSess(user)

	vlt.AllSessions.Delete(user)

	// cancel any searches in progress: you are about to do a .CancelFnc()
	vlt.WSInfo.Reset <- user

	// [a] two-part searches are not canceled yet; and the incomplete results will be handed to the next function
	// canceling the subsequent parts happens via SSBuildQueries()
	// if !vlt.AllSessions.IsInVault(s.User) no actual queries will be loaded into the ss so the search ends instantly

	// [b] a different mechanism is used to halt a nn vector search once it starts training and the wego code has taken over
	// but the supplied context can cancel a training loop, yield empty embeddings, and then skip storage

	// [c] lda uses a similar mechanism: context inserted into nlp.LatentDirichletAllocation in the nlp code

	// reset the user ID and session
	newid := vlt.WriteUUIDCookie(c)
	newsess := vlt.AllSessions.GetSess(newid)

	// let's retain font and color info...
	newsess.FontSel = oldsess.FontSel
	newsess.ZapLunates = oldsess.ZapLunates
	newsess.DarkMode = oldsess.DarkMode

	vlt.AllSessions.InsertSess(newsess)

	e := c.Redirect(http.StatusFound, "/")
	Msg.EC(e)
	return nil
}

// cleancookie - try to mitigate SQL injection threat from the insertion of forged cookies
func cleancookie(s string) string {
	// it is probably possible to set up an 'evil' local cookie w/ SQL in it and then feed it to the server
	// so there really should be some parsing done of this new ServerSession that goes beyond conformity to the struct

	// session02:"{%22ID%22:%22dbe4067b-a9e1-4040-88d0-02d259c9a1ef%22%2C%22Inclusions%22:{%22AuGenres%22:null%2C%22WkGenres%22:null%2C%22AuLocations%22:null%2C%22WkLocations%22:null%2C%22Authors%22:[%22lt0119%22%2C%22lt0134%22]%2C%22Works%22:[]%2C%22Passages%22:[]%2C%22MappedPsgByName%22:{}%2C%22MappedAuthByName%22:null%2C%22MappedWkByName%22:null%2C%22ListedPBN%22:null%2C%22ListedABN%22:null%2C%22ListedWBN%22:null}%2C%22Exclusions%22:{%22AuGenres%22:null%2C%22WkGenres%22:null%2C%22AuLocations%22:null%2C%22WkLocations%22:null%2C%22Authors%22:null%2C%22Works%22:null%2C%22Passages%22:null%2C%22MappedPsgByName%22:{}%2C%22MappedAuthByName%22:null%2C%22MappedWkByName%22:null%2C%22ListedPBN%22:null%2C%22ListedABN%22:null%2C%22ListedWBN%22:null}%2C%22ActiveCorp%22:{%22ch%22:false%2C%22dp%22:false%2C%22gr%22:false%2C%22in%22:false%2C%22lt%22:true}%2C%22varia%22:true%2C%22incerta%22:true%2C%22spuria%22:true%2C%22rawinputstyle%22:false%2C%22onehit%22:false%2C%22headwordindexing%22:false%2C%22indexbyfrequency%22:false%2C%22vocbycount%22:false%2C%22vocscansion%22:false%2C%22nearornot%22:%22near%22%2C%22searchscope%22:%22lines%22%2C%22sortorder%22:%22shortname%22%2C%22proximity%22:1%2C%22BrowseCtx%22:24%2C%22InputStyle%22:%22%22%2C%22HitLimit%22:250%2C%22HitContext%22:4%2C%22Earliest%22:%22-850%22%2C%22Latest%22:%221500%22%2C%22TmpInt%22:0%2C%22TmpStr%22:%22%22%2C%22LoginName%22:%22Anonymous%22%2C%22VecGraphExt%22:false%2C%22VecModeler%22:%22w2v%22%2C%22VecNeighbCt%22:16%2C%22VecNNSearch%22:false%2C%22VecTextPrep%22:%22winner%22%2C%22VecLDASearch%22:false%2C%22LDAgraph%22:false%2C%22LDAtopics%22:0%2C%22LDA2D%22:true}"

	// one way to handle this would be to send everything through RtSetOption + RtSelectionMake, BUT...
	// gen.Purgechars() is going to be key one way or the other; so maybe just valiantly try to clean the cookie string

	// default "BadChars": "\"'!@:,=_/",
	// skipping lnch.Config.BadChars because if you use lnch.Config.BadChars and it contains "|", etc. we have a problem
	// must have: |:-%{}[] and 'evil %' is dropped later

	// awkward bit: SearchIncExl has 'byName' fields, and in rare cases we will have interesting characters there
	// but these are *supposed* to be present via 'special' variants alone (SMALL QUESTION MARK instead of QM, etc.)
	// in any case, cleaning the names should not in fact change the underlying selections and only yield 'typos'

	dropping := vv.USELESSINPUT + "\"`'!@,=_/$#&;^*<>~"

	return gen.Purgechars(dropping, s)
}
