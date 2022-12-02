package main

import (
	_ "embed"
	"html/template"
	"net/http"
)

//go:embed index.html
var idxTmpl []byte

var idxT = template.Must(template.New("index").Parse(string(idxTmpl)))

func (s *server) index(w http.ResponseWriter, r *http.Request) {
	links, _ := s.store.List(user)

	w.Header().Set("Content-Type", "text/html")
	if err := idxT.Execute(w, links); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
