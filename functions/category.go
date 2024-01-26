package functions

type Category struct {
	ID      string
	Text    string
	Created string
}

func GetCategoriesFromDb() ([]Category, error) {
	rows, err := db.Query("SELECT id, text, created FROM category")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category

	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Text, &category.Created); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}
