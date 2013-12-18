function doAuth() {
    authorize(true, function(authResult) {
        if (authResult && !authResult.error) {
            handleAuthResult(authResult);
        } else {
            authorize(false, handleAuthResult);
        }
    });
}
