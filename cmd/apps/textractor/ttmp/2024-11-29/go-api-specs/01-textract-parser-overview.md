# AWS Textract Go Library Specification

## Overview
This specification defines a Go library for processing AWS Textract responses, providing both low-level access to raw Textract data and high-level abstractions for document processing.

## Design Principles

1. **Dual Layer Architecture**
   - Low-level layer: Direct mappings to Textract JSON responses
   - High-level layer: Intuitive document traversal APIs

2. **Idiomatic Go**
   - Use Go interfaces for abstraction
   - Follow standard Go error handling patterns
   - Implement standard Go interfaces where appropriate (e.g., Stringer)

3. **Memory Efficiency**
   - Lazy loading of relationships where possible
   - Smart caching of frequently accessed data
   - Clear ownership hierarchy of objects

4. **Type Safety**
   - Strong typing for all Textract concepts
   - Type-safe enums for block types, entity types, etc.
   - Compile-time type checking where possible

5. **Error Handling**
   - Rich error types with context
   - Clear error hierarchies
   - Wrapped errors for better debugging

6. **Extensibility**
   - Plugin system for custom block processors
   - Middleware support for processing pipeline
   - Configuration options for behavior customization

7. **Performance**
   - Efficient relationship traversal
   - Index important relationships on document load
   - Minimize redundant computations

## Key Components

### 1. Core Types
- Document: Top-level container
- Page: Single page representation
- Block: Basic unit of Textract data
- Relationship: Connection between blocks

### 2. Content Types
- Line: Text line container
- Word: Individual word
- Table: Table structure
- Cell: Table cell
- Form: Form container
- KeyValue: Key-value pair

### 3. Special Types
- SelectionElement: Checkboxes and radio buttons
- Signature: Signature detection
- Query: Query and results
- Layout: Document layout elements

### 4. Support Types
- Geometry: Position information
- Confidence: Confidence scoring
- BoundingBox: Location data
- Point: Coordinate data

## Design Decisions

1. **Relationship Handling**
   - Pre-process relationships on document load
   - Build bidirectional navigation
   - Cache common relationship paths

2. **Memory vs Speed Tradeoffs**
   - Index frequently accessed relationships
   - Lazy load less common relationships
   - Clear documentation of performance characteristics

3. **Configuration**
   - Use functional options pattern
   - Allow runtime configuration
   - Support per-operation settings

4. **Error Handling Strategy**
   - Return errors explicitly
   - Use custom error types
   - Include operation context


# Textract Go Client Library Design RFC

## Overview
The library should provide two main layers:
1. A low-level layer that maps directly to AWS Textract JSON responses
2. A high-level abstraction layer that provides an intuitive API for document traversal

## Design Philosophy

### Low-Level Layer
- Direct struct mappings to Textract JSON responses
- Minimal processing, maximum fidelity
- Useful for cases where raw access is needed
- Tagged for JSON unmarshaling

Example:
```go
type Block struct {
    BlockType    string       `json:"BlockType"`
    Confidence   float64      `json:"Confidence"`
    Text         string       `json:"Text"`
    Relationships []Relationship `json:"Relationships"`
    // ... other fields
}
```

### High-Level Layer
- Built on top of low-level layer
- Provides intuitive navigation of document structure
- Handles relationship traversal internally
- Focuses on developer ergonomics

Example:
```go
// Instead of manually traversing relationships
doc := document.New(rawResponse)
for _, page := range doc.Pages() {
    for _, line := range page.Lines() {
        fmt.Printf("Line: %s\n", line.Text())
    }
}
```

## Key Components

### Document
Central object that manages both raw response and provides high-level access:

```go
type Document struct {
    raw *TextractResponse  // Low-level access
    pages []*Page         // Processed pages for high-level access
}

func New(response *TextractResponse) *Document
func (d *Document) Pages() []*Page
func (d *Document) Raw() *TextractResponse
```

### Page
Represents a single page with convenient accessors:

```go
type Page struct {
    doc *Document
    block *Block
}

func (p *Page) Lines() []*Line
func (p *Page) Tables() []*Table
func (p *Page) Forms() []*Form
```

### Navigation
All objects should support both forward and backward navigation:

```go
type Line struct {
    page *Page
    block *Block
}

func (l *Line) Words() []*Word
func (l *Line) Page() *Page
```

## Error Handling
- Use explicit error returns for operations that can fail
- Provide rich error types for different failure modes
- Include context in errors for easier debugging

## Extensibility
- Allow custom processors for special block types
- Support middleware for custom processing
- Enable custom confidence thresholds

## Example Usage

Basic usage:
```go
doc := textract.New(response)
for _, page := range doc.Pages() {
    // Process lines
    for _, line := range page.Lines() {
        fmt.Printf("Line: %s (%.2f%%)\n", line.Text(), line.Confidence())
    }
    
    // Process tables
    for _, table := range page.Tables() {
        for i, row := range table.Rows() {
            for j, cell := range row.Cells() {
                fmt.Printf("Cell[%d][%d]: %s\n", i, j, cell.Text())
            }
        }
    }
}
```

Advanced usage with options:
```go
doc := textract.New(response, 
    textract.WithConfidenceThreshold(0.9),
    textract.WithCustomProcessor("CUSTOM_BLOCK", myProcessor),
)
```