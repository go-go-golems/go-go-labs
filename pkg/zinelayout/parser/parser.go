package parser

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Value represents a numeric value with its associated unit and position in the input string
type Value struct {
	Val      float64
	Unit     string
	StartPos int
	EndPos   int
}

// ExpressionParser represents the parser and interpreter for unit expressions
type ExpressionParser struct {
	input         string
	pos           int
	PPI           float64
	Debug         bool
	depth         int
	unitConverter *UnitConverter
}

// Parse parses and evaluates the input expression
func (p *ExpressionParser) Parse(input string) (Value, error) {
	p.input = input
	p.pos = 0
	p.depth = 0
	p.unitConverter = &UnitConverter{PPI: p.PPI}
	if p.Debug {
		fmt.Printf("Parsing expression: %s\n", input)
	}
	res, err := p.parseExpression()
	if err != nil {
		return Value{}, err
	}

	p.skipWhitespace()
	if p.pos < len(p.input) {
		return Value{}, fmt.Errorf("unexpected character: %c", p.currentChar())
	}
	if p.Debug {
		fmt.Printf("Final result: %+v\n", res)
	}
	return res, nil
}

func (p *ExpressionParser) debugPrint(format string, a ...interface{}) {
	if p.Debug {
		indent := strings.Repeat("  ", p.depth)
		fmt.Printf(indent+format+"\n", a...)
	}
}

func (p *ExpressionParser) parseExpression() (Value, error) {
	p.depth++
	defer func() { p.depth-- }()
	p.debugPrint("Parsing expression at position %d", p.pos)

	left, err := p.parseTerm()
	if err != nil {
		return Value{}, err
	}
	p.debugPrint("Initial term: %+v", left)

	for p.pos < len(p.input) {
		p.skipWhitespace()
		c := p.currentChar()
		switch c {
		case '+', '-':
			p.debugPrint("Found operator: %c", c)
			p.pos++
			right, err := p.parseTerm()
			if err != nil {
				return Value{}, err
			}
			left, err = p.performOperation(left, right, c)
			if err != nil {
				return Value{}, err
			}
			p.debugPrint("After operation: %+v", left)
		default:
			p.debugPrint("Ending expression, final result: %+v", left)
			return left, nil
		}
	}

	p.debugPrint("Reached end of input, final result: %+v", left)
	return left, nil
}

func (p *ExpressionParser) parseTerm() (Value, error) {
	p.depth++
	defer func() { p.depth-- }()
	p.debugPrint("Parsing term at position %d", p.pos)

	left, err := p.parseFactor()
	if err != nil {
		return Value{}, err
	}

	for p.pos < len(p.input) {
		p.skipWhitespace()
		c := p.currentChar()
		switch c {
		case '*', '/':
			p.debugPrint("Found operator: %c", c)
			p.pos++
			right, err := p.parseFactor()
			if err != nil {
				return Value{}, err
			}
			left, err = p.performOperation(left, right, c)
			if err != nil {
				return Value{}, err
			}
			p.debugPrint("After operation: %+v", left)
		default:
			return left, nil
		}
	}

	return left, nil
}

func (p *ExpressionParser) parseFactor() (Value, error) {
	p.depth++
	defer func() { p.depth-- }()
	p.debugPrint("Parsing factor at position %d", p.pos)

	p.skipWhitespace()
	startPos := p.pos

	if p.currentChar() == '(' {
		p.pos++
		result, err := p.parseExpression()
		if err != nil {
			return Value{}, err
		}
		p.skipWhitespace()
		if p.currentChar() != ')' {
			return Value{}, fmt.Errorf("missing closing parenthesis")
		}
		p.pos++
		result.StartPos = startPos
		result.EndPos = p.pos
		return result, nil
	}

	if p.currentChar() == '-' {
		p.pos++
		p.skipWhitespace()
		if p.currentChar() == '-' {
			return Value{}, fmt.Errorf("double negative is not allowed")
		}
		factor, err := p.parseFactor()
		if err != nil {
			return Value{}, err
		}
		factor.Val = -factor.Val
		factor.StartPos = startPos
		return factor, nil
	}

	return p.parseNumber()
}

func (p *ExpressionParser) parseNumber() (Value, error) {
	p.depth++
	defer func() { p.depth-- }()
	p.debugPrint("Parsing number at position %d", p.pos)

	startPos := p.pos
	for p.pos < len(p.input) && (unicode.IsDigit(rune(p.currentChar())) || p.currentChar() == '.') {
		p.pos++
	}

	numStr := p.input[startPos:p.pos]
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return Value{}, fmt.Errorf("invalid number: %s", numStr)
	}

	unit, err := p.parseUnit()
	if err != nil {
		return Value{}, err
	}

	return Value{Val: num, Unit: unit, StartPos: startPos, EndPos: p.pos}, nil
}

func (p *ExpressionParser) parseUnit() (string, error) {
	p.depth++
	defer func() { p.depth-- }()
	p.debugPrint("Parsing unit at position %d", p.pos)

	p.skipWhitespace()
	unitStart := p.pos
	for p.pos < len(p.input) && unicode.IsLetter(rune(p.currentChar())) {
		p.pos++
	}
	return p.input[unitStart:p.pos], nil
}

func (p *ExpressionParser) performOperation(left, right Value, op byte) (Value, error) {
	if left.Unit == "" {
		left.Unit = right.Unit
	} else if right.Unit != "" && left.Unit != right.Unit {
		// Convert right to left's unit
		convertedRight, err := p.convertUnit(right, left.Unit)
		if err != nil {
			return Value{}, err
		}
		right = convertedRight
	}

	var result Value
	result.Unit = right.Unit
	result.StartPos = left.StartPos
	result.EndPos = right.EndPos

	switch op {
	case '+':
		result.Val = left.Val + right.Val
	case '-':
		result.Val = left.Val - right.Val
	case '*':
		result.Val = left.Val * right.Val
	case '/':
		if right.Val == 0 {
			return Value{}, fmt.Errorf("division by zero")
		}
		result.Val = left.Val / right.Val
	default:
		return Value{}, fmt.Errorf("unknown operator: %c", op)
	}

	return result, nil
}

func (p *ExpressionParser) convertUnit(value Value, targetUnit string) (Value, error) {
	if value.Unit == targetUnit {
		return value, nil
	}

	// Convert to pixels first
	pixels, err := p.unitConverter.ToPixels(value.Val, value.Unit)
	if err != nil {
		return Value{}, err
	}

	// Then convert from pixels to target unit
	result, err := p.unitConverter.FromPixels(pixels, targetUnit)
	if err != nil {
		return Value{}, err
	}

	return Value{Val: result, Unit: targetUnit, StartPos: value.StartPos, EndPos: value.EndPos}, nil
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
