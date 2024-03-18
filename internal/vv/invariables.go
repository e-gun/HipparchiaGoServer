package vv

import (
	"github.com/e-gun/HipparchiaGoServer/internal/structs"
	"regexp"
	"time"
)

var (
	Config structs.CurrentConfiguration
	// SQLPool       *pgxpool.Pool
	UserPassPairs = make(map[string]string)
	//AllWorks      = make(map[string]*structs.DbWork)
	//AllAuthors    = make(map[string]*structs.DbAuthor) // populated by authormap.go
	//AllLemm       = make(map[string]*structs.DbLemma)
	//NestedLemm    = make(map[string]map[string]*structs.DbLemma)
	//WkCorpusMap   = make(map[string][]string)
	//AuCorpusMap   = make(map[string][]string)
	// LoadedCorp = make(map[string]bool)
	//AuGenres      = make(map[string]bool)
	//WkGenres      = make(map[string]bool)
	//AuLocs        = make(map[string]bool)
	//WkLocs        = make(map[string]bool)
	TheCorpora    = []string{GREEKCORP, LATINCORP, INSCRIPTCORP, CHRISTINSC, PAPYRUSCORP}
	TheLanguages  = []string{"greek", "latin"}
	ServableFonts = map[string]structs.FontTempl{"Noto": NotoFont, "Roboto": RobotoFont, "Fira": FiraFont} // cf rt-embhcss.go
	LaunchTime    = time.Now()
)

var (
	HasAccent = regexp.MustCompile("[äëïöüâêîôûàèìòùáéíóúᾂᾒᾢᾃᾓᾣᾄᾔᾤᾅᾕᾥᾆᾖᾦᾇᾗᾧἂἒἲὂὒἢὢἃἓἳὃὓἣὣἄἔἴὄὔἤὤἅἕἵὅὕἥὥἆἶὖἦὦἇἷὗἧὧᾲῂῲᾴῄῴᾷῇῷᾀᾐᾠᾁᾑᾡῒῢΐΰῧἀἐἰὀὐἠὠῤἁἑἱὁὑἡὡῥὰὲὶὸὺὴὼάέίόύήώᾶῖῦῆῶϊϋ]")
	IsGreek   = regexp.MustCompile("[α-ωϲῥἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάἐἑἒἓἔἕὲέἰἱἲἳἴἵἶἷὶίῐῑῒΐῖῗὀὁὂὃὄὅόὸὐὑὒὓὔὕὖὗϋῠῡῢΰῦῧύὺᾐᾑᾒᾓᾔᾕᾖᾗῂῃῄῆῇἤἢἥἣὴήἠἡἦἧὠὡὢὣὤὥὦὧᾠᾡᾢᾣᾤᾥᾦᾧῲῳῴῶῷώὼ]")
)

//
// FONTS
//

// the fonts we know how to serve
// NB: Inter, SourceSans and Ubuntu have been toyed with: Inter lacks both condensed and semi-condensed

var (
	NotoFont = structs.FontTempl{
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
	FiraFont = structs.FontTempl{
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
	RobotoFont = structs.FontTempl{
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
