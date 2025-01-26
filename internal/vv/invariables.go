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
	ServableFonts = map[string]str.FontTempl{"Alegreya": AlegreyaFont, "Brill": BrillFont, "Fira": FiraFont,
		"Gentium": GentiumFont, "Inter": InterFont, "Lato": LatoFont, "Noto": NotoFont, "Roboto": RobotoFont,
		"SourceSans": SourceSansFont, "Ubuntu": UbuntuFont} // cf rt-embhcss.go
	LaunchTime = time.Now()
)

var (
	IsGreek = regexp.MustCompile("[α-ωϲῥἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάἐἑἒἓἔἕὲέἰἱἲἳἴἵἶἷὶίῐῑῒΐῖῗὀὁὂὃὄὅόὸὐὑὒὓὔὕὖὗϋῠῡῢΰῦῧύὺᾐᾑᾒᾓᾔᾕᾖᾗῂῃῄῆῇἤἢἥἣὴήἠἡἦἧὠὡὢὣὤὥὦὧᾠᾡᾢᾣᾤᾥᾦᾧῲῳῴῶῷώὼ]")
)

//
// FONTS
//

// the fonts we know how to serve; need to run "pyftsubset.sh" to generate these

// note that a search for "Ͻ" or "⏙" is a good way to peek at support for obscure chars
// even Noto and Brill will fail a lot inside of Aristides Quintilianus

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
		SemiCondItalic:   "AlegreyaSans-RegularSubset.ttf",
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
		SemiCondItalic:   "Brill-RomanSubset.ttf",
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
		SubFolder:        "fira",
	}
	GentiumFont = str.FontTempl{
		Type:             "truetype",
		ShrtType:         "ttf",
		Bold:             "GentiumPlus-BoldSubset.ttf",
		BoldItalic:       "GentiumPlus-BoldItalicSubset.ttf",
		CondensedBold:    "GentiumPlus-BoldSubset.ttf",
		CondensedItalic:  "GentiumPlus-ItalicSubset.ttf",
		CondensedRegular: "GentiumPlus-RegularSubset.ttf",
		SemiCondRegular:  "GentiumPlus-RegularSubset.ttf",
		SemiCondItalic:   "GentiumPlus-RegularSubset.ttf",
		Italic:           "GentiumPlus-ItalicSubset.ttf",
		Light:            "GentiumPlus-RegularSubset.ttf",
		Mono:             "NotoSansMono_Condensed-RegularSubset.ttf",
		Regular:          "GentiumPlus-RegularSubset.ttf",
		SemiBold:         "GentiumPlus-BoldSubset.ttf",
		Thin:             "GentiumPlus-RegularSubset.ttf",
		HasLunateSigma:   true,
		NeedsManualStyle: []string{"hipparchialightstatic", "hipparchiasemicondensedstatic", "hipparchiasemicondenseditalicstatic",
			"hipparchiacondensedboldstatic", "hipparchiacondenseditalicstatic", "hipparchiathinstatic", "hipparchiacondensedstatic"},
		SubFolder: "gentium",
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
		SubFolder:        "noto",
	}
	PlexFont = str.FontTempl{
		Type:             "truetype",
		ShrtType:         "ttf",
		Bold:             "IBMPlexSans-BoldSubset.ttf",
		BoldItalic:       "IBMPlexSans-BoldItalicSubset.ttf",
		CondensedBold:    "IBMPlexSansCondensed-SemiBoldSubset.ttf",
		CondensedItalic:  "IBMPlexSansCondensed-ItalicSubset.ttf",
		CondensedRegular: "IBMPlexSansCondensed-RegularSubset.ttf",
		SemiCondRegular:  "IBMPlexSansCondensed-LightSubset.ttf",
		SemiCondItalic:   "IBMPlexSansCondensed-LightItalicSubset.ttf",
		Italic:           "IBMPlexSans-ItalicSubset.ttf",
		Light:            "IBMPlexSans-ExtraLightSubset.ttf",
		Mono:             "IBMPlexMono-LightSubset.ttf",
		Regular:          "IBMPlexSans-RegularSubset.ttf",
		SemiBold:         "IBMPlexSans-SemiBoldSubset.ttf",
		Thin:             "IBMPlexSans-ThinSubset.ttf",
		HasLunateSigma:   true,
		NeedsManualStyle: []string{},
		SubFolder:        "plex",
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
		SemiBold:         "InterTight-BoldSubset.ttf", // Inter_18pt-MediumSubset.ttf is too light?
		Thin:             "Inter_18pt-ThinSubset.ttf",
		HasLunateSigma:   true,
		NeedsManualStyle: []string{},
		SubFolder:        "inter",
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
		Mono:             "iosevka-regularSubset.woff2",
		Regular:          "Lato-RegularSubset.ttf",
		SemiBold:         "Lato-MediumSubset.ttf", // Inter_18pt-MediumSubset.ttf is too light?
		Thin:             "Lato-ThinSubset.ttf",
		HasLunateSigma:   true,
		NeedsManualStyle: []string{"hipparchiasemicondensedstatic", "hipparchiasemicondenseditalicstatic",
			"hipparchiacondensedboldstatic", "hipparchiacondenseditalicstatic"},
		SubFolder: "lato",
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
		SubFolder:        "roboto",
	}
	SourceSansFont = str.FontTempl{
		Type:             "truetype",
		ShrtType:         "ttf",
		Bold:             "SourceSans3-VariableFont_wghtSubset.ttf",
		BoldItalic:       "SourceSans3-Italic-VariableFont_wghtSubset.ttf",
		CondensedBold:    "SourceSans3-VariableFont_wghtSubset.ttf",
		CondensedItalic:  "SourceSans3-Italic-VariableFont_wghtSubset.ttf",
		CondensedRegular: "SourceSans3-VariableFont_wghtSubset.ttf",
		SemiCondRegular:  "SourceSans3-VariableFont_wghtSubset.ttf",
		SemiCondItalic:   "SourceSans3-Italic-VariableFont_wghtSubset.ttf",
		Italic:           "SourceSans3-Italic-VariableFont_wghtSubset.ttf",
		Light:            "SourceSans3-VariableFont_wghtSubset.ttf",
		Mono:             "SourceCodePro-VariableFont_wghtSubset.ttf",
		Regular:          "SourceSans3-VariableFont_wghtSubset.ttf",
		SemiBold:         "SourceSans3-VariableFont_wghtSubset.ttf",
		Thin:             "SourceSans3-VariableFont_wghtSubset.ttf",
		HasLunateSigma:   false,
		NeedsManualStyle: []string{"hipparchialightstatic", "hipparchiasemicondensedstatic", "hipparchiasemicondenseditalicstatic",
			"hipparchiacondensedboldstatic", "hipparchiacondenseditalicstatic", "hipparchiathinstatic", "hipparchiaboldstatic",
			"hipparchiasemiboldstatic"},
		SubFolder: "source",
	}
	UbuntuFont = str.FontTempl{
		Type:             "truetype",
		ShrtType:         "ttf",
		Bold:             "Ubuntu-BoldSubset.ttf",
		BoldItalic:       "Ubuntu-BoldItalicSubset.ttf",
		CondensedBold:    "Ubuntu-RegularSubset.ttf",
		CondensedItalic:  "Ubuntu-RegularSubset.ttf",
		CondensedRegular: "Ubuntu-RegularSubset.ttf",
		SemiCondRegular:  "Ubuntu-RegularSubset.ttf", // semi dne
		SemiCondItalic:   "Ubuntu-RegularSubset.ttf",
		Italic:           "Ubuntu-ItalicSubset.ttf",
		Light:            "Ubuntu-LightSubset.ttf",
		Mono:             "UbuntuMono-RegularSubset.ttf",
		Regular:          "Ubuntu-RegularSubset.ttf",
		SemiBold:         "Ubuntu-MediumSubset.ttf",
		Thin:             "Ubuntu-LightSubset.ttf",
		HasLunateSigma:   false,
		NeedsManualStyle: []string{"hipparchiasemicondensedstatic", "hipparchiasemicondenseditalicstatic",
			"hipparchiacondensedboldstatic", "hipparchiacondenseditalicstatic"},
		SubFolder: "ubuntu",
	}
)
