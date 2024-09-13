package zinelayout

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/go-go-labs/pkg/zinelayout/parser"
	"gopkg.in/yaml.v3"
)

type Margin struct {
	Top    MarginValue `yaml:"top"`
	Bottom MarginValue `yaml:"bottom"`
	Left   MarginValue `yaml:"left"`
	Right  MarginValue `yaml:"right"`
	PPI    float64     `yaml:"-"`
}

type MarginValue struct {
	Expression string `yaml:"expression"`
	Pixels     int    `yaml:"-"`
}

func (m *Margin) UnmarshalYAML(value *yaml.Node) error {
	type rawMargin struct {
		Top    string `yaml:"top"`
		Bottom string `yaml:"bottom"`
		Left   string `yaml:"left"`
		Right  string `yaml:"right"`
	}

	var raw rawMargin
	if err := value.Decode(&raw); err != nil {
		return err
	}

	m.Top = MarginValue{Expression: raw.Top}
	m.Bottom = MarginValue{Expression: raw.Bottom}
	m.Left = MarginValue{Expression: raw.Left}
	m.Right = MarginValue{Expression: raw.Right}

	return nil
}

func (m Margin) MarshalYAML() (interface{}, error) {
	return struct {
		Top    string `yaml:"top"`
		Bottom string `yaml:"bottom"`
		Left   string `yaml:"left"`
		Right  string `yaml:"right"`
	}{
		Top:    m.Top.Expression,
		Bottom: m.Bottom.Expression,
		Left:   m.Left.Expression,
		Right:  m.Right.Expression,
	}, nil
}

func (m *Margin) ComputePixelValues(ppi float64) error {
	m.PPI = ppi
	converter := &parser.UnitConverter{PPI: ppi}

	for _, mv := range []*MarginValue{&m.Top, &m.Bottom, &m.Left, &m.Right} {
		if strings.TrimSpace(mv.Expression) == "" {
			mv.Pixels = 0
			continue
		}
		pixels, err := converter.ToPixels(mv.Expression)
		if err != nil {
			return err
		}
		fmt.Println("Computing pixels for", mv.Expression, "=", pixels)
		mv.Pixels = int(pixels)
	}

	return nil
}

func (mv *MarginValue) UpdatePixels(pixels int, ppi float64) {
	mv.Pixels = pixels
	converter := &parser.UnitConverter{PPI: ppi}
	mv.Expression, _ = converter.FromPixels(float64(pixels), "px")
}

func (mv *MarginValue) UpdateExpression(expression string, ppi float64) error {
	converter := &parser.UnitConverter{PPI: ppi}
	pixels, err := converter.ToPixels(expression)
	if err != nil {
		return err
	}
	mv.Expression = expression
	mv.Pixels = int(pixels)
	return nil
}
