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
	functions.InitDb()
	defer functions.CloseDb()
	tpl, _ = template.ParseGlob("templates/*.html")
	port := "8080"
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/post/", PostHandler)
	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/register", RegisterHandler)
	http.HandleFunc("/logout", LogoutHandler)
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

	posts, err := functions.GetPostsFromDb()
	if err != nil || posts == nil {
		w.Header().Set("Content-Type", "text/html")
		tpl.ExecuteTemplate(w, "index.html", nil) //replace nil with data
		return
	}

	data := struct {
		Posts []functions.Post
	}{
		Posts: posts,
	}

	fmt.Print(data)

	w.Header().Set("Content-Type", "text/html")
	tpl.ExecuteTemplate(w, "index.html", data) //replace nil with data
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received (/register) a request with method:", r.Method)
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

		exists, err := functions.UserExists(username, email)
		if err != nil {
			http.Error(w, "Error checking user existence", http.StatusInternalServerError)
			return
		}
		if exists {
			http.Error(w, "Username or Email already in use", http.StatusConflict)
			return
		}
		passwordHash, _ := functions.HashPassword(password)
		/* match := functions.CheckPasswordHash(password, passwordHash)
		fmt.Println(match) */
		fmt.Println("Form data:", username, firstname, lastname, email)

		functions.RegisterUserToDb(username, firstname, lastname, passwordHash, email)

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		fmt.Fprintln(w, "Welcome, you are registered, please login in!")
	}

}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "templates/login.html")
		_, err := functions.AuthenticateUser(w, r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusMovedPermanently)
			return
		}
		// http.Redirect(w, r, "/", http.StatusMovedPermanently)
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
			http.Error(w, "Invalid login credentials", http.StatusUnauthorized)
			return
		} else {
			// log.Printf("Retrieved user data: %+v\n", user)
		}
		match := functions.CheckPasswordHash(password, user.Password)
		if !match {
			fmt.Println("Wrong password!")
			http.Redirect(w, r, "/login", http.StatusUnauthorized)
		}

		sessionID, err := functions.GenerateSessionID(user.Password)
		if err != nil {
			ErrorHandler(w, "Error generating session ID", http.StatusInternalServerError)
			return
		}

		// cookieName, err := functions.GenerateCookieName(user.Email)
		// if err != nil {
		// 	fmt.Print(err)
		// }

		cookieName := "brownie" //??vb peaks kasutama nime generaatorit??
		fmt.Printf("cookie name: %s\ncookie value: %s\n", cookieName, sessionID)

		functions.StoreSessionInDb(sessionID, *user)

		functions.NewCookie(w, cookieName, sessionID)
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {

	user_id, err := functions.AuthenticateUser(w, r)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
	}
	err = functions.DeleteSessionFromDb(user_id)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "brownie",
		Path:   "/",
		MaxAge: -1, //MaxAge <0 means delete cookie now
	})
	fmt.Printf("Deleted %v's session", user_id)
	http.Redirect(w, r, "/", http.StatusPermanentRedirect)
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

func PostHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/post/") {
		http.NotFound(w, r)
		return
	}

	postID := strings.TrimPrefix(r.URL.Path, "/post/")
	post_id, err := strconv.Atoi(postID)
	if err != nil {
		fmt.Println("Error converting id from string to int")
	}
	post, err := functions.GetPostById(post_id)
	if err != nil {
		fmt.Println("Error getting post info from database")
	}

	tpl.ExecuteTemplate(w, "templates/post.html", post)
}
