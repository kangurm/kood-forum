package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"log"
	"net/http"
)

var tpl *template.Template

//This is a huge focking comment

func main() {
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
	tpl, _ = template.ParseGlob("templates/*.html")
	port := "8080"
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/login", LoginHandler)
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
