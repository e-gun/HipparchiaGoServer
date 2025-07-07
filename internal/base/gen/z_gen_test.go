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
