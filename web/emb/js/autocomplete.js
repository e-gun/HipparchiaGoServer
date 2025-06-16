//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-25
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

// this is how you can allow shifts from a universalid of `gr0001w001` to `tlg0001w001`
// NB: coreinterfaceclicks.js also uses these values

const WORKNAMELEN = 3;
const CORPPREFIXLEN = 2;
const AUTHNAMELEN = 4;
const AUIDLEN = CORPPREFIXLEN + AUTHNAMELEN;

const NUMBEROFLEVELS = 6;

//
// AUTHORS
//

function reloadselections(selectiondata){
    hidemany(endpointbuttons);
    let holder = $('#selectionscriptholder');
    // the data comes back from the server as a dict with three keys: timeexclusions, selections, exclusions
    if (selectiondata.numberofselections > 0) {
            $('#selectionstable').show();
        } else {
            $('#selectionstable').hide();
        }
    $('#timerestrictions').html(selectiondata.timeexclusions);
    $('#selectioninfocell').html(selectiondata.selections);
    $('#exclusioninfocell').html(selectiondata.exclusions);
    holder.children().remove();
    holder.html(selectiondata['newjs']);
    }

function reloadAuthorlist(){
    let ids = Array('#lexica', '#authorholdings', '#selectionendpoint');
    hidemany(ids);
    hidemany(actionbuttons);
    hidemany(nonessentialautofills);
    hidemany(loadandsaveslots);
    hidemany(levelsids);
    hidemany(endpointnoticesandbuttons);
    $.getJSON('/selection/fetch', function (selectiondata) {
        reloadselections(selectiondata);
    });
}

function resetworksautocomplete(){
    showmany(postauthorpickui);
    hidemany(inputids);
    hidemany(endpointids);
    hidemany(endpointnotices);
    clearmany(inputids);
    clearmany(endpointids);
}


$('#authorsautocomplete').autocomplete({
    change: reloadAuthorlist(),
    source: "/hints/author/_",
    select: function (event, ui) {
        let thisselector = $('#authorsautocomplete');
        let selector = $('#worksautocomplete');
        selector.val('');
        resetworksautocomplete();
        // stupid timing issue if you select with mouse instead of keyboard: nothing happens
        // see: http://stackoverflow.com/questions/9809013/jqueryui-autocomplete-select-does-not-work-on-mouse-click-but-works-on-key-eve
        let origEvent = event;
        let auid = '';
        while (origEvent.originalEvent !== undefined){ origEvent = origEvent.originalEvent; }

        // console.log(thisselector.val());
        // `M. Tullius Cicero [lt0474]`
        // and you are targeting `lt0474` in that string

        if (origEvent.type === 'click'){
            document.getElementById('authorsautocomplete').value = ui.item.value;
            auid = thisselector.val().slice(-1 * (AUIDLEN + 1), -1);
        } else {
            auid = thisselector.val().slice(-1 * (AUIDLEN + 1), -1);
        }
        loadWorklist(auid);
        selector.prop('placeholder', '(Pick a work)');
        $('#authorendpoint').val(thisselector.val());
        clearmany(allautofills);
        }
    });


