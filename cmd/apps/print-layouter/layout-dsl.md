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

- `<dimension>`: A measurement (e.g., "210mm", "8.5in")
- `<integer>`: A whole number

## Guide Types

### Horizontal or Vertical Guide

```yaml
- type: <"horizontal" | "vertical">
  position: <dimension | percentage>
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

## Examples

### Basic Single Page Layout

```yaml
page:
  width: 210mm
  height: 297mm
  margins:
    top: 20mm
    right: 15mm
    bottom: 20mm
    left: 15mm
  guides:
    - type: horizontal
      position: 50%
      from: top
      reference: margin
      gutter: 2mm
```

### Multi-column Subpage Layout

```yaml
page:
  width: 210mm
  height: 297mm
  subpages:
    rows: 2
    columns: 1
  margins:
    top: 10mm
    right: 10mm
    bottom: 10mm
    left: 10mm

subpage:
  margins:
    top: 5mm
    right: 5mm
    bottom: 5mm
    left: 5mm
  columns:
    count: 3
    gutter: 5mm
  guides:
    - type: vertical
      position: 33.33%
      from: left
      reference: margin
      gutter: 1mm
    - type: vertical
      position: 66.66%
      from: left
      reference: margin
      gutter: 1mm
```

### Complex Layout with Multiple Guide Types

```yaml
page:
  width: 11in
  height: 8.5in
  margins:
    top: 0.5in
    right: 0.75in
    bottom: 0.5in
    left: 0.75in
  guides:
    - type: horizontal
      position: 1in
      from: top
      reference: page
      gutter: 0.1in
    - type: vertical
      position: 50%
      from: left
      reference: page
      gutter: 0.1in
    - type: rect
      x: 1in
      y: 1.5in
      width: 4in
      height: 3in
      gutter: 0.2in

subpage:
  margins:
    top: 0.25in
    right: 0.25in
    bottom: 0.25in
    left: 0.25in
  columns:
    count: 2
    gutter: 0.3in
  guides:
    - type: horizontal
      position: 2in
      from: bottom
      reference: margin
      gutter: 0.05in
```

This specification provides a flexible way to define layout guides for various print designs, from simple single-page layouts to complex multi-page booklets with multiple columns and custom guide placements.