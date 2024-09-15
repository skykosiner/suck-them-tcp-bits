// Reload page after user enters in username
document.addEventListener("htmx:afterRequest", function(evt) {
    if (evt.target.id === "usernameForm") {
        window.location.reload();
    }
});

