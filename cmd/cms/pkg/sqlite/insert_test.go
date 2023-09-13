package sqlite

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/go-go-golems/go-go-labs/cmd/cms/pkg"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestInsertData(t *testing.T) {
	testCases := []struct {
		name   string
		schema string
		data   map[string]interface{}
	}{
		{
			name:   "single field",
			schema: "../test-data/01-single-field.yaml",
			data:   map[string]interface{}{"name": "John"},
		},
		{
			name:   "multiple fields",
			schema: "../test-data/02-multiple-fields.yaml",
			data:   map[string]interface{}{"name": "John", "age": int64(18)},
		},
		{
			name:   "multiple tables",
			schema: "../test-data/03-multiple-tables.yaml",
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
			name:   "additional field properties",
			schema: "../test-data/04-additional-field-properties.yaml",
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
			name:   "plant",
			schema: "../test-data/05-plant.yaml",
			data: map[string]interface{}{
				"name":           "Rose",
				"botanical_name": "Rosa",
				"sunlight_needs": []string{"Full Sun", "Partial Shade"},
				"categories":     []string{"Flowering", "Perennial"},
			},
		},
		{
			name:   "plant2",
			schema: "../test-data/06-plant2.yaml",
			data: map[string]interface{}{
				"name":           "Rose",
				"botanical_name": "Rosa",
				"rating":         int64(3),
				"sunlight_needs": []string{"Full Sun", "Partial Shade"},
				"categories":     []string{"Flowering", "Perennial"},
				"special_features": []map[string]interface{}{
					{"feature": "fragrant", "description": "Smells nice"},
					{"feature": "edible", "description": "Tastes good"},
				},
			},

			// TODO(manuel, 2023-09-13) We need to test for insertion errors
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the schema from the test file
			schemaBytes, err := os.ReadFile(tc.schema)
			require.NoError(t, err)

			schema, err := pkg.ParseSchemaFromYAML(schemaBytes)
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
