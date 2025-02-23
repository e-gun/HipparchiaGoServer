package clr

// if the body is a '5' or '6', then it is safest to keep most of the other colors ina the 2-0 range...

// gotacha colors:
// .dictquote is --hgsmono01
// .dicttrans is --hgsmono07
// modal backgrounds are --hgscol10
// morphcell is --hgscol03
// nthrow is --hgsmono02
// sought is --hgscol05
// resultlocus --hgscol09

const (
	MONO = `
	--main-body-color: {{.colorA5}};
	--main-font-color: {{.colorA0}};
	--input-border-color: {{.colorA4}};
	
	--buttoncolor: {{.colorA6}};
	--button-hover: {{.colorA3}};
	
	--fieldset-background: {{.colorA4}};
	--focus-shadow: rgba(0, 0, 0, 0.5);
	--icons-color: {{.colorA1}};
	
	--greyborders: {{.colorA3}};
	--boxshadows: {{.colorA4}};
	--modalbackground: {{.colorA2}};
	--sidednavcolor: {{.colorA6}};
	--nthrowcol: {{.colorA6}};

	--locuscolor: {{.colorA6}};
	--notations: {{.colorA6}};
	--dottedborders: {{.colorA6}};
	--bracketcolor1: {{.colorA2}};
	--bracketcolor2: {{.colorA2}};
	--bracketcolor3: {{.colorA2}};
	--bracketcolor4: {{.colorA2}};
	
	--dictauthwkcolor: {{.colorA6}};
	--dictdefaultcolor: {{.colorA6}};
	--dictlevelscolor: {{.colorA1}};
	--dictquotecolor: {{.colorA1}};
	--dicttranslcolor: {{.colorA6}};

	--invisible: hsla(0, 100%, 100%, 0);
	--transparentgrey: hsla(0, 0%, 67%, .8);
	
	--hgscol01: {{.colorA6}};
	--hgscol02: {{.colorA6}};
	--hgscol03: {{.colorA1}};
	--hgscol04: {{.colorA0}};
	--hgscol05: {{.colorA6}};
	--hgscol06: {{.colorA6}};
	--hgscol07: {{.colorA0}};
	--hgscol07: {{.colorA5}};
	--hgscol08: {{.colorA0}};
	--hgscol09: {{.colorA6}};
	--hgscol10: {{.colorA4}};
	--hgscol11: {{.colorA4}};
	--hgscol12: {{.colorA3}};
	--hgscol13: {{.colorA3}};
	--hgscol14: {{.colorA5}};
	--hgscol15: {{.colorA5}};
	--hgscol16: {{.colorA2}};
	--hgscol17: {{.colorA0}};
	--hgscol18: {{.colorA3}};
	--hgscol19: {{.colorA5}};
	--hgscol20: {{.colorA3}};
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
	
	--invisible: hsla(0, 100%, 0%, 0);
	--transparentgrey: hsla(0, 0%, 10%, .7);`

	// SPLITCOMP - only A, B, C available; the geometry is *very* similar to triadic...
	SPLITCOMP = `
    --main-body-color: {{.colorA1}};
    --main-font-color: {{.colorA6}};
    --input-border-color: {{.colorA2}};
    
    --buttoncolor: {{.colorA3}};
    --button-hover: {{.colorA4}};
    
    --fieldset-background: {{.colorA2}};
    --focus-shadow: rgba(0, 0, 0, 0.5);
    --icons-color: {{.colorA4}};

	--greyborders: {{.colorA3}};
	--boxshadows: {{.colorA4}};
	--modalbackground: {{.colorB5}};
	--sidednavcolor: {{.colorA2}};
	--nthrowcol: {{.colorA6}};

	--locuscolor: {{.colorA6}};
	--notations: {{.colorB6}};
	--dottedborders: {{.colorA6}};
	--bracketcolor1: {{.colorA5}};
	--bracketcolor2: {{.colorB5}};
	--bracketcolor3: {{.colorC5}};
	--bracketcolor4: {{.colorA6}};
	
	--dictauthwkcolor: {{.colorC5}};
	--dictdefaultcolor: {{.colorA6}};
	--dictlevelscolor: {{.colorA1}};
	--dictquotecolor: {{.colorA6}};
	--dicttranslcolor: {{.colorB6}};

	--invisible: hsla(0, 100%, 100%, 0);
	--transparentgrey: hsla(0, 0%, 67%, .8);

	/* R */
    --hgscol13: {{.colorB1}};
    --hgscol04: {{.colorB2}};
    --hgscol12: {{.colorB3}};
    --hgscol16: {{.colorB3}}; 
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
    --hgscol19: {{.colorC4}};
    --hgscol03: {{.colorC5}};
    --hgscol01: {{.colorC4}};
    --hgscol05: {{.colorC5}}; 
    --hgscol06: {{.colorC6}};  

    --hgscol09: {{.colorA6}};
    --hgscol10: {{.colorA0}};

	/* b/w */
	--hgsmono09: {{.colorA6}};
	--hgsmono05: {{.colorB6}};
	--hgsmono08: {{.colorC6}};
	--hgsmono02: {{.colorA3}};
	--hgsmono03: {{.colorB5}};
	--hgsmono04: {{.colorC5}};
	--hgsmono01: {{.colorA5}};
	--hgsmono07: {{.colorB5}};
	--hgsmono00: {{.colorC6}};

	--invisible: hsla(0, 100%, 0%, 0);
	--transparentgrey: hsla(0, 0%, 10%, .7);`

	TETRADTMPL = `
    --main-body-color: {{.colorA0}};
    --main-font-color: {{.colorA6}};
    --input-border-color: {{.colorA2}};
    
    --buttoncolor: {{.colorB3}};
    --button-hover: {{.colorB1}};
    
    --fieldset-background: {{.colorA2}};
    --focus-shadow: rgba(0, 0, 0, 0.5);
    --icons-color: {{.colorA4}};
    
 	--greyborders: {{.colorA3}};
	--boxshadows: {{.colorA4}};
	--modalbackground: {{.colorC2}};
	--sidednavcolor: {{.colorA2}};
	--nthrowcol: {{.colorA6}};

	--locuscolor: {{.colorA6}};
	--notations: {{.colorB5}};
	--dottedborders: {{.colorA6}};
	--bracketcolor1: {{.colorA5}};
	--bracketcolor2: {{.colorB5}};
	--bracketcolor3: {{.colorC5}};
	--bracketcolor4: {{.colorA6}};
	
	--dictauthwkcolor: {{.colorC5}};
	--dictdefaultcolor: {{.colorA6}};
	--dictlevelscolor: {{.colorA1}};
	--dictquotecolor: {{.colorA6}};
	--dicttranslcolor: {{.colorB6}};

	--invisible: hsla(0, 100%, 100%, 0);
	--transparentgrey: hsla(0, 0%, 67%, .8);

	/* R */
    --hgscol13: {{.colorC2}};
    --hgscol04: {{.colorB4}};
    --hgscol12: {{.colorC0}};
    --hgscol16: {{.colorB4}}; 
    --hgscol17: {{.colorA4}}; 
    --hgscol02: {{.colorC0}};
    --hgscol14: {{.colorA1}};

	/* G */
    --hgscol18: {{.colorC0}};
    --hgscol15: {{.colorC0}};
    --hgscol20: {{.colorC4}};
    --hgscol07: {{.colorB6}};
    --hgscol08: {{.colorB6}};
    --hgscol21: {{.colorB0}};

	/* B */
    --hgscol11: {{.colorB2}};
    --hgscol19: {{.colorB1}};
    --hgscol03: {{.colorC5}};
    --hgscol01: {{.colorB4}};
    --hgscol05: {{.colorC6}}; 
    --hgscol06: {{.colorB4}};  

    --hgscol09: {{.colorC5}};
    --hgscol10: {{.colorB0}};

	/* b/w */
	--hgsmono09: {{.colorA6}};
	--hgsmono05: {{.colorB5}};
	--hgsmono08: {{.colorC5}};
	--hgsmono02: {{.colorA1}};
	--hgsmono03: {{.colorB4}};
	--hgsmono04: {{.colorC4}};
	--hgsmono01: {{.colorB2}};
	--hgsmono07: {{.colorC2}};
	--hgsmono00: {{.colorA0}};

	--invisible: hsla(0, 100%, 0%, 0);
	--transparentgrey: hsla(0, 0%, 10%, .7);`

	// TRIADIC - only A, B, C available; the geometry is *very* similar to splitcomp...
	TRIADIC = `
    /* TRIADIC: only A, B, C */
    --main-body-color: {{.colorA1}};
    --main-font-color: {{.colorA6}};
    --input-border-color: {{.colorA2}};
    
    --buttoncolor: {{.colorA3}};
    --button-hover: {{.colorA4}};
    
    --fieldset-background: {{.colorA2}};
    --focus-shadow: rgba(0, 0, 0, 0.5);
    --icons-color: {{.colorA4}};

	--greyborders: {{.colorA3}};
	--boxshadows: {{.colorA4}};
	--modalbackground: {{.colorB5}};
	--sidednavcolor: {{.colorA2}};
	--nthrowcol: {{.colorB1}};

	--locuscolor: {{.colorC6}};
	--notations: {{.colorB6}};
	--dottedborders: {{.colorA6}};
	--bracketcolor1: {{.colorA5}};
	--bracketcolor2: {{.colorB5}};
	--bracketcolor3: {{.colorC5}};
	--bracketcolor4: {{.colorA6}};
	
	--dictauthwkcolor: {{.colorC5}};
	--dictdefaultcolor: {{.colorA6}};
	--dictlevelscolor: {{.colorB5}};
	--dictquotecolor: {{.colorA6}};
	--dicttranslcolor: {{.colorB6}};

	--invisible: hsla(0, 100%, 100%, 0);
	--transparentgrey: hsla(0, 0%, 67%, .8);

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
    --hgscol19: {{.colorC4}};
    --hgscol03: {{.colorC6}};
    --hgscol01: {{.colorC4}};
    --hgscol05: {{.colorC5}}; 
    --hgscol06: {{.colorC6}};  

    --hgscol09: {{.colorA6}};
    --hgscol10: {{.colorA0}};

	/* b/w */
	--hgsmono09: {{.colorA6}};
	--hgsmono05: {{.colorB6}};
	--hgsmono08: {{.colorC6}};
	--hgsmono02: {{.colorA5}};
	--hgsmono03: {{.colorB5}};
	--hgsmono04: {{.colorC5}};
	--hgsmono01: {{.colorA5}};
	--hgsmono07: {{.colorB5}};
	--hgsmono00: {{.colorC6}};

	--invisible: hsla(0, 100%, 0%, 0);
	--transparentgrey: hsla(0, 0%, 10%, .7);`
)
