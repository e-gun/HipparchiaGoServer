package lnch

import (
	"github.com/e-gun/HipparchiaGoServer/internal/mm"
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
