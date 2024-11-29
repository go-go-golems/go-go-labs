package parser

import (
	"fmt"
)

// TableRow represents a row in a table
type TableRow interface {
	Cells() []Cell
	RowIndex() int
	Table() Table
}

// Cell represents a table cell
type Cell interface {
	// Content
	Text() string
	Confidence() float64

	// Position
	RowIndex() int
	ColumnIndex() int
	RowSpan() int
	ColumnSpan() int

	// Navigation
	Table() Table

	// Geometry
	BoundingBox() BoundingBox
	Polygon() []Point
}

// MergedCell represents a cell that spans multiple rows or columns
type MergedCell interface {
	Cell
	MergedRowSpan() int
	MergedColumnSpan() int
	ContainedCells() []Cell
}

// tableImpl implements the Table interface
type tableImpl struct {
	block       Block
	page        Page
	rows        []TableRow
	cells       [][]Cell
	mergedCells []MergedCell
	rowCount    int
	columnCount int
}

func newTable(block Block, page Page) (Table, error) {
	if block.BlockType() != BlockTypeTable {
		return nil, fmt.Errorf("block type must be TABLE, got %s", block.BlockType())
	}

	table := &tableImpl{
		block: block,
		page:  page,
	}

	if err := table.processTableStructure(); err != nil {
		return nil, fmt.Errorf("processing table structure: %w", err)
	}

	return table, nil
}

// Rows returns all rows in the table
func (t *tableImpl) Rows() []TableRow {
	return t.rows
}

// Cells returns the 2D array of cells
func (t *tableImpl) Cells() [][]Cell {
	return t.cells
}

// MergedCells returns all merged cells
func (t *tableImpl) MergedCells() []MergedCell {
	return t.mergedCells
}

// RowCount returns the number of rows
func (t *tableImpl) RowCount() int {
	return t.rowCount
}

// ColumnCount returns the number of columns
func (t *tableImpl) ColumnCount() int {
	return t.columnCount
}

// Page returns the parent page
func (t *tableImpl) Page() Page {
	return t.page
}

// GetCellByPosition returns a cell at the specified position
func (t *tableImpl) GetCellByPosition(row, col int) (Cell, error) {
	if row < 0 || row >= t.rowCount || col < 0 || col >= t.columnCount {
		return nil, fmt.Errorf("invalid cell position: row=%d, col=%d", row, col)
	}
	return t.cells[row][col], nil
}

// Internal methods

func (t *tableImpl) processTableStructure() error {
	// Find dimensions
	for _, child := range t.block.Children() {
		if child.BlockType() == BlockTypeCell {
			rowIdx := getRowIndex(child)
			colIdx := getColumnIndex(child)
			t.rowCount = max(t.rowCount, rowIdx+1)
			t.columnCount = max(t.columnCount, colIdx+1)
		}
	}

	// Initialize cells array
	t.cells = make([][]Cell, t.rowCount)
	for i := range t.cells {
		t.cells[i] = make([]Cell, t.columnCount)
	}

	// Create cells
	for _, child := range t.block.Children() {
		if child.BlockType() == BlockTypeCell {
			cell, err := newCell(child, t)
			if err != nil {
				return fmt.Errorf("creating cell: %w", err)
			}

			rowIdx := getRowIndex(child)
			colIdx := getColumnIndex(child)
			t.cells[rowIdx][colIdx] = cell
		}
	}

	// Process merged cells
	if err := t.processMergedCells(); err != nil {
		return fmt.Errorf("processing merged cells: %w", err)
	}

	// Create rows
	if err := t.createRows(); err != nil {
		return fmt.Errorf("creating rows: %w", err)
	}

	return nil
}

func (t *tableImpl) processMergedCells() error {
	for _, child := range t.block.Children() {
		if child.BlockType() == BlockTypeCell {
			rowSpan := getRowSpan(child)
			colSpan := getColumnSpan(child)

			if rowSpan > 1 || colSpan > 1 {
				merged, err := newMergedCell(child, t)
				if err != nil {
					return fmt.Errorf("creating merged cell: %w", err)
				}
				t.mergedCells = append(t.mergedCells, merged)
			}
		}
	}
	return nil
}

func (t *tableImpl) createRows() error {
	t.rows = make([]TableRow, t.rowCount)
	for i := 0; i < t.rowCount; i++ {
		row, err := newTableRow(t, i)
		if err != nil {
			return fmt.Errorf("creating row %d: %w", i, err)
		}
		t.rows[i] = row
	}
	return nil
}

// Helper functions

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
