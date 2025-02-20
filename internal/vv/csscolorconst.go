package vv

import "github.com/e-gun/HipparchiaGoServer/internal/base/gen"

var (
	CssColorModes = map[string]string{
		"Light":     LIGHTCOLORS,
		"Dark":      DARKCOLORS,
		"MonoSand":  gen.GenerateMonoScheme(60, 15, 80),
		"MonoAsh":   gen.GenerateMonoScheme(220, 15, 80),
		"Tetradic":  gen.GenerateTetradScheme(250, 15, 70),
		"Splitcomp": SPLITCOMP,
		"Tridaic":   gen.GenerateTriadScheme(10, 25, 75),
	}
)

const (
	DEFAULCOLORS = "Light"

	LIGHTCOLORS = `
	--main-body-color: hsla(0, 0%, 98%, 1);
	--main-font-color: hsla(0, 0%, 6%, 1);
	--input-border-color: hsla(0, 0%, 92%, 1);
	
	--buttoncolor: hsla(0, 0%, 93%, 1);
	--button-hover: hsla(0, 0%, 90%, 1);
	--fieldset-background: hsla(0, 0%, 98%, 1);
	--focus-shadow: rgba(0, 0, 0, 0.5);
	--icons-color: rgba(0, 0, 0, 0.54);
	
	--black: hsla(0, 0%, 0%, 1);
	--blue: hsla(240, 100%, 27%, 1);
	--brown: hsla(22, 22%, 26%, 1);
	--brtblue: hsla(236, 44%, 40%, 1);
	--copper: hsla(11, 53%, 30%);
	--deepblue: hsla(233, 77%, 26%, 1);
	--dkbabyblue: hsla(237, 43%, 57%, 1);
	--dkgreen: hsla(120, 80%, 20%, 1);
	--dkgrey: hsla(0, 0%, 33%, 1);
	--dkteal: hsla(203, 22%, 26%, 1);
	--huedgrey: hsl(240, 10%, 61%);
	--invisible: hsla(0, 100%, 100%, 0);
	--lessoffwhite: hsla(0, 0%, 98%, 1);
	--ltbabyblue: hsla(200, 33%, 95%, 1);
	--ltgrey: hsla(0, 0%, 90%, 1);
	--ltgrey2: hsl(206, 7%, 81%);
	--midgrey: hsla(0, 0%, 67%, 1);
	--offwhite: hsla(0, 0%, 99%, 1);
	--orange: hsla(47, 100%, 30%, 1);
	--pink: hsla(0, 33%, 96%, 1);
	--pinker: hsl(0, 73%, 80%);
	--plum: hsla(291, 15%, 38%);
	--pukegreen: hsl(71, 95%, 22%);
	--red: hsla(346, 77%, 26%, 1);
	--rustedorange: hsla(23, 37%, 39%, 1);
	--sicklyyellow: hsl(45, 16%, 53%);
	--skyblue: hsla(205, 92%, 37%, 1);
	--teal: hsla(196, 14%, 39%, 1);
	--transparentgrey: hsla(0, 0%, 67%, .8);
	--vdkteal: hsla(196, 27%, 20%, 1);
	--vdkgrey: hsla(0, 0%, 20%, 1);
	--vltgrey: hsla(0, 0%, 96%, 1);
	--white: hsla(0, 0%, 100%, 1);`

	DARKCOLORS = `
	--main-body-color: hsla(60,17%,11%, 1);
	--main-font-color: hsla(60,30%,97%, 1);
	--input-border-color: hsla(60,17%, 22%, 1);
	
	--buttoncolor: hsla(0, 0%, 20%, 1);
	--button-hover: hsla(187, 23%, 24%, 1);
	
	--fieldset-background: hsla(0, 0%, 2%, 1);
	--focus-shadow: rgba(0, 0, 0, 0.5);
	--icons-color: rgba(255, 255, 255, 0.54);
	
	--black: hsla(0, 0%, 95%, 1);
	--blue: hsl(167, 35%, 77%);  /* pea green... */
	--brown: hsla(22, 22%, 74%, 1);
	--brtblue: hsla(236, 44%, 85%, 1);
	--copper: hsla(11, 53%, 83%);
	--deepblue: hsla(57, 65%, 83%, 1);  /* yellow... */
	--dkbabyblue:  hsla(64, 52%, 84%, 1);  /* yellow... */
	--dkgreen: hsla(120, 80%, 80%, 1);
	--dkgrey: hsla(0, 0%, 82%, 1);
	--dkteal: hsla(203, 22%, 82%, 1);
	--huedgrey: hsl(113, 35%, 85%);
	--invisible: hsla(0, 100%, 0%, 0);
	--lessoffwhite: hsla(0, 0%, 2%, 1);
	--ltbabyblue: hsla(200, 33%, 15%, 1);
	--ltgrey: hsla(0, 0%, 20%, 1);
	--ltgrey2: hsl(206, 7%, 11%);
	--midgrey: hsla(0, 0%, 53%, 1);
	--offwhite: hsla(0, 0%, 1%, 1);
	--orange: hsla(47, 100%, 75%, 1);
	--pink: hsla(0, 33%, 15%, 1);
	--pinker: hsl(0, 73%, 20%);
	--plum: hsla(291, 15%, 72%);
	--pukegreen: hsl(71, 95%, 78%);
	--red: hsl(186, 32%, 65%); /* teal... */
	--dkgreen: hsla(120, 80%, 80%, 1);
	--rustedorange: hsla(223, 10%, 76%, 1); /* blue-grey... */
	--sicklyyellow: hsl(45, 16%, 67%);
	--skyblue: hsla(205, 92%, 73%, 1);
	--teal: hsla(196, 14%, 71%, 1);
	--transparentgrey: hsla(0, 0%, 10%, .7);
	--vdkteal: hsla(196, 27%, 80%, 1);
	--vdkgrey: hsla(0, 0%, 80%, 1);
	--vltgrey: hsla(0, 0%, 4%, 1);
	--white: hsla(0, 0%, 0%, 1);`

	// MONOCHROMESANDY - from dark to light...
	// #1B1B19
	// #40403B
	// #64645C
	// #89897E
	// #ADAD9F
	// #D1D1C0
	// #FAFAE5 - cheated up from F6F6E2...
	MONOCHROMESANDY = `
	--main-body-color: #D1D1C0;
	--main-font-color: #1B1B19;
	--input-border-color: #ADAD9F;
	
	--buttoncolor: #FAFAE5;
	--button-hover: #ADAD9F;
	
	--fieldset-background: #ADAD9F;
	--focus-shadow: rgba(0, 0, 0, 0.5);
	--icons-color: #3F3F3A;
	
	--black: #FAFAE5;
	--blue: #FAFAE5; 
	--brown: #FAFAE5;
	--brtblue: #64645C;
	--copper: #1B1B19;
	--deepblue: #3F3F3A; 
	--dkbabyblue:  #3F3F3A;  
	--dkgreen: #1B1B19;
	--dkgrey: #1B1B19;
	--dkteal: #1B1B19;
	--huedgrey: #FAFAE5;
	--invisible: hsla(0, 100%, 0%, 0);
	--lessoffwhite: #ADAD9F;
	--ltbabyblue: #89897E;
	--ltgrey: #FAFAE5;
	--ltgrey2: #FAFAE5;
	--midgrey: #FAFAE5;
	--offwhite: #FAFAE5;
	--orange: #89897E;
	--pink: #89897E;
	--pinker: #89897E;
	--plum: #D1D1C0;
	--pukegreen: #D1D1C0;
	--red: #64645C; 
	--dkgreen: #D1D1C0;
	--rustedorange: #1B1B19; 
	--sicklyyellow: #89897E;
	--skyblue: #D1D1C0;
	--teal: #89897E;
	--transparentgrey: hsla(0, 0%, 10%, .7);
	--vdkteal: #40403B;
	--vdkgrey: #FAFAE5;
	--vltgrey: #FAFAE5;
	--white: #1B1B19;`

	// MONOCHROMEASH - from dark to light...
	// #141518
	// #31343D
	// #4F5461
	// #6D7485
	// #8A93AA
	// #A8B3CE
	// #D0DCFF - cheated up from C6D2F3
	MONOCHROMEASH = `
    --main-body-color: #A8B3CE;
    --main-font-color: #141518;
    --input-border-color: #8A93AA;
    
    --buttoncolor: #D0DCFF;
    --button-hover: #8A93AA;
    
    --fieldset-background: #8A93AA;
    --focus-shadow: rgba(0, 0, 0, 0.5);
    --icons-color: #3F3F3A;
    
    --black: #D0DCFF;
    --blue: #D0DCFF;
    --brown: #D0DCFF;
    --brtblue: #4F5461;
    --copper: #141518;
    --deepblue: #3F3F3A;
    --dkbabyblue:  #3F3F3A;
    --dkgreen: #141518;
    --dkgrey: #141518;
    --dkteal: #31343D;
    --huedgrey: #D0DCFF;
    --invisible: hsla(0, 100%, 0%, 0);
    --lessoffwhite: #8A93AA;
    --ltbabyblue: #6D7485;
    --ltgrey: #D0DCFF;
    --ltgrey2: #D0DCFF;
    --midgrey: #D0DCFF;
    --offwhite: #D0DCFF;
    --orange: #6D7485;
    --pink: #6D7485;
    --pinker: #6D7485;
    --plum: #A8B3CE;
    --pukegreen: #A8B3CE;
    --red: #4F5461;
    --dkgreen: #A8B3CE;
    --rustedorange: #141518;
    --sicklyyellow: #6D7485;
    --skyblue: #A8B3CE;
    --teal: #6D7485;
    --transparentgrey: hsla(0, 0%, 10%, .7);
    --vdkteal: #D0DCFF;
    --vdkgrey: #D0DCFF;
    --vltgrey: #D0DCFF;
    --white: #141518;`

	// TETRADIC - a tatradic + a mono scheme
	// tetradic:
	//
	// #E1E1F5
	// #F5E1EB
	// #F5F5E1
	// #E1F5EB
	//
	// + mono:
	//
	// #19191B
	// #3A3A3F
	// #5C5C64
	// #7D7D88
	// #9F9FAD
	// #C0C0D1
	// #E1E1F5
	TETRADIC = `
	--main-body-color: #3A3A3F;
	--main-font-color: #E1E1F5;
	--input-border-color: #5C5C64;
	
	--buttoncolor: #7D7D88;
	--button-hover: #C0C0D1;
	
	--fieldset-background: #5C5C64;
	--focus-shadow: rgba(0, 0, 0, 0.5);
	--icons-color: #9F9FAD;
	
	--black: #19191B;
	--blue: #F5F5E1; 
	--brown: #E1F5EB;
	--brtblue: #F5F5E1;
	--copper: #E1E1F5;
	--deepblue: #F5E1EB; 
	--dkbabyblue:  #F5F5E1;  
	--dkgreen: #E1E1F5;
	--dkgrey: #F5F5E1;
	--dkteal: #E1E1F5;
	--huedgrey: #E1F5EB;
	--invisible: hsla(0, 100%, 0%, 0);
	--lessoffwhite: #5C5C64;
	--ltbabyblue: #7D7D88;
	--ltgrey: #5C5C64;
	--ltgrey2: #F5F5E1;
	--midgrey: #F5F5E1;
	--offwhite: #E1E1F5;
	--orange: #E1F5EB;
	--pink: #E1F5EB;
	--pinker: #7D7D88;
	--plum: #3A3A3F;
	--pukegreen: #F5F5E1;
	--red: #E1F5EB; 
	--dkgreen: #E1E1F5;
	--rustedorange: #9F9FAD; 
	--sicklyyellow: #E1F5EB;
	--skyblue: #E1F5EB;
	--teal: #E1F5EB;
	--transparentgrey: hsla(0, 0%, 10%, .7);
	--vdkteal: #F5E1EB;
	--vdkgrey: #E1F5EB;
	--vltgrey: #F5E1EB;
	--white: #E1E1F5;`

	// SPLITCOMP - a spit complementary + a mono scheme
	// #DCF2F1
	// #DFDCF2
	// #F2E8DC
	//
	// #151717
	// #363C3C
	// #586060
	// #798584
	// #9AA9A8
	// #BBCECD
	// #DCF2F1
	SPLITCOMP = `
	--main-body-color: #363C3C;
	--main-font-color: #DCF2F1;
	--input-border-color: #586060;
	
	--buttoncolor: #798584;
	--button-hover: #9AA9A8;
	
	--fieldset-background: #586060;
	--focus-shadow: rgba(0, 0, 0, 0.5);
	--icons-color: #9AA9A8;

	--black: #151717;
	--blue: #F2E8DC; 
	--brown: #F2E8DC;
	--brtblue: #F2E8DC;
	--copper: #DCF2F1;
	--deepblue: #DFDCF2; 
	--dkbabyblue:  #F2E8DC;  
	--dkgreen: #DCF2F1;
	--dkgrey: #F2E8DC;
	--dkteal: #DCF2F1;
	--huedgrey: #DCF2F1;
	--invisible: hsla(0, 100%, 0%, 0);
	--lessoffwhite: #586060;
	--ltbabyblue: #798584;
	--ltgrey: #586060;
	--ltgrey2: #F2E8DC;
	--midgrey: #F2E8DC;
	--offwhite: #DCF2F1;
	--orange: #F2E8DC;
	--pink: #F2E8DC;
	--pinker: #798584;
	--plum: #363C3C;
	--pukegreen: #F2E8DC;
	--red: #F2E8DC; 
	--dkgreen: #DCF2F1;
	--rustedorange: #9AA9A8; 
	--sicklyyellow: #F2E8DC;
	--skyblue: #F2E8DC;
	--teal: #F2E8DC;
	--transparentgrey: hsla(0, 0%, 10%, .7);
	--vdkteal: #DFDCF2;
	--vdkgrey: #DCF2F1;
	--vltgrey: #DFDCF2;
	--white: #DCF2F1;`

	// TRIADIC - triadic + monochrome
	// #FFFFD6
	// #D6FFFF
	// #FFD6FF
	//
	// #24241F
	// #49493D
	// #6D6D5C
	// #92927A
	// #B6B699
	// #DBDBB7
	// #FFFFD6
	TRIADIC = `
	--main-body-color: #49493D;
	--main-font-color: #FFFFD6;
	--input-border-color: #6D6D5C;
	
	--buttoncolor: #92927A;
	--button-hover: #DBDBB7;
	
	--fieldset-background: #6D6D5C;
	--focus-shadow: rgba(0, 0, 0, 0.5);
	--icons-color: #B6B699;
	
	--black: #24241F;
	--blue: #F2E8DC; 
	--brown: #F2E8DC;
	--brtblue: #F2E8DC;
	--copper: #FFFFD6;
	--deepblue: #D6FFFF; 
	--dkbabyblue:  #F2E8DC;  
	--dkgreen: #FFFFD6;
	--dkgrey: #F2E8DC;
	--dkteal: #FFFFD6;
	--huedgrey: #D6FFFF;
	--invisible: hsla(0, 100%, 0%, 0);
	--lessoffwhite: #6D6D5C;
	--ltbabyblue: #92927A;
	--ltgrey: #6D6D5C;
	--ltgrey2: #F2E8DC;
	--midgrey: #F2E8DC;
	--offwhite: #FFFFD6;
	--orange: #F2E8DC;
	--pink: #F2E8DC;
	--pinker: #92927A;
	--plum: #49493D;
	--pukegreen: #F2E8DC;
	--red: #F2E8DC; 
	--dkgreen: #FFFFD6;
	--rustedorange: #B6B699; 
	--sicklyyellow: #F2E8DC;
	--skyblue: #F2E8DC;
	--teal: #F2E8DC;
	--transparentgrey: hsla(0, 0%, 10%, .7);
	--vdkteal: #D6FFFF;
	--vdkgrey: #D6FFFF;
	--vltgrey: #D6FFFF;
	--white: #FFFFD6;`
)
