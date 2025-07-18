//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)


function browseuponclick(url){
   // console.log("browseuponclick url is '" + url + "'");
    consolidatedbrowseonclick('/browse/', url)
}

function rawlocusbrowseuponclick(url){
    consolidatedbrowseonclick('/browse/rawlocus/', url);
}

function consolidatedbrowseonclick(pfx, url) {
    $.getJSON(
        { url: pfx + url,
            success: function (passagereturned) {
                let bf = $('#browseforward');
                let bb = $('#browseback');
                bf.unbind('click');
                bb.unbind('click');

                let fb = parsepassagereturned(passagereturned);
                // left and right arrow keys

                bf.bind('click', function(){ browseuponclick(fb[0]); });
                bb.bind('click', function(){ browseuponclick(fb[1]); });
            }
        }
    );
}

function parsepassagereturned(passagereturned) {
    const bdt = $('#browserdialogtext');
    const aac = $('#authorsautocomplete');
    const wac = $('#worksautocomplete');
    const jshld = $('#lexicaljsscriptholder');
    bdt.text('');
    let fwdurl = passagereturned['browseforwards'];  // e.g. 'linenumber/lt1254w001/4868'
    let bkdurl = passagereturned['browseback'];      // e.g. 'linenumber/lt1254w001/4840'

    resetworksautocomplete();
    aac.val(passagereturned['authorboxcontents']);
    aac.prop('placeholder', '');
    wac.val(passagereturned['workboxcontents']);
    wac.prop('placeholder', '');
    loadWorklist(passagereturned['authornumber']);
    if ($('#autofillinput').is(':checked')) {
        // autofill option
        loadLevellist(passagereturned['authornumber'], passagereturned['worknumber'], 'firstline');
    } else {
        // rawentry
        loadsamplecitation(passagereturned['authornumber'], passagereturned['worknumber']);
        $('#rawlocationinput').show();
    }

    bdt.html(passagereturned['browserhtml']);

    document.title = passagereturned['newtitle'];

    showmany(postbrowsepickui);

    let browsedauthorid = document.getElementById('browsertableuid').attributes.uid.value;
    const observed = document.querySelectorAll('observed');
    for (let i = 0; i < observed.length; i++) {
        const id = observed[i].id;
        // example: `<observed id="τράχηλον--205">τράχηλον</observed>`
        // observed[i].id will be `τράχηλον--205`
        // but you need to turn `τράχηλον--205` into `τράχηλον` to do the lookup
        const theword = id.split('--')[0];
        observed[i].addEventListener('click', function(e) {
            $.getJSON('/lex/findbyform/' + theword + '/' + browsedauthorid, function (definitionreturned) {
                document.getElementById('leftmodalheadertext').innerHTML = definitionreturned['entryname'];
                document.getElementById('lexmodalbody').innerHTML = definitionreturned['newhtml'];
                document.getElementById('lexmodal').style.display = "block";
                jshld.html(definitionreturned['newjs']);
            });
            return false;
        });
    }

	return [fwdurl, bkdurl]
}

function clickandbrowseforward(url) {
    // need a named function to add/remove eventlisteners; also called by vv.CLICKTOBROWSE for injected JS
    // anonymous functions produce event pile-ups
    browseuponclick(url);
}

function clickandbrowseback(url) {
    browseuponclick(url);
}