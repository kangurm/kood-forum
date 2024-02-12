package functions

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func RegisterUserToDb(username, firstname, lastname, password, email string) {

	statement, err := db.Prepare("INSERT INTO user(username, firstname, lastname, password, email) VALUES(?, ?, ?, ?, ?)")
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
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM user WHERE username = ? OR email = ?)", username, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
