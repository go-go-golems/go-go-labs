package zinelayout

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/go-go-labs/pkg/zinelayout/parser"
	"github.com/rs/zerolog/log"
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

func (mv *MarginValue) String() string {
	if mv.Expression == "" {
		return fmt.Sprintf("%dpx", mv.Pixels)
	}
	return fmt.Sprintf("%s (%dpx)", mv.Expression, mv.Pixels)
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

func (m *Margin) String() string {
	return fmt.Sprintf("Margin(Top: %s, Bottom: %s, Left: %s, Right: %s)", m.Top.String(), m.Bottom.String(), m.Left.String(), m.Right.String())
}

func (m *Margin) ComputePixelValues(ppi float64) error {
	m.PPI = ppi
	uc := parser.UnitConverter{PPI: ppi}
	p := parser.ExpressionParser{PPI: ppi}

	for _, mv := range []*MarginValue{&m.Top, &m.Bottom, &m.Left, &m.Right} {
		log.Trace().
			Interface("marginValue", mv).
			Float64("ppi", ppi).
			Msg("MarginValue before")

		if strings.TrimSpace(mv.Expression) == "" {
			mv.Pixels = 0
			continue
		}
		val, err := p.Parse(mv.Expression)
		if err != nil {
			return err
		}
		log.Trace().
			Str("value", val.String()).
			Msg("MarginValue after parse")

		pixels, err := uc.ToPixels(val.Val, val.Unit)
		if err != nil {
			return err
		}
		log.Trace().
			Float64("pixels", pixels).
			Msg("MarginValue after to pixels")

		mv.Pixels = int(pixels)
		log.Trace().
			Interface("marginValue", mv).
			Msg("MarginValue after")
	}

	return nil
}

func (mv *MarginValue) UpdatePixels(pixels int, ppi float64) {
	mv.Pixels = pixels
	mv.Expression = fmt.Sprintf("%dpx", pixels)
}

func (mv *MarginValue) UpdateExpression(expression string, ppi float64) error {
	p := parser.ExpressionParser{PPI: ppi}
	uc := parser.UnitConverter{PPI: ppi}
	val, err := p.Parse(expression)
	if err != nil {
		return err
	}
	mv.Expression = expression
	pixels, err := uc.ToPixels(val.Val, val.Unit)
	if err != nil {
		return err
	}
	mv.Pixels = int(pixels)
	return nil
}
