package sqlite

import (
	"bytes"
	"github.com/go-go-golems/go-go-labs/cmd/cms/pkg"
	"text/template"
)

const createTableTPL = `CREATE TABLE IF NOT EXISTS
    {{.TableName}} (
        id INTEGER PRIMARY KEY AUTOINCREMENT
        {{- range $i, $field := .Fields}},
        {{.Name}} {{$field.Type|sqliteDataType}}
        {{- if .Unique}} UNIQUE{{end}}
        {{- end}});`

func GenerateSQLiteCreateTable(schema *pkg.Schema) (map[string][]string, error) {
	tpl := template.Must(template.New("createTable").Funcs(template.FuncMap{
		"sqliteDataType": sqliteDataType,
	}).Parse(createTableTPL))

	tableQueries := make(map[string][]string, len(schema.Tables))

	// creating set tableNames for checking existence of table with same name as field
	tableNames := make(map[string]bool)
	for tableName := range schema.Tables {
		tableNames[tableName] = true
	}
	for tableName, table := range schema.Tables {
		newFields := make([]pkg.Field, 0)

		newFields = append(newFields, table.Fields...)

		// if this is a secondary table, add the parent_id field
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
			return nil, err
		}
		tableQueries[tableName] = []string{SQL.String()}

		// create indexes
		for _, field := range table.Fields {
			if field.Index != nil && *field.Index {
				indexQuery := "CREATE INDEX IF NOT EXISTS " + tableName + "_" + field.Name + "_idx ON " + tableName + "(" + field.Name + ")"
				tableQueries[tableName] = append(tableQueries[tableName], indexQuery)
			}
		}
	}

	return tableQueries, nil
}
