package main

import (
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
)

// for now, can add auth later
const user = "global"

type server struct {
	store *Store
}

type Template struct {
	Path  string
	Query string
}

func (s *server) redirect(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	key, args, _ := strings.Cut(key, " ")

	outURL, err := s.store.Get(user, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hasQueryVar := strings.Contains(outURL, ".Query")
	hasPathVar := strings.Contains(outURL, ".Path")

	if !hasQueryVar && !hasPathVar {
		http.Redirect(w, r, outURL, http.StatusFound)
		return
	}

	tmpl, err := template.New("url").Parse(outURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	path := chi.URLParam(r, "*")
	if args != "" && path != "" {
		path = args + "/" + path
	}

	path, _, _ = strings.Cut(path, "?")
	query := r.URL.RawQuery

	if !hasPathVar {
		query = path
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, Template{
		Path:  path,
		Query: query,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, buf.String(), http.StatusTemporaryRedirect)
}

func (s *server) set(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	key := r.FormValue("key")
	url := r.FormValue("url")

	slog.Info("set", "key", key, "url", url)

	if err := s.store.Set(user, key, url); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *server) delete(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	if err := s.store.Delete(user, key); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer slog.Info("received",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
			"took", time.Since(start),
			"time", start,
		)

		next.ServeHTTP(w, r)
	})
}

func main() {
	dbPath := "store.db"
	if path := os.Getenv("DB_PATH"); path != "" {
		dbPath = path
	}

	s := &server{
		store: Open(dbPath),
	}
	defer s.store.Close()

	r := chi.NewMux()

	r.Use(logger)

	r.Get("/", s.index)
	r.Get("/static/*", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path, _ = url.JoinPath("/front", r.URL.Path)
		http.FileServer(http.FS(frontFS)).ServeHTTP(w, r)
	})

	r.Post("/", s.set)
	r.Get("/{key}", s.redirect)
	r.Get("/{key}/*", s.redirect)
	r.Delete("/", s.delete)

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	slog.Info("starting", "port", port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		slog.Error("ls", err)
	}
}
