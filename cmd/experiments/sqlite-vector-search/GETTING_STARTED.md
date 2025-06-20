# Getting Started with SQLite Vector Search

This guide will walk you through building and understanding the SQLite vector search implementation step by step.

## Prerequisites

### 1. Install Ollama

```bash
# On macOS
brew install ollama

# On Linux
curl -fsSL https://ollama.ai/install.sh | sh

# On Windows
# Download from https://ollama.ai/download
```

### 2. Start Ollama and Download Model

```bash
# Start Ollama server
ollama serve

# In another terminal, download the embedding model
ollama pull all-minilm:latest
```

### 3. Verify Ollama is Working

```bash
# Test the embeddings endpoint
curl http://127.0.0.1:11434/api/embeddings \
  -d '{"model":"all-minilm:latest","prompt":"test"}'
```

You should see a response with an "embedding" array of numbers.

## Quick Start

### 1. Build the Program

```bash
cd go-go-labs
go build ./cmd/experiments/sqlite-vector-search
```

### 2. Run the Demo

```bash
# Run the full demo with sample data
go run ./cmd/experiments/sqlite-vector-search

# Or use the built binary
./sqlite-vector-search
```

This will:
- Connect to Ollama
- Create a SQLite database
- Register custom similarity functions
- Insert sample documents with embeddings
- Perform a similarity search

### 3. Test Individual Components

```bash
# Test basic SQLite function registration
go run ./cmd/experiments/sqlite-vector-search test basic

# Test cosine similarity calculations
go run ./cmd/experiments/sqlite-vector-search test cosine

# Test JSON handling
go run ./cmd/experiments/sqlite-vector-search test json

# Test vector operations
go run ./cmd/expressions/sqlite-vector-search test vectors

# Test error handling
go run ./cmd/experiments/sqlite-vector-search test errors
```

## Interactive Usage

### Add Your Own Documents

```bash
# Add a document
go run ./cmd/experiments/sqlite-vector-search add "Python is a programming language"

# Add another
go run ./cmd/experiments/sqlite-vector-search add "Go is also a programming language"
```

### Search for Similar Content

```bash
# Search for programming-related content
go run ./cmd/experiments/sqlite-vector-search search "coding and software development" --limit 3

# Search for specific topics
go run ./cmd/experiments/sqlite-vector-search search "machine learning AI" --limit 5
```

### Configure Options

```bash
# Use a different Ollama model
go run ./cmd/experiments/sqlite-vector-search --ollama-model "nomic-embed-text:latest" search "example query"

# Use a different database file
go run ./cmd/experiments/sqlite-vector-search --db-path "my_vectors.db" search "example query"

# Enable debug logging
go run ./cmd/experiments/sqlite-vector-search --log-level debug search "example query"
```

## Understanding the Code

### 1. SQLite Function Registration

The core concept is registering Go functions that SQLite can call:

```go
// Get access to the raw SQLite connection
conn.Raw(func(driverConn interface{}) error {
    sqliteConn := driverConn.(*sqlite3.SQLiteConn)
    
    // Register a function called "cosine_similarity"
    return sqliteConn.RegisterFunc("cosine_similarity", 
        func(a, b string) float64 {
            // Your Go code here
            return calculateSimilarity(a, b)
        }, 
        true) // true = deterministic/pure function
})
```

### 2. Data Flow

```
Text Input → Ollama API → Embedding Vector → JSON → SQLite
                                                      ↓
Search Query → Ollama API → Query Vector → SQL Function → Results
```

### 3. SQL Usage

Once registered, you can use the custom function in SQL:

```sql
SELECT 
    content,
    cosine_similarity(embedding, '[0.1, 0.2, 0.3, ...]') as similarity
FROM documents
ORDER BY similarity DESC
LIMIT 5;
```

## Building Your Own Functions

### Step 1: Define Your Function

```go
func myCustomFunction(input string) float64 {
    // Your logic here
    return result
}
```

### Step 2: Register It

```go
err = conn.Raw(func(driverConn interface{}) error {
    sqliteConn := driverConn.(*sqlite3.SQLiteConn)
    return sqliteConn.RegisterFunc("my_function", myCustomFunction, true)
})
```

