//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package gen

//
// STRINGS and []RUNE
//

// Purgechars - drop any of the chars in the bad-string from the check-string
func Purgechars(bad string, checking string) string {
	rb := []rune(bad)
	reducer := make(map[rune]bool, len(rb))
	for _, r := range rb {
		reducer[r] = true
	}

	var stripped []rune
	for _, x := range []rune(checking) {
		if _, skip := reducer[x]; !skip {
			stripped = append(stripped, x)
		}
	}
	s := string(stripped)
	return s
}
