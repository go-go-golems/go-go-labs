# OCR Mistral View

A web application to view OCR results from Mistral in a user-friendly, browsable format.

## Features

- View OCR results from Mistral in a web browser
- Navigate through pages with a page browser
- Option to view all pages at once
- Automatic extraction and rendering of images from the OCR results
- Markdown rendering for text content

## Installation

```bash
# Clone the repository
git clone https://github.com/go-go-golems/go-go-labs.git
cd go-go-labs/cmd/apps/ocr-mistral-view

# Build the application
go generate ./...
go build
```

## Usage

```bash
# Run the application with an input JSON file
./ocr-mistral-view -i path/to/ocr_result.json

# Specify a custom port (default is 8080)
./ocr-mistral-view -i path/to/ocr_result.json -p 8081
```

Once the application is running, open your web browser and navigate to http://localhost:8080 (or your specified port).

## Input JSON Format

The application expects a JSON file in the following format:

```json
{
  "model": "mistral-ocr-2503-completion",
  "pages": [
    {
      "dimensions": {
        "dpi": 200,
        "height": 2200,
        "width": 1700
      },
      "images": [],
      "index": 0,
      "markdown": "# LEVERAGING UNLABELED DATA TO PREDICT OUT-OF-DIST..."
    }
    // More pages...
  ],
  "usage_info": {
    "doc_size_bytes": 3002783,
    "pages_processed": 29
  }
}
```

## License

MIT
