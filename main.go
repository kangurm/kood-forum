package main

import (
	"fmt"
	"forum/functions"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var tpl *template.Template

func main() {
	functions.InitDb()
	defer functions.CloseDb()
	tpl, _ = template.ParseGlob("templates/*.html")
	port := "8080"
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/register", RegisterHandler)
	http.HandleFunc("/create-a-post", CreateAPostHandler)
	fmt.Println("Server running at http://localhost:" + port)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":"+port, nil)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		ErrorHandler(w, "Page not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	tpl.ExecuteTemplate(w, "index.html", nil) //replace nil with data
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "login.html", nil)
}
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a request with method:", r.Method)
	if r.Method == "GET" {
		http.ServeFile(w, r, "templates/register.html")
		return
	}
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing the form", http.StatusInternalServerError)
			return
		}

		username := r.FormValue("username")
		firstname := r.FormValue("firstname")
		lastname := r.FormValue("lastname")
		password := r.FormValue("password")
		email := r.FormValue("email")

		fmt.Println("Form data:", username, firstname, lastname, email)

		functions.RegisterUserToDb(username, firstname, lastname, password, email)
	}

}

func CreateAPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// TODO: Check for cookie/if user is logged in
		http.ServeFile(w, r, "templates/create-a-post.html")
		return
	}
	if r.Method == "POST" {
		// TODO: Double check for cookie/if user is logged in?!

		err := r.ParseForm()
		if err != nil {
			ErrorHandler(w, "Error parsing the form", http.StatusInternalServerError)
		}

		postTitle := r.FormValue("userPostTitle")
		postBody := r.FormValue("userPostBodyText")

		// TODO: Find it out using a cookie
		user_id := 0

		functions.RegisterPostToDb(user_id, postTitle, postBody)
	}
}

func ErrorHandler(w http.ResponseWriter, s string, i int) {

	data := struct {
		StatusCode int
		Message    string
	}{
		StatusCode: i,
		Message:    s,
	}

	w.WriteHeader(i)
	tpl.ExecuteTemplate(w, "error.html", data)
}
