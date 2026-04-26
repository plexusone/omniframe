# Release Notes - v0.1.0

**Release Date:** 2026-04-26

## Overview

Initial release of omniframe, a DataFrame library for Go with XLSX export support.

omniframe provides pandas/Polars-inspired data structures for Go, focusing on tabular data manipulation and multi-format export capabilities.

## Features

### Core Data Structures

- **Frame** - Tabular data with named columns, supporting operations like `Head()`, `Tail()`, `Select()`, and `Clone()`
- **Column** - Typed column with automatic type inference for string, int64, float64, bool, and time.Time
- **FrameSet** - Collection of frames for multi-sheet workbook support

### Frame Construction

- `FromColumns()` - Create frame from map of column name to values
- `FromRows()` - Create frame from column names and row data

### XLSX Export

- `WriteXLSX()` for single-frame export
- `FrameSet.WriteXLSX()` for multi-sheet workbooks
- Excel number formats: currency, percent, date, datetime, and more
- Configurable column widths

## Installation

```bash
go get github.com/plexusone/omniframe
```

## Quick Start

```go
package main

import "github.com/plexusone/omniframe"

func main() {
    df, _ := omniframe.FromColumns("Users", map[string]any{
        "name":   []string{"Alice", "Bob", "Carol"},
        "age":    []int64{30, 25, 35},
        "salary": []float64{75000.50, 65000.25, 85000.75},
    })

    df.SetFormat("salary", omniframe.FormatCurrency)
    df.WriteXLSX("users.xlsx")
}
```

## Roadmap

This release completes Phase 1 (XLSX Export). Future phases include:

- Phase 2: Core DataFrame Operations
- Phase 3: Expression System
- Phase 4: Lazy Execution
- Phase 5: Multi-Format Support (CSV, JSON, Parquet, Arrow)
- Phase 6: Query Optimizer

See [PRD.md](PRD.md) for the complete roadmap.

## License

MIT