$('#addauthortosearchlist').click( function() {
        let name = $('#authorsautocomplete').val();
        let authorid = name.slice(-1 * (AUIDLEN + 1), -1);
        let locus = locusdataloader();
        let endpoint = endpointdataloader();

        // console.log($('#worksautocomplete').val());
        // `Academica (w045)`
        // and you are targeting `045` in that string

        let wrk = $('#worksautocomplete').val().slice(-1 * (WORKNAMELEN + 1), -1); // want the `001` in `gr0007w001`
        let rawlocus = $('#rawlocationinput').val();
        let rawendpoint = $('#rawendpointinput').val();
        if ($('#endpointnotice').is(':hidden')) {
            rawendpoint = '';
            endpoint = '';
        }
        resetworksautocomplete();
        if (authorid !== '') {
            if (wrk === '') {
                // note the '_' is safe from depunct() but '.' is not and that you have to have at least one character
                $.getJSON('/selection/make/_?auth=' + authorid, function (selectiondata) {
                    reloadselections(selectiondata);
                    loadWorklist(authorid);
                    $('#worksautocomplete').prop('placeholder', '(Pick a work)');
                });
            } else if ($('#autofillinput').is(':checked')) {
                // you are using the autofill boxes
                if (locus === '') {
                   $.getJSON('/selection/make/_?auth=' + authorid + '&work=' + wrk, function (selectiondata) {
                       reloadselections(selectiondata);
                   });
                } else if (locus === endpoint){
                   $.getJSON('/selection/make/_?auth=' + authorid + '&work=' + wrk + '&locus=' + locus, function (selectiondata) {
                       reloadselections(selectiondata);
                   });
                } else {
                   $.getJSON('/selection/make/_?auth=' + authorid + '&work=' + wrk + '&locus=' + locus + '&endpoint=' + endpoint, function (selectiondata) {
                       reloadselections(selectiondata);
                   });
                }
            } else {
                // you are using the raw entry subsystem
                if (rawlocus === '') {
                   $.getJSON('/selection/make/_?auth=' + authorid + '&work=' + wrk, function (selectiondata) {
                       reloadselections(selectiondata);
                       });
                } else if (rawendpoint === '') {
                   $.getJSON('/selection/make/_?auth=' + authorid + '&work=' + wrk + '&locus=' + rawlocus + '&raw=t', function (selectiondata) {
                       reloadselections(selectiondata);
                   });
                } else {
                   $.getJSON('/selection/make/_?auth=' + authorid + '&work=' + wrk + '&locus=' + rawlocus + '&endpoint=' + rawendpoint + '&raw=t', function (selectiondata) {
                       reloadselections(selectiondata);
                   });
                }
            }
            hidemany(endpointnotification);
        }
        $('#searchlistcontents').hide();
});


$('#excludeauthorfromsearchlist').click( function() {
        let name = $('#authorsautocomplete').val();
        let authorid = name.slice(-1 * (AUIDLEN + 1), -1);
        let locus = locusdataloader();
        let endpoint = endpointdataloader();
        let wrk = $('#worksautocomplete').val().slice(-1 * (WORKNAMELEN + 1), -1);
        let rawlocus = $('#rawlocationinput').val();
        let rawendpoint = $('#rawendpointinput').val();
        resetworksautocomplete();
        if (authorid !== '') {
            if (wrk === '') {
                $.getJSON('/selection/make/_?auth=' + authorid + '&exclude=t', function (selectiondata) {
                    reloadselections(selectiondata);
                    loadWorklist(authorid);
                    $('#worksautocomplete').prop('placeholder', '(Pick a work)');
                });
            } else if ($('#autofillinput').is(':checked')) {
                // you are using the autofill boxes
                if (locus === '') {
                    $.getJSON('/selection/make/_?auth=' + authorid + '&work=' + wrk + '&exclude=t', function (selectiondata) {
                        reloadselections(selectiondata);
                    });
                } else if (locus === endpoint) {
                    $.getJSON('/selection/make/_?auth=' + authorid + '&work=' + wrk + '&locus=' + locus + '&exclude=t', function (selectiondata) {
                        reloadselections(selectiondata);
                    });
                } else {
                    $.getJSON('/selection/make/_?auth=' + authorid + '&work=' + wrk + '&locus=' + locus + '&endpoint=' + endpoint + '&exclude=t', function (selectiondata) {
                        reloadselections(selectiondata);
                    });
                }
            } else {
                // you are using the raw entry subsystem
                if (rawlocus === '') {
                    $.getJSON('/selection/make/_?auth=' + authorid + '&work=' + wrk + '&exclude=t', function (selectiondata) {
                        reloadselections(selectiondata);
                    });
                } else if (rawendpoint === '') {
                    $.getJSON('/selection/make/_?auth=' + authorid + '&work=' + wrk + '&locus=' + rawlocus + '&raw=t' + '&exclude=t', function (selectiondata) {
                        reloadselections(selectiondata);
                    });
                } else {
                    $.getJSON('/selection/make/_?auth=' + authorid + '&work=' + wrk + '&locus=' + rawlocus + '&endpoint=' + rawendpoint + '&raw=t' + '&exclude=t', function (selectiondata) {
                        reloadselections(selectiondata);
                    });
                }
                $('#fromnotice').hide();
                $('#endpointnotice').hide();
                $('#endpointbutton-isclosed').hide();
                $('#endpointbutton-isopen').hide();
            }
        }
        $('#searchlistcontents').hide();
});


