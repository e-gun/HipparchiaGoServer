//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package str

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/base/gen"
	"regexp"
	"strings"
)

var (
	UVSubs         = true //to enable setting dbw.GetMarked() outside the module; configatlaunch.go sets the real value
	smallcapsregex = regexp.MustCompile("(<span class=\"smallcapitals\">)([^<]*)(</span>)")
)

// hipparchiaDB-# \d gr0001
//                                     Table "public.gr0001"
//      Column      |          Type          | Collation | Nullable |           Default
//------------------+------------------------+-----------+----------+-----------------------------
// index            | integer                |           | not null | nextval('gr0001'::regclass)
// wkuniversalid    | character varying(10)  |           |          |
// level_05_value   | character varying(64)  |           |          |
// level_04_value   | character varying(64)  |           |          |
// level_03_value   | character varying(64)  |           |          |
// level_02_value   | character varying(64)  |           |          |
// level_01_value   | character varying(64)  |           |          |
// level_00_value   | character varying(64)  |           |          |
// marked_up_line   | text                   |           |          |
// accented_line    | text                   |           |          |
// stripped_line    | text                   |           |          |
// hyphenated_words | character varying(128) |           |          |
// annotations      | character varying(256) |           |          |
//Indexes:
//    "gr0001_index_key" UNIQUE CONSTRAINT, btree (index)
//    "gr0001_mu_trgm_idx" gin (accented_line gin_trgm_ops)
//    "gr0001_st_trgm_idx" gin (stripped_line gin_trgm_ops)

const (
	WKLNHYPERLNKTEMPL = `index/%s/%s/%d`
	WLNMETADATATEMPL  = `<span class="embeddedannotations foundwork">$1</span>`
	EMPTYLEVELINFO    = "-1"
)

var (
	NoHTML   = regexp.MustCompile("<[^>]*>") // crude, and will not do all of everything
	Metadata = regexp.MustCompile(`<hmu_metadata_(.*?) value="(.*?)" />`)
	MDFormat = regexp.MustCompile(`&3(.*?)&`) // see andsubstitutes in betacodefontshifts.py
	// MDRemap has to be kept in sync w/ l 150 of rt-browser.go ()
	MDRemap = map[string]string{"provenance": "loc:", "documentnumber": "#", "publicationinfo": "pub:", "notes": "",
		"city": "c:", "region": "r:", "date": "d:"}
)

type DbWorkline struct {
	WkUID       string
	TbIndex     int
	Lvl5Value   string
	Lvl4Value   string
	Lvl3Value   string
	Lvl2Value   string
	Lvl1Value   string
	Lvl0Value   string
	MarkedUp    string // converting this and others to pointers will not save you memory
	Accented    string // converting to pointers might give you a very slight speed boost
	Stripped    string // converting to pointers can produce nil pointer problems that need constant checks
	Hyphenated  string
	Annotations string
	// beyond the db stuff; do not make this "public": pgx.RowToStructByPos will balk
	embnotes map[string]string
}

func (dbw *DbWorkline) GetNotes() map[string]string {
	return dbw.embnotes
}

func (dbw *DbWorkline) FindLocus() []string {
	loc := [NUMBEROFCITATIONLEVELS]string{
		dbw.Lvl5Value,
		dbw.Lvl4Value,
		dbw.Lvl3Value,
		dbw.Lvl2Value,
		dbw.Lvl1Value,
		dbw.Lvl0Value,
	}

	var trim []string
	for _, l := range loc {
		if l != EMPTYLEVELINFO {
			trim = append(trim, l)
		}
	}
	return trim
}

// AuID - gr0001w001 --> gr0001
func (dbw *DbWorkline) AuID() string {
	return dbw.WkUID[:LENGTHOFAUTHORID]
}

// WkID - gr0001w001 --> 001
func (dbw *DbWorkline) WkID() string {
	return dbw.WkUID[LENGTHOFAUTHORID+1:]
}

func (dbw *DbWorkline) FindCorpus() string {
	// gr0001w001 --> gr
	return dbw.WkUID[0:2]
}

func (dbw *DbWorkline) BuildHyperlink() string {
	if len(dbw.WkUID) == 0 {
		// FormatWithContextResults() will trigger this
		// Msg.TMI("BuildHyperlink() on empty dbworkline")
		return ""
	}
	return fmt.Sprintf(WKLNHYPERLNKTEMPL, dbw.AuID(), dbw.WkID(), dbw.TbIndex)
}

func (dbw *DbWorkline) GatherMetadata() {
	md := make(map[string]string)
	if Metadata.MatchString(dbw.MarkedUp) {
		mm := Metadata.FindAllStringSubmatch(dbw.MarkedUp, -1)
		for _, m := range mm {
			// sample location:
			// hipparchiaDB=# select index, marked_up_line from lt0474 where index = 116946;
			//  <hmu_metadata_documentnumber value="177" /><hmu_metadata_provenance value="in Cumano" /><hmu_metadata_date value="xiii Kal. Dec.(?) 46" />&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;CICERO PAETO
			md[m[1]] = m[2]
		}

		dbw.MarkedUp = Metadata.ReplaceAllString(dbw.MarkedUp, "")
		for k, v := range md {
			md[k] = MDFormat.ReplaceAllString(v, WLNMETADATATEMPL)
			if _, y := MDRemap[k]; y {
				md[MDRemap[k]] = md[k]
				delete(md, k)
			}
		}
	}
	dbw.embnotes = md
}

