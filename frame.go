package omniframe

import (
	"errors"
	"fmt"
)

var (
	ErrColumnNotFound = errors.New("column not found")
	ErrColumnExists   = errors.New("column already exists")
	ErrLengthMismatch = errors.New("column length mismatch")
	ErrEmptyFrame     = errors.New("frame is empty")
	ErrInvalidSlice   = errors.New("values must be a slice")
)

// Frame represents a tabular data structure with named columns.
type Frame struct {
	name    string
	columns []string           // Column names in order
	data    map[string]*Column // Column data by name
}

// NewFrame creates a new empty frame with the given name.
func NewFrame(name string) *Frame {
	return &Frame{
		name:    name,
		columns: []string{},
		data:    make(map[string]*Column),
	}
}

// FromColumns creates a frame from a map of column name to values.
// All columns must have the same length.
func FromColumns(name string, data map[string]any) (*Frame, error) {
	f := NewFrame(name)

	var expectedLen int
	first := true

	for colName, values := range data {
		col, err := NewColumn(colName, values)
		if err != nil {
			return nil, fmt.Errorf("column %q: %w", colName, err)
		}

		if first {
			expectedLen = col.Len()
			first = false
		} else if col.Len() != expectedLen {
			return nil, fmt.Errorf("column %q: %w: expected %d, got %d",
				colName, ErrLengthMismatch, expectedLen, col.Len())
		}

		f.columns = append(f.columns, colName)
		f.data[colName] = col
	}

	return f, nil
}

// FromRows creates a frame from column names and row data.
func FromRows(name string, columns []string, rows [][]any) (*Frame, error) {
	f := NewFrame(name)
	f.columns = columns

	// Initialize columns
	for _, colName := range columns {
		f.data[colName] = &Column{
			Name:   colName,
			Type:   TypeUnknown,
			values: make([]any, 0, len(rows)),
		}
	}

	// Add rows
	for rowIdx, row := range rows {
		if len(row) != len(columns) {
			return nil, fmt.Errorf("row %d: %w: expected %d columns, got %d",
				rowIdx, ErrLengthMismatch, len(columns), len(row))
		}

		for colIdx, val := range row {
			col := f.data[columns[colIdx]]
			col.values = append(col.values, val)

			// Infer type from first non-nil value
			if col.Type == TypeUnknown && val != nil {
				col.Type = inferType(val)
			}
		}
	}

	return f, nil
}

// Name returns the frame name.
func (f *Frame) Name() string {
	return f.name
}

// SetName sets the frame name.
func (f *Frame) SetName(name string) {
	f.name = name
}

// Columns returns the column names in order.
func (f *Frame) Columns() []string {
	return f.columns
}

// Len returns the number of rows.
func (f *Frame) Len() int {
	if len(f.columns) == 0 {
		return 0
	}
	return f.data[f.columns[0]].Len()
}

// Width returns the number of columns.
func (f *Frame) Width() int {
	return len(f.columns)
}

// Col returns the column with the given name.
func (f *Frame) Col(name string) *Column {
	return f.data[name]
}

// HasColumn returns true if the frame has a column with the given name.
func (f *Frame) HasColumn(name string) bool {
	_, ok := f.data[name]
	return ok
}

// AddColumn adds a new column to the frame.
func (f *Frame) AddColumn(name string, values any) error {
	if f.HasColumn(name) {
		return fmt.Errorf("%w: %s", ErrColumnExists, name)
	}

	col, err := NewColumn(name, values)
	if err != nil {
		return err
	}

	// Check length matches existing columns
	if len(f.columns) > 0 && col.Len() != f.Len() {
		return fmt.Errorf("%w: expected %d rows, got %d", ErrLengthMismatch, f.Len(), col.Len())
	}

	f.columns = append(f.columns, name)
	f.data[name] = col
	return nil
}

