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
	LikedPost      bool
	DislikePost    bool
}

// GenerateSessionID function generates a session id by hashing the given password
// by using bcrypt algorithm
// function is needed to not track users from sessions
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

// GenerateCookieName generates a cookie name by hashing the given email
// using the bcrypt algorithm
// NOT IN USE
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

// NewCookie creates a new HTTP cookie with specified name and value
// Sets it on the HTTP response.
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

// GetUserIdFromSession requests user id from session table
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

// DeleteSessionFromDb deletes user's session from the session table using the user id
func DeleteSessionFromDb(user_id int) error {

	fmt.Println("User id: ", user_id)

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

// AuthenticateUser function is needed to understand if user is logged in or not
// based on cookie in an HTTP request
func AuthenticateUser(w http.ResponseWriter, r *http.Request) (LoggedUser, error) {

	var loggedUser LoggedUser
	//request cookie from HTTP request
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

// RemoveCookieFromClient removes cookie named "Forum" from clients browser
func RemoveCookieFromClient(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "forum",
		Path:   "/",
		MaxAge: -1, //MaxAge <0 means delete cookie now
	})
}

// NoCacheHeaders sets HTTP headers to prevent caching of the response
// Header().Set() method is called on w to set various HTTP headers
func NoCacheHeaders(w http.ResponseWriter) {
	//tells to browser that not cache the response
	//must validate cache with server before using it
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	//No cache HTTP/1.0 older version headers
	w.Header().Set("Pragma", "no-cache")
	//Expires gives date/time after which the response is considered to stale.
	//0 means that response is already expired
	w.Header().Set("Expires", "0")
	//client caches that can response might be different for every request,
	//they should not use a cached response to satisfy the request
	w.Header().Set("Vary", "*")
}
