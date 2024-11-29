package parser

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/textract"
)

// blockImpl implements the Block interface
type blockImpl struct {
	id              string
	blockType       BlockType
	entityTypes     []EntityType
	text            string
	textType        string
	confidence      float64
	geometry        Geometry
	children        []Block
	parents         []Block
	relationships   []Relationship
	document        Document
	rawBlock        *textract.Block
	page            int
	rowIndex        int
	columnIndex     int
	rowSpan         int
	columnSpan      int
	selectionStatus string
}

func newBlock(raw *textract.Block, doc Document) (Block, error) {
	if raw == nil {
		return nil, fmt.Errorf("raw block cannot be nil")
	}

	block := &blockImpl{
		id:         stringValue(raw.Id),
		blockType:  BlockType(stringValue(raw.BlockType)),
		text:       stringValue(raw.Text),
		textType:   stringValue(raw.TextType),
		confidence: floatValue(raw.Confidence),
		document:   doc,
		rawBlock:   raw,
		page:       int(intValue(raw.Page)),
	}

	// Handle entity types
	if raw.EntityTypes != nil {
		block.entityTypes = make([]EntityType, len(raw.EntityTypes))
		for i, et := range raw.EntityTypes {
			block.entityTypes[i] = EntityType(stringValue(et))
		}
	}

	// Handle table specific fields
	block.rowIndex = int(intValue(raw.RowIndex))
	block.columnIndex = int(intValue(raw.ColumnIndex))
	block.rowSpan = int(intValue(raw.RowSpan))
	block.columnSpan = int(intValue(raw.ColumnSpan))

	block.selectionStatus = stringValue(raw.SelectionStatus)

	if raw.Geometry != nil {
		block.geometry = convertGeometry(raw.Geometry)
	}

	// Handle relationships
	if raw.Relationships != nil {
		block.relationships = make([]Relationship, len(raw.Relationships))
		for i, rel := range raw.Relationships {
			block.relationships[i] = Relationship{
				Type: stringValue(rel.Type),
				IDs:  stringSliceValue(rel.Ids),
			}
		}
	}

	return block, nil
}

// ID returns the block's unique identifier
func (b *blockImpl) ID() string {
	return b.id
}

// BlockType returns the type of block
func (b *blockImpl) BlockType() BlockType {
	return b.blockType
}

// EntityTypes returns the entity types (if any)
func (b *blockImpl) EntityTypes() []EntityType {
	return b.entityTypes
}

// Text returns the block's text content
func (b *blockImpl) Text() string {
	switch b.blockType {
	case BlockTypeWord, BlockTypeLine, BlockTypeSelectionElement:
		// These block types have direct text content
		return b.text

	case BlockTypeCell:
		// For lines and cells, concatenate child word texts
		var texts []string
		for _, child := range b.children {
			if text := child.Text(); text != "" {
				texts = append(texts, text)
			}
		}
		return strings.Join(texts, " ")

	case BlockTypeTable:
		// For tables, build a text representation row by row
		var rows []string
		for _, child := range b.children {
			if child.BlockType() == BlockTypeCell && child.RowIndex() == len(rows) {
				rows = append(rows, child.Text())
			}
		}
		return strings.Join(rows, "\n")

	case BlockTypeKeyValueSet:
		// For key-value sets, combine key and value texts
		var parts []string
		for _, child := range b.children {
			if text := child.Text(); text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, ": ")

	case BlockTypePage:
		// For pages, combine all line texts
		var lines []string
		for _, child := range b.children {
			if child.BlockType() == BlockTypeLine {
				if text := child.Text(); text != "" {
					lines = append(lines, text)
				}
			}
		}
		return strings.Join(lines, "\n")

	default:
		// For other block types, return raw text if any
		return b.text
	}
}

// TextType returns the text type (if any)
func (b *blockImpl) TextType() string {
	return b.textType
}

// Confidence returns the confidence score
func (b *blockImpl) Confidence() float64 {
	return b.confidence
}

// Children returns child blocks
func (b *blockImpl) Children() []Block {
	return b.children
}

// Parents returns parent blocks
func (b *blockImpl) Parents() []Block {
	return b.parents
}

// Relationships returns relationships
func (b *blockImpl) Relationships() []Relationship {
	return b.relationships
}

// BoundingBox returns the block's bounding box
func (b *blockImpl) BoundingBox() BoundingBox {
	return b.geometry.BoundingBox
}

// Polygon returns the block's polygon points
func (b *blockImpl) Polygon() []Point {
	return b.geometry.Polygon
}

// Page returns the block's page number
func (b *blockImpl) Page() int {
	return b.page
}

// RowIndex returns the block's row index
func (b *blockImpl) RowIndex() int {
	return b.rowIndex
}

// ColumnIndex returns the block's column index
func (b *blockImpl) ColumnIndex() int {
	return b.columnIndex
}

// RowSpan returns the block's row span
func (b *blockImpl) RowSpan() int {
	return b.rowSpan
}

// ColumnSpan returns the block's column span
func (b *blockImpl) ColumnSpan() int {
	return b.columnSpan
}

// SelectionStatus returns the block's selection status
func (b *blockImpl) SelectionStatus() string {
	return b.selectionStatus
}

// Helper functions

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func floatValue(f *float64) float64 {
	if f == nil {
		return 0.0
	}
	return *f
}

func convertGeometry(g *textract.Geometry) Geometry {
	if g == nil {
		return Geometry{}
	}

	geo := Geometry{}

	if g.BoundingBox != nil {
		geo.BoundingBox = BoundingBox{
			Width:  floatValue(g.BoundingBox.Width),
			Height: floatValue(g.BoundingBox.Height),
			Left:   floatValue(g.BoundingBox.Left),
			Top:    floatValue(g.BoundingBox.Top),
		}
	}

	if g.Polygon != nil {
		geo.Polygon = make([]Point, len(g.Polygon))
		for i, p := range g.Polygon {
			geo.Polygon[i] = Point{
				X: floatValue(p.X),
				Y: floatValue(p.Y),
			}
		}
	}

	return geo
}

func intValue(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func stringSliceValue(ss []*string) []string {
	if ss == nil {
		return nil
	}
	result := make([]string, len(ss))
	for i, s := range ss {
		result[i] = stringValue(s)
	}
	return result
}
