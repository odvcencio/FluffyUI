package widgets

import "m31labs.dev/fluffyui/runtime"

// GridBuilder provides a declarative API for constructing grids.
//
// Usage:
//
//	grid := BuildGrid().
//	    Columns(3).
//	    Gap(1).
//	    Row(nameInput, emailInput, phoneInput).
//	    Row(addressInput, cityInput, zipInput).
//	    Build()
type GridBuilder struct {
	columns int
	gap     int
	rows    []gridBuilderRow
}

// gridBuilderRow represents a row in the builder with optional span metadata.
type gridBuilderRow struct {
	children []runtime.Widget
	span     bool // if true, the single child spans all columns
}

// BuildGrid starts a new grid builder with a default of 1 column.
func BuildGrid() *GridBuilder {
	return &GridBuilder{columns: 1}
}

// Columns sets the number of columns in the grid.
func (b *GridBuilder) Columns(n int) *GridBuilder {
	if n > 0 {
		b.columns = n
	}
	return b
}

// Gap sets the gap (in cells) between grid children.
func (b *GridBuilder) Gap(gap int) *GridBuilder {
	b.gap = gap
	return b
}

// Row adds a row of widgets placed left-to-right into columns.
// If fewer widgets than columns are provided, the remaining cells are empty.
// If more widgets than columns are provided, excess widgets are ignored.
func (b *GridBuilder) Row(children ...runtime.Widget) *GridBuilder {
	b.rows = append(b.rows, gridBuilderRow{children: children})
	return b
}

// SpanRow adds a single widget that spans all columns in the row.
func (b *GridBuilder) SpanRow(child runtime.Widget) *GridBuilder {
	b.rows = append(b.rows, gridBuilderRow{
		children: []runtime.Widget{child},
		span:     true,
	})
	return b
}

// Build constructs the Grid widget from the builder configuration.
func (b *GridBuilder) Build() *Grid {
	totalRows := len(b.rows)
	if totalRows == 0 {
		totalRows = 1
	}
	grid := NewGrid(totalRows, b.columns)
	grid.Gap = b.gap

	for rowIdx, row := range b.rows {
		if row.span {
			// SpanRow: single child spans all columns.
			if len(row.children) > 0 && row.children[0] != nil {
				grid.Add(row.children[0], rowIdx, 0, 1, b.columns)
			}
			continue
		}
		for colIdx, child := range row.children {
			if colIdx >= b.columns {
				break
			}
			if child != nil {
				grid.Add(child, rowIdx, colIdx, 1, 1)
			}
		}
	}

	return grid
}