### Step 3: Use in SQL

```sql
SELECT my_function(column_name) FROM table_name;
```

## Common Patterns

### Working with JSON Arrays

```go
func processVector(jsonStr string) float64 {
    var vector []float64
    if err := json.Unmarshal([]byte(jsonStr), &vector); err != nil {
        return 0 // Handle error gracefully
    }
    
    // Process vector
    return result
}
```

### Multiple Parameters

```go
func similarity(vecA, vecB string, threshold float64) int {
    // Parse vectors
    // Calculate similarity
    // Return 1 if above threshold, 0 otherwise
}
```

### Error Handling

```go
func safeFunction(input string) float64 {
    defer func() {
        if r := recover(); r != nil {
            // Log error in production
            fmt.Printf("Function panicked: %v\n", r)
        }
    }()
    
    // Your logic that might panic
    return result
}
```

## Troubleshooting

### Ollama Connection Issues

```bash
# Check if Ollama is running
curl http://127.0.0.1:11434/api/tags

# Check if model is available
ollama list
```

### SQLite Function Not Found

- Ensure you're using the same connection that registered the function
- Functions are connection-specific, not database-specific
- Use `conn.QueryRow()` instead of `db.QueryRow()` after registration

### Performance Issues

```go
// Mark functions as pure/deterministic when possible
sqliteConn.RegisterFunc("my_func", func, true) // true = pure

// Use connection pooling appropriately
db.SetMaxOpenConns(1) // For function registration consistency
```

### Memory Issues with Large Vectors

```go
// Consider chunking large operations
// Use streaming for large result sets
// Close connections and statements properly
```

## Advanced Topics

### 1. Batch Operations

```go
func insertBatch(db *sql.DB, documents []Document) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    stmt, err := tx.Prepare("INSERT INTO documents (content, embedding) VALUES (?, ?)")
    if err != nil {
        return err
    }
    defer stmt.Close()
    
    for _, doc := range documents {
        embedding, err := getEmbedding(doc.Content)
        if err != nil {
            return err
        }
        
        embeddingJSON, _ := json.Marshal(embedding)
        _, err = stmt.Exec(doc.Content, string(embeddingJSON))
        if err != nil {
            return err
        }
    }
    
    return tx.Commit()
}
```

### 2. Custom Distance Metrics

```go
// Euclidean distance
func euclideanDistance(a, b []float64) float64 {
    var sum float64
    for i := 0; i < len(a); i++ {
        diff := a[i] - b[i]
        sum += diff * diff
    }
    return math.Sqrt(sum)
}

// Manhattan distance
func manhattanDistance(a, b []float64) float64 {
    var sum float64
    for i := 0; i < len(a); i++ {
        sum += math.Abs(a[i] - b[i])
    }
    return sum
}
```

### 3. Approximate Nearest Neighbors

For large datasets, consider:
- Pre-filtering with metadata
- Locality Sensitive Hashing (LSH)
- Vector quantization
- External libraries like Faiss

## Next Steps

1. **Experiment with different embedding models** in Ollama
2. **Add more distance metrics** (Euclidean, Manhattan, etc.)
3. **Implement approximate search** for better performance
4. **Add vector indexing** for large datasets
5. **Build a web interface** for easier interaction

## Resources

- [SQLite Documentation](https://sqlite.org/docs.html)
- [go-sqlite3 Library](https://github.com/mattn/go-sqlite3)
- [Ollama API Reference](https://github.com/ollama/ollama/blob/main/docs/api.md)
- [Vector Similarity Metrics](https://en.wikipedia.org/wiki/Cosine_similarity)

## Example Projects to Build

1. **Document Search Engine**: Index PDF/text files and search by content
2. **Code Similarity Finder**: Find similar code snippets in a codebase
3. **Product Recommendation**: Find similar products based on descriptions
4. **FAQ Chatbot**: Match questions to similar previously asked questions
5. **Content Deduplication**: Find and remove duplicate articles/posts

This foundation gives you everything needed to build sophisticated vector search applications with SQLite and Go!
