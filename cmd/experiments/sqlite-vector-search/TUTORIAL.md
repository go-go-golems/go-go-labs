# SQLite Custom Functions in Go: A Complete Tutorial

This tutorial explains how to create custom SQLite functions in Go, specifically for vector similarity search using embeddings from Ollama.

## Table of Contents

1. [Understanding SQLite Custom Functions](#understanding-sqlite-custom-functions)
2. [Go SQLite Integration](#go-sqlite-integration)
3. [Vector Similarity Concepts](#vector-similarity-concepts)
4. [Step-by-Step Implementation](#step-by-step-implementation)
5. [Ollama Integration](#ollama-integration)
6. [Complete Working Example](#complete-working-example)
7. [Testing and Debugging](#testing-and-debugging)

## Understanding SQLite Custom Functions

### What are SQLite Custom Functions?

SQLite allows you to register custom functions that can be called from SQL queries. These functions are written in the host language (Go in our case) and can perform complex calculations that would be difficult or impossible in pure SQL.

### Why Use Custom Functions for Vector Search?

Vector similarity search requires:
- Complex mathematical operations (dot products, norms)
- Working with array data structures
- Efficient similarity calculations

While SQLite has JSON functions, performing vector math in SQL would be:
- Extremely verbose and hard to read
- Inefficient for large vectors
- Difficult to maintain and debug

### Function Registration Process

1. **Connection Access**: Get direct access to the SQLite connection
2. **Function Registration**: Register your Go function with SQLite
3. **Type Mapping**: Handle data type conversion between Go and SQLite
4. **SQL Usage**: Call the function from SQL queries

## Go SQLite Integration

### Required Dependencies

```go
import (
    "database/sql"
    "github.com/mattn/go-sqlite3"
    _ "github.com/mattn/go-sqlite3" // SQLite driver
)
```

### Key Concepts

#### 1. Database Connection Hierarchy

```
sql.DB (Go standard library)
    ↓
sql.Conn (Connection wrapper)
    ↓
sqlite3.SQLiteConn (Driver-specific connection)
```

#### 2. Raw Connection Access

To register custom functions, you need access to the underlying SQLite connection:

```go
func registerSQLiteFunctions(db *sql.DB) error {
    // Get a connection from the pool
    sqliteConn, err := db.Conn(context.Background())
    if err != nil {
        return err
    }
    defer sqliteConn.Close()

    // Access the raw driver connection
    return sqliteConn.Raw(func(driverConn interface{}) error {
        conn := driverConn.(*sqlite3.SQLiteConn)
        // Now we can register functions
        return conn.RegisterFunc("my_function", myFunction, true)
    })
}
```

#### 3. Function Registration Parameters

```go
conn.RegisterFunc(name, function, pure)
```

- **name**: SQL function name
- **function**: Go function to execute
- **pure**: Whether function is deterministic (same input → same output)

## Vector Similarity Concepts

### What are Embeddings?

Embeddings are numerical representations of text where:
- Similar meanings → similar vectors
- Each dimension captures semantic features
- Typically 100-1500 dimensions

Example:
```
"cat" → [0.2, -0.1, 0.8, ...]
"dog" → [0.3, -0.1, 0.7, ...]  // Similar to "cat"
"car" → [-0.5, 0.9, -0.2, ...] // Different from animals
```

### Cosine Similarity

Measures the cosine of the angle between two vectors:

```
similarity = (A · B) / (||A|| × ||B||)

Where:
- A · B = dot product
- ||A|| = magnitude/norm of vector A
```

Range: -1 to 1 (1 = identical, 0 = orthogonal, -1 = opposite)

### Why Cosine Similarity?

- **Scale invariant**: Focuses on direction, not magnitude
- **Normalized**: Always between -1 and 1
- **Intuitive**: Higher values = more similar

## Step-by-Step Implementation

### Step 1: Basic Cosine Similarity Function

```go
func cosineSimilarity(a, b []float64) float64 {
    if len(a) != len(b) {
        return 0 // Invalid input
    }

    var dotProduct, normA, normB float64
    
    // Calculate dot product and norms in one pass
    for i := 0; i < len(a); i++ {
        dotProduct += a[i] * b[i]  // A · B
        normA += a[i] * a[i]       // ||A||²
        normB += b[i] * b[i]       // ||B||²
    }

    // Handle zero vectors
    if normA == 0 || normB == 0 {
        return 0
    }

    // Return cosine similarity
    return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
```

### Step 2: SQLite Function Wrapper

SQLite functions receive and return basic types. For vectors, we use JSON encoding:

```go
func sqliteCosineSimilarity(a, b string) float64 {
    var vecA, vecB []float64
    
    // Parse JSON vectors
    if err := json.Unmarshal([]byte(a), &vecA); err != nil {
        return 0 // Return 0 for invalid JSON
    }
    if err := json.Unmarshal([]byte(b), &vecB); err != nil {
        return 0
    }

    return cosineSimilarity(vecA, vecB)
}
```

### Step 3: Function Registration

```go
func registerSQLiteFunctions(db *sql.DB) error {
    sqliteConn, err := db.Conn(context.Background())
    if err != nil {
        return errors.Wrap(err, "failed to get connection")
    }
    defer sqliteConn.Close()

    return sqliteConn.Raw(func(driverConn interface{}) error {
        conn := driverConn.(*sqlite3.SQLiteConn)

        // Register our cosine similarity function
        return conn.RegisterFunc("cosine_similarity", sqliteCosineSimilarity, true)
    })
}
```

### Step 4: Database Schema

```sql
CREATE TABLE documents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content TEXT NOT NULL,
    embedding TEXT NOT NULL  -- JSON array of floats
);
```

### Step 5: Using the Function in SQL

```sql
-- Find most similar documents to a query vector
SELECT 
    id, 
    content, 
    cosine_similarity(embedding, ?) as similarity
FROM documents
ORDER BY similarity DESC
LIMIT 5;
```

## Ollama Integration

### Understanding Ollama API

Ollama provides a local API for running language models:

```
POST http://127.0.0.1:11434/api/embeddings
{
    "model": "all-minilm:latest",
    "prompt": "Your text here"
}

Response:
{
    "embedding": [0.1, -0.2, 0.8, ...]
}
```

### HTTP Client Implementation

```go
type OllamaClient struct {
    BaseURL string
    Model   string
    logger  zerolog.Logger
}

func (c *OllamaClient) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
    // Prepare request
    reqData := OllamaEmbeddingRequest{
        Model:  c.Model,
        Prompt: text,
    }

    jsonData, err := json.Marshal(reqData)
    if err != nil {
        return nil, errors.Wrap(err, "failed to marshal request")
    }

    // Make HTTP request
    url := fmt.Sprintf("%s/api/embeddings", c.BaseURL)
    req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
    if err != nil {
        return nil, errors.Wrap(err, "failed to create request")
    }

    req.Header.Set("Content-Type", "application/json")

    // Execute request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, errors.Wrap(err, "failed to make request")
    }
    defer resp.Body.Close()

    // Parse response
    var response OllamaEmbeddingResponse
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, errors.Wrap(err, "failed to decode response")
    }

    return response.Embedding, nil
}
```

### Error Handling Best Practices

```go
// Always check HTTP status
if resp.StatusCode != http.StatusOK {
    return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
}

// Use context for timeouts
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Wrap errors with context
return nil, errors.Wrap(err, "failed to get embedding from Ollama")
```

## Complete Working Example

### Document Storage

```go
func insertDocument(db *sql.DB, content string, embedding []float64) error {
    // Convert embedding to JSON
    embeddingJSON, err := json.Marshal(embedding)
    if err != nil {
        return errors.Wrap(err, "failed to marshal embedding")
    }

    // Store in database
    _, err = db.Exec(
        "INSERT INTO documents (content, embedding) VALUES (?, ?)", 
        content, 
        string(embeddingJSON),
    )
    return errors.Wrap(err, "failed to insert document")
}
```

### Similarity Search

```go
func searchSimilarDocuments(db *sql.DB, queryEmbedding []float64, limit int) ([]SearchResult, error) {
    // Convert query embedding to JSON
    queryEmbeddingJSON, err := json.Marshal(queryEmbedding)
    if err != nil {
        return nil, errors.Wrap(err, "failed to marshal query embedding")
    }

    // Execute similarity search
    query := `
        SELECT id, content, cosine_similarity(embedding, ?) as similarity
        FROM documents
        ORDER BY similarity DESC
        LIMIT ?
    `

    rows, err := db.Query(query, string(queryEmbeddingJSON), limit)
    if err != nil {
        return nil, errors.Wrap(err, "failed to execute search query")
    }
    defer rows.Close()

    var results []SearchResult
    for rows.Next() {
        var result SearchResult
        if err := rows.Scan(&result.ID, &result.Content, &result.Similarity); err != nil {
            return nil, errors.Wrap(err, "failed to scan row")
        }
        results = append(results, result)
    }

    return results, nil
}
```

## Testing and Debugging

### Function Testing

```go
func TestCosineSimilarity(t *testing.T) {
    // Test identical vectors
    vec1 := []float64{1, 0, 0}
    vec2 := []float64{1, 0, 0}
    similarity := cosineSimilarity(vec1, vec2)
    assert.Equal(t, 1.0, similarity)

    // Test orthogonal vectors
    vec3 := []float64{1, 0, 0}
    vec4 := []float64{0, 1, 0}
    similarity = cosineSimilarity(vec3, vec4)
    assert.Equal(t, 0.0, similarity)
}
```

### SQL Function Testing

```sql
-- Test the registered function directly
SELECT cosine_similarity('[1,0,0]', '[1,0,0]') as similarity;
-- Should return 1.0

SELECT cosine_similarity('[1,0,0]', '[0,1,0]') as similarity;
-- Should return 0.0
```

### Debugging Tips

1. **Check Ollama Connection**:
   ```bash
   curl http://127.0.0.1:11434/api/embeddings \
     -d '{"model":"all-minilm:latest","prompt":"test"}'
   ```

2. **Verify Function Registration**:
   ```go
   // Add logging to confirm registration
   logger.Info().Msg("cosine_similarity function registered successfully")
   ```

3. **Inspect Embeddings**:
   ```go
   logger.Debug().Int("embedding_size", len(embedding)).Msg("got embedding")
   ```

4. **Test JSON Serialization**:
   ```go
   embeddingJSON, _ := json.Marshal(embedding)
   logger.Debug().Str("embedding_json", string(embeddingJSON)).Msg("serialized embedding")
   ```

## Common Pitfalls and Solutions

### 1. Connection Pool Issues

**Problem**: Function registration fails randomly
**Solution**: Use `db.Conn()` to get a dedicated connection

### 2. JSON Parsing Errors

**Problem**: Invalid JSON in database causes function to return 0
**Solution**: Add validation and error logging

```go
func sqliteCosineSimilarity(a, b string) float64 {
    var vecA, vecB []float64
    
    if err := json.Unmarshal([]byte(a), &vecA); err != nil {
        log.Printf("Invalid JSON for vector A: %s", a)
        return 0
    }
    // ... rest of function
}
```

### 3. Vector Dimension Mismatches

**Problem**: Comparing vectors of different sizes
**Solution**: Always validate dimensions

```go
if len(a) != len(b) {
    log.Printf("Vector dimension mismatch: %d vs %d", len(a), len(b))
    return 0
}
```

### 4. Zero Vectors

**Problem**: Division by zero in cosine similarity
**Solution**: Handle zero norms explicitly

```go
if normA == 0 || normB == 0 {
    return 0 // Undefined similarity for zero vectors
}
```

## Performance Considerations

### 1. Function Purity

Mark functions as pure when possible:
```go
conn.RegisterFunc("cosine_similarity", func, true) // true = pure/deterministic
```

Benefits:
- SQLite can cache results
- Better query optimization
- Parallel execution possible

### 2. Index Strategies

For large datasets, consider:
- Approximate similarity search (LSH, Annoy)
- Pre-filtering with metadata
- Batch similarity calculations

### 3. Memory Management

- Use connection pooling appropriately
- Close connections and rows properly
- Consider streaming for large result sets

## Next Steps

To extend this implementation:

1. **Add More Vector Functions**: L2 distance, dot product
2. **Implement Approximate Search**: For better performance on large datasets
3. **Add Vector Indexing**: Using libraries like Faiss
4. **Support Different Embedding Models**: Multiple Ollama models
5. **Add Batch Operations**: Process multiple vectors at once

This tutorial provides a complete foundation for understanding and implementing SQLite custom functions for vector similarity search in Go.
