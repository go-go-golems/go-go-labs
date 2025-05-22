# JavaScript Docstring Extractor

A command-line tool that extracts docstrings and function signatures from JavaScript files and outputs them in Markdown format.

## Features

- Extracts function declarations, function expressions, and arrow functions
- Identifies JSDoc comments and regular comments
- Processes single files, directories, or input from stdin
- Generates Markdown documentation with function signatures and docstrings
- Option to extract only JSDoc comments (starting with `/**`)
- Support for recursive directory processing

## Installation

```bash
go install github.com/go-go-golems/go-go-labs/cmd/js-docstring-extractor-2@latest
```

Or build from source:

```bash
git clone https://github.com/go-go-golems/go-go-labs.git
cd go-go-labs
go build -o js-docstring-extractor-2 ./cmd/js-docstring-extractor-2
```

## Usage

### Process a single file:

```bash
js-docstring-extractor-2 path/to/file.js
```

### Process all JavaScript files in a directory:

```bash
js-docstring-extractor-2 path/to/directory
```

### Process files recursively:

```bash
js-docstring-extractor-2 -r path/to/directory
```

### Read from stdin:

```bash
cat file.js | js-docstring-extractor-2
```

### Write output to a file:

```bash
js-docstring-extractor-2 -o docs.md path/to/file.js
```

### Extract only JSDoc comments:

```bash
js-docstring-extractor-2 -j path/to/file.js
```

## Examples

Extract docstrings from a single file:

```bash
js-docstring-extractor-2 src/utils.js
```

Process multiple files:

```bash
js-docstring-extractor-2 src/utils.js src/components.js
```

Recursively process all JavaScript files in a directory and save to a file:

```bash
js-docstring-extractor-2 -r -o docs/api.md src/
```

## Output Format

The tool generates Markdown documentation with the following structure:

- File name (if multiple files are processed)
- Top-level docstring (if present)
- Function name as heading
- Function docstring
- Function signature in code block

## Implementation Details

This tool uses the [tdewolff/parse](https://github.com/tdewolff/parse) library to parse JavaScript code and extract docstrings. It matches comments to functions based on their proximity in the source code.
