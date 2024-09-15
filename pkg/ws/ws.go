package ws

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"text/template"

	"github.com/gorilla/websocket"
	"github.com/skykosiner/golang-context/pkg/user"
)

type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

var (
	clients   = make(map[*websocket.Conn]string)
	clientsMu sync.Mutex // Ensure thread-safe access to the clients map
	upgrader  = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	messages []Message
)

func HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	db, ok := r.Context().Value("db").(*sql.DB)
	if !ok {
		slog.Error("It's so over there' no db")
		http.Error(w, "Database in not available it's over.", http.StatusInternalServerError)
		return
	}

	username := r.URL.Query().Get("name")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Websocket connection upgrade failed", "err", err)
	}

	clientsMu.Lock()
	clients[ws] = username
	fmt.Println(clients)
	clientsMu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		clientsMu.Lock()
		username := clients[ws]

		err := user.DeleteUserFromDb(username, db)
		if err != nil {
			slog.Error("it's so over", "error", err)
			http.Error(w, "It's Joever", http.StatusInternalServerError)
			return
		}

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
		messages = append(messages, msg)

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

func GetMessages(w http.ResponseWriter, r *http.Request) {
	tmpl := `<div>
	{{range .}}
		<p>{{.Username}}:  {{.Message}}</p>
	{{end}}
</div>`
	t, err := template.New("messages").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, messages)
}
