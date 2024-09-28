package svg

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// SVGDSL represents the root of the YAML DSL.
type SVGDSL struct {
	SVG Canvas `yaml:"svg"`
}

// Canvas represents the SVG canvas configuration.
type Canvas struct {
	Width      int              `yaml:"width"`
	Height     int              `yaml:"height"`
	Background Background       `yaml:"background"`
	Elements   []ElementWrapper `yaml:"elements"`
}

// GetElements converts []ElementWrapper to []Element.
func (c *Canvas) GetElements() []Element {
	elements := make([]Element, len(c.Elements))
	for i, ew := range c.Elements {
		elements[i] = ew.Element
	}
	return elements
}

// Background represents the canvas background.
type Background struct {
	Color string `yaml:"color,omitempty"`
	Image string `yaml:"image,omitempty"`
}

// ElementWrapper is a proxy for unmarshaling different Element types.
type ElementWrapper struct {
	Element
}

// UnmarshalYAML implements custom unmarshaling for the Element interface.
func (ew *ElementWrapper) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Temporary map to extract the "type" field
	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	typ, ok := raw["type"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid type field")
	}

	// Based on the type, unmarshal into the appropriate struct
	switch typ {
	case "rectangle":
		var rect Rectangle
		if err := mapToStruct(raw, &rect); err != nil {
			return err
		}
		ew.Element = &rect
	case "line":
		var line Line
		if err := mapToStruct(raw, &line); err != nil {
			return err
		}
		ew.Element = &line
	case "image":
		var img Image
		if err := mapToStruct(raw, &img); err != nil {
			return err
		}
		ew.Element = &img
	case "text":
		var txt Text
		if err := mapToStruct(raw, &txt); err != nil {
			return err
		}
		ew.Element = &txt
	case "group":
		var grp Group
		if err := mapToStruct(raw, &grp); err != nil {
			return err
		}
		ew.Element = &grp
	default:
		return fmt.Errorf("unsupported element type: %s", typ)
	}

	return nil
}

// mapToStruct helps in converting a map to a struct using YAML marshalling.
func mapToStruct(m map[string]interface{}, out interface{}) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, out)
}
