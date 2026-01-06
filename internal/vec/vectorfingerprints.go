//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vec

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"slices"
)

// FingerprintNNVectorSearch - derive a unique md5 for any given mix of search items & vector settings
func FingerprintNNVectorSearch(srch str.SearchStruct) string {
	const (
		MSG1 = "FingerprintNNVectorSearch() fingerprint: "
		FAIL = "FingerprintNNVectorSearch() failed to Marshal"
	)

	// vectorbot vs normal surfer requires passing the model type and textprep style (bot: configtype; surfer: sessiontype)

	// unless you sort, you do not get repeatable results with a md5sum of srch.SearchIn if you look at "all latin"

	// [1] start with the searchlist + the stoplists + VecTextPrep (which are all collections of strings + one string)

	in := srch.SearchIn
	ex := srch.SearchEx

	incat := slices.Concat(in.Authors, in.AuGenres, in.AuLocations, in.Works, in.WkLocations, in.WkGenres, in.Passages)
	excat := slices.Concat(ex.Authors, ex.AuGenres, ex.AuLocations, ex.Works, ex.WkLocations, ex.WkGenres, ex.Passages)

	fp := slices.Concat(incat, excat, readstopconfig("greek"), readstopconfig("latin"))

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
		// os.Exit(1)
		f2 = []byte{}
	}

	// [3] merge the previous two into a single byte array

	f1 = append(f1, f2...)

	// [4] generate the md5 fingerprint from this

	m := fmt.Sprintf("%x", md5.Sum(f1))
	Msg.TMI(MSG1 + m)

	return m
}
