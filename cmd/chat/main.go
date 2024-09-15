package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/mattn/go-sqlite3"
	"github.com/skykosiner/golang-context/pkg/user"
	"github.com/skykosiner/golang-context/pkg/ws"
)

var templates = template.Must(template.ParseGlob("views/*.html"))

type Page struct {
	Title    string
	LoggedIn bool
	Username string
}

type HTTPHandler struct {
	ctx context.Context
	db  *sql.DB
}

func (p *Page) UpdateValues(loggedIn bool, username string) {
	p.LoggedIn = loggedIn
	p.Username = username
}

func NewHTTPHandler(db *sql.DB, ctx context.Context) *HTTPHandler {
	return &HTTPHandler{
		ctx,
		db,
	}
}

func (h *HTTPHandler) getUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(h.ctx, 2*time.Second)
	defer cancel()

	users, err := user.GetUsers(ctx, h.db)
	if err != nil {
		slog.Error("Error getting all the users", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("It's so over"))
		return
	}

	tmpl := `<ul>
	{{range .}}
		<li>{{.Username}}</li>
	{{end}}
</ul>`
	t, err := template.New("users").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, users)
}

func (h *HTTPHandler) addUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	ctx, cancel := context.WithTimeout(h.ctx, 2*time.Second)
	defer cancel()

	username := r.FormValue("username")
	if len(username) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please make sure you provide a username."))
		return
	}

	exists, err := user.UserExists(username, h.db, ctx)
	if err != nil {
		slog.Error("It's so over", "error", err)
		w.Write([]byte("JOEVER"))
		return
	}

	if exists {
		w.Write([]byte("Sorry a user with that username already exists."))
		return
	}

	query, args, err := sq.Insert("users").Columns("username").Values(username).ToSql()
	if err != nil {
		slog.Error("Error building sql query for new user", "error", err)
		w.Write([]byte("Error please try again."))
		return
	}

	if _, err = h.db.Exec(query, args...); err != nil {
		slog.Error("Error adding new user to the db.", "error", err)
		w.Write([]byte("Error please try again."))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "user",
		Value: username,
	})

	w.Write([]byte("WE'RE SO BACK"))
}

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

	ctx := context.Background()
	httpHandler := NewHTTPHandler(db, ctx)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/ws", ws.HandleWebsocket)
	http.HandleFunc("/get-messages", ws.GetMessages)
	http.HandleFunc("/get-users", httpHandler.getUsers)
	http.HandleFunc("/user", httpHandler.addUser)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var page Page
		page.Title = "Home Page"
		cookie, err := r.Cookie("user")
		if err == nil && len(cookie.Value) > 0 {
			if exists, _ := user.UserExists(cookie.Value, db, ctx); exists {
				page.UpdateValues(true, cookie.Value)
			} else {
				http.SetCookie(w, &http.Cookie{
					Name:    "user",
					Value:   "",
					Expires: time.Unix(0, 0),
					MaxAge:  -1,
				})
			}
		}

		if page.LoggedIn {
			err = templates.ExecuteTemplate(w, "loggedIn", page)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			err = templates.ExecuteTemplate(w, "login", page)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	})

	slog.Warn("Listening now on", "port", port)
	slog.Error(
		"Server is so joever",
		"error", http.ListenAndServe(fmt.Sprintf(":%s", port), nil),
	)
}
