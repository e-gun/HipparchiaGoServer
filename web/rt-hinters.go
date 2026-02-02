//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/mps"
	"github.com/e-gun/HipparchiaGoServer/internal/vlt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/labstack/echo/v5"
	"net/http"
	"regexp"
	"slices"
	"strings"
)

// RtAuthorHints - /hints/author/_?term=auf --> [{'value': 'Aufidius Bassus [lt0809]'}, {'value': 'Aufustius [lt0401]'}]
func RtAuthorHints(c *echo.Context) error {
	const (
		TEMPL = "%s [%s]"
	)

	skg := c.QueryParam("term")
	if len(skg) < 2 {
		return emptyjsreturn(c)
	}
	skg = strings.ToLower(skg)

	// is what we have a match?
	var auu [][2]string
	if isidhint(skg) {
		auu = auidhint(skg)
	} else {
		auu = aunamehint(skg)
	}

	// trim by active corpora
	user := vlt.ReadUUIDCookie(c)
	s := vlt.AllSessions.GetSess(user)

	var trimmed [][2]string
	for _, a := range auu {
		co := a[1][0:vv.ABBREVLEN]
		if s.ActiveCorp[co] {
			trimmed = append(trimmed, a)
		}
	}

	auf := make([]string, len(trimmed))
	for i := 0; i < len(trimmed); i++ {
		t := trimmed[i]
		auf[i] = fmt.Sprintf(TEMPL, t[0], t[1])
	}

	slices.Sort(auf)
	out := tojsstructslice(auf)

	return c.JSONPretty(http.StatusOK, out, vv.JSONINDENT)
}

// RtLemmaHints - /hints/lemmata/_?term=dol --> [{"value": "dolabella\u00b9"}, {"value": "dolabra"}, {"value": "dolamen"}, ... ]
func RtLemmaHints(c *echo.Context) error {
	const (
		MAXLEMMARETURN = 33 // hates "προ" and "προτ": so many come back that you will lag the DOM
		MAXMSG         = "... (and %d more) ..."
	)

	user := vlt.ReadUUIDCookie(c)
	s := vlt.AllSessions.GetSess(user)

	term := c.QueryParam("term")
	// can't slice a unicode string...
	skg := []rune(term)

	if len(skg) < 2 {
		return emptyjsreturn(c)
	}

	skg = gen.StripaccentsRUNE(skg)
	nl := string(skg[0:2])

	var matches []string

	if _, ok := mps.NestedLemm[nl]; ok {
		for _, l := range mps.NestedLemm[nl] {
			er := l.EntryRune()

			// do not overshoot "er"...: "slice bounds out of range"
			lim := len(skg)
			if lim > len(er) {
				lim = len(er)
			}

			potential := gen.StripaccentsRUNE(er[0:lim])
			if len(er) >= len(skg) && string(potential) == string(skg) {
				// need to filter ab-cedo¹ --> abcedo
				matches = append(matches, l.Entry)
			}
		}
	}

	matches = gen.PolytonicSort(matches)

	if len(matches) > MAXLEMMARETURN {
		diff := len(matches) - MAXLEMMARETURN
		matches = matches[0:MAXLEMMARETURN]
		matches = append(matches, fmt.Sprintf(MAXMSG, diff))
	}

	if s.ZapLunates {
		for i, m := range matches {
			matches[i] = gen.DeLunate(m)
		}
	}

	jss := tojsstructslice(matches)

	return c.JSONPretty(http.StatusOK, jss, vv.JSONINDENT)
}

func RtAuGenreHints(c *echo.Context) error {
	return basichinter(c, mps.AuGenres)
}

func RtWkGenreHints(c *echo.Context) error {
	return basichinter(c, mps.WkGenres)
}

func RtAuLocHints(c *echo.Context) error {
	return basichinter(c, mps.AuLocs)
}

func RtWkLocHints(c *echo.Context) error {
	return basichinter(c, mps.WkLocs)
}

// basichinter - which substrings of the request are members of the master map?
func basichinter(c *echo.Context, mastermap map[string]bool) error {
	skg := c.QueryParam("term")
	if len(skg) < 2 {
		return emptyjsreturn(c)
	}
	skg = strings.ToLower(skg)
	skg = strings.Title(skg)

	// is what we have a match?
	var ff []string
	for f := range mastermap {
		if strings.Contains(f, skg) {
			ff = append(ff, f)
		}
	}

	// not really needed in practice?
	ff = gen.PolytonicSort(ff)
	jsstr := tojsstructslice(ff)

	return c.JSONPretty(http.StatusOK, jsstr, vv.JSONINDENT)
}

// tojsstructslice - []string -> []jsstruct for web output
func tojsstructslice(ss []string) []jsstruct {
	jss := make([]jsstruct, len(ss))
	for i := 0; i < len(ss); i++ {
		jss[i] = jsstruct{ss[i]}
	}
	return jss
}

//
// HELPERS for RtAuthorHints
//

// isidhint - is the author hint in the format of a UID?
func isidhint(skg string) bool {
	if len(skg) < 3 {
		return false
	}
	// note that ids for papyrus, etc are not necc. going to have a digit in third place; many will be unfindable
	// this is a to-do of sorts; but we are aiming for 'good enough' right now; nobody is really using this as it
	// is basically a debugger's function
	preff := []string{vv.TLGABBREV, vv.LATABBREV, vv.DDPABREV, vv.INSABBREV, vv.CHRABBREV}

	pref := fmt.Sprintf("(%s)", strings.Join(preff, ")|("))
	r := fmt.Sprintf("(%s)[0-9]", pref) // --> ((gr)|(lt)|(dp)|(in)|(ch))[0-9]
	// nb: the next is Sprintf "%s[0-9]" and is bad; will yield 'true' below for 'pindar': (gr)|(lt)|(dp)|(in)|(ch)[0-9]

	re, e := regexp.Compile(r)
	Msg.EC(e)
	if re.MatchString(skg) {
		return true
	} else {
		return false
	}
}

// aunamehint - which authors names contain the letters you have sent? "Var..." (used to do a 'starts with' check instead)
func aunamehint(skg string) [][2]string {
	// 'var' yields:
	// L. Varius Rufus
	// M. Terentius Varro
	// P Aufenus Varus
	// ...

	var auu [][2]string
	for _, a := range mps.AllAuthors {
		var who string

		// [sosthenes], et al. can be found via "sos" or "[so"
		if strings.Contains(skg, "[") {
			who = a.Cleaname
		} else {
			who = strings.Replace(a.Cleaname, "[", "", 1)
		}

		if strings.Contains(strings.ToLower(who), strings.ToLower(skg)) {
			ai := [2]string{a.Cleaname, a.UID}
			auu = append(auu, ai)
		}
	}
	return auu
}

// auidhint - which authors ids start with the letters you have sent? "gr00..."
func auidhint(skg string) [][2]string {
	var auu [][2]string
	for _, a := range mps.AllAuthors {
		var who string
		var an string

		who = a.UID
		if len(who) >= len(skg) {
			an = strings.ToLower(who[0:len(skg)])
		}
		if an == skg {
			ai := [2]string{a.Cleaname, a.UID}
			auu = append(auu, ai)
		}
	}
	return auu
}
