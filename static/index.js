const host = window.location.host;
const ws = new WebSocket(`ws://${host}/ws`)

ws.onopen = function() {
    console.log("We're so in");
}

/**
    @param {import("./types.ts").WsMessage} event
*/
ws.onmessage = function(event) {
    const msg = document.getElementById("messages");
    if (msg) {
        msg.innerHTML += `<p>${event.data}</p>`;  // Update DOM if the element exists
    } else {
        console.error("Element #messages not found");
    }
}

ws.onerror = function(event) {
    console.log(event);
}

const newMessageForm = document.getElementById("sendMessage");

newMessageForm.addEventListener("submit", (e) => {
    e.preventDefault();

    const messageInput = document.getElementById("newMessage");
    const message = messageInput.value;  // Get the value from input

    if (message.trim() !== "") {
        ws.send(message);  // Send the message to WebSocket
        messageInput.value = "";  // Clear the input field after sending
    } else {
        console.warn("Message is empty. Not sending.");
    }
});
