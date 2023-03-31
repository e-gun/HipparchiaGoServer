//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

//
// the radio ui-click options
//

$('#headwordindexing_y').click( function(){
    setoptions('headwordindexing', 'yes'); $('#headwordindexingactive').show(); $('#headwordindexinginactive').hide();
});

$('#headwordindexing_n').click( function(){
    setoptions('headwordindexing', 'no'); $('#headwordindexinginactive').show(); $('#headwordindexingactive').hide();
});

$('#frequencyindexing_y').click( function(){
    setoptions('indexbyfrequency', 'yes'); $('#frequencyindexingactive').show(); $('#frequencyindexinginactive').hide();
});

$('#frequencyindexing_n').click( function(){
    setoptions('indexbyfrequency', 'no'); $('#frequencyindexinginactive').show(); $('#frequencyindexingactive').hide();
});


$('#onehit_y').click( function(){
    setoptions('onehit', 'yes'); $('#onehitistrue').show(); $('#onehitisfalse').hide();
});

$('#onehit_n').click( function(){
    setoptions('onehit', 'no'); $('#onehitisfalse').show(); $('#onehitistrue').hide();
});

$('#autofillinput').click( function(){
    setoptions('rawinputstyle', 'no'); $('#usingautoinput').show(); $('#usingrawinput').hide();
    hidemany(rawinputuielements);
});

$('#manualinput').click( function(){
    setoptions('rawinputstyle', 'yes'); $('#usingrawinput').show(); $('#usingautoinput').hide();
    let elementarray = Array().concat(levelsids, endpointlevelssids);
    hidemany(elementarray);
});

$('#includespuria').change(function() {
    if(this.checked) { setoptions('spuria', 'yes'); } else { setoptions('spuria', 'no'); }
    refreshselections();
    loadoptions();
    });

$('#includevaria').change(function() {
    if(this.checked) { setoptions('varia', 'yes'); } else { setoptions('varia', 'no'); }
    refreshselections();
    loadoptions();
    });

$('#includeincerta').change(function() {
    if(this.checked) { setoptions('incerta', 'yes'); } else { setoptions('incerta', 'no'); }
    refreshselections();
    loadoptions();
    });

$('#greekcorpus').change(function() {
    if(this.checked) { setoptions('greekcorpus', 'yes'); } else { setoptions('greekcorpus', 'no'); }
    refreshselections();
    loadoptions();
    });

$('#latincorpus').change(function() {
    if(this.checked) { setoptions('latincorpus', 'yes'); } else { setoptions('latincorpus', 'no'); }
    refreshselections();
    loadoptions();
    });

$('#inscriptioncorpus').change(function() {
    if(this.checked) { setoptions('inscriptioncorpus', 'yes'); } else { setoptions('inscriptioncorpus', 'no'); }
    refreshselections();
    loadoptions();
    });

$('#papyruscorpus').change(function() {
    if(this.checked) { setoptions('papyruscorpus', 'yes'); } else { setoptions('papyruscorpus', 'no'); }
    refreshselections();
    loadoptions();
    });

$('#christiancorpus').change(function() {
    if(this.checked) { setoptions('christiancorpus', 'yes'); } else { setoptions('christiancorpus', 'no'); }
    refreshselections();
    loadoptions();
    });

$('#vocbycount').change(function() {
    if(this.checked) { setoptions('vocbycount', 'yes'); } else { setoptions('vocbycount', 'no'); }
    refreshselections();
    loadoptions();
});

$('#vocscansion').change(function() {
    if(this.checked) { setoptions('vocscansion', 'yes'); } else { setoptions('vocscansion', 'no'); }
    refreshselections();
    loadoptions();
});

$('#extendedgraph').change(function() {
    if(this.checked) { setoptions('extendedgraph', 'yes'); } else { setoptions('extendedgraph', 'no'); }
    refreshselections();
    loadoptions();
});

// lemmata and vectors

$('#isvectorsearch').change(function() {
    if(this.checked) {
        setoptions('isvectorsearch', 'yes');
        showvectornotification();
        refreshselections();
        loadoptions();
        lsf.attr('placeholder', '(semantic neighbors of...)');
    } else {
        setoptions('isvectorsearch', 'no');
        hidevectornotification();
        refreshselections();
        loadoptions();
        lsf.attr('placeholder', '(all forms of...)');
    }
});

const lsf = $('#lemmatasearchform');
const vschon = $('#vectorizing-ison');
const vschoff  = $('#vectorizing-isoff');

function hidevectornotification() {
    vschon.hide();
    vschoff.show();
}

function showvectornotification() {
    vschon.show();
    vschoff.hide();
}


const lschon = $('#lemmatizing-ison');
const lschoff= $('#lemmatizing-isoff');

function hidelemmatanotification() {
    lschon.hide();
    lschoff.show();
}

function showlemmatanotification() {
    lschon.show();
    lschoff.hide();
}

const trmonelem = $('#termoneisalemma');
const trmtwolem = $('#termtwoisalemma');
const wsf = $('#wordsearchform');

trmonelem.change(function() {
    if(this.checked) {
        wsf.hide();
        wsf.val('');
        lsf.show();
        vct.show();
        showlemmatanotification();
    } else {
        lsf.hide();
        lsf.val('');
        wsf.show();
        vct.hide();
        setoptions('isvectorsearch', 'no');
        loadoptions();
        hidelemmatanotification();
        if(!trmtwolem.is(':checked')) {
            hidelemmatanotification();
        }
    }
});

trmtwolem.change(function() {
    if(this.checked) {
        psf.hide();
        psf.val('');
        plsf.show();
        showlemmatanotification();
    } else {
        plsf.hide();
        plsf.val('');
        psf.show();
        if(!trmonelem.is(':checked')) {
            hidelemmatanotification();
        }
    }
});
