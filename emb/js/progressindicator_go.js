//
//	HipparchiaServer: an interface to a database of Greek and Latin texts
//	Copyright: E Gunderson 2016-22
//	License: License: GNU GENERAL PUBLIC LICENSE 3
//      (see LICENSE in the top level directory of the distribution)
//

//
// PROGRESS INDICATOR
//

function checkactivityviawebsocket(searchid) {
    $.getJSON('/search/confirm/'+searchid, function(portnumber) {
        let pd = $('#pollingdata');
        let ip = location.hostname;
        // but /etc/nginx/nginx.conf might have a WS proxy and not the actual WS host...
        let s = new WebSocket('ws://'+ip+':'+portnumber+'/ws');
        let amready = setInterval(function(){
            if (s.readyState === 1) { s.send(JSON.stringify(searchid)); clearInterval(amready); }
            }, 10);
        s.onmessage = function(e){
            let progress = JSON.parse(e.data);
            displayprogress(searchid, progress);
            if  (progress['active'] === 'inactive') { pd.html(''); s.close(); s = null; }
            }
    });
}

function displayprogress(searchid, progress){
    if (progress['ID'] === searchid) {
        $('#pollingdata').html(progress['value']);
    } else {
        console.log("id", progress['ID']);
        console.log("searchid", searchid);
    }
}


function simpleactivityviawebsocket(searchid) {
    $.getJSON('/search/confirm/'+searchid, function(portnumber) {
        let pd = $('#pollingdata');
        let ip = location.hostname;
        // but /etc/nginx/nginx.conf might have a WS proxy and not the actual WS host...
        let s = new WebSocket('ws://'+ip+':'+portnumber+'/ws');
        let amready = setInterval(function(){
            if (s.readyState === 1) { s.send(JSON.stringify(searchid)); clearInterval(amready); }
        }, 10);
        s.onmessage = function(e){
            let progress = JSON.parse(e.data);
            displayprogress(searchid, progress);
            if  (progress['active'] === 'inactive') { pd.html(''); s.close(); s = null; }
        }
    });
}
