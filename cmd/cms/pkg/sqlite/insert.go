package sqlite

import (
	"bytes"
	"context"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/go-go-golems/go-go-labs/cmd/cms/pkg"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"text/template"
)

const insertTPL = `INSERT INTO {{.TableName}} ({{range $i, $field := .Fields}}{{if $i}}, {{end}}{{.Name}}{{end}})
VALUES ({{range $i, $field := .Fields}}{{if $i}}, {{end}}:{{.Name}}{{end}})
RETURNING id;`

func InsertData(ctx context.Context, db *sqlx.DB, schema *pkg.Schema, data map[string]interface{}) (int64, error) {
	tpl := template.Must(template.New("insertData").Parse(insertTPL))

	tableQueries := make(map[string]string, len(schema.Tables))

	for tableName, table := range schema.Tables {
		newFields := make([]pkg.Field, 0)
		newFields = append(newFields, table.Fields...)

		// for secondary tables, we need to insert the parent_id
		if tableName != schema.MainTable {
			t := true
			newFields = append(newFields, pkg.Field{
				Name:  "parent_id",
				Type:  "int",
				Index: &t,
			})
		}

		tableData := TableData{TableName: tableName, Fields: newFields}
		var SQL bytes.Buffer
		err := tpl.Execute(&SQL, tableData)
		if err != nil {
			return 0, err
		}
		tableQueries[tableName] = SQL.String()
	}

	tx, err := db.Beginx()
	if err != nil {
		return 0, err
	}

	var id int64
	// first, create the main table to get the parent id
	query, ok := tableQueries[schema.MainTable]
	if !ok {
		return 0, errors.New("main table not found")
	}
	res, err := tx.NamedExecContext(ctx, query, data)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	id64, err := res.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	id = id64

	for tableName, query := range tableQueries {
		if tableName == schema.MainTable {
			continue
		} else {
			tableDefinition, ok := schema.Tables[tableName]
			if !ok {
				return 0, errors.New("table not found")
			}

			// for secondary table, we need to:
			// - insert the parent_id field
			// - iterate over the individual rows in data
			// - insert each row

			// check if the table is a list
			if tableDefinition.IsList {
				l, err := cast.CastListToInterfaceList(data[tableName])
				if err != nil {
					return 0, err
				}
				row := map[string]interface{}{
					"parent_id": id,
				}

				_, err = tx.NamedExecContext(ctx, query, row)
				if err != nil {
					_ = tx.Rollback()
					return 0, err
				}

				_ = l
			} else {
				v, ok := data[tableName].([]map[string]interface{})
				if !ok {
					return 0, errors.New("data not found")
				}

				for _, row := range v {
					row["parent_id"] = id
					_, err := tx.NamedExecContext(ctx, query, row)
					if err != nil {
						_ = tx.Rollback()
						return 0, err
					}
				}
			}

		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}
