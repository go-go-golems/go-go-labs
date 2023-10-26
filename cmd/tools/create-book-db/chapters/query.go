package chapters

import (
	"database/sql"
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/tools/create-book-db/utils"
)

func loadTocItems(db *sql.DB, chapterID int64) ([]TocItem, error) {
	tocRows, err := db.Query("SELECT name, level FROM TocItems WHERE chapterID = ?", chapterID)
	if err != nil {
		return nil, err
	}
	defer func(tocRows *sql.Rows) {
		_ = tocRows.Close()
	}(tocRows)

	var tocItems []TocItem
	for tocRows.Next() {
		var item TocItem
		if err := tocRows.Scan(&item.Name, &item.Level); err != nil {
			return nil, err
		}
		tocItems = append(tocItems, item)
	}
	return tocItems, nil
}

func loadSources(db *sql.DB, chapterID int64) ([]Source, error) {
	sourceRows, err := db.Query("SELECT author, source FROM Sources WHERE chapterID = ?", chapterID)
	if err != nil {
		return nil, err
	}
	defer func(sourceRows *sql.Rows) {
		_ = sourceRows.Close()
	}(sourceRows)

	var sources []Source
	for sourceRows.Next() {
		var s Source
		if err := sourceRows.Scan(&s.Author, &s.Source); err != nil {
			return nil, err
		}
		sources = append(sources, s)
	}
	return sources, nil
}

func GetAllChapters(db *sql.DB) ([]*Chapter, error) {
	rows, err := db.Query("SELECT id, filename, title, created_at, comment,summaryParagraph FROM Chapters")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var chapters []*Chapter
	for rows.Next() {
		c := &Chapter{}
		if err := rows.Scan(&c.ID, &c.Filename, &c.Title, &c.CreatedAt, &c.Comment, &c.SummaryParagraph); err != nil {
			return nil, err
		}

		err = loadChapterFields(db, c)
		if err != nil {
			return nil, err
		}

		chapters = append(chapters, c)
	}
	return chapters, nil
}

func loadChapterFields(db *sql.DB, c *Chapter) error {
	var err error
	c.KeyPoints, err = utils.LoadStringsFromTable(db, "KeyPoints", "chapterID", "point", uint64(c.ID))
	if err != nil {
		return err
	}

	c.Keywords, err = utils.LoadStringsFromTable(db, "Keywords", "chapterID", "keyword", uint64(c.ID))
	if err != nil {
		return err
	}

	c.Toc, err = loadTocItems(db, c.ID)
	if err != nil {
		return err
	}

	c.Sources, err = loadSources(db, c.ID)
	if err != nil {
		return err
	}
	return nil
}

func GetChapterByID(db *sql.DB, id int64) (*Chapter, error) {
	c := &Chapter{}
	err := db.QueryRow("SELECT id, filename, title, created_at, comment, summaryParagraph FROM Chapters WHERE id = ?", id).Scan(&c.ID, &c.Filename, &c.Title, &c.CreatedAt, &c.Comment, &c.SummaryParagraph)
	if err != nil {
		return nil, err
	}

	err = loadChapterFields(db, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

type Query struct {
	Title     string   `json:"title"`
	Keywords  []string `json:"keywords"`
	TocItems  []string `json:"tocItems"`
	KeyPoints []string `json:"keyPoints"`
	Sources   []string `json:"sources"`
}

func FindChapterIDs(db *sql.DB, query Query) ([]int64, error) {
	var chapterIDs []int64

	// Build SQL query for Chapters based on title
	var titleQuery string
	if query.Title != "" {
		titleQuery = fmt.Sprintf("SELECT id FROM Chapters WHERE title LIKE '%%%s%%'", query.Title)
	}

	// Execute Title query
	if titleQuery != "" {
		rows, err := db.Query(titleQuery)
		if err != nil {
			return nil, err
		}
		defer func(rows *sql.Rows) {
			_ = rows.Close()
		}(rows)

		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				return nil, err
			}
			chapterIDs = append(chapterIDs, id)
		}
	}

	// For each Keywords, TocItems, KeyPoints, and Sources, build and execute a query to further filter chapter IDs
	subQueries := []struct {
		fieldName  string
		tableName  string
		queryItems []string
	}{
		{"keyword", "Keywords", query.Keywords},
		{"name", "TocItems", query.TocItems},
		{"point", "KeyPoints", query.KeyPoints},
		{"source", "Sources", query.Sources},
	}

	for _, sub := range subQueries {
		if len(sub.queryItems) == 0 {
			continue
		}
		for _, item := range sub.queryItems {
			itemQuery := fmt.Sprintf("SELECT chapterID FROM %s WHERE %s LIKE '%%%s%%'", sub.tableName, sub.fieldName, item)
			rows, err := db.Query(itemQuery)
			if err != nil {
				return nil, err
			}
			defer func(rows *sql.Rows) {
				_ = rows.Close()
			}(rows)

			var newChapterIDs []int64
			for rows.Next() {
				var id int64
				if err := rows.Scan(&id); err != nil {
					return nil, err
				}
				newChapterIDs = append(newChapterIDs, id)
			}

			// Intersect newChapterIDs with existing chapterIDs
			chapterIDs = union(chapterIDs, newChapterIDs)
		}
	}

	return chapterIDs, nil
}

func union(a, b []int64) []int64 {
	m := make(map[int64]bool)
	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		m[item] = true
	}

	var result []int64
	for item := range m {
		result = append(result, item)
	}

	return result
}
