package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/cms/pkg"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"strings"
	"text/template"
)

type QueryData struct {
	Fields    []pkg.Field
	TableName string
	IdField   string
}

// SQLite query template for getting all objects from table.
const getObjectTpl = `SELECT id,{{range $i, $field := .Fields}}{{if $i}}, {{end}}{{.Name}}{{end}} FROM {{.TableName}} WHERE {{.IdField}} = ?`

// GenerateSQLiteGetObject generates SQLite query for getting the object with id from table.
func GenerateSQLiteGetObject(
	ctx context.Context,
	db *sqlx.DB,
	schema *pkg.Schema,
	id int,
) (map[string]interface{}, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			err = tx.Rollback() //
			log.Error().Err(err).Msg("rollback failed")
		} else if err != nil {
			err = tx.Rollback()
			if err != nil {
				log.Error().Err(err).Msg("rollback failed")
			} // err is non-nil; don't change it
		} else {
			err = tx.Commit() // err is nil; if Commit returns error update err
			if err != nil {
				log.Error().Err(err).Msg("commit failed")
			}
		}
	}()

	tmpl, err := template.New("query").Parse(getObjectTpl)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	mainTable, ok := schema.Tables[schema.MainTable]
	if !ok {
		return nil, fmt.Errorf("main table %q not found in schema", schema.MainTable)
	}

	queryData := QueryData{
		Fields:    mainTable.Fields,
		TableName: schema.MainTable,
		IdField:   "id",
	}

	queryBuilder := &strings.Builder{}
	err = tmpl.Execute(queryBuilder, queryData)
	if err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	query := queryBuilder.String()
	rows, err := tx.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	results := make(map[string]interface{})
	for rows.Next() {
		var result map[string]interface{}
		err := rows.Scan(&result)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		for k, v := range result {
			results[k] = v
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	// Query additional tables
	for tableName, table := range schema.Tables {
		if tableName == schema.MainTable {
			continue
		}

		queryData := QueryData{
			Fields:    table.Fields,
			TableName: tableName,
			IdField:   "parent_id",
		}

		queryBuilder := &strings.Builder{}
		err = tmpl.Execute(queryBuilder, queryData)
		if err != nil {
			return nil, fmt.Errorf("execute template: %w", err)
		}

		query := queryBuilder.String()
		rows, err := tx.QueryContext(ctx, query, id)
		if err != nil {
			return nil, fmt.Errorf("query: %w", err)
		}
		defer func(rows *sql.Rows) {
			_ = rows.Close()
		}(rows)

		var additionalResults []map[string]interface{}
		for rows.Next() {
			var result map[string]interface{}
			err := rows.Scan(result)
			if err != nil {
				return nil, fmt.Errorf("scan: %w", err)
			}
			additionalResults = append(additionalResults, result)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("rows: %w", err)
		}

		results[tableName] = additionalResults
	}

	return results, nil
}
