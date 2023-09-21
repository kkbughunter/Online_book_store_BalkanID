package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

var db *sql.DB

type User struct {
	ID   int
	Name string
}

func main() {
	// Create a PostgreSQL database connection.
	connStr := "user=myuser dbname=mydb password=mypassword sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Initialize the HTTP server.
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/add", addHandler)
	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	name := r.FormValue("name")
	idStr := r.FormValue("password")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID value", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO account VALUES ($1, $2)", id, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
