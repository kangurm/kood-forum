package functions

import (
	"encoding/base64"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// Returns sessionID, err
func GenerateSessionID() (string, error) {

	// Generate a random 16-byte session ID
	sessionIDBytes := make([]byte, 16)
	_, err := bcrypt.GenerateFromPassword(sessionIDBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Encode the hashed session ID to base64 to create a string
	sessionID := base64.URLEncoding.EncodeToString(sessionIDBytes)

	return sessionID, nil
}

func SetNewSession(w http.ResponseWriter, key string, value string) {
	cookie := http.Cookie{
		Name:  key,
		Value: value,
		Path:  "/",
	}

	http.SetCookie(w, &cookie)
}

func GetSession(r *http.Request, key string) (string, error) {
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
