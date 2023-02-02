//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import "time"

const (
	MYNAME                   = "Hipparchia Golang Server"
	SHORTNAME                = "HGS"
	VERSION                  = "1.0.16"
	AVGWORDSPERLINE          = 8 // hard coding a suspect assumption
	BLACKANDWHITE            = false
	CONFIGLOCATION           = "."
	CONFIGALTAPTH            = "%s/.config/" // %s = os.UserHomeDir()
	CONFIGAUTH               = "hgs-users.json"
	CONFIGBASIC              = "hgs-conf.json"
	CONFIGPROLIX             = "hgs-prolix-conf.json"
	DBAUMAPSIZE              = 3455   //[HGS] [A2: 0.436s][Δ: 0.051s] 3455 authors built: map[string]DbAuthor
	DBLMMAPSIZE              = 151701 //[HGS] [B1: 0.310s][Δ: 0.310s] unnested lemma map built (151701 items)
	DBWKMAPSIZE              = 236835 //[HGS] [A1: 0.385s][Δ: 0.385s] 236835 works built: map[string]DbWork
	DEFAULTBROWSERCTX        = 12
	DEFAULTCOLUMN            = "stripped_line"
	DEFAULTCORPORA           = "{\"gr\": true, \"lt\": true, \"in\": false, \"ch\": false, \"dp\": false}"
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
	DEFAULTSYNTAX            = "~"
	FIRSTSEARCHLIM           = 750000          // 149570 lines in Cicero (lt0474); all 485 forms of »δείκνυμι« will pass 50k
	FONTSETTING              = "SERVEALLFONTS" // will send Noto when this "font" is not found
	GENRESTOCOUNT            = 5
	INCERTADATE              = 2500
	JSONINDENT               = "  "
	MAXBROWSERCONTEXT        = 60
	MAXDATE                  = 1500
	MAXDATESTR               = "1500"
	MAXDICTLOOKUP            = 125
	MAXDISTANCE              = 10
	MAXECHOREQPERSECONDPERIP = 40 // it takes c. 20 to load the front page for the first time; 40 lets you double-load
	MAXHITLIMIT              = 2500
	MAXINPUTLEN              = 50
	MAXLEMMACHUNKSIZE        = 20
	MAXLINESHITCONTEXT       = 30
	MAXSEARCHINFOLISTLEN     = 100
	MAXTEXTLINEGENERATION    = 35000 // euripides is 33517 lines, sophocles is 15729, cicero is 149570, e.g.; jQuery slows exponentially as lines increase
	MAXVOCABLINEGENERATION   = 1     // this is a multiplier for Config.MaxText; the browser does not get overwhelmed by these lists
	MAXTITLELENGTH           = 110
	MINBROWSERWIDTH          = 90
	MINDATE                  = -850
	MINDATESTR               = "-850"
	MINORGENREWTCAP          = 250
	MSGMAND                  = -1
	MSGCRIT                  = 0
	MSGWARN                  = 1
	MSGNOTE                  = 2
	MSGFYI                   = 3
	MSGPEEK                  = 4
	MSGTMI                   = 5
	NESTEDLEMMASIZE          = 543
	ORDERBY                  = "index"
	POLLEVERYNTABLES         = 50 // 3455 is the max number of tables in a search...
	SERVEDFROMHOST           = "127.0.0.1"
	SERVEDFROMPORT           = 8000
	SIMULTANEOUSSEARCHES     = 3 // cap on the number of db connections at (S * Config.WorkerCount)
	SHOWCITATIONEVERYNLINES  = 10
	SORTBY                   = "shortname"
	TEMPTABLETHRESHOLD       = 100 // if a table requires N "between" clauses, build a temptable instead to gather the needed lines
	TERMINATIONS             = `(\s|\.|\]|\<|⟩|’|”|\!|,|:|;|\?|·|$)`
	TIMEOUTRD                = 15 * time.Second  // only set if Config.Authenticate is true (and so in a "serve the net" situation)
	TIMEOUTWR                = 120 * time.Second // this is *very* generous, but some searches are slow/long
	TIMETRACKERMSGTHRESH     = 3
	USEGZIP                  = false
	VARIADATE                = 2000
	VOCABSCANSION            = false
	VOCABBYCOUNT             = false
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

	HELPTEXT = `command line options:
   C1-acC0 C2{string}C0 set corpora active on startup and reset C4(*)C0
   C1-auC0          require authentication 
                   also implies "C3%sC0" exists and has been properly configured C4(**)C0
   C1-bcC0 C2{num}C0    default lines of browser context to display [current: C3%dC0]
   C1-bwC0          disable color output in the console
   C1-cfC0 C2{file}C0   read PSQL password from file [default: looks for "C3%s/%sC0" and "C3%s%sC0"]
   C1-dbC0          debug database: show internal references in browsed passages
   C1-elC0 C2{num}C0    set echo server log level (C10-3C0) [default: C3%dC0]
   C1-ftC0 C2{string}C0 force a client-side font instead of serving a font [served font: C3NotoSansC0]
                   names with spaces need quotes around them: "C4Gentium Plus CompactC0"
   C1-glC0 C2{num}C0    set golang log level (C10-5C0) [default: C3%dC0]
   C1-gzC0          enable gzip compression of the server's output
   C1-hC0           print this help information
   C1-pgC0 C2{string}C0 supply full PostgreSQL credentials C4(†)C0
   C1-qC0           quiet startup: suppress copyright notice
   C1-saC0 C2{string}C0 server IP address [default: "C3%sC0"]
   C1-spC0 C2{num}C0    server port [default: C3%dC0]
   C1-uiC0 C2{string}C0 unacceptable input characters [default: C3%sC0]
   C1-vC0           print version info and exit
   C1-vvC0          print full version info and exit
   C1-wcC0 C2{int}C0    number of workers [default: C1cpu_countC0 (C3%dC0)]
   C1-zlC0          zap lunate sigmas and replace them with σ/ς

     (*) example: 
         C4"{\"gr\": true, \"lt\": true, \"in\": false, \"ch\": false, \"dp\": false}"C0

     (**) example:
         C4[{"User": "user1","Pass": "pass1"}, {"User":"user2","Pass":"pass2"}, ...]C0

     (†) example: 
         C4"{\"Pass\": \"YOURPASSWORDHERE\" ,\"Host\": \"127.0.0.1\", \"Port\": 5432, \"DBName\": \"hipparchiaDB\" ,\"User\": \"hippa_wr\"}"C0
     
     NB: place a properly formatted version of 'C3%sC0' in 'C3%sC0' 
         if you want to avoid constantly setting multiple options. 
         See 'C3sample_hgs-prolix-conf.jsonC0' as well as other sample configuration files at
             C3%sC0
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
			$( '#lexicadialogtext' ).dialog( 'open' );
			$( '#lexicadialogtext' ).html('[searching...]');
			$.getJSON('/lex/findbyform/'+this.id, function (definitionreturned) {
				$( '#lexicadialogtext' ).html(definitionreturned['newhtml']);
				$( '#lexicaljsscriptholder' ).html(definitionreturned['newjs']);
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
					ldt.html(definitionreturned['newhtml']);
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
		
		ldt.dialog({
			closeOnEscape: true,
			autoOpen: false,
			minWidth: windowWidth*.33,
			maxHeight: windowHeight*.9,
			// position: { my: "left top", at: "left top", of: window },
			title: this.id,
			draggable: true,
			icons: { primary: 'ui-icon-close' },
			click: function() { $(this).dialog('close'); }
			});
		
		ldt.dialog('open');
		ldt.html('[searching...]');
		
		$.getJSON('/lex/lookup/^'+this.id+'$', function (definitionreturned) {
			ldt.html(definitionreturned['newhtml']);
			jshld.html(definitionreturned['newjs']);		
			});
		return false;
		
		});

	$('dictionaryidsearch').click( function(){
			$('#imagearea').empty();

			let ldt = $('#lexicadialogtext');
			let jshld = $('#lexicaljsscriptholder');
	
			let entryid = this.getAttribute("entryid");
			let language = this.getAttribute("language");

			let url = '/lex/idlookup/' + language + '/' + entryid;
			
			$.getJSON(url, function (definitionreturned) { 
				ldt.html(definitionreturned['newhtml']);
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
		
		ldt.dialog({
			closeOnEscape: true,
			autoOpen: false,
			minWidth: windowWidth*.33,
			maxHeight: windowHeight*.9,
			// position: { my: "left top", at: "left top", of: window },
			title: headword,
			draggable: true,
			icons: { primary: 'ui-icon-close' },
			click: function() { $(this).dialog('close'); }
			});
		
		ldt.dialog('open');
		ldt.html('[searching...]');
		
		$.getJSON('/lex/chart/'+this.lang+'/'+lexid+'/'+parserxref+'/'+headword, function (definitionreturned) {
			ldt.html(definitionreturned['newhtml']);
			jshld.html(definitionreturned['newjs']);		
			});
			
		return false;
		
		});
`
	AUTHHTML = `    
	<div id="currentuser" class="unobtrusive">
        <span id="userid" class="user">{{index . "user" }}</span>
        <span id="executelogout" class="material-icons">person_remove</span>
        <span id="executelogin" class="material-icons">login</span>
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
