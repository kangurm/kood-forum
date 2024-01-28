package functions

import (
	"fmt"
	"log"
)

type Reaction struct {
	Reaction_id string
	Post_id     string
	User_id     string
	Comment_id  string
	Like        bool
	Created     string
}

func RegisterReactionToDb(post_id int, user_id int, like int) error {

	statement, err := db.Prepare("INSERT INTO reaction(post_id, user_id, reaction_bool) VALUES(?, ?, ?)")
	if err != nil {
		log.Printf("Error preparing data: %v", err)
		return err
	}
	defer statement.Close()
	_, err = statement.Exec(post_id, user_id, like)
	if err != nil {
		log.Printf("Error executing data: %v", err)
		return err
	}
	fmt.Println("Inserted reaction data into database:", post_id, user_id, like)
	return nil
}

// reactionToRemove must be one of the following: "like_count", "dislike_count", "comment_count"
func RemoveReactionFromPost(post_id int, user_id int, reactionToRemove string) error {

	err := UpdateReactionCount(post_id, "", true, reactionToRemove)
	if err != nil {
		return fmt.Errorf("error removing reaction from post because of UpdateReactionCount")
	}

	statement, err := db.Prepare("DELETE FROM reaction WHERE post_id = ? AND user_id = ?")
	if err != nil {
		fmt.Println("STUCK 35")
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(post_id, user_id)
	if err != nil {
		fmt.Println("STUCK 42")
		return err
	}
	return nil
}

func UpdateReactionCount(post_id int, reactionTypeToAdd string, remove bool, reactionTypeToRemove string) error {

	if reactionTypeToAdd == "" && !remove {
		return fmt.Errorf("returned early from UpdateReactionCount, because input was invalid")
	}

	var reactionType string
	doRecursive := true

	// If we only want to remove count and not add a new one.
	if reactionTypeToAdd == "" && remove && reactionTypeToRemove != "" {
		reactionType = reactionTypeToRemove
		doRecursive = false
	}

	// If we only need to add count
	if reactionTypeToRemove == "" && !remove && reactionTypeToAdd != "" {
		reactionType = reactionTypeToAdd
		doRecursive = false
	}

	var exists bool
	// Check if reaction count (like_count etc) arent '0' for making sure count doesnt go to negatives.
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM post WHERE id = ? AND "+reactionType+" = ?)", post_id, 0).Scan(&exists)
	if remove && exists {
		fmt.Println("broken here?")
		return err
	}

	var template string
	// Looks like this if we want to remove: ("UPDATE post SET like_count = like_count - 1 WHERE post_id = ?")
	// or this if we want to add ("UPDATE post SET like_count = like_count + 1 WHERE post_id = ?")
	if remove {
		template = "UPDATE post SET " + reactionTypeToRemove + " = " + reactionTypeToRemove + " - 1 WHERE id = ?"
		reactionType = reactionTypeToRemove
	} else {
		template = "UPDATE post SET " + reactionTypeToAdd + " = " + reactionTypeToAdd + " + 1 WHERE id = ?"
		reactionType = reactionTypeToAdd
	}

	switch reactionType {
	case "like_count":
		statement, err := db.Prepare(template)
		if err != nil {
			fmt.Println("Error preparing like_count update")
		}
		_, err = statement.Exec(post_id)
		if err != nil {
			fmt.Println("Error updating like count")
		}
	case "dislike_count":
		statement, err := db.Prepare(template)
		if err != nil {
			fmt.Println("Error preparing dislike_count update")
		}
		_, err = statement.Exec(post_id)
		if err != nil {
			fmt.Println("Error updating dislike count")
		}
	case "comment_count":
		statement, err := db.Prepare(template)
		if err != nil {
			fmt.Println("Error preparing comment_count update")
		}
		_, err = statement.Exec(post_id)
		if err != nil {
			fmt.Println("Error updating comment count")
		}
	default:
		return fmt.Errorf("error in UpdateReactionCount switchcase")
	}

	if doRecursive {
		UpdateReactionCount(post_id, reactionTypeToAdd, false, "")
	}
	return nil
}

// Adds reaction to post, deals with reaction counts on post and automatically removes previous reactions.
// like = false is dislike, like = true is like, leave comment to false if no comment. If comment = true then adds comment.
func AddReactionToPost(post_id int, user_id int, like bool, comment bool) {
	reaction := 0
	reactionType := "dislike_count"
	if like {
		reaction = 1
		reactionType = "like_count"
	}

	var exists bool
	var previousReactionInt int
	var previousReactionStr string

	if comment {
		reaction = 2
		reactionType = "comment_count"
	}
	// Check if user has a like/dislike on the post already
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM reaction WHERE post_id = ? AND user_id = ?)", post_id, user_id).Scan(&exists)
	if err != nil {
		fmt.Println("Error checking if reaction exists")
	}
	// Get user's previous reaction
	err = db.QueryRow("SELECT reaction_bool FROM reaction WHERE post_id = ? AND user_id = ?", post_id, user_id).Scan(&previousReactionInt)
	if err != nil {
		fmt.Println("No previous reaction, proceeding...")
	}

	if previousReactionInt == 0 {
		if previousReactionInt == reaction && exists {
			RemoveReactionFromPost(post_id, user_id, reactionType)
			return
		}
		previousReactionStr = "dislike_count"
	} else if previousReactionInt == 1 {
		if previousReactionInt == reaction && exists {
			RemoveReactionFromPost(post_id, user_id, reactionType)
			return
		}
		previousReactionStr = "like_count"
	}

	// If user had like/dislike on the post, remove reaction count from POST table and add a new one, then update entry in REACTION table
	if exists {
		UpdateReactionCount(post_id, reactionType, true, previousReactionStr)

		statement, err := db.Prepare("UPDATE reaction SET reaction_bool = ? WHERE post_id = ? AND user_id = ?")
		if err != nil {
			fmt.Println("Error preparing update reaction")
		}
		_, err = statement.Exec(reaction, post_id, user_id)
		if err != nil {
			fmt.Println("Error updating reaction")
		}

		// If user doesnt have like/dislike on the post, then add reaction count to POST table and add a new entry to REACTION table.
	} else {
		UpdateReactionCount(post_id, reactionType, false, "")

		err := RegisterReactionToDb(post_id, user_id, reaction)
		if err != nil {
			fmt.Println("Error registering reaction to db ln187")
		}
	}
}
