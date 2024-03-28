//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vec

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"os"
	"slices"
)

// FingerprintNNVectorSearch - derive a unique md5 for any given mix of search items & vector settings
func FingerprintNNVectorSearch(srch str.SearchStruct) string {
	const (
		MSG1 = "RtNeighborsSearch() fingerprint: "
		FAIL = "FingerprintNNVectorSearch() failed to Marshal"
	)

	// vectorbot vs normal surfer requires passing the model type and textprep style (bot: configtype; surfer: sessiontype)

	// unless you sort, you do not get repeatable results with a md5sum of srch.SearchIn if you look at "all latin"

	var fp []string

	// [1] start with the searchlist + the stoplists + VecTextPrep (which are all collections of strings + one string)

	// includes
	fp = append(fp, srch.SearchIn.AuGenres...)
	fp = append(fp, srch.SearchIn.WkGenres...)
	fp = append(fp, srch.SearchIn.AuLocations...)
	fp = append(fp, srch.SearchIn.WkLocations...)
	fp = append(fp, srch.SearchIn.Authors...)
	fp = append(fp, srch.SearchIn.Works...)
	fp = append(fp, srch.SearchIn.Passages...)

	// excludes
	fp = append(fp, srch.SearchEx.AuGenres...)
	fp = append(fp, srch.SearchEx.WkGenres...)
	fp = append(fp, srch.SearchEx.AuLocations...)
	fp = append(fp, srch.SearchEx.WkLocations...)
	fp = append(fp, srch.SearchEx.Authors...)
	fp = append(fp, srch.SearchEx.Works...)
	fp = append(fp, srch.SearchEx.Passages...)

	// stops
	fp = append(fp, readstopconfig("greek")...)
	fp = append(fp, readstopconfig("latin")...)

	// one last item...
	fp = append(fp, srch.VecTextPrep)
	slices.Sort(fp)

	f1, e1 := json.Marshal(fp)

	// [2] now add in the vector settings (which have an underlying Options struct)

	var f2 []byte
	var e2 error

	switch srch.VecModeler {
	case "glove":
		ff, ee := json.Marshal(glovevectorconfig())
		f2 = ff
		e2 = ee
	case "lexvec":
		ff, ee := json.Marshal(lexvecvectorconfig())
		f2 = ff
		e2 = ee
	default: // w2v
		ff, ee := json.Marshal(w2vvectorconfig())
		f2 = ff
		e2 = ee
	}

	if e1 != nil || e2 != nil {
		Msg.MAND(FAIL)
		os.Exit(1)
	}

	// [3] merge the previous two into a single byte array

	f1 = append(f1, f2...)

	// [4] generate the md5 fingerprint from this

	m := fmt.Sprintf("%x", md5.Sum(f1))
	Msg.TMI(MSG1 + m)

	return m
}
