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
	IsGreek       = regexp.MustCompile("[α-ωϲῥἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάἐἑἒἓἔἕὲέἰἱἲἳἴἵἶἷὶίῐῑῒΐῖῗὀὁὂὃὄὅόὸὐὑὒὓὔὕὖὗϋῠῡῢΰῦῧύὺᾐᾑᾒᾓᾔᾕᾖᾗῂῃῄῆῇἤἢἥἣὴήἠἡἦἧὠὡὢὣὤὥὦὧᾠᾡᾢᾣᾤᾥᾦᾧῲῳῴῶῷώὼ]")
	TheCorpora    = []string{GREEKCORP, LATINCORP, INSCRIPTCORP, CHRISTINSC, PAPYRUSCORP}
	TheLanguages  = []string{"greek", "latin"}
	LaunchTime    = time.Now()
	ServableFonts = map[string]str.FontTempl{"Alegreya": AlegreyaFont, "Brill": BrillFont, "CMU": CMUFont, "Fira": FiraFont,
		"Gentium": GentiumFont, "Inter": InterFont, "Lato": LatoFont, "MPlus1": MPlusOneFont, "Noto": NotoFont, "Roboto": RobotoFont,
		"SourceSans": SourceSansFont, "Ubuntu": UbuntuFont} // cf rt-embhcss.go
)

//
// FONTS
//

// the fonts we know how to serve; need to look at https://github.com/e-gun/fontsubsetting for how to generate the ttf files

// note that a search for "Ͻ" or "⏙" is a good way to peek at support for obscure chars as others cluster near them
// even Noto and Brill will fail a lot inside of Aristides Quintilianus

// IMBPlex is missing GreekExtended?

// hipparchiasemicondenseditalicstatic, etc: setting font-style as oblique on a Regular src file will not give you oblique...

