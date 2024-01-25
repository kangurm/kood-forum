package functions

import (
	"fmt"
	"log"
)

type Comment struct {
	Comment_id  string
	User_id     string
	Post_id     string
	Reaction_id string
	Created     string
}

func RegisterCommentToDb(user_id int, post_id int, text string) {
	statement, err := db.Prepare("INSERT INTO comment(user_id, post_id, text) VALUES(?, ?, ?)")
	if err != nil {
		log.Printf("Error preparing data: %v", err)
		return
	}
	defer statement.Close()
	_, err = statement.Exec(user_id, post_id, text)
	if err != nil {
		log.Printf("Error executing data: %v", err)
		return
	}
	fmt.Println("Inserted data into database:", user_id, post_id, text)
}
