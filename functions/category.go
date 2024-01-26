package functions

import "fmt"

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
