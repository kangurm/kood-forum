package functions

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite3", "/Users/rainpraks/golang/forum/db")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
}

func registerUserToDb(username, firstname, lastname, password, email string) {

	statment, err := db.Prepare("INSERT INTO projects(id, username, firstname, lastname, password, email) VALUES( ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

}
