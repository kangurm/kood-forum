package functions

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

// Returns sessionID, err

func GenerateSessionID(password string) (string, error) {

	// Generate a random 16-byte session ID
	sessionIDBytes := []byte(password)
	sessionID, err := bcrypt.GenerateFromPassword(sessionIDBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Encode the hashed session ID to base64 to create a string
	//sessionID := base64.URLEncoding.EncodeToString(sessionIDBytes)
	sessionIDstring := string(sessionID)
	return sessionIDstring, nil
}

func GenerateCookieName(email string) (string, error) {

	emailBytes := []byte(email)
	nameBytes, err := bcrypt.GenerateFromPassword(emailBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	name := string(nameBytes)
	fmt.Printf("Cookie name is: %s\n", name)
	return name, err
}

func NewCookie(w http.ResponseWriter, key string, value string) {
	cookie := http.Cookie{
		Name:  key,
		Value: value,
	}
	http.SetCookie(w, &cookie)
}

// Return cookie's value after providing the name of cookie
func GetCookieValue(r *http.Request, key string) (string, error) {
	cookie, err := r.Cookie(key)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// Associates user and sessionID, saves it to the database.
func StoreSessionInDb(sessionID string, userData User) error {
	_, err := db.Exec("INSERT INTO session (session_id, user_id, email) VALUES (?, ?, ?)",
		sessionID, userData.Id, userData.Email)
	return err
}
