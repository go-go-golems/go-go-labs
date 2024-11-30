package parser

// BlockType represents the type of a Textract block
type BlockType string

// EntityType represents the type of entity in a Textract block
type EntityType string

const (
	BlockTypePage             BlockType = "PAGE"
	BlockTypeLine             BlockType = "LINE"
	BlockTypeWord             BlockType = "WORD"
	BlockTypeTable            BlockType = "TABLE"
	BlockTypeCell             BlockType = "CELL"
	BlockTypeKeyValueSet      BlockType = "KEY_VALUE_SET"
	BlockTypeSelectionElement BlockType = "SELECTION_ELEMENT"
	BlockTypeSignature        BlockType = "SIGNATURE"
	BlockTypeQuery            BlockType = "QUERY"
	BlockTypeQueryResult      BlockType = "QUERY_RESULT"
)

const (
	EntityTypeKey                 EntityType = "KEY"
	EntityTypeValue               EntityType = "VALUE"
	EntityTypeColumnHeader        EntityType = "COLUMN_HEADER"
	EntityTypeTableTitle          EntityType = "TABLE_TITLE"
	EntityTypeTableSectionTitle   EntityType = "TABLE_SECTION_TITLE"
	EntityTypeTableFooter         EntityType = "TABLE_FOOTER"
	EntityTypeTableSummary        EntityType = "TABLE_SUMMARY"
	EntityTypeStructuredTable     EntityType = "STRUCTURED_TABLE"
	EntityTypeSemiStructuredTable EntityType = "SEMI_STRUCTURED_TABLE"
)

const (
	SelectionStatusSelected    SelectionStatus = "SELECTED"
	SelectionStatusNotSelected SelectionStatus = "NOT_SELECTED"
)

// Geometry represents position information for blocks
type Geometry struct {
	BoundingBox BoundingBox
	Polygon     []Point
}

// BoundingBox represents a coarse-grained boundary
type BoundingBox struct {
	Width  float64
	Height float64
	Left   float64
	Top    float64
}

// Point represents a coordinate pair
type Point struct {
	X float64
	Y float64
}

// DocumentMetadata contains document-level metadata
type DocumentMetadata struct {
	Pages int
	// Other metadata fields can be added here
}

// FilterOptions provides filtering criteria for blocks
type FilterOptions struct {
	MinConfidence float64
	BlockTypes    []BlockType
	EntityTypes   []EntityType
}

// DocumentOptions configures document processing behavior
type DocumentOptions struct {
	ConfidenceThreshold float64
	EnableMergedCells   bool
	CustomProcessors    map[BlockType]BlockProcessor
}
