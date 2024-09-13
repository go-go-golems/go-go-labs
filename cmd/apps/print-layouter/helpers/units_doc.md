# Unit Converter Documentation

The `UnitConverter` is a utility for converting between different units of measurement, primarily focused on print and digital layout units. This documentation covers the `ExpressionParser` struct, the `Distance` type, and their associated methods.

## ExpressionParser

The `ExpressionParser` is responsible for parsing and evaluating unit expressions. It supports basic arithmetic operations and various units.

### Fields

- `input string`: The input expression to be parsed.
- `pos int`: The current position in the input string during parsing.
- `PPI float64`: Pixels Per Inch, used as the base for conversions.

### Methods

#### Parse(input string) (float64, error)

Parses and evaluates the input expression, returning the result in pixels.

#### parseExpression() (float64, error)

Parses addition and subtraction operations.

#### parsePower() (float64, error)

Parses exponentiation operations.

#### parseTerm() (float64, error)

Parses multiplication and division operations.

#### parseFactor() (float64, error)

Parses parentheses, negative numbers, and number-unit combinations.

#### parseNumberUnit() (float64, error)

Parses a number followed by an optional unit.

#### convertToPixels(value float64, unit string) (float64, error)

Converts a value from the given unit to pixels.

### Supported Units

- mm (millimeters)
- cm (centimeters)
- in (inches)
- pc (picas)
- pt (points)
- px (pixels)
- em (relative to font size)
- rem (relative to root font size)

### Supported Operations

- Addition (+)
- Subtraction (-)
- Multiplication (*)
- Division (/)
- Exponentiation (^)
- Parentheses for grouping
- Negative numbers

## Distance

The `Distance` type represents a length that can be expressed in various units.

### Methods

- `NewDistance(pixels float64) Distance`: Creates a new Distance from a float64 value (assumed to be in pixels).
- `Pixels() float64`: Returns the Distance value in pixels.
- `MarshalJSON() ([]byte, error)`: Implements the json.Marshaler interface.
- `UnmarshalJSON(data []byte) error`: Implements the json.Unmarshaler interface.

## Usage

To use the `ExpressionParser`, create an instance with the desired PPI and call the `Parse` method with a unit expression string:

```go
parser := &ExpressionParser{PPI: 96}
result, err := parser.Parse("1in + 2.54cm")
if err != nil {
    // Handle error
}
fmt.Printf("Result: %.2f pixels\n", result)
```

The `Distance` type can be used to represent and convert between different units:

```go
d := NewDistance(100) // 100 pixels
jsonData, _ := json.Marshal(d)
fmt.Println(string(jsonData)) // Output: 100

var d2 Distance
json.Unmarshal([]byte(`"2in"`), &d2)
fmt.Printf("%.2f pixels\n", d2.Pixels()) // Output: 192.00 pixels
```

## Testing

The `parser_test.go` file contains extensive tests for the `ExpressionParser`, covering various scenarios including:

- Simple unit conversions
- Complex expressions with multiple operations and units
- Whitespace handling
- Unit case variations
- Edge cases (e.g., very small or large values, leading/trailing decimals)
- Error cases (e.g., invalid syntax, unknown units, division by zero)

These tests ensure the robustness and accuracy of the parsing and conversion functionality.
