package widgets

import (
	"testing"
)

func TestBuildGrid_DefaultOneColumn(t *testing.T) {
	grid := BuildGrid().Build()
	if grid == nil {
		t.Fatal("expected non-nil Grid from BuildGrid")
	}
	if grid.Cols != 1 {
		t.Fatalf("expected 1 column by default, got %d", grid.Cols)
	}
	// No rows added, but at least 1 row is forced.
	if grid.Rows != 1 {
		t.Fatalf("expected 1 row for empty builder, got %d", grid.Rows)
	}
}

func TestBuildGrid_Columns(t *testing.T) {
	grid := BuildGrid().Columns(4).Build()
	if grid.Cols != 4 {
		t.Fatalf("expected 4 columns, got %d", grid.Cols)
	}
}

func TestBuildGrid_ColumnsZeroIgnored(t *testing.T) {
	grid := BuildGrid().Columns(0).Build()
	// Zero should be ignored; stays at default 1.
	if grid.Cols != 1 {
		t.Fatalf("expected 1 column when 0 provided, got %d", grid.Cols)
	}
}

func TestBuildGrid_ColumnsNegativeIgnored(t *testing.T) {
	grid := BuildGrid().Columns(-3).Build()
	if grid.Cols != 1 {
		t.Fatalf("expected 1 column when negative provided, got %d", grid.Cols)
	}
}

func TestBuildGrid_Gap(t *testing.T) {
	grid := BuildGrid().Gap(2).Build()
	if grid.Gap != 2 {
		t.Fatalf("expected gap 2, got %d", grid.Gap)
	}
}

func TestBuildGrid_SingleRow(t *testing.T) {
	w1 := NewFixedSpacer(1, 1)
	w2 := NewFixedSpacer(1, 1)
	grid := BuildGrid().
		Columns(2).
		Row(w1, w2).
		Build()

	if grid.Rows != 1 {
		t.Fatalf("expected 1 row, got %d", grid.Rows)
	}
	if len(grid.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(grid.Children))
	}

	// Verify positions.
	if grid.Children[0].Row != 0 || grid.Children[0].Col != 0 {
		t.Fatalf("expected child 0 at (0,0), got (%d,%d)", grid.Children[0].Row, grid.Children[0].Col)
	}
	if grid.Children[1].Row != 0 || grid.Children[1].Col != 1 {
		t.Fatalf("expected child 1 at (0,1), got (%d,%d)", grid.Children[1].Row, grid.Children[1].Col)
	}
}

func TestBuildGrid_MultipleRows(t *testing.T) {
	w1 := NewFixedSpacer(1, 1)
	w2 := NewFixedSpacer(1, 1)
	w3 := NewFixedSpacer(1, 1)
	w4 := NewFixedSpacer(1, 1)
	grid := BuildGrid().
		Columns(2).
		Row(w1, w2).
		Row(w3, w4).
		Build()

	if grid.Rows != 2 {
		t.Fatalf("expected 2 rows, got %d", grid.Rows)
	}
	if len(grid.Children) != 4 {
		t.Fatalf("expected 4 children, got %d", len(grid.Children))
	}

	// Check second row positions.
	if grid.Children[2].Row != 1 || grid.Children[2].Col != 0 {
		t.Fatalf("expected child 2 at (1,0), got (%d,%d)", grid.Children[2].Row, grid.Children[2].Col)
	}
	if grid.Children[3].Row != 1 || grid.Children[3].Col != 1 {
		t.Fatalf("expected child 3 at (1,1), got (%d,%d)", grid.Children[3].Row, grid.Children[3].Col)
	}
}

func TestBuildGrid_FewerWidgetsThanColumns(t *testing.T) {
	w1 := NewFixedSpacer(1, 1)
	grid := BuildGrid().
		Columns(3).
		Row(w1). // Only 1 widget for 3 columns.
		Build()

	if len(grid.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(grid.Children))
	}
	if grid.Children[0].Col != 0 {
		t.Fatalf("expected child at col 0, got %d", grid.Children[0].Col)
	}
}

