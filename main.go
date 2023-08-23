package main

import (
	"fmt"
	"log"
	"net/http"
)

// For handling dynamic routing
type ListHandler struct {}
func (p *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  id := r.URL.Path[6:]

  if r.Method == "GET" {
    fmt.Fprintf(w, "opening %s...", id)
    return
  }

  http.NotFound(w, r)
	return
}

type TodoHandler struct {}
func (p *TodoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  id := r.URL.Path[6:]

  if r.Method == "GET" {
    fmt.Fprintf(w, "opening todo %s...", id)
    return
  }

  http.NotFound(w, r)
  return
}

func main() {
	newMux := http.NewServeMux()

	newMux.HandleFunc("/lists", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			return
		}
		http.NotFound(w, r)
		return
	})

	newMux.Handle("/list/", &ListHandler{})

	newMux.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			return
		}
		if r.Method == "PUT" {
			return
		}
		if r.Method == "PATCH" {
			return
		}
		if r.Method == "DELETE" {
			return
		}
		http.NotFound(w, r)
		return
	})

  newMux.Handle("/todo/", &TodoHandler{})

	log.Fatal(http.ListenAndServe(":4000", newMux))
}
