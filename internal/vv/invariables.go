//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vv

import (
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"regexp"
	"time"
)

var (
	TheCorpora    = []string{GREEKCORP, LATINCORP, INSCRIPTCORP, CHRISTINSC, PAPYRUSCORP}
	TheLanguages  = []string{"greek", "latin"}
	ServableFonts = map[string]str.FontTempl{"Noto": NotoFont, "Roboto": RobotoFont, "Fira": FiraFont} // cf rt-embhcss.go
	LaunchTime    = time.Now()
)

var (
	IsGreek = regexp.MustCompile("[α-ωϲῥἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάἐἑἒἓἔἕὲέἰἱἲἳἴἵἶἷὶίῐῑῒΐῖῗὀὁὂὃὄὅόὸὐὑὒὓὔὕὖὗϋῠῡῢΰῦῧύὺᾐᾑᾒᾓᾔᾕᾖᾗῂῃῄῆῇἤἢἥἣὴήἠἡἦἧὠὡὢὣὤὥὦὧᾠᾡᾢᾣᾤᾥᾦᾧῲῳῴῶῷώὼ]")
)

//
// FONTS
//

// the fonts we know how to serve
// NB: Inter, SourceSans and Ubuntu have been toyed with: Inter lacks both condensed and semi-condensed

var (
	NotoFont = str.FontTempl{
		Type:             "truetype",
		ShrtType:         "ttf",
		Bold:             "NotoSansDisplay-Bold.ttf",
		BoldItalic:       "NotoSansDisplay-BoldItalic.ttf",
		CondensedBold:    "NotoSansDisplay_Condensed-SemiBold.ttf",
		CondensedItalic:  "NotoSansDisplay_Condensed-Italic.ttf",
		CondensedRegular: "NotoSansDisplay_Condensed-Regular.ttf",
		SemiCondRegular:  "NotoSansDisplay_SemiCondensed-Regular.ttf",
		SemiCondItalic:   "NotoSansDisplay_SemiCondensed-Italic.ttf",
		Italic:           "NotoSansDisplay-Italic.ttf",
		Light:            "NotoSansDisplay-ExtraLight.ttf",
		Mono:             "NotoSansMono_Condensed-Regular.ttf",
		Regular:          "NotoSansDisplay-Regular.ttf",
		SemiBold:         "NotoSansDisplay-SemiBold.ttf",
		Thin:             "NotoSansDisplay-Thin.ttf",
	}
	NotoFontSS = str.FontTempl{
		Type:             "truetype",
		ShrtType:         "ttf",
		Bold:             "NotoSansDisplay-BoldSubset.ttf",
		BoldItalic:       "NotoSansDisplay-BoldItalicSubset.ttf",
		CondensedBold:    "NotoSansDisplay_Condensed-SemiBoldSubset.ttf",
		CondensedItalic:  "NotoSansDisplay_Condensed-ItalicSubset.ttf",
		CondensedRegular: "NotoSansDisplay_Condensed-RegularSubset.ttf",
		SemiCondRegular:  "NotoSansDisplay_SemiCondensed-RegularSubset.ttf",
		SemiCondItalic:   "NotoSansDisplay_SemiCondensed-ItalicSubset.ttf",
		Italic:           "NotoSansDisplay-ItalicSubset.ttf",
		Light:            "NotoSansDisplay-ExtraLightSubset.ttf",
		Mono:             "NotoSansMono_Condensed-RegularSubset.ttf",
		Regular:          "NotoSansDisplay-RegularSubset.ttf",
		SemiBold:         "NotoSansDisplay-SemiBoldSubset.ttf",
		Thin:             "NotoSansDisplay-ThinSubset.ttf",
	}
	FiraFont = str.FontTempl{
		Type:             "truetype",
		ShrtType:         "ttf",
		Bold:             "FiraSans-Bold.ttf",
		BoldItalic:       "FiraSans-BoldItalic.ttf",
		CondensedBold:    "FiraSansCondensed-Bold.ttf",
		CondensedItalic:  "FiraSansCondensed-Italic.ttf",
		CondensedRegular: "FiraSansCondensed-Regular.ttf",
		SemiCondRegular:  "FiraSansCondensed-Regular.ttf", // semi dne
		SemiCondItalic:   "FiraSansCondensed-Italic.ttf",
		Italic:           "FiraSans-Italic.ttf",
		Light:            "FiraSans-Light.ttf",
		Mono:             "FiraMono-Regular.ttf",
		Regular:          "FiraSans-Regular.ttf",
		SemiBold:         "FiraSans-SemiBold.ttf",
		Thin:             "FiraSans-Thin.ttf",
	}
	FiraFontSS = str.FontTempl{
		Type:             "truetype",
		ShrtType:         "ttf",
		Bold:             "FiraSans-BoldSubset.ttf",
		BoldItalic:       "FiraSans-BoldItalicSubset.ttf",
		CondensedBold:    "FiraSansCondensed-BoldSubset.ttf",
		CondensedItalic:  "FiraSansCondensed-ItalicSubset.ttf",
		CondensedRegular: "FiraSansCondensed-RegularSubset.ttf",
		SemiCondRegular:  "FiraSansCondensed-RegularSubset.ttf", // semi dne
		SemiCondItalic:   "FiraSansCondensed-ItalicSubset.ttf",
		Italic:           "FiraSans-ItalicSubset.ttf",
		Light:            "FiraSans-LightSubset.ttf",
		Mono:             "FiraMono-RegularSubset.ttf",
		Regular:          "FiraSans-RegularSubset.ttf",
		SemiBold:         "FiraSans-SemiBoldSubset.ttf",
		Thin:             "FiraSans-ThinSubset.ttf",
	}
	RobotoFont = str.FontTempl{
		Type:             "truetype",
		ShrtType:         "ttf",
		Bold:             "Roboto-Bold.ttf",
		BoldItalic:       "Roboto-BoldItalic.ttf",
		CondensedBold:    "RobotoCondensed-Bold.ttf",
		CondensedItalic:  "RobotoCondensed-Italic.ttf",
		CondensedRegular: "RobotoCondensed-Regular.ttf",
		SemiCondRegular:  "RobotoCondensed-Regular.ttf", // semi dne
		SemiCondItalic:   "RobotoCondensed-Italic.ttf",
		Italic:           "Roboto-Italic.ttf",
		Light:            "Roboto-Light.ttf",
		Mono:             "RobotoMono-Regular.ttf",
		Regular:          "Roboto-Regular.ttf",
		SemiBold:         "Roboto-Medium.ttf",
		Thin:             "Roboto-Thin.ttf",
	}
)
