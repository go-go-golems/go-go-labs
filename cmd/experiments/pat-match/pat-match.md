Porting paip pattern matching to go.

Chat: https://chatgpt.com/c/6748d336-31a4-8012-9333-0b5b9606b876

# Pat-Match: A Go Pattern Matching Library

## Overview

**Pat-Match** is a Go library that implements a powerful and extensible pattern matching system inspired by the pattern matcher from Peter Norvig's *Paradigms of Artificial Intelligence Programming* (PAIP). It allows developers to define complex patterns using a concise Domain-Specific Language (DSL) and match them against inputs to extract variable bindings. The library supports variables, constants, sequences, segment patterns, and special patterns like `?is`, `?and`, `?or`, `?not`, and `?if`.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Getting Started](#getting-started)
  - [Basic Patterns](#basic-patterns)
  - [Variables and Constants](#variables-and-constants)
  - [Sequences](#sequences)
- [Advanced Patterns](#advanced-patterns)
  - [Segment Patterns](#segment-patterns)
  - [Single Patterns](#single-patterns)
  - [Custom Predicates](#custom-predicates)
- [Pattern Construction DSL](#pattern-construction-dsl)
- [Examples](#examples)
  - [Example 1: Variable Matching](#example-1-variable-matching)
  - [Example 2: Segment Matching](#example-2-segment-matching)
  - [Example 3: Conditional Matching with ?if](#example-3-conditional-matching-with-if)
- [Unit Testing](#unit-testing)
- [Extensibility](#extensibility)
- [Contributing](#contributing)
- [License](#license)

---

## Features

- **Variable Matching**: Match variables and bind them to input values.
- **Constant Matching**: Match specific constant values.
- **Sequence Matching**: Match sequences (lists) of patterns against lists of inputs.
- **Segment Patterns**: Match segments of input lists using `?*`, `?+`, and `??`.
- **Single Patterns**: Use special patterns like `?is`, `?and`, `?or`, and `?not` for advanced matching.
- **Conditional Matching**: Use `?if` patterns with custom predicates for conditional logic.
- **Pattern Construction DSL**: Build patterns using a concise and readable DSL.
- **Extensibility**: Easily extend the library with custom patterns and predicates.

---

## Getting Started

### Basic Patterns

At its core, Pat-Match allows you to define patterns and match them against inputs. The matching process attempts to bind variables in the pattern to corresponding parts of the input.

### Variables and Constants

- **Variables**: Symbols starting with `?`, e.g., `?x`, represent variables that can match any input.
- **Constants**: Specific values that must match exactly in the input.

### Sequences

Patterns can be sequences (lists) of other patterns, allowing you to match lists of inputs.

---

## Advanced Patterns

### Segment Patterns

Segment patterns allow you to match segments of the input list:

- **`?*` (Zero or more elements)**:
  ```go
  Seg("?x", restPattern, 0)
  ```
  Matches zero or more elements and binds them to `?x`.

- **`?+` (One or more elements)**:
  ```go
  Seg("?x", restPattern, 1)
  ```
  Matches one or more elements and binds them to `?x`.

- **`??` (Zero or one element)**:
  You can simulate `??` by setting the minimum to 0 and handling the maximum match in your logic.

### Single Patterns

Special patterns that match a single input element:

- **`?is`**: Matches an input if it satisfies a predicate.
  ```go
  Single("?is", Var("?n"), Const("numberp"))
  ```
- **`?and`**: Matches if all sub-patterns match.
  ```go
  Single("?and", pattern1, pattern2, ...)
  ```
- **`?or`**: Matches if any sub-pattern matches.
  ```go
  Single("?or", pattern1, pattern2, ...)
  ```
- **`?not`**: Matches if none of the sub-patterns match.
  ```go
  Single("?not", pattern)
  ```

### Custom Predicates

Use `?if` patterns with custom Go functions to perform conditional matching based on variable bindings.

```go
SingleWithPredicate("?if", func(input interface{}, bindings Bindings) bool {
    x, ok1 := bindings["?x"].(int)
    y, ok2 := bindings["?y"].(int)
    return ok1 && ok2 && x > y
})
```

---

## Pattern Construction DSL

Pat-Match provides helper functions to construct patterns in a readable and concise manner:

- **Variables**:
  ```go
  Var("?x")
  ```
- **Constants**:
  ```go
  Const("a")
  ```
- **Sequences**:
  ```go
  Seq(pattern1, pattern2, ...)
  ```
- **Segments**:
  ```go
  Seg("?x", restPattern, minElements)
  ```
- **Single Patterns**:
  ```go
  Single(operator, args...)
  ```
- **Custom Predicate Patterns**:
  ```go
  SingleWithPredicate("?if", predicateFunc)
  ```

---

## Examples

### Example 1: Variable Matching

**Pattern**: Match any input and bind it to `?x`.

```go
pattern := Var("?x")
input := "hello"
bindings := Bindings{}

result, err := pattern.Match(input, bindings)
if err != nil {
    fmt.Println("No match:", err)
} else {
    fmt.Println("Bindings:", result)
}
// Output: Bindings: map[?x:hello]
```

### Example 2: Segment Matching

**Pattern**: Match a list starting with `"a"`, followed by zero or more elements bound to `?x`, ending with `"d"`.

```go
pattern := Seq(Const("a"), Seg("?x", Seq(Const("d")), 0))
input := []interface{}{"a", "b", "c", "d"}
bindings := Bindings{}

result, err := pattern.Match(input, bindings)
if err != nil {
    fmt.Println("No match:", err)
} else {
    fmt.Println("Bindings:", result)
}
// Output: Bindings: map[?x:[b c]]
```

### Example 3: Conditional Matching with `?if`

**Pattern**: Match inputs of the form `(?x > ?y)` where `?x` and `?y` are integers and `?x` is greater than `?y`.

```go
pattern := Seq(
    Var("?x"),
    Const(">"),
    Var("?y"),
    SingleWithPredicate("?if", func(input interface{}, bindings Bindings) bool {
        x, ok1 := bindings["?x"].(int)
        y, ok2 := bindings["?y"].(int)
        return ok1 && ok2 && x > y
    }),
)
input := []interface{}{5, ">", 3}
bindings := Bindings{}

result, err := pattern.Match(input, bindings)
if err != nil {
    fmt.Println("No match:", err)
} else {
    fmt.Println("Bindings:", result)
}
// Output: Bindings: map[?x:5 ?y:3]
```

---

## Unit Testing

The library includes a comprehensive suite of unit tests covering various pattern matching scenarios:

- Variable and constant matching
- Sequence matching
- Segment patterns (`?*`, `?+`, `??`)
- Single patterns (`?is`, `?and`, `?or`, `?not`)
- Conditional patterns with `?if`
- Nested patterns
- Failure cases and edge conditions

To run the tests, execute:

```bash
go test
```

---

## Extensibility

Pat-Match is designed to be extensible:

- **Custom Patterns**: Implement the `Pattern` interface to create new pattern types.
- **Custom Predicates**: Add new predicates to the `predicateFuncs` map or use `SingleWithPredicate` for inline predicates.
- **DSL Enhancements**: Extend the DSL helper functions to include additional pattern constructors.

---

## Additional Details

### Data Structures

- **Pattern Interface**: Central to the library, all pattern types implement the `Match` method.
- **Bindings**: A map from variable names to their matched values, allowing for variable binding across the pattern.

### Pattern Types

1. **VariablePattern**: Matches variables (e.g., `?x`) and binds them to input values.
2. **ConstantPattern**: Matches constants, ensuring the input equals a specific value.
3. **ListPattern**: Matches a list of patterns against a list of inputs.
4. **SegmentPattern**: Matches a segment of the input list, handling patterns like `(?* ?x)`.
5. **SinglePattern**: Handles special patterns (`?is`, `?and`, `?or`, `?not`, `?if`).

### Helper Functions

- **Var(name string)**: Creates a variable pattern.
- **Const(value interface{})**: Creates a constant pattern.
- **Seq(patterns ...Pattern)**: Creates a sequence pattern.
- **Seg(varName string, rest Pattern, min int)**: Creates a segment pattern with a minimum match length.
- **Single(operator string, args ...Pattern)**: Creates a single pattern with specified operator and arguments.
- **SingleWithPredicate(operator string, predicate func(input interface{}, bindings Bindings) bool)**: Creates a single pattern with a custom predicate function.

### Predicates

Predicates are functions that determine if an input satisfies certain conditions. The library includes predefined predicates:

- **`numberp`**: Checks if the input is a number (`int` or `float64`).
- **`oddp`**: Checks if an integer input is odd.

You can add more predicates by updating the `predicateFuncs` map.

### Error Handling

The `Match` methods return an error when a pattern does not match the input. This allows you to handle matching failures gracefully.

### Best Practices

- **Immutable Bindings**: The `copyBindings` function ensures bindings are not mutated during matching, preserving the integrity of previous bindings.
- **Type Assertions**: When working with variable bindings, use type assertions carefully to avoid panics.
