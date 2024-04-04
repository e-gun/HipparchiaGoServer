//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package str

import "fmt"

type WorkLineBundle struct {
	Lines []DbWorkline
	Abort chan struct{} // YieldSome() uses this; and only YieldSome() should initialize it
}

// YieldAll - don't copy everything at once; send everything over a channel
func (wlb *WorkLineBundle) YieldAll() chan DbWorkline {
	// assuming the receiver will grab everything; this is everywhere true with one exception
	// the code is always of the following format: `ll = wlb.YieldAll()` + `for l := range ll { ... }`

	Msg.TMI(fmt.Sprintf("WorkLineBundle.YieldAll() sending %d lines", wlb.Len()))

	c := make(chan DbWorkline)
	go func() {
		for i := 0; i < len(wlb.Lines); i++ {
			c <- wlb.Lines[i]
		}
		close(c)
	}()
	return c
}

// YieldSome - maybe send everything over a channel
func (wlb *WorkLineBundle) YieldSome() chan DbWorkline {
	// see notes at SearchAndInsertResults() and FindPhrasesAcrossLines() on why not YieldAll()
	wlb.Abort = make(chan struct{})

	const (
		MSG1 = "WorkLineBundle.YieldSome() has %d lines available"
		MSG2 = "WorkLineBundle.YieldSome() yielded all"
	)

	Msg.TMI(fmt.Sprintf(MSG1, wlb.Len()))

	c := make(chan DbWorkline)
	go func() {
		defer close(c)
		for i := 0; i < len(wlb.Lines); i++ {
			select {
			case <-wlb.Abort:
				return
			default:
				c <- wlb.Lines[i]
			}
		}
		Msg.TMI(MSG2)
	}()
	return c
}

func (wlb *WorkLineBundle) ResizeTo(i int) {
	if i < len(wlb.Lines) {
		wlb.Lines = wlb.Lines[0:i]
	}
}

func (wlb *WorkLineBundle) Len() int {
	return len(wlb.Lines)
}

func (wlb *WorkLineBundle) IsEmpty() bool {
	if len(wlb.Lines) == 0 {
		return true
	} else {
		return false
	}
}

func (wlb *WorkLineBundle) FirstLine() DbWorkline {
	if len(wlb.Lines) != 0 {
		return wlb.Lines[0]
	} else {
		return DbWorkline{}
	}
}

func (wlb *WorkLineBundle) AppendLines(toadd []DbWorkline) {
	wlb.Lines = append(wlb.Lines, toadd...)
}

func (wlb *WorkLineBundle) AppendOne(toadd DbWorkline) {
	wlb.Lines = append(wlb.Lines, toadd)
}
