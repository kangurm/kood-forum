package main

import (
	"fmt"
	"forum/functions"
	"forum/handlers"
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
	http.HandleFunc("/", handlers.IndexHandler)
	http.HandleFunc("/post/", handlers.PostHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/create-a-post", handlers.CreateAPostHandler)
	http.HandleFunc("/post/react", handlers.ReactionHandler)
	http.HandleFunc("/post/comment", handlers.CreateACommentHandler)
	fmt.Println("Server running at http://localhost:" + port)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":"+port, nil)
}
