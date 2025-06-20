# SQLite Vector Search with Ollama Embeddings

This experiment demonstrates SQLite vector search capabilities using custom functions and real embeddings from Ollama.

## Features

- Custom SQLite functions for cosine similarity calculation
- Integration with Ollama for generating embeddings using `all-minilm:latest` model
- CLI interface for adding documents and searching
- Sample data insertion and demonstration

## Prerequisites

- Ollama running locally on `127.0.0.1:11434`
- `all-minilm:latest` model available in Ollama

## Usage

### Basic Demo
Run the default demo which inserts sample documents and performs a search:

```bash
go run ./cmd/experiments/sqlite-vector-search
```

### Add Documents
Add a new document to the database:

```bash
go run ./cmd/experiments/sqlite-vector-search add "Your document text here"
```

### Search Documents
Search for similar documents:

```bash
go run ./cmd/experiments/sqlite-vector-search search "your search query" --limit 5
```

### Options

- `--log-level`: Set log level (debug, info, warn, error)
- `--ollama-url`: Ollama server URL (default: http://127.0.0.1:11434)
- `--ollama-model`: Ollama model to use (default: all-minilm:latest)
- `--db-path`: SQLite database path (default: vector_search.db)

## How It Works

1. **Custom SQLite Functions**: Registers a `cosine_similarity` function in SQLite
2. **Embedding Generation**: Uses Ollama API to generate embeddings for text
3. **Vector Storage**: Stores embeddings as JSON in SQLite
4. **Similarity Search**: Calculates cosine similarity between query and stored embeddings

## Example Output

```
Search results for: artificial intelligence and neural networks
============================================================
1. [0.8234] Machine learning is a subset of artificial intelligence
2. [0.7891] Deep learning uses neural networks with multiple layers
3. [0.6456] Natural language processing involves understanding text
4. [0.5234] Embeddings capture semantic meaning of text
5. [0.3456] Vector databases are useful for similarity search
```

## Technical Details

- Uses `github.com/mattn/go-sqlite3` for SQLite integration
- Implements cosine similarity calculation in Go
- JSON encoding for vector storage in SQLite
- Zerolog for structured logging
- Cobra for CLI interface
