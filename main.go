package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

var (
	addr = flag.String("http", "127.0.0.1:8080", "listen on http")
)

func main() {
	flag.Parse()
	var rd io.Reader = os.Stdin
	if flag.Arg(0) != "" {
		file, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		rd = file
	}

	data, err := io.ReadAll(rd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	index := NewIndex()

	dir, _ := filepath.Abs(".")
	index.Parse(dir, data)

	fmt.Printf("Listening on http://%v\n", *addr)
	err = http.ListenAndServe(*addr, &Server{index})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

type Server struct {
	Index *Index
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "" || r.URL.Path == "/" {
		defaultPath := r.URL.Query().Get("path")

		err := T.Execute(w, map[string]interface{}{
			"StatCount":   statCount,
			"Stats":       statSpecs,
			"Files":       server.Index.Files,
			"DefaultPath": defaultPath,
		})
		if err != nil {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		return
	}

	if r.URL.Path == "/file" {
		path := r.FormValue("path")
		if path == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "No path specified.")
			return
		}

		annotated, err := server.Index.LoadAnnotatedFile(path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(os.Stderr, "%v\n", err)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err = json.NewEncoder(w).Encode(annotated)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

//go:embed index.html
var indexTemplate string

var T = template.Must(template.New("").Funcs(template.FuncMap{
	"mul": func(a, b int) int { return a * b },
}).Parse(indexTemplate))
