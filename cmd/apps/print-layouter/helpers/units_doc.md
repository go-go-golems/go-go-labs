# Unit Converter Documentation

The Unit Converter is a utility for converting between different units of measurement, primarily focused on print and digital layout units. This documentation covers the `ExpressionParser` struct and the `Distance` type, which are the main components of the public API.

## ExpressionParser

The `ExpressionParser` is responsible for parsing and evaluating unit expressions. It supports basic arithmetic operations and various units.

### Usage

To use the `ExpressionParser`, create an instance with the desired PPI (Pixels Per Inch) and call the `Parse` method with a unit expression string:

```go
import "github.com/go-go-golems/go-go-labs/pkg/zinelayout/parser"

parser := &ExpressionParser{PPI: 96}
result, err := parser.Parse("1in + 2.54cm")
if err != nil {
    // Handle error
}
fmt.Printf("Result: %.2f pixels\n", result)
```

### Public Methods

#### Parse(input string) (float64, error)

Parses and evaluates the input expression, returning the result in pixels.

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

### Usage

The `Distance` type can be used to represent and convert between different units:

```go
import "github.com/go-go-golems/go-go-labs/pkg/zinelayout/parser"

d := NewDistance(100) // 100 pixels
jsonData, _ := json.Marshal(d)
fmt.Println(string(jsonData)) // Output: 100

var d2 Distance
json.Unmarshal([]byte(`"2in"`), &d2)
fmt.Printf("%.2f pixels\n", d2.Pixels()) // Output: 192.00 pixels
```

### Public Methods

- `NewDistance(pixels float64) Distance`: Creates a new Distance from a float64 value (assumed to be in pixels).
- `Pixels() float64`: Returns the Distance value in pixels.
- `MarshalJSON() ([]byte, error)`: Implements the json.Marshaler interface.
- `UnmarshalJSON(data []byte) error`: Implements the json.Unmarshaler interface.

## Error Handling

The `Parse` method of `ExpressionParser` returns an error if there are issues with the input expression. Developers should always check the error returned and handle it appropriately. Common errors include:

- Invalid syntax
- Unknown units
- Division by zero

## Best Practices

1. Always set the correct PPI when initializing the `ExpressionParser` to ensure accurate conversions.
2. Use parentheses to group operations in complex expressions for clarity and to ensure correct order of operations.
3. Handle errors returned by the `Parse` method to provide meaningful feedback to users.
4. When working with `Distance` values, use the `Pixels()` method to get the standardized pixel value for comparisons or further calculations.
