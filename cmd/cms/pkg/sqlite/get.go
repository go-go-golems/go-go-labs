package sqlite

import (
	"context"
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/cms/pkg"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"strings"
	"text/template"
)

type QueryData struct {
	Fields    []pkg.Field
	TableName string
	IdField   string
}

// SQLite query template for getting all objects from table.
const getObjectTpl = `SELECT id,
{{range $i, $field := .Fields}}{{if $i}}, {{end}}{{.Name}}{{end}}
FROM {{.TableName}} WHERE {{.IdField}} = ?`

var getObjectParsedTemplate *template.Template

func generateSQLiteGet(queryData QueryData) (string, error) {
	if getObjectParsedTemplate == nil {
		tmpl, err := template.New("query").Parse(getObjectTpl)
		if err != nil {
			return "", fmt.Errorf("parse template: %w", err)
		}
		getObjectParsedTemplate = tmpl
	}

	queryBuilder := &strings.Builder{}
	err := getObjectParsedTemplate.Execute(queryBuilder, queryData)
	if err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	query := queryBuilder.String()
	return query, nil
}

// GenerateSQLiteGetObjectMainTable generates SQLite query for getting the object with id from table.
func GenerateSQLiteGetObjectMainTable(schema *pkg.Schema) (string, error) {
	mainTable, ok := schema.Tables[schema.MainTable]
	if !ok {
		return "", fmt.Errorf("main table %q not found in schema", schema.MainTable)
	}

	queryData := QueryData{
		Fields:    mainTable.Fields,
		TableName: schema.MainTable,
		IdField:   "id",
	}

	return generateSQLiteGet(queryData)
}

func GenerateSQLiteGetObjectSecondaryTable(schema *pkg.Schema, name string) (string, error) {
	if name == schema.MainTable {
		return "", errors.New("main table is not a secondary table")
	}

	table := schema.Tables[name]
	fields := make([]pkg.Field, 0)
	if table.IsList {
		if table.ValueField == nil {
			return "", errors.New("value field not found")
		}
		fields = append(fields, *table.ValueField)
	} else {
		fields = append(fields, table.Fields...)
	}

	queryData := QueryData{
		Fields:    fields,
		TableName: name,
		IdField:   "parent_id",
	}

	return generateSQLiteGet(queryData)
}

func GetObject(
	ctx context.Context,
	db *sqlx.DB,
	schema *pkg.Schema,
	id int64,
) (map[string]interface{}, error) {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}

	defer func(tx *sqlx.Tx) {
		// only a read transaction, so we can rollback
		_ = tx.Rollback()
	}(tx)

	query, err := GenerateSQLiteGetObjectMainTable(schema)
	if err != nil {
		return nil, fmt.Errorf("generate query: %w", err)
	}

	rows, err := tx.QueryxContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer func(rows *sqlx.Rows) {
		_ = rows.Close()
	}(rows)

	results := make(map[string]interface{})
	for rows.Next() {
		result := map[string]interface{}{}
		err := rows.MapScan(result)
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

		query, err = GenerateSQLiteGetObjectSecondaryTable(schema, tableName)
		if err != nil {
			return nil, fmt.Errorf("generate query: %w", err)
		}

		rows, err := tx.QueryxContext(ctx, query, id)
		if err != nil {
			return nil, fmt.Errorf("query: %w", err)
		}
		defer func(rows *sqlx.Rows) {
			_ = rows.Close()
		}(rows)

		if table.IsList {
			additionalResults := []interface{}{}
			for rows.Next() {
				results := map[string]interface{}{}
				err := rows.MapScan(results)
				if err != nil {
					return nil, fmt.Errorf("scan: %w", err)
				}
				additionalResults = append(additionalResults, results[table.ValueField.Name])
			}
			if err := rows.Err(); err != nil {
				return nil, fmt.Errorf("rows: %w", err)
			}

			results[tableName] = additionalResults
		} else {
			var additionalResults []map[string]interface{}
			for rows.Next() {
				result := map[string]interface{}{}
				err := rows.MapScan(result)
				if err != nil {
					return nil, fmt.Errorf("scan: %w", err)
				}
				delete(result, "id")
				delete(result, "parent_id")
				additionalResults = append(additionalResults, result)
			}
			if err := rows.Err(); err != nil {
				return nil, fmt.Errorf("rows: %w", err)
			}

			results[tableName] = additionalResults
		}
	}

	return results, nil
}
