package functions

import "fmt"

type Post struct {
	Post_id string
	User_id string
	Title   string
	Text    string
	Created string
}

func GetCategoriesFromDb() []string {
	rows, err := db.Query("SELECT text FROM category")
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var categories []string

	for rows.Next() {
		var categoryText string
		if err := rows.Scan(&categoryText); err != nil {
			fmt.Println(err)
		}
		categories = append(categories, categoryText)
	}
	return categories
}

func GetPostsFromDb() ([]Post, error) {
	rows, err := db.Query("SELECT id, user_id, postTitle, postBody, created FROM post")
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var posts []Post

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.Post_id, &post.User_id, &post.Title, &post.Text, &post.Created); err != nil {
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
	rows, err := db.Query("SELECT user_id, postTitle, postBody, created FROM post WHERE id = ?", postID)
	if err != nil {
		return Post{}, err
	}
	defer rows.Close()

	var post Post

	for rows.Next() {
		if err := rows.Scan(&post.Post_id, &post.User_id, &post.Title, &post.Text, &post.Created); err != nil {
			return Post{}, err
		}
	}

	if err := rows.Err(); err != nil {
		return Post{}, err
	}
	fmt.Println("Retrieved post info: ", post)
	return post, nil
}
