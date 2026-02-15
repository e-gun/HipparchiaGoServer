package gen

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// you are not supposed to use `strings.Title()`: `cases.Title()` is the replacement
// the issue here is whether `cases.Title(language.Greek)` really works with polytonic, etc

// all of this likely deserves rigorous testing; this is why a separate file has been built for so little code
// at the moment it looks like `cases.Title(language.Greek)` will get both `socrates` and `ὅτι` right
// so maybe this is a non-issue: famous last words

var (
	thecaser = GreekCaser()
)

func HipparchiaUppercaser(s string) string {
	return thecaser.String(s)
}

func GreekCaser() cases.Caser {
	return cases.Title(language.Greek)
}

//func LatinCaser() cases.Caser {
//	return cases.Title(language.English)
//}
