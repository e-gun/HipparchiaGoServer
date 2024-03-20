//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//

package debug

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/str"
	"slices"
)

// ShowInterimResults - print out the current results
func ShowInterimResults(s *str.SearchStruct) {
	const (
		NB  = "ShowInterimResults()"
		FMT = "[%d] %s\t%s\t%s"
	)

	Msg.WARN(NB)

	mp := make(map[string]str.DbWorkline, s.Results.Len())
	kk := make([]string, s.Results.Len())

	for i := 0; i < s.Results.Len(); i++ {
		r := s.Results.Lines[i]
		mp[r.BuildHyperlink()] = r
		kk[i] = r.BuildHyperlink()
	}

	slices.Sort(kk)

	for i, k := range kk {
		r := mp[k]
		v := fmt.Sprintf(FMT, i, r.BuildHyperlink(), s.Seeking, r.MarkedUp)
		Msg.NOTE(v)
	}
}
