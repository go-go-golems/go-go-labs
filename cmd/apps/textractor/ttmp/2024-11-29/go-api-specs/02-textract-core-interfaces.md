# Core Interfaces and Types

## Base Types

### Document
Primary container for Textract response data.

```go
type Document interface {
    // Core access
    Pages() []Page
    Raw() *TextractResponse
    
    // Navigation
    GetPageByIndex(idx int) (Page, error)
    
    // Search and filtering
    FindKeyValuePairs(key string) []KeyValue
    FilterBlocks(opts FilterOptions) []Block
    
    // Metadata
    PageCount() int
    DocumentMetadata() DocumentMetadata
}

type DocumentMetadata struct {
    Pages int
    // Other metadata fields
}
```

### Page
Represents a single document page.

```go
type Page interface {
    // Content access
    Lines() []Line
    Tables() []Table
    Forms() []Form
    Words() []Word
    
    // Layout elements
    GetLayoutElements(layoutType LayoutType) []LayoutElement
    
    // Navigation
    Document() Document
    Number() int
    
    // Geometry
    BoundingBox() BoundingBox
    Polygon() []Point
}
```

### Block
Basic unit of Textract data.

```go
type Block interface {
    // Identification
    ID() string
    BlockType() BlockType
    EntityType() EntityType
    
    // Content
    Text() string
    Confidence() float64
    
    // Relationships
    Relationships() []Relationship
    Children() []Block
    Parents() []Block
    
    // Geometry
    BoundingBox() BoundingBox
    Polygon() []Point
}
```

## Content Types

### Line
```go
type Line interface {
    // Content
    Text() string
    Words() []Word
    Confidence() float64
    
    // Navigation
    Page() Page
    
    // Geometry
    BoundingBox() BoundingBox
    Polygon() []Point
}
```

### Table
```go
type Table interface {
    // Structure
    Rows() []TableRow
    Cells() [][]Cell
    MergedCells() []MergedCell
    
    // Metadata
    RowCount() int
    ColumnCount() int
    
    // Special elements
    Title() *TableTitle
    Footer() *TableFooter
    
    // Navigation
    Page() Page
    GetCellByPosition(row, col int) (Cell, error)
}
```

### Form
```go
type Form interface {
    // Content
    Fields() []KeyValue
    SelectionElements() []SelectionElement
    
    // Search
    GetFieldByKey(key string) KeyValue
    SearchFieldsByKey(key string) []KeyValue
    
    // Navigation
    Page() Page
}
```

### KeyValue
```go
type KeyValue interface {
    // Content
    Key() Block
    Value() Block
    
    // Helper methods
    KeyText() string
    ValueText() string
    Confidence() float64
    
    // Navigation
    Form() Form
}
```

## Special Types

### SelectionElement
```go
type SelectionElement interface {
    // Status
    IsSelected() bool
    SelectionStatus() SelectionStatus
    Confidence() float64
    
    // Navigation
    Block() Block
    Page() Page
}
```

### Query
```go
type Query interface {
    // Content
    Text() string
    Alias() string
    Results() []QueryResult
    
    // Navigation
    Page() Page
}
```

## Support Types

### Geometry
```go
type BoundingBox struct {
    Width  float64
    Height float64
    Left   float64
    Top    float64
}

type Point struct {
    X float64
    Y float64
}

type Geometry struct {
    BoundingBox BoundingBox
    Polygon     []Point
}
```

### Options
```go
type FilterOptions struct {
    MinConfidence float64
    BlockTypes    []BlockType
    EntityTypes   []EntityType
}

type DocumentOptions struct {
    ConfidenceThreshold float64
    EnableMergedCells   bool
    CustomProcessors    map[BlockType]BlockProcessor
}
```
