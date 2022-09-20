//
//	HipparchiaServer: an interface to a database of Greek and Latin texts
//	Copyright: E Gunderson 2016-22
//	License: License: GNU GENERAL PUBLIC LICENSE 3
//      (see LICENSE in the top level directory of the distribution)


//
// COMPLETE INDEX TO
//

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


$('#makeanindex').click( function() {
        $('#searchsummary').html('');
        $('#displayresults').html('');
        let searchid = generateId(8);
        let url = '/text/index/' + searchid;
        simpleactivityviawebsocket(searchid);
        $.getJSON(url, function (indexdata) {
            loadindexintodisplayresults(indexdata);
        });
});


function loadindexintodisplayresults(indexdata) {
        $('#searchsummary').html(indexdata['searchsummary']);
        $('#displayresults').html(indexdata['indexhtml']);
        let bcsh = document.getElementById("indexclickscriptholder");
        if (bcsh.hasChildNodes()) { bcsh.removeChild(bcsh.firstChild); }
        $('#indexclickscriptholder').html(indexdata['newjs']);
}


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
        loadtextintodisplayresults(returnedtext);
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
        loadtextintodisplayresults(returnedtext);
    });
});


function loadtextintodisplayresults(returnedtext) {
    $('#searchsummary').html(returnedtext['searchsummary']);
    $('#displayresults').html(returnedtext['texthtml']);
    $('#indexclickscriptholder').html(returnedtext['newjs']);
    }