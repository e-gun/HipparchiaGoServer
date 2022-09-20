//
//	HipparchiaServer: an interface to a database of Greek and Latin texts
//	Copyright: E Gunderson 2016-22
//	License: License: GNU GENERAL PUBLIC LICENSE 3
//      (see LICENSE in the top level directory of the distribution)
//

$(document).ready( function () {

    $(document).keydown(function(e) {
        // 27 - escape
        // 38 & 40 - up and down arrow
        // 37 & 39 - forward and back arrow; but the click does not exist until you open a passage browser
        switch(e.which) {
            case 27: $('#browserdialog').hide(); break;
            case 37: $('#browseback').click(); break;
            case 39: $('#browseforward').click(); break;
            }
        });

    $('#clear_button').click( function() { window.location.href = '/reset/session'; });
    $('#alt_clear_button').click( function() { window.location.href = '/reset/session'; });
    $('#vectoralt_clear_button').click( function() { window.location.href = '/reset/session'; });
    $('#helptabs').tabs();
    $('#helpbutton').click( function() {
        if (document.getElementById('Interface').innerHTML === '<!-- placeholder -->') {
            $.getJSON('/get/json/helpdata', function (data) {
                let l = data.helpcategories.length;
                let hc = data.helpcategories;
                for (let i = 0; i < l; i++) {
                    if (data["HT"][hc[i]].length > 0) {
                        document.getElementById(hc[i]).innerHTML = data["HT"][hc[i]];
                        }
                    }
                });
            }
        $('#helptabs').toggle();
        $('#executesearch').toggle();
        $('#extendsearchbutton').toggle();
    });

    $('#extendsearchbutton-ispresentlyopen').click( function() {
        closeextendedsearcharea();
    });

    $('#extendsearchbutton-ispresentlyclosed').click( function() {
        openextendedsearcharea();
        });

    // https://stackoverflow.com/questions/1349404/generate-random-string-characters-in-javascript
    // dec2hex :: Integer -> String
    function dec2hex (dec) {
      return ('0' + dec.toString(16)).substr(-2);
    }

    // generateId :: Integer -> String
    function generateId (len) {
      let arr = new Uint8Array((len || 40) / 2);
      window.crypto.getRandomValues(arr);
      return Array.from(arr, dec2hex).join('');
    }

    function areWeWearchingVectors () {
        let xor = [];
        for (let i = 0; i < vectorboxes.length; i++) {
            let opt = $(vectorboxes[i]);
            if (opt.prop('checked')) { xor.push(1); }
            }
        return xor.length;
    }

    function whichVectorChoice () {
        let xor = [];
        for (let i = 0; i < vectorboxes.length; i++) {
            let opt = $(vectorboxes[i]);
            if (opt.prop('checked')) { xor.push(vectorboxes[i].slice(1)); }
            }
        return xor[0];
    }

    $('#executesearch').click( function(){
        $('#imagearea').empty();
        $('#searchsummary').html('');
        $('#displayresults').html('');

        let pd = $('#pollingdata');
        pd.html('');
        pd.show();

        // the script additions can pile up: so first kill off any scripts we have already added
        let bcsh = document.getElementById("browserclickscriptholder");
        if (bcsh.hasChildNodes()) { bcsh.removeChild(bcsh.firstChild); }

        const terms = {
            'skg': $('#wordsearchform').val(),
            'prx': $('#proximatesearchform').val(),
            'lem': $('#lemmatasearchform').val(),
            'plm': $('#proximatelemmatasearchform').val()
            };
        // disgustingly, if you send 'STRING ' to window.location it strips the whitespace and turns it into 'STRING'
        if (terms['skg'].slice(-1) === ' ') { terms['skg'] = terms['skg'].slice(0,-1) + '%20'; }
        if (terms['prx'].slice(-1) === ' ') { terms['prx'] = terms['prx'].slice(0,-1) + '%20'; }

        let qstringarray = Array();
        for (let t in terms) {
            if (terms[t] !== '') {qstringarray.push(t+'='+terms[t]); }
            }
        let qstring = qstringarray.join('&');

        let searchid = generateId(8);
        let flaskpath = '';
        let url = '';

        // let searchid = uuidv4();

        if (areWeWearchingVectors() === 0) {
            flaskpath = '/search/standard/';
            url = flaskpath + searchid + '?' + qstring;
        } else {
            let lsv = $('#lemmatasearchform').val();
            let vtype = whichVectorChoice();
            if (lsv.length === 0) { lsv = '_'; }
            flaskpath = '/vectors/';
            url = flaskpath + vtype + '/' + searchid + '/' + lsv;
        }

        checkactivityviawebsocket(searchid);
        $.getJSON(url, function (returnedresults) { loadsearchresultsintodisplayresults(returnedresults); });
        });

    function loadsearchresultsintodisplayresults(output) {
        document.title = output['title'];
        $('#searchsummary').html(output['searchsummary']);
        $('#displayresults').html(output['found']);

        //
        // THE GRAPH: if there is one... Note that if it is embedded in the output table, then
        // that table has to be created and  $('#imagearea') with it before you do any of the following
        //

        let imagetarget = $('#imagearea');
        if (typeof output['image'] !== 'undefined' && output['image'] !== '') {
            let w = window.innerWidth * .9;
            let h = window.innerHeight * .9;
            jQuery('<img/>').prependTo(imagetarget).attr({
                src: '/get/response/vectorfigure/' + output['image'],
                alt: '[vector graph]',
                id: 'insertedfigure',
                height: h
            });
        }

        //
        // JS UPDATE
        // [http://stackoverflow.com/questions/9413737/how-to-append-script-script-in-javascript#9413803]
        //

        let browserclickscript = document.createElement('script');
        browserclickscript.innerHTML = output['js'];
        document.getElementById('browserclickscriptholder').appendChild(browserclickscript);
    }

    // setoptions() defined in coreinterfaceclicks_go.js
    $('#searchlines').click( function(){ setoptions('searchscope', 'lines'); });
    $('#searchwords').click( function(){ setoptions('searchscope', 'words'); });

    $('#wordisnear').click( function(){ setoptions('nearornot', 'near'); });
    $('#wordisnotnear').click( function(){ setoptions('nearornot', 'notnear'); });

    $('#proximityspinner').spinner({
        min: 1,
        max: 10,
        value: 1,
        step: 1,
        stop: function( event, ui ) {
            let result = $('#proximityspinner').spinner('value');
            setoptions('proximity', String(result));
            },
        spin: function( event, ui ) {
            let result = $('#proximityspinner').spinner('value');
            setoptions('proximity', String(result));
            }
        });

    $('#browserclose').bind("click", function(){
    		$('#browserdialog').hide();
    		$('#browseback').unbind('click');
    		$('#browseforward').unbind('click');
    		}
		);
	});

