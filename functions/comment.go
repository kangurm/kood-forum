package functions

import (
	"fmt"
	"log"
)

type Comment struct {
	Comment_id   string
	User_id      string
	Post_id      string
	Text         string
	Reaction_id  string
	Created      string
	Username     string
	LikeCount    int
	DislikeCount int
}

// RegisterCommentToDb stores new comment into the comment table.
func RegisterCommentToDb(user_id int, post_id int, text string, username string) {
	statement, err := db.Prepare("INSERT INTO comment(user_id, post_id, text, username) VALUES(?, ?, ?, ?)")
	if err != nil {
		log.Printf("Error preparing data: %v", err)
		return
	}
	defer statement.Close()
	_, err = statement.Exec(user_id, post_id, text, username)
	if err != nil {
		log.Printf("Error executing data: %v", err)
		return
	}
	fmt.Println("Inserted data into database:", user_id, post_id, text)
}

// GetCommentsByPostId retrieves the comments of a post with given id from comment table
func GetCommentsByPostId(post_id int) ([]Comment, error) {
	rows, err := db.Query("SELECT id, post_id, user_id, text, created, username, like_count, dislike_count FROM comment WHERE post_id = ?", post_id)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.Comment_id, &comment.Post_id, &comment.User_id, &comment.Text, &comment.Created, &comment.Username, &comment.LikeCount, &comment.DislikeCount); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}
