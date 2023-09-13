package sqlite

import (
	"context"
	"encoding/json"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/go-go-golems/go-go-labs/cmd/cms/pkg"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"os"
	"reflect"
	"testing"
)

func TestGetObject(t *testing.T) {
	testCases := []struct {
		name   string
		schema string
		files  []string
	}{
		{
			name:   "Single plant with all properties",
			schema: "../test-data/06-plant2.yaml",
			files:  []string{"../test-data/06-plants/04-normal.json"},
		},
		{
			name:   "Empty sunlight",
			schema: "../test-data/06-plant2.yaml",
			files:  []string{"../test-data/06-plants/05-empty-sunlight.json"},
		},
		{
			name:   "Missing property",
			schema: "../test-data/06-plant2.yaml",
			files:  []string{"../test-data/06-plants/06-missing-property.json"},
		},
		{
			name:   "Multiple secondary",
			schema: "../test-data/06-plant2.yaml",
			files:  []string{"../test-data/06-plants/07-multiple-secondary.json"},
		},
		{
			name:   "All plants",
			schema: "../test-data/06-plant2.yaml",
			files: []string{
				"../test-data/06-plants/04-normal.json",
				"../test-data/06-plants/05-empty-sunlight.json",
				"../test-data/06-plants/06-missing-property.json",
				"../test-data/06-plants/07-multiple-secondary.json",
			},
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

			ids := []int64{}
			var objects []map[string]interface{}

			for _, file := range tc.files {
				// Parse the data from the test file
				dataBytes, err := os.ReadFile(file)
				require.NoError(t, err)

				data := map[string]interface{}{}

				err = json.Unmarshal(dataBytes, &data)
				require.NoError(t, err)

				id, err := InsertData(ctx, db, schema, data)
				require.NoError(t, err)

				ids = append(ids, id)
				objects = append(objects, data)
			}

			for idx := range objects {
				id := ids[idx]

				// Get the object
				object, err := GetObject(ctx, db, schema, int(id))
				require.NoError(t, err)

				// delete id field
				delete(object, "id")

				err = compareExpectedActual(objects[idx], object)
				require.NoError(t, err)

			}
		})
	}
}

func isEmptySlice(v interface{}) bool {
	if reflect.TypeOf(v).Kind() != reflect.Slice {
		return false
	}

	if reflect.ValueOf(v).Len() == 0 {
		return true
	}

	return false
}

func compareObject(v map[string]interface{}, vName string, v2 map[string]interface{}, v2Name string) error {
	for k, v_ := range v {
		if v_ == nil || isEmptySlice(v_) {
			// check that v2[k] is not present, nil or an empty list
			v2_, ok := v2[k]
			if !ok {
				continue
			}

			if v2_ == nil || isEmptySlice(v2_) {
				continue
			}
		}

		v2_, ok := v2[k]
		if !ok {
			return errors.Errorf("field %s not found in %s, %s %v", k, v2Name, vName, v_)
		}

		if v2_ == nil {
			continue
		}

		// special case lists
		if reflect.TypeOf(v_).Kind() == reflect.Slice {
			if reflect.TypeOf(v2_).Kind() != reflect.Slice {
				return errors.Errorf("field %s type mismatch, %s %s != %s %s", k,
					vName, reflect.TypeOf(v_),
					v2Name, reflect.TypeOf(v2_))
			}

			v__, err := cast.CastListToInterfaceList(v_)
			if err != nil {
				return errors.Errorf("could not cast field %s %s to []interface{}", vName, k)
			}

			v2__, err := cast.CastListToInterfaceList(v2_)
			if err != nil {
				return errors.Errorf("could not cast field %s %s to []interface{}", v2Name, k)
			}

			if len(v__) != len(v2__) {
				return errors.Errorf("field %s length mismatch, %s %d != %s %d", k,
					vName, len(v__),
					v2Name, len(v2__))
			}

			for idx := range v__ {
				value := v__[idx]
				value2 := v2__[idx]

				// check that the underlying type is the same
				if reflect.TypeOf(value) != reflect.TypeOf(value2) {
					return errors.Errorf("field %s type mismatch, %s %s != %s %s", k,
						vName, reflect.TypeOf(value),
						v2Name, reflect.TypeOf(value2))
				}

				// if the underlying type is a map, compare the maps
				if reflect.TypeOf(value).Kind() == reflect.Map {
					return compareExpectedActual(v__[idx].(map[string]interface{}), v2__[idx].(map[string]interface{}))
				}

				// otherwise, compare the values
				if value != value2 {
					return errors.Errorf("field %s value mismatch, %s %v != %s %v", k,
						vName, v_,
						v2Name, v2_)
				}
			}

			continue
		}

		if reflect.TypeOf(v_) != reflect.TypeOf(v2_) {
			return errors.Errorf("field %s type mismatch, %s %s != %s %s", k,
				vName, reflect.TypeOf(v_),
				v2Name, reflect.TypeOf(v2_))
		}

		if reflect.TypeOf(v_).Kind() == reflect.Map {
			var err error
			if vName == "expected" {
				err = compareExpectedActual(v_.(map[string]interface{}), v2_.(map[string]interface{}))
			} else {
				err = compareExpectedActual(v2_.(map[string]interface{}), v_.(map[string]interface{}))
			}
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// compare object by checking that all fields in v are in v2, except if v[field] is nil
// compare object by checking that all fields in v2 are in v, except if v2[field] is nil
func compareExpectedActual(v map[string]interface{}, v2 map[string]interface{}) error {
	err := compareObject(v, "expected", v2, "actual")
	if err != nil {
		return err
	}

	err = compareObject(v2, "actual", v, "expected")
	if err != nil {
		return err
	}

	return nil
}
