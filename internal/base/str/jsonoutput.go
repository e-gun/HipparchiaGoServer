//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package str

type SearchOutputJSON struct {
	Title         string `json:"title"`
	Searchsummary string `json:"searchsummary"`
	Found         string `json:"found"`
	Image         string `json:"image"`
	JS            string `json:"js"`
}

type JSONOutFeeder struct {
	SU string `json:"searchsummary"`
	HT string `json:"thehtml"`
	NJ string `json:"newjs"`
	NT string `json:"title"`
	JS string `json:"js"`
}
