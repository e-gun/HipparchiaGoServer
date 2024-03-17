package structs

type DbLemma struct {
	// dictionary_entry | xref_number |    derivative_forms
	Entry string
	Xref  int
	Deriv []string
}

func (dbl DbLemma) EntryRune() []rune {
	return []rune(dbl.Entry)
}
