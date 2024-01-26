package main

import (
	"fmt"
	"forum/functions"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var tpl *template.Template

func main() {
	var err error
	functions.InitDb()
	defer functions.CloseDb()
	tpl, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Error parsing remplates: %v", err)
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

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	functions.NoCacheHeaders(w)

	if r.URL.Path != "/" {
		ErrorHandler(w, "Page not found", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		ErrorHandler(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	// Posts sorting logic
	var posts []functions.Post
	var err error
	action := r.URL.Query().Get("sort")
	switch action {
	case "top":
		posts, err = functions.SortByTop()
		if err != nil {
			fmt.Println("Error sorting")
		}
	case "new":
		functions.SortByNew()
	case "hot":
		functions.SortByHot()
	default:
		posts, err = functions.GetPostsFromDb()
		if err != nil {
			w.Header().Set("Content-Type", "text/html")
			tpl.ExecuteTemplate(w, "index.html", nil) //replace nil with data
			return
		}
	}

	loggedUser, err := functions.AuthenticateUser(w, r)
	if err != nil {
		fmt.Println("Not logged in")
		loggedUser.IsLoggedIn = false
	}

	categories, err := functions.GetCategoriesFromDb()
	if err != nil {
		fmt.Println("Error getting categories: ", err)
	}

	var comments struct{}

	data := functions.BuildResponse(loggedUser, posts, comments, categories)

	w.Header().Set("Content-Type", "text/html")
	tpl.ExecuteTemplate(w, "index.html", data)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received (/register) a request with method:", r.Method)
	if r.Method == "GET" {
		err := tpl.ExecuteTemplate(w, "register.html", nil)
		if err != nil {
			log.Printf("Error executing template: %v", err)
		}
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

		exists, err := functions.UserExists(username, email)
		if err != nil {
			http.Error(w, "Error checking user existence", http.StatusInternalServerError)
			return
		}
		if exists {
			w.Header().Set("Content-Type", "text/html")
			tpl.ExecuteTemplate(w, "register.html", functions.LoggedUser{UserExists: "Username or Email already in use"})
			return
		}
		passwordHash, _ := functions.HashPassword(password)
		fmt.Println("Form data:", username, firstname, lastname, email)

		functions.RegisterUserToDb(username, firstname, lastname, passwordHash, email)
		w.Header().Set("Content-Type", "text/html")
		tpl.ExecuteTemplate(w, "login.html", functions.LoggedUser{WelcomeMessage: "Welcome, you are registered, please login in!"})
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		loggedUser, err := functions.AuthenticateUser(w, r)
		if err != nil {
			loggedUser.IsLoggedIn = false
			tpl.ExecuteTemplate(w, "login.html", nil)
			return
		}
		fmt.Println("User is already logged in, redirecting to index.")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	if r.Method == "POST" {
		loggedUser, err := functions.AuthenticateUser(w, r)
		if err != nil {
			loggedUser.IsLoggedIn = false
		}

		err = r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing the form (login)", http.StatusInternalServerError)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		user, err := functions.GetUserByEmail(email)
		if err != nil {
			log.Printf("Error retrieving user: %v\n", err)
			loggedUser.ErrorMessage = "Invalid email or password"
			data := functions.BuildResponse(loggedUser)
			tpl.ExecuteTemplate(w, "login.html", data)
			return
		}

		match := functions.CheckPasswordHash(password, user.Password)
		if !match {
			loggedUser.ErrorMessage = "Invalid email or password"
			data := functions.BuildResponse(loggedUser)
			tpl.ExecuteTemplate(w, "login.html", data)
			return
		}

		err = functions.DeleteSessionFromDb(user.Id)
		if err != nil {
			fmt.Println("Failed to delete session from database after user logged in")
		}

		functions.RemoveCookieFromClient(w)

		sessionID, err := functions.GenerateSessionID(user.Password)
		if err != nil {
			ErrorHandler(w, "Error generating session ID", http.StatusInternalServerError)
			return
		}

		functions.StoreSessionInDb(sessionID, *user)

		cookieName := "forum"

		functions.NewCookie(w, cookieName, sessionID)
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {

	loggedUser, err := functions.AuthenticateUser(w, r)
	if err != nil || loggedUser.Id == 0 {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
		return
	}

	err = functions.DeleteSessionFromDb(loggedUser.Id)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
		return
	}
	functions.RemoveCookieFromClient(w)
	fmt.Printf("Deleted %v's session", loggedUser.Id)
	http.Redirect(w, r, "/", http.StatusPermanentRedirect)
}

func CreateAPostHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		loggedUser, err := functions.AuthenticateUser(w, r)
		if err != nil || loggedUser.Id == 0 {
			fmt.Println("User is not logged in. Redirecting to login.")
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		data := functions.BuildResponse(loggedUser)
		tpl.ExecuteTemplate(w, "create-a-post.html", data)
		return
	}
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			ErrorHandler(w, "Error parsing the form", http.StatusInternalServerError)
		}

		// TODO: Add categories to html and use them here.

		postTitle := r.FormValue("userPostTitle")
		postBody := r.FormValue("userPostBodyText")

		loggedUser, err := functions.AuthenticateUser(w, r)
		if err != nil || loggedUser.Id == 0 {
			fmt.Println("User is not logged in. Redirecting to login.")
			// TODO: Replace with message to login
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		}

		functions.RegisterPostToDb(loggedUser.Id, postTitle, postBody)
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
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

func PostHandler(w http.ResponseWriter, r *http.Request) {
	functions.NoCacheHeaders(w)

	if !strings.HasPrefix(r.URL.Path, "/post/") {
		ErrorHandler(w, "Status not found", http.StatusNotFound)
		return
	}

	loggedUser, err := functions.AuthenticateUser(w, r)
	if err != nil || loggedUser.Id == 0 {
		loggedUser.IsLoggedIn = false
	}

	postID := strings.TrimPrefix(r.URL.Path, "/post/")
	post_id, err := strconv.Atoi(postID)
	if err != nil {
		fmt.Println("Error converting id from string to int")
		ErrorHandler(w, "Internal server error", http.StatusInternalServerError)
	}
	currentPost, err := functions.GetPostById(post_id)
	if err != nil {
		fmt.Println("Error getting post info from database")
		ErrorHandler(w, "Internal server error", http.StatusInternalServerError)
	}
	currentComments, err := functions.GetCommentsByPostId(post_id)
	if err != nil {
		fmt.Println("Error getting comment info from database")
	}
	data := functions.BuildResponse(loggedUser, currentPost, currentComments)

	tpl.ExecuteTemplate(w, "post.html", data)
	fmt.Printf("%+v\n", currentComments)
}

func CreateACommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		err := r.ParseForm()
		if err != nil {
			ErrorHandler(w, "Error parsing the form", http.StatusInternalServerError)
			return
		}

		commentBody := r.FormValue("comment")
		fmt.Println(commentBody)

		postIDStr := r.URL.Query().Get("post_id")
		fmt.Println("postIDStr:", postIDStr)
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			ErrorHandler(w, "Invalid post ID", http.StatusBadRequest)
			return
		}
		loggedUser, err := functions.AuthenticateUser(w, r)
		if err != nil || loggedUser.Id == 0 {
			fmt.Println("User is not logged in. To make a comment, the user must be logged in.")
			http.Redirect(w, r, "/post/", http.StatusTemporaryRedirect)
		}

		username, err := functions.GetUserByID(loggedUser.Id)
		if err != nil {
			fmt.Println("Error getting username")
		}

		functions.RegisterCommentToDb(loggedUser.Id, postID, commentBody, username)
		http.Redirect(w, r, "/post/"+postIDStr, http.StatusMovedPermanently)
	}
}

func ReactionHandler(w http.ResponseWriter, r *http.Request) {
	functions.NoCacheHeaders(w)
	if r.Method == "GET" {
		loggedUser, err := functions.AuthenticateUser(w, r)
		if err != nil || loggedUser.Id == 0 {
			loggedUser.IsLoggedIn = false
		}

		postIDStr := r.URL.Query().Get("post_id")
		post_id, err := strconv.Atoi(postIDStr)
		if err != nil {
			ErrorHandler(w, "Post might be deleted", http.StatusBadRequest)
			return
		}

		action := r.URL.Query().Get("action")
		var like bool
		switch action {
		case "like":
			like = true
		case "dislike":
			like = false
		}

		if loggedUser.IsLoggedIn {
			functions.AddReactionToPost(post_id, loggedUser.Id, like, false)
		} else {
			// TODO: Asenda see error messagiga
			http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		}
		http.Redirect(w, r, "/post/"+postIDStr, http.StatusMovedPermanently)
	}
}
