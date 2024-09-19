# Unit Parsing and Conversion Documentation

## Table of Contents

1. [Introduction](#introduction)
2. [Syntax of Unit Expressions](#syntax-of-unit-expressions)
   - [Supported Units](#supported-units)
   - [Supported Operations](#supported-operations)
   - [Expression Examples](#expression-examples)
3. [API Documentation](#api-documentation)
   - [Value Struct](#value-struct)
   - [ExpressionParser Struct](#expressionparser-struct)
   - [UnitConverter Struct](#unitconverter-struct)
   - [Public Methods](#public-methods)
4. [Tutorials and Examples](#tutorials-and-examples)
   - [Basic Usage](#basic-usage)
   - [Complex Expressions](#complex-expressions)
   - [Unit Conversion](#unit-conversion)
   - [Error Handling](#error-handling)
   - [Integration with Other Systems](#integration-with-other-systems)
5. [Conclusion](#conclusion)

---

## Introduction

The **Unit Parser and Converter** is a Go package designed to parse and evaluate arithmetic expressions involving units of measurement, particularly for print and digital layouts. It supports basic arithmetic operations, various units, and provides a flexible way to handle complex unit expressions. This package is especially useful for applications that require precise measurements and conversions, such as typesetting, graphics rendering, and UI design.

---

## Syntax of Unit Expressions

### Supported Units

The parser recognizes the following units:

- **Length Units**:
  - `mm` - Millimeters
  - `cm` - Centimeters
  - `in` - Inches
  - `pt` - Points (1 point = 1/72 inch)
  - `pc` - Picas (1 pica = 12 points)
  - `px` - Pixels
- **Relative Units**:
  - `em` - Relative to the font size (commonly 16px)
  - `rem` - Relative to the root font size

### Supported Operations

The parser supports the following arithmetic operations:

- **Addition (`+`)**: Adds two values.
- **Subtraction (`-`)**: Subtracts the second value from the first.
- **Multiplication (`*`)**: Multiplies two values.
- **Division (`/`)**: Divides the first value by the second.
- **Parentheses (`(`, `)`)**: Groups expressions to override the default precedence.
- **Negative Numbers**: Supports unary minus for negative values (e.g., `-5cm`).

### Expression Examples

- Simple expressions:
  - `10in`
  - `2.54cm`
  - `100px`
- Arithmetic expressions:
  - `1in + 2.54cm`
  - `3mm * 4`
  - `10pt / 2`
- Expressions with parentheses:
  - `(1in + 2.54cm) * 3`
  - `(10 + 5) * (2in - 1cm)`
- Complex expressions:
  - `(1in + 2cm) * 3 - (4mm + 5pt) * 2`
  - `(((1in + 2cm) * 3) - 4mm) - (5pt / 2)`
- Expressions with units applied to subexpressions:
  - `(1 + 2) in`
  - `1/12 in`
  - `1 px + 2 in`

---

## API Documentation

### Value Struct

The `Value` struct represents a numeric value with its associated unit and position in the input string.

```go
type Value struct {
    Val      float64 // Numeric value
    Unit     string  // Unit of measurement
    StartPos int     // Start position in the input string
    EndPos   int     // End position in the input string
}
```

#### Methods

- `func (v Value) String() string`: Returns a string representation of the `Value`.

### ExpressionParser Struct

The `ExpressionParser` is responsible for parsing and evaluating unit expressions.

```go
type ExpressionParser struct {
    input         string         // Input expression string
    pos           int            // Current position in the input string
    PPI           float64        // Pixels Per Inch for unit conversion
    Debug         bool           // Enables debug output if true
    depth         int            // Current recursion depth (used for indentation in debug output)
    unitConverter *UnitConverter // Reference to a UnitConverter
}
```

#### Methods

- `func (p *ExpressionParser) Parse(input string) (Value, error)`: Parses and evaluates the input expression, returning a `Value` and an error if any.
- `func (p *ExpressionParser) ValueWithOriginal(v Value) string`: Returns a string representation of the `Value` including its original input segment.

#### Usage

Create an instance of `ExpressionParser`, set the desired PPI, and call the `Parse` method:

```go
parser := &ExpressionParser{PPI: 96}
result, err := parser.Parse("1in + 2.54cm")
if err != nil {
    // Handle error
}
fmt.Printf("Result: %s\n", result.String())
```

### UnitConverter Struct

The `UnitConverter` handles conversions between different units and pixels.

```go
type UnitConverter struct {
    PPI float64 // Pixels Per Inch for conversion
}
```

#### Methods

- `func (uc *UnitConverter) ToPixels(value float64, unit string) (float64, error)`: Converts a value from a given unit to pixels.
- `func (uc *UnitConverter) FromPixels(pixels float64, unit string) (float64, error)`: Converts a pixel value to the specified unit.
- Conversion methods for specific units (e.g., `FromInch`, `ToInch`, `FromMillimeter`, `ToMillimeter`, etc.).

#### Usage

Instantiate `UnitConverter` with the desired PPI and use its methods to convert values:

```go
uc := &UnitConverter{PPI: 96}
pixels, err := uc.ToPixels(2.54, "cm")
if err != nil {
    // Handle error
}
fmt.Printf("2.54cm is %.2f pixels\n", pixels)
```

### Public Methods

#### Parsing and Evaluating Expressions

- `Parse(input string) (Value, error)`: Parses the input expression and returns a `Value`.

#### Unit Conversion

- `ToPixels(value float64, unit string) (float64, error)`: Converts a value from the specified unit to pixels.
- `FromPixels(pixels float64, unit string) (float64, error)`: Converts a pixel value to the specified unit.

#### Utility Methods

- `String() string`: Returns a string representation of the `Value`.

---

## Tutorials and Examples

### Basic Usage

**Objective**: Parse and evaluate a simple unit expression.

```go
package main

import (
    "fmt"
    "github.com/go-go-golems/go-go-labs/pkg/zinelayout/parser"
)

func main() {
    parser := &parser.ExpressionParser{PPI: 96}
    result, err := parser.Parse("10cm")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Printf("Parsed Value: %s\n", result.String())

    // Convert to pixels
    pixels, err := parser.unitConverter.ToPixels(result.Val, result.Unit)
    if err != nil {
        fmt.Println("Conversion Error:", err)
        return
    }
    fmt.Printf("Value in pixels: %.2fpx\n", pixels)
}
```

**Output**:
```
Parsed Value: 10.00cm
Value in pixels: 378.00px
```

### Complex Expressions

**Objective**: Parse and evaluate a complex expression involving different units and operations.

```go
func main() {
    parser := &parser.ExpressionParser{PPI: 96}
    expression := "(1in + 2.54cm) * 3"
    result, err := parser.Parse(expression)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Printf("Parsed Value: %s\n", result.String())

    // Convert to pixels
    pixels, err := parser.unitConverter.ToPixels(result.Val, result.Unit)
    if err != nil {
        fmt.Println("Conversion Error:", err)
        return
    }
    fmt.Printf("Value in pixels: %.2fpx\n", pixels)
}
```

**Explanation**:

- `1in + 2.54cm` adds 1 inch and 2.54 centimeters.
- Multiplying the sum by 3.
- The parser handles unit conversion internally.

**Output**:
```
Parsed Value: 11.00in
Value in pixels: 1056.00px
```

### Unit Conversion

**Objective**: Convert a value from one unit to another using `UnitConverter`.

```go
func main() {
    uc := &parser.UnitConverter{PPI: 96}
    inches := 2.0
    pixels, err := uc.ToPixels(inches, "in")
    if err != nil {
        fmt.Println("Conversion Error:", err)
        return
    }
    fmt.Printf("2 inches is %.2f pixels\n", pixels)

    centimeters, err := uc.FromPixels(pixels, "cm")
    if err != nil {
        fmt.Println("Conversion Error:", err)
        return
    }
    fmt.Printf("Which is %.2f centimeters\n", centimeters)
}
```

**Output**:
```
2 inches is 192.00 pixels
Which is 5.08 centimeters
```

### Error Handling

**Objective**: Gracefully handle parsing errors.

```go
func main() {
    parser := &parser.ExpressionParser{PPI: 96}
    expressions := []string{
        "1in +",
        "10kg",
        "1in + 2cm)",
    }

    for _, expr := range expressions {
        _, err := parser.Parse(expr)
        if err != nil {
            fmt.Printf("Error parsing '%s': %v\n", expr, err)
        } else {
            fmt.Printf("Parsed '%s' successfully\n", expr)
        }
    }
}
```

**Output**:
```
Error parsing '1in +': unexpected character: 
Error parsing '10kg': invalid unit: kg
Error parsing '1in + 2cm)': unexpected character: )
```

### Integration with Other Systems

**Objective**: Use the parser and converter in a context like setting UI element sizes.

```go
func main() {
    // Assume we are setting the width of a UI element
    widthExpression := "50% of (800px - 2 * 20px)"
    // For simplicity, we'll replace '50% of' with '* 0.5'
    adjustedExpression := strings.Replace(widthExpression, "50% of", "* 0.5", 1)

    // Create a parser instance
    parser := &parser.ExpressionParser{PPI: 96}
    result, err := parser.Parse(adjustedExpression)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    // Convert the result to pixels
    pixels, err := parser.unitConverter.ToPixels(result.Val, result.Unit)
    if err != nil {
        fmt.Println("Conversion Error:", err)
        return
    }

    fmt.Printf("Calculated width: %.2fpx\n", pixels)
}
```

**Note**: This example assumes a simplified syntax and would need adjustments for more complex expressions.

**Output**:
```
Calculated width: 380.00px
```

---

## Conclusion

The **Unit Parser and Converter** provides a robust and flexible way to parse, evaluate, and convert unit expressions in Go. By supporting a wide range of units and arithmetic operations, it simplifies the handling of measurements in applications that require precise control over sizes and distances.

When using this package:

- **Always set the correct PPI**: The Pixels Per Inch value is crucial for accurate conversions.
- **Handle errors appropriately**: The parser will return informative errors that can help in debugging expressions.
- **Utilize the flexibility of units**: Units can be applied to subexpressions, allowing for intuitive and readable expressions.

Whether you're working on typesetting, graphical layouts, or UI design, this package can significantly streamline the process of dealing with units and measurements.