var (
	AlegreyaFont = str.FontTempl{
		Type:             "truetype",
		ShrtType:         "ttf",
		Bold:             "AlegreyaSans-BoldSubset.ttf",
		BoldItalic:       "AlegreyaSans-BoldItalicSubset.ttf",
		CondensedBold:    "AlegreyaSans-BoldSubset.ttf",
		CondensedItalic:  "AlegreyaSans-ItalicSubset.ttf",
		CondensedRegular: "AlegreyaSans-RegularSubset.ttf",
		SemiCondRegular:  "AlegreyaSans-RegularSubset.ttf",
		SemiCondItalic:   "AlegreyaSans-ItalicSubset.ttf",
		Italic:           "AlegreyaSans-ItalicSubset.ttf",
		Light:            "AlegreyaSans-LightSubset.ttf",
		Mono:             "NotoSansMono_Condensed-RegularSubset.ttf",
		Regular:          "AlegreyaSans-RegularSubset.ttf",
		SemiBold:         "AlegreyaSans-MediumSubset.ttf",
		Thin:             "AlegreyaSans-ThinSubset.ttf",
		HasLunateSigma:   false,
		NeedsManualStyle: []string{"hipparchiasemicondensedstatic", "hipparchiasemicondenseditalicstatic",
			"hipparchiacondensedboldstatic", "hipparchiacondenseditalicstatic", "hipparchiacondensedstatic"},
		SubFolder: "alegreya",
	}
	BrillFont = str.FontTempl{
		Type:             "truetype",
		ShrtType:         "ttf",
		Bold:             "Brill-BoldSubset.ttf",
		BoldItalic:       "Brill-BoldItalicSubset.ttf",
		CondensedBold:    "Brill-BoldSubset.ttf",
		CondensedItalic:  "Brill-ItalicSubset.ttf",
		CondensedRegular: "Brill-RomanSubset.ttf",
		SemiCondRegular:  "Brill-RomanSubset.ttf",
		SemiCondItalic:   "Brill-ItalicSubset.ttf",
		Italic:           "Brill-ItalicSubset.ttf",
		Light:            "Brill-RomanSubset.ttf",
		Mono:             "NotoSansMono_Condensed-RegularSubset.ttf",
		Regular:          "Brill-RomanSubset.ttf",
		SemiBold:         "Brill-BoldSubset.ttf",
		Thin:             "Brill-RomanSubset.ttf",
		HasLunateSigma:   true,
		NeedsManualStyle: []string{"hipparchialightstatic", "hipparchiasemicondensedstatic", "hipparchiasemicondenseditalicstatic",
			"hipparchiacondensedboldstatic", "hipparchiacondenseditalicstatic", "hipparchiathinstatic", "hipparchiacondensedstatic"},
		SubFolder: "brill",
	}
	CMUFont = str.FontTempl{
		Type:             "opentype",
		ShrtType:         "otf",
		Bold:             "cmunsxSubset.otf",
		BoldItalic:       "cmunsoSubset.otf",
		CondensedBold:    "cmunssdcSubset.otf",
		CondensedItalic:  "cmunsiSubset.otf",
		CondensedRegular: "cmunssdcSubset.otf",
		SemiCondRegular:  "cmunssdcSubset.otf", // semi dne
		SemiCondItalic:   "cmunsiSubset.otf",
		Italic:           "cmunsiSubset.otf",
		Light:            "cmunssSubset.otf",
		Mono:             "cmunbtlSubset.otf",
		Regular:          "cmunssSubset.otf",
		SemiBold:         "cmunsxSubset.otf",
		Thin:             "cmunsxSubset.otf",
		HasLunateSigma:   false,
		NeedsManualStyle: []string{"hipparchiathinstatic", "hipparchialightstatic", "hipparchiacondenseditalicstatic",
			"hipparchiacondensedboldstatic", "hipparchiasemicondensedstatic", "hipparchiasemicondenseditalicstatic"},
		SubFolder: "cmu",
	}
	FiraFont = str.FontTempl{
		Type:             "opentype",
		ShrtType:         "otf",
		Bold:             "FiraSans-BoldSubset.otf",
		BoldItalic:       "FiraSans-BoldItalicSubset.otf",
		CondensedBold:    "FiraSansCondensed-BoldSubset.otf",
		CondensedItalic:  "FiraSansCondensed-ItalicSubset.otf",
		CondensedRegular: "FiraSansCondensed-RegularSubset.otf",
		SemiCondRegular:  "FiraSansCondensed-RegularSubset.otf", // semi dne
		SemiCondItalic:   "FiraSansCondensed-ItalicSubset.otf",
		Italic:           "FiraSans-ItalicSubset.otf",
		Light:            "FiraSans-LightSubset.otf",
		Mono:             "FiraMono-RegularSubset.otf",
		Regular:          "FiraSans-RegularSubset.otf",
		SemiBold:         "FiraSans-SemiBoldSubset.otf",
		Thin:             "FiraSans-ThinSubset.otf",
		HasLunateSigma:   true,
		NeedsManualStyle: []string{},
		SubFolder:        "fira",
	}
	GentiumFont = str.FontTempl{
		Type:             "truetype",
		ShrtType:         "ttf",
		Bold:             "GentiumPlusCompact-BoldSubset.ttf",
		BoldItalic:       "GentiumPlusCompact-BoldItalicSubset.ttf",
		CondensedBold:    "GentiumPlusCompact-BoldSubset.ttf",
		CondensedItalic:  "GentiumPlusCompact-ItalicSubset.ttf",
		CondensedRegular: "GentiumPlusCompact-RegularSubset.ttf",
		SemiCondRegular:  "GentiumPlusCompact-RegularSubset.ttf",
		SemiCondItalic:   "GentiumPlusCompact-ItalicSubset.ttf",
		Italic:           "GentiumPlusCompact-ItalicSubset.ttf",
		Light:            "GentiumPlusCompact-RegularSubset.ttf",
		Mono:             "NotoSansMono_Condensed-RegularSubset.ttf",
		Regular:          "GentiumPlusCompact-RegularSubset.ttf",
		SemiBold:         "GentiumPlusCompact-BoldSubset.ttf",
		Thin:             "GentiumPlusCompact-RegularSubset.ttf",
		HasLunateSigma:   true,
		NeedsManualStyle: []string{"hipparchialightstatic", "hipparchiasemicondensedstatic",
			"hipparchiasemicondenseditalicstatic", "hipparchiacondensedboldstatic", "hipparchiacondenseditalicstatic",
			"hipparchiathinstatic", "hipparchiacondensedstatic"},
		SubFolder: "gentium",
	}
	NotoFont = str.FontTempl{
		Type:             "opentype",
		ShrtType:         "otf",
		Bold:             "NotoSans-BoldSubset.otf",
		BoldItalic:       "NotoSans-BoldItalicSubset.otf",
		CondensedBold:    "NotoSans-CondensedSemiBoldSubset.otf",
		CondensedItalic:  "NotoSans-CondensedItalicSubset.otf",
		CondensedRegular: "NotoSans-CondensedSubset.otf",
		SemiCondRegular:  "NotoSans-SemiCondensedSubset.otf",
		SemiCondItalic:   "NotoSans-SemiCondensedItalicSubset.otf",
		Italic:           "NotoSans-ItalicSubset.otf",
		Light:            "NotoSans-LightSubset.otf",
		Mono:             "NotoSansMono-SemiCondensedSubset.otf",
		Regular:          "NotoSans-RegularSubset.otf",
		SemiBold:         "NotoSans-SemiBoldSubset.otf",
		Thin:             "NotoSans-ThinSubset.otf",
		HasLunateSigma:   true,
		NeedsManualStyle: []string{},
		SubFolder:        "noto",
	}
	InterFont = str.FontTempl{
		Type:             "opentype",
		ShrtType:         "otf",
		Bold:             "Inter-BoldSubset.otf",
		BoldItalic:       "Inter-BoldItalicSubset.otf",
		CondensedBold:    "InterTight-BoldSubset.otf",
		CondensedItalic:  "Inter-ItalicSubset.otf",
		CondensedRegular: "Inter-RegularSubset.otf",
		SemiCondRegular:  "Inter-RegularSubset.otf",
		SemiCondItalic:   "Inter-ItalicSubset.otf",
		Italic:           "Inter-ItalicSubset.otf",
		Light:            "Inter-LightSubset.otf",
		Mono:             "NotoSansMono-SemiCondensedSubset.otf",
		Regular:          "Inter-RegularSubset.otf",
		SemiBold:         "Inter-SemiBoldSubset.otf",
		Thin:             "Inter-ThinSubset.otf",
		HasLunateSigma:   true,
		NeedsManualStyle: []string{"hipparchiasemicondensedstatic", "hipparchiasemicondenseditalicstatic",
			"hipparchiacondensedboldstatic", "hipparchiacondenseditalicstatic", "hipparchiacondensedstatic"},
		SubFolder: "inter",
	}
	LatoFont = str.FontTempl{
		Type:             "truetype",
		ShrtType:         "ttf",
		Bold:             "Lato-BoldSubset.ttf",
		BoldItalic:       "Lato-BoldItalicSubset.ttf",
		CondensedBold:    "Lato-BoldSubset.ttf",
		CondensedItalic:  "Lato-ItalicSubset.ttf",
		CondensedRegular: "Lato-RegularSubset.ttf",
		SemiCondRegular:  "Lato-RegularSubset.ttf",
		SemiCondItalic:   "Lato-ItalicSubset.ttf",
		Italic:           "Lato-ItalicSubset.ttf",
		Light:            "Lato-LightSubset.ttf",
		Mono:             "Iosevka-RegularSubset.ttf",
		Regular:          "Lato-RegularSubset.ttf",
		SemiBold:         "Lato-MediumSubset.ttf", // Inter_18pt-MediumSubset.ttf is too light?
		Thin:             "Lato-ThinSubset.ttf",
		HasLunateSigma:   true,
		NeedsManualStyle: []string{"hipparchiasemicondensedstatic", "hipparchiasemicondenseditalicstatic",
			"hipparchiacondensedboldstatic", "hipparchiacondenseditalicstatic"},
		SubFolder: "lato",
	}
	// MPlusOneFont - has not Italic ...
	MPlusOneFont = str.FontTempl{
		Type:             "opentype",
		ShrtType:         "otf",
		Bold:             "Mplus1-BoldSubset.otf",
		BoldItalic:       "Mplus1-BoldSubset.otf",
		CondensedBold:    "Mplus1-RegularSubset.otf",
		CondensedItalic:  "Mplus1-RegularSubset.otf",
		CondensedRegular: "Mplus1-RegularSubset.otf",
		SemiCondRegular:  "Mplus1-RegularSubset.otf",
		SemiCondItalic:   "Mplus1-RegularSubset.otf",
		Italic:           "Mplus1-RegularSubset.otf",
		Light:            "Mplus1-LightSubset.otf",
		Mono:             "Mplus1Code-RegulaSubset.otf",
		Regular:          "Mplus1-RegularSubset.otf",
		SemiBold:         "Mplus1-SemiBoldSubset.otf",
		Thin:             "Mplus1-ThinSubset.otf",
		HasLunateSigma:   true,
		NeedsManualStyle: []string{"hipparchiasemicondensedstatic", "hipparchiasemicondenseditalicstatic",
			"hipparchiacondensedboldstatic", "hipparchiacondenseditalicstatic", "hipparchiacondensedstatic",
			"hipparchiabolditalicstatic", "hipparchiaobliquestatic"},
		SubFolder: "mplus1",
	}
	RobotoFont = str.FontTempl{
		Type:             "truetype",
		ShrtType:         "ttf",
		Bold:             "Roboto-BoldSubset.ttf",
		BoldItalic:       "Roboto-BoldItalicSubset.ttf",
		CondensedBold:    "Roboto_Condensed-BoldSubset.ttf",
		CondensedItalic:  "Roboto_Condensed-ItalicSubset.ttf",
		CondensedRegular: "Roboto_Condensed-RegularSubset.ttf",
		SemiCondRegular:  "Roboto_SemiCondensed-RegularSubset.ttf", // semi dne
		SemiCondItalic:   "Roboto_SemiCondensed-ItalicSubset.ttf",
		Italic:           "Roboto-ItalicSubset.ttf",
		Light:            "Roboto-LightSubset.ttf",
		Mono:             "RobotoMono-RegularSubset.ttf",
		Regular:          "Roboto-RegularSubset.ttf",
		SemiBold:         "Roboto-SemiBoldSubset.ttf",
		Thin:             "Roboto-ThinSubset.ttf",
		HasLunateSigma:   true,
		NeedsManualStyle: []string{},
		SubFolder:        "roboto",
	}
	SourceSansFont = str.FontTempl{
		Type:             "truetype",
		ShrtType:         "ttf",
		Bold:             "SourceSans3-BoldSubset.ttf",
		BoldItalic:       "SourceSans3-BoldItalicSubset.ttf",
		CondensedBold:    "SourceSans3-RegularSubset.ttf",
		CondensedItalic:  "SourceSans3-ItalicSubset.ttf",
		CondensedRegular: "SourceSans3-RegularSubset.ttf",
		SemiCondRegular:  "SourceSans3-RegularSubset.ttf",
		SemiCondItalic:   "SourceSans3-ItalicSubset.ttf",
		Italic:           "SourceSans3-ItalicSubset.ttf",
		Light:            "SourceSans3-LightSubset.ttf",
		Mono:             "SourceCodePro-VariableFont_wghtSubset.ttf",
		Regular:          "SourceSans3-RegularSubset.ttf",
		SemiBold:         "SourceSans3-BoldSubset.ttf",
		Thin:             "SourceSans3-LightSubset.ttf",
		HasLunateSigma:   false,
		NeedsManualStyle: []string{"hipparchiasemicondensedstatic", "hipparchiasemicondenseditalicstatic",
			"hipparchiacondensedboldstatic", "hipparchiacondenseditalicstatic"},
		SubFolder: "source",
	}
	UbuntuFont = str.FontTempl{
		Type:             "truetype",
		ShrtType:         "ttf",
		Bold:             "UbuntuSans-BoldSubset.ttf",
		BoldItalic:       "UbuntuSans-BoldItalicSubset.ttf",
		CondensedBold:    "UbuntuSans_Condensed-BoldSubset.ttf",
		CondensedItalic:  "UbuntuSans_Condensed-ItalicSubset.ttf",
		CondensedRegular: "UbuntuSans_Condensed-RegularSubset.ttf",
		SemiCondRegular:  "UbuntuSans_SemiCondensed-RegularSubset.ttf", // semi dne
		SemiCondItalic:   "UbuntuSans_SemiCondensed-ItalicSubset.ttf",
		Italic:           "UbuntuSans-ItalicSubset.ttf",
		Light:            "UbuntuSans-LightSubset.ttf",
		Mono:             "UbuntuMono-RegularSubset.ttf",
		Regular:          "UbuntuSans-RegularSubset.ttf",
		SemiBold:         "UbuntuSans-MediumSubset.ttf",
		Thin:             "UbuntuSans-LightSubset.ttf",
		HasLunateSigma:   false,
		NeedsManualStyle: []string{},
		SubFolder:        "ubuntu",
	}

	// inactive / unused (but github.com/e-gun/fontsubsetting probably still knows about them)

	FiraFontTTF = str.FontTempl{
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
		SubFolder:        "fira",
	}
	NotoDispFont = str.FontTempl{
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
		SubFolder:        "notodisplay",
	}
	NotoFontTTF = str.FontTempl{
		Type:             "truetype",
		ShrtType:         "ttf",
		Bold:             "NotoSans-BoldSubset.ttf",
		BoldItalic:       "NotoSans-BoldItalicSubset.ttf",
		CondensedBold:    "NotoSans-CondensedSemiBoldSubset.ttf",
		CondensedItalic:  "NotoSans-CondensedItalicSubset.ttf",
		CondensedRegular: "NotoSans-CondensedSubset.ttf",
		SemiCondRegular:  "NotoSans-SemiCondensedSubset.ttf",
		SemiCondItalic:   "NotoSans-SemiCondensedItalicSubset.ttf",
		Italic:           "NotoSans-ItalicSubset.ttf",
		Light:            "NotoSans-LightSubset.ttf",
		Mono:             "NotoSansMono-SemiCondensedSubset.ttf",
		Regular:          "NotoSans-RegularSubset.ttf",
		SemiBold:         "NotoSans-SemiBoldSubset.ttf",
		Thin:             "NotoSans-ThinSubset.ttf",
		HasLunateSigma:   true,
		NeedsManualStyle: []string{},
		SubFolder:        "noto",
	}
	InterFontTTF = str.FontTempl{
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
		SemiBold:         "InterTight-BoldSubset.ttf", // Inter_18pt-MediumSubset.ttf is too light?
		Thin:             "Inter_18pt-ThinSubset.ttf",
		HasLunateSigma:   true,
		NeedsManualStyle: []string{},
		SubFolder:        "inter",
	}
)
