package main

import (
	"database/sql"
	"embed"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io/fs"
	"strings"
)

//go:embed data.sql
var embeddedSQL embed.FS

func main() {
	// Create in-memory SQLite database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		fmt.Println("Error while creating in-memory database:", err)
		return
	}
	defer db.Close()

	// Read the embedded SQL content
	sqlContent, err := fs.ReadFile(embeddedSQL, "data.sql")
	if err != nil {
		fmt.Println("Error reading embedded SQL:", err)
		return
	}

	// Execute the SQL commands one by one
	queries := strings.Split(string(sqlContent), ";")
	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query != "" {
			_, err := db.Exec(query)
			if err != nil {
				fmt.Printf("Error executing query %s: %v\n", query, err)
				return
			}
		}
	}

	fmt.Println("Database filled successfully!")

	// Sample query: List all books and their authors
	rows, err := db.Query("SELECT b.title, a.name FROM books b INNER JOIN authors a ON b.author_id = a.author_id")
	if err != nil {
		fmt.Println("Error executing sample query:", err)
		return
	}
	defer rows.Close()

	fmt.Println("Books and their Authors:")
	for rows.Next() {
		var title, authorName string
		if err := rows.Scan(&title, &authorName); err != nil {
			fmt.Println("Error scanning row:", err)
			return
		}
		fmt.Printf("- %s by %s\n", title, authorName)
	}

	if err := rows.Err(); err != nil {
		fmt.Println("Error with rows:", err)
		return
	}
}
