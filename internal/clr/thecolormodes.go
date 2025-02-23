package clr

var (
	CssColorModes = map[string]string{
		"Light":     LIGHTCOLORSORIGINAL,
		"Dark":      DARKCOLORSMANUAL,
		"MonoSand":  GenerateMonoScheme(60, 12, 85, 8),
		"MonoAsh":   GenerateMonoScheme(220, 15, 85, 10),
		"Splitcomp": GenerateSplitcompScheme(100, 15, 25, 85),
		"Square":    GenerateTetradScheme(300, 20, 85, 15),
		"Tetradic":  GenerateTetradScheme(20, 15, 15, 85),
		"Triadic":   GenerateTriadScheme(180, 20, 85, 20),
		// "Splitcomp": SPLITCOMPMANUAL,
		// "MonoSand":  MONOCHROMESANDYMANUAL,
		// "MonoAsh":  MONOCHROMEASHMANUAL,
		// "Tetradic": TETRADICMANUAL,
		// "Triadic": TRIADICMANUAL,
	}
)
