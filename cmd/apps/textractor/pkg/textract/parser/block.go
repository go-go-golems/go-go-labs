package parser

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/textract"
)

// blockImpl implements the Block interface
type blockImpl struct {
	id         string
	blockType  BlockType
	entityType EntityType
	text       string
	confidence float64
	geometry   Geometry
	children   []Block
	parents    []Block
	document   Document
	rawBlock   *textract.Block
}

func newBlock(raw *textract.Block, doc Document) (Block, error) {
	if raw == nil {
		return nil, fmt.Errorf("raw block cannot be nil")
	}

	block := &blockImpl{
		id:         *raw.Id,
		blockType:  BlockType(*raw.BlockType),
		text:       stringValue(raw.Text),
		confidence: floatValue(raw.Confidence),
		document:   doc,
		rawBlock:   raw,
	}

	if len(raw.EntityTypes) > 0 {
		block.entityType = EntityType(*raw.EntityTypes[0])
	}

	if raw.Geometry != nil {
		block.geometry = convertGeometry(raw.Geometry)
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

// EntityType returns the entity type (if any)
func (b *blockImpl) EntityType() EntityType {
	return b.entityType
}

// Text returns the block's text content
func (b *blockImpl) Text() string {
	return b.text
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

// BoundingBox returns the block's bounding box
func (b *blockImpl) BoundingBox() BoundingBox {
	return b.geometry.BoundingBox
}

// Polygon returns the block's polygon points
func (b *blockImpl) Polygon() []Point {
	return b.geometry.Polygon
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
