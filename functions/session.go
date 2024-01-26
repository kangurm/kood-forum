package functions

import (
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type LoggedUser struct {
	Id             int
	Username       string
	IsLoggedIn     bool
	ErrorMessage   string
	WelcomeMessage string
	UserExists     string
}

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

func GetUserIdFromSession(sessionID string) (int, error) {
	rows, err := db.Query("SELECT user_id FROM session WHERE session_id = ?", sessionID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var user_id int

	for rows.Next() {
		if err := rows.Scan(&user_id); err != nil {
			return 0, err
		}
	}

	return user_id, nil
}

func DeleteSessionFromDb(user_id int) error {

	statement, err := db.Prepare("DELETE FROM session WHERE user_id = ?")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(user_id)
	if err != nil {
		return err
	}
	fmt.Printf("Deleted session for user with user_id: %v\n", user_id)
	return nil
}

func AuthenticateUser(w http.ResponseWriter, r *http.Request) (LoggedUser, error) {

	var loggedUser LoggedUser

	cookie, err := r.Cookie("forum")
	if err != nil {
		return LoggedUser{}, err
	}

	// Remove cookie from client if user not logged in
	user_id, err := GetUserIdFromSession(cookie.Value)
	if user_id == 0 || err != nil {
		RemoveCookieFromClient(w)
		return LoggedUser{}, err
	}

	username, err := GetUserByID(user_id)
	if err != nil {
		return LoggedUser{}, err
	}

	loggedUser.Id = user_id
	loggedUser.Username = username
	loggedUser.IsLoggedIn = true

	return loggedUser, nil
}

func RemoveCookieFromClient(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "forum",
		Path:   "/",
		MaxAge: -1, //MaxAge <0 means delete cookie now
	})
}

func NoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}
