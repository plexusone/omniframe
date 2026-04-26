# OmniFrame

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Docs][docs-mkdoc-svg]][docs-mkdoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/omniframe/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/omniframe/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/omniframe/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/omniframe/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/omniframe/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/omniframe/actions/workflows/go-sast-codeql.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/plexusone/omniframe
 [goreport-url]: https://goreportcard.com/report/github.com/plexusone/omniframe
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/omniframe
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/omniframe
 [docs-mkdoc-svg]: https://img.shields.io/badge/Go-dev%20guide-blue.svg
 [docs-mkdoc-url]: https://plexusone.dev/omniframe
 [viz-svg]: https://img.shields.io/badge/Go-visualizaton-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fomniframe
 [loc-svg]: https://tokei.rs/b1/github/plexusone/omniframe
 [repo-url]: https://github.com/plexusone/omniframe
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/omniframe/blob/main/LICENSE

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
