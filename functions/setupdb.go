package functions

import (
	"database/sql"
	"fmt"

	// Import the SQLite3 driver
	_ "github.com/mattn/go-sqlite3"
)

func SetUpDatabase(db *sql.DB) error {

	templates := CreateTemplatesArray()
	for i, template := range templates {
		statement, err := db.Prepare(template)
		if err != nil {
			return err
		}
		defer statement.Close()

		_, err = statement.Exec()
		if err != nil {
			return err
		}
		fmt.Println("Succesfully create table if not exists index: ", i)
	}

	return nil // Ensure there's a return at the end of the function
}

func CreateTemplatesArray() []string {
	var templates []string

	// Define all your templates outside of the loop
	templates = append(templates, `CREATE TABLE IF NOT EXISTS category (
		"id"	INTEGER NOT NULL UNIQUE,
		"text"	TEXT NOT NULL,
		"url"	TEXT NOT NULL,
		"created"	TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY("id" AUTOINCREMENT)
	);`)

	templates = append(templates, `CREATE TABLE IF NOT EXISTS comment (
		"id"	INTEGER NOT NULL UNIQUE,
		"post_id"	INTEGER NOT NULL,
		"user_id"	INTEGER NOT NULL,
		"username"	TEXT NOT NULL,
		"text"	TEXT NOT NULL,
		"like_count"	INTEGER DEFAULT 0,
		"dislike_count"	INTEGER DEFAULT 0,
		"created"	TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY("user_id") REFERENCES "user"("id"),
		PRIMARY KEY("id" AUTOINCREMENT)
		);`)

	templates = append(templates, `CREATE TABLE IF NOT EXISTS post (
		"id"	INTEGER NOT NULL UNIQUE,
		"user_id"	INTEGER NOT NULL,
		"postTitle"	TEXT NOT NULL,
		"postBody"	TEXT NOT NULL,
		"created"	NUMERIC NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"like_count"	INTEGER DEFAULT 0,
		"dislike_count"	INTEGER DEFAULT 0,
		"comment_count"	INTEGER DEFAULT 0,
		"username"	TEXT NOT NULL,
		FOREIGN KEY("user_id") REFERENCES "user"("id"),
		PRIMARY KEY("id" AUTOINCREMENT)
		);`)

	templates = append(templates, `CREATE TABLE IF NOT EXISTS post_category (
		"id"	INTEGER NOT NULL UNIQUE,
		"category_id"	INTEGER NOT NULL,
		"post_id"	INTEGER NOT NULL,
		"created"	TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY("category_id") REFERENCES "category"("id"),
		FOREIGN KEY("post_id") REFERENCES "post"("id"),
		PRIMARY KEY("id" AUTOINCREMENT)
	);`)

	templates = append(templates, `CREATE TABLE IF NOT EXISTS reaction (
		"id"	INTEGER NOT NULL UNIQUE,
		"post_id"	INTEGER NOT NULL,
		"user_id"	INTEGER NOT NULL,
		"comment_id"	INTEGER NOT NULL,
		"reaction_bool"	INTEGER,
		"created"	TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY("user_id") REFERENCES "user"("id"),
		FOREIGN KEY("post_id") REFERENCES "post"("id"),
		FOREIGN KEY("comment_id") REFERENCES "comment"("id"),
		PRIMARY KEY("id" AUTOINCREMENT)
		);`)

	templates = append(templates, `CREATE TABLE IF NOT EXISTS session (
		"session_id"	INTEGER NOT NULL,
		"user_id"	INTEGER NOT NULL,
		"email"	TEXT NOT NULL UNIQUE,
		CONSTRAINT "id" FOREIGN KEY("user_id") REFERENCES "user"("id")
	);`)

	templates = append(templates, ` CREATE TABLE IF NOT EXISTS user (
		"id"	INTEGER NOT NULL UNIQUE,
		"username"	TEXT NOT NULL UNIQUE,
		"firstname"	TEXT NOT NULL,
		"lastname"	TEXT NOT NULL,
		"password"	TEXT NOT NULL,
		"email"	TEXT NOT NULL UNIQUE,
		"created"	TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY("id" AUTOINCREMENT)
	);`)

	return templates
}
