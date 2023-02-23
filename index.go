package main

import (
	"embed"
	"html/template"
	"net/http"
	"os"
	"sync"
)

//go:embed front/*
var frontFS embed.FS

var idxT = template.Must(template.New("index.html").ParseFS(frontFS, "front/index.html"))
var idxTMu = sync.Mutex{}

func (s *server) index(w http.ResponseWriter, r *http.Request) {
	links, _ := s.store.List(user)

	if os.Getenv("DEBUG") != "" {
		idxTMu.Lock()
		idxT = template.Must(template.ParseFiles("front/index.html"))
		defer idxTMu.Unlock()
	}

	w.Header().Set("Content-Type", "text/html")
	if err := idxT.Execute(w, links); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
