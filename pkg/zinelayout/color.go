package zinelayout

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// CustomColor is a wrapper around color.RGBA that implements yaml.Unmarshaler
type CustomColor struct {
	color.RGBA
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (c *CustomColor) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		// Handle hex color string or color name
		return c.unmarshalScalar(value.Value)
	} else if value.Kind == yaml.SequenceNode {
		// Handle list of numbers
		return c.unmarshalList(value)
	}
	return fmt.Errorf("invalid color format")
}

func (c *CustomColor) unmarshalScalar(s string) error {
	// Check if it's a hex color
	if strings.HasPrefix(s, "#") {
		return c.unmarshalHex(s)
	}
	// Check if it's a standard color name
	if rgba, ok := standardColors[strings.ToLower(s)]; ok {
		c.RGBA = rgba
		return nil
	}
	return fmt.Errorf("invalid color: %s", s)
}

func (c *CustomColor) unmarshalHex(s string) error {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return fmt.Errorf("invalid hex color: %s", s)
	}
	rgb, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return err
	}
	c.R = uint8(rgb >> 16)
	c.G = uint8((rgb >> 8) & 0xFF)
	c.B = uint8(rgb & 0xFF)
	c.A = 255
	return nil
}

func (c *CustomColor) unmarshalList(node *yaml.Node) error {
	var values []uint8
	if err := node.Decode(&values); err != nil {
		return err
	}
	if len(values) != 3 && len(values) != 4 {
		return fmt.Errorf("invalid color list length: %d", len(values))
	}
	c.R = values[0]
	c.G = values[1]
	c.B = values[2]
	if len(values) == 4 {
		c.A = values[3]
	} else {
		c.A = 255
	}
	return nil
}

// standardColors maps color names to their RGBA values
var standardColors = map[string]color.RGBA{
	"black":     {0, 0, 0, 255},
	"white":     {255, 255, 255, 255},
	"red":       {255, 0, 0, 255},
	"green":     {0, 255, 0, 255},
	"blue":      {0, 0, 255, 255},
	"yellow":    {255, 255, 0, 255},
	"cyan":      {0, 255, 255, 255},
	"magenta":   {255, 0, 255, 255},
	"gray":      {128, 128, 128, 255},
	"grey":      {128, 128, 128, 255},
	"lightgray": {211, 211, 211, 255},
	"lightgrey": {211, 211, 211, 255},
	"darkgray":  {169, 169, 169, 255},
	"darkgrey":  {169, 169, 169, 255},
	"orange":    {255, 165, 0, 255},
	"purple":    {128, 0, 128, 255},
	"brown":     {165, 42, 42, 255},
	"pink":      {255, 192, 203, 255},
	// Add more colors as needed
}
