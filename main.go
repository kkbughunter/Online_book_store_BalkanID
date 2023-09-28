package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

type User struct {
	password int
	Name     string
	Email    string
}
type Book struct {
	ISBN        int
	title       string
	author      string
	genre       string
	publisher   string
	releasedata string
	format      string
	price       int
	stock       int
}

func main() {
	// Create a PostgreSQL database connection.
	connStr := "user=myuser dbname=mydb password=mypassword sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close() // this close will run at the end of the main get close.
	fmt.Println("Server Start Listening  http://localhost:8080")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static")))) // this line is use to display the static images, fonts etc...

	// Initialize the HTTP server.
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/index.html", loginPageHandler)
	http.HandleFunc("/register.html", registerPageHandler)

	http.HandleFunc("/reg", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)

	http.HandleFunc("/dashboard.html", dashboardHandler)
	http.HandleFunc("/Admindashboard.html", adminDashboardHandler)

	http.HandleFunc("/delete-book", deleteBookHandler)

	http.HandleFunc("/add-book", addBookHandler)
	http.HandleFunc("/remove-book", removeBookHandler)

	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		fmt.Println("Error in indesHandler")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	// Serve the login page here.
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		fmt.Println("Error in loginPageHandler")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
func registerPageHandler(w http.ResponseWriter, r *http.Request) {
	// Serve the login page here.
	tmpl, err := template.ParseFiles("register.html")
	if err != nil {
		fmt.Println("Error in loginPageHandler")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func adminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("user")
	var cookieValue string
	cookie, cerr := r.Cookie(name)
	if cerr != nil {
		fmt.Println("Error in Cookie")
	} else {
		cookieValue = cookie.Value
	}

	fmt.Println("\n", name, " Entered in dashboard")

	var ServerKey string
	err := db.QueryRow("SELECT key_value FROM section_keys WHERE userid = $1", name).Scan(&ServerKey)
	if err != nil {
		fmt.Println("app: No Key Exit - Query Error fetching the key_value ")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("app: user database key", ServerKey)

	if cookieValue != ServerKey {
		fmt.Println("app: Unauthorized Access User section already present...")
		http.Error(w, "Unauthorized Access", http.StatusUnauthorized)
		return
	}

	// If the section key is valid, query user information and serve the dashboard
	var user User
	err = db.QueryRow("SELECT adminid, emailid FROM sadmin WHERE adminid = $1", name).Scan(&user.Name, &user.Email)
	if err != nil {
		fmt.Println("app: User info not present. admin")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a data structure to hold the user information
	data := struct {
		User User
	}{
		User: user,
	}

	// Serve the regular user dashboard page with user data
	tmpl, err := template.ParseFiles("Admindashboard.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	name := r.FormValue("name")
	email := r.FormValue("email")
	password1 := r.FormValue("password")
	password2 := r.FormValue("confirm")

	if name[0] != 'A' {
		name = "U" + name
	}
	var errors []string

	// Check if a user with the same userid already exists.
	var existingUser bool
	row := db.QueryRow("SELECT EXISTS (SELECT 1 FROM suser WHERE userid = $1)", name)
	if err := row.Scan(&existingUser); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existingUser {
		errors = append(errors, "User already exists...")
	}

	if password1 != password2 {
		errors = append(errors, "Password not matching...")
	}
	if len(email) < 4 || email[len(email)-4:] != ".com" {
		errors = append(errors, "Incorrect Email format...")
	}

	// Structure for the error password, email, and user already exists.
	data := struct {
		Errors []string
	}{
		Errors: errors,
	}

	// Pass the error structure to the HTML page and display the errors on the page.
	if len(errors) > 0 {
		tmpl, err := template.ParseFiles("register.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, data)
		return
	}

	// If there are no errors, perform the database insert.
	if name[0] == 'A' {

		_, err := db.Exec("INSERT INTO sadmin (adminid, emailid, password) VALUES ($1, $2, $3)", name, email, password1)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		_, err := db.Exec("INSERT INTO suser (userid, emailid, password) VALUES ($1, $2, $3)", name, email, password1)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	fmt.Printf("REGISTERED: New User %s", name)
	http.Redirect(w, r, "/index.html", http.StatusSeeOther)
}
func ErrorHandling(errors []string, path string, w http.ResponseWriter) {
	data := struct {
		Errors []string
	}{
		Errors: errors,
	}
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println("app EH0x00: ParseFiles error:", err)
		return
	}
	tmpl.Execute(w, data)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Login Algorithm
	// ---------------
	// 0: start
	// 1: Check the given information is Admin or Normal User.
	// 	1.1: If user => check given information present in the user database (check the password also)
	// 		1.1.1: if Yes => check for the section is already present.
	// 			1.1.1.1: if Yes => Delete the old section and generate new section id.
	// 			1.1.1.2: And store it in the browser as a cookie.
	// 			1.1.1.3: Redirect to User Dashboard.
	// 		1.1.2: if No => Display the error Message
	// 	1.2: else admin => check given information present in the admin database(check the password also)
	// 		1.2.1: if Yes => check for the section is already present.
	// 			1.2.1.1: if Yes => Delete the old section and generate new section id.
	// 			1.2.1.2: And store it in the browser as a cookie.
	// 			1.2.1.3: Redirect to Admin Dashboard.
	// 		1.2.2: if No => Display the error Message
	// 2: if No => Display the error Message

	r.ParseForm()
	name := r.FormValue("name")
	passwd := r.FormValue("password")

	var errors []string

	if name[0] != 'A' {
		// Normal User
		var Checkname string
		var Checkpassword string
		name = "U" + name
		err1 := db.QueryRow("SELECT userid, password FROM suser WHERE userid = $1", name).Scan(&Checkname, &Checkpassword)
		if err1 != nil {
			errors = append(errors, "Register Your Self")
			fmt.Println("app L0x01: User Not Found in the table:", err1)
			ErrorHandling(errors, "index.html", w)
			return
		}
		if passwd != Checkpassword {
			errors = append(errors, "Incorrect Password")
			fmt.Println("app L0x01: Incorrect password try:", err1)
			ErrorHandling(errors, "index.html", w)
			return
		}
	} else {
		var Checkname string
		var Checkpassword string
		err1 := db.QueryRow("SELECT adminid, password FROM sadmin WHERE adminid = $1", name).Scan(&Checkname, &Checkpassword)
		if err1 != nil {
			errors = append(errors, "Register Your Self")
			fmt.Println("app L0x02: User Not Found in the table:", err1)
			ErrorHandling(errors, "index.html", w)
			return
		}

		if passwd != Checkpassword {
			errors = append(errors, "Incorrect Password")
			fmt.Println("app L1x02: Incorrect password try:", err1)
			ErrorHandling(errors, "index.html", w)
			return
		}
	}

	var Oldkey string
	err1 := db.QueryRow("SELECT key_value FROM section_keys WHERE userid = $1", name).Scan(&Oldkey)
	if err1 == nil {
		// there is old cookie so delete the old and create a new one
		db.QueryRow("DELETE FROM section_keys WHERE userid = $1", name) // deleting the old section id.
	}

	section_key, err := generateUniqueString(16)
	if err != nil {
		fmt.Println("app L2x01: generateUniqueString error", err)
		// return
	}
	cookie := &http.Cookie{
		Name:  name,
		Value: section_key,
	}
	http.SetCookie(w, cookie)
	_, err = db.Exec("INSERT INTO section_keys (userid, key_value) VALUES ($1, $2)", name, section_key)

	// login success
	if name[0] == 'U' {
		http.Redirect(w, r, "/dashboard.html?user="+name, http.StatusSeeOther)
		fmt.Println("app: login success for", name)
	} else {
		http.Redirect(w, r, "/Admindashboard.html?user="+name, http.StatusSeeOther)
		fmt.Println("app: login success for", name)
	}
}

func verifyCookie(r *http.Request) bool {
	name := r.URL.Query().Get("user")
	var cookieValue string
	cookie, cerr := r.Cookie(name)
	if cerr != nil {
		fmt.Println("Error in Cookie")
		return false
	}
	cookieValue = cookie.Value

	var ServerKey string
	err := db.QueryRow("SELECT key_value FROM section_keys WHERE userid = $1", name).Scan(&ServerKey)
	if err != nil {
		fmt.Println("app: No Key Exit - Query Error fetching the key_value ")
		return false
	}

	return cookieValue == ServerKey
}
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	type Book struct {
		ISBN        int
		Title       string
		Author      string
		Genre       string
		Publisher   string
		ReleaseDate time.Time // Use time.Time for date fields
		Format      string
		Price       float64
		Stock       int
	}
	name := r.URL.Query().Get("user")
	if !verifyCookie(r) {
		fmt.Println("app: Unauthorized Access User section already present...")
		http.Error(w, "Unauthorized Access", http.StatusUnauthorized)
		return
	}

	// Query user information
	var user User
	err := db.QueryRow("SELECT userid, emailid FROM suser WHERE userid = $1", name).Scan(&user.Name, &user.Email)
	if err != nil {
		fmt.Println("app: User info not present. user")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Query book details
	rows, err := db.Query("SELECT ISBN, title, author, genre, publisher, releasedata, format, price, stock FROM books")
	if err != nil {
		fmt.Println("app: Error fetching book details")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Create a slice to hold book details
	var books []Book
	for rows.Next() {
		var book Book
		err := rows.Scan(&book.ISBN, &book.Title, &book.Author, &book.Genre, &book.Publisher, &book.ReleaseDate, &book.Format, &book.Price, &book.Stock)
		if err != nil {
			fmt.Println("app: Error scanning book details")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		books = append(books, book)
	}

	// Create a data structure to hold user and book information
	data := struct {
		User  User
		Books []Book
	}{
		User:  user,
		Books: books,
	}

	// Serve the user dashboard page with user and book data
	tmpl, err := template.ParseFiles("dashboard.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

func deleteBookHandler(w http.ResponseWriter, r *http.Request) {
	isbn := r.URL.Query().Get("isbn")
	name := r.URL.Query().Get("user")
	if err := deleteBookFromDatabase(isbn); err != nil {
		fmt.Println("Error deleting book:", err)
		http.Error(w, "Failed to delete the book", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/dashboard.html?user="+name, http.StatusSeeOther)
}
func deleteBookFromDatabase(isbn string) error {
	db.QueryRow("UPDATE books SET stock = stock - 1 WHERE ISBN = $1", isbn)
	return nil
}
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the user's name from the request
	name := r.URL.Query().Get("user")
	println(name)
	// Delete the user's session data from the database
	_, err := db.Exec("DELETE FROM section_keys WHERE userid = $1", name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("app: logout by", name)
	// Redirect to the login page after logout
	http.Redirect(w, r, "/login.html", http.StatusSeeOther)
}

func generateUniqueString(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}
func addBookHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("user")
	// fmt.Println("start", name)
	var book Book
	var errors []string
	r.ParseForm()
	isbnStr := r.FormValue("ISBN")
	priceStr := r.FormValue("price")
	stockStr := r.FormValue("stock")

	// Check if the input values are empty
	if isbnStr == "" || priceStr == "" || stockStr == "" {
		errors = append(errors, "ISBN, price, and stock fields are required.")
	} else {
		// Convert the input values to integers
		isbn, err1 := strconv.Atoi(isbnStr)
		price, err2 := strconv.Atoi(priceStr)
		stock, err3 := strconv.Atoi(stockStr)

		if err1 != nil || err2 != nil || err3 != nil {
			errors = append(errors, "Invalid ISBN, price, or stock value.")
		}

		// Populate the book struct
		book.ISBN = isbn
		book.title = r.FormValue("title") // Corrected field name
		book.author = r.FormValue("author")
		book.genre = r.FormValue("genre")
		book.publisher = r.FormValue("publisher")
		book.releasedata = r.FormValue("releasedata") // Corrected field name
		book.format = "eBook"
		book.price = price
		book.stock = stock
	}
	fmt.Println("New Book Added: ", book.ISBN, book.title, book.author, book.genre, book.publisher, book.releasedata, book.format, book.price, book.stock)
	// Prepare the database query
	_, err := db.Exec("INSERT INTO books (isbn, title, author, genre, publisher, releasedata, format, price, stock) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		book.ISBN, book.title, book.author, book.genre, book.publisher, book.releasedata, book.format, book.price, book.stock)

	if err != nil {
		fmt.Println("app AB00x02: Book INSERT Error.", err)
		errors = append(errors, "Something went wrong. Try Again.")
	} else {
		errors = append(errors, "Book Added Successfully.")
	}

	// data := struct {
	// 	Errors []string
	// }{
	// 	Errors: errors,
	// }

	if len(errors) > 0 {
		// tmpl, err := template.ParseFiles("Admindashboard.html")

		// fmt.Println(name)
		// if err != nil {
		// 	fmt.Println("app AB00x03: ParseFiles error.", err)
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
		// tmpl.Execute(w, data)
		// return
		http.Redirect(w, r, "/Admindashboard.html?user="+name, http.StatusSeeOther)
	}

	// fmt.Println(name + "out")
	http.Redirect(w, r, "/Admindashboard.html?user="+name, http.StatusSeeOther)
}

func removeBookHandler(w http.ResponseWriter, r *http.Request) {
	// Get the user name from the query parameters
	name := r.URL.Query().Get("user")
	fmt.Println("remove book", name)
	// Parse the form data
	err := r.ParseForm()

	ISBNstr := r.FormValue("ISBN")
	ISBN, err := strconv.Atoi(ISBNstr)
	if err != nil {
		http.Error(w, "Invalid ISBN value.", http.StatusBadRequest)
		return
	}

	_, qerr := db.Exec("DELETE FROM books WHERE ISBN = $1", ISBN)
	if qerr != nil {
		if qerr == sql.ErrNoRows {
			errors := []string{"Book Not Found"}

			data := struct {
				Error []string
			}{
				Error: errors,
			}
			http.Redirect(w, r, "/Admindashboard.html?user="+name, http.StatusSeeOther)
			tmpl, err := template.ParseFiles("Admindashboard.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = tmpl.Execute(w, data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			return
		}

		http.Error(w, qerr.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/Admindashboard.html?user="+name, http.StatusSeeOther)
}
