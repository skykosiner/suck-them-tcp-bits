package main

import (
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/skykosiner/golang-context/pkg/ws"
)

type Page struct {
	// Port string
}

var templates = template.Must(template.ParseFiles("./view/index.gohtmltmpl"))

func main() {
	var port string
	flag.StringVar(&port, "port", "42069", "The port of the chat server")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := templates.ExecuteTemplate(w, "index.gohtmltmpl", Page{})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/ws", ws.HandleWebsocket)

	slog.Warn("Listening now on", "port", port)
	slog.Error("Server is so joever", http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
