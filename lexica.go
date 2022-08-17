package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4"
	"reflect"
	"regexp"
	"strings"
)

// initial target: be able to respond to "GET /lexica/findbyform/ἀμιϲθὶ/gr0062 HTTP/1.1"
// full set of verbs: lookup, findbyform, idlookup, morphologychart

func findbyform(word string, author string) []byte {

	// [a] clean the search term

	// TODO...

	// [b] pick a dictionary

	// naive and could/should be improved

	var d string
	if author[0:2] != "lt" {
		d = "greek"
	} else {
		d = "latin"
	}

	// [c] search for morphology matches

	// the python is funky because we need to poke at words several times and to try combinations of fixes
	// skipping that stuff here for now because 'findbyform' should usually see a known form
	// TODO: accute/grave issues should be handled ASAP

	dbpool := grabpgsqlconnection()
	fld := `observed_form, xrefs, prefixrefs, possible_dictionary_forms, related_headwords`
	psq := fmt.Sprintf("SELECT %s FROM %s_morphology WHERE observed_form = '%s'", fld, d, word)

	var foundrows pgx.Rows
	var err error

	foundrows, err = dbpool.Query(context.Background(), psq)
	checkerror(err)

	var thesefinds []DbMorphology
	defer foundrows.Close()
	for foundrows.Next() {
		var thehit DbMorphology
		err := foundrows.Scan(&thehit.Observed, &thehit.Xrefs, &thehit.PrefixXrefs, &thehit.RawPossib, &thehit.RelatedHW)
		checkerror(err)
		thesefinds = append(thesefinds, thehit)
	}

	// [c1] turn morph matches into []MorphPossib

	var mpp []MorphPossib

	for _, h := range thesefinds {
		// RawPossib is JSON + JSON; nested JSON is a PITA, but the structure is: {"1": {...}, "2": {...}, ...}
		// that is splittable
		// just need to clean the '}}' at the end

		boundary := regexp.MustCompile(`(\{|, )"\d": `)
		possible := boundary.Split(h.RawPossib, -1)

		for _, p := range possible {
			// fmt.Println(p)
			p = strings.Replace(p, "}}", "}", -1)
			p = strings.TrimSpace(p)
			var mp MorphPossib
			if len(p) > 0 {
				err := json.Unmarshal([]byte(p), &mp)
				checkerror(err)
			}
			mpp = append(mpp, mp)
		}
	}

	// fmt.Println(mpp)

	// [d] take the []MorphPossib and find the set of headwords we are interested in

	var hwm []string
	for _, p := range mpp {
		if len(p.Headwd) > 0 {
			hwm = append(hwm, p.Headwd)
		}
	}

	// the next is primed to produce problems: see καρποῦ which will turn καρπόϲ1 and καρπόϲ2 into just καρπόϲ; need xref_value?
	// but we have probably taken care of this below: see the comments
	hwm = unique(hwm)
	// fmt.Println(hwm)

	// [e] get the wordobjects for each unique headword: probedictionary()

	// note that the greek and latin dictionaries have extra fields that we are not using (right?)
	//var ec string
	//if d == "latin" {
	//	ec = "entry_key"
	//} else {
	//	ec = "unaccented_entry"
	//}

	// fld = fmt.Sprintf(`entry_name, metrical_entry, id_number, pos, translations, entry_body, %s`, ec)

	fld = `entry_name, metrical_entry, id_number, pos, translations, entry_body`
	psq = `SELECT %s FROM %s_dictionary WHERE %s ~* '^%s(|¹|²|³|⁴)$' ORDER BY id_number ASC`
	col := "entry_name"

	var lexicalfinds []DbLexicon
	dedup := make(map[int64]bool)
	for _, w := range hwm {
		// var foundrows pgx.Rows
		var err error
		q := fmt.Sprintf(psq, fld, d, col, w)
		foundrows, err = dbpool.Query(context.Background(), q)
		checkerror(err)

		defer foundrows.Close()
		for foundrows.Next() {
			var thehit DbLexicon
			err := foundrows.Scan(&thehit.Word, &thehit.Metrical, &thehit.ID, &thehit.POS, &thehit.Transl, &thehit.Entry)
			checkerror(err)
			if _, dup := dedup[thehit.ID]; !dup {
				// use ID and not Word because καρπόϲ.53442 is not καρπόϲ.53443
				dedup[thehit.ID] = true
				lexicalfinds = append(lexicalfinds, thehit)
			}
		}
	}

	//for _, x := range lexicalfinds {
	//	fmt.Println(fmt.Sprintf("%s: %s", x.Word, x.Transl))
	//}

	// [f] generate and format the prevalence data for this form: cf formatprevalencedata() in lexicalformatting.py

	fld = `entry_name, total_count, gr_count, lt_count, dp_count, in_count, ch_count`
	psq = `SELECT %s FROM wordcounts_%s where entry_name = '%s'`
	// golang hates indexing unicode strings: strings are bytes, and unicode chars take more than one byte
	c := []rune(word)
	q := fmt.Sprintf(psq, fld, stripaccents(string(c[0])), word)

	foundrows, err = dbpool.Query(context.Background(), q)
	checkerror(err)
	var wc DbWordCount
	defer foundrows.Close()
	for foundrows.Next() {
		// only one should ever return...
		err := foundrows.Scan(&wc.Word, &wc.Total, &wc.Gr, &wc.Lt, &wc.Dp, &wc.In, &wc.Ch)
		checkerror(err)
	}

	label := wc.Word
	allformpd := formatprevalencedata(wc, label)

	// [g] format the parsing summary

	parsing := formatparsingdata(mpp)

	// [h] generate the lexical output: multiple entries possible - <div id="δημόϲιοϲ_23337644"> ... <div id="δημοϲίᾳ_23333080"> ...

	var entries string
	for _, w := range lexicalfinds {
		entries += formatlexicaloutput(w)
	}

	// [i] add the HTML + JS to inject `{"newhtml": "...", "newjs":"..."}`

	//fmt.Println(allformpd)
	//fmt.Println(parsing)

	html := allformpd + parsing + entries
	// html = strings.Replace(html, `"`, `\"`, -1)
	js := insertlexicaljs()

	type JSB struct {
		HTML string `json:"newhtml"`
		JS   string `json:"newjs"`
	}

	var jb JSB
	jb.HTML = html
	jb.JS = js

	jsonbundle, ee := json.Marshal(jb)
	checkerror(ee)

	// jsonbundle := []byte(fmt.Sprintf(`{"newhtml":"%s","newjs":"%s"}`, html, js))
	return jsonbundle
}

