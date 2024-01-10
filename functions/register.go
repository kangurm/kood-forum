package functions

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDb() {
	var err error
	db, err = sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Fatal(err)
	}

}

func RegisterUserToDb(username, firstname, lastname, password, email string) {

	statement, err := db.Prepare("INSERT INTO projects(username, firstname, lastname, password, email) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec(username, firstname, lastname, password, email)
	if err != nil {
		log.Fatalf("Error executing statement: %v", err)
	}
	fmt.Println("Inserted data into database:", username, firstname, lastname, email)

}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received a request with method:", r.Method)
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

		RegisterUserToDb(username, firstname, lastname, password, email)
	}

}
