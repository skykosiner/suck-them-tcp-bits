const port = window.location.port;
const ws = new WebSocket(`ws://localhost:${port}/ws`)

ws.onopen = function() {
    ws.send("Deez Nuts");
}

/**
    @param {import("./types.ts").WsMessage} event
*/
ws.onmessage = function(event) {
    const msg = document.getElementById("messages");
    if (msg) {
        msg.innerHTML = `<p>${event.data}</p>`;  // Update DOM if the element exists
    } else {
        console.error("Element #messages not found");
    }
}

ws.onerror = function(event) {
    console.log(event);
}
