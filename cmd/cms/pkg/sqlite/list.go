package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/cms/pkg"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"strings"
	"text/template"
)

type PaginationRequest struct {
	Offset int
	Limit  int
}

type PaginationResponse struct {
	Total  int
	Data   []map[string]interface{}
	Offset int
}

const listObjectsTpl = `SELECT id,{{range $i, $field := .Fields}}{{if $i}}, {{end}}{{.Name}}{{end}} FROM {{.TableName}} LIMIT ? OFFSET ?`

var listObjectsParsedTemplate *template.Template

func generateSQLiteList(queryData QueryData, limit, offset int) (string, []interface{}, error) {
	if listObjectsParsedTemplate == nil {
		tmpl, err := template.New("query").Parse(listObjectsTpl)
		if err != nil {
			return "", nil, fmt.Errorf("parse template: %w", err)
		}
		listObjectsParsedTemplate = tmpl
	}

	queryBuilder := &strings.Builder{}
	err := listObjectsParsedTemplate.Execute(queryBuilder, queryData)
	if err != nil {
		return "", nil, fmt.Errorf("execute template: %w", err)
	}

	query := queryBuilder.String()
	return query, []interface{}{limit, offset}, nil
}

func GenerateSQLiteListObjectsMainTable(schema *pkg.Schema, limit, offset int) (string, []interface{}, error) {
	mainTable, ok := schema.Tables[schema.MainTable]
	if !ok {
		return "", nil, fmt.Errorf("main table %q not found in schema", schema.MainTable)
	}

	queryData := QueryData{
		Fields:    mainTable.Fields,
		TableName: schema.MainTable,
	}

	return generateSQLiteList(queryData, limit, offset)
}

const listObjectsSecondaryTpl = `SELECT parent_id,{{range $i, $field := .Fields}}{{if $i}}, {{end}}{{.Name}}{{end}} FROM {{.TableName}} WHERE parent_id IN (?)`

var listObjectsSecondaryParsedTemplate *template.Template

func generateSQLiteListSecondary(queryData QueryData, ids []int) (string, []interface{}, error) {
	if listObjectsSecondaryParsedTemplate == nil {
		tmpl, err := template.New("query").Parse(listObjectsSecondaryTpl)
		if err != nil {
			return "", nil, fmt.Errorf("parse template: %w", err)
		}
		listObjectsSecondaryParsedTemplate = tmpl
	}

	queryBuilder := &strings.Builder{}
	err := listObjectsSecondaryParsedTemplate.Execute(queryBuilder, queryData)
	if err != nil {
		return "", nil, fmt.Errorf("execute template: %w", err)
	}

	query := queryBuilder.String()
	return query, []interface{}{ids}, nil
}

func GenerateSQLiteListObjectsSecondaryTable(schema *pkg.Schema, name string, ids []int) (string, []interface{}, error) {
	if name == schema.MainTable {
		return "", nil, errors.New("main table is not a secondary table")
	}

	table := schema.Tables[name]

	queryData := QueryData{
		Fields:    table.Fields,
		TableName: name,
	}

	return generateSQLiteListSecondary(queryData, ids)
}

func ListObjects(
	ctx context.Context,
	db *sqlx.DB,
	schema *pkg.Schema,
	paginationRequest *PaginationRequest,
) (*PaginationResponse, error) {
	query, args, err := GenerateSQLiteListObjectsMainTable(schema, paginationRequest.Limit, paginationRequest.Offset)
	if err != nil {
		return nil, fmt.Errorf("generate query: %w", err)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var ids []int
	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int
		result := make(map[string]interface{})
		err := rows.Scan(&id, &result)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		ids = append(ids, id)
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	for tableName := range schema.Tables {
		if tableName == schema.MainTable {
			continue
		}

		query, args, err = GenerateSQLiteListObjectsSecondaryTable(schema, tableName, ids)
		if err != nil {
			return nil, fmt.Errorf("generate query: %w", err)
		}

		rows, err := db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("query: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var parentId int
			result := make(map[string]interface{})
			err := rows.Scan(&parentId, &result)
			if err != nil {
				return nil, fmt.Errorf("scan: %w", err)
			}

			for i, r := range results {
				if r["id"] == parentId {
					results[i][tableName] = result
					break
				}
			}
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("rows: %w", err)
		}
	}

	return &PaginationResponse{
		Total:  len(results),
		Data:   results,
		Offset: paginationRequest.Offset,
	}, nil
}
