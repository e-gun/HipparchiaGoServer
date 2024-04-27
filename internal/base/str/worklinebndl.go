//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package str

type WorkLineBundle struct {
	Lines []DbWorkline
	Halt  chan struct{}
}

// Yield - don't copy everything at once; send everything over a channel
func (wlb *WorkLineBundle) Yield() chan DbWorkline {
	wlb.Halt = make(chan struct{}) // only FindPhrasesAcrossLines() uses this channel; everyone else asks for all
	c := make(chan DbWorkline)
	go func() {
		defer close(c)
		for i := 0; i < len(wlb.Lines); i++ {
			select {
			case <-wlb.Halt:
				return
			default:
				c <- wlb.Lines[i]
			}
		}
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
