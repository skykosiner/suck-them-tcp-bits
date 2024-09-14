package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/skykosiner/golang-context/pkg/ws"
)

type Page struct {
	// Port string
}

var templates = template.Must(template.ParseGlob("views/*.html"))

func main() {
	var port string
	flag.StringVar(&port, "port", "42069", "The port of the chat server")
	flag.Parse()

	db, err := sql.Open("sqlite3", "./dvorak-btw.sqlite")
	if err != nil {
		slog.Error("Couldn't seem to open database, my bad tbh", "error", err)
		return
	}

	defer db.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := templates.ExecuteTemplate(w, "index", Page{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	mux.HandleFunc("/ws", ws.HandleWebsocket)
	mux.HandleFunc("/get-messages", ws.GetMessages)

	slog.Warn("Listening now on", "port", port)
	slog.Error(
		"Server is so joever",
		"error", http.ListenAndServe(fmt.Sprintf(":%s", port), mux),
	)
}
