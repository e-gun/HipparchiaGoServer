//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vv

const (
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
