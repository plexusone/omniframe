package omniframe

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFrameWriteXLSX(t *testing.T) {
	// Create a simple frame
	df, err := FromColumns("Users", map[string]any{
		"name":   []string{"Alice", "Bob", "Carol"},
		"age":    []int64{30, 25, 35},
		"salary": []float64{75000.50, 65000.25, 85000.75},
		"active": []bool{true, true, false},
	})
	if err != nil {
		t.Fatalf("Failed to create frame: %v", err)
	}

	// Set formats
	_ = df.SetFormat("salary", FormatCurrency)
	_ = df.SetColumnWidth("name", 20)

	// Write to temp file
	tmpFile := filepath.Join(os.TempDir(), "omniframe_test_single.xlsx")
	if err := df.WriteXLSX(tmpFile); err != nil {
		t.Fatalf("Failed to write XLSX: %v", err)
	}

	// Verify file was created
	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Output file is empty")
	}

	t.Logf("Created single-sheet XLSX: %s (%d bytes)", tmpFile, info.Size())

	// Cleanup
	_ = os.Remove(tmpFile)
}

func TestFrameSetWriteXLSX(t *testing.T) {
	// Create multiple frames
	users, _ := FromColumns("Users", map[string]any{
		"id":    []int64{1, 2, 3},
		"name":  []string{"Alice", "Bob", "Carol"},
		"email": []string{"alice@example.com", "bob@example.com", "carol@example.com"},
	})

	orders, _ := FromColumns("Orders", map[string]any{
		"order_id": []int64{101, 102, 103, 104},
		"user_id":  []int64{1, 1, 2, 3},
		"amount":   []float64{99.99, 149.50, 75.00, 200.00},
		"order_date": []time.Time{
			time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 2, 10, 0, 0, 0, 0, time.UTC),
		},
	})
	_ = orders.SetFormat("amount", FormatCurrency)
	_ = orders.SetFormat("order_date", FormatDate)

	summary, _ := FromRows("Summary", []string{"Metric", "Value"}, [][]any{
		{"Total Users", 3},
		{"Total Orders", 4},
		{"Total Revenue", 524.49},
	})

	// Create frame set
	fs := NewFrameSet("Q1 Report")
	if err := fs.AddFrame(users); err != nil {
		t.Fatalf("Failed to add users frame: %v", err)
	}
	if err := fs.AddFrame(orders); err != nil {
		t.Fatalf("Failed to add orders frame: %v", err)
	}
	if err := fs.AddFrame(summary); err != nil {
		t.Fatalf("Failed to add summary frame: %v", err)
	}

	// Verify frame set
	if fs.Len() != 3 {
		t.Errorf("Expected 3 frames, got %d", fs.Len())
	}

	// Write to temp file
	tmpFile := filepath.Join(os.TempDir(), "omniframe_test_multi.xlsx")
	if err := fs.WriteXLSX(tmpFile); err != nil {
		t.Fatalf("Failed to write XLSX: %v", err)
	}

	// Verify file was created
	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Output file is empty")
	}

	t.Logf("Created multi-sheet XLSX: %s (%d bytes)", tmpFile, info.Size())
	t.Logf("Sheets: %v", fs.Names())

	// Cleanup
	_ = os.Remove(tmpFile)
}

func TestColumnTypes(t *testing.T) {
	col, err := NewColumn("test", []int64{1, 2, 3})
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}

	if col.Type != TypeInt64 {
		t.Errorf("Expected TypeInt64, got %v", col.Type)
	}

	if col.Len() != 3 {
		t.Errorf("Expected length 3, got %d", col.Len())
	}

	// Test float column
	colFloat, _ := NewColumn("prices", []float64{1.5, 2.5, 3.5})
	if colFloat.Type != TypeFloat64 {
		t.Errorf("Expected TypeFloat64, got %v", colFloat.Type)
	}

	// Test string column
	colStr, _ := NewColumn("names", []string{"a", "b", "c"})
	if colStr.Type != TypeString {
		t.Errorf("Expected TypeString, got %v", colStr.Type)
	}

	// Test bool column
	colBool, _ := NewColumn("flags", []bool{true, false, true})
	if colBool.Type != TypeBool {
		t.Errorf("Expected TypeBool, got %v", colBool.Type)
	}
}

func TestFrameOperations(t *testing.T) {
	df, _ := FromColumns("test", map[string]any{
		"a": []int64{1, 2, 3, 4, 5},
		"b": []string{"one", "two", "three", "four", "five"},
	})

	// Test Len
	if df.Len() != 5 {
		t.Errorf("Expected 5 rows, got %d", df.Len())
	}

	// Test Width
	if df.Width() != 2 {
		t.Errorf("Expected 2 columns, got %d", df.Width())
	}

	// Test Head
	head := df.Head(3)
	if head.Len() != 3 {
		t.Errorf("Head: expected 3 rows, got %d", head.Len())
	}

	// Test Tail
	tail := df.Tail(2)
	if tail.Len() != 2 {
		t.Errorf("Tail: expected 2 rows, got %d", tail.Len())
	}

	// Test Row
	row := df.Row(2)
	if row["a"] != int64(3) {
		t.Errorf("Expected row[a]=3, got %v", row["a"])
	}
	if row["b"] != "three" {
		t.Errorf("Expected row[b]='three', got %v", row["b"])
	}

	// Test Select
	selected, err := df.Select("b")
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}
	if selected.Width() != 1 {
		t.Errorf("Select: expected 1 column, got %d", selected.Width())
	}

	// Test Clone
	clone := df.Clone()
	if clone.Len() != df.Len() || clone.Width() != df.Width() {
		t.Error("Clone doesn't match original dimensions")
	}
}

func TestColLetter(t *testing.T) {
	tests := []struct {
		col      int
		expected string
	}{
		{0, "A"},
		{1, "B"},
		{25, "Z"},
		{26, "AA"},
		{27, "AB"},
		{51, "AZ"},
		{52, "BA"},
		{701, "ZZ"},
		{702, "AAA"},
	}

	for _, tt := range tests {
		result := colLetter(tt.col)
		if result != tt.expected {
			t.Errorf("colLetter(%d) = %q, expected %q", tt.col, result, tt.expected)
		}
	}
}
