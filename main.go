package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// For handling dynamic routing
type ListHandler struct{}

func (p *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pathname := r.URL.Path[6:]

	if r.Method == http.MethodGet {
		listId, err := strconv.Atoi(pathname)

		if err != nil {
			errorResponse(w, "Invalid id", http.StatusUnprocessableEntity)
			return
		}

		selectListQuery := fmt.Sprintf(`SELECT id, name, description FROM lists WHERE id = %d LIMIT 1;`, listId)
		selectTodoQuery := fmt.Sprintf(`SELECT id, description, status FROM todos WHERE listId = %d LIMIT 1`, listId)

		listRow := db.QueryRow(selectListQuery)
		todoRows, todoErr := db.Query(selectTodoQuery)

		if todoErr != nil {
			log.Fatal("connection error", todoErr.Error())
		}

		fetchedList := List{}
		fetchedTodos := []Todo{}

		listRow.Scan(&fetchedList.Id, &fetchedList.Name, &fetchedList.Description)

		for todoRows.Next() {
			todo := Todo{}
			todoRows.Scan(&todo.Id, &todo.Description, &todo.Status, &todo.ListId)
			fetchedTodos = append(fetchedTodos, todo)
		}

		fetchedListJson, _ := json.Marshal(fetchedList)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(fetchedListJson)
		return
	}

	if r.Method == http.MethodPut {
		pathSlice := strings.Split(pathname, "/")

		if pathSlice[1] == "createTodo" {
			contentType := r.Header.Get("Content-Type")

			if contentType != "application/json" {
				errorResponse(w, "Content type is not JSON", http.StatusUnprocessableEntity)
			}

			newTodo := Todo{}

			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&newTodo)

			newTodo.ListId = pathSlice[0]
			newTodo.Status = "PENDING"

			if err != nil {
				errorResponse(w, "Bad Request", http.StatusBadRequest)
				return
			}

			insertTodoQuery := fmt.Sprintf(`INSERT INTO todos (description, listId) VALUES('%s', '%s')`, newTodo.Description, newTodo.ListId)

			result, dbErr := db.Exec(insertTodoQuery)

			if dbErr != nil {
				errorResponse(w, "Unable to create row "+dbErr.Error(), http.StatusInternalServerError)
				return
			}

			newTodoId, _ := result.LastInsertId()
			newTodo.Id = int8(newTodoId)

			newTodoJSON, _ := json.Marshal(newTodo)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write(newTodoJSON)

			return
		}
	}

	http.NotFound(w, r)
	return
}

type TodoHandler struct{}

type List struct {
	Id          int8   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Todo struct {
	Id          int8   `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"`
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
CREATE TABLE IF NOT EXISTS todos (id INTEGER PRIMARY KEY AUTOINCREMENT, description TEXT, status TEXT DEFAULT 'PENDING',listId TEXT, FOREIGN KEY(listId) REFERENCES lists(id));
`

	_, initDbError := db.Exec(initTableQuery)

	if initDbError != nil {
		log.Fatal(initDbError.Error())
	}

	newMux.HandleFunc("/lists/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[6:]

		if path == "/" {
			if r.Method == http.MethodGet {
				rows, err := db.Query(`SELECT id, name, description FROM lists`)

				if err != nil {
					errorResponse(w, "Unable to get rows", http.StatusInternalServerError)
				}

				fetchedLists := []List{}

				for rows.Next() {
					list := List{}
					scanErr := rows.Scan(&list.Id, &list.Name, &list.Description)
					if scanErr != nil {
						errorResponse(w, scanErr.Error(), http.StatusInternalServerError)
						return
					}

					fetchedLists = append(fetchedLists, list)
				}

				fetchedListsJson, _ := json.Marshal(fetchedLists)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(200)
				w.Write(fetchedListsJson)

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

	log.Fatal(http.ListenAndServe(":4000", newMux))
}
