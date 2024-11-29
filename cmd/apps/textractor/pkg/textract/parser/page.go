package parser

import (
	"fmt"
	"slices"
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

	page := &pageImpl{
		document:   doc,
		block:      block,
		number:     number,
		blockIndex: make(map[string]Block),
	}

	// Process all blocks to build page structure
	if err := page.processBlocks(); err != nil {
		return nil, fmt.Errorf("processing page blocks: %w", err)
	}

	return page, nil
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

		// Add words from line's children
		for _, child := range block.Children() {
			if child.BlockType() == BlockTypeWord {
				p.words = append(p.words, child.Text())
			}
		}

	case BlockTypeTable:
		table, err := newTable(block, p)
		if err != nil {
			return fmt.Errorf("creating table: %w", err)
		}
		p.tables = append(p.tables, table)

	case BlockTypeKeyValueSet:
		// Key-value sets are processed later in buildForms()
		return nil

	case BlockTypeWord:
		// Individual words are added through their parent lines
		return nil

	case BlockTypeSelectionElement:
		// Selection elements are processed as part of forms
		return nil

	default:
		return fmt.Errorf("unexpected block type on page: %s", block.BlockType())
	}

	return nil
}

func (p *pageImpl) processBlocks() error {
	// First process all direct children of the page
	for _, child := range p.block.Children() {
		if err := p.addBlock(child); err != nil {
			return fmt.Errorf("adding block: %w", err)
		}

		// Index the child and its descendants
		if err := p.indexBlockTree(child); err != nil {
			return fmt.Errorf("indexing block tree: %w", err)
		}
	}

	// Build forms after all blocks are processed and indexed
	if err := p.buildForms(); err != nil {
		return fmt.Errorf("building forms: %w", err)
	}

	return nil
}

func (p *pageImpl) indexBlockTree(block Block) error {
	// Add this block to the index
	p.blockIndex[block.ID()] = block

	// Recursively index all children
	for _, child := range block.Children() {
		if err := p.indexBlockTree(child); err != nil {
			return err
		}
	}

	return nil
}

func (p *pageImpl) buildForms() error {
	// Find all key-value sets that are keys (not values)
	var keyBlocks []Block
	for _, block := range p.blockIndex {
		if block.BlockType() == BlockTypeKeyValueSet &&
			slices.Contains(block.EntityTypes(), EntityTypeKey) {
			keyBlocks = append(keyBlocks, block)
		}
	}

	if len(keyBlocks) == 0 {
		return nil // No forms on this page
	}

	// Create a single form for the page
	form := newForm(p)
	p.forms = append(p.forms, form)

	// Process each key block
	for _, keyBlock := range keyBlocks {
		// Find corresponding value block through parent relationships
		for _, parent := range keyBlock.Parents() {
			if parent.BlockType() == BlockTypeKeyValueSet &&
				slices.Contains(parent.EntityTypes(), EntityTypeValue) {
				// Create key-value pair
				kv, err := newKeyValue(keyBlock, parent, form)
				if err != nil {
					return fmt.Errorf("creating key-value pair: %w", err)
				}

				// Add to form
				formImpl := form.(*formImpl)
				formImpl.addField(kv)
				break
			}
		}
	}

	return nil
}
