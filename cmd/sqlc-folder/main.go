package main

import (
	"context"
	"database/sql"
	_ "embed"
	"github.com/go-go-golems/go-go-labs/cmd/sqlc-folder/tutorial"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

//go:embed schema.sql
var schema string

func main() {
	// Open the SQLite database
	db, err := sql.Open("sqlite3", "./tutorial.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	// Execute the schema.sql commands
	_, err = db.Exec(schema)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new instance of Queries
	q := tutorial.New(db)

	// Use the context package for cancellation
	ctx := context.Background()

	// Create a new author
	author, err := q.CreateAuthor(ctx, tutorial.CreateAuthorParams{
		Name: "John Doe",
		Bio:  sql.NullString{String: "Author's biography", Valid: true},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Print the ID of the new author
	log.Println("New author ID:", author.ID)

	// Update the author's name
	err = q.UpdateAuthor(ctx, tutorial.UpdateAuthorParams{
		ID:   author.ID,
		Name: "Jane Doe",
		Bio:  sql.NullString{String: "Updated biography", Valid: true},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Fetch the updated author
	updatedAuthor, err := q.GetAuthor(ctx, author.ID)
	if err != nil {
		log.Fatal(err)
	}

	// Print the updated author's name
	log.Println("Updated author name:", updatedAuthor.Name)

	author2, err := q.CreateAuthor(ctx, tutorial.CreateAuthorParams{
		Name: "Edgar Allan Poe",
		Bio:  sql.NullString{String: "Author's biography", Valid: true},
	})
	if err != nil {
		log.Fatal(err)
	}

	// print out author
	log.Println("Author:", author2)

	// Fetch all authors
	authors, err := q.ListAuthors(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Print the number of authors
	log.Println("Number of authors:", len(authors))
	// Print authors in a loop
	for _, author := range authors {
		log.Println("Author:", author)
	}

	// Delete the author
	err = q.DeleteAuthor(ctx, author.ID)
	if err != nil {
		log.Fatal(err)
	}

	// Verify the author was deleted
	_, err = q.GetAuthor(ctx, author.ID)
	if err != nil {
		log.Println("Author was successfully deleted")
	} else {
		log.Println("Failed to delete author")
	}
}
