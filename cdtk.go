package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"html/template"
	"log"
	"net/http"
	"regexp"
)

const (
	DB_USER     = "postgres"
	DB_PASSWORD = "123456"
	DB_NAME     = "postgres"
)

var dbinfo = fmt.Sprintf("user=%s password = %s dbname=%s sslmode=disable", DB_USER, DB_PASSWORD, DB_NAME)

type TodoItem struct {
	Title string
	Body  string
}

type TodoList struct {
	TodoItems []TodoItem
}

func (todoItem *TodoItem) save() error {
	query := fmt.Sprintf("INSERT INTO todo(title, body) VALUES($1, $2)")
	db, err := sql.Open("postgres", dbinfo)

	if err != nil {
		return err
	}

	defer db.Close()

	db.QueryRow(query, todoItem.Title, todoItem.Body)

	return nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index")
}

var templates = template.Must(template.ParseFiles("html/index.html",
	"html/todo_detail.html",
	"html/todo.html",
	"html/todo_edit.html"))

func renderTemplate(w http.ResponseWriter, tmpl string) {
	err := templates.ExecuteTemplate(w, tmpl+".html", nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderTodoItemTemplate(w http.ResponseWriter, tmpl string, todoItem *TodoItem) {
	err := templates.ExecuteTemplate(w, tmpl+".html", todoItem)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderTodoListTemplate(w http.ResponseWriter, tmpl string, todoList *TodoList) {
	err := templates.ExecuteTemplate(w, tmpl+".html", todoList)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func fetchTodoItem(title string) (*TodoItem, error) {
	db, err := sql.Open("postgres", dbinfo)
	query := fmt.Sprintf("SELECT body FROM todo WHERE title = '%s'", title)
	rows, err := db.Query(query)
	defer db.Close()

	if err != nil {
		fmt.Println(err)
	}

	var body string
	for rows.Next() {
		err = rows.Scan(&body)

		if err != nil {
			fmt.Println(err)
		}
	}

	return &TodoItem{Title: title, Body: body}, nil
}

func fetchTodoList() (*TodoList, error) {
	dbinfo := fmt.Sprintf("user=%s password = %s dbname=%s sslmode=disable", DB_USER, DB_PASSWORD, DB_NAME)

	db, err := sql.Open("postgres", dbinfo)
	query := fmt.Sprintf("SELECT title, body FROM todo")
	rows, err := db.Query(query)
	defer db.Close()

	if err != nil {
		fmt.Println(err)
	}

	var todoList TodoList
	var body string
	var title string

	for rows.Next() {
		err = rows.Scan(&title, &body)

		if err != nil {
			fmt.Println(err)
		}

		item := &TodoItem{Title: title, Body: body}
		todoList.TodoItems = append(todoList.TodoItems, *item)
	}

	return &todoList, nil
}

func todoListHandler(w http.ResponseWriter, r *http.Request) {
	todoList, err := fetchTodoList()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	renderTodoListTemplate(w, "todo", todoList)
}

var todoValidPath = regexp.MustCompile("^/todo/(view|save|edit)/([a-zA-Z0-9]+)$")

func makeTodoHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := todoValidPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}

		fn(w, r, m[2])
	}
}

func viewTodoHandler(w http.ResponseWriter, r *http.Request, title string) {
	todoItem, err := fetchTodoItem(title)

	if err != nil {
		http.NotFound(w, r)
	}

	renderTodoItemTemplate(w, "todo_detail", todoItem)
}

func saveTodoHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	item := &TodoItem{Title: title, Body: body}
	err := item.save()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/todo/view/"+title, http.StatusFound)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	item, err := fetchTodoItem(title)

	if err != nil {
		item = &TodoItem{Title: title}
	}

	renderTodoItemTemplate(w, "todo_edit", item)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/index", indexHandler)
	http.HandleFunc("/todo", todoListHandler)
	http.HandleFunc("/todo/view/", makeTodoHandler(viewTodoHandler))
	http.HandleFunc("/todo/save/", makeTodoHandler(saveTodoHandler))
	http.HandleFunc("/todo/edit/", makeTodoHandler(editHandler))

	log.Fatal(http.ListenAndServe(":1025", nil))
}
