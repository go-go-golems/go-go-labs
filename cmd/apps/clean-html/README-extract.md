# HTML Content Extractor

This project provides a command-line tool to extract specific content from an HTML document using a YAML-based configuration file that specifies CSS selectors and titles. The extracted content is outputted in YAML format, making it easy to parse and reuse.

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
  - [Configuration File Format](#configuration-file-format)
  - [Command-Line Usage](#command-line-usage)
- [Examples](#examples)
- [Dependencies](#dependencies)
- [License](#license)

## Introduction

The `extractor.py` script allows you to extract specific elements from an HTML file using CSS selectors defined in a YAML configuration file. This tool is particularly useful for extracting structured content such as headings, paragraphs, or code snippets from HTML documents, and presenting the extracted data in a YAML format.

## Features

- **Customizable Selectors**: Define CSS selectors in a YAML file to control which parts of the HTML are extracted.
- **Flexible Input/Output**: Reads HTML input from `stdin` and outputs extracted YAML content to `stdout`.
- **Command-Line Flags**: Uses `click` to provide an easy-to-use command-line interface for configuration.
- **YAML Output**: Outputs extracted content in a structured YAML format.

## Installation

To get started, clone this repository and install the required dependencies:

```bash
git clone https://github.com/yourusername/html-extractor.git
cd html-extractor
pip install -r requirements.txt
```

The dependencies are:
- `click`
- `pyyaml`
- `beautifulsoup4`

Alternatively, you can install them manually:

```bash
pip install click pyyaml beautifulsoup4
```

## Usage

### Configuration File Format

The configuration file should be written in YAML and define the list of selectors and titles for each section you want to extract. Here's an example `config.yaml` file:

```yaml
selectors:
  - title: Repository Title
    selector: "h1.heading-element"
  - title: Repository Description
    selector: ".markdown-body p"
  - title: Code Snippets
    selector: ".highlight"
  - title: Policy Documentation
    selector: ".markdown-body h2.heading-element"
```

### Command-Line Usage

1. **Create a YAML configuration file** (`config.yaml`) that specifies the CSS selectors and titles for extraction.
2. **Run the extractor** with the following command:

   ```bash
   cat your_html_file.html | python extractor.py --config config.yaml
   ```

This command will:
- Read the HTML content from `your_html_file.html`.
- Extract elements defined in the `config.yaml` file using the provided CSS selectors.
- Output the extracted content as YAML to `stdout`.

### Options

- `--config, -c`: Specifies the path to the YAML configuration file.

## Examples

Suppose you have the following HTML content in a file named `example.html`:

```html
<!DOCTYPE html>
<html>
  <body>
    <h1 class="heading-element">Repository Title Example</h1>
    <div class="markdown-body">
      <p>This is a description of the repository.</p>
      <div class="highlight">
        <pre>def example_function():
    print("Hello, world!")</pre>
      </div>
    </div>
  </body>
</html>
```

And the following YAML configuration file (`config.yaml`):

```yaml
selectors:
  - title: Repository Title
    selector: "h1.heading-element"
  - title: Repository Description
    selector: ".markdown-body p"
  - title: Code Snippets
    selector: ".highlight pre"
```

Running the command:

```bash
cat example.html | python extractor.py --config config.yaml
```

Would produce the following output:

```yaml
Repository Title:
- Repository Title Example

Repository Description:
- This is a description of the repository.

Code Snippets:
- |
  def example_function():
      print("Hello, world!")
```
