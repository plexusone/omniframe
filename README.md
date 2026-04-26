# omniframe

A DataFrame library for Go with multi-format support.

## Overview

omniframe provides a pandas/Polars-inspired DataFrame implementation for Go, with a focus on:

- **Multi-format export**: XLSX, CSV, JSON, Parquet, Arrow (planned)
- **Columnar storage**: Efficient memory layout for analytics
- **Type-safe columns**: Strong typing with automatic inference
- **Expression system**: Polars-inspired operations (planned)

## Installation

```bash
go get github.com/plexusone/omniframe
```

## Quick Start

### Single DataFrame to XLSX

```go
package main

import "github.com/plexusone/omniframe"

func main() {
    df, _ := omniframe.FromColumns("Users", map[string]any{
        "name":   []string{"Alice", "Bob", "Carol"},
        "age":    []int64{30, 25, 35},
        "salary": []float64{75000.50, 65000.25, 85000.75},
    })

    // Set Excel formatting
    df.SetFormat("salary", omniframe.FormatCurrency)
    df.SetColumnWidth("name", 20)

    df.WriteXLSX("users.xlsx")
}
```

### Multi-Sheet Workbook

```go
fs := omniframe.NewFrameSet("Q1 Report")
fs.AddFrame(users)
fs.AddFrame(orders)
fs.AddFrame(summary)

fs.WriteXLSX("report.xlsx")
```

## Column Types

- `TypeString` - String values
- `TypeInt64` - Integer values
- `TypeFloat64` - Floating-point values
- `TypeBool` - Boolean values
- `TypeTime` - time.Time values
- `TypeAny` - Mixed types

## Excel Formats

- `FormatGeneral` - Default
- `FormatText` - Text (@)
- `FormatNumber` - #,##0
- `FormatNumber2` - #,##0.00
- `FormatPercent` - 0%
- `FormatPercent2` - 0.00%
- `FormatCurrency` - $#,##0.00
- `FormatDate` - yyyy-mm-dd
- `FormatDateTime` - yyyy-mm-dd hh:mm:ss
- `FormatTime` - hh:mm:ss

## Phase 1 Status

Current implementation (Phase 1) focuses on XLSX export. See [PRD.md](PRD.md) for the full roadmap including:

- Phase 2: Core DataFrame Operations
- Phase 3: Expression System
- Phase 4: Lazy Execution
- Phase 5: Multi-Format Support
- Phase 6: Query Optimizer

## License

MIT
