# omniframe - Product Requirements Document

## Vision

**omniframe** is a high-performance, pandas/Polars-like DataFrame library for Go that provides a unified abstraction over multiple tabular data formats and backends.

## Problem Statement

Go developers working with tabular data face a fragmented ecosystem:

- **gota** (the most popular DataFrame library) is archived
- Existing alternatives have limited maintenance and feature gaps
- Teams fall back to Python/pandas/Polars for data manipulation, breaking "all-in-Go" workflows
- No unified abstraction for reading/writing multiple formats (XLSX, CSV, SQL, JSON, etc.)
- No Go library offers modern features like lazy evaluation or expression-based APIs

## Goals

1. **Unified Interface** - Single API for multiple tabular backends (following plexusone "omni" pattern)
2. **Modern API** - Expression system and lazy evaluation inspired by Polars
3. **High Performance** - Columnar storage, parallel execution, query optimization
4. **Go-Native** - Type-safe, idiomatic Go with generics where appropriate
5. **Incremental Adoption** - Start simple (eager API), opt-in to advanced features (lazy API)

## Target Users

- Go developers doing data manipulation/ETL
- Teams wanting to avoid Python context-switching
- Applications generating reports (XLSX, CSV)
- Services integrating with databases and data warehouses
- Data pipelines requiring high performance

## Inspiration: Pandas vs Polars

| Feature | Pandas | Polars | omniframe (Goal) |
|---------|--------|--------|------------------|
| Language | Python/C | Rust/Python | Go |
| Memory Model | Row-based (NumPy) | Columnar (Arrow) | Columnar (Arrow optional) |
| Evaluation | Eager only | Eager + Lazy | Eager + Lazy |
| Parallelism | Limited | Automatic | Automatic |
| Expression API | No | Yes | Yes |
| Query Optimizer | No | Yes | Yes (with lazy) |
| Type Safety | Runtime | Compile-time (Rust) | Compile-time (Go) |

## Core Concepts

### 1. Frame (Eager API)

Immediate execution, simple API for small-to-medium datasets.

```go
df := omniframe.NewFrame("users")
df.AddColumn("name", []string{"Alice", "Bob"})
df.AddColumn("age", []int64{30, 25})

// Operations execute immediately
filtered := df.Filter(func(row Row) bool {
    return row.Int64("age") >= 18
})
```

### 2. LazyFrame (Lazy API)

Deferred execution with query optimization for large datasets.

```go
lf := omniframe.Scan("users.csv")  // Returns LazyFrame

result := lf.
    Filter(col("age").Gte(18)).
    Select(col("name"), col("email")).
    GroupBy(col("department")).
    Agg(
        col("salary").Sum(),
        col("age").Mean(),
    ).
    Sort(col("salary").Desc()).
    Collect()  // Executes optimized query plan
```

### 3. Expression System

Composable, type-safe expressions for column operations.

```go
import . "github.com/plexusone/omniframe/expr"

// Column references
col("name")
col("age")

// Comparisons
col("age").Gte(18)
col("status").Eq("active")
col("name").Contains("John")

// Arithmetic
col("price").Mul(col("quantity"))
col("total").Div(100)

// Aggregations
col("salary").Sum()
col("age").Mean()
col("id").Count()
col("price").Min()
col("price").Max()

// Composition
col("revenue").Sub(col("cost")).Alias("profit")
```

### 4. Query Optimizer (Lazy Mode)

Automatic optimizations when using LazyFrame:

| Optimization | Description |
|--------------|-------------|
| **Predicate Pushdown** | Move filters closer to data source |
| **Projection Pushdown** | Load only required columns |
| **Common Subexpression Elimination** | Reuse repeated computations |
| **Join Reordering** | Optimize join order for smaller intermediates |
| **Parallel Execution** | Distribute operations across CPU cores |

```go
// User writes:
lf.Select(col("a"), col("b")).Filter(col("a").Gt(10))

// Optimizer rewrites to:
lf.Filter(col("a").Gt(10)).Select(col("a"), col("b"))
// Filter first = less data to select
```

## Supported Formats (Roadmap)

### Phase 1: MVP - XLSX Export

| Format | Read | Write | Notes |
|--------|------|-------|-------|
| XLSX | - | Yes | Multi-sheet workbooks |

### Phase 2: Core I/O

| Format | Read | Write | Notes |
|--------|------|-------|-------|
| CSV | Yes | Yes | Standard tabular format |
| XLSX | Yes | Yes | Full read support |
| JSON | Yes | Yes | Array of objects, nested |
| TOON | Yes | Yes | Token-optimized notation |