// formatprevalencedata - turn a wordcount into an HTML summary
func formatprevalencedata(w DbWordCount, s string) string {
	// <p class="wordcounts">Prevalence (all forms): <span class="prevalence">Ⓣ</span> 1482 / <span class="prevalence">Ⓖ</span> 1415 / <span class="prevalence">Ⓓ</span> 54 / <span class="prevalence">Ⓘ</span> 11 / <span class="prevalence">Ⓒ</span> 2</p>

	pdp := `<p class="wordcounts">Prevalence of %s: %s</p>`
	pds := `<span class="prevalence">%s</span> %d`
	labels := map[string]string{"Total": "Ⓣ", "Gr": "Ⓖ", "Lt": "Ⓛ", "Dp": "Ⓓ", "In": "Ⓘ", "Ch": "Ⓒ"}

	var pdd []string
	for _, l := range []string{"Total", "Gr", "Lt", "Dp", "In", "Ch"} {
		v := reflect.ValueOf(w).FieldByName(l).Int()
		if v > 0 {
			pd := fmt.Sprintf(pds, labels[l], v)
			pdd = append(pdd, pd)
		}
	}

	spans := strings.Join(pdd, " / ")
	html := fmt.Sprintf(pdp, s, spans)
	return html
}

// formatparsingdata - turn []MorphPossib into HTML
func formatparsingdata(mpp []MorphPossib) string {
	obs := `
	<span class="obsv"><a class="parsing" href="#%s_%s"><span class="obsv"> from <span class="baseform">%s</span>
	<span class="baseformtranslation">&nbsp;(“%s”)</span></span></a></span>`
	mtb := `
	<table class="morphtable">
		<tbody>
		%s
		</tbody>
	</table>
	`
	mtr := `<tr>%s</tr>`
	mtd := `<td class="%s">%s</td>`

	var html string
	usecounter := false
	if len(mpp) > 1 {
		usecounter = true
	}
	for i, m := range mpp {
		if usecounter {
			html += fmt.Sprintf("(%d)&nbsp;", i+1)
		}
		html += fmt.Sprintf(obs, m.Headwd, m.Xrefval, m.Headwd, m.Transl)
		pos := strings.Split(m.Anal, " ")
		var tab string
		for _, p := range pos {
			tab += fmt.Sprintf(mtd, "morphcell", p)
		}
		tab = fmt.Sprintf(mtr, tab)
		tab = fmt.Sprintf(mtb, tab)
		html += tab
	}

	return html
}

