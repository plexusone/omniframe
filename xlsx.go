package omniframe

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// WriteXLSX writes the frame to an Excel file.
func (f *Frame) WriteXLSX(filename string) error {
	fs := NewFrameSet("")
	if err := fs.AddFrame(f); err != nil {
		return err
	}
	return fs.WriteXLSX(filename)
}

// WriteXLSX writes all frames to an Excel file, each as a separate sheet.
func (fs *FrameSet) WriteXLSX(filename string) error {
	if fs.Len() == 0 {
		return ErrEmptyFrameSet
	}

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()

	// Track if we need to delete the default sheet
	deleteDefault := true

	for i, frameName := range fs.order {
		frame := fs.frames[frameName]
		if frame == nil {
			continue
		}

		sheetName := frameName
		if sheetName == "" {
			sheetName = fmt.Sprintf("Sheet%d", i+1)
		}

		// Create sheet (first sheet reuses "Sheet1" or creates new)
		if i == 0 {
			// Rename default sheet
			if err := file.SetSheetName("Sheet1", sheetName); err != nil {
				return fmt.Errorf("failed to rename sheet: %w", err)
			}
			deleteDefault = false
		} else {
			if _, err := file.NewSheet(sheetName); err != nil {
				return fmt.Errorf("failed to create sheet %q: %w", sheetName, err)
			}
		}

		if err := writeFrameToSheet(file, sheetName, frame); err != nil {
			return fmt.Errorf("failed to write sheet %q: %w", sheetName, err)
		}
	}

	// Delete default sheet if we didn't use it
	if deleteDefault {
		_ = file.DeleteSheet("Sheet1") // Ignore error, sheet may not exist
	}

	// Set first sheet as active
	if len(fs.order) > 0 {
		idx, err := file.GetSheetIndex(fs.order[0])
		if err == nil {
			file.SetActiveSheet(idx)
		}
	}

	return file.SaveAs(filename)
}

// writeFrameToSheet writes a frame's data to an Excel sheet.
func writeFrameToSheet(file *excelize.File, sheetName string, frame *Frame) error {
	if frame.Width() == 0 {
		return nil
	}

	// Create styles for formatted columns
	styleCache := make(map[Format]int)

	// Write header row
	for colIdx, colName := range frame.columns {
		cell := cellRef(colIdx, 0)
		if err := file.SetCellValue(sheetName, cell, colName); err != nil {
			return err
		}

		// Set column width if specified
		col := frame.data[colName]
		if col.Width > 0 {
			colLetter := colLetter(colIdx)
			if err := file.SetColWidth(sheetName, colLetter, colLetter, col.Width); err != nil {
				return err
			}
		}
	}

	// Style header row (bold)
	headerStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	if err == nil {
		startCell := cellRef(0, 0)
		endCell := cellRef(frame.Width()-1, 0)
		_ = file.SetCellStyle(sheetName, startCell, endCell, headerStyle)
	}

	// Write data rows
	for rowIdx := 0; rowIdx < frame.Len(); rowIdx++ {
		excelRow := rowIdx + 1 // +1 for header row

		for colIdx, colName := range frame.columns {
			col := frame.data[colName]
			val := col.At(rowIdx)
			cell := cellRef(colIdx, excelRow)

			if err := file.SetCellValue(sheetName, cell, val); err != nil {
				return err
			}

			// Apply format if specified
			if col.Format != FormatGeneral {
				styleID, ok := styleCache[col.Format]
				if !ok {
					styleID, err = file.NewStyle(&excelize.Style{
						NumFmt: excelNumFmt(col.Format),
					})
					if err != nil {
						return err
					}
					styleCache[col.Format] = styleID
				}
				if err := file.SetCellStyle(sheetName, cell, cell, styleID); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// cellRef returns an Excel cell reference like "A1", "B2", etc.
func cellRef(col, row int) string {
	return fmt.Sprintf("%s%d", colLetter(col), row+1)
}

// colLetter returns the Excel column letter(s) for a 0-based index.
func colLetter(col int) string {
	result := ""
	col++ // Convert to 1-based

	for col > 0 {
		col--
		result = string(rune('A'+col%26)) + result
		col /= 26
	}

	return result
}

// excelNumFmt returns the Excel number format ID for a Format.
func excelNumFmt(f Format) int {
	switch f {
	case FormatText:
		return 49 // @
	case FormatNumber:
		return 3 // #,##0
	case FormatNumber2:
		return 4 // #,##0.00
	case FormatPercent:
		return 9 // 0%
	case FormatPercent2:
		return 10 // 0.00%
	case FormatCurrency:
		return 44 // Currency with 2 decimals
	case FormatDate:
		return 14 // m/d/yyyy (will use custom)
	case FormatDateTime:
		return 22 // m/d/yyyy h:mm
	case FormatTime:
		return 21 // h:mm:ss
	default:
		return 0 // General
	}
}

// ToExcelizeFile returns an excelize.File for advanced customization.
func (fs *FrameSet) ToExcelizeFile() (*excelize.File, error) {
	if fs.Len() == 0 {
		return nil, ErrEmptyFrameSet
	}

	file := excelize.NewFile()

	for i, frameName := range fs.order {
		frame := fs.frames[frameName]
		if frame == nil {
			continue
		}

		sheetName := frameName
		if sheetName == "" {
			sheetName = fmt.Sprintf("Sheet%d", i+1)
		}

		if i == 0 {
			if err := file.SetSheetName("Sheet1", sheetName); err != nil {
				return nil, err
			}
		} else {
			if _, err := file.NewSheet(sheetName); err != nil {
				return nil, err
			}
		}

		if err := writeFrameToSheet(file, sheetName, frame); err != nil {
			return nil, err
		}
	}

	return file, nil
}
