package functions

import (
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
	err := db.QueryRow("SELECT id, email, password FROM user WHERE email = ?", email).Scan(&user.Id, &user.Email, &user.Password)
	// fmt.Print(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil

}
