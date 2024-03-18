package lnch

import (
	"github.com/e-gun/HipparchiaGoServer/internal/m"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"runtime"
	"time"
)

func NewMessageMakerConfigured() *m.MessageMaker {
	w := false
	if runtime.GOOS == "windows" {
		w = true
	}
	return &m.MessageMaker{
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

func NewMessageMakerWithDefaults() *m.MessageMaker {
	w := false
	if runtime.GOOS == "windows" {
		w = true
	}
	return &m.MessageMaker{
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
