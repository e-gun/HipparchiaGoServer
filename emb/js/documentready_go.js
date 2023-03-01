//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

$(document).ready( function () {

    $(document).keydown(function(e) {
        // 27 - escape
        // 38 & 40 - up and down arrow
        // 37 & 39 - forward and back arrow; but the click does not exist until you open a passage browser
        switch(e.which) {
            case 27: $('#lexmodal').hide(); break;
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

//
// SEARCHING
//
    // if you press "enter" when the cursor is in one of these boxes, execute a search
    document.getElementById('wordsearchform').addEventListener('keydown', searchifenterkeypressed);
    document.getElementById('proximatesearchform').addEventListener('keydown', searchifenterkeypressed);
    document.getElementById('lemmatasearchform').addEventListener('keydown', searchifenterkeypressed);
    document.getElementById('proximatelemmatasearchform').addEventListener('keydown', searchifenterkeypressed);

    function searchifenterkeypressed(e) {
        if (e.code === "Enter") {
            srch();
        }
    }

    $('#executesearch').click( function() {
        srch();
    });

    function srch() {
        $('#imagearea').empty();
        $('#searchsummary').html('');
        $('#displayresults').html('');

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
        let serverroot = '';
        let url = '';

        serverroot = '/srch/exec/';
        url = serverroot + searchid + '?' + qstring;

        // if (areWeWearchingVectors() === 0) {
        //     serverroot = '/srch/exec/';
        //     url = serverroot + searchid + '?' + qstring;
        // } else {
        //     let lsv = $('#lemmatasearchform').val();
        //     let vtype = whichVectorChoice();
        //     if (lsv.length === 0) { lsv = '_'; }
        //     serverroot = '/vectors/';
        //     url = serverroot + vtype + '/' + searchid + '/' + lsv;
        // }

        checkactivityviawebsocket(searchid);
        $.getJSON(url, function (returnedresults) { loadsearchresultsintodisplayresults(returnedresults); });
    }

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

    // function areWeWearchingVectors () {
    //     let xor = [];
    //     for (let i = 0; i < vectorboxes.length; i++) {
    //         let opt = $(vectorboxes[i]);
    //         if (opt.prop('checked')) { xor.push(1); }
    //         }
    //     return xor.length;
    // }

    function whichVectorChoice () {
        let xor = [];
        for (let i = 0; i < vectorboxes.length; i++) {
            let opt = $(vectorboxes[i]);
            if (opt.prop('checked')) { xor.push(vectorboxes[i].slice(1)); }
            }
        return xor[0];
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
// togglemany(vectorcheckboxspans);
closeextendedsearcharea();

if ($('#termoneisalemma').is(":checked")) {
    $('#termonecheckbox').show();
}


//
// PROGRESS INDICATOR
//

function checkactivityviawebsocket(searchid) {
    $.getJSON('/srch/conf/'+searchid, function(portnumber) {
        let pd = $('#pollingdata');
        pd.html('');
        pd.show();
        let ip = location.hostname;
        let s = new WebSocket('ws://'+ip+':'+portnumber+'/ws');
        s.onclose = function(e){ s = null; }
        s.onerror = function(e){ s.close(); s = null; }
        s.onopen = function(e){ s.send(JSON.stringify(searchid)); }
        s.onmessage = function(e){
            let progress = JSON.parse(e.data);
            // console.log(progress);
            if (progress['ID'] === searchid) {
                $('#pollingdata').html(progress['value']);
                if  (progress['close'] === 'close') { s.close(); s = null; }
            }
        }
    });
}

//
// AUTHENTICATION
//


$.getJSON('/auth/check', function(data){
    var u = data['userid'];
    var a = Boolean(data['authorized']);
    $('#userid').html(u);
    if (a !== true) {
        $('#executelogin').show();
        $('#executelogout').hide();
    } else {
        $('#executelogin').hide();
        $('#executelogout').show();
        }
    });

$('#validateusers').dialog({ autoOpen: false });

$('#executelogin').click( function() {
    $('#validateusers').dialog( 'open' );
});

$('#executelogout').click( function() {
    $.getJSON('/auth/logout', function(data){
        $('#userid').html(data);
    });
    $('#executelogout').hide();
    $('#executelogin').show();
});


//
// INDEXING, VOCAB LISTS, TEXT GENERATION
//

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
    checkactivityviawebsocket(searchid);
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
    checkactivityviawebsocket(searchid);
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

//
// COOKIES
//

$('#togglesaveslots').click( function(){ $('#saveslots').toggle()});

$('#toggleloadslots').click( function(){ $('#loadslots').toggle()});

function javascriptsessionintocookie(cookienumberstr){
    $.getJSON('/sc/set/'+cookienumberstr, function () {});
}

$('#save01').click( function(){ javascriptsessionintocookie('01'); $('#setoptions').hide(); $('#saveslots').hide(); });
$('#save02').click( function(){ javascriptsessionintocookie('02'); $('#setoptions').hide(); $('#saveslots').hide(); });
$('#save03').click( function(){ javascriptsessionintocookie('03'); $('#setoptions').hide(); $('#saveslots').hide(); });
$('#save04').click( function(){ javascriptsessionintocookie('04'); $('#setoptions').hide(); $('#saveslots').hide(); });
$('#save05').click( function(){ javascriptsessionintocookie('05'); $('#setoptions').hide(); $('#saveslots').hide(); });

// timing issues: you will get the cookie properly, but the selections will not show up right unless you use the misleadingly named .always()
//  'the .always() method replaces the deprecated .complete() method.'

$('#load01').click( function(){ $.getJSON('/sc/get/01').always( function() { $.getJSON('/selection/fetch', function(selectiondata) { reloadselections(selectiondata); }); location.reload(); }); });
$('#load02').click( function(){ $.getJSON('/sc/get/02').always( function() { $.getJSON('/selection/fetch', function(selectiondata) { reloadselections(selectiondata); }); location.reload(); }); });
$('#load03').click( function(){ $.getJSON('/sc/get/03').always( function() { $.getJSON('/selection/fetch', function(selectiondata) { reloadselections(selectiondata); }); location.reload(); }); });
$('#load04').click( function(){ $.getJSON('/sc/get/04').always( function() { $.getJSON('/selection/fetch', function(selectiondata) { reloadselections(selectiondata); }); location.reload(); }); });
$('#load05').click( function(){ $.getJSON('/sc/get/05').always( function() { $.getJSON('/selection/fetch', function(selectiondata) { reloadselections(selectiondata); }); location.reload(); }); });


//
// NON-VECTOR SPINNERS
//


$('#linesofcontextspinner').spinner({
    max: 20,
    min: 0,
    value: 2,
    step: 2,
    stop: function( event, ui ) {
        let result = $('#linesofcontextspinner').spinner('value');
        setoptions('linesofcontext', String(result));
    },
    spin: function( event, ui ) {
        let result = $('#linesofcontextspinner').spinner('value');
        setoptions('linesofcontext', String(result));
    }
});

$('#browserspinner').spinner({
    max: 50,
    min: 5,
    value: 1,
    stop: function( event, ui ) {
        let result = $('#browserspinner').spinner('value');
        setoptions('browsercontext', String(result));
    },
    spin: function( event, ui ) {
        let result = $('#browserspinner').spinner('value');
        setoptions('browsercontext', String(result));
    }
});

$( '#hitlimitspinner' ).spinner({
    min: 1,
    value: 1000,
    step: 50,
    stop: function( event, ui ) {
        let result = $('#hitlimitspinner').spinner('value');
        setoptions('maxresults', String(result));
    },
    spin: function( event, ui ) {
        let result = $('#hitlimitspinner').spinner('value');
        setoptions('maxresults', String(result));
    }
});

$( '#latestdate' ).spinner({
    min: -850,
    max: 1500,
    value: 1500,
    step: 50,
    stop: function( event, ui ) {
        let result = $('#latestdate').spinner('value');
        setoptions('latestdate', String(result));
        refreshselections();
    },
    spin: function( event, ui ) {
        let result = $('#latestdate').spinner('value');
        setoptions('latestdate', String(result));
        refreshselections();
    }
});


$( '#earliestdate' ).spinner({
    min: -850,
    max: 1500,
    value: -850,
    step: 50,
    stop: function( event, ui ) {
        let result = $('#earliestdate').spinner('value');
        setoptions('earliestdate', String(result));
        refreshselections();
    },
    spin: function( event, ui ) {
        let result = $('#earliestdate').spinner('value');
        setoptions('earliestdate', String(result));
        refreshselections();
    }
});


// 'width' property not working when you define the spinners
for (let i = 0; i < nonvectorspinners.length; i++) {
    const mywidth = 90;
    $(nonvectorspinners[i]).width(mywidth);
}


var lexmodal = document.getElementById("lexmodal");
var leftmodalclose = document.getElementById("leftmodalclose");

leftmodalclose.onclick = function() {
    lexmodal.style.display = "none";
}