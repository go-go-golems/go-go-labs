package questions

import (
	"database/sql"
	"github.com/go-go-golems/go-go-labs/cmd/tools/create-book-db/utils"
	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

func GetAllQuestions(db *sql.DB) ([]*Question, error) {
	rows, err := db.Query(`SELECT id, chapterID, chapterName, relevancyScore, explanationForRelevance, 
                                  recommendationsToReader, created_at, comment, filename 
                           FROM Questions`)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var questions []*Question
	for rows.Next() {
		q := &Question{}
		if err := rows.Scan(&q.ChapterID, &q.ChapterName, &q.RelevancyScore, &q.ExplanationForRelevance,
			&q.RecommendationsToReader, &q.CreatedAt, &q.Comment, &q.Filename); err != nil {
			return nil, err
		}

		err = loadQuestionFields(db, q)
		if err != nil {
			return nil, err
		}

		questions = append(questions, q)
	}
	return questions, nil
}

func GetQuestionByID(db *sql.DB, id int64) (*Question, error) {
	q := &Question{}
	err := db.QueryRow(`SELECT id, chapterID, chapterName, relevancyScore, explanationForRelevance, 
                               recommendationsToReader, created_at, comment, filename 
                        FROM Questions WHERE id = ?`, id).Scan(&q.ChapterID, &q.ChapterName, &q.RelevancyScore,
		&q.ExplanationForRelevance, &q.RecommendationsToReader,
		&q.CreatedAt, &q.Comment, &q.Filename)
	if err != nil {
		return nil, err
	}

	err = loadQuestionFields(db, q)
	if err != nil {
		return nil, err
	}

	return q, nil
}

func loadQuestionFields(db *sql.DB, q *Question) error {
	var err error
	q.RelevantSections, err = utils.LoadStringsFromTable(db, "RelevantSections", "questionID", "section", uint64(q.ChapterID))
	if err != nil {
		return err
	}

	q.RelevantKeywords, err = utils.LoadStringsFromTable(db, "RelevantKeywords", "questionID", "keyword", uint64(q.ChapterID))
	if err != nil {
		return err
	}

	q.RelevantKeyPoints, err = utils.LoadStringsFromTable(db, "RelevantKeyPoints", "questionID", "keyPoint", uint64(q.ChapterID))
	if err != nil {
		return err
	}

	q.RelevantReferences, err = utils.LoadStringsFromTable(db, "RelevantReferences", "questionID", "reference", uint64(q.ChapterID))
	if err != nil {
		return err
	}

	return nil
}
