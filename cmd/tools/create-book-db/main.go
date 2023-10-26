package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"time"
)

type Chapter struct {
	Filename         string    `json:"filename"`
	Title            string    `json:"title"`
	Toc              []TocItem `json:"toc"`
	KeyPoints        []string  `json:"keyPoints"`
	SummaryParagraph string    `json:"summaryParagraph"`
	Keywords         []string  `json:"keywords"`
	Sources          []Source  `json:"references"`
	CreatedAt        string    `json:"createdAt"`
	Comment          string    `json:"comment"`
}

type TocItem struct {
	Name  string `json:"name"`
	Level int    `json:"level"`
}

type Source struct {
	Author string `json:"author"`
	Source string `json:"source"`
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./chapter.db")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func createTables(db *sql.DB) error {
	chapterTable := `
    CREATE TABLE IF NOT EXISTS Chapters (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        filename TEXT NOT NULL,
        created_at TEXT,
        comment TEXT,
        summaryParagraph TEXT
    );`

	tocTable := `
    CREATE TABLE IF NOT EXISTS TocItems (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        chapterID INTEGER,
        name TEXT,
        level INTEGER,
        FOREIGN KEY(chapterID) REFERENCES Chapters(id)
    );`

	keyPointsTable := `
    CREATE TABLE IF NOT EXISTS KeyPoints (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        chapterID INTEGER,
        point TEXT,
        FOREIGN KEY(chapterID) REFERENCES Chapters(id)
    );`

	keywordsTable := `
    CREATE TABLE IF NOT EXISTS Keywords (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        chapterID INTEGER,
        keyword TEXT,
        FOREIGN KEY(chapterID) REFERENCES Chapters(id)
    );`

	referencesTable := `
    CREATE TABLE IF NOT EXISTS Sources (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        chapterID INTEGER,
        author TEXT,
        source TEXT,
        FOREIGN KEY(chapterID) REFERENCES Chapters(id)
    );`

	tables := []string{chapterTable, tocTable, keyPointsTable, keywordsTable, referencesTable}

	for _, table := range tables {
		_, err := db.Exec(table)
		if err != nil {
			return err
		}
	}

	return nil
}
func readJSONFile(filename string) (*Chapter, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var chapter Chapter
	if err := json.Unmarshal(file, &chapter); err != nil {
		return nil, err
	}

	chapter.Filename = filename
	chapter.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
	chapter.Comment = ""

	return &chapter, nil
}

func insertChapter(db *sql.DB, chapter *Chapter) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO Chapters (title, summaryParagraph, filename, created_at, comment) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	res, err := stmt.Exec(chapter.Title, chapter.SummaryParagraph, chapter.Filename, chapter.CreatedAt, chapter.Comment)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	return id, err
}

func insertTocItem(db *sql.DB, chapterID int64, toc TocItem) error {
	_, err := db.Exec("INSERT INTO TocItems (chapterID, name, level) VALUES (?, ?, ?)", chapterID, toc.Name, toc.Level)
	return err
}

func insertKeyPoint(db *sql.DB, chapterID int64, keyPoint string) error {
	_, err := db.Exec("INSERT INTO KeyPoints (chapterID, point) VALUES (?, ?)", chapterID, keyPoint)
	return err
}

func insertKeyword(db *sql.DB, chapterID int64, keyword string) error {
	_, err := db.Exec("INSERT INTO Keywords (chapterID, keyword) VALUES (?, ?)", chapterID, keyword)
	return err
}

func insertSource(db *sql.DB, chapterID int64, reference Source) error {
	_, err := db.Exec("INSERT INTO Sources (chapterID, author, source) VALUES (?, ?, ?)", chapterID, reference.Author, reference.Source)
	return err
}

func main() {
	db, err := initDB()
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}

	if err := createTables(db); err != nil {
		fmt.Println("Error creating tables:", err)
		return
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: program <json file path>")
		return
	}

	for _, file := range os.Args[1:] {
		chapter, err := readJSONFile(file)
		if err != nil {
			fmt.Println("Error reading JSON:", err)
			return
		}

		chapterID, err := insertChapter(db, chapter)
		if err != nil {
			fmt.Println("Error inserting chapter:", err)
			continue
		}

		for _, toc := range chapter.Toc {
			err := insertTocItem(db, chapterID, toc)
			if err != nil {
				fmt.Println("Error inserting TOC item:", err)
				return
			}
		}

		for _, keyPoint := range chapter.KeyPoints {
			err := insertKeyPoint(db, chapterID, keyPoint)
			if err != nil {
				fmt.Println("Error inserting key point:", err)
				return
			}
		}

		for _, keyword := range chapter.Keywords {
			err := insertKeyword(db, chapterID, keyword)
			if err != nil {
				fmt.Println("Error inserting keyword:", err)
				return
			}
		}

		for _, sources := range chapter.Sources {
			err := insertSource(db, chapterID, sources)
			if err != nil {
				fmt.Println("Error inserting sources:", err)
				return
			}
		}

	}
}
