package functions

import "fmt"

type Post struct {
	Post_id      string
	User_id      string
	Title        string
	Text         string
	Created      string
	LikeCount    int
	DislikeCount int
	CommentCount int
}

func GetPostsFromDb() ([]Post, error) {
	rows, err := db.Query("SELECT id, user_id, postTitle, postBody, created, like_count, dislike_count, comment_count FROM post")
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var posts []Post

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.Post_id, &post.User_id, &post.Title, &post.Text, &post.Created, &post.LikeCount, &post.DislikeCount, &post.CommentCount); err != nil {
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
	rows, err := db.Query("SELECT id, user_id, postTitle, postBody, created, like_count, dislike_count, comment_count FROM post WHERE id = ?", postID)
	if err != nil {
		return Post{}, err
	}
	defer rows.Close()

	var post Post
	for rows.Next() {
		if err := rows.Scan(&post.Post_id, &post.User_id, &post.Title, &post.Text, &post.Created, &post.LikeCount, &post.DislikeCount, &post.CommentCount); err != nil {
			return Post{}, err
		}
	}

	if err := rows.Err(); err != nil {
		return Post{}, err
	}

	return post, nil
}
