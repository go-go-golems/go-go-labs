package parser

import "fmt"

// cellImpl implements the Cell interface
type cellImpl struct {
	block       Block
	table       Table
	rowIndex    int
	columnIndex int
	rowSpan     int
	columnSpan  int
}

func newCell(block Block, table Table) (Cell, error) {
	if block.BlockType() != BlockTypeCell {
		return nil, fmt.Errorf("block type must be CELL, got %s", block.BlockType())
	}

	return &cellImpl{
		block:       block,
		table:       table,
		rowIndex:    getRowIndex(block),
		columnIndex: getColumnIndex(block),
		rowSpan:     getRowSpan(block),
		columnSpan:  getColumnSpan(block),
	}, nil
}

// Text returns the cell's text content
func (c *cellImpl) Text() string {
	return c.block.Text()
}

// Confidence returns the confidence score
func (c *cellImpl) Confidence() float64 {
	return c.block.Confidence()
}

// RowIndex returns the cell's row index
func (c *cellImpl) RowIndex() int {
	return c.rowIndex
}

// ColumnIndex returns the cell's column index
func (c *cellImpl) ColumnIndex() int {
	return c.columnIndex
}

// RowSpan returns the number of rows the cell spans
func (c *cellImpl) RowSpan() int {
	return c.rowSpan
}

// ColumnSpan returns the number of columns the cell spans
func (c *cellImpl) ColumnSpan() int {
	return c.columnSpan
}

// Table returns the parent table
func (c *cellImpl) Table() Table {
	return c.table
}

// BoundingBox returns the cell's bounding box
func (c *cellImpl) BoundingBox() BoundingBox {
	return c.block.BoundingBox()
}

// Polygon returns the cell's polygon points
func (c *cellImpl) Polygon() []Point {
	return c.block.Polygon()
}

// EntityTypes returns the cell's entity types
func (c *cellImpl) EntityTypes() []EntityType {
	return c.block.EntityTypes()
}

// Add IsColumnHeader method to cellImpl
func (c *cellImpl) IsColumnHeader() bool {
	for _, et := range c.EntityTypes() {
		if et == EntityTypeColumnHeader {
			return true
		}
	}
	return false
}

// mergedCellImpl implements the MergedCell interface
type mergedCellImpl struct {
	cellImpl
	containedCells []Cell
}

func newMergedCell(block Block, table Table) (MergedCell, error) {
	base, err := newCell(block, table)
	if err != nil {
		return nil, err
	}

	baseCell := base.(*cellImpl)
	merged := &mergedCellImpl{
		cellImpl: *baseCell,
	}

	// Find contained cells
	if err := merged.findContainedCells(); err != nil {
		return nil, fmt.Errorf("finding contained cells: %w", err)
	}

	return merged, nil
}

func (m *mergedCellImpl) MergedRowSpan() int {
	return m.rowSpan
}

func (m *mergedCellImpl) MergedColumnSpan() int {
	return m.columnSpan
}

func (m *mergedCellImpl) ContainedCells() []Cell {
	return m.containedCells
}

func (m *mergedCellImpl) findContainedCells() error {
	// Implementation will find all cells contained within this merged cell
	return nil
}

// Add EntityTypes method to mergedCellImpl
func (m *mergedCellImpl) EntityTypes() []EntityType {
	return m.block.EntityTypes()
}

// Add IsColumnHeader method to mergedCellImpl (though it inherits from cellImpl, being explicit)
func (m *mergedCellImpl) IsColumnHeader() bool {
	return m.cellImpl.IsColumnHeader()
}
