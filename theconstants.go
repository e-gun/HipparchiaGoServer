package main

const (
	MYNAME                = "Hipparchia Golang Server"
	SHORTNAME             = "HGS"
	VERSION               = "0.1.3"
	PSQ                   = `{"Host": "localhost", "Port": 5432, "User": "hippa_wr", "Pass": "", "DBName": "hipparchiaDB"}`
	PSDefaultHost         = "localhost"
	PSDefaultUser         = "hippa_wr"
	PSDefaultPort         = 5432
	PSDefaultDB           = "hipparchiaDB"
	TwoPassThresh         = 100 // cicero has >70 works
	VARIADATE             = 2000
	INCERTADATE           = 2500
	MINDATE               = -850
	MAXDATE               = 1500
	DEFAULTBROWSERCTX     = 20
	DEFAULTHITLIMIT       = 200
	DEFAULTPROXIMITY      = 1
	DEFAULTPROXIMITYSCOPE = "lines"
	DEFAULTLOGLEVEL       = 3
	DEFAULTECHOLOGLEVEL   = 0
	MINBROWSERWIDTH       = 90
	MAXLEMMACHUNKSIZE     = 20
	FIRSTSEARCHLIM        = 500000
	DBWKMAPSIZE           = 236835 //[HGS] [A1: 0.385s][Δ: 0.385s] 236835 works built: map[string]DbWork
	DBAUMAPSIZE           = 3455   //[HGS] [A2: 0.436s][Δ: 0.051s] 3455 authors built: map[string]DbAuthor
	DBLMMAPSIZE           = 151701 //[HGS] [B1: 0.310s][Δ: 0.310s] unnested lemma map built (151701 items)
	TEMPTABLETHRESHOLD    = 2

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
	
	$('browser').click( function() {
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
