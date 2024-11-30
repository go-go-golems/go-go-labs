package parser

import "fmt"

// selectionElementImpl implements the SelectionElement interface
type selectionElementImpl struct {
	block  Block
	form   Form
	status SelectionStatus
}

func newSelectionElement(block Block, form Form) (SelectionElement, error) {
	if block.BlockType() != BlockTypeSelectionElement {
		return nil, fmt.Errorf("block type must be SELECTION_ELEMENT, got %s", block.BlockType())
	}

	// Get selection status from block
	status := getSelectionStatus(block)

	return &selectionElementImpl{
		block:  block,
		form:   form,
		status: status,
	}, nil
}

// IsSelected returns whether the element is selected
func (s *selectionElementImpl) IsSelected() bool {
	return s.status == SelectionStatusSelected
}

// SelectionStatus returns the selection status
func (s *selectionElementImpl) SelectionStatus() SelectionStatus {
	return s.status
}

// Confidence returns the confidence score
func (s *selectionElementImpl) Confidence() float64 {
	return s.block.Confidence()
}

// Block returns the underlying block
func (s *selectionElementImpl) Block() Block {
	return s.block
}

// Form returns the parent form
func (s *selectionElementImpl) Form() Form {
	return s.form
}

// BoundingBox returns the element's bounding box
func (s *selectionElementImpl) BoundingBox() BoundingBox {
	return s.block.BoundingBox()
}

// Polygon returns the element's polygon points
func (s *selectionElementImpl) Polygon() []Point {
	return s.block.Polygon()
}

// EntityTypes returns the entity types of the element
func (s *selectionElementImpl) EntityTypes() []EntityType {
	return s.block.EntityTypes()
}
