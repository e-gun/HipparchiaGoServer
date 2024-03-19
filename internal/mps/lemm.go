//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package mps

import (
	"context"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/db"
	"github.com/e-gun/HipparchiaGoServer/internal/gen"
	"github.com/e-gun/HipparchiaGoServer/internal/str"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"strings"
)

// LemmaMapper - map[string]DbLemma for all lemmata
func LemmaMapper() map[string]*str.DbLemma {
	// example: {dorsum 24563373 [dorsum dorsone dorsa dorsoque dorso dorsoue dorsis dorsi dorsisque dorsumque]}

	// hipparchiaDB=# \d greek_lemmata
	//                       Table "public.greek_lemmata"
	//      Column      |         Type          | Collation | Nullable | Default
	//------------------+-----------------------+-----------+----------+---------
	// dictionary_entry | character varying(64) |           |          |
	// xref_number      | integer               |           |          |
	// derivative_forms | text[]                |           |          |
	//Indexes:
	//    "greek_lemmata_idx" btree (dictionary_entry)

	// a list of 152k words is too long to send to 'getlemmahint' without offering quicker access
	// [HGS] [B1: 0.167s][Δ: 0.167s] unnested lemma map built (152382 items)

	// move to pgx v5 slows this function down (and will add .1s to startup time...):
	// [HGS] [B1: 0.436s][Δ: 0.436s] unnested lemma map built (152382 items)
	// see devel-mutex 1.0.7 at e841c135f22ffaae26cb5cc29e20be58bf4801d7 vs 9457ace03e048c0e367d132cef595ed1661a8c12
	// but pgx v5 does seem faster and more memory efficient in general: must not like returning huge lists

	const (
		THEQUERY = `SELECT dictionary_entry, xref_number, derivative_forms FROM %s_lemmata`
	)

	// note that the v --> u here will push us to stripped_line SearchMap instead of accented_line
	// clean := strings.NewReplacer("-", "", "¹", "", "²", "", "³", "", "j", "i", "v", "u")
	clean := strings.NewReplacer("-", "", "j", "i", "v", "u")

	unnested := make(map[string]*str.DbLemma, vv.DBLMMAPSIZE)

	// use the older iterative idiom to facilitate working with pointers: "foreach" idiom will fight you...
	for _, lg := range vv.TheLanguages {
		q := fmt.Sprintf(THEQUERY, lg)
		foundrows, err := db.SQLPool.Query(context.Background(), q)
		Msg.EC(err)
		for foundrows.Next() {
			thehit := &str.DbLemma{}
			e := foundrows.Scan(&thehit.Entry, &thehit.Xref, &thehit.Deriv)
			Msg.EC(e)
			thehit.Entry = clean.Replace(thehit.Entry)
			unnested[thehit.Entry] = thehit
		}
		foundrows.Close()
	}

	return unnested
}

// NestedLemmaMapper - map[string]map[string]DbLemma for the hinter
func NestedLemmaMapper(unnested map[string]*str.DbLemma) map[string]map[string]*str.DbLemma {
	// 20.96MB    20.96MB (flat, cum)  7.91% of Total
	// you need both a nested and the unnested version; nested is for the hinter

	nested := make(map[string]map[string]*str.DbLemma, vv.NESTEDLEMMASIZE)
	swap := strings.NewReplacer("j", "i", "v", "u")
	for k, v := range unnested {
		rbag := []rune(v.Entry)[0:2]
		rbag = gen.StripaccentsRUNE(rbag)
		bag := strings.ToLower(string(rbag))
		bag = swap.Replace(bag)
		if _, y := nested[bag]; !y {
			nested[bag] = make(map[string]*str.DbLemma)
		}
		nested[bag][k] = v
	}
	return nested
}
