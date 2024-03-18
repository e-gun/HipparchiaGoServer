//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"bytes"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/generic"
	"github.com/e-gun/HipparchiaGoServer/internal/launch"
	"github.com/e-gun/HipparchiaGoServer/internal/m"
	"github.com/e-gun/HipparchiaGoServer/internal/vaults"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/labstack/echo/v4"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"text/template"
	"time"
)

var (
	// have the option to return/generate some sort of fail message...
	emptyjsreturn = func(c echo.Context) error { return c.JSONPretty(http.StatusOK, "", vv.JSONINDENT) }
)

//
// ROUTING
//

// RtFrontpage - send the html for "/"
func RtFrontpage(c echo.Context) error {
	const (
		UPSTR    = "[%v] HGS uptime: %v [%s]"
		PADDING  = " ----------------- "
		STATTMPL = "%s: %d"
		SPACER   = "    "
		VECTORS  = `
        <span id="vectorsearchcheckbox">
            <span class="rarechars small">v⃗</span><input type="checkbox" id="isvectorsearch" value="yes">
        </span>
       <span id="ldasearches">
            <span class="rarechars small">τ⃗</span><input type="checkbox" id="isldasearch" value="yes">
        </span>`
	)
	// will set if missing
	user := vaults.ReadUUIDCookie(c)
	s := vaults.AllSessions.GetSess(user)

	ahtm := vv.AUTHHTML
	if !launch.Config.Authenticate {
		ahtm = ""
	}

	gc := launch.GitCommit
	if gc == "" {
		gc = "UNKNOWN"
	}
	ver := fmt.Sprintf("Version: %s [git: %s]", vv.VERSION+launch.VersSuppl, gc)

	env := fmt.Sprintf("%s: %s - %s (%d workers)", runtime.Version(), runtime.GOOS, runtime.GOARCH, launch.Config.WorkerCount)

	// t() will give the uptime
	var mem runtime.MemStats

	t := func(up time.Duration) string {
		runtime.ReadMemStats(&mem)
		heap := fmt.Sprintf("%dM", mem.HeapAlloc/1024/1024)
		tick := fmt.Sprintf(UPSTR, time.Now().Format(time.TimeOnly), up.Truncate(time.Minute), heap)
		return PADDING + tick + PADDING
	}

	// svd() will report what requests have been made
	svd := func() string {
		responder := m.PIReply{Request: true, Response: make(chan map[string]int)}
		m.PIRequest <- responder
		ctr := <-responder.Response

		exclude := []string{"main() post-initialization"}
		keys := generic.StringMapKeysIntoSlice(ctr)
		keys = generic.SetSubtraction(keys, exclude)
		sort.Strings(keys)

		var pairs []string
		for k := range keys {
			this := strings.TrimPrefix(keys[k], "Rt")
			this = strings.TrimSuffix(this, "()")
			pairs = append(pairs, fmt.Sprintf(SPACER+STATTMPL, this, ctr[keys[k]]))
		}
		return strings.Join(pairs, "\n")
	}

	vec := ""
	if !launch.Config.VectorsDisabled {
		vec = VECTORS
	}

	// sample ticker output

	//      ----------------- [13:29:41] HGS uptime: 1m0s -----------------
	//
	//    BrowseLine: 5
	//    LexFindByForm: 2
	//    Search: 4

	subs := map[string]interface{}{
		"version":       vv.VERSION + launch.VersSuppl,
		"longver":       ver,
		"authhtm":       ahtm,
		"vec":           vec,
		"env":           env,
		"ticker":        t(time.Since(vv.LaunchTime)) + "\n\n" + svd(),
		"user":          "Anonymous",
		"resultcontext": s.HitContext,
		"browsecontext": s.BrowseCtx,
		"proxval":       s.Proximity}

	f, e := efs.ReadFile("emb/frontpage.html")
	msg.EC(e)

	tmpl, e := template.New("fp").Parse(string(f))
	msg.EC(e)

	var b bytes.Buffer
	err := tmpl.Execute(&b, subs)
	msg.EC(err)

	return c.HTML(http.StatusOK, b.String())
}
