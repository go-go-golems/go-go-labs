# clean_html.py

A Python script to simplify HTML files by removing unnecessary elements and attributes.

## Features

- Removes comments, scripts, styles, meta tags, and other non-content elements
- Preserves only 'id' and 'class' attributes
- Optional whitespace cleanup
- Outputs a cleaned HTML file for easier content extraction and CSS selector creation
- Supports reading from stdin and writing to stdout
- Flexible command-line interface using Click

## Requirements

- Python 3.x
- BeautifulSoup4
- Click

## Installation

1. Ensure Python 3.x is installed on your system.
2. Install the required packages:

```bash
pip install beautifulsoup4 click
```

## Usage

Run the script from the command line:

```bash
# Using stdin and stdout
cat input.html | ./clean_html.py > output.html

# Specifying input and output files
./clean_html.py -i input.html -o output.html

# Enable whitespace cleanup
./clean_html.py -i input.html -o output.html --cleanup-whitespace

# Get help
./clean_html.py --help
```

Options:
- `-i, --input FILE`: Input HTML file (default: stdin)
- `-o, --output FILE`: Output HTML file (default: stdout)
- `--cleanup-whitespace`: Enable whitespace cleanup (default: False)

## Example

Input HTML:

```html
<!DOCTYPE html>
<html>
<head>
    <title>Example</title>
    <style>/* CSS */</style>
</head>
<body>
    <div id="main" class="container" style="color: red;">
        <h1>Hello, World!</h1>
    </div>
</body>
</html>
```

Output HTML (without whitespace cleanup):

```html
<!DOCTYPE html>
<html>
<head><title>Example</title></head>
<body>
<div id="main" class="container">
<h1>Hello, World!</h1>
</div>
</body>
</html>
```

Output HTML (with whitespace cleanup):

```html
<!DOCTYPE html> <html> <head><title>Example</title></head> <body> <div id="main" class="container"> <h1>Hello, World!</h1> </div> </body> </html>
```

## Customization

Modify the `simplify_html` function in the script to preserve additional attributes or remove more elements as needed.

## Integration

The script can be easily integrated into Unix-style pipelines or used with file input/output as needed, making it versatile for various text processing workflows.

## Whitespace Cleanup

The `--cleanup-whitespace` flag enables more aggressive whitespace cleanup:
- Collapses multiple whitespace characters into a single space
- Trims leading and trailing whitespace from text nodes
- Removes unnecessary whitespace between HTML tags

Use this option when you need a more compact HTML output, but be aware that it may affect the rendering of whitespace-sensitive content.