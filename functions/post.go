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

func GetPostsFromDb() ([]Post, error) {
	rows, err := db.Query("SELECT id, user_id, postTitle, postBody, created, like_count, dislike_count, comment_count, username FROM post")
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var posts []Post

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.Post_id, &post.User_id, &post.Title, &post.Text, &post.Created, &post.LikeCount, &post.DislikeCount, &post.CommentCount, &post.Username); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func GetPostById(postID int) (Post, error) {
	log.Printf("Fetching post with ID: %d", postID)
	rows, err := db.Query("SELECT id, user_id, postTitle, postBody, created, like_count, dislike_count, comment_count, username FROM post WHERE id = ?", postID)
	if err != nil {
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
