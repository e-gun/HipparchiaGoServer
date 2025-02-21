package clr

const (
	MONO = `
    --main-body-color: {{.colorA5}};
    --main-font-color: {{.colorA0}};
    --input-border-color: {{.colorA4}};
    
    --buttoncolor: {{.colorA6}};
    --button-hover: {{.colorA4}};
    
    --fieldset-background: {{.colorA4}};
    --focus-shadow: rgba(0, 0, 0, 0.5);
    --icons-color: {{.colorA1}};
    
    --black: {{.colorA6}};
    --blue: {{.colorA6}};
    --brown: {{.colorA6}};
    --brtblue: {{.colorA2}};
    --copper: {{.colorA0}};
    --deepblue: {{.colorA6}};
    --dkbabyblue: {{.colorA6}};
    --dkgreen: {{.colorA0}};
    --dkgrey: {{.colorA0}};
    --dkteal: {{.colorA0}};
    --huedgrey: {{.colorA6}};
    --invisible: hsla(0, 100%, 0%, 0);
    --lessoffwhite: {{.colorA4}};
    --ltbabyblue: {{.colorA3}};
    --ltgrey: {{.colorA6}};
    --ltgrey2: {{.colorA6}};
    --midgrey: {{.colorA6}};
    --offwhite: {{.colorA6}};
    --orange: {{.colorA3}};
    --pink: {{.colorA3}};
    --pinker: {{.colorA3}};
    --plum: {{.colorA5}};
    --pukegreen: {{.colorA6}};
    --red: {{.colorA2}};
    --dkgreen: {{.colorA5}};
    --rustedorange: {{.colorA0}};
    --sicklyyellow: {{.colorA3}};
    --skyblue: {{.colorA5}};
    --teal: {{.colorA3}};
    --transparentgrey: hsla(0, 0%, 10%, .7);
    --vdkteal: {{.colorA1}};
    --vdkgrey: {{.colorA6}};
    --vltgrey: {{.colorA6}};
    --white: {{.colorA0}};`

	TRIADIC = `
    --main-body-color: {{.colorA1}};
    --main-font-color: {{.colorA6}};
    --input-border-color: {{.colorA2}};
    
    --buttoncolor: {{.colorA3}};
    --button-hover: {{.colorA5}};
    
    --fieldset-background: {{.colorA2}};
    --focus-shadow: rgba(0, 0, 0, 0.5);
    --icons-color: {{.colorA4}};
    
    --black: {{.colorA0}};
    --blue: {{.colorC0}}; 
    --brown: {{.colorC0}};
    --brtblue: {{.colorA6}};
    --copper: {{.colorA6}};
    --deepblue: {{.colorB0}}; 
    --dkbabyblue:  {{.colorA6}};  
    --dkgreen: {{.colorA6}};
    --dkgrey: {{.colorA6}};
    --dkteal: {{.colorA6}};
    --huedgrey: {{.colorB0}};
    --invisible: hsla(0, 100%, 0%, 0);
    --lessoffwhite: {{.colorA2}};
    --ltbabyblue: {{.colorA3}};
    --ltgrey: {{.colorA2}};
    --ltgrey2: {{.colorA6}};
    --midgrey: {{.colorA6}};
    --offwhite: {{.colorA6}};
    --orange: {{.colorA6}};
    --pink: {{.colorA6}};
    --pinker: {{.colorA3}};
    --plum: {{.colorA1}};
    --pukegreen: {{.colorA6}};
    --red: {{.colorA6}}; 
    --dkgreen: {{.colorA6}};
    --rustedorange: {{.colorA4}}; 
    --sicklyyellow: {{.colorA6}};
    --skyblue: {{.colorA6}};
    --teal: {{.colorA6}};
    --transparentgrey: hsla(0, 0%, 10%, .7);
    --vdkteal: {{.colorB0}};
    --vdkgrey: {{.colorB0}};
    --vltgrey: {{.colorB0}};
    --white: {{.colorA6}};`

	TETRADTMPL = `
    --main-body-color: {{.colorA1}};
    --main-font-color: {{.colorA6}};
    --input-border-color: {{.colorA2}};
    
    --buttoncolor: {{.colorA3}};
    --button-hover: {{.colorA5}};
    
    --fieldset-background: {{.colorA2}};
    --focus-shadow: rgba(0, 0, 0, 0.5);
    --icons-color: {{.colorA4}};
    
    --black: {{.colorA0}};
    --blue: {{.colorC0}}; 
    --brown: {{.colorD0}};
    --brtblue: {{.colorC0}};
    --copper: {{.colorA6}};
    --deepblue: {{.colorB0}}; 
    --dkbabyblue:  {{.colorC0}};  
    --dkgreen: {{.colorA6}};
    --dkgrey: {{.colorC0}};
    --dkteal: {{.colorA6}};
    --huedgrey: {{.colorD0}};
    --invisible: hsla(0, 100%, 0%, 0);
    --lessoffwhite: {{.colorA2}};
    --ltbabyblue: {{.colorA3}};
    --ltgrey: {{.colorA2}};
    --ltgrey2: {{.colorC0}};
    --midgrey: {{.colorC0}};
    --offwhite: {{.colorA6}};
    --orange: {{.colorD0}};
    --pink: {{.colorD0}};
    --pinker: {{.colorA3}};
    --plum: {{.colorA1}};
    --pukegreen: {{.colorC0}};
    --red: {{.colorD0}}; 
    --dkgreen: {{.colorA6}};
    --rustedorange: {{.colorA4}}; 
    --sicklyyellow: {{.colorD0}};
    --skyblue: {{.colorD0}};
    --teal: {{.colorD0}};
    --transparentgrey: hsla(0, 0%, 10%, .7);
    --vdkteal: {{.colorB0}};
    --vdkgrey: {{.colorD0}};
    --vltgrey: {{.colorB0}};
    --white: {{.colorA6}};`

	SPLITCOMP = `
    --main-body-color: {{.colorA1}};
    --main-font-color: {{.colorA6}};
    --input-border-color: {{.colorA2}};
    
    --buttoncolor: {{.colorA3}};
    --button-hover: {{.colorA4}};
    
    --fieldset-background: {{.colorA2}};
    --focus-shadow: rgba(0, 0, 0, 0.5);
    --icons-color: {{.colorA4}};

    --black: {{.colorA0}};
    --blue: {{.colorC0}}; 
    --brown: {{.colorC0}};
    --brtblue: {{.colorC0}};
    --copper: {{.colorA6}};
    --deepblue: {{.colorB0}}; 
    --dkbabyblue:  {{.colorC0}};  
    --dkgreen: {{.colorA6}};
    --dkgrey: {{.colorC0}};
    --dkteal: {{.colorA6}};
    --huedgrey: {{.colorA6}};
    --invisible: hsla(0, 100%, 0%, 0);
    --lessoffwhite: {{.colorA2}};
    --ltbabyblue: {{.colorA3}};
    --ltgrey: {{.colorA2}};
    --ltgrey2: {{.colorC0}};
    --midgrey: {{.colorC0}};
    --offwhite: {{.colorA6}};
    --orange: {{.colorC0}};
    --pink: {{.colorC0}};
    --pinker: {{.colorA3}};
    --plum: {{.colorA1}};
    --pukegreen: {{.colorC0}};
    --red: {{.colorC0}}; 
    --dkgreen: {{.colorA6}};
    --rustedorange: {{.colorA4}}; 
    --sicklyyellow: {{.colorC0}};
    --skyblue: {{.colorC0}};
    --teal: {{.colorC0}};
    --transparentgrey: hsla(0, 0%, 10%, .7);
    --vdkteal: {{.colorB0}};
    --vdkgrey: {{.colorA6}};
    --vltgrey: {{.colorB0}};
    --white: {{.colorA6}};`
)
