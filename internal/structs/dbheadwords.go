package structs

import "reflect"

var (
	CORPUSWEIGTING = map[string]float32{"Ⓖ": 1.0, "Ⓛ": 12.7, "Ⓘ": 15.19, "Ⓓ": 18.14, "Ⓒ": 85.78}
	ERAWEIGHTING   = map[string]float32{"ℯ": 6.93, "𝓂": 1.87, "ℓ": 1}
	GKGENREWEIGHT  = map[string]float32{"Acta": 85.38, "Alchem": 72.13, "Anthol": 17.68, "Apocal": 117.69, "Apocr": 89.77,
		"Apol": 7.0, "Astrol": 20.68, "Astron": 44.72, "Biogr": 6.39, "Bucol": 416.66, "Caten": 5.21,
		"Chron": 4.55, "Comic": 29.61, "Comm": 1.0, "Concil": 16.75, "Coq": 532.74, "Dial": 7.1,
		"Docu": 2.66, "Doxogr": 130.84, "Eccl": 7.57, "Eleg": 188.08, "Encom": 13.17, "Epic": 19.36,
		"Epigr": 10.87, "Epist": 4.7, "Evang": 118.66, "Exeg": 1.24, "Fab": 140.87,
		"Geog": 10.74, "Gnom": 88.54, "Gram": 8.65, "Hagiog": 22.83, "Hexam": 110.78,
		"Hist": 1.44, "Homil": 6.87, "Hymn": 48.18, "Hypoth": 12.95, "Iamb": 122.22,
		"Ignot": 122914.2, "Invect": 238.54, "Inscr": 1.91, "Juris": 51.42, "Lexic": 4.14,
		"Litur": 531.5, "Lyr": 213.43, "Magica": 85.38, "Math": 9.91, "Mech": 103.44, "Med": 2.25,
		"Metro": 276.78, "Mim": 2183.94, "Mus": 96.32, "Myth": 201.78, "NarrFic": 14.62,
		"NatHis": 9.67, "Onir": 145.15, "Orac": 240.47, "Orat": 6.67, "Paradox": 267.32,
		"Parod": 831.51, "Paroem": 65.58, "Perig": 220.38, "Phil": 3.69, "Physiog": 628.77,
		"Poem": 62.82, "Polyhist": 24.91, "Proph": 95.51, "Pseud": 611.65, "Rhet": 8.67,
		"Satura": 291.58, "Satyr": 96.78, "Schol": 5.56, "Tact": 52.01, "Test": 66.53, "Theol": 6.28,
		"Trag": 35.8, "AllRelig": 0.58, "AllRhet": 2.9}
	LATGENREWEIGHT = map[string]float32{"Agric": 5.27, "Astron": 17.15, "Biogr": 9.88, "Bucol": 40.39, "Bomic": 4.21, "Comm": 2.25,
		"Coq": 60.0, "Dial": 1134.73, "Docu": 6.19, "Eleg": 8.35, "Encom": 404.6, "Epic": 2.37,
		"Epigr": 669.3, "Epist": 2.06, "Fab": 25.4, "Gnom": 147.23, "Gramm": 5.74, "Hexam": 20.06,
		"Hist": 1.0, "Hypoth": 762.59, "Ignotum": 586.58, "Inscr": 1.29, "Juris": 1.11,
		"Lexic": 27.71, "Lyr": 24.76, "Med": 7.26, "Mim": 1045.69, "NarrFic": 11.7,
		"Nathist": 1.94, "Orat": 1.81, "Parod": 339.23, "Phil": 2.3, "Poem": 14.34,
		"Polyhist": 4.75, "Rhet": 2.71, "Satura": 23.0, "Tact": 37.6, "Trag": 13.29, "Allrelig": 0,
		"Allrhet": 1.08}
)

type DbHeadwordTimeCounts struct {
	Early  int
	Middle int
	Late   int
}

type DbHeadwordCorpusCounts struct {
	TGrk int
	TLat int
	TDP  int
	TIN  int
	TCh  int
}

