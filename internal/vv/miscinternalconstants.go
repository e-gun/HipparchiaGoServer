//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vv

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"regexp"
	"time"
)

const (
	AVGWORDSPERLINE        = 8  // hard coding a suspect assumption
	CHARSPERLINE           = 60 // used by vector to preallocate memory: set it closer to a max than a real average
	GENRESTOCOUNT          = 5
	INCERTADATE            = 9999
	JSONINDENT             = "  "
	LENGTHOFAUTHORID       = 6
	LENGTHOFWORKID         = 3
	NESTEDLEMMASIZE        = 544
	NUMBEROFCITATIONLEVELS = 6
	TERMINATIONS           = `(\s|\.|\]|\<|⟩|\)|’|”|\!|,|:|;|}|\?|⸥|«|·|$)` // circular imports means this is declared 2x... see also "gen.greekandlatin.go"
	VARIADATE              = 8888
)

var (
	IsGreek      = regexp.MustCompile("[α-ωϲῥἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάἐἑἒἓἔἕὲέἰἱἲἳἴἵἶἷὶίῐῑῒΐῖῗὀὁὂὃὄὅόὸὐὑὒὓὔὕὖὗϋῠῡῢΰῦῧύὺᾐᾑᾒᾓᾔᾕᾖᾗῂῃῄῆῇἤἢἥἣὴήἠἡἦἧὠὡὢὣὤὥὦὧᾠᾡᾢᾣᾤᾥᾦᾧῲῳῴῶῷώὼ]")
	TheLanguages = []string{"greek", "latin"}
	LaunchTime   = time.Now()
	DbLmMapSize  = 159265 // will be (re)set at launch time; but the supplied value is almost certainly right
)

func SetLemmCount(p *pgxpool.Pool) int {
	const (
		Q = `SELECT count(*) FROM %s_lemmata`
	)

	ct := 0
	langs := []string{"greek", "latin"}

	for _, lang := range langs {
		theresult, err := p.Query(context.Background(), fmt.Sprintf(Q, lang))
		if err != nil {
			panic(err)
		}
		for theresult.Next() {
			theval := 0
			e := theresult.Scan(&theval)
			if e != nil {
				panic(e)
			}
			ct += theval
		}
		theresult.Close()
	}

	return ct
}
