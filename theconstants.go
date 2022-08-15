package main

const (
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
	// universalid | language | idxname | akaname | shortname | cleanname | genres | recorded_date | converted_date | location
	//-------------+----------+---------+---------+-----------+-----------+--------+---------------+----------------+----------
	//(0 rows)

	AUTHORTEMPLATE = `
			universalid,
			language,
			idxname,
			akaname,
			shortname,
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
)
