package mps

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"os"
	"slices"
)

// figure out all the characters we are actually using

// the db results need to be supplemented with the special characters in "frontpage.html"
// so the easiest thing to do is to add "fontsubsetting/inuse.txt" into "frontpage.html" in the
// comments and then to subset off of that...

// find the subset via
// % glyphhanger ./web/emb/frontpage.html

//U+A,U+20,U+22,U+28,U+29,U+2C,U+2E,U+2F,U+32,U+33,U+3A,U+41-49,U+4C-4E,U+50,U+52,U+53,U+55-59,U+5B,U+5D,U+5F,U+61-70,U+72-7B,U+7D,U+A0,U+B7,U+3BB,U+3C4,U+20D7,U+2474,U+249C,U+24AD,U+24B8,U+24B9,U+24BC,U+24BE,U+24C1,U+2780-2784,U+278A-278E,U+28B1,U+FE5F,U+1F130,U+1F135,U+1F144,U+1F146

func SubsetChars() {
	m := graballtables()
	s := mapintostring(m)
	outputtofile("fontsubsetting/inuse.txt", s)
}

// graballtables
func graballtables() map[rune]bool {
	var maps []map[rune]bool
	count := 0
	for _, a := range AllAuthors {
		count++
		t := grabonetable(a.UID)
		c := allunicodeinuse(t)
		maps = append(maps, c)
		if count%250 == 0 {
			fmt.Println(count)
		}
	}
	m := collapsemaps(maps)
	return m
}

// grabonetable get everything from a table
func grabonetable(table string) *str.WorkLineBundle {
	const (
		QTMPL = "SELECT %s FROM %s"
	)

	var prq str.PrerolledQuery
	prq.TempTable = ""
	prq.PsqlQuery = fmt.Sprintf(QTMPL, db.WORLINETEMPLATE, table)

	foundlines := db.GetWorklineBundle(prq)
	return foundlines
}

func allunicodeinuse(wlb *str.WorkLineBundle) map[rune]bool {
	rr := wlb.Yield()

	// i32 is the unicode value of the character
	chars := map[rune]bool{}
	for r := range rr {
		runes := []rune(r.MarkedUp)
		for _, c := range runes {
			chars[c] = true
		}
	}
	return chars
}

func collapsemaps(maps []map[rune]bool) map[rune]bool {
	chars := map[rune]bool{}
	for _, m := range maps {
		for c := range m {
			chars[c] = true
		}
	}
	return chars
}

func mapintostring(m map[rune]bool) string {
	var runes []rune
	for r := range m {
		runes = append(runes, r)
	}
	slices.Sort(runes)
	return string(runes)
}

func outputtofile(fn string, s string) {
	if err := os.WriteFile(fn, []byte(s), 0666); err != nil {
		fmt.Println("Error writing file:", err)
	}

}
