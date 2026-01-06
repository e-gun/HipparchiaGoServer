//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-26
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package clr

var (
	// CssColorHSLs -
	// need access to the numbers and not just CSS so you can reset vector dot colors too...
	// this also explains why Light and Dark have values assigned to them
	// command line "-cd" can alter one of the k/v pairs; this will then affect CssColorModes and
	// CssSamples
	CssColorHSLs = map[string][]int{
		"Light": {140, 25, 100, 0},
		"Dark":  {50, 40, 0, 100},
		// "MonochromeDK": {0, 55, 8, 95}, // Blood Red; 0 0 0 100 is White on Black...
		"Monochrome": {50, 40, 95, 8}, // Sandy
		"SplitComp":  {175, 12, 85, 10},
		"Square":     {111, 25, 85, 15},
		"Tetradic":   {40, 17, 80, 15},
		"Triadic":    {55, 40, 15, 100},
	}
	CssColorModes = map[string]string{} // set in "main.go" after "configatlaunch.go"
	CssSamples    = map[string]string{} // set in "main.go" after "configatlaunch.go"
)

func GenerateCSSSamples() map[string]string {
	return map[string]string{
		"Monochrome": colorsamplesheet("Monochrome", CssColorHSLs["Monochrome"]),
		"SplitComp":  colorsamplesheet("SplitComp", CssColorHSLs["SplitComp"]),
		"Square":     colorsamplesheet("Square", CssColorHSLs["Square"]),
		"Tetradic":   colorsamplesheet("Tetradic", CssColorHSLs["Tetradic"]),
		"Triadic":    colorsamplesheet("Triadic", CssColorHSLs["Triadic"]),
	}
}

func GenerateColorModes() map[string]string {
	return map[string]string{
		"Light":      LIGHTCOLORSORIGINAL,
		"Dark":       DARKCOLORSMANUAL,
		"Monochrome": generatemonoscheme(CssColorHSLs["Monochrome"]),
		"SplitComp":  generatesplitcompscheme(CssColorHSLs["SplitComp"]),
		"Square":     generatesquarescheme(CssColorHSLs["Square"]),
		"Tetradic":   generatetetradscheme(CssColorHSLs["Tetradic"]),
		"Triadic":    generatetriatscheme(CssColorHSLs["Triadic"]),
	}
}
