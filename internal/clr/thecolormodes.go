package clr

// need access to these numbers so you can reset vector dot colors too...
// this also explains why Light and Dark have values assigned to them

var (
	CssColorHSLs = map[string][]int{
		"Light":     {236, 0, 50, 0},
		"Dark":      {50, 0, 80, 0},
		"MonoSand":  {60, 12, 85, 8},
		"MonoAsh":   {220, 15, 85, 10},
		"Splitcomp": {220, 15, 85, 10},
		"Square":    {300, 20, 85, 15},
		"Tetradic":  {20, 15, 15, 85},
		"Triadic":   {180, 20, 85, 20},
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
		// "Splitcomp": SPLITCOMPMANUAL,
		// "MonoSand":  MONOCHROMESANDYMANUAL,
		// "MonoAsh":  MONOCHROMEASHMANUAL,
		// "Tetradic": TETRADICMANUAL,
		// "Triadic": TRIADICMANUAL,
	}
)
