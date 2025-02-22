const messagesDiv = document.getElementById("messages");
const usersDiv = document.getElementById("users");
const newMessageForm = document.getElementById("sendMessage");
const host = window.location.host;
const username = document.cookie.split("=")[1]
const ws = new WebSocket(`ws://${host}/ws?name=${username}`)

/**
    * @param {import("./types.ts").WsMessage} event
*/
ws.onmessage = function(event) {
    /** @type {import("./types.ts").Message}*/
    let msg;
    if (!event.data.includes("leaved")) {
        msg = JSON.parse(event.data);
    } else {
        htmx.trigger("#users", "refresh");
        return
    }

    const messages = document.getElementById("messages");
    messages.innerHTML += `<p>${msg.username}: ${msg.message}</p>`;  // Update DOM if the element exists
}

ws.onerror = function(event) {
    console.log("It's so over. Websocket error", event);
}

newMessageForm.addEventListener("submit", (e) => {
    e.preventDefault();

    const messageInput = document.getElementById("newMessage");
    const message = messageInput.value;  // Get the value from input

    if (message.trim() !== "") {
        /** @type {import("./types.ts").Message}*/
        const msg = {
            username: username,
            message,
        }

        ws.send(JSON.stringify(msg));  // Send the message to WebSocket
        messageInput.value = "";  // Clear the input field after sending
    } else {
        console.warn("Message is empty. Not sending.");
    }
});
