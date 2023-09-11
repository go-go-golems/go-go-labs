package sqlite

import (
	"bytes"
	"github.com/go-go-golems/go-go-labs/cmd/cms/pkg"
	"text/template"
)

// This assumes that all the fields in the table are updated.
const updateObjectTpl = `UPDATE {{.TableName}} SET {{range $i, $field := .Fields}}{{if $i}}, {{end}}{{.Name}} = :{{.Name}}{{end}} WHERE id = :id`

func GenerateSQLiteUpdateObject(schema *pkg.Schema) (map[string]string, error) {
	tpl := template.Must(template.New("updateObject").Funcs(template.FuncMap{}).Parse(updateObjectTpl))

	updateQuery := make(map[string]string, len(schema.Tables))

	for tableName, table := range schema.Tables {
		tableData := TableData{TableName: tableName, Fields: table.Fields}
		var SQL bytes.Buffer
		err := tpl.Execute(&SQL, tableData)
		if err != nil {
			return nil, err
		}
		updateQuery[tableName] = SQL.String()
	}
	return updateQuery, nil
}
