package functions

import "fmt"

type Post struct {
	Post_id string
	User_id string
	Title   string
	Text    string
	Created string
	LikeCount string
	DislikeCount string
	CommentCount string
}

type Reaction struct {
	Reaction_id string
	Post_id     string
	User_id     string
	Comment_id  string
	Like        bool
	Created     string
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
	rows, err := db.Query("SELECT id, user_id, postTitle, postBody, created FROM post WHERE id = ?", postID)
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

// func GetReactionsForPost(postID int) ([]Reaction, error) {
// 	rows, err := db.Query("SELECT id, post_id, user_id, comment_id, reaction_bool, created FROM reaction WHERE post_id = ?", postID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var reactions []Reaction

// 	for rows.Next() {
// 		var reaction Reaction
// 		if err := rows.Scan(&reaction.Reaction_id, &reaction.Post_id, &reaction.User_id, &reaction.Comment_id, &reaction.Like, &reaction.Created); err != nil {
// 			return nil, err
// 		}
// 		reactions = append(reactions, reaction)
// 	}

// 	if err := rows.Err(); err != nil {
// 		return nil, err
// 	}

// 	return reactions, nil
// }
