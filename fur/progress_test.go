package fur

import (
	"strings"
	"testing"
)

func TestProgressBasic(t *testing.T) {
	p := NewProgress(100)
	p.Set(50)

	lines := p.Render(40)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}

	text := extractText(lines)
	if !strings.Contains(text, "50%") {
		t.Errorf("expected 50%% in output, got %q", text)
	}
}

func TestProgressWithLabel(t *testing.T) {
	p := NewProgress(100).WithLabel("Downloading")
	p.Set(25)

	lines := p.Render(50)
	text := extractText(lines)

	if !strings.Contains(text, "Downloading") {
		t.Errorf("expected label in output, got %q", text)
	}
}

func TestProgressZero(t *testing.T) {
	p := NewProgress(100)
	p.Set(0)

	lines := p.Render(40)
	text := extractText(lines)

	if !strings.Contains(text, "0%") {
		t.Errorf("expected 0%% in output, got %q", text)
	}
}

func TestProgressComplete(t *testing.T) {
	p := NewProgress(100)
	p.Set(100)

	lines := p.Render(40)
	text := extractText(lines)

	if !strings.Contains(text, "100%") {
		t.Errorf("expected 100%% in output, got %q", text)
	}
}

func TestProgressOverflow(t *testing.T) {
	p := NewProgress(100)
	p.Set(150) // Over the total

	lines := p.Render(40)
	text := extractText(lines)

	// Should clamp to 100%
	if !strings.Contains(text, "100%") {
		t.Errorf("expected clamped to 100%%, got %q", text)
	}
}

func TestProgressNegative(t *testing.T) {
	p := NewProgress(100)
	p.Set(-10) // Negative value

	lines := p.Render(40)
	text := extractText(lines)

	// Should clamp to 0%
	if !strings.Contains(text, "0%") {
		t.Errorf("expected clamped to 0%%, got %q", text)
	}
}

func TestProgressHidePercent(t *testing.T) {
	p := NewProgress(100).WithShowPercent(false)
	p.Set(50)

	lines := p.Render(40)
	text := extractText(lines)

	if strings.Contains(text, "%") {
		t.Errorf("percent should be hidden, got %q", text)
	}
}

func TestProgressNil(t *testing.T) {
	var p *Progress

	lines := p.Render(40)
	if lines != nil {
		t.Error("nil Progress should return nil lines")
	}

	// These should not panic
	p.Set(50)
	p.WithLabel("test")
	p.WithShowPercent(true)
}

func TestProgressZeroTotal(t *testing.T) {
	p := NewProgress(0)
	p.Set(0)

	lines := p.Render(40)
	if len(lines) == 0 {
		t.Error("should render even with zero total")
	}
}

func TestProgressNegativeTotal(t *testing.T) {
	p := NewProgress(-100)
	p.Set(50)

	// Should handle gracefully
	lines := p.Render(40)
	if len(lines) == 0 {
		t.Error("should render even with negative total")
	}
}

func TestProgressBarCharacters(t *testing.T) {
	p := NewProgress(100)
	p.Set(50)

	lines := p.Render(40)
	text := extractText(lines)

	// Should contain bar characters
	if !strings.Contains(text, "[") || !strings.Contains(text, "]") {
		t.Error("expected bar brackets in output")
	}
	if !strings.Contains(text, "=") {
		t.Error("expected filled bar character")
	}
	if !strings.Contains(text, "-") {
		t.Error("expected empty bar character")
	}
}

func TestProgressDefaultWidth(t *testing.T) {
	p := NewProgress(100)
	p.Set(50)

	// Zero width should use default
	lines := p.Render(0)
	if len(lines) == 0 {
		t.Error("should render with default width when given 0")
	}
}

func TestProgressChaining(t *testing.T) {
	p := NewProgress(100).
		WithLabel("Test").
		WithShowPercent(true)

	if p == nil {
		t.Error("chaining should return non-nil")
	}
}
