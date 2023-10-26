package main

import (
	"database/sql"
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/tools/create-book-db/chapters"
	"github.com/go-go-golems/go-go-labs/cmd/tools/create-book-db/questions"
	"github.com/go-go-golems/go-go-labs/cmd/tools/create-book-db/web"
	"os"
	"strings"
)

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./chapter.db")
	if err != nil {
		return nil, err
	}
	return db, nil
}
func main() {
	db, err := initDB()
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}

	if err := chapters.CreateTables(db); err != nil {
		fmt.Println("Error creating tables:", err)
		return
	}

	if err = questions.CreateTables(db); err != nil {
		fmt.Println("Error creating questions tables:", err)
		return
	}

	if len(os.Args) == 1 {
		err := web.Serve(db)
		if err != nil {
			fmt.Println("Error serving web:", err)
		}
	}

	importFiles(db, os.Args[1:])
}

func importFiles(db *sql.DB, files []string) {
	for _, file := range files {
		fmt.Println("Processing file:", file)

		if strings.HasSuffix(file, ".md") {
			fmt.Println("Processing markdown file:", file)
			_, err := chapters.InsertChapterContent(db, file)
			if err != nil {
				fmt.Println("Error inserting chapter content:", err)
				continue
			}
			fmt.Println("Inserted chapter content")
			continue
		}

		chapter, err := chapters.ReadChapterFromFile(file)
		if err != nil {
			question, err := questions.ReadQuestionFromFile(file)
			if err != nil {
				fmt.Println("Error reading JSON:", err)
				return
			}

			_, err = questions.InsertQuestion(db, question)
			if err != nil {
				fmt.Println("Error inserting question:", err)
				continue
			}

			fmt.Printf("Inserted question %s\n", question.ExplanationForRelevance)
			continue
		}

		_, err = chapters.InsertChapter(db, chapter)
		if err != nil {
			fmt.Println("Error inserting chapter:", err)
			continue
		}

		fmt.Printf("Inserted chapter %s\n", chapter.Title)
	}
}
