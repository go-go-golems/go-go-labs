package parser

import "fmt"

// tableRowImpl implements the TableRow interface
type tableRowImpl struct {
	table    Table
	rowIndex int
	cells    []Cell
}

func newTableRow(table Table, rowIndex int) (TableRow, error) {
	if rowIndex < 0 || rowIndex >= table.RowCount() {
		return nil, fmt.Errorf("invalid row index: %d", rowIndex)
	}

	row := &tableRowImpl{
		table:    table,
		rowIndex: rowIndex,
	}

	// Get cells for this row
	row.cells = table.Cells()[rowIndex]

	return row, nil
}

// Cells returns the cells in this row
func (r *tableRowImpl) Cells() []Cell {
	return r.cells
}

// RowIndex returns the row's index
func (r *tableRowImpl) RowIndex() int {
	return r.rowIndex
}

// Table returns the parent table
func (r *tableRowImpl) Table() Table {
	return r.table
}
