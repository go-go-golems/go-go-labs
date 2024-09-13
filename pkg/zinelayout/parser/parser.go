package parser

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

// ExpressionParser represents the parser and interpreter for unit expressions
type ExpressionParser struct {
	input string
	pos   int
	PPI   float64
}

// Parse parses and evaluates the input expression
func (p *ExpressionParser) Parse(input string) (float64, error) {
	p.input = input
	p.pos = 0
	res, err := p.parseExpression()
	if err != nil {
		return 0, err
	}

	p.skipWhitespace()
	if p.pos < len(p.input) {
		return 0, fmt.Errorf("unexpected character: %c", p.currentChar())
	}
	return res, nil
}

func (p *ExpressionParser) parseExpression() (float64, error) {
	left, err := p.parseTerm()
	if err != nil {
		return 0, err
	}

	for p.pos < len(p.input) {
		p.skipWhitespace()
		c := p.currentChar()
		switch c {
		case '+':
			p.pos++
			right, err := p.parseTerm()
			if err != nil {
				return 0, err
			}
			left += right
		case '-':
			p.pos++
			right, err := p.parseTerm()
			if err != nil {
				return 0, err
			}
			left -= right
		default:
			return left, nil
		}
	}

	return left, nil
}

func (p *ExpressionParser) parsePower() (float64, error) {
	left, err := p.parseFactor()
	if err != nil {
		return 0, err
	}

	for p.pos < len(p.input) {
		p.skipWhitespace()
		if p.currentChar() != '^' {
			return left, nil
		}
		p.pos++
		right, err := p.parseFactor()
		if err != nil {
			return 0, err
		}
		left = math.Pow(left, right)
	}

	return left, nil
}

func (p *ExpressionParser) parseTerm() (float64, error) {
	left, err := p.parsePower()
	if err != nil {
		return 0, err
	}

	for p.pos < len(p.input) {
		p.skipWhitespace()
		c := p.currentChar()
		switch c {
		case '*':
			p.pos++
			right, err := p.parsePower()
			if err != nil {
				return 0, err
			}
			left *= right
		case '/':
			p.pos++
			right, err := p.parsePower()
			if err != nil {
				return 0, err
			}
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			left /= right
		default:
			return left, nil
		}
	}

	return left, nil
}

func (p *ExpressionParser) parseFactor() (float64, error) {
	p.skipWhitespace()

	if p.currentChar() == '(' {
		p.pos++
		result, err := p.parseExpression()
		if err != nil {
			return 0, err
		}
		p.skipWhitespace()
		if p.currentChar() != ')' {
			return 0, fmt.Errorf("missing closing parenthesis")
		}
		p.pos++
		return result, nil
	}

	if p.currentChar() == '-' {
		p.pos++
		p.skipWhitespace()
		if p.currentChar() == '-' {
			return 0, fmt.Errorf("double negative is not allowed")
		}
		factor, err := p.parseFactor()
		if err != nil {
			return 0, err
		}
		return -factor, nil
	}

	return p.parseNumberUnit()
}

func (p *ExpressionParser) parseNumberUnit() (float64, error) {
	start := p.pos
	for p.pos < len(p.input) && (unicode.IsDigit(rune(p.currentChar())) || p.currentChar() == '.') {
		p.pos++
	}

	numStr := p.input[start:p.pos]
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", numStr)
	}

	p.skipWhitespace()
	unitStart := p.pos
	for p.pos < len(p.input) && unicode.IsLetter(rune(p.currentChar())) {
		p.pos++
	}
	unit := p.input[unitStart:p.pos]

	return p.convertToPixels(num, unit)
}

func (p *ExpressionParser) convertToPixels(value float64, unit string) (float64, error) {
	switch strings.ToLower(unit) {
	case "mm":
		return value * p.PPI / 25.4, nil
	case "cm":
		return value * p.PPI / 2.54, nil
	case "in":
		return value * p.PPI, nil
	case "pc":
		return value * p.PPI / 6, nil
	case "pt":
		return value * p.PPI / 72, nil
	case "px":
		return value, nil
	case "em", "rem":
		return value * 16 * (p.PPI / 96), nil
	case "":
		return value, nil // Assume pixels if no unit is specified
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}
}

func (p *ExpressionParser) currentChar() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

func (p *ExpressionParser) skipWhitespace() {
	for p.pos < len(p.input) && unicode.IsSpace(rune(p.currentChar())) {
		p.pos++
	}
}

// Distance represents a length that can be expressed in various units
type Distance struct {
	value float64
}

// NewDistance creates a new Distance from a float64 value (assumed to be in pixels)
func NewDistance(pixels float64) Distance {
	return Distance{value: pixels}
}

// Pixels returns the Distance value in pixels
func (d Distance) Pixels() float64 {
	return d.value
}

// MarshalJSON implements the json.Marshaler interface
func (d Distance) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.value)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (d *Distance) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case float64:
		d.value = value
	case string:
		parser := ExpressionParser{PPI: 96} // Assume default 96 PPI
		pixels, err := parser.Parse(value)
		if err != nil {
			return err
		}
		d.value = pixels
	default:
		return fmt.Errorf("invalid distance value: %v", v)
	}

	return nil
}
