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

	var err error
	functions.InitDb()
	defer functions.CloseDb()
	tpl, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}
	port := "8080"
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/post/", PostHandler)
	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/register", RegisterHandler)
	http.HandleFunc("/logout", LogoutHandler)
	http.HandleFunc("/create-a-post", CreateAPostHandler)
	http.HandleFunc("/post/react", ReactionHandler)
	http.HandleFunc("/post/comment", CreateACommentHandler)
	fmt.Println("Server running at http://localhost:" + port)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":"+port, nil)
}
