# Fur Rich Format Renderers

The `fur` package provides rich format renderers for displaying structured data beautifully in the terminal. These work standalone or integrate with FluffyUI widgets.

---

## Table of Contents

- [CSV Tables](#csv-tables)
- [Diff View](#diff-view)
- [JSON Formatter](#json-formatter)
- [Data Visualization](#data-visualization)
- [Simple Diagrams](#simple-diagrams)

---

## CSV Tables

Render CSV data as beautifully formatted tables with Unicode box-drawing characters.

```go
import "github.com/odvcencio/fluffyui/fur"

// From CSV string
csvData := `Name,Role,Department
Alice,Engineer,Engineering
Bob,Designer,Design`

table := fur.CSVTable(csvData, true) // true = has header
fur.Default().Render(table)
```

**Output:**
```
╭───────┬───────────┬────────────╮
│ Name  │ Role      │ Department │
├───────┼───────────┼────────────┤
│ Alice │ Engineer  │ Engineering│
│ Bob   │ Designer  │ Design     │
╰───────┴───────────┴────────────╯
```

### Table Styles

```go
// Default style with rounded corners
table := fur.CSVTable(csvData, true)

// Compact style - minimal borders
table := fur.CSVTableFromRecords(records, true).
    WithStyle(fur.CompactTableStyle())

// Custom style
style := fur.TableStyle{
    BoxDrawings: fur.HeavyBoxDrawings,  // Thick borders
    HeaderStyle: fur.Style{}.Bold().Foreground(fur.ColorCyan),
    CellStyle:   fur.DefaultStyle(),
    BorderStyle: fur.Style{}.Foreground(fur.ColorBrightBlack),
    Padding:     1,
    MinColWidth: 3,
}
table := fur.CSVTableFromRecords(records, true).WithStyle(style)
```

**Available Box Styles:**
- `RoundedBoxDrawings` — Rounded corners (╭╮╰╯)
- `SharpBoxDrawings` — Sharp corners (┌┐└┘)
- `HeavyBoxDrawings` — Thick lines (┏┓┗┛)
- `DoubleBoxDrawings` — Double lines (╔╗╚╝)

---

## Diff View

Syntax-highlighted diff output for git diffs or unified diffs.

```go
diff := `diff --git a/main.go b/main.go
index 3a4b5c6..7d8e9f0 100644
--- a/main.go
+++ b/main.go
@@ -10,7 +10,7 @@ func main() {
 	fmt.Println("Hello!")
 
-	oldFunction()
+	newFunction()
 }`

fur.Default().Render(fur.Diff(diff))
```

**Output:**
```
diff --git a/main.go b/main.go           ← Magenta (file header)
index 3a4b5c6..7d8e9f0 100644          ← Yellow (metadata)
--- a/main.go                          ← Red (old file)
+++ b/main.go                          ← Green (new file)
@@ -10,7 +10,7 @@ func main() {        ← Cyan (hunk header)
 	fmt.Println("Hello!")               ← Default (context)
 
-	oldFunction()                      ← Red (deleted)
+	newFunction()                      ← Green (added)
```

### Diff Statistics

```go
stats := fur.DiffStats(added, deleted, modified)
fur.Default().Render(stats)
// Output: +12 added, -5 deleted, ~3 modified
```

---

## JSON Formatter

Pretty-print JSON with syntax highlighting.

```go
jsonData := `{"name":"FluffyUI","version":"1.0.0","features":["widgets","graphics"]}`

// Pretty-printed with indentation
fur.Default().Render(fur.JSON(jsonData))

// From a Go value
fur.Default().Render(fur.JSONFromValue(myStruct))
```

**Output:**
```json
{
  "features": [
    "widgets",     ← Green (strings)
    "graphics"
  ],
  "name": "FluffyUI",
  "version": "1.0.0"  ← Cyan (keys)
}
```

**Color Scheme:**
- Keys: Cyan
- Strings: Green
- Numbers: Yellow
- Booleans: Magenta
- Null: Bright Black (dim)

---

## Data Visualization

### Sparkline

Compact trend visualization using Unicode block characters.

```go
values := []float64{10, 25, 18, 32, 45, 38, 52, 48, 61, 55, 70}
sparkline := fur.Sparkline(values, 40)
fur.Default().Render(sparkline)
// Output: ▁▂▄▃▄▄▅▄▆▅▇
```

### Bar Chart

Horizontal bar charts for comparing values.

```go
barChart := fur.BarChart(
    []string{"Jan", "Feb", "Mar", "Apr", "May"},
    []float64{120, 190, 150, 220, 280},
    30, // max bar width
)
fur.Default().Render(barChart)
```

**Output:**
```
Jan ████████████ 120.0
Feb ████████████████████ 190.0
Mar ████████████████ 150.0
Apr ███████████████████████ 220.0
May ██████████████████████████████ 280.0
```

### Progress Bar

Visual progress indicator with color coding.

```go
progress := fur.ProgressBar(75, 100, 30)
fur.Default().Render(progress)
// Output: [██████████████████████▌░░░░░░░]  75%
```

Colors change based on progress:
- < 30%: Red
- 30-70%: Yellow
- > 70%: Green

### Bullet Graph

Shows actual vs target values in a compact format.

```go
bullet := fur.BulletGraph(65, 80, 100, 40)
// actual=65, target=80, max=100, width=40
fur.Default().Render(bullet)
// Output: █████████████████████████●░░░░░│░░░░░░░░ 65/80
```

### Statistics

Calculate and display summary statistics.

```go
data := []float64{12.5, 18.2, 15.8, 22.1, 19.5, 14.3, 16.7, 20.9}
stats := fur.Statistics(data)
fur.Default().Render(stats)
```

**Output:**
```
Count:    8
Min:      12.50
Max:      22.10
Mean:     17.50
Std Dev:  3.08
```

### Heatmap

2D data visualization with color gradients.

```go
data := [][]float64{
    {0.2, 0.4, 0.6, 0.8, 1.0},
    {0.1, 0.3, 0.5, 0.7, 0.9},
    {0.3, 0.5, 0.7, 0.9, 0.6},
}
heatmap := fur.Heatmap(data, 40)
fur.Default().Render(heatmap)
```

**Output:**
```
░▒▒▓█
░░▒▓▓
░▒▓▓▒
```

Uses blue→green→yellow→red gradient for values.

---

## Simple Diagrams

### Flow Diagrams

Simple text-based flowcharts.

```go
flow := fur.SimpleDiagram("Start -> Process -> Decision -> End")
fur.Default().Render(flow)
// Output: [ Start ] --> [ Process ] --> [ Decision ] --> [ End ]
```

### Mermaid Flowcharts

Basic support for Mermaid-style flowchart syntax.

```go
diagram := `
A[Start] --> B{Decision}
B -->|Yes| C[Process]
B -->|No| D[End]
C --> D
`
fur.Default().Render(fur.MermaidFlowchart(diagram))
```

**Supported Syntax:**
- `A --> B` — Simple connection
- `A -->|label| B` — Labeled connection
- `A[Label]` — Box node
- `A(Label)` — Rounded node
- `A{Label}` — Diamond node

---

## Integration with FluffyUI

All renderers work inside FluffyUI widgets:

```go
type DataPanel struct {
    widgets.Base
    tableData string
}

func (p *DataPanel) Render(ctx runtime.RenderContext) {
    bounds := p.ContentBounds()
    
    // Create CSV table
    table := fur.CSVTable(p.tableData, true)
    lines := table.Render(bounds.Width)
    
    // Render to buffer
    for y, line := range lines {
        if y >= bounds.Height {
            break
        }
        x := bounds.X
        for _, span := range line {
            style := span.Style.ToBackend()
            ctx.Buffer.SetString(x, bounds.Y+y, span.Text, style)
            x += len(span.Text)
        }
    }
}
```

---

## Complete Example

```go
package main

import (
    "github.com/odvcencio/fluffyui/fur"
)

func main() {
    c := fur.Default()
    
    // CSV Table
    c.Println("[bold]Employee Data:[/]")
    csv := `Name,Department,Salary
Alice,Engineering,145000
Bob,Design,115000`
    c.Render(fur.CSVTable(csv, true))
    c.Println()
    
    // Progress
    c.Println("[bold]Progress:[/]")
    c.Render(fur.ProgressBar(75, 100, 30))
    c.Println()
    
    // Sparkline
    c.Println("[bold]Traffic Trend:[/]")
    values := []float64{10, 25, 18, 32, 45, 38, 52, 48, 61}
    c.Render(fur.Sparkline(values, 40))
    c.Println()
    
    // Statistics
    c.Println("[bold]Statistics:[/]")
    data := []float64{12.5, 18.2, 15.8, 22.1, 19.5}
    c.Render(fur.Statistics(data))
}
```

**Output:**
```
Employee Data:
╭───────┬─────────────┬────────╮
│ Name  │ Department  │ Salary │
├───────┼─────────────┼────────┤
│ Alice │ Engineering │ 145000 │
│ Bob   │ Design      │ 115000 │
╰───────┴─────────────┴────────╯

Progress:
[██████████████████████▌░░░░░░░]  75%

Traffic Trend:
▁▂▄▃▄▄▅▄▆

Statistics:
  Count:    5
  Min:      12.50
  Max:      22.10
  Mean:     17.62
  Std Dev:  3.62
```

---

## API Reference

### CSV

```go
func CSVTable(data string, hasHeader bool) Renderable
func CSVTableFromRecords(records [][]string, hasHeader bool) csvTableRenderable
```

### Diff

```go
func Diff(content string) Renderable
func DiffStats(added, deleted, modified int) Renderable
```

### JSON

```go
func JSON(data string) jsonRenderable
func JSONFromValue(v any) jsonRenderable
func JSONCompact(data string) Renderable
```

### Charts

```go
func Sparkline(values []float64, width int) sparklineRenderable
func BarChart(labels []string, values []float64, maxWidth int) barChartRenderable
func PieChart(values []float64, labels []string) Renderable
func Heatmap(data [][]float64, width int) heatmapRenderable
```

### Progress

```go
func ProgressBar(current, total float64, width int) progressBarRenderable
func BulletGraph(actual, target, max float64, width int) Renderable
func Gauge(value, min, max float64, width int) Renderable
```

### Statistics

```go
func Statistics(values []float64) Renderable
```

### Diagrams

```go
func SimpleDiagram(description string) Renderable
func MermaidFlowchart(diagram string) Renderable
```
