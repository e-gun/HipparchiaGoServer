package vv

var (
	CssColorModes = map[string]string{
		"Light": LIGHTCOLORS,
		"Dark":  DARKCOLORS,
		"Sand":  MONOCHROMESANDY,
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

	DARKCOLORS_ORIGINAL = `
	--main-body-color: hsla(0, 0%, 10%, 1);
	--main-font-color: hsla(0, 0%, 95%, 1);
	--input-border-color: hsla(0, 0%, 20%, 1);
	
	--buttoncolor: hsla(0, 0%, 20%, 1);
	--button-hover: hsla(0, 0%, 10%, 1);
	
	--fieldset-background: hsla(0, 0%, 2%, 1);
	--focus-shadow: rgba(0, 0, 0, 0.5);
	--icons-color: rgba(255, 255, 255, 0.54);
	
	--black: hsla(0, 0%, 95%, 1);
	--blue: hsl(167, 35%, 77%);  /* pea green... */
	--brown: hsla(22, 22%, 74%, 1);
	--brtblue: hsla(236, 44%, 85%, 1);
	--copper: hsla(11, 53%, 83%);
	--deepblue: hsla(64, 88%, 84%, 1);  /* yellow... */
	--dkbabyblue:  hsla(64, 52%, 84%, 1);  /* yellow... */
	--dkgreen: hsla(120, 80%, 80%, 1);
	--dkgrey: hsla(0, 0%, 82%, 1);
	--dkteal: hsla(203, 22%, 82%, 1);
	--huedgrey: hsl(113, 35%, 79%);
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
	--red: hsl(258, 35%, 83%); /* lt purple... */
	--dkgreen: hsla(120, 80%, 80%, 1);
	--rustedorange: hsla(23, 37%, 71%, 1);
	--sicklyyellow: hsl(45, 16%, 67%);
	--skyblue: hsla(205, 92%, 73%, 1);
	--teal: hsla(196, 14%, 71%, 1);
	--transparentgrey: hsla(0, 0%, 23%, .8);
	--vdkteal: hsla(196, 27%, 80%, 1);
	--vdkgrey: hsla(0, 0%, 80%, 1);
	--vltgrey: hsla(0, 0%, 4%, 1);
	--white: hsla(0, 0%, 0%, 1);`

	MONOCHROMESANDY = `
	--main-body-color: #D1D1C0;
	--main-font-color: #1B1B19;
	--input-border-color: #ADAD9F;
	
	--buttoncolor: #F5F5E1;
	--button-hover: #88887D;
	
	--fieldset-background: #ADAD9F;
	--focus-shadow: rgba(0, 0, 0, 0.5);
	--icons-color: #3F3F3A;
	
	--black: #F5F5E1;
	--blue: #ADAD9F; 
	--brown: #88887D;
	--brtblue: #64645C;
	--copper: #1B1B19;
	--deepblue: #3F3F3A; 
	--dkbabyblue:  #3F3F3A;  
	--dkgreen: #1B1B19;
	--dkgrey: #1B1B19;
	--dkteal: #1B1B19;
	--huedgrey: #F5F5E1;
	--invisible: hsla(0, 100%, 0%, 0);
	--lessoffwhite: #ADAD9F;
	--ltbabyblue: #88887D;
	--ltgrey: #F5F5E1;
	--ltgrey2: #F5F5E1;
	--midgrey: #F5F5E1;
	--offwhite: #F5F5E1;
	--orange: #88887D;
	--pink: #88887D;
	--pinker: #88887D;
	--plum: #D1D1C0;
	--pukegreen: #D1D1C0;
	--red: #64645C; 
	--dkgreen: #D1D1C0;
	--rustedorange: #64645C; 
	--sicklyyellow: #88887D;
	--skyblue: #D1D1C0;
	--teal: #88887D;
	--transparentgrey: hsla(0, 0%, 10%, .7);
	--vdkteal: #F5F5E1;
	--vdkgrey: #F5F5E1;
	--vltgrey: #F5F5E1;
	--white: #1B1B19;`
)
