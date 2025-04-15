package format

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"regexp"
)

// FormatPublicationInfo - does just what you think it does
func FormatPublicationInfo(w str.DbWork) string {
	// 	in:
	//		<volumename>FHG </volumename>4 <press>Didot </press><city>Paris </city><year>1841–1870</year><pages>371 </pages><pagesintocitations>Frr. 1–2</pagesintocitations><editor>Müller, K. </editor>
	//	out:
	//		<span class="pubvolumename">FHG <br /></span><span class="pubpress">Didot , </span><span class="pubcity">Paris , </span><span class="pubyear">1841–1870. </span><span class="pubeditor"> (Müller, K. )</span>

	const (
		REGS = "<%s>(?P<data>.*?)</%s>"
		REGD = "<%d>(?P<data>.*?)</%d>"
	)

	type Swapper struct {
		Name  string
		Sub   int
		Left  string
		Right string
	}

	tags := []Swapper{
		{"volumename", 1, "", " "},
		{"press", 2, " ", ", "},
		{"city", 3, " ", ", "},
		{"year", 4, " ", ". "},
		{"yearreprinted", 5, "[", "] "},
		{"series", 6, " ", ""},
		{"editor", 7, "(", ")"},
		{"work", 8, " ", " "},
		{"pages", 9, " pp. ", ". "},
	}

	pubinfo := ""

	// shorten the strings so you can split
	for _, t := range tags {
		tag := fmt.Sprintf(REGS, t.Name, t.Name)
		pattern := regexp.MustCompile(tag)
		found := pattern.MatchString(w.Pub)
		if found {
			subs := pattern.FindStringSubmatch(w.Pub)
			data := subs[pattern.SubexpIndex("data")]
			pub := fmt.Sprintf(`<%d>%s%s%s</%d>`, t.Sub, t.Left, data, t.Right, t.Sub)
			pubinfo += pub
		}
	}

	pubinfo = gen.AvoidLongLines(pubinfo, vv.MINBROWSERWIDTH+(vv.MINBROWSERWIDTH/2))

	// restore the strings
	var reconstituted string
	for _, t := range tags {
		tag := fmt.Sprintf(REGD, t.Sub, t.Sub)
		pattern := regexp.MustCompile(tag)
		found := pattern.MatchString(pubinfo)
		if found {
			subs := pattern.FindStringSubmatch(pubinfo)
			data := subs[pattern.SubexpIndex("data")]
			pub := fmt.Sprintf(`<span class="pub%s">%s</span>`, t.Name, data)
			reconstituted += pub
		}
	}

	readability := `<br>
	%s
	`
	return fmt.Sprintf(readability, reconstituted)
}
