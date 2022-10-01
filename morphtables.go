//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

const (
	DECLTABTMPL = `
	<table class="verbanalysis">
	<tbody>
	{{.Header}}
	{{.Rows}}
	</tbody>
	</table>
	<hr class="styled">
`
	DECLHEADERTMPL = `
	<tr align="center">
		<td rowspan="1" colspan="{{.Span}}" class="dialectlabel">{{.Dialect}}<br>
		</td>
	</tr>
	{{.Gendlabel}}
	</tr>
`
	DECLGENDLABEL = `<tr>
		<td class="genderlabel">&nbsp;</td>
		{{.AllGenders}}
	</tr>`

	DECLBLANK = `
	<tr><td>&nbsp;</td>{{.Columns}}</tr>
`
	DECLGENDERCELL = `<td class="gendercell">{{.G}}<br></td>`

	DECLMORPHROW = `	
	<tr class="morphrow">
		{{.AllCells}}
	</tr>`

	MORPHLABELCELL = `<td class="morphlabelcell">{{.Ml}}</td>`
	MORPHCELL      = `<td class="morphcell">{{.Mo}}</td>`

	// 	Smythe §383ff
	//
	//	verb cells look like:
	//
	//		<td class="morphcell">_attic_subj_pass_pl_2nd_pres_</td>

	VBTABTMPL = `
	<table class="verbanalysis">
	<tbody>
	{{.Header}}
	{{.Rows}}
	</tbody>
	</table>
	<hr class="styled">`

	VBHEADERTMPL = `	
	<tr align="center">
		<td rowspan="1" colspan="{s}" class="dialectlabel">{dialect}<br>
		</td>
	</tr>
	<tr align="center">
		<td rowspan="1" colspan="{s}" class="voicelabel">{voice}<br>
		</td>
	</tr>
	<tr align="center">
		<td rowspan="1" colspan="{s}" class="moodlabel">{mood}<br>
		</td>
	{{.Tenseheader}}
	</tr>`

	VBTENSETEMPL = `
	<tr>
		<td class="tenselabel">&nbsp;</td>
		{{.Alltenses}}
	</tr>`

	VBBLANK      = DECLBLANK
	VMMORPHROW   = DECLMORPHROW
	VBREGEXTEMPL = `_{{.D}}_{{.M}}_{{.V}}_{{.N}}_{{.P}}_{{.T}}_`
	PCPLTEMPL    = `_{{.D}}_{{.M}}_{{.V}}_{{.N}}_{{.T}}_{{.G}}_{{.C}}_`
)

var (
	GKCASES  = []string{"nom", "gen", "dat", "acc", "voc"}
	GKNUMB   = []string{"sg", "dual", "pl"}
	GKMOODS  = []string{"ind", "subj", "opt", "imperat", "inf", "part"}
	GKVOICE  = []string{"act", "mid", "pass"}
	GKTENSES = map[int]string{1: "Present", 2: "Imperfect", 3: "Future", 4: "Aorist", 5: "Perfect", 6: "Pluperfect", 7: "Future Perfect"}
	GKVERBS  = getgkvbmap()
	LTCASES  = []string{"nom", "gen", "dat", "acc", "abl", "voc"}
	LTNUMB   = []string{"sg", "pl"}
	LTMOODS  = []string{"ind", "subj", "imperat", "inf", "part", "gerundive", "supine"}
	LTVOICES = []string{"act", "pass"}
	LTTENSES = map[int]string{1: "Present", 2: "Imperfect", 3: "Future", 5: "Perfect", 6: "Pluperfect", 7: "Future Perfect"}
	LTVERBS  = getltvbmap()
	GENDERS  = []string{"masc", "fem", "neut"}
	PERSONS  = []string{"1st", "2nd", "3rd"}
)

func getgkvbmap() map[string]map[string]map[int]bool {
	gvm := make(map[string]map[string]map[int]bool)
	gvm["act"]["ind"] = map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: false}
	gvm["act"]["subj"] = map[int]bool{1: true, 2: false, 3: false, 4: true, 5: true, 6: false, 7: false}
	gvm["act"]["opt"] = map[int]bool{1: true, 2: false, 3: true, 4: true, 5: true, 6: false, 7: false}
	gvm["act"]["imperat"] = map[int]bool{1: true, 2: false, 3: false, 4: true, 5: true, 6: false, 7: false}
	gvm["act"]["inf"] = map[int]bool{1: true, 2: false, 3: true, 4: true, 5: true, 6: false, 7: false}
	gvm["act"]["part"] = map[int]bool{1: true, 2: false, 3: true, 4: true, 5: true, 6: false, 7: false}
	gvm["mid"]["ind"] = map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: false}
	gvm["mid"]["subj"] = map[int]bool{1: true, 2: false, 3: false, 4: true, 5: true, 6: false, 7: false}
	gvm["mid"]["opt"] = map[int]bool{1: true, 2: false, 3: true, 4: true, 5: true, 6: false, 7: false}
	gvm["mid"]["imperat"] = map[int]bool{1: true, 2: false, 3: false, 4: true, 5: true, 6: false, 7: false}
	gvm["mid"]["inf"] = map[int]bool{1: true, 2: false, 3: true, 4: true, 5: true, 6: false, 7: false}
	gvm["mid"]["part"] = map[int]bool{1: true, 2: false, 3: true, 4: true, 5: true, 6: false, 7: false}
	gvm["pass"]["ind"] = map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true}
	gvm["pass"]["subj"] = map[int]bool{1: true, 2: false, 3: false, 4: true, 5: true, 6: false, 7: false}
	gvm["pass"]["opt"] = map[int]bool{1: true, 2: false, 3: true, 4: true, 5: true, 6: false, 7: true}
	gvm["pass"]["imperat"] = map[int]bool{1: true, 2: false, 3: false, 4: true, 5: true, 6: false, 7: false}
	gvm["pass"]["inf"] = map[int]bool{1: true, 2: false, 3: true, 4: true, 5: true, 6: false, 7: true}
	gvm["pass"]["part"] = map[int]bool{1: true, 2: false, 3: true, 4: true, 5: true, 6: false, 7: true}
	return gvm
}

func getltvbmap() map[string]map[string]map[int]bool {
	// note that ppf subj pass, etc are "false" because "laudātus essem" is not going to be found
	lvm := make(map[string]map[string]map[int]bool)
	lvm["act"]["ind"] = map[int]bool{1: true, 2: true, 3: true, 5: true, 6: true, 7: true}
	lvm["act"]["subj"] = map[int]bool{1: true, 2: false, 3: false, 5: true, 6: true, 7: false}
	lvm["act"]["imperat"] = map[int]bool{1: true, 2: false, 3: true, 5: false, 6: false, 7: false}
	lvm["act"]["inf"] = map[int]bool{1: true, 2: false, 3: false, 5: true, 6: false, 7: false}
	lvm["act"]["part"] = map[int]bool{1: true, 2: false, 3: true, 5: false, 6: false, 7: false}
	lvm["pass"]["ind"] = map[int]bool{1: true, 2: true, 3: true, 5: false, 6: false, 7: false}
	lvm["pass"]["subj"] = map[int]bool{1: true, 2: true, 3: false, 5: false, 6: false, 7: false}
	lvm["pass"]["imperat"] = map[int]bool{1: true, 2: false, 3: true, 5: false, 6: false, 7: false}
	lvm["pass"]["inf"] = map[int]bool{1: true, 2: false, 3: false, 5: false, 6: false, 7: false}
	lvm["pass"]["part"] = map[int]bool{1: false, 2: false, 3: false, 5: true, 6: false, 7: false}
	return lvm
}
