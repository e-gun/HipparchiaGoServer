package gen

import (
	"fmt"
	"strings"
	"testing"
)

func TestStripextraaccent(t *testing.T) {
	tg := `ἀληθῶϲ οἷόν τε θαυμάζεταί ὀνομάτων δριμεῖαί`
	words := strings.Split(tg, " ")
	for _, word := range words {
		w := StripExtraAccent(word)
		fmt.Println(word, w)
	}
}

func TestCapsVariants(t *testing.T) {
	tg := `ἔρωτα·`
	cv := CapsVariants(tg)
	fmt.Println(cv)
}

func TestDeLunate(t *testing.T) {
	tg := `Τὴν οὖν τῶν ϲωμάτων ϲύνταξιν ϲκεψαμένουϲ πρὸϲ || τἀν Πρυτανείῳ ς. public maintenance`
	cv := DeLunate(tg)
	fmt.Println(cv)
}

func TestLexDeLunate(t *testing.T) {
	tg := `τἀν Πρυτανείῳ ς. public maintenance || quote lang="grc">ἐργώναις ς</quote>.`
	cv := LexDeLunate(tg)
	fmt.Println(cv)
}

func TestReLunate(t *testing.T) {
	tg := `Τὴν οὖν τῶν σωμάτων σύνταξιν σκεψαμένους πρὸς`
	cv := ReLunate(tg)
	fmt.Println(cv)
}
