package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/textract"
)

// JSONBlock represents a Textract block in JSON format
type JSONBlock struct {
	ID              string             `json:"Id"`
	BlockType       string             `json:"BlockType"`
	EntityTypes     []string           `json:"EntityTypes,omitempty"`
	Text            string             `json:"Text,omitempty"`
	Confidence      float64            `json:"Confidence"`
	Geometry        *JSONGeometry      `json:"Geometry"`
	Relationships   []JSONRelationship `json:"Relationships,omitempty"`
	RowIndex        int                `json:"RowIndex,omitempty"`
	ColumnIndex     int                `json:"ColumnIndex,omitempty"`
	RowSpan         int                `json:"RowSpan,omitempty"`
	ColumnSpan      int                `json:"ColumnSpan,omitempty"`
	SelectionStatus string             `json:"SelectionStatus,omitempty"`
	Page            int                `json:"Page,omitempty"`
}

// JSONGeometry represents geometry information in JSON format
type JSONGeometry struct {
	BoundingBox *JSONBoundingBox `json:"BoundingBox"`
	Polygon     []JSONPoint      `json:"Polygon"`
}

// JSONBoundingBox represents a bounding box in JSON format
type JSONBoundingBox struct {
	Width  float64 `json:"Width"`
	Height float64 `json:"Height"`
	Left   float64 `json:"Left"`
	Top    float64 `json:"Top"`
}

// JSONPoint represents a point in JSON format
type JSONPoint struct {
	X float64 `json:"X"`
	Y float64 `json:"Y"`
}

// JSONRelationship represents a relationship in JSON format
type JSONRelationship struct {
	Type string   `json:"Type"`
	IDs  []string `json:"Ids"`
}

// JSONResponse represents a complete Textract response in JSON format
type JSONResponse struct {
	DocumentMetadata *JSONDocumentMetadata `json:"DocumentMetadata"`
	Blocks           []*JSONBlock          `json:"Blocks"`
}

// JSONDocumentMetadata represents document metadata in JSON format
type JSONDocumentMetadata struct {
	Pages int `json:"Pages"`
}

// LoadFromJSON creates Documents from a JSON file
func LoadFromJSON(filename string) ([]Document, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening JSON file: %w", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	return LoadFromJSONReader(file)
}

// LoadFromJSONReader creates Documents from a JSON reader
func LoadFromJSONReader(r io.Reader) ([]Document, error) {
	// First try to decode as array
	var jsonResps []JSONResponse
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&jsonResps); err != nil {
		// If it fails, try as single document
		if _, ok := err.(*json.UnmarshalTypeError); ok {
			// Reset reader to start
			if seeker, ok := r.(io.Seeker); ok {
				_, err = seeker.Seek(0, 0)
				if err != nil {
					return nil, fmt.Errorf("seeking reader: %w", err)
				}
			} else {
				return nil, fmt.Errorf("reader is not seekable and JSON is not an array")
			}

			var singleResp JSONResponse
			if err := json.NewDecoder(r).Decode(&singleResp); err != nil {
				return nil, fmt.Errorf("decoding JSON: %w", err)
			}
			jsonResps = []JSONResponse{singleResp}
		} else {
			return nil, fmt.Errorf("decoding JSON: %w", err)
		}
	}

	docs := make([]Document, len(jsonResps))
	for i, jsonResp := range jsonResps {
		// Convert JSON response to TextractResponse
		resp := &TextractResponse{
			Blocks: make([]*textract.Block, len(jsonResp.Blocks)),
			Metadata: &textract.DocumentMetadata{
				Pages: aws.Int64(int64(jsonResp.DocumentMetadata.Pages)),
			},
		}

		// Convert blocks
		for j, jsonBlock := range jsonResp.Blocks {
			block, err := convertJSONBlock(jsonBlock)
			if err != nil {
				return nil, fmt.Errorf("converting block %s in document %d: %w", jsonBlock.ID, i, err)
			}
			resp.Blocks[j] = block
		}

		// Create document using existing parser
		doc, err := NewDocument(resp)
		if err != nil {
			return nil, fmt.Errorf("creating document %d: %w", i, err)
		}
		docs[i] = doc
	}

	return docs, nil
}

// Helper functions to convert JSON types to Textract types

func convertJSONBlock(jb *JSONBlock) (*textract.Block, error) {
	block := &textract.Block{
		Id:              &jb.ID,
		BlockType:       &jb.BlockType,
		Text:            &jb.Text,
		Confidence:      &jb.Confidence,
		Page:            aws.Int64(int64(jb.Page)),
		SelectionStatus: &jb.SelectionStatus,
	}

	if jb.Geometry != nil {
		block.Geometry = convertJSONGeometry(jb.Geometry)
	}

	if len(jb.EntityTypes) > 0 {
		block.EntityTypes = make([]*string, len(jb.EntityTypes))
		for i, et := range jb.EntityTypes {
			block.EntityTypes[i] = &et
		}
	}

	if len(jb.Relationships) > 0 {
		block.Relationships = make([]*textract.Relationship, len(jb.Relationships))
		for i, rel := range jb.Relationships {
			block.Relationships[i] = convertJSONRelationship(&rel)
		}
	}

	// Handle table cell specific fields
	if jb.RowIndex != 0 {
		block.RowIndex = aws.Int64(int64(jb.RowIndex))
	}
	if jb.ColumnIndex != 0 {
		block.ColumnIndex = aws.Int64(int64(jb.ColumnIndex))
	}
	if jb.RowSpan != 0 {
		block.RowSpan = aws.Int64(int64(jb.RowSpan))
	}
	if jb.ColumnSpan != 0 {
		block.ColumnSpan = aws.Int64(int64(jb.ColumnSpan))
	}

	return block, nil
}

func convertJSONGeometry(jg *JSONGeometry) *textract.Geometry {
	geo := &textract.Geometry{}

	if jg.BoundingBox != nil {
		geo.BoundingBox = &textract.BoundingBox{
			Width:  &jg.BoundingBox.Width,
			Height: &jg.BoundingBox.Height,
			Left:   &jg.BoundingBox.Left,
			Top:    &jg.BoundingBox.Top,
		}
	}

	if len(jg.Polygon) > 0 {
		geo.Polygon = make([]*textract.Point, len(jg.Polygon))
		for i, p := range jg.Polygon {
			geo.Polygon[i] = &textract.Point{
				X: &p.X,
				Y: &p.Y,
			}
		}
	}

	return geo
}

func convertJSONRelationship(jr *JSONRelationship) *textract.Relationship {
	rel := &textract.Relationship{
		Type: &jr.Type,
	}

	if len(jr.IDs) > 0 {
		rel.Ids = make([]*string, len(jr.IDs))
		for i, id := range jr.IDs {
			rel.Ids[i] = &id
		}
	}

	return rel
}
