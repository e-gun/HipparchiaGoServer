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
    // note that morphologychartjs() has its own version of this function: changes here should be made there too
    // console.log(progress);
    let r = progress['Remaining'];
    let t = progress['Poolofwork'];
    let h = progress['Hitcount'];
    let pct = Math.round((t-r) / t * 100);
    // let pct = 100;
    let m = progress['Statusmessage'];
    let e = progress['Elapsed'];
    let x = progress['Notes'];
    let id = progress['ID'];
    let a = progress['Activity'];

    // let thehtml = '[' + id + '] ';

    if (id === searchid) {
        let thehtml = '';
        if (r !== undefined && t !== undefined && !isNaN(pct)) {
            // let e = Math.round((new Date().getTime() / 1000) - l);

            if (t !== -1 && pct < 100) {
                thehtml += m + ': <span class="progress">' + pct + '%</span> completed&nbsp;(' + e + ')';
            } else {
                thehtml += m + '&nbsp;(' + e + ')';
            }

            if (h > 0) {
                let hc = h.toLocaleString();
                thehtml += '<br />(<span class="progress">' + hc + '</span> found)';
            }

            thehtml += '<br />' + x;
        }
        $('#pollingdata').html(thehtml);
    } else {
        console.log("id", id);
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
            simpleprogress(searchid, progress);
            if  (progress['active'] === 'inactive') { pd.html(''); s.close(); s = null; }
        }
    });
}

function simpleprogress(searchid, progress) {
    let m = progress['Statusmessage'];
    let e = progress['Elapsed'];
    let x = progress['Notes'];
    let id = progress['ID'];
    let a = progress['Activity'];

    // console.log("id", id);
    // console.log("searchid", searchid);

    if (id === searchid) {
        let thehtml = '';
        thehtml += m + '&nbsp;(' + e + ')';
        $('#pollingdata').html(thehtml);
    }
}