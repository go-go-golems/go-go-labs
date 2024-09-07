# Zine Layout YAML DSL Specification

## 1. Overview

This YAML DSL (Domain-Specific Language) is designed for laying out pages in a zine, including complex layouts where multiple input pages can be placed on a single output page. It allows users to specify page positions, rotations, and margins at various levels of granularity.

## 2. Structure

The YAML document consists of three main sections:

1. `global`: Contains settings that apply to the entire document.
2. `page_setup`: Defines the layout and margins for output pages.
3. `output_pages`: Specifies how input pages are arranged on output pages.

## 3. Global Settings

The `global` section can contain the following properties:

- `margin`: Specifies default margins for all output pages.
  - `top`: Top margin (number)
  - `bottom`: Bottom margin (number)
  - `left`: Left margin (number)
  - `right`: Right margin (number)

## 4. Page Setup

The `page_setup` section defines properties for output pages:

- `orientation`: Page orientation ("portrait" or "landscape")
- `grid_size`: Grid size for the output page (optional, overrides page_setup margins)
  - `rows`: Number of rows (number)
  - `columns`: Number of columns (number)
- `margins`: Default margins for output pages (overrides global margins)
  - `top`: Top margin (number)
  - `bottom`: Bottom margin (number)
  - `left`: Left margin (number)
  - `right`: Right margin (number)

## 5. Output Page Definition

Each page in the `output_pages` list defines how input pages are arranged:

- `id`: Unique identifier for the output page (string, required)
- `margin`: Individual margins for this output page (optional, overrides page_setup margins)
  - `top`: Top margin (number)
  - `bottom`: Bottom margin (number)
  - `left`: Left margin (number)
  - `right`: Right margin (number)
- `layout`: List of input pages and their placement on this output page
  - `input_index`: Index of the input page (number, required, 1-based)
  - `position`: [row, column] coordinates on the output page (array of two integers, required)
  - `rotation`: Rotation of the input page (0, 90, 180, or 270, optional)
  - `margin`: Individual margins for this input page on this output page (optional)
    - `top`: Top margin (number)
    - `bottom`: Bottom margin (number)
    - `left`: Left margin (number)
    - `right`: Right margin (number)

## 6. Example: 8-page Zine Layout

```yaml
global:
  margin:
    top: 10
    bottom: 10
    left: 10
    right: 10

page_setup:
  size: [297, 210]  # A4 landscape
  orientation: landscape
  grid_size:
    rows: 2
    columns: 2
  margins:
    top: 5
    bottom: 5
    left: 5
    right: 5

output_pages:
  - id: output1
    layout:
      - input_index: 2
        position: [0, 0]
      - input_index: 7
        position: [1, 0]
      - input_index: 3
        position: [0, 1]
      - input_index: 6
        position: [1, 1]
  - id: output2
    margin:
      left: 15
      right: 15
    layout:
      - input_index: 1
        position: [0, 0]
      - input_index: 8
        position: [1, 0]
      - input_index: 4
        position: [0, 1]
      - input_index: 5
        position: [1, 1]
        margin:
          top: 2
          bottom: 2
```

## 7. Notes

- The `global` margins apply to all output pages unless overridden.
- `page_setup` margins override `global` margins for all output pages.
- Individual output page margins override both `global` and `page_setup` margins.
- Input page margins (specified in the `layout` section) are applied within the space allocated on the output page.
- Position coordinates in the `layout` section are integers, where [0, 0] is top-left and [3, 3] would be bottom-right for a 4x4 grid layout.
- The examples show layouts for an 8-page zine and a 32-page zine that can be printed on two sheets and one sheet, respectively.
- Input pages are referred to by their index (1-based) in the `input_index` field.
- The content of input pages is not specified in this YAML; it's assumed to be handled separately (e.g., in a PDF generation step).

## 8. Additional Example: 32-page Zine Layout (8 input pages per output page)

```yaml
global:
  margin:
    top: 5
    bottom: 5
    left: 5
    right: 5

page_setup:
  orientation: landscape
  grid_size:
    rows: 4
    columns: 2
  margins:
    top: 2
    bottom: 2
    left: 2
    right: 2

output_pages:
  - id: output1
    layout:
      - input_index: 32
        position: [0, 0]
      - input_index: 1
        position: [1, 0]
      - input_index: 2
        position: [2, 0]
      - input_index: 31
        position: [3, 0]
      - input_index: 30
        position: [0, 1]
      - input_index: 3
        position: [1, 1]
      - input_index: 4
        position: [2, 1]
      - input_index: 29
        position: [3, 1]
  - id: output2
    layout:
      - input_index: 28
        position: [0, 0]
      - input_index: 5
        position: [1, 0]
      - input_index: 6
        position: [2, 0]
      - input_index: 27
        position: [3, 0]
      - input_index: 26
        position: [0, 1]
      - input_index: 7
        position: [1, 1]
      - input_index: 8
        position: [2, 1]
      - input_index: 25
        position: [3, 1]
  - id: output3
    layout:
      - input_index: 24
        position: [0, 0]
      - input_index: 9
        position: [1, 0]
      - input_index: 10
        position: [2, 0]
      - input_index: 23
        position: [3, 0]
      - input_index: 22
        position: [0, 1]
      - input_index: 11
        position: [1, 1]
      - input_index: 12
        position: [2, 1]
      - input_index: 21
        position: [3, 1]
  - id: output4
    layout:
      - input_index: 20
        position: [0, 0]
      - input_index: 13
        position: [1, 0]
      - input_index: 14
        position: [2, 0]
      - input_index: 19
        position: [3, 0]
      - input_index: 18
        position: [0, 1]
      - input_index: 15
        position: [1, 1]
      - input_index: 16
        position: [2, 1]
      - input_index: 17
        position: [3, 1]
```

## 9. Printing and Assembly Considerations

For the 32-page zine layout example:

1. Print double-sided: Output1 and Output2 on the front, Output3 and Output4 on the back.
2. Cut the sheet in half horizontally.
3. Stack the resulting two pieces with Output1/Output3 on top and Output2/Output4 at the bottom.
4. Fold the stack in half vertically, then in half horizontally to create the final zine.

This layout ensures that when the zine is assembled, the pages will be in the correct order: 1, 2, 3, ..., 31, 32.