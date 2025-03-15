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

type CommonJSONOutput struct {
	Sum string `json:"searchsummary"`
	Htm string `json:"thehtml"`
	JS  string `json:"newjs"`
	Tit string `json:"title"`
}
