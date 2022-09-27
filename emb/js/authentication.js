//    HipparchiaGoServer
//    Copyright: E Gunderson 2022
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

$('#validateusers').dialog({ autoOpen: false });

$('#executelogin').click( function() {
    $('#validateusers').dialog( 'open' );
});

$('#executelogout').click( function() {
    $.getJSON('/authentication/logout', function(data){
         $('#userid').html(data);
    });
    $('#executelogout').hide();
    $('#executelogin').show();
});
