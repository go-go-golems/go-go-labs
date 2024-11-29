package parser

// Document is the primary container for Textract response data
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

// Page represents a single document page
type Page interface {
	// Content access
	Lines() []Line
	Tables() []Table
	Forms() []Form
	Words() []string

	// Navigation
	Document() Document
	Number() int

	// Geometry
	BoundingBox() BoundingBox
	Polygon() []Point
}

// Block is the basic unit of Textract data
type Block interface {
	// Identification
	ID() string
	BlockType() BlockType
	EntityTypes() []EntityType
	Page() int

	// Content
	Text() string
	TextType() string
	Confidence() float64

	// Table specific
	RowIndex() int
	ColumnIndex() int
	RowSpan() int
	ColumnSpan() int

	// Selection elements
	SelectionStatus() string

	// Relationships
	Children() []Block
	Parents() []Block
	Relationships() []Relationship

	// Geometry
	BoundingBox() BoundingBox
	Polygon() []Point
}

// Line represents a text line
type Line interface {
	// Content
	Text() string
	Words() []string
	Confidence() float64

	// Navigation
	Page() Page

	// Geometry
	BoundingBox() BoundingBox
	Polygon() []Point
}

// Table represents a table structure
type Table interface {
	// Structure
	Rows() []TableRow
	Cells() [][]Cell
	MergedCells() []MergedCell

	// Metadata
	RowCount() int
	ColumnCount() int

	// Navigation
	Page() Page
	GetCellByPosition(row, col int) (Cell, error)

	// Geometry
	BoundingBox() BoundingBox
	Polygon() []Point
}

// Form represents a form structure
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

// KeyValue represents a key-value pair
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

// Query represents a question asked of a document and its results
type Query interface {
	// Text returns the query text
	Text() string

	// Alias returns the query alias if one was specified
	Alias() string

	// Results returns all answers found for this query
	Results() []QueryResult

	// Page returns the parent page containing this query
	Page() Page
}

// QueryResult represents an answer to a query
type QueryResult interface {
	// Text returns the result text
	Text() string

	// Confidence returns the confidence score for this result
	Confidence() float64

	// Query returns the parent query
	Query() Query

	// Block returns the underlying block
	Block() Block
}

// TextType represents the type of text in a block
type TextType string

// SelectionStatus represents the selection status of a block
type SelectionStatus string

// Relationship represents a relationship between blocks
type Relationship struct {
	Type string
	IDs  []string
}
