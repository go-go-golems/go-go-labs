# Layout Guide DSL Specification

This document specifies a YAML-based Domain-Specific Language (DSL) for describing layout guides for printable pages or booklets.

## Page Structure

```yaml
page:
  width: <dimension>
  height: <dimension>
  subpages:
    rows: <integer>
    columns: <integer>
  margins:
    top: <dimension>
    right: <dimension>
    bottom: <dimension>
    left: <dimension>
  guides: [<guide>...]

subpage:
  margins:
    top: <dimension>
    right: <dimension>
    bottom: <dimension>
    left: <dimension>
  columns:
    count: <integer>
    gutter: <dimension>
  guides: [<guide>...]
```

- `<dimension>`: A measurement (e.g., "210mm", "8.5in") or an arithmetic expression (e.g., "(210 / 2)mm", "(8.5 + 0.5)in")
- `<integer>`: A whole number

## Guide Types

### Horizontal or Vertical Guide

```yaml
- type: <"horizontal" | "vertical">
  position: <dimension | percentage | arithmetic_expression>
  from: <"top" | "bottom" | "left" | "right">
  reference: <"page" | "margin">
  gutter: <dimension>
```

### Rectangular Guide

```yaml
- type: rect
  x: <dimension>
  y: <dimension>
  width: <dimension>
  height: <dimension>
  gutter: <dimension>
```

## Arithmetic Expressions

Arithmetic expressions can be used in place of simple dimensions when more complex calculations are needed. They must be enclosed in parentheses and can include basic arithmetic operations (+, -, *, /, ^, %). For example:

- `"(210 / 2)mm"`
- `"(8.5 + 0.5)in"`
- `"(20 * 3 + 5)pt"`
- `"(100 - 20)%"`

## Examples

### Basic Single Page Layout with Arithmetic

```yaml
page:
  width: 220mm
  height: 297mm
  margins:
    top: 22mm
    right: 18mm
    bottom: 22mm
    left: 18mm
  guides:
    - type: horizontal
      position: 52%
      from: top
      reference: margin
      gutter: 3mm
```

### Multi-column Subpage Layout with Arithmetic

```yaml
page:
  width: 215mm
  height: 297mm
  subpages:
    rows: 2
    columns: 1
  margins:
    top: 11mm
    right: 11mm
    bottom: 11mm
    left: 11mm

subpage:
  margins:
    top: 6mm
    right: 6mm
    bottom: 6mm
    left: 6mm
  columns:
    count: 3
    gutter: 6mm
  guides:
    - type: vertical
      position: (100 / 3)mm
      from: left
      reference: margin
      gutter: 1.5mm
    - type: vertical
      position: (200 / 3) mm
      from: left
      reference: margin
      gutter: 1.5mm
```

### Complex Layout with Multiple Guide Types and Arithmetic

```yaml
page:
  width: (11 + 0.5)in
  height: 8.5in
  margins:
    top: 0.55in
    right: 0.825in
    bottom: 0.55in
    left: 0.825in
  guides:
    - type: horizontal
      position: 1.1in
      from: top
      reference: page
      gutter: 0.12in
    - type: vertical
      position: 52%
      from: left
      reference: page
      gutter: 0.12in
    - type: rect
      x: 1.1in
      y: 1.6in
      width: 4.2in
      height: 3.15in
      gutter: 0.22in

subpage:
  margins:
    top: 0.275in
    right: 0.275in
    bottom: 0.275in
    left: 0.275in
  columns:
    count: 2
    gutter: 0.33in
  guides:
    - type: horizontal
      position: 2.1in
      from: bottom
      reference: margin
      gutter: 0.06in
```

This specification provides a flexible way to define layout guides for various print designs, from simple single-page layouts to complex multi-page booklets with multiple columns and custom guide placements. Arithmetic expressions are used sparingly, only when they add value to the layout definitions.
