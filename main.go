package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// For handling dynamic routing
type ListHandler struct{}

func (p *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[6:]

	if r.Method == "GET" {
		fmt.Fprintf(w, "opening %s...", id)
		return
	}

	http.NotFound(w, r)
	return
}

type TodoHandler struct{}

type List struct {
	id          string
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Todo struct {
	id          string
	Description string `json:"description"`
	ListId      string `json:"listId"`
}

func (p *TodoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[6:]

	if r.Method == "GET" {
		fmt.Fprintf(w, "opening todo %s...", id)
		return
	}

	http.NotFound(w, r)
	return
}

func errorResponse(w http.ResponseWriter, message string, httpStatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	response := map[string]interface{}{
		"message": message,
		"status":  httpStatusCode,
		"error":   http.StatusText(httpStatusCode),
	}

	jsonResponse, err := json.Marshal(response)

	if err != nil {
		log.Fatal("Invalid json")
	}

	w.Write(jsonResponse)
}

func main() {
	newMux := http.NewServeMux()

	db, err := sql.Open("sqlite3", "todo.db")
	defer db.Close()

	if err != nil {
		log.Fatal(err)
	}

	if db != nil {
		log.Println("connected to database")
	}

	initTableQuery := `
CREATE TABLE IF NOT EXISTS lists (name TEXT, description TEXT, id TEXT PRIMARY KEY);
CREATE TABLE IF NOT EXISTS todos (id TEXT PRIMARY KEY, description TEXT, listId TEXT, FOREIGN KEY(listId) REFERENCES lists(id));
`

	_, initDbError := db.Exec(initTableQuery)

	if initDbError != nil {
		log.Fatal(initDbError.Error())
	}

	newMux.HandleFunc("/lists/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[6:]

		if path == "/" {
			if r.Method == "GET" {
				return
			}
		}

		if path == "/add" {
			if r.Method == http.MethodPut {
				contentType := r.Header.Get("Content-Type")

				if contentType != "application/json" {
					errorResponse(w, "Content Type is not JSON", http.StatusUnprocessableEntity)
					return
				}

				var newList List

				decoder := json.NewDecoder(r.Body)
				decoder.DisallowUnknownFields()
				err := decoder.Decode(&newList)

				if err != nil {
					errorResponse(w, "Bad Request", http.StatusBadRequest)
					return
				}

				createListQuery := fmt.Sprintf(`INSERT INTO lists (name, description, id) VALUES('%s', '%s', '%s')`, newList.Name, newList.Description, uuid.NewString())

				_, dbErr := db.Exec(createListQuery)

				if dbErr != nil {
					errorResponse(w, "Internal Server Error", http.StatusInternalServerError)
					log.Fatal(err.Error())
					return
				}

				return
			}
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