type DbHeadwordGenreCounts struct {
	Agric    int
	Alchem   int
	Anthol   int
	Astrol   int
	Astron   int
	Biogr    int
	Bucol    int
	Chron    int
	Comic    int
	Comm     int
	Concil   int
	Coq      int
	Dial     int
	Docu     int
	Doxog    int
	Eleg     int
	Epic     int
	Epigr    int
	Epist    int
	Exeg     int
	Fab      int
	Geog     int
	Gnom     int
	Gram     int
	Hexam    int
	Hist     int
	Hymn     int
	Hypoth   int
	Iamb     int
	Ignot    int
	Inscr    int
	Juris    int
	Lexic    int
	Lyr      int
	Magica   int
	Math     int
	Mech     int
	Med      int
	Meteor   int
	Mim      int
	Mus      int
	Myth     int
	NarrFic  int
	NatHis   int
	Onir     int
	Orac     int
	Paradox  int
	Parod    int
	Paroem   int
	Perig    int
	Phil     int
	Physiog  int
	Poem     int
	Polyhist int
	Pseud    int
	Satura   int
	Satyr    int
	Schol    int
	Tact     int
	Test     int
	Trag     int
	AllRhet  int
	AllRelig int
}

type DbHeadwordRhetoricaCounts struct {
	Encom  int
	Invect int
	Orat   int
	Rhet   int
}

type DbHeadwordTheologyCounts struct {
	Acta   int
	Apocal int
	Apocr  int
	Apol   int
	Caten  int
	Eccl   int
	Evang  int
	Hagiog int
	Homil  int
	Litur  int
	Proph  int
	Theol  int
}

// HWData - to help sort values inside DbHeadwordCount
type HWData struct {
	Name  string
	Count int
}

type DbHeadwordCount struct {
	Entry     string
	Total     int
	FrqCla    string
	Chron     DbHeadwordTimeCounts
	Genre     DbHeadwordGenreCounts
	Corpus    DbHeadwordCorpusCounts
	Rhetorica DbHeadwordRhetoricaCounts
	Theology  DbHeadwordTheologyCounts
	CorpVal   []HWData
	TimeVal   []HWData
	TagVal    []HWData
	GenreVal  []HWData
}

func (hw *DbHeadwordCount) LoadCorpVals() {
	// Prevalence (all forms): Ⓖ 95,843 / Ⓛ 10 / Ⓘ 151 / Ⓓ 751 / Ⓒ 64 / Ⓣ 96,819
	var vv []HWData
	vv = append(vv, HWData{"Ⓖ", hw.Corpus.TGrk})
	vv = append(vv, HWData{"Ⓛ", hw.Corpus.TLat})
	vv = append(vv, HWData{"Ⓘ", hw.Corpus.TIN})
	vv = append(vv, HWData{"Ⓓ", hw.Corpus.TDP})
	vv = append(vv, HWData{"Ⓒ", hw.Corpus.TCh})
	hw.CorpVal = vv
}

func (hw *DbHeadwordCount) LoadTimeVals() {
	// Weighted chronological distribution: ℯ 100 / ℓ 84 / 𝓂 62
	var vv []HWData
	vv = append(vv, HWData{"ℯ", hw.Chron.Early})
	vv = append(vv, HWData{"ℓ", hw.Chron.Late})
	vv = append(vv, HWData{"𝓂", hw.Chron.Middle})
	hw.TimeVal = vv
}

func (hw *DbHeadwordCount) LoadGenreVals() {
	// Weighted genre distribution: Predominant genres: bucol (100), iamb (98), epic (95),...
	gvv := reflect.ValueOf(hw.Rhetorica)
	gvtype := gvv.Type()
	sum := 0
	for i := 0; i < gvv.NumField(); i++ {
		sum += gvv.Field(i).Interface().(int)
	}
	hw.Genre.AllRhet = sum

	gvv = reflect.ValueOf(hw.Theology)
	gvtype = gvv.Type()
	sum = 0
	for i := 0; i < gvv.NumField(); i++ {
		sum += gvv.Field(i).Interface().(int)
	}
	hw.Genre.AllRelig = sum

	gvv = reflect.ValueOf(hw.Genre)
	gvtype = gvv.Type()
	var vv []HWData
	for i := 0; i < gvv.NumField(); i++ {
		var v HWData
		v.Name = gvtype.Field(i).Name
		v.Count = gvv.Field(i).Interface().(int)
		vv = append(vv, v)
	}
	hw.GenreVal = vv
}
