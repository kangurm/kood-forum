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
	fmt.Println("database opened")
	if err != nil {
		log.Fatal(err)
	}
}
func CloseDb() {
	db.Close()
	fmt.Println("Database closed")
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
func UserExists(username, email string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ? OR email = ?)", username, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