//
// WORKS
//

function loadWorklist(authornumber){
    $('#fromnotice').hide();
    $('#endpointnotice').hide();
    $('#endpointbutton-isclosed').hide();
    $('#endpointbutton-isopen').hide();
    $.getJSON('/get/json/worksof/'+authornumber, function (selectiondata) {
        let dLen = selectiondata.length;
        let worksfound = Array();
        let wac = $('#worksautocomplete');
        for (let i = 0; i < dLen; i++) { worksfound.push(selectiondata[i]); }
        wac.autocomplete( "enable" );
        wac.autocomplete({ source: worksfound });
    });
}


$('#worksautocomplete').autocomplete({
    focus: function (event, ui) {
        resetworksautocomplete();
        let auth = $("#authorsautocomplete").val().slice(-1 * (AUIDLEN + 1), -1);
        let wrk = ui.item.value.slice(-1 * (WORKNAMELEN + 1), -1);
        if ($('#autofillinput').is(':checked')) {
            loadLevellist(auth, wrk, 'firstline');
        } else {
            $('#rawlocationinput').show();
            loadsamplecitation(auth, wrk);
        }
        },
     select: function (event, ui) {
        let thisselector = $('#worksautocomplete');
        $('#workendpoint').val(thisselector.val());
        toggleendpointarrows();
     }
});

function toggleendpointarrows() {
    let isclosed = $('#endpointbutton-isclosed');
    let isopen = $('#endpointbutton-isopen');
    if ( isclosed.is(':hidden') && isopen.is(':hidden')) {
        isclosed.show();
    }
}


//
// LEVELS
//

// async issue: if you type "1" then click "12" from the pulldown, locusdataloader() can read "1" before "12" arrives
// and so you will get the values appropriate to "1" in the next box not the values that go with "12"

function waitAndReturnLocusValue() {
    return new Promise(resolve => {
        setTimeout(() => {
            let loc = locusdataloader();
            resolve(loc);  // This returns the value after 3ms
        }, 3);
    });
}

function waitAndReturnEndPointLocusValue() {
    return new Promise(resolve => {
        setTimeout(() => {
            let loc = endpointdataloader();
            resolve(loc);  // This returns the value after 3ms
        }, 3);
    });
}

function locusdataloader() {
    let l5 = document.getElementById("level05").value;
    let l4 = document.getElementById("level04").value;
    let l3 = document.getElementById("level03").value;
    let l2 = document.getElementById("level02").value;
    let l1 = document.getElementById("level01").value;
    let l0 = document.getElementById("level00").value;
    let lvls = [ l5, l4, l3, l2, l1, l0];
    let locusdata = '';
    for (let i = 0; i < NUMBEROFLEVELS; i++ ) {
        if (lvls[i] !== '') { locusdata += lvls[i]+'|' } }
    locusdata = locusdata.slice(0, (locusdata.length)-1);
    return locusdata;
}

function endpointdataloader() {
    let l5 = $('#level05endpoint').val();
    let l4 = $('#level04endpoint').val();
    let l3 = $('#level03endpoint').val();
    let l2 = $('#level02endpoint').val();
    let l1 = $('#level01endpoint').val();
    let l0 = $('#level00endpoint').val();
    let lvls = [ l5, l4, l3, l2, l1, l0];
    let locusdata = '';
    for (let i = 0; i < NUMBEROFLEVELS; i++ ) {
        if (lvls[i] !== '') { locusdata += lvls[i]+'|' } }
    locusdata = locusdata.slice(0, (locusdata.length)-1);

    return locusdata;
}

