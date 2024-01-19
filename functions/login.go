package functions

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       int
	Email    string
	Password string
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 5)
	return string(bytes), err

}
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
func GetUserByEmail(email string) (*User, error) {
	var user User
	err := db.QueryRow("SELECT id, email, password FROM users WHERE email = ?", email).Scan(&user.Id, &user.Email, &user.Password)
	// fmt.Print(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil

}

func GetUserByID(userID int) (string, error) {
	fmt.Println("GetUserByID called with userID:", userID)

	rows, err := db.Query("SELECT username FROM users WHERE id= ?", userID)
	if err != nil {
		fmt.Println("Error executing query:", err)
		return "", err
	}
	defer rows.Close()

	var username string

	for rows.Next() {
		if err := rows.Scan(&username); err != nil {
			fmt.Println("Error scanning row:", err)
			return "", err
		}
		fmt.Println("Scanned username:", username)
	}

	if username == "" {
		fmt.Println("No username found for userID:", userID)
		return "", errors.New("no username found for the given userID")
	}

	fmt.Println("Returning username:", username)
	return username, nil
}
