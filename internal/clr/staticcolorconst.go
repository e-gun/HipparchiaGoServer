package clr

const (
	DEFAULTCOLORSCHEME = "Light"

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
	--modalbackground: hsla(0, 0%, 98%, 1);
	--modalheader: hsla(0, 0%, 67%, 1);
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
	--dictbiblcolor: hsla(237, 43%, 57%, 1);
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
	--modalbackground: hsla(0, 0%, 0%, 1);
	--modalheader: hsl(206, 7%, 11%);
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
	--dictbiblcolor: hsla(64, 52%, 84%, 1);
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
)
