# Severance-Inspired Museum Display Webapp - Enhanced Version

## Overview

This enhanced version of the Severance-inspired museum display webapp includes three major new features:

1. **Mermaid Diagram Support** - Render diagrams using the mermaid.js syntax
2. **Search Functionality** - Search across all loaded museum displays
3. **Print View** - Optimized view for printing content

The webapp remains a self-contained frontend application with no backend requirements, allowing you to upload and display JSON files with museum content in a Severance-inspired interface.

## Features

### Core Features

- Upload and parse JSON files containing museum display content
- Switch between multiple loaded displays
- Navigate through different pages and content types
- Severance-inspired retro terminal aesthetic with blue/whitish text on black background

### New Features

#### Mermaid Diagram Support

- Integrated mermaid.js library for rendering diagrams
- Support for various diagram types:
  - Flowcharts
  - Sequence diagrams
  - Class diagrams
  - State diagrams
  - Entity relationship diagrams
  - Gantt charts
  - Pie charts
- Diagrams can be included in:
  - Slide decks
  - Tutorials
  - Interactive code examples
  - Hardware visuals
  - Dedicated diagram pages

#### Search Functionality

- Search across all loaded museum displays
- Content indexing when JSON files are loaded
- Search results with context highlighting
- Navigation to search results
- Keyboard shortcut support (Ctrl+F/Cmd+F)

#### Print View

- Toggle between display and print views
- Print-optimized styling with improved readability
- Removal of interactive elements in print view
- Automatic page breaks at appropriate locations
- Keyboard shortcut support (Ctrl+P/Cmd+P)

## Getting Started

1. Extract the zip file to a local directory
2. Open `index.html` in any modern web browser
3. Click "SELECT FILE" to upload a JSON file
4. Navigate through your museum display using the sidebar

## Using the New Features

### Mermaid Diagrams

To include mermaid diagrams in your JSON files, add a `mermaid` property to the appropriate content element with the mermaid syntax as its value. For example:

```json
{
  "title": "Flowchart Example",
  "mermaid": "graph TD;\n    A[Start] --> B{Decision};\n    B -->|Yes| C[Action 1];\n    B -->|No| D[Action 2];\n    C --> E[End];\n    D --> E;"
}
```

See the included `test-mermaid.json` for examples of different diagram types and how to include them in various page types.

### Search

1. Use the search bar in the top-right corner of the interface
2. Enter your search term and press Enter or click the search icon
3. Results will show matching content with context
4. Click "GO TO" to navigate to a specific result
5. Click "CLOSE" to return to your previous view

### Print View

1. Click the "PRINT" button in the top-right corner
2. The interface will switch to a print-optimized view
3. Use your browser's print function (Ctrl+P/Cmd+P) to print the content
4. Click "EXIT PRINT" to return to the normal view

## JSON Specification

The JSON specification has been updated to include support for mermaid diagrams. See the included `museum-display-json-specification.md` for detailed documentation on the JSON format.

## Browser Compatibility

The webapp is compatible with all modern browsers:
- Chrome/Edge (latest versions)
- Firefox (latest version)
- Safari (latest version)

## Known Limitations

- Large or complex mermaid diagrams may take longer to render
- Print view may vary slightly between browsers
- Very large JSON files may cause performance issues with search indexing

## Troubleshooting

- If diagrams don't render, check your mermaid syntax for errors
- If search doesn't find expected results, try using different keywords
- If print view doesn't display correctly, try using Chrome for best results

## License

This project is provided for educational and demonstration purposes.
