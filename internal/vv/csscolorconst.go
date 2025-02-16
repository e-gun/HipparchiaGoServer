package vv

var (
	CssColorModes = map[string]string{
		"Light":    LIGHTCOLORS,
		"Dark":     DARKCOLORS,
		"Sand":     MONOCHROMESANDY,
		"Ash":      MONOCHROMEASH,
		"Tetradic": TETRADIC,
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
	// #F6F6E2
	MONOCHROMESANDY = `
	--main-body-color: #D1D1C0;
	--main-font-color: #1B1B19;
	--input-border-color: #ADAD9F;
	
	--buttoncolor: #F6F6E2;
	--button-hover: #89897E;
	
	--fieldset-background: #ADAD9F;
	--focus-shadow: rgba(0, 0, 0, 0.5);
	--icons-color: #3F3F3A;
	
	--black: #F6F6E2;
	--blue: #F6F6E2; 
	--brown: #D1D1C0;
	--brtblue: #64645C;
	--copper: #1B1B19;
	--deepblue: #3F3F3A; 
	--dkbabyblue:  #3F3F3A;  
	--dkgreen: #1B1B19;
	--dkgrey: #1B1B19;
	--dkteal: #1B1B19;
	--huedgrey: #F6F6E2;
	--invisible: hsla(0, 100%, 0%, 0);
	--lessoffwhite: #ADAD9F;
	--ltbabyblue: #89897E;
	--ltgrey: #F6F6E2;
	--ltgrey2: #F6F6E2;
	--midgrey: #F6F6E2;
	--offwhite: #F6F6E2;
	--orange: #89897E;
	--pink: #89897E;
	--pinker: #89897E;
	--plum: #D1D1C0;
	--pukegreen: #D1D1C0;
	--red: #64645C; 
	--dkgreen: #D1D1C0;
	--rustedorange: #64645C; 
	--sicklyyellow: #89897E;
	--skyblue: #D1D1C0;
	--teal: #89897E;
	--transparentgrey: hsla(0, 0%, 10%, .7);
	--vdkteal: #F6F6E2;
	--vdkgrey: #F6F6E2;
	--vltgrey: #F6F6E2;
	--white: #1B1B19;`

	// MONOCHROMEASH - from dark to light...
	// #141518
	// #31343D
	// #4F5461
	// #6D7485
	// #8A93AA
	// #A8B3CE
	// #C6D2F3
	MONOCHROMEASH = `
    --main-body-color: #A8B3CE;
    --main-font-color: #141518;
    --input-border-color: #8A93AA;
    
    --buttoncolor: #C6D2F3;
    --button-hover: #6D7485;
    
    --fieldset-background: #8A93AA;
    --focus-shadow: rgba(0, 0, 0, 0.5);
    --icons-color: #3F3F3A;
    
    --black: #C6D2F3;
    --blue: #C6D2F3;
    --brown: #A8B3CE;
    --brtblue: #4F5461;
    --copper: #141518;
    --deepblue: #3F3F3A;
    --dkbabyblue:  #3F3F3A;
    --dkgreen: #141518;
    --dkgrey: #141518;
    --dkteal: #141518;
    --huedgrey: #C6D2F3;
    --invisible: hsla(0, 100%, 0%, 0);
    --lessoffwhite: #8A93AA;
    --ltbabyblue: #6D7485;
    --ltgrey: #C6D2F3;
    --ltgrey2: #C6D2F3;
    --midgrey: #C6D2F3;
    --offwhite: #C6D2F3;
    --orange: #6D7485;
    --pink: #6D7485;
    --pinker: #6D7485;
    --plum: #A8B3CE;
    --pukegreen: #A8B3CE;
    --red: #4F5461;
    --dkgreen: #A8B3CE;
    --rustedorange: #4F5461;
    --sicklyyellow: #6D7485;
    --skyblue: #A8B3CE;
    --teal: #6D7485;
    --transparentgrey: hsla(0, 0%, 10%, .7);
    --vdkteal: #C6D2F3;
    --vdkgrey: #C6D2F3;
    --vltgrey: #C6D2F3;
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
	--huedgrey: #E1E1F5;
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
	--vdkgrey: #E1E1F5;
	--vltgrey: #F5E1EB;
	--white: #E1E1F5;`
)
