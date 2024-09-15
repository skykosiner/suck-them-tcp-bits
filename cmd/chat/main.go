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

var (
	templates = template.Must(template.ParseGlob("views/*.html"))
)

type Page struct {
	Title    string
	LoggedIn bool
	Username string
}

type HTTPHandler struct {
	ctx context.Context
	db  *sql.DB
}

func NewHTTPHandler(db *sql.DB, ctx context.Context) *HTTPHandler {
	return &HTTPHandler{ctx: ctx, db: db}
}

func (h *HTTPHandler) renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	if err := templates.ExecuteTemplate(w, tmpl, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *HTTPHandler) getUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(h.ctx, 2*time.Second)
	defer cancel()

	users, err := user.GetUsers(ctx, h.db)
	if err != nil {
		slog.Error("Error getting users", "error", err)
		http.Error(w, "Error getting users", http.StatusInternalServerError)
		return
	}

	tmpl := `<ul>{{range .}}<li>{{.Username}}</li>{{end}}</ul>`
	if err := template.Must(template.New("users").Parse(tmpl)).Execute(w, users); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *HTTPHandler) addUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	ctx, cancel := context.WithTimeout(h.ctx, 2*time.Second)
	defer cancel()

	username := r.FormValue("username")
	if username == "" {
		http.Error(w, "Please provide a username.", http.StatusBadRequest)
		return
	}

	if exists, err := user.UserExists(username, h.db, ctx); err != nil || exists {
		msg := "Error occurred"

		if exists {
			msg = "User already exists"
		}

		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if _, err := sq.Insert("users").Columns("username").Values(username).RunWith(h.db).ExecContext(ctx); err != nil {
		slog.Error("Error adding user", "error", err)
		http.Error(w, "Error adding user", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "user", Value: username})
	w.Write([]byte("User added"))
}

func (h *HTTPHandler) home(w http.ResponseWriter, r *http.Request) {
	page := Page{Title: "Home Page"}
	if cookie, err := r.Cookie("user"); err == nil && cookie.Value != "" {
		if exists, _ := user.UserExists(cookie.Value, h.db, h.ctx); exists {
			page.LoggedIn, page.Username = true, cookie.Value
		} else {
			http.SetCookie(w, &http.Cookie{Name: "user", Value: "", Expires: time.Unix(0, 0), MaxAge: -1})
		}
	}

	tmpl := "login"
	if page.LoggedIn {
		tmpl = "loggedIn"
	}

	h.renderTemplate(w, tmpl, page)
}

func withDB(next http.Handler, db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "db", db)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	port := flag.String("port", "42069", "The port to listen on")
	flag.Parse()

	db, err := sql.Open("sqlite3", "./dvorak-btw.sqlite")
	if err != nil {
		slog.Error("Error opening database", "error", err)
		return
	}
	defer db.Close()

	httpHandler := NewHTTPHandler(db, context.Background())

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.Handle("/ws", withDB(http.HandlerFunc(ws.HandleWebsocket), db))
	http.HandleFunc("/get-messages", ws.GetMessages)
	http.HandleFunc("/get-users", httpHandler.getUsers)
	http.HandleFunc("/user", httpHandler.addUser)
	http.HandleFunc("/", httpHandler.home)

	slog.Warn("Listening on", "port", *port)
	slog.Error("Server error", "error", http.ListenAndServe(fmt.Sprintf(":%s", *port), nil))
}
