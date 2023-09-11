package sqlite

import (
	"github.com/go-go-golems/go-go-labs/cmd/cms/pkg"
	"io/ioutil"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func TestGenerateSQLiteCreateTable(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
	}{
		{
			name:     "Single field",
			filePath: "../test-data/01-single-field.yaml",
		},
		{
			name:     "Multiple fields",
			filePath: "../test-data/02-multiple-fields.yaml",
		},
		{
			name:     "Multiple tables",
			filePath: "../test-data/03-multiple-tables.yaml",
		},
		{
			name:     "Additional field properties",
			filePath: "../test-data/04-additional-field-properties.yaml",
		},
		{
			name:     "Plant",
			filePath: "../test-data/05-plant.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read the test data
			data, err := ioutil.ReadFile(tt.filePath)
			require.NoError(t, err)

			// Parse the schema
			schema, err := pkg.ParseSchema(data)
			require.NoError(t, err)

			// Generate the SQL
			sql, err := GenerateSQLiteCreateTable(schema)
			require.NoError(t, err)

			// Create an in-memory SQLite database
			db, err := sqlx.Connect("sqlite3", ":memory:")
			require.NoError(t, err)
			defer func(db *sqlx.DB) {
				_ = db.Close()
			}(db)

			// Execute the SQL
			for _, queries := range sql {
				for _, query := range queries {
					_, err := db.Exec(query)
					require.NoError(t, err, "failed to execute query: "+query)
				}
			}
			for tableName := range sql {
				// Check that the table exists
				row := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", tableName)
				var name string
				err = row.Scan(&name)
				require.NoError(t, err)
				require.Equal(t, tableName, name)

				columns := make([]map[string]interface{}, 0)

				// load all rows into a column map
				rows, err := db.Queryx("PRAGMA table_info(" + tableName + ")")
				require.NoError(t, err)
				for rows.Next() {
					column := make(map[string]interface{})
					err := rows.MapScan(column)
					require.NoError(t, err)
					columns = append(columns, column)
				}

				// Check the columns
				for _, field := range schema.Tables[tableName].Fields {

					// Find the column
					var column map[string]interface{}
					for _, col := range columns {
						if col["name"] == field.Name {
							column = col
							break
						}
					}
					require.NotNil(t, column, "column not found: "+field.Name)

					// Check the column type
					expectedType := sqliteDataType(field.Type)
					require.Equal(t, expectedType, column["type"], "wrong type for column: "+field.Name)
				}

				// for secondary tables, check that we have a parent_id field
				if tableName != schema.MainTable {
					var column map[string]interface{}
					for _, col := range columns {
						if col["name"] == "parent_id" {
							column = col
							break
						}
					}
					require.NotNil(t, column, "column not found: parent_id")
					require.Equal(t, "INTEGER", column["type"], "wrong type for column: parent_id")
				}
			}
		})
	}
}
