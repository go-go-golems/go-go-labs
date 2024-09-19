# Zine Layout DSL Guide

## Introduction

The **Zine Layout Domain-Specific Language (DSL)** is a powerful and flexible YAML-based language designed to help you create complex layouts for zines, booklets, and other printed materials. It allows you to specify how input pages (content) are arranged on output pages (print sheets), including precise control over positions, rotations, margins, borders, and units of measurement.

This guide will provide an in-depth explanation of how the layout files work, a comprehensive description of the units calculation syntax, and a step-by-step tutorial on writing a layout from scratch. We'll also showcase examples of what's possible with the Zine Layout DSL.

---

## Table of Contents

1. [Understanding the Layout File Structure](#1-understanding-the-layout-file-structure)
2. [Units Calculation Syntax](#2-units-calculation-syntax)
3. [Writing a Layout from Scratch](#3-writing-a-layout-from-scratch)
   - [Step 1: Define Global Settings](#step-1-define-global-settings)
   - [Step 2: Configure Page Setup](#step-2-configure-page-setup)
   - [Step 3: Specify Output Pages](#step-3-specify-output-pages)
   - [Step 4: Putting It All Together](#step-4-putting-it-all-together)
4. [Examples of Possible Layouts](#4-examples-of-possible-layouts)
   - [Example 1: Simple Two-Page Spread](#example-1-simple-two-page-spread)
   - [Example 2: Complex Booklet with Custom Margins and Rotations](#example-2-complex-booklet-with-custom-margins-and-rotations)
5. [Conclusion](#5-conclusion)
6. [Appendix](#6-appendix)
   - [Complete Syntax Reference](#complete-syntax-reference)
   - [Common Units and Conversions](#common-units-and-conversions)

---

## 1. Understanding the Layout File Structure

A Zine Layout DSL file is a YAML document that defines how input images (pages) are arranged on output pages. The structure is divided into several key sections:

- **Global Settings**: General settings that apply to the entire document.
- **Page Setup**: Configuration for the overall page layout, such as grid size and margins.
- **Output Pages**: Detailed specifications for each output page, including the placement of input images.

Here's a high-level overview of the YAML structure:

```yaml
global:
  # Global settings

page_setup:
  # Page configuration

output_pages:
  # List of output page specifications
```

Let's delve into each section in detail.

### Global Settings

The `global` section contains settings that affect the entire document:

- **Border**: Defines a global border around the entire output image.
- **PPI**: Pixels per inch, used for unit conversions (required, normally set to 300)

Example:

```yaml
global:
  ppi: 300
  border:
    enabled: true
    color: black
    type: dotted
```

### Page Setup

The `page_setup` section configures the overall layout of the pages:

- **Grid Size**: Defines how many rows and columns the output page grid has.
- **Margin**: Sets default margins for all output pages.
- **Page Border**: Specifies a border around each output page.

Example:

```yaml
page_setup:
  grid_size:
    rows: 2
    columns: 2
  margin:
    top: 0.5in
    bottom: 0.5in
    left: 0.5in
    right: 0.5in
  border:
    enabled: true
    color: red
    type: plain
```

### Output Pages

The `output_pages` section is a list of output page specifications:

- **ID**: A unique identifier for the output page.
- **Margin**: Overrides default margins for this output page.
- **Layout**: Defines how input images are placed on this output page.
- **Layout Border**: Border around the layout area.

Each layout item within an output page includes:

- **Input Index**: The index (1-based) of the input image to place.
- **Position**: The row and column in the grid where the image is placed.
- **Rotation**: Rotation angle (0, 90, 180, 270 degrees).
- **Margin**: Margins specific to this input image.
- **Inner Layout Border**: Border around the input image area.

Example:

```yaml
output_pages:
  - id: page1
    margin:
      top: 0.25in
      bottom: 0.25in
      left: 0.25in
      right: 0.25in
    layout:
      - input_index: 1
        position:
          row: 0
          column: 0
        rotation: 0
        margin:
          top: 10
          bottom: 10
          left: 10
          right: 10
        border:
          enabled: true
          color: blue
          type: dashed
```

---

## 2. Units Calculation Syntax

The Zine Layout DSL allows you to specify sizes and margins using various units of measurement, along with arithmetic expressions. This provides flexibility and precision in defining your layout.

### Supported Units

- **Pixels (`px`)**: Default unit if none is specified.
- **Inches (`in`)**
- **Centimeters (`cm`)**
- **Millimeters (`mm`)**
- **Points (`pt`)**
- **Picas (`pc`)**
- **Ems (`em`)**: Relative to font size (assumed to be 16px at 96 PPI).
- **Rems (`rem`)**: Relative to root font size (same as `em` in this context).

### Arithmetic Expressions

You can use arithmetic expressions to calculate values:

- **Operators**: `+`, `-`, `*`, `/`
- **Parentheses**: `(`, `)` for grouping
- **Example**: `1in + 5mm`, `(10 + 5) * 2px`

### Examples

- **Simple Margin**: `0.5in`
- **Calculated Margin**: `1/8 in` (equivalent to `0.125in`)
- **Combined Units**: `10mm + 0.5in`

### Unit Conversion

The DSL automatically converts units to pixels using the specified PPI (Pixels Per Inch). The PPI is required and normally set to 300.

**Conversion Formulas**:

- **Inches to Pixels**: `pixels = inches * PPI`
- **Centimeters to Pixels**: `pixels = (centimeters / 2.54) * PPI`
- **Millimeters to Pixels**: `pixels = (millimeters / 25.4) * PPI`
- **Points to Pixels**: `pixels = (points / 72) * PPI`
- **Picas to Pixels**: `pixels = (picas / 6) * PPI`
- **Ems to Pixels**: `pixels = ems * 16 * (PPI / 96)`

---

## 3. Writing a Layout from Scratch

Let's walk through creating a layout file step by step.

### Step 1: Define Global Settings

Start by setting global configurations that apply to the entire document.

```yaml
global:
  ppi: 300  # Higher PPI for print quality (required)
  border:
    enabled: true
    color: black
    type: dotted
```

### Step 2: Configure Page Setup

Set up the overall page layout.

```yaml
page_setup:
  grid_size:
    rows: 2
    columns: 2
  margin:
    top: 0.5in
    bottom: 0.5in
    left: 0.5in
    right: 0.5in
  border:
    enabled: true
    color: red
    type: plain
```

- **Grid Size**: 2 rows and 2 columns.
- **Margins**: Half an inch on all sides.
- **Page Border**: Plain red border around each output page.

### Step 3: Specify Output Pages

Define how input images are placed on each output page.

```yaml
output_pages:
  - id: page1
    layout:
      - input_index: 1
        position:
          row: 0
          column: 0
        rotation: 0
      - input_index: 2
        position:
          row: 0
          column: 1
        rotation: 0
      - input_index: 3
        position:
          row: 1
          column: 0
        rotation: 0
      - input_index: 4
        position:
          row: 1
          column: 1
        rotation: 0
```

- **Input Images**: Images 1 to 4 are placed on the first output page.
- **Positions**: Placed according to the grid coordinates.

### Step 4: Putting It All Together

Combine all sections into one YAML file.

```yaml
global:
  ppi: 300
  border:
    enabled: true
    color: black
    type: dotted

page_setup:
  grid_size:
    rows: 2
    columns: 2
  margin:
    top: 0.5in
    bottom: 0.5in
    left: 0.5in
    right: 0.5in
  border:
    enabled: true
    color: red
    type: plain

output_pages:
  - id: page1
    layout:
      - input_index: 1
        position:
          row: 0
          column: 0
        rotation: 0
      - input_index: 2
        position:
          row: 0
          column: 1
        rotation: 0
      - input_index: 3
        position:
          row: 1
          column: 0
        rotation: 0
      - input_index: 4
        position:
          row: 1
          column: 1
        rotation: 0
```

---

## 4. Examples of Possible Layouts

### Example 1: Simple Two-Page Spread

**Objective**: Create a layout where two input images are placed side by side on a landscape-oriented output page, suitable for printing a booklet.

**Layout File**:

```yaml
global:
  ppi: 300

page_setup:
  grid_size:
    rows: 1
    columns: 2
  orientation: landscape
  margin:
    top: 0.25in
    bottom: 0.25in
    left: 0.25in
    right: 0.25in

output_pages:
  - id: spread1
    layout:
      - input_index: 1
        position:
          row: 0
          column: 0
        rotation: 0
      - input_index: 2
        position:
          row: 0
          column: 1
        rotation: 0
```

**Explanation**:

- **Orientation**: Landscape.
- **Grid**: 1 row, 2 columns.
- **Input Images**: Images 1 and 2 placed side by side.

**Usage**: Ideal for creating a center spread in a zine or booklet.

### Example 2: Complex Booklet with Custom Margins and Rotations

**Objective**: Design a 16-page booklet with custom margins, rotations, and borders to ensure correct page ordering when folded.

**Layout File**:

```yaml
global:
  ppi: 300
  border:
    enabled: true
    color: black
    type: corner

page_setup:
  grid_size:
    rows: 2
    columns: 2
  orientation: portrait
  margin:
    top: 0.5in
    bottom: 0.5in
    left: 0.5in
    right: 0.5in
  border:
    enabled: true
    color: gray
    type: dashed

output_pages:
  - id: page1
    layout:
      - input_index: 16
        position:
          row: 0
          column: 0
        rotation: 180
        margin:
          top: 0.1in
          bottom: 0.1in
          left: 0.1in
          right: 0.1in
      - input_index: 1
        position:
          row: 0
          column: 1
        rotation: 0
      - input_index: 2
        position:
          row: 1
          column: 0
        rotation: 180
      - input_index: 15
        position:
          row: 1
          column: 1
        rotation: 0
  - id: page2
    layout:
      - input_index: 14
        position:
          row: 0
          column: 0
        rotation: 180
      - input_index: 3
        position:
          row: 0
          column: 1
        rotation: 0
      - input_index: 4
        position:
          row: 1
          column: 0
        rotation: 180
      - input_index: 13
        position:
          row: 1
          column: 1
        rotation: 0
```

**Explanation**:

- **Orientation**: Portrait.
- **Grid**: 2 rows, 2 columns.
- **Rotations**: Applied to ensure pages are upright after folding.
- **Custom Margins**: Specific to certain input images.
- **Borders**: Corner borders to mark cutting/folding lines.

**Usage**: Suitable for creating a booklet where pages need to be in a specific order after folding and binding.

---

## 5. Conclusion

The Zine Layout DSL provides a flexible and powerful way to design complex page layouts for zines, booklets, and other print media. By understanding the structure of the layout files and utilizing the units calculation syntax, you can create precise and professional layouts tailored to your project's needs.

---

## 6. Appendix

### Complete Syntax Reference

#### Global Section

```yaml
global:
  ppi: <number>  # Pixels per inch (required, normally set to 300)
  border:
    enabled: <boolean>  # Enable or disable global border
    color: <color>      # Border color (name or hex code)
    type: <type>        # Border type ('plain', 'dotted', 'dashed', 'corner')
```

#### Page Setup Section

```yaml
page_setup:
  grid_size:
    rows: <integer>     # Number of rows in the grid
    columns: <integer>  # Number of columns in the grid
  orientation: <string> # 'portrait' or 'landscape' (optional)
  margin:
    top: <expression>    # Margin expressions (see units syntax)
    bottom: <expression>
    left: <expression>
    right: <expression>
  border:
    enabled: <boolean>
    color: <color>
    type: <type>
```

#### Output Pages Section

```yaml
output_pages:
  - id: <string>       # Unique identifier for the output page
    margin:
      top: <expression>
      bottom: <expression>
      left: <expression>
      right: <expression>
    layout_border:
      enabled: <boolean>
      color: <color>
      type: <type>
    layout:
      - input_index: <integer>  # Index of the input image (1-based)
        position:
          row: <integer>        # Row position in the grid
          column: <integer>     # Column position in the grid
        rotation: <integer>     # Rotation angle (0, 90, 180, 270)
        margin:
          top: <expression>
          bottom: <expression>
          left: <expression>
          right: <expression>
        border:
          enabled: <boolean>
          color: <color>
          type: <type>
```

### Common Units and Conversions

- **1 inch (in)** = 2.54 centimeters (cm) = 25.4 millimeters (mm) = 72 points (pt) = 6 picas (pc)
- **Pixels (px)**: Dependent on PPI (pixels per inch)
- **Points (pt)**: Commonly used in typography (1pt = 1/72 inch)
- **Picas (pc)**: Used in printing (1pc = 12pt)
- **Ems (em)**: Relative to the current font size (1em = current font size)

---

Feel free to experiment with different configurations and settings to achieve the desired layout for your project. The Zine Layout DSL is designed to accommodate a wide range of layouts, from simple page arrangements to complex booklets with precise measurements and rotations.