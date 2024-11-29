package parser

import (
	"fmt"
)

// pageImpl implements the Page interface
type pageImpl struct {
	document   Document
	block      Block
	number     int
	lines      []Line
	tables     []Table
	forms      []Form
	words      []string
	blockIndex map[string]Block
}

func newPage(doc Document, block Block, number int) (Page, error) {
	if block.BlockType() != BlockTypePage {
		return nil, fmt.Errorf("block type must be PAGE, got %s", block.BlockType())
	}

	return &pageImpl{
		document:   doc,
		block:      block,
		number:     number,
		blockIndex: make(map[string]Block),
	}, nil
}

// Lines returns all lines on the page
func (p *pageImpl) Lines() []Line {
	return p.lines
}

// Tables returns all tables on the page
func (p *pageImpl) Tables() []Table {
	return p.tables
}

// Forms returns all forms on the page
func (p *pageImpl) Forms() []Form {
	return p.forms
}

// Words returns all words on the page
func (p *pageImpl) Words() []string {
	return p.words
}

// Document returns the parent document
func (p *pageImpl) Document() Document {
	return p.document
}

// Number returns the page number
func (p *pageImpl) Number() int {
	return p.number
}

// BoundingBox returns the page's bounding box
func (p *pageImpl) BoundingBox() BoundingBox {
	return p.block.BoundingBox()
}

// Polygon returns the page's polygon points
func (p *pageImpl) Polygon() []Point {
	return p.block.Polygon()
}

// Internal methods

func (p *pageImpl) addBlock(block Block) error {
	p.blockIndex[block.ID()] = block

	switch block.BlockType() {
	case BlockTypeLine:
		line, err := newLine(block, p)
		if err != nil {
			return fmt.Errorf("creating line: %w", err)
		}
		p.lines = append(p.lines, line)

	case BlockTypeTable:
		table, err := newTable(block, p)
		if err != nil {
			return fmt.Errorf("creating table: %w", err)
		}
		p.tables = append(p.tables, table)

	case BlockTypeWord:
		p.words = append(p.words, block.Text())
	}

	return nil
}

func (p *pageImpl) processBlocks() error {
	// Process child blocks
	for _, child := range p.block.Children() {
		if err := p.addBlock(child); err != nil {
			return fmt.Errorf("adding block: %w", err)
		}
	}

	// Build forms after all blocks are processed
	if err := p.buildForms(); err != nil {
		return fmt.Errorf("building forms: %w", err)
	}

	return nil
}

func (p *pageImpl) buildForms() error {
	// Group key-value sets into forms
	// Implementation will go here
	return nil
}
