package clr

const (
	DEFAULTCOLORS = "Light"

	LIGHTCOLORSORIGINAL = `
	--main-body-color: hsla(0, 0%, 98%, 1);
	--main-font-color: hsla(0, 0%, 6%, 1);
	--input-border-color: hsla(0, 0%, 92%, 1);
	
	--buttoncolor: hsla(0, 0%, 93%, 1);
	--button-hover: hsla(0, 0%, 90%, 1);
	--fieldset-background: hsla(0, 0%, 98%, 1);
	--focus-shadow: rgba(0, 0, 0, 0.5);
	--icons-color: rgba(0, 0, 0, 0.54);
	
	--greyborders: hsla(0, 0%, 67%, 1);
	--boxshadows: hsla(0, 0%, 67%, 1);
	--modalbackground: hsla(0, 0%, 33%, 1);
	--sidednavcolor: hsla(0, 0%, 90%, 1);
	--nthrowcol: hsla(0, 0%, 90%, 1);

	--locuscolor: hsl(240, 10%, 61%);
	--notations: hsla(196, 14%, 39%, 1);
	--dottedborders: hsla(0, 0%, 0%, 1);
	--bracketcolor1: hsla(22, 22%, 26%, 1);
	--bracketcolor2: hsl(240, 10%, 61%);
	--bracketcolor3: hsl(71, 95%, 22%);
	--bracketcolor4: hsla(237, 43%, 57%, 1);

	--dictauthwkcolor: hsla(237, 43%, 57%, 1);
	--dictdefaultcolor: hsla(0, 0%, 20%, 1);
	--dictlevelscolor: hsla(346, 77%, 26%, 1);
	--dictquotecolor: hsla(0, 0%, 0%, 1);
	
	--invisible: hsla(0, 100%, 100%, 0);
	--transparentgrey: hsla(0, 0%, 67%, .8);
	
	--hgsmono00: hsla(0, 0%, 0%, 1);
	--hgsmono01: hsla(0, 0%, 33%, 1);
	--hgsmono02: hsla(0, 0%, 90%, 1);
	--hgsmono03: hsl(206, 7%, 81%);
	--hgsmono03: hsla(0, 0%, 67%, 1);
	--hgsmono05: hsla(0, 0%, 99%, 1);
	--hgsmono07: hsla(0, 0%, 20%, 1);
	--hgsmono08: hsla(0, 0%, 96%, 1);
	--hgsmono09: hsla(0, 0%, 100%, 1);
	
	--hgscol01: hsla(240, 100%, 27%, 1);
	--hgscol02: hsla(22, 22%, 26%, 1);
	--hgscol03: hsla(236, 44%, 40%, 1);
	--hgscol04: hsla(11, 53%, 30%);
	--hgscol05: hsla(233, 77%, 26%, 1);
	--hgscol06: hsla(237, 43%, 57%, 1);
	--hgscol07: hsla(120, 80%, 20%, 1);
	--hgscol08: hsla(203, 22%, 26%, 1);
	--hgscol09: hsl(240, 10%, 61%);
	--hgscol10: hsla(0, 0%, 98%, 1);
	--hgscol11: hsla(200, 33%, 95%, 1);
	--hgscol12: hsla(47, 100%, 30%, 1);
	--hgscol13: hsla(0, 33%, 96%, 1);
	--hgscol14: hsla(291, 15%, 38%);
	--hgscol15: hsl(71, 95%, 22%);
	--hgscol16: hsla(346, 77%, 26%, 1);
	--hgscol17: hsla(23, 37%, 39%, 1);
	--hgscol18: hsl(45, 16%, 53%);
	--hgscol19: hsla(205, 92%, 37%, 1);
	--hgscol20: hsla(196, 14%, 39%, 1);
	--hgscol21: hsla(196, 27%, 20%, 1);
`

	DARKCOLORSMANUAL = `
	--main-body-color: hsla(60,17%,11%, 1);
	--main-font-color: hsla(60,30%,97%, 1);
	--input-border-color: hsla(60,17%, 22%, 1);
	
	--buttoncolor: hsla(0, 0%, 20%, 1);
	--button-hover: hsla(187, 23%, 24%, 1);
	
	--fieldset-background: hsla(0, 0%, 2%, 1);
	--focus-shadow: rgba(0, 0, 0, 0.5);
	--icons-color: rgba(255, 255, 255, 0.54);
	
	--greyborders: hsla(0, 0%, 53%, 1);
    --boxshadows: hsla(0, 0%, 53%, 1);
	--modalbackground: hsla(0, 0%, 82%, 1);
	--sidednavcolor: hsla(0, 0%, 20%, 1);
	--nthrowcol: hsla(0, 0%, 20%, 1);

	--locuscolor: hsl(113, 35%, 85%);
	--notations: hsla(196, 14%, 71%, 1);
	--dottedborders: hsla(0, 0%, 100%, 1);
	--bracketcolor1: hsla(22, 22%, 74%, 1);
	--bracketcolor2: hsl(113, 35%, 85%);
	--bracketcolor3: hsl(71, 95%, 78%);
	--bracketcolor4: hsla(64, 52%, 84%, 1);

	--dictauthwkcolor: hsla(64, 52%, 84%, 1);
	--dictdefaultcolor: hsla(0, 0%, 80%, 1);
	--dictlevelscolor: hsl(186, 32%, 65%);
	--dictquotecolor: hsla(0, 0%, 82%, 1);
	--dicttranslcolor: hsla(0, 0%, 80%, 1);

	--hgsmono00: hsla(0, 0%, 95%, 1);
	--hgscol01: hsl(167, 35%, 77%);  /* pea green... */
	--hgscol02: hsla(22, 22%, 74%, 1);
	--hgscol03: hsla(236, 44%, 85%, 1);
	--hgscol04: hsla(11, 53%, 83%);
	--hgscol05: hsla(57, 65%, 83%, 1);  /* yellow... */
	--hgscol06:  hsla(64, 52%, 84%, 1);  /* yellow... */
	--hgscol07: hsla(120, 80%, 80%, 1);
	--hgsmono01: hsla(0, 0%, 82%, 1);
	--hgscol08: hsla(203, 22%, 82%, 1);
	--hgscol09: hsl(113, 35%, 85%);
	--invisible: hsla(0, 100%, 0%, 0);
	--hgscol10: hsla(0, 0%, 2%, 1);
	--hgscol11: hsla(200, 33%, 15%, 1);
	--hgsmono02: hsla(0, 0%, 20%, 1);
	--hgsmono03: hsl(206, 7%, 11%);
	--hgsmono04: hsla(0, 0%, 53%, 1);
	--hgsmono05: hsla(0, 0%, 1%, 1);
	--hgscol12: hsla(47, 100%, 75%, 1);
	--hgscol13: hsla(0, 33%, 15%, 1);
	--hgscol14: hsla(291, 15%, 72%);
	--hgscol15: hsl(71, 95%, 78%);
	--hgscol16: hsl(186, 32%, 65%); /* teal... */
	--hgscol07: hsla(120, 80%, 80%, 1);
	--hgscol17: hsla(223, 10%, 76%, 1); /* blue-grey... */
	--hgscol18: hsl(45, 16%, 67%);
	--hgscol19: hsla(205, 92%, 73%, 1);
	--hgscol20: hsla(196, 14%, 71%, 1);
	--transparentgrey: hsla(0, 0%, 90%, .7);
	--hgscol21: hsla(196, 27%, 80%, 1);
	--hgsmono07: hsla(0, 0%, 80%, 1);
	--hgsmono08: hsla(0, 0%, 4%, 1);
	--hgsmono09: hsla(0, 0%, 0%, 1);`

	// MONOCHROMESANDYMANUAL - from dark to light...
	// #1B1B19
	// #40403B
	// #64645C
	// #89897E
	// #ADAD9F
	// #D1D1C0
	// #FAFAE5 - cheated up from F6F6E2...
	MONOCHROMESANDYMANUAL = `
	--main-body-color: #D1D1C0;
	--main-font-color: #1B1B19;
	--input-border-color: #ADAD9F;
	
	--buttoncolor: #FAFAE5;
	--button-hover: #ADAD9F;
	
	--fieldset-background: #ADAD9F;
	--focus-shadow: rgba(0, 0, 0, 0.5);
	--icons-color: #3F3F3A;
	
	--greyborders: #FAFAE5;
    --boxshadows: #FAFAE5;
	--modalbackground: #1B1B19;
	--sidednavcolor: #FAFAE5;
	--nthrowcol: #FAFAE5;

	--locuscolor: #FAFAE5;
	--notations: #89897E;
	--dottedborders: #FAFAE5;
	--bracketcolor1: #FAFAE5;
	--bracketcolor2: #FAFAE5;
	--bracketcolor3: #D1D1C0;
	--bracketcolor4: 3F3F3A;

	--dictauthwkcolor: #3F3F3A;
	--dictdefaultcolor: #FAFAE5;
	--dictlevelscolor: #64645C; 
	--dictquotecolor: #FAFAE5;
	--dicttranslcolor: #FAFAE5;

	--hgsmono00: #FAFAE5;
	--hgscol01: #FAFAE5; 
	--hgscol02: #FAFAE5;
	--hgscol03: #64645C;
	--hgscol04: #1B1B19;
	--hgscol05: #3F3F3A; 
	--hgscol06:  #3F3F3A;  
	--hgscol07: #1B1B19;
	--hgsmono01: #1B1B19;
	--hgscol08: #1B1B19;
	--hgscol09: #FAFAE5;
	--invisible: hsla(0, 100%, 0%, 0);
	--hgscol10: #ADAD9F;
	--hgscol11: #89897E;
	--hgsmono02: #FAFAE5;
	--hgsmono03: #FAFAE5;
	--hgsmono04: #FAFAE5;
	--hgsmono05: #FAFAE5;
	--hgscol12: #89897E;
	--hgscol13: #89897E;
	--hgscol14: #D1D1C0;
	--hgscol15: #D1D1C0;
	--hgscol16: #64645C; 
	--hgscol07: #D1D1C0;
	--hgscol17: #1B1B19; 
	--hgscol18: #89897E;
	--hgscol19: #D1D1C0;
	--hgscol20: #89897E;
	--transparentgrey: hsla(0, 0%, 10%, .7);
	--hgscol21: #40403B;
	--hgsmono07: #FAFAE5;
	--hgsmono08: #FAFAE5;
	--hgsmono09: #1B1B19;`

	// MONOCHROMEASHMANUAL - from dark to light...
	// #141518
	// #31343D
	// #4F5461
	// #6D7485
	// #8A93AA
	// #A8B3CE
	// #D0DCFF - cheated up from C6D2F3
	MONOCHROMEASHMANUAL = `
    --main-body-color: #A8B3CE;
    --main-font-color: #141518;
    --input-border-color: #8A93AA;
    
    --buttoncolor: #D0DCFF;
    --button-hover: #8A93AA;
    
    --fieldset-background: #8A93AA;
    --focus-shadow: rgba(0, 0, 0, 0.5);
    --icons-color: #3F3F3A;
    
	--greyborders: #D0DCFF;
    --boxshadows: #D0DCFF;
	--modalbackground: #141518;
	--locuscolor: #D0DCFF;
	--notations: #6D7485;
	--sidednavcolor: #D0DCFF;
	--nthrowcol: #D0DCFF;

	--dottedborders: #FAFAE5;
	--bracketcolor1: #D0DCFF;
	--bracketcolor2: #D0DCFF;
	--bracketcolor3: #A8B3CE;
	--bracketcolor4: 3F3F3A;

	--dictauthwkcolor: #4F5461;
	--dictdefaultcolor: #D0DCFF;
	--dictlevelscolor: #4F5461; 
	--dictquotecolor: #D0DCFF;
	--dicttranslcolor: #D0DCFF;

    --hgsmono00: #D0DCFF;
    --hgscol01: #D0DCFF;
    --hgscol02: #D0DCFF;
    --hgscol03: #4F5461;
    --hgscol04: #141518;
    --hgscol05: #3F3F3A;
    --hgscol06:  #3F3F3A;
    --hgscol07: #141518;
    --hgsmono01: #141518;
    --hgscol08: #31343D;
    --hgscol09: #D0DCFF;
    --invisible: hsla(0, 100%, 0%, 0);
    --hgscol10: #8A93AA;
    --hgscol11: #6D7485;
    --hgsmono02: #D0DCFF;
    --hgsmono03: #D0DCFF;
    --hgsmono04: #D0DCFF;
    --hgsmono05: #D0DCFF;
    --hgscol12: #6D7485;
    --hgscol13: #6D7485;
    --hgscol14: #A8B3CE;
    --hgscol15: #A8B3CE;
    --hgscol16: #4F5461;
    --hgscol07: #A8B3CE;
    --hgscol17: #141518;
    --hgscol18: #6D7485;
    --hgscol19: #A8B3CE;
    --hgscol20: #6D7485;
    --transparentgrey: hsla(0, 0%, 10%, .7);
    --hgscol21: #D0DCFF;
    --hgsmono07: #D0DCFF;
    --hgsmono08: #D0DCFF;
    --hgsmono09: #141518;`

	// TETRADICMANUAL - a tatradic + a mono scheme
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
	TETRADICMANUAL = `
	--main-body-color: #3A3A3F;
	--main-font-color: #E1E1F5;
	--input-border-color: #5C5C64;
	
	--buttoncolor: #7D7D88;
	--button-hover: #C0C0D1;
	
	--fieldset-background: #5C5C64;
	--focus-shadow: rgba(0, 0, 0, 0.5);
	--icons-color: #9F9FAD;
	
	--greyborders: #F5F5E1;
    --boxshadows: #F5F5E1;
	--modalbackground: #F5F5E1;
	--sidednavcolor: #5C5C64;
	--nthrowcol: #5C5C64;

	--locuscolor: #E1F5EB;
	--notations: #E1F5EB;
	--dottedborders: #19191B;
	--bracketcolor1: #E1F5EB;
	--bracketcolor2: #E1F5EB;
	--bracketcolor3: #F5F5E1;
	--bracketcolor4: F5F5E1;

	--dictauthwkcolor: #E1F5EB;
	--dictdefaultcolor: #E1F5EB;
	--dictlevelscolor: #E1F5EB; 
	--dictquotecolor: #F5F5E1;
	--dicttranslcolor: #E1F5EB;

	--hgsmono00: #19191B;
	--hgscol01: #F5F5E1; 
	--hgscol02: #E1F5EB;
	--hgscol03: #F5F5E1;
	--hgscol04: #E1E1F5;
	--hgscol05: #F5E1EB; 
	--hgscol06:  #F5F5E1;  
	--hgscol07: #E1E1F5;
	--hgsmono01: #F5F5E1;
	--hgscol08: #E1E1F5;
	--hgscol09: #E1F5EB;
	--invisible: hsla(0, 100%, 0%, 0);
	--hgscol10: #5C5C64;
	--hgscol11: #7D7D88;
	--hgsmono02: #5C5C64;
	--hgsmono03: #F5F5E1;
	--hgsmono04: #F5F5E1;
	--hgsmono05: #E1E1F5;
	--hgscol12: #E1F5EB;
	--hgscol13: #E1F5EB;
	--hgscol14: #3A3A3F;
	--hgscol15: #F5F5E1;
	--hgscol16: #E1F5EB; 
	--hgscol07: #E1E1F5;
	--hgscol17: #9F9FAD; 
	--hgscol18: #E1F5EB;
	--hgscol19: #E1F5EB;
	--hgscol20: #E1F5EB;
	--transparentgrey: hsla(0, 0%, 10%, .7);
	--hgscol21: #F5E1EB;
	--hgsmono07: #E1F5EB;
	--hgsmono08: #F5E1EB;
	--hgsmono09: #E1E1F5;`

	// SPLITCOMPMANUAL - a spit complementary + a mono scheme
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
	SPLITCOMPMANUAL = `
	--main-body-color: #363C3C;
	--main-font-color: #DCF2F1;
	--input-border-color: #586060;
	
	--buttoncolor: #798584;
	--button-hover: #9AA9A8;
	
	--fieldset-background: #586060;
	--focus-shadow: rgba(0, 0, 0, 0.5);
	--icons-color: #9AA9A8;

	--greyborders: #F2E8DC;
    --boxshadows: #F2E8DC;
	--modalbackground: #F2E8DC;
	--sidednavcolor: #586060;
	--nthrowcol: #586060;

	--locuscolor: #DCF2F1;
	--notations: #F2E8DC;
	--dottedborders: #151717;
	--bracketcolor1: #F2E8DC;
	--bracketcolor2: #DCF2F1;
	--bracketcolor3: #F2E8DC;
	--bracketcolor4: F2E8DC;

	--dictauthwkcolor: #F2E8DC;
	--dictdefaultcolor: #DCF2F1;
	--dictlevelscolor: #F2E8DC; 
	--dictquotecolor: #F2E8DC;
	--dicttranslcolor: #DCF2F1;

	--hgsmono00: #151717;
	--hgscol01: #F2E8DC; 
	--hgscol02: #F2E8DC;
	--hgscol03: #F2E8DC;
	--hgscol04: #DCF2F1;
	--hgscol05: #DFDCF2; 
	--hgscol06:  #F2E8DC;  
	--hgscol07: #DCF2F1;
	--hgsmono01: #F2E8DC;
	--hgscol08: #DCF2F1;
	--hgscol09: #DCF2F1;
	--invisible: hsla(0, 100%, 0%, 0);
	--hgscol10: #586060;
	--hgscol11: #798584;
	--hgsmono02: #586060;
	--hgsmono03: #F2E8DC;
	--hgsmono04: #F2E8DC;
	--hgsmono05: #DCF2F1;
	--hgscol12: #F2E8DC;
	--hgscol13: #F2E8DC;
	--hgscol14: #363C3C;
	--hgscol15: #F2E8DC;
	--hgscol16: #F2E8DC; 
	--hgscol07: #DCF2F1;
	--hgscol17: #9AA9A8; 
	--hgscol18: #F2E8DC;
	--hgscol19: #F2E8DC;
	--hgscol20: #F2E8DC;
	--transparentgrey: hsla(0, 0%, 10%, .7);
	--hgscol21: #DFDCF2;
	--hgsmono07: #DCF2F1;
	--hgsmono08: #DFDCF2;
	--hgsmono09: #DCF2F1;`

	// TRIADICMANUAL - triadic + monochrome
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
	TRIADICMANUAL = `
	--main-body-color: #49493D;
	--main-font-color: #FFFFD6;
	--input-border-color: #6D6D5C;
	
	--buttoncolor: #92927A;
	--button-hover: #DBDBB7;
	
	--fieldset-background: #6D6D5C;
	--focus-shadow: rgba(0, 0, 0, 0.5);
	--icons-color: #B6B699;
	
	--greyborders: #F2E8DC;
    --boxshadows: #F2E8DC;
	--modalbackground: #F2E8DC;
	--locuscolor: #D6FFFF;
	--notations: #F2E8DC;
	--sidednavcolor: #6D6D5C;
	--nthrowcol: #6D6D5C;

	--dottedborders: #24241F;
	--bracketcolor1: #F2E8DC;
	--bracketcolor2: #D6FFFF;
	--bracketcolor3: #F2E8DC;
	--bracketcolor4: F2E8DC;

	--dictauthwkcolor: #F2E8DC;
	--dictdefaultcolor: #D6FFFF;
	--dictlevelscolor: #F2E8DC; 
	--dictquotecolor: #F2E8DC;
	--dicttranslcolor: #D6FFFF;

	--hgsmono00: #24241F;
	--hgscol01: #F2E8DC; 
	--hgscol02: #F2E8DC;
	--hgscol03: #F2E8DC;
	--hgscol04: #FFFFD6;
	--hgscol05: #D6FFFF; 
	--hgscol06:  #F2E8DC;  
	--hgscol07: #FFFFD6;
	--hgsmono01: #F2E8DC;
	--hgscol08: #FFFFD6;
	--hgscol09: #D6FFFF;
	--invisible: hsla(0, 100%, 0%, 0);
	--hgscol10: #6D6D5C;
	--hgscol11: #92927A;
	--hgsmono02: #6D6D5C;
	--hgsmono03: #F2E8DC;
	--hgsmono04: #F2E8DC;
	--hgsmono05: #FFFFD6;
	--hgscol12: #F2E8DC;
	--hgscol13: #F2E8DC;
	--hgscol14: #49493D;
	--hgscol15: #F2E8DC;
	--hgscol16: #F2E8DC; 
	--hgscol07: #FFFFD6;
	--hgscol17: #B6B699; 
	--hgscol18: #F2E8DC;
	--hgscol19: #F2E8DC;
	--hgscol20: #F2E8DC;
	--transparentgrey: hsla(0, 0%, 10%, .7);
	--hgscol21: #D6FFFF;
	--hgsmono07: #D6FFFF;
	--hgsmono08: #D6FFFF;
	--hgsmono09: #FFFFD6;`
)
