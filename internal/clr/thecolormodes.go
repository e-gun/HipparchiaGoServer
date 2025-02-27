package clr

var (
	// CssColorHSLs -
	// need access to the numbers and not just CSS so you can reset vector dot colors too...
	// this also explains why Light and Dark have values assigned to them
	CssColorHSLs = map[string][]int{
		"Light": {236, 0, 50, 0},
		"Dark":  {50, 0, 80, 0},
		// "MonoSand":  {60, 15, 8, 85},
		"Monochrome": {220, 15, 100, 0},
		"Splitcomp":  {200, 12, 85, 10},
		"Square":     {45, 20, 25, 85},
		"Tetradic":   {20, 17, 88, 22},
		"Triadic":    {240, 12, 20, 90},
	}
	CssColorModes = map[string]string{
		"Light":      LIGHTCOLORSORIGINAL,
		"Dark":       DARKCOLORSMANUAL,
		"Monochrome": GenerateMonoScheme(CssColorHSLs["Monochrome"]),
		"Splitcomp":  GenerateSplitcompScheme(CssColorHSLs["Splitcomp"]),
		"Square":     GenerateSquareScheme(CssColorHSLs["Square"]),
		"Tetradic":   GenerateTetradScheme(CssColorHSLs["Tetradic"]),
		"Triadic":    GenerateTriadScheme(CssColorHSLs["Triadic"]),
	}
	CssSamples = map[string]string{
		"Monochrome": colorsamplesheet("mono", CssColorHSLs["Monochrome"]),
		"Splitcomp":  colorsamplesheet("splitcomp", CssColorHSLs["Splitcomp"]),
		"Square":     colorsamplesheet("square", CssColorHSLs["Square"]),
		"Tetradic":   colorsamplesheet("tetrad", CssColorHSLs["Tetradic"]),
		"Triadic":    colorsamplesheet("triad", CssColorHSLs["Triadic"]),
	}
)
