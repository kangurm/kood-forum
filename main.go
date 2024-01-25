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

type LoggedUser struct {
	Username       string
	IsLoggedIn     bool
	ErrorMessage   string
	WelcomeMessage string
	UserExists     string
}

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
	http.HandleFunc("/post/", PostAndCommentHandler)
	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/register", RegisterHandler)
	http.HandleFunc("/logout", LogoutHandler)
	http.HandleFunc("/create-a-post", CreateAPostHandler)
	//http.HandleFunc("/post", PostHandler)
	http.HandleFunc("/post/comment", CreateACommentHandler)
	fmt.Println("Server running at http://localhost:" + port)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":"+port, nil)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		ErrorHandler(w, "Page not found", http.StatusNotFound)
		return
	}

	posts, err := functions.GetPostsFromDb()
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		tpl.ExecuteTemplate(w, "index.html", nil) //replace nil with data
		return
	}

	var username string
	logUser := LoggedUser{Username: username, IsLoggedIn: true}

	user_id, err := functions.AuthenticateUser(w, r)
	if err != nil {
		logUser.IsLoggedIn = false
	} else {
		username, err := functions.GetUserByID(user_id)
		if err != nil {
			http.Error(w, "cant find username from database", http.StatusInternalServerError)
			fmt.Print(username)
		}
		logUser.Username = username
		logUser.IsLoggedIn = true
	}

	data := struct {
		Posts      []functions.Post
		LoggedUser LoggedUser
	}{
		Posts:      posts,
		LoggedUser: logUser,
	}

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
			tpl.ExecuteTemplate(w, "register.html", LoggedUser{UserExists: "Username or Email already in use"})
			return
		}
		passwordHash, _ := functions.HashPassword(password)
		/* match := functions.CheckPasswordHash(password, passwordHash)
		fmt.Println(match) */
		fmt.Println("Form data:", username, firstname, lastname, email)

		functions.RegisterUserToDb(username, firstname, lastname, passwordHash, email)
		w.Header().Set("Content-Type", "text/html")
		tpl.ExecuteTemplate(w, "login.html", LoggedUser{WelcomeMessage: "Welcome, you are registered, please login in!"})

		return
	}

}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		user_id, err := functions.AuthenticateUser(w, r)
		if err != nil || user_id == 0 {
			tpl.ExecuteTemplate(w, "login.html", nil)
			return
		}
		fmt.Println("User is already logged in, redirecting to index.")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing the form(login)", http.StatusInternalServerError)
			return
		}
		email := r.FormValue("email")
		password := r.FormValue("password")

		user, err := functions.GetUserByEmail(email)
		if err != nil {
			log.Printf("Error retrieving user: %v\n", err)
			tpl.ExecuteTemplate(w, "login.html", LoggedUser{ErrorMessage: "Invalid email or password"})
			return
		}

		match := functions.CheckPasswordHash(password, user.Password)
		if !match {
			fmt.Println("Wrong password!")
			tpl.ExecuteTemplate(w, "login.html", LoggedUser{ErrorMessage: "Invalid email or password"})
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

		cookieName := "brownie" //??vb peaks kasutama nime generaatorit??
		fmt.Printf("cookie name: %s\ncookie value: %s\n", cookieName, sessionID)

		functions.StoreSessionInDb(sessionID, *user)

		functions.NewCookie(w, cookieName, sessionID)
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {

	user_id, err := functions.AuthenticateUser(w, r)
	if err != nil || user_id == 0 {
		fmt.Println(err)
		fmt.Println("aaa")
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
	}

	err = functions.DeleteSessionFromDb(user_id)
	if err != nil {
		fmt.Println(err)
		fmt.Println("eee")
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
	}
	functions.RemoveCookieFromClient(w)
	fmt.Printf("Deleted %v's session", user_id)
	http.Redirect(w, r, "/", http.StatusPermanentRedirect)
}

func CreateAPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		user_id, err := functions.AuthenticateUser(w, r)
		if err != nil || user_id == 0 {
			fmt.Println("User is not logged in. Redirecting to login.")
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		}
		http.ServeFile(w, r, "templates/create-a-post.html")
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

		user_id, err := functions.AuthenticateUser(w, r)
		if err != nil || user_id == 0 {
			fmt.Println("User is not logged in. Redirecting to login.")
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		}

		functions.RegisterPostToDb(user_id, postTitle, postBody)
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

func PostAndCommentHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/post/") {
		http.NotFound(w, r)
		return
	}

	postID := strings.TrimPrefix(r.URL.Path, "/post/")
	post_id, err := strconv.Atoi(postID)
	if err != nil {
		fmt.Println("Error converting id from string to int")
	}
	currentPost, err := functions.GetPostById(post_id)
	if err != nil {
		fmt.Println("Error getting post info from database")
	}
	currentComments, err := functions.GetCommentsByPostId(post_id)
	if err != nil {
		fmt.Println("Error getting comment info from database")
	}
	data := struct {
		Post     functions.Post
		PostID   int
		Comments []functions.Comment
	}{
		Post:     currentPost,
		PostID:   post_id,
		Comments: currentComments,
	}

	tpl.ExecuteTemplate(w, "post.html", data)
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
		userID, err := functions.AuthenticateUser(w, r)
		if err != nil || userID == 0 {
			fmt.Println("User is not logged in. To make a comment, the user must be logged in.")
			http.Redirect(w, r, "/post/", http.StatusTemporaryRedirect)
		}

		functions.RegisterCommentToDb(userID, postID, commentBody)
		http.Redirect(w, r, "/post/"+postIDStr, http.StatusMovedPermanently)
	}
}
