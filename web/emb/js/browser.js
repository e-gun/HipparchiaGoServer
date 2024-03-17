//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)


function browseuponclick(url){
	$.getJSON(
	    { url: '/browse/' + url,
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

function rawlocusbrowseuponclick(url){
	$.getJSON(
	    { url: '/browse/rawlocus/' + url,
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
    const ldt = $('#lexicadialogtext');
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

    showmany(postbrowsepickui);

    $('observed').click( function(e) {
        e.preventDefault();
        let browsedauthorid = document.getElementById('browsertableuid').attributes.uid.value;
        var htxt = this.id;
        $.getJSON('/lex/findbyform/' + this.id + '/' + browsedauthorid, function (definitionreturned) {
            document.getElementById('leftmodalheadertext').innerHTML = htxt;
            document.getElementById('lexmodalbody').innerHTML = definitionreturned['newhtml'];
            document.getElementById('lexmodal').style.display = "block";
            jshld.html(definitionreturned['newjs']);
        });
        return false;
    });
	return [fwdurl, bkdurl]
}
