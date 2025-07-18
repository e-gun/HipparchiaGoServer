//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package vv

const (
	CLICKTOLOOKUP = `
    var vocabList = document.querySelectorAll('**REPLACEME**');
    for (let i = 0; i < vocabList.length; i++) {
        const theitem = vocabList[i];
        theitem.addEventListener('click', function(e) {
            const theword = e.target.id.split('--')[0];
            const windowWidth = window.innerWidth;
            const windowHeight = window.innerHeight;
            e.preventDefault();
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
            // Make a GET request to the /lex/findbyform endpoint
            fetch('/lex/findbyform/' + theword)
                .then(response => response.json())
                .then(definitionreturned => {
                    // Update the content of the dialog box with the returned data
                    const newJS = definitionreturned['newjs'];
                    const newHTML = definitionreturned['newhtml'];
                    document.getElementById('lexmodalbody').innerHTML = newHTML;
                    document.getElementById('lexmodal').style.display = 'block';
                });
        });
    }`

	// CLICKTOLOOKUPINPROGRESS - todo: replace jquery $( '#lexicadialogtext' ).dialog():
	CLICKTOLOOKUPINPROGRESS = `
    var vocabList = document.querySelectorAll('**REPLACEME**');
    for (let i = 0; i < vocabList.length; i++) {
        const theitem = vocabList[i];
        theitem.addEventListener('click', function(e) {
            const theword = e.target.id.split('--')[0];
            e.preventDefault();
            // Get the window dimensions
           const windowWidth = window.innerWidth;
           const windowHeight = window.innerHeight;

           // Create a new dialog box element
           const dialogBox = document.createElement('div');
           dialogBox.id = 'lexmodalbody';
           dialogBox.style.width = (windowWidth * 0.33).toString() + 'px';
           dialogBox.style.height = (windowHeight * 0.9).toString() + 'px';
           dialogBox.style.position = 'absolute';

           // Set the title of the dialog box
           const title = document.createElement('h3');
           title.textContent = this.id;
           dialogBox.appendChild(title);

           // Make the dialog box draggable
           let isDragging = false;
           dialogBox.addEventListener('mousedown', function(e) {
               isDragging = true;
           });
           document.addEventListener('mouseup', function(e) {
               isDragging = false;
           });
           dialogBox.addEventListener('mousemove', function(e) {
               if (isDragging) {
                   const x = e.clientX - dialogBox.offsetWidth / 2;
                   const y = e.clientY - dialogBox.offsetHeight / 2;
                   dialogBox.style.left = x.toString()+ 'px';
                   dialogBox.style.top = y.toString()+ 'px';
               }
           });

           // Close the dialog box when the user clicks on the close button
           const closeButton = document.createElement('button');
           closeButton.textContent = 'Close';
           closeButton.addEventListener('click', function(e) {
               dialogBox.remove();
           });
           dialogBox.appendChild(closeButton);

           // Add the dialog box to the document
           document.body.appendChild(dialogBox);
            // Make a GET request to the /lex/findbyform endpoint
            fetch('/lex/findbyform/' + theword)
                .then(response => response.json())
                .then(definitionreturned => {
                    // Update the content of the dialog box with the returned data
                    const newJS = definitionreturned['newjs'];
                    const newHTML = definitionreturned['newhtml'];
                    document.getElementById('lexmodalbody').innerHTML = newHTML;
                    document.getElementById('lexmodal').style.display = 'block';
                });
        });
    }`

	// OLDCLICKTOBROWSE - uses more memory
	OLDCLICKTOBROWSE = `
	$('#pollingdata').hide();

	var indexedLocations = document.querySelectorAll('%s');
	for (let i = 0; i < indexedLocations.length; i++) {
		const location = indexedLocations[i];
		location.addEventListener('click', function(e) {
			// Extract the ID from the clicked element's ID attribute
	
			// example: <td class="passages"><indexedlocation id="index/tlg0612/001/4270--31">11.11.1</indexedlocation></td>
			// indexedlocation[i].id will be "index/tlg0612/001/4270--31""
			// but you need to turn "index/tlg0612/001/4270--31" into "index/tlg0612/001/4270" to do the lookup
			// the "--31" is added to keep the ids distinct and it means "I was assigned by the 31st word in the index"
	
			const locus = location.id.split('--')[0];
			$.getJSON('/browse/'+locus, function (passagereturned) {
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
    }`

	// CLICKTOBROWSE - in Firefox an index to all of Ovid consumes 960MB of memory: 1.6M objects and 1.1M domNode
	// the older jQuery version of vv.CLICKTOBROWSE: 1.5GB of memory: 3.2M objects and 3.2M domNode; but there are big
	// potential problems here: you need to remove the eventlisteners and you cannot remove anonymous handlers ...
	CLICKTOBROWSE = `
	$('#pollingdata').hide();

    var indexedLocations = document.querySelectorAll('%s');
    for (let i = 0; i < indexedLocations.length; i++) {
        const location = indexedLocations[i];
        location.addEventListener('click', function(e) {
            // Extract the ID from the clicked element's ID attribute

            // example: <td class="passages"><indexedlocation id="index/tlg0612/001/4270--31">11.11.1</indexedlocation></td>
            // indexedlocation[i].id will be "index/tlg0612/001/4270--31""
            // but you need to turn "index/tlg0612/001/4270--31" into "index/tlg0612/001/4270" to do the lookup
            // the "--31" is added to keep the ids distinct and it means "I was assigned by the 31st word in the index"

            const locus = location.id.split('--')[0]; 
            fetch('/browse/' + locus)
                .then(response => response.json())
                .then(passagereturned => {
                    const fb = parsepassagereturned(passagereturned);
                    // left and right arrow keys
                    document.getElementById('browserdialogtext').addEventListener('keydown', e => {
                        switch (e.which) {
                            case 37: browseuponclick(fb[1]); break;
                            case 39: browseuponclick(fb[0]); break;
                        }
                    });
                    document.getElementById('browseforward').addEventListener('click', clickandbrowseforward(fb[0]));
                    document.getElementById('browseback').addEventListener('click', clickandbrowseback(fb[1]));
                })
        });
    }`

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
