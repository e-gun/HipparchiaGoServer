//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

const (
	MYNAME                = "Hipparchia Golang Server"
	SHORTNAME             = "HGS"
	VERSION               = "0.2.8"
	PSQ                   = `{"Host": "localhost", "Port": 5432, "User": "hippa_wr", "Pass": "", "DBName": "hipparchiaDB"}`
	PSDefaultHost         = "localhost"
	PSDefaultUser         = "hippa_wr"
	PSDefaultPort         = 5432
	PSDefaultDB           = "hipparchiaDB"
	DBAUMAPSIZE           = 3455   //[HGS] [A2: 0.436s][Δ: 0.051s] 3455 authors built: map[string]DbAuthor
	DBLMMAPSIZE           = 151701 //[HGS] [B1: 0.310s][Δ: 0.310s] unnested lemma map built (151701 items)
	DBWKMAPSIZE           = 236835 //[HGS] [A1: 0.385s][Δ: 0.385s] 236835 works built: map[string]DbWork
	DEFAULTBROWSERCTX     = 20
	DEFAULTCOLUMN         = "stripped_line"
	DEFAULTECHOLOGLEVEL   = 0
	DEFAULTLINESOFCONTEXT = 4
	DEFAULTHITLIMIT       = 200
	DEFAULTLOGLEVEL       = 3
	DEFAULTPROXIMITY      = 1
	DEFAULTPROXIMITYSCOPE = "lines"
	DEFAULTSYNTAX         = "~*"
	FIRSTSEARCHLIM        = 500000
	INCERTADATE           = 2500
	MAXBROWSERCONTEXT     = 60
	MAXDATE               = 1500
	MAXDATESTR            = "1500"
	MAXHITLIMIT           = 2500
	MAXINPUTLEN           = 50
	MAXLEMMACHUNKSIZE     = 20
	MAXLINESHITCONTEXT    = 30
	MAXTEXTLINEGENERATION = 7500
	MINBROWSERWIDTH       = 90
	MINDATE               = -850
	MINDATESTR            = "-850"
	ORDERBY               = "index"
	TEMPTABLETHRESHOLD    = 100          // if a table requirce N "between" clauses, build a temptable instead to gather the needed lines
	UNACCEPTABLEINPUT     = `|""'!@:,=+` // we want to be able to do regex...; echo+net/url means some can't make it into a parser: #%&;
	VARIADATE             = 2000
	ARCHIVEFOLDER         = "~"
	TARGETDIR             = "hipparchia-archive"

	// hipparchiaDB=# select * from gr0001 limit 0;
	// index | wkuniversalid | level_05_value | level_04_value | level_03_value | level_02_value | level_01_value | level_00_value | marked_up_line | accented_line | stripped_line | hyphenated_words | annotations
	//-------+---------------+----------------+----------------+----------------+----------------+----------------+----------------+----------------+---------------+---------------+------------------+-------------
	//(0 rows)

	WORLINETEMPLATE = `wkuniversalid,
			index,
			level_05_value,
			level_04_value,
			level_03_value,
			level_02_value,
			level_01_value,
			level_00_value,
			marked_up_line,
			accented_line,
			stripped_line,
			hyphenated_words,
			annotations`

	// hipparchiaDB=# select * from authors limit 0;
	// universalid | language | idxname | akaname | SHORTNAME | cleanname | genres | recorded_date | converted_date | location
	//-------------+----------+---------+---------+-----------+-----------+--------+---------------+----------------+----------
	//(0 rows)

	AUTHORTEMPLATE = `
			universalid,
			language,
			idxname,
			akaname,
			SHORTNAME,
			cleanname,
			genres,
			recorded_date,
			converted_date,
			location`

	// hipparchiaDB=# select * from works limit 0;
	// universalid | title | language | publication_info | levellabels_00 | levellabels_01 | levellabels_02 | levellabels_03 | levellabels_04 | levellabels_05 | workgenre | transmission | worktype | provenance | recorded_date | converted_date | wordcount | firstline | lastline | authentic
	//-------------+-------+----------+------------------+----------------+----------------+----------------+----------------+----------------+----------------+-----------+--------------+----------+------------+---------------+----------------+-----------+-----------+----------+-----------
	//(0 rows)

	WORKTEMPLATE = `
		universalid,
		title,
		language,
		publication_info,
		levellabels_00,
		levellabels_01,
		levellabels_02,
		levellabels_03,
		levellabels_04,
		levellabels_05,
		workgenre,
		transmission,
		worktype,
		provenance,
		recorded_date,
		converted_date,
		wordcount,
		firstline,
		lastline,
		authentic`

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
)

/*
corporaweights = {'gr': 1.0, 'lt': 12.7, 'in': 15.19, 'dp': 18.14, 'ch': 85.78}
Ⓖ 95,843 / Ⓛ 10 / Ⓘ 151 / Ⓓ 751 / Ⓒ 64 / Ⓣ 96,819
*/
