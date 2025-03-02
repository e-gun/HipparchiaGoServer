package clr

var (
	// CssColorHSLs -
	// need access to the numbers and not just CSS so you can reset vector dot colors too...
	// this also explains why Light and Dark have values assigned to them
	CssColorHSLs = map[string][]int{
		"Light": {236, 0, 50, 0},
		"Dark":  {50, 0, 80, 0},
		// "MonoSand":  {60, 15, 8, 85},
		"Monochrome": {220, 15, 95, 0},
		"SplitComp":  {200, 12, 85, 10},
		"Square":     {90, 25, 85, 25},
		"Tetradic":   {20, 17, 88, 22},
		"Triadic":    {230, 15, 20, 90},
	}
	CssColorModes = map[string]string{
		"Light":      LIGHTCOLORSORIGINAL,
		"Dark":       DARKCOLORSMANUAL,
		"Monochrome": generatemonoscheme(CssColorHSLs["Monochrome"]),
		"SplitComp":  generatesplitcompscheme(CssColorHSLs["SplitComp"]),
		"Square":     generatesquarescheme(CssColorHSLs["Square"]),
		"Tetradic":   generatetetradscheme(CssColorHSLs["Tetradic"]),
		"Triadic":    generatetriatscheme(CssColorHSLs["Triadic"]),
	}
	CssSamples = map[string]string{
		"Monochrome": colorsamplesheet("Monochrome", CssColorHSLs["Monochrome"]),
		"SplitComp":  colorsamplesheet("SplitComp", CssColorHSLs["SplitComp"]),
		"Square":     colorsamplesheet("Square", CssColorHSLs["Square"]),
		"Tetradic":   colorsamplesheet("Tetradic", CssColorHSLs["Tetradic"]),
		"Triadic":    colorsamplesheet("Triadic", CssColorHSLs["Triadic"]),
	}
)
