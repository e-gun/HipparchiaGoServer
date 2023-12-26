//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import "time"

const (
	MYNAME    = "Hipparchia Golang Server"
	SHORTNAME = "HGS"
	VERSION   = "0.0.0"

	GREEKCORP      = "gr"
	LATINCORP      = "lt"
	PAPYRUSCORP    = "dp"
	INSCRIPTCORP   = "in"
	CHRISTINSC     = "ch"
	DEFAULTCORPORA = "{\"gr\": true, \"lt\": true, \"in\": false, \"ch\": false, \"dp\": false}"

	AVGWORDSPERLINE      = 8 // hard coding a suspect assumption
	BLACKANDWHITE        = false
	CHARSPERLINE         = 60 // used by vector to preallocate memory: set it closer to a max than a real average
	CONFIGLOCATION       = "."
	CONFIGALTAPTH        = "%s/.config/" // %s = os.UserHomeDir()
	CONFIGAUTH           = "hgs-users.json"
	CONFIGBASIC          = "hgs-conf.json"
	CONFIGPROLIX         = "hgs-prolix-conf.json"
	CONFIGVECTORW2V      = "hgs-vector-conf-w2v.json"
	CONFIGVECTORGLOVE    = "hgs-vector-conf-glove.json"
	CONFIGVECTORLEXVEC   = "hgs-vector-conf-lexvec.json"
	CONFIGVECTORLDA      = "hgs-vector-conf-lda.json"
	CONFIGVECTORSTOPSLAT = "hgs-vector-stops-latin.json"
	CONFIGVECTORSTOPSGRK = "hgs-vector-stops-greek.json"
	CUSTOMCSSFILENAME    = "custom-hipparchiastyles.css"
	// DBAUMAPSIZE              = 3455   //[HGS] [A2: 0.436s][Δ: 0.051s] 3455 authors built: map[string]DbAuthor
	DBLMMAPSIZE = 151701 //[HGS] [B1: 0.310s][Δ: 0.310s] unnested lemma map built (151701 items)
	// DBWKMAPSIZE              = 236835 //[HGS] [A1: 0.385s][Δ: 0.385s] 236835 works built: map[string]DbWork
	DEFAULTBROWSERCTX        = 14
	DEFAULTCHRTWIDTH         = "1500px"
	DEFAULTCHRTHEIGHT        = "1200px"
	DEFAULTCOLUMN            = "stripped_line"
	DEFAULTECHOLOGLEVEL      = 0
	DEFAULTGOLOGLEVEL        = 0
	DEFAULTHITLIMIT          = 250
	DEFAULTLINESOFCONTEXT    = 4
	DEFAULTPROXIMITY         = 1
	DEFAULTPROXIMITYSCOPE    = "lines"
	DEFAULTPSQLHOST          = "127.0.0.1"
	DEFAULTPSQLUSER          = "hippa_wr"
	DEFAULTPSQLPORT          = 5432
	DEFAULTPSQLDB            = "hipparchiaDB"
	FIRSTSEARCHLIM           = 750000 // 149570 lines in Cicero (lt0474); all 485 forms of »δείκνυμι« will pass 50k
	FONTSETTING              = "Noto"
	GENRESTOCOUNT            = 5
	HDBFOLDER                = "hDB"
	INCERTADATE              = 2500
	JSONINDENT               = "  "
	LDATOPICS                = 8
	LDAMAXTOPICS             = 30
	LDASENTPERBAG            = 1
	LDAITER                  = 60
	LDAXFORMPASSES           = 30
	LDABURNINPASSES          = 1
	LDACHGEVALFRQ            = 15
	LDAPERPEVALFRQ           = 15
	LDAPERPTOL               = 1e-2
	LDAMAXGRAPHLINES         = 30000
	MAXBROWSERCONTEXT        = 60
	MAXDATE                  = 1500
	MAXDATESTR               = "1500"
	MAXDICTLOOKUP            = 125
	MAXDISTANCE              = 10
	MAXECHOREQPERSECONDPERIP = 60 // it takes c. 20 to load the front page for the first time; 40 lets you double-load; selftestsuite needs 60
	MAXHITLIMIT              = 2500
	MAXINPUTLEN              = 50
	MAXLEMMACHUNKSIZE        = 25
	MAXLINESHITCONTEXT       = 30
	MAXSEARCHINFOLISTLEN     = 100
	MAXSEARCHPERIPADDR       = 2
	MAXSEARCHTOTAL           = 4     // note that vectors and two-part searches generate subsearches and kick your total active search count over the number of "clicked" searches from RtSearch()
	MAXTEXTLINEGENERATION    = 35000 // euripides is 33517 lines, sophocles is 15729, cicero is 149570, e.g.; jQuery slows exponentially as lines increase
	MAXVOCABLINEGENERATION   = 1     // this is a multiplier for Config.MaxText; the browser does not get overwhelmed by these lists
	MAXTITLELENGTH           = 110
	MINBROWSERWIDTH          = 90
	MINDATE                  = -850
	MINDATESTR               = "-850"
	MINORGENREWTCAP          = 250
	NESTEDLEMMASIZE          = 543
	ORDERBY                  = "index"
	POLLEVERYNTABLES         = 34 // 3455 is the max number of tables in a search...
	QUERYSYNTAXPGSQL         = "~"
	QUERYSYNTAXSQLITE        = "regexp"
	SERVEDFROMHOST           = "127.0.0.1"
	SERVEDFROMPORT           = 8000
	SIMULTANEOUSSEARCHES     = 3 // cap on the number of db connections at (S * Config.WorkerCount)
	SHOWCITATIONEVERYNLINES  = 10
	SORTBY                   = "shortname"
	TEMPTABLETHRESHOLD       = 100 // if a table requires N "between" clauses, build a temptable instead to gather the needed lines
	TERMINATIONS             = `(\s|\.|\]|\<|⟩|’|”|\!|,|:|;|\?|·|$)`
	TICKERISACTIVE           = false
	TICKERDELAY              = 30 * time.Second
	TIMEOUTRD                = 15 * time.Second  // only set if Config.Authenticate is true (and so in a "serve the net" situation)
	TIMEOUTWR                = 120 * time.Second // this is *very* generous, but some searches are slow/long
	USEGZIP                  = false
	VARIADATE                = 2000
	VECTORNEIGHBORS          = 16
	VECTORNEIGHBORSMAX       = 40
	VECTORNEIGHBORSMIN       = 4
	VECTORTABLENAMENN        = "semantic_vectors_nn"
	VECTORTABLENAMELDA       = "semantic_vectors_lda"
	VECTORMAXLINES           = 1000000 // 964403 lines will get you all of Latin
	VECTORMODELDEFAULT       = "w2v"
	VECTORTEXTPREPDEFAULT    = "winner"
	VECTROWEBEXTDEFAULT      = false
	VOCABSCANSION            = false
	VOCABBYCOUNT             = false
	WRITEPERMS               = 0644
	WSPOLLINGPAUSE           = 10000000 * 10 // 10000000 * 10 = every .1s

	// UNACCEPTABLEINPUT       = `|"'!@:,=+_\/`
	UNACCEPTABLEINPUT = `"'!@:,=_/` // we want to be able to do regex...; echo+net/url means some can't even make it into a parser: #%&;
	USELESSINPUT      = `’“”̣`      // these can't be found and so should be dropped; note the subscript dot at the end

	MINCONFIG = `
{"PostgreSQLPassword": "YOURPASSWORDHERE"}
`
	TERMINALTEXT = `Copyright (C) %s / %s
      %s

      This program comes with ABSOLUTELY NO WARRANTY; without even the  
      implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

      This is free software, and you are welcome to redistribute it and/or 
      modify it under the terms of the GNU General Public License version 3.`

	PROJYEAR = "2022-23"
	PROJAUTH = "E. Gunderson"
	PROJMAIL = "Department of Classics, 125 Queen’s Park, Toronto, ON  M5S 2C7 Canada"
	PROJURL  = "https://github.com/e-gun/HipparchiaGoServer"

	HELPTEXTTEMPLATE = `S3command line optionsS0:
   C1-auC0          require authentication; also implies "C3{{.confauth}}C0" exists and has been properly configured
   C1-avC0          automatically generate vector models for every author in the database
                   default settings will consume c. C11.3GBC0 of extra disk space
   C1-bcC0 C2{num}C0    default lines of browser context to display [C6currentC0: C3{{.ctxlines}}C0]
   C1-bwC0          disable color output in the console
   C1-csC0          use a custom CSS file; will try to read "C3{{.home}}{{.css}}C0"
   C1-dbC0          debug database: show internal references in browsed passages
   C1-dvC0          disable semantic vector searching
   C1-elC0 C2{num}C0    set echo server log level (C10-3C0) [C6currentC0: C3{{.echoll}}C0]
   C1-exC0          extract the data to an archive folder in the same directory as the application; data sent to: "C3{{.cwd}}C0"
   C1-ftC0 C2{string}C0 force a client-side font instead of serving a font [C6served font:C0 C3NotoSansC0]
                   names with spaces need quotes around them: "C4Gentium Plus CompactC0"
   C1-glC0 C2{num}C0    set golang log level (C10-5C0) [C6currentC0: C3{{.hgsll}}C0]
   C1-gzC0          enable gzip compression of the server's output
   C1-hC0           print this help information
   C1-mdC0 C2{string}C0 set the default vector model type; available: C3gloveC0, C3lexvecC0, & C3w2vC0 [C6currentC0: C3{{.vmodel}}C0]
   C1-miC0 C2{num}C0    maximum number of concurrent searches per IP address [C6currentC0: C3{{.maxipsrch}}C0]
   C1-msC0 C2{num}C0    maximum total number of concurrent searches [C6currentC0: C3{{.maxtotscrh}}C0]
   C1-pcC0          enable CPU profiling run
   C1-pdC0          write a copy of the embedded PDF instructions to the current directory
   C1-pmC0          enable MEM profiling run
   C1-pgC0 C2{string}C0 supply full PostgreSQL credentials C4(*)C0
   C1-qC0           quiet startup: suppress copyright notice
   C1-rlC0          reload the database tables; data will be read from: "C3{{.dbf}}C0" in "C3{{.cwd}}C0"
   C1-rvC0          reset the stored semantic vector table
   C1-saC0 C2{string}C0 server IP address [C6currentC0: C3{{.host}}C0]
   C1-spC0 C2{num}C0    server port [C6currentC0: C3{{.port}}C0]
   C1-stC0          run the self-test suite at launch; repeat the flag to iterate: e.g., "C1-st -stC0" will run twice
   C1-tkC0          turn on the uptime UptimeTicker [unavailable if OS is Windows]
   C1-uiC0 C2{string}C0 unacceptable input characters [C6currentC0: C3{{.badchars}}C0]
   C1-vC0           print version info and exit
   C1-vvC0          print full version info and exit
   C1-wcC0 C2{int}C0    number of workers [C6currentC0: C3{{.workers}}C0][C1cpu_countC0 is C3{{.cpus}}C0]
   C1-zlC0          zap lunate sigmas and replace them with C1σ/ςC0
   C1-00C0          completely erase the database and reset the tables
                   the application cannot run again until you restore its data from an archive 
                   you probably want to run with the "C1-exC0" flag before you try this.
     (*) S3exampleS0: 
         C4"{\"Pass\": \"YOURPASSWORDHERE\" ,\"Host\": \"127.0.0.1\", \"Port\": 5432, \"DBName\": \"hipparchiaDB\" ,\"User\": \"hippa_wr\"}"C0
     
     S1NB:S0 a properly formatted version of "C3{{.conffile}}C0" in "C3{{.home}}C0" configures everything for you. 
         See 'C3sample_hgs-prolix-conf.jsonC0' as well as other sample configuration files at
             C3{{.projurl}}C0
`
	LEXFINDJS = `
		$('%s').click( function(e) {
			e.preventDefault();
			var windowWidth = $(window).width();
			var windowHeight = $(window).height();
			$( '#lexicadialogtext' ).dialog({
				closeOnEscape: true, 
				autoOpen: false,
				minWidth: windowWidth*.33,
				maxHeight: windowHeight*.9,
				// position: { my: "left top", at: "left top", of: window },
				title: this.id,
				draggable: true,
				icons: { primary: 'ui-icon-close' },
				click: function() { $( this ).dialog( 'close' ); }
				});
			$.getJSON('/lex/findbyform/'+this.id, function (definitionreturned) {
				$( '#lexicaljsscriptholder' ).html(definitionreturned['newjs']);
				document.getElementById('lexmodalbody').innerHTML = definitionreturned['newhtml']
				document.getElementById('lexmodal').style.display = "block";
			});
		return false;
		});`

	BROWSERJS = `
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
	`

	MORPHJS = `
	<script>
		function displayresults(output) {
			document.title = output['title'];
			$('#searchsummary').html(output['searchsummary']);
			$('#displayresults').html(output['found']);
			let browserclickscript = document.createElement('script');
			browserclickscript.innerHTML = output['js'];
			document.getElementById('browserclickscriptholder').appendChild(browserclickscript);
		}

		$('verbform').click( function(){
			$('#imagearea').empty();
			$('#searchsummary').html('');
			$('#displayresults').html('');
			$('#pollingdata').show();
			
			let bcsh = document.getElementById("browserclickscriptholder");
			if (bcsh.hasChildNodes()) { bcsh.removeChild(bcsh.firstChild); }
	
			let searchterm = this.getAttribute("searchterm");
			
			let searchid = generateId(8);
			let url = '/srch/exec/' + searchid + '?skg=%20' + searchterm + '%20';
			
			$.getJSON(url, function (returnedresults) { displayresults(returnedresults); });
			
			checkactivityviawebsocket(searchid);
		});

		$('dictionaryidsearch').click( function(){
				$('#imagearea').empty();
	
				let ldt = $('#lexicadialogtext');
				let jshld = $('#lexicaljsscriptholder');
		
				let entryid = this.getAttribute("entryid");
				let language = this.getAttribute("language");
	
				let url = '/lex/idlookup/' + language + '/' + entryid;
				
				$.getJSON(url, function (definitionreturned) {
					document.getElementById('leftmodalheadertext').innerHTML = entryid;
					document.getElementById('lexmodalbody').innerHTML = definitionreturned['newhtml'];
					document.getElementById('lexmodal').style.display = "block";
					jshld.html(definitionreturned['newjs']);
				});
			});
	</script>`

	DICTIDJS = `
	$('dictionaryentry').click( function(e) {
		e.preventDefault();
		var windowWidth = $(window).width();
		var windowHeight = $(window).height();
		let ldt = $('#lexicadialogtext');
		let jshld = $('#lexicaljsscriptholder');
		var htxt = this.id;

		$.getJSON('/lex/lookup/^'+this.id+'$', function (definitionreturned) {
				document.getElementById('leftmodalheadertext').innerHTML = htxt;
				document.getElementById('lexmodalbody').innerHTML = definitionreturned['newhtml'];
				document.getElementById('lexmodal').style.display = "block";
				jshld.html(definitionreturned['newjs']);
			});
		return false;	
		});

	$('dictionaryidsearch').click( function(){
			$('#imagearea').empty();
			let jshld = $('#lexicaljsscriptholder');
			let entryid = this.getAttribute("entryid");
			let language = this.getAttribute("language");

			let url = '/lex/idlookup/' + language + '/' + entryid;
			document.getElementById('leftmodalheadertext').innerHTML = "[searching...]";
			document.getElementById('lexmodal').style.display = "block";
			$.getJSON(url, function (definitionreturned) {
				document.getElementById('leftmodalheadertext').innerHTML = entryid;
				document.getElementById('lexmodalbody').innerHTML = definitionreturned['newhtml'];
				jshld.html(definitionreturned['newjs']);
			});

		});
	
	$('formsummary').click( function(e) {
		e.preventDefault();
		var windowWidth = $(window).width();
		var windowHeight = $(window).height();
		let ldt = $('#lexicadialogtext');
		let jshld = $('#lexicaljsscriptholder');
		let headword = this.getAttribute("headword");
		let parserxref = this.getAttribute("parserxref");
		let lexid = this.getAttribute("lexicalid");
		document.getElementById('morphmodal').style.display = "block";
		document.getElementById('rightmodalheadertext').innerHTML = "[building chart...]";
		$.getJSON('/lex/chart/'+this.lang+'/'+lexid+'/'+parserxref+'/'+headword, function (definitionreturned) {
				document.getElementById('rightmodalheadertext').innerHTML = lexid;
				document.getElementById('morphmodalbody').innerHTML = definitionreturned['newhtml'];
				jshld.html(definitionreturned['newjs']);
			});
			
		return false;
		
		});
`

	VECTORJS = `
        $('#pollingdata').html('');

		$('vectorheadword').click( function(e) { 
			var searchid = generateId(8);
			url = '/srch/exec/' + searchid + '?lem=' + this.id;
			$('#imagearea').empty();
			$('#searchsummary').html(''); 
			$('#displayresults').html('');
			$('#vectorgraphing').html('');
			$('#wordsearchform').hide();
			$('#lemmatasearchform').show();
			$('#lemmatasearchform').val(this.id);
			$('#lexicon').val(' '+this.id+' ');
        	checkactivityviawebsocket(searchid);
        	$.getJSON(url, function (returnedresults) { loadnewres(returnedresults); });
		});

    function loadnewres(output) {
        document.title = output['title'];
        $('#searchsummary').html(output['searchsummary']);
        $('#displayresults').html(output['found']);
        $('#vectorgraphing').html(output['image']);
        let browserclickscript = document.createElement('script');
        browserclickscript.innerHTML = output['js'];
        document.getElementById('browserclickscriptholder').appendChild(browserclickscript);
    }`

	AUTHHTML = `    
	<div id="currentuser" class="unobtrusive">
        <span id="userid" class="user">{{index . "user" }}</span>
        <span id="executelogout" class="material-icons material-icons-outline">verified_user</span>
        <span id="executelogin" class="material-icons material-icons-outline">shield</span>
        <br>
        <span id="alertarea"></span>
    </div>
    <div id="validateusers" class="center unobtrusive ui-dialog-content ui-widget-content" title="Please log in...">
        <form id="hipparchiauserlogin" method="POST" action="/auth/login">
            <input id="user" name="user" placeholder="[username]" required="" size="12" type="text" value="">
            <input id="pw" name="pw" placeholder="[password]" required="" size="12" type="password" value="">
            <p class="center"><input type="submit" name="login" value="Login"></p>
        </form>
    </div>`
	AUTHWARN      = "Please log in first..."
	VALIDATIONBOX = "$('#validateusers').dialog( 'open' );"
	JSVALIDATION  = "<script>" + VALIDATIONBOX + "</script>"
)
