package parser

import (
	"fmt"
)

// documentImpl implements the Document interface
type documentImpl struct {
	raw        *TextractResponse
	pages      []Page
	metadata   DocumentMetadata
	blockIndex map[string]Block
	pageIndex  map[int]Page
}

// NewDocument creates a new Document from a TextractResponse
func NewDocument(response *TextractResponse, opts ...DocumentOption) (Document, error) {
	if response == nil {
		return nil, fmt.Errorf("textract response cannot be nil")
	}

	// Apply options
	options := &DocumentOptions{
		ConfidenceThreshold: 0.0,
		EnableMergedCells:   true,
	}
	for _, opt := range opts {
		opt(options)
	}

	doc := &documentImpl{
		raw:        response,
		blockIndex: make(map[string]Block),
		pageIndex:  make(map[int]Page),
	}

	// Process blocks and build indexes
	if err := doc.processBlocks(options); err != nil {
		return nil, fmt.Errorf("processing blocks: %w", err)
	}

	return doc, nil
}

// Pages returns all pages in the document
func (d *documentImpl) Pages() []Page {
	return d.pages
}

// Raw returns the underlying TextractResponse
func (d *documentImpl) Raw() *TextractResponse {
	return d.raw
}

// GetPageByIndex returns a specific page by its index
func (d *documentImpl) GetPageByIndex(idx int) (Page, error) {
	if page, ok := d.pageIndex[idx]; ok {
		return page, nil
	}
	return nil, fmt.Errorf("page index %d not found", idx)
}

// FindKeyValuePairs finds all key-value pairs with matching key text
func (d *documentImpl) FindKeyValuePairs(key string) []KeyValue {
	var results []KeyValue
	for _, page := range d.pages {
		for _, form := range page.Forms() {
			kvs := form.SearchFieldsByKey(key)
			results = append(results, kvs...)
		}
	}
	return results
}

// FilterBlocks returns blocks matching the given criteria
func (d *documentImpl) FilterBlocks(opts FilterOptions) []Block {
	var results []Block
	for _, block := range d.blockIndex {
		if d.matchesFilter(block, opts) {
			results = append(results, block)
		}
	}
	return results
}

// PageCount returns the total number of pages
func (d *documentImpl) PageCount() int {
	return d.metadata.Pages
}

// DocumentMetadata returns the document metadata
func (d *documentImpl) DocumentMetadata() DocumentMetadata {
	return d.metadata
}

// Internal helper methods

func (d *documentImpl) processBlocks(options *DocumentOptions) error {
	// First pass: create all blocks and build index
	for _, rawBlock := range d.raw.Blocks {
		block, err := newBlock(rawBlock, d)
		if err != nil {
			return fmt.Errorf("creating block: %w", err)
		}
		d.blockIndex[block.ID()] = block
	}

	// Second pass: process relationships
	for _, block := range d.blockIndex {
		if err := d.processBlockRelationships(block); err != nil {
			return fmt.Errorf("processing relationships: %w", err)
		}
	}

	// Build page hierarchy
	if err := d.buildPages(); err != nil {
		return fmt.Errorf("building pages: %w", err)
	}

	return nil
}

func (d *documentImpl) matchesFilter(block Block, opts FilterOptions) bool {
	if block.Confidence() < opts.MinConfidence {
		return false
	}

	if len(opts.BlockTypes) > 0 {
		matched := false
		for _, bt := range opts.BlockTypes {
			if block.BlockType() == bt {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	if len(opts.EntityTypes) > 0 {
		matched := false
		for _, et := range opts.EntityTypes {
			if block.EntityType() == et {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

func (d *documentImpl) processBlockRelationships(block Block) error {
	// Implementation will process parent/child relationships
	// and other relationship types based on BlockType
	return nil
}

func (d *documentImpl) buildPages() error {
	// Implementation will organize blocks into page hierarchy
	return nil
}
