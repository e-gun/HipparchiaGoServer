//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package lnch

import (
	"github.com/e-gun/HipparchiaGoServer/internal/base/mm"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"runtime"
	"time"
)

func NewMessageMakerConfigured() *mm.MessageMaker {
	w := false
	if runtime.GOOS == "windows" {
		w = true
	}
	return &mm.MessageMaker{
		Lnc:  time.Now(),
		BW:   Config.BlackAndWhite,
		Clr:  "",
		GC:   Config.ManualGC,
		LLvl: Config.LogLevel,
		LNm:  vv.MYNAME,
		SNm:  vv.SHORTNAME,
		Tick: Config.TickerActive,
		Ver:  vv.VERSION,
		Win:  w,
	}
}

func NewMessageMakerWithDefaults() *mm.MessageMaker {
	w := false
	if runtime.GOOS == "windows" {
		w = true
	}
	return &mm.MessageMaker{
		Lnc:  time.Now(),
		BW:   false,
		Clr:  "",
		GC:   false,
		LLvl: 0,
		LNm:  vv.MYNAME,
		SNm:  vv.SHORTNAME,
		Tick: false,
		Ver:  vv.VERSION,
		Win:  w,
	}
}

func UpdateMessageMakerWithConfig(m *mm.MessageMaker) {
	m.BW = Config.BlackAndWhite
	m.GC = Config.ManualGC
	m.LLvl = Config.LogLevel
}