### Phase 3: Databases

| Format | Read | Write | Notes |
|--------|------|-------|-------|
| SQL | Yes | Yes | PostgreSQL, MySQL, SQLite |
| Parquet | Yes | Yes | Columnar analytics format |

### Phase 4: Big Data

| Format | Read | Write | Notes |
|--------|------|-------|-------|
| Apache Arrow | Yes | Yes | In-memory columnar, IPC |
| GoNum | Yes | Yes | Numerical computing |
| Cassandra | Yes | Yes | Wide-column store |
| BigQuery | Yes | Yes | Google data warehouse |
| DuckDB | Yes | Yes | Embedded analytics |

## Architecture

### Internal Storage Model

**Phase 1-3:** Simple columnar storage

```go
type Frame struct {
    name    string
    columns []string
    schema  Schema
    data    map[string]*Column
}

type Column struct {
    Name     string
    Type     ColumnType
    Nullable bool
    values   any  // []int64, []float64, []string, []bool, []time.Time
    nullMask []bool
}

type ColumnType uint8

const (
    TypeInt64 ColumnType = iota
    TypeFloat64
    TypeString
    TypeBool
    TypeTime
    TypeDuration
    TypeAny
)
```

**Phase 4+:** Optional Arrow backend

```go
type ArrowFrame struct {
    record arrow.Record
}

// Seamless conversion
func (f *Frame) ToArrow() *ArrowFrame
func FromArrow(record arrow.Record) *Frame
```

### Lazy Evaluation Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  User Code  │────▶│  DSL/Expr   │────▶│  Logical    │────▶│  Physical   │
│  (LazyFrame)│     │  Builder    │     │  Plan (IR)  │     │  Plan       │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                                               │                   │
                                               ▼                   ▼
                                        ┌─────────────┐     ┌─────────────┐
                                        │  Optimizer  │     │  Executor   │
                                        │  (Rewrites) │     │  (Parallel) │
                                        └─────────────┘     └─────────────┘
```

### Logical Plan Nodes

```go
type LogicalPlan interface {
    Schema() Schema
    Children() []LogicalPlan
}

type ScanNode struct {
    source     DataSource
    projection []string  // columns to read
    predicate  Expr      // pushed-down filter
}

type FilterNode struct {
    input     LogicalPlan
    predicate Expr
}

type ProjectNode struct {
    input   LogicalPlan
    columns []Expr
}

type AggregateNode struct {
    input    LogicalPlan
    groupBy  []Expr
    aggregates []AggExpr
}

type JoinNode struct {
    left     LogicalPlan
    right    LogicalPlan
    on       []Expr
    joinType JoinType
}

type SortNode struct {
    input   LogicalPlan
    orderBy []SortExpr
}
```

## API Design

### Frame Creation

```go
// Empty frame
df := omniframe.NewFrame("users")

// From columnar data
df := omniframe.FromColumns(map[string]any{
    "name": []string{"Alice", "Bob", "Carol"},
    "age":  []int64{30, 25, 35},
    "salary": []float64{75000, 65000, 85000},
})

// From rows
df := omniframe.FromRows(
    []string{"name", "age", "salary"},
    [][]any{
        {"Alice", 30, 75000.0},
        {"Bob", 25, 65000.0},
    },
)

// From structs (reflection)
type User struct {
    Name   string  `frame:"name"`
    Age    int64   `frame:"age"`
    Salary float64 `frame:"salary"`
}
df := omniframe.FromStructs(users)
```

### Reading Data

```go
// Eager (immediate load)
df, err := omniframe.ReadCSV("data.csv")
df, err := omniframe.ReadXLSX("data.xlsx")
df, err := omniframe.ReadXLSXSheet("data.xlsx", "Sheet1")
df, err := omniframe.ReadJSON("data.json")
df, err := omniframe.ReadParquet("data.parquet")
df, err := omniframe.ReadSQL(db, "SELECT * FROM users")

// Lazy (deferred, optimized)
lf := omniframe.ScanCSV("data.csv")
lf := omniframe.ScanParquet("data.parquet")
lf := omniframe.ScanSQL(db, "users")
```

### Writing Data

```go
err := df.WriteCSV("output.csv")
err := df.WriteXLSX("output.xlsx")
err := df.WriteJSON("output.json")
err := df.WriteParquet("output.parquet")
err := df.WriteSQL(db, "output_table")

