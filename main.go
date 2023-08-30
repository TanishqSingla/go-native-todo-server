package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

// For handling dynamic routing
type ListHandler struct{}

func (p *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[6:]

	if r.Method == http.MethodGet {
		listId, err := strconv.Atoi(id)

		if err != nil {
			errorResponse(w, "Invalid id", http.StatusUnprocessableEntity)
			return
		}

		selectListQuery := fmt.Sprintf(`SELECT id, name, description FROM lists WHERE id = %d LIMIT 1;`, listId)

		row := db.QueryRow(selectListQuery)

		fetchedList := List{}

		row.Scan(&fetchedList.Id, &fetchedList.Name, &fetchedList.Description)

		fetchedListJson, _:= json.Marshal(fetchedList)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(fetchedListJson)
		return
	}

	http.NotFound(w, r)
	return
}

type TodoHandler struct{}

type List struct {
	Id          int8 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Todo struct {
	Id          int8 `json:"id"`
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

var db *sql.DB
var dbErr error

func main() {
	newMux := http.NewServeMux()

	db, dbErr = sql.Open("sqlite3", "todo.db")
	defer db.Close()

	if dbErr != nil {
		log.Fatal(dbErr)
	}

	if db != nil {
		log.Println("connected to database")
	}

	initTableQuery := `
CREATE TABLE IF NOT EXISTS lists (name TEXT, description TEXT, id INTEGER PRIMARY KEY AUTOINCREMENT);
CREATE TABLE IF NOT EXISTS todos (id INTEGER PRIMARY KEY AUTOINCREMENT, description TEXT, listId TEXT, FOREIGN KEY(listId) REFERENCES lists(id));
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

				createListQuery := fmt.Sprintf(`INSERT INTO lists (name, description) VALUES('%s', '%s')`, newList.Name, newList.Description)

				result, dbErr := db.Exec(createListQuery)

				if dbErr != nil {
					errorResponse(w, "Unable to create row", http.StatusInternalServerError)
					log.Fatal(err.Error())
					return
				}
				resultId, _ := result.LastInsertId()
				newList.Id = int8(resultId)

				newListJson, _ := json.Marshal(newList)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				w.Write(newListJson)

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
