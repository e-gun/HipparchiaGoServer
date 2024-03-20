//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package lnch

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"runtime"
)

//
// VERSION INFO BUILD TIME INJECTION
//

// these next variables should be injected at build time: 'go build -ldflags "-X main.GitCommit=$GIT_COMMIT"', etc
// values are loaded into this file at runtime by main.go

var GitCommit string
var VersSuppl string
var BuildDate string
var PGOInfo string

func PrintVersion(cc str.CurrentConfiguration) {
	// example:
	// [HGS] Hipparchia Golang Server (v1.2.16-pre) [git: 64974732] [default.pgo] [gl=3; el=0]
	const (
		SN = "[C1%sC0] "
		GC = " [C4git: C4%sC0]"
		LL = " [C6gl=%d; el=%dC0]"
		ME = "C5%sC0 (C2v%sC0)"
		PG = " [C3%sC0]"
	)
	sn := fmt.Sprintf(SN, vv.SHORTNAME)
	gc := ""
	if GitCommit != "" {
		gc = fmt.Sprintf(GC, GitCommit)
	}

	pg := ""
	if PGOInfo != "" {
		pg = fmt.Sprintf(PG, PGOInfo)
	} else {
		pg = fmt.Sprintf(PG, "no pgo")
	}

	ll := fmt.Sprintf(LL, cc.LogLevel, cc.EchoLog)
	versioninfo := fmt.Sprintf(ME, vv.MYNAME, vv.VERSION+VersSuppl)
	versioninfo = sn + versioninfo + gc + pg + ll
	versioninfo = Msg.ColStyle(versioninfo)
	fmt.Println(versioninfo)
}

func PrintBuildInfo(cc str.CurrentConfiguration) {
	// example:
	// 	Built:	2023-11-14@19:02:51		Golang:	go1.21.4
	//	System:	darwin-arm64			WKvCPU:	20/20
	const (
		BD = "\tS1Built:S0\tC3%sC0\t"
		GV = "\tS1Golang:S0\tC3%sC0\n"
		SY = "\tS1System:S0\tC3%s-%sC0\t"
		WC = "\t\tS1WKvCPU:S0\tC3%dC0/C3%dC0"
	)

	bi := ""
	if BuildDate != "" {
		bi = Msg.ColStyle(fmt.Sprintf(BD, BuildDate))
	}
	bi += Msg.ColStyle(fmt.Sprintf(GV, runtime.Version()))
	bi += Msg.ColStyle(fmt.Sprintf(SY, runtime.GOOS, runtime.GOARCH))
	bi += Msg.ColStyle(fmt.Sprintf(WC, cc.WorkerCount, runtime.NumCPU()))
	fmt.Println(bi)
}
