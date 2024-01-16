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

//This is a huge focking comment

func main() {
	functions.InitDb()
	defer functions.CloseDb()
	tpl, _ = template.ParseGlob("templates/*.html")
	port := "8080"
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/register", RegisterHandler)
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

		cookieName := "wtf"
		fmt.Printf("cookie name: %s\ncookie value: %s\n", cookieName, sessionID)

		functions.StoreSessionInDb(sessionID, *user)

		// Ei ole kindel kas see on oige tegu. -Marcus
		functions.NewCookie(w, cookieName, sessionID)
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
