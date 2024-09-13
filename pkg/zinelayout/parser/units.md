# Unit Expression Grammar and Parser

## Grammar

The grammar for parsing arithmetic expressions on distance units is defined as follows:

```
Expression := Term {('+' | '-') Term}

Term := Factor {('*' | '/') Factor}

Factor := Number Unit | '(' Expression ')' | '-' Factor

Number := Digit+ ['.' Digit+]

Digit := '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9'

Unit := 'mm' | 'cm' | 'in' | 'pc' | 'pt' | 'px' | 'em' | 'rem' |
```

This grammar allows for:
- Basic arithmetic operations: addition, subtraction, multiplication, and division
- Parentheses for grouping
- Negative numbers
- Decimal numbers
- Various distance units

## Recursive Descent Parser

The recursive descent parser implemented for this grammar works by creating a set of mutually recursive functions, each corresponding to a non-terminal symbol in the grammar. Here's a detailed explanation of how each function in the parser works:

### 1. parseExpression

This function handles the `Expression` rule in the grammar. It:
1. Calls `parseTerm` to get the first term.
2. Enters a loop to handle any additional terms connected by '+' or '-'.
3. For each additional term, it calls `parseTerm` again and performs the addition or subtraction.

### 2. parseTerm

This function handles the `Term` rule. It:
1. Calls `parseFactor` to get the first factor.
2. Enters a loop to handle any additional factors connected by '*' or '/'.
3. For each additional factor, it calls `parseFactor` again and performs the multiplication or division.

### 3. parseFactor

This function handles the `Factor` rule. It:
1. Checks for parentheses. If found, it recursively calls `parseExpression` for the contents inside the parentheses.
2. Checks for a negative sign. If found, it recursively calls `parseFactor` and negates the result.
3. If neither of the above, it calls `parseNumberUnit` to handle a number with a unit.

### 4. parseNumberUnit

This function handles the `Number` and `Unit` rules. It:
1. Parses a sequence of digits and an optional decimal point to form a number.
2. Parses any following letters to identify the unit.
3. Calls `convertToPixels` to convert the value to pixels based on the unit and the configured PPI (pixels per inch).

### Helper Functions

- `currentChar`: Returns the current character being processed.
- `skipWhitespace`: Skips any whitespace characters.
- `convertToPixels`: Converts a value from a given unit to pixels.

### Parsing Process

1. The parsing starts with `parseExpression`.
2. As it encounters different parts of the input, it calls the appropriate functions.
3. The parser maintains a current position (`pos`) in the input string, which is advanced as characters are consumed.
4. Each function returns a float64 value representing the result of its parsing and evaluation.

### Error Handling

The parser includes error handling at various stages:
- Invalid numbers
- Unknown units
- Missing closing parentheses
- Division by zero

When an error is encountered, it's immediately returned, halting the parsing process.

### Integration with UnitConverter

The `ExpressionParser` is integrated into the `UnitConverter` struct via the `ToPixels` method. This method creates a new `ExpressionParser` instance with the configured PPI and uses it to parse and evaluate the input expression.

This recursive descent parser provides a flexible and powerful way to handle complex unit expressions while maintaining a clear and maintainable structure that closely mirrors the grammar definition.