func TestBuildGrid_MoreWidgetsThanColumns(t *testing.T) {
	w1 := NewFixedSpacer(1, 1)
	w2 := NewFixedSpacer(1, 1)
	w3 := NewFixedSpacer(1, 1)
	grid := BuildGrid().
		Columns(2).
		Row(w1, w2, w3). // 3 widgets, 2 columns; w3 should be ignored.
		Build()

	if len(grid.Children) != 2 {
		t.Fatalf("expected 2 children (excess ignored), got %d", len(grid.Children))
	}
}

func TestBuildGrid_SpanRow(t *testing.T) {
	header := NewFixedSpacer(1, 1)
	w1 := NewFixedSpacer(1, 1)
	w2 := NewFixedSpacer(1, 1)
	grid := BuildGrid().
		Columns(3).
		SpanRow(header).
		Row(w1, w2).
		Build()

	if grid.Rows != 2 {
		t.Fatalf("expected 2 rows, got %d", grid.Rows)
	}

	// First child is the span row.
	span := grid.Children[0]
	if span.ColSpan != 3 {
		t.Fatalf("expected span row ColSpan 3, got %d", span.ColSpan)
	}
	if span.Row != 0 || span.Col != 0 {
		t.Fatalf("expected span row at (0,0), got (%d,%d)", span.Row, span.Col)
	}

	// Second row children should be at row 1.
	if grid.Children[1].Row != 1 {
		t.Fatalf("expected second row children at row 1, got %d", grid.Children[1].Row)
	}
}

func TestBuildGrid_NilWidgetsSkipped(t *testing.T) {
	w1 := NewFixedSpacer(1, 1)
	grid := BuildGrid().
		Columns(3).
		Row(w1, nil, nil).
		Build()

	// nil widgets should be skipped.
	if len(grid.Children) != 1 {
		t.Fatalf("expected 1 child (nils skipped), got %d", len(grid.Children))
	}
}

func TestBuildGrid_SpanRowNilChild(t *testing.T) {
	grid := BuildGrid().
		Columns(3).
		SpanRow(nil).
		Build()

	// Nil span child should be skipped.
	if len(grid.Children) != 0 {
		t.Fatalf("expected 0 children with nil span, got %d", len(grid.Children))
	}
}

func TestBuildGrid_ChainingFluency(t *testing.T) {
	// Verify the fluent API returns the same builder at every step.
	b := BuildGrid()
	b2 := b.Columns(2)
	if b != b2 {
		t.Fatal("expected Columns to return same builder")
	}
	b3 := b.Gap(1)
	if b != b3 {
		t.Fatal("expected Gap to return same builder")
	}
	w := NewFixedSpacer(1, 1)
	b4 := b.Row(w)
	if b != b4 {
		t.Fatal("expected Row to return same builder")
	}
	b5 := b.SpanRow(w)
	if b != b5 {
		t.Fatal("expected SpanRow to return same builder")
	}
}

func TestBuildGrid_ComplexLayout(t *testing.T) {
	// Build a typical 3-column form layout:
	// Row 0: header (spans all 3)
	// Row 1: field1, field2, field3
	// Row 2: footer (spans all 3)
	header := NewFixedSpacer(80, 1)
	f1 := NewFixedSpacer(20, 1)
	f2 := NewFixedSpacer(20, 1)
	f3 := NewFixedSpacer(20, 1)
	footer := NewFixedSpacer(80, 1)

	grid := BuildGrid().
		Columns(3).
		Gap(1).
		SpanRow(header).
		Row(f1, f2, f3).
		SpanRow(footer).
		Build()

	if grid.Rows != 3 {
		t.Fatalf("expected 3 rows, got %d", grid.Rows)
	}
	if grid.Cols != 3 {
		t.Fatalf("expected 3 cols, got %d", grid.Cols)
	}
	if grid.Gap != 1 {
		t.Fatalf("expected gap 1, got %d", grid.Gap)
	}
	// 1 (header span) + 3 (fields) + 1 (footer span) = 5 children.
	if len(grid.Children) != 5 {
		t.Fatalf("expected 5 children, got %d", len(grid.Children))
	}
}
