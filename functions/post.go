package functions

import (
	"fmt"
	"log"
)

type Post struct {
	Post_id      int
	User_id      string
	Title        string
	Text         string
	Created      string
	LikeCount    int
	DislikeCount int
	CommentCount int
	Categories   []string
	Username     string
}

// GetPostsFromDb retrieves all posts from db
// returns them as slice of Posts structs
func GetPostsFromDb() ([]Post, error) {
	//selects all rows from post tabel and returns *sql.rows pointer value
	rows, err := db.Query("SELECT id, user_id, postTitle, postBody, created, like_count, dislike_count, comment_count, username FROM post")
	if err != nil {
		fmt.Println(err)
	}
	//execution of row.Close soon as function returns
	defer rows.Close()

	var posts []Post

	//iterates over the rows in the result set
	for rows.Next() {
		var post Post
		//scans the values from current row into Post struct.
		if err := rows.Scan(&post.Post_id, &post.User_id, &post.Title, &post.Text, &post.Created, &post.LikeCount, &post.DislikeCount, &post.CommentCount, &post.Username); err != nil {
			//if an error occured during scanning, it returns nil and and the error
			return nil, err
		}
		posts = append(posts, post)
	}
	//if any error occured during the iteration. If an error occured, it returns nil and error
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// if no errors then function returns error slice
	return posts, nil
}

func GetPostById(postID int) (Post, error) {
	log.Printf("Fetching post with ID: %d", postID)
	rows, err := db.Query("SELECT id, user_id, postTitle, postBody, created, like_count, dislike_count, comment_count, username FROM post WHERE id = ?", postID)
	if err != nil {
		fmt.Println("aiaiai")
		return Post{}, err

	}
	defer rows.Close()

	var post Post
	for rows.Next() {
		if err := rows.Scan(&post.Post_id, &post.User_id, &post.Title, &post.Text, &post.Created, &post.LikeCount, &post.DislikeCount, &post.CommentCount, &post.Username); err != nil {
			return Post{}, err
		}
	}

	if err := rows.Err(); err != nil {
		return Post{}, err
	}

	return post, nil
}

func RegisterPostToDb(user_id int, postTitle, postBody string, username string) {

	statement, err := db.Prepare("INSERT INTO post(user_id, postTitle, postBody, username) VALUES(?, ?, ?, ?)")
	if err != nil {
		log.Printf("Error preparing data: %v", err)
		return
	}
	defer statement.Close()
	_, err = statement.Exec(user_id, postTitle, postBody, username)
	if err != nil {
		log.Printf("Error executing data: %v", err)
		return
	}
	fmt.Println("Inserted data into database:", user_id, postTitle, postBody, username)
}

func GetPostByContent(user_id int, postTitle, postBody string) int {
	var post_id int
	err := db.QueryRow("SELECT id FROM post WHERE user_id = ? AND postTitle = ? AND postBody = ?", user_id, postTitle, postBody).Scan(&post_id)
	if err != nil {
		fmt.Println("EEEE")
		return 0
	}
	return post_id
}

func CheckIfPostExists(postID int) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM post WHERE id = ?)", postID).Scan(&exists)
	if err != nil {
		fmt.Println("Error in checkIfPostExists line 102")
		return exists, err
	}
	return exists, nil
}