// PurgeMetadata - delete the line Metadata
func (dbw *DbWorkline) PurgeMetadata() {
	if Metadata.MatchString(dbw.MarkedUp) {
		dbw.MarkedUp = Metadata.ReplaceAllString(dbw.MarkedUp, "")
	}
}

// GetMarked - do a v --> u transformation on the marked up line
func (dbw *DbWorkline) GetMarked() string {
	return dbw.UVXform(dbw.MarkedUp)
}

// GetAccented - do a v --> u transformation on the accented up line (rt-browser.go needs this for clicks)
func (dbw *DbWorkline) GetAccented() string {
	return dbw.UVXform(dbw.Accented)
}

func (dbw *DbWorkline) UVXform(s string) string {
	if !UVSubs {
		return s
	}

	// toggling this via lnch.Config is not possible since that would lead to circular imports

	// NB: no checks for inside/outside the markup and so a real chance of generating invalid markup
	// the penalty should only be a failure to format/display correctly; and the fix can be made in the CSS rules
	// u/v vigilance there is probably a better idea than regex on all the embedded html...

	// negative side-effect is that all-caps CSS will now look like ADUOCATI instead of ADVOCATI (Pl., Poen. 515)
	// as the underlying text is "Advocati" before you rewrite

	// but, relatively speaking, you have gained more than you have lost here?

	// another issue: roman numerals...
	// if level_00_value ends in "t", then this is something like III.iv as an act marker or letter number
	// but there is no way to check for internal numbers in Pliny the Elder, et al.
	// "number" passages will look bad as "v" for 5 turns into "u": "O casum mirificum! u Id. cum ante lucem..."

	if strings.HasSuffix(dbw.Lvl0Value, "t") {
		return s
	} else {
		// preserve smallcaps V
		s = gen.UVcaps(s)
		return smallcapsuv(s)
	}
}

// ShowMarkup - reveal markup in a line
func (dbw *DbWorkline) ShowMarkup() string {
	clean := strings.NewReplacer("<", "&lt;", ">", "&gt;")
	return clean.Replace(dbw.MarkedUp)
}

func (dbw *DbWorkline) SameLevelAs(other DbWorkline) bool {
	// to help toggle the counters when building texts
	one := dbw.Lvl1Value == other.Lvl1Value
	two := dbw.Lvl2Value == other.Lvl2Value
	three := dbw.Lvl3Value == other.Lvl3Value
	four := dbw.Lvl4Value == other.Lvl4Value
	five := dbw.Lvl5Value == other.Lvl5Value
	if one && two && three && four && five {
		return true
	} else {
		return false
	}
}

func (dbw *DbWorkline) StrippedSlice() []string {
	return strings.Split(dbw.Stripped, " ")
}

func (dbw *DbWorkline) AccentedSlice() []string {
	return strings.Split(dbw.Accented, " ")
}

func (dbw *DbWorkline) MarkedUpSlice() []string {
	cln := NoHTML.ReplaceAllString(dbw.MarkedUp, "")
	return strings.Split(cln, " ")
}

func (dbw *DbWorkline) Citation() string {
	return strings.Join(dbw.FindLocus(), ".")
}

// Lvls - report the number of active levels for this line
func (dbw *DbWorkline) Lvls() int {
	//alternate is: "return dbw.MyWk().CountLevels()"
	vl := []string{dbw.Lvl0Value, dbw.Lvl1Value, dbw.Lvl2Value, dbw.Lvl3Value, dbw.Lvl4Value, dbw.Lvl5Value}
	empty := gen.ContainsN(vl, EMPTYLEVELINFO)
	return NUMBEROFCITATIONLEVELS - empty
}

func (dbw *DbWorkline) LvlVal(lvl int) string {
	// what is the value at level N?
	switch lvl {
	case 0:
		return dbw.Lvl0Value
	case 1:
		return dbw.Lvl1Value
	case 2:
		return dbw.Lvl2Value
	case 3:
		return dbw.Lvl3Value
	case 4:
		return dbw.Lvl4Value
	case 5:
		return dbw.Lvl5Value
	default:
		return ""
	}
}

type LevelValues struct {
	// for JSON output...
	// {"totallevels": 3, "level": 2, "label": "book", "low": "1", "high": "3", "range": ["1", "2", "3"]}
	Total int      `json:"totallevels"`
	AtLvl int      `json:"level"`
	Label string   `json:"label"`
	Low   string   `json:"low"`
	High  string   `json:"high"`
	Range []string `json:"range"`
}

// smallcapsuv - preserve V in Roman numerals
func smallcapsuv(s string) string {
	// this is a roman numeral...: <span class="smallcapitals">u</span>
	if !strings.Contains(s, "<span class=\"smallcapitals\">") {
		return s
	}

	replacer := func(match string) string {
		f := smallcapsregex.FindStringSubmatch(match)
		if len(f) == 0 {
			fmt.Println("smallcapsuv failed FindStringSubmatch(): " + match)
		}
		// f[2] is the thing inside the span
		vv := strings.ReplaceAll(f[2], "u", "v")
		return f[1] + vv + f[3]
	}
	return smallcapsregex.ReplaceAllStringFunc(s, replacer)
}
