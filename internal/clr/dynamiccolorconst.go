package clr

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
	--notations: {{.colorA6}};
	--dottedborders: {{.colorA6}};
	--bracketcolor1: {{.colorA2}};
	--bracketcolor2: {{.colorA2}};
	--bracketcolor3: {{.colorA2}};
	--bracketcolor4: {{.colorA2}};
	
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
	--hgscol11: {{.colorA3}};
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

	SPLITCOMP = `
    --main-body-color: {{.colorA1}};
    --main-font-color: {{.colorA6}};
    --input-border-color: {{.colorA2}};
    
    --buttoncolor: {{.colorA3}};
    --button-hover: {{.colorA4}};
    
    --fieldset-background: {{.colorA2}};
    --focus-shadow: rgba(0, 0, 0, 0.5);
    --icons-color: {{.colorA4}};

    --hgsmono00: {{.colorA0}};
    --hgscol01: {{.colorC0}}; 
    --hgscol02: {{.colorC0}};
    --hgscol03: {{.colorC0}};
    --hgscol04: {{.colorA6}};
    --hgscol05: {{.colorB0}}; 
    --hgscol06:  {{.colorC0}};  
    --hgscol07: {{.colorA6}};
    --hgsmono01: {{.colorC0}};
    --hgscol08: {{.colorA6}};
    --hgscol09: {{.colorA6}};
    --invisible: hsla(0, 100%, 0%, 0);
    --hgscol10: {{.colorA2}};
    --hgscol11: {{.colorA3}};
    --hgsmono02: {{.colorA2}};
    --hgsmono03: {{.colorC0}};
    --hgsmono04: {{.colorC0}};
    --hgsmono05: {{.colorA6}};
    --hgscol12: {{.colorC0}};
    --hgscol13: {{.colorC0}};
    --hgscol14: {{.colorA1}};
    --hgscol15: {{.colorC0}};
    --hgscol16: {{.colorC0}}; 
    --hgscol07: {{.colorA6}};
    --hgscol17: {{.colorA4}}; 
    --hgscol18: {{.colorC0}};
    --hgscol19: {{.colorC0}};
    --hgscol20: {{.colorC0}};
    --transparentgrey: hsla(0, 0%, 10%, .7);
    --hgscol21: {{.colorB0}};
    --hgsmono07: {{.colorA6}};
    --hgsmono08: {{.colorB0}};
    --hgsmono09: {{.colorA6}};`

	TETRADTMPL = `
    --main-body-color: {{.colorA1}};
    --main-font-color: {{.colorA6}};
    --input-border-color: {{.colorA2}};
    
    --buttoncolor: {{.colorA3}};
    --button-hover: {{.colorA5}};
    
    --fieldset-background: {{.colorA2}};
    --focus-shadow: rgba(0, 0, 0, 0.5);
    --icons-color: {{.colorA4}};
    
    --hgsmono00: {{.colorA0}};
    --hgscol01: {{.colorC0}}; 
    --hgscol02: {{.colorD0}};
    --hgscol03: {{.colorC0}};
    --hgscol04: {{.colorA6}};
    --hgscol05: {{.colorB0}}; 
    --hgscol06:  {{.colorC0}};  
    --hgscol07: {{.colorA6}};
    --hgsmono01: {{.colorC0}};
    --hgscol08: {{.colorA6}};
    --hgscol09: {{.colorD0}};
    --invisible: hsla(0, 100%, 0%, 0);
    --hgscol10: {{.colorA2}};
    --hgscol11: {{.colorA3}};
    --hgsmono02: {{.colorA2}};
    --hgsmono03: {{.colorC0}};
    --hgsmono04: {{.colorC0}};
    --hgsmono05: {{.colorA6}};
    --hgscol12: {{.colorD0}};
    --hgscol13: {{.colorD0}};
    --hgscol14: {{.colorA1}};
    --hgscol15: {{.colorC0}};
    --hgscol16: {{.colorD0}}; 
    --hgscol17: {{.colorA4}}; 
    --hgscol18: {{.colorD0}};
    --hgscol19: {{.colorD0}};
    --hgscol20: {{.colorD0}};
    --transparentgrey: hsla(0, 0%, 10%, .7);
    --hgscol21: {{.colorB0}};
    --hgsmono07: {{.colorD0}};
    --hgsmono08: {{.colorB0}};
    --hgsmono09: {{.colorA6}};`

	TRIADIC = `
    --main-body-color: {{.colorA1}};
    --main-font-color: {{.colorA6}};
    --input-border-color: {{.colorA2}};
    
    --buttoncolor: {{.colorA3}};
    --button-hover: {{.colorA5}};
    
    --fieldset-background: {{.colorA2}};
    --focus-shadow: rgba(0, 0, 0, 0.5);
    --icons-color: {{.colorA4}};
    
    --hgsmono00: {{.colorA0}};
    --hgscol01: {{.colorC0}}; 
    --hgscol02: {{.colorC0}};
    --hgscol03: {{.colorA6}};
    --hgscol04: {{.colorA6}};
    --hgscol05: {{.colorB0}}; 
    --hgscol06:  {{.colorA6}};  
    --hgscol07: {{.colorA6}};
    --hgsmono01: {{.colorA6}};
    --hgscol08: {{.colorA6}};
    --hgscol09: {{.colorB0}};
    --invisible: hsla(0, 100%, 0%, 0);
    --hgscol10: {{.colorA2}};
    --hgscol11: {{.colorA3}};
    --hgsmono02: {{.colorA2}};
    --hgsmono03: {{.colorA6}};
    --hgsmono04: {{.colorA6}};
    --hgsmono05: {{.colorA6}};
    --hgscol12: {{.colorA6}};
    --hgscol13: {{.colorA6}};
    --hgscol14: {{.colorA1}};
    --hgscol15: {{.colorA6}};
    --hgscol16: {{.colorA6}}; 
    --hgscol07: {{.colorA6}};
    --hgscol17: {{.colorA4}}; 
    --hgscol18: {{.colorA6}};
    --hgscol19: {{.colorA6}};
    --hgscol20: {{.colorA6}};
    --transparentgrey: hsla(0, 0%, 10%, .7);
    --hgscol21: {{.colorB0}};
    --hgsmono07: {{.colorB0}};
    --hgsmono08: {{.colorB0}};
    --hgsmono09: {{.colorA6}};`
)
