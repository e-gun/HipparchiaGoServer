//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

const (
	MYNAME                  = "Hipparchia Golang Server"
	SHORTNAME               = "HGS"
	VERSION                 = "0.6.7"
	AUTHENTICATIONREQUIRED  = false // unused ATM
	CONFIGLOCATION          = "."
	CONFIGNAME              = "hgs-conf.json"
	DBAUMAPSIZE             = 3455 //[HGS] [A2: 0.436s][Δ: 0.051s] 3455 authors built: map[string]DbAuthor
	DBLEMMACOUNT            = 152759
	DBLMMAPSIZE             = 151701 //[HGS] [B1: 0.310s][Δ: 0.310s] unnested lemma map built (151701 items)
	DBWKMAPSIZE             = 236835 //[HGS] [A1: 0.385s][Δ: 0.385s] 236835 works built: map[string]DbWork
	DEFAULTBROWSERCTX       = 20
	DEFAULTCOLUMN           = "stripped_line"
	DEFAULTECHOLOGLEVEL     = 0
	DEFAULTGOLOGLEVEL       = 0
	DEFAULTHITLIMIT         = 200
	DEFAULTLINESOFCONTEXT   = 4
	DEFAULTPROXIMITY        = 2
	DEFAULTPROXIMITYSCOPE   = "lines"
	DEFAULTSYNTAX           = "~"
	FIRSTSEARCHLIM          = 500000 // 149570 lines in Cicero (lt0474)
	FONTSETTING             = "SERVEALLFONTS"
	GENRESTOCOUNT           = 5
	INCERTADATE             = 2500
	MAXBROWSERCONTEXT       = 60
	MAXDATE                 = 1500
	MAXDATESTR              = "1500"
	MAXDICTLOOKUP           = 100
	MAXDISTANCE             = 10
	MAXHITLIMIT             = 2500
	MAXINPUTLEN             = 50
	MAXLEMMACHUNKSIZE       = 20
	MAXLINESHITCONTEXT      = 30
	MAXSEARCHINFOLISTLEN    = 100
	MAXTEXTLINEGENERATION   = 7500
	MINBROWSERWIDTH         = 90
	MINDATE                 = -850
	MINDATESTR              = "-850"
	MINORGENREWTCAP         = 250
	NESTEDLEMMASIZE         = 543
	ORDERBY                 = "index"
	POLLEVERYNTABLES        = 50 // 3455 is the max number of tables in a search...
	SERVEDFROMHOST          = "127.0.0.1"
	SERVEDFROMPORT          = 8000
	SHOWCITATIONEVERYNLINES = 10
	SORTBY                  = "shortname"
	TEMPTABLETHRESHOLD      = 100 // if a table requires N "between" clauses, build a temptable instead to gather the needed lines
	TIMETRACKERMSGTHRESH    = 3
	// UNACCEPTABLEINPUT       = `|"'!@:,=+_\/` // we want to be able to do regex...; echo+net/url means some can't make it into a parser: #%&;
	UNACCEPTABLEINPUT = `"'!@:,=_/̣` // we want to be able to do regex...; note the subscript dot at the end; echo+net/url means some can't make it into a parser: #%&;
	VARIADATE         = 2000
	WSPOLLINGPAUSE    = 800000

	PSQLHOST  = "127.0.0.1"
	PSQLUSER  = "hippa_wr"
	PSQLPORT  = 5432
	PSQLDB    = "hipparchiaDB"
	MINCONFIG = `
{"PosgreSQLPassword": "YOURPASSWORDHERE"}
`
	TERMINALTEXT = `
	%s / Copyright (C) %s / %s
	%s

	This program comes with ABSOLUTELY NO WARRANTY;
	without even the implied warranty of MERCHANTABILITY
	or FITNESS FOR A PARTICULAR PURPOSE.

	This is free software, and you are welcome to redistribute
	it and/or modify it under the terms of the GNU General
	Public License version 3.
`

	PROJ     = MYNAME
	PROJYEAR = "2022"
	PROJAUTH = "E. Gunderson"
	PROJMAIL = "Department of Classics, 125 Queen’s Park, Toronto, ON  M5S 2C7 Canada"

	HELPTEXT = `command line options:
   -cf {file}   read PSQL password from file [default: '%s/%s']
   -el {num}    set echo server log level (0-2) [default: %d]
   -ft {name}   force a client-side font instead of serving Noto fonts
                   names with spaces need quotes: "Gentium Plus Compact"
   -gl {num}    set golang log level (0-5) [default: %d]
   -h           print this help information
   -p  {string} supply full PostgreSQL credentials(*)
   -sa {string} server IP address [default: '%s']
   -sp {num}    server port [default: %d]
   -v           print version and exit

     (*) example: "{\"Pass\": \"YOURPASSWORDHERE\" ,\"Host\": \"127.0.0.1\", \"Port\": 5432, \"DBName\": \"hipparchiaDB\" ,\"User\": \"hippa_wr\"}"
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
			$.getJSON('/lexica/findbyform/'+this.id, function (definitionreturned) {
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
		
		$.getJSON('/lexica/lookup/^'+this.id+'$', function (definitionreturned) {
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

			let url = '/lexica/idlookup/' + language + '/' + entryid;
			
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
		
		$.getJSON('/lexica/morphologychart/'+this.lang+'/'+lexid+'/'+parserxref+'/'+headword, function (definitionreturned) {
			ldt.html(definitionreturned['newhtml']);
			jshld.html(definitionreturned['newjs']);		
			});
			
		return false;
		
		});
`
)
