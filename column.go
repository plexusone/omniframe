// Package omniframe provides a DataFrame library for Go with support for
// multiple tabular data formats.
package omniframe

import (
	"fmt"
	"reflect"
	"time"
)

// ColumnType represents the data type of a column.
type ColumnType uint8

const (
	TypeUnknown ColumnType = iota
	TypeString
	TypeInt64
	TypeFloat64
	TypeBool
	TypeTime
	TypeAny // Mixed types
)

// String returns the string representation of the column type.
func (ct ColumnType) String() string {
	switch ct {
	case TypeString:
		return "string"
	case TypeInt64:
		return "int64"
	case TypeFloat64:
		return "float64"
	case TypeBool:
		return "bool"
	case TypeTime:
		return "time"
	case TypeAny:
		return "any"
	default:
		return "unknown"
	}
}

// Format represents Excel number format for a column.
type Format string

const (
	FormatGeneral  Format = ""
	FormatText     Format = "@"
	FormatNumber   Format = "#,##0"
	FormatNumber2  Format = "#,##0.00"
	FormatPercent  Format = "0%"
	FormatPercent2 Format = "0.00%"
	FormatCurrency Format = "$#,##0.00"
	FormatDate     Format = "yyyy-mm-dd"
	FormatDateTime Format = "yyyy-mm-dd hh:mm:ss"
	FormatTime     Format = "hh:mm:ss"
)

// Column represents a single column of data.
type Column struct {
	Name   string
	Type   ColumnType
	Format Format
	Width  float64 // Column width in Excel units (0 = auto)
	values []any
}

// NewColumn creates a new column with the given name and values.
func NewColumn(name string, values any) (*Column, error) {
	col := &Column{
		Name: name,
	}

	if err := col.setValues(values); err != nil {
		return nil, err
	}

	return col, nil
}

// setValues sets the column values and infers the type.
func (c *Column) setValues(values any) error {
	rv := reflect.ValueOf(values)
	if rv.Kind() != reflect.Slice {
		return fmt.Errorf("values must be a slice, got %T", values)
	}

	c.values = make([]any, rv.Len())

	if rv.Len() == 0 {
		c.Type = TypeAny
		return nil
	}

	// Infer type from first non-nil element
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i)
		c.values[i] = elem.Interface()

		if c.Type == TypeUnknown && !elem.IsZero() {
			c.Type = inferType(elem.Interface())
		}
	}

	if c.Type == TypeUnknown {
		c.Type = TypeAny
	}

	return nil
}

// inferType infers the ColumnType from a value.
func inferType(v any) ColumnType {
	switch v.(type) {
	case string:
		return TypeString
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return TypeInt64
	case float32, float64:
		return TypeFloat64
	case bool:
		return TypeBool
	case time.Time:
		return TypeTime
	default:
		return TypeAny
	}
}

// Len returns the number of values in the column.
func (c *Column) Len() int {
	return len(c.values)
}

// Values returns all values as []any.
func (c *Column) Values() []any {
	return c.values
}

// At returns the value at the given index.
func (c *Column) At(i int) any {
	if i < 0 || i >= len(c.values) {
		return nil
	}
	return c.values[i]
}

// Strings returns values as []string.
func (c *Column) Strings() []string {
	result := make([]string, len(c.values))
	for i, v := range c.values {
		if v == nil {
			result[i] = ""
			continue
		}
		switch val := v.(type) {
		case string:
			result[i] = val
		default:
			result[i] = fmt.Sprintf("%v", val)
		}
	}
	return result
}

// Int64s returns values as []int64.
func (c *Column) Int64s() []int64 {
	result := make([]int64, len(c.values))
	for i, v := range c.values {
		if v == nil {
			continue
		}
		switch val := v.(type) {
		case int64:
			result[i] = val
		case int:
			result[i] = int64(val)
		case int32:
			result[i] = int64(val)
		case float64:
			result[i] = int64(val)
		case float32:
			result[i] = int64(val)
		}
	}
	return result
}

// Float64s returns values as []float64.
func (c *Column) Float64s() []float64 {
	result := make([]float64, len(c.values))
	for i, v := range c.values {
		if v == nil {
			continue
		}
		switch val := v.(type) {
		case float64:
			result[i] = val
		case float32:
			result[i] = float64(val)
		case int64:
			result[i] = float64(val)
		case int:
			result[i] = float64(val)
		case int32:
			result[i] = float64(val)
		}
	}
	return result
}

// Bools returns values as []bool.
func (c *Column) Bools() []bool {
	result := make([]bool, len(c.values))
	for i, v := range c.values {
		if val, ok := v.(bool); ok {
			result[i] = val
		}
	}
	return result
}

// Times returns values as []time.Time.
func (c *Column) Times() []time.Time {
	result := make([]time.Time, len(c.values))
	for i, v := range c.values {
		if val, ok := v.(time.Time); ok {
			result[i] = val
		}
	}
	return result
}
