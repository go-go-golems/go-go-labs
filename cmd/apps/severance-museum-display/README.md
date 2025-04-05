# Severance-Inspired Museum Display Webapp

A self-contained frontend application for creating interactive museum displays with a Severance-inspired aesthetic. This webapp allows you to upload and switch between multiple JSON files containing museum display content.

## Features

- **Severance-inspired UI**: Blue/whitish text on black background with terminal-like aesthetics
- **No backend required**: Completely client-side, runs in any modern browser
- **Multiple file support**: Upload and switch between different museum displays
- **Responsive design**: Works on desktop and mobile devices
- **Support for various content types**:
  - Slide decks
  - Step-by-step tutorials
  - Interactive code examples
  - Hardware visualizations
  - Bio galleries
  - Resource lists
  - Interactive quizzes

## Getting Started

1. Extract the `severance-museum-display.zip` file to a location of your choice
2. Open the `index.html` file in a web browser
3. Click "SELECT FILE" to upload a JSON file in the specified format
4. Navigate through the museum display using the sidebar navigation

## JSON File Format

The webapp accepts JSON files in a specific format. See the included `museum-display-json-specification.md` for detailed documentation on the JSON structure.

Two example JSON files are included:
- `micro-planner.json`: The full example with all page types
- `test-json.json`: A simplified example with just slide decks

## Creating Your Own Displays

To create your own museum displays:

1. Use the JSON specification document as a reference
2. Create a new JSON file following the required structure
3. Include all required fields for each page type
4. For images, use absolute URLs or relative paths to publicly accessible images
5. Upload your JSON file to the webapp

## Browser Compatibility

This webapp is compatible with all modern browsers including:
- Chrome/Edge (latest versions)
- Firefox (latest versions)
- Safari (latest versions)
- Mobile browsers (iOS Safari, Android Chrome)

## License

This project is provided for your use without restrictions.

## Acknowledgments

- Inspired by the aesthetic of the TV show "Severance"
- Created as a self-contained museum display solution
