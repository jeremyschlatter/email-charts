function redirectPOST(page, args) {
    var inputs = "";
    for (arg in args) {
        inputs += '<input type="text" name="' + arg + '" value="' + args[arg] + '" />';
    }
    var form = $('<form action="' + page + '" method="post">' + inputs + '</form>');
    $(form).submit();
}

function handleAuthResult(authResult) {
    if (authResult && !authResult.error) {
        $.ajax({url: 'https://www.googleapis.com/userinfo/email?alt=json&oauth_token=' + authResult.access_token,
                success: function(result) {
                    redirectPOST('loading', { user: result.data.email, token: authResult.access_token });
                }});
    } else {
        alert("Looks like you didn't accept the GMail login. Come back if you change your mind!");
    }
}

function authorize(immediate, callback) {
    gapi.auth.authorize({client_id: '582390535564-1iinodq9ttbrgb47e3pchsgdbf25hcis.apps.googleusercontent.com',
                         scope: 'https://mail.google.com/ https://www.googleapis.com/auth/userinfo.email',
                         immediate: immediate},
                        callback);
}

function doIt() {
    doAuth();
}