function loadsamplecitation(author, work) {
    // we are using the maual input style on the web page
    // so we need some hint on how to do things: check the end line for a sample citation
    // "Cic., In Verr" ==> 2.5.189.7
    // all of these '/get/json/...' URLS end up at infogetter() in getteroutes.py
    $.getJSON('/get/json/samplecitation/' + author + '/' + work, function (citationdata) {
        let firstline = citationdata['firstline'];
        let lastline = citationdata['lastline'];
        $('#rawlocationinput').prop('placeholder', '(' + firstline + ' to ' + lastline + ')');
        $('#rawendpointinput').prop('placeholder', '(' + firstline + ' to ' + lastline + ')');
    });
}

function loadLevellist(author, work, pariallocus){
    // the server is hoping to be sent something like:
    //
    //  /get/json/workstructure/lt1254/001
    //  /get/json/workstructure/lt0474/043/3|12
    //
    // bad things happen if you send level00 info
    //
    // the server will return info about the next level down such as:
    //  ws =  {'totallevels': 5, 'level': 2, 'label': 'par', 'low': '1', 'high': '10', 'range': ['1', '2', '3', '4', '5', '6', '7', '8', '9', '10']}

    let getpath = '';
    if ( pariallocus !== 'firstline' ) {
        getpath = author + '/' + work + '/' + pariallocus;
    } else {
        getpath = author + '/' + work;
    }

    let openbutton = $('#endpointbutton-isclosed');
    let closebutton = $('#endpointbutton-isopen');
    let workboxval = $('#worksautocomplete').val();

    $.getJSON('/get/json/workstructure/' + getpath, function (selectiondata) {
        let top = selectiondata['totallevels']-1;
        let atlevel = selectiondata['level'];
        let label = selectiondata['label'];
        let low = selectiondata['low'];
        let high = selectiondata['high'];
        let possibilities = selectiondata['range'];

        let generateme = '#level0'+String(atlevel);

        if ( low !== '-9999') { $(generateme).prop('placeholder', '('+label+' '+String(low)+' to '+String(high)+')'); }
        else { $(generateme).prop('placeholder', '(awaiting a valid selection...)'); }

        $(generateme).show();
        $(generateme).autocomplete ({
            source: possibilities, // alas: the select function can execute before the possibilities have loaded...
            select: async function (event, ui) {
                // as per the preceding: if select() executes before source() is done resetting the
                // document.getElementById("level0N").value, you will read bad values off of getElementById()
                const loc = await waitAndReturnLocusValue();
                loadLevellist(author, work, loc);
                if ( atlevel && workboxval && openbutton.is(':hidden') === true && closebutton.is(':hidden') === true) {
                    openbutton.show();
                }
                if (!workboxval) { openbutton.hide(); closebutton.hide(); }
            }});

        let endpointlevel = atlevel+1;
        let endpointbox = '#level0'+String(endpointlevel)+'endpoint';
        let startpointbox = '#level0'+endpointlevel;
        $(endpointbox).val($(startpointbox).val());

        let generateendpoint = '#level0'+String(atlevel)+'endpoint';
        if ( low !== '-9999') { $(generateendpoint).prop('placeholder', '('+label+' '+String(low)+' to '+String(high)+')'); }
        else { $(generateendpoint).prop('placeholder', '(awaiting a valid selection...)'); }
        $(generateendpoint).autocomplete ({
            source: possibilities,
            select: async function (event, ui) {
                const loc = await waitAndReturnEndPointLocusValue();
                endpointloadLevellist(author, work, String(loc));
            }});
    });
}


