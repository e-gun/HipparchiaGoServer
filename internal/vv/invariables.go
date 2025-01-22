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
	ServableFonts = map[string]str.FontTempl{"Noto": NotoFont, "Roboto": RobotoFont, "Fira": FiraFont,
		"Inter": InterFont, "Brill": BrillFont} // cf rt-embhcss.go
	LaunchTime = time.Now()
)

var (
	IsGreek = regexp.MustCompile("[α-ωϲῥἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάἐἑἒἓἔἕὲέἰἱἲἳἴἵἶἷὶίῐῑῒΐῖῗὀὁὂὃὄὅόὸὐὑὒὓὔὕὖὗϋῠῡῢΰῦῧύὺᾐᾑᾒᾓᾔᾕᾖᾗῂῃῄῆῇἤἢἥἣὴήἠἡἦἧὠὡὢὣὤὥὦὧᾠᾡᾢᾣᾤᾥᾦᾧῲῳῴῶῷώὼ]")
)

//
// FONTS
//

// the fonts we know how to serve
// NB: Inter, SourceSans and Ubuntu have been toyed with: Inter lacks both condensed and semi-condensed
// now serving subsetted fonts; used to serve whole fonts: NotoFontR, etc.

var (
	BrillFont = str.FontTempl{
		Type:             "truetype",
		ShrtType:         "ttf",
		Bold:             "Brill-BoldSubset.ttf",
		BoldItalic:       "Brill-BoldItalicSubset.ttf",
		CondensedBold:    "Brill-RomanSubset.ttf",
		CondensedItalic:  "Brill-RomanSubset.ttf",
		CondensedRegular: "Brill-RomanSubset.ttf",
		SemiCondRegular:  "Brill-RomanSubset.ttf",
		SemiCondItalic:   "Brill-RomanSubset.ttf",
		Italic:           "Brill-ItalicSubset.ttf",
		Light:            "Brill-RomanSubset.ttf",
		Mono:             "NotoSansMono_Condensed-RegularSubset.ttf",
		Regular:          "Brill-RomanSubset.ttf",
		SemiBold:         "Brill-RomanSubset.ttf",
		Thin:             "Brill-RomanSubset.ttf",
		HasLunateSigma:   true,
		NeedsManualStyle: []string{"hipparchialightstatic", "hipparchiasemicondensedstatic", "hipparchiasemicondenseditalicstatic",
			"hipparchiacondensedboldstatic", "hipparchiacondenseditalicstatic", "hipparchiathinstatic"},
	}
	FiraFont = str.FontTempl{
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
		HasLunateSigma:   true,
		NeedsManualStyle: []string{},
	}
	NotoFont = str.FontTempl{
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
		HasLunateSigma:   true,
		NeedsManualStyle: []string{},
	}
	InterFont = str.FontTempl{
		Type:             "truetype",
		ShrtType:         "ttf",
		Bold:             "Inter_18pt-BoldSubset.ttf",
		BoldItalic:       "Inter_18pt-BoldItalicSubset.ttf",
		CondensedBold:    "InterTight-BoldSubset.ttf",
		CondensedItalic:  "InterTight-ItalicSubset.ttf",
		CondensedRegular: "InterTight-RegularSubset.ttf",
		SemiCondRegular:  "InterTight-RegularSubset.ttf",
		SemiCondItalic:   "InterTight-ItalicSubset.ttf",
		Italic:           "Inter_18pt-ItalicSubset.ttf",
		Light:            "Inter_18pt-LightSubset.ttf",
		Mono:             "iosevka-regularSubset.woff2",
		Regular:          "Inter_18pt-RegularSubset.ttf",
		SemiBold:         "Inter_18pt-MediumSubset.ttf",
		Thin:             "Inter_18pt-ThinSubset.ttf",
		HasLunateSigma:   true,
		NeedsManualStyle: []string{},
	}
	RobotoFont = str.FontTempl{
		Type:             "truetype",
		ShrtType:         "ttf",
		Bold:             "Roboto-BoldSubset.ttf",
		BoldItalic:       "Roboto-BoldItalicSubset.ttf",
		CondensedBold:    "RobotoCondensed-BoldSubset.ttf",
		CondensedItalic:  "RobotoCondensed-ItalicSubset.ttf",
		CondensedRegular: "RobotoCondensed-RegularSubset.ttf",
		SemiCondRegular:  "RobotoCondensed-RegularSubset.ttf", // semi dne
		SemiCondItalic:   "RobotoCondensed-ItalicSubset.ttf",
		Italic:           "Roboto-ItalicSubset.ttf",
		Light:            "Roboto-LightSubset.ttf",
		Mono:             "RobotoMono-RegularSubset.ttf",
		Regular:          "Roboto-RegularSubset.ttf",
		SemiBold:         "Roboto-MediumSubset.ttf",
		Thin:             "Roboto-ThinSubset.ttf",
		HasLunateSigma:   true,
		NeedsManualStyle: []string{},
	}
	NotoFontR = str.FontTempl{
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
		HasLunateSigma:   true,
		NeedsManualStyle: []string{},
	}
	FiraFontR = str.FontTempl{
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
		HasLunateSigma:   true,
		NeedsManualStyle: []string{},
	}
	RobotoFontR = str.FontTempl{
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
		HasLunateSigma:   true,
		NeedsManualStyle: []string{},
	}
)
