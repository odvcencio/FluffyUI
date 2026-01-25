package fur

import (
	"testing"
)

func TestColumnsBasic(t *testing.T) {
	r := Columns(
		Text("Left"),
		Text("Right"),
	)
	lines := r.Render(40)

	if len(lines) == 0 {
		t.Fatal("expected at least one line")
	}
	text := extractText(lines)
	if len(text) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestColumnsEmpty(t *testing.T) {
	r := Columns()
	lines := r.Render(40)

	if lines != nil {
		t.Error("expected nil for empty columns")
	}
}

func TestColumnsWithOpts(t *testing.T) {
	r := ColumnsWith([]Renderable{
		Text("A"),
		Text("B"),
		Text("C"),
	}, ColumnsOpts{
		Padding: 4,
		Equal:   true,
	})
	lines := r.Render(60)

	if len(lines) == 0 {
		t.Fatal("expected output")
	}
}

func TestColumnsFixedWidth(t *testing.T) {
	r := ColumnsWith([]Renderable{
		Text("Column1"),
		Text("Column2"),
	}, ColumnsOpts{
		Width: 10,
	})
	lines := r.Render(80)

	if len(lines) == 0 {
		t.Fatal("expected output")
	}
}

func TestColumnsExpand(t *testing.T) {
	r := ColumnsWith([]Renderable{
		Text("A"),
		Text("B"),
	}, ColumnsOpts{
		Expand: true,
	})
	lines := r.Render(40)

	if len(lines) == 0 {
		t.Fatal("expected output")
	}
}

func TestColumnsMultiline(t *testing.T) {
	r := Columns(
		Text("Line1\nLine2\nLine3"),
		Text("Single"),
	)
	lines := r.Render(40)

	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines, got %d", len(lines))
	}
}

func TestColumnsNilItem(t *testing.T) {
	r := Columns(
		Text("Valid"),
		nil,
		Text("Also Valid"),
	)
	lines := r.Render(40)

	if len(lines) == 0 {
		t.Fatal("expected output even with nil item")
	}
}
