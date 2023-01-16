package main

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed front/*
var frontFS embed.FS

var idxT = template.Must(template.New("index.html").ParseFS(frontFS, "front/index.html"))

func (s *server) index(w http.ResponseWriter, r *http.Request) {
	links, _ := s.store.List(user)

	w.Header().Set("Content-Type", "text/html")
	if err := idxT.Execute(w, links); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
