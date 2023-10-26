package questions

import (
	"database/sql"
	"encoding/json"
	"os"
	"time"
)

// Additional struct for Questions
type Question struct {
	ChapterName             string   `json:"chapterName"`
	RelevancyScore          int      `json:"relevancyScore"`
	RelevantSections        []string `json:"relevantSections"`
	RelevantKeywords        []string `json:"relevantKeywords"`
	RelevantKeyPoints       []string `json:"relevantKeyPoints"`
	RelevantReferences      []string `json:"relevantReferences"`
	ExplanationForRelevance string   `json:"explanationForRelevance"`
	RecommendationsToReader string   `json:"recommendationsToReader"`
	ChapterID               int64    `json:"chapterId"`
	CreatedAt               string   `json:"createdAt"`
	Comment                 string   `json:"comment"`
	Filename                string   `json:"filename"`
}

// Additional tables creation SQL strings
var questionTable = `
CREATE TABLE IF NOT EXISTS Questions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chapterID INTEGER,
    chapterName TEXT NOT NULL,
    relevancyScore INTEGER,
    explanationForRelevance TEXT,
    recommendationsToReader TEXT,
    created_at TEXT,
    comment TEXT,
    filename TEXT,
    FOREIGN KEY(chapterID) REFERENCES Chapters(id)
);`

var relevantSectionsTable = `
CREATE TABLE IF NOT EXISTS RelevantSections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    questionID INTEGER,
    section TEXT,
    FOREIGN KEY(questionID) REFERENCES Questions(id)
);`

var relevantKeywordsTable = `
CREATE TABLE IF NOT EXISTS RelevantKeywords (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    questionID INTEGER,
    keyword TEXT,
    FOREIGN KEY(questionID) REFERENCES Questions(id)
);`

var relevantKeyPointsTable = `
CREATE TABLE IF NOT EXISTS RelevantKeyPoints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    questionID INTEGER,
    keyPoint TEXT,
    FOREIGN KEY(questionID) REFERENCES Questions(id)
);`

var relevantReferencesTable = `
CREATE TABLE IF NOT EXISTS RelevantReferences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    questionID INTEGER,
    reference TEXT,
    FOREIGN KEY(questionID) REFERENCES Questions(id)
);`

func CreateTables(db *sql.DB) error {
	newTables := []string{questionTable, relevantSectionsTable, relevantKeywordsTable, relevantKeyPointsTable, relevantReferencesTable}

	for _, table := range newTables {
		_, err := db.Exec(table)
		if err != nil {
			return err
		}
	}

	return nil
}

func ReadQuestionFromFile(filename string) (*Question, error) {
	var q Question

	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(file, &q); err != nil {
		return nil, err
	}

	q.Filename = filename
	q.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
	q.Comment = ""

	return &q, nil
}

func createQuestionsTable(db *sql.DB) error {
	questionsTable := `
    CREATE TABLE IF NOT EXISTS Questions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        chapterID INTEGER,
        chapterName TEXT NOT NULL,
        relevancyScore INTEGER,
        explanationForRelevance TEXT,
        recommendationsToReader TEXT,
        created_at TEXT,
        comment TEXT,
        filename TEXT,
        FOREIGN KEY(chapterID) REFERENCES Chapters(id)
    );`
	_, err := db.Exec(questionsTable)
	return err
}

// Inserting data into Questions table
func InsertQuestion(db *sql.DB, q *Question) (int64, error) {
	// Lookup chapterID by chapterName
	var chapterID int64
	err := db.QueryRow("SELECT id FROM Chapters WHERE name = ?", q.ChapterName).Scan(&chapterID)
	if err != nil {
		return 0, err
	}

	// Insert the main question record
	stmt, err := db.Prepare("INSERT INTO Questions(chapterID, chapterName, relevancyScore, explanationForRelevance, recommendationsToReader, created_at, comment, filename) VALUES(?,?,?,?,?,?,?,?)")
	if err != nil {
		return 0, err
	}
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)

	res, err := stmt.Exec(chapterID, q.ChapterName, q.RelevancyScore, q.ExplanationForRelevance, q.RecommendationsToReader, q.CreatedAt, q.Comment, q.Filename)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert relevant sections
	for _, section := range q.RelevantSections {
		if err := insertRelevantSection(db, section, lastID); err != nil {
			return 0, err
		}
	}

	// Insert relevant keywords
	for _, keyword := range q.RelevantKeywords {
		if err := insertRelevantKeyword(db, keyword, lastID); err != nil {
			return 0, err
		}
	}

	// Insert relevant key points
	for _, keyPoint := range q.RelevantKeyPoints {
		if err := insertRelevantKeyPoint(db, keyPoint, lastID); err != nil {
			return 0, err
		}
	}

	// Insert relevant references
	for _, reference := range q.RelevantReferences {
		if err := insertRelevantReference(db, reference, lastID); err != nil {
			return 0, err
		}
	}

	return lastID, nil
}

// Inserting data into RelevantSections table
func insertRelevantSection(db *sql.DB, section string, questionID int64) error {
	_, err := db.Exec("INSERT INTO RelevantSections(questionID, section) VALUES(?,?)", questionID, section)
	return err
}

// Inserting data into RelevantKeywords table
func insertRelevantKeyword(db *sql.DB, keyword string, questionID int64) error {
	_, err := db.Exec("INSERT INTO RelevantKeywords(questionID, keyword) VALUES(?,?)", questionID, keyword)
	return err
}

// Inserting data into RelevantKeyPoints table
func insertRelevantKeyPoint(db *sql.DB, keyPoint string, questionID int64) error {
	_, err := db.Exec("INSERT INTO RelevantKeyPoints(questionID, keyPoint) VALUES(?,?)", questionID, keyPoint)
	return err
}

// Inserting data into RelevantReferences table
func insertRelevantReference(db *sql.DB, reference string, questionID int64) error {
	_, err := db.Exec("INSERT INTO RelevantReferences(questionID, reference) VALUES(?,?)", questionID, reference)
	return err
}
