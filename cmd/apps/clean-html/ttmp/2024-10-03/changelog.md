# Clean HTML Script Enhancements

Improved the clean_html.py script for better command-line usage and integration.

- Integrated Click command-line framework for improved argument handling
- Added support for reading from stdin and writing to stdout by default
- Introduced -i/--input and -o/--output options for specifying input/output files
- Updated usage to support both file-based and pipeline-based operations


# Clean HTML Script Documentation Update

Updated the README.md to reflect recent changes in the clean_html.py script.

- Added information about Click framework usage
- Updated installation instructions to include Click
- Modified usage examples to show stdin/stdout and file-based operations
- Included details about new command-line options
- Added an Integration section to highlight script versatility

# HTML Cleaner Enhancements

## Add whitespace collapsing functionality

- Added a new `collapse_whitespace` function to reduce multiple whitespace characters to a single space.
- Updated `simplify_html` function to apply whitespace collapsing to text nodes.
- This change helps create more compact and readable HTML output by removing unnecessary spaces and line breaks.