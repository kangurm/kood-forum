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

type TemplateData struct {
	Username   string
	IsLoggedIn bool
}

func main() {
	functions.InitDb()
	defer functions.CloseDb()
	tpl, _ = template.ParseGlob("templates/*.html")
	port := "8080"
	http.HandleFunc("/", IndexHandler)
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
	logged, err := functions.AuthenticateUser(w, r)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		tpl.ExecuteTemplate(w, "index.html", TemplateData{IsLoggedIn: false})
		fmt.Println("user is not logged in")
	} else {
		username, err := functions.GetUserByID(logged)
		if err != nil {
			http.Error(w, "cant find username from database", http.StatusInternalServerError)
			fmt.Print(username)
		}
		w.Header().Set("Content-Type", "text/html")
		tpl.ExecuteTemplate(w, "index.html", TemplateData{Username: username, IsLoggedIn: true})
	}
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

		postTitle := r.FormValue("userPostTitle")
		postBody := r.FormValue("userPostBodyText")

		// TODO: Find it out using a cookie
		user_id := 0

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
