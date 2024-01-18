package functions

import (
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
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
	fmt.Printf("Deleted session for user with user_id: %v", user_id)
	return nil
}

func AuthenticateUser(w http.ResponseWriter, r *http.Request) (userid int, err error) {
	cookie, err := r.Cookie("brownie")
	if err != nil {
		fmt.Println("Cookie not found from client:", err)
		return 0, err
	}
	fmt.Println(cookie.Value)
	user_id, err := GetUserIdFromSession(cookie.Value)
	if user_id == 0 || err != nil {
		fmt.Println("Cookie not found from database:", err, user_id)
		return 0, err
	}
	fmt.Println("user_id of user that is logged in: ", user_id)
	return user_id, nil
}
