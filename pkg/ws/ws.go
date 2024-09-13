package ws

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex // Ensure thread-safe access to the clients map
	upgrader  = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Websocket connection upgrade failed", "err", err)
	}

	clientsMu.Lock()
	clients[ws] = true
	clientsMu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		clientsMu.Lock()
		delete(clients, ws)
		clientsMu.Unlock()
		ws.Close()
	}()

	go func() {
		<-ctx.Done()
		slog.Warn("It's so over one of the ws conections has been closed")
	}()

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			slog.Error("Failed to read websocket message", "err", err)
			cancel()
			break
		}

		slog.Warn("Got a message from client", "msg", msg)

		clientsMu.Lock()
		for client := range clients {
			err := client.WriteJSON(&msg)
			if err != nil {
				slog.Error("It's so over couldn't send a message to a client", "err", err)
				client.Close()
				delete(clients, client)
			}
		}
		clientsMu.Unlock()

	}
}
