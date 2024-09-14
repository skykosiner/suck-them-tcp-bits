const usernameForm = document.getElementById("setUsername");
const messagesDiv = document.getElementById("messages");
const newMessageForm = document.getElementById("sendMessage");
const host = window.location.host;
const ws = new WebSocket(`ws://${host}/ws`)

/**
    * Updates the #messages div with all the old messages currently stored on the server
*/
async function getOldMessages() {
    /** @type {import("./types.ts").Message[]}*/
    const messages = await fetch("/get-messages").then(resp => resp.json());

    messages.map(msg => {
        messagesDiv.innerHTML += `<p>${msg.name}: ${msg.message}</p>`;
    })
}

getOldMessages();

/**
    * @param {import("./types.ts").WsMessage} event
*/
ws.onmessage = function(event) {
    /** @type {import("./types.ts").Message}*/
    const msg = JSON.parse(event.data);

    const messages = document.getElementById("messages");
    messages.innerHTML += `<p>${msg.name}: ${msg.message}</p>`;  // Update DOM if the element exists
}

ws.onerror = function(event) {
    console.log("It's so over. Websocket error", event);
}

usernameForm.addEventListener("submit", (e) => {
    e.preventDefault();
    /** @type {string} */
    const username = document.getElementById("username").value.trim();

    if (username !== "") {
        sessionStorage.setItem("username", username);
    }

    if (username) {
        // Proceed with username setting
        usernameForm.classList.add('hide');
        newMessageForm.classList.remove('hide');
        messagesDiv.classList.remove('hide');
    }
});


newMessageForm.addEventListener("submit", (e) => {
    e.preventDefault();

    const messageInput = document.getElementById("newMessage");
    const message = messageInput.value;  // Get the value from input

    if (message.trim() !== "") {
        /** @type {import("./types.ts").Message}*/
        const msg = {
            name: sessionStorage.getItem("username"),
            message,
        }

        ws.send(JSON.stringify(msg));  // Send the message to WebSocket
        messageInput.value = "";  // Clear the input field after sending
    } else {
        console.warn("Message is empty. Not sending.");
    }
});
