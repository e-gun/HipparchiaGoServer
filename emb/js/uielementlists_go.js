//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-23
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

//
// BULK OPERATIONS ON ARRAYS OF ELEMENTS
//

function hidemany(arrayofelements) {
    for (let i = 0; i < arrayofelements.length; i++) {
        $(arrayofelements[i]).hide();
        }
}

function showmany(arrayofelements) {
    for (let i = 0; i < arrayofelements.length; i++) {
        $(arrayofelements[i]).show();
        }
}

function hidemanyrevealone(arrayofelementstohide, elementtoshow) {
    for (let i = 0; i < arrayofelementstohide.length; i++) {
        if (arrayofelementstohide[i] === elementtoshow) {
            $(arrayofelementstohide[i]).show();
        } else {
            $(arrayofelementstohide[i]).hide();
        }
        }
}

function clearmany(arrayofelements) {
    for (let i = 0; i < arrayofelements.length; i++) {
        $(arrayofelements[i]).val('');
        }
}

function togglemany(arrayofelements) {
        for (let i = 0; i < arrayofelements.length; i++) {
        $(arrayofelements[i]).toggle();
        }
}

function setoptions(sessionvar, value){
	$.getJSON( {url: '/setoption/' + sessionvar + '/' + value,
	    async: false,
	    success: function (resultdata) {
		 // do nothing special: the return exists but is not relevant
	    }
	    });
}

function setmultiple(arrayofvars, value) {
    for (let i = 0; i < arrayofvars.length; i++) {
        setoptions(arrayofvars[i], value);
        }
}

//
// ID COLLECTIONS
//

// searchfield.html structure

const toplevelofsearchfieldhtml = Array('#searchfield', '#authorholdings', '#selectionendpoint');

// passage selection UI

const levelsids = Array('#level05', '#level04', '#level03', '#level02', '#level01', '#level00');

const inputids = levelsids.concat(Array('#rawlocationinput'));

const endpointlevelssids = Array('#level05endpoint', '#level04endpoint', '#level03endpoint', '#level02endpoint',
    '#level01endpoint', '#level00endpoint');

const endpointids = endpointlevelssids.concat(Array('#rawendpointinput', '#authorendpoint', '#workendpoint'));

const endpointnotices = Array('#endpointnotice', '#fromnotice');

const endpointbuttons = Array('#endpointbutton-isopen', '#endpointbutton-isclosed');

const endpointnotification = Array().concat(endpointnotices, endpointbuttons);

const rawinputboxes = Array('#rawlocationinput', '#rawendpointinput');

const rawinputuielements = Array().concat(rawinputboxes, endpointnotification);

const endpointnoticesandbuttons = endpointnotices.concat(endpointbuttons);

// category selection ui

const categoryautofills = Array('#genresautocomplete', '#workgenresautocomplete', '#locationsautocomplete',
    '#provenanceautocomplete');

const nonessentialautofills = Array().concat(categoryautofills, ['#worksautocomplete']);

const allautofills = Array().concat(nonessentialautofills, ['#authorsautocomplete']);

// info buttons

const infobuttons = Array('#authinfobutton', '#genreinfobutton');

// action buttons

const coreactionbuttons = Array('#addauthortosearchlist', '#excludeauthorfromsearchlist');
const extendedactionbuttons = Array('#browseto', '#fewerchoices');
const genreselectbuttons = Array('#pickgenrebutton', '#excludegenrebutton');
const actionbuttons = Array().concat(coreactionbuttons, extendedactionbuttons, genreselectbuttons);

// active options

const lemmatabagoptions = Array('#winnertakesall-ison', '#unlemmatized-ison', '#flatlemma-ison', '#alternates-ison');
const trimmingoptions = Array('#trimming-none', '#trimming-declined', '#trimming-conjugated');
// datespinners and includespuria checkboxes

const datespinners = Array('#edts', '#ldts');
const extrasearchcriteria = Array().concat(datespinners, ['#spuriacheckboxes']);

// infoboxes + infotables

const infoboxes = Array('#genrelistcontents', '#selectionstable', '#searchlistcontents');

// loadandsave UI

const loadandsaveslots = Array('#loadslots', '#saveslots');

// searchforms

const extrasearchform = Array('#proximatesearchform');
const lemmatasearchforms = Array('#lemmatasearchform', '#proximatelemmatasearchform');
const allextrasearchfroms = Array().concat(extrasearchform, lemmatasearchforms);
const extrasearchuielements = Array('#nearornot', '#termonecheckbox', '#termtwocheckbox', '#complexsearching');

// spinners

const nonvectorspinners = ["#earliestdate", "#latestdate", "#hitlimitspinner", "#linesofcontextspinner", "#browserspinner"];

// vectors

const vectorcheckboxspans = ['#cosinedistancesentencecheckbox', '#cosinedistancelineorwordcheckbox', '#semanticvectorquerycheckbox',
    '#semanticvectornnquerycheckbox', '#tensorflowgraphcheckbox', '#sentencesimilaritycheckbox', '#topicmodelcheckbox',
    '#analogiescheckbox', '#vectortestcheckbox'];

const vectorboxes = ['#cosdistbysentence', '#cosdistbylineorword', '#semanticvectorquery', '#nearestneighborsquery',
    '#tensorflowgraph', '#sentencesimilarity', '#topicmodel', '#vectortestfunction'];

const vectorformattingdotpyids = Array(['#analogiescheckbox', '#analogyfinder', '#cosdistbylineorword',
    '#cosdistbysentence', '#cosinedistancelineorwordcheckbox', '#cosinedistancesentencecheckbox', '#nearestneighborsquery',
    '#semanticvectornnquerycheckbox', '#semanticvectorquery', '#semanticvectorquerycheckbox', '#sentencesimilarity',
    '#sentencesimilaritycheckbox', '#tensorflowgraph', '#tensorflowgraphcheckbox', '#topicmodel', '#topicmodelcheckbox',
    '#vectortestcheckbox', '#vectortestfunction']);

// the checkbox names can be found via: vectorhtmlforfrontpage() in vectorformatting.py
// >>> f = re.compile(r'type="checkbox" id="(.*?)"')
// >>> re.findall(f,x)

let vectoroptionarray = Array('cosdistbysentence', 'cosdistbylineorword', 'semanticvectorquery',
    'nearestneighborsquery', 'tensorflowgraph', 'sentencesimilarity', 'topicmodel', 'analogyfinder');


// collections of elements that have logical connections

const corepickui = ['#worksautocomplete', '#makeanindex', '#textofthis', '#browseto', '#authinfobutton', '#makevocablist'];
const postauthorpickui = Array().concat(corepickui, coreactionbuttons);
const postbrowsepickui = Array().concat(corepickui, ['#browserdialog']);
const extrauichoices = Array().concat(categoryautofills);

// firstload hiding
const miscfirstloadhides = Array('#browserdialog', '#helptabs', '#fewerchoicesbutton', '#lemmatizing-ison',
    '#vectorizing-ison', '#alt_upperleftbuttons', '#analogiesinputarea', '#extendsearchbutton-ispresentlyopen',
    '#vectorsearchcheckbox', '#trimmingcheckboxes');
const tohideonfirstload = Array().concat(miscfirstloadhides, endpointnoticesandbuttons,
    endpointids, inputids, actionbuttons, infobuttons, infoboxes, lemmatasearchforms, extrasearchcriteria,
    lemmatabagoptions, extrasearchuielements);

