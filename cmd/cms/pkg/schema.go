package pkg

import (
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"gopkg.in/yaml.v3"
)

// Schema is a struct that represents the schema of a CMS object.
// This contains all the necessary tables, as well as the main table for the object
// to which all other tables are joined on its `id` using the `parent_id` field.
//
// A CMS object is represents by a main table and multiple additional tables used
// to represent list or object fields.
//
// When mapped to SQL, every table contains an id field that can be used to do foreign relational
// joins.
//
// - For list fields, the value of the list item is stored in a column named `value` in the respective secondary table.
// - For object fields, every column of the secondary table represents one key of the object. There is supposed to be only one row.
// - For object list fields, the table is the same as for object fields, but there can be multiple rows.
//
// For example, an object for users with a list of addresses would have the schema.
//
// tables:
//
//		users:
//		  fields:
//		    - name: name
//		      type: string
//			  unique: true
//		    - name: age
//		      type: int
//		      index: true
//		addresses:
//		  fields:
//		    - name: street
//		      type: string
//		    - name: city
//		      type: string
//		    - name: country
//		      type: string
//		      index: true
//	 categories:
//	   value-field:
//	     name: category
//	     type: string
//	   is-list: true
//	   help: "A list of categories"
//
// main-table: users
type Schema struct {
	Tables    map[string]Table `yaml:"tables"`
	MainTable string           `yaml:"main-table"`
}

// Table is a struct that represents a table in a database.
type Table struct {
	Help string `yaml:"help,omitempty"`
	// The fields of the table, empty is IsList is true
	Fields []Field `yaml:"fields,omitempty"`
	// Indicates where this table only stores a list of values, and thus only has a single field `ValueField`
	IsList bool `yaml:"is-list,omitempty"`
	// A single field storing elements of the list, only used if `IsList` is true
	ValueField *Field `yaml:"value-field,omitempty"`
}

// Field is a struct that represents a field in a table.
// It specifies the type, which can be:
// -  ParameterTypeString         = "string"
// -  ParameterTypeInteger     = "int"
// -  ParameterTypeFloat       = "float"
// -  ParameterTypeBool        = "bool"
// -  ParameterTypeDate        = "date"
// - ParameterTypeChoice       = "choice" -> string
type Field struct {
	Name string                   `yaml:"name"`
	Type parameters.ParameterType `yaml:"type"`
	Help string                   `yaml:"help,omitempty"`
	// Specifies if a unique index should be created for the given field
	Unique *bool `yaml:"unique,omitempty"`
	// Specifies if an index should be created for the given field
	Index *bool `yaml:"index,omitempty"`
	// For numeric types, specifies the minimum and maximum values
	Min *float32 `yaml:"min,omitempty"`
	Max *float32 `yaml:"max,omitempty"`
	// For documentation purposes, specifies the unit of the field
	Unit *string `yaml:"unit,omitempty"`
	// For ParameterTypeChoice, specifies the possible choices
	Choices []string `yaml:"choices,omitempty"`
}

type CMSObject struct {
	Schema *Schema `yaml:"schema"`
	Layout *Layout `yaml:"layout"`
}

func ParseSchema(input []byte) (*Schema, error) {
	schema := &Schema{
		Tables: make(map[string]Table),
	}

	err := yaml.Unmarshal(input, schema)
	if err != nil {
		return nil, err
	}
	return schema, nil
}