// formatlexicaloutput - turn a DbLexicon word into HTML
func formatlexicaloutput(w DbLexicon) string {
	// [h1] first part of a lexical entry:

	// [h1a] known forms in use

	// requires probing dictionary_headword_wordcounts

	// 		SELECT
	//			entry_name , total_count, gr_count, lt_count, dp_count, in_count, ch_count,
	//			frequency_classification, early_occurrences, middle_occurrences ,late_occurrences,
	//			acta, agric, alchem, anthol, apocalyp, apocryph, apol, astrol, astron, biogr, bucol, caten, chronogr, comic, comm,
	//			concil, coq, dialog, docu, doxogr, eccl, eleg, encom, epic, epigr, epist, evangel, exeget, fab, geogr, gnom, gramm,
	//			hagiogr, hexametr, hist, homilet, hymn, hypoth, iamb, ignotum, invectiv, inscr, jurisprud, lexicogr, liturg, lyr,
	//			magica, math, mech, med, metrolog, mim, mus, myth, narrfict, nathist, onir, orac, orat, paradox, parod, paroem,
	//			perieg, phil, physiognom, poem, polyhist, prophet, pseudepigr, rhet, satura, satyr, schol, tact, test, theol, trag
	//		FROM dictionary_headword_wordcounts WHERE entry_name='%s'

	// [h1b] principle parts

	// [h2] wordcounts data including weighted distributions

	// [h3]  _buildentrysummary() which gives senses, flagged senses, citations, quotes (summarized or not)

	// [h4] the actual body of the entry

	// more formatting to come

	entrybody := w.Entry

	// [h5] previous & next entry
	enfolded := `<div id="%s_%d">%s</div>
	`
	html := fmt.Sprintf(enfolded, w.Word, w.ID, entrybody)
	return html
}

func insertlexicaljs() string {
	js := `
	<script>
	// Chromium can send poll data after the search is done... 
	$('#pollingdata').hide();
	
	$('%s').click( function() {
		$.getJSON('/browse/'+this.id, function (passagereturned) {
			$('#browseforward').unbind('click');
			$('#browseback').unbind('click');
			var fb = parsepassagereturned(passagereturned)
			// left and right arrow keys
			$('#browserdialogtext').keydown(function(e) {
				switch(e.which) {
					case 37: browseuponclick(fb[1]); break;
					case 39: browseuponclick(fb[0]); break;
				}
			});
			$('#browseforward').bind('click', function(){ browseuponclick(fb[0]); });
			$('#browseback').bind('click', function(){ browseuponclick(fb[1]); });
		});
	});
	</script>`

	tag := "bibl"

	thejs := fmt.Sprintf(js, tag)
	return thejs
}