// Multi-sheet XLSX
fs := omniframe.NewFrameSet("Report")
fs.AddFrame(salesDF)
fs.AddFrame(inventoryDF)
err := fs.WriteXLSX("report.xlsx")
```

### Column Operations

```go
// Access
names := df.Col("name").Strings()
ages := df.Col("age").Int64s()
prices := df.Col("price").Float64s()

// Add/modify
df = df.WithColumn("full_name",
    col("first").Concat(lit(" ")).Concat(col("last")))

df = df.WithColumn("tax",
    col("price").Mul(0.08))

// Drop
df = df.Drop("temp_column")

// Rename
df = df.Rename(map[string]string{
    "old_name": "new_name",
})
```

### Filtering

```go
// Function-based (eager)
adults := df.Filter(func(row Row) bool {
    return row.Int64("age") >= 18
})

// Expression-based
adults := df.Filter(col("age").Gte(18))

// Multiple conditions
active_adults := df.Filter(
    col("age").Gte(18).And(col("status").Eq("active")),
)

// Lazy with optimization
result := lf.
    Filter(col("age").Gte(18)).
    Filter(col("country").Eq("US")).  // Optimizer combines filters
    Collect()
```

### Selection & Slicing

```go
subset := df.Select("name", "email", "phone")
subset := df.SelectExpr(col("name"), col("salary").Mul(12).Alias("annual"))

first10 := df.Head(10)
last10 := df.Tail(10)
sample := df.Sample(100)
slice := df.Slice(10, 20)  // rows 10-19
```

### Sorting

```go
sorted := df.Sort("name")
sorted := df.SortDesc("salary")
sorted := df.SortBy(
    col("department").Asc(),
    col("salary").Desc(),
)
```

### Aggregation

```go
// Simple
total := df.Col("salary").Sum()
average := df.Col("age").Mean()
count := df.Len()

// GroupBy
summary := df.GroupBy("department").Agg(
    col("salary").Sum().Alias("total_salary"),
    col("salary").Mean().Alias("avg_salary"),
    col("id").Count().Alias("headcount"),
    col("age").Min().Alias("youngest"),
    col("age").Max().Alias("oldest"),
)

// Multiple grouping columns
summary := df.GroupBy("department", "level").Agg(
    col("salary").Mean(),
)
```

### Joins

```go
// Inner join
result := df1.Join(df2, col("user_id"))

// Left join
result := df1.LeftJoin(df2, col("user_id"))

// Multiple join keys
result := df1.Join(df2, col("year"), col("month"))

// Different column names
result := df1.Join(df2,
    col("id").EqCol(col("user_id")),
)
```

### Pivot & Reshape

```go
// Pivot (wide format)
pivoted := df.Pivot(
    index: "date",
    columns: "category",
    values: "amount",
    aggFunc: Sum,
)

// Melt (long format)
melted := df.Melt(
    idVars: []string{"id", "date"},
    valueVars: []string{"revenue", "cost", "profit"},
)
```

### Window Functions (Phase 4)

```go
df = df.WithColumn("rank",
    col("salary").Rank().Over(col("department")),
)

df = df.WithColumn("running_total",
    col("amount").CumSum().Over(col("account")).OrderBy(col("date")),
)
```

## FrameSet (Multi-Sheet)

```go
type FrameSet struct {
    name   string
    frames map[string]*Frame
    order  []string
}

fs := omniframe.NewFrameSet("Q4 Report")
fs.AddFrame(revenueDF)    // Sheet: "Revenue" (from df.Name)
fs.AddFrame(expensesDF)   // Sheet: "Expenses"
fs.AddFrame(summaryDF)    // Sheet: "Summary"

// Write multi-sheet XLSX
err := fs.WriteXLSX("q4_report.xlsx")

// Custom sheet names
fs.AddFrameAs(df, "Custom Sheet Name")
```

## Formatting & Styling

```go
// Column formats (for XLSX export)
df.SetFormat("price", omniframe.FormatCurrency)     // $1,234.56
df.SetFormat("rate", omniframe.FormatPercent)       // 12.5%
df.SetFormat("date", omniframe.FormatDate)          // 2025-01-15
df.SetFormat("amount", omniframe.FormatNumber(2))   // 1,234.56

// Conditional styling (Phase 3)
df.SetStyle("status", func(val any, row Row) *CellStyle {
    switch val {
    case "error":
        return &CellStyle{Fill: "#FFCCCC"}
    case "warning":
        return &CellStyle{Fill: "#FFFFCC"}
    default:
        return nil
    }
})

