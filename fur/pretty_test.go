package fur

import (
	"strings"
	"testing"
)

func TestPrettyString(t *testing.T) {
	r := Pretty("hello")
	lines := r.Render(80)

	if len(lines) == 0 {
		t.Fatal("expected at least 1 line")
	}
	text := extractText(lines)
	if !strings.Contains(text, `"hello"`) {
		t.Errorf("expected quoted string, got %q", text)
	}
}

func TestPrettyInt(t *testing.T) {
	r := Pretty(42)
	lines := r.Render(80)

	if len(lines) == 0 {
		t.Fatal("expected at least 1 line")
	}
	text := extractText(lines)
	if !strings.Contains(text, "42") {
		t.Errorf("expected 42, got %q", text)
	}
}

func TestPrettyStruct(t *testing.T) {
	type Config struct {
		Name string
		Port int
	}
	cfg := Config{Name: "test", Port: 8080}

	r := Pretty(cfg)
	lines := r.Render(80)

	text := extractText(lines)
	if !strings.Contains(text, "Name") {
		t.Errorf("expected Name field, got %q", text)
	}
	if !strings.Contains(text, "test") {
		t.Errorf("expected test value, got %q", text)
	}
	if !strings.Contains(text, "Port") {
		t.Errorf("expected Port field, got %q", text)
	}
	if !strings.Contains(text, "8080") {
		t.Errorf("expected 8080 value, got %q", text)
	}
}

func TestPrettySlice(t *testing.T) {
	r := Pretty([]string{"a", "b", "c"})
	lines := r.Render(80)

	text := extractText(lines)
	if !strings.Contains(text, `"a"`) {
		t.Errorf("expected a, got %q", text)
	}
	if !strings.Contains(text, `"b"`) {
		t.Errorf("expected b, got %q", text)
	}
}

func TestPrettyMap(t *testing.T) {
	r := Pretty(map[string]int{"x": 1, "y": 2})
	lines := r.Render(80)

	text := extractText(lines)
	if !strings.Contains(text, "x") {
		t.Errorf("expected x key, got %q", text)
	}
	if !strings.Contains(text, "y") {
		t.Errorf("expected y key, got %q", text)
	}
}

func TestPrettyNil(t *testing.T) {
	r := Pretty(nil)
	lines := r.Render(80)

	text := extractText(lines)
	if !strings.Contains(text, "nil") {
		t.Errorf("expected nil, got %q", text)
	}
}

func TestPrettyPointer(t *testing.T) {
	value := 42
	r := Pretty(&value)
	lines := r.Render(80)

	text := extractText(lines)
	if !strings.Contains(text, "&") {
		t.Errorf("expected & prefix, got %q", text)
	}
	if !strings.Contains(text, "42") {
		t.Errorf("expected 42, got %q", text)
	}
}

func TestPrettyMaxDepth(t *testing.T) {
	type Nested struct {
		Child *Nested
	}
	deep := &Nested{Child: &Nested{Child: &Nested{Child: &Nested{}}}}

	r := PrettyWith(deep, PrettyOpts{MaxDepth: 2})
	lines := r.Render(80)

	text := extractText(lines)
	if !strings.Contains(text, "...") {
		t.Errorf("expected truncation at depth, got %q", text)
	}
}

func TestPrettyCycleDetection(t *testing.T) {
	type Node struct {
		Next *Node
	}
	a := &Node{}
	a.Next = a // self-cycle

	r := Pretty(a)
	lines := r.Render(80)

	text := extractText(lines)
	if !strings.Contains(text, "<cycle>") {
		t.Errorf("expected cycle detection, got %q", text)
	}
}

func extractText(lines []Line) string {
	var sb strings.Builder
	for _, line := range lines {
		for _, span := range line {
			sb.WriteString(span.Text)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}
