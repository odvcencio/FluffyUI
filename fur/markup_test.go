package fur

import (
	"testing"
)

func TestMarkupParserBasic(t *testing.T) {
	p := NewMarkupParser()
	lines := p.Parse("Hello")

	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if len(lines[0]) != 1 {
		t.Fatalf("expected 1 span, got %d", len(lines[0]))
	}
	if lines[0][0].Text != "Hello" {
		t.Errorf("got %q, want %q", lines[0][0].Text, "Hello")
	}
}

func TestMarkupParserBold(t *testing.T) {
	p := NewMarkupParser()
	lines := p.Parse("[bold]Bold[/]")

	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if len(lines[0]) != 1 {
		t.Fatalf("expected 1 span, got %d", len(lines[0]))
	}
	if lines[0][0].Text != "Bold" {
		t.Errorf("got %q, want %q", lines[0][0].Text, "Bold")
	}
	if !lines[0][0].Style.bold {
		t.Error("expected bold style")
	}
}

func TestMarkupParserColor(t *testing.T) {
	p := NewMarkupParser()
	lines := p.Parse("[red]Red[/]")

	if len(lines) != 1 || len(lines[0]) != 1 {
		t.Fatalf("unexpected structure")
	}
	if lines[0][0].Text != "Red" {
		t.Errorf("got %q, want %q", lines[0][0].Text, "Red")
	}
	if lines[0][0].Style.fg != ColorRed {
		t.Errorf("expected red foreground, got %v", lines[0][0].Style.fg)
	}
}

func TestMarkupParserHexColor(t *testing.T) {
	p := NewMarkupParser()
	lines := p.Parse("[#ff6600]Orange[/]")

	if len(lines) != 1 || len(lines[0]) != 1 {
		t.Fatalf("unexpected structure")
	}
	if lines[0][0].Text != "Orange" {
		t.Errorf("got %q, want %q", lines[0][0].Text, "Orange")
	}
}

func TestMarkupParserBackground(t *testing.T) {
	p := NewMarkupParser()
	lines := p.Parse("[on blue]BG[/]")

	if len(lines) != 1 || len(lines[0]) != 1 {
		t.Fatalf("unexpected structure")
	}
	if lines[0][0].Style.bg != ColorBlue {
		t.Errorf("expected blue background, got %v", lines[0][0].Style.bg)
	}
}

func TestMarkupParserMultipleStyles(t *testing.T) {
	p := NewMarkupParser()
	lines := p.Parse("[bold italic red]Styled[/]")

	if len(lines) != 1 || len(lines[0]) != 1 {
		t.Fatalf("unexpected structure")
	}
	style := lines[0][0].Style
	if !style.bold {
		t.Error("expected bold")
	}
	if !style.italic {
		t.Error("expected italic")
	}
	if style.fg != ColorRed {
		t.Errorf("expected red, got %v", style.fg)
	}
}

func TestMarkupParserNewlines(t *testing.T) {
	p := NewMarkupParser()
	lines := p.Parse("Line1\nLine2")

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0][0].Text != "Line1" {
		t.Errorf("got %q, want %q", lines[0][0].Text, "Line1")
	}
	if lines[1][0].Text != "Line2" {
		t.Errorf("got %q, want %q", lines[1][0].Text, "Line2")
	}
}

func TestMarkupParserEscape(t *testing.T) {
	p := NewMarkupParser()
	lines := p.Parse(`\[not a tag\]`)

	if len(lines) != 1 || len(lines[0]) != 1 {
		t.Fatalf("unexpected structure")
	}
	if lines[0][0].Text != "[not a tag]" {
		t.Errorf("got %q, want %q", lines[0][0].Text, "[not a tag]")
	}
}

func TestMarkupParserNested(t *testing.T) {
	p := NewMarkupParser()
	lines := p.Parse("[bold]Bold [italic]BoldItalic[/italic] Bold[/bold]")

	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	// Should have 3 spans: "Bold ", "BoldItalic", " Bold"
	if len(lines[0]) < 3 {
		t.Fatalf("expected at least 3 spans, got %d", len(lines[0]))
	}
}
