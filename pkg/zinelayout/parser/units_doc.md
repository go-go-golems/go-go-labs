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

### Enhanced Unit Handling

The `ExpressionParser` now supports more flexible unit handling:

- Units can follow any term or subexpression, not just numbers.
- Conversion to pixels happens when a unit is encountered.
- Examples of valid expressions:
  - `(1 + 2) in`
  - `1/12 in`
  - `1 px + 2 in`
  - `(1 in + 2 cm) * 3 px`

This allows for more intuitive and flexible expressions when working with different units.

## Distance

(The Distance section remains unchanged)

## Error Handling

The `Parse` method of `ExpressionParser` returns an error if there are issues with the input expression. Developers should always check the error returned and handle it appropriately. Common errors include:

- Invalid syntax
- Unknown units
- Division by zero
- Missing closing parenthesis

## Best Practices

1. Always set the correct PPI when initializing the `ExpressionParser` to ensure accurate conversions.
2. Use parentheses to group operations in complex expressions for clarity and to ensure correct order of operations.
3. Handle errors returned by the `Parse` method to provide meaningful feedback to users.
4. When working with `Distance` values, use the `Pixels()` method to get the standardized pixel value for comparisons or further calculations.
5. Take advantage of the flexible unit handling to create more intuitive expressions, such as `1/2 in` or `(1 + 2) cm`.
6. Be aware that units can be applied to entire subexpressions, not just individual numbers.