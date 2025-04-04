//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vv

import (
	"regexp"
	"time"
)

const (
	AVGWORDSPERLINE        = 8      // hard coding a suspect assumption
	CHARSPERLINE           = 60     // used by vector to preallocate memory: set it closer to a max than a real average
	DBLMMAPSIZE            = 151701 //[HGS] [B1: 0.310s][Δ: 0.310s] unnested lemma map built (151701 items)
	GENRESTOCOUNT          = 5
	INCERTADATE            = 2500
	JSONINDENT             = "  "
	LENGTHOFAUTHORID       = 6
	LENGTHOFWORKID         = 3
	NESTEDLEMMASIZE        = 543
	NUMBEROFCITATIONLEVELS = 6
	TERMINATIONS           = `(\s|\.|\]|\<|⟩|\)|’|”|\!|,|:|;|\?|⸥|«|·|$)` // circular imports means this is declared 2x... see also "gen.greekandlatin.go"
	VARIADATE              = 2000
)

var (
	IsGreek      = regexp.MustCompile("[α-ωϲῥἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάἐἑἒἓἔἕὲέἰἱἲἳἴἵἶἷὶίῐῑῒΐῖῗὀὁὂὃὄὅόὸὐὑὒὓὔὕὖὗϋῠῡῢΰῦῧύὺᾐᾑᾒᾓᾔᾕᾖᾗῂῃῄῆῇἤἢἥἣὴήἠἡἦἧὠὡὢὣὤὥὦὧᾠᾡᾢᾣᾤᾥᾦᾧῲῳῴῶῷώὼ]")
	TheLanguages = []string{"greek", "latin"}
	LaunchTime   = time.Now()
)
