# HTML Content Extractor

This project provides a powerful command-line tool to extract specific content from HTML documents using a flexible YAML-based configuration file. The extracted content is outputted in YAML format, making it easy to parse and reuse.

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Usage](#usage)
  - [Configuration File Format](#configuration-file-format)
  - [Command-Line Usage](#command-line-usage)
- [Examples](#examples)
- [Dependencies](#dependencies)
- [License](#license)

## Introduction

The HTML Content Extractor allows you to extract specific elements from an HTML file using CSS selectors defined in a YAML configuration file. This tool is particularly useful for extracting structured content such as headings, paragraphs, attributes, or code snippets from HTML documents, and presenting the extracted data in a YAML format.

## Features

- **Customizable Selectors**: Define CSS selectors in a YAML file to control which parts of the HTML are extracted.
- **Flexible Assembly Strategies**: Choose from various assembly methods like list, single, concatenate, code_blocks, and hash.
- **Nested Selectors**: Support for hierarchical data extraction with nested selector objects.
- **Attribute Extraction**: Extract specific attributes from HTML elements.
- **Text Transformations**: Apply various transformations to extracted text, such as stripping whitespace, capitalizing, or converting to lowercase.
- **Flexible Input/Output**: Reads HTML input from `stdin` and outputs extracted YAML content to `stdout`.
- **Command-Line Interface**: Uses `click` to provide an easy-to-use command-line interface for configuration.


## Usage

### Configuration File Format

The configuration file should be written in YAML and define the list of selectors and extraction rules. Here's a basic example:

```yaml
selectors:
  - title: Content
    selector: ".content"
    assemble: "hash"
    children:
      - title: Description
        selector: "p.description"
        assemble: "single"
        transformations:
          - "remove_newlines"
          - "trim_spaces"
          - "capitalize"
      - title: List Items
        selector: "ul.list li"
        assemble: "list"
        transformations:
          - "strip"
          - "to_lowercase"
```

For a complete guide on the configuration options, refer to the [HTML Extraction DSL Specification](extract-dsl.md).

### Command-Line Usage

1. Create a YAML configuration file (e.g., `config.yaml`) that specifies the CSS selectors and extraction rules.
2. Run the extractor with the following command:

   ```bash
   cat your_html_file.html | python extractor.py --config config.yaml
   ```

This command will:
- Read the HTML content from `your_html_file.html`.
- Extract elements defined in the `config.yaml` file using the provided CSS selectors and rules.
- Output the extracted content as YAML to `stdout`.

## Examples

Here's an example of extracting information from a GitHub repository page:

```yaml
selectors:
  - title: Repository Information
    selector: ".markdown-body.entry-content.container-lg"
    assemble: "hash"
    children:
      - title: Title
        selector: "h1.heading-element"
        assemble: "single"
      - title: Description
        selector: "p"
        assemble: "concatenate"
        transformations:
          - "strip"
      - title: Links
        selector: ".markdown-heading h2:contains(\"Links\") a"
        assemble: "hash"
        key_attribute: "text"
        value_attribute: "href"
  - title: Repository Statistics
    selector: "#repository-container-header"
    assemble: "hash"
    children:
      - title: Stars
        selector: "#repo-stars-counter"
        assemble: "single"
      - title: Forks
        selector: "#repo-network-counter"
        assemble: "single"
```

This configuration would extract the repository title, description, links, and statistics from a GitHub repository page.

For more examples and detailed explanations of the configuration options, please refer to the [HTML Extraction DSL Specification](extract-dsl.md).

## Dependencies

- `click`: For creating the command-line interface
- `pyyaml`: For parsing YAML configuration files
- `beautifulsoup4`: For HTML parsing and content extraction
