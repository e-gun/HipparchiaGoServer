//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package clr

// if the body is a '5' or '6', then it is safest to keep most of the other colors ina the 2-0 range...

const (
	MONO = `
	--main-body-color: {{.colorA6}};
	--main-font-color: {{.colorA0}};
	--input-border-color: {{.colorA4}};
	
	--buttoncolor: {{.colorA4}};
	--button-hover: {{.colorA3}};
	
	--fieldset-background: {{.colorA4}};
	--focus-shadow: rgba(0, 0, 0, 0.7);
	--icons-color: {{.colorA1}};
	
	--greyborders: {{.colorA3}};
	--boxshadows: {{.colorA4}};
	--dropdown-backgrounds: {{.colorA5}};
	--modalbackground: {{.colorA5}};
	--modalheader: {{.colorA4}};
	--sidednavcolor: {{.colorA4}};
	--nthrowcol: {{.colorA5}};
	--buttonandcheckboxborder: {{.colorA1}};
	--buttonandcheckboxcontents: {{.colorA2}};

	--locuscolor: {{.colorA1}};
	--notations: {{.colorA2}};
	--dottedborders: {{.colorA1}};
	--bracketcolor1: {{.colorA2}};
	--bracketcolor2: {{.colorA2}};
	--bracketcolor3: {{.colorA2}};
	--bracketcolor4: {{.colorA2}};
	
	--dictauthwkcolor: {{.colorA0}};
	--dictbiblcolor: {{.colorA1}};
	--dictdefaultcolor: {{.colorA1}};
	--dictlevelscolor: {{.colorA2}};
	--dictquotecolor: {{.colorA2}};
	--dicttranslcolor: {{.colorA0}};

	--invisible: hsla(0, 100%, 100%, 0);
	--transparentgrey: hsla(0, 0%, 50%, .5);
	--textshadowcolor: {{.textshadowColor}};
	
	--hgscol01: {{.colorA3}};
	--hgscol02: {{.colorA6}};
	--hgscol03: {{.colorA1}};
	--hgscol04: {{.colorA0}};
	--hgscol05: {{.colorA2}};
	--hgscol06: {{.colorA2}};
	--hgscol07: {{.colorA0}};
	--hgscol08: {{.colorA0}};
	--hgscol09: {{.colorA3}};
	--hgscol10: {{.colorA4}};
	--hgscol11: {{.colorA5}};
	--hgscol12: {{.colorA2}};
	--hgscol13: {{.colorA4}};
	--hgscol14: {{.colorA5}};
	--hgscol15: {{.colorA5}};
	--hgscol16: {{.colorA2}};
	--hgscol17: {{.colorA0}};
	--hgscol18: {{.colorA3}};
	--hgscol19: {{.colorA0}};
	--hgscol20: {{.colorA2}};
	--hgscol21: {{.colorA1}};
	
	--hgsmono09: {{.colorA0}};
	--hgsmono05: {{.colorA6}};
	--hgsmono08: {{.colorA6}};
	--hgsmono02: {{.colorA6}};
	--hgsmono03: {{.colorA5}};
	--hgsmono04: {{.colorA5}};
	--hgsmono01: {{.colorA0}};
	--hgsmono07: {{.colorA6}};
	--hgsmono00: {{.colorA6}};
`

	// SPLITCOMP - only A, B, C available; the geometry is *very* similar to triadic...
	SPLITCOMP = `
    --main-body-color: {{.colorA1}};
    --main-font-color: {{.colorA6}};
    --input-border-color: {{.colorA2}};
    
    --buttoncolor: {{.colorA3}};
    --button-hover: {{.colorA4}};
    
    --fieldset-background: {{.colorA2}};
    --focus-shadow: rgba(0, 0, 0, 0.5);
    --icons-color: {{.colorC4}};

	--greyborders: {{.colorA3}};
	--boxshadows: {{.colorA4}};
	--dropdown-backgrounds: {{.colorA0}};
	--modalbackground: {{.colorB0}};
	--modalheader: {{.colorB2}};
	--sidednavcolor: {{.colorA2}};
	--nthrowcol: {{.colorB1}};
	--buttonandcheckboxborder: {{.colorA5}};
	--buttonandcheckboxcontents: {{.colorA4}};

	--locuscolor: {{.colorB5}};
	--notations: {{.colorB6}};
	--dottedborders: {{.colorA6}};
	--bracketcolor1: {{.colorA5}};
	--bracketcolor2: {{.colorB5}};
	--bracketcolor3: {{.colorC5}};
	--bracketcolor4: {{.colorA6}};
	
	--dictauthwkcolor: {{.colorC5}};
	--dictbiblcolor: {{.colorB5}};
	--dictdefaultcolor: {{.colorA6}};
	--dictlevelscolor: {{.colorA4}};
	--dictquotecolor: {{.colorA6}};
	--dicttranslcolor: {{.colorB6}};

	--invisible: hsla(0, 100%, 100%, 0);
	--transparentgrey: hsla(0, 0%, 50%, .5);
	--textshadowcolor: {{.textshadowColor}};

	/* R */
    --hgscol13: {{.colorB1}};
    --hgscol04: {{.colorB2}};
    --hgscol12: {{.colorC5}};
    --hgscol16: {{.colorB5}}; 
    --hgscol17: {{.colorB4}}; 
    --hgscol02: {{.colorB5}};
    --hgscol14: {{.colorB6}};

	/* G */
    --hgscol18: {{.colorA1}};
    --hgscol15: {{.colorA2}};
    --hgscol20: {{.colorA3}};
    --hgscol07: {{.colorA4}};
    --hgscol08: {{.colorA5}};
    --hgscol21: {{.colorA6}};

	/* B */
    --hgscol11: {{.colorC2}};
    --hgscol19: {{.colorC5}};
    --hgscol03: {{.colorC5}};
    --hgscol01: {{.colorC4}};
    --hgscol05: {{.colorC5}}; 
    --hgscol06: {{.colorC6}};  

    --hgscol09: {{.colorC5}};
    --hgscol10: {{.colorA0}};

	/* b/w */
	--hgsmono09: {{.colorA6}};
	--hgsmono05: {{.colorB6}};
	--hgsmono08: {{.colorC6}};
	--hgsmono02: {{.colorA3}};
	--hgsmono03: {{.colorB3}};
	--hgsmono04: {{.colorC5}};
	--hgsmono01: {{.colorA5}};
	--hgsmono07: {{.colorB5}};
	--hgsmono00: {{.colorC6}};
`

	// SQUARE - A, B, C, D available
	SQUARE = `
    --main-body-color: {{.colorA0}};
    --main-font-color: {{.colorA6}};
    --input-border-color: {{.colorA2}};
    
    --buttoncolor: {{.colorB2}};
    --button-hover: {{.colorB4}};
    
    --fieldset-background: {{.colorA2}};
    --focus-shadow: rgba(0, 0, 0, 0.5);
    --icons-color: {{.colorB4}};
    
 	--greyborders: {{.colorA3}};
	--boxshadows: {{.colorA4}};
	--dropdown-backgrounds: {{.colorA1}};
	--modalbackground: {{.colorD0}};
	--modalheader: {{.colorD2}};
	--sidednavcolor: {{.colorC0}};
	--nthrowcol: {{.colorD1}};
	--buttonandcheckboxborder: {{.colorC5}};
	--buttonandcheckboxcontents: {{.colorC4}};

	--locuscolor: {{.colorD5}};
	--notations: {{.colorC5}};
	--dottedborders: {{.colorA6}};
	--bracketcolor1: {{.colorA5}};
	--bracketcolor2: {{.colorB5}};
	--bracketcolor3: {{.colorC5}};
	--bracketcolor4: {{.colorD6}};
	
	--dictauthwkcolor: {{.colorC5}};
	--dictbiblcolor: {{.colorB5}};
	--dictdefaultcolor: {{.colorA6}};
	--dictlevelscolor: {{.colorB4}};
	--dictquotecolor: {{.colorA6}};
	--dicttranslcolor: {{.colorB6}};

	--invisible: hsla(0, 100%, 100%, 0);
	--transparentgrey: hsla(0, 0%, 50%, .5);
	--textshadowcolor: {{.textshadowColor}};

	/* R */
    --hgscol13: {{.colorC2}};
    --hgscol04: {{.colorB4}};
    --hgscol12: {{.colorC5}};
    --hgscol16: {{.colorB4}}; 
    --hgscol17: {{.colorD5}}; 
    --hgscol02: {{.colorC0}};
    --hgscol14: {{.colorA1}};

	/* G */
    --hgscol18: {{.colorC0}};
    --hgscol15: {{.colorC0}};
    --hgscol20: {{.colorC4}};
    --hgscol07: {{.colorB6}};
    --hgscol08: {{.colorB6}};
    --hgscol21: {{.colorC4}};

	/* B */
    --hgscol11: {{.colorB2}};
    --hgscol19: {{.colorB5}};
    --hgscol03: {{.colorC5}};
    --hgscol01: {{.colorB4}};
    --hgscol05: {{.colorC6}}; 
    --hgscol06: {{.colorB4}};  

    --hgscol09: {{.colorC4}};
    --hgscol10: {{.colorB0}};

	/* b/w */
	--hgsmono09: {{.colorA6}};
	--hgsmono05: {{.colorB5}};
	--hgsmono08: {{.colorC5}};
	--hgsmono02: {{.colorA1}};
	--hgsmono03: {{.colorB4}};
	--hgsmono04: {{.colorC4}};
	--hgsmono01: {{.colorD4}};
	--hgsmono07: {{.colorC2}};
	--hgsmono00: {{.colorA0}};
`

	TETRADTMPL = `
    --main-body-color: {{.colorA0}};
    --main-font-color: {{.colorA6}};
    --input-border-color: {{.colorA2}};
    
    --buttoncolor: {{.colorB1}};
    --button-hover: {{.colorB4}};
    
    --fieldset-background: {{.colorA2}};
    --focus-shadow: rgba(0, 0, 0, 0.5);
    --icons-color: {{.colorE5}};
    
 	--greyborders: {{.colorA2}};
	--boxshadows: {{.colorA4}};
	--modalbackground: {{.colorE0}};
	--dropdown-backgrounds: {{.colorA1}};
	--modalheader: {{.colorD2}};
	--sidednavcolor: {{.colorE0}};
	--nthrowcol: {{.colorD1}};
	--buttonandcheckboxborder: {{.colorB2}};
	--buttonandcheckboxcontents: {{.colorB3}};

	--locuscolor: {{.colorD5}};
	--notations: {{.colorE5}};
	--dottedborders: {{.colorA6}};
	--bracketcolor1: {{.colorA5}};
	--bracketcolor2: {{.colorB5}};
	--bracketcolor3: {{.colorC5}};
	--bracketcolor4: {{.colorD6}};
	
	--dictauthwkcolor: {{.colorC5}};
	--dictbiblcolor: {{.colorB5}};
	--dictdefaultcolor: {{.colorA6}};
	--dictlevelscolor: {{.colorB4}};
	--dictquotecolor: {{.colorA6}};
	--dicttranslcolor: {{.colorB6}};

	--invisible: hsla(0, 100%, 100%, 0);
	--transparentgrey: hsla(0, 0%, 50%, .5);
	--textshadowcolor: {{.textshadowColor}};

	/* R */
    --hgscol13: {{.colorC2}};
    --hgscol04: {{.colorB4}};
    --hgscol12: {{.colorC4}};
    --hgscol16: {{.colorB4}}; 
    --hgscol17: {{.colorD5}}; 
    --hgscol02: {{.colorC0}};
    --hgscol14: {{.colorA1}};

	/* G */
    --hgscol18: {{.colorC0}};
    --hgscol15: {{.colorC0}};
    --hgscol20: {{.colorC4}};
    --hgscol07: {{.colorB6}};
    --hgscol08: {{.colorB6}};
    --hgscol21: {{.colorC4}};

	/* B */
    --hgscol11: {{.colorB2}};
    --hgscol19: {{.colorB5}};
    --hgscol03: {{.colorC5}};
    --hgscol01: {{.colorB4}};
    --hgscol05: {{.colorC6}}; 
    --hgscol06: {{.colorB4}};  

    --hgscol09: {{.colorC3}};
    --hgscol10: {{.colorB0}};

	/* b/w */
	--hgsmono09: {{.colorA6}};
	--hgsmono05: {{.colorB5}};
	--hgsmono08: {{.colorC5}};
	--hgsmono02: {{.colorA1}};
	--hgsmono03: {{.colorB4}};
	--hgsmono04: {{.colorC4}};
	--hgsmono01: {{.colorD4}};
	--hgsmono07: {{.colorE2}};
	--hgsmono00: {{.colorA0}};
`

	// TRIADIC - only A, B, C available; the geometry is *very* similar to splitcomp...
	TRIADIC = `
    /* TRIADIC: only A, B, C */
    --main-body-color: {{.colorA0}};
    --main-font-color: {{.colorA6}};
    --input-border-color: {{.colorB3}};
    
    --buttoncolor: {{.colorB3}};
    --button-hover: {{.colorB5}};
    
    --fieldset-background: {{.colorA2}};
    --focus-shadow: rgba(0, 0, 0, 0.5);
    --icons-color: {{.colorB5}};

	--greyborders: {{.colorB5}};
	--boxshadows: {{.colorA4}};
	--dropdown-backgrounds: {{.colorC1}};
	--modalbackground: {{.colorA0}};
	--modalheader: {{.colorB3}};
	--sidednavcolor: {{.colorA1}};
	--nthrowcol: {{.colorA2}};
	--buttonandcheckboxborder: {{.colorA4}};
	--buttonandcheckboxcontents: {{.colorB4}};

	--locuscolor: {{.colorC4}};
	--notations: {{.colorB6}};
	--dottedborders: {{.colorA6}};
	--bracketcolor1: {{.colorA5}};
	--bracketcolor2: {{.colorB5}};
	--bracketcolor3: {{.colorC5}};
	--bracketcolor4: {{.colorA6}};
	
	--dictauthwkcolor: {{.colorC5}};
	--dictbiblcolor: {{.colorB5}};
	--dictdefaultcolor: {{.colorA6}};
	--dictlevelscolor: {{.colorB5}};
	--dictquotecolor: {{.colorA6}};
	--dicttranslcolor: {{.colorB6}};

	--invisible: hsla(0, 100%, 100%, 0);
	--transparentgrey: hsla(0, 0%, 50%, .5);
	--textshadowcolor: {{.textshadowColor}};

	/* R */
    --hgscol13: {{.colorB1}};
    --hgscol04: {{.colorB4}};
    --hgscol12: {{.colorB4}};
    --hgscol16: {{.colorB5}}; 
    --hgscol17: {{.colorB5}}; 
    --hgscol02: {{.colorB6}};
    --hgscol14: {{.colorB6}};

	/* G */
    --hgscol18: {{.colorA4}};
    --hgscol15: {{.colorA4}};
    --hgscol20: {{.colorA5}};
    --hgscol07: {{.colorA5}};
    --hgscol08: {{.colorA6}};
    --hgscol21: {{.colorB5}};

	/* B */
    --hgscol11: {{.colorC1}};
    --hgscol19: {{.colorC5}};
    --hgscol03: {{.colorC6}};
    --hgscol01: {{.colorC4}};
    --hgscol05: {{.colorC5}}; 
    --hgscol06: {{.colorC6}};  

    --hgscol09: {{.colorB5}};
    --hgscol10: {{.colorA0}};

	/* b/w */
	--hgsmono09: {{.colorA6}};
	--hgsmono05: {{.colorB6}};
	--hgsmono08: {{.colorC6}};
	--hgsmono02: {{.colorA5}};
	--hgsmono03: {{.colorB4}};
	--hgsmono04: {{.colorC5}};
	--hgsmono01: {{.colorA4}};
	--hgsmono07: {{.colorB5}};
	--hgsmono00: {{.colorC6}};
`
)