// Column widths
df.SetColumnWidth("description", 50)
df.SetColumnWidthAuto("name")
```

## Phased Roadmap

### Phase 1: XLSX Export MVP

**Goal:** Multi-sheet XLSX export for immediate use in prism project

- [x] `Frame` struct with columnar storage
- [x] `FrameSet` for multi-sheet workbooks
- [ ] `Frame.WriteXLSX()` single sheet
- [ ] `FrameSet.WriteXLSX()` multi-sheet
- [ ] Column type support: string, int64, float64, bool, time.Time
- [ ] Basic formatting (number, percent, date, currency)
- [ ] Column width control

**Dependencies:** `github.com/xuri/excelize/v2`

### Phase 2: Core I/O & Eager API

**Goal:** Full read/write support, basic DataFrame operations

- [ ] CSV read/write
- [ ] XLSX read
- [ ] JSON read/write
- [ ] TOON read/write
- [ ] `FromColumns()`, `FromRows()`, `FromStructs()`
- [ ] Column access: `Col()`, `Strings()`, `Int64s()`, etc.
- [ ] `Filter()` with function
- [ ] `Select()`, `Drop()`, `Rename()`
- [ ] `Head()`, `Tail()`, `Sample()`, `Slice()`
- [ ] `Sort()`, `SortDesc()`, `SortBy()`

### Phase 3: Expression System & Aggregation

**Goal:** Polars-style expressions, GroupBy, type inference

- [ ] Expression DSL: `col()`, `lit()`, comparisons, arithmetic
- [ ] `Filter()` with expressions
- [ ] `WithColumn()` for computed columns
- [ ] `GroupBy().Agg()` with Sum, Mean, Count, Min, Max
- [ ] Type inference from data
- [ ] Null handling with `Option[T]` semantics
- [ ] SQL read/write

### Phase 4: Lazy Evaluation & Optimization

**Goal:** Deferred execution with query optimization

- [ ] `LazyFrame` type
- [ ] `Scan*()` functions for lazy loading
- [ ] Logical plan representation
- [ ] Query optimizer: predicate pushdown, projection pushdown
- [ ] `Collect()` to execute plan
- [ ] `Explain()` to view plan
- [ ] Parquet read/write

### Phase 5: Joins & Reshaping

**Goal:** Complete relational operations

- [ ] Inner, Left, Right, Outer joins
- [ ] Multi-key joins
- [ ] Cross joins
- [ ] Pivot (long to wide)
- [ ] Melt (wide to long)
- [ ] Concat, Union

### Phase 6: Advanced & Performance

**Goal:** High-performance features, extended backends

- [ ] Apache Arrow backend (optional)
- [ ] Parallel execution engine
- [ ] Window functions (Rank, RowNumber, CumSum, etc.)
- [ ] Apply/Transform with UDFs
- [ ] Streaming for large datasets
- [ ] GoNum matrix interop
- [ ] DuckDB integration
- [ ] Cassandra connector
- [ ] BigQuery connector

## Dependencies

### Phase 1

```go
require (
    github.com/xuri/excelize/v2 v2.9.0
)
```

### Phase 4+

```go
require (
    github.com/apache/arrow-go/v18
    github.com/xitongsys/parquet-go
    github.com/marcboeker/go-duckdb
)
```

## Success Metrics

1. **Adoption** - GitHub stars, go.pkg.dev imports
2. **Completeness** - % of pandas/Polars operations supported
3. **Performance** - Benchmarks vs gota, dataframe-go, and Python alternatives
4. **Stability** - Test coverage > 80%
5. **Ergonomics** - API usability feedback

## Non-Goals (Initially)

- GPU acceleration (CUDA/Metal)
- Distributed computing (Spark-like clustering)
- Real-time streaming ingestion
- Built-in visualization
- Machine learning integration

## References

- [Polars Documentation](https://docs.pola.rs/)
- [Polars GitHub](https://github.com/pola-rs/polars)
- [Polars Architecture Deep Dive](https://endjin.com/blog/2026/01/under-the-hood-what-makes-polars-so-scalable-and-fast)
- [Apache Arrow Go](https://github.com/apache/arrow-go)
- [pandas Documentation](https://pandas.pydata.org/docs/)
- [gota (archived)](https://github.com/go-gota/gota)
- [dataframe-go](https://github.com/rocketlaunchr/dataframe-go)
