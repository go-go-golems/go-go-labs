package pkg

import "gopkg.in/yaml.v3"

// InputType is used to specify the type of input for a CMS field.
type InputType string

const (
	InputTypeTextInput    InputType = "text-input"
	InputTypeTextArea     InputType = "textarea"
	InputTypeIntegerInput InputType = "integer-input"
	InputTypeIntegerRange InputType = "integer-range"
	InputTypeEditableList InputType = "editable-list"
)

// Input is a struct that represents an input for a CMS field.
type Input struct {
	Name  string    `yaml:"name"`
	Type  InputType `yaml:"type"`
	Label string    `yaml:"label,omitempty"`
	// Optional help tooltip
	Tooltip string `yaml:"tooltip,omitempty"`
	// Target is a string `table.field` that specifies the field to which the input is linked.
	Target string `yaml:"target,omitempty"`
}

// Row is a struct that represents a row in a CMS section. It is basically a list of fields.
type Row struct {
	Inputs []Input `yaml:"inputs"`
}

// Section is a struct that represents a section in a CMS object.
// It has multiple rows and can be used to represent a list of objects or a list of fields.
type Section struct {
	Title        string `yaml:"title"`
	HasAddRemove bool   `yaml:"has-add-remove,omitempty"`
	Rows         []Row  `yaml:"rows"`
}

// Layout is a struct that represents the layout of a CMS object.
type Layout struct {
	Sections []Section `yaml:"section"`
}

func ParseLayout(input []byte) (*Layout, error) {
	var layout Layout
	err := yaml.Unmarshal(input, &layout)
	if err != nil {
		return nil, err
	}
	return &layout, nil
}
