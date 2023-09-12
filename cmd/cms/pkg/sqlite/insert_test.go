package sqlite

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/go-go-golems/go-go-labs/cmd/cms/pkg"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestInsertData(t *testing.T) {
	testCases := []struct {
		name     string
		filePath string
		data     map[string]interface{}
	}{
		{
			name:     "single field",
			filePath: "../test-data/01-single-field.yaml",
			data:     map[string]interface{}{"name": "John"},
		},
		{
			name:     "multiple fields",
			filePath: "../test-data/02-multiple-fields.yaml",
			data:     map[string]interface{}{"name": "John", "age": int64(18)},
		},
		{
			name:     "multiple tables",
			filePath: "../test-data/03-multiple-tables.yaml",
			data: map[string]interface{}{
				"name": "John",
				"age":  int64(18),
				"addresses": []map[string]interface{}{{
					"street": "123 Main St",
					"city":   "Springfield",
				},
				},
			},
		},
		{
			name:     "additional field properties",
			filePath: "../test-data/04-additional-field-properties.yaml",
			data: map[string]interface{}{
				"name": "John",
				"age":  int64(18),
				"addresses": []map[string]interface{}{{
					"street":  "123 Main St",
					"city":    "Springfield",
					"country": "USA",
				}},
			},
		},
		{
			name:     "plant",
			filePath: "../test-data/05-plant.yaml",
			data: map[string]interface{}{
				"name":           "Rose",
				"botanical_name": "Rosa",
				"sunlight_needs": []string{"Full Sun", "Partial Shade"},
				"categories":     []string{"Flowering", "Perennial"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the schema from the test file
			schemaBytes, err := ioutil.ReadFile(tc.filePath)
			require.NoError(t, err)

			schema, err := pkg.ParseSchema(schemaBytes)
			require.NoError(t, err)

			// Create an in-memory SQLite database
			db, err := sqlx.Connect("sqlite3", ":memory:")
			require.NoError(t, err)
			defer func(db *sqlx.DB) {
				_ = db.Close()
			}(db)

			ctx := context.Background()
			// Create the schema in the database
			err = CreateSchema(ctx, schema, db)
			require.NoError(t, err)

			// Insert the data
			id, err := InsertData(ctx, db, schema, tc.data)
			require.NoError(t, err)
			_ = id

			// Query the data back out
			rows, err := db.Queryx("SELECT * FROM " + schema.MainTable)
			require.NoError(t, err)

			// Assert that the data matches what was inserted
			for rows.Next() {
				result := map[string]interface{}{}
				err = rows.MapScan(result)
				require.NoError(t, err)

				// check that id is the same
				assert.Equal(t, id, result["id"])

				for _, field := range schema.Tables[schema.MainTable].Fields {
					value, ok := tc.data[field.Name]
					require.True(t, ok)

					assert.Equal(t, value, result[field.Name])
				}
			}

			// Now go over the secondary tables
			for tableName, table := range schema.Tables {
				if tableName == schema.MainTable {
					continue
				}

				// Query the data back out
				rows, err := db.Queryx("SELECT * FROM " + tableName)
				require.NoError(t, err, "failed to query table:"+tableName)

				idx := 0

				if table.IsList {
					dataRows, err := cast.CastListToInterfaceList(tc.data[tableName])
					require.NoError(t, err, "failed to cast data for table:"+tableName)

					// Assert that the data matches what was inserted
					for rows.Next() {
						result := map[string]interface{}{}
						err = rows.MapScan(result)
						require.NoError(t, err, "failed to scan row for table:"+tableName)

						// check that parent_id is the same
						assert.Equal(t, id, result["parent_id"])

						require.GreaterOrEqual(t, len(dataRows), idx, fmt.Sprintf("not enough rows for table: %s", tableName))
						data := dataRows[idx]

						field := table.ValueField
						require.Equal(t, data, result[field.Name], "wrong value for field: "+field.Name)

						idx++
					}

				} else {
					dataRows, ok := tc.data[tableName].([]map[string]interface{})
					require.True(t, ok, "failed to get data for table:"+tableName)

					// Assert that the data matches what was inserted
					for rows.Next() {
						result := map[string]interface{}{}
						err = rows.MapScan(result)
						require.NoError(t, err, "failed to scan row for table:"+tableName)

						// check that parent_id is the same
						assert.Equal(t, id, result["parent_id"])

						require.GreaterOrEqual(t, len(dataRows), idx, fmt.Sprintf("not enough rows for table: %s", tableName))
						data := dataRows[idx]

						for _, field := range table.Fields {
							value, ok := data[field.Name]
							require.True(t, ok, "field not found: "+field.Name)

							assert.Equal(t, value, result[field.Name], "wrong value for field: "+field.Name)
						}

						idx++
					}
				}
			}
		})
	}
}
