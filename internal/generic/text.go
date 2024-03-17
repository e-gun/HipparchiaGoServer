package generic

import (
	"fmt"
	"strings"
)

// AvoidLongLines - insert "<br>" into strings that are too long
func AvoidLongLines(untrimmed string, maxlen int) string {
	if len(untrimmed) > maxlen {
		untrimmed = strings.Replace(untrimmed, ";", "; ", -1)
		pi := strings.Split(untrimmed, " ")
		var trimmed string
		breaks := 0
		reset := 0
		crop := maxlen
		for i := 0; i < len(pi); i++ {
			trimmed += pi[i] + " "
			if len(trimmed) > reset+crop {
				trimmed += "<br>"
				breaks += 1
				reset = len(trimmed)
			}
		}
		untrimmed = trimmed
	}
	return untrimmed
}

// UniversalPatternMaker - feeder for SearchTermFinder()
func UniversalPatternMaker(term string) string {
	// also used by searchformatting.go
	// converter := extendedrunefeeder()
	converter := ERuneFd // see top of setsandslices.go
	st := []rune(term)
	var stre string
	for _, r := range st {
		if _, ok := converter[r]; ok {
			re := fmt.Sprintf("[%s]", string(converter[r]))
			stre += re
		} else {
			stre += string(r)
		}
	}
	stre = fmt.Sprintf("(%s)", stre)
	return stre
}
