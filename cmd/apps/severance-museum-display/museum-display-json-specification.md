# Museum Display JSON Specification

This document outlines the JSON format for creating interactive museum displays in a Severance-inspired web application.

## Overview

The JSON structure defines a complete museum display with multiple pages of different types, navigation options, theming, and footer information. Each museum display is self-contained within a single JSON file that can be loaded into the web application.

## Root Structure

```json
[
  {
    "microplanner_museum_display": {
      "title": "String - Main title of the museum display",
      "theme": "String - 'dark' or 'light'",
      "navigation": { /* Navigation object */ },
      "pages": [ /* Array of page objects */ ],
      "footer": { /* Footer object */ }
    }
  }
]
```

The root element is an array containing a single object with a key that represents the museum display ID (e.g., "microplanner_museum_display"). This structure allows for potential future expansion to include multiple displays in a single file.

## Navigation Object

```json
"navigation": {
  "type": "String - 'sidebar', 'top', or 'bottom'",
  "persistent_menu": "Boolean - whether menu is always visible",
  "show_progress": "Boolean - whether to show progress indicator"
}
```

## Footer Object

```json
"footer": {
  "text": "String - footer text content",
  "logos": [
    "String - URL or path to logo image",
    "String - URL or path to logo image",
    ...
  ]
}
```

## Pages Array

The `pages` array contains multiple page objects, each representing a different section of the museum display. Each page has a common structure with type-specific properties:

```json
{
  "id": "String - unique identifier for the page",
  "title": "String - page title",
  "type": "String - page type identifier",
  /* Type-specific properties */
}
```

### Page Types

#### 1. Slide Deck (`type: "slide_deck"`)

A sequence of slides with title and content.

```json
{
  "id": "intro",
  "title": "Welcome to Micro-Planner",
  "type": "slide_deck",
  "slides": [
    {
      "title": "What is Micro-Planner?",
      "content": "String - slide content with optional markdown"
    },
    /* More slides */
  ]
}
```

#### 2. Tutorial (`type: "tutorial"`)

Step-by-step instructions.

```json
{
  "id": "tutorial_id",
  "title": "Tutorial Title",
  "type": "tutorial",
  "steps": [
    {
      "title": "Step 1: Title",
      "description": "String - step description with optional markdown"
    },
    /* More steps */
  ]
}
```

#### 3. Interactive Code (`type: "interactive_code"`)

Code examples with descriptions.

```json
{
  "id": "code_demos",
  "title": "Interactive Examples",
  "type": "interactive_code",
  "language": "String - programming language (e.g., 'lisp')",
  "examples": [
    {
      "id": "example1",
      "title": "Example Title",
      "description": "String - example description",
      "code": "String - code content"
    },
    /* More examples */
  ]
}
```

#### 4. Hardware Visual (`type: "hardware_visual"`)

Visual representation of hardware with interactive elements.

```json
{
  "id": "hardware_id",
  "title": "Hardware Title",
  "type": "hardware_visual",
  "panels": [
    {
      "name": "Panel Name",
      "image": "String - URL or path to image",
      "description": "String - panel description",
      "interactions": [
        {
          "label": "Button Label",
          "action": "String - description of action"
        },
        /* More interactions */
      ]
    },
    /* More panels */
  ]
}
```

#### 5. Bio Gallery (`type: "bio_gallery"`)

Collection of biographical information.

```json
{
  "id": "people",
  "title": "People Title",
  "type": "bio_gallery",
  "bios": [
    {
      "name": "Person Name",
      "role": "Person Role",
      "image": "String - URL or path to image",
      "quote": "String - notable quote"
    },
    /* More bios */
  ]
}
```

#### 6. Resource List (`type: "resource_list"`)

List of external resources with links.

```json
{
  "id": "resources",
  "title": "Resources Title",
  "type": "resource_list",
  "resources": [
    {
      "title": "Resource Title",
      "link": "String - URL to resource"
    },
    /* More resources */
  ]
}
```

#### 7. Quiz (`type: "quiz"`)

Interactive quiz with questions and answers.

```json
{
  "id": "quiz",
  "title": "Quiz Title",
  "type": "quiz",
  "questions": [
    {
      "question": "Question text?",
      "options": [
        "Option 1",
        "Option 2",
        "Option 3",
        "Option 4"
      ],
      "answer": "String - correct answer (must match one of the options exactly)"
    },
    /* More questions */
  ]
}
```

## Content Formatting

- Text content fields (`content`, `description`, etc.) support basic Markdown formatting
- Code blocks can be included using triple backticks (```)
- Line breaks are preserved with `\n`
- HTML entities can be used for special characters (e.g., `&amp;` for &)

## Image References

- Image paths can be absolute URLs or relative paths
- Relative paths are resolved relative to the web application's base URL
- Recommended image formats: JPEG, PNG, SVG

## Best Practices

1. **Unique IDs**: Ensure all page IDs and example IDs are unique within the display
2. **Content Length**: Keep slide content concise for better readability
3. **Image Optimization**: Use appropriately sized and optimized images
4. **Markdown Usage**: Use markdown for formatting rather than HTML when possible
5. **Quiz Options**: Provide 3-5 options for each quiz question
6. **Consistent Naming**: Use consistent naming conventions for IDs and titles

## Example

See the provided `micro-planner.json` for a complete example of a museum display JSON file.
