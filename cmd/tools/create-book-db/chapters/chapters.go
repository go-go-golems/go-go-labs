package chapters

import (
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"os"
	"time"
)

type Chapter struct {
	ID               int64     `json:"id"`
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

func CreateTables(db *sql.DB) error {
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
func ReadChapterFromFile(filename string) (*Chapter, error) {
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

func InsertChapter(db *sql.DB, chapter *Chapter) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO Chapters (title, summaryParagraph, filename, created_at, comment) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	res, err := stmt.Exec(chapter.Title, chapter.SummaryParagraph, chapter.Filename, chapter.CreatedAt, chapter.Comment)
	if err != nil {
		return 0, err
	}

	chapterID, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	for _, toc := range chapter.Toc {
		err := insertTocItem(db, chapterID, toc)
		if err != nil {
			return -1, errors.Wrapf(err, "Error inserting toc item %v", toc)
		}
	}

	for _, keyPoint := range chapter.KeyPoints {
		err := insertKeyPoint(db, chapterID, keyPoint)
		if err != nil {
			return -1, errors.Wrapf(err, "Error inserting key point %s", keyPoint)
		}
	}

	for _, keyword := range chapter.Keywords {
		err := insertKeyword(db, chapterID, keyword)
		if err != nil {
			return -1, errors.Wrapf(err, "Error inserting keyword %s", keyword)
		}
	}

	for _, sources := range chapter.Sources {
		err := insertSource(db, chapterID, sources)
		if err != nil {
			return -1, errors.Wrapf(err, "Error inserting source %s", sources)
		}
	}

	return chapterID, nil
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

func InsertChapterContent(db *sql.DB, filename string) (int64, error) {
	// Reading markdown file
	contentBytes, err := os.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	content := string(contentBytes)

	// Inserting into database
	stmt, err := db.Prepare("INSERT INTO ChapterContent(content, filename) VALUES(?, ?)")
	if err != nil {
		return 0, err
	}
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)

	res, err := stmt.Exec(content, filename)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastID, nil
}
