package main

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

type Form struct {
	Name       string   `yaml:"name,omitempty"`
	Theme      string   `yaml:"theme,omitempty"`
	Accessible bool     `yaml:"accessible,omitempty"`
	Groups     []*Group `yaml:"groups"`
}

type Group struct {
	Name   string   `yaml:"name,omitempty"`
	Fields []*Field `yaml:"fields"`
}

type Field struct {
	Type       string        `yaml:"type"`
	Key        string        `yaml:"key,omitempty"`
	Title      string        `yaml:"title,omitempty"`
	Value      interface{}   `yaml:"value,omitempty"`
	Options    []*Option     `yaml:"options,omitempty"`
	Validation []*Validation `yaml:"validation,omitempty"`
	Attributes *Attributes   `yaml:"attributes,omitempty"`
}

type Option struct {
	Label string      `yaml:"label"`
	Value interface{} `yaml:"value"`
}

type Validation struct {
	Condition string `yaml:"condition"`
	Error     string `yaml:"error"`
}

type Attributes struct {
	Prompt      string `yaml:"prompt,omitempty"`
	Limit       int    `yaml:"limit,omitempty"`
	Affirmative string `yaml:"affirmative,omitempty"`
	Negative    string `yaml:"negative,omitempty"`
	Height      int    `yaml:"height,omitempty"`
	CharLimit   int    `yaml:"char_limit,omitempty"`
}

type FieldWithValidation interface {
	Validate(func(string) error) huh.Field
}

func addValidation(field huh.Field, validations []*Validation) huh.Field {
	switch f := field.(type) {
	case *huh.Input:
		return f.Validate(func(s string) error {
			for _, v := range validations {
				if s == v.Condition {
					return fmt.Errorf(v.Error)
				}
			}
			return nil
		})
	case *huh.Select[string]:
		return f.Validate(func(s string) error {
			for _, v := range validations {
				if s == v.Condition {
					return fmt.Errorf(v.Error)
				}
			}
			return nil
		})
	case *huh.Confirm:
		return f.Validate(func(b bool) error {
			for _, v := range validations {
				if fmt.Sprintf("%v", b) == v.Condition {
					return fmt.Errorf(v.Error)
				}
			}
			return nil
		})
	// Add more cases for other field types as needed
	default:
		return field
	}
}

// Run executes the form and returns a map of the input values and an error if any
func (f *Form) Run() (map[string]interface{}, error) {
	// Create a map to store the input values
	values := make(map[string]interface{})

	// Create huh Form groups
	var huhGroups []*huh.Group

	// Iterate through groups and fields to build the huh Form
	for _, group := range f.Groups {
		huhFields := make([]huh.Field, 0, len(group.Fields))

		for _, field := range group.Fields {
			// Initialize the value in the map
			if field.Value != nil {
				values[field.Key] = field.Value
			} else {
				values[field.Key] = getDefaultValue(field.Type)
			}

			// Create the appropriate huh field based on the type
			var huhField huh.Field
			switch field.Type {
			case "input":
				value := values[field.Key].(string)
				input := huh.NewInput().
					Title(field.Title).
					Value(&value)
				huhField = input
				values[field.Key] = &value
			case "text":
				value := values[field.Key].(string)
				text := huh.NewText().
					Title(field.Title).
					Value(&value)
				huhField = text
				values[field.Key] = &value
			case "select":
				value := values[field.Key].(string)
				select_ := huh.NewSelect[string]().
					Title(field.Title).
					Options(createOptions(field.Options)...).
					Value(&value)
				huhField = select_
				values[field.Key] = &value
			case "multiselect":
				value := values[field.Key].([]string)
				multiSelect := huh.NewMultiSelect[string]().
					Title(field.Title).
					Options(createOptions(field.Options)...).
					Value(&value)
				huhField = multiSelect
				values[field.Key] = &value
			case "confirm":
				value := values[field.Key].(bool)
				confirm := huh.NewConfirm().
					Title(field.Title).
					Value(&value)
				huhField = confirm
				values[field.Key] = &value
			default:
				return nil, fmt.Errorf("unsupported field type: %s", field.Type)
			}

			// Add validation if specified
			if len(field.Validation) > 0 {
				huhField = addValidation(huhField, field.Validation)
			}

			huhFields = append(huhFields, huhField)
		}

		// Create the huh Group
		huhGroup := huh.NewGroup(huhFields...)
		huhGroups = append(huhGroups, huhGroup)
	}

	// Create the huh Form
	huhForm := huh.NewForm(huhGroups...)

	// Set the theme if specified
	if f.Theme != "" {
		theme, err := getTheme(f.Theme)
		if err != nil {
			return nil, err
		}
		huhForm = huhForm.WithTheme(theme)
	}

	// Set accessibility mode if specified
	if f.Accessible {
		huhForm = huhForm.WithAccessible(true)
	}

	// Run the form
	err := huhForm.Run()
	if err != nil {
		return nil, err
	}

	// Extract final values from the map
	finalValues := make(map[string]interface{})
	for key, value := range values {
		switch v := value.(type) {
		case *string:
			finalValues[key] = *v
		case *[]string:
			finalValues[key] = *v
		case *bool:
			finalValues[key] = *v
		default:
			finalValues[key] = v
		}
	}

	return finalValues, nil
}

// Helper function to get the default value based on field type
func getDefaultValue(fieldType string) interface{} {
	switch fieldType {
	case "input", "text":
		return ""
	case "select":
		return ""
	case "multiselect":
		return []string{}
	case "confirm":
		return false
	default:
		return nil
	}
}

// Helper function to create huh options from our Option structs
func createOptions(options []*Option) []huh.Option[string] {
	var huhOptions []huh.Option[string]
	for _, opt := range options {
		huhOptions = append(huhOptions, huh.NewOption(opt.Label, opt.Value.(string)))
	}
	return huhOptions
}

// Helper function to get the huh theme based on the theme name
func getTheme(themeName string) (*huh.Theme, error) {
	switch themeName {
	case "Charm":
		return huh.ThemeCharm(), nil
	case "Dracula":
		return huh.ThemeDracula(), nil
	case "Catppuccin":
		return huh.ThemeCatppuccin(), nil
	case "Base16":
		return huh.ThemeBase16(), nil
	case "Default":
		return huh.ThemeBase(), nil
	default:
		return nil, fmt.Errorf("unsupported theme: %s", themeName)
	}
}
