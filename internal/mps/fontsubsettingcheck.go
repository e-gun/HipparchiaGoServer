//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package mps

import (
	"fmt"
	"os"
	"slices"

	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"golang.org/x/exp/maps"
)

const (
	FRONTPAGE = `web/emb/htm/frontpage.html`
	OUTFILE   = `inuse.txt`
)

// figure out all the characters we are actually using

// the db results need to be supplemented with the special characters in "frontpage.html"
// so the easiest thing to do is to add "fontsubsetting/inuse.txt" into "frontpage.html" in the
// comments and then to subset off of that...

// find the subset via
// % glyphhanger ./web/emb/frontpage.html

//U+A,U+20,U+22,U+28,U+29,U+2C,U+2E,U+2F,U+32,U+33,U+3A,U+41-49,U+4C-4E,U+50,U+52,U+53,U+55-59,U+5B,U+5D,U+5F,U+61-70,U+72-7B,U+7D,U+A0,U+B7,U+3BB,U+3C4,U+20D7,U+2474,U+249C,U+24AD,U+24B8,U+24B9,U+24BC,U+24BE,U+24C1,U+2780-2784,U+278A-278E,U+28B1,U+FE5F,U+1F130,U+1F135,U+1F144,U+1F146

// SubsetChars - only runs if the right line in "main.go" is uncommented; will build the string of distinct chars needed
func SubsetChars() {
	// get the runes
	m := graballtables()
	l := grabsomelexrunes()
	f := grabthefrontpage()

	// merge them
	maps.Copy(m, l) // (dst, src)
	maps.Copy(m, f)

	// stringify and output
	s := mapintostring(m)
	// rare dictionary chars that might not get caught, etc
	kludge := "ṛṣσ"
	s = s + kludge
	outputtofile(OUTFILE, s)
}

// graballtables - get all lines from all authors(!) and then see what runes are in use in them
func graballtables() map[rune]bool {
	fmt.Println("graballtables start")
	var mymaps []map[rune]bool
	count := 0
	for _, a := range AllAuthors {
		count++
		t := grabonetable(a.UID)
		c := allunicodeinuse(t)
		mymaps = append(mymaps, c)
		if count%250 == 0 {
			fmt.Println(count)
		}
	}
	m := collapsemaps(mymaps)
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

// grabsomelexrunes - grab some lex entries and then see what runes are in use in them
func grabsomelexrunes() map[rune]bool {
	fmt.Println("grabsomelexrunes")
	// vv.MAXDICTLOOKUP will cap this; but you should see everything there is to see
	gl := db.DictEntryGrabber("", "greek", "entry_name", "~*")
	lt := db.DictEntryGrabber("", "latin", "entry_name", "~*")

	chars := map[rune]bool{}
	for _, l := range gl {
		runes := []rune(l.EntryName)
		for _, c := range runes {
			chars[c] = true
		}
	}

	for _, l := range lt {
		runes := []rune(l.EntryName)
		for _, c := range runes {
			chars[c] = true
		}
	}
	fmt.Println(len(chars))
	return chars
}

// grabthefrontpage - read the front page to add in the special runes there such as "➀"
func grabthefrontpage() map[rune]bool {
	fp, err := os.ReadFile(FRONTPAGE)
	if err != nil {
		fmt.Println(err)
		fp = []byte{}
	}

	st := string(fp)
	chars := map[rune]bool{}
	for _, l := range st {
		chars[l] = true
	}
	return chars
}

// allunicodeinuse - find the unique runes in a WorkLineBundle
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

// collapsemaps - turn a slice of maps into a single map
func collapsemaps(maps []map[rune]bool) map[rune]bool {
	chars := map[rune]bool{}
	for _, m := range maps {
		for c := range m {
			chars[c] = true
		}
	}
	return chars
}

// mapintostring - convert a map of unique runes into a string of unique characters
func mapintostring(m map[rune]bool) string {
	var runes []rune
	for r := range m {
		runes = append(runes, r)
	}
	slices.Sort(runes)
	return string(runes)
}

// outputtofile - save the results to a textfile
func outputtofile(fn string, s string) {
	if err := os.WriteFile(fn, []byte(s), 0666); err != nil {
		fmt.Println("Error writing file:", err)
	}
}
