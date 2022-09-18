//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

const (
	MYNAME                  = "Hipparchia Golang Server"
	SHORTNAME               = "HGS"
	VERSION                 = "0.5.2"
	SERVEDFROMHOST          = "127.0.0.1"
	SERVEDFROMPORT          = 8000
	DBAUMAPSIZE             = 3455   //[HGS] [A2: 0.436s][Δ: 0.051s] 3455 authors built: map[string]DbAuthor
	DBLMMAPSIZE             = 151701 //[HGS] [B1: 0.310s][Δ: 0.310s] unnested lemma map built (151701 items)
	DBWKMAPSIZE             = 236835 //[HGS] [A1: 0.385s][Δ: 0.385s] 236835 works built: map[string]DbWork
	POLLEVERYNTABLES        = 50     // 3455 is the max number of tables in a search...
	WSPOLLINGPAUSE          = 800000
	DEFAULTBROWSERCTX       = 20
	DEFAULTCOLUMN           = "stripped_line"
	DEFAULTLINESOFCONTEXT   = 4
	DEFAULTHITLIMIT         = 200
	DEFAULTPROXIMITY        = 2
	MAXDISTANCE             = 10
	DEFAULTPROXIMITYSCOPE   = "lines"
	DEFAULTSYNTAX           = "~"
	FIRSTSEARCHLIM          = 500000
	INCERTADATE             = 2500
	MAXBROWSERCONTEXT       = 60
	MAXDATE                 = 1500
	MAXDATESTR              = "1500"
	MAXHITLIMIT             = 2500
	MAXINPUTLEN             = 50
	MAXLEMMACHUNKSIZE       = 20
	MAXLINESHITCONTEXT      = 30
	MAXTEXTLINEGENERATION   = 7500
	MAXDICTLOOKUP           = 100
	MINBROWSERWIDTH         = 90
	MINDATE                 = -850
	MINORGENREWTCAP         = 250
	TIMETRACKERMSGTHRESH    = 3
	SHOWCITATIONEVERYNLINES = 10
	MINDATESTR              = "-850"
	ORDERBY                 = "index"
	SORTBY                  = "shortname"
	TEMPTABLETHRESHOLD      = 100            // if a table requires N "between" clauses, build a temptable instead to gather the needed lines
	UNACCEPTABLEINPUT       = `|"'!@:,=+_\/` // we want to be able to do regex...; echo+net/url means some can't make it into a parser: #%&;
	VARIADATE               = 2000
	AUTHENTICATIONREQUIRED  = false
	GENRESTOCOUNT           = 5
	CONFIGNAME              = "hgs-conf.json"
	CONFIGLOCATION          = "."
	DEFAULTECHOLOGLEVEL     = 0
	DEFAULTGOLOGLEVEL       = 0

	PSQLHOST  = "127.0.0.1"
	PSQLUSER  = "hippa_wr"
	PSQLPORT  = 5432
	PSQLDB    = "hipparchiaDB"
	MINCONFIG = `
{"PosgreSQLPassword": "YOURPASSWORDHERE"}
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
		-cf {file}   read PSQL password from file [default: './%s']
		-el {num}    set echo server log level (0-2) [default: %d]
		-gl {num}    set golang log level (0-5) [default: %d]
		-h           print this help information
		-p  {string} supply full PostgreSQL credentials(*)
		-sa {string} server IP address [default: '%s']
		-sp {num}    server port [default: %d]
		-v           print version and exit

	(*) example: "{\"Pass\": \"YOURPASSWORDHERE\" ,\"Host\": \"127.0.0.1\", \"Port\": 5432, \"DBName\": \"hipparchiaDB\" ,\"User\": \"hippa_wr\"}"
`
)
