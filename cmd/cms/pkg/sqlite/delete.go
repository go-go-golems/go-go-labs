package sqlite

import (
	"bytes"
	"github.com/go-go-golems/go-go-labs/cmd/cms/pkg"
	"text/template"
)

const deleteObjectTpl = `DELETE FROM {{.TableName}} WHERE id = ?`

func GenerateSQLiteDeleteObject(schema *pkg.Schema) (map[string]string, error) {
	tpl := template.Must(template.New("deleteObject").Funcs(template.FuncMap{}).Parse(deleteObjectTpl))

	deleteQuery := make(map[string]string, len(schema.Tables))
	for tableName := range schema.Tables {
		tableData := TableData{TableName: tableName}
		var SQL bytes.Buffer
		err := tpl.Execute(&SQL, tableData)
		if err != nil {
			return nil, err
		}
		deleteQuery[tableName] = SQL.String()
	}
	return deleteQuery, nil
}
