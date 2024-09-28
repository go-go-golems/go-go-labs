package svg

import (
	"fmt"

	"gopkg.in/yaml.v3"
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
func (ew *ElementWrapper) UnmarshalYAML(value *yaml.Node) error {
	var temp struct {
		Type string `yaml:"type"`
	}
	if err := value.Decode(&temp); err != nil {
		return err
	}

	// Based on the type, unmarshal into the appropriate struct
	var element Element
	switch temp.Type {
	case "rectangle":
		element = &Rectangle{}
	case "line":
		element = &Line{}
	case "image":
		element = &Image{}
	case "text":
		element = &Text{}
	case "group":
		element = &Group{}
	case "circle":
		element = &Circle{}
	default:
		return fmt.Errorf("unsupported element type: %s", temp.Type)
	}

	if err := value.Decode(element); err != nil {
		return err
	}

	ew.Element = element
	return nil
}

// MarshalYAML implements custom marshaling for the Element interface.
func (ew ElementWrapper) MarshalYAML() (interface{}, error) {
	return ew.Element, nil
}