function endpointloadLevellist(author, work, pariallocus){
    // python is hoping to be sent something like:
    //
    //  /get/json/workstructure/lt0474/043/3|12
    //
    // bad things happen if you send level00 info
    //
    // python will return info about the next level down such as:
    //  ws =  {'totallevels': 5, 'level': 2, 'label': 'par', 'low': '1', 'high': '10', 'range': ['1', '2', '3', '4', '5', '6', '7', '8', '9', '10']}

    let getpath = '';
    if ( pariallocus !== 'firstline' ) {
        getpath = author + '/' + work + '/' + pariallocus;
    } else {
        getpath = author + '/' + work;
    }

    $.getJSON('/get/json/workstructure/' + getpath, function (selectiondata) {
        let top = selectiondata['totallevels']-1;
        let atlevel = selectiondata['level'];
        let label = selectiondata['label'];
        let low = selectiondata['low'];
        let high = selectiondata['high'];
        let possibilities = selectiondata['range'];

        let generateme = '#level0'+String(atlevel)+'endpoint';

        if ( low !== '-9999') { $(generateme).prop('placeholder', '('+label+' '+String(low)+' to '+String(high)+')'); }
        else { $(generateme).prop('placeholder', '(awaiting a valid selection...)'); }

        $(generateme).show();
        $(generateme).autocomplete ({
            source: possibilities,
            select: async function (event, ui) {
                const loc = await waitAndReturnEndPointLocusValue();
                endpointdataloader(author, work, String(loc));
            }});
    });
}

//
// GENRES
//

$('#genresautocomplete').autocomplete({
    source: '/hints/authgenre/_'
    });

$('#workgenresautocomplete').autocomplete({
    source: '/hints/workgenre/_'
    });

$('#locationsautocomplete').autocomplete({
    source: '/hints/authlocation/_'
    });

$('#provenanceautocomplete').autocomplete({
    source: '/hints/worklocation/_'
    });

$('#pickgenrebutton').click( function() {
        let genre = $('#genresautocomplete').val();
        let wkgenre = $('#workgenresautocomplete').val();
        let loc = $('#locationsautocomplete').val();
        let prov = $('#provenanceautocomplete').val();

        if (genre !== undefined && genre !== '') {
            $.getJSON('/selection/make/_?genre=' + genre, function (selectiondata) {
                reloadselections(selectiondata);
             });
        }
        if (wkgenre !== undefined && wkgenre !== '') {
            $.getJSON('/selection/make/_?wkgenre=' + wkgenre, function (selectiondata) {
                reloadselections(selectiondata);
             });
        }
        if (loc !== undefined && loc !== '') {
            $.getJSON('/selection/make/_?auloc=' + loc, function (selectiondata) {
                reloadselections(selectiondata);
             });
        }
        if (prov !== undefined && prov !== '') {
            $.getJSON('/selection/make/_?wkprov=' + prov, function (selectiondata) {
                reloadselections(selectiondata);
             });
        }
        clearmany(categoryautofills);
        $('#searchlistcontents').hide();
    });



$('#excludegenrebutton').click( function() {
        let genre = $('#genresautocomplete').val();
        let wkgenre = $('#workgenresautocomplete').val();
        let loc = $('#locationsautocomplete').val();
        let prov = $('#provenanceautocomplete').val();

        if (genre !== undefined && genre !== '') {
            $.getJSON('/selection/make/_?genre=' + genre +'&exclude=t', function (selectiondata) {
                reloadselections(selectiondata);
             });
        }
        if (wkgenre !== undefined && wkgenre !== '') {
            $.getJSON('/selection/make/_?wkgenre=' + wkgenre +'&exclude=t', function (selectiondata) {
                reloadselections(selectiondata);
             });
        }

        if (loc !== undefined && loc !== '') {
            $.getJSON('/selection/make/_?auloc=' + loc +'&exclude=t', function (selectiondata) {
                reloadselections(selectiondata);
             });
        }
        if (prov !== undefined && prov !== '') {
            $.getJSON('/selection/make/_?wkprov=' + prov +'&exclude=t', function (selectiondata) {
                reloadselections(selectiondata);
             });
        }
        clearmany(categoryautofills);
        $('#searchlistcontents').hide();
    });


//
// LEMMATA
//

$('#lemmatasearchform').autocomplete({
    source: '/hints/lemmata/_'
    });

$('#proximatelemmatasearchform').autocomplete({
    source: '/hints/lemmata/_'
    });

$('#analogiesinputA').autocomplete({
    source: '/hints/lemmata/_'
    });

$('#analogiesinputB').autocomplete({
    source: '/hints/lemmata/_'
    });

$('#analogiesinputC').autocomplete({
    source: '/hints/lemmata/_'
    });
