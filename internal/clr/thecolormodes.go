package clr

var (
	// CssColorHSLs -
	// need access to the numbers and not just CSS so you can reset vector dot colors too...
	// this also explains why Light and Dark have values assigned to them
	CssColorHSLs = map[string][]int{
		"Light":     {236, 0, 50, 0},
		"Dark":      {50, 0, 80, 0},
		"MonoSand":  {60, 12, 85, 8},
		"MonoAsh":   {220, 15, 90, 15},
		"Splitcomp": {200, 12, 85, 10},
		"Square":    {325, 10, 20, 90},
		"Tetradic":  {40, 15, 85, 20},
		"Triadic":   {240, 12, 20, 90},
	}
	CssColorModes = map[string]string{
		"Light":     LIGHTCOLORSORIGINAL,
		"Dark":      DARKCOLORSMANUAL,
		"MonoSand":  GenerateMonoScheme(CssColorHSLs["MonoSand"]),
		"MonoAsh":   GenerateMonoScheme(CssColorHSLs["MonoAsh"]),
		"Splitcomp": GenerateSplitcompScheme(CssColorHSLs["Splitcomp"]),
		"Square":    GenerateTetradScheme(CssColorHSLs["Square"]),
		"Tetradic":  GenerateTetradScheme(CssColorHSLs["Tetradic"]),
		"Triadic":   GenerateTriadScheme(CssColorHSLs["Triadic"]),
	}
)
