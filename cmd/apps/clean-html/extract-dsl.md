# HTML Extraction DSL Specification

This document outlines the Domain Specific Language (DSL) used for configuring HTML content extraction. The DSL is written in YAML format and provides a flexible way to define selectors and extraction rules.

## Basic Structure

The configuration file consists of a list of selector objects under the `selectors` key:

```yaml
selectors:
  - title: Selector Title
    selector: "CSS Selector"
    assemble: Assembly Strategy
    # Additional options...
```

## Selector Object Properties

- `title`: A string representing the name of the extracted data.
- `selector`: A string containing a valid CSS selector (always quoted).
- `assemble`: Defines how the selected elements should be processed. Options include:
  - `list`: Returns a list of text content or attribute values (default).
  - `single`: Returns the text content of the first matched element.
  - `concatenate`: Joins the text content of all matched elements with newlines.
  - `code_blocks`: Similar to concatenate, but formats each element as a code block.
  - `hash`: Creates a key-value pair for each element.
- `attributes`: A list of attribute names to extract (used with `list` assembly).
- `transformations`: A list of transformations to apply to the extracted text. Options include:
  - `strip`: Removes leading and trailing whitespace.
  - `capitalize`: Capitalizes the first character of the text.
- `children`: A list of nested selector objects for hierarchical extraction.
- `key_attribute`: Specifies the attribute or "text" to use as the key in `hash` assembly (default: "text").
- `value_attribute`: Specifies the attribute to use as the value in `hash` assembly (default: "href").

## Examples

### 1. Simple Text Extraction

```yaml
selectors:
  - title: Page Title
    selector: "h1"
    assemble: single
```

### 2. List of Elements

```yaml
selectors:
  - title: Navigation Links
    selector: "nav ul li a"
    assemble: list
```

### 3. Attribute Extraction

```yaml
selectors:
  - title: Image Sources
    selector: "img"
    assemble: list
    attributes:
      - src
      - alt
```

### 4. Text Concatenation

```yaml
selectors:
  - title: Article Content
    selector: "article p"
    assemble: concatenate
```

### 5. Code Block Extraction

```yaml
selectors:
  - title: Code Snippets
    selector: "pre code"
    assemble: code_blocks
```

### 6. Key-Value Pairs

```yaml
selectors:
  - title: Meta Tags
    selector: "meta"
    assemble: hash
    key_attribute: name
    value_attribute: content
```

### 7. Nested Selectors

```yaml
selectors:
  - title: Blog Posts
    selector: "article"
    assemble: list
    children:
      - title: Post Title
        selector: "h2"
        assemble: single
      - title: Post Content
        selector: "p"
        assemble: concatenate
      - title: Post Tags
        selector: ".tags span"
        assemble: list
```

### 8. Complex Example with Multiple Techniques

```yaml
selectors:
  - title: Product Information
    selector: ".product"
    assemble: hash
    children:
      - title: Name
        selector: "h1"
        assemble: single
        transformations:
          - strip
          - capitalize
      - title: Price
        selector: ".price"
        assemble: single
      - title: Description
        selector: ".description p"
        assemble: concatenate
      - title: Features
        selector: ".features li"
        assemble: list
      - title: Images
        selector: ".gallery img"
        assemble: list
        attributes:
          - src
      - title: Reviews
        selector: ".review"
        assemble: list
        children:
          - title: Author
            selector: ".author"
            assemble: single
          - title: Rating
            selector: ".rating"
            assemble: single
          - title: Comment
            selector: ".comment"
            assemble: concatenate
  - title: Related Products
    selector: ".related-products a"
    assemble: hash
    key_attribute: text
    value_attribute: href
```

### 9. ID Selector Example

```yaml
selectors:
  - title: Main Content
    selector: "#main-content"
    assemble: single
```

Note: When using ID selectors (e.g., "#main-content"), it's crucial to enclose the selector in quotes. This is because the "#" character is used for comments in YAML. Quoting the selector ensures that it's interpreted correctly as a CSS selector rather than the start of a comment.

This complex example demonstrates nested selectors, multiple assembly strategies, transformations, and attribute extractions within a single configuration. All selectors are now explicitly quoted to ensure proper parsing.
