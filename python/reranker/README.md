# Document Reranker Web App

A Flask + HTMX web application for reranking documents based on their relevance to a query using the BAAI/bge-reranker-large model.

## Features

- Upload YAML files with queries and documents
- Paste YAML content directly in the browser
- View example YAML files and use them as templates
- Rerank documents with scores displayed in a clean UI
- Informational cheatsheet about reranking

## Installation

1. Clone the repository
2. Install dependencies:
   ```
   pip install -r requirements.txt
   ```

## Usage

1. Start the application:
   ```
   python app.py
   ```
2. Open a web browser and navigate to `http://localhost:5000`
3. Upload a YAML file or paste YAML content containing:
   - A query
   - A list of documents
   - Optional top_k parameter

## YAML Format

```yaml
query: "Your query text here"
documents:
  - "First document text"
  - "Second document text"
  - "Third document text"
top_k: 5  # Optional
```

## Example Files

The `examples/` directory contains sample YAML files that can be used as templates:

- `rag_pipeline.yaml` - Example about semantic chunking in RAG systems
- `llm_architecture.yaml` - Example about transformer architecture components
- `coffee_brewing.yaml` - Example about coffee brewing temperatures

## Model Information

This application uses the [BAAI/bge-reranker-large](https://huggingface.co/BAAI/bge-reranker-large) model to calculate relevance scores for query-document pairs. 