package parser

import "github.com/aws/aws-sdk-go/service/textract"

// TextractBlock represents the raw AWS Textract Block structure
type TextractBlock struct {
	BlockType     *string
	Confidence    *float64
	Text          *string
	RowIndex      *int64
	ColumnIndex   *int64
	RowSpan       *int64
	ColumnSpan    *int64
	Geometry      *TextractGeometry
	ID            *string
	Relationships []*TextractRelationship
	EntityTypes   []*string
	Page          *int64
}

// TextractGeometry represents the position information for a block
type TextractGeometry struct {
	BoundingBox *TextractBoundingBox
	Polygon     []*TextractPoint
}

// TextractBoundingBox represents a coarse-grained boundary
type TextractBoundingBox struct {
	Width  *float64
	Height *float64
	Left   *float64
	Top    *float64
}

// TextractPoint represents a coordinate pair
type TextractPoint struct {
	X *float64
	Y *float64
}

// TextractRelationship represents a relationship between blocks
type TextractRelationship struct {
	Type *string
	IDs  []*string
}

// TextractDocument represents the top-level document structure
type TextractDocument struct {
	DocumentMetadata *TextractDocumentMetadata
	Blocks           []*TextractBlock
}

// TextractDocumentMetadata contains document-level metadata
type TextractDocumentMetadata struct {
	Pages *int64
}

// ConvertFromAWS converts from AWS Textract types to our internal types
func ConvertFromAWS(response *textract.GetDocumentAnalysisOutput) (*TextractDocument, error) {
	// Implementation will go here
	panic("Not implemented")
	return nil, nil
}
