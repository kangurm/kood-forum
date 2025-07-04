package functions

import (
	"fmt"
	"log"
)

type Category struct {
	ID      int
	Text    string
	URL     string
	Created string
	NoPosts bool
}

// GetAllCategoriesFromDb retrieves all categories from category table
func GetAllCategoriesFromDb() ([]Category, error) {
	rows, err := db.Query("SELECT id, text, url, created FROM category")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category

	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Text, &category.URL, &category.Created); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

// RegisterPostCategoriestoDb iterates over each category name, gets the id of the category,
// Prepares a SQL statment to insert a new row into post_category table and executes the statment
// Data will be stored in post_category table
func RegisterPostCategoriesToDb(post_id int, categoryNames []string) error {
	for _, categoryName := range categoryNames {
		category_id := GetCategoryID(categoryName)
		statement, err := db.Prepare("INSERT INTO post_category(post_id, category_id) VALUES(?, ?)")
		if err != nil {
			log.Printf("Error preparing data: %v", err)
			return err
		}
		defer statement.Close()
		_, err = statement.Exec(post_id, category_id)
		if err != nil {
			log.Printf("Error executing data: %v", err)
			return err
		}
		fmt.Println("Inserted data into database:", post_id, category_id)
	}
	return nil
}

// getCategoryNamesForPost takes all categories by id from category table
func GetCategoryNamesForPost(category_ids []int) ([]string, error) {
	var categoryNames []string
	for _, category_id := range category_ids {
		rows, err := db.Query("SELECT text FROM category WHERE id = ?", category_id)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var categoryName string
			if err := rows.Scan(&categoryName); err != nil {
				return nil, err
			}
			categoryNames = append(categoryNames, categoryName)
		}
	}
	return categoryNames, nil
}

// GetAllCategoryIDsForPost is needed to get all category_ids from post category table
func GetAllCategoryIDsForPost(post_id int) ([]int, error) {
	rows, err := db.Query("SELECT category_id FROM post_category WHERE post_id = ?", post_id)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var category_ids []int

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		category_ids = append(category_ids, id)
	}
	//if any error occured during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return category_ids, nil
}

// GetCategoryID gets id of category with the given name.
func GetCategoryID(categoryName string) int {
	category_id := 0
	err := db.QueryRow("SELECT id FROM category WHERE text = ?", categoryName).Scan(&category_id)
	if err != nil {
		fmt.Println("No previous reaction, proceeding...")
	}
	return category_id
}

// DoesCategoryExist finds catogory from db and returns bool
func DoesCategoryExist(categoryURL string) bool {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM category WHERE url = ?)", categoryURL).Scan(&exists)
	if err != nil {
		fmt.Println("Didnt find category from database.")
		return false
	}
	return exists
}

// GetCurrentCategory function retrieves a category from a database based on its URL.
func GetCurrentCategory(categoryURL string) (Category, error) {
	var currentCategory Category
	err := db.QueryRow("SELECT id, text, url, created FROM category WHERE url = ?", categoryURL).Scan(&currentCategory.ID, &currentCategory.Text, &currentCategory.URL, &currentCategory.Created)
	if err != nil {
		return Category{}, err
	}
	return currentCategory, nil
}

// GetAllPostIDSByCategory retrieves all post id-s associated with specific category
// from post category table
func GetAllPostIDsByCategory(category_id int) ([]int, error) {
	rows, err := db.Query("SELECT post_id FROM post_category WHERE category_id = ?", category_id)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var category_ids []int

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		category_ids = append(category_ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return category_ids, nil
}

// GetAllPostsByPostIDs retrieves all posts associated with a list of post IDS from post table
func GetAllPostsByPostIDs(post_ids []int) ([]Post, error) {
	var posts []Post

	for _, post_id := range post_ids {
		rows, err := db.Query("SELECT id, user_id, postTitle, postBody, created, like_count, dislike_count, comment_count FROM post WHERE id = ?", post_id)
		if err != nil {
			return []Post{}, err
		}
		defer rows.Close()

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
	}
	return posts, nil
}
