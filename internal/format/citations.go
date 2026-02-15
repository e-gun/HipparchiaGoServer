//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package format

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/search"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
)

// BasicCitation - produce a comma-separated citation from a DbWorkline: e.g., "book 5, chapter 37, section 5, line 3"
func BasicCitation(l str.DbWorkline) string {
	w := search.DbWlnMyWk(&l)
	cf := w.CitationFormat()
	loc := l.FindLocus()
	cf = cf[vv.NUMBEROFCITATIONLEVELS-(len(loc)) : vv.NUMBEROFCITATIONLEVELS]

	var cit []string
	for i := range loc {
		cit = append(cit, fmt.Sprintf("%s %s", cf[i], loc[i]))
	}
	fullcit := strings.Join(cit, ", ")
	return fullcit
}

// SelectivelyDisplayCitations - only show line numbers every N lines, etc.
func SelectivelyDisplayCitations(theline str.DbWorkline, previous str.DbWorkline, focus int) string {
	// figure out whether to display a citation
	// pulled this out because it is common with the textbuilder (who will always send "0" as the focus)

	// [a] if thisline.samelevelas(previousline) is not True:...
	// [b] if linenumber % linesevery == 0
	// [c] always give a citation for the focus line
	citation := strings.Join(theline.FindLocus(), ".")

	z, e := strconv.Atoi(theline.Lvl0Value)
	if e != nil {
		z = 0
	}

	if !theline.SameLevelAs(previous) || z%vv.SHOWCITATIONEVERYNLINES == 0 || theline.TbIndex == focus {
		// display citation
	} else {
		citation = ""
	}
	return citation
}
