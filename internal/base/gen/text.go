//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package gen

import (
	"fmt"
	"strings"
)

var (
	// QuantityFixer = strings.NewReplacer("ă_", "ā̆", "ā^", "ā̆", "ē^", "ē̆", "ĭ_", "ī̆", "ō^", "ō̆", "A_^", "Ā̆", "A^", "Ă", "A_", "Ā", "E_", "Ē", "E^", "Ĕ", "I_^", "Ī̆", "I_", "Ī", "I^", "Ĭ", "O_", "Ō", "O^", "Ŏ", "U^", "Ŭ", "U_", "Ū", "_^", "̆̄", "_", "̄", "^", "̆")
	QuantityFixer  = strings.NewReplacer("_^", "̄̆", "_", "̄", "^", "̆")
	NumberStripper = strings.NewReplacer("-", "", "¹", "", "²", "", "³", "")
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
	converter := exrunefd // see top of setsandslices.go
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
