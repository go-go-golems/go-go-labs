package parser

import "fmt"

// lineImpl implements the Line interface
type lineImpl struct {
	block Block
	page  Page
	words []string
}

func newLine(block Block, page Page) (Line, error) {
	if block.BlockType() != BlockTypeLine {
		return nil, fmt.Errorf("block type must be LINE, got %s", block.BlockType())
	}

	line := &lineImpl{
		block: block,
		page:  page,
	}

	// Process child words
	for _, child := range block.Children() {
		if child.BlockType() == BlockTypeWord {
			line.words = append(line.words, child.Text())
		}
	}

	return line, nil
}

// Text returns the line's text content
func (l *lineImpl) Text() string {
	return l.block.Text()
}

// Words returns the individual words in the line
func (l *lineImpl) Words() []string {
	return l.words
}

// Confidence returns the confidence score
func (l *lineImpl) Confidence() float64 {
	return l.block.Confidence()
}

// Page returns the parent page
func (l *lineImpl) Page() Page {
	return l.page
}

// BoundingBox returns the line's bounding box
func (l *lineImpl) BoundingBox() BoundingBox {
	return l.block.BoundingBox()
}

// Polygon returns the line's polygon points
func (l *lineImpl) Polygon() []Point {
	return l.block.Polygon()
}