// SetFormat sets the Excel format for a column.
func (f *Frame) SetFormat(colName string, format Format) error {
	col, ok := f.data[colName]
	if !ok {
		return fmt.Errorf("%w: %s", ErrColumnNotFound, colName)
	}
	col.Format = format
	return nil
}

// SetColumnWidth sets the width for a column in Excel units.
func (f *Frame) SetColumnWidth(colName string, width float64) error {
	col, ok := f.data[colName]
	if !ok {
		return fmt.Errorf("%w: %s", ErrColumnNotFound, colName)
	}
	col.Width = width
	return nil
}

// Row returns the values at the given row index as a map.
func (f *Frame) Row(idx int) map[string]any {
	if idx < 0 || idx >= f.Len() {
		return nil
	}

	row := make(map[string]any, len(f.columns))
	for _, colName := range f.columns {
		row[colName] = f.data[colName].At(idx)
	}
	return row
}

// RowSlice returns the values at the given row index as a slice.
func (f *Frame) RowSlice(idx int) []any {
	if idx < 0 || idx >= f.Len() {
		return nil
	}

	row := make([]any, len(f.columns))
	for i, colName := range f.columns {
		row[i] = f.data[colName].At(idx)
	}
	return row
}

// Head returns a new frame with the first n rows.
func (f *Frame) Head(n int) *Frame {
	if n <= 0 {
		return NewFrame(f.name)
	}
	if n > f.Len() {
		n = f.Len()
	}

	result := NewFrame(f.name)
	result.columns = make([]string, len(f.columns))
	copy(result.columns, f.columns)

	for _, colName := range f.columns {
		srcCol := f.data[colName]
		newCol := &Column{
			Name:   srcCol.Name,
			Type:   srcCol.Type,
			Format: srcCol.Format,
			Width:  srcCol.Width,
			values: make([]any, n),
		}
		copy(newCol.values, srcCol.values[:n])
		result.data[colName] = newCol
	}

	return result
}

// Tail returns a new frame with the last n rows.
func (f *Frame) Tail(n int) *Frame {
	if n <= 0 {
		return NewFrame(f.name)
	}
	if n > f.Len() {
		n = f.Len()
	}

	start := f.Len() - n

	result := NewFrame(f.name)
	result.columns = make([]string, len(f.columns))
	copy(result.columns, f.columns)

	for _, colName := range f.columns {
		srcCol := f.data[colName]
		newCol := &Column{
			Name:   srcCol.Name,
			Type:   srcCol.Type,
			Format: srcCol.Format,
			Width:  srcCol.Width,
			values: make([]any, n),
		}
		copy(newCol.values, srcCol.values[start:])
		result.data[colName] = newCol
	}

	return result
}

// Select returns a new frame with only the specified columns.
func (f *Frame) Select(columns ...string) (*Frame, error) {
	result := NewFrame(f.name)

	for _, colName := range columns {
		srcCol, ok := f.data[colName]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrColumnNotFound, colName)
		}

		newCol := &Column{
			Name:   srcCol.Name,
			Type:   srcCol.Type,
			Format: srcCol.Format,
			Width:  srcCol.Width,
			values: make([]any, len(srcCol.values)),
		}
		copy(newCol.values, srcCol.values)

		result.columns = append(result.columns, colName)
		result.data[colName] = newCol
	}

	return result, nil
}

// Clone creates a deep copy of the frame.
func (f *Frame) Clone() *Frame {
	result := NewFrame(f.name)
	result.columns = make([]string, len(f.columns))
	copy(result.columns, f.columns)

	for _, colName := range f.columns {
		srcCol := f.data[colName]
		newCol := &Column{
			Name:   srcCol.Name,
			Type:   srcCol.Type,
			Format: srcCol.Format,
			Width:  srcCol.Width,
			values: make([]any, len(srcCol.values)),
		}
		copy(newCol.values, srcCol.values)
		result.data[colName] = newCol
	}

	return result
}
