//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package format

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"testing"
)

var (
	wl1 = str.DbWorkline{
		WkUID:       "tlg0658w001",
		TbIndex:     962,
		Lvl5Value:   "-1",
		Lvl4Value:   "-1",
		Lvl3Value:   "1",
		Lvl2Value:   "26",
		Lvl1Value:   "2",
		Lvl0Value:   "5",
		MarkedUp:    "ἴϲθι, Θεάγενεϲ, οὐδ’ ἂν τὸ παρὸν τοῦτο ἀλλήλοιϲ διελεγό-",
		Accented:    "ἴϲθι θεάγενεϲ οὐδ ἂν τὸ παρὸν τοῦτο ἀλλήλοιϲ διελεγόμεθα",
		Stripped:    "ιϲθι θεαγενεϲ ουδ αν το παρον τουτο αλληλοιϲ διελεγομεθα",
		Hyphenated:  "διελεγόμεθα",
		Annotations: "",
	}
	wl2 = str.DbWorkline{
		WkUID:       "tlg0658w001",
		TbIndex:     962,
		Lvl5Value:   "-1",
		Lvl4Value:   "-1",
		Lvl3Value:   "1",
		Lvl2Value:   "26",
		Lvl1Value:   "2",
		Lvl0Value:   "5",
		MarkedUp:    "ρον ἐπιτροπεύειν ἔρωτα· πολλὰ μία ἡμέρα καὶ δύο πολλά-",
		Accented:    "ρον ἐπιτροπεύειν ἔρωτα πολλὰ μία ἡμέρα καὶ δύο πολλάκις",
		Stripped:    "",
		Hyphenated:  "πολλάκις",
		Annotations: "",
	}
)

// TestBuildBrowserTable - BuildBrowserTable(focus int, lines []str.DbWorkline, zaplunates bool, regularizewidth bool)
func TestBuildBrowserTable(t *testing.T) {
	// <!-- tlg0658 962 --><!-- ἴϲθι, Θεάγενεϲ, οὐδ’ ἂν τὸ παρὸν τοῦτο ἀλλήλοιϲ διελεγό- -->
	wkln := wl2
	lines := []str.DbWorkline{wkln}
	zaplunates := false
	regularizewidth := false
	out := BuildBrowserTable(wkln.TbIndex, lines, zaplunates, regularizewidth)
	fmt.Println(out)
	//if out != WANT {
	//	t.Errorf("\nmismatch:\n%v\n\n", out)
	//}

}
