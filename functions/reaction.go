package functions

import (
	"fmt"
	"log"
)

func RegisterReactionToDb(post_id int, user_id int, like bool) {
	var reaction_bool int
	if like {
		reaction_bool = 1
	} else {
		reaction_bool = 0
	}

	statement, err := db.Prepare("INSERT INTO reaction(post_id, user_id, reaction_bool) VALUES(?, ?, ?)")
	if err != nil {
		log.Printf("Error preparing data: %v", err)
		return
	}
	defer statement.Close()
	_, err = statement.Exec(post_id, user_id, reaction_bool)
	if err != nil {
		log.Printf("Error executing data: %v", err)
		return
	}
	fmt.Println("Inserted reaction data into database:", post_id, user_id, reaction_bool)
}

func CheckForReaction(post_id int, user_id int) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM reaction WHERE post_id = ? AND user_id = ?)", post_id, user_id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func RemoveReactionFromDb(post_id int, user_id int) error {
	statement, err := db.Prepare("DELETE FROM reaction WHERE post_id = ? AND user_id = ?")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(post_id, user_id)
	if err != nil {
		return err
	}
	fmt.Printf("Deleted reaction for user (%v) on post (%v).\n", user_id, post_id)
	return nil
}
