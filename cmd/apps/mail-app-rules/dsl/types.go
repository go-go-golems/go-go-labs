package dsl

// Rule represents a complete IMAP DSL rule
type Rule struct {
	Name        string       `yaml:"name"`
	Description string       `yaml:"description"`
	Search      SearchConfig `yaml:"search"`
	Output      OutputConfig `yaml:"output"`
}

// SearchConfig defines search criteria
type SearchConfig struct {
	Since      string `yaml:"since"`
	Before     string `yaml:"before"`
	On         string `yaml:"on"`
	WithinDays int    `yaml:"within_days"`
	From       string `yaml:"from"`
}

// OutputConfig defines output formatting
type OutputConfig struct {
	Format string        `yaml:"format"` // json, text, table
	Fields []interface{} `yaml:"fields"`
}

// UnmarshalYAML implements custom unmarshaling for fields
func (o *OutputConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Define a temporary struct to unmarshal into
	type tempOutputConfig struct {
		Format string        `yaml:"format"`
		Fields []interface{} `yaml:"fields"`
	}

	// Unmarshal into the temporary struct
	var temp tempOutputConfig
	if err := unmarshal(&temp); err != nil {
		return err
	}

	// Copy the simple fields
	o.Format = temp.Format
	o.Fields = make([]interface{}, len(temp.Fields))

	// Process each field
	for i, field := range temp.Fields {
		switch f := field.(type) {
		case string:
			// Simple field like "subject", "from", etc.
			o.Fields[i] = Field{Name: f}
		case map[string]interface{}:
			// Complex field like body: {type: "text/plain", max_length: 1000}
			if bodyMap, ok := f["body"].(map[string]interface{}); ok {
				bodyField := &BodyField{}
				if t, ok := bodyMap["type"].(string); ok {
					bodyField.Type = t
				}
				if ml, ok := bodyMap["max_length"].(int); ok {
					bodyField.MaxLength = ml
				}
				o.Fields[i] = Field{Name: "body", Body: bodyField}
			} else {
				// Just store as is for now
				o.Fields[i] = field
			}
		default:
			// Just store as is
			o.Fields[i] = field
		}
	}

	return nil
}

// Field represents an output field, which can be a simple string or complex field
type Field struct {
	Name string
	Body *BodyField
	// More field types will be added later
}

// BodyField represents body output configuration
type BodyField struct {
	Type      string `yaml:"type"`
	MaxLength int    `yaml:"max_length"`
}