//func main() {
//	// findbyform("ἐρχόμενον", "gr0062")
//	// findbyform("ἧκεν", "gr0062")
//	r := findbyform("ὀκνεῖϲ", "gr0062")
//	// findbyform("miles", "lt0448")
//	fmt.Println(string(r))
//}
//
//// DELETE LATER: in other files
//
//
//func stripaccents(u string) string {
//	// ὀκνεῖϲ --> οκνειϲ
//	feeder := make(map[rune][]rune)
//	feeder['α'] = []rune("αἀἁἂἃἄἅἆἇᾀᾁᾂᾃᾄᾅᾆᾇᾲᾳᾴᾶᾷᾰᾱὰάᾈᾉᾊᾋᾌᾍᾎᾏἈἉἊἋἌἍἎἏΑ")
//	feeder['ε'] = []rune("εἐἑἒἓἔἕὲέἘἙἚἛἜἝΕ")
//	feeder['ι'] = []rune("ιἰἱἲἳἴἵἶἷὶίῐῑῒΐῖῗΐἸἹἺἻἼἽἾἿΙ")
//	feeder['ο'] = []rune("οὀὁὂὃὄὅόὸὈὉὊὋὌὍΟ")
//	feeder['υ'] = []rune("υὐὑὒὓὔὕὖὗϋῠῡῢΰῦῧύὺὙὛὝὟΥ")
//	feeder['η'] = []rune("ηᾐᾑᾒᾓᾔᾕᾖᾗῂῃῄῆῇἤἢἥἣὴήἠἡἦἧᾘᾙᾚᾛᾜᾝᾞᾟἨἩἪἫἬἭἮἯΗ")
//	feeder['ω'] = []rune("ωὠὡὢὣὤὥὦὧᾠᾡᾢᾣᾤᾥᾦᾧῲῳῴῶῷώὼᾨᾩᾪᾫᾬᾭᾮᾯὨὩὪὫὬὭὮὯ")
//	feeder['ρ'] = []rune("ρῤῥῬ")
//	feeder['β'] = []rune("βΒ")
//	feeder['ψ'] = []rune("ψΨ")
//	feeder['δ'] = []rune("δΔ")
//	feeder['φ'] = []rune("φΦ")
//	feeder['γ'] = []rune("γΓ")
//	feeder['ξ'] = []rune("ξΞ")
//	feeder['κ'] = []rune("κΚ")
//	feeder['λ'] = []rune("λΛ")
//	feeder['μ'] = []rune("μΜ")
//	feeder['ν'] = []rune("νΝ")
//	feeder['π'] = []rune("πΠ")
//	feeder['ϙ'] = []rune("ϙϘ")
//	feeder['ϲ'] = []rune("ϲσΣςϹ")
//	feeder['τ'] = []rune("τΤ")
//	feeder['χ'] = []rune("χΧ")
//	feeder['θ'] = []rune("θΘ")
//	feeder['ζ'] = []rune("ζΖ")
//
//	reducer := make(map[rune]rune)
//	for f, _ := range feeder {
//		for _, r := range feeder[f] {
//			reducer[r] = f
//		}
//	}
//
//	var stripped []rune
//	for _, x := range []rune(u) {
//		stripped = append(stripped, reducer[x])
//	}
//
//	s := string(stripped)
//	return s
//}
//
//
//type RawPossib struct {
//	Number string
//	MP     string
//}
//
//type DbWordCount struct {
//	Word  string
//	Total int64
//	Gr    int64
//	Lt    int64
//	Dp    int64
//	In    int64
//	Ch    int64
//}
//
//type DbLexicon struct {
//	// skipping 'unaccented_entry' from greek_dictionary
//	// skipping 'entry_key' from latin_dictionary
//	Word     string
//	Metrical string
//	ID       int64
//	POS      string
//	Transl   string
//	Entry    string
//}
//
//type MorphPossib struct {
//	Transl   string `json:"transl"`
//	Anal     string `json:"analysis"`
//	Headwd   string `json:"headword"`
//	Scansion string `json:"scansion"`
//	Xrefkind string `json:"xref_kind"`
//	Xrefval  string `json:"xref_value"`
//}
//
//type DbMorphology struct {
//	Observed    string
//	Xrefs       string
//	PrefixXrefs string
//	RawPossib   string
//	RelatedHW   string
//}
//
//func grabpgsqlconnection() *pgxpool.Pool {
//	var pl PostgresLogin
//	pl.User = "hippa_wr"
//	pl.Host = "127.0.0.1"
//	pl.Pass = "8rnX8KBcbwvW8zH"
//	pl.Port = 5432
//	pl.DBName = "hipparchiaDB"
//
//	// using 'workers' was causing an m1 to choke when the worker count got high: no available connections to db
//	// panic: failed to connect to `host=localhost user=hippa_wr database=hipparchiaDB`: server error (FATAL: remaining connection slots are reserved for non-replication superuser connections (SQLSTATE 53300))
//	// workers := cfg.WorkerCount
//
//	url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", pl.User, pl.Pass, pl.Host, pl.Port, pl.DBName)
//
//	config, oops := pgxpool.ParseConfig(url)
//	if oops != nil {
//		msg(fmt.Sprintf("Could not execute pgxpool.ParseConfig(url) via %s", url), -1)
//		panic(oops)
//	}
//
//	// config.ConnConfig.PreferSimpleProtocol = true
//	// config.MaxConns = int32(workers * 3)
//	// config.MinConns = int32(workers + 2)
//
//	// the boring way if you don't want to go via pgxpool.ParseConfig(url)
//	// pooledconnection, err := pgxpool.Connect(context.Background(), url)
//
//	pooledconnection, err := pgxpool.ConnectConfig(context.Background(), config)
//
//	if err != nil {
//		msg(fmt.Sprintf("Could not connect to PostgreSQL via %s", url), -1)
//		panic(err)
//	}
//
//	msg(fmt.Sprintf("Connected to %s on PostgreSQL", pl.DBName), 4)
//
//	return pooledconnection
//}
//
//func checkerror(err error) {
//	if err != nil {
//		fmt.Println(fmt.Sprintf("UNRECOVERABLE ERROR: PLEASE TAKE NOTE OF THE FOLLOWING PANIC MESSAGE [%s v.%s]", myname, version))
//		panic(err)
//	}
//}
//
//func msg(message string, threshold int) {
//	if 5 >= threshold {
//		message = fmt.Sprintf("[%s] %s", shortname, message)
//		fmt.Println(message)
//	}
//}
//
//func unique[T comparable](s []T) []T {
//	// https://gosamples.dev/generics-remove-duplicates-slice/
//	inResult := make(map[T]bool)
//	var result []T
//	for _, str := range s {
//		if _, ok := inResult[str]; !ok {
//			inResult[str] = true
//			result = append(result, str)
//		}
//	}
//	return result
//}
//
//func flatten[T any](lists [][]T) []T {
//	var res []T
//	for _, list := range lists {
//		res = append(res, list...)
//	}
//	return res
//}
//
//const (
//	myname          = "Hipparchia Golang Server"
//	shortname       = "HGS"
//	version         = "0.0.1"
//	tesquery        = "SELECT * FROM %s WHERE index BETWEEN %d and %d"
//	testdb          = "lt0448"
//	teststart       = 1
//	testend         = 26
//	linelength      = 72
//	pollinginterval = 333 * time.Millisecond
//	skipheadwords   = "unus verum omne sum¹ ab δύο πρότεροϲ ἄνθρωποϲ τίϲ δέω¹ ὅϲτιϲ homo πᾶϲ οὖν εἶπον ἠμί ἄν² tantus μένω μέγαϲ οὐ verus neque eo¹ nam μέν ἡμόϲ aut Sue διό reor ut ἐγώ is πωϲ ἐκάϲ enim ὅτι² παρά ἐν Ἔχιϲ sed ἐμόϲ οὐδόϲ ad de ita πηρόϲ οὗτοϲ an ἐπεί a γάρ αὐτοῦ ἐκεῖνοϲ ἀνά ἑαυτοῦ quam αὐτόϲε et ὑπό quidem Alius¹ οἷοϲ noster γίγνομαι ἄνα προϲάμβ ἄν¹ οὕτωϲ pro² tamen ἐάν atque τε qui² si multus idem οὐδέ ἐκ omnes γε δεῖ πολύϲ in ἔδω ὅτι¹ μή Ios ἕτεροϲ cum meus ὅλοξ suus omnis ὡϲ sua μετά Ἀλλά ne¹ jam εἰϲ ἤ² ἄναξ ἕ ὅϲοϲ dies ipse ὁ hic οὐδείϲ suo ἔτι ἄνω¹ ὅϲ νῦν ὁμοῖοϲ edo¹ εἰ qui¹ πάλιν ὥϲπερ ne³ ἵνα τιϲ διά φύω per τοιοῦτοϲ for eo² huc locum neo¹ sui non ἤ¹ χάω ex κατά δή ἁμόϲ ὅμοιοϲ αὐτόϲ etiam vaco πρόϲ Ζεύϲ ϲύ quis¹ tuus b εἷϲ Eos οὔτε τῇ καθά ego tu ille pro¹ ἀπό suum εἰμί ἄλλοϲ δέ alius² pars vel ὥϲτε χέω res ἡμέρα quo δέομαι modus ὑπέρ ϲόϲ ito τῷ περί Τήιοϲ ἕκαϲτοϲ autem καί ἐπί nos θεάω γάρον γάροϲ Cos²"
//	skipinflected   = "ἀρ ita a inquit ego die nunc nos quid πάντων ἤ με θεόν δεῖ for igitur ϲύν b uers p ϲου τῷ εἰϲ ergo ἐπ ὥϲτε sua me πρό sic aut nisi rem πάλιν ἡμῶν φηϲί παρά ἔϲτι αὐτῆϲ τότε eos αὐτούϲ λέγει cum τόν quidem ἐϲτιν posse αὐτόϲ post αὐτῶν libro m hanc οὐδέ fr πρῶτον μέν res ἐϲτι αὐτῷ οὐχ non ἐϲτί modo αὐτοῦ sine ad uero fuit τοῦ ἀπό ea ὅτι parte ἔχει οὔτε ὅταν αὐτήν esse sub τοῦτο i omnes break μή ἤδη ϲοι sibi at mihi τήν in de τούτου ab omnia ὃ ἦν γάρ οὐδέν quam per α autem eius item ὡϲ sint length οὗ eum ἀντί ex uel ἐπειδή re ei quo ἐξ δραχμαί αὐτό ἄρα ἔτουϲ ἀλλ οὐκ τά ὑπέρ τάϲ μάλιϲτα etiam haec nihil οὕτω siue nobis si itaque uac erat uestig εἶπεν ἔϲτιν tantum tam nec unde qua hoc quis iii ὥϲπερ semper εἶναι e ½ is quem τῆϲ ἐγώ καθ his θεοῦ tibi ubi pro ἄν πολλά τῇ πρόϲ l ἔϲται οὕτωϲ τό ἐφ ἡμῖν οἷϲ inter idem illa n se εἰ μόνον ac ἵνα ipse erit μετά μοι δι γε enim ille an sunt esset γίνεται omnibus ne ἐπί τούτοιϲ ὁμοίωϲ παρ causa neque cr ἐάν quos ταῦτα h ante ἐϲτίν ἣν αὐτόν eo ὧν ἐπεί οἷον sed ἀλλά ii ἡ t te ταῖϲ est sit cuius καί quasi ἀεί o τούτων ἐϲ quae τούϲ minus quia tamen iam d διά primum r τιϲ νῦν illud u apud c ἐκ δ quod f quoque tr τί ipsa rei hic οἱ illi et πῶϲ φηϲίν τοίνυν s magis unknown οὖν dum text μᾶλλον habet τοῖϲ qui αὐτοῖϲ suo πάντα uacat τίϲ pace ἔχειν οὐ κατά contra δύο ἔτι αἱ uet οὗτοϲ deinde id ut ὑπό τι lin ἄλλων τε tu ὁ cf δή potest ἐν eam tum μου nam θεόϲ κατ ὦ cui nomine περί atque δέ quibus ἡμᾶϲ τῶν eorum"
//	memoutputfile   = "mem_profiler_output.bin"
//	cpuoutputfile   = "cpu_profiler_output.bin"
//	browseauthor    = "gr0062"
//	browsework      = "028"
//	browseline      = 14672
//	browsecontext   = 4
//)
//
//type PostgresLogin struct {
//	Host   string
//	Port   int
//	User   string
//	Pass   string
//	DBName string
//}
