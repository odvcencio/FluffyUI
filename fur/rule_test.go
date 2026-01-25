package fur

import (
	"strings"
	"testing"
)

func TestRuleNoTitle(t *testing.T) {
	r := Rule()
	lines := r.Render(20)

	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	text := extractText(lines)
	text = strings.TrimSpace(text)
	// Rule uses Unicode box characters - verify string width, not byte length
	width := 0
	for range text {
		width++
	}
	if width != 20 {
		t.Errorf("expected 20 runes, got %d", width)
	}
	if !strings.Contains(text, "─") {
		t.Error("expected line character")
	}
}

func TestRuleWithTitle(t *testing.T) {
	r := Rule("Test")
	lines := r.Render(20)

	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	text := extractText(lines)
	text = strings.TrimSpace(text)
	if !strings.Contains(text, "Test") {
		t.Errorf("expected title, got %q", text)
	}
}

func TestRuleWithAlignment(t *testing.T) {
	r := RuleWith("Left", RuleOpts{Align: AlignLeft})
	lines := r.Render(20)

	text := extractText(lines)
	text = strings.TrimSpace(text)
	// Title should be near the start
	idx := strings.Index(text, "Left")
	if idx > 2 {
		t.Errorf("expected left alignment, title at index %d", idx)
	}
}

func TestRuleWithCustomChar(t *testing.T) {
	r := RuleWith("", RuleOpts{Character: '='})
	lines := r.Render(10)

	text := extractText(lines)
	if !strings.Contains(text, "=") {
		t.Errorf("expected = character, got %q", text)
	}
}
