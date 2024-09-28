package svg

import (
	"fmt"
	"strings"
)

// Transform represents transformations applied to SVG elements.
type Transform struct {
	Translate []int     `yaml:"translate,omitempty"` // [x, y]
	Rotate    float64   `yaml:"rotate,omitempty"`    // degrees
	Scale     []float64 `yaml:"scale,omitempty"`     // [x, y]
}

// buildTransform constructs the transformation string based on Translate, Rotate, and Scale.
func buildTransform(t *Transform) string {
	var transforms []string
	if len(t.Translate) == 2 {
		transforms = append(transforms, fmt.Sprintf("translate(%d,%d)", t.Translate[0], t.Translate[1]))
	}
	if t.Rotate != 0 {
		transforms = append(transforms, fmt.Sprintf("rotate(%f)", t.Rotate))
	}
	if len(t.Scale) == 2 {
		transforms = append(transforms, fmt.Sprintf("scale(%f,%f)", t.Scale[0], t.Scale[1]))
	}
	return strings.Join(transforms, " ")
}
