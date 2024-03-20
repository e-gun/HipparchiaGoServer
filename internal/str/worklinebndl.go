//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package str

import "fmt"

type WorkLineBundle struct {
	Lines []DbWorkline
}

// YieldAll - don't copy everything at once; send everything over a channel
func (wlb *WorkLineBundle) YieldAll() chan DbWorkline {
	// assuming the receiver will grab everything
	// the code is always of the following format: `ll = wlb.YieldAll()` + `for l := range ll { ... }`

	// a YieldSome() is not yet needed: yield some but listen on a stop channel, etc.

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
