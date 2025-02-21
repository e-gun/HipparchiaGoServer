package clr

var (
	CssColorModes = map[string]string{
		"Light":    LIGHTCOLORSORIGINAL,
		"Dark":     DARKCOLORSMANUAL,
		"MonoSand": GenerateMonoScheme(60, 12, 83, 5),
		"MonoAsh":  GenerateMonoScheme(220, 15, 80, 10),
		//"Tetradic": clr.GenerateTetradScheme(250, 15, 70, 15),
		//"Tridaic":  clr.GenerateTriadScheme(10, 25, 75, 15),
		"Splitcomp": SPLITCOMPMANUAL,
		// "MonoSand":  MONOCHROMESANDYMANUAL,
		// "MonoAsh":  MONOCHROMEASHMANUAL,
		"Tetradic": TETRADICMANUAL,
		"Tridaic":  TRIADICMANUAL,
	}
)
