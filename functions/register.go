package functions

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDb() {
	var err error
	db, err = sql.Open("sqlite3", "db/database.db")
	if err != nil {
		log.Fatal(err)
	}
}
func CloseDb() {
	db.Close()
}

func RegisterUserToDb(username, firstname, lastname, password, email string) {

	statement, err := db.Prepare("INSERT INTO users(username, firstname, lastname, password, email) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		log.Printf("Error preparing data: %v", err)
		return
	}
	defer statement.Close()
	_, err = statement.Exec(username, firstname, lastname, password, email)
	if err != nil {
		log.Printf("Error executing data: %v", err)
		return
	}
	fmt.Println("Inserted data into database:", username, firstname, lastname, email)
}

func InsertUserToDb(username, firstname, lastname, password, email string) {

	action := "INSERT INTO"
	table := "users"
	dataLayout := "(username, firstname, lastname, password, email) VALUES(?, ?, ?, ?, ?)"


	statement, err := db.Prepare(action+table+dataLayout)
	if err != nil {
		log.Printf("Error preparing data: %v", err)
		return
	}
	defer statement.Close()
	_, err = statement.Exec(username, firstname, lastname, password, email)
	if err != nil {
		log.Printf("Error executing data: %v", err)
		return
	}
	fmt.Println("Inserted data into database:", username, firstname, lastname, email)
}

func RegisterPostToDb(user_id, postTitle, postBody string) {

}