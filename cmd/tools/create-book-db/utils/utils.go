package utils

import (
	"database/sql"
	"fmt"
)

func LoadStringsFromTable(db *sql.DB, table string, idFieldName string, fieldName string, chapterID uint64) ([]string, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", fieldName, table, idFieldName)
	rows, err := db.Query(query, chapterID)

	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	results := make([]string, 0)

	for rows.Next() {
		var result string
		err := rows.Scan(&result)
		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}