loadoptions();

function checkCookie(){
    let c = navigator.cookieEnabled;
    if (!c){
        document.cookie = "testcookie";
        c = document.cookie.indexOf("testcookie")!=-1;
        document.cookie = "testcookie=1; expires=Thu, 01-Jan-1970 00:00:01 GMT";
    }

    if (c) {
        $('#cookiemessage').hide();
    } else {
        $('#cookiemessage').show();
    }
}

checkCookie();

hidemany(tohideonfirstload);
togglemany(vectorcheckboxspans);
closeextendedsearcharea();

if ($('#termoneisalemma').is(":checked")) {
    $('#termonecheckbox').show(); }


//
// authentication
//


$.getJSON('/authentication/checkuser', function(data){
    var u = data['userid'];
    $('#userid').html(u);
    if (u === 'Anonymous') {
        $('#executelogin').show();
        $('#executelogout').hide();
        } else {
        $('#executelogin').hide();
        $('#executelogout').show();
        }
    });

function dec2hex (dec) {
    return ('0' + dec.toString(16)).substr(-2);
}

// generateId :: Integer -> String
function generateId (len) {
    let arr = new Uint8Array((len || 40) / 2);
    window.crypto.getRandomValues(arr);
    return Array.from(arr, dec2hex).join('');
}


$('#makeanindex').click( function() {
    $('#searchsummary').html('');
    $('#displayresults').html('');
    let searchid = generateId(8);
    let url = '/text/index/' + searchid;
    simpleactivityviawebsocket(searchid);
    $.getJSON(url, function (indexdata) {
        loadintodisplayresults(indexdata);
    });
});

//
// VOCABLISTS
//

$('#makevocablist').click( function() {
    $('#searchsummary').html('');
    $('#displayresults').html('');
    let searchid = generateId(8);
    let url = '/text/vocab/' + searchid;
    simpleactivityviawebsocket(searchid);
    $.getJSON(url, function (returnedtext) {
        loadintodisplayresults(returnedtext);
    });

});


//
// TEXTMAKER
//

$('#textofthis').click( function() {
    $('#searchsummary').html('');
    $('#displayresults').html('');

    let url = '/text/make/_';
    $.getJSON(url, function (returnedtext) {
        loadintodisplayresults(returnedtext);
    });
});


function loadintodisplayresults(indexdata) {
    $('#searchsummary').html(indexdata['searchsummary']);
    $('#displayresults').html(indexdata['thehtml']);
    let bcsh = document.getElementById("indexclickscriptholder");
    if (bcsh.hasChildNodes()) { bcsh.removeChild(bcsh.firstChild); }
    $('#indexclickscriptholder').html(indexdata['newjs']);
}
