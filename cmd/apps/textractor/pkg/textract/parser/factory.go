package parser

import (
	"context"
	"github.com/aws/aws-sdk-go/service/textract"
)

// TextractResponse represents the raw response from AWS Textract
type TextractResponse struct {
	Blocks []*textract.Block
	// Add other fields as needed
}

// Parser is the main interface for creating Document instances
type Parser interface {
	ParseDocument(ctx context.Context, response *TextractResponse, opts ...DocumentOption) (Document, error)
}

// DocumentOption configures document parsing behavior
type DocumentOption func(*DocumentOptions)

// BlockProcessor is an interface for custom block processing
type BlockProcessor interface {
	ProcessBlock(ctx context.Context, block *textract.Block) (Block, error)
}
