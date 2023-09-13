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

func generateSQLiteListSecondary(queryData QueryData, ids []int64) (string, []interface{}, error) {
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

func GenerateSQLiteListObjectsSecondaryTable(schema *pkg.Schema, name string, ids []int64) (
	string,
	[]interface{},
	error,
) {
	if name == schema.MainTable {
		return "", nil, errors.New("main table is not a secondary table")
	}

	fields := make([]pkg.Field, 0)

	table := schema.Tables[name]

	if table.IsList {
		if table.ValueField == nil {
			return "", nil, errors.New("value field not found")
		}
		fields = append(fields, *table.ValueField)
	} else {
		fields = append(fields, table.Fields...)
	}

	queryData := QueryData{
		Fields:    fields,
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
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}

	defer func(tx *sqlx.Tx) {
		// only a read transaction, so rollback
		_ = tx.Rollback()
	}(tx)

	query, args, err := GenerateSQLiteListObjectsMainTable(schema, paginationRequest.Limit, paginationRequest.Offset)
	if err != nil {
		return nil, fmt.Errorf("generate query: %w", err)
	}

	rows, err := tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer func(rows *sqlx.Rows) {
		_ = rows.Close()
	}(rows)

	var ids []int64
	results := make(map[int64]map[string]interface{}, 0)
	for rows.Next() {
		var id int64
		result := make(map[string]interface{})
		err := rows.Scan(&id, &result)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		ids = append(ids, id)
		results[id] = result
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	for tableName, table := range schema.Tables {
		if tableName == schema.MainTable {
			continue
		}

		query, args, err = GenerateSQLiteListObjectsSecondaryTable(schema, tableName, ids)
		if err != nil {
			return nil, fmt.Errorf("generate query: %w", err)
		}

		rows, err := tx.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("query: %w", err)
		}
		defer func(rows *sqlx.Rows) {
			_ = rows.Close()
		}(rows)

		for rows.Next() {
			result := make(map[string]interface{})
			err := rows.MapScan(result)
			if err != nil {
				return nil, fmt.Errorf("scan: %w", err)
			}

			parentId, ok := result["parent_id"].(int64)
			if !ok {
				return nil, fmt.Errorf("parent id is not an int64")
			}

			v, ok := results[parentId]
			if !ok {
				return nil, fmt.Errorf("parent id %d not found", parentId)
			}

			if table.IsList {
				additionalResults := []interface{}{}
				for rows.Next() {
					results_ := map[string]interface{}{}
					err := rows.MapScan(results_)
					if err != nil {
						return nil, fmt.Errorf("scan: %w", err)
					}
					additionalResults = append(additionalResults, results_[table.ValueField.Name])
				}
				v[tableName] = additionalResults
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

				v[tableName] = additionalResults
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

	results_ := make([]map[string]interface{}, 0)
	// keep ordering
	for _, id := range ids {
		results_ = append(results_, results[id])
	}

	return &PaginationResponse{
		Total:  len(results),
		Data:   results_,
		Offset: paginationRequest.Offset,
	}, nil
